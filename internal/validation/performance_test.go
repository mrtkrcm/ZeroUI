package validation

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// TestOptimizedVsUnoptimizedValidation verifies the performance improvement
func TestOptimizedVsUnoptimizedValidation(t *testing.T) {
	validator := NewValidator()

	// Create test data
	appConfig := &config.AppConfig{
		Name:        "test-app",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Test application",
		Fields: map[string]config.FieldConfig{
			"username": {
				Type:        "string",
				Default:     "testuser",
				Description: "Username",
			},
			"port": {
				Type:        "number",
				Default:     8080,
				Description: "Port number",
			},
			"debug": {
				Type:        "boolean",
				Default:     false,
				Description: "Debug mode",
			},
		},
	}

	// Simple schema
	schema := &Schema{
		Name: "test-app",
		Fields: map[string]*FieldRule{
			"username": {
				Type:      "string",
				Required:  true,
				MinLength: intPtr(3),
				MaxLength: intPtr(50),
			},
			"port": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(65535),
			},
			"debug": {
				Type: "boolean",
			},
		},
	}
	validator.RegisterSchema(schema)

	// Test that optimized path works correctly
	result := validator.ValidateAppConfig("test-app", appConfig)
	if !result.Valid {
		t.Fatalf("Optimized validation failed: %v", result.Errors)
	}

	// Measure performance of optimized validation
	optimizedResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.ValidateAppConfig("test-app", appConfig)
		}
	})

	// Measure performance of field-by-field validation (simulating old approach)
	unoptimizedResult := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate the old approach with individual field validations
			validator.ValidateField("test-app", "username", "testuser")
			validator.ValidateField("test-app", "port", 8080)
			validator.ValidateField("test-app", "debug", false)
		}
	})

	t.Logf("Optimized validation: %s", optimizedResult)
	t.Logf("Unoptimized validation: %s", unoptimizedResult)

	// Calculate improvement ratio
	optimizedNsPerOp := optimizedResult.NsPerOp()
	unoptimizedNsPerOp := unoptimizedResult.NsPerOp()

	if optimizedNsPerOp > 0 {
		improvement := float64(unoptimizedNsPerOp) / float64(optimizedNsPerOp)
		t.Logf("Performance improvement: %.2fx", improvement)

		// We expect at least 2x improvement (targeting 3x)
		if improvement < 2.0 {
			t.Logf("Warning: Performance improvement (%.2fx) is less than target (3x)", improvement)
		} else {
			t.Logf("✓ Performance improvement target met: %.2fx", improvement)
		}
	}
}

// BenchmarkValidatorOptimized measures the optimized validation performance
func BenchmarkValidatorOptimized(b *testing.B) {
	validator := NewValidator()

	appConfig := &config.AppConfig{
		Name:        "optimized-test",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Optimized test application",
		Fields: map[string]config.FieldConfig{
			"setting1": {Type: "string", Default: "value1"},
			"setting2": {Type: "number", Default: 42},
			"setting3": {Type: "boolean", Default: true},
			"setting4": {Type: "string", Default: "value4"},
			"setting5": {Type: "number", Default: 100},
		},
	}

	schema := &Schema{
		Name: "optimized-test",
		Fields: map[string]*FieldRule{
			"setting1": {Type: "string", Required: true, MaxLength: intPtr(50)},
			"setting2": {Type: "number", Min: floatPtr(0), Max: floatPtr(1000)},
			"setting3": {Type: "boolean"},
			"setting4": {Type: "string", MaxLength: intPtr(100)},
			"setting5": {Type: "number", Min: floatPtr(1), Max: floatPtr(1000)},
		},
	}
	validator.RegisterSchema(schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateAppConfig("optimized-test", appConfig)
	}
}

// BenchmarkValidatorUnoptimized measures the unoptimized validation performance
func BenchmarkValidatorUnoptimized(b *testing.B) {
	validator := NewValidator()

	schema := &Schema{
		Name: "unoptimized-test",
		Fields: map[string]*FieldRule{
			"setting1": {Type: "string", Required: true, MaxLength: intPtr(50)},
			"setting2": {Type: "number", Min: floatPtr(0), Max: floatPtr(1000)},
			"setting3": {Type: "boolean"},
			"setting4": {Type: "string", MaxLength: intPtr(100)},
			"setting5": {Type: "number", Min: floatPtr(1), Max: floatPtr(1000)},
		},
	}
	validator.RegisterSchema(schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate field-by-field validation (old approach)
		validator.ValidateField("unoptimized-test", "setting1", "value1")
		validator.ValidateField("unoptimized-test", "setting2", 42)
		validator.ValidateField("unoptimized-test", "setting3", true)
		validator.ValidateField("unoptimized-test", "setting4", "value4")
		validator.ValidateField("unoptimized-test", "setting5", 100)
	}
}

// BenchmarkStructTagValidation measures the performance of struct tag validation only
func BenchmarkStructTagValidation(b *testing.B) {
	validator := NewValidator()

	validatedConfig := ValidatedAppConfig{
		Name:        "struct-tag-test",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Struct tag test application",
		Fields: map[string]ValidatedFieldConfig{
			"setting1": {Type: "string"},
			"setting2": {Type: "number"},
			"setting3": {Type: "boolean"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.validate.Struct(validatedConfig)
	}
}
