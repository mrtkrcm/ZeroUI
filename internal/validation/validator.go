package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// Validator provides configuration validation functionality
type Validator struct {
	schemas map[string]*Schema
}

// Schema represents a validation schema for an application
type Schema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Fields      map[string]*FieldRule  `json:"fields"`
	Global      *GlobalRules           `json:"global,omitempty"`
}

// FieldRule defines validation rules for a specific field
type FieldRule struct {
	Type         string          `json:"type"`                   // string, number, boolean, choice, array
	Required     bool            `json:"required,omitempty"`     // Field is required
	Pattern      string          `json:"pattern,omitempty"`      // Regex pattern for strings
	MinLength    *int            `json:"min_length,omitempty"`   // Minimum string length
	MaxLength    *int            `json:"max_length,omitempty"`   // Maximum string length
	Min          *float64        `json:"min,omitempty"`          // Minimum numeric value
	Max          *float64        `json:"max,omitempty"`          // Maximum numeric value
	Enum         []string        `json:"enum,omitempty"`         // Valid values for choice type
	Default      interface{}     `json:"default,omitempty"`      // Default value
	Dependencies []string        `json:"dependencies,omitempty"` // Fields that must be present if this field is set
	ConflictsWith []string       `json:"conflicts_with,omitempty"` // Fields that cannot be set together
	Format       string          `json:"format,omitempty"`       // Format specification (email, url, etc.)
	Custom       *CustomRule     `json:"custom,omitempty"`       // Custom validation rule
}

// GlobalRules defines global validation rules
type GlobalRules struct {
	MinFields      *int     `json:"min_fields,omitempty"`      // Minimum number of fields
	MaxFields      *int     `json:"max_fields,omitempty"`      // Maximum number of fields
	RequiredFields []string `json:"required_fields,omitempty"` // Globally required fields
	ForbiddenFields []string `json:"forbidden_fields,omitempty"` // Forbidden field names
}

// CustomRule represents a custom validation rule
type CustomRule struct {
	Function string                 `json:"function"`           // Function name
	Args     map[string]interface{} `json:"args,omitempty"`     // Function arguments
	Message  string                 `json:"message,omitempty"`  // Custom error message
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Errors  []*ValidationError `json:"errors,omitempty"`
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
	return &Validator{
		schemas: make(map[string]*Schema),
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
func (v *Validator) ValidateAppConfig(appName string, appConfig *config.AppConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}
	_ = result // Mark as used

	schema, exists := v.schemas[appName]
	if !exists {
		// No schema available, perform basic validation
		return v.validateBasic(appConfig)
	}

	// Validate using schema
	return v.validateWithSchema(appConfig, schema)
}

// ValidateTargetConfig validates a target configuration file
func (v *Validator) ValidateTargetConfig(appName string, configData map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}
	_ = result // Mark as used

	schema, exists := v.schemas[appName]
	if !exists {
		// No schema available, perform basic validation
		return v.validateBasicConfig(configData)
	}

	// Validate configuration against schema
	return v.validateConfigWithSchema(configData, schema)
}

// ValidateField validates a single field value
func (v *Validator) ValidateField(appName string, fieldName string, value interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	schema, exists := v.schemas[appName]
	if !exists {
		// No schema available, perform basic type checking
		return v.validateFieldBasic(fieldName, value)
	}

	rule, exists := schema.Fields[fieldName]
	if !exists {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("Field '%s' is not defined in schema", fieldName),
			Code:    "undefined_field",
		})
		return result
	}

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
	urlRegex := `^https?://[^\s]+$`
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