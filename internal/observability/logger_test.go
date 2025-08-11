package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// setupLoggerTest creates a test environment for logging
func setupLoggerTest(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "logger-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// TestNewLogger tests creating a new logger
func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	if logger.logger == nil {
		t.Error("Expected internal logger to be initialized")
	}

	if logger.context == nil {
		t.Error("Expected context map to be initialized")
	}

	if logger.hooks == nil {
		t.Error("Expected hooks slice to be initialized")
	}
}

// TestLoggerWithConfig tests creating logger with configuration
func TestLoggerWithConfig(t *testing.T) {
	tmpDir, cleanup := setupLoggerTest(t)
	defer cleanup()

	logFile := filepath.Join(tmpDir, "test.log")

	config := &LogConfig{
		Level: LevelDebug,
		Console: ConsoleConfig{
			Enabled:   true,
			UseStderr: false,
			Format:    "json",
		},
		File: FileConfig{
			Enabled:  true,
			Path:     logFile,
			MaxSize:  100,
			MaxAge:   30,
			MaxFiles: 5,
		},
		DefaultContext: map[string]interface{}{
			"service": "test",
			"version": "1.0.0",
		},
	}

	logger := NewLoggerWithConfig(config)

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	if logger.level != slog.LevelDebug {
		t.Errorf("Expected debug level, got %v", logger.level)
	}

	// Test logging to file
	logger.Info("Test message", "key", "value")

	// Give it a moment to write
	time.Sleep(10 * time.Millisecond)

	// Check if file was created and contains log
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Test message") {
		t.Error("Expected log file to contain test message")
	}

	if !strings.Contains(string(content), "service") {
		t.Error("Expected log file to contain default context")
	}
}

// TestLoggerLevels tests different log levels
func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer

	// Create logger that writes to buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Test all levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	output := buf.String()

	if !strings.Contains(output, "Debug message") {
		t.Error("Expected debug message in output")
	}
	if !strings.Contains(output, "Info message") {
		t.Error("Expected info message in output")
	}
	if !strings.Contains(output, "Warning message") {
		t.Error("Expected warning message in output")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Expected error message in output")
	}

	// Test level filtering
	buf.Reset()
	logger.SetLevel(LevelError)

	logger.Debug("Should not appear")
	logger.Info("Should not appear")
	logger.Warn("Should not appear")
	logger.Error("Should appear")

	output = buf.String()

	if strings.Contains(output, "Should not appear") {
		t.Error("Expected filtered messages not to appear")
	}
	if !strings.Contains(output, "Should appear") {
		t.Error("Expected error message to appear")
	}
}

// TestLoggerWithContext tests context functionality
func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Add context
	contextLogger := logger.WithContext("request_id", "12345").WithContext("user", "testuser")

	contextLogger.Info("Test with context")

	output := buf.String()

	if !strings.Contains(output, "request_id") {
		t.Error("Expected request_id in output")
	}
	if !strings.Contains(output, "12345") {
		t.Error("Expected request_id value in output")
	}
	if !strings.Contains(output, "user") {
		t.Error("Expected user in output")
	}
	if !strings.Contains(output, "testuser") {
		t.Error("Expected user value in output")
	}
}

// TestLoggerHooks tests hook functionality
func TestLoggerHooks(t *testing.T) {
	var buf bytes.Buffer
	var hookCalls int

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Add test hook
	logger.AddHook(func(ctx context.Context, record slog.Record, extra map[string]interface{}) {
		hookCalls++
	})

	logger.Info("Test message 1")
	logger.Error("Test message 2")

	if hookCalls != 2 {
		t.Errorf("Expected 2 hook calls, got %d", hookCalls)
	}
}

