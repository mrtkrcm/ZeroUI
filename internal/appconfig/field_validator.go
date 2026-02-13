package appconfig

import (
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// FieldValidator handles field validation logic
type FieldValidator struct{}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{}
}

// ValidateFieldExists checks if a field exists in the app config
func (fv *FieldValidator) ValidateFieldExists(appConfig *AppConfig, key string) error {
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
func (fv *FieldValidator) ValidateFieldValue(appConfig *AppConfig, key, value string) error {
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

// ValidateConfig validates all fields in the target config against the app config
func (fv *FieldValidator) ValidateConfig(appConfig *AppConfig, k *koanf.Koanf) error {
	// Iterate over all fields defined in AppConfig
	for key, fieldConfig := range appConfig.Fields {
		// If the key exists in the config, validate it
		// Note: k.Exists(key) works for nested keys if dot separator is used
		if k.Exists(key) {
			val := k.Get(key)
			if err := fv.validateValue(appConfig.Name, key, &fieldConfig, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fv *FieldValidator) validateValue(appName, key string, fieldConfig *FieldConfig, value interface{}) error {
	if len(fieldConfig.Values) == 0 {
		return nil
	}

	// Convert value to string for comparison with Values
	// This is a simplified comparison since Values are strings
	strValue := fmt.Sprintf("%v", value)

	for _, validValue := range fieldConfig.Values {
		if validValue == strValue {
			return nil
		}
	}

	return errors.NewInvalidValueError(appName, key, strValue, fieldConfig.Values)
}

// GetFieldConfig returns the field configuration for validation
func (fv *FieldValidator) GetFieldConfig(appConfig *AppConfig, key string) (*FieldConfig, error) {
	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		return nil, errors.NewFieldNotFoundError(appConfig.Name, key, nil)
	}
	return &fieldConfig, nil
}
