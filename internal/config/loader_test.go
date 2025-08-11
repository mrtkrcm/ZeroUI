package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/knadh/koanf/v2"
)

// TestLoader_LoadAppConfig tests loading application configurations
func TestLoader_LoadAppConfig(t *testing.T) {
	// Create temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create apps directory
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app config
	testConfig := `name: test-app
path: ~/.config/test-app/config.json
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"
    description: "Application theme"
  
  font-size:
    type: number
    values: ["12", "14", "16"]
    default: 14
    description: "Font size"

presets:
  default:
    name: default
    description: Default settings
    values:
      theme: dark
      font-size: 14

hooks:
  post-toggle: "echo 'Config updated'"
`

	configPath := filepath.Join(appsDir, "test-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create loader with custom config directory
	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	// Test loading app config
	appConfig, err := loader.LoadAppConfig("test-app")
	if err != nil {
		t.Fatalf("Failed to load app config: %v", err)
	}

	// Verify app config
	if appConfig.Name != "test-app" {
		t.Errorf("Expected app name 'test-app', got '%s'", appConfig.Name)
	}

	if appConfig.Path != "~/.config/test-app/config.json" {
		t.Errorf("Expected path '~/.config/test-app/config.json', got '%s'", appConfig.Path)
	}

	if appConfig.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", appConfig.Format)
	}

	// Test fields
	if len(appConfig.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(appConfig.Fields))
	}

	themeField, exists := appConfig.Fields["theme"]
	if !exists {
		t.Error("Expected 'theme' field to exist")
	} else {
		if themeField.Type != "choice" {
			t.Errorf("Expected theme type 'choice', got '%s'", themeField.Type)
		}
		expectedValues := []string{"dark", "light"}
		if !reflect.DeepEqual(themeField.Values, expectedValues) {
			t.Errorf("Expected theme values %v, got %v", expectedValues, themeField.Values)
		}
	}

	// Test presets
	if len(appConfig.Presets) != 1 {
		t.Errorf("Expected 1 preset, got %d", len(appConfig.Presets))
	}

	defaultPreset, exists := appConfig.Presets["default"]
	if !exists {
		t.Error("Expected 'default' preset to exist")
	} else {
		if defaultPreset.Name != "default" {
			t.Errorf("Expected preset name 'default', got '%s'", defaultPreset.Name)
		}
	}

	// Test hooks
	if len(appConfig.Hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(appConfig.Hooks))
	}

	hook, exists := appConfig.Hooks["post-toggle"]
	if !exists {
		t.Error("Expected 'post-toggle' hook to exist")
	} else if hook != "echo 'Config updated'" {
		t.Errorf("Expected hook 'echo 'Config updated'', got '%s'", hook)
	}
}

// TestLoader_LoadAppConfig_NotFound tests loading non-existent app config
func TestLoader_LoadAppConfig_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	_, err = loader.LoadAppConfig("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent app config")
	}
}

// TestLoader_ListApps tests listing applications
func TestLoader_ListApps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create apps directory
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app configs
	testApps := []string{"app1", "app2", "app3"}
	for _, app := range testApps {
		configContent := fmt.Sprintf("name: %s\npath: ~/.config/%s/config.json", app, app)
		configPath := filepath.Join(appsDir, app+".yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config for %s: %v", app, err)
		}
	}

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	apps, err := loader.ListApps()
	if err != nil {
		t.Fatalf("Failed to list apps: %v", err)
	}

	if len(apps) != len(testApps) {
		t.Errorf("Expected %d apps, got %d", len(testApps), len(apps))
	}

	// Check that all test apps are present
	appSet := make(map[string]bool)
	for _, app := range apps {
		appSet[app] = true
	}

	for _, expectedApp := range testApps {
		if !appSet[expectedApp] {
			t.Errorf("Expected app '%s' not found in results", expectedApp)
		}
	}
}

// TestLoader_ListApps_Empty tests listing apps when directory is empty
func TestLoader_ListApps_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	apps, err := loader.ListApps()
	if err != nil {
		t.Fatalf("Failed to list apps: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("Expected 0 apps, got %d", len(apps))
	}
}

// TestNewLoader tests creating a new loader
func TestNewLoader(t *testing.T) {
	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	if loader.configDir == "" {
		t.Error("Expected config directory to be set")
	}

	// Check that config directory exists
	if _, err := os.Stat(loader.configDir); os.IsNotExist(err) {
		t.Error("Expected config directory to exist after creation")
	}
}

