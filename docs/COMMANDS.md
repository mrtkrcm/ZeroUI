# Command reference

Use `zeroui --help` and `zeroui <command> --help` for full flag and subcommand details.

## Common commands

| Command   | Purpose                                                  | Example                                     |
| --------- | -------------------------------------------------------- | ------------------------------------------- |
| (no args) | Launch the interactive app grid (terminal required)      | `zeroui`                                    |
| `ui`      | Launch the interactive TUI (optionally for a single app) | `zeroui ui ghostty`                         |
| `list`    | List apps, presets, or keys                              | `zeroui list apps`                          |
| `toggle`  | Set a specific configuration value                       | `zeroui toggle ghostty theme dark`          |
| `cycle`   | Cycle to the next value for a key                        | `zeroui cycle ghostty theme`                |
| `preset`  | Apply a preset (or preview changes)                      | `zeroui preset ghostty minimal --show-diff` |
| `backup`  | List/create/restore/cleanup backups                      | `zeroui backup list ghostty`                |
| `ref`     | Browse and validate reference settings                   | `zeroui ref search ghostty font`            |
| `extract` | Extract configuration from apps                          | `zeroui extract ghostty`                    |
| `design-system` | Launch native design system showcase               | `zeroui design-system`                      |
| `ui-select` | Select and configure UI implementation                 | `zeroui ui-select`                          |

## Shell completion

```bash
zeroui completion zsh > "${fpath[1]}/_zeroui"
```

## Global flags

All subcommands support:

- `--config` (override config file path)
- `-v, --verbose`
- `-n, --dry-run` (show what would change without writing)
