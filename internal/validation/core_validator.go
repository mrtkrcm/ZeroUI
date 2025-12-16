package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
)

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := &Validator{
		schemas:  make(map[string]*Schema),
		validate: validator.New(),
	}

	// Register custom validators
	v.validate.RegisterValidation("color", validateColorTag)
	v.validate.RegisterValidation("pathformat", validatePathFormatTag)
	v.validate.RegisterValidation("regex", validateRegexTag)
	v.validate.RegisterValidation("fieldtype", validateFieldTypeTag)

	return v
}

// ValidateAppConfig validates an entire application configuration
func (v *Validator) ValidateAppConfig(appName string, appConfig *appconfig.AppConfig) *ValidationResult {
	// Check if schema exists
	schema, ok := v.schemas[appName]
	if !ok {
		// No schema, use basic validation
		return v.validateBasic(appConfig)
	}

	// Use optimized validation for simple schemas
	if v.isSimpleSchema(schema) {
		return v.validateAppConfigFast(appConfig, schema)
	}

	// Full validation with schema
	return v.validateAppConfigWithSchema(appConfig, schema)
}

// ValidateTargetConfig validates a target configuration (loaded config file)
func (v *Validator) ValidateTargetConfig(appName string, configData map[string]interface{}) *ValidationResult {
	// Check if schema exists
	schema, ok := v.schemas[appName]
	if !ok {
		// No schema, use basic validation
		return v.validateBasicConfig(configData)
	}

	// Validate with schema
	return v.validateConfigWithSchema(configData, schema)
}

// validateBasic performs basic validation without a schema
func (v *Validator) validateBasic(appConfig *appconfig.AppConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate using struct tags
	validatedConfig := v.convertToValidatedAppConfig(appConfig)
	if err := v.validate.Struct(validatedConfig); err != nil {
		return v.convertValidatorError(err, "app_config")
	}

	// Basic field validation
	for fieldName, field := range appConfig.Fields {
		// Check field type is valid
		if !isValidFieldType(field.Type) {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("invalid field type: %s", field.Type),
				Code:    "invalid_field_type",
			})
		}

		// Check default value type matches field type
		if field.Default != nil {
			if !v.validateFieldType(fieldName, field.Default, field.Type) {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("default value type mismatch for field type %s", field.Type),
					Code:    "default_type_mismatch",
					Value:   field.Default,
				})
			}
		}
	}

	// Validate presets
	for presetName, preset := range appConfig.Presets {
		for fieldName, value := range preset.Values {
			if field, ok := appConfig.Fields[fieldName]; ok {
				if !v.validateFieldType(fieldName, value, field.Type) {
					result.Valid = false
					result.Errors = append(result.Errors, &ValidationError{
						Field:   fmt.Sprintf("preset.%s.%s", presetName, fieldName),
						Message: fmt.Sprintf("value type mismatch for field type %s", field.Type),
						Code:    "preset_type_mismatch",
						Value:   value,
					})
				}
			}
		}
	}

	return result
}

// validateBasicConfig performs basic validation on config data
func (v *Validator) validateBasicConfig(configData map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	for key, value := range configData {
		if !isValidConfigValue(value) {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Field:   key,
				Message: fmt.Sprintf("invalid value type: %T", value),
				Code:    "invalid_type",
				Value:   value,
			})
		}
	}

	return result
}

