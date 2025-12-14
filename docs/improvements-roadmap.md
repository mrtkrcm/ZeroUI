# Engineering improvements roadmap

This document tracks internal improvements (testing, reliability, performance, security). It intentionally stays high level to avoid going stale.

## Reliability

- Keep CLI/TUI behavior deterministic and testable (fixtures, stubs, hermetic integration tests).
- Standardize user-facing errors and output across commands.
- Reduce hidden side effects (explicit IO, explicit plugin loading, timeouts, cancellation).

## Security

- Tighten path validation and safe file writes (especially backups and restores).
- Validate inputs to external processes and enforce timeouts.
- Periodically run and review `make security` and `make vuln`.

## Performance

- Optimize hot TUI paths (render caching, fewer allocations, incremental updates where feasible).
- Avoid repeated config parsing and add caching where correctness allows.
- Add small benchmarks for hotspots that regress easily.

## Maintainability

- Keep dependency wiring in the container; avoid constructors that create hidden dependencies.
- Keep `pkg/` small and stable; keep implementation details in `internal/`.
- Prefer clear interfaces at subsystem boundaries (config load/save, plugins, TUI state).
