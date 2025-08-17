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

---

## Test stubs and deterministic test environment (developer notes)

To improve test reliability and make CI reproducible, the repository includes repo-local test stubs and helpers. Follow these conventions when adding or modifying tests.

- Repository-local stub binaries
  - Path: `testdata/bin/`
  - Example: `testdata/bin/ghostty` — a minimal stub used by tests that would otherwise exec the real `ghostty` binary.
  - Purpose: let tests run without requiring system-installed CLIs and to provide deterministic, stable output.

- Makefile helper
  - Use `make test-setup` to prepare test stub binaries before running tests locally or in CI. This target ensures executable bits are set:
    - chmod +x ./testdata/bin/\*
  - `make test` already depends on `test-setup`, so CI can simply run `make test`.

- Deterministic HOME & PATH for tests
  - Reusable test helper: `test/helpers/testing_env.go`
    - Functions: `SetupTestEnv(t *testing.T)` and `SetupTestEnvWithHome(t *testing.T, homeDir string)`
    - Behavior: prepends `testdata/bin` to `PATH` (if repo root found) and sets `HOME` to a temporary directory for test isolation. Registers cleanup via `t.Cleanup`.
  - Package-level approach: For broad package coverage, prefer adding a `TestMain` to the package (e.g., `internal/tui/testmain_test.go`) that prepends `testdata/bin` to `PATH`. Avoid globally changing `HOME` in package TestMain for integration packages that manage `HOME` per-test.

- Best practices
  - Call `helpers.SetupTestEnv(t)` at the start of tests that run external binaries or interact with the user's home directory.
  - Prefer package-level `TestMain` for packages where nearly every test benefits from the deterministic PATH.
  - Commit stub binaries with the executable bit set (git mode 100755) so CI picks them up correctly. If necessary, run `git update-index --chmod=+x testdata/bin/*`.
  - In CI pipelines, run `make test-setup` (or `make test`) before `go test` to ensure stubs are prepared.

- Example usage
  - Per-test:
    ```go
    func TestSomething(t *testing.T) {
        helpers.SetupTestEnv(t)
        // ... test code ...
    }
    ```
  - Package-level TestMain (illustrative):
    ```go
    func TestMain(m *testing.M) {
        // prepend testdata/bin to PATH, set HOME if desired, run tests
        os.Exit(m.Run())
    }
    ```

These practices are intentionally non-invasive: the repo-local stubs are only preferred for test runs (by PATH ordering) and the fallback behavior is limited to tests so production behavior is unchanged.

If you add new tests that exec external tools, add a small stub under `testdata/bin` and update `test-setup` if additional permissions are required.
