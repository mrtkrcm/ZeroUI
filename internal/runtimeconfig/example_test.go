package runtimeconfig_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
	"github.com/spf13/pflag"
)

// Example_basic demonstrates basic usage with defaults
func Example_basic() {
	// Clean environment for example
	os.Unsetenv("ZEROUI_CONFIG_DIR")
	os.Unsetenv("ZEROUI_LOG_LEVEL")

	loader := runtimeconfig.NewLoader(nil)
	cfg, err := loader.Load("", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)
	fmt.Printf("Theme: %s\n", cfg.DefaultTheme)
	// Output:
	// Log Level: info
	// Log Format: text
	// Theme: modern
}

// Example_withEnvironment demonstrates loading from environment variables
func Example_withEnvironment() {
	// Set environment variables
	os.Setenv("ZEROUI_LOG_LEVEL", "debug")
	os.Setenv("ZEROUI_LOG_FORMAT", "json")
	os.Setenv("ZEROUI_VERBOSE", "true")
	defer func() {
		os.Unsetenv("ZEROUI_LOG_LEVEL")
		os.Unsetenv("ZEROUI_LOG_FORMAT")
		os.Unsetenv("ZEROUI_VERBOSE")
	}()

	loader := runtimeconfig.NewLoader(nil)
	cfg, err := loader.Load("", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)
	fmt.Printf("Verbose: %v\n", cfg.Verbose)
	// Output:
	// Log Level: debug
	// Log Format: json
	// Verbose: true
}

// Example_withFlags demonstrates loading with command-line flags
func Example_withFlags() {
	// Clean environment
	os.Unsetenv("ZEROUI_LOG_LEVEL")

	// Create and configure flags
	flags := pflag.NewFlagSet("example", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	flags.String("default-theme", "", "default theme")
	flags.Bool("dry-run", false, "dry run mode")

	// Set flag values
	flags.Set("log-level", "warn")
	flags.Set("default-theme", "nord")
	flags.Set("dry-run", "true")

	loader := runtimeconfig.NewLoader(nil)
	cfg, err := loader.Load("", flags)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Theme: %s\n", cfg.DefaultTheme)
	fmt.Printf("Dry Run: %v\n", cfg.DryRun)
	// Output:
	// Log Level: warn
	// Theme: nord
	// Dry Run: true
}

// Example_withConfigFile demonstrates loading from a YAML config file
func Example_withConfigFile() {
	// Clean environment
	os.Unsetenv("ZEROUI_LOG_LEVEL")

	// Create a temporary config file
	tmpDir := os.TempDir()
	configFile := filepath.Join(tmpDir, "example-config.yaml")
	configContent := `log_level: error
log_format: json
default_theme: dracula
verbose: true
`
	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		fmt.Printf("Error creating config file: %v\n", err)
		return
	}
	defer os.Remove(configFile)

	loader := runtimeconfig.NewLoader(nil)
	cfg, err := loader.Load(configFile, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)
	fmt.Printf("Theme: %s\n", cfg.DefaultTheme)
	fmt.Printf("Verbose: %v\n", cfg.Verbose)
	// Output:
	// Log Level: error
	// Log Format: json
	// Theme: dracula
	// Verbose: true
}

// Example_precedence demonstrates configuration precedence
func Example_precedence() {
	// Clean environment
	os.Unsetenv("ZEROUI_LOG_LEVEL")
	os.Unsetenv("ZEROUI_DEFAULT_THEME")

	// Create a config file (lowest priority)
	tmpDir := os.TempDir()
	configFile := filepath.Join(tmpDir, "precedence-config.yaml")
	configContent := `log_level: info
default_theme: default
log_format: text
`
	if err := os.WriteFile(configFile, []byte(configContent), 0o644); err != nil {
		fmt.Printf("Error creating config file: %v\n", err)
		return
	}
	defer os.Remove(configFile)

	// Set environment variable (medium priority)
	os.Setenv("ZEROUI_LOG_LEVEL", "warn")
	os.Setenv("ZEROUI_DEFAULT_THEME", "catppuccin")
	defer func() {
		os.Unsetenv("ZEROUI_LOG_LEVEL")
		os.Unsetenv("ZEROUI_DEFAULT_THEME")
	}()

	// Create flags (highest priority)
	flags := pflag.NewFlagSet("precedence", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	flags.Set("log-level", "debug")

	loader := runtimeconfig.NewLoader(nil)
	cfg, err := loader.Load(configFile, flags)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// log_level: flag (debug) > env (warn) > file (info)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	// default_theme: env (catppuccin) > file (default)
	fmt.Printf("Theme: %s\n", cfg.DefaultTheme)
	// log_format: file (text) > default
	fmt.Printf("Log Format: %s\n", cfg.LogFormat)
	// Output:
	// Log Level: debug
	// Theme: catppuccin
	// Log Format: text
}

// Example_validation demonstrates validation errors
func Example_validation() {
	// Clean environment
	os.Unsetenv("ZEROUI_LOG_LEVEL")

	flags := pflag.NewFlagSet("validation", pflag.ContinueOnError)
	flags.String("log-level", "", "log level")
	flags.Set("log-level", "invalid-level")

	loader := runtimeconfig.NewLoader(nil)
	_, err := loader.Load("", flags)
	if err != nil {
		fmt.Println("Validation failed as expected")
		fmt.Println("Error contains 'invalid log_level':", err.Error()[:50])
	}
	// Output:
	// Validation failed as expected
	// Error contains 'invalid log_level': config validation failed: invalid log_level: inval
}

// Example_defaultConfigDir demonstrates DefaultConfigDir function
func Example_defaultConfigDir() {
	// Without environment variable
	os.Unsetenv("ZEROUI_CONFIG_DIR")
	defaultDir := runtimeconfig.DefaultConfigDir()
	fmt.Printf("Uses home directory: %v\n", defaultDir != "")

	// With environment variable
	os.Setenv("ZEROUI_CONFIG_DIR", "/custom/config/path")
	defer os.Unsetenv("ZEROUI_CONFIG_DIR")

	customDir := runtimeconfig.DefaultConfigDir()
	fmt.Printf("Custom directory: %s\n", customDir)
	// Output:
	// Uses home directory: true
	// Custom directory: /custom/config/path
}
