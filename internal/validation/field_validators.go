package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// Field type constants
const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeBoolean = "boolean"
	TypeChoice  = "choice"
	TypeArray   = "array"
)

// Format constants
const (
	FormatEmail = "email"
	FormatURL   = "url"
	FormatPath  = "path"
	FormatColor = "color"
)

// ValidateField validates a single field with its schema rule
func (v *Validator) ValidateField(appName string, fieldName string, value interface{}) *ValidationResult {

	// Check if schema exists
	schema, ok := v.schemas[appName]
	if !ok {
		// Try basic validation without schema
		return v.validateFieldBasic(fieldName, value)
	}

	// Get field rule
	rule, ok := schema.Fields[fieldName]
	if !ok {
		// Field not in schema - validate basic types but also emit a warning so callers
		// can detect undefined fields without failing the validation.
		result := v.validateFieldBasic(fieldName, value)

		// Append an undefined-field warning to the result so tests and callers see it.
		result.Warnings = append(result.Warnings, &ValidationError{
			Field:   fieldName,
			Message: "field is not defined in schema",
			Code:    "undefined_field",
			Value:   value,
		})
		return result
	}

	// Validate with rule
	return v.validateFieldWithRule(fieldName, value, rule)
}

// validateFieldBasic performs basic field validation without a schema
func (v *Validator) validateFieldBasic(fieldName string, value interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if !isValidConfigValue(value) {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("invalid value type: %T", value),
			// Use lower_snake_case codes to match existing tests/consumers
			Code:  "invalid_type",
			Value: value,
		})
	}

	return result
}

// validateFieldWithRule validates a field against a specific rule
func (v *Validator) validateFieldWithRule(fieldName string, value interface{}, rule *FieldRule) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check required
	if rule.Required && value == nil {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Message: "field is required",
			// normalized code
			Code: "required",
		})
		return result
	}

	// Allow nil for optional fields
	if value == nil {
		return result
	}

	// Type validation
	if !v.validateFieldType(fieldName, value, rule.Type) {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("expected type %s, got %T", rule.Type, value),
			// normalized code
			Code:  "type_mismatch",
			Value: value,
		})
		return result
	}

	// String-specific validations
	if rule.Type == TypeString {
		if err := v.validateStringField(fieldName, value, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err)
		}
	}

	// Number-specific validations
	if rule.Type == TypeNumber {
		if err := v.validateNumberField(fieldName, value, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err)
		}
	}

	// Choice validation
	if rule.Type == TypeChoice && len(rule.Enum) > 0 {
		if err := v.validateChoiceField(fieldName, value, rule); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err)
		}
	}

	// Custom validation
	if rule.Custom != nil {
		if err := v.validateCustom(fieldName, value, rule.Custom); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Message: err.Error(),
				Code:    "CUSTOM_VALIDATION",
				Value:   value,
			})
		}
	}

	return result
}

// validateStringField validates string-specific rules
func (v *Validator) validateStringField(fieldName string, value interface{}, rule *FieldRule) *ValidationError {
	str, ok := value.(string)
	if !ok {
		return nil // Type already validated
	}

	// Length validation
	if rule.MinLength != nil && len(str) < *rule.MinLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("minimum length is %d", *rule.MinLength),
			// normalized code expected by tests
			Code:  "too_short",
			Value: value,
		}
	}

	if rule.MaxLength != nil && len(str) > *rule.MaxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("maximum length is %d", *rule.MaxLength),
			// normalized code expected by tests
			Code:  "too_long",
			Value: value,
		}
	}

	// Pattern validation (use pre-compiled regex)
	if rule.compiledRegex != nil {
		if !rule.compiledRegex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("does not match pattern: %s", rule.Pattern),
				// normalized pattern code
				Code:  "pattern_mismatch",
				Value: value,
			}
		}
	}

	// Format validation
	if rule.Format != "" {
		if err := v.validateFormat(str, rule.Format); err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: err.Error(),
				// normalized format error code used in tests
				Code:  "invalid_format",
				Value: value,
			}
		}
	}

	return nil
}

// validateNumberField validates number-specific rules
func (v *Validator) validateNumberField(fieldName string, value interface{}, rule *FieldRule) *ValidationError {
	num, err := convertToFloat64(value)
	if err != nil {
		return nil // Type already validated
	}

	if rule.Min != nil && num < *rule.Min {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("minimum value is %v", *rule.Min),
			// normalized codes used by tests
			Code:  "too_small",
			Value: value,
		}
	}

	if rule.Max != nil && num > *rule.Max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("maximum value is %v", *rule.Max),
			Code:    "too_large",
			Value:   value,
		}
	}

	return nil
}

// validateChoiceField validates choice-specific rules
func (v *Validator) validateChoiceField(fieldName string, value interface{}, rule *FieldRule) *ValidationError {
	str := fmt.Sprintf("%v", value)

	// Use optimized enum map for O(1) lookup
	if rule.enumMap != nil {
		if _, ok := rule.enumMap[str]; !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be one of: %s", strings.Join(rule.Enum, ", ")),
				// normalized code
				Code:  "invalid_choice",
				Value: value,
			}
		}
	}

	return nil
}