// validateConfigWithSchema validates config data against a schema
func (v *Validator) validateConfigWithSchema(configData map[string]interface{}, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate each field
	for fieldName, value := range configData {
		rule, ok := schema.Fields[fieldName]
		if !ok {
			// Field not in schema - check if forbidden
			isForbidden := false
			if schema.Global != nil && len(schema.Global.ForbiddenFields) > 0 {
				for _, forbidden := range schema.Global.ForbiddenFields {
					if fieldName == forbidden {
						isForbidden = true
						result.Valid = false
						result.Errors = append(result.Errors, &ValidationError{
							Field:   fieldName,
							Message: "field is forbidden",
							Code:    "forbidden_field",
						})
						break
					}
				}
			}
			// If the field is not explicitly forbidden, emit a warning so callers
			// are informed that an undefined field was present but do not fail the
			// entire validation. Tests expect an 'undefined_field' warning.
			if !isForbidden {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field:   fieldName,
					Message: "field is not defined in schema",
					Code:    "undefined_field",
					Value:   value,
				})
			}
			continue
		}

		// Validate with rule
		fieldResult := v.validateFieldWithRule(fieldName, value, rule)
		if !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
		}
	}

	// Check required fields
	for fieldName, rule := range schema.Fields {
		if rule.Required {
			if _, ok := configData[fieldName]; !ok {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: "required field is missing",
					Code:    "missing_required_field",
				})
			}
		}
	}

	// Check dependencies
	for fieldName, value := range configData {
		if rule, ok := schema.Fields[fieldName]; ok && value != nil {
			for _, dep := range rule.Dependencies {
				if _, ok := configData[dep]; !ok {
					result.Valid = false
					result.Errors = append(result.Errors, &ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("depends on field: %s", dep),
						Code:    "missing_dependency",
					})
				}
			}

			// Check conflicts
			for _, conflict := range rule.ConflictsWith {
				if _, ok := configData[conflict]; ok {
					result.Valid = false
					result.Errors = append(result.Errors, &ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("conflicts with field: %s", conflict),
						Code:    "field_conflict",
					})
				}
			}
		}
	}

	// Validate global rules
	if schema.Global != nil {
		globalErrors := v.validateGlobalRules(configData, schema.Global)
		if len(globalErrors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, globalErrors...)
		}
	}

	return result
}

// validateGlobalRules validates global configuration rules
func (v *Validator) validateGlobalRules(configData interface{}, global *GlobalRules) []*ValidationError {
	var errors []*ValidationError

	// Convert to map if it's an AppConfig
	var data map[string]interface{}
	switch c := configData.(type) {
	case *appconfig.AppConfig:
		data = make(map[string]interface{})
		for k, field := range c.Fields {
			data[k] = field.Default
		}
	case map[string]interface{}:
		data = c
	default:
		return errors
	}

	// Check min/max fields
	fieldCount := len(data)
	if global.MinFields != nil && fieldCount < *global.MinFields {
		errors = append(errors, &ValidationError{
			Field:   "_global",
			Message: fmt.Sprintf("minimum %d fields required, got %d", *global.MinFields, fieldCount),
			Code:    "too_few_fields",
		})
	}

	if global.MaxFields != nil && fieldCount > *global.MaxFields {
		errors = append(errors, &ValidationError{
			Field:   "_global",
			Message: fmt.Sprintf("maximum %d fields allowed, got %d", *global.MaxFields, fieldCount),
			Code:    "too_many_fields",
		})
	}

	// Check required fields
	for _, required := range global.RequiredFields {
		if _, ok := data[required]; !ok {
			errors = append(errors, &ValidationError{
				Field:   required,
				Message: "globally required field is missing",
				Code:    "global_required",
			})
		}
	}

	// Check forbidden fields
	for _, forbidden := range global.ForbiddenFields {
		if _, ok := data[forbidden]; ok {
			errors = append(errors, &ValidationError{
				Field:   forbidden,
				Message: "globally forbidden field is present",
				Code:    "global_forbidden",
			})
		}
	}

	return errors
}

// validateAppConfigWithSchema validates app config with full schema
func (v *Validator) validateAppConfigWithSchema(appConfig *appconfig.AppConfig, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate each field definition
	for fieldName, field := range appConfig.Fields {
		if rule, ok := schema.Fields[fieldName]; ok {
			fieldResult := v.validateFieldDefinition(fieldName, &field, rule)
			if !fieldResult.Valid {
				result.Valid = false
				result.Errors = append(result.Errors, fieldResult.Errors...)
			}
		}
	}

	// Check required fields in schema
	for fieldName, rule := range schema.Fields {
		if rule.Required {
			if _, ok := appConfig.Fields[fieldName]; !ok {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: "required field is not defined",
					Code:    "missing_required_field",
				})
			}
		}
	}

	// Validate presets against schema
	for presetName, preset := range appConfig.Presets {
		for fieldName, value := range preset.Values {
			if rule, ok := schema.Fields[fieldName]; ok {
				fieldResult := v.validateFieldWithRule(
					fmt.Sprintf("preset.%s.%s", presetName, fieldName),
					value,
					rule,
				)
				if !fieldResult.Valid {
					result.Valid = false
					result.Errors = append(result.Errors, fieldResult.Errors...)
				}
			}
		}
	}

	// Validate global rules
	if schema.Global != nil {
		globalErrors := v.validateGlobalRules(appConfig, schema.Global)
		if len(globalErrors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, globalErrors...)
		}
	}

	return result
}

