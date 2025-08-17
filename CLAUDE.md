# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It features both CLI and interactive TUI interfaces built with Go, Cobra, and Bubble Tea.

## Development Commands

### Building and Running
```bash
make build          # Build binary to build/zeroui
make install        # Install to GOPATH/bin
make run            # Build and run the application
go run .            # Direct run (launches interactive grid)
```

### Testing
```bash
make test           # Run all tests with coverage
go test ./...       # Run tests without coverage
go test -v ./internal/plugins/rpc  # Test specific package
go test -run TestSpecificFunction  # Run specific test
make test-integration              # Run integration tests
```

### Code Quality
```bash
make lint           # Run golangci-lint (requires installation)
make fmt            # Format code with gofmt and goimports  
make security       # Run gosec security checks
make check          # Run all quality checks (lint + test + security)
```

### Development Workflow
```bash
make dev            # Start with file watching (requires entr)
make watch          # Watch for changes and rebuild
make pre-commit     # Run pre-commit checks (fmt + lint + test + security)
```

### Plugin Development
```bash
# Build RPC plugin
cd plugins/ghostty-rpc && go build -o zeroui-plugin-ghostty-rpc

# Test plugin
cd plugins/ghostty-rpc && go test -v

# Plugin naming convention: zeroui-plugin-{name}
```

## Architecture Overview

### Core Structure
- **cmd/** - CLI commands and entry points using Cobra
- **internal/** - Application internals (private packages)
  - **config/** - Configuration management with Koanf
  - **tui/** - Terminal UI built with Bubble Tea and Huh
  - **toggle/** - Core configuration toggle operations  
  - **service/** - Business logic layer
  - **plugins/rpc/** - RPC-based plugin system using hashicorp/go-plugin
  - **observability/** - Metrics, logging, and tracing
- **plugins/** - External RPC plugins
- **pkg/** - Reusable public packages

### Plugin Architecture

ZeroUI implements a streamlined RPC-based plugin system for cross-language compatibility and process isolation:

```go
// Plugin interface - all plugins implement this via gRPC
type ConfigPlugin interface {
    GetInfo(ctx context.Context) (*PluginInfo, error)
    DetectConfig(ctx context.Context) (*ConfigInfo, error)
    ParseConfig(ctx context.Context, path string) (*ConfigData, error)
    WriteConfig(ctx context.Context, path string, data *ConfigData) error
    ValidateField(ctx context.Context, field string, value interface{}) error
    ValidateConfig(ctx context.Context, data *ConfigData) error
    GetSchema(ctx context.Context) (*ConfigMetadata, error)
    SupportsFeature(ctx context.Context, feature string) (bool, error)
}

// Registry usage
registry := rpc.NewRegistry("/path/to/plugins")
plugin, err := registry.LoadPlugin("ghostty-rpc")
```

**Plugin Discovery**: Plugins must follow naming convention `zeroui-plugin-{name}` for automatic discovery.

**Benefits**: Process isolation, cross-language support, hot reload capability, sandboxed execution.

### TUI Architecture

The Terminal UI is built with Bubble Tea and uses a component-based architecture:

- **App Grid** - Main interface showing available applications
- **Config Editor** - In-app configuration editing with Huh forms
- **Responsive Design** - Adapts to different terminal sizes
- **Visual Testing** - Comprehensive snapshot and regression testing

### Configuration Flow

1. **Detection** - Plugins detect configuration files automatically
2. **Parsing** - Structured parsing of config formats (YAML, TOML, custom)
3. **Validation** - Field-level and config-level validation
4. **Editing** - Interactive forms or direct CLI toggles
5. **Writing** - Safe config file updates preserving structure

### Key Dependencies

- **CLI**: `spf13/cobra` for command structure
- **TUI**: `charmbracelet/bubbletea` and `charmbracelet/huh` for interactive interface
- **Config**: `knadh/koanf` for configuration management
- **Plugins**: `hashicorp/go-plugin` for RPC plugin system
- **Testing**: Comprehensive visual regression and snapshot testing

### Commands Structure

- `zeroui` (no args) - Launch interactive app grid
- `zeroui toggle <app> <field> <value>` - Direct configuration toggle
- `zeroui ui <app>` - Launch app-specific configuration UI
- `zeroui list apps` - List supported applications
- `zeroui preset <app> <preset>` - Apply configuration preset

### Testing Strategy

- **Unit Tests** - Individual component testing
- **Integration Tests** - Plugin and service integration 
- **Visual Regression** - TUI appearance and behavior testing with snapshots
- **Performance Tests** - Benchmarking critical paths
- **Plugin Tests** - RPC communication and plugin lifecycle

The codebase emphasizes clean architecture with separation of concerns, comprehensive testing, and maintainable plugin architecture for extensibility.