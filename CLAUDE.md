# Claude Code Integration Guide

This file provides essential guidance for Claude Code (claude.ai/code) when working with the ZeroUI repository.

## Project Overview

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both CLI commands and an interactive TUI built with Charm libraries (Bubble Tea, Huh, Lipgloss).

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
make test-setup     # Prepare test stubs (required)
make test-fast      # Run fast unit tests
make test-deterministic # Full deterministic test suite
make test           # Complete test run with coverage
go test ./...       # Run tests without coverage
go test -v ./internal/plugins/rpc  # Test specific package
go test -run TestSpecificFunction  # Run specific test
make test-integration # Run integration tests
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

## Testing Environment

**Important**: Run `make test-setup` before testing to prepare deterministic stubs in `testdata/bin/`.

### Test Categories

- **Unit tests**: Fast, isolated component tests
- **Integration tests**: Full end-to-end workflows
- **Visual tests**: TUI regression testing with baselines
- **Performance tests**: Benchmarking critical paths
- **Plugin tests**: RPC communication and plugin lifecycle
- **Deterministic mode**: `ZEROUI_TEST_MODE=true` for CI consistency

### Test Stubs and Deterministic Environment

To improve test reliability and make CI reproducible, the repository includes repo-local test stubs and helpers. Follow these conventions when adding or modifying tests.

- **Repository-local stub binaries**
  - Path: `testdata/bin/`
  - Example: `testdata/bin/ghostty` — a minimal stub used by tests that would otherwise exec the real `ghostty` binary.
  - Purpose: let tests run without requiring system-installed CLIs and to provide deterministic, stable output.

- **Makefile helper**
  - Use `make test-setup` to prepare test stub binaries before running tests locally or in CI. This target ensures executable bits are set:
    - chmod +x ./testdata/bin/\*
  - `make test` already depends on `test-setup`, so CI can simply run `make test`.

- **Deterministic HOME & PATH for tests**
  - Reusable test helper: `test/helpers/testing_env.go`
    - Functions: `SetupTestEnv(t *testing.T)` and `SetupTestEnvWithHome(t *testing.T, homeDir string)`
    - Behavior: prepends `testdata/bin` to `PATH` (if repo root found) and sets `HOME` to a temporary directory for test isolation. Registers cleanup via `t.Cleanup`.
  - Package-level approach: For broad package coverage, prefer adding a `TestMain` to the package (e.g., `internal/tui/testmain_test.go`) that prepends `testdata/bin` to `PATH`. Avoid globally changing `HOME` in package TestMain for integration packages that manage `HOME` per-test.

- **Best practices**
  - Call `helpers.SetupTestEnv(t)` at the start of tests that run external binaries or interact with the user's home directory.
  - Prefer package-level `TestMain` for packages where nearly every test benefits from the deterministic PATH.
  - Commit stub binaries with the executable bit set (git mode 100755) so CI picks them up correctly. If necessary, run `git update-index --chmod=+x testdata/bin/*`.
  - In CI pipelines, run `make test-setup` (or `make test`) before `go test` to ensure stubs are prepared.

- **Example usage**
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

## Commands Structure

- `zeroui` (no args) - Launch interactive app grid
- `zeroui backup <subcommand>` - Configuration backup management
  - `backup list <app>` - List available backups
  - `backup create <app>` - Create new backup
  - `backup restore <app> <backup>` - Restore from backup
  - `backup cleanup <app>` - Clean old backups
- `zeroui cycle <app> <field>` - Cycle through field values
- `zeroui design-system` - Display design system showcase
- `zeroui enhanced-ui` - Launch enhanced UI with themes
- `zeroui extract [apps]` - Extract configurations to files
- `zeroui list apps` - List supported applications
- `zeroui preset <app> <preset>` - Apply configuration preset
- `zeroui reference <subcommand>` - Reference system management
  - `reference list` - List reference configurations
  - `reference show <app>` - Show reference configuration
  - `reference validate <app>` - Validate reference configuration
  - `reference search <query>` - Search reference configurations
- `zeroui toggle <app> <field> <value>` - Direct configuration toggle
- `zeroui ui <app>` - Launch app-specific configuration UI
- `zeroui ui-select` - Select UI implementation
- `zeroui validate-reference [apps]` - Validate reference configurations

## Environment Variables

- `FAST_TUI_TESTS=true` - Skip slow visual tests
- `ZEROUI_TEST_MODE=true` - Enable deterministic test mode
- `UPDATE_TUI_BASELINES=true` - Update visual baselines
- `ZEROUI_CONFIG_DIR=<path>` - Override config directory location
- `RUN_SNAPSHOTS=true` - Enable snapshot tests (CI mode)
- `GENERATE_TUI_IMAGES=true` - Enable visual regression image generation

### Accessibility Variables

- `ACCESSIBLE=true` / `ACCESSIBILITY=true` / `A11Y=true` - Enable accessibility mode
- `SCREEN_READER=true` / `NVDA=true` / `JAWS=true` / `ORCA=true` - Enable screen reader mode
- `HIGH_CONTRAST=true` / `CONTRAST=high` - Enable high contrast mode
- `REDUCED_MOTION=true` / `MOTION=reduced` - Enable reduced motion
- `NO_COLOR=true` / `MONOCHROME=true` - Disable color output
- `VERBOSE_DESCRIPTIONS=true` / `ACCESSIBILITY_VERBOSE=true` - Enable verbose descriptions
- `SIMPLE_UI=true` / `ACCESSIBILITY_SIMPLE=true` - Enable simplified UI

## Dependencies

- **Core**: Go 1.24+, Charm libraries (bubbletea, lipgloss, huh)
- **Config**: koanf, viper, go-playground/validator
- **Plugins**: gRPC, protobuf
- **Dev**: golangci-lint, air, entr
