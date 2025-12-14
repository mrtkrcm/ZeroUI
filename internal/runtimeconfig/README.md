# Runtime Configuration Loader

A flexible runtime configuration loader for ZeroUI that supports loading from multiple sources with proper precedence handling.

## Features

- **Multi-source loading**: Flags, environment variables, config files, and defaults
- **Proper precedence**: Flags > Env > File > Defaults
- **Multiple formats**: Supports YAML, JSON, and TOML config files
- **Automatic validation**: Built-in validation for all configuration options
- **Type-safe**: Strongly typed configuration structure
- **Environment integration**: Automatic ZEROUI_ prefix for environment variables

## Installation

This is an internal package. Import it in your ZeroUI code:

```go
import "github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
)

func main() {
    loader := runtimeconfig.NewLoader(nil)
    cfg, err := loader.Load("", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Config Dir: %s\n", cfg.ConfigDir)
    fmt.Printf("Log Level: %s\n", cfg.LogLevel)
    fmt.Printf("Log Format: %s\n", cfg.LogFormat)
}
```

### With Config File

```go
loader := runtimeconfig.NewLoader(nil)
cfg, err := loader.Load("/path/to/config.yaml", nil)
if err != nil {
    log.Fatal(err)
}
```

### With Command-Line Flags

```go
import (
    "github.com/spf13/pflag"
    "github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
)

func main() {
    flags := pflag.NewFlagSet("myapp", pflag.ExitOnError)
    flags.String("config-dir", "", "Configuration directory")
    flags.String("log-level", "", "Log level")
    flags.String("log-format", "", "Log format")
    flags.String("default-theme", "", "Default theme")
    flags.Bool("verbose", false, "Verbose output")
    flags.Bool("dry-run", false, "Dry run mode")
    flags.Parse(os.Args[1:])

    loader := runtimeconfig.NewLoader(nil)
    cfg, err := loader.Load("", flags)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Options

### Config Structure

```go
type Config struct {
    ConfigFile   string  // Path to config file (if used)
    ConfigDir    string  // ZeroUI configuration directory
    LogLevel     string  // Log level: debug, info, warn, error
    LogFormat    string  // Log format: text, json
    DefaultTheme string  // Default theme: default, catppuccin, nord, dracula
    Verbose      bool    // Enable verbose output
    DryRun       bool    // Enable dry-run mode
}
```

### Environment Variables

All configuration options can be set via environment variables:

| Variable | Description | Valid Values |
|----------|-------------|--------------|
| `ZEROUI_CONFIG_DIR` | Configuration directory | Any valid directory path |
| `ZEROUI_LOG_LEVEL` | Logging level | debug, info, warn, error |
| `ZEROUI_LOG_FORMAT` | Log output format | text, json |
| `ZEROUI_DEFAULT_THEME` | Default UI theme | default, modern, catppuccin, nord, dracula |
| `ZEROUI_VERBOSE` | Verbose output | true, false |
| `ZEROUI_DRY_RUN` | Dry-run mode | true, false |

### Config File Formats

#### YAML Example

```yaml
config_dir: /home/user/.config/zeroui
log_level: debug
log_format: json
default_theme: catppuccin
verbose: true
dry_run: false
```

#### JSON Example

```json
{
  "config_dir": "/home/user/.config/zeroui",
  "log_level": "debug",
  "log_format": "json",
  "default_theme": "catppuccin",
  "verbose": true,
  "dry_run": false
}
```

#### TOML Example

```toml
config_dir = "/home/user/.config/zeroui"
log_level = "debug"
log_format = "json"
default_theme = "catppuccin"
verbose = true
dry_run = false
```

## Configuration Precedence

Configuration values are resolved in the following order (highest to lowest priority):

1. **Command-line flags** - Explicitly set via CLI
2. **Environment variables** - ZEROUI_ prefixed env vars
3. **Configuration file** - Values from loaded config file
4. **Default values** - Built-in defaults

### Example

Given:
- Config file: `log_level: info`
- Environment: `ZEROUI_LOG_LEVEL=warn`
- Flag: `--log-level=debug`

Result: `log_level = "debug"` (flag wins)

## Validation

The loader performs automatic validation:

- **LogLevel**: Must be one of: debug, info, warn, error
- **LogFormat**: Must be one of: text, json
- **DefaultTheme**: Must be one of: default, modern, catppuccin, nord, dracula
- **ConfigDir**: Must not be empty
- **ConfigFile**: Must exist if specified

Invalid configurations return descriptive errors.

## Default Values

| Option | Default Value |
|--------|---------------|
| `ConfigDir` | `$HOME/.config/zeroui` (or `ZEROUI_CONFIG_DIR` if set) |
| `LogLevel` | `info` |
| `LogFormat` | `text` |
| `DefaultTheme` | `modern` |
| `Verbose` | `false` |
| `DryRun` | `false` |

## Advanced Usage

### Custom Viper Instance

```go
import "github.com/spf13/viper"

v := viper.New()
v.SetConfigType("yaml")
v.AddConfigPath("/etc/zeroui/")
v.AddConfigPath("$HOME/.config/zeroui/")

loader := runtimeconfig.NewLoader(v)
cfg, err := loader.Load("", nil)
```

### Programmatic Configuration

```go
loader := runtimeconfig.NewLoader(nil)

// Load with environment variables
os.Setenv("ZEROUI_LOG_LEVEL", "debug")
os.Setenv("ZEROUI_VERBOSE", "true")

cfg, err := loader.Load("", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Log Level: %s\n", cfg.LogLevel)  // Output: debug
fmt.Printf("Verbose: %v\n", cfg.Verbose)      // Output: true
```

## Testing

Run the test suite:

```bash
go test ./internal/runtimeconfig/...
```

Run with coverage:

```bash
go test -coverprofile=coverage.out ./internal/runtimeconfig/...
go tool cover -html=coverage.out
```

## API Reference

### NewLoader

```go
func NewLoader(v *viper.Viper) *Loader
```

Creates a new runtime configuration loader. If `v` is nil, a new viper instance is created.

### Load

```go
func (l *Loader) Load(cfgFile string, flags *pflag.FlagSet) (*Config, error)
```

Loads configuration from all sources with proper precedence. Returns the loaded configuration or an error.

### DefaultConfigDir

```go
func DefaultConfigDir() string
```

Returns the default configuration directory. Checks `ZEROUI_CONFIG_DIR` environment variable first, then falls back to `$HOME/.config/zeroui`.

## License

Part of the ZeroUI project.
