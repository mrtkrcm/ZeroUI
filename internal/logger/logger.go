package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type contextKey string

// Field represents a structured log field with optional redaction.
type Field struct {
	Key    string
	Value  interface{}
	Redact bool
}

// Logger defines structured logging capabilities with contextual scoping.
type Logger interface {
	With(fields ...Field) Logger
	WithApp(app string) Logger
	WithField(key string) Logger
	WithRedacted(key string, value interface{}) Logger
	WithRequest(name string) Logger

	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, err error, fields ...Field)
	Success(msg string, fields ...Field)
}

// Config holds logger configuration sourced from CLI flags and config.
type Config struct {
	Level         string
	Format        string // json, console
	Output        io.Writer
	TimeFormat    string
	Verbose       bool
	DryRun        bool
	EnableTracing bool
	Session       string
}

// DefaultConfig returns a default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:         "info",
		Format:        "console",
		Output:        os.Stderr,
		TimeFormat:    time.RFC3339,
		EnableTracing: true,
	}
}

// New creates a new logger with the given configuration.
func New(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}

	cfg := *config
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	level := parseLevel(cfg.Level, cfg.Verbose)

	var base zerolog.Logger
	if cfg.Format == "console" {
		output := zerolog.ConsoleWriter{
			Out:        cfg.Output,
			TimeFormat: cfg.TimeFormat,
		}
		base = zerolog.New(output).With().Timestamp().Logger()
	} else {
		base = zerolog.New(cfg.Output).With().Timestamp().Logger()
	}

	builder := base.Level(level).With()
	if cfg.DryRun {
		builder = builder.Bool("dry_run", true)
	}
	if cfg.Session != "" {
		builder = builder.Str("session", cfg.Session)
	}

	logger := builder.Logger()
	if cfg.EnableTracing {
		logger = logger.Hook(tracingHook{session: cfg.Session})
	}

	return &zerologLogger{
		logger: logger,
		cfg:    cfg,
	}
}

// ContextWithLogger embeds a logger in a context for request-scoped logging.
func ContextWithLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, contextKey("logger"), log)
}

// FromContext extracts a logger from context if available.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return nil
	}
	if log, ok := ctx.Value(contextKey("logger")).(Logger); ok {
		return log
	}
	return nil
}

type zerologLogger struct {
	logger zerolog.Logger
	cfg    Config
}

func (l *zerologLogger) With(fields ...Field) Logger {
	ctx := l.logger.With()
	for _, field := range fields {
		ctx = appendField(ctx, field)
	}
	return &zerologLogger{
		logger: ctx.Logger(),
		cfg:    l.cfg,
	}
}

func (l *zerologLogger) WithApp(app string) Logger {
	return l.With(Field{Key: "app", Value: app})
}

func (l *zerologLogger) WithField(key string) Logger {
	return l.With(Field{Key: "field", Value: key})
}

func (l *zerologLogger) WithRedacted(key string, value interface{}) Logger {
	return l.With(Field{Key: key, Value: value, Redact: true})
}

func (l *zerologLogger) WithRequest(name string) Logger {
	return l.With(Field{Key: "request", Value: name})
}

func (l *zerologLogger) Debug(msg string, fields ...Field) {
	l.writeEvent(l.logger.Debug(), msg, fields...)
}

func (l *zerologLogger) Info(msg string, fields ...Field) {
	l.writeEvent(l.logger.Info(), msg, fields...)
}

func (l *zerologLogger) Warn(msg string, fields ...Field) {
	l.writeEvent(l.logger.Warn(), msg, fields...)
}

func (l *zerologLogger) Error(msg string, err error, fields ...Field) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.writeEvent(event, msg, fields...)
}

func (l *zerologLogger) Fatal(msg string, err error, fields ...Field) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	l.writeEvent(event, msg, fields...)
}

func (l *zerologLogger) Success(msg string, fields ...Field) {
	l.Info("✓ "+msg, fields...)
}

func (l *zerologLogger) writeEvent(event *zerolog.Event, msg string, fields ...Field) {
	for _, field := range fields {
		event = appendFieldToEvent(event, field)
	}
	event.Msg(msg)
}

func parseLevel(level string, verbose bool) zerolog.Level {
	if verbose {
		return zerolog.DebugLevel
	}
	parsed, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return parsed
}

func appendField(ctx zerolog.Context, field Field) zerolog.Context {
	if field.Redact {
		return ctx.Str(field.Key, "[REDACTED]")
	}
	return ctx.Interface(field.Key, field.Value)
}

func appendFieldToEvent(event *zerolog.Event, field Field) *zerolog.Event {
	if field.Redact {
		return event.Str(field.Key, "[REDACTED]")
	}
	return event.Interface(field.Key, field.Value)
}

// tracingHook attaches tracing metadata to log events.
type tracingHook struct {
	session string
}

func (h tracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if h.session != "" {
		e.Str("trace_session", h.session)
	}
	e.Str("trace_phase", "event")
}
