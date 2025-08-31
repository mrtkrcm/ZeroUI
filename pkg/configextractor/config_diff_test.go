package configextractor

import (
	"testing"
)

func TestConfigDiffer_DiffConfigurations(t *testing.T) {
	differ := NewConfigDiffer()

	tests := []struct {
		name     string
		old      map[string]interface{}
		new      map[string]interface{}
		expected ConfigDiff
	}{
		{
			name: "no changes",
			old: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
			},
			new: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
			},
			expected: ConfigDiff{
				Added:    map[string]interface{}{},
				Modified: map[string]ValueDiff{},
				Removed:  map[string]interface{}{},
				Unchanged: map[string]interface{}{
					"font-family": "JetBrains Mono",
					"font-size":   12,
				},
			},
		},
		{
			name: "added keys",
			old: map[string]interface{}{
				"font-family": "JetBrains Mono",
			},
			new: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
				"theme":       "dark",
			},
			expected: ConfigDiff{
				Added: map[string]interface{}{
					"font-size": 12,
					"theme":     "dark",
				},
				Modified: map[string]ValueDiff{},
				Removed:  map[string]interface{}{},
				Unchanged: map[string]interface{}{
					"font-family": "JetBrains Mono",
				},
			},
		},
		{
			name: "modified keys",
			old: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
			},
			new: map[string]interface{}{
				"font-family": "Fira Code",
				"font-size":   14,
			},
			expected: ConfigDiff{
				Added:   map[string]interface{}{},
				Removed: map[string]interface{}{},
				Modified: map[string]ValueDiff{
					"font-family": {Old: "JetBrains Mono", New: "Fira Code"},
					"font-size":   {Old: 12, New: 14},
				},
				Unchanged: map[string]interface{}{},
			},
		},
		{
			name: "removed keys",
			old: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
				"theme":       "dark",
			},
			new: map[string]interface{}{
				"font-family": "JetBrains Mono",
			},
			expected: ConfigDiff{
				Added: map[string]interface{}{},
				Removed: map[string]interface{}{
					"font-size": 12,
					"theme":     "dark",
				},
				Modified: map[string]ValueDiff{},
				Unchanged: map[string]interface{}{
					"font-family": "JetBrains Mono",
				},
			},
		},
		{
			name: "mixed changes",
			old: map[string]interface{}{
				"font-family": "JetBrains Mono",
				"font-size":   12,
				"theme":       "dark",
			},
			new: map[string]interface{}{
				"font-family": "Fira Code",
				"font-size":   12,
				"background":  "#000000",
			},
			expected: ConfigDiff{
				Added: map[string]interface{}{
					"background": "#000000",
				},
				Removed: map[string]interface{}{
					"theme": "dark",
				},
				Modified: map[string]ValueDiff{
					"font-family": {Old: "JetBrains Mono", New: "Fira Code"},
				},
				Unchanged: map[string]interface{}{
					"font-size": 12,
				},
			},
		},
		{
			name: "string changes",
			old: map[string]interface{}{
				"keybind": "ctrl+c=copy",
			},
			new: map[string]interface{}{
				"keybind": "ctrl+c=copy,ctrl+v=paste",
			},
			expected: ConfigDiff{
				Added:   map[string]interface{}{},
				Removed: map[string]interface{}{},
				Modified: map[string]ValueDiff{
					"keybind": {
						Old: "ctrl+c=copy",
						New: "ctrl+c=copy,ctrl+v=paste",
					},
				},
				Unchanged: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := differ.DiffConfigurations(tt.old, tt.new)

			// Check added keys
			if len(result.Added) != len(tt.expected.Added) {
				t.Errorf("Added count mismatch: got %d, expected %d", len(result.Added), len(tt.expected.Added))
			}
			for key, expectedValue := range tt.expected.Added {
				if actualValue, exists := result.Added[key]; !exists || actualValue != expectedValue {
					t.Errorf("Added key %s: got %v, expected %v", key, actualValue, expectedValue)
				}
			}

			// Check modified keys
			if len(result.Modified) != len(tt.expected.Modified) {
				t.Errorf("Modified count mismatch: got %d, expected %d", len(result.Modified), len(tt.expected.Modified))
			}
			for key, expectedDiff := range tt.expected.Modified {
				if actualDiff, exists := result.Modified[key]; !exists {
					t.Errorf("Modified key %s not found", key)
				} else if actualDiff.Old != expectedDiff.Old || actualDiff.New != expectedDiff.New {
					t.Errorf("Modified key %s: got {%v → %v}, expected {%v → %v}",
						key, actualDiff.Old, actualDiff.New, expectedDiff.Old, expectedDiff.New)
				}
			}

			// Check removed keys
			if len(result.Removed) != len(tt.expected.Removed) {
				t.Errorf("Removed count mismatch: got %d, expected %d", len(result.Removed), len(tt.expected.Removed))
			}
			for key, expectedValue := range tt.expected.Removed {
				if actualValue, exists := result.Removed[key]; !exists || actualValue != expectedValue {
					t.Errorf("Removed key %s: got %v, expected %v", key, actualValue, expectedValue)
				}
			}

			// Check unchanged keys
			if len(result.Unchanged) != len(tt.expected.Unchanged) {
				t.Errorf("Unchanged count mismatch: got %d, expected %d", len(result.Unchanged), len(tt.expected.Unchanged))
			}
			for key, expectedValue := range tt.expected.Unchanged {
				if actualValue, exists := result.Unchanged[key]; !exists || actualValue != expectedValue {
					t.Errorf("Unchanged key %s: got %v, expected %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigDiff_HasChanges(t *testing.T) {
	tests := []struct {
		name     string
		diff     ConfigDiff
		expected bool
	}{
		{
			name: "no changes",
			diff: ConfigDiff{
				Added:     map[string]interface{}{},
				Modified:  map[string]ValueDiff{},
				Removed:   map[string]interface{}{},
				Unchanged: map[string]interface{}{"key": "value"},
			},
			expected: false,
		},
		{
			name: "has added",
			diff: ConfigDiff{
				Added:    map[string]interface{}{"key": "value"},
				Modified: map[string]ValueDiff{},
				Removed:  map[string]interface{}{},
			},
			expected: true,
		},
		{
			name: "has modified",
			diff: ConfigDiff{
				Added:    map[string]interface{}{},
				Modified: map[string]ValueDiff{"key": {Old: "old", New: "new"}},
				Removed:  map[string]interface{}{},
			},
			expected: true,
		},
		{
			name: "has removed",
			diff: ConfigDiff{
				Added:    map[string]interface{}{},
				Modified: map[string]ValueDiff{},
				Removed:  map[string]interface{}{"key": "value"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diff.HasChanges()
			if result != tt.expected {
				t.Errorf("HasChanges() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConfigDiff_Summary(t *testing.T) {
	tests := []struct {
		name     string
		diff     ConfigDiff
		expected string
	}{
		{
			name: "no changes",
			diff: ConfigDiff{
				Added:     map[string]interface{}{},
				Modified:  map[string]ValueDiff{},
				Removed:   map[string]interface{}{},
				Unchanged: map[string]interface{}{},
			},
			expected: "No changes",
		},
		{
			name: "mixed changes",
			diff: ConfigDiff{
				Added:     map[string]interface{}{"key1": "value1", "key2": "value2"},
				Modified:  map[string]ValueDiff{"key3": {Old: "old", New: "new"}},
				Removed:   map[string]interface{}{"key4": "value4"},
				Unchanged: map[string]interface{}{"key5": "value5"},
			},
			expected: "+2 added, ~1 modified, -1 removed, =1 unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diff.Summary()
			if result != tt.expected {
				t.Errorf("Summary() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestConfigDiff_FormatDiff(t *testing.T) {
	diff := ConfigDiff{
		Added: map[string]interface{}{
			"font-size": 14,
			"theme":     "dark",
		},
		Modified: map[string]ValueDiff{
			"font-family": {Old: "JetBrains Mono", New: "Fira Code"},
		},
		Removed: map[string]interface{}{
			"cursor-blink": true,
		},
		Unchanged: map[string]interface{}{
			"background": "#000000",
		},
	}

	result := diff.FormatDiff()

	// Check that the formatted output contains expected sections
	expectedStrings := []string{
		"Added:",
		"  + font-size = 14",
		"  + theme = dark",
		"Modified:",
		"  ~ font-family: JetBrains Mono → Fira Code",
		"Removed:",
		"  - cursor-blink = true",
	}

	for _, expected := range expectedStrings {
		if !containsString(result, expected) {
			t.Errorf("FormatDiff() missing expected string: %q", expected)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