// TestLogError tests structured error logging
func TestLogError(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Test with ZeroUIError
	ctErr := errors.New(errors.AppNotFound, "App not found").
		WithApp("test-app").
		WithField("test-field").
		WithValue("test-value").
		WithSeverity(errors.Error).
		WithSuggestions("Try checking the app name").
		WithActions("Run 'list apps' command")

	logger.LogError(ctErr)

	output := buf.String()

	if !strings.Contains(output, "APP_NOT_FOUND") {
		t.Error("Expected error type in output")
	}
	if !strings.Contains(output, "test-app") {
		t.Error("Expected app name in output")
	}
	if !strings.Contains(output, "test-field") {
		t.Error("Expected field name in output")
	}
	if !strings.Contains(output, "Try checking the app name") {
		t.Error("Expected suggestions in output")
	}

	// Test with regular error
	buf.Reset()
	regularErr := fmt.Errorf("regular error")
	logger.LogError(regularErr)

	output = buf.String()
	if !strings.Contains(output, "regular error") {
		t.Error("Expected regular error message in output")
	}
}

// TestLogOperation tests operation logging
func TestLogOperation(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Test successful operation
	logger.LogOperation("toggle", 150*time.Millisecond, true, map[string]interface{}{
		"app":   "test-app",
		"field": "theme",
	})

	output := buf.String()

	if !strings.Contains(output, "toggle") {
		t.Error("Expected operation name in output")
	}
	if !strings.Contains(output, "150") {
		t.Error("Expected duration in output")
	}
	if !strings.Contains(output, "true") {
		t.Error("Expected success flag in output")
	}
	if !strings.Contains(output, "test-app") {
		t.Error("Expected app name in output")
	}

	// Test failed operation
	buf.Reset()
	logger.LogOperation("preset", 50*time.Millisecond, false, map[string]interface{}{
		"error": "preset not found",
	})

	output = buf.String()

	if !strings.Contains(output, "false") {
		t.Error("Expected failure flag in output")
	}
	if !strings.Contains(output, "preset not found") {
		t.Error("Expected error details in output")
	}
}

// TestLogConfigChange tests configuration change logging
func TestLogConfigChange(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	logger.LogConfigChange("myapp", "theme", "dark", "light", "testuser")

	output := buf.String()

	if !strings.Contains(output, "Configuration changed") {
		t.Error("Expected config change message in output")
	}
	if !strings.Contains(output, "myapp") {
		t.Error("Expected app name in output")
	}
	if !strings.Contains(output, "theme") {
		t.Error("Expected field name in output")
	}
	if !strings.Contains(output, "dark") {
		t.Error("Expected old value in output")
	}
	if !strings.Contains(output, "light") {
		t.Error("Expected new value in output")
	}
	if !strings.Contains(output, "testuser") {
		t.Error("Expected user in output")
	}
}

// TestLogHook tests hook execution logging
func TestLogHook(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	logger.LogHook("post-toggle", "echo 'test'", true, "test", 25*time.Millisecond)

	output := buf.String()

	if !strings.Contains(output, "Hook executed") {
		t.Error("Expected hook execution message in output")
	}
	if !strings.Contains(output, "post-toggle") {
		t.Error("Expected hook type in output")
	}
	if !strings.Contains(output, "echo 'test'") {
		t.Error("Expected command in output")
	}
	if !strings.Contains(output, "25") {
		t.Error("Expected duration in output")
	}
}

// TestMetrics tests metrics collection
func TestMetrics(t *testing.T) {
	metrics, err := NewMetrics(nil)
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	ctx := context.Background()

	// Test error recording
	metrics.RecordError(ctx, "test_operation", "test_error")

	// Test operation recording
	metrics.RecordToggleOperation(ctx, "test_app", "test_key", true, 100*time.Millisecond)
	metrics.RecordCycleOperation(ctx, "test_app", "test_key", true, 200*time.Millisecond)

	// Metrics verification would require exposing internal state
	// The actual metrics are recorded via OpenTelemetry
	// Just ensure operations don't panic
}

// TestMetricsHook tests metrics hook
func TestMetricsHook(t *testing.T) {
	var buf bytes.Buffer
	metrics, err := NewMetrics(nil)
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Add metrics hook
	logger.AddHook(MetricsHook(metrics))

	// Log some messages
	logger.Info("Info message")
	logger.Error("Error message")
	logger.LogOperation("toggle", 100*time.Millisecond, true, map[string]interface{}{
		"app": "test",
	})

	// Just ensure the hook doesn't panic
	// Actual metrics verification would require exposing internal state
	// The metrics are recorded via OpenTelemetry
}

