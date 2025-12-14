package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// CharmLogger wraps the Charm log library with ZeroUI-specific configuration
type CharmLogger struct {
	logger     *log.Logger
	fileLogger *log.Logger
	level      LogLevel
	logFile    *os.File
}

// LoggerConfig configures the logger behavior
type LoggerConfig struct {
	Level           LogLevel
	EnableFile      bool
	FileLocation    string
	EnableTimestamp bool
	EnableCaller    bool
	Prefix          string
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		Level: func() LogLevel {
			if os.Getenv("ZEROUI_TEST_MODE") == "true" {
				return LevelWarn
			}
			return LevelInfo
		}(),
		EnableFile:      true,
		FileLocation:    getDefaultLogPath(),
		EnableTimestamp: os.Getenv("ZEROUI_TEST_MODE") != "true",
		EnableCaller:    false,
		Prefix:          "zeroui",
	}
}

// NewCharmLogger creates a new logger with beautiful styling
func NewCharmLogger(config LoggerConfig) (*CharmLogger, error) {
	// Create the main logger for terminal output
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    config.EnableCaller,
		ReportTimestamp: config.EnableTimestamp,
		TimeFormat:      "15:04:05",
		Prefix:          config.Prefix,
	})

	// Set up beautiful styles
	logger.SetStyles(createZeroUIStyles())
	logger.SetLevel(mapLogLevel(config.Level))

	charmLogger := &CharmLogger{
		logger: logger,
		level:  config.Level,
	}

	// Set up file logging if enabled
	if config.EnableFile {
		if err := charmLogger.setupFileLogging(config.FileLocation); err != nil {
			return nil, err
		}
	}

	return charmLogger, nil
}

// setupFileLogging configures file-based logging
func (cl *CharmLogger) setupFileLogging(filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open the log file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	cl.logFile = file

	// Create a file logger with structured output
	cl.fileLogger = log.NewWithOptions(file, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		Prefix:          "zeroui",
	})

	// Use plain styles for file output
	cl.fileLogger.SetStyles(createFileStyles())
	cl.fileLogger.SetLevel(log.DebugLevel) // Always log everything to file

	return nil
}

// createZeroUIStyles creates beautiful terminal styles matching ZeroUI theme
func createZeroUIStyles() *log.Styles {
	styles := log.DefaultStyles()

	// Customize level indicators
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255"))

	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("39")).
		Foreground(lipgloss.Color("255"))

	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("220")).
		Foreground(lipgloss.Color("0"))

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("196")).
		Foreground(lipgloss.Color("255"))

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("FATAL").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("129")).
		Foreground(lipgloss.Color("255"))

	// Customize other elements
	styles.Timestamp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	styles.Prefix = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	styles.Message = lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	styles.Key = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)

	styles.Value = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	styles.Separator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		SetString("•")

	return styles
}

// createFileStyles creates plain styles for file output
func createFileStyles() *log.Styles {
	styles := log.DefaultStyles()

	// Remove colors and fancy formatting for files
	for level := range styles.Levels {
		styles.Levels[level] = lipgloss.NewStyle()
	}

	styles.Timestamp = lipgloss.NewStyle()
	styles.Prefix = lipgloss.NewStyle()
	styles.Message = lipgloss.NewStyle()
	styles.Key = lipgloss.NewStyle()
	styles.Value = lipgloss.NewStyle()
	styles.Separator = lipgloss.NewStyle().SetString(" ")

	return styles
}

// mapLogLevel converts our LogLevel to Charm's log level
func mapLogLevel(level LogLevel) log.Level {
	switch level {
	case LevelDebug:
		return log.DebugLevel
	case LevelInfo:
		return log.InfoLevel
	case LevelWarn:
		return log.WarnLevel
	case LevelError:
		return log.ErrorLevel
	case LevelFatal:
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}

// getDefaultLogPath returns the default log file path
func getDefaultLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./zeroui.log"
	}
	return filepath.Join(homeDir, ".local", "state", "zeroui", "zeroui.log")
}

// Logging methods with both terminal and file output

// Debug logs a debug message
func (cl *CharmLogger) Debug(msg string, fields ...interface{}) {
	cl.logger.Debug(msg, fields...)
	if cl.fileLogger != nil {
		cl.fileLogger.Debug(msg, fields...)
	}
}

// Info logs an info message
func (cl *CharmLogger) Info(msg string, fields ...interface{}) {
	cl.logger.Info(msg, fields...)
	if cl.fileLogger != nil {
		cl.fileLogger.Info(msg, fields...)
	}
}

