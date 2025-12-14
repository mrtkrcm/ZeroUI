// Package runtimeconfig provides runtime configuration management for ZeroUI.
//
// This package handles loading configuration from multiple sources with proper
// precedence ordering: flags > environment variables > config file > defaults.
//
// # Basic Usage
//
// Load configuration with defaults only:
//
//	loader := runtimeconfig.NewLoader(nil)
//	cfg, err := loader.Load("", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Log level: %s\n", cfg.LogLevel)
//
// # Loading from a Config File
//
// Load configuration from a YAML, JSON, or TOML file:
//
//	loader := runtimeconfig.NewLoader(nil)
//	cfg, err := loader.Load("/path/to/config.yaml", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Using with Command-Line Flags
//
// Integrate with pflag for CLI applications:
//
//	flags := pflag.NewFlagSet("myapp", pflag.ExitOnError)
//	flags.String("config-dir", "", "Configuration directory")
//	flags.String("log-level", "", "Log level (debug, info, warn, error)")
//	flags.String("log-format", "", "Log format (text, json)")
//	flags.String("default-theme", "", "Default theme")
//	flags.Bool("verbose", false, "Verbose output")
//	flags.Bool("dry-run", false, "Dry run mode")
//	flags.Parse(os.Args[1:])
//
//	loader := runtimeconfig.NewLoader(nil)
//	cfg, err := loader.Load("", flags)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Environment Variables
//
// All configuration options can be set via environment variables with the
// ZEROUI_ prefix:
//
//   - ZEROUI_CONFIG_DIR - Configuration directory
//   - ZEROUI_LOG_LEVEL - Log level (debug, info, warn, error)
//   - ZEROUI_LOG_FORMAT - Log format (text, json)
//   - ZEROUI_DEFAULT_THEME - Default theme (default, modern, catppuccin, nord, dracula)
//   - ZEROUI_VERBOSE - Verbose output (true, false)
//   - ZEROUI_DRY_RUN - Dry run mode (true, false)
//
// # Configuration Precedence
//
// Configuration values are loaded with the following precedence (highest to lowest):
//
//  1. Command-line flags
//  2. Environment variables (with ZEROUI_ prefix)
//  3. Configuration file
//  4. Default values
//
// # Validation
//
// The loader automatically validates configuration values:
//
//   - LogLevel must be one of: debug, info, warn, error
//   - LogFormat must be one of: text, json
//   - DefaultTheme must be one of: default, modern, catppuccin, nord, dracula
//   - ConfigDir must not be empty
//   - ConfigFile path must exist if specified
//
// # Custom Viper Instance
//
// You can provide your own viper instance for advanced use cases:
//
//	v := viper.New()
//	v.SetConfigType("yaml")
//	// ... additional viper configuration
//
//	loader := runtimeconfig.NewLoader(v)
//	cfg, err := loader.Load("/path/to/config.yaml", nil)
package runtimeconfig