// validateFieldDefinition validates a field configuration definition
func (v *Validator) validateFieldDefinition(fieldName string, field *config.FieldConfig, rule *FieldRule) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate field type
	if field.Type != rule.Type {
		result.Valid = false
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("field type mismatch: expected %s, got %s", rule.Type, field.Type),
			Code:    "type_mismatch",
		})
	}

	// Validate enum values match
	if rule.Type == TypeChoice && len(rule.Enum) > 0 {
		for _, val := range field.Values {
			if _, ok := rule.enumMap[val]; !ok {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("invalid choice value: %s", val),
					// align with choice error code
					Code:  "invalid_choice",
					Value: val,
				})
			}
		}
	}

	// Validate default value
	if field.Default != nil {
		if err := v.validateFieldWithRule(fieldName, field.Default, rule); !err.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, err.Errors...)
		}
	}

	return result
}

// validateFieldType checks if a value matches the expected type
func (v *Validator) validateFieldType(fieldName string, value interface{}, expectedType string) bool {
	switch expectedType {
	case TypeString:
		_, ok := value.(string)
		return ok
	case TypeNumber:
		_, err := convertToFloat64(value)
		return err == nil
	case TypeBoolean:
		_, ok := value.(bool)
		return ok
	case TypeChoice:
		// Choices are typically strings
		_, ok := value.(string)
		return ok
	case TypeArray:
		// Arrays can be various types
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			return true
		}
		return false
	default:
		return false
	}
}

// validateFormat validates string format specifications
func (v *Validator) validateFormat(value, format string) error {
	switch strings.ToLower(format) {
	case FormatEmail:
		return validateEmail(value)
	case FormatURL:
		return validateURL(value)
	case FormatPath:
		return validatePath(value)
	case FormatColor:
		return validateColor(value)
	case "regex", "pattern":
		// Treat value as a regex pattern to validate
		if _, err := regexp.Compile(value); err != nil {
			return errors.New("invalid regex pattern")
		}
		return nil
	default:
		return errors.New("unknown format: " + format)
	}
}

// validateCustom executes custom validation logic
func (v *Validator) validateCustom(fieldName string, value interface{}, custom *CustomRule) error {
	// Custom validation functions are handled here
	switch custom.Function {
	case "validate_positive":
		if num, err := convertToFloat64(value); err == nil {
			if num <= 0 {
				if custom.Message != "" {
					return errors.New(custom.Message)
				}
				return errors.New("value must be positive")
			}
		}
	case "validate_non_empty":
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			if custom.Message != "" {
				return errors.New(custom.Message)
			}
			return errors.New("value cannot be empty")
		}
	case "unique":
		// Delegate to package-level uniqueness helper implemented in uniqueness.go
		valStr := strings.TrimSpace(fmt.Sprintf("%v", value))
		if err := v.validateUniqueness(fieldName, valStr, custom.Args); err != nil {
			// If custom message is provided, prefer it
			if custom.Message != "" {
				return errors.New(custom.Message)
			}
			return err
		}
		return nil
	default:
		return errors.New("unknown custom validation function: " + custom.Function)
	}
	return nil
}

// Helper validation functions

func isValidFieldType(fieldType string) bool {
	switch fieldType {
	case TypeString, TypeNumber, TypeBoolean, TypeChoice, TypeArray:
		return true
	default:
		return false
	}
}

func isValidConfigValue(value interface{}) bool {
	switch value.(type) {
	case string, int, int64, float64, bool, []interface{}, []string, map[string]interface{}:
		return true
	default:
		return false
	}
}

func convertToFloat64(value interface{}) (float64, error) {
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

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must have scheme and host")
	}
	return nil
}

func validatePath(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}
	return nil
}

func validateColor(color string) error {
	c := strings.TrimSpace(color)
	if c == "" {
		return fmt.Errorf("invalid color format")
	}

	// Validate hex color
	if strings.HasPrefix(c, "#") {
		hex := strings.TrimPrefix(c, "#")
		if len(hex) != 3 && len(hex) != 6 {
			return fmt.Errorf("invalid hex color format")
		}
		for _, ch := range hex {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return fmt.Errorf("invalid hex color character")
			}
		}
		return nil
	}

	// Validate RGB/RGBA
	rgbPattern := regexp.MustCompile(`^rgba?\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*(,\s*[\d.]+\s*)?\)$`)
	if rgbPattern.MatchString(c) {
		return nil
	}

	// Support common named colors (basic set) - tests expect 'red' to be valid
	named := map[string]struct{}{
		"red": {}, "green": {}, "blue": {}, "white": {}, "black": {}, "yellow": {},
		"cyan": {}, "magenta": {}, "gray": {}, "grey": {}, "orange": {}, "purple": {},
	}
	if _, ok := named[strings.ToLower(c)]; ok {
		return nil
	}

	return fmt.Errorf("invalid color format")
}
