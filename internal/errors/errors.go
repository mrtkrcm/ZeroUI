package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// Config related errors
	ConfigNotFound      ErrorType = "CONFIG_NOT_FOUND"
	ConfigParseError    ErrorType = "CONFIG_PARSE_ERROR"
	ConfigWriteError    ErrorType = "CONFIG_WRITE_ERROR"
	ConfigInvalidFormat ErrorType = "CONFIG_INVALID_FORMAT"
	ConfigPermission    ErrorType = "CONFIG_PERMISSION"

	// App related errors
	AppNotFound    ErrorType = "APP_NOT_FOUND"
	AppInvalid     ErrorType = "APP_INVALID"
	AppConfigError ErrorType = "APP_CONFIG_ERROR"

	// Field related errors
	FieldNotFound     ErrorType = "FIELD_NOT_FOUND"
	FieldInvalidValue ErrorType = "FIELD_INVALID_VALUE"
	FieldInvalidType  ErrorType = "FIELD_INVALID_TYPE"

	// Preset related errors
	PresetNotFound ErrorType = "PRESET_NOT_FOUND"
	PresetInvalid  ErrorType = "PRESET_INVALID"

	// System related errors
	SystemPermission ErrorType = "SYSTEM_PERMISSION"
	SystemFileError  ErrorType = "SYSTEM_FILE_ERROR"
	SystemCommand    ErrorType = "SYSTEM_COMMAND"

	// Hook related errors
	HookFailed   ErrorType = "HOOK_FAILED"
	HookNotFound ErrorType = "HOOK_NOT_FOUND"

	// Plugin related errors
	PluginNotFound ErrorType = "PLUGIN_NOT_FOUND"
	PluginError    ErrorType = "PLUGIN_ERROR"

	// Validation related errors
	ValidationError   ErrorType = "VALIDATION_ERROR"
	SchemaError      ErrorType = "SCHEMA_ERROR"
	TypeConversion   ErrorType = "TYPE_CONVERSION"

	// User input errors
	UserInputError   ErrorType = "USER_INPUT_ERROR"
	CommandLineError ErrorType = "COMMAND_LINE_ERROR"
)

// ZeroUIError represents a structured error with context
type ZeroUIError struct {
	Type        ErrorType
	Message     string
	App         string
	Field       string
	Value       string
	Path        string        // File path where error occurred
	Line        int          // Line number in file (if applicable)
	Column      int          // Column number in file (if applicable)
	Cause       error
	Suggestions []string
	Actions     []string     // Actionable next steps
	Context     map[string]string // Additional context information
	Severity    Severity     // Error severity level
}

// Error implements the error interface
func (e *ZeroUIError) Error() string {
	var parts []string
	
	if e.App != "" {
		parts = append(parts, fmt.Sprintf("app: %s", e.App))
	}
	if e.Field != "" {
		parts = append(parts, fmt.Sprintf("field: %s", e.Field))
	}
	if e.Value != "" {
		parts = append(parts, fmt.Sprintf("value: %s", e.Value))
	}
	
	context := ""
	if len(parts) > 0 {
		context = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
	}
	
	msg := e.Message + context
	
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	
	return msg
}

// Unwrap returns the underlying cause error
func (e *ZeroUIError) Unwrap() error {
	return e.Cause
}

// String returns a user-friendly error message with suggestions
func (e *ZeroUIError) String() string {
	msg := e.Error()
	
	if len(e.Suggestions) > 0 {
		msg += "\n\nSuggestions:"
		for _, suggestion := range e.Suggestions {
			msg += fmt.Sprintf("\n  • %s", suggestion)
		}
	}
	
	return msg
}

