package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/mrtkrcm/ZeroUI/internal/security"
)

// Loader handles loading and parsing configuration files with caching
type Loader struct {
	configDir     string
	yamlValidator *security.YAMLValidator

	// Performance optimization: LRU cache for app configs
	appConfigCache *lru.Cache[string, *AppConfig]
	cacheMutex     sync.RWMutex

	// File watching for cache invalidation
	watcher         *fsnotify.Watcher
	watcherInitOnce sync.Once

	// Cache statistics for monitoring
	cacheHits   uint64
	cacheMisses uint64

	// Memory pools for reusable buffers
	bufferPool        sync.Pool
	stringBuilderPool sync.Pool
}

// NewLoader creates a new config loader with caching
func NewLoader() (*Loader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "configtoggle")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize YAML validator with security limits
	yamlValidator := security.NewYAMLValidator(security.DefaultYAMLLimits())

	// Initialize LRU cache with 1000 entry limit (same as toggle engine)
	appCache, err := lru.New[string, *AppConfig](1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create app config cache: %w", err)
	}

	return &Loader{
		configDir:      configDir,
		yamlValidator:  yamlValidator,
		appConfigCache: appCache,
		// Initialize memory pools for better allocation patterns
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 4096) // 4KB buffer
			},
		},
		stringBuilderPool: sync.Pool{
			New: func() interface{} {
				var sb strings.Builder
				sb.Grow(1024) // Pre-allocate 1KB
				return &sb
			},
		},
	}, nil
}

// SetConfigDir sets the config directory (for testing)
func (l *Loader) SetConfigDir(dir string) {
	l.configDir = dir
	// Initialize YAML validator if not already set (for testing)
	if l.yamlValidator == nil {
		l.yamlValidator = security.NewYAMLValidator(security.DefaultYAMLLimits())
	}
	// Initialize cache if not already set (for testing)
	if l.appConfigCache == nil {
		cache, err := lru.New[string, *AppConfig](1000)
		if err == nil {
			l.appConfigCache = cache
		}
	}
}

// AppConfig represents the configuration for a single application
type AppConfig struct {
	Name        string                  `yaml:"name"`
	Path        string                  `yaml:"path"`
	Format      string                  `yaml:"format"`
	Description string                  `yaml:"description,omitempty"`
	Fields      map[string]FieldConfig  `yaml:"fields"`
	Presets     map[string]PresetConfig `yaml:"presets"`
	Hooks       map[string]string       `yaml:"hooks,omitempty"`
	Env         map[string]string       `yaml:"env,omitempty"`
}

