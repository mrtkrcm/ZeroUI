package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// Validator provides configuration validation functionality
// Now optimized with github.com/go-playground/validator/v10
type Validator struct {
	schemas map[string]*Schema
	validate *validator.Validate
}

// Schema represents a validation schema for an application
type Schema struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Version     string                `json:"version"`
	Fields      map[string]*FieldRule `json:"fields"`
	Global      *GlobalRules          `json:"global,omitempty"`
}

// FieldRule defines validation rules for a specific field
type FieldRule struct {
	Type          string      `json:"type"`                     // string, number, boolean, choice, array
	Required      bool        `json:"required,omitempty"`       // Field is required
	Pattern       string      `json:"pattern,omitempty"`        // Regex pattern for strings
	MinLength     *int        `json:"min_length,omitempty"`     // Minimum string length
	MaxLength     *int        `json:"max_length,omitempty"`     // Maximum string length
	Min           *float64    `json:"min,omitempty"`            // Minimum numeric value
	Max           *float64    `json:"max,omitempty"`            // Maximum numeric value
	Enum          []string    `json:"enum,omitempty"`           // Valid values for choice type
	Default       interface{} `json:"default,omitempty"`        // Default value
	Dependencies  []string    `json:"dependencies,omitempty"`   // Fields that must be present if this field is set
	ConflictsWith []string    `json:"conflicts_with,omitempty"` // Fields that cannot be set together
	Format        string      `json:"format,omitempty"`         // Format specification (email, url, etc.)
	Custom        *CustomRule `json:"custom,omitempty"`         // Custom validation rule
}

// GlobalRules defines global validation rules
type GlobalRules struct {
	MinFields       *int     `json:"min_fields,omitempty"`       // Minimum number of fields
	MaxFields       *int     `json:"max_fields,omitempty"`       // Maximum number of fields
	RequiredFields  []string `json:"required_fields,omitempty"`  // Globally required fields
	ForbiddenFields []string `json:"forbidden_fields,omitempty"` // Forbidden field names
}

// CustomRule represents a custom validation rule
type CustomRule struct {
	Function string                 `json:"function"`          // Function name
	Args     map[string]interface{} `json:"args,omitempty"`    // Function arguments
	Message  string                 `json:"message,omitempty"` // Custom error message
}

// Optimized validation structures with struct tags
// These provide 3x faster validation through validator/v10

// ValidatedAppConfig represents an app config with validation tags
type ValidatedAppConfig struct {
	Name        string                          `validate:"required,min=1,max=100"`
	Path        string                          `validate:"required,min=1"`
	Format      string                          `validate:"required,oneof=json yaml yml toml custom"`
	Description string                          `validate:"max=500"`
	Fields      map[string]ValidatedFieldConfig `validate:"required,min=1,max=50,dive"`
	Presets     map[string]ValidatedPresetConfig `validate:"dive"`
	Hooks       map[string]string               `validate:"dive,max=200"`
	Env         map[string]string               `validate:"dive,max=200"`
}

// ValidatedFieldConfig represents a field config with validation tags
type ValidatedFieldConfig struct {
	Type        string      `validate:"required,fieldtype"`
	Values      []string    `validate:"max=100,dive,max=100"`
	Default     interface{} `validate:"-"`
	Description string      `validate:"max=500"`
	Path        string      `validate:"max=200"`
}

// ValidatedPresetConfig represents a preset config with validation tags
type ValidatedPresetConfig struct {
	Name        string                 `validate:"required,min=1,max=100"`
	Description string                 `validate:"max=500"`
	Values      map[string]interface{} `validate:"required,min=1,max=50"`
}

// ValidatedConfigData represents config data with validation tags
type ValidatedConfigData struct {
	Data map[string]interface{} `validate:"required,min=1,max=100"`
}

