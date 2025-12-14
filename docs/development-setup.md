# Development setup

## Prerequisites

- Go 1.24+ (see `go.mod`)
- Git
- Make

## Quick start

```bash
make build
make test-fast
make run
```

## Common commands

```bash
make test
make lint
make fmt
make check
```

## Notes

- Tests use deterministic stubs in `testdata/bin/` (run `make test-setup` if needed).
- If TUI snapshots change, update baselines with `make test-update-baselines`.
