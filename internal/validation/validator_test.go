package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// setupValidatorTest creates a test environment with sample schemas
func setupValidatorTest(t *testing.T) (*Validator, string, func()) {
	tmpDir, err := os.MkdirTemp("", "validator-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	validator := NewValidator()

	// Create sample schema
	schema := &Schema{
		Name:        "test-app",
		Description: "Test application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"theme": {
				Type:     "choice",
				Required: true,
				Enum:     []string{"dark", "light", "auto"},
				Default:  "dark",
			},
			"font-size": {
				Type:    "number",
				Min:     floatPtr(8),
				Max:     floatPtr(72),
				Default: 14,
			},
			"debug": {
				Type:    "boolean",
				Default: false,
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
			"website": {
				Type:   "string",
				Format: "url",
			},
			"username": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(3),
				MaxLength: intPtr(20),
				Pattern:   "^[a-zA-Z0-9_]+$",
			},
		},
		Global: &GlobalRules{
			MinFields:       intPtr(2),
			MaxFields:       intPtr(10),
			RequiredFields:  []string{"theme", "username"},
			ForbiddenFields: []string{"password"},
		},
	}

	validator.RegisterSchema(schema)

	// Create schema file for loading tests with a different name
	loadedSchema := &Schema{
		Name:        "loaded-app",
		Description: "Test loaded application schema",
		Version:     "1.0.0",
		Fields:      make(map[string]*FieldRule),
	}
	schemaPath := filepath.Join(tmpDir, "test-schema.json")
	schemaData, _ := json.MarshalIndent(loadedSchema, "", "  ")
	if err := os.WriteFile(schemaPath, schemaData, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return validator, tmpDir, cleanup
}

// Helper functions
func intPtr(i int) *int           { return &i }
func floatPtr(f float64) *float64 { return &f }

// TestNewValidator tests validator creation
func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("Expected non-nil validator")
	}

	if validator.schemas == nil {
		t.Error("Expected schemas map to be initialized")
	}
}

// TestValidator_RegisterSchema tests schema registration
func TestValidator_RegisterSchema(t *testing.T) {
	validator := NewValidator()

	schema := &Schema{
		Name:        "test-app",
		Description: "Test schema",
		Fields:      make(map[string]*FieldRule),
	}

	validator.RegisterSchema(schema)

	if len(validator.schemas) != 1 {
		t.Errorf("Expected 1 schema, got %d", len(validator.schemas))
	}

	retrievedSchema, exists := validator.schemas["test-app"]
	if !exists {
		t.Error("Expected schema to be registered")
	}

	if retrievedSchema.Name != "test-app" {
		t.Errorf("Expected schema name 'test-app', got '%s'", retrievedSchema.Name)
	}
}

// TestValidator_LoadSchema tests loading schema from file
func TestValidator_LoadSchema(t *testing.T) {
	validator, tmpDir, cleanup := setupValidatorTest(t)
	defer cleanup()

	// Test loading existing schema
	schemaPath := filepath.Join(tmpDir, "test-schema.json")
	err := validator.LoadSchema(schemaPath)
	if err != nil {
		t.Fatalf("Failed to load schema: %v", err)
	}

	// Should now have 2 schemas (one from setup, one loaded)
	if len(validator.schemas) != 2 {
		t.Errorf("Expected 2 schemas after loading, got %d", len(validator.schemas))
	}

	// Test loading non-existent schema
	err = validator.LoadSchema("/nonexistent/schema.json")
	if err == nil {
		t.Error("Expected error for non-existent schema file")
	}

	// Test loading invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	err = validator.LoadSchema(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON schema")
	}
}

