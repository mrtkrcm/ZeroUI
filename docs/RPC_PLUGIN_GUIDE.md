# RPC Plugin Development Guide

## Quick Start

1. **Create plugin directory**:
   ```bash
   mkdir plugins/my-plugin
   cd plugins/my-plugin
   go mod init github.com/mrtkrcm/ZeroUI/plugins/my-plugin
   ```

2. **Create main.go**:
   ```go
   package main

   import (
       "context"
       "log"
       "os"

       "github.com/hashicorp/go-plugin"
       "github.com/mrtkrcm/ZeroUI/internal/plugins/rpc"
   )

   type MyPlugin struct {
       logger *log.Logger
   }

   func (p *MyPlugin) GetInfo(ctx context.Context) (*rpc.PluginInfo, error) {
       return &rpc.PluginInfo{
           Name:        "my-plugin",
           Version:     "1.0.0", 
           Description: "My awesome plugin",
           ApiVersion:  rpc.CurrentAPIVersion,
           Capabilities: []string{
               rpc.CapabilityConfigParsing,
               rpc.CapabilityConfigWriting,
           },
       }, nil
   }

   // Implement other ConfigPlugin methods...

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

3. **Build**: 
   ```bash
   go build -o zeroui-plugin-my-plugin
   ```

4. **Deploy**: Place executable in plugin directory

## Interface Implementation

### Required Methods

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

### Data Types

- **PluginInfo**: Basic plugin metadata
- **ConfigInfo**: Config file detection results
- **ConfigData**: Parsed configuration with protobuf fields
- **ConfigMetadata**: Schema definition with field types and presets

## Type Conversion

Use helper functions for protobuf compatibility:

```go
// Convert Go values to protobuf Any
anyValue, err := convertInterfaceToAny(value)

// Convert protobuf Any back to Go values  
value, err := convertAnyToInterface(anyValue)
```

## Plugin Naming

**Critical**: Plugins must be named `zeroui-plugin-{name}` for discovery.

Examples:
- `zeroui-plugin-ghostty-rpc`
- `zeroui-plugin-vscode`
- `zeroui-plugin-tmux`

## Testing

```go
func TestMyPlugin(t *testing.T) {
    p := &MyPlugin{}
    ctx := context.Background()
    
    info, err := p.GetInfo(ctx)
    assert.NoError(t, err)
    assert.Equal(t, "my-plugin", info.Name)
}
```

## Example Plugin

See `plugins/ghostty-rpc/main.go` for a complete implementation with:
- Config file parsing
- Field validation
- Preset management  
- Error handling
- Comprehensive tests

The Ghostty RPC plugin provides identical functionality to the legacy version while running in a separate process for better isolation and reliability.