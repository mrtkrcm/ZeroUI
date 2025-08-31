package toggle

import (
	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// FieldValidator handles field validation logic
type FieldValidator struct{}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{}
}

// ValidateFieldExists checks if a field exists in the app config
func (fv *FieldValidator) ValidateFieldExists(appConfig *config.AppConfig, key string) error {
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
func (fv *FieldValidator) ValidateFieldValue(appConfig *config.AppConfig, key, value string) error {
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
func (fv *FieldValidator) GetFieldConfig(appConfig *config.AppConfig, key string) (*config.FieldConfig, error) {
	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		return nil, errors.NewFieldNotFoundError(appConfig.Name, key, nil)
	}
	return &fieldConfig, nil
}