package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// Logger provides structured logging with observability features
type Logger struct {
	logger  *slog.Logger
	level   slog.Level
	outputs []io.Writer
	hooks   []Hook
	context map[string]interface{}
	mu      sync.RWMutex
}

// Hook represents a logging hook function
type Hook func(ctx context.Context, record slog.Record, extra map[string]interface{})

// LogLevel represents log severity levels
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// NewLogger creates a new logger with default configuration
func NewLogger() *Logger {
	return &Logger{
		logger:  slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
		level:   slog.LevelInfo,
		outputs: []io.Writer{os.Stdout},
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}
}

// NewLoggerWithConfig creates a logger with custom configuration
func NewLoggerWithConfig(config *LogConfig) *Logger {
	level := slog.LevelInfo
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	}

	var outputs []io.Writer
	var handlers []slog.Handler

	// Add stdout/stderr
	if config.Console.Enabled {
		if config.Console.UseStderr {
			outputs = append(outputs, os.Stderr)
		} else {
			outputs = append(outputs, os.Stdout)
		}

		var handler slog.Handler
		if config.Console.Format == "text" {
			handler = slog.NewTextHandler(outputs[len(outputs)-1], &slog.HandlerOptions{
				Level: level,
			})
		} else {
			handler = slog.NewJSONHandler(outputs[len(outputs)-1], &slog.HandlerOptions{
				Level: level,
			})
		}
		handlers = append(handlers, handler)
	}

	// Add file output
	if config.File.Enabled && config.File.Path != "" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(config.File.Path), 0755); err == nil {
			if file, err := os.OpenFile(config.File.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				outputs = append(outputs, file)
				handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
					Level: level,
				})
				handlers = append(handlers, handler)
			}
		}
	}

	// Create multi-handler if multiple outputs
	var finalHandler slog.Handler
	if len(handlers) == 1 {
		finalHandler = handlers[0]
	} else if len(handlers) > 1 {
		finalHandler = &MultiHandler{handlers: handlers}
	} else {
		// Fallback to stdout JSON handler
		finalHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := &Logger{
		logger:  slog.New(finalHandler),
		level:   level,
		outputs: outputs,
		context: make(map[string]interface{}),
		hooks:   make([]Hook, 0),
	}

	// Add default context
	if config.DefaultContext != nil {
		for k, v := range config.DefaultContext {
			logger.context[k] = v
		}
	}

	return logger
}

// LogConfig represents logger configuration
type LogConfig struct {
	Level          LogLevel               `json:"level"`
	Console        ConsoleConfig          `json:"console"`
	File           FileConfig             `json:"file"`
	DefaultContext map[string]interface{} `json:"default_context,omitempty"`
	Hooks          []string               `json:"hooks,omitempty"`
}

// ConsoleConfig configures console output
type ConsoleConfig struct {
	Enabled   bool   `json:"enabled"`
	UseStderr bool   `json:"use_stderr"`
	Format    string `json:"format"` // "json" or "text"
}

// FileConfig configures file output
type FileConfig struct {
	Enabled  bool   `json:"enabled"`
	Path     string `json:"path"`
	MaxSize  int    `json:"max_size"`  // MB
	MaxAge   int    `json:"max_age"`   // days
	MaxFiles int    `json:"max_files"` // number of files to keep
}

// MultiHandler handles logging to multiple destinations
type MultiHandler struct {
	handlers []slog.Handler
}

func (mh *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (mh *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range mh.handlers {
		if h.Enabled(ctx, r.Level) {
			// Clone the record for each handler
			newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
			r.Attrs(func(a slog.Attr) bool {
				newRecord.AddAttrs(a)
				return true
			})
			if err := h.Handle(ctx, newRecord); err != nil {
				return err
			}
		}
	}
	return nil
}

func (mh *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (mh *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(mh.handlers))
	for i, h := range mh.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch level {
	case LevelDebug:
		l.level = slog.LevelDebug
	case LevelInfo:
		l.level = slog.LevelInfo
	case LevelWarn:
		l.level = slog.LevelWarn
	case LevelError:
		l.level = slog.LevelError
	}
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		logger:  l.logger,
		level:   l.level,
		outputs: l.outputs,
		hooks:   l.hooks,
		context: make(map[string]interface{}),
	}

	// Copy existing context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context
	newLogger.context[key] = value

	return newLogger
}

