package runtimeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config represents the consolidated runtime configuration for the CLI/TUI.
type Config struct {
	ConfigFile   string `mapstructure:"config"`
	ConfigDir    string `mapstructure:"config_dir"`
	LogLevel     string `mapstructure:"log_level"`
	LogFormat    string `mapstructure:"log_format"`
	DefaultTheme string `mapstructure:"default_theme"`
	Verbose      bool   `mapstructure:"verbose"`
	DryRun       bool   `mapstructure:"dry_run"`
}

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "console"
	defaultTheme     = "Modern"
	envPrefix        = "ZEROUi"
)

// Loader merges configuration sources (flags, env vars, config files) into a typed struct.
type Loader struct {
	v *viper.Viper
}

// NewLoader constructs a Loader using the provided viper instance. If v is nil, the global
// viper instance is used so downstream code relying on viper keeps working.
func NewLoader(v *viper.Viper) *Loader {
	if v == nil {
		v = viper.GetViper()
	}
	return &Loader{v: v}
}

// DefaultConfigDir returns the default directory for ZeroUI configuration.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zeroui"
	}
	return filepath.Join(home, ".config", "zeroui")
}

// DefaultConfig returns a Config populated with built-in defaults.
func DefaultConfig() *Config {
	return &Config{
		ConfigDir:    DefaultConfigDir(),
		LogLevel:     defaultLogLevel,
		LogFormat:    defaultLogFormat,
		DefaultTheme: defaultTheme,
	}
}

// Load processes configuration from the provided sources and validates the result.
func (l *Loader) Load(configFile string, flags *pflag.FlagSet) (*Config, error) {
	cfg := DefaultConfig()

	l.v.SetEnvPrefix(envPrefix)
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	l.v.AutomaticEnv()

	l.setDefaults(cfg)

	if flags != nil {
		bindFlag(l.v, flags, "config", "config")
		bindFlag(l.v, flags, "config_dir", "config-dir")
		bindFlag(l.v, flags, "log_level", "log-level")
		bindFlag(l.v, flags, "log_format", "log-format")
		bindFlag(l.v, flags, "default_theme", "default-theme")
		bindFlag(l.v, flags, "verbose", "verbose")
		bindFlag(l.v, flags, "dry_run", "dry-run")
	}

	if configFile != "" {
		l.v.SetConfigFile(configFile)
	} else {
		l.v.SetConfigName("config")
		l.v.SetConfigType("yaml")
		l.v.AddConfigPath(DefaultConfigDir())
	}

	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok || configFile != "" {
			return nil, fmt.Errorf("failed to read config file %q: %w", configFile, err)
		}
	} else {
		cfg.ConfigFile = l.v.ConfigFileUsed()
	}

	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	cfg.ConfigDir = expandPath(cfg.ConfigDir)
	if cfg.ConfigFile == "" && configFile != "" {
		cfg.ConfigFile = configFile
	}
	if cfg.ConfigFile != "" {
		cfg.ConfigFile = expandPath(cfg.ConfigFile)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	// Keep existing viper-based callers working by setting the resolved values.
	l.applyGlobals(cfg)

	return cfg, nil
}

func (l *Loader) setDefaults(cfg *Config) {
	l.v.SetDefault("config_dir", cfg.ConfigDir)
	l.v.SetDefault("log_level", cfg.LogLevel)
	l.v.SetDefault("log_format", cfg.LogFormat)
	l.v.SetDefault("default_theme", cfg.DefaultTheme)
}

func bindFlag(v *viper.Viper, flags *pflag.FlagSet, key, name string) {
	if flag := flags.Lookup(name); flag != nil {
		_ = v.BindPFlag(key, flag)
	}
}

func (l *Loader) applyGlobals(cfg *Config) {
	l.v.Set("verbose", cfg.Verbose)
	l.v.Set("dry-run", cfg.DryRun)
	l.v.Set("log_level", cfg.LogLevel)
	l.v.Set("log_format", cfg.LogFormat)
	l.v.Set("default_theme", cfg.DefaultTheme)
	l.v.Set("config_dir", cfg.ConfigDir)
	if cfg.ConfigFile != "" {
		l.v.Set("config", cfg.ConfigFile)
	}
}

func expandPath(path string) string {
	if path == "" {
		return ""
	}

	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

func validateConfig(cfg *Config) error {
	var issues []string

	allowedLogLevels := map[string]struct{}{ // zerolog-compatible
		"trace": {},
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
		"fatal": {},
		"panic": {},
	}

	allowedFormats := map[string]struct{}{
		"console": {},
		"json":    {},
	}

	allowedThemes := make(map[string]struct{})
	for _, name := range styles.GetThemeNames() {
		allowedThemes[strings.ToLower(name)] = struct{}{}
	}

	if cfg.ConfigFile != "" {
		if _, err := os.Stat(cfg.ConfigFile); err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("config file %q does not exist; create it or point --config to an existing file", cfg.ConfigFile))
			} else {
				issues = append(issues, fmt.Sprintf("cannot read config file %q: %v", cfg.ConfigFile, err))
			}
		}
	}

	if cfg.ConfigDir == "" {
		issues = append(issues, "config_dir cannot be empty; set it to the directory containing application configs")
	} else if info, err := os.Stat(cfg.ConfigDir); err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("config_dir %q does not exist; create the directory or update the setting", cfg.ConfigDir))
		} else {
			issues = append(issues, fmt.Sprintf("cannot access config_dir %q: %v", cfg.ConfigDir, err))
		}
	} else if !info.IsDir() {
		issues = append(issues, fmt.Sprintf("config_dir %q is not a directory; point it to a valid folder", cfg.ConfigDir))
	}

	if _, ok := allowedLogLevels[strings.ToLower(cfg.LogLevel)]; !ok {
		issues = append(issues, fmt.Sprintf("invalid log_level %q; expected one of trace, debug, info, warn, error, fatal, panic", cfg.LogLevel))
	}

	if _, ok := allowedFormats[strings.ToLower(cfg.LogFormat)]; !ok {
		issues = append(issues, fmt.Sprintf("invalid log_format %q; use either 'console' or 'json'", cfg.LogFormat))
	}

	if _, ok := allowedThemes[strings.ToLower(cfg.DefaultTheme)]; !ok {
		issues = append(issues, fmt.Sprintf("unknown default_theme %q; available themes: %s", cfg.DefaultTheme, strings.Join(styles.GetThemeNames(), ", ")))
	}

	if len(issues) > 0 {
		return fmt.Errorf("configuration validation failed:\n - %s", strings.Join(issues, "\n - "))
	}

	return nil
}
