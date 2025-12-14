# Architecture Overview — ZeroUI

ZeroUI follows clean architecture principles with separation of concerns.

Primary layout

```
cmd/                    # CLI commands and entry points
internal/               # Application internals (config, service, tui, etc.)
pkg/                    # Reusable packages
tools/                  # Development tools
.github/workflows/      # CI/CD pipelines
```

Key components

- `internal/config` - Configuration management
- `internal/tui` - Terminal UI components and views
- `internal/toggle` - Core toggle operations
- `internal/service` - Business logic
- `internal/observability` - Metrics and tracing

For more details see docs/README.md and the module code.
