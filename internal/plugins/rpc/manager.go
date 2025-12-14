package rpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// HandshakeConfig is used to prevent non-plugin binaries from connecting
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ZEROUI_PLUGIN",
	MagicCookieValue: "zeroui-config-plugin",
}

// PluginMap maps plugin names to their implementations
var PluginMap = map[string]plugin.Plugin{
	"config": &ConfigPluginGRPC{},
}

// PluginManager manages the lifecycle of RPC plugins
type PluginManager struct {
	mu         sync.RWMutex
	clients    map[string]*plugin.Client
	plugins    map[string]ConfigPlugin
	pluginDir  string
	logger     hclog.Logger
	lifecycles map[string]*PluginLifecycle
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir string) *PluginManager {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin-manager",
		Output: os.Stderr,
		Level:  hclog.Info,
	})

	return &PluginManager{
		clients:    make(map[string]*plugin.Client),
		plugins:    make(map[string]ConfigPlugin),
		pluginDir:  pluginDir,
		logger:     logger,
		lifecycles: make(map[string]*PluginLifecycle),
	}
}

// LoadPlugin loads an RPC plugin by name
func (pm *PluginManager) LoadPlugin(name string) (ConfigPlugin, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already loaded
	if p, exists := pm.plugins[name]; exists {
		return p, nil
	}

	// Find plugin binary
	pluginPath := filepath.Join(pm.pluginDir, fmt.Sprintf("zeroui-plugin-%s", name))
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// Try with .exe extension on Windows
		pluginPath += ".exe"
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("plugin binary not found: %s", name)
		}
	}

	// Create plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          PluginMap,
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           pm.logger,
		Stderr:           os.Stderr,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to connect to plugin %s: %w", name, err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("config")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin %s: %w", name, err)
	}

	// Cast to our interface
	configPlugin, ok := raw.(ConfigPlugin)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin %s does not implement ConfigPlugin interface", name)
	}

	// Store references
	pm.clients[name] = client
	pm.plugins[name] = configPlugin

	pm.logger.Info("Successfully loaded plugin", "name", name, "path", pluginPath)
	return configPlugin, nil
}

// GetPlugin returns a loaded plugin by name
func (pm *PluginManager) GetPlugin(name string) (ConfigPlugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// UnloadPlugin unloads a plugin and cleans up resources
func (pm *PluginManager) UnloadPlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	client, exists := pm.clients[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	// Kill the plugin process
	client.Kill()

	// Remove from maps
	delete(pm.clients, name)
	delete(pm.plugins, name)

	pm.logger.Info("Unloaded plugin", "name", name)
	return nil
}

// ListLoadedPlugins returns a list of currently loaded plugin names
func (pm *PluginManager) ListLoadedPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// DiscoverPlugins scans the plugin directory for available plugins
func (pm *PluginManager) DiscoverPlugins() ([]string, error) {
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin directory does not exist: %s", pm.pluginDir)
	}

	entries, err := os.ReadDir(pm.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	var plugins []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".exe" {
			name = name[:len(name)-4] // Remove .exe extension
		}

		// Check if it's a zeroui plugin
		if len(name) > 14 && name[:14] == "zeroui-plugin-" {
			pluginName := name[14:] // Remove "zeroui-plugin-" prefix
			plugins = append(plugins, pluginName)
		}
	}

	return plugins, nil
}

// HealthCheck checks if a plugin is healthy and responsive
func (pm *PluginManager) HealthCheck(name string) error {
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	// Create a context with timeout for health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get plugin info as a health check
	_, err := plugin.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("plugin %s failed health check: %w", name, err)
	}

	return nil
}

// Shutdown gracefully shuts down all plugins
func (pm *PluginManager) Shutdown() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.logger.Info("Shutting down plugin manager", "loaded_plugins", len(pm.clients))

	var errors []error
	for name, client := range pm.clients {
		pm.logger.Info("Shutting down plugin", "name", name)
		client.Kill()
		delete(pm.clients, name)
		delete(pm.plugins, name)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	pm.logger.Info("Plugin manager shutdown complete")
	return nil
}

// GetPluginInfo returns info for a specific plugin
func (pm *PluginManager) GetPluginInfo(name string) (*PluginInfo, error) {
	pm.mu.RLock()
	plugin, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s is not loaded", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return plugin.GetInfo(ctx)
}

