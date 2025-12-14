package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool // whether logger should be non-nil
	}{
		{
			name:   "with nil config",
			config: nil,
			want:   true,
		},
		{
			name: "with valid config",
			config: &Config{
				Level:  "debug",
				Format: "json",
			},
			want: true,
		},
		{
			name: "with console format",
			config: &Config{
				Level:  "info",
				Format: "console",
			},
			want: true,
		},
		{
			name: "with invalid level defaults to info",
			config: &Config{
				Level:  "invalid",
				Format: "json",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.config)
			if (got != nil) != tt.want {
				t.Errorf("New() = %v, want non-nil = %v", got, tt.want)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		fields  []Field
		wantMsg string
		wantKey string
	}{
		{
			name:    "simple message",
			msg:     "test message",
			fields:  nil,
			wantMsg: "test message",
		},
		{
			name: "with single field",
			msg:  "user action",
			fields: []Field{
				{Key: "user_id", Value: "123"},
			},
			wantMsg: "user action",
			wantKey: "user_id",
		},
		{
			name: "with multiple fields",
			msg:  "operation complete",
			fields: []Field{
				{Key: "operation", Value: "toggle"},
				{Key: "duration_ms", Value: 42},
			},
			wantMsg: "operation complete",
			wantKey: "operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &Config{
				Level:  "info",
				Format: "json",
				Output: &buf,
			}
			logger := New(cfg)

			// Use the adapter to test LoggerInterface methods
			adapter := logger.With() // Returns LoggerInterface
			adapter.Info(tt.msg, tt.fields...)

			output := buf.String()
			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("Info() output = %q, want to contain %q", output, tt.wantMsg)
			}

			if tt.wantKey != "" && !strings.Contains(output, tt.wantKey) {
				t.Errorf("Info() output = %q, want to contain key %q", output, tt.wantKey)
			}
		})
	}
}

func TestLogger_Error(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		err     error
		fields  []Field
		wantMsg string
		wantErr string
		wantKey string
	}{
		{
			name:    "error without fields",
			msg:     "operation failed",
			err:     errors.New("test error"),
			fields:  nil,
			wantMsg: "operation failed",
			wantErr: "test error",
		},
		{
			name: "error with fields",
			msg:  "database error",
			err:  errors.New("connection timeout"),
			fields: []Field{
				{Key: "database", Value: "postgres"},
				{Key: "retry_count", Value: 3},
			},
			wantMsg: "database error",
			wantErr: "connection timeout",
			wantKey: "database",
		},
		{
			name:    "nil error",
			msg:     "warning message",
			err:     nil,
			fields:  nil,
			wantMsg: "warning message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &Config{
				Level:  "error",
				Format: "json",
				Output: &buf,
			}
			logger := New(cfg)

			// Use the adapter to test LoggerInterface methods
			adapter := logger.With() // Returns LoggerInterface
			adapter.Error(tt.msg, tt.err, tt.fields...)

			output := buf.String()
			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("Error() output = %q, want to contain %q", output, tt.wantMsg)
			}

			if tt.wantErr != "" && !strings.Contains(output, tt.wantErr) {
				t.Errorf("Error() output = %q, want to contain error %q", output, tt.wantErr)
			}

			if tt.wantKey != "" && !strings.Contains(output, tt.wantKey) {
				t.Errorf("Error() output = %q, want to contain key %q", output, tt.wantKey)
			}
		})
	}
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	// Create a child logger with additional context
	childLogger := logger.With(
		Field{Key: "service", Value: "api"},
		Field{Key: "version", Value: "1.0"},
	)

	// Log with the child logger
	childLogger.Info("request processed")

	output := buf.String()

	// Verify the message is present
	if !strings.Contains(output, "request processed") {
		t.Errorf("With() output = %q, want to contain message", output)
	}

	// Verify the fields are present
	if !strings.Contains(output, "service") || !strings.Contains(output, "api") {
		t.Errorf("With() output = %q, want to contain service field", output)
	}

	if !strings.Contains(output, "version") || !strings.Contains(output, "1.0") {
		t.Errorf("With() output = %q, want to contain version field", output)
	}
}

func TestLogger_WithRequest(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	requestID := "req-123-456"
	requestLogger := logger.WithRequest(requestID)

	requestLogger.Info("handling request")

	output := buf.String()

	// Verify the request_id field is present
	if !strings.Contains(output, "request_id") {
		t.Errorf("WithRequest() output = %q, want to contain request_id field", output)
	}

	if !strings.Contains(output, requestID) {
		t.Errorf("WithRequest() output = %q, want to contain request ID %q", output, requestID)
	}
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		shouldHaveData bool // whether result should have specific context data
	}{
		{
			name:           "nil context returns global adapter",
			ctx:            nil,
			shouldHaveData: false,
		},
		{
			name:           "empty context returns global adapter",
			ctx:            context.Background(),
			shouldHaveData: false,
		},
		{
			name: "context with logger",
			ctx: func() context.Context {
				logger := New(DefaultConfig())
				adapter := logger.With(Field{Key: "test_ctx", Value: "value"})
				return ContextWithLogger(context.Background(), adapter)
			}(),
			shouldHaveData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromContext(tt.ctx)
			if got == nil {
				t.Error("FromContext() returned nil")
			}

			// Verify it implements LoggerInterface
			var _ LoggerInterface = got
		})
	}
}

func TestContextWithLogger(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "nil context creates background",
			ctx:  nil,
		},
		{
			name: "existing context",
			ctx:  context.Background(),
		},
		{
			name: "context with values",
			ctx:  context.WithValue(context.Background(), "key", "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(DefaultConfig())
			adapter := logger.With() // Get LoggerInterface
			ctx := ContextWithLogger(tt.ctx, adapter)

			if ctx == nil {
				t.Error("ContextWithLogger() returned nil context")
			}

			// Retrieve the logger from context
			retrieved := FromContext(ctx)
			if retrieved == nil {
				t.Error("FromContext() returned nil after ContextWithLogger()")
			}

			// Verify it implements LoggerInterface
			var _ LoggerInterface = retrieved
		})
	}
}

