package legacyextractor

import (
	"fmt"
	"strconv"
	"strings"
)

// Validator provides simple, rule-based validation for configs
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines how to validate a setting
type ValidationRule interface {
	Validate(value interface{}) (bool, string)
}

// ValidationResult contains validation outcome
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warning []string
}

// NewValidator creates a simple validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string]ValidationRule),
	}
}

// AddRule adds a validation rule for a setting
func (v *Validator) AddRule(setting string, rule ValidationRule) {
	v.rules[setting] = rule
}

// Validate checks a single value
func (v *Validator) Validate(setting string, value interface{}) (bool, string) {
	rule, exists := v.rules[setting]
	if !exists {
		return true, "" // No rule = valid
	}
	return rule.Validate(value)
}

// ValidateConfig validates an entire config
func (v *Validator) ValidateConfig(config *Config) ValidationResult {
	result := ValidationResult{Valid: true}

	for name, setting := range config.Settings {
		if rule, exists := v.rules[name]; exists {
			if valid, msg := rule.Validate(setting.Default); !valid {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", name, msg))
			}
		}
	}

	return result
}

// Built-in validation rules

// StringRule validates string settings
type StringRule struct {
	Required  bool
	MinLength int
	MaxLength int
	Pattern   string
}

func (r StringRule) Validate(value interface{}) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, "must be a string"
	}

	if r.Required && str == "" {
		return false, "is required"
	}

	if r.MinLength > 0 && len(str) < r.MinLength {
		return false, fmt.Sprintf("must be at least %d characters", r.MinLength)
	}

	if r.MaxLength > 0 && len(str) > r.MaxLength {
		return false, fmt.Sprintf("must be at most %d characters", r.MaxLength)
	}

	if r.Pattern != "" && !strings.Contains(str, r.Pattern) {
		return false, fmt.Sprintf("must contain '%s'", r.Pattern)
	}

	return true, ""
}

// NumberRule validates numeric settings
type NumberRule struct {
	Required bool
	Min      *float64
	Max      *float64
}

func (r NumberRule) Validate(value interface{}) (bool, string) {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		if v == "" && !r.Required {
			return true, ""
		}
		// Try to parse string number
		if !isNumeric(v) {
			return false, "must be a number"
		}
		// Parse the numeric string to validate against min/max
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			num = parsed
		} else {
			return false, "must be a number"
		}
	default:
		return false, "must be a number"
	}

	if r.Min != nil && num < *r.Min {
		return false, fmt.Sprintf("must be at least %v", *r.Min)
	}

	if r.Max != nil && num > *r.Max {
		return false, fmt.Sprintf("must be at most %v", *r.Max)
	}

	return true, ""
}

// BooleanRule validates boolean settings
type BooleanRule struct {
	Required bool
}

func (r BooleanRule) Validate(value interface{}) (bool, string) {
	switch v := value.(type) {
	case bool:
		return true, ""
	case string:
		if v == "true" || v == "false" || (!r.Required && v == "") {
			return true, ""
		}
		return false, "must be true or false"
	default:
		if !r.Required && value == nil {
			return true, ""
		}
		return false, "must be a boolean"
	}
}

// ChoiceRule validates enum/choice settings
type ChoiceRule struct {
	Required bool
	Choices  []string
}

func (r ChoiceRule) Validate(value interface{}) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, "must be a string choice"
	}

	if !r.Required && str == "" {
		return true, ""
	}

	for _, choice := range r.Choices {
		if str == choice {
			return true, ""
		}
	}

	return false, fmt.Sprintf("must be one of: %s", strings.Join(r.Choices, ", "))
}

// Factory functions for common rules

// Min creates a minimum value
func Min(v float64) *float64 {
	return &v
}

// Max creates a maximum value
func Max(v float64) *float64 {
	return &v
}

// Required creates a required string rule
func Required() ValidationRule {
	return StringRule{Required: true}
}

// Choice creates a choice validation rule
func Choice(choices ...string) ValidationRule {
	return ChoiceRule{Choices: choices}
}
