package toggle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// setupTestEngine creates a test engine with a temporary config directory
func setupTestEngine(t testing.TB) (*Engine, string, func()) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create apps directory
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app config
	testConfig := `name: test-app
path: ~/.config/test-app/appconfig.json
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light", "auto"]
    default: "dark"
    description: "Application theme"
  
  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 14
    description: "Font size"
    
  debug:
    type: boolean
    default: false
    description: "Enable debug mode"

presets:
  default:
    name: default
    description: Default settings
    values:
      theme: dark
      font-size: 14
      debug: false
      
  light-mode:
    name: light-mode
    description: Light theme setup
    values:
      theme: light
      font-size: 16
      debug: false

hooks:
  post-toggle: "echo 'Config updated'"
`

	configPath := filepath.Join(appsDir, "test-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create test target config file
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	targetConfigPath := filepath.Join(targetDir, "appconfig.json")
	targetConfig := `{
  "theme": "dark",
  "font-size": 14,
  "debug": false
}`

	if err := os.WriteFile(targetConfigPath, []byte(targetConfig), 0o644); err != nil {
		t.Fatalf("Failed to write target config: %v", err)
	}

	// Update test config to point to actual target file
	updatedConfig := `name: test-app
path: ` + targetConfigPath + `
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light", "auto"]
    default: "dark"
    description: "Application theme"
  
  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 14
    description: "Font size"
    
  debug:
    type: boolean
    default: false
    description: "Enable debug mode"

presets:
  default:
    name: default
    description: Default settings
    values:
      theme: dark
      font-size: 14
      debug: false
      
  light-mode:
    name: light-mode
    description: Light theme setup
    values:
      theme: light
      font-size: 16
      debug: false

hooks:
  post-toggle: "echo 'Config updated'"
`

	if err := os.WriteFile(configPath, []byte(updatedConfig), 0o644); err != nil {
		t.Fatalf("Failed to update test config: %v", err)
	}

	// Create loader and engine
	loader := &appconfig.Loader{}
	loader.SetConfigDir(tmpDir) // Assume we add this method

	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	engine.loader = loader

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return engine, tmpDir, cleanup
}

// TestEngine_Toggle tests the basic toggle functionality
func TestEngine_Toggle(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test valid toggle
	err := engine.Toggle("test-app", "theme", "light")
	if err != nil {
		t.Fatalf("Failed to toggle theme: %v", err)
	}

	// Test invalid app
	err = engine.Toggle("nonexistent", "theme", "light")
	if err == nil {
		t.Error("Expected error for non-existent app")
	}

	ctErr, ok := err.(*errors.ZeroUIError)
	if !ok {
		t.Error("Expected ZeroUIError for non-existent app")
	} else if ctErr.Type != errors.AppNotFound {
		t.Errorf("Expected AppNotFound error, got %s", ctErr.Type)
	}

	// Test invalid field
	err = engine.Toggle("test-app", "nonexistent", "value")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}

	ctErr, ok = err.(*errors.ZeroUIError)
	if !ok {
		t.Error("Expected ZeroUIError for non-existent field")
	} else if ctErr.Type != errors.FieldNotFound {
		t.Errorf("Expected FieldNotFound error, got %s", ctErr.Type)
	}

	// Test invalid value
	err = engine.Toggle("test-app", "theme", "invalid")
	if err == nil {
		t.Error("Expected error for invalid value")
	}

	ctErr, ok = err.(*errors.ZeroUIError)
	if !ok {
		t.Error("Expected ZeroUIError for invalid value")
	} else if ctErr.Type != errors.FieldInvalidValue {
		t.Errorf("Expected FieldInvalidValue error, got %s", ctErr.Type)
	}
}

// TestEngine_Cycle tests the cycle functionality
func TestEngine_Cycle(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test valid cycle
	err := engine.Cycle("test-app", "theme")
	if err != nil {
		t.Fatalf("Failed to cycle theme: %v", err)
	}

	// Test invalid app
	err = engine.Cycle("nonexistent", "theme")
	if err == nil {
		t.Error("Expected error for non-existent app")
	}

	// Test invalid field
	err = engine.Cycle("test-app", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}

	// Test field without values (can't cycle)
	err = engine.Cycle("test-app", "debug")
	if err == nil {
		t.Error("Expected error for field without predefined values")
	}
}

