package forms

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFieldSeparation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		fields         []ConfigField
		expectedGroups map[string]int // group name -> number of fields
	}{
		{
			name: "existing_vs_available_separation",
			fields: []ConfigField{
				{
					Key:    "theme",
					Type:   FieldTypeString,
					Value:  "dark",
					IsSet:  true,
					Source: "config file",
				},
				{
					Key:    "font_size",
					Type:   FieldTypeInt,
					Value:  14,
					IsSet:  true,
					Source: "config file",
				},
				{
					Key:    "auto_save",
					Type:   FieldTypeBool,
					IsSet:  false,
					Source: "available option",
				},
				{
					Key:    "background_color",
					Type:   FieldTypeString,
					IsSet:  false,
					Source: "available option",
				},
			},
			expectedGroups: map[string]int{
				"[*] Current Configuration - Appearance": 1, // theme
				"[*] Current Configuration - Typography": 1, // font_size
				"[+] Available Options - Behavior":       1, // auto_save
				"[+] Available Options - Appearance":     1, // background_color
			},
		},
		{
			name: "all_existing_configurations",
			fields: []ConfigField{
				{
					Key:    "window_width",
					Type:   FieldTypeInt,
					Value:  800,
					IsSet:  true,
					Source: "config file",
				},
				{
					Key:    "window_height",
					Type:   FieldTypeInt,
					Value:  600,
					IsSet:  true,
					Source: "config file",
				},
			},
			expectedGroups: map[string]int{
				"[*] Current Configuration - Window": 2,
			},
		},
		{
			name: "all_available_options",
			fields: []ConfigField{
				{
					Key:    "enable_notifications",
					Type:   FieldTypeBool,
					IsSet:  false,
					Source: "available option",
				},
				{
					Key:    "cursor_blink",
					Type:   FieldTypeBool,
					IsSet:  false,
					Source: "available option",
				},
			},
			expectedGroups: map[string]int{
				"[+] Available Options - Behavior": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := NewHuhConfigForm("test-app")
			form.SetFields(tt.fields)

			groups := form.groupFields()

			// Verify expected groups exist with correct field counts
			for expectedGroup, expectedCount := range tt.expectedGroups {
				fields, exists := groups[expectedGroup]
				assert.True(t, exists, "Expected group %s to exist", expectedGroup)
				assert.Len(t, fields, expectedCount, "Expected %d fields in group %s", expectedCount, expectedGroup)
			}

			// Verify no unexpected groups
			assert.Len(t, groups, len(tt.expectedGroups), "Unexpected number of groups")
		})
	}
}