// FieldConfig represents a configurable field
type FieldConfig struct {
	Type        string      `yaml:"type"` // choice, string, number, boolean
	Values      []string    `yaml:"values,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Path        string      `yaml:"path,omitempty"` // JSON path for nested values
}

// PresetConfig represents a preset configuration
type PresetConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Values      map[string]interface{} `yaml:"values"`
}

// LoadAppConfig loads configuration for a specific application with caching
func (l *Loader) LoadAppConfig(appName string) (*AppConfig, error) {
	// Check cache first (fast path)
	l.cacheMutex.RLock()
	if cached, ok := l.appConfigCache.Get(appName); ok {
		l.cacheHits++
		l.cacheMutex.RUnlock()
		return cached, nil
	}
	l.cacheMisses++
	l.cacheMutex.RUnlock()

	// Cache miss - load from disk
	configPath := filepath.Join(l.configDir, "apps", appName+".yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("app config not found: %s", appName)
	}

	// Initialize file watcher on first use
	l.initFileWatcher()

	// Use secure YAML reading with complexity limits (Security hardening)
	data, err := l.yamlValidator.SafeReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to securely read config file: %w", err)
	}

	// Parse YAML with validated content
	var config AppConfig
	if err := yamlv3.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.Name = appName

	// Update cache
	l.cacheMutex.Lock()
	l.appConfigCache.Add(appName, &config)
	l.cacheMutex.Unlock()

	// Add file to watcher for automatic cache invalidation
	if l.watcher != nil {
		_ = l.watcher.Add(configPath)
	}

	return &config, nil
}

// ListApps returns a list of available applications
func (l *Loader) ListApps() ([]string, error) {
	appsDir := filepath.Join(l.configDir, "apps")

	entries, err := os.ReadDir(appsDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read apps directory: %w", err)
	}

	var apps []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			appName := strings.TrimSuffix(entry.Name(), ".yaml")
			apps = append(apps, appName)
		}
	}

	return apps, nil
}

// LoadTargetConfig loads the actual configuration file that the app uses
func (l *Loader) LoadTargetConfig(appConfig *AppConfig) (*koanf.Koanf, error) {
	// Expand home directory in path
	configPath := appConfig.Path
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[1:])
	}

	k := koanf.New(".")

	// Load the file based on format
	var parser koanf.Parser
	format := strings.ToLower(appConfig.Format)
	switch format {
	case "":
		// Auto-detect from file extension only when format is not provided
		ext := strings.ToLower(filepath.Ext(configPath))
		switch ext {
		case ".json":
			parser = json.Parser()
		case ".yaml", ".yml":
			parser = yaml.Parser()
		case ".toml":
			parser = toml.Parser()
		default:
			return nil, fmt.Errorf("unsupported config format: %s", appConfig.Format)
		}
	case "json":
		parser = json.Parser()
	case "yaml", "yml":
		parser = yaml.Parser()
	case "toml":
		parser = toml.Parser()
	case "custom":
		// For custom formats, special handling
		return l.loadCustomFormat(configPath)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", appConfig.Format)
	}

	if err := k.Load(file.Provider(configPath), parser); err != nil {
		return nil, fmt.Errorf("failed to load target config: %w", err)
	}

	return k, nil
}

// SaveTargetConfig saves the configuration back to the target file
func (l *Loader) SaveTargetConfig(appConfig *AppConfig, k *koanf.Koanf) error {
	configPath := appConfig.Path
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[1:])
	}

	// Create backup
	backupPath := configPath + ".backup"
	if _, err := os.Stat(configPath); err == nil {
		if err := copyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	var data []byte
	var err error

	switch strings.ToLower(appConfig.Format) {
	case "json":
		data, err = k.Marshal(json.Parser())
	case "yaml", "yml":
		data, err = k.Marshal(yaml.Parser())
	case "toml":
		data, err = k.Marshal(toml.Parser())
	case "custom":
		return l.saveCustomFormat(configPath, k)
	default:
		ext := strings.ToLower(filepath.Ext(configPath))
		switch ext {
		case ".json":
			data, err = k.Marshal(json.Parser())
		case ".yaml", ".yml":
			data, err = k.Marshal(yaml.Parser())
		case ".toml":
			data, err = k.Marshal(toml.Parser())
		default:
			return fmt.Errorf("unsupported config format: %s", appConfig.Format)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadCustomFormat handles custom configuration formats (like Ghostty's format)
func (l *Loader) loadCustomFormat(configPath string) (*koanf.Koanf, error) {
	// Use the Ghostty custom parser
	return ParseGhosttyConfig(configPath)
}

// saveCustomFormat handles saving custom configuration formats
func (l *Loader) saveCustomFormat(configPath string, k *koanf.Koanf) error {
	// Use the Ghostty custom parser to write back
	return WriteGhosttyConfig(configPath, k, configPath)
}

// initFileWatcher initializes the file watcher for cache invalidation
func (l *Loader) initFileWatcher() {
	l.watcherInitOnce.Do(func() {
		var err error
		l.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			// Log error but don't fail - caching will still work without file watching
			return
		}

		// Start watching for file changes
		go l.watchFiles()
	})
}

// watchFiles handles file system events for cache invalidation
func (l *Loader) watchFiles() {
	if l.watcher == nil {
		return
	}

	for {
		select {
		case event, ok := <-l.watcher.Events:
			if !ok {
				return
			}

			// Invalidate cache on file changes
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {
				l.invalidateCacheForPath(event.Name)
			}

		case err, ok := <-l.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err
		}
	}
}

// invalidateCacheForPath removes cached config for a specific file path
func (l *Loader) invalidateCacheForPath(filePath string) {
	// Extract app name from file path
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".yaml") {
		appName := strings.TrimSuffix(fileName, ".yaml")

		l.cacheMutex.Lock()
		l.appConfigCache.Remove(appName)
		l.cacheMutex.Unlock()
	}
}

// ClearCache clears all cached configurations
func (l *Loader) ClearCache() {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()
	l.appConfigCache.Purge()
}

// GetCacheStats returns cache statistics for monitoring
func (l *Loader) GetCacheStats() map[string]interface{} {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_hits":   l.cacheHits,
		"cache_misses": l.cacheMisses,
		"cache_size":   l.appConfigCache.Len(),
		"hit_ratio":    float64(l.cacheHits) / float64(l.cacheHits+l.cacheMisses),
	}
}

// Close closes the file watcher and cleans up resources
func (l *Loader) Close() error {
	if l.watcher != nil {
		return l.watcher.Close()
	}
	return nil
}

// copyFile creates a copy of a file using streaming I/O for efficiency
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	// Get source file info for future optimizations (permissions, size hints)
	_, err = srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	// Use io.Copy for optimal streaming - handles buffer management automatically
	_, err = io.Copy(dstFile, srcFile)
	return err
}
