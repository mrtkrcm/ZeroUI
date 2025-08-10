package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger provides structured logging functionality
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

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	event := l.logger.Warn()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
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

// addFields adds fields to a log event
func (l *Logger) addFields(event *zerolog.Event, fields ...map[string]interface{}) {
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			event = event.Interface(key, value)
		}
	}
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

// Convenience functions for global logger
func Debug(msg string, fields ...map[string]interface{}) {
	Global().Debug(msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	Global().Info(msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
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
