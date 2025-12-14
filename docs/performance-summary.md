# Performance notes

This directory captures performance-related notes and references for ZeroUI.

## Where optimizations live

- Rendering and allocation helpers: `internal/performance/`
- TUI hot paths: `internal/tui/`
- Validation performance: `internal/validation/`

## How to profile locally

```bash
make benchmark
make profile
```
