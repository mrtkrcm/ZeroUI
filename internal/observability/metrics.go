package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	instrumentationName = "github.com/mrtkrcm/ZeroUI"
)

// Metrics holds all application metrics
type Metrics struct {
	// Operation counters
	toggleOperations metric.Int64Counter
	cycleOperations  metric.Int64Counter
	presetOperations metric.Int64Counter

	// Duration histograms
	operationDuration metric.Float64Histogram

	// Error counters
	operationErrors metric.Int64Counter

	// Application metrics
	activeApps    metric.Int64UpDownCounter
	configChanges metric.Int64Counter

	// TUI metrics
	tuiSessions metric.Int64Counter
	tuiDuration metric.Float64Histogram
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	EnableMetrics  bool
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		ServiceName:    "zeroui",
		ServiceVersion: "unknown",
		Environment:    "development",
		EnableMetrics:  true,
	}
}

// NewMetrics creates a new metrics instance
func NewMetrics(config *MetricsConfig) (*Metrics, error) {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	if !config.EnableMetrics {
		return &Metrics{}, nil // Return empty metrics if disabled
	}

	// Create Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// Create meter provider
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	// Get meter
	meter := otel.Meter(instrumentationName)

	// Initialize metrics
	m := &Metrics{}

	// Operation counters
	if m.toggleOperations, err = meter.Int64Counter(
		"zeroui_toggle_operations_total",
		metric.WithDescription("Total number of toggle operations"),
	); err != nil {
		return nil, err
	}

	if m.cycleOperations, err = meter.Int64Counter(
		"zeroui_cycle_operations_total",
		metric.WithDescription("Total number of cycle operations"),
	); err != nil {
		return nil, err
	}

	if m.presetOperations, err = meter.Int64Counter(
		"zeroui_preset_operations_total",
		metric.WithDescription("Total number of preset operations"),
	); err != nil {
		return nil, err
	}

	// Duration histogram
	if m.operationDuration, err = meter.Float64Histogram(
		"zeroui_operation_duration_seconds",
		metric.WithDescription("Duration of operations in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	// Error counter
	if m.operationErrors, err = meter.Int64Counter(
		"zeroui_operation_errors_total",
		metric.WithDescription("Total number of operation errors"),
	); err != nil {
		return nil, err
	}

	// Application metrics
	if m.activeApps, err = meter.Int64UpDownCounter(
		"zeroui_active_apps",
		metric.WithDescription("Number of active applications"),
	); err != nil {
		return nil, err
	}

	if m.configChanges, err = meter.Int64Counter(
		"zeroui_config_changes_total",
		metric.WithDescription("Total number of configuration changes"),
	); err != nil {
		return nil, err
	}

	// TUI metrics
	if m.tuiSessions, err = meter.Int64Counter(
		"zeroui_tui_sessions_total",
		metric.WithDescription("Total number of TUI sessions"),
	); err != nil {
		return nil, err
	}

	if m.tuiDuration, err = meter.Float64Histogram(
		"zeroui_tui_session_duration_seconds",
		metric.WithDescription("Duration of TUI sessions in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	return m, nil
}

// RecordToggleOperation records a toggle operation
func (m *Metrics) RecordToggleOperation(ctx context.Context, app, key string, success bool, duration time.Duration) {
	if m.toggleOperations == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("app", app),
		attribute.String("key", key),
		attribute.Bool("success", success),
	}

	m.toggleOperations.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if !success {
		m.operationErrors.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "toggle"),
			attribute.String("app", app),
		))
	} else {
		m.configChanges.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordCycleOperation records a cycle operation
func (m *Metrics) RecordCycleOperation(ctx context.Context, app, key string, success bool, duration time.Duration) {
	if m.cycleOperations == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("app", app),
		attribute.String("key", key),
		attribute.Bool("success", success),
	}

	m.cycleOperations.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if !success {
		m.operationErrors.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "cycle"),
			attribute.String("app", app),
		))
	} else {
		m.configChanges.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordPresetOperation records a preset operation
func (m *Metrics) RecordPresetOperation(ctx context.Context, app, preset string, success bool, duration time.Duration) {
	if m.presetOperations == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("app", app),
		attribute.String("preset", preset),
		attribute.Bool("success", success),
	}

	m.presetOperations.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if !success {
		m.operationErrors.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "preset"),
			attribute.String("app", app),
		))
	} else {
		// Count all config changes in the preset
		m.configChanges.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordActiveApps records the number of active applications
func (m *Metrics) RecordActiveApps(ctx context.Context, count int64) {
	if m.activeApps == nil {
		return
	}
	m.activeApps.Add(ctx, count)
}

// RecordTUISession records a TUI session
func (m *Metrics) RecordTUISession(ctx context.Context, duration time.Duration) {
	if m.tuiSessions == nil {
		return
	}

	m.tuiSessions.Add(ctx, 1)
	m.tuiDuration.Record(ctx, duration.Seconds())
}

// RecordError records a general error
func (m *Metrics) RecordError(ctx context.Context, operation, errorType string) {
	if m.operationErrors == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("error_type", errorType),
	}

	m.operationErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// OperationTimer helps measure operation duration
type OperationTimer struct {
	start   time.Time
	metrics *Metrics
	ctx     context.Context
}

// NewOperationTimer creates a new operation timer
func (m *Metrics) NewOperationTimer(ctx context.Context) *OperationTimer {
	return &OperationTimer{
		start:   time.Now(),
		metrics: m,
		ctx:     ctx,
	}
}

// RecordToggle records a toggle operation with the timer
func (t *OperationTimer) RecordToggle(app, key string, success bool) {
	duration := time.Since(t.start)
	t.metrics.RecordToggleOperation(t.ctx, app, key, success, duration)
}

// RecordCycle records a cycle operation with the timer
func (t *OperationTimer) RecordCycle(app, key string, success bool) {
	duration := time.Since(t.start)
	t.metrics.RecordCycleOperation(t.ctx, app, key, success, duration)
}

// RecordPreset records a preset operation with the timer
func (t *OperationTimer) RecordPreset(app, preset string, success bool) {
	duration := time.Since(t.start)
	t.metrics.RecordPresetOperation(t.ctx, app, preset, success, duration)
}
