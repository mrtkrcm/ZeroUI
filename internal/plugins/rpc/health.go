package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the health status of a plugin
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusStarting  HealthStatus = "starting"
	HealthStatusStopping  HealthStatus = "stopping"
)

// HealthCheck represents a plugin health check result
type HealthCheck struct {
	Status           HealthStatus
	LastCheck        time.Time
	ResponseTime     time.Duration
	Error            error
	ConsecutiveFails int
	Metadata         map[string]string
}

// PluginHealth manages health checks for plugins
type PluginHealth struct {
	mu            sync.RWMutex
	checks        map[string]*HealthCheck
	checkInterval time.Duration
	maxRetries    int
	timeout       time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewPluginHealth creates a new plugin health manager
func NewPluginHealth(checkInterval time.Duration, timeout time.Duration) *PluginHealth {
	return &PluginHealth{
		checks:        make(map[string]*HealthCheck),
		checkInterval: checkInterval,
		maxRetries:    3,
		timeout:       timeout,
		stopChan:      make(chan struct{}),
	}
}

// Start begins health checking for all registered plugins
func (ph *PluginHealth) Start(manager *PluginManager) {
	ph.wg.Add(1)
	go ph.healthCheckLoop(manager)
}

// Stop stops all health checks
func (ph *PluginHealth) Stop() {
	close(ph.stopChan)
	ph.wg.Wait()
}

// healthCheckLoop continuously checks plugin health
func (ph *PluginHealth) healthCheckLoop(manager *PluginManager) {
	defer ph.wg.Done()

	ticker := time.NewTicker(ph.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ph.stopChan:
			return
		case <-ticker.C:
			ph.checkAllPlugins(manager)
		}
	}
}

// checkAllPlugins checks health of all plugins
func (ph *PluginHealth) checkAllPlugins(manager *PluginManager) {
	plugins := manager.ListPlugins()

	var wg sync.WaitGroup
	for _, pluginName := range plugins {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			ph.checkPlugin(manager, name)
		}(pluginName)
	}
	wg.Wait()
}

// checkPlugin checks health of a single plugin
func (ph *PluginHealth) checkPlugin(manager *PluginManager, pluginName string) {
	ctx, cancel := context.WithTimeout(context.Background(), ph.timeout)
	defer cancel()

	startTime := time.Now()

	// Get plugin
	pluginInstance, exists := manager.GetPlugin(pluginName)
	if !exists {
		ph.updateHealth(pluginName, &HealthCheck{
			Status:    HealthStatusUnhealthy,
			LastCheck: time.Now(),
			Error:     fmt.Errorf("plugin not found"),
		})
		return
	}

	// Call GetInfo as health check
	info, err := pluginInstance.GetInfo(ctx)
	responseTime := time.Since(startTime)

	ph.mu.Lock()
	defer ph.mu.Unlock()

	check, exists := ph.checks[pluginName]
	if !exists {
		check = &HealthCheck{
			Metadata: make(map[string]string),
		}
		ph.checks[pluginName] = check
	}

	check.LastCheck = time.Now()
	check.ResponseTime = responseTime

	if err != nil {
		check.Error = err
		check.ConsecutiveFails++
		if check.ConsecutiveFails >= ph.maxRetries {
			check.Status = HealthStatusUnhealthy
			// Trigger restart if unhealthy
			go ph.handleUnhealthyPlugin(manager, pluginName)
		} else {
			check.Status = HealthStatusUnknown
		}
	} else {
		check.Status = HealthStatusHealthy
		check.Error = nil
		check.ConsecutiveFails = 0
		// Store plugin info as metadata
		if info != nil {
			check.Metadata["name"] = info.Name
			check.Metadata["version"] = info.Version
			check.Metadata["description"] = info.Description
		}
	}
}

// updateHealth updates the health status of a plugin
func (ph *PluginHealth) updateHealth(pluginName string, check *HealthCheck) {
	ph.mu.Lock()
	defer ph.mu.Unlock()
	ph.checks[pluginName] = check
}

// GetHealth returns the health status of a plugin
func (ph *PluginHealth) GetHealth(pluginName string) (*HealthCheck, bool) {
	ph.mu.RLock()
	defer ph.mu.RUnlock()
	check, exists := ph.checks[pluginName]
	return check, exists
}

// GetAllHealth returns health status of all plugins
func (ph *PluginHealth) GetAllHealth() map[string]*HealthCheck {
	ph.mu.RLock()
	defer ph.mu.RUnlock()

	result := make(map[string]*HealthCheck)
	for k, v := range ph.checks {
		result[k] = v
	}
	return result
}

// IsHealthy checks if a plugin is healthy
func (ph *PluginHealth) IsHealthy(pluginName string) bool {
	check, exists := ph.GetHealth(pluginName)
	return exists && check.Status == HealthStatusHealthy
}

// handleUnhealthyPlugin handles an unhealthy plugin by attempting restart
func (ph *PluginHealth) handleUnhealthyPlugin(manager *PluginManager, pluginName string) {
	// Log the issue
	fmt.Printf("Plugin %s is unhealthy, attempting restart...\n", pluginName)

	// Attempt to restart the plugin
	if err := manager.RestartPlugin(pluginName); err != nil {
		fmt.Printf("Failed to restart plugin %s: %v\n", pluginName, err)
	} else {
		fmt.Printf("Successfully restarted plugin %s\n", pluginName)
		// Reset health status
		ph.mu.Lock()
		if check, exists := ph.checks[pluginName]; exists {
			check.Status = HealthStatusStarting
			check.ConsecutiveFails = 0
		}
		ph.mu.Unlock()
	}
}

// WaitForHealthy waits for a plugin to become healthy
func (ph *PluginHealth) WaitForHealthy(pluginName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if ph.IsHealthy(pluginName) {
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for plugin %s to become healthy", pluginName)
			}
		}
	}
}
