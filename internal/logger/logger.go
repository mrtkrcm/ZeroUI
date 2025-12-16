package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const loggerContextKey contextKey = "logger"

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// LoggerInterface defines the contract for structured logging
type LoggerInterface interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	With(fields ...Field) LoggerInterface
	WithRequest(requestID string) LoggerInterface
}

// Logger provides structured logging functionality
// It implements the LoggerInterface
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string // json, console
	Output     io.Writer
	TimeFormat string
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "console",
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
	}
}

// New creates a new logger with the given configuration
func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	// Set global log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if config.Format == "console" {
		output := zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: config.TimeFormat,
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(config.Output).With().Timestamp().Logger()
	}

	return &Logger{
		logger: logger,
	}
}

// WithContext adds contextual fields to the logger
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	return &Logger{
		logger: ctx.Logger(),
	}
}

// WithApp adds app context to the logger
func (l *Logger) WithApp(app string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("app", app).Logger(),
	}
}

// WithField adds a field context to the logger
func (l *Logger) WithField(field string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("field", field).Logger(),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	event := l.logger.Debug()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Warn logs a warning message with structured Field types
func (l *Logger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	l.addStructuredFields(event, fields...)
	event.Msg(msg)
}

// Info logs an info message with map-based fields (backward compatibility)
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// InfoStructured logs an info message with structured Field types
func (l *Logger) InfoStructured(msg string, fields ...Field) {
	event := l.logger.Info()
	l.addStructuredFields(event, fields...)
	event.Msg(msg)
}

// DebugStructured logs a debug message with structured Field types
func (l *Logger) DebugStructured(msg string, fields ...Field) {
	event := l.logger.Debug()
	l.addStructuredFields(event, fields...)
	event.Msg(msg)
}

// Error logs an error message with map-based fields (backward compatibility)
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// ErrorStructured logs an error message with structured Field types
func (l *Logger) ErrorStructured(msg string, err error, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addStructuredFields(event, fields...)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Success logs a success message with a green checkmark
func (l *Logger) Success(msg string, fields ...map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg("✓ " + msg)
}

// addFields adds fields to a log event (backward compatibility)
func (l *Logger) addFields(event *zerolog.Event, fields ...map[string]interface{}) {
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			event = event.Interface(key, value)
		}
	}
}

// addStructuredFields adds structured Field types to a log event
func (l *Logger) addStructuredFields(event *zerolog.Event, fields ...Field) {
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
}

// With adds structured fields to the logger and returns a new logger instance
// This implements the LoggerInterface.With method
func (l *Logger) With(fields ...Field) LoggerInterface {
	ctx := l.logger.With()
	for _, field := range fields {
		ctx = ctx.Interface(field.Key, field.Value)
	}
	return &loggerAdapter{
		logger: &Logger{logger: ctx.Logger()},
	}
}

// WithRequest adds a request ID to the logger
// This implements the LoggerInterface.WithRequest method
func (l *Logger) WithRequest(requestID string) LoggerInterface {
	return &loggerAdapter{
		logger: &Logger{
			logger: l.logger.With().Str("request_id", requestID).Logger(),
		},
	}
}

// loggerAdapter adapts the Logger to implement LoggerInterface
// This allows the Logger to maintain backward compatibility with map-based fields
// while also implementing the interface with Field-based methods
type loggerAdapter struct {
	logger *Logger
}

// Debug implements LoggerInterface.Debug with Field-based fields
func (a *loggerAdapter) Debug(msg string, fields ...Field) {
	a.logger.DebugStructured(msg, fields...)
}

// Info implements LoggerInterface.Info with Field-based fields
func (a *loggerAdapter) Info(msg string, fields ...Field) {
	a.logger.InfoStructured(msg, fields...)
}

// Warn implements LoggerInterface.Warn with Field-based fields
func (a *loggerAdapter) Warn(msg string, fields ...Field) {
	a.logger.Warn(msg, fields...)
}

// Error implements LoggerInterface.Error with Field-based fields
func (a *loggerAdapter) Error(msg string, err error, fields ...Field) {
	a.logger.ErrorStructured(msg, err, fields...)
}

// With implements LoggerInterface.With
func (a *loggerAdapter) With(fields ...Field) LoggerInterface {
	return a.logger.With(fields...)
}

// WithRequest implements LoggerInterface.WithRequest
func (a *loggerAdapter) WithRequest(requestID string) LoggerInterface {
	return a.logger.WithRequest(requestID)
}

// FromContext retrieves a logger from the context
// If no logger is found, it returns the global logger as an adapter
func FromContext(ctx context.Context) LoggerInterface {
	if ctx == nil {
		return &loggerAdapter{logger: Global()}
	}

	if logger, ok := ctx.Value(loggerContextKey).(LoggerInterface); ok {
		return logger
	}

	return &loggerAdapter{logger: Global()}
}

// ContextWithLogger adds a logger to the context
func ContextWithLogger(ctx context.Context, l LoggerInterface) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerContextKey, l)
}

// Global logger instance
var global *Logger

// InitGlobal initializes the global logger
func InitGlobal(config *Config) {
	global = New(config)
}

// Global returns the global logger instance
func Global() *Logger {
	if global == nil {
		global = New(DefaultConfig())
	}
	return global
}

// Convenience functions for global logger (backward compatibility with map-based fields)
func Debug(msg string, fields ...map[string]interface{}) {
	Global().Debug(msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	Global().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	Global().Warn(msg, fields...)
}

func Error(msg string, err error, fields ...map[string]interface{}) {
	Global().Error(msg, err, fields...)
}

func Fatal(msg string, err error, fields ...map[string]interface{}) {
	Global().Fatal(msg, err, fields...)
}

func Success(msg string, fields ...map[string]interface{}) {
	Global().Success(msg, fields...)
}

func WithApp(app string) *Logger {
	return Global().WithApp(app)
}

func WithField(field string) *Logger {
	return Global().WithField(field)
}

func WithContext(fields map[string]interface{}) *Logger {
	return Global().WithContext(fields)
}