// TestValidator_ValidateField tests single field validation
func TestValidator_ValidateField(t *testing.T) {
	validator, _, cleanup := setupValidatorTest(t)
	defer cleanup()

	tests := []struct {
		name        string
		app         string
		field       string
		value       interface{}
		expectValid bool
		expectCode  string
	}{
		{"Valid theme", "test-app", "theme", "dark", true, ""},
		{"Invalid theme", "test-app", "theme", "invalid", false, "invalid_choice"},
		{"Valid font-size", "test-app", "font-size", 16, true, ""},
		{"Font-size too small", "test-app", "font-size", 5, false, "too_small"},
		{"Font-size too large", "test-app", "font-size", 100, false, "too_large"},
		{"Valid boolean", "test-app", "debug", true, true, ""},
		{"Valid email", "test-app", "email", "test@example.com", true, ""},
		{"Invalid email", "test-app", "email", "invalid-email", false, "invalid_format"},
		{"Valid URL", "test-app", "website", "https://example.com", true, ""},
		{"Invalid URL", "test-app", "website", "not-a-url", false, "invalid_format"},
		{"Valid username", "test-app", "username", "user123", true, ""},
		{"Username too short", "test-app", "username", "ab", false, "too_short"},
		{"Username too long", "test-app", "username", "verylongusernamethatexceedslimit", false, "too_long"},
		{"Username invalid pattern", "test-app", "username", "user@domain", false, "pattern_mismatch"},
		{"Undefined field", "test-app", "unknown", "value", true, "undefined_field"}, // Should be warning
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateField(tt.app, tt.field, tt.value)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Valid)
			}

			if tt.expectCode != "" {
				found := false
				for _, err := range result.Errors {
					if err.Code == tt.expectCode {
						found = true
						break
					}
				}
				for _, warn := range result.Warnings {
					if warn.Code == tt.expectCode {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected error/warning with code '%s', but not found. Errors: %v, Warnings: %v",
						tt.expectCode, result.Errors, result.Warnings)
				}
			}
		})
	}
}

// TestValidator_ValidateTargetConfig tests configuration validation
func TestValidator_ValidateTargetConfig(t *testing.T) {
	validator, _, cleanup := setupValidatorTest(t)
	defer cleanup()

	// Test valid configuration
	t.Run("Valid config", func(t *testing.T) {
		config := map[string]interface{}{
			"theme":     "dark",
			"username":  "testuser",
			"font-size": 16,
			"debug":     false,
			"email":     "test@example.com",
		}

		result := validator.ValidateTargetConfig("test-app", config)
		if !result.Valid {
			t.Errorf("Expected valid config. Errors: %v", result.Errors)
		}
	})

	// Test missing required fields
	t.Run("Missing required fields", func(t *testing.T) {
		config := map[string]interface{}{
			"font-size": 16,
			"debug":     false,
		}

		result := validator.ValidateTargetConfig("test-app", config)
		if result.Valid {
			t.Error("Expected invalid config due to missing required fields")
		}

		// Should have errors for missing theme and username
		requiredErrors := 0
		for _, err := range result.Errors {
			if err.Code == "missing_required_field" {
				requiredErrors++
			}
		}

		if requiredErrors < 2 {
			t.Errorf("Expected at least 2 missing required field errors, got %d", requiredErrors)
		}
	})

	// Test forbidden fields
	t.Run("Forbidden fields", func(t *testing.T) {
		config := map[string]interface{}{
			"theme":    "dark",
			"username": "testuser",
			"password": "secret", // Forbidden field
		}

		result := validator.ValidateTargetConfig("test-app", config)
		if result.Valid {
			t.Error("Expected invalid config due to forbidden field")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "forbidden_field" && err.Field == "password" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected forbidden field error for 'password'")
		}
	})

	// Test field count limits
	t.Run("Too few fields", func(t *testing.T) {
		config := map[string]interface{}{
			"theme": "dark",
		}

		result := validator.ValidateTargetConfig("test-app", config)
		if result.Valid {
			t.Error("Expected invalid config due to too few fields")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "too_few_fields" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected too_few_fields error")
		}
	})

	// Test with no schema (basic validation)
	t.Run("No schema", func(t *testing.T) {
		config := map[string]interface{}{
			"setting1": "value1",
			"setting2": 42,
			"setting3": true,
		}

		result := validator.ValidateTargetConfig("unknown-app", config)
		if !result.Valid {
			t.Errorf("Expected valid config with basic validation. Errors: %v", result.Errors)
		}
	})
}

