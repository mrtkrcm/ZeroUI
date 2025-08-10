package errors

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Severity represents the severity level of an error
type Severity int

const (
	Info Severity = iota
	Warning
	Error
	Critical
)

// String returns the string representation of severity
func (s Severity) String() string {
	switch s {
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case Critical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Color returns the ANSI color code for the severity
func (s Severity) Color() string {
	switch s {
	case Info:
		return "\033[36m" // Cyan
	case Warning:
		return "\033[33m" // Yellow
	case Error:
		return "\033[31m" // Red
	case Critical:
		return "\033[35m" // Magenta
	default:
		return "\033[0m"  // Reset
	}
}

// ColorReset returns the ANSI reset code
func ColorReset() string {
	return "\033[0m"
}

// Enhanced error methods for ZeroUIError

// WithPath adds file path context to the error
func (e *ZeroUIError) WithPath(path string) *ZeroUIError {
	e.Path = path
	return e
}

// WithLocation adds line and column context to the error
func (e *ZeroUIError) WithLocation(line, column int) *ZeroUIError {
	e.Line = line
	e.Column = column
	return e
}

// WithSeverity sets the error severity
func (e *ZeroUIError) WithSeverity(severity Severity) *ZeroUIError {
	e.Severity = severity
	return e
}

// WithActions adds actionable next steps
func (e *ZeroUIError) WithActions(actions ...string) *ZeroUIError {
	e.Actions = actions
	return e
}

// WithContext adds additional context information
func (e *ZeroUIError) WithContext(key, value string) *ZeroUIError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// DetailedString returns a detailed error message with all context
func (e *ZeroUIError) DetailedString() string {
	var parts []string
	
	// Add severity and error type
	header := fmt.Sprintf("[%s] %s", e.Severity.String(), e.Type)
	parts = append(parts, header)
	
	// Add main message
	parts = append(parts, e.Message)
	
	// Add location information
	if e.Path != "" {
		location := fmt.Sprintf("File: %s", e.Path)
		if e.Line > 0 {
			location += fmt.Sprintf(":%d", e.Line)
			if e.Column > 0 {
				location += fmt.Sprintf(":%d", e.Column)
			}
		}
		parts = append(parts, location)
	}
	
	// Add context
	if e.App != "" || e.Field != "" || e.Value != "" {
		context := "Context:"
		if e.App != "" {
			context += fmt.Sprintf(" app=%s", e.App)
		}
		if e.Field != "" {
			context += fmt.Sprintf(" field=%s", e.Field)
		}
		if e.Value != "" {
			context += fmt.Sprintf(" value=%s", e.Value)
		}
		parts = append(parts, context)
	}
	
	// Add additional context
	for key, value := range e.Context {
		parts = append(parts, fmt.Sprintf("%s: %s", key, value))
	}
	
	// Add cause if present
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Caused by: %v", e.Cause))
	}
	
	msg := strings.Join(parts, "\n")
	
	// Add suggestions
	if len(e.Suggestions) > 0 {
		msg += "\n\nSuggestions:"
		for _, suggestion := range e.Suggestions {
			msg += fmt.Sprintf("\n  💡 %s", suggestion)
		}
	}
	
	// Add actions
	if len(e.Actions) > 0 {
		msg += "\n\nNext Steps:"
		for i, action := range e.Actions {
			msg += fmt.Sprintf("\n  %d. %s", i+1, action)
		}
	}
	
	return msg
}

// ColoredString returns a colored error message for terminal output
func (e *ZeroUIError) ColoredString() string {
	color := e.Severity.Color()
	reset := ColorReset()
	
	// Color the header
	header := fmt.Sprintf("%s[%s] %s%s", color, e.Severity.String(), e.Type, reset)
	msg := header + "\n" + e.Message
	
	// Add context with subtle coloring
	if e.App != "" || e.Field != "" || e.Value != "" {
		msg += "\n\033[2m" // Dim
		if e.App != "" {
			msg += fmt.Sprintf(" app: %s", e.App)
		}
		if e.Field != "" {
			msg += fmt.Sprintf(" field: %s", e.Field)
		}
		if e.Value != "" {
			msg += fmt.Sprintf(" value: %s", e.Value)
		}
		msg += reset
	}
	
	// Add location with file icon
	if e.Path != "" {
		msg += "\n📁 " + e.Path
		if e.Line > 0 {
			msg += fmt.Sprintf(":%d", e.Line)
			if e.Column > 0 {
				msg += fmt.Sprintf(":%d", e.Column)
			}
		}
	}
	
	// Add cause
	if e.Cause != nil {
		msg += fmt.Sprintf("\n🔗 %v", e.Cause)
	}
	
	// Add suggestions with light bulb
	if len(e.Suggestions) > 0 {
		msg += "\n\n💡 Suggestions:"
		for _, suggestion := range e.Suggestions {
			msg += fmt.Sprintf("\n   • %s", suggestion)
		}
	}
	
	// Add actions with numbered steps
	if len(e.Actions) > 0 {
		msg += "\n\n🔧 Next Steps:"
		for i, action := range e.Actions {
			msg += fmt.Sprintf("\n   %d. %s", i+1, action)
		}
	}
	
	return msg
}

// Enhanced error constructors

// NewValidationError creates a validation error
func NewValidationError(message string, field string, value interface{}) *ZeroUIError {
	suggestions := []string{
		"Check the field type and format requirements",
		"Ensure the value matches the expected pattern",
	}
	
	actions := []string{
		"Review the field documentation",
		"Try a different value that matches the requirements",
	}
	
	return New(ValidationError, message).
		WithField(field).
		WithValue(fmt.Sprintf("%v", value)).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// NewTypeConversionError creates a type conversion error
func NewTypeConversionError(value string, expectedType string, field string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Value must be a valid %s", expectedType),
	}
	
	actions := []string{
		fmt.Sprintf("Provide a valid %s value", expectedType),
		"Check the field documentation for examples",
	}
	
	// Add type-specific suggestions
	switch expectedType {
	case "boolean":
		suggestions = append(suggestions, "Valid boolean values: true, false, yes, no, 1, 0")
		actions = append(actions, "Try: true or false")
	case "number", "integer":
		suggestions = append(suggestions, "Value should be a numeric value (e.g., 42, 3.14)")
		actions = append(actions, "Try a numeric value like 10 or 1.5")
	case "string":
		suggestions = append(suggestions, "Value should be text, optionally in quotes")
		actions = append(actions, "Try wrapping the value in quotes if it contains special characters")
	}
	
	return New(TypeConversion, fmt.Sprintf("cannot convert '%s' to %s", value, expectedType)).
		WithField(field).
		WithValue(value).
		WithSeverity(Error).
		WithContext("expected_type", expectedType).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// NewUserInputError creates a user input error
func NewUserInputError(message string, input string) *ZeroUIError {
	suggestions := []string{
		"Check the command syntax and arguments",
		"Use --help to see available options",
	}
	
	actions := []string{
		"Review your input and try again",
		"Run the command with --help for usage information",
	}
	
	return New(UserInputError, message).
		WithValue(input).
		WithSeverity(Warning).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// Enhanced versions of existing constructors

// NewAppNotFoundErrorEnhanced creates an enhanced app not found error
func NewAppNotFoundErrorEnhanced(app string, availableApps []string) *ZeroUIError {
	suggestions := []string{
		"Check available apps with: configtoggle list apps",
	}
	
	actions := []string{
		"Run 'configtoggle list apps' to see all available applications",
	}
	
	if len(availableApps) > 0 {
		// Find similar app names using simple string matching
		similar := findSimilarStrings(app, availableApps, 3)
		if len(similar) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s?", strings.Join(similar, ", ")))
			actions = append(actions, fmt.Sprintf("Try one of these similar apps: %s", strings.Join(similar, ", ")))
		}
		suggestions = append(suggestions, fmt.Sprintf("Available apps: %s", strings.Join(availableApps, ", ")))
	}
	
	return New(AppNotFound, fmt.Sprintf("application '%s' not found", app)).
		WithApp(app).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...).
		WithContext("available_count", strconv.Itoa(len(availableApps)))
}

// NewFieldNotFoundErrorEnhanced creates an enhanced field not found error
func NewFieldNotFoundErrorEnhanced(app, field string, availableFields []string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Check available fields with: configtoggle list keys %s", app),
	}
	
	actions := []string{
		fmt.Sprintf("Run 'configtoggle list keys %s' to see all available fields", app),
	}
	
	if len(availableFields) > 0 {
		// Find similar field names
		similar := findSimilarStrings(field, availableFields, 3)
		if len(similar) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s?", strings.Join(similar, ", ")))
			actions = append(actions, fmt.Sprintf("Try one of these similar fields: %s", strings.Join(similar, ", ")))
		}
		suggestions = append(suggestions, fmt.Sprintf("Available fields: %s", strings.Join(availableFields, ", ")))
	}
	
	return New(FieldNotFound, fmt.Sprintf("field '%s' not found in app '%s'", field, app)).
		WithApp(app).
		WithField(field).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...).
		WithContext("field_count", strconv.Itoa(len(availableFields)))
}

