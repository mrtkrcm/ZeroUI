package observability

import (
	"context"
	"testing"
	"time"
)

// TestMetricsIntegration tests the complete OpenTelemetry metrics integration
func TestMetricsIntegration(t *testing.T) {
	metrics, err := NewMetrics(nil)
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	ctx := context.Background()

	t.Run("Toggle operations", func(t *testing.T) {
		// Test successful toggle
		metrics.RecordToggleOperation(ctx, "test-app", "theme", true, 100*time.Millisecond)
		
		// Test failed toggle
		metrics.RecordToggleOperation(ctx, "test-app", "theme", false, 50*time.Millisecond)
	})

	t.Run("Cycle operations", func(t *testing.T) {
		// Test successful cycle
		metrics.RecordCycleOperation(ctx, "test-app", "font-size", true, 200*time.Millisecond)
		
		// Test failed cycle
		metrics.RecordCycleOperation(ctx, "test-app", "font-size", false, 150*time.Millisecond)
	})

	t.Run("Preset operations", func(t *testing.T) {
		// Test successful preset
		metrics.RecordPresetOperation(ctx, "test-app", "dark-theme", true, 300*time.Millisecond)
		
		// Test failed preset
		metrics.RecordPresetOperation(ctx, "test-app", "dark-theme", false, 250*time.Millisecond)
	})

	t.Run("Error recording", func(t *testing.T) {
		metrics.RecordError(ctx, "config_load", "file_not_found")
		metrics.RecordError(ctx, "config_save", "permission_denied")
	})

	t.Run("App and session tracking", func(t *testing.T) {
		metrics.RecordActiveApps(ctx, 5)
		metrics.RecordTUISession(ctx, 30*time.Second)
	})

	t.Run("Operation timer", func(t *testing.T) {
		timer := metrics.NewOperationTimer(ctx)
		
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		
		timer.RecordToggle("test-app", "theme", true)
		timer.RecordCycle("test-app", "font-size", true)
		timer.RecordPreset("test-app", "minimal", true)
	})
}

// TestMetricsDisabled tests that metrics can be disabled
func TestMetricsDisabled(t *testing.T) {
	config := &MetricsConfig{
		EnableMetrics: false,
	}
	
	metrics, err := NewMetrics(config)
	if err != nil {
		t.Fatalf("Failed to create disabled metrics: %v", err)
	}

	ctx := context.Background()

	// These operations should not panic when metrics are disabled
	metrics.RecordToggleOperation(ctx, "test-app", "theme", true, 100*time.Millisecond)
	metrics.RecordError(ctx, "test_operation", "test_error")
	metrics.RecordActiveApps(ctx, 3)
	
	timer := metrics.NewOperationTimer(ctx)
	timer.RecordToggle("test-app", "theme", true)
}

// TestMetricsConfiguration tests various metrics configurations
func TestMetricsConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		config   *MetricsConfig
		wantErr  bool
	}{
		{
			name:    "Default config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "Enabled metrics",
			config: &MetricsConfig{
				EnableMetrics: true,
			},
			wantErr: false,
		},
		{
			name: "Disabled metrics",
			config: &MetricsConfig{
				EnableMetrics: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := NewMetrics(tt.config)
			
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if metrics == nil {
				t.Error("Expected metrics instance")
			}
		})
	}
}

// BenchmarkMetricsOperations benchmarks metrics recording performance
func BenchmarkMetricsOperations(b *testing.B) {
	metrics, err := NewMetrics(nil)
	if err != nil {
		b.Fatalf("Failed to create metrics: %v", err)
	}

	ctx := context.Background()

	b.Run("RecordToggleOperation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.RecordToggleOperation(ctx, "test-app", "theme", true, 100*time.Millisecond)
		}
	})

	b.Run("RecordError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.RecordError(ctx, "test_operation", "test_error")
		}
	})

	b.Run("OperationTimer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timer := metrics.NewOperationTimer(ctx)
			timer.RecordToggle("test-app", "theme", true)
		}
	})
}