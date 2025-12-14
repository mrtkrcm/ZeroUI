# Claude Code Integration Guide

This file provides essential guidance for Claude Code (claude.ai/code) when working with the ZeroUI repository.

## Project Overview

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both CLI commands and an interactive TUI built with Charm libraries (Bubble Tea, Huh, Lipgloss).

## Essential Commands

### Build & Run

- `make build` - Build binary to `./build/zeroui`
- `make run` - Build and run TUI application
- `./build/zeroui` - Run interactive TUI

### Testing

- `make test-setup` - Prepare test stubs (required)
- `make test-fast` - Run fast unit tests
- `make test-deterministic` - Full deterministic test suite
- `make test` - Complete test run with coverage

### Code Quality

- `make fmt` - Format code
- `make lint` - Run linters
- `make security` - Security checks

## Architecture

### Core Structure

```
cmd/           # CLI commands (Cobra)
internal/      # Application internals
├── config/   # Configuration management
├── tui/      # Terminal UI (Bubble Tea)
├── toggle/   # Core toggle operations
└── plugins/  # Plugin system (gRPC)
pkg/          # Public packages
testdata/     # Test fixtures and stubs
```

### Key Patterns

- **Clean Architecture**: CLI ↔ Business Logic ↔ UI separation
- **Bubble Tea MVC**: Model/Update/View pattern for TUI components
- **Plugin Architecture**: gRPC-based extensible plugins
- **Test Isolation**: Deterministic stubs in `testdata/bin/`

## Testing Environment

**Important**: Run `make test-setup` before testing to prepare deterministic stubs in `testdata/bin/`.

### Test Categories

- **Unit tests**: Fast, isolated component tests
- **Integration tests**: Full end-to-end workflows
- **Visual tests**: TUI regression testing with baselines
- **Deterministic mode**: `ZEROUI_TEST_MODE=true` for CI consistency

## Development Workflow

1. **Setup**: Install Go 1.24+
2. **Build**: `make build`
3. **Test**: `make test-setup && make test-fast`
4. **Format**: `make fmt`
5. **Quality**: `make lint`

## Key Implementation Notes

### Configuration Management

- `koanf` library for flexible config parsing (YAML/TOML/JSON)
- App-specific providers in `internal/config/providers/`
- Validation with Go playground validator

### TUI Components

- Charm ecosystem (Bubble Tea, Lipgloss, Huh)
- Responsive design with theme support
- Visual regression testing

### Plugin System

- gRPC-based plugins for extensibility
- Plugin registry with health checking
- Example: `plugins/ghostty-rpc/`

## Common Patterns

### Adding New App Support

1. Create provider in `internal/config/providers/`
2. Register in app registry
3. Add config templates and validation
4. Create tests with stubs

### TUI Component Development

1. Follow Bubble Tea Model/Update/View pattern
2. Use styles from `internal/tui/styles/theme.go`
3. Add visual regression tests
4. Update baselines with `make test-update-baselines`

## Environment Variables

- `FAST_TUI_TESTS=true` - Skip slow visual tests
- `ZEROUI_TEST_MODE=true` - Enable deterministic test mode
- `UPDATE_TUI_BASELINES=true` - Update visual baselines

## Dependencies

- **Core**: Go 1.24+, Charm libraries (bubbletea, lipgloss, huh)
- **Config**: koanf, viper, go-playground/validator
- **Plugins**: gRPC, protobuf
- **Dev**: golangci-lint, air, entr