// NewInvalidValueErrorEnhanced creates an enhanced invalid value error
func NewInvalidValueErrorEnhanced(app, field, value string, validValues []string) *ZeroUIError {
	suggestions := []string{}
	actions := []string{}
	
	if len(validValues) > 0 {
		// Find similar values
		similar := findSimilarStrings(value, validValues, 3)
		if len(similar) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s?", strings.Join(similar, ", ")))
			actions = append(actions, fmt.Sprintf("Try: configtoggle toggle %s %s %s", app, field, similar[0]))
		}
		suggestions = append(suggestions, fmt.Sprintf("Valid values: %s", strings.Join(validValues, ", ")))
		actions = append(actions, fmt.Sprintf("Use one of the valid values: %s", strings.Join(validValues, ", ")))
	} else {
		suggestions = append(suggestions, "This field accepts any string value")
		actions = append(actions, "Ensure the value format is correct")
	}
	
	return New(FieldInvalidValue, fmt.Sprintf("invalid value '%s' for field '%s'", value, field)).
		WithApp(app).
		WithField(field).
		WithValue(value).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...).
		WithContext("valid_count", strconv.Itoa(len(validValues)))
}

// NewPresetNotFoundErrorEnhanced creates an enhanced preset not found error
func NewPresetNotFoundErrorEnhanced(app, preset string, availablePresets []string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Check available presets with: configtoggle list presets %s", app),
	}
	
	actions := []string{
		fmt.Sprintf("Run 'configtoggle list presets %s' to see all available presets", app),
	}
	
	if len(availablePresets) > 0 {
		// Find similar preset names
		similar := findSimilarStrings(preset, availablePresets, 3)
		if len(similar) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s?", strings.Join(similar, ", ")))
			actions = append(actions, fmt.Sprintf("Try: configtoggle preset %s %s", app, similar[0]))
		}
		suggestions = append(suggestions, fmt.Sprintf("Available presets: %s", strings.Join(availablePresets, ", ")))
	} else {
		suggestions = append(suggestions, "This app has no configured presets")
		actions = append(actions, "Create a preset in the app configuration file")
	}
	
	return New(PresetNotFound, fmt.Sprintf("preset '%s' not found in app '%s'", preset, app)).
		WithApp(app).
		WithValue(preset).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...).
		WithContext("preset_count", strconv.Itoa(len(availablePresets)))
}