func TestFormatTitle(t *testing.T) {
	t.Parallel()
	form := NewHuhConfigForm("test-app")

	tests := []struct {
		name     string
		field    ConfigField
		expected string
	}{
		{
			name: "existing_field_with_value",
			field: ConfigField{
				Key:    "theme_mode",
				Value:  "dark",
				IsSet:  true,
				Source: "config file",
			},
			expected: "[*] Theme Mode (current: dark)",
		},
		{
			name: "existing_field_empty_value",
			field: ConfigField{
				Key:    "background_image",
				Value:  "",
				IsSet:  true,
				Source: "config file",
			},
			expected: "[!] Background Image (set but empty)",
		},
		{
			name: "available_option",
			field: ConfigField{
				Key:    "auto_save_interval",
				IsSet:  false,
				Source: "available option",
			},
			expected: "[ ] Auto Save Interval (available)",
		},
		{
			name: "snake_case_conversion",
			field: ConfigField{
				Key:    "font_family_monospace",
				Value:  "JetBrains Mono",
				IsSet:  true,
				Source: "config file",
			},
			expected: "[*] Font Family Monospace (current: JetBrains Mono)",
		},
		{
			name: "kebab_case_conversion",
			field: ConfigField{
				Key:    "cursor-blink-rate",
				Value:  500,
				IsSet:  true,
				Source: "config file",
			},
			expected: "[*] Cursor Blink Rate (current: 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := form.formatTitle(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDescription(t *testing.T) {
	t.Parallel()
	form := NewHuhConfigForm("test-app")

	tests := []struct {
		name     string
		field    ConfigField
		expected string
	}{
		{
			name: "existing_field_with_source",
			field: ConfigField{
				Key:         "theme",
				Description: "Color theme for the application",
				IsSet:       true,
				Source:      "config file",
			},
			expected: "Color theme for the application\n>> Source: config file",
		},
		{
			name: "existing_field_without_source",
			field: ConfigField{
				Key:         "font_size",
				Description: "Font size in pixels",
				IsSet:       true,
			},
			expected: "Font size in pixels\n>> Currently configured",
		},
		{
			name: "available_option",
			field: ConfigField{
				Key:         "auto_save",
				Description: "Automatically save changes",
				IsSet:       false,
				Source:      "available option",
			},
			expected: "Automatically save changes\n>> Available option - not currently set",
		},
		{
			name: "field_without_description",
			field: ConfigField{
				Key:    "custom_field",
				IsSet:  false,
				Source: "available option",
			},
			expected: "Configure custom_field\n>> Available option - not currently set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := form.formatDescription(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldTypeHandling(t *testing.T) {
	t.Parallel()
	form := NewHuhConfigForm("test-app")

	tests := []struct {
		name       string
		field      ConfigField
		shouldWork bool
	}{
		{
			name: "string_field",
			field: ConfigField{
				Key:   "app_name",
				Type:  FieldTypeString,
				Value: "MyApp",
				IsSet: true,
			},
			shouldWork: true,
		},
		{
			name: "int_field",
			field: ConfigField{
				Key:   "port",
				Type:  FieldTypeInt,
				Value: 8080,
				IsSet: true,
			},
			shouldWork: true,
		},
		{
			name: "bool_field",
			field: ConfigField{
				Key:   "debug_mode",
				Type:  FieldTypeBool,
				Value: true,
				IsSet: true,
			},
			shouldWork: true,
		},
		{
			name: "select_field",
			field: ConfigField{
				Key:     "log_level",
				Type:    FieldTypeSelect,
				Value:   "info",
				Options: []string{"debug", "info", "warn", "error"},
				IsSet:   true,
			},
			shouldWork: true,
		},
		{
			name: "float_field",
			field: ConfigField{
				Key:   "opacity",
				Type:  FieldTypeFloat,
				Value: 0.8,
				IsSet: true,
			},
			shouldWork: true,
		},
		{
			name: "select_field_no_options",
			field: ConfigField{
				Key:     "invalid_select",
				Type:    FieldTypeSelect,
				Value:   "value",
				Options: []string{}, // No options provided
				IsSet:   true,
			},
			shouldWork: false, // Should return nil for invalid select field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form.SetFields([]ConfigField{tt.field})

			// Test field creation
			huhField := form.createHuhField(tt.field)

			if tt.shouldWork {
				assert.NotNil(t, huhField, "Expected field to be created successfully")
			} else {
				assert.Nil(t, huhField, "Expected field creation to fail")
			}
		})
	}
}

func TestConfigFormIntegration(t *testing.T) {
	t.Parallel()
	// Test realistic configuration scenarios
	t.Run("ghostty_terminal_config", func(t *testing.T) {
		form := NewHuhConfigForm("ghostty")

		// Simulate realistic Ghostty configuration
		fields := []ConfigField{
			{
				Key:         "theme",
				Type:        FieldTypeString,
				Value:       "tokyo-night",
				Description: "Color theme for terminal",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "font-family",
				Type:        FieldTypeString,
				Value:       "JetBrains Mono",
				Description: "Font family for terminal text",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "font-size",
				Type:        FieldTypeInt,
				Value:       14,
				Description: "Font size in points",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "window-opacity",
				Type:        FieldTypeFloat,
				Value:       0.95,
				Description: "Window transparency level",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "cursor-style",
				Type:        FieldTypeSelect,
				Options:     []string{"block", "underline", "bar"},
				Value:       "block",
				Description: "Cursor appearance style",
				IsSet:       true,
				Source:      "config file",
			},
			// Available options not yet configured
			{
				Key:         "auto-hide-mouse",
				Type:        FieldTypeBool,
				Description: "Hide mouse cursor when typing",
				IsSet:       false,
				Source:      "available option",
			},
			{
				Key:         "bell-sound",
				Type:        FieldTypeString,
				Description: "Custom bell sound file path",
				IsSet:       false,
				Source:      "available option",
			},
			{
				Key:         "scrollback-lines",
				Type:        FieldTypeInt,
				Description: "Number of scrollback lines to keep",
				IsSet:       false,
				Source:      "available option",
			},
		}

		form.SetFields(fields)
		groups := form.groupFields()

		// Verify existing configurations are properly grouped
		assert.Contains(t, groups, "[*] Current Configuration - Appearance")
		assert.Contains(t, groups, "[*] Current Configuration - Typography")
		assert.Contains(t, groups, "[*] Current Configuration - Window")

		// Verify available options are properly grouped
		assert.Contains(t, groups, "[+] Available Options - Behavior")
		assert.Contains(t, groups, "[+] Available Options - Appearance")

		// Test field formatting
		themeField := fields[0] // theme field
		title := form.formatTitle(themeField)
		assert.Contains(t, title, "[*]")
		assert.Contains(t, title, "tokyo-night")

		availableField := fields[5] // auto-hide-mouse
		availableTitle := form.formatTitle(availableField)
		assert.Contains(t, availableTitle, "[ ]")
		assert.Contains(t, availableTitle, "available")
	})

	t.Run("zed_editor_config", func(t *testing.T) {
		form := NewHuhConfigForm("zed")

		// Simulate realistic Zed editor configuration
		fields := []ConfigField{
			{
				Key:         "theme",
				Type:        FieldTypeString,
				Value:       "One Dark",
				Description: "Editor color theme",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "buffer_font_family",
				Type:        FieldTypeString,
				Value:       "JetBrains Mono",
				Description: "Font for editor buffers",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "ui_font_size",
				Type:        FieldTypeInt,
				Value:       14,
				Description: "UI font size",
				IsSet:       true,
				Source:      "config file",
			},
			{
				Key:         "format_on_save",
				Type:        FieldTypeBool,
				Value:       true,
				Description: "Format code on save",
				IsSet:       true,
				Source:      "config file",
			},
			// Available options
			{
				Key:         "vim_mode",
				Type:        FieldTypeBool,
				Description: "Enable Vim key bindings",
				IsSet:       false,
				Source:      "available option",
			},
			{
				Key:         "soft_wrap",
				Type:        FieldTypeSelect,
				Options:     []string{"none", "editor_width", "preferred_line_length"},
				Description: "Text wrapping behavior",
				IsSet:       false,
				Source:      "available option",
			},
		}

		form.SetFields(fields)

		// Test that form can be built without errors
		require.NotPanics(t, func() {
			form.buildForm()
		})

		// Verify form is created
		assert.NotNil(t, form.form, "Form should be created successfully")
	})
}

func TestValidationScenarios(t *testing.T) {
	t.Parallel()
	form := NewHuhConfigForm("test-app")

	t.Run("required_field_validation", func(t *testing.T) {
		field := ConfigField{
			Key:      "api_key",
			Type:     FieldTypeString,
			Required: true,
			IsSet:    false,
			Source:   "available option",
		}

		form.SetFields([]ConfigField{field})
		huhField := form.createHuhField(field)
		assert.NotNil(t, huhField, "Required field should be created")
	})

	t.Run("numeric_validation", func(t *testing.T) {
		minVal := float64(1)
		maxVal := float64(100)

		field := ConfigField{
			Key:    "percentage",
			Type:   FieldTypeInt,
			Min:    &minVal,
			Max:    &maxVal,
			IsSet:  false,
			Source: "available option",
		}

		form.SetFields([]ConfigField{field})
		huhField := form.createHuhField(field)
		assert.NotNil(t, huhField, "Field with numeric validation should be created")
	})
}

// TestConfigFormPerformance tests that form creation and updates are fast
func TestConfigFormPerformance(t *testing.T) {
	t.Parallel()
	// Create a large number of fields to test performance
	var fields []ConfigField
	for i := 0; i < 100; i++ {
		fields = append(fields, ConfigField{
			Key:         fmt.Sprintf("field_%d", i),
			Type:        FieldTypeString,
			Value:       fmt.Sprintf("value_%d", i),
			Description: fmt.Sprintf("Description for field %d", i),
			IsSet:       i%2 == 0, // Alternate between set and unset
			Source:      "config file",
		})
	}

	form := NewHuhConfigForm("performance-test")

	// Test that setting many fields doesn't cause performance issues
	start := time.Now()
	form.SetFields(fields)
	duration := time.Since(start)

	// Should complete within reasonable time (less than 500ms for 100 fields)
	assert.Less(t, duration.Milliseconds(), int64(500), "Form creation should be fast")

	groups := form.groupFields()
	assert.Greater(t, len(groups), 0, "Should create at least one group")
}

func TestCalculateLayout(t *testing.T) {
	t.Parallel()
	form := NewHuhConfigForm("test")
	form.SetSize(50, 20)
	w, h, cols := form.calculateLayout()
	assert.GreaterOrEqual(t, w, 40)
	assert.GreaterOrEqual(t, h, 10)
	assert.Equal(t, 1, cols)

	form.SetSize(200, 60)
	w, h, cols = form.calculateLayout()
	assert.Equal(t, 2, cols)
	assert.Greater(t, w, 100)
	assert.Greater(t, h, 10)
}
