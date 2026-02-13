package validation

import (
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// FieldAdapter provides an abstraction for field-level validation logic
type FieldAdapter struct{}

// NewFieldAdapter creates a new field adapter
func NewFieldAdapter() *FieldAdapter {
	return &FieldAdapter{}
}

// ValidateFieldExists checks if a field exists in the app config
func (fa *FieldAdapter) ValidateFieldExists(appConfig *appconfig.AppConfig, key string) error {
	if _, exists := appConfig.Fields[key]; !exists {
		var availableFields []string
		for field := range appConfig.Fields {
			availableFields = append(availableFields, field)
		}
		return errors.NewFieldNotFoundError(appConfig.Name, key, availableFields)
	}
	return nil
}

// ValidateFieldValue checks if a value is valid for a field
func (fa *FieldAdapter) ValidateFieldValue(appConfig *appconfig.AppConfig, key, value string) error {
	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		return errors.NewFieldNotFoundError(appConfig.Name, key, nil)
	}

	// Check if field has value constraints
	if len(fieldConfig.Values) > 0 {
		for _, validValue := range fieldConfig.Values {
			if validValue == value {
				return nil // Valid value found
			}
		}
		return errors.NewInvalidValueError(appConfig.Name, key, value, fieldConfig.Values)
	}

	return nil // No constraints, value is valid
}

// GetFieldConfig returns the field configuration for validation
func (fa *FieldAdapter) GetFieldConfig(appConfig *appconfig.AppConfig, key string) (*appconfig.FieldConfig, error) {
	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		return nil, errors.NewFieldNotFoundError(appConfig.Name, key, nil)
	}
	return &fieldConfig, nil
}