// NewConfigNotFoundErrorEnhanced creates an enhanced config file not found error
func NewConfigNotFoundErrorEnhanced(path string) *ZeroUIError {
	suggestions := []string{
		"Create the configuration file manually",
		"Use --dry-run to see what would be changed without creating the file",
		"Check if the path is correct and accessible",
	}
	
	actions := []string{
		fmt.Sprintf("Create the directory: mkdir -p %s", filepath.Dir(path)),
		fmt.Sprintf("Create an empty config: touch %s", path),
		"Run the command again",
	}
	
	return New(ConfigNotFound, fmt.Sprintf("configuration file not found at '%s'", path)).
		WithPath(path).
		WithSeverity(Warning).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// NewConfigParseErrorEnhanced creates an enhanced config parsing error
func NewConfigParseErrorEnhanced(path string, cause error) *ZeroUIError {
	suggestions := []string{
		"Check the configuration file syntax",
		"Ensure the file format matches the expected format",
		"Try validating the file with a format-specific tool",
		"Look for common issues: missing quotes, trailing commas, invalid escapes",
	}
	
	actions := []string{
		fmt.Sprintf("Open and check the file: %s", path),
		"Validate the syntax with a format-specific validator",
		"Fix any syntax errors and try again",
	}
	
	return Wrap(ConfigParseError, fmt.Sprintf("failed to parse configuration file '%s'", path), cause).
		WithPath(path).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// NewPermissionErrorEnhanced creates an enhanced permission error
func NewPermissionErrorEnhanced(path string, operation string) *ZeroUIError {
	suggestions := []string{
		fmt.Sprintf("Ensure you have %s permissions for '%s'", operation, path),
		"Check file and directory permissions",
		"Try running with appropriate privileges if needed",
	}
	
	actions := []string{
		fmt.Sprintf("Check permissions: ls -la %s", filepath.Dir(path)),
		"Fix permissions if needed",
		"Retry the operation",
	}
	
	return New(SystemPermission, fmt.Sprintf("permission denied: cannot %s '%s'", operation, path)).
		WithPath(path).
		WithSeverity(Error).
		WithSuggestions(suggestions...).
		WithActions(actions...)
}

// Utility functions

// findSimilarStrings finds strings that are similar to the target using simple edit distance
func findSimilarStrings(target string, candidates []string, maxResults int) []string {
	type candidate struct {
		value    string
		distance int
	}
	
	var similar []candidate
	
	for _, c := range candidates {
		dist := levenshteinDistance(target, c)
		// Only consider strings that are reasonably similar
		if dist <= len(target)/2+1 {
			similar = append(similar, candidate{value: c, distance: dist})
		}
	}
	
	// Sort by distance (closest first)
	for i := 0; i < len(similar)-1; i++ {
		for j := i + 1; j < len(similar); j++ {
			if similar[i].distance > similar[j].distance {
				similar[i], similar[j] = similar[j], similar[i]
			}
		}
	}
	
	// Return up to maxResults
	result := make([]string, 0, maxResults)
	for i, c := range similar {
		if i >= maxResults {
			break
		}
		result = append(result, c.value)
	}
	
	return result
}

// levenshteinDistance computes the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return utf8.RuneCountInString(s2)
	}
	if len(s2) == 0 {
		return utf8.RuneCountInString(s1)
	}
	
	runes1 := []rune(s1)
	runes2 := []rune(s2)
	
	len1, len2 := len(runes1), len(runes2)
	
	// Create a matrix to store distances
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if runes1[i-1] != runes2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = minThree(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len1][len2]
}

// minThree returns the minimum of three integers
func minThree(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}