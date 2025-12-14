# Reference system

ZeroUI ships with curated reference metadata for supported applications. Reference data is used to:

- Browse available settings and their meaning.
- Validate values before writing config.
- Improve the TUI/CLI experience (search, descriptions, allowed values).

## Data source

Reference metadata lives in `configs/` (for example: `configs/ghostty.yaml`, `configs/zed.yaml`, `configs/mise.yaml`).

## CLI usage

The `ref` command reads the curated metadata:

```bash
zeroui ref list
zeroui ref show ghostty
zeroui ref search ghostty font
zeroui ref validate ghostty font_size 14
```

## Development utilities

`validate-reference` validates that reference metadata can be loaded and mapped for one or all apps:

```bash
zeroui validate-reference ghostty
zeroui validate-reference --all
```

## Adding or updating reference data

1. Edit or add an app file under `configs/`.
2. Confirm `zeroui ref show <app>` and `zeroui ref validate <app> <setting> <value>` behave as expected.
3. Run `zeroui validate-reference <app>` to catch mapping issues early.
