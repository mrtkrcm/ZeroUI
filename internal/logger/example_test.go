package logger_test

import (
	"context"
	"errors"
	"os"

	"github.com/mrtkrcm/ZeroUI/internal/logger"
)

// Example demonstrates the basic usage of the logger with structured fields
func Example_basicUsage() {
	// Create a new logger
	log := logger.New(&logger.Config{
		Level:  "info",
		Format: "console",
		Output: os.Stdout,
	})

	// Get the LoggerInterface adapter
	structuredLogger := log.With()

	// Log with structured fields
	structuredLogger.Info("Processing request",
		logger.Field{Key: "user_id", Value: "123"},
		logger.Field{Key: "action", Value: "login"},
	)

	// Log errors with context
	err := errors.New("database connection failed")
	structuredLogger.Error("Failed to connect", err,
		logger.Field{Key: "retry_count", Value: 3},
		logger.Field{Key: "database", Value: "postgres"},
	)
}

// Example demonstrates creating child loggers with additional context
func Example_childLogger() {
	log := logger.New(logger.DefaultConfig())

	// Create a logger with service context
	serviceLogger := log.With(
		logger.Field{Key: "service", Value: "api"},
		logger.Field{Key: "version", Value: "1.0"},
	)

	// Add request-specific context
	requestLogger := serviceLogger.WithRequest("req-123-456")

	// All logs from this logger will include the service, version, and request_id
	requestLogger.Info("Handling API request")
}

// Example demonstrates using logger with context
func Example_contextLogger() {
	// Create a logger
	log := logger.New(logger.DefaultConfig())
	structuredLogger := log.With(
		logger.Field{Key: "component", Value: "worker"},
	)

	// Add logger to context
	ctx := logger.ContextWithLogger(context.Background(), structuredLogger)

	// Later, retrieve the logger from context
	loggerFromCtx := logger.FromContext(ctx)
	loggerFromCtx.Info("Processing task",
		logger.Field{Key: "task_id", Value: "task-789"},
	)
}

// Example demonstrates backward compatibility with map-based fields
func Example_backwardCompatibility() {
	log := logger.New(logger.DefaultConfig())

	// Old style with maps (still supported)
	log.Info("Old style logging", map[string]interface{}{
		"key": "value",
	})

	log.Error("Old style error", errors.New("something went wrong"), map[string]interface{}{
		"retry": true,
	})

	// With app context (returns *Logger)
	appLogger := log.WithApp("myapp")
	appLogger.Info("App-specific message")
}

// Example demonstrates chaining logger context
func Example_chainingContext() {
	log := logger.New(logger.DefaultConfig())

	// Chain multiple context additions
	chainedLogger := log.
		With(logger.Field{Key: "service", Value: "api"}).
		WithRequest("req-123").
		With(logger.Field{Key: "user", Value: "john"})

	// All fields are included in the log
	chainedLogger.Info("Request processed successfully",
		logger.Field{Key: "duration_ms", Value: 42},
	)
}
