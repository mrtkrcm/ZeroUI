package errors

import (
	"fmt"
	"testing"
)

// TestZeroUIError_Error tests error message formatting
func TestZeroUIError_Error(t *testing.T) {
	err := &ZeroUIError{
		Type:    AppNotFound,
		Message: "application not found",
		App:     "test-app",
		Field:   "theme",
		Value:   "invalid",
	}

	expected := "application not found (app: test-app, field: theme, value: invalid)"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestZeroUIError_String tests user-friendly error formatting
func TestZeroUIError_String(t *testing.T) {
	err := &ZeroUIError{
		Type:        AppNotFound,
		Message:     "application not found",
		App:         "test-app",
		Suggestions: []string{"Check available apps", "Verify app name"},
	}

	result := err.String()
	if !containsString(result, "application not found") {
		t.Error("Expected error message to contain main error")
	}

	if !containsString(result, "Suggestions:") {
		t.Error("Expected error message to contain suggestions section")
	}

	if !containsString(result, "Check available apps") {
		t.Error("Expected error message to contain first suggestion")
	}

	if !containsString(result, "Verify app name") {
		t.Error("Expected error message to contain second suggestion")
	}
}

// TestZeroUIError_IsType tests error type checking
func TestZeroUIError_IsType(t *testing.T) {
	err := &ZeroUIError{
		Type: AppNotFound,
	}

	if !err.IsType(AppNotFound) {
		t.Error("Expected error to be of type AppNotFound")
	}

	if err.IsType(FieldNotFound) {
		t.Error("Expected error not to be of type FieldNotFound")
	}
}

// TestZeroUIError_Unwrap tests error unwrapping
func TestZeroUIError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := &ZeroUIError{
		Type:  SystemFileError,
		Cause: cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Expected unwrapped error to match cause")
	}

	// Test with no cause
	errNoCause := &ZeroUIError{
		Type: AppNotFound,
	}

	if errNoCause.Unwrap() != nil {
		t.Error("Expected unwrapped error to be nil when no cause")
	}
}

// TestNew tests creating a new error
func TestNew(t *testing.T) {
	err := New(AppNotFound, "test message")

	if err.Type != AppNotFound {
		t.Errorf("Expected type AppNotFound, got %s", err.Type)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	if err.Cause != nil {
		t.Error("Expected no cause for new error")
	}
}

// TestWrap tests wrapping an error
func TestWrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	err := Wrap(SystemFileError, "wrapped message", cause)

	if err.Type != SystemFileError {
		t.Errorf("Expected type SystemFileError, got %s", err.Type)
	}

	if err.Message != "wrapped message" {
		t.Errorf("Expected message 'wrapped message', got '%s'", err.Message)
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}
}

// TestZeroUIError_WithApp tests adding app context
func TestZeroUIError_WithApp(t *testing.T) {
	err := New(FieldNotFound, "field not found")
	err = err.WithApp("test-app")

	if err.App != "test-app" {
		t.Errorf("Expected app 'test-app', got '%s'", err.App)
	}

	// Method chaining should return the same instance
	err2 := err.WithField("theme").WithValue("dark")
	if err2 != err {
		t.Error("Expected method chaining to return same instance")
	}

	if err.Field != "theme" {
		t.Errorf("Expected field 'theme', got '%s'", err.Field)
	}

	if err.Value != "dark" {
		t.Errorf("Expected value 'dark', got '%s'", err.Value)
	}
}

// TestZeroUIError_WithSuggestions tests adding suggestions
func TestZeroUIError_WithSuggestions(t *testing.T) {
	err := New(AppNotFound, "app not found")
	err = err.WithSuggestions("suggestion 1", "suggestion 2")

	if len(err.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.Suggestions))
	}

	if err.Suggestions[0] != "suggestion 1" {
		t.Errorf("Expected first suggestion 'suggestion 1', got '%s'", err.Suggestions[0])
	}

	if err.Suggestions[1] != "suggestion 2" {
		t.Errorf("Expected second suggestion 'suggestion 2', got '%s'", err.Suggestions[1])
	}
}

