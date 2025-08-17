package reference

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStaticConfigLoader(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()

	// Create a test config file
	testConfig := `
app_name: "test_app"
config_path: "~/.test/config"
config_type: "json"

settings:
  test_setting:
    name: "test_setting"
    type: "string"
    description: "A test setting"
    default_value: "default"
    valid_values: ["option1", "option2"]
    category: "test"
`

	configFile := filepath.Join(tempDir, "test_app.yaml")
	if err := os.WriteFile(configFile, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test loader
	loader := NewStaticConfigLoader(tempDir)
	ref, err := loader.LoadReference("test_app")
	if err != nil {
		t.Fatalf("Failed to load reference: %v", err)
	}

	// Validate loaded data
	if ref.AppName != "test_app" {
		t.Errorf("Expected app name 'test_app', got '%s'", ref.AppName)
	}

	if len(ref.Settings) != 1 {
		t.Errorf("Expected 1 setting, got %d", len(ref.Settings))
	}

	setting, exists := ref.Settings["test_setting"]
	if !exists {
		t.Error("Expected 'test_setting' to exist")
	}

	if setting.Type != TypeString {
		t.Errorf("Expected type 'string', got '%s'", setting.Type)
	}

	if len(setting.ValidValues) != 2 {
		t.Errorf("Expected 2 valid values, got %d", len(setting.ValidValues))
	}
}

func TestReferenceManager(t *testing.T) {
	// Create temporary config directory with test data
	tempDir := t.TempDir()

	testConfig := `
app_name: "test_app"
config_path: "~/.test/config"
config_type: "toml"

settings:
  string_setting:
    name: "string_setting"
    type: "string"
    description: "String setting"
    category: "basic"
    
  number_setting:
    name: "number_setting"
    type: "number"
    description: "Number setting"
    default_value: 42
    category: "basic"
    
  boolean_setting:
    name: "boolean_setting"
    type: "boolean"
    description: "Boolean setting"
    default_value: true
    category: "basic"
    
  enum_setting:
    name: "enum_setting"
    type: "string"
    description: "Enum setting"
    valid_values: ["option1", "option2", "option3"]
    category: "advanced"
`

	configFile := filepath.Join(tempDir, "test_app.yaml")
	if err := os.WriteFile(configFile, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create manager
	loader := NewStaticConfigLoader(tempDir)
	manager := NewReferenceManager(loader)

	t.Run("GetReference", func(t *testing.T) {
		ref, err := manager.GetReference("test_app")
		if err != nil {
			t.Fatalf("Failed to get reference: %v", err)
		}

		if ref.AppName != "test_app" {
			t.Errorf("Expected app name 'test_app', got '%s'", ref.AppName)
		}

		if len(ref.Settings) != 4 {
			t.Errorf("Expected 4 settings, got %d", len(ref.Settings))
		}

		// Test caching - second call should use cache
		ref2, err := manager.GetReference("test_app")
		if err != nil {
			t.Fatalf("Failed to get cached reference: %v", err)
		}

		if ref != ref2 {
			t.Error("Expected cached reference to be the same instance")
		}
	})

	t.Run("ValidateConfiguration", func(t *testing.T) {
		// Valid string
		result, err := manager.ValidateConfiguration("test_app", "string_setting", "test value")
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if !result.Valid {
			t.Error("Expected string validation to be valid")
		}

		// Valid number
		result, err = manager.ValidateConfiguration("test_app", "number_setting", 123)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if !result.Valid {
			t.Error("Expected number validation to be valid")
		}

		// Valid boolean
		result, err = manager.ValidateConfiguration("test_app", "boolean_setting", true)
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if !result.Valid {
			t.Error("Expected boolean validation to be valid")
		}

		// Valid enum value
		result, err = manager.ValidateConfiguration("test_app", "enum_setting", "option1")
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if !result.Valid {
			t.Error("Expected valid enum value to be valid")
		}

		// Invalid enum value
		result, err = manager.ValidateConfiguration("test_app", "enum_setting", "invalid_option")
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if result.Valid {
			t.Error("Expected invalid enum value to be invalid")
		}
		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for invalid enum value")
		}

		// Invalid setting name (similar to existing "string_setting")
		result, err = manager.ValidateConfiguration("test_app", "string", "value")
		if err != nil {
			t.Fatalf("Validation failed: %v", err)
		}
		if result.Valid {
			t.Error("Expected nonexistent setting to be invalid")
		}
		if len(result.Suggestions) == 0 {
			t.Error("Expected suggestions for similar settings")
		}
	})

	t.Run("SearchSettings", func(t *testing.T) {
		// Search by name
		results, err := manager.SearchSettings("test_app", "string")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if results[0].Name != "string_setting" {
			t.Errorf("Expected 'string_setting', got '%s'", results[0].Name)
		}

		// Search by category
		results, err = manager.SearchSettings("test_app", "basic")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		// Search with no matches
		results, err = manager.SearchSettings("test_app", "nonexistent")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})
}

func TestEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	loader := NewStaticConfigLoader(tempDir)
	manager := NewReferenceManager(loader)

	t.Run("NonexistentApp", func(t *testing.T) {
		_, err := manager.GetReference("nonexistent_app")
		if err == nil {
			t.Error("Expected error for nonexistent app")
		}
	})

	t.Run("EmptyConfigDirectory", func(t *testing.T) {
		apps, err := manager.ListApps()
		if err != nil {
			t.Fatalf("ListApps failed: %v", err)
		}
		if len(apps) != 0 {
			t.Errorf("Expected 0 apps, got %d", len(apps))
		}
	})
}

func BenchmarkReferenceOperations(b *testing.B) {
	// Setup test data
	tempDir := b.TempDir()

	testConfig := `
app_name: "bench_app"
settings:
  setting1:
    name: "setting1"
    type: "string"
    description: "Test setting 1"
  setting2:
    name: "setting2"
    type: "number"
    description: "Test setting 2"
  setting3:
    name: "setting3"
    type: "boolean"
    description: "Test setting 3"
`

	configFile := filepath.Join(tempDir, "bench_app.yaml")
	if err := os.WriteFile(configFile, []byte(testConfig), 0644); err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	loader := NewStaticConfigLoader(tempDir)
	manager := NewReferenceManager(loader)

	b.Run("GetReference", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.GetReference("bench_app")
			if err != nil {
				b.Fatalf("GetReference failed: %v", err)
			}
		}
	})

	b.Run("ValidateConfiguration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.ValidateConfiguration("bench_app", "setting1", "test")
			if err != nil {
				b.Fatalf("ValidateConfiguration failed: %v", err)
			}
		}
	})

	b.Run("SearchSettings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.SearchSettings("bench_app", "setting")
			if err != nil {
				b.Fatalf("SearchSettings failed: %v", err)
			}
		}
	})
}
