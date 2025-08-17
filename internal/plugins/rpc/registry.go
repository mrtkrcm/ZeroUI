package rpc

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages RPC plugins only
type Registry struct {
	mu      sync.RWMutex
	manager *PluginManager
	plugins map[string]*PluginInfo
}

// NewRegistry creates a new RPC-only plugin registry
func NewRegistry(pluginDir string) *Registry {
	return &Registry{
		manager: NewPluginManager(pluginDir),
		plugins: make(map[string]*PluginInfo),
	}
}

// LoadPlugin loads an RPC plugin
func (r *Registry) LoadPlugin(name string) (ConfigPlugin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, err := r.manager.LoadPlugin(name)
	if err != nil {
		return nil, err
	}

	// Cache plugin info
	info, err := plugin.GetInfo(context.Background())
	if err != nil {
		r.manager.UnloadPlugin(name)
		return nil, fmt.Errorf("failed to get plugin info: %w", err)
	}
	r.plugins[name] = info

	return plugin, nil
}

// GetPlugin returns a loaded plugin
func (r *Registry) GetPlugin(name string) (ConfigPlugin, error) {
	plugin, exists := r.manager.GetPlugin(name)
	if !exists {
		return r.LoadPlugin(name)
	}
	return plugin, nil
}

// ListPlugins returns all loaded plugins
func (r *Registry) ListPlugins() []*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*PluginInfo, 0, len(r.plugins))
	for _, info := range r.plugins {
		plugins = append(plugins, info)
	}
	return plugins
}

// DiscoverPlugins finds available plugins
func (r *Registry) DiscoverPlugins() ([]string, error) {
	return r.manager.DiscoverPlugins()
}

// UnloadPlugin unloads a plugin
func (r *Registry) UnloadPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.plugins, name)
	return r.manager.UnloadPlugin(name)
}

// HealthCheck checks plugin health
func (r *Registry) HealthCheck(name string) error {
	return r.manager.HealthCheck(name)
}

// Shutdown gracefully shuts down all plugins
func (r *Registry) Shutdown() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins = make(map[string]*PluginInfo)
	return r.manager.Shutdown()
}

// GetStats returns registry statistics
func (r *Registry) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"loaded_plugins": len(r.plugins),
		"manager":        r.manager.GetStats(),
	}
}