// TestEngine_ApplyPreset tests the preset application functionality
func TestEngine_ApplyPreset(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test valid preset
	err := engine.ApplyPreset("test-app", "light-mode")
	if err != nil {
		t.Fatalf("Failed to apply preset: %v", err)
	}

	// Test invalid app
	err = engine.ApplyPreset("nonexistent", "default")
	if err == nil {
		t.Error("Expected error for non-existent app")
	}

	ctErr, ok := err.(*errors.ZeroUIError)
	if !ok {
		t.Error("Expected ZeroUIError for non-existent app")
	} else if ctErr.Type != errors.AppNotFound {
		t.Errorf("Expected AppNotFound error, got %s", ctErr.Type)
	}

	// Test invalid preset
	err = engine.ApplyPreset("test-app", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent preset")
	}

	ctErr, ok = err.(*errors.ZeroUIError)
	if !ok {
		t.Error("Expected ZeroUIError for non-existent preset")
	} else if ctErr.Type != errors.PresetNotFound {
		t.Errorf("Expected PresetNotFound error, got %s", ctErr.Type)
	}
}

// TestEngine_GetApps tests listing applications
func TestEngine_GetApps(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	apps, err := engine.GetApps()
	if err != nil {
		t.Fatalf("Failed to get apps: %v", err)
	}

	if len(apps) != 1 {
		t.Errorf("Expected 1 app, got %d", len(apps))
	}

	if apps[0] != "test-app" {
		t.Errorf("Expected 'test-app', got '%s'", apps[0])
	}
}

// TestEngine_GetAppConfig tests getting app configuration
func TestEngine_GetAppConfig(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	appConfig, err := engine.GetAppConfig("test-app")
	if err != nil {
		t.Fatalf("Failed to get app config: %v", err)
	}

	if appConfig.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", appConfig.Name)
	}

	if len(appConfig.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(appConfig.Fields))
	}

	if len(appConfig.Presets) != 2 {
		t.Errorf("Expected 2 presets, got %d", len(appConfig.Presets))
	}
}

// TestEngine_GetCurrentValues tests getting current config values
func TestEngine_GetCurrentValues(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	values, err := engine.GetCurrentValues("test-app")
	if err != nil {
		t.Fatalf("Failed to get current values: %v", err)
	}

	if len(values) == 0 {
		t.Error("Expected some current values")
	}

	// Check specific values
	if theme, exists := values["theme"]; !exists {
		t.Error("Expected theme value to exist")
	} else if theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%v'", theme)
	}
}

// BenchmarkEngine_Toggle benchmarks the toggle operation
func BenchmarkEngine_Toggle(b *testing.B) {
	engine, _, cleanup := setupTestEngine(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between dark and light themes
		theme := "dark"
		if i%2 == 1 {
			theme = "light"
		}

		err := engine.Toggle("test-app", "theme", theme)
		if err != nil {
			b.Fatalf("Failed to toggle: %v", err)
		}
	}
}

// BenchmarkEngine_ApplyPreset benchmarks preset application
func BenchmarkEngine_ApplyPreset(b *testing.B) {
	engine, _, cleanup := setupTestEngine(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between presets
		preset := "default"
		if i%2 == 1 {
			preset = "light-mode"
		}

		err := engine.ApplyPreset("test-app", preset)
		if err != nil {
			b.Fatalf("Failed to apply preset: %v", err)
		}
	}
}

// TestEngine_EdgeCases tests various edge cases and error scenarios
func TestEngine_EdgeCases(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test toggle with boolean field
	t.Run("Toggle boolean field", func(t *testing.T) {
		err := engine.Toggle("test-app", "debug", "true")
		if err != nil {
			t.Fatalf("Failed to toggle boolean field: %v", err)
		}

		// Toggle to false
		err = engine.Toggle("test-app", "debug", "false")
		if err != nil {
			t.Fatalf("Failed to toggle boolean field to false: %v", err)
		}
	})

	// Test toggle with number field
	t.Run("Toggle number field", func(t *testing.T) {
		err := engine.Toggle("test-app", "font-size", "16")
		if err != nil {
			t.Fatalf("Failed to toggle number field: %v", err)
		}
	})

	// Test field type conversion errors
	t.Run("Type conversion errors", func(t *testing.T) {
		// Create engine with invalid type conversion scenario
		// This would need special setup with a field that has invalid default
		t.Skip("Type conversion error testing requires special setup")
	})

	// Test config loading edge cases
	t.Run("Config loading edge cases", func(t *testing.T) {
		// Test with corrupted target config
		// This requires creating a corrupted config file
		t.Skip("Corrupted config testing requires special setup")
	})
}