// BenchmarkLoader_LoadAppConfig benchmarks app config loading
func BenchmarkLoader_LoadAppConfig(b *testing.B) {
	// Create temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "configtoggle-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create apps directory and test config
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		b.Fatalf("Failed to create apps dir: %v", err)
	}

	testConfig := `name: bench-app
path: ~/.config/bench-app/config.json
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
`

	configPath := filepath.Join(appsDir, "bench-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		b.Fatalf("Failed to write test config: %v", err)
	}

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAppConfig("bench-app")
		if err != nil {
			b.Fatalf("Failed to load app config: %v", err)
		}
	}
}

// TestLoader_LoadTargetConfig tests loading target configuration files
func TestLoader_LoadTargetConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	// Test JSON format
	t.Run("JSON format", func(t *testing.T) {
		jsonPath := filepath.Join(tmpDir, "test.json")
		jsonContent := `{
  "theme": "dark",
  "font-size": 14,
  "debug": true
}`
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("Failed to write JSON file: %v", err)
		}

		appConfig := &AppConfig{
			Path:   jsonPath,
			Format: "json",
		}

		config, err := loader.LoadTargetConfig(appConfig)
		if err != nil {
			t.Fatalf("Failed to load JSON config: %v", err)
		}

		if config.String("theme") != "dark" {
			t.Errorf("Expected theme 'dark', got '%s'", config.String("theme"))
		}

		if config.Int("font-size") != 14 {
			t.Errorf("Expected font-size 14, got %d", config.Int("font-size"))
		}

		if !config.Bool("debug") {
			t.Error("Expected debug to be true")
		}
	})

	// Test YAML format
	t.Run("YAML format", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "test.yaml")
		yamlContent := `theme: light
font-size: 16
debug: false
nested:
  value: test
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		appConfig := &AppConfig{
			Path:   yamlPath,
			Format: "yaml",
		}

		config, err := loader.LoadTargetConfig(appConfig)
		if err != nil {
			t.Fatalf("Failed to load YAML config: %v", err)
		}

		if config.String("theme") != "light" {
			t.Errorf("Expected theme 'light', got '%s'", config.String("theme"))
		}

		if config.Int("font-size") != 16 {
			t.Errorf("Expected font-size 16, got %d", config.Int("font-size"))
		}

		if config.Bool("debug") {
			t.Error("Expected debug to be false")
		}

		if config.String("nested.value") != "test" {
			t.Errorf("Expected nested.value 'test', got '%s'", config.String("nested.value"))
		}
	})

	// Test custom format (Ghostty)
	t.Run("Custom format", func(t *testing.T) {
		customPath := filepath.Join(tmpDir, "ghostty.conf")
		customContent := `# Ghostty config