// AddHook adds a logging hook
func (l *Logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(slog.LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(slog.LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(slog.LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(slog.LevelError, msg, args...)
}

// LogError logs a ZeroUIError with full context
func (l *Logger) LogError(err error) {
	if ctErr, ok := errors.GetZeroUIError(err); ok {
		l.logZeroUIError(ctErr)
	} else {
		l.Error("Error occurred", "error", err.Error())
	}
}

// LogOperation logs an operation with timing
func (l *Logger) LogOperation(operation string, duration time.Duration, success bool, details map[string]interface{}) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelError
	}

	args := []interface{}{
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
		"success", success,
	}

	for k, v := range details {
		args = append(args, k, v)
	}

	l.log(level, fmt.Sprintf("Operation %s completed", operation), args...)
}

// LogConfigChange logs configuration changes
func (l *Logger) LogConfigChange(app, field string, oldValue, newValue interface{}, user string) {
	l.Info("Configuration changed",
		"app", app,
		"field", field,
		"old_value", oldValue,
		"new_value", newValue,
		"user", user,
		"timestamp", time.Now().UTC(),
	)
}

// LogBackup logs backup operations
func (l *Logger) LogBackup(operation string, app string, backupId string, success bool) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelError
	}

	l.log(level, fmt.Sprintf("Backup %s", operation),
		"operation", operation,
		"app", app,
		"backup_id", backupId,
		"success", success,
	)
}

// LogHook logs hook execution
func (l *Logger) LogHook(hookType string, command string, success bool, output string, duration time.Duration) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
	}

	l.log(level, "Hook executed",
		"hook_type", hookType,
		"command", command,
		"success", success,
		"output", output,
		"duration_ms", duration.Milliseconds(),
	)
}

// log is the internal logging method
func (l *Logger) log(level slog.Level, msg string, args ...interface{}) {
	// Respect the Logger's internal level threshold independent of handler configuration
	l.mu.RLock()
	currentLevel := l.level
	l.mu.RUnlock()
	if level < currentLevel {
		return
	}

	// Also honor handler capability to avoid unnecessary work
	if l.logger != nil && !l.logger.Enabled(context.Background(), level) {
		return
	}

	// Create record
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip log, Debug/Info/Warn/Error, and caller
	pc = pcs[0]

	record := slog.NewRecord(time.Now(), level, msg, pc)

	// Add context fields
	l.mu.RLock()
	for k, v := range l.context {
		record.AddAttrs(slog.Any(k, v))
	}
	hooks := l.hooks
	l.mu.RUnlock()

	// Add provided args
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				record.AddAttrs(slog.Any(key, args[i+1]))
			}
		}
	}

	// Execute hooks
	ctx := context.Background()
	extraData := make(map[string]interface{})
	l.mu.RLock()
	for k, v := range l.context {
		extraData[k] = v
	}
	l.mu.RUnlock()

	for _, hook := range hooks {
		hook(ctx, record, extraData)
	}

	// Log the record using Handler
	if l.logger.Handler() != nil {
		l.logger.Handler().Handle(ctx, record)
	}
}

// logZeroUIError logs a structured ZeroUIError
func (l *Logger) logZeroUIError(ctErr *errors.ZeroUIError) {
	args := []interface{}{
		"error_type", string(ctErr.Type),
		"error_message", ctErr.Message,
	}

	if ctErr.App != "" {
		args = append(args, "app", ctErr.App)
	}
	if ctErr.Field != "" {
		args = append(args, "field", ctErr.Field)
	}
	if ctErr.Value != "" {
		args = append(args, "value", ctErr.Value)
	}
	if ctErr.Path != "" {
		args = append(args, "file_path", ctErr.Path)
	}
	if ctErr.Line > 0 {
		args = append(args, "line", ctErr.Line)
	}
	if ctErr.Column > 0 {
		args = append(args, "column", ctErr.Column)
	}
	if ctErr.Cause != nil {
		args = append(args, "cause", ctErr.Cause.Error())
	}
	if len(ctErr.Suggestions) > 0 {
		args = append(args, "suggestions", ctErr.Suggestions)
	}
	if len(ctErr.Actions) > 0 {
		args = append(args, "actions", ctErr.Actions)
	}
	if len(ctErr.Context) > 0 {
		args = append(args, "context", ctErr.Context)
	}

	level := slog.LevelError
	switch ctErr.Severity {
	case errors.Info:
		level = slog.LevelInfo
	case errors.Warning:
		level = slog.LevelWarn
	case errors.Error:
		level = slog.LevelError
	case errors.Critical:
		level = slog.LevelError
	}

	l.log(level, "ZeroUI error occurred", args...)
}


