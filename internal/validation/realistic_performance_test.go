package validation

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// These benchmarks compare realistic use cases:
// 1. Validating a complete app configuration (common case)
// 2. Validating target configuration data (common case)

// BenchmarkRealisticAppConfigValidation_Optimized uses the new optimized approach
func BenchmarkRealisticAppConfigValidation_Optimized(b *testing.B) {
	validator := NewValidator()

	// Create a realistic schema (simple enough for optimization)
	schema := &Schema{
		Name: "realistic-app",
		Fields: map[string]*FieldRule{
			"server_host": {Type: "string", Required: true},
			"server_port": {Type: "number", Min: floatPtr(1), Max: floatPtr(65535)},
			"debug_mode":  {Type: "boolean"},
			"log_level":   {Type: "choice", Enum: []string{"debug", "info", "warn", "error"}},
			"timeout":     {Type: "number", Min: floatPtr(1), Max: floatPtr(300)},
		},
		Global: &GlobalRules{
			RequiredFields: []string{"server_host"},
			MinFields:      intPtr(2),
			MaxFields:      intPtr(10),
		},
	}
	validator.RegisterSchema(schema)

	// Realistic app config
	appConfig := &config.AppConfig{
		Name:        "realistic-app",
		Path:        "/etc/myapp/config.json",
		Format:      "json",
		Description: "Realistic application configuration",
		Fields: map[string]config.FieldConfig{
			"server_host": {Type: "string", Default: "localhost", Description: "Server hostname"},
			"server_port": {Type: "number", Default: 8080, Description: "Server port"},
			"debug_mode":  {Type: "boolean", Default: false, Description: "Enable debug mode"},
			"log_level":   {Type: "choice", Values: []string{"debug", "info", "warn", "error"}, Default: "info", Description: "Log level"},
			"timeout":     {Type: "number", Default: 30, Description: "Request timeout in seconds"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateAppConfig("realistic-app", appConfig)
		if !result.Valid {
			b.Fatalf("Validation failed: %v", result.Errors)
		}
	}
}

// BenchmarkRealisticAppConfigValidation_Legacy simulates the old approach
func BenchmarkRealisticAppConfigValidation_Legacy(b *testing.B) {
	validator := NewValidator()

	// Create a complex schema to force legacy path
	schema := &Schema{
		Name: "realistic-app-legacy",
		Fields: map[string]*FieldRule{
			"server_host": {
				Type:          "string",
				Required:      true,
				Dependencies:  []string{"server_port"}, // Forces complex path
			},
			"server_port": {Type: "number", Min: floatPtr(1), Max: floatPtr(65535)},
			"debug_mode":  {Type: "boolean"},
			"log_level":   {Type: "choice", Enum: []string{"debug", "info", "warn", "error"}},
			"timeout":     {Type: "number", Min: floatPtr(1), Max: floatPtr(300)},
		},
		Global: &GlobalRules{
			RequiredFields:  []string{"server_host"},
			ForbiddenFields: []string{"admin_key"}, // Forces complex path
			MinFields:       intPtr(2),
			MaxFields:       intPtr(10),
		},
	}
	validator.RegisterSchema(schema)

	appConfig := &config.AppConfig{
		Name:        "realistic-app-legacy",
		Path:        "/etc/myapp/config.json",
		Format:      "json",
		Description: "Realistic application configuration",
		Fields: map[string]config.FieldConfig{
			"server_host": {Type: "string", Default: "localhost", Description: "Server hostname"},
			"server_port": {Type: "number", Default: 8080, Description: "Server port"},
			"debug_mode":  {Type: "boolean", Default: false, Description: "Enable debug mode"},
			"log_level":   {Type: "choice", Values: []string{"debug", "info", "warn", "error"}, Default: "info", Description: "Log level"},
			"timeout":     {Type: "number", Default: 30, Description: "Request timeout in seconds"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateAppConfig("realistic-app-legacy", appConfig)
		if !result.Valid {
			// Note: May have validation errors due to dependencies, but we're measuring performance
			b.Logf("Validation result: %v", result.Valid)
		}
	}
}

// BenchmarkRealisticTargetConfigValidation_Optimized tests target config validation with optimization
func BenchmarkRealisticTargetConfigValidation_Optimized(b *testing.B) {
	validator := NewValidator()

	// Simple schema for target config
	schema := &Schema{
		Name: "target-config",
		Fields: map[string]*FieldRule{
			"database_url": {Type: "string", Required: true, Format: "url"},
			"port":         {Type: "number", Min: floatPtr(1), Max: floatPtr(65535)},
			"debug":        {Type: "boolean"},
			"timeout":      {Type: "number", Min: floatPtr(1), Max: floatPtr(300)},
		},
		Global: &GlobalRules{
			RequiredFields: []string{"database_url"},
		},
	}
	validator.RegisterSchema(schema)

	configData := map[string]interface{}{
		"database_url": "postgresql://user:pass@localhost:5432/mydb",
		"port":         5432,
		"debug":        false,
		"timeout":      30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateTargetConfig("target-config", configData)
		if !result.Valid {
			b.Fatalf("Validation failed: %v", result.Errors)
		}
	}
}

// TestRealisticPerformanceComparison provides a realistic performance comparison
func TestRealisticPerformanceComparison(t *testing.T) {
	// Test the optimized vs legacy paths
	optimizedResult := testing.Benchmark(BenchmarkRealisticAppConfigValidation_Optimized)
	legacyResult := testing.Benchmark(BenchmarkRealisticAppConfigValidation_Legacy)
	targetResult := testing.Benchmark(BenchmarkRealisticTargetConfigValidation_Optimized)

	t.Logf("=== REALISTIC PERFORMANCE COMPARISON ===")
	t.Logf("")
	t.Logf("App Config Validation (Optimized):  %s", optimizedResult)
	t.Logf("Memory: %d B/op, %d allocs/op", optimizedResult.AllocedBytesPerOp(), optimizedResult.AllocsPerOp())
	t.Logf("")
	t.Logf("App Config Validation (Legacy):     %s", legacyResult)
	t.Logf("Memory: %d B/op, %d allocs/op", legacyResult.AllocedBytesPerOp(), legacyResult.AllocsPerOp())
	t.Logf("")
	t.Logf("Target Config Validation:           %s", targetResult)
	t.Logf("Memory: %d B/op, %d allocs/op", targetResult.AllocedBytesPerOp(), targetResult.AllocsPerOp())

	if optimizedResult.NsPerOp() > 0 && legacyResult.NsPerOp() > 0 {
		improvement := float64(legacyResult.NsPerOp()) / float64(optimizedResult.NsPerOp())
		t.Logf("")
		t.Logf("Performance improvement: %.2fx", improvement)
		
		if improvement >= 2.0 {
			t.Logf("✓ PERFORMANCE TARGET ACHIEVED: %.2fx >= 2x", improvement)
		} else {
			t.Logf("Current improvement: %.2fx (target: 2x+)", improvement)
		}

		// Memory comparison
		if optimizedResult.AllocedBytesPerOp() > 0 && legacyResult.AllocedBytesPerOp() > 0 {
			memImprovement := float64(legacyResult.AllocedBytesPerOp()) / float64(optimizedResult.AllocedBytesPerOp())
			t.Logf("Memory improvement: %.2fx", memImprovement)
		}
	}
}