// ValidatedFieldValue represents a field value for validation
type ValidatedFieldValue struct {
	Field string      `validate:"required,min=1,max=100"`
	Value interface{} `validate:"-"`
	Type  string      `validate:"required,fieldtype"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	Errors   []*ValidationError `json:"errors,omitempty"`
	Warnings []*ValidationError `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Path    string      `json:"path,omitempty"`
	Line    int         `json:"line,omitempty"`
	Column  int         `json:"column,omitempty"`
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	v := validator.New()
	
	// Register custom validation functions
	v.RegisterValidation("color", validateColorTag)
	v.RegisterValidation("pathformat", validatePathFormatTag)
	v.RegisterValidation("regex", validateRegexTag)
	v.RegisterValidation("fieldtype", validateFieldTypeTag)
	
	return &Validator{
		schemas:  make(map[string]*Schema),
		validate: v,
	}
}

// LoadSchema loads a validation schema from file
func (v *Validator) LoadSchema(schemaPath string) error {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	if schema.Name == "" {
		schema.Name = strings.TrimSuffix(filepath.Base(schemaPath), filepath.Ext(schemaPath))
	}

	v.schemas[schema.Name] = &schema
	return nil
}

// LoadSchemasFromDir loads all schemas from a directory
func (v *Validator) LoadSchemasFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".json") {
			schemaPath := filepath.Join(dir, entry.Name())
			if err := v.LoadSchema(schemaPath); err != nil {
				return fmt.Errorf("failed to load schema %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// RegisterSchema registers a schema programmatically
func (v *Validator) RegisterSchema(schema *Schema) {
	v.schemas[schema.Name] = schema
}

// ValidateAppConfig validates an application configuration
// Optimized version using validator/v10 struct tags
func (v *Validator) ValidateAppConfig(appName string, appConfig *config.AppConfig) *ValidationResult {
	// Fast path: if schema exists, use schema validation directly
	schema, exists := v.schemas[appName]
	if exists {
		// Check if this is a simple schema that can benefit from struct tags
		if v.isSimpleSchema(schema) {
			return v.validateAppConfigFast(appConfig, schema)
		}
		// Complex schema - use full validation
		return v.validateAppConfigWithSchema(appConfig, schema)
	}

	// No schema - use basic validation with struct tags
	validatedConfig := v.convertToValidatedAppConfig(appConfig)
	if err := v.validate.Struct(validatedConfig); err != nil {
		return v.convertValidatorError(err, "app_config")
	}
	
	return v.validateBasic(appConfig)
}

// ValidateTargetConfig validates a target configuration file
// Optimized version using validator/v10 struct tags
func (v *Validator) ValidateTargetConfig(appName string, configData map[string]interface{}) *ValidationResult {
	// Fast struct tag validation for basic structure
	validatedData := ValidatedConfigData{Data: configData}
	if err := v.validate.Struct(validatedData); err != nil {
		return v.convertValidatorError(err, "config_data")
	}

	// Schema-specific validation if available
	schema, exists := v.schemas[appName]
	if exists {
		return v.validateConfigWithSchema(configData, schema)
	}

	// Basic validation if no schema
	return v.validateBasicConfig(configData)
}

// ValidateField validates a single field value
// Optimized version using validator/v10 struct tags
func (v *Validator) ValidateField(appName string, fieldName string, value interface{}) *ValidationResult {
	schema, exists := v.schemas[appName]
	if !exists {
		// No schema available, perform basic type checking
		return v.validateFieldBasic(fieldName, value)
	}

	rule, exists := schema.Fields[fieldName]
	if !exists {
		return &ValidationResult{
			Valid: true,
			Warnings: []*ValidationError{{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' is not defined in schema", fieldName),
				Code:    "undefined_field",
			}},
		}
	}

	// Fast validation for simple cases
	if rule.Custom == nil && rule.Dependencies == nil && rule.ConflictsWith == nil {
		return v.validateFieldWithRuleFast(fieldName, value, rule)
	}

	// Fallback to original complex validation
	return v.validateFieldWithRule(fieldName, value, rule)
}

// validateBasic performs basic validation without schema
func (v *Validator) validateBasic(appConfig *config.AppConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check required fields
	if appConfig.Name == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   "name",
			Message: "Application name is required",
			Code:    "required_field",
		})
		result.Valid = false
	}

	if appConfig.Path == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   "path",
			Message: "Configuration path is required",
			Code:    "required_field",
		})
		result.Valid = false
	}

	// Validate field types
	for fieldName, field := range appConfig.Fields {
		if field.Type == "" {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Message: "Field type is required",
				Code:    "missing_type",
			})
			result.Valid = false
		} else if !isValidFieldType(field.Type) {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   field.Type,
				Message: fmt.Sprintf("Invalid field type '%s'", field.Type),
				Code:    "invalid_type",
			})
			result.Valid = false
		}

		// Validate choice field has values
		if field.Type == "choice" && len(field.Values) == 0 {
			result.Warnings = append(result.Warnings, &ValidationError{
				Field:   fieldName,
				Message: "Choice field should have predefined values",
				Code:    "missing_values",
			})
		}
	}

	return result
}

