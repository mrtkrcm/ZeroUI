package runtimeconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	t.Run("WithProvidedViper", func(t *testing.T) {
		v := viper.New()
		loader := NewLoader(v)
		require.NotNil(t, loader)
		assert.Equal(t, v, loader.v)
	})

	t.Run("WithNilViper", func(t *testing.T) {
		loader := NewLoader(nil)
		require.NotNil(t, loader)
		require.NotNil(t, loader.v)
	})
}

func TestDefaultConfigDir(t *testing.T) {
	t.Run("WithEnvironmentVariable", func(t *testing.T) {
		// Save original value
		original := os.Getenv("ZEROUI_CONFIG_DIR")
		defer func() {
			if original != "" {
				os.Setenv("ZEROUI_CONFIG_DIR", original)
			} else {
				os.Unsetenv("ZEROUI_CONFIG_DIR")
			}
		}()

		// Set test value
		expected := "/tmp/test-config"
		os.Setenv("ZEROUI_CONFIG_DIR", expected)

		result := DefaultConfigDir()
		assert.Equal(t, expected, result)
	})

	t.Run("WithoutEnvironmentVariable", func(t *testing.T) {
		// Save original value
		original := os.Getenv("ZEROUI_CONFIG_DIR")
		defer func() {
			if original != "" {
				os.Setenv("ZEROUI_CONFIG_DIR", original)
			} else {
				os.Unsetenv("ZEROUI_CONFIG_DIR")
			}
		}()

		// Unset environment variable
		os.Unsetenv("ZEROUI_CONFIG_DIR")

		result := DefaultConfigDir()
		home, err := os.UserHomeDir()
		require.NoError(t, err)
		expected := filepath.Join(home, ".config", "zeroui")
		assert.Equal(t, expected, result)
	})
}

func TestLoader_Load_Defaults(t *testing.T) {
	// Clean environment
	cleanEnv(t)

	loader := NewLoader(nil)
	cfg, err := loader.Load("", nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check defaults
	assert.Equal(t, DefaultConfigDir(), cfg.ConfigDir)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.Equal(t, "modern", cfg.DefaultTheme)
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
}

func TestLoader_Load_FromEnvironment(t *testing.T) {
	// Clean environment
	cleanEnv(t)

	// Set environment variables
	os.Setenv("ZEROUI_CONFIG_DIR", "/tmp/zeroui-test")
	os.Setenv("ZEROUI_LOG_LEVEL", "debug")
	os.Setenv("ZEROUI_LOG_FORMAT", "json")
	os.Setenv("ZEROUI_DEFAULT_THEME", "catppuccin")
	os.Setenv("ZEROUI_VERBOSE", "true")
	os.Setenv("ZEROUI_DRY_RUN", "true")
	defer cleanEnv(t)

	loader := NewLoader(nil)
	cfg, err := loader.Load("", nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check environment values
	assert.Equal(t, "/tmp/zeroui-test", cfg.ConfigDir)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "catppuccin", cfg.DefaultTheme)
	assert.True(t, cfg.Verbose)
	assert.True(t, cfg.DryRun)
}

func TestLoader_Load_FromFlags(t *testing.T) {
	// Clean environment
	cleanEnv(t)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("config-dir", "", "config directory")
	flags.String("log-level", "", "log level")
	flags.String("log-format", "", "log format")
	flags.String("default-theme", "", "default theme")
	flags.Bool("verbose", false, "verbose output")
	flags.Bool("dry-run", false, "dry run mode")

	// Set flag values
	flags.Set("config-dir", "/tmp/flags-test")
	flags.Set("log-level", "warn")
	flags.Set("log-format", "json")
	flags.Set("default-theme", "nord")
	flags.Set("verbose", "true")
	flags.Set("dry-run", "true")

	loader := NewLoader(nil)
	cfg, err := loader.Load("", flags)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check flag values
	assert.Equal(t, "/tmp/flags-test", cfg.ConfigDir)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "nord", cfg.DefaultTheme)
	assert.True(t, cfg.Verbose)
	assert.True(t, cfg.DryRun)
}

