# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ZeroUI is a delightful zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both CLI commands and an enhanced interactive TUI built with Charm libraries (Bubble Tea, Huh, Lipgloss), featuring intelligent notifications, contextual help, and smooth animations.

## Essential Commands

### Building and Running
- `make build` - Build the binary to `./build/zeroui`
- `make install` - Install to GOPATH/bin
- `make run` - Build and run the application
- `./build/zeroui` - Run the interactive TUI (main use case)
- `./build/zeroui toggle <app> <key> <value>` - CLI toggle functionality

### Testing
- `make test-setup` - Prepare deterministic test stubs (required before testing)
- `make test-fast` - Run fast unit tests only (short mode, no integration/perf)
- `make test-deterministic` - Full test suite in deterministic/fast mode (CI-like)
- `make test` - Full test run with coverage and HTML report
- `make test-update-baselines` - Update visual test baselines (run and review diffs carefully)
- `make test-watch` - Run tests in watch mode (requires entr)

### Code Quality
- `make fmt` - Format code with gofmt and goimports
- `make lint` - Run golangci-lint (requires golangci-lint installed)
- `make security` - Run security checks (gosec, nancy)
- `make vuln` - Check for known vulnerabilities

### Development
- `make dev` - Start development environment with file watching (requires entr)
- `make air` - Hot reloading development server (installs air if needed)
- `go test ./internal/tui -run TestSpecificTest` - Run individual tests

## Architecture

### Core Structure
```
cmd/                    # CLI commands and Cobra entry points
├── root.go            # Main CLI root command
├── ui.go              # TUI command entry point
└── toggle.go          # Core toggle functionality

internal/               # Application internals
├── config/            # Configuration management and parsing
│   ├── apps_registry.go    # Application discovery and registry
│   ├── loader.go          # Config file loading and validation
│   └── providers/         # App-specific config providers (ghostty, etc)
├── tui/               # Terminal UI implementation (Bubble Tea)
│   ├── app_core.go        # Main TUI application struct
│   ├── components/        # Reusable UI components
│   │   ├── app/           # App selection and scanning
│   │   ├── forms/         # Configuration editing forms
│   │   └── display/       # Status bars, help, notifications
│   ├── styles/theme.go    # Design system and themes
│   ├── animations/        # Animation and transition effects
│   ├── feedback/          # Notifications and loading states
│   └── help/              # Contextual help system
├── toggle/            # Core toggle operations engine
├── service/           # Business logic services
└── plugins/rpc/       # Plugin system with gRPC

pkg/                   # Public reusable packages
├── configextractor/   # Config extraction utilities
└── reference/         # Reference system for app configs
```

### Key Design Patterns

1. **Clean Architecture**: Separation between CLI commands, business logic (toggle engine), and UI components
2. **Bubble Tea MVC**: TUI components follow the Bubble Tea pattern with Model, Update, View methods
3. **Plugin Architecture**: Extensible via gRPC plugins for new application support
4. **Configuration Registry**: Centralized app discovery and config management
5. **Test Isolation**: Comprehensive test environment with deterministic stubs in `testdata/bin/`

### TUI State Machine
The TUI uses a state machine with these primary views:
- **ListView**: App grid selection with responsive layout
- **FormView**: Configuration editing with enhanced UX
- **HelpView**: Contextual help overlay
- **ProgressView**: Loading states and operation feedback

## Testing Environment

**IMPORTANT**: Always run `make test-setup` before running tests. This ensures:
- Test stub binaries in `testdata/bin/` are executable
- Git index marks them as executable
- Deterministic test environment is prepared

### Test Categories
- **Fast tests**: Use `FAST_TUI_TESTS=true` environment variable
- **Deterministic tests**: Use `ZEROUI_TEST_MODE=true` for CI-like behavior
- **Visual tests**: TUI visual regression tests with baseline comparison
- **Integration tests**: Full end-to-end testing with real config files

### Test Stubs
The repository uses deterministic test stubs in `testdata/bin/` to avoid dependency on system-installed tools. When adding tests that exec external tools, create corresponding stubs here.

## Development Workflow

1. **Setup**: Ensure Go 1.24+ is installed
2. **Dependencies**: Run `make deps` to download and verify modules
3. **Build**: Use `make build` to create the binary
4. **Test**: Always run `make test-setup` first, then `make test-fast` for quick iteration
5. **Format**: Run `make fmt` before committing
6. **Quality**: Use `make lint` to catch issues

## Key Implementation Notes

### Configuration Management
- Uses `koanf` library for flexible config parsing
- Supports YAML, TOML, JSON config formats
- App-specific providers in `internal/config/providers/`
- Configuration validation with Go playground validator

### TUI Components
- Built with Charm ecosystem (Bubble Tea, Lipgloss, Huh)
- Responsive design system with theme support
- Animation system for smooth transitions
- Comprehensive visual regression testing

### Plugin System
- gRPC-based plugins for extending app support
- Plugin registry with health checking
- Example plugin: `plugins/ghostty-rpc/`

### Performance Optimizations
- Object pooling for frequent allocations
- Concurrent configuration loading
- Optimized string interning for config keys
- Memory-efficient parser pools

## Common Patterns

### Adding New App Support
1. Create provider in `internal/config/providers/`
2. Register in `apps_registry.yaml`
3. Add configuration templates and validation
4. Create comprehensive tests with stubs

### TUI Component Development
1. Follow Bubble Tea Model/Update/View pattern
2. Use existing styles from `internal/tui/styles/theme.go`
3. Add visual regression tests
4. Update baselines with `make test-update-baselines`

### Plugin Development
1. Use `internal/plugins/rpc/` as reference
2. Implement gRPC protocol defined in `protocol.proto`
3. Register with plugin manager
4. Add integration tests

## Environment Variables

- `FAST_TUI_TESTS=true` - Skip slow visual tests
- `ZEROUI_TEST_MODE=true` - Enable deterministic test mode
- `UPDATE_TUI_BASELINES=true` - Update visual test baselines
- `CGO_ENABLED=0` - Disabled for static binary builds

## Dependencies

- **Core**: Go 1.24+, Charm libraries (bubbletea, lipgloss, huh)
- **Testing**: testify, golang/mock for mocking
- **Config**: koanf, viper, go-playground/validator
- **Plugins**: gRPC, protobuf
- **Development**: golangci-lint, air (hot reloading), entr (file watching)