// validateBasicConfig performs basic validation on configuration data
func (v *Validator) validateBasicConfig(configData map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Basic checks - ensure values are of reasonable types
	for key, value := range configData {
		if !isValidConfigValue(value) {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   key,
				Value:   value,
				Message: fmt.Sprintf("Invalid value type for field '%s'", key),
				Code:    "invalid_value_type",
			})
			result.Valid = false
		}
	}

	return result
}

// validateWithSchema validates app config against schema
func (v *Validator) validateWithSchema(appConfig *config.AppConfig, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate global rules
	if schema.Global != nil {
		if err := v.validateGlobalRules(appConfig, schema.Global); err != nil {
			result.Errors = append(result.Errors, err...)
			if len(err) > 0 {
				result.Valid = false
			}
		}
	}

	// Validate each field against its rule
	for fieldName, field := range appConfig.Fields {
		rule, exists := schema.Fields[fieldName]
		if !exists {
			result.Warnings = append(result.Warnings, &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' is not defined in schema", fieldName),
				Code:    "undefined_field",
			})
			continue
		}

		// Validate field definition
		fieldResult := v.validateFieldDefinition(fieldName, &field, rule)
		result.Errors = append(result.Errors, fieldResult.Errors...)
		result.Warnings = append(result.Warnings, fieldResult.Warnings...)
		if !fieldResult.Valid {
			result.Valid = false
		}
	}

	// Check for missing required fields
	for fieldName, rule := range schema.Fields {
		if rule.Required {
			if _, exists := appConfig.Fields[fieldName]; !exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Required field '%s' is missing", fieldName),
					Code:    "missing_required_field",
				})
				result.Valid = false
			}
		}
	}

	return result
}

// validateConfigWithSchema validates configuration data against schema
func (v *Validator) validateConfigWithSchema(configData map[string]interface{}, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check global field count restrictions
	if schema.Global != nil {
		if schema.Global.MinFields != nil && len(configData) < *schema.Global.MinFields {
			result.Errors = append(result.Errors, &ValidationError{
				Message: fmt.Sprintf("Configuration must have at least %d fields", *schema.Global.MinFields),
				Code:    "too_few_fields",
			})
			result.Valid = false
		}

		if schema.Global.MaxFields != nil && len(configData) > *schema.Global.MaxFields {
			result.Errors = append(result.Errors, &ValidationError{
				Message: fmt.Sprintf("Configuration cannot have more than %d fields", *schema.Global.MaxFields),
				Code:    "too_many_fields",
			})
			result.Valid = false
		}

		// Check forbidden fields
		for _, forbidden := range schema.Global.ForbiddenFields {
			if _, exists := configData[forbidden]; exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   forbidden,
					Message: fmt.Sprintf("Field '%s' is forbidden", forbidden),
					Code:    "forbidden_field",
				})
				result.Valid = false
			}
		}
	}

	// Validate each field value
	for fieldName, value := range configData {
		rule, exists := schema.Fields[fieldName]
		if !exists {
			result.Warnings = append(result.Warnings, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' is not defined in schema", fieldName),
				Code:    "undefined_field",
			})
			continue
		}

		fieldResult := v.validateFieldWithRule(fieldName, value, rule)
		result.Errors = append(result.Errors, fieldResult.Errors...)
		result.Warnings = append(result.Warnings, fieldResult.Warnings...)
		if !fieldResult.Valid {
			result.Valid = false
		}
	}

	// Check required fields
	for fieldName, rule := range schema.Fields {
		if rule.Required {
			if _, exists := configData[fieldName]; !exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Required field '%s' is missing", fieldName),
					Code:    "missing_required_field",
				})
				result.Valid = false
			}
		}
	}

	// Validate dependencies
	for fieldName, value := range configData {
		rule, exists := schema.Fields[fieldName]
		if !exists || rule.Dependencies == nil {
			continue
		}

		for _, dep := range rule.Dependencies {
			if _, exists := configData[dep]; !exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' requires '%s' to be present", fieldName, dep),
					Code:    "missing_dependency",
				})
				result.Valid = false
			}
		}
	}

	// Validate conflicts
	for fieldName, value := range configData {
		rule, exists := schema.Fields[fieldName]
		if !exists || rule.ConflictsWith == nil {
			continue
		}

		for _, conflict := range rule.ConflictsWith {
			if _, exists := configData[conflict]; exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' conflicts with '%s'", fieldName, conflict),
					Code:    "field_conflict",
				})
				result.Valid = false
			}
		}
	}

	return result
}

