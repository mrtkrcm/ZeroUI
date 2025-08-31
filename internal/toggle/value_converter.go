package toggle

import (
	"strconv"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// ValueConverter handles value conversion and type checking
type ValueConverter struct{}

// NewValueConverter creates a new value converter
func NewValueConverter() *ValueConverter {
	return &ValueConverter{}
}

// ConvertValue converts a string value to the appropriate type
func (vc *ValueConverter) ConvertValue(value string, fieldConfig *config.FieldConfig) (interface{}, error) {
	return vc.convertValueByType(value, fieldConfig.Type)
}

// convertValueByType converts a value based on its type
func (vc *ValueConverter) convertValueByType(value, fieldType string) (interface{}, error) {
	switch strings.ToLower(fieldType) {
	case "bool", "boolean":
		return vc.convertToBool(value)
	case "int", "integer":
		return vc.convertToInt(value)
	case "float", "number":
		return vc.convertToFloat(value)
	case "string", "":
		return value, nil
	default:
		// Default to string for unknown types
		return value, nil
	}
}

// convertToBool converts string to boolean
func (vc *ValueConverter) convertToBool(value string) (bool, error) {
	lowered := strings.ToLower(value)
	switch lowered {
	case "true", "1", "yes", "on", "enabled":
		return true, nil
	case "false", "0", "no", "off", "disabled":
		return false, nil
	default:
		return false, errors.New(errors.FieldInvalidType, "invalid boolean value: "+value)
	}
}

// convertToInt converts string to integer
func (vc *ValueConverter) convertToInt(value string) (int, error) {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(errors.FieldInvalidType, "invalid integer value: "+value)
	}
	return result, nil
}

// convertToFloat converts string to float
func (vc *ValueConverter) convertToFloat(value string) (float64, error) {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, errors.New(errors.FieldInvalidType, "invalid float value: "+value)
	}
	return result, nil
}

// GetNextValue returns the next value in a field's cycle
func (vc *ValueConverter) GetNextValue(fieldConfig *config.FieldConfig, currentValue string) (string, error) {
	if len(fieldConfig.Values) == 0 {
		return "", errors.New(errors.FieldInvalidType, "field has no predefined values to cycle through")
	}

	// Find current value index
	currentIndex := -1
	for i, value := range fieldConfig.Values {
		if value == currentValue {
			currentIndex = i
			break
		}
	}

	// Get next value (wrap around if needed)
	nextIndex := (currentIndex + 1) % len(fieldConfig.Values)
	return fieldConfig.Values[nextIndex], nil
}