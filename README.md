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

## Presets

- Press `p` in a form to open the presets selector.
- Use Up/Down to navigate and `enter` to apply a preset.
- A success toast is shown and the form reloads to reflect new values.

## Undo

- Press `u` in a form to restore the most recent backup created during a previous save.
- A toast reports success or failure.