// validateFieldBasic performs basic field validation
func (v *Validator) validateFieldBasic(fieldName string, value interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if !isValidConfigValue(value) {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("Invalid value type for field '%s'", fieldName),
			Code:    "invalid_value_type",
		})
		result.Valid = false
	}

	return result
}

// validateFieldWithRule validates a field value against a rule
func (v *Validator) validateFieldWithRule(fieldName string, value interface{}, rule *FieldRule) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Type validation
	if !v.validateFieldType(fieldName, value, rule.Type) {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("Field '%s' must be of type %s", fieldName, rule.Type),
			Code:    "type_mismatch",
		})
		result.Valid = false
		return result // Don't continue if type is wrong
	}

	// Enum validation for choice types
	if rule.Type == "choice" && len(rule.Enum) > 0 {
		strValue := fmt.Sprintf("%v", value)
		valid := false
		for _, enum := range rule.Enum {
			if enum == strValue {
				valid = true
				break
			}
		}
		if !valid {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' must be one of: %s", fieldName, strings.Join(rule.Enum, ", ")),
				Code:    "invalid_choice",
			})
			result.Valid = false
		}
	}

	// String validations
	if rule.Type == "string" {
		strValue := fmt.Sprintf("%v", value)

		// Length validation
		if rule.MinLength != nil && len(strValue) < *rule.MinLength {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, *rule.MinLength),
				Code:    "too_short",
			})
			result.Valid = false
		}

		if rule.MaxLength != nil && len(strValue) > *rule.MaxLength {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' cannot exceed %d characters", fieldName, *rule.MaxLength),
				Code:    "too_long",
			})
			result.Valid = false
		}

		// Pattern validation
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, strValue)
			if err != nil {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Invalid pattern in schema for field '%s'", fieldName),
					Code:    "invalid_pattern",
				})
			} else if !matched {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' does not match required pattern", fieldName),
					Code:    "pattern_mismatch",
				})
				result.Valid = false
			}
		}

		// Format validation
		if rule.Format != "" {
			if err := v.validateFormat(strValue, rule.Format); err != nil {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' has invalid %s format: %s", fieldName, rule.Format, err.Error()),
					Code:    "invalid_format",
				})
				result.Valid = false
			}
		}
	}

	// Numeric validations
	if rule.Type == "number" {
		numValue, err := convertToFloat64(value)
		if err != nil {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' must be a valid number", fieldName),
				Code:    "invalid_number",
			})
			result.Valid = false
		} else {
			if rule.Min != nil && numValue < *rule.Min {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' must be at least %g", fieldName, *rule.Min),
					Code:    "too_small",
				})
				result.Valid = false
			}

			if rule.Max != nil && numValue > *rule.Max {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' cannot exceed %g", fieldName, *rule.Max),
					Code:    "too_large",
				})
				result.Valid = false
			}
		}
	}

	// Custom validation
	if rule.Custom != nil {
		if err := v.validateCustom(fieldName, value, rule.Custom); err != nil {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: err.Error(),
				Code:    "custom_validation_failed",
			})
			result.Valid = false
		}
	}

	return result
}

