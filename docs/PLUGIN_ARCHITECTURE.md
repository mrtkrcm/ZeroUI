# Plugin Architecture

## Overview

ConfigToggle implements a streamlined RPC-based plugin architecture using hashicorp/go-plugin for cross-language compatibility and process isolation.

## Core Components

### Registry (`internal/plugins/rpc/registry.go`)

Central plugin management:

```go
registry := rpc.NewRegistry("/path/to/plugins")

// Load plugin
plugin, err := registry.LoadPlugin("ghostty-rpc")

// List all plugins
plugins := registry.ListPlugins()
```

### RPC Infrastructure

- **Protocol**: gRPC with Protocol Buffers (`protocol.proto`)
- **Manager**: Plugin lifecycle management (`manager.go`)
- **Interface**: ConfigPlugin interface (`interface.go`)
- **Conversion**: Type conversion utilities (`convert.go`)

## Plugin Interface

```go
type ConfigPlugin interface {
    GetInfo(ctx context.Context) (*PluginInfo, error)
    DetectConfig(ctx context.Context) (*ConfigInfo, error)
    ParseConfig(ctx context.Context, path string) (*ConfigData, error)
    WriteConfig(ctx context.Context, path string, data *ConfigData) error
    ValidateField(ctx context.Context, field string, value interface{}) error
    ValidateConfig(ctx context.Context, data *ConfigData) error
    GetSchema(ctx context.Context) (*ConfigMetadata, error)
    SupportsFeature(ctx context.Context, feature string) (bool, error)
}
```

## Plugin Discovery

Plugins must follow naming convention: `zeroui-plugin-{name}`

Example: `zeroui-plugin-ghostty-rpc`

## Benefits

- ✅ **Process isolation** - crashes don't affect main app
- ✅ **Cross-language** - write plugins in any language
- ✅ **Security** - sandboxed execution
- ✅ **Hot reload** - update without restarting
- ✅ **Concurrent** - multiple plugins run in parallel

## Example: Ghostty Plugin

See `plugins/ghostty-rpc/main.go` for complete implementation:
- 7 configurable fields (theme, font-family, etc.)
- 4 presets (dark-mode, light-mode, cyberpunk, minimal)
- Config file parsing and writing
- Field validation

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Main App      │    │   RPC Plugin    │
│                 │    │                 │
│  ┌───────────┐  │    │  ┌───────────┐  │
│  │ Registry  │◄─┼────┼─►│ConfigPlugin│  │
│  └───────────┘  │    │  └───────────┘  │
│                 │    │                 │
└─────────────────┘    └─────────────────┘
        │                       │
        └─────── gRPC/Unix ─────┘
              Socket IPC
```

Simple, secure, and extensible.