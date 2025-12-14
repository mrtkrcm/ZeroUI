# Plugin architecture

ZeroUI supports out-of-process plugins for config detection, parsing, validation, and writing. Plugins run as separate processes and communicate with the main binary over gRPC via `hashicorp/go-plugin`.

## Components

- Registry and discovery: `internal/plugins/rpc/registry.go`
- Lifecycle management: `internal/plugins/rpc/manager.go`
- Protocol and types: `internal/plugins/rpc/protocol.proto` and generated code

## Discovery

Plugins are discovered by executable name:

- Convention: `zeroui-plugin-{name}`
- Example: `zeroui-plugin-ghostty-rpc`

## Why separate processes

- Isolation: plugin crashes do not take down the main app.
- Language-agnostic: plugins can be implemented outside Go.
- Security and control: explicit boundaries around file access and parsing.

## Implementation notes

For a working example and a development walkthrough, see `rpc-plugin-guide.md` and `plugins/ghostty-rpc/main.go`.
