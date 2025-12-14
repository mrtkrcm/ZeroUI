# Logger Package

A structured logging package built on top of [zerolog](https://github.com/rs/zerolog) that provides both a modern interface with structured fields and backward compatibility with existing map-based logging.

## Features

- **Structured Logging**: Type-safe field definitions with the `Field` struct
- **Context Support**: Store and retrieve loggers from `context.Context`
- **Request Tracing**: Built-in support for request ID tracking
- **Backward Compatible**: Existing code using map-based fields continues to work
- **Chainable API**: Fluent interface for building contextual loggers
- **JSON & Console Output**: Configurable output formats

## Quick Start

### Basic Usage

```go
import "github.com/mrtkrcm/ZeroUI/internal/logger"

// Create a logger
log := logger.New(logger.DefaultConfig())

// Get the structured interface
structuredLog := log.With()

// Log with structured fields
structuredLog.Info("User logged in",
    logger.Field{Key: "user_id", Value: "123"},
    logger.Field{Key: "ip", Value: "192.168.1.1"},
)

// Log errors
err := errors.New("connection timeout")
structuredLog.Error("Database error", err,
    logger.Field{Key: "retry_count", Value: 3},
)
```

### Creating Child Loggers

```go
// Add persistent context to logger
serviceLogger := log.With(
    logger.Field{Key: "service", Value: "api"},
    logger.Field{Key: "version", Value: "1.0"},
)

// All subsequent logs include the service and version
serviceLogger.Info("Service started")
```

### Request Tracing

```go
// Add request ID to logger
requestLogger := log.With().WithRequest("req-abc-123")

// All logs will include request_id field
requestLogger.Info("Processing request")
requestLogger.Info("Request completed")
```

### Context Integration

```go
import "context"

// Store logger in context
ctx := logger.ContextWithLogger(context.Background(), structuredLog)

// Later, retrieve logger from context
loggerFromCtx := logger.FromContext(ctx)
loggerFromCtx.Info("Using logger from context")
```

### Chaining Context

```go
// Chain multiple context additions
chainedLogger := log.
    With(logger.Field{Key: "service", Value: "api"}).
    WithRequest("req-123").
    With(logger.Field{Key: "user", Value: "john"})

chainedLogger.Info("Request processed")
// Output includes: service, request_id, and user fields
```

## Configuration

```go
cfg := &logger.Config{
    Level:      "debug",        // debug, info, warn, error
    Format:     "json",         // json or console
    Output:     os.Stdout,      // any io.Writer
    TimeFormat: time.RFC3339,   // timestamp format
}

log := logger.New(cfg)
```

### Default Configuration

```go
log := logger.New(logger.DefaultConfig())
// Level: info
// Format: console
// Output: os.Stdout
// TimeFormat: RFC3339
```

## API Reference

### LoggerInterface

The primary interface for structured logging:

```go
type LoggerInterface interface {
    Info(msg string, fields ...Field)
    Error(msg string, err error, fields ...Field)
    With(fields ...Field) LoggerInterface
    WithRequest(requestID string) LoggerInterface
}
```

### Field Struct

```go
type Field struct {
    Key   string
    Value interface{}
}
```

### Context Functions

```go
// Store logger in context
func ContextWithLogger(ctx context.Context, l LoggerInterface) context.Context

// Retrieve logger from context (returns global logger if not found)
func FromContext(ctx context.Context) LoggerInterface
```

## Backward Compatibility

Existing code using `*logger.Logger` with map-based fields continues to work:

```go
log := logger.New(logger.DefaultConfig())

// Old style (still supported)
log.Info("Old message", map[string]interface{}{
    "key": "value",
})

log.Error("Error message", err, map[string]interface{}{
    "retry": true,
})

// App and field context helpers
appLogger := log.WithApp("myapp")
fieldLogger := log.WithField("myfield")
```

## Global Logger

```go
// Use the global logger instance
logger.InitGlobal(&logger.Config{Level: "debug"})

// Access global logger
global := logger.Global()

// Convenience functions
logger.Info("message", map[string]interface{}{"key": "value"})
logger.Error("error", err, map[string]interface{}{"key": "value"})
```

## Examples

See `example_test.go` for comprehensive examples.

## Testing

```bash
go test ./internal/logger/... -v
```

## Migration Guide

### From Map-Based to Structured Fields

Before:
```go
log.Info("User action", map[string]interface{}{
    "user_id": "123",
    "action": "login",
})
```

After:
```go
structuredLog := log.With()
structuredLog.Info("User action",
    logger.Field{Key: "user_id", Value: "123"},
    logger.Field{Key: "action", Value: "login"},
)
```

Both patterns are supported, so migration can be gradual.
