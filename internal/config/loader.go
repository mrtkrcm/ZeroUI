package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/mrtkrcm/ZeroUI/internal/performance"
	"github.com/mrtkrcm/ZeroUI/internal/security"
)

// Loader handles loading and parsing configuration files with caching.
type Loader struct {
	configDir     string
	yamlValidator *security.YAMLValidator
	pathValidator *security.PathValidator

	// Performance optimization: LRU cache for app configs
	appConfigCache *lru.Cache[string, *AppConfig]
	cacheMutex     sync.RWMutex

	// File watching for cache invalidation with debouncing
	fileWatcher     *DebouncedWatcher
	watcherInitOnce sync.Once

	// Cache statistics for monitoring
	cacheHits   uint64
	cacheMisses uint64

	// Memory pools for reusable buffers
	bufferPool        sync.Pool
	stringBuilderPool sync.Pool
}

// NewLoader creates a new config loader with caching.
//
// Behavior:
//   - If the environment variable ZEROUI_CONFIG_DIR is set, it is used as the
//     config directory (useful for tests/CI).
//   - Otherwise the loader uses $HOME/.config/zeroui as the default path.
//   - The directory is created if it does not exist.
func NewLoader() (*Loader, error) {
	// Honor explicit override (useful for tests/CI)
	if cfg := os.Getenv("ZEROUI_CONFIG_DIR"); cfg != "" {
		configDir := cfg
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		yamlValidator := security.NewYAMLValidator(security.DefaultYAMLLimits())
		pathValidator := security.NewPathValidator(configDir)
		appCache, err := lru.New[string, *AppConfig](1000)
		if err != nil {
			return nil, fmt.Errorf("failed to create app config cache: %w", err)
		}
		return &Loader{
			configDir:      configDir,
			yamlValidator:  yamlValidator,
			pathValidator:  pathValidator,
			appConfigCache: appCache,
			bufferPool: sync.Pool{
				New: func() interface{} { return make([]byte, 0, 4096) },
			},
			stringBuilderPool: sync.Pool{
				New: func() interface{} { var sb strings.Builder; sb.Grow(1024); return &sb },
			},
		}, nil
	}

	home, err := performance.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "zeroui")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize YAML validator with security limits
	yamlValidator := security.NewYAMLValidator(security.DefaultYAMLLimits())
	
	// Initialize path validator with allowed config directories
	pathValidator := security.NewPathValidator(configDir, home)

	// Initialize LRU cache with 1000 entry limit
	appCache, err := lru.New[string, *AppConfig](1000)
	if err != nil {
		return nil, fmt.Errorf("failed to create app config cache: %w", err)
	}

	return &Loader{
		configDir:      configDir,
		yamlValidator:  yamlValidator,
		pathValidator:  pathValidator,
		appConfigCache: appCache,
		bufferPool: sync.Pool{
			New: func() interface{} { return make([]byte, 0, 4096) },
		},
		stringBuilderPool: sync.Pool{
			New: func() interface{} { var sb strings.Builder; sb.Grow(1024); return &sb },
		},
	}, nil
}

// SetConfigDir sets the config directory (for testing or overrides).
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

// AppConfig represents the configuration for a single application.
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

