package validation

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// BenchmarkSuperOptimizedValidation tests validation with a very simple schema
func BenchmarkSuperOptimizedValidation(b *testing.B) {
	validator := NewValidator()

	// Create a very simple schema that will trigger the fast path
	schema := &Schema{
		Name: "super-simple",
		Fields: map[string]*FieldRule{
			"name": {Type: "string", Required: true},
			"port": {Type: "number"},
			"enabled": {Type: "boolean"},
		},
		Global: &GlobalRules{
			RequiredFields: []string{"name"},
		},
	}
	validator.RegisterSchema(schema)

	// Simple config
	appConfig := &config.AppConfig{
		Name:   "super-simple",
		Path:   "/config",
		Format: "json",
		Fields: map[string]config.FieldConfig{
			"name":    {Type: "string", Default: "test"},
			"port":    {Type: "number", Default: 8080},
			"enabled": {Type: "boolean", Default: true},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateAppConfig("super-simple", appConfig)
		if !result.Valid {
			b.Fatalf("Validation failed: %v", result.Errors)
		}
	}
}

// BenchmarkComplexSchemaValidation tests validation with a complex schema
func BenchmarkComplexSchemaValidation(b *testing.B) {
	validator := NewValidator()

	// Complex schema that will use full validation
	schema := &Schema{
		Name: "complex",
		Fields: map[string]*FieldRule{
			"username": {
				Type:          "string",
				Required:      true,
				MinLength:     intPtr(3),
				MaxLength:     intPtr(50),
				Pattern:       "^[a-zA-Z0-9_]+$",
				ConflictsWith: []string{"email"}, // This makes it complex
			},
			"email": {
				Type:         "string",
				Format:       "email",
				Dependencies: []string{"username"}, // This makes it complex
			},
			"password": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(8),
				Custom: &CustomRule{ // This makes it complex
					Function: "strong_password",
				},
			},
		},
		Global: &GlobalRules{
			RequiredFields:  []string{"username", "password"},
			ForbiddenFields: []string{"admin_key"},
			MinFields:       intPtr(2),
			MaxFields:       intPtr(10),
		},
	}
	validator.RegisterSchema(schema)

	appConfig := &config.AppConfig{
		Name:   "complex",
		Path:   "/config",
		Format: "json",
		Fields: map[string]config.FieldConfig{
			"username": {Type: "string", Default: "testuser"},
			"password": {Type: "string", Default: "securepass123"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateAppConfig("complex", appConfig)
		// Note: this will likely have errors due to custom validation, but we're measuring performance
		_ = result
	}
}

// TestFastPathDetection verifies that simple schemas use the fast path
func TestFastPathDetection(t *testing.T) {
	validator := NewValidator()

	// Simple schema
	simpleSchema := &Schema{
		Name: "simple",
		Fields: map[string]*FieldRule{
			"name": {Type: "string", Required: true},
			"port": {Type: "number"},
		},
		Global: &GlobalRules{
			RequiredFields: []string{"name"},
		},
	}

	// Complex schema
	complexSchema := &Schema{
		Name: "complex",
		Fields: map[string]*FieldRule{
			"name": {
				Type:          "string",
				Required:      true,
				ConflictsWith: []string{"alias"}, // Makes it complex
			},
		},
	}

	if !validator.isSimpleSchema(simpleSchema) {
		t.Error("Expected simple schema to be detected as simple")
	}

	if validator.isSimpleSchema(complexSchema) {
		t.Error("Expected complex schema to be detected as complex")
	}
}