theme = GruvboxDark
font-family = JetBrains Mono
font-size = 14
background-opacity = 0.9
`
		if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
			t.Fatalf("Failed to write custom config file: %v", err)
		}

		appConfig := &AppConfig{
			Path:   customPath,
			Format: "custom",
		}

		config, err := loader.LoadTargetConfig(appConfig)
		if err != nil {
			t.Fatalf("Failed to load custom config: %v", err)
		}

		if config.String("theme") != "GruvboxDark" {
			t.Errorf("Expected theme 'GruvboxDark', got '%s'", config.String("theme"))
		}

		if config.String("font-family") != "JetBrains Mono" {
			t.Errorf("Expected font-family 'JetBrains Mono', got '%s'", config.String("font-family"))
		}

		if config.String("font-size") != "14" {
			t.Errorf("Expected font-size '14', got '%s'", config.String("font-size"))
		}
	})

	// Test format detection from extension
	t.Run("Format detection", func(t *testing.T) {
		jsonPath := filepath.Join(tmpDir, "auto.json")
		jsonContent := `{"detected": "json"}`
		if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("Failed to write auto-detect JSON file: %v", err)
		}

		appConfig := &AppConfig{
			Path:   jsonPath,
			Format: "", // Empty format should trigger detection
		}

		config, err := loader.LoadTargetConfig(appConfig)
		if err != nil {
			t.Fatalf("Failed to load config with auto-detection: %v", err)
		}

		if config.String("detected") != "json" {
			t.Errorf("Expected detected 'json', got '%s'", config.String("detected"))
		}
	})

	// Test error cases
	t.Run("Error cases", func(t *testing.T) {
		// Non-existent file
		appConfig := &AppConfig{
			Path:   "/nonexistent/file.json",
			Format: "json",
		}

		_, err := loader.LoadTargetConfig(appConfig)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		// Unsupported format
		appConfig2 := &AppConfig{
			Path:   filepath.Join(tmpDir, "test.json"),
			Format: "unsupported",
		}

		_, err = loader.LoadTargetConfig(appConfig2)
		if err == nil {
			t.Error("Expected error for unsupported format")
		}

		// Invalid JSON
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(invalidPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid JSON: %v", err)
		}

		appConfig3 := &AppConfig{
			Path:   invalidPath,
			Format: "json",
		}

		_, err = loader.LoadTargetConfig(appConfig3)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestLoader_SaveTargetConfig tests saving target configuration files
func TestLoader_SaveTargetConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtoggle-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	loader := &Loader{
		configDir: tmpDir,
	}
	loader.SetConfigDir(tmpDir) // Initialize yamlValidator for testing

	// Test JSON format
	t.Run("JSON format", func(t *testing.T) {
		jsonPath := filepath.Join(tmpDir, "save-test.json")

		// Create koanf config
		k := koanf.New(".")
		k.Set("theme", "dark")
		k.Set("font-size", 16)
		k.Set("debug", true)
		k.Set("nested.value", "test")

		appConfig := &AppConfig{
			Path:   jsonPath,
			Format: "json",
		}

		err := loader.SaveTargetConfig(appConfig, k)
		if err != nil {
			t.Fatalf("Failed to save JSON config: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			t.Error("Expected JSON file to be created")
		}

		// Verify content by loading it back
		content, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("Failed to read saved JSON: %v", err)
		}

		// Parse back to verify
		var parsed map[string]interface{}
		if err := json.Unmarshal(content, &parsed); err != nil {
			t.Fatalf("Failed to parse saved JSON: %v", err)
		}

		if parsed["theme"] != "dark" {
			t.Errorf("Expected theme 'dark', got '%v'", parsed["theme"])
		}
	})

	// Test YAML format
	t.Run("YAML format", func(t *testing.T) {
		yamlPath := filepath.Join(tmpDir, "save-test.yaml")

		k := koanf.New(".")
		k.Set("theme", "light")
		k.Set("font-size", 18)
		k.Set("list", []string{"a", "b", "c"})

		appConfig := &AppConfig{
			Path:   yamlPath,
			Format: "yaml",
		}

		err := loader.SaveTargetConfig(appConfig, k)
		if err != nil {
			t.Fatalf("Failed to save YAML config: %v", err)
		}

		// Verify content
		content, err := os.ReadFile(yamlPath)
		if err != nil {
			t.Fatalf("Failed to read saved YAML: %v", err)
		}

		if !strings.Contains(string(content), "theme: light") {
			t.Error("Expected 'theme: light' in YAML content")
		}
	})

	// Test custom format
	t.Run("Custom format", func(t *testing.T) {
		customPath := filepath.Join(tmpDir, "save-test.conf")

		k := koanf.New(".")
		k.Set("theme", "GruvboxDark")
		k.Set("font-family", "SF Mono")
		k.Set("font-size", "14")

		appConfig := &AppConfig{
			Path:   customPath,
			Format: "custom",
		}

		err := loader.SaveTargetConfig(appConfig, k)
		if err != nil {
			t.Fatalf("Failed to save custom config: %v", err)
		}

		// Verify content
		content, err := os.ReadFile(customPath)
		if err != nil {
			t.Fatalf("Failed to read saved custom config: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected 'theme = GruvboxDark' in custom content")
		}
		if !strings.Contains(contentStr, "font-family = SF Mono") {
			t.Error("Expected 'font-family = SF Mono' in custom content")
		}
	})

	// Test error cases
	t.Run("Error cases", func(t *testing.T) {
		k := koanf.New(".")
		k.Set("test", "value")

		// Unsupported format
		appConfig := &AppConfig{
			Path:   filepath.Join(tmpDir, "unsupported.xyz"),
			Format: "unsupported",
		}

		err := loader.SaveTargetConfig(appConfig, k)
		if err == nil {
			t.Error("Expected error for unsupported format")
		}
	})
}