// TestAuditHook tests audit hook
func TestAuditHook(t *testing.T) {
	tmpDir, cleanup := setupLoggerTest(t)
	defer cleanup()

	auditFile := filepath.Join(tmpDir, "audit.log")

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := &Logger{
		logger:  slog.New(handler),
		level:   slog.LevelDebug,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Add audit hook
	logger.AddHook(AuditHook(auditFile))

	// Log an operation that should be audited
	logger.LogOperation("toggle", 100*time.Millisecond, true, map[string]interface{}{
		"app":       "test-app",
		"field":     "theme",
		"old_value": "dark",
		"new_value": "light",
		"user":      "testuser",
	})

	// Give it time to write
	time.Sleep(10 * time.Millisecond)

	// Check audit file
	if _, err := os.Stat(auditFile); os.IsNotExist(err) {
		t.Error("Expected audit file to be created")
	}

	content, err := os.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	if !strings.Contains(string(content), "toggle") {
		t.Error("Expected toggle operation in audit file")
	}
	if !strings.Contains(string(content), "test-app") {
		t.Error("Expected app name in audit file")
	}
	if !strings.Contains(string(content), "testuser") {
		t.Error("Expected user in audit file")
	}

	// Log a non-audited operation
	logger.Info("Regular log message")

	// Audit file should not grow
	newContent, err := os.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("Failed to read audit file: %v", err)
	}

	if len(newContent) != len(content) {
		t.Error("Expected audit file size to remain the same for non-audited logs")
	}
}

// TestLoadLogConfig tests loading configuration from file
func TestLoadLogConfig(t *testing.T) {
	tmpDir, cleanup := setupLoggerTest(t)
	defer cleanup()

	// Create config file
	configPath := filepath.Join(tmpDir, "log-config.json")
	config := CreateDefaultLogConfig()
	config.Level = LevelDebug
	config.File.Enabled = true
	config.File.Path = "/tmp/test.log"

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	loadedConfig, err := LoadLogConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.Level != LevelDebug {
		t.Errorf("Expected debug level, got %v", loadedConfig.Level)
	}

	if !loadedConfig.File.Enabled {
		t.Error("Expected file logging to be enabled")
	}

	if loadedConfig.File.Path != "/tmp/test.log" {
		t.Errorf("Expected file path '/tmp/test.log', got '%s'", loadedConfig.File.Path)
	}

	// Test loading non-existent file
	_, err = LoadLogConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}

	// Test loading invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err = LoadLogConfig(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON config")
	}
}

// TestGlobalInstances tests global logger and metrics
func TestGlobalInstances(t *testing.T) {
	// Reset global state
	globalLogger = nil
	globalMetrics = nil
	once = sync.Once{}

	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("Expected same logger instance from multiple GetLogger calls")
	}

	metrics1 := GetMetrics()
	metrics2 := GetMetrics()

	if metrics1 != metrics2 {
		t.Error("Expected same metrics instance from multiple GetMetrics calls")
	}

	// Test setting global logger
	customLogger := NewLogger()
	SetGlobalLogger(customLogger)

	if globalLogger != customLogger {
		t.Error("Expected global logger to be set to custom logger")
	}
}

// TestMultiHandler tests the multi-handler functionality
func TestMultiHandler(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	handler1 := slog.NewJSONHandler(&buf1, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler2 := slog.NewJSONHandler(&buf2, &slog.HandlerOptions{Level: slog.LevelDebug})

	multiHandler := &MultiHandler{handlers: []slog.Handler{handler1, handler2}}
	logger := slog.New(multiHandler)

	logger.Info("Test message", "key", "value")
	logger.Debug("Debug message", "debug", true)

	output1 := buf1.String()
	output2 := buf2.String()

	// Both handlers should receive the info message
	if !strings.Contains(output1, "Test message") {
		t.Error("Expected info message in first handler output")
	}
	if !strings.Contains(output2, "Test message") {
		t.Error("Expected info message in second handler output")
	}

	// Only the debug-level handler should receive the debug message
	if strings.Contains(output1, "Debug message") {
		t.Error("First handler (info level) should not receive debug message")
	}
	if !strings.Contains(output2, "Debug message") {
		t.Error("Second handler (debug level) should receive debug message")
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