// TestNewAppNotFoundError tests the app not found constructor
func TestNewAppNotFoundError(t *testing.T) {
	availableApps := []string{"app1", "app2", "app3"}
	err := NewAppNotFoundError("missing-app", availableApps)

	if err.Type != AppNotFound {
		t.Errorf("Expected type AppNotFound, got %s", err.Type)
	}

	if err.App != "missing-app" {
		t.Errorf("Expected app 'missing-app', got '%s'", err.App)
	}

	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be set")
	}

	// Check that available apps are mentioned in suggestions
	found := false
	for _, suggestion := range err.Suggestions {
		if containsString(suggestion, "app1") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected suggestions to mention available apps")
	}
}

// TestNewFieldNotFoundError tests the field not found constructor
func TestNewFieldNotFoundError(t *testing.T) {
	availableFields := []string{"theme", "font-size", "debug"}
	err := NewFieldNotFoundError("test-app", "missing-field", availableFields)

	if err.Type != FieldNotFound {
		t.Errorf("Expected type FieldNotFound, got %s", err.Type)
	}

	if err.App != "test-app" {
		t.Errorf("Expected app 'test-app', got '%s'", err.App)
	}

	if err.Field != "missing-field" {
		t.Errorf("Expected field 'missing-field', got '%s'", err.Field)
	}

	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be set")
	}
}

// TestNewInvalidValueError tests the invalid value constructor
func TestNewInvalidValueError(t *testing.T) {
	validValues := []string{"dark", "light", "auto"}
	err := NewInvalidValueError("test-app", "theme", "invalid", validValues)

	if err.Type != FieldInvalidValue {
		t.Errorf("Expected type FieldInvalidValue, got %s", err.Type)
	}

	if err.App != "test-app" {
		t.Errorf("Expected app 'test-app', got '%s'", err.App)
	}

	if err.Field != "theme" {
		t.Errorf("Expected field 'theme', got '%s'", err.Field)
	}

	if err.Value != "invalid" {
		t.Errorf("Expected value 'invalid', got '%s'", err.Value)
	}

	if len(err.Suggestions) == 0 {
		t.Error("Expected suggestions to be set")
	}

	// Check that valid values are mentioned in suggestions
	found := false
	for _, suggestion := range err.Suggestions {
		if containsString(suggestion, "dark") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected suggestions to mention valid values")
	}
}

// TestIsZeroUIError tests error type detection
func TestIsZeroUIError(t *testing.T) {
	ctErr := New(AppNotFound, "test")
	if !IsZeroUIError(ctErr) {
		t.Error("Expected ZeroUIError to be detected")
	}

	stdErr := fmt.Errorf("standard error")
	if IsZeroUIError(stdErr) {
		t.Error("Expected standard error not to be detected as ZeroUIError")
	}
}

// TestGetZeroUIError tests extracting ZeroUIError
func TestGetZeroUIError(t *testing.T) {
	ctErr := New(AppNotFound, "test")
	extracted, ok := GetZeroUIError(ctErr)
	if !ok {
		t.Error("Expected to extract ZeroUIError")
	}
	if extracted != ctErr {
		t.Error("Expected extracted error to be same instance")
	}

	stdErr := fmt.Errorf("standard error")
	_, ok = GetZeroUIError(stdErr)
	if ok {
		t.Error("Expected standard error not to be extractable")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkZeroUIError_Error benchmarks error message creation
func BenchmarkZeroUIError_Error(b *testing.B) {
	err := &ZeroUIError{
		Type:    FieldInvalidValue,
		Message: "invalid value for field",
		App:     "test-app",
		Field:   "theme",
		Value:   "invalid-value",
		Cause:   fmt.Errorf("underlying error"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// BenchmarkZeroUIError_String benchmarks user-friendly error formatting
func BenchmarkZeroUIError_String(b *testing.B) {
	err := &ZeroUIError{
		Type:    FieldInvalidValue,
		Message: "invalid value for field",
		App:     "test-app",
		Field:   "theme",
		Value:   "invalid-value",
		Suggestions: []string{
			"Check valid values",
			"Use configtoggle list keys",
			"Verify field type",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.String()
	}
}
