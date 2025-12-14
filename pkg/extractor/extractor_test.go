package extractor

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Mock strategy for testing
type mockStrategy struct {
	name       string
	confidence float64
	config     *Config
	err        error
	delay      time.Duration
}

func (m *mockStrategy) Name() string        { return m.name }
func (m *mockStrategy) Confidence() float64 { return m.confidence }
func (m *mockStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

func TestExtractor_Extract(t *testing.T) {
	tests := []struct {
		name       string
		strategies []Strategy
		wantApp    string
		wantErr    bool
	}{
		{
			name: "successful extraction",
			strategies: []Strategy{
				&mockStrategy{
					name:       "mock1",
					confidence: 0.9,
					config: &Config{
						App: "test",
						Settings: map[string]Setting{
							"option1": {Name: "option1", Type: "string"},
						},
					},
				},
			},
			wantApp: "test",
			wantErr: false,
		},
		{
			name: "highest confidence wins",
			strategies: []Strategy{
				&mockStrategy{
					name:       "low",
					confidence: 0.5,
					config:     &Config{App: "low"},
				},
				&mockStrategy{
					name:       "high",
					confidence: 0.9,
					config:     &Config{App: "high"},
				},
			},
			wantApp: "high",
			wantErr: false,
		},
		{
			name: "all strategies fail",
			strategies: []Strategy{
				&mockStrategy{name: "fail1", err: fmt.Errorf("error1")},
				&mockStrategy{name: "fail2", err: fmt.Errorf("error2")},
			},
			wantApp: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Extractor{
				strategies: tt.strategies,
				cache:      &NoOpCache{},
				pool:       make(chan struct{}, 1),
				timeout:    1 * time.Second,
			}

			ctx := context.Background()
			got, err := e.Extract(ctx, "test")

			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.App != tt.wantApp {
				t.Errorf("Extract() got app = %v, want %v", got.App, tt.wantApp)
			}
		})
	}
}

func TestExtractor_ExtractBatch(t *testing.T) {
	strategy := &mockStrategy{
		name:       "test",
		confidence: 0.9,
		config: &Config{
			App: "test",
			Settings: map[string]Setting{
				"option1": {Name: "option1", Type: "string"},
			},
		},
	}

	e := New(
		WithStrategy(strategy),
		WithConcurrency(2),
	)

	ctx := context.Background()
	apps := []string{"app1", "app2", "app3"}

	results, err := e.ExtractBatch(ctx, apps)
	if err != nil {
		t.Fatalf("ExtractBatch() error = %v", err)
	}

	if len(results) != len(apps) {
		t.Errorf("ExtractBatch() got %d results, want %d", len(results), len(apps))
	}
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(2, 1*time.Hour)

	config1 := &Config{App: "app1"}
	config2 := &Config{App: "app2"}
	config3 := &Config{App: "app3"}

	// Test set and get
	cache.Set("app1", config1)
	cache.Set("app2", config2)

	if got, ok := cache.Get("app1"); !ok || got.App != "app1" {
		t.Errorf("Cache.Get(app1) failed")
	}

	// Test LRU eviction
	cache.Set("app3", config3) // Should evict app2

	if _, ok := cache.Get("app2"); ok {
		t.Errorf("Cache should have evicted app2")
	}

	if got, ok := cache.Get("app3"); !ok || got.App != "app3" {
		t.Errorf("Cache.Get(app3) failed")
	}

	// Test clear
	cache.Clear()
	if _, ok := cache.Get("app1"); ok {
		t.Errorf("Cache.Clear() failed")
	}
}

func TestValidator(t *testing.T) {
	v := NewValidator()

	// Add rules
	v.AddRule("font-size", NumberRule{
		Min: Min(8),
		Max: Max(72),
	})

	v.AddRule("theme", ChoiceRule{
		Choices: []string{"light", "dark", "auto"},
	})

	// Test valid values
	if valid, msg := v.Validate("font-size", "14"); !valid {
		t.Errorf("Validate(font-size, 14) failed: %s", msg)
	}

	if valid, msg := v.Validate("theme", "dark"); !valid {
		t.Errorf("Validate(theme, dark) failed: %s", msg)
	}

	// Test invalid values
	if valid, _ := v.Validate("font-size", "5"); valid {
		t.Errorf("Validate(font-size, 5) should fail")
	}

	if valid, _ := v.Validate("theme", "invalid"); valid {
		t.Errorf("Validate(theme, invalid) should fail")
	}

	// Test no rule = valid
	if valid, _ := v.Validate("unknown", "anything"); !valid {
		t.Errorf("Validate(unknown) should be valid when no rule")
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"true", "boolean"},
		{"false", "boolean"},
		{"123", "number"},
		{"12.34", "number"},
		{"-42", "number"},
		{"#ffffff", "color"},
		{"0xDEADBEEF", "color"},
		{"foo,bar,baz", "array"},
		{"hello world", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := inferType(tt.input); got != tt.want {
				t.Errorf("inferType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestInferCategory(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"font-family", "font"},
		{"text-size", "font"},
		{"background-color", "appearance"},
		{"theme-mode", "appearance"},
		{"window-width", "window"},
		{"keybind-copy", "keybindings"},
		{"editor-tabsize", "editor"},
		{"git-diff", "git"},
		{"random-option", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := inferCategory(tt.input); got != tt.want {
				t.Errorf("inferCategory(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkExtractor_Extract(b *testing.B) {
	e := New(
		WithStrategy(&mockStrategy{
			name:       "bench",
			confidence: 0.9,
			config:     &Config{App: "test"},
		}),
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Extract(ctx, "test")
	}
}

func BenchmarkCache_GetSet(b *testing.B) {
	cache := NewLRUCache(100, 1*time.Hour)
	config := &Config{App: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("app%d", i%10)
		cache.Set(key, config)
		cache.Get(key)
	}
}