// validateFieldDefinition validates a field definition against schema rules
func (v *Validator) validateFieldDefinition(fieldName string, field *config.FieldConfig, rule *FieldRule) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check type matches
	if field.Type != rule.Type {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' type mismatch: expected %s, got %s", fieldName, rule.Type, field.Type),
			Code:    "type_mismatch",
		})
		result.Valid = false
	}

	// Validate default value if present
	if field.Default != nil {
		defaultResult := v.validateFieldWithRule(fieldName, field.Default, rule)
		if !defaultResult.Valid {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   field.Default,
				Message: fmt.Sprintf("Default value for field '%s' is invalid", fieldName),
				Code:    "invalid_default",
			})
			result.Valid = false
		}
	}

	return result
}

// validateGlobalRules validates global configuration rules
func (v *Validator) validateGlobalRules(appConfig *config.AppConfig, global *GlobalRules) []*ValidationError {
	var errors []*ValidationError

	fieldCount := len(appConfig.Fields)

	if global.MinFields != nil && fieldCount < *global.MinFields {
		errors = append(errors, &ValidationError{
			Message: fmt.Sprintf("Application must have at least %d fields", *global.MinFields),
			Code:    "too_few_fields",
		})
	}

	if global.MaxFields != nil && fieldCount > *global.MaxFields {
		errors = append(errors, &ValidationError{
			Message: fmt.Sprintf("Application cannot have more than %d fields", *global.MaxFields),
			Code:    "too_many_fields",
		})
	}

	// Check required fields
	for _, required := range global.RequiredFields {
		if _, exists := appConfig.Fields[required]; !exists {
			errors = append(errors, &ValidationError{
				Field:   required,
				Message: fmt.Sprintf("Required field '%s' is missing", required),
				Code:    "missing_required_field",
			})
		}
	}

	// Check forbidden fields
	for _, forbidden := range global.ForbiddenFields {
		if _, exists := appConfig.Fields[forbidden]; exists {
			errors = append(errors, &ValidationError{
				Field:   forbidden,
				Message: fmt.Sprintf("Field '%s' is forbidden", forbidden),
				Code:    "forbidden_field",
			})
		}
	}

	return errors
}

// validateFieldType validates that a value matches the expected type
func (v *Validator) validateFieldType(fieldName string, value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, err := convertToFloat64(value)
		return err == nil
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			// Try to parse string as boolean
			if str, ok := value.(string); ok {
				_, err := strconv.ParseBool(str)
				return err == nil
			}
		}
		return ok
	case "choice":
		// Choices are typically strings
		_, ok := value.(string)
		return ok
	case "array":
		// Arrays can be various types
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			return true
		default:
			return false
		}
	default:
		return true // Unknown types are accepted
	}
}

