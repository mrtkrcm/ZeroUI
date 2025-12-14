# ZeroUI

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both CLI commands and an interactive TUI built with Charm libraries (Bubble Tea, Huh, Lipgloss).

## Quick Start

**Prerequisites:**

- Go 1.24+ (see `go.mod`)
- Make and Git

**Setup:**

```bash
make test-setup  # Prepare test stubs
make build       # Build binary
./build/zeroui   # Run TUI
make test-fast   # Run tests
```

## Development

### Build & Run

```bash
make build     # Build binary to ./build/zeroui
make install   # Install to GOPATH/bin
make run       # Build and run TUI
```

### Testing

```bash
make test-setup           # Prepare test stubs
make test-fast            # Fast unit tests
make test-deterministic   # CI-like full test suite
make test                 # Full tests with coverage
make test-update-baselines # Update visual baselines
```

### Code Quality

```bash
make fmt       # Format code
make lint      # Run linters
make security  # Security checks
```

### Development Tools

```bash
make dev        # Watch mode (requires entr)
make test-watch # Test watch mode (requires entr)
```

**Note:** Test stubs in `testdata/bin/` provide deterministic testing without external dependencies.

## TUI Controls

**Global:**

- `q` / `Ctrl+C`: Quit
- `?`: Toggle help
- `/`: Search (where supported)

**App List:**

- `Enter` / `Space`: Select app
- `r`: Refresh apps

**Configuration Editor:**

- `Tab` / `Shift+Tab`: Navigate fields
- `Enter`: Select/confirm
- `Ctrl+S`: Save
- `C`: Changed-only view
- `p`: Open presets
- `u`: Undo last save
- `Esc`: Back to app list

## CLI Examples

```bash
# List available apps
./build/zeroui list apps

# Toggle configuration
./build/zeroui toggle ghostty maximize true

# Apply preset
./build/zeroui preset ghostty minimal

# Preview changes
./build/zeroui preset ghostty minimal --show-diff

# Backup management
./build/zeroui backup list ghostty
./build/zeroui backup create ghostty
./build/zeroui backup cleanup ghostty --keep 3
```

## Testing

ZeroUI uses deterministic test stubs in `testdata/bin/` to avoid external dependencies. Run `make test-setup` before testing to prepare stubs.

When adding tests that execute external tools, create stubs in `testdata/bin/` and use `test/helpers/testing_env.go` utilities.

## Contributing

**PR Checklist:**

- Run `make fmt` to format code
- Run `make test-fast` (or `make test-deterministic` for CI parity)
- Run `make lint` and address critical warnings
- Update visual baselines with `make test-update-baselines` if TUI visuals change
- Keep changes focused and update documentation for user-visible changes

## Troubleshooting

- **Missing test stubs**: Run `make test-setup`
- **Linter not found**: Install golangci-lint or use local formatting
- **Visual test diffs**: Run `make test-update-baselines` and review changes

## Codebase Overview

- `cmd/` - CLI commands (Cobra)
- `internal/tui/` - TUI implementation (Bubble Tea)
- `internal/config/` - Configuration management
- `internal/plugins/` - Plugin system (gRPC)
- `pkg/` - Public packages
- `testdata/bin/` - Test stubs
