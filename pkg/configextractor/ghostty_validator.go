package configextractor

import (
	"fmt"
	"regexp"
	"strings"
)

// GhosttySchemaValidator validates Ghostty configuration fields
type GhosttySchemaValidator struct {
	validFields map[string]FieldSchema
}

// FieldSchema defines the schema for a configuration field
type FieldSchema struct {
	Type    string
	Enum    []string
	Pattern *regexp.Regexp
}

// NewGhosttySchemaValidator creates a new Ghostty schema validator
func NewGhosttySchemaValidator() *GhosttySchemaValidator {
	validator := &GhosttySchemaValidator{
		validFields: make(map[string]FieldSchema),
	}
	validator.initValidFields()
	return validator
}

// initValidFields initializes the valid Ghostty configuration fields
func (v *GhosttySchemaValidator) initValidFields() {
	// Color fields (hex colors or named colors)
	colorPattern := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$|^[a-zA-Z]+$`)

	// Define valid fields
	v.validFields = map[string]FieldSchema{
		"font-family":         {Type: "string"},
		"font-size":           {Type: "number"},
		"cursor-style":        {Type: "enum", Enum: []string{"block", "bar", "underline", "outline"}},
		"cursor-color":        {Type: "color", Pattern: colorPattern},
		"window-padding-x":    {Type: "number"},
		"window-padding-y":    {Type: "number"},
		"background":          {Type: "color", Pattern: colorPattern},
		"foreground":          {Type: "color", Pattern: colorPattern},
		"cursor-invert-fg-bg": {Type: "boolean"},
		// Add more fields as needed
	}
}

// ValidateField validates a single field
func (v *GhosttySchemaValidator) ValidateField(field string, value interface{}) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}}

	schema, exists := v.validFields[field]
	if !exists {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("field '%s' is not a valid Ghostty configuration option", field))
		return result
	}

	// Type validation
	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be of type string", field))
		}
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid number types
		default:
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be of type number", field))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be of type boolean", field))
		}
	case "enum":
		strVal, ok := value.(string)
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be of type string", field))
		} else {
			valid := false
			for _, enumVal := range schema.Enum {
				if strVal == enumVal {
					valid = true
					break
				}
			}
			if !valid {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be one of: %s", field, strings.Join(schema.Enum, ", ")))
			}
		}
	case "color":
		strVal, ok := value.(string)
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be of type color", field))
		} else if schema.Pattern != nil && !schema.Pattern.MatchString(strVal) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field '%s' must be a valid color (hex or named color)", field))
		}
	}

	return result
}

// ValidateConfig validates an entire configuration map
func (v *GhosttySchemaValidator) ValidateConfig(config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}}

	for field, value := range config {
		fieldResult := v.ValidateField(field, value)
		if !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	return result
}

// ValidateGhosttyConfig is a convenience function for validating Ghostty configs
func ValidateGhosttyConfig(config map[string]interface{}) ValidationResult {
	validator := NewGhosttySchemaValidator()
	return validator.ValidateConfig(config)
}