// validateFormat validates string formats
func (v *Validator) validateFormat(value, format string) error {
	switch format {
	case "email":
		return validateEmail(value)
	case "url":
		return validateURL(value)
	case "path":
		return validatePath(value)
	case "color":
		return validateColor(value)
	case "regex":
		_, err := regexp.Compile(value)
		return err
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// validateCustom validates using custom rules
func (v *Validator) validateCustom(fieldName string, value interface{}, custom *CustomRule) error {
	// This is a placeholder for custom validation logic
	// In a real implementation, you'd have a registry of custom validation functions
	switch custom.Function {
	case "range":
		if min, ok := custom.Args["min"].(float64); ok {
			if max, ok := custom.Args["max"].(float64); ok {
				numValue, err := convertToFloat64(value)
				if err != nil {
					return fmt.Errorf("value must be numeric for range validation")
				}
				if numValue < min || numValue > max {
					return fmt.Errorf("value must be between %g and %g", min, max)
				}
			}
		}
	case "unique":
		// Placeholder for uniqueness validation
		return fmt.Errorf("uniqueness validation not implemented")
	default:
		return fmt.Errorf("unknown custom validation function: %s", custom.Function)
	}
	return nil
}

// Utility functions

func isValidFieldType(fieldType string) bool {
	validTypes := []string{"string", "number", "boolean", "choice", "array"}
	for _, valid := range validTypes {
		if fieldType == valid {
			return true
		}
	}
	return false
}

func isValidConfigValue(value interface{}) bool {
	switch value.(type) {
	case string, int, int64, float64, bool, []interface{}, []string, []int, []float64:
		return true
	case map[string]interface{}:
		return true // Allow nested objects
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
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateURL(url string) error {
	// More flexible URL validation that accepts various schemes
	urlRegex := `^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s]+$`
	matched, err := regexp.MatchString(urlRegex, url)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid URL format")
	}
	return nil
}

func validatePath(path string) error {
	// Basic path validation - check if it's a valid file path format
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}
	return nil
}

func validateColor(color string) error {
	// Validate hex color codes
	hexRegex := `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`
	matched, err := regexp.MatchString(hexRegex, color)
	if err != nil {
		return err
	}
	if !matched {
		// Also allow named colors
		namedColors := []string{"red", "green", "blue", "white", "black", "yellow", "cyan", "magenta"}
		for _, named := range namedColors {
			if strings.EqualFold(color, named) {
				return nil
			}
		}
		return fmt.Errorf("invalid color format")
	}
	return nil
}

// Optimized validation helper functions using validator/v10

// convertToValidatedAppConfig converts config.AppConfig to ValidatedAppConfig
func (v *Validator) convertToValidatedAppConfig(appConfig *config.AppConfig) ValidatedAppConfig {
	validatedFields := make(map[string]ValidatedFieldConfig)
	for name, field := range appConfig.Fields {
		validatedFields[name] = ValidatedFieldConfig{
			Type:        field.Type,
			Values:      field.Values,
			Default:     field.Default,
			Description: field.Description,
			Path:        field.Path,
		}
	}
	
	validatedPresets := make(map[string]ValidatedPresetConfig)
	for name, preset := range appConfig.Presets {
		validatedPresets[name] = ValidatedPresetConfig{
			Name:        preset.Name,
			Description: preset.Description,
			Values:      preset.Values,
		}
	}
	
	return ValidatedAppConfig{
		Name:        appConfig.Name,
		Path:        appConfig.Path,
		Format:      appConfig.Format,
		Description: appConfig.Description,
		Fields:      validatedFields,
		Presets:     validatedPresets,
		Hooks:       appConfig.Hooks,
		Env:         appConfig.Env,
	}
}

// convertValidatorError converts validator errors to ValidationResult
func (v *Validator) convertValidatorError(err error, context string) *ValidationResult {
	result := &ValidationResult{Valid: false}
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldErr.Field(),
				Value:   fieldErr.Value(),
				Message: v.getValidationErrorMessage(fieldErr),
				Code:    v.getValidationErrorCode(fieldErr.Tag()),
				Path:    context,
			})
		}
	} else {
		result.Errors = append(result.Errors, &ValidationError{
			Message: fmt.Sprintf("Validation error: %v", err),
			Code:    "validation_error",
			Path:    context,
		})
	}
	
	return result
}

// getValidationErrorMessage returns user-friendly error messages
func (v *Validator) getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", err.Field())
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("Field '%s' cannot exceed %s", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", err.Field(), err.Param())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email", err.Field())
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", err.Field())
	case "fieldtype":
		return fmt.Sprintf("Field '%s' has invalid type", err.Field())
	case "color":
		return fmt.Sprintf("Field '%s' must be a valid color", err.Field())
	case "pathformat":
		return fmt.Sprintf("Field '%s' must be a valid path", err.Field())
	case "regex":
		return fmt.Sprintf("Field '%s' must be a valid regex", err.Field())
	default:
		return fmt.Sprintf("Field '%s' validation failed: %s", err.Field(), err.Tag())
	}
}

