# ZeroUI

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It includes a CLI and an interactive TUI built with Charm (Bubble Tea, Huh, Lipgloss).

## Quick start

```bash
make build
./build/zeroui           # Launch TUI
./build/zeroui ui ghostty # Open Ghostty config form directly
```

## Keybindings (TUI)

- Global
  - `q`/`Ctrl+C`: quit
  - `?`: toggle Help
  - `/`: search (where supported)
- App List
  - `enter`/`space`: select app
  - `r`: refresh apps
- Form (Config Editor)
  - `tab`/`shift+tab`: navigate fields
  - `enter`: select/confirm
  - `ctrl+s`: save
  - `C`: toggle changed-only view
  - `p`: open presets selector
  - `u`: undo last save (restore most recent backup)
  - `esc`: back to app list
- Help
  - `?`/`esc`: close

## Development

- Fast test suite (short mode):

```bash
make test-fast
```

- Deterministic full test run (relaxed visuals for CI and PRs):

```bash
make test-deterministic
```

- Update visual baselines (run locally and review diffs before committing):

```bash
make test-update-baselines
```

### Testing environment (quick notes)

The repo includes repo-local test stub binaries and a small Makefile helper to make tests deterministic and CI-friendly.

- Repo-local test stubs
  - Put lightweight stub binaries under `testdata/bin/` (a `ghostty` stub is provided).
  - Tests and package-level `TestMain` implementations prefer `testdata/bin` by prepending it to `PATH`.

- Prepare stubs for CI / local runs
  - Run `make test-setup` to ensure the files under `testdata/bin` are executable before running tests.
  - `make test` will also call `test-setup` automatically.

- Reusable test helper
  - A helper is available at `test/helpers/testing_env.go` exposing `SetupTestEnv(t *testing.T)` and `SetupTestEnvWithHome(t *testing.T, homeDir string)`.
  - Call `helpers.SetupTestEnv(t)` at the start of tests that need deterministic `PATH` and an isolated `HOME`.

Example (local workflow)

```bash
make test-setup    # ensure test stubs are executable
make test          # runs the test-suite (test-setup is invoked automatically)
```

If you add tests that exec external tools, add a simple deterministic stub under `testdata/bin` and prefer injecting the test helper in your package tests.

## Presets

- Press `p` in a form to open the presets selector.
- Use Up/Down to navigate and `enter` to apply a preset.
- A success toast is shown and the form reloads to reflect new values.

## Undo

- Press `u` in a form to restore the most recent backup created during a previous save.
- A toast reports success or failure.
