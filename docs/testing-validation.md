# Testing and validation

## Quick checks

```bash
make test-fast
make lint
```

## Full checks (CI-like)

```bash
make check
```

## Deterministic test environment

Some tests use stub binaries in `testdata/bin/` to avoid relying on external tools. If tests fail due to missing stubs:

```bash
make test-setup
```

## TUI baselines

If you change TUI visuals and snapshot tests fail:

```bash
make test-update-baselines
```
