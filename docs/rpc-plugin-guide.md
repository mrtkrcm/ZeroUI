# RPC plugin guide

This guide describes how to write an out-of-process plugin that ZeroUI can discover and run.

For system-level context, see `plugin-architecture.md`.

## Create a plugin

1. Create a directory (in-repo example):

```bash
mkdir -p plugins/my-plugin
```

2. Implement a `main.go` that serves the gRPC plugin:

```go
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/mrtkrcm/ZeroUI/internal/plugins/rpc"
)

type MyPlugin struct{}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpc.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"config": &rpc.ConfigPluginGRPC{Impl: &MyPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

3. Build the binary with the discovery name:

```bash
go build -o zeroui-plugin-my-plugin ./plugins/my-plugin
```

## Discovery name

ZeroUI discovers plugin executables named `zeroui-plugin-{name}`.

If you build `zeroui-plugin-my-plugin`, the plugin name to load/discover is `my-plugin`.

## Implementation checklist

- Implement the `ConfigPlugin` interface from `internal/plugins/rpc`.
- Keep parsing/writing deterministic and avoid implicit network calls.
- Return actionable errors (path, key, and validation details).

## Example

See `plugins/ghostty-rpc/main.go`.
