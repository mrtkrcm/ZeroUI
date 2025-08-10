package validation

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// BenchmarkValidator_ValidateAppConfig compares old vs new validation performance
func BenchmarkValidator_ValidateAppConfig(b *testing.B) {
	validator := NewValidator()

	// Create a complex schema
	schema := &Schema{
		Name:        "benchmark-app",
		Description: "Benchmark application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"theme": {
				Type:     "choice",
				Required: true,
				Enum:     []string{"dark", "light", "auto", "system", "custom"},
				Default:  "dark",
			},
			"font-size": {
				Type:    "number",
				Min:     floatPtr(6),
				Max:     floatPtr(144),
				Default: 12,
			},
			"debug": {
				Type:    "boolean",
				Default: false,
			},
			"email": {
				Type:     "string",
				Required: true,
				Format:   "email",
				MinLength: intPtr(5),
				MaxLength: intPtr(254),
			},
			"website": {
				Type:   "string",
				Format: "url",
			},
			"username": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(3),
				MaxLength: intPtr(32),
				Pattern:   "^[a-zA-Z0-9_-]+$",
			},
			"password": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(8),
				MaxLength: intPtr(128),
			},
			"api-key": {
				Type:      "string",
				Pattern:   "^[a-zA-Z0-9]{32}$",
				MinLength: intPtr(32),
				MaxLength: intPtr(32),
			},
			"timeout": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(3600),
			},
			"max-connections": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(10000),
			},
		},
		Global: &GlobalRules{
			MinFields:       intPtr(3),
			MaxFields:       intPtr(20),
			RequiredFields:  []string{"theme", "username", "email", "password"},
			ForbiddenFields: []string{"secret", "private-key"},
		},
	}

	validator.RegisterSchema(schema)

	// Create test app config
	appConfig := &config.AppConfig{
		Name:        "benchmark-app",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Benchmark test application",
		Fields: map[string]config.FieldConfig{
			"theme": {
				Type:        "choice",
				Values:      []string{"dark", "light", "auto", "system", "custom"},
				Default:     "dark",
				Description: "UI theme",
			},
			"font-size": {
				Type:        "number",
				Default:     12,
				Description: "Font size",
			},
			"debug": {
				Type:        "boolean",
				Default:     false,
				Description: "Debug mode",
			},
			"email": {
				Type:        "string",
				Default:     "user@example.com",
				Description: "User email",
			},
			"website": {
				Type:        "string",
				Default:     "https://example.com",
				Description: "Website URL",
			},
			"username": {
				Type:        "string",
				Default:     "testuser",
				Description: "Username",
			},
			"password": {
				Type:        "string",
				Default:     "securepassword123",
				Description: "Password",
			},
			"api-key": {
				Type:        "string",
				Default:     "abcdef1234567890abcdef1234567890",
				Description: "API key",
			},
			"timeout": {
				Type:        "number",
				Default:     30,
				Description: "Request timeout",
			},
			"max-connections": {
				Type:        "number",
				Default:     100,
				Description: "Maximum connections",
			},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := validator.ValidateAppConfig("benchmark-app", appConfig)
			if !result.Valid {
				b.Errorf("Expected valid config, got errors: %v", result.Errors)
			}
		}
	})
}

// BenchmarkValidator_ValidateTargetConfig benchmarks target config validation
func BenchmarkValidator_ValidateTargetConfig(b *testing.B) {
	validator := NewValidator()

	// Create a complex schema
	schema := &Schema{
		Name:        "benchmark-target-app",
		Description: "Benchmark target application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"database_url": {
				Type:     "string",
				Required: true,
				Format:   "url",
			},
			"port": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(65535),
			},
			"host": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(1),
				MaxLength: intPtr(255),
			},
			"ssl_enabled": {
				Type:    "boolean",
				Default: true,
			},
			"log_level": {
				Type: "choice",
				Enum: []string{"debug", "info", "warn", "error", "fatal"},
			},
		},
		Global: &GlobalRules{
			MinFields:      intPtr(2),
			MaxFields:      intPtr(10),
			RequiredFields: []string{"database_url", "host"},
		},
	}

	validator.RegisterSchema(schema)

	// Create test config data
	configData := map[string]interface{}{
		"database_url": "postgresql://user:pass@localhost:5432/mydb",
		"port":         8080,
		"host":         "localhost",
		"ssl_enabled":  true,
		"log_level":    "info",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := validator.ValidateTargetConfig("benchmark-target-app", configData)
			if !result.Valid {
				b.Errorf("Expected valid config, got errors: %v", result.Errors)
			}
		}
	})
}

