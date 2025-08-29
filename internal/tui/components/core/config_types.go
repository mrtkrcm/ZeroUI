package core

import (
	"fmt"
	"strconv"
	"time"
)

// ConfigValue represents a strongly-typed configuration value
type ConfigValue struct {
	Key         string      `json:"key"`
	Type        ValueType   `json:"type"`
	StringVal   *string     `json:"string_value,omitempty"`
	IntVal      *int        `json:"int_value,omitempty"`
	FloatVal    *float64    `json:"float_value,omitempty"`
	BoolVal     *bool       `json:"bool_value,omitempty"`
	DurationVal *time.Duration `json:"duration_value,omitempty"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     *ConfigValue `json:"default,omitempty"`
}

// ValueType represents the type of a configuration value
type ValueType string

const (
	StringType   ValueType = "string"
	IntType      ValueType = "int"
	FloatType    ValueType = "float"
	BoolType     ValueType = "bool"
	DurationType ValueType = "duration"
)

// ConfigData represents a collection of typed configuration values
type ConfigData map[string]ConfigValue

// NewConfigData creates a new empty configuration data collection
func NewConfigData() ConfigData {
	return make(ConfigData)
}

// Set sets a configuration value with type safety
func (cd ConfigData) Set(key string, value ConfigValue) {
	cd[key] = value
}

// Get retrieves a configuration value by key
func (cd ConfigData) Get(key string) (ConfigValue, bool) {
	value, exists := cd[key]
	return value, exists
}

// GetString retrieves a string value with type safety
func (cd ConfigData) GetString(key string) (string, error) {
	value, exists := cd[key]
	if !exists {
		return "", fmt.Errorf("key %s not found", key)
	}
	if value.Type != StringType || value.StringVal == nil {
		return "", fmt.Errorf("key %s is not a string type", key)
	}
	return *value.StringVal, nil
}

// GetInt retrieves an integer value with type safety
func (cd ConfigData) GetInt(key string) (int, error) {
	value, exists := cd[key]
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}
	if value.Type != IntType || value.IntVal == nil {
		return 0, fmt.Errorf("key %s is not an int type", key)
	}
	return *value.IntVal, nil
}

// GetBool retrieves a boolean value with type safety
func (cd ConfigData) GetBool(key string) (bool, error) {
	value, exists := cd[key]
	if !exists {
		return false, fmt.Errorf("key %s not found", key)
	}
	if value.Type != BoolType || value.BoolVal == nil {
		return false, fmt.Errorf("key %s is not a bool type", key)
	}
	return *value.BoolVal, nil
}

// GetFloat retrieves a float value with type safety
func (cd ConfigData) GetFloat(key string) (float64, error) {
	value, exists := cd[key]
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}
	if value.Type != FloatType || value.FloatVal == nil {
		return 0, fmt.Errorf("key %s is not a float type", key)
	}
	return *value.FloatVal, nil
}

// GetDuration retrieves a duration value with type safety
func (cd ConfigData) GetDuration(key string) (time.Duration, error) {
	value, exists := cd[key]
	if !exists {
		return 0, fmt.Errorf("key %s not found", key)
	}
	if value.Type != DurationType || value.DurationVal == nil {
		return 0, fmt.Errorf("key %s is not a duration type", key)
	}
	return *value.DurationVal, nil
}

// SetString sets a string configuration value
func (cd ConfigData) SetString(key, value, description string, required bool) {
	cd[key] = ConfigValue{
		Key:         key,
		Type:        StringType,
		StringVal:   &value,
		Description: description,
		Required:    required,
	}
}

// SetInt sets an integer configuration value
func (cd ConfigData) SetInt(key string, value int, description string, required bool) {
	cd[key] = ConfigValue{
		Key:         key,
		Type:        IntType,
		IntVal:      &value,
		Description: description,
		Required:    required,
	}
}

// SetBool sets a boolean configuration value
func (cd ConfigData) SetBool(key string, value bool, description string, required bool) {
	cd[key] = ConfigValue{
		Key:         key,
		Type:        BoolType,
		BoolVal:     &value,
		Description: description,
		Required:    required,
	}
}

// SetFloat sets a float configuration value
func (cd ConfigData) SetFloat(key string, value float64, description string, required bool) {
	cd[key] = ConfigValue{
		Key:         key,
		Type:        FloatType,
		FloatVal:    &value,
		Description: description,
		Required:    required,
	}
}

// SetDuration sets a duration configuration value
func (cd ConfigData) SetDuration(key string, value time.Duration, description string, required bool) {
	cd[key] = ConfigValue{
		Key:         key,
		Type:        DurationType,
		DurationVal: &value,
		Description: description,
		Required:    required,
	}
}

// Validate validates all configuration values
func (cd ConfigData) Validate() []ValidationError {
	var errors []ValidationError
	
	for key, value := range cd {
		if value.Required && value.IsEmpty() {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: "required field is empty",
				Type:    "required",
			})
		}
		
		// Type-specific validation could be added here
	}
	
	return errors
}

// IsEmpty returns true if the configuration value is empty
func (cv ConfigValue) IsEmpty() bool {
	switch cv.Type {
	case StringType:
		return cv.StringVal == nil || *cv.StringVal == ""
	case IntType:
		return cv.IntVal == nil
	case FloatType:
		return cv.FloatVal == nil
	case BoolType:
		return cv.BoolVal == nil
	case DurationType:
		return cv.DurationVal == nil
	default:
		return true
	}
}

// String returns the string representation of the configuration value
func (cv ConfigValue) String() string {
	switch cv.Type {
	case StringType:
		if cv.StringVal != nil {
			return *cv.StringVal
		}
	case IntType:
		if cv.IntVal != nil {
			return strconv.Itoa(*cv.IntVal)
		}
	case FloatType:
		if cv.FloatVal != nil {
			return strconv.FormatFloat(*cv.FloatVal, 'f', -1, 64)
		}
	case BoolType:
		if cv.BoolVal != nil {
			return strconv.FormatBool(*cv.BoolVal)
		}
	case DurationType:
		if cv.DurationVal != nil {
			return cv.DurationVal.String()
		}
	}
	return ""
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Type    string
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ToLegacyMap converts ConfigData to map[string]interface{} for backward compatibility
func (cd ConfigData) ToLegacyMap() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range cd {
		switch value.Type {
		case StringType:
			if value.StringVal != nil {
				result[key] = *value.StringVal
			}
		case IntType:
			if value.IntVal != nil {
				result[key] = *value.IntVal
			}
		case FloatType:
			if value.FloatVal != nil {
				result[key] = *value.FloatVal
			}
		case BoolType:
			if value.BoolVal != nil {
				result[key] = *value.BoolVal
			}
		case DurationType:
			if value.DurationVal != nil {
				result[key] = *value.DurationVal
			}
		}
	}
	return result
}

// FromLegacyMap creates ConfigData from map[string]interface{} for backward compatibility
func FromLegacyMap(data map[string]interface{}) ConfigData {
	cd := NewConfigData()
	for key, value := range data {
		switch v := value.(type) {
		case string:
			cd.SetString(key, v, "", false)
		case int:
			cd.SetInt(key, v, "", false)
		case float64:
			cd.SetFloat(key, v, "", false)
		case bool:
			cd.SetBool(key, v, "", false)
		case time.Duration:
			cd.SetDuration(key, v, "", false)
		default:
			// For unknown types, convert to string
			cd.SetString(key, fmt.Sprintf("%v", v), "", false)
		}
	}
	return cd
}