# Architecture overview

ZeroUI is a Go CLI + TUI application designed around small, testable packages and a clear separation between command wiring, business logic, and UI.

## Repository layout

```text
cmd/       Cobra commands and CLI wiring
internal/  Application internals (config, service, tui, plugins, etc.)
pkg/       Reusable packages (public API surface)
configs/   Curated reference metadata used by `zeroui ref`
test/      Integration tests
testdata/  Deterministic fixtures and stub binaries
```

## Major subsystems

- `internal/config/`: app registry, config loading/merging, format handling
- `internal/toggle/`: toggle/cycle/preset execution and safe write behavior
- `internal/tui/`: Bubble Tea application and views
- `internal/plugins/`: plugin interfaces and implementations
- `internal/container/`: dependency wiring for commands and runtime services
- `internal/runtimeconfig/`: runtime config loading and defaults
- `internal/security/` and `internal/validation/`: input/path validation and hardening

## Related documents

- Plugin system: `plugin-architecture.md` and `rpc-plugin-guide.md`
- Reference metadata: `reference-system.md`