func TestLogger_BackwardCompatibility(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	// Test that old methods still work with map-based fields
	t.Run("Info with maps", func(t *testing.T) {
		buf.Reset()
		logger.Info("old style info", map[string]interface{}{
			"key": "value",
		})
		output := buf.String()
		if !strings.Contains(output, "old style info") {
			t.Errorf("Info() output = %q, want to contain message", output)
		}
	})

	t.Run("Error with maps", func(t *testing.T) {
		buf.Reset()
		logger.Error("old style error", errors.New("test"), map[string]interface{}{
			"key": "value",
		})
		output := buf.String()
		if !strings.Contains(output, "old style error") {
			t.Errorf("Error() output = %q, want to contain message", output)
		}
	})

	t.Run("WithApp", func(t *testing.T) {
		buf.Reset()
		appLogger := logger.WithApp("myapp")
		appLogger.Info("app message")
		output := buf.String()
		if !strings.Contains(output, "myapp") {
			t.Errorf("WithApp() output = %q, want to contain app name", output)
		}
	})

	t.Run("WithField", func(t *testing.T) {
		buf.Reset()
		fieldLogger := logger.WithField("myfield")
		fieldLogger.Info("field message")
		output := buf.String()
		if !strings.Contains(output, "myfield") {
			t.Errorf("WithField() output = %q, want to contain field name", output)
		}
	})
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	adapter := logger.With() // Get LoggerInterface
	adapter.Info("test message", Field{Key: "key1", Value: "value1"})

	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	// Verify the message
	if msg, ok := logEntry["message"].(string); !ok || msg != "test message" {
		t.Errorf("JSON output message = %v, want %q", logEntry["message"], "test message")
	}

	// Verify the custom field
	if val, ok := logEntry["key1"].(string); !ok || val != "value1" {
		t.Errorf("JSON output key1 = %v, want %q", logEntry["key1"], "value1")
	}

	// Verify log level
	if level, ok := logEntry["level"].(string); !ok || level != "info" {
		t.Errorf("JSON output level = %v, want %q", logEntry["level"], "info")
	}
}

func TestLogger_ChainedContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	// Chain multiple context additions
	chainedLogger := logger.
		With(Field{Key: "service", Value: "api"}).
		WithRequest("req-123").
		With(Field{Key: "user", Value: "john"})

	chainedLogger.Info("chained context test")

	output := buf.String()

	// Verify all contexts are present
	expectedFields := []string{"service", "api", "request_id", "req-123", "user", "john"}
	for _, expected := range expectedFields {
		if !strings.Contains(output, expected) {
			t.Errorf("Chained context output = %q, want to contain %q", output, expected)
		}
	}
}

func TestGlobal(t *testing.T) {
	// Get global logger
	logger1 := Global()
	if logger1 == nil {
		t.Fatal("Global() returned nil")
	}

	// Get it again, should be the same instance
	logger2 := Global()
	if logger1 != logger2 {
		t.Error("Global() should return the same instance")
	}

	// Initialize global with custom config
	var buf bytes.Buffer
	InitGlobal(&Config{
		Level:  "debug",
		Format: "json",
		Output: &buf,
	})

	// Get global logger again
	logger3 := Global()
	if logger3 == nil {
		t.Fatal("Global() returned nil after InitGlobal()")
	}

	// Test that it uses the new config
	logger3.Info("test")
	if buf.Len() == 0 {
		t.Error("Global logger did not write to custom output")
	}
}

func TestLogger_InterfaceCompliance(t *testing.T) {
	// Verify that loggerAdapter implements LoggerInterface
	var _ LoggerInterface = (*loggerAdapter)(nil)

	logger := New(DefaultConfig())
	adapter := logger.With() // Returns LoggerInterface

	// Test that all interface methods are callable
	t.Run("Info", func(t *testing.T) {
		adapter.Info("test")
	})

	t.Run("Error", func(t *testing.T) {
		adapter.Error("test", errors.New("error"))
	})

	t.Run("With", func(t *testing.T) {
		childLogger := adapter.With(Field{Key: "test", Value: "value"})
		if childLogger == nil {
			t.Error("With() returned nil")
		}
	})

	t.Run("WithRequest", func(t *testing.T) {
		reqLogger := adapter.WithRequest("req-123")
		if reqLogger == nil {
			t.Error("WithRequest() returned nil")
		}
	})
}

func TestLogger_FieldTypes(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}
	logger := New(cfg)

	// Test various field value types using the adapter
	adapter := logger.With() // Get LoggerInterface
	adapter.Info("test various types",
		Field{Key: "string", Value: "text"},
		Field{Key: "int", Value: 42},
		Field{Key: "float", Value: 3.14},
		Field{Key: "bool", Value: true},
		Field{Key: "slice", Value: []string{"a", "b"}},
		Field{Key: "map", Value: map[string]int{"x": 1}},
	)

	output := buf.String()

	// Verify all field keys are present
	expectedKeys := []string{"string", "int", "float", "bool", "slice", "map"}
	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("Output = %q, want to contain key %q", output, key)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.Level != "info" {
		t.Errorf("DefaultConfig().Level = %q, want %q", cfg.Level, "info")
	}

	if cfg.Format != "console" {
		t.Errorf("DefaultConfig().Format = %q, want %q", cfg.Format, "console")
	}

	if cfg.Output == nil {
		t.Error("DefaultConfig().Output is nil")
	}
}
