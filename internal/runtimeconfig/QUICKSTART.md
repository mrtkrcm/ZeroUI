# Runtime Config Loader - Quick Start Guide

## 5-Minute Integration Guide

### 1. Basic Usage (No Config File, No Flags)

```go
import "github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"

loader := runtimeconfig.NewLoader(nil)
cfg, err := loader.Load("", nil)
if err != nil {
    log.Fatal(err)
}

// Use the config
fmt.Printf("Config directory: %s\n", cfg.ConfigDir)
fmt.Printf("Log level: %s\n", cfg.LogLevel)
```

### 2. With Config File

Create `config.yaml`:
```yaml
log_level: debug
log_format: json
default_theme: catppuccin
verbose: true
```

Load it:
```go
loader := runtimeconfig.NewLoader(nil)
cfg, err := loader.Load("/path/to/config.yaml", nil)
```

### 3. With CLI Flags (Cobra Integration)

```go
import (
    "github.com/spf13/cobra"
    "github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    Run: func(cmd *cobra.Command, args []string) {
        loader := runtimeconfig.NewLoader(nil)
        cfg, err := loader.Load("", cmd.Flags())
        if err != nil {
            log.Fatal(err)
        }

        // Use cfg...
    },
}

func init() {
    rootCmd.Flags().String("config-dir", "", "config directory")
    rootCmd.Flags().String("log-level", "", "log level")
    rootCmd.Flags().String("log-format", "", "log format")
    rootCmd.Flags().String("default-theme", "", "default theme")
    rootCmd.Flags().Bool("verbose", false, "verbose output")
    rootCmd.Flags().Bool("dry-run", false, "dry run")
}
```

### 4. With Environment Variables

```bash
export ZEROUI_LOG_LEVEL=debug
export ZEROUI_VERBOSE=true
./myapp
```

### 5. All Together (Full Precedence)

```bash
# config.yaml has: log_level: info
export ZEROUI_LOG_LEVEL=warn
./myapp --log-level=debug
# Result: log_level = "debug" (flag wins)
```

## Common Patterns

### Pattern 1: Config File with Flag Override

```go
var cfgFile string

rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

// In command execution:
loader := runtimeconfig.NewLoader(nil)
cfg, err := loader.Load(cfgFile, cmd.Flags())
```

### Pattern 2: Environment-First Development

```bash
# .env file
export ZEROUI_CONFIG_DIR=./dev-config
export ZEROUI_LOG_LEVEL=debug
export ZEROUI_VERBOSE=true

source .env
go run .
```

### Pattern 3: Testing with Custom Config

```go
func TestMyFeature(t *testing.T) {
    // Set test environment
    os.Setenv("ZEROUI_LOG_LEVEL", "debug")
    defer os.Unsetenv("ZEROUI_LOG_LEVEL")

    loader := runtimeconfig.NewLoader(nil)
    cfg, err := loader.Load("", nil)
    require.NoError(t, err)

    // Use cfg in test...
}
```

## Configuration Reference

### Valid Values

| Option | Valid Values |
|--------|--------------|
| log_level | `debug`, `info`, `warn`, `error` |
| log_format | `text`, `json` |
| default_theme | `default`, `modern`, `catppuccin`, `nord`, `dracula` |
| verbose | `true`, `false` |
| dry_run | `true`, `false` |

### Environment Variable Names

| Config Key | Environment Variable | Flag Name |
|------------|---------------------|-----------|
| config_dir | ZEROUI_CONFIG_DIR | --config-dir |
| log_level | ZEROUI_LOG_LEVEL | --log-level |
| log_format | ZEROUI_LOG_FORMAT | --log-format |
| default_theme | ZEROUI_DEFAULT_THEME | --default-theme |
| verbose | ZEROUI_VERBOSE | --verbose |
| dry_run | ZEROUI_DRY_RUN | --dry-run |

## Error Handling

```go
loader := runtimeconfig.NewLoader(nil)
cfg, err := loader.Load(cfgFile, flags)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "invalid log_level"):
        log.Fatal("Invalid log level specified")
    case strings.Contains(err.Error(), "failed to read config file"):
        log.Fatal("Config file not found or not readable")
    default:
        log.Fatalf("Configuration error: %v", err)
    }
}
```

## Tips

1. **Use environment variables for development** - Easy to change without editing code
2. **Use config files for production** - Versioned, auditable configuration
3. **Use flags for one-off overrides** - Quick testing without changing files
4. **Check errors** - The loader validates all input, always check for errors
5. **Test all sources** - Verify your app works with env vars, files, and flags

## Troubleshooting

### Config file not loading
- Check file path is absolute or relative to working directory
- Verify file extension matches content (`.yaml`, `.json`, `.toml`)
- Check file permissions

### Environment variables not working
- Ensure `ZEROUI_` prefix is used
- Check variable name uses underscores not hyphens
- Verify export was successful: `echo $ZEROUI_LOG_LEVEL`

### Flags not overriding
- Ensure flag names use hyphens: `--log-level` not `--log_level`
- Verify flags are parsed before `Load()` is called
- Check flag is actually set: `cmd.Flags().Changed("log-level")`

## Examples

See `example_test.go` for runnable examples:
```bash
go test -v -run Example ./internal/runtimeconfig/...
```

## Full Documentation

- `README.md` - Comprehensive package documentation
- `doc.go` - Go package documentation
- `IMPLEMENTATION_SUMMARY.md` - Technical implementation details