// Global logger and metrics instances
var (
	globalLogger  *Logger
	globalMetrics *Metrics
	once          sync.Once
)

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	once.Do(func() {
		globalLogger = NewLogger()
		metrics, _ := NewMetrics(nil)
		globalMetrics = metrics
	})
	return globalLogger
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		globalLogger = NewLogger()
		metrics, _ := NewMetrics(nil)
		globalMetrics = metrics
	})
	return globalMetrics
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// LoadLogConfig loads logging configuration from file
func LoadLogConfig(configPath string) (*LogConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log config: %w", err)
	}

	var config LogConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse log config: %w", err)
	}

	return &config, nil
}

// CreateDefaultLogConfig creates a default logging configuration
func CreateDefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: LevelInfo,
		Console: ConsoleConfig{
			Enabled:   true,
			UseStderr: false,
			Format:    "json",
		},
		File: FileConfig{
			Enabled:  false,
			Path:     "",
			MaxSize:  100,
			MaxAge:   30,
			MaxFiles: 5,
		},
		DefaultContext: map[string]interface{}{
			"service": "configtoggle",
			"version": "1.0.0",
		},
	}
}

// Common hooks

// MetricsHook creates a hook that records metrics
func MetricsHook(metrics *Metrics) Hook {
	return func(ctx context.Context, record slog.Record, extra map[string]interface{}) {
		// Record errors if it's an error level
		if record.Level >= slog.LevelError {
			operation := "unknown"
			if op, ok := extra["operation"].(string); ok {
				operation = op
			}
			metrics.RecordError(ctx, operation, "log_error")
		}

		// Record operation timers if present
		if _, ok := extra["operation"].(string); ok {
			// Use the metrics API that exists
			// For now, we'll skip timer recording since the Metrics type
			// has specific operation methods rather than generic ones
		}
	}
}

// AuditHook creates a hook for audit logging
func AuditHook(auditFile string) Hook {
	return func(ctx context.Context, record slog.Record, extra map[string]interface{}) {
		// Collect attributes from the record and merge into extra so hooks can discover operation details
		attrMap := make(map[string]interface{})
		record.Attrs(func(a slog.Attr) bool {
			attrMap[a.Key] = a.Value.Any()
			return true
		})

		// Merge record attributes into a new map so we don't mutate the original 'extra'
		merged := make(map[string]interface{}, len(extra)+len(attrMap))
		for k, v := range extra {
			merged[k] = v
		}
		for k, v := range attrMap {
			merged[k] = v
		}

		// Determine operation from merged attributes
		operation, _ := merged["operation"].(string)
		if operation == "" {
			// Try to infer from message like "Operation toggle completed"
			msg := strings.ToLower(record.Message)
			switch {
			case strings.Contains(msg, "toggle"):
				operation = "toggle"
			case strings.Contains(msg, "preset"):
				operation = "preset"
			case strings.Contains(msg, "backup"):
				operation = "backup"
			}
		}

		if operation == "toggle" || operation == "preset" || operation == "backup" {
			auditData := map[string]interface{}{
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"level":     record.Level.String(),
				"message":   record.Message,
				"operation": operation,
			}

			// Add relevant fields from merged attributes (record attrs + context)
			for key, value := range merged {
				switch key {
				case "app", "field", "old_value", "new_value", "user", "backup_id", "duration_ms", "success":
					auditData[key] = value
				}
			}

			// Write to audit file
			data, _ := json.Marshal(auditData)
			data = append(data, '\n')

			// Ensure audit directory exists
			os.MkdirAll(filepath.Dir(auditFile), 0755)

			// Append to audit file
			if file, err := os.OpenFile(auditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				file.Write(data)
				file.Close()
			}
		}
	}
}