// ReloadPlugin unloads and reloads a plugin
func (pm *PluginManager) ReloadPlugin(name string) error {
	pm.logger.Info("Reloading plugin", "name", name)

	// Unload if currently loaded
	if _, exists := pm.GetPlugin(name); exists {
		if err := pm.UnloadPlugin(name); err != nil {
			return fmt.Errorf("failed to unload plugin for reload: %w", err)
		}
	}

	// Load again
	_, err := pm.LoadPlugin(name)
	if err != nil {
		return fmt.Errorf("failed to reload plugin: %w", err)
	}

	pm.logger.Info("Successfully reloaded plugin", "name", name)
	return nil
}

// SetLogLevel changes the log level for the plugin manager
func (pm *PluginManager) SetLogLevel(level string) {
	switch level {
	case "trace":
		pm.logger.SetLevel(hclog.Trace)
	case "debug":
		pm.logger.SetLevel(hclog.Debug)
	case "info":
		pm.logger.SetLevel(hclog.Info)
	case "warn":
		pm.logger.SetLevel(hclog.Warn)
	case "error":
		pm.logger.SetLevel(hclog.Error)
	default:
		pm.logger.Warn("Unknown log level", "level", level)
	}
}

// GetStats returns plugin manager statistics
func (pm *PluginManager) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"loaded_plugins":   len(pm.plugins),
		"active_clients":   len(pm.clients),
		"plugin_directory": pm.pluginDir,
		"loaded_names":     pm.ListLoadedPlugins(),
	}
}

// ListPlugins returns names of all loaded plugins
func (pm *PluginManager) ListPlugins() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// RestartPlugin restarts a plugin
func (pm *PluginManager) RestartPlugin(name string) error {
	pm.mu.Lock()
	lifecycle, exists := pm.lifecycles[name]
	pm.mu.Unlock()

	if !exists {
		// Create new lifecycle if it doesn't exist
		lifecycle = NewPluginLifecycle(name)
		pm.mu.Lock()
		pm.lifecycles[name] = lifecycle
		pm.mu.Unlock()
	}

	// Restart using lifecycle manager
	return lifecycle.Restart(func() (*plugin.Client, ConfigPlugin, error) {
		// Unload first if loaded
		pm.UnloadPlugin(name)

		// Load fresh
		pluginInstance, err := pm.LoadPlugin(name)
		if err != nil {
			return nil, nil, err
		}

		pm.mu.RLock()
		client := pm.clients[name]
		pm.mu.RUnlock()

		return client, pluginInstance, nil
	})
}

// LoadPluginsConcurrently loads multiple plugins in parallel for faster startup
func (pm *PluginManager) LoadPluginsConcurrently(names []string) error {
	if len(names) == 0 {
		return nil
	}

	// Use a semaphore to limit concurrent plugin loads to CPU count
	semaphore := make(chan struct{}, runtime.NumCPU())
	errChan := make(chan error, len(names))
	var wg sync.WaitGroup

	pm.logger.Info("Loading plugins concurrently", "count", len(names), "max_concurrent", runtime.NumCPU())

	for _, name := range names {
		wg.Add(1)
		go func(pluginName string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			pm.logger.Debug("Loading plugin", "name", pluginName)

			if _, err := pm.LoadPlugin(pluginName); err != nil {
				pm.logger.Error("Failed to load plugin", "name", pluginName, "error", err)
				errChan <- fmt.Errorf("plugin %s: %w", pluginName, err)
			} else {
				pm.logger.Info("Successfully loaded plugin", "name", pluginName)
			}
		}(name)
	}

	// Wait for all plugins to load
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to load %d plugins: %v", len(errors), errors)
	}

	pm.logger.Info("All plugins loaded successfully", "count", len(names))
	return nil
}

// DiscoverAndLoadPlugins discovers and loads all available plugins in parallel
func (pm *PluginManager) DiscoverAndLoadPlugins() error {
	// Find all plugin binaries
	pattern := filepath.Join(pm.pluginDir, "zeroui-plugin-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	// Extract plugin names
	var pluginNames []string
	for _, match := range matches {
		base := filepath.Base(match)
		// Remove prefix "zeroui-plugin-" and any extension
		name := strings.TrimPrefix(base, "zeroui-plugin-")
		name = strings.TrimSuffix(name, filepath.Ext(name))
		pluginNames = append(pluginNames, name)
	}

	if len(pluginNames) == 0 {
		pm.logger.Info("No plugins found", "dir", pm.pluginDir)
		return nil
	}

	// Load all plugins concurrently
	return pm.LoadPluginsConcurrently(pluginNames)
}