func TestLoader_Load_FromFile(t *testing.T) {
	// Clean environment
	cleanEnv(t)

	// Create temporary config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
config_dir: /tmp/file-test
log_level: error
log_format: json
default_theme: dracula
verbose: true
dry_run: true
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check file values
	assert.Equal(t, "/tmp/file-test", cfg.ConfigDir)
	assert.Equal(t, "error", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "dracula", cfg.DefaultTheme)
	assert.True(t, cfg.Verbose)
	assert.True(t, cfg.DryRun)
}

func TestLoader_Load_Precedence(t *testing.T) {
	// Test that flags override environment which overrides file which overrides defaults
	cleanEnv(t)

	// Create temporary config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
log_level: error
log_format: json
default_theme: dracula
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("ZEROUI_LOG_LEVEL", "warn")
	os.Setenv("ZEROUI_DEFAULT_THEME", "nord")
	defer cleanEnv(t)

	// Set flags (highest priority)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	flags.Set("log-level", "debug")

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, flags)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify precedence:
	// - log_level: flag (debug) > env (warn) > file (error)
	assert.Equal(t, "debug", cfg.LogLevel)
	// - default_theme: env (nord) > file (dracula)
	assert.Equal(t, "nord", cfg.DefaultTheme)
	// - log_format: file (json) > default (text)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLoader_Load_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() *pflag.FlagSet
		expectError string
	}{
		{
			name: "InvalidLogLevel",
			setupFunc: func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("log-level", "", "log level")
				flags.Set("log-level", "invalid")
				return flags
			},
			expectError: "invalid log_level",
		},
		{
			name: "InvalidLogFormat",
			setupFunc: func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("log-format", "", "log format")
				flags.Set("log-format", "invalid")
				return flags
			},
			expectError: "invalid log_format",
		},
		{
			name: "InvalidTheme",
			setupFunc: func() *pflag.FlagSet {
				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.String("default-theme", "", "default theme")
				flags.Set("default-theme", "invalid")
				return flags
			},
			expectError: "invalid default_theme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanEnv(t)
			loader := NewLoader(nil)
			flags := tt.setupFunc()

			cfg, err := loader.Load("", flags)

			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestLoader_Load_NonexistentConfigFile(t *testing.T) {
	cleanEnv(t)

	loader := NewLoader(nil)
	cfg, err := loader.Load("/nonexistent/config.yaml", nil)

	// Should fail validation because file doesn't exist
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoader_Load_InvalidConfigFile(t *testing.T) {
	cleanEnv(t)

	// Create temporary invalid config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
invalid: yaml: content:
  - this is
  - malformed
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoader_Load_ComplexScenario(t *testing.T) {
	// Test a realistic scenario with all sources
	cleanEnv(t)

	// Setup config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
config_dir: /tmp/complex-test
log_level: info
log_format: text
default_theme: default
verbose: false
dry_run: false
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	// Setup environment (overrides some file values)
	os.Setenv("ZEROUI_LOG_LEVEL", "warn")
	os.Setenv("ZEROUI_VERBOSE", "true")
	defer cleanEnv(t)

	// Setup flags (overrides environment and file)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("default-theme", "", "default theme")
	flags.Bool("dry-run", false, "dry run")
	flags.Set("default-theme", "catppuccin")
	flags.Set("dry-run", "true")

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, flags)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify final values with correct precedence
	assert.Equal(t, "/tmp/complex-test", cfg.ConfigDir) // from file
	assert.Equal(t, "warn", cfg.LogLevel)               // from env (overrides file)
	assert.Equal(t, "text", cfg.LogFormat)              // from file
	assert.Equal(t, "catppuccin", cfg.DefaultTheme)     // from flags (overrides file)
	assert.True(t, cfg.Verbose)                         // from env (overrides file)
	assert.True(t, cfg.DryRun)                          // from flags (overrides file)
}

func TestLoader_Load_JSONConfigFile(t *testing.T) {
	cleanEnv(t)

	// Create temporary JSON config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.json")
	cfgContent := `{
  "config_dir": "/tmp/json-test",
  "log_level": "debug",
  "log_format": "json",
  "default_theme": "nord",
  "verbose": true,
  "dry_run": false
}`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check JSON file values
	assert.Equal(t, "/tmp/json-test", cfg.ConfigDir)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "nord", cfg.DefaultTheme)
	assert.True(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
}

func TestLoader_Load_EmptyConfigFile(t *testing.T) {
	cleanEnv(t)

	// Create temporary empty config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(cfgFile, []byte(""), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use defaults since file is empty
	assert.Equal(t, DefaultConfigDir(), cfg.ConfigDir)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.Equal(t, "modern", cfg.DefaultTheme)
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
}

func TestLoader_Load_ConfigFileWithFlags(t *testing.T) {
	cleanEnv(t)

	// Create config file with some values
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
log_level: info
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	// Setup flags that override config file
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("config", "", "config file")
	flags.Set("config", cfgFile)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, flags)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, cfgFile, cfg.ConfigFile)
}

