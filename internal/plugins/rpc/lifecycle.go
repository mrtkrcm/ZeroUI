package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	plugin "github.com/hashicorp/go-plugin"
)

// PluginState represents the state of a plugin
type PluginState string

const (
	PluginStateUnknown      PluginState = "unknown"
	PluginStateInitializing PluginState = "initializing"
	PluginStateRunning      PluginState = "running"
	PluginStateStopping     PluginState = "stopping"
	PluginStateStopped      PluginState = "stopped"
	PluginStateError        PluginState = "error"
	PluginStateRestarting   PluginState = "restarting"
)

// PluginLifecycle manages the lifecycle of a plugin
type PluginLifecycle struct {
	Name         string
	State        PluginState
	Client       *plugin.Client
	Plugin       ConfigPlugin
	StartedAt    time.Time
	StoppedAt    time.Time
	RestartCount int
	LastError    error
	mu           sync.RWMutex
}

// NewPluginLifecycle creates a new plugin lifecycle manager
func NewPluginLifecycle(name string) *PluginLifecycle {
	return &PluginLifecycle{
		Name:  name,
		State: PluginStateUnknown,
	}
}

// Start starts the plugin
func (pl *PluginLifecycle) Start(client *plugin.Client, pluginInstance ConfigPlugin) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if pl.State == PluginStateRunning {
		return fmt.Errorf("plugin %s is already running", pl.Name)
	}

	pl.State = PluginStateInitializing
	pl.Client = client
	pl.Plugin = pluginInstance
	pl.StartedAt = time.Now()
	pl.LastError = nil

	// Verify plugin is responsive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := pluginInstance.GetInfo(ctx); err != nil {
		pl.State = PluginStateError
		pl.LastError = err
		return fmt.Errorf("plugin %s failed to initialize: %w", pl.Name, err)
	}

	pl.State = PluginStateRunning
	return nil
}

// Stop stops the plugin
func (pl *PluginLifecycle) Stop() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	if pl.State != PluginStateRunning {
		return fmt.Errorf("plugin %s is not running", pl.Name)
	}

	pl.State = PluginStateStopping

	if pl.Client != nil {
		pl.Client.Kill()
		pl.Client = nil
	}

	pl.Plugin = nil
	pl.StoppedAt = time.Now()
	pl.State = PluginStateStopped

	return nil
}

// Restart restarts the plugin
func (pl *PluginLifecycle) Restart(factory func() (*plugin.Client, ConfigPlugin, error)) error {
	pl.mu.Lock()
	pl.State = PluginStateRestarting
	pl.RestartCount++
	pl.mu.Unlock()

	// Stop if running
	if pl.State == PluginStateRunning {
		if err := pl.Stop(); err != nil {
			return fmt.Errorf("failed to stop plugin during restart: %w", err)
		}
	}

	// Wait a moment before restarting
	time.Sleep(100 * time.Millisecond)

	// Start with new client
	client, pluginInstance, err := factory()
	if err != nil {
		pl.mu.Lock()
		pl.State = PluginStateError
		pl.LastError = err
		pl.mu.Unlock()
		return fmt.Errorf("failed to restart plugin %s: %w", pl.Name, err)
	}

	return pl.Start(client, pluginInstance)
}

// GetState returns the current state
func (pl *PluginLifecycle) GetState() PluginState {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.State
}

// IsRunning checks if the plugin is running
func (pl *PluginLifecycle) IsRunning() bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.State == PluginStateRunning
}

// GetUptime returns the uptime of the plugin
func (pl *PluginLifecycle) GetUptime() time.Duration {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	if pl.State != PluginStateRunning {
		return 0
	}

	return time.Since(pl.StartedAt)
}

// GetInfo returns lifecycle information
func (pl *PluginLifecycle) GetInfo() map[string]interface{} {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	info := map[string]interface{}{
		"name":          pl.Name,
		"state":         string(pl.State),
		"restart_count": pl.RestartCount,
	}

	if pl.State == PluginStateRunning {
		info["started_at"] = pl.StartedAt
		info["uptime"] = time.Since(pl.StartedAt).String()
	}

	if !pl.StoppedAt.IsZero() {
		info["stopped_at"] = pl.StoppedAt
	}

	if pl.LastError != nil {
		info["last_error"] = pl.LastError.Error()
	}

	return info
}

// LifecycleManager manages lifecycle for multiple plugins
type LifecycleManager struct {
	plugins map[string]*PluginLifecycle
	mu      sync.RWMutex
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		plugins: make(map[string]*PluginLifecycle),
	}
}

// Register registers a plugin for lifecycle management
func (lm *LifecycleManager) Register(name string, lifecycle *PluginLifecycle) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.plugins[name] = lifecycle
}

// Unregister removes a plugin from lifecycle management
func (lm *LifecycleManager) Unregister(name string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.plugins, name)
}

// Get returns the lifecycle for a plugin
func (lm *LifecycleManager) Get(name string) (*PluginLifecycle, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	lifecycle, exists := lm.plugins[name]
	return lifecycle, exists
}

// StopAll stops all managed plugins
func (lm *LifecycleManager) StopAll() error {
	lm.mu.RLock()
	plugins := make([]*PluginLifecycle, 0, len(lm.plugins))
	for _, pl := range lm.plugins {
		plugins = append(plugins, pl)
	}
	lm.mu.RUnlock()

	var wg sync.WaitGroup
	var errors []error
	var errorMu sync.Mutex

	for _, pl := range plugins {
		wg.Add(1)
		go func(lifecycle *PluginLifecycle) {
			defer wg.Done()
			if lifecycle.IsRunning() {
				if err := lifecycle.Stop(); err != nil {
					errorMu.Lock()
					errors = append(errors, err)
					errorMu.Unlock()
				}
			}
		}(pl)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d plugins", len(errors))
	}

	return nil
}

// GetAllStates returns the states of all managed plugins
func (lm *LifecycleManager) GetAllStates() map[string]PluginState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	states := make(map[string]PluginState)
	for name, pl := range lm.plugins {
		states[name] = pl.GetState()
	}
	return states
}

// GetRunningPlugins returns names of all running plugins
func (lm *LifecycleManager) GetRunningPlugins() []string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var running []string
	for name, pl := range lm.plugins {
		if pl.IsRunning() {
			running = append(running, name)
		}
	}
	return running
}

// RestartUnhealthy restarts all unhealthy plugins
func (lm *LifecycleManager) RestartUnhealthy(health *PluginHealth, factory func(string) (*plugin.Client, ConfigPlugin, error)) error {
	lm.mu.RLock()
	plugins := make(map[string]*PluginLifecycle)
	for k, v := range lm.plugins {
		plugins[k] = v
	}
	lm.mu.RUnlock()

	var errors []error
	for name, pl := range plugins {
		if !health.IsHealthy(name) && pl.GetState() != PluginStateRestarting {
			if err := pl.Restart(func() (*plugin.Client, ConfigPlugin, error) {
				return factory(name)
			}); err != nil {
				errors = append(errors, fmt.Errorf("failed to restart %s: %w", name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("restart errors: %v", errors)
	}

	return nil
}
