# Monorepo notes

This repository contains the main `zeroui` Go module plus additional components (plugins and a Raycast extension).

## Structure

```text
cmd/                CLI commands
internal/           application internals
pkg/                reusable packages
plugins/            plugin implementations (some are separate Go modules)
raycast-extension/  Raycast extension (Node workspace)
docs/               documentation
scripts/            build and maintenance scripts
```

## Go modules

- Root module: `go.mod` at the repository root.
- Some plugins are separate modules (example: `plugins/ghostty-rpc/go.mod`) and use a `replace` directive to point back to the root during development.

## Node workspace (Raycast)

The Raycast extension is managed via `package.json` workspaces:

```bash
npm install
npm run build --workspace=raycast-extension
```
