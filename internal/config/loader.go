package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	yamlv3 "gopkg.in/yaml.v3"
)

// Loader handles loading and parsing configuration files
type Loader struct {
	configDir string
}

// NewLoader creates a new config loader
func NewLoader() (*Loader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "configtoggle")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Loader{
		configDir: configDir,
	}, nil
}

// SetConfigDir sets the config directory (for testing)
func (l *Loader) SetConfigDir(dir string) {
	l.configDir = dir
}

// AppConfig represents the configuration for a single application
type AppConfig struct {
	Name        string                     `yaml:"name"`
	Path        string                     `yaml:"path"`
	Format      string                     `yaml:"format"`
	Description string                     `yaml:"description,omitempty"`
	Fields      map[string]FieldConfig     `yaml:"fields"`
	Presets     map[string]PresetConfig    `yaml:"presets"`
	Hooks       map[string]string          `yaml:"hooks,omitempty"`
	Env         map[string]string          `yaml:"env,omitempty"`
}

// FieldConfig represents a configurable field
type FieldConfig struct {
	Type        string      `yaml:"type"`        // choice, string, number, boolean
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

// LoadAppConfig loads configuration for a specific application
func (l *Loader) LoadAppConfig(appName string) (*AppConfig, error) {
	configPath := filepath.Join(l.configDir, "apps", appName+".yaml")
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("app config not found: %s", appName)
	}

	// TODO: Implement lazy loading with file watchers (Week 3)
	// TODO: Add in-memory caching to avoid repeated file reads
	// TODO: Performance: File I/O on every operation causes 100ms+ latency
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// TODO: Add YAML complexity limits to prevent resource exhaustion (Security)
	// TODO: Implement max file size, depth, and key count limits
	var config AppConfig
	if err := yamlv3.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.Name = appName
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
	switch strings.ToLower(appConfig.Format) {
	case "json":
		parser = json.Parser()
	case "yaml", "yml":
		parser = yaml.Parser()
	case "toml":
		parser = toml.Parser()
	case "custom":
		// For custom formats, we'll need special handling
		return l.loadCustomFormat(configPath)
	default:
		// Try to detect format from file extension
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

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}