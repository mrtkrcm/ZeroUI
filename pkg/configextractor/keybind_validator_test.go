package configextractor

import (
	"testing"
)

func TestKeybindValidator_ValidateKeybind(t *testing.T) {
	validator := NewKeybindValidator()

	tests := []struct {
		name     string
		keybind  string
		expected bool
		errorMsg string
		hasWarning bool
	}{
		// Valid keybinds
		{
			name:     "simple modifier + key",
			keybind:  "ctrl+c=copy",
			expected: true,
		},
		{
			name:     "multiple modifiers",
			keybind:  "ctrl+shift+c=copy",
			expected: true,
		},
		{
			name:     "super key",
			keybind:  "super+v=paste",
			expected: true,
		},
		{
			name:     "function key",
			keybind:  "f1=reload_config",
			expected: true,
		},
		{
			name:     "special key",
			keybind:  "escape=quit",
			expected: true,
		},
		{
			name:     "action with argument",
			keybind:  "ctrl+shift+right=resize_split_right:10",
			expected: true,
		},
		{
			name:     "single character",
			keybind:  "a=select_all",
			expected: true,
		},

		// Invalid keybinds
		{
			name:     "missing equals",
			keybind:  "ctrl+c copy",
			expected: false,
			errorMsg: "keybind must contain '=' separator",
		},
		{
			name:     "empty keys",
			keybind:  "=copy",
			expected: false,
			errorMsg: "keybind keys cannot be empty",
		},
		{
			name:     "empty action",
			keybind:  "ctrl+c=",
			expected: false,
			errorMsg: "keybind action cannot be empty",
		},
		{
			name:     "invalid key component",
			keybind:  "invalid+c=copy",
			expected: false,
			errorMsg: "invalid key component: 'invalid'",
		},
		{
			name:     "invalid function key",
			keybind:  "f99=copy",
			expected: false,
			errorMsg: "invalid key component: 'f99'",
		},

		// Keybinds with warnings (unknown actions)
		{
			name:     "unknown action",
			keybind:  "ctrl+c=custom_action",
			expected: true,
			hasWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateKeybind(tt.keybind)

			if result.Valid != tt.expected {
				t.Errorf("ValidateKeybind() valid = %v, expected %v", result.Valid, tt.expected)
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

			if tt.hasWarning && len(result.Warnings) == 0 {
				t.Errorf("Expected warning but got none")
			}
		})
	}
}

func TestKeybindValidator_ValidateKeyCombination(t *testing.T) {
	validator := NewKeybindValidator()

	tests := []struct {
		name     string
		keys     string
		expected bool
		errorMsg string
	}{
		// Valid key combinations
		{
			name:     "single modifier",
			keys:     "ctrl",
			expected: true,
		},
		{
			name:     "modifier + key",
			keys:     "ctrl+c",
			expected: true,
		},
		{
			name:     "multiple modifiers + key",
			keys:     "ctrl+shift+c",
			expected: true,
		},
		{
			name:     "super modifier",
			keys:     "super+v",
			expected: true,
		},
		{
			name:     "function key",
			keys:     "f12",
			expected: true,
		},
		{
			name:     "special key",
			keys:     "escape",
			expected: true,
		},
		{
			name:     "single character",
			keys:     "a",
			expected: true,
		},
		{
			name:     "arrow key",
			keys:     "up",
			expected: true,
		},

		// Invalid key combinations
		{
			name:     "invalid modifier",
			keys:     "invalid+c",
			expected: false,
			errorMsg: "invalid key component: 'invalid'",
		},
		{
			name:     "empty string",
			keys:     "",
			expected: false,
			errorMsg: "keybind must contain at least one valid key",
		},
		{
			name:     "invalid function key",
			keys:     "f99",
			expected: false,
			errorMsg: "invalid key component: 'f99'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateKeyCombination(tt.keys)

			if result.Valid != tt.expected {
				t.Errorf("validateKeyCombination() valid = %v, expected %v", result.Valid, tt.expected)
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

func TestKeybindValidator_ValidateAction(t *testing.T) {
	validator := NewKeybindValidator()

	tests := []struct {
		name      string
		action    string
		expected  bool
		hasWarning bool
	}{
		// Valid actions
		{
			name:     "copy action",
			action:   "copy",
			expected: true,
		},
		{
			name:     "action with argument",
			action:   "resize_split_right:10",
			expected: true,
		},
		{
			name:     "paste action",
			action:   "paste",
			expected: true,
		},
		{
			name:     "complex action",
			action:   "goto_tab:3",
			expected: true,
		},

		// Actions with warnings
		{
			name:       "unknown action",
			action:     "custom_action",
			expected:   true,
			hasWarning: true,
		},
		{
			name:       "unknown action with arg",
			action:     "custom:arg",
			expected:   true,
			hasWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.validateAction(tt.action)

			if result.Valid != tt.expected {
				t.Errorf("validateAction() valid = %v, expected %v", result.Valid, tt.expected)
			}

			if tt.hasWarning && len(result.Warnings) == 0 {
				t.Errorf("Expected warning but got none")
			}
		})
	}
}
