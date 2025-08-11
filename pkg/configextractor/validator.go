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
	if !v.validateType(value, rule.Type) {
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
		if num, err := v.toFloat64(value); err == nil {
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
func (v *Validator) validateType(value interface{}, expectedType SettingType) bool {
	switch expectedType {
	case TypeString:
		_, ok := value.(string)
		return ok
	case TypeNumber:
		_, err := v.toFloat64(value)
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

// toFloat64 converts various numeric types to float64
func (v *Validator) toFloat64(value interface{}) (float64, error) {
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

// DefaultValidationRules returns common validation rules for typical apps
func DefaultValidationRules() map[string]ValidationRule {
	return map[string]ValidationRule{
		// Font settings
		"font-family": StringRule(false, "", "monospace", "serif", "sans-serif"),
		"font-size":   NumberRule(false, Min(8), Max(72)),
		
		// Color settings
		"background":  StringRule(false, `^#[0-9A-Fa-f]{6}$`),
		"foreground":  StringRule(false, `^#[0-9A-Fa-f]{6}$`),
		
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