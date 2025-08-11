package configextractor

import (
	"time"
	"github.com/mrtkrcm/ZeroUI/pkg/configextractor/cache"
)

// CacheAdapter adapts the cache.LRUCache to the Cache interface
type CacheAdapter struct {
	lru *cache.LRUCache
}

// NewLRUCache creates a new cache adapter
func NewLRUCache(maxSize int, defaultTTL time.Duration) Cache {
	return &CacheAdapter{
		lru: cache.NewLRU(maxSize, defaultTTL),
	}
}

// Get retrieves a config from cache
func (a *CacheAdapter) Get(key string) (*Config, bool) {
	cachedConfig, found := a.lru.Get(key)
	if !found {
		return nil, false
	}
	
	// Convert from cache.Config to configextractor.Config
	config := &Config{
		App:        cachedConfig.App,
		ConfigPath: cachedConfig.ConfigPath,
		Format:     cachedConfig.Format,
		Timestamp:  cachedConfig.Timestamp,
		Source:     ExtractionSource{Method: cachedConfig.Source},
		Settings:   make(map[string]Setting),
	}
	
	// Convert settings if needed
	if cachedConfig.Settings != nil {
		for k, v := range cachedConfig.Settings {
			if settingMap, ok := v.(map[string]interface{}); ok {
				setting := Setting{
					Name: k,
				}
				if desc, ok := settingMap["description"].(string); ok {
					setting.Desc = desc
				}
				if def, ok := settingMap["default"]; ok {
					setting.Default = def
				}
				config.Settings[k] = setting
			}
		}
	}
	
	return config, true
}

// Set stores a config in cache
func (a *CacheAdapter) Set(key string, config *Config) {
	// Convert from configextractor.Config to cache.Config
	cachedConfig := &cache.Config{
		App:        config.App,
		ConfigPath: config.ConfigPath,
		Format:     config.Format,
		Source:     config.Source.Method,
		Timestamp:  config.Timestamp,
		Settings:   make(map[string]interface{}),
	}
	
	// Convert settings
	for k, v := range config.Settings {
		settingMap := map[string]interface{}{
			"name":        v.Name,
			"type":        string(v.Type),
			"description": v.Desc,
			"default":     v.Default,
		}
		cachedConfig.Settings[k] = settingMap
	}
	
	a.lru.Set(key, cachedConfig, 24*time.Hour)
}

// Delete removes a config from cache
func (a *CacheAdapter) Delete(key string) {
	// LRUCache doesn't have a Delete method, we'll use Clear for now
	// This could be improved by adding a Delete method to LRUCache
	a.lru.Clear()
}

// Clear removes all entries
func (a *CacheAdapter) Clear() {
	a.lru.Clear()
}