// BenchmarkValidator_ValidateField benchmarks field validation
func BenchmarkValidator_ValidateField(b *testing.B) {
	validator := NewValidator()

	// Create a complex schema
	schema := &Schema{
		Name:        "benchmark-field-app",
		Description: "Benchmark field application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"email": {
				Type:      "string",
				Required:  true,
				Format:    "email",
				MinLength: intPtr(5),
				MaxLength: intPtr(254),
				Pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
		},
	}

	validator.RegisterSchema(schema)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := validator.ValidateField("benchmark-field-app", "email", "user@example.com")
			if !result.Valid {
				b.Errorf("Expected valid field, got errors: %v", result.Errors)
			}
		}
	})
}

// BenchmarkValidator_ValidateMultipleFields benchmarks validation of multiple fields
func BenchmarkValidator_ValidateMultipleFields(b *testing.B) {
	validator := NewValidator()

	// Create a complex schema
	schema := &Schema{
		Name:        "benchmark-multi-app",
		Description: "Benchmark multi-field application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"username": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(3),
				MaxLength: intPtr(32),
				Pattern:   "^[a-zA-Z0-9_-]+$",
			},
			"age": {
				Type: "number",
				Min:  floatPtr(13),
				Max:  floatPtr(120),
			},
			"country": {
				Type: "choice",
				Enum: []string{"US", "UK", "CA", "AU", "DE", "FR", "JP", "BR", "IN", "CN"},
			},
			"website": {
				Type:   "string",
				Format: "url",
			},
			"bio": {
				Type:      "string",
				MaxLength: intPtr(500),
			},
		},
	}

	validator.RegisterSchema(schema)

	fields := []struct {
		name  string
		value interface{}
	}{
		{"username", "testuser123"},
		{"age", 25},
		{"country", "US"},
		{"website", "https://example.com"},
		{"bio", "Software engineer with 5 years of experience"},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, field := range fields {
				result := validator.ValidateField("benchmark-multi-app", field.name, field.value)
				if !result.Valid {
					b.Errorf("Expected valid field %s, got errors: %v", field.name, result.Errors)
				}
			}
		}
	})
}

// BenchmarkValidator_ComplexValidation benchmarks complex validation scenarios
func BenchmarkValidator_ComplexValidation(b *testing.B) {
	validator := NewValidator()

	// Create a very complex schema with dependencies and conflicts
	schema := &Schema{
		Name:        "complex-benchmark-app",
		Description: "Complex benchmark application schema",
		Version:     "1.0.0",
		Fields: map[string]*FieldRule{
			"auth_type": {
				Type:     "choice",
				Required: true,
				Enum:     []string{"oauth", "basic", "token", "certificate"},
			},
			"oauth_client_id": {
				Type:         "string",
				Dependencies: []string{"auth_type"},
				MinLength:    intPtr(10),
				MaxLength:    intPtr(100),
			},
			"oauth_client_secret": {
				Type:         "string",
				Dependencies: []string{"oauth_client_id"},
				MinLength:    intPtr(20),
				MaxLength:    intPtr(200),
			},
			"basic_username": {
				Type:          "string",
				ConflictsWith: []string{"oauth_client_id", "token_value"},
				MinLength:     intPtr(3),
				MaxLength:     intPtr(50),
			},
			"basic_password": {
				Type:         "string",
				Dependencies: []string{"basic_username"},
				MinLength:    intPtr(8),
				MaxLength:    intPtr(128),
			},
			"token_value": {
				Type:          "string",
				ConflictsWith: []string{"basic_username", "oauth_client_id"},
				MinLength:     intPtr(32),
				MaxLength:     intPtr(512),
			},
			"ssl_verify": {
				Type:    "boolean",
				Default: true,
			},
			"timeout": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(300),
			},
		},
		Global: &GlobalRules{
			MinFields:      intPtr(2),
			MaxFields:      intPtr(15),
			RequiredFields: []string{"auth_type"},
		},
	}

	validator.RegisterSchema(schema)

	// Create test config with OAuth
	configData := map[string]interface{}{
		"auth_type":          "oauth",
		"oauth_client_id":    "1234567890abcdef",
		"oauth_client_secret": "super_secret_oauth_secret_key_123456789",
		"ssl_verify":         true,
		"timeout":            30,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := validator.ValidateTargetConfig("complex-benchmark-app", configData)
			if !result.Valid {
				b.Errorf("Expected valid config, got errors: %v", result.Errors)
			}
		}
	})
}

// Helper functions are defined in validator_test.go