// FieldConfig represents a configurable field.
type FieldConfig struct {
	Type        string      `yaml:"type"` // choice, string, number, boolean
	Values      []string    `yaml:"values,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Path        string      `yaml:"path,omitempty"` // JSON path for nested values
}

// PresetConfig represents a preset configuration.
type PresetConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Values      map[string]interface{} `yaml:"values"`
}

// LoadAppConfig loads configuration for a specific application with caching.
func (l *Loader) LoadAppConfig(appName string) (*AppConfig, error) {
	// Check cache first
	l.cacheMutex.RLock()
	if cached, ok := l.appConfigCache.Get(appName); ok {
		l.cacheHits++
		l.cacheMutex.RUnlock()
		return cached, nil
	}
	l.cacheMisses++
	l.cacheMutex.RUnlock()

	// Cache miss - load from disk
	// Validate app name for security
	if err := l.pathValidator.ValidatePath(appName); err != nil {
		return nil, fmt.Errorf("invalid app name: %w", err)
	}
	
	configPath := filepath.Join(l.configDir, "apps", appName+".yaml")
	
	// Validate the resulting config path
	if err := l.pathValidator.ValidatePath(configPath); err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("app config not found: %s", appName)
	}

	// Initialize file watcher on first use
	l.initFileWatcher()

	// Use secure YAML reading
	data, err := l.yamlValidator.SafeReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to securely read config file: %w", err)
	}

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
	if l.fileWatcher != nil {
		_ = l.fileWatcher.Watch(configPath)
	}

	return &config, nil
}

// ListApps returns a list of available applications.
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

// LoadTargetConfig loads the actual configuration file that the app uses.
func (l *Loader) LoadTargetConfig(appConfig *AppConfig) (*koanf.Koanf, error) {
	// Expand home directory in path
	configPath := appConfig.Path
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		// support both "~" and "~/..." usage
		trimmed := strings.TrimPrefix(configPath, "~")
		if strings.HasPrefix(trimmed, string(os.PathSeparator)) {
			configPath = filepath.Join(home, trimmed[1:])
		} else {
			configPath = filepath.Join(home, trimmed)
		}
	}
	
	// Validate the resolved config path for security
	if err := l.pathValidator.ValidatePath(configPath); err != nil {
		return nil, fmt.Errorf("invalid target config path '%s': %w", configPath, err)
	}

	k := koanf.New(".")

	var parser koanf.Parser
	format := strings.ToLower(appConfig.Format)
	switch format {
	case "":
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
		return l.loadCustomFormat(configPath)
	default:
		return nil, fmt.Errorf("unsupported config format: %s", appConfig.Format)
	}

	if err := k.Load(file.Provider(configPath), parser); err != nil {
		return nil, fmt.Errorf("failed to load target config: %w", err)
	}

	return k, nil
}

// SaveTargetConfig saves the configuration back to the target file using temporary files for safety.
func (l *Loader) SaveTargetConfig(appConfig *AppConfig, k *koanf.Koanf) error {
	configPath := appConfig.Path
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		trimmed := strings.TrimPrefix(configPath, "~")
		if strings.HasPrefix(trimmed, string(os.PathSeparator)) {
			configPath = filepath.Join(home, trimmed[1:])
		} else {
			configPath = filepath.Join(home, trimmed)
		}
	}

	// Initialize temp file manager and integrity checker
	tempManager, err := NewTempFileManager()
	if err != nil {
		return fmt.Errorf("failed to initialize temp manager: %w", err)
	}
	defer tempManager.CleanupAll()

	integrityChecker := NewIntegrityChecker()

	// Create temporary copy of the file
	tempFile, err := tempManager.CreateTempCopy(configPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary copy: %w", err)
	}

	// Marshal configuration data
	var data []byte

	switch strings.ToLower(appConfig.Format) {
	case "json":
		data, err = k.Marshal(json.Parser())
	case "yaml", "yml":
		data, err = k.Marshal(yaml.Parser())
	case "toml":
		data, err = k.Marshal(toml.Parser())
	case "custom":
		// For custom formats, handle separately
		if err := l.saveCustomFormatWithTemp(tempFile.TempPath, k); err != nil {
			tempManager.Rollback(tempFile)
			return fmt.Errorf("failed to save custom format: %w", err)
		}
		// Validate and commit
		if err := integrityChecker.ValidateFormat(tempFile.TempPath); err != nil {
			tempManager.Rollback(tempFile)
			return fmt.Errorf("validation failed: %w", err)
		}
		return tempManager.CommitTemp(tempFile)
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
			tempManager.Rollback(tempFile)
			return fmt.Errorf("unsupported config format: %s", appConfig.Format)
		}
	}

	if err != nil {
		tempManager.Rollback(tempFile)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tempFile.TempPath, data, 0o644); err != nil {
		tempManager.Rollback(tempFile)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Validate the temporary file before committing
	if err := integrityChecker.ValidateFormat(tempFile.TempPath); err != nil {
		tempManager.Rollback(tempFile)
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Verify content integrity
	if err := integrityChecker.ValidateContent(tempFile.TempPath, nil); err != nil {
		tempManager.Rollback(tempFile)
		return fmt.Errorf("content validation failed: %w", err)
	}

	// Commit the temporary file to the actual location
	if err := tempManager.CommitTemp(tempFile); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Verify final file integrity
	finalChecksum, err := integrityChecker.CalculateChecksum(configPath)
	if err != nil {
		return fmt.Errorf("failed to verify saved file: %w", err)
	}

	// Log success with checksum for audit
	if finalChecksum != "" {
		// Could log this for audit purposes
		_ = finalChecksum
	}

	return nil
}

// loadCustomFormat handles custom formats (like Ghostty).
func (l *Loader) loadCustomFormat(configPath string) (*koanf.Koanf, error) {
	return ParseGhosttyConfig(configPath)
}

// saveCustomFormat handles saving custom formats.
func (l *Loader) saveCustomFormat(configPath string, k *koanf.Koanf) error {
	return WriteGhosttyConfig(configPath, k, configPath)
}

// saveCustomFormatWithTemp handles saving custom formats to a temporary file.
func (l *Loader) saveCustomFormatWithTemp(tempPath string, k *koanf.Koanf) error {
	// For now, write directly to temp path
	// This would be enhanced for specific custom formats
	data := make(map[string]interface{})
	for _, key := range k.Keys() {
		data[key] = k.Get(key)
	}

	// Write as simple key=value format for custom configs
	var lines []string
	for key, value := range data {
		lines = append(lines, fmt.Sprintf("%s = %v", key, value))
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(tempPath, []byte(content), 0o644)
}

// initFileWatcher initializes the file watcher for cache invalidation.
func (l *Loader) initFileWatcher() {
	l.watcherInitOnce.Do(func() {
		var err error
		l.fileWatcher, err = NewDebouncedWatcher(func(path string) {
			l.invalidateCacheForPath(path)
		})
		if err != nil {
			// Log error but do not fail; caching still works without watcher.
			return
		}
	})
}

// invalidateCacheForPath removes cached config for a specific file path.
func (l *Loader) invalidateCacheForPath(filePath string) {
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".yaml") {
		appName := strings.TrimSuffix(fileName, ".yaml")
		l.cacheMutex.Lock()
		l.appConfigCache.Remove(appName)
		l.cacheMutex.Unlock()
	}
}

// ClearCache clears all cached configurations.
func (l *Loader) ClearCache() {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()
	l.appConfigCache.Purge()
}

// GetCacheStats returns cache statistics for monitoring.
func (l *Loader) GetCacheStats() map[string]interface{} {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	hits := l.cacheHits
	misses := l.cacheMisses
	total := hits + misses
	var hitRatio float64
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}
	return map[string]interface{}{
		"cache_hits":   hits,
		"cache_misses": misses,
		"cache_size":   l.appConfigCache.Len(),
		"hit_ratio":    hitRatio,
	}
}

// Close closes the file watcher and cleans up resources.
func (l *Loader) Close() error {
	if l.fileWatcher != nil {
		return l.fileWatcher.Close()
	}
	return nil
}

// copyFile creates a copy of a file using streaming I/O for efficiency.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	_, err = srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