// validateAppConfigFast performs optimized validation for simple schemas
func (v *Validator) validateAppConfigFast(appConfig *appconfig.AppConfig, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Fast path: check field types and required fields
	for fieldName, rule := range schema.Fields {
		field, exists := appConfig.Fields[fieldName]

		if rule.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Message: "required field is missing",
				Code:    "required",
			})
			continue
		}

		if exists {
			// Quick type check
			if field.Type != rule.Type {
				result.Valid = false
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("type mismatch: expected %s, got %s", rule.Type, field.Type),
					Code:    "type_mismatch",
				})
			}

			// Validate enum values if present
			if rule.enumMap != nil && len(field.Values) > 0 {
				for _, val := range field.Values {
					if _, ok := rule.enumMap[val]; !ok {
						result.Valid = false
						result.Errors = append(result.Errors, &ValidationError{
							Field:   fieldName,
							Message: fmt.Sprintf("invalid enum value: %s", val),
							Code:    "invalid_enum",
							Value:   val,
						})
					}
				}
			}
		}
	}

	return result
}

// convertToValidatedAppConfig converts AppConfig to ValidatedAppConfig
func (v *Validator) convertToValidatedAppConfig(appConfig *appconfig.AppConfig) ValidatedAppConfig {
	validated := ValidatedAppConfig{
		Name:        appConfig.Name,
		Path:        appConfig.Path,
		Format:      appConfig.Format,
		Description: appConfig.Description,
		Fields:      make(map[string]ValidatedFieldConfig),
		Presets:     make(map[string]ValidatedPresetConfig),
		Hooks:       appConfig.Hooks,
		Env:         appConfig.Env,
	}

	for name, field := range appConfig.Fields {
		validated.Fields[name] = ValidatedFieldConfig{
			Type:        field.Type,
			Values:      field.Values,
			Default:     field.Default,
			Description: field.Description,
			Path:        field.Path,
		}
	}

	for name, preset := range appConfig.Presets {
		validated.Presets[name] = ValidatedPresetConfig{
			Name:        preset.Name,
			Description: preset.Description,
			Values:      preset.Values,
		}
	}

	return validated
}

// convertValidatorError converts go-playground/validator errors to ValidationResult
func (v *Validator) convertValidatorError(err error, context string) *ValidationResult {
	result := &ValidationResult{Valid: false}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   strings.ToLower(e.Field()),
				Message: v.getValidationErrorMessage(e),
				Code:    v.getValidationErrorCode(e.Tag()),
			})
		}
	} else {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   context,
			Message: err.Error(),
			Code:    "validation_error",
		})
	}

	return result
}

// getValidationErrorMessage generates user-friendly error messages
func (v *Validator) getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("minimum length is %s", err.Param())
	case "max":
		return fmt.Sprintf("maximum length is %s", err.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())
	case "dive":
		return "nested validation failed"
	case "fieldtype":
		return "invalid field type"
	case "color":
		return "invalid color format"
	case "pathformat":
		return "invalid path format"
	default:
		return fmt.Sprintf("validation failed: %s", err.Tag())
	}
}

// getValidationErrorCode returns error codes for validation errors
func (v *Validator) getValidationErrorCode(tag string) string {
	switch tag {
	case "required":
		return "required"
	case "min":
		return "min_length"
	case "max":
		return "max_length"
	case "oneof":
		return "invalid_choice"
	case "dive":
		return "nested_validation"
	case "fieldtype":
		return "invalid_type"
	case "color":
		return "invalid_color"
	case "pathformat":
		return "invalid_path"
	default:
		return "validation_error"
	}
}

// Custom validation tag functions

func validateColorTag(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	return validateColor(color) == nil
}

func validatePathFormatTag(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return validatePath(path) == nil
}

func validateRegexTag(fl validator.FieldLevel) bool {
	pattern := fl.Field().String()
	_, err := regexp.Compile(pattern)
	return err == nil
}

func validateFieldTypeTag(fl validator.FieldLevel) bool {
	fieldType := fl.Field().String()
	return isValidFieldType(fieldType)
}