// Warn logs a warning message
func (cl *CharmLogger) Warn(msg string, fields ...interface{}) {
	cl.logger.Warn(msg, fields...)
	if cl.fileLogger != nil {
		cl.fileLogger.Warn(msg, fields...)
	}
}

// Error logs an error message
func (cl *CharmLogger) Error(msg string, fields ...interface{}) {
	cl.logger.Error(msg, fields...)
	if cl.fileLogger != nil {
		cl.fileLogger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func (cl *CharmLogger) Fatal(msg string, fields ...interface{}) {
	cl.logger.Fatal(msg, fields...)
	if cl.fileLogger != nil {
		cl.fileLogger.Fatal(msg, fields...)
	}
}

// WithFields creates a sub-logger with additional fields
func (cl *CharmLogger) WithFields(fields ...interface{}) *CharmLogger {
	newLogger := &CharmLogger{
		logger:  cl.logger.With(fields...),
		level:   cl.level,
		logFile: cl.logFile,
	}

	if cl.fileLogger != nil {
		newLogger.fileLogger = cl.fileLogger.With(fields...)
	}

	return newLogger
}

// WithComponent creates a sub-logger for a specific component
func (cl *CharmLogger) WithComponent(component string) *CharmLogger {
	return cl.WithFields("component", component)
}

// WithOperation creates a sub-logger for a specific operation
func (cl *CharmLogger) WithOperation(operation string) *CharmLogger {
	return cl.WithFields("operation", operation)
}

// WithApp creates a sub-logger for a specific app
func (cl *CharmLogger) WithApp(app string) *CharmLogger {
	return cl.WithFields("app", app)
}

// SetLevel updates the logging level
func (cl *CharmLogger) SetLevel(level LogLevel) {
	cl.level = level
	cl.logger.SetLevel(mapLogLevel(level))
}

// SetOutput redirects the terminal output to a different writer
func (cl *CharmLogger) SetOutput(w io.Writer) {
	cl.logger.SetOutput(w)
}

// Close closes the file logger if it exists
func (cl *CharmLogger) Close() error {
	if cl.logFile != nil {
		return cl.logFile.Close()
	}
	return nil
}

// GetFileLocation returns the current log file path
func (cl *CharmLogger) GetFileLocation() string {
	if cl.logFile != nil {
		return cl.logFile.Name()
	}
	return ""
}

// IsLevelEnabled checks if a log level is enabled
func (cl *CharmLogger) IsLevelEnabled(level LogLevel) bool {
	return level >= cl.level
}

// Structured logging helpers for common ZeroUI operations

// LogAppOperation logs an application-related operation
func (cl *CharmLogger) LogAppOperation(app, operation string, fields ...interface{}) {
	cl.WithApp(app).WithOperation(operation).Info("App operation", fields...)
}

// LogConfigChange logs a configuration change
func (cl *CharmLogger) LogConfigChange(app, field, oldValue, newValue string) {
	cl.WithApp(app).Info("Configuration changed",
		"field", field,
		"old_value", oldValue,
		"new_value", newValue,
	)
}

// LogUIEvent logs a user interface event
func (cl *CharmLogger) LogUIEvent(event, view string, fields ...interface{}) {
	allFields := append([]interface{}{"event", event, "view", view}, fields...)
	cl.WithComponent("tui").Debug("UI event", allFields...)
}

// LogPerformance logs performance metrics
func (cl *CharmLogger) LogPerformance(operation string, duration time.Duration, fields ...interface{}) {
	allFields := append([]interface{}{"operation", operation, "duration_ms", duration.Milliseconds()}, fields...)
	cl.WithComponent("performance").Info("Performance metric", allFields...)
}

// LogPanic logs a panic with context
func (cl *CharmLogger) LogPanic(r interface{}, context string, fields ...interface{}) {
	cl.Error("panic occurred",
		"context", context,
		"panic", fmt.Sprintf("%v", r),
		"fields", fields)
}

// LogError logs an error with context
func (cl *CharmLogger) LogError(err error, context string, fields ...interface{}) {
	allFields := append([]interface{}{"error", err.Error(), "context", context}, fields...)
	cl.Error("Error occurred", allFields...)
}

// Global logger instance
var globalLogger *CharmLogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config LoggerConfig) error {
	logger, err := NewCharmLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *CharmLogger {
	if globalLogger == nil {
		// Create a default logger if none exists
		logger, _ := NewCharmLogger(DefaultConfig())
		globalLogger = logger
	}
	return globalLogger
}

// Global convenience functions
func Debug(msg string, fields ...interface{}) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	GetGlobalLogger().Fatal(msg, fields...)
}
