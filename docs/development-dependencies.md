# Dependencies

This document summarizes the main dependencies used by ZeroUI. For exact versions, see `go.mod`.

## Core

- TUI: Bubble Tea (`github.com/charmbracelet/bubbletea`) and Bubbles (`github.com/charmbracelet/bubbles`)
- Styling: Lipgloss (`github.com/charmbracelet/lipgloss`)
- Forms: Huh (`github.com/charmbracelet/huh`)
- CLI: Cobra (`github.com/spf13/cobra`)
- Config: Koanf (`github.com/knadh/koanf/v2`) and Viper (`github.com/spf13/viper`)
- Plugins: `hashicorp/go-plugin`, gRPC, Protobuf

## Dev tooling (optional)

- Lint: `golangci-lint`
- Security: `gosec`, `govulncheck`

## Reference

Library-specific notes live under `docs/cheatsheets/`.