func TestLoader_Load_AllValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			cleanEnv(t)

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String("log-level", "", "log level")
			flags.Set("log-level", level)

			loader := NewLoader(nil)
			cfg, err := loader.Load("", flags)

			require.NoError(t, err)
			assert.Equal(t, level, cfg.LogLevel)
		})
	}
}

func TestLoader_Load_AllValidThemes(t *testing.T) {
	validThemes := []string{"default", "modern", "catppuccin", "nord", "dracula"}

	for _, theme := range validThemes {
		t.Run(theme, func(t *testing.T) {
			cleanEnv(t)

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flags.String("default-theme", "", "default theme")
			flags.Set("default-theme", theme)

			loader := NewLoader(nil)
			cfg, err := loader.Load("", flags)

			require.NoError(t, err)
			assert.Equal(t, theme, cfg.DefaultTheme)
		})
	}
}

func TestLoader_Load_TOMLConfigFile(t *testing.T) {
	cleanEnv(t)

	// Create temporary TOML config file
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.toml")
	cfgContent := `
config_dir = "/tmp/toml-test"
log_level = "debug"
log_format = "json"
default_theme = "catppuccin"
verbose = true
dry_run = false
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check TOML file values
	assert.Equal(t, "/tmp/toml-test", cfg.ConfigDir)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "catppuccin", cfg.DefaultTheme)
	assert.True(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
}

func TestLoader_Load_PartialFlags(t *testing.T) {
	// Test that we can set only some flags and others use defaults
	cleanEnv(t)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	flags.Bool("verbose", false, "verbose")
	// Only set log-level, not verbose
	flags.Set("log-level", "debug")

	loader := NewLoader(nil)
	cfg, err := loader.Load("", flags)

	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.False(t, cfg.Verbose)           // Should use default
	assert.Equal(t, "text", cfg.LogFormat) // Should use default
}

func TestLoader_bindFlags_ErrorHandling(t *testing.T) {
	// Test that bindFlags handles missing flags gracefully
	loader := NewLoader(nil)

	// Create a flagset with only some flags
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	// Don't add all flags - bindFlags should handle missing ones

	err := loader.bindFlags(flags)
	require.NoError(t, err) // Should not error on missing flags
}

func TestLoader_Load_EmptyConfigDir(t *testing.T) {
	// Test that empty config_dir fails validation
	cleanEnv(t)

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	cfgContent := `
config_dir: ""
`
	err := os.WriteFile(cfgFile, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	loader := NewLoader(nil)
	cfg, err := loader.Load(cfgFile, nil)

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "config_dir cannot be empty")
}

// Helper function to clean test environment
func cleanEnv(t *testing.T) {
	t.Helper()

	envVars := []string{
		"ZEROUI_CONFIG_DIR",
		"ZEROUI_LOG_LEVEL",
		"ZEROUI_LOG_FORMAT",
		"ZEROUI_DEFAULT_THEME",
		"ZEROUI_VERBOSE",
		"ZEROUI_DRY_RUN",
		"ZEROUI_CONFIG",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}

	// Cleanup function to restore environment after test
	t.Cleanup(func() {
		for _, env := range envVars {
			os.Unsetenv(env)
		}
	})
}
