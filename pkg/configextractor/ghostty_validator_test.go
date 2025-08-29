package configextractor

import (
	"testing"
)

func TestGhosttySchemaValidator_ValidateField(t *testing.T) {
	validator := NewGhosttySchemaValidator()

	tests := []struct {
		name     string
		field    string
		value    interface{}
		expected bool
		errorMsg string
	}{
		// Valid fields
		{
			name:     "valid cursor-color",
			field:    "cursor-color",
			value:    "#ff0000",
			expected: true,
		},
		{
			name:     "valid cursor-style",
			field:    "cursor-style",
			value:    "block",
			expected: true,
		},
		{
			name:     "valid window-padding-x",
			field:    "window-padding-x",
			value:    10,
			expected: true,
		},
		{
			name:     "valid font-family",
			field:    "font-family",
			value:    "JetBrains Mono",
			expected: true,
		},
		{
			name:     "valid boolean cursor-invert-fg-bg",
			field:    "cursor-invert-fg-bg",
			value:    true,
			expected: true,
		},

		// Invalid fields (should be rejected)
		{
			name:     "invalid cursor-blink",
			field:    "cursor-blink",
			value:    true,
			expected: false,
			errorMsg: "field 'cursor-blink' is not a valid Ghostty configuration option",
		},
		{
			name:     "invalid window-padding",
			field:    "window-padding",
			value:    10,
			expected: false,
			errorMsg: "field 'window-padding' is not a valid Ghostty configuration option",
		},
		{
			name:     "invalid cursor-shape (wrong enum)",
			field:    "cursor-style",
			value:    "invalid-shape",
			expected: false,
			errorMsg: "field 'cursor-style' must be one of: block, bar, underline, outline",
		},
		{
			name:     "invalid color format",
			field:    "cursor-color",
			value:    "not-a-color",
			expected: false,
			errorMsg: "field 'cursor-color' must be a valid color (hex or named color)",
		},
		{
			name:     "invalid type for cursor-color",
			field:    "cursor-color",
			value:    123,
			expected: false,
			errorMsg: "field 'cursor-color' must be of type color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateField(tt.field, tt.value)

			if result.Valid != tt.expected {
				t.Errorf("ValidateField() valid = %v, expected %v", result.Valid, tt.expected)
			}

			if !result.Valid && tt.errorMsg != "" {
				found := false
				for _, err := range result.Errors {
					if err == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message '%s', got errors: %v", tt.errorMsg, result.Errors)
				}
			}
		})
	}
}

func TestGhosttySchemaValidator_ValidateConfig(t *testing.T) {
	validator := NewGhosttySchemaValidator()

	t.Run("valid configuration", func(t *testing.T) {
		config := map[string]interface{}{
			"font-family":        "JetBrains Mono",
			"font-size":          12,
			"cursor-style":       "block",
			"cursor-color":       "#ff0000",
			"window-padding-x":   10,
			"window-padding-y":   10,
			"background":         "#000000",
			"foreground":         "#ffffff",
		}

		result := validator.ValidateConfig(config)
		if !result.Valid {
			t.Errorf("Expected valid configuration, got errors: %v", result.Errors)
		}
	})

	t.Run("configuration with invalid fields", func(t *testing.T) {
		config := map[string]interface{}{
			"font-family":        "JetBrains Mono",
			"cursor-blink":       true,        // Invalid field
			"window-padding":     10,          // Invalid field
			"cursor-style":       "invalid",   // Invalid value
			"cursor-color":       "not-color", // Invalid color
		}

		result := validator.ValidateConfig(config)
		if result.Valid {
			t.Error("Expected invalid configuration")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}

		// Check that we get errors for invalid fields
		errorText := ""
		for _, err := range result.Errors {
			errorText += err + " "
		}

		if !containsError(errorText, "cursor-blink") {
			t.Error("Expected error for invalid field 'cursor-blink'")
		}
		if !containsError(errorText, "window-padding") {
			t.Error("Expected error for invalid field 'window-padding'")
		}
		if !containsError(errorText, "cursor-style") {
			t.Error("Expected error for invalid cursor-style value")
		}
		if !containsError(errorText, "cursor-color") {
			t.Error("Expected error for invalid cursor-color value")
		}
	})
}

func TestValidateGhosttyConfig(t *testing.T) {
	t.Run("convenience function", func(t *testing.T) {
		config := map[string]interface{}{
			"cursor-blink": true, // Invalid field
		}

		result := ValidateGhosttyConfig(config)
		if result.Valid {
			t.Error("Expected invalid configuration")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
	})
}

// Helper function to check if error contains specific text
func containsError(errorText, fieldName string) bool {
	return len(errorText) > 0 && (errorText == fieldName ||
		len(errorText) >= len(fieldName) &&
		func() bool {
			for i := 0; i <= len(errorText)-len(fieldName); i++ {
				if errorText[i:i+len(fieldName)] == fieldName {
					return true
				}
			}
			return false
		}())
}
