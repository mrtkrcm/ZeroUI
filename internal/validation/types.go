package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
)

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

	// Performance optimization: cached enum map for O(1) lookups
	enumMap       map[string]struct{} `json:"-"` // Cached for fast validation
	compiledRegex *regexp.Regexp      `json:"-"` // Pre-compiled regex
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

// ValidatedAppConfig represents an app config with validation tags
type ValidatedAppConfig struct {
	Name        string                           `validate:"required,min=1,max=100"`
	Path        string                           `validate:"required,min=1"`
	Format      string                           `validate:"required,oneof=json yaml yml toml custom"`
	Description string                           `validate:"max=500"`
	Fields      map[string]ValidatedFieldConfig  `validate:"required,min=1,max=50,dive"`
	Presets     map[string]ValidatedPresetConfig `validate:"dive"`
	Hooks       map[string]string                `validate:"dive,max=200"`
	Env         map[string]string                `validate:"dive,max=200"`
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
	Fields map[string]ValidatedFieldValue `validate:"required,min=1,dive"`
}

// ValidatedFieldValue represents a field value with validation tags
type ValidatedFieldValue struct {
	Value interface{} `validate:"required"`
	Type  string      `validate:"required,fieldtype"`
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid    bool                   `json:"valid"`
	Errors   []*ValidationError     `json:"errors,omitempty"`
	Warnings []*ValidationError     `json:"warnings,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

// Validator provides configuration validation functionality
type Validator struct {
	schemas  map[string]*Schema
	validate *validator.Validate
}

// Manager defines the interface for a validation service
type Manager interface {
	ValidateAppConfig(appName string, appConfig *appconfig.AppConfig) *ValidationResult
	ValidateTargetConfig(appName string, configData map[string]interface{}) *ValidationResult
	RegisterSchema(appName string, schema *Schema) error
	GetSchema(appName string) (*Schema, error)
}
