package configextractor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator provides lightweight configuration validation
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines simple validation rules for settings
type ValidationRule struct {
	Type     SettingType `json:"type"`
	Required bool        `json:"required,omitempty"`
	Min      *float64    `json:"min,omitempty"`
	Max      *float64    `json:"max,omitempty"`
	Pattern  string      `json:"pattern,omitempty"`
	Values   []string    `json:"values,omitempty"` // Valid choices
}

// ValidationResult represents validation outcome
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// NewValidator creates a minimal validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule adds a validation rule for a setting
func (v *Validator) AddRule(setting string, rule ValidationRule) {
	v.rules[setting] = rule
}

// Validate validates a configuration setting
func (v *Validator) Validate(setting string, value interface{}) ValidationResult {
	rule, exists := v.rules[setting]
	if !exists {
		// No rule = valid (permissive by default)
		return ValidationResult{Valid: true}
	}

	var errors []string

	// Type validation
	if !v.validateValueType(value, rule.Type) {
		errors = append(errors, fmt.Sprintf("%s must be of type %s", setting, rule.Type))
	}

	// Value-specific validations
	switch rule.Type {
	case TypeString:
		if str, ok := value.(string); ok {
			if rule.Pattern != "" {
				if matched, _ := regexp.MatchString(rule.Pattern, str); !matched {
					errors = append(errors, fmt.Sprintf("%s does not match required pattern", setting))
				}
			}
			if len(rule.Values) > 0 {
				valid := false
				for _, validValue := range rule.Values {
					if str == validValue {
						valid = true
						break
					}
				}
				if !valid {
					errors = append(errors, fmt.Sprintf("%s must be one of: %s", setting, strings.Join(rule.Values, ", ")))
				}
			}
		}

	case TypeNumber:
		if num, err := v.ToFloat64(value); err == nil {
			if rule.Min != nil && num < *rule.Min {
				errors = append(errors, fmt.Sprintf("%s must be at least %g", setting, *rule.Min))
			}
			if rule.Max != nil && num > *rule.Max {
				errors = append(errors, fmt.Sprintf("%s cannot exceed %g", setting, *rule.Max))
			}
		}

	case TypeChoice:
		if str := fmt.Sprintf("%v", value); len(rule.Values) > 0 {
			valid := false
			for _, validValue := range rule.Values {
				if str == validValue {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, fmt.Sprintf("%s must be one of: %s", setting, strings.Join(rule.Values, ", ")))
			}
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateConfig validates an entire configuration
func (v *Validator) ValidateConfig(config *Config) ValidationResult {
	var allErrors []string

	// Check required settings
	for setting, rule := range v.rules {
		if rule.Required {
			if _, exists := config.Settings[setting]; !exists {
				allErrors = append(allErrors, fmt.Sprintf("required setting %s is missing", setting))
			}
		}
	}

	// Validate existing settings
	for setting, settingConfig := range config.Settings {
		result := v.Validate(setting, settingConfig.Default)
		if !result.Valid {
			allErrors = append(allErrors, result.Errors...)
		}
	}

	return ValidationResult{
		Valid:  len(allErrors) == 0,
		Errors: allErrors,
	}
}

// validateType checks if value matches expected type
func (v *Validator) validateValueType(value interface{}, expectedType SettingType) bool {
	switch expectedType {
	case TypeString:
		_, ok := value.(string)
		return ok
	case TypeNumber:
		_, err := v.ToFloat64(value)
		return err == nil
	case TypeBoolean:
		if _, ok := value.(bool); ok {
			return true
		}
		// Also accept string boolean representations
		if str, ok := value.(string); ok {
			_, err := strconv.ParseBool(str)
			return err == nil
		}
		return false
	case TypeChoice:
		// Choices are typically strings
		_, ok := value.(string)
		return ok
	case TypeArray:
		// Accept any slice type
		switch value.(type) {
		case []interface{}, []string, []int, []float64, []bool:
			return true
		default:
			return false
		}
	default:
		return true // Unknown types are valid
	}
}

// ToFloat64 converts various numeric types to float64
func (v *Validator) ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// Common validation rule factories

// StringRule creates a string validation rule
func StringRule(required bool, pattern string, values ...string) ValidationRule {
	rule := ValidationRule{
		Type:     TypeString,
		Required: required,
	}
	if pattern != "" {
		rule.Pattern = pattern
	}
	if len(values) > 0 {
		rule.Values = values
	}
	return rule
}

// NumberRule creates a number validation rule
func NumberRule(required bool, min, max *float64) ValidationRule {
	return ValidationRule{
		Type:     TypeNumber,
		Required: required,
		Min:      min,
		Max:      max,
	}
}

// BooleanRule creates a boolean validation rule
func BooleanRule(required bool) ValidationRule {
	return ValidationRule{
		Type:     TypeBoolean,
		Required: required,
	}
}

// ChoiceRule creates a choice validation rule
func ChoiceRule(required bool, values ...string) ValidationRule {
	return ValidationRule{
		Type:     TypeChoice,
		Required: required,
		Values:   values,
	}
}

// Helper functions for creating common rules

// Min creates a minimum value constraint
func Min(value float64) *float64 {
	return &value
}

// Max creates a maximum value constraint
func Max(value float64) *float64 {
	return &value
}

// GhosttySchemaValidator provides schema-aware validation for Ghostty configurations
type GhosttySchemaValidator struct {
	schema map[string]GhosttyFieldSchema
}

// GhosttyFieldSchema represents the schema for a Ghostty configuration field
type GhosttyFieldSchema struct {
	Name         string      `yaml:"name"`
	Type         string      `yaml:"type"`
	Description  string      `yaml:"description"`
	DefaultValue interface{} `yaml:"default_value"`
	Category     string      `yaml:"category"`
	Values       []string    `yaml:"values,omitempty"` // For enum types
}

// NewGhosttySchemaValidator creates a validator that validates against the official Ghostty schema
func NewGhosttySchemaValidator() *GhosttySchemaValidator {
	return &GhosttySchemaValidator{
		schema: LoadGhosttySchema(),
	}
}

// ValidateField validates a single field against the Ghostty schema
func (v *GhosttySchemaValidator) ValidateField(fieldName string, value interface{}) ValidationResult {
	fieldSchema, exists := v.schema[fieldName]
	if !exists {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("field '%s' is not a valid Ghostty configuration option", fieldName)},
		}
	}

	var errors []string

	// Type validation
	if !v.validateGhosttyType(value, fieldSchema.Type) {
		errors = append(errors, fmt.Sprintf("field '%s' must be of type %s", fieldName, fieldSchema.Type))
	}

	// Value validation for enum types
	if len(fieldSchema.Values) > 0 {
		if str := fmt.Sprintf("%v", value); !v.isValidEnum(str, fieldSchema.Values) {
			errors = append(errors, fmt.Sprintf("field '%s' must be one of: %s", fieldName, strings.Join(fieldSchema.Values, ", ")))
		}
	}

	// Special validation for specific field types
	switch fieldSchema.Name {
	case "cursor-color", "cursor-text":
		if str, ok := value.(string); ok && !v.isValidColorValue(str) {
			errors = append(errors, fmt.Sprintf("field '%s' must be a valid color (hex or named color)", fieldName))
		}
	case "font-family":
		if str, ok := value.(string); ok && strings.Contains(str, ",") {
			// Multiple font families should be space-separated, not comma-separated
			if !strings.Contains(str, ", ") {
				errors = append(errors, fmt.Sprintf("field '%s' should use spaces between font families, not commas", fieldName))
			}
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// ValidateConfig validates an entire Ghostty configuration
func (v *GhosttySchemaValidator) ValidateConfig(config map[string]interface{}) ValidationResult {
	var allErrors []string

	for fieldName, value := range config {
		result := v.ValidateField(fieldName, value)
		if !result.Valid {
			allErrors = append(allErrors, result.Errors...)
		}
	}

	return ValidationResult{
		Valid:  len(allErrors) == 0,
		Errors: allErrors,
	}
}

// validateGhosttyType validates the type of a value against Ghostty's expected types
func (v *GhosttySchemaValidator) validateGhosttyType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		if _, ok := value.(float64); ok {
			return true
		}
		if _, ok := value.(int); ok {
			return true
		}
		if str, ok := value.(string); ok {
			_, err := strconv.ParseFloat(str, 64)
			return err == nil
		}
		return false
	case "boolean":
		if _, ok := value.(bool); ok {
			return true
		}
		if str, ok := value.(string); ok {
			_, err := strconv.ParseBool(str)
			return err == nil
		}
		return false
	case "color":
		if str, ok := value.(string); ok {
			return v.isValidColorValue(str)
		}
		return false
	default:
		// For unknown types, accept string representation
		return true
	}
}

// isValidEnum checks if a value is in the list of allowed enum values
func (v *GhosttySchemaValidator) isValidEnum(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

// isValidColorValue validates color formats (hex or named colors)
func (v *GhosttySchemaValidator) isValidColorValue(color string) bool {
	color = strings.TrimSpace(color)

	// Check for hex colors
	if strings.HasPrefix(color, "#") {
		hex := color[1:]
		if len(hex) == 3 || len(hex) == 6 {
			for _, r := range hex {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
					return false
				}
			}
			return true
		}
	}

	// Check for named colors (basic validation - could be expanded)
	namedColors := []string{
		"black", "white", "red", "green", "blue", "yellow", "magenta", "cyan",
		"gray", "grey", "background", "foreground", "extend",
	}

	for _, named := range namedColors {
		if color == named {
			return true
		}
	}

	return false
}

// loadGhosttySchema loads the Ghostty schema from the embedded configuration
func LoadGhosttySchema() map[string]GhosttyFieldSchema {
	// This would ideally load from the actual ghostty.yaml file
	// For now, we'll define the most common invalid fields that we've seen
	return map[string]GhosttyFieldSchema{
		// Valid cursor fields
		"cursor-color": {
			Name:        "cursor-color",
			Type:        "color",
			Description: "The color of the cursor",
		},
		"cursor-invert-fg-bg": {
			Name:        "cursor-invert-fg-bg",
			Type:        "boolean",
			Description: "Swap foreground and background colors under cursor",
		},
		"cursor-opacity": {
			Name:        "cursor-opacity",
			Type:        "number",
			Description: "Opacity level of the cursor",
		},
		"cursor-style": {
			Name:        "cursor-style",
			Type:        "string",
			Description: "Style of the cursor",
			Values:      []string{"block", "bar", "underline", "outline"},
		},
		"cursor-style-blink": {
			Name:        "cursor-style-blink",
			Type:        "boolean",
			Description: "Whether cursor blinks",
		},
		"cursor-text": {
			Name:        "cursor-text",
			Type:        "color",
			Description: "Color of text under cursor",
		},
		"cursor-click-to-move": {
			Name:        "cursor-click-to-move",
			Type:        "boolean",
			Description: "Enable cursor click-to-move",
		},

		// Valid window padding fields
		"window-padding-x": {
			Name:        "window-padding-x",
			Type:        "number",
			Description: "Horizontal window padding",
		},
		"window-padding-y": {
			Name:        "window-padding-y",
			Type:        "number",
			Description: "Vertical window padding",
		},
		"window-padding-balance": {
			Name:        "window-padding-balance",
			Type:        "boolean",
			Description: "Balance padding dimensions",
		},
		"window-padding-color": {
			Name:        "window-padding-color",
			Type:        "color",
			Description: "Color of window padding area",
		},

		// Common font fields
		"font-family": {
			Name:        "font-family",
			Type:        "string",
			Description: "Font family for terminal text",
		},
		"font-size": {
			Name:        "font-size",
			Type:        "number",
			Description: "Font size",
		},

		// Common theme fields
		"theme": {
			Name:        "theme",
			Type:        "string",
			Description: "Color theme",
		},
		"background": {
			Name:        "background",
			Type:        "color",
			Description: "Background color",
		},
		"foreground": {
			Name:        "foreground",
			Type:        "color",
			Description: "Foreground color",
		},
	}
}

// ValidateGhosttyConfig is a convenience function for validating Ghostty configurations
func ValidateGhosttyConfig(config map[string]interface{}) ValidationResult {
	validator := NewGhosttySchemaValidator()
	return validator.ValidateConfig(config)
}

// ConfigDiff represents the differences between two configuration states
type ConfigDiff struct {
	Added     map[string]interface{} `json:"added,omitempty"`
	Modified  map[string]ValueDiff   `json:"modified,omitempty"`
	Removed   map[string]interface{} `json:"removed,omitempty"`
	Unchanged map[string]interface{} `json:"unchanged,omitempty"`
}

// ValueDiff represents a change in a single configuration value
type ValueDiff struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// ConfigDiffer provides functionality to compare configuration states
type ConfigDiffer struct{}

// NewConfigDiffer creates a new configuration differ
func NewConfigDiffer() *ConfigDiffer {
	return &ConfigDiffer{}
}

// DiffConfigurations compares two configuration maps and returns the differences
func (d *ConfigDiffer) DiffConfigurations(oldConfig, newConfig map[string]interface{}) *ConfigDiff {
	diff := &ConfigDiff{
		Added:     make(map[string]interface{}),
		Modified:  make(map[string]ValueDiff),
		Removed:   make(map[string]interface{}),
		Unchanged: make(map[string]interface{}),
	}

	// Track keys we've seen in new config
	seenKeys := make(map[string]bool)

	// Check for added and modified keys
	for key, newValue := range newConfig {
		seenKeys[key] = true
		if oldValue, exists := oldConfig[key]; exists {
			// Key exists in both - check if value changed
			if d.valuesEqual(oldValue, newValue) {
				diff.Unchanged[key] = newValue
			} else {
				diff.Modified[key] = ValueDiff{
					Old: oldValue,
					New: newValue,
				}
			}
		} else {
			// Key only in new config - it's added
			diff.Added[key] = newValue
		}
	}

	// Check for removed keys
	for key, oldValue := range oldConfig {
		if !seenKeys[key] {
			diff.Removed[key] = oldValue
		}
	}

	return diff
}

// valuesEqual compares two values for equality, handling different types appropriately
func (d *ConfigDiffer) valuesEqual(a, b interface{}) bool {
	// Handle nil values
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle arrays/slices
	if aSlice, ok := a.([]interface{}); ok {
		if bSlice, ok := b.([]interface{}); ok {
			if len(aSlice) != len(bSlice) {
				return false
			}
			for i := range aSlice {
				if !d.valuesEqual(aSlice[i], bSlice[i]) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Handle string arrays
	if aSlice, ok := a.([]string); ok {
		if bSlice, ok := b.([]string); ok {
			if len(aSlice) != len(bSlice) {
				return false
			}
			for i := range aSlice {
				if aSlice[i] != bSlice[i] {
					return false
				}
			}
			return true
		}
		return false
	}

	// For other types, use standard equality
	return a == b
}

// HasChanges returns true if the diff contains any changes
func (d *ConfigDiff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Modified) > 0 || len(d.Removed) > 0
}

// Summary returns a human-readable summary of the changes
func (d *ConfigDiff) Summary() string {
	var parts []string

	if len(d.Added) > 0 {
		parts = append(parts, fmt.Sprintf("+%d added", len(d.Added)))
	}
	if len(d.Modified) > 0 {
		parts = append(parts, fmt.Sprintf("~%d modified", len(d.Modified)))
	}
	if len(d.Removed) > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", len(d.Removed)))
	}
	if len(d.Unchanged) > 0 {
		parts = append(parts, fmt.Sprintf("=%d unchanged", len(d.Unchanged)))
	}

	if len(parts) == 0 {
		return "No changes"
	}

	return strings.Join(parts, ", ")
}

// FormatDiff returns a formatted string representation of the diff
func (d *ConfigDiff) FormatDiff() string {
	var result strings.Builder

	if len(d.Added) > 0 {
		result.WriteString("Added:\n")
		for key, value := range d.Added {
			result.WriteString(fmt.Sprintf("  + %s = %v\n", key, value))
		}
		result.WriteString("\n")
	}

	if len(d.Modified) > 0 {
		result.WriteString("Modified:\n")
		for key, diff := range d.Modified {
			result.WriteString(fmt.Sprintf("  ~ %s: %v → %v\n", key, diff.Old, diff.New))
		}
		result.WriteString("\n")
	}

	if len(d.Removed) > 0 {
		result.WriteString("Removed:\n")
		for key, value := range d.Removed {
			result.WriteString(fmt.Sprintf("  - %s = %v\n", key, value))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// KeybindValidator provides specialized validation for keybind configurations across multiple apps
type KeybindValidator struct {
	app string // The target application (ghostty, vscode, zed, etc.)
}

// NewKeybindValidator creates a new keybind validator for a specific app
func NewKeybindValidator() *KeybindValidator {
	return &KeybindValidator{}
}

// NewKeybindValidatorForApp creates a new keybind validator for a specific application
func NewKeybindValidatorForApp(app string) *KeybindValidator {
	return &KeybindValidator{app: strings.ToLower(app)}
}

// ValidateKeybind validates a single keybind string
func (v *KeybindValidator) ValidateKeybind(keybind string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	// Parse keybind based on application format
	keys, action, err := v.parseKeybindFormat(keybind)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	if keys == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind keys cannot be empty")
	}

	if action == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind action cannot be empty")
	}

	if !result.Valid {
		return result
	}

	// Validate key combination
	keyResult := v.validateKeyCombination(keys)
	if !keyResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, keyResult.Errors...)
	}

	// Validate action
	actionResult := v.validateAction(action)
	if !actionResult.Valid {
		result.Valid = false
		result.Errors = append(result.Errors, actionResult.Errors...)
	}

	// Include warnings from action validation
	if len(actionResult.Warnings) > 0 {
		result.Warnings = append(result.Warnings, actionResult.Warnings...)
	}

	result.Keys = keys
	result.Action = action
	result.ParsedKeybind = KeybindComponents{
		Keys:   keys,
		Action: action,
	}

	return result
}

// parseKeybindFormat parses keybind format based on the target application
func (v *KeybindValidator) parseKeybindFormat(keybind string) (keys, action string, err error) {
	switch v.app {
	case "ghostty":
		// Ghostty format: "[global:]keys=action[:arg]"
		if !strings.Contains(keybind, "=") {
			return "", "", fmt.Errorf("ghostty keybind must contain '=' separator")
		}
		parts := strings.SplitN(keybind, "=", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid ghostty keybind format")
		}
		keys = strings.TrimSpace(parts[0])
		action = strings.TrimSpace(parts[1])

		// Handle global prefix
		if strings.HasPrefix(keys, "global:") {
			keys = strings.TrimPrefix(keys, "global:")
		}

	case "vscode", "code":
		// VSCode format: "keys": "action" (JSON-like)
		if strings.Contains(keybind, `": "`) {
			parts := strings.SplitN(keybind, `": "`, 2)
			if len(parts) == 2 {
				keys = strings.Trim(strings.TrimSpace(parts[0]), `"`)
				action = strings.Trim(strings.TrimSpace(parts[1]), `"`)
			} else {
				return "", "", fmt.Errorf("invalid vscode keybind format")
			}
		} else {
			// Fallback to ghostty format
			return v.parseKeybindFormat(keybind)
		}

	case "zed":
		// Zed format: similar to ghostty but may use different action names
		if !strings.Contains(keybind, "=") {
			return "", "", fmt.Errorf("zed keybind must contain '=' separator")
		}
		parts := strings.SplitN(keybind, "=", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid zed keybind format")
		}
		keys = strings.TrimSpace(parts[0])
		action = strings.TrimSpace(parts[1])

	default:
		// Default to ghostty format
		if !strings.Contains(keybind, "=") {
			return "", "", fmt.Errorf("keybind must contain '=' separator")
		}
		parts := strings.SplitN(keybind, "=", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid keybind format")
		}
		keys = strings.TrimSpace(parts[0])
		action = strings.TrimSpace(parts[1])
	}

	return keys, action, nil
}

// validateKeyCombination validates the key combination part of a keybind
func (v *KeybindValidator) validateKeyCombination(keys string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	// Common modifiers
	validModifiers := []string{
		"ctrl", "alt", "shift", "super", "cmd", "meta", "hyper",
		"primary", "secondary",
	}

	// Special keys
	specialKeys := []string{
		"escape", "tab", "capslock", "backspace", "enter", "return",
		"space", "up", "down", "left", "right",
		"home", "end", "pageup", "pagedown", "page_up", "page_down", "insert", "delete",
		"equal", "plus", "minus", "comma", "period", "slash", "backslash",
		"semicolon", "apostrophe", "bracketleft", "bracketright",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	}

	lowerKeys := strings.ToLower(keys)
	keyParts := strings.Split(lowerKeys, "+")

	// Check for valid modifiers and keys
	hasValidComponent := false
	for _, part := range keyParts {
		part = strings.TrimSpace(part)

		// Check if it's a valid modifier
		isModifier := false
		for _, modifier := range validModifiers {
			if part == modifier {
				isModifier = true
				break
			}
		}

		// Check if it's a special key
		isSpecialKey := false
		for _, special := range specialKeys {
			if part == special {
				isSpecialKey = true
				break
			}
		}

		// Check if it's a single character
		isSingleChar := len(part) == 1 && part[0] >= 32 && part[0] <= 126

		// Check if it's a function key
		isFunctionKey := strings.HasPrefix(part, "f") && len(part) >= 2 && len(part) <= 3
		if isFunctionKey {
			numStr := part[1:]
			if num, err := strconv.Atoi(numStr); err != nil || num < 1 || num > 12 {
				isFunctionKey = false
			}
		}

		if !isModifier && !isSpecialKey && !isSingleChar && !isFunctionKey {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("invalid key component: '%s'", part))
		} else {
			hasValidComponent = true
		}
	}

	if !hasValidComponent {
		result.Valid = false
		result.Errors = append(result.Errors, "keybind must contain at least one valid key")
	}

	return result
}

// validateAction validates the action part of a keybind
func (v *KeybindValidator) validateAction(action string) KeybindValidationResult {
	result := KeybindValidationResult{Valid: true}

	// Handle action with arguments: "action:arg"
	var baseAction string
	if strings.Contains(action, ":") {
		parts := strings.SplitN(action, ":", 2)
		baseAction = strings.TrimSpace(parts[0])
	} else {
		baseAction = action
	}

	// Common valid actions
	validActions := []string{
		"copy", "paste", "cut", "select_all",
		"new_tab", "new_window", "close_tab", "close_window",
		"next_tab", "prev_tab", "goto_tab",
		"split_right", "split_down", "split_left", "split_up",
		"select_split_right", "select_split_down", "select_split_left", "select_split_up",
		"resize_split_right", "resize_split_down", "resize_split_left", "resize_split_up",
		"equalize_splits", "close_split",
		"increase_font_size", "decrease_font_size", "reset_font_size",
		"reload_config", "inspect", "show_inspector",
		"scroll_to_top", "scroll_to_bottom", "scroll_page_up", "scroll_page_down",
		"scroll_line_up", "scroll_line_down",
		"clear_screen", "reset_terminal",
		"toggle_fullscreen", "toggle_maximize",
		"quit", "minimize_window",
	}

	lowerAction := strings.ToLower(baseAction)
	actionValid := false
	for _, valid := range validActions {
		if lowerAction == valid {
			actionValid = true
			break
		}
	}

	// Allow custom actions (we can't validate all possible actions)
	if !actionValid {
		// Log a warning but don't fail validation for unknown actions
		result.Warnings = append(result.Warnings, fmt.Sprintf("unknown action: '%s' (may still be valid)", baseAction))
	}

	return result
}

// KeybindValidationResult represents the result of keybind validation
type KeybindValidationResult struct {
	Valid         bool              `json:"valid"`
	Errors        []string          `json:"errors,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
	Keys          string            `json:"keys,omitempty"`
	Action        string            `json:"action,omitempty"`
	ParsedKeybind KeybindComponents `json:"parsed_keybind,omitempty"`
}

// KeybindComponents represents the parsed components of a keybind
type KeybindComponents struct {
	Keys   string `json:"keys"`
	Action string `json:"action"`
}

// DefaultValidationRules returns common validation rules for typical apps
func DefaultValidationRules() map[string]ValidationRule {
	return map[string]ValidationRule{
		// Font settings
		"font-family": StringRule(false, "", "monospace", "serif", "sans-serif"),
		"font-size":   NumberRule(false, Min(8), Max(72)),

		// Color settings
		"background": StringRule(false, `^#[0-9A-Fa-f]{6}$`),
		"foreground": StringRule(false, `^#[0-9A-Fa-f]{6}$`),

		// Window settings
		"window-width":  NumberRule(false, Min(100), Max(3000)),
		"window-height": NumberRule(false, Min(100), Max(2000)),

		// Boolean settings
		"mouse-support": BooleanRule(false),
		"confirm-quit":  BooleanRule(false),

		// Common choices
		"cursor-shape": ChoiceRule(false, "block", "underline", "bar"),
		"shell":        StringRule(false, ""),
	}
}