// TestEngine_ConvertValue tests the value conversion function
func TestEngine_ConvertValue(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	tests := []struct {
		name      string
		value     string
		fieldType string
		expected  interface{}
		expectErr bool
	}{
		{"Boolean true", "true", "boolean", true, false},
		{"Boolean false", "false", "boolean", false, false},
		{"Boolean invalid", "maybe", "boolean", nil, true},
		{"Number integer", "42", "number", int64(42), false},
		{"Number float", "42.5", "number", 42.5, false},
		{"Number invalid", "not-a-number", "number", nil, true},
		{"String value", "hello world", "string", "hello world", false},
		{"Choice value", "option1", "choice", "option1", false},
		{"Unknown type", "value", "unknown", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.convertValue(tt.value, tt.fieldType)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

// TestEngine_RunHooks tests hook execution
func TestEngine_RunHooks(t *testing.T) {
	engine, tmpDir, cleanup := setupTestEngine(t)
	defer cleanup()

	// Create app config with hooks
	appsDir := filepath.Join(tmpDir, "apps")
	configWithHooks := `name: hook-test
path: ` + filepath.Join(tmpDir, "hook-appconfig.json") + `
format: json

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"

presets:
  default:
    name: default
    values:
      theme: dark

hooks:
  post-toggle: "echo 'Toggle hook executed'"
  post-preset: "echo 'Preset hook executed'"

env:
  TEST_VAR: "test_value"
`

	hookConfigPath := filepath.Join(appsDir, "hook-test.yaml")
	if err := os.WriteFile(hookConfigPath, []byte(configWithHooks), 0o644); err != nil {
		t.Fatalf("Failed to write hook config: %v", err)
	}

	// Create target config
	targetConfig := `{"theme": "dark"}`
	targetPath := filepath.Join(tmpDir, "hook-appconfig.json")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0o644); err != nil {
		t.Fatalf("Failed to write hook target config: %v", err)
	}

	// Test toggle with hook
	t.Run("Toggle with hook", func(t *testing.T) {
		err := engine.Toggle("hook-test", "theme", "light")
		if err != nil {
			t.Fatalf("Failed to toggle with hook: %v", err)
		}
	})

	// Test preset with hook
	t.Run("Preset with hook", func(t *testing.T) {
		err := engine.ApplyPreset("hook-test", "default")
		if err != nil {
			t.Fatalf("Failed to apply preset with hook: %v", err)
		}
	})

	// Test hook with invalid command
	t.Run("Invalid hook command", func(t *testing.T) {
		configWithBadHook := `name: bad-hook-test
path: ` + filepath.Join(tmpDir, "bad-hook-appconfig.json") + `
format: json

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"

hooks:
  post-toggle: "nonexistent-command-that-should-fail"
`

		badHookConfigPath := filepath.Join(appsDir, "bad-hook-test.yaml")
		if err := os.WriteFile(badHookConfigPath, []byte(configWithBadHook), 0o644); err != nil {
			t.Fatalf("Failed to write bad hook config: %v", err)
		}

		badTargetConfig := `{"theme": "dark"}`
		badTargetPath := filepath.Join(tmpDir, "bad-hook-appconfig.json")
		if err := os.WriteFile(badTargetPath, []byte(badTargetConfig), 0o644); err != nil {
			t.Fatalf("Failed to write bad hook target config: %v", err)
		}

		// This should fail due to bad hook
		err := engine.Toggle("bad-hook-test", "theme", "light")
		if err == nil {
			t.Error("Expected error from bad hook, but got none")
		}
	})
}

// TestEngine_ListMethods tests the various list methods
func TestEngine_ListMethods(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test ListApps (prints to stdout)
	t.Run("ListApps", func(t *testing.T) {
		err := engine.ListApps()
		if err != nil {
			t.Fatalf("Failed to list apps: %v", err)
		}
	})

	// Test ListPresets
	t.Run("ListPresets", func(t *testing.T) {
		err := engine.ListPresets("test-app")
		if err != nil {
			t.Fatalf("Failed to list presets: %v", err)
		}
	})

	// Test ListKeys
	t.Run("ListKeys", func(t *testing.T) {
		err := engine.ListKeys("test-app")
		if err != nil {
			t.Fatalf("Failed to list keys: %v", err)
		}
	})

	// Test with non-existent app
	t.Run("ListPresets non-existent app", func(t *testing.T) {
		err := engine.ListPresets("nonexistent-app")
		if err == nil {
			t.Error("Expected error for non-existent app")
		}
	})

	t.Run("ListKeys non-existent app", func(t *testing.T) {
		err := engine.ListKeys("nonexistent-app")
		if err == nil {
			t.Error("Expected error for non-existent app")
		}
	})
}

// TestEngine_CycleEdgeCases tests edge cases in cycle functionality
func TestEngine_CycleEdgeCases(t *testing.T) {
	engine, tmpDir, cleanup := setupTestEngine(t)
	defer cleanup()

	// Create app with single-value field
	appsDir := filepath.Join(tmpDir, "apps")
	singleValueConfig := `name: single-value-test
path: ` + filepath.Join(tmpDir, "single-appconfig.json") + `
format: json

fields:
  single-choice:
    type: choice
    values: ["only-option"]
    default: "only-option"
    
  no-values:
    type: string
    description: "Field without predefined values"
`

	singleConfigPath := filepath.Join(appsDir, "single-value-test.yaml")
	if err := os.WriteFile(singleConfigPath, []byte(singleValueConfig), 0o644); err != nil {
		t.Fatalf("Failed to write single value config: %v", err)
	}

	// Create target config
	targetConfig := `{"single-choice": "only-option", "no-values": "some-value"}`
	targetPath := filepath.Join(tmpDir, "single-appconfig.json")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0o644); err != nil {
		t.Fatalf("Failed to write single target config: %v", err)
	}

	// Test cycle with single option (should wrap around to same value)
	t.Run("Cycle single option", func(t *testing.T) {
		err := engine.Cycle("single-value-test", "single-choice")
		if err != nil {
			t.Fatalf("Failed to cycle single option: %v", err)
		}
	})

	// Test cycle with no predefined values
	t.Run("Cycle no values", func(t *testing.T) {
		err := engine.Cycle("single-value-test", "no-values")
		if err == nil {
			t.Error("Expected error for field with no predefined values")
		}
	})
}

// TestEngine_PresetEdgeCases tests edge cases in preset functionality
func TestEngine_PresetEdgeCases(t *testing.T) {
	engine, tmpDir, cleanup := setupTestEngine(t)
	defer cleanup()

	// Create app with preset that has unknown fields
	appsDir := filepath.Join(tmpDir, "apps")
	unknownFieldConfig := `name: unknown-field-test
path: ` + filepath.Join(tmpDir, "unknown-appconfig.json") + `
format: json

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"

presets:
  with-unknown:
    name: with-unknown
    description: "Preset with unknown fields"
    values:
      theme: light
      unknown-field: "unknown-value"
      another-unknown: 42
`

	unknownConfigPath := filepath.Join(appsDir, "unknown-field-test.yaml")
	if err := os.WriteFile(unknownConfigPath, []byte(unknownFieldConfig), 0o644); err != nil {
		t.Fatalf("Failed to write unknown field config: %v", err)
	}

	// Create target config
	targetConfig := `{"theme": "dark"}`
	targetPath := filepath.Join(tmpDir, "unknown-appconfig.json")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0o644); err != nil {
		t.Fatalf("Failed to write unknown target config: %v", err)
	}

	// Test preset with unknown fields (should still work)
	t.Run("Preset with unknown fields", func(t *testing.T) {
		err := engine.ApplyPreset("unknown-field-test", "with-unknown")
		if err != nil {
			t.Fatalf("Failed to apply preset with unknown fields: %v", err)
		}
	})
}

func TestEngine_GetPresets(t *testing.T) {
	engine, _, cleanup := setupTestEngine(t)
	defer cleanup()

	// Test with existing app
	presets, err := engine.GetPresets("test-app")
	if err != nil {
		t.Fatalf("GetPresets failed: %v", err)
	}

	// Should return at least the "development" preset from test config
	if len(presets) == 0 {
		t.Error("Expected at least one preset, got none")
	}

	// Test with non-existent app
	_, err = engine.GetPresets("non-existent-app")
	if err == nil {
		t.Error("Expected error for non-existent app, got nil")
	}
}

// TestEngine_NewEngineError tests NewEngine error scenarios
func TestEngine_NewEngineError(t *testing.T) {
	// This would require mocking or creating invalid conditions
	// For now, just test the happy path
	t.Run("Valid creation", func(t *testing.T) {
		engine, err := NewEngine()
		if err != nil {
			t.Fatalf("Failed to create new engine: %v", err)
		}

		if engine.loader == nil {
			t.Error("Expected loader to be initialized")
		}
	})
}