// TestValidator_ValidateAppConfig tests application configuration validation
func TestValidator_ValidateAppConfig(t *testing.T) {
	validator, _, cleanup := setupValidatorTest(t)
	defer cleanup()

	// Test valid app config
	t.Run("Valid app config", func(t *testing.T) {
		appConfig := &config.AppConfig{
			Name:   "test-app",
			Path:   "/path/to/config",
			Format: "json",
			Fields: map[string]config.FieldConfig{
				"theme": {
					Type:        "choice",
					Values:      []string{"dark", "light", "auto"},
					Default:     "dark",
					Description: "Theme setting",
				},
				"username": {
					Type:        "string",
					Default:     "testuser",
					Description: "Username",
				},
			},
		}

		result := validator.ValidateAppConfig("test-app", appConfig)
		if !result.Valid {
			t.Errorf("Expected valid app config. Errors: %v", result.Errors)
		}
	})

	// Test app config with type mismatch
	t.Run("Type mismatch", func(t *testing.T) {
		appConfig := &config.AppConfig{
			Name:   "test-app",
			Path:   "/path/to/config",
			Format: "json",
			Fields: map[string]config.FieldConfig{
				"theme": {
					Type:        "string", // Schema expects choice
					Values:      []string{"dark", "light"},
					Default:     "dark",
					Description: "Theme setting",
				},
			},
		}

		result := validator.ValidateAppConfig("test-app", appConfig)
		if result.Valid {
			t.Error("Expected invalid app config due to type mismatch")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "type_mismatch" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected type_mismatch error")
		}
	})

	// Test basic validation without schema
	t.Run("Basic validation", func(t *testing.T) {
		appConfig := &config.AppConfig{
			Name:   "unknown-app",
			Path:   "/path/to/config",
			Format: "json",
			Fields: map[string]config.FieldConfig{
				"setting1": {
					Type:        "string",
					Description: "A string setting",
				},
				"setting2": {
					Type:        "invalid-type", // Invalid type
					Description: "Invalid setting",
				},
			},
		}

		result := validator.ValidateAppConfig("unknown-app", appConfig)
		if result.Valid {
			t.Error("Expected invalid app config due to invalid field type")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "invalid_type" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected invalid_type error")
		}
	})
}

// TestValidationResult tests validation result structure
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []*ValidationError{
			{
				Field:   "test_field",
				Value:   "test_value",
				Message: "Test error",
				Code:    "test_error",
			},
		},
		Warnings: []*ValidationError{
			{
				Field:   "warn_field",
				Message: "Test warning",
				Code:    "test_warning",
			},
		},
	}

	if result.Valid {
		t.Error("Expected result to be invalid")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if result.Errors[0].Field != "test_field" {
		t.Errorf("Expected error field 'test_field', got '%s'", result.Errors[0].Field)
	}

	if result.Warnings[0].Code != "test_warning" {
		t.Errorf("Expected warning code 'test_warning', got '%s'", result.Warnings[0].Code)
	}
}