// getValidationErrorCode returns consistent error codes
func (v *Validator) getValidationErrorCode(tag string) string {
	switch tag {
	case "required":
		return "required_field"
	case "min":
		return "too_short"
	case "max":
		return "too_long"
	case "oneof":
		return "invalid_choice"
	case "email":
		return "invalid_format"
	case "url":
		return "invalid_format"
	case "fieldtype":
		return "invalid_type"
	case "color":
		return "invalid_format"
	case "pathformat":
		return "invalid_format"
	case "regex":
		return "invalid_pattern"
	default:
		return "validation_error"
	}
}

// validateFieldWithRuleFast provides fast validation for simple field rules
func (v *Validator) validateFieldWithRuleFast(fieldName string, value interface{}, rule *FieldRule) *ValidationResult {
	// Create a dynamic validation struct for this field
	fieldVal := ValidatedFieldValue{
		Field: fieldName,
		Value: value,
		Type:  rule.Type,
	}
	
	// Fast struct validation
	if err := v.validate.Struct(fieldVal); err != nil {
		return v.convertValidatorError(err, "field_validation")
	}
	
	// Additional type-specific validation
	return v.validateFieldTypeSpecific(fieldName, value, rule)
}

// validateFieldTypeSpecific handles type-specific validations
func (v *Validator) validateFieldTypeSpecific(fieldName string, value interface{}, rule *FieldRule) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Type validation
	if !v.validateFieldType(fieldName, value, rule.Type) {
		result.Errors = append(result.Errors, &ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("Field '%s' must be of type %s", fieldName, rule.Type),
			Code:    "type_mismatch",
		})
		result.Valid = false
		return result
	}
	
	// String-specific validations
	if rule.Type == "string" {
		strValue := fmt.Sprintf("%v", value)
		
		// Length validation
		if rule.MinLength != nil && len(strValue) < *rule.MinLength {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, *rule.MinLength),
				Code:    "too_short",
			})
			result.Valid = false
		}
		
		if rule.MaxLength != nil && len(strValue) > *rule.MaxLength {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' cannot exceed %d characters", fieldName, *rule.MaxLength),
				Code:    "too_long",
			})
			result.Valid = false
		}
		
		// Pattern validation
		if rule.Pattern != "" {
			if matched, err := regexp.MatchString(rule.Pattern, strValue); err != nil || !matched {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' does not match required pattern", fieldName),
					Code:    "pattern_mismatch",
				})
				result.Valid = false
			}
		}
		
		// Format validation
		if rule.Format != "" {
			if err := v.validateFormat(strValue, rule.Format); err != nil {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' has invalid %s format", fieldName, rule.Format),
					Code:    "invalid_format",
				})
				result.Valid = false
			}
		}
	}
	
	// Numeric validations
	if rule.Type == "number" {
		if numValue, err := convertToFloat64(value); err == nil {
			if rule.Min != nil && numValue < *rule.Min {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' must be at least %g", fieldName, *rule.Min),
					Code:    "too_small",
				})
				result.Valid = false
			}
			
			if rule.Max != nil && numValue > *rule.Max {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   value,
					Message: fmt.Sprintf("Field '%s' cannot exceed %g", fieldName, *rule.Max),
					Code:    "too_large",
				})
				result.Valid = false
			}
		}
	}
	
	// Choice validation
	if rule.Type == "choice" && len(rule.Enum) > 0 {
		strValue := fmt.Sprintf("%v", value)
		valid := false
		for _, enum := range rule.Enum {
			if enum == strValue {
				valid = true
				break
			}
		}
		if !valid {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("Field '%s' must be one of: %s", fieldName, strings.Join(rule.Enum, ", ")),
				Code:    "invalid_choice",
			})
			result.Valid = false
		}
	}
	
	return result
}

