package runtimeconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config represents the runtime configuration for ZeroUI.
// It supports loading from multiple sources with the following precedence:
// flags > environment variables > config file > defaults
type Config struct {
	ConfigFile   string `mapstructure:"config" validate:"omitempty,filepath"`
	ConfigDir    string `mapstructure:"config_dir" validate:"required,dirpath"`
	LogLevel     string `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
	LogFormat    string `mapstructure:"log_format" validate:"required,oneof=text json"`
	DefaultTheme string `mapstructure:"default_theme" validate:"required,oneof=default modern dracula light nord catppuccin"`
	Verbose      bool   `mapstructure:"verbose"`
	DryRun       bool   `mapstructure:"dry_run"`
}

// Loader manages loading runtime configuration from multiple sources.
type Loader struct {
	v *viper.Viper
}

// NewLoader creates a new runtime configuration loader with the provided viper instance.
// If v is nil, a new viper instance is created.
func NewLoader(v *viper.Viper) *Loader {
	if v == nil {
		v = viper.New()
	}
	return &Loader{v: v}
}

// Load loads the runtime configuration with the following precedence:
// 1. Command-line flags (highest priority)
// 2. Environment variables (ZEROUI_ prefix)
// 3. Configuration file (if provided)
// 4. Default values (lowest priority)
//
// Parameters:
//   - cfgFile: Path to the configuration file (optional, can be empty)
//   - flags: Command-line flags to bind (optional, can be nil)
//
// Returns:
//   - *Config: The loaded configuration
//   - error: Any error encountered during loading
func (l *Loader) Load(cfgFile string, flags *pflag.FlagSet) (*Config, error) {
	// Set defaults
	l.setDefaults()

	// Bind environment variables with ZEROUI_ prefix
	l.v.SetEnvPrefix("ZEROUI")
	l.v.AutomaticEnv()

	// Bind command-line flags if provided
	if flags != nil {
		if err := l.bindFlags(flags); err != nil {
			return nil, fmt.Errorf("failed to bind flags: %w", err)
		}
	}

	// Load config file if provided
	if cfgFile != "" {
		l.v.SetConfigFile(cfgFile)
		if err := l.v.ReadInConfig(); err != nil {
			// Only return error if file was explicitly provided but can't be read
			return nil, fmt.Errorf("failed to read config file %s: %w", cfgFile, err)
		}
	}

	// Unmarshal into Config struct
	cfg := &Config{}
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the configuration
	if err := l.validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values for all configuration options.
func (l *Loader) setDefaults() {
	l.v.SetDefault("config_dir", DefaultConfigDir())
	l.v.SetDefault("log_level", "info")
	l.v.SetDefault("log_format", "text")
	l.v.SetDefault("default_theme", "modern")
	l.v.SetDefault("verbose", false)
	l.v.SetDefault("dry_run", false)
}

// bindFlags binds command-line flags to viper configuration keys.
// It maps flag names (with hyphens) to viper keys (with underscores).
func (l *Loader) bindFlags(flags *pflag.FlagSet) error {
	// Map of flag names to viper keys
	flagMap := map[string]string{
		"config-dir":    "config_dir",
		"log-level":     "log_level",
		"log-format":    "log_format",
		"default-theme": "default_theme",
		"verbose":       "verbose",
		"dry-run":       "dry_run",
		"config":        "config",
	}

	// Bind each flag individually with proper key mapping
	for flagName, viperKey := range flagMap {
		flag := flags.Lookup(flagName)
		if flag != nil {
			if err := l.v.BindPFlag(viperKey, flag); err != nil {
				return fmt.Errorf("failed to bind flag %s: %w", flagName, err)
			}
		}
	}

	return nil
}

// validate performs validation on the loaded configuration.
// It checks for basic requirements and path validity.
func (l *Loader) validate(cfg *Config) error {
	// Validate ConfigDir is set
	if cfg.ConfigDir == "" {
		return fmt.Errorf("config_dir cannot be empty")
	}

	// Validate LogLevel
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[cfg.LogLevel] {
		return fmt.Errorf("invalid log_level: %s (must be one of: debug, info, warn, error)", cfg.LogLevel)
	}

	// Validate LogFormat
	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validLogFormats[cfg.LogFormat] {
		return fmt.Errorf("invalid log_format: %s (must be one of: text, json)", cfg.LogFormat)
	}

	// Validate DefaultTheme
	validThemes := map[string]bool{
		"default":    true,
		"modern":     true,
		"dracula":    true,
		"light":      true,
		"nord":       true,
		"catppuccin": true,
	}
	if !validThemes[cfg.DefaultTheme] {
		return fmt.Errorf("invalid default_theme: %s (must be one of: default, modern, dracula, light, nord, catppuccin)", cfg.DefaultTheme)
	}

	// Validate ConfigFile exists if specified
	if cfg.ConfigFile != "" {
		if _, err := os.Stat(cfg.ConfigFile); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", cfg.ConfigFile)
		}
	}

	return nil
}

// DefaultConfigDir returns the default configuration directory for ZeroUI.
// It checks the following in order:
// 1. ZEROUI_CONFIG_DIR environment variable
// 2. $HOME/.config/zeroui (default)
func DefaultConfigDir() string {
	// Check environment variable first
	if dir := os.Getenv("ZEROUI_CONFIG_DIR"); dir != "" {
		return dir
	}

	// Fall back to default
	home, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, use relative path as last resort
		return ".config/zeroui"
	}

	return filepath.Join(home, ".config", "zeroui")
}
