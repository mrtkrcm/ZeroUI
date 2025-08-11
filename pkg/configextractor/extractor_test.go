package configextractor

import (
	"context"
	"testing"
	"time"
)

func TestExtractor_Extract(t *testing.T) {
	tests := []struct {
		name        string
		app         string
		timeout     time.Duration
		expectError bool
		expectCLI   bool
	}{
		{
			name:        "extract_ghostty_success",
			app:         "ghostty",
			timeout:     30 * time.Second,
			expectError: false,
			expectCLI:   true,
		},
		{
			name:        "extract_unknown_app",
			app:         "nonexistent-app",
			timeout:     5 * time.Second,
			expectError: true,
			expectCLI:   false,
		},
		{
			name:        "extract_with_timeout",
			app:         "tmux",
			timeout:     1 * time.Millisecond, // Very short timeout
			expectError: true,
			expectCLI:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create extractor with test-friendly configuration
			extractor := New(
				WithTimeout(tt.timeout),
				WithConcurrency(2),
			)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			config, err := extractor.Extract(ctx, tt.app)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && config != nil {
				// Verify config structure
				if config.App != tt.app {
					t.Errorf("expected app %s, got %s", tt.app, config.App)
				}
				if config.Settings == nil {
					t.Errorf("expected settings to be non-nil")
				}
				if config.Source.Method == "" {
					t.Errorf("expected source method to be set")
				}
				if config.Timestamp.IsZero() {
					t.Errorf("expected timestamp to be set")
				}
			}
		})
	}
}

func TestExtractor_ExtractBatch(t *testing.T) {
	extractor := New(
		WithTimeout(10*time.Second),
		WithConcurrency(3),
	)

	apps := []string{"tmux", "git", "nonexistent-app"}
	ctx := context.Background()

	results, err := extractor.ExtractBatch(ctx, apps)

	// Batch extraction should not error even if some apps fail
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have some successful results
	if len(results) == 0 {
		t.Errorf("expected at least one successful result")
	}

	// Check that successful extractions have proper structure
	for app, config := range results {
		if config.App != app {
			t.Errorf("config app mismatch: expected %s, got %s", app, config.App)
		}
		if config.Settings == nil {
			t.Errorf("config settings should not be nil for app %s", app)
		}
	}
}

func TestExtractor_SupportedApps(t *testing.T) {
	extractor := New()
	apps := extractor.SupportedApps()

	if len(apps) == 0 {
		t.Errorf("expected at least one supported app")
	}

	// Check for some expected apps
	expectedApps := []string{"ghostty", "zed", "alacritty", "tmux", "git"}
	for _, expected := range expectedApps {
		found := false
		for _, app := range apps {
			if app == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected app %s to be supported", expected)
		}
	}
}

func TestExtractor_Cache(t *testing.T) {
	extractor := New()

	ctx := context.Background()
	app := "tmux" // Use an app that should work

	// First extraction
	config1, err1 := extractor.Extract(ctx, app)
	if err1 != nil {
		t.Skipf("skipping cache test, extraction failed: %v", err1)
	}

	// Second extraction (should use cache)
	start := time.Now()
	config2, err2 := extractor.Extract(ctx, app)
	elapsed := time.Since(start)

	if err2 != nil {
		t.Errorf("second extraction failed: %v", err2)
	}

	// Cache hit should be very fast (< 1ms typically)
	if elapsed > 10*time.Millisecond {
		t.Errorf("cache lookup took too long: %v", elapsed)
	}

	// Configs should be identical
	if config1.App != config2.App {
		t.Errorf("cached config differs from original")
	}
}

func BenchmarkExtractor_Extract(b *testing.B) {
	extractor := New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(ctx, "tmux")
	}
}

func BenchmarkExtractor_ExtractBatch(b *testing.B) {
	extractor := New()
	ctx := context.Background()
	apps := []string{"tmux", "git"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractBatch(ctx, apps)
	}
}

// MockStrategy for testing
type MockStrategy struct {
	name       string
	priority   int
	canExtract map[string]bool
	config     *Config
	err        error
	delay      time.Duration
}

func NewMockStrategy(name string, priority int) *MockStrategy {
	return &MockStrategy{
		name:       name,
		priority:   priority,
		canExtract: make(map[string]bool),
	}
}

func (m *MockStrategy) Name() string {
	return m.name
}

func (m *MockStrategy) Priority() int {
	return m.priority
}

func (m *MockStrategy) CanExtract(app string) bool {
	return m.canExtract[app]
}

func (m *MockStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
		}
	}
	return m.config, m.err
}

func (m *MockStrategy) SetCanExtract(app string, canExtract bool) {
	m.canExtract[app] = canExtract
}

func (m *MockStrategy) SetResult(config *Config, err error) {
	m.config = config
	m.err = err
}

func (m *MockStrategy) SetDelay(delay time.Duration) {
	m.delay = delay
}

func TestExtractor_StrategyPriority(t *testing.T) {
	// Create mock strategies with different priorities
	highPrio := NewMockStrategy("high", 100)
	lowPrio := NewMockStrategy("low", 10)

	highPrio.SetCanExtract("test-app", true)
	lowPrio.SetCanExtract("test-app", true)

	highPrio.SetResult(&Config{
		App:    "test-app",
		Format: "high-priority",
		Settings: map[string]Setting{
			"source": {Name: "source", Type: TypeString, Default: "high"},
		},
	}, nil)

	lowPrio.SetResult(&Config{
		App:    "test-app",
		Format: "low-priority",
		Settings: map[string]Setting{
			"source": {Name: "source", Type: TypeString, Default: "low"},
		},
	}, nil)

	// Create extractor with custom strategies
	extractor := New(
		WithStrategy(lowPrio),  // Add low priority first
		WithStrategy(highPrio), // Add high priority second
	)

	ctx := context.Background()
	config, err := extractor.Extract(ctx, "test-app")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should get result from high-priority strategy
	if config.Format != "high-priority" {
		t.Errorf("expected high-priority result, got %s", config.Format)
	}
}