// TestUtilityFunctions tests utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test isValidFieldType
	t.Run("isValidFieldType", func(t *testing.T) {
		validTypes := []string{"string", "number", "boolean", "choice", "array"}
		invalidTypes := []string{"invalid", "object", "null"}

		for _, validType := range validTypes {
			if !isValidFieldType(validType) {
				t.Errorf("Expected '%s' to be valid field type", validType)
			}
		}

		for _, invalidType := range invalidTypes {
			if isValidFieldType(invalidType) {
				t.Errorf("Expected '%s' to be invalid field type", invalidType)
			}
		}
	})

	// Test isValidConfigValue
	t.Run("isValidConfigValue", func(t *testing.T) {
		validValues := []interface{}{
			"string",
			42,
			int64(42),
			3.14,
			true,
			[]interface{}{"a", "b"},
			[]string{"a", "b"},
			map[string]interface{}{"key": "value"},
		}

		invalidValues := []interface{}{
			complex(1, 2),
			make(chan int),
		}

		for i, validValue := range validValues {
			if !isValidConfigValue(validValue) {
				t.Errorf("Expected value %d (%T) to be valid", i, validValue)
			}
		}

		for i, invalidValue := range invalidValues {
			if isValidConfigValue(invalidValue) {
				t.Errorf("Expected value %d (%T) to be invalid", i, invalidValue)
			}
		}
	})

	// Test convertToFloat64
	t.Run("convertToFloat64", func(t *testing.T) {
		tests := []struct {
			input     interface{}
			expected  float64
			shouldErr bool
		}{
			{42, 42.0, false},
			{int64(42), 42.0, false},
			{3.14, 3.14, false},
			{"42.5", 42.5, false},
			{"not-a-number", 0, true},
			{true, 0, true},
		}

		for _, test := range tests {
			result, err := convertToFloat64(test.input)

			if test.shouldErr {
				if err == nil {
					t.Errorf("Expected error for input %v (%T)", test.input, test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %v: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %g, got %g for input %v", test.expected, result, test.input)
				}
			}
		}
	})
}

// TestFormatValidation tests format validation
func TestFormatValidation(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		format    string
		value     string
		shouldErr bool
	}{
		{"email", "test@example.com", false},
		{"email", "invalid-email", true},
		{"url", "https://example.com", false},
		{"url", "not-a-url", true},
		{"path", "/valid/path", false},
		{"path", "path\x00with\x00nulls", true},
		{"color", "#FF0000", false},
		{"color", "#FFF", false},
		{"color", "red", false},
		{"color", "invalid-color", true},
		{"regex", "^[a-z]+$", false},
		{"regex", "[invalid", true},
		{"unknown", "value", true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.format, test.value), func(t *testing.T) {
			err := validator.validateFormat(test.value, test.format)

			if test.shouldErr {
				if err == nil {
					t.Errorf("Expected error for format '%s' with value '%s'", test.format, test.value)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for format '%s' with value '%s': %v", test.format, test.value, err)
				}
			}
		})
	}
}

// TestSchemaLoadingFromDir tests loading multiple schemas from directory
func TestSchemaLoadingFromDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "schema-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create multiple schema files
	schemas := []*Schema{
		{Name: "app1", Description: "App 1", Fields: make(map[string]*FieldRule)},
		{Name: "app2", Description: "App 2", Fields: make(map[string]*FieldRule)},
		{Name: "app3", Description: "App 3", Fields: make(map[string]*FieldRule)},
	}

	for _, schema := range schemas {
		schemaData, _ := json.MarshalIndent(schema, "", "  ")
		schemaPath := filepath.Join(tmpDir, schema.Name+".json")
		if err := os.WriteFile(schemaPath, schemaData, 0644); err != nil {
			t.Fatalf("Failed to write schema file: %v", err)
		}
	}

	// Create a non-JSON file that should be ignored
	if err := os.WriteFile(filepath.Join(tmpDir, "not-a-schema.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("Failed to write non-JSON file: %v", err)
	}

	validator := NewValidator()
	err = validator.LoadSchemasFromDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load schemas from directory: %v", err)
	}

	if len(validator.schemas) != 3 {
		t.Errorf("Expected 3 schemas to be loaded, got %d", len(validator.schemas))
	}

	for _, schema := range schemas {
		if _, exists := validator.schemas[schema.Name]; !exists {
			t.Errorf("Expected schema '%s' to be loaded", schema.Name)
		}
	}

	// Test with non-existent directory
	err = validator.LoadSchemasFromDir("/nonexistent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