// IsType checks if the error is of a specific type
func (e *ZeroUIError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// New creates a new ZeroUIError
func New(errorType ErrorType, message string) *ZeroUIError {
	return &ZeroUIError{
		Type:    errorType,
		Message: message,
	}
}

// Wrap creates a new ZeroUIError that wraps another error
func Wrap(errorType ErrorType, message string, cause error) *ZeroUIError {
	return &ZeroUIError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// WithApp adds app context to the error
func (e *ZeroUIError) WithApp(app string) *ZeroUIError {
	e.App = app
	return e
}

// WithField adds field context to the error
func (e *ZeroUIError) WithField(field string) *ZeroUIError {
	e.Field = field
	return e
}

// WithValue adds value context to the error
func (e *ZeroUIError) WithValue(value string) *ZeroUIError {
	e.Value = value
	return e
}

// WithSuggestions adds helpful suggestions
func (e *ZeroUIError) WithSuggestions(suggestions ...string) *ZeroUIError {
	e.Suggestions = suggestions
	return e
}

// Common error constructors for convenience

// NewAppNotFoundError creates an app not found error
func NewAppNotFoundError(app string, availableApps []string) *ZeroUIError {
	suggestions := []string{
		"Check available apps with: zeroui list apps",
	}
	if len(availableApps) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Available apps: %s", strings.Join(availableApps, ", ")))
	}
	
	return New(AppNotFound, fmt.Sprintf("application '%s' not found", app)).
		WithApp(app).
		WithSuggestions(suggestions...)
}

// NewFieldNotFoundError creates a field not found error
func NewFieldNotFoundError(app, field string, availableFields []string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Check available fields with: zeroui list keys %s", app),
	}
	if len(availableFields) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Available fields: %s", strings.Join(availableFields, ", ")))
	}
	
	return New(FieldNotFound, fmt.Sprintf("field '%s' not found", field)).
		WithApp(app).
		WithField(field).
		WithSuggestions(suggestions...)
}

// NewInvalidValueError creates an invalid value error
func NewInvalidValueError(app, field, value string, validValues []string) *ZeroUIError {
	suggestions := []string{}
	if len(validValues) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Valid values: %s", strings.Join(validValues, ", ")))
	}
	
	return New(FieldInvalidValue, fmt.Sprintf("invalid value '%s' for field '%s'", value, field)).
		WithApp(app).
		WithField(field).
		WithValue(value).
		WithSuggestions(suggestions...)
}

// NewPresetNotFoundError creates a preset not found error
func NewPresetNotFoundError(app, preset string, availablePresets []string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Check available presets with: zeroui list presets %s", app),
	}
	if len(availablePresets) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Available presets: %s", strings.Join(availablePresets, ", ")))
	}
	
	return New(PresetNotFound, fmt.Sprintf("preset '%s' not found", preset)).
		WithApp(app).
		WithValue(preset).
		WithSuggestions(suggestions...)
}

// NewConfigNotFoundError creates a config file not found error
func NewConfigNotFoundError(path string) *ZeroUIError {
	suggestions := []string{
		"Create the configuration file manually, or",
		"Use --dry-run to see what would be changed without creating the file",
		"Check if the path is correct and accessible",
	}
	
	return New(ConfigNotFound, fmt.Sprintf("configuration file not found at '%s'", path)).
		WithSuggestions(suggestions...)
}

// NewConfigParseError creates a config parsing error
func NewConfigParseError(path string, cause error) *ZeroUIError {
	suggestions := []string{
		"Check the configuration file syntax",
		"Ensure the file format matches the expected format",
		"Try validating the file with a format-specific tool",
	}
	
	return Wrap(ConfigParseError, fmt.Sprintf("failed to parse configuration file '%s'", path), cause).
		WithSuggestions(suggestions...)
}

// NewPermissionError creates a permission error
func NewPermissionError(path string, operation string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Ensure you have %s permissions for '%s'", operation, path),
		"Check file and directory permissions",
		"Try running with appropriate privileges if needed",
	}
	
	return New(SystemPermission, fmt.Sprintf("permission denied: cannot %s '%s'", operation, path)).
		WithSuggestions(suggestions...)
}

// IsZeroUIError checks if an error is a ZeroUIError
func IsZeroUIError(err error) bool {
	_, ok := err.(*ZeroUIError)
	return ok
}

// GetZeroUIError extracts a ZeroUIError from an error
func GetZeroUIError(err error) (*ZeroUIError, bool) {
	if ctErr, ok := err.(*ZeroUIError); ok {
		return ctErr, true
	}
	return nil, false
}