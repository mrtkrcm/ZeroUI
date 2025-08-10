package validation

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// BenchmarkOptimizedValidation tests the new optimized approach
func BenchmarkOptimizedValidation(b *testing.B) {
	validator := NewValidator()

	// Create test app config
	appConfig := &config.AppConfig{
		Name:        "perf-test",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Performance test application",
		Fields: map[string]config.FieldConfig{
			"username": {
				Type:        "string",
				Default:     "testuser",
				Description: "Username",
			},
			"port": {
				Type:        "number",
				Default:     8080,
				Description: "Port",
			},
			"debug": {
				Type:        "boolean",
				Default:     false,
				Description: "Debug mode",
			},
			"email": {
				Type:        "string",
				Default:     "user@example.com",
				Description: "Email",
			},
			"timeout": {
				Type:        "number",
				Default:     30,
				Description: "Timeout",
			},
		},
	}

	// Create schema for the test
	schema := &Schema{
		Name: "perf-test",
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
			"email": {
				Type:   "string",
				Format: "email",
			},
			"timeout": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(300),
			},
		},
		Global: &GlobalRules{
			MinFields:      intPtr(2),
			MaxFields:      intPtr(10),
			RequiredFields: []string{"username"},
		},
	}
	validator.RegisterSchema(schema)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateAppConfig("perf-test", appConfig)
		if !result.Valid {
			b.Fatalf("Validation failed: %v", result.Errors)
		}
	}
}

// BenchmarkLegacyValidation simulates the old approach without struct tags
func BenchmarkLegacyValidation(b *testing.B) {
	validator := NewValidator()

	// Create the same schema
	schema := &Schema{
		Name: "perf-test-legacy",
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
			"email": {
				Type:   "string",
				Format: "email",
			},
			"timeout": {
				Type: "number",
				Min:  floatPtr(1),
				Max:  floatPtr(300),
			},
		},
		Global: &GlobalRules{
			MinFields:      intPtr(2),
			MaxFields:      intPtr(10),
			RequiredFields: []string{"username"},
		},
	}
	validator.RegisterSchema(schema)

	// Create test data to validate
	configData := map[string]interface{}{
		"username": "testuser",
		"port":     8080,
		"debug":    false,
		"email":    "user@example.com",
		"timeout":  30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Force use of the legacy path (without fast struct validation)
		result := validator.ValidateTargetConfig("perf-test-legacy", configData)
		if !result.Valid {
			b.Fatalf("Validation failed: %v", result.Errors)
		}
	}
}

// BenchmarkPureStructTagValidation measures just the struct tag validation
func BenchmarkPureStructTagValidation(b *testing.B) {
	validator := NewValidator()

	// Test data that uses pure struct validation
	testData := ValidatedAppConfig{
		Name:        "struct-test",
		Path:        "/path/to/config",
		Format:      "json",
		Description: "Test description",
		Fields: map[string]ValidatedFieldConfig{
			"field1": {Type: "string"},
			"field2": {Type: "number"},
			"field3": {Type: "boolean"},
			"field4": {Type: "string"},
			"field5": {Type: "number"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.validate.Struct(testData)
		if err != nil {
			b.Fatalf("Struct validation failed: %v", err)
		}
	}
}

// BenchmarkManualFieldValidation measures individual field validations
func BenchmarkManualFieldValidation(b *testing.B) {
	validator := NewValidator()

	schema := &Schema{
		Name: "manual-test",
		Fields: map[string]*FieldRule{
			"field1": {Type: "string", MaxLength: intPtr(50)},
			"field2": {Type: "number", Min: floatPtr(0), Max: floatPtr(1000)},
			"field3": {Type: "boolean"},
			"field4": {Type: "string", Format: "email"},
			"field5": {Type: "number", Min: floatPtr(1), Max: floatPtr(100)},
		},
	}
	validator.RegisterSchema(schema)

	testFields := []struct {
		name  string
		value interface{}
	}{
		{"field1", "test value"},
		{"field2", 500},
		{"field3", true},
		{"field4", "test@example.com"},
		{"field5", 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, field := range testFields {
			result := validator.ValidateField("manual-test", field.name, field.value)
			if !result.Valid {
				b.Fatalf("Field validation failed for %s: %v", field.name, result.Errors)
			}
		}
	}
}

// TestPerformanceComparison runs a side-by-side comparison
func TestPerformanceComparison(t *testing.T) {
	// Run optimized benchmark
	optimizedResult := testing.Benchmark(BenchmarkOptimizedValidation)
	
	// Run legacy benchmark
	legacyResult := testing.Benchmark(BenchmarkLegacyValidation)
	
	// Run struct tag benchmark
	structTagResult := testing.Benchmark(BenchmarkPureStructTagValidation)
	
	// Run manual field benchmark
	manualResult := testing.Benchmark(BenchmarkManualFieldValidation)

	t.Logf("=== Performance Comparison Results ===")
	t.Logf("Optimized validation:     %s", optimizedResult)
	t.Logf("Legacy validation:        %s", legacyResult)
	t.Logf("Pure struct tag:          %s", structTagResult)  
	t.Logf("Manual field validation:  %s", manualResult)

	// Calculate improvements
	if optimizedResult.NsPerOp() > 0 && legacyResult.NsPerOp() > 0 {
		improvement := float64(legacyResult.NsPerOp()) / float64(optimizedResult.NsPerOp())
		t.Logf("Optimized vs Legacy improvement: %.2fx", improvement)
		
		if improvement >= 2.0 {
			t.Logf("✓ Performance target met (2x+ improvement)")
		} else {
			t.Logf("⚠ Performance target not met (%.2fx < 2x)", improvement)
		}
	}
	
	if structTagResult.NsPerOp() > 0 && manualResult.NsPerOp() > 0 {
		structImprovement := float64(manualResult.NsPerOp()) / float64(structTagResult.NsPerOp())
		t.Logf("Struct tag vs Manual improvement: %.2fx", structImprovement)
	}

	// Memory allocation comparison
	t.Logf("\n=== Memory Allocation Comparison ===")
	t.Logf("Optimized: %d B/op, %d allocs/op", optimizedResult.AllocedBytesPerOp(), optimizedResult.AllocsPerOp())
	t.Logf("Legacy:    %d B/op, %d allocs/op", legacyResult.AllocedBytesPerOp(), legacyResult.AllocsPerOp())
	t.Logf("Struct:    %d B/op, %d allocs/op", structTagResult.AllocedBytesPerOp(), structTagResult.AllocsPerOp())
	t.Logf("Manual:    %d B/op, %d allocs/op", manualResult.AllocedBytesPerOp(), manualResult.AllocsPerOp())
}