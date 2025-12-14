# ZeroUI

ZeroUI is a zero-configuration UI toolkit manager for developers. It provides a CLI and an interactive TUI for managing application settings, presets, and safe config edits.

## Install

Using Go (requires repository access):

```bash
go install github.com/mrtkrcm/ZeroUI@latest
```

From source:

```bash
make build
./build/zeroui --help
```

## Quick usage

```bash
zeroui                 # Launch interactive app grid (terminal required)
zeroui list apps
zeroui toggle ghostty theme dark
zeroui preset ghostty minimal --show-diff
zeroui backup list ghostty
```

## Documentation

- `docs/INDEX.md` (start here)
- `docs/QUICKSTART.md` (install + first run)
- `docs/COMMANDS.md` (CLI reference)
- `docs/dev/SETUP.md` (development setup)
- `docs/ARCHITECTURE.md` and `docs/PLUGIN_ARCHITECTURE.md` (system design)

## Development

```bash
make build
make run
make test-fast
make test-deterministic
make lint
```

Notes:

- Tests use deterministic stubs in `testdata/bin/` (run `make test-setup` if needed).
- If TUI snapshots change, update baselines with `make test-update-baselines`.

## Contributing

See `docs/CONTRIBUTING.md`.