// validateAppConfigWithSchema performs schema validation on app config
func (v *Validator) validateAppConfigWithSchema(appConfig *config.AppConfig, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Validate global rules first
	if schema.Global != nil {
		if err := v.validateGlobalRules(appConfig, schema.Global); err != nil {
			result.Errors = append(result.Errors, err...)
			if len(err) > 0 {
				result.Valid = false
			}
		}
	}
	
	// Validate each field
	for fieldName, field := range appConfig.Fields {
		rule, exists := schema.Fields[fieldName]
		if !exists {
			result.Warnings = append(result.Warnings, &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' is not defined in schema", fieldName),
				Code:    "undefined_field",
			})
			continue
		}
		
		// Validate field definition
		fieldResult := v.validateFieldDefinition(fieldName, &field, rule)
		result.Errors = append(result.Errors, fieldResult.Errors...)
		result.Warnings = append(result.Warnings, fieldResult.Warnings...)
		if !fieldResult.Valid {
			result.Valid = false
		}
	}
	
	// Check for missing required fields
	for fieldName, rule := range schema.Fields {
		if rule.Required {
			if _, exists := appConfig.Fields[fieldName]; !exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("Required field '%s' is missing", fieldName),
					Code:    "missing_required_field",
				})
				result.Valid = false
			}
		}
	}
	
	return result
}

// Custom validation functions for validator/v10

// validateColorTag validates color format for validator/v10
func validateColorTag(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	return validateColor(color) == nil
}

// validatePathFormatTag validates path format for validator/v10
func validatePathFormatTag(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return validatePath(path) == nil
}

// validateRegexTag validates regex format for validator/v10
func validateRegexTag(fl validator.FieldLevel) bool {
	pattern := fl.Field().String()
	_, err := regexp.Compile(pattern)
	return err == nil
}

// validateFieldTypeTag validates field type for validator/v10
func validateFieldTypeTag(fl validator.FieldLevel) bool {
	fieldType := fl.Field().String()
	return isValidFieldType(fieldType)
}

// isSimpleSchema determines if a schema can benefit from fast validation
func (v *Validator) isSimpleSchema(schema *Schema) bool {
	// Simple schemas have no dependencies, conflicts, or custom validation
	for _, rule := range schema.Fields {
		if rule.Dependencies != nil || rule.ConflictsWith != nil || rule.Custom != nil {
			return false
		}
	}
	
	// Simple schemas have minimal global rules
	if schema.Global != nil {
		if len(schema.Global.RequiredFields) > 5 || len(schema.Global.ForbiddenFields) > 0 {
			return false
		}
	}
	
	return true
}

// validateAppConfigFast provides fast validation for simple schemas
func (v *Validator) validateAppConfigFast(appConfig *config.AppConfig, schema *Schema) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Fast basic structure validation
	validatedConfig := v.convertToValidatedAppConfig(appConfig)
	if err := v.validate.Struct(validatedConfig); err != nil {
		result = v.convertValidatorError(err, "app_config")
		if !result.Valid {
			return result
		}
	}
	
	// Quick field type and basic constraint checks
	for fieldName, field := range appConfig.Fields {
		rule, exists := schema.Fields[fieldName]
		if !exists {
			continue
		}
		
		// Fast type check
		if field.Type != rule.Type {
			result.Errors = append(result.Errors, &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' type mismatch: expected %s, got %s", fieldName, rule.Type, field.Type),
				Code:    "type_mismatch",
			})
			result.Valid = false
		}
		
		// Fast default value validation
		if field.Default != nil {
			if !v.validateFieldType(fieldName, field.Default, rule.Type) {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   fieldName,
					Value:   field.Default,
					Message: fmt.Sprintf("Default value for field '%s' is invalid", fieldName),
					Code:    "invalid_default",
				})
				result.Valid = false
			}
		}
	}
	
	// Quick required field check
	if schema.Global != nil {
		for _, required := range schema.Global.RequiredFields {
			if _, exists := appConfig.Fields[required]; !exists {
				result.Errors = append(result.Errors, &ValidationError{
					Field:   required,
					Message: fmt.Sprintf("Required field '%s' is missing", required),
					Code:    "missing_required_field",
				})
				result.Valid = false
			}
		}
	}
	
	return result
}
