package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockPlugin is a test plugin for testing
type mockPlugin struct {
	name         string
	description  string
	configPath   string
	fields       map[string]FieldMeta
	presets      map[string]Preset
	hooks        map[string]string
	failDetect   bool
	failParse    bool
	failWrite    bool
	failValidate bool
}

// newMockPlugin creates a new mock plugin
func newMockPlugin() *mockPlugin {
	return &mockPlugin{
		name:        "test-plugin",
		description: "Test plugin for unit testing",
		configPath:  "/tmp/test-config",
		fields: map[string]FieldMeta{
			"theme": {
				Type:        "choice",
				Values:      []string{"dark", "light", "auto"},
				Default:     "dark",
				Description: "Application theme",
			},
			"font-size": {
				Type:        "number",
				Values:      []string{"12", "14", "16", "18"},
				Default:     14,
				Description: "Font size",
			},
			"debug": {
				Type:        "boolean",
				Default:     false,
				Description: "Enable debug mode",
			},
		},
		presets: map[string]Preset{
			"default": {
				Name:        "default",
				Description: "Default configuration",
				Values: map[string]interface{}{
					"theme":     "dark",
					"font-size": 14,
					"debug":     false,
				},
			},
			"light-mode": {
				Name:        "light-mode",
				Description: "Light theme configuration",
				Values: map[string]interface{}{
					"theme":     "light",
					"font-size": 16,
					"debug":     false,
				},
			},
		},
		hooks: map[string]string{
			"post-toggle": "echo 'Config updated'",
			"post-preset": "echo 'Preset applied'",
		},
	}
}

// Plugin interface implementation
func (p *mockPlugin) Name() string                           { return p.name }
func (p *mockPlugin) Description() string                    { return p.description }
func (p *mockPlugin) GetFieldMetadata() map[string]FieldMeta { return p.fields }
func (p *mockPlugin) GetPresets() map[string]Preset          { return p.presets }
func (p *mockPlugin) GetHooks() map[string]string            { return p.hooks }

func (p *mockPlugin) DetectConfigPath() (string, error) {
	if p.failDetect {
		return "", fmt.Errorf("mock detect failure")
	}
	return p.configPath, nil
}

func (p *mockPlugin) ParseConfig(configPath string) (map[string]interface{}, error) {
	if p.failParse {
		return nil, fmt.Errorf("mock parse failure")
	}
	// Return a simple mock config
	return map[string]interface{}{
		"theme":     "dark",
		"font-size": 14,
		"debug":     false,
	}, nil
}

func (p *mockPlugin) WriteConfig(configPath string, config map[string]interface{}) error {
	if p.failWrite {
		return fmt.Errorf("mock write failure")
	}
	// Just pretend to write
	return nil
}

func (p *mockPlugin) ValidateValue(field string, value interface{}) error {
	if p.failValidate {
		return fmt.Errorf("mock validation failure")
	}

	fieldMeta, exists := p.fields[field]
	if !exists {
		return nil // Allow unknown fields
	}

	// Simple validation
	if fieldMeta.Type == "choice" {
		strValue := fmt.Sprintf("%v", value)
		for _, validValue := range fieldMeta.Values {
			if validValue == strValue {
				return nil
			}
		}
		return fmt.Errorf("invalid value %s for field %s", strValue, field)
	}

	return nil
}

// TestNewRegistry tests registry creation
func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.plugins == nil {
		t.Error("Expected plugins map to be initialized")
	}

	if len(registry.plugins) != 0 {
		t.Errorf("Expected empty registry, got %d plugins", len(registry.plugins))
	}
}

// TestRegistry_Register tests plugin registration
func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	plugin := newMockPlugin()

	// Test successful registration
	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	if len(registry.plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(registry.plugins))
	}

	// Test duplicate registration
	err = registry.Register(plugin)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("Expected 'already registered' error, got: %s", err.Error())
	}
}

// TestRegistry_Get tests plugin retrieval
func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	plugin := newMockPlugin()

	// Test getting non-existent plugin
	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent plugin")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %s", err.Error())
	}

	// Register and test getting existing plugin
	if err := registry.Register(plugin); err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	retrieved, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if retrieved != plugin {
		t.Error("Retrieved plugin is not the same as registered")
	}

	if retrieved.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", retrieved.Name())
	}
}

// TestRegistry_List tests listing plugins
func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %d items", len(list))
	}

	// Register plugins
	plugin1 := newMockPlugin()
	plugin2 := newMockPlugin()
	plugin2.name = "test-plugin-2"

	if err := registry.Register(plugin1); err != nil {
		t.Fatalf("Failed to register plugin1: %v", err)
	}

	if err := registry.Register(plugin2); err != nil {
		t.Fatalf("Failed to register plugin2: %v", err)
	}

	// Test list
	list = registry.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(list))
	}

	// Check that both plugins are in the list
	found1, found2 := false, false
	for _, name := range list {
		if name == "test-plugin" {
			found1 = true
		}
		if name == "test-plugin-2" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Expected to find 'test-plugin' in list")
	}
	if !found2 {
		t.Error("Expected to find 'test-plugin-2' in list")
	}
}

// TestAutoGenerate tests auto-generation of config files
func TestAutoGenerate(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plugin-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Set HOME to tmpDir so config goes there
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	plugin := newMockPlugin()
	plugin.configPath = "/custom/config/path"

	// Test successful generation
	err = AutoGenerate(plugin)
	if err != nil {
		t.Fatalf("Failed to auto-generate config: %v", err)
	}

	// Check that config file was created
	configPath := filepath.Join(tmpDir, ".config", "configtoggle", "apps", "test-plugin.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Check for expected content
	if !strings.Contains(contentStr, "name: test-plugin") {
		t.Error("Expected 'name: test-plugin' in generated config")
	}

	if !strings.Contains(contentStr, "path: /custom/config/path") {
		t.Error("Expected custom config path in generated config")
	}

	if !strings.Contains(contentStr, "description: Test plugin for unit testing") {
		t.Error("Expected plugin description in generated config")
	}

	if !strings.Contains(contentStr, "theme:") {
		t.Error("Expected theme field in generated config")
	}

	if !strings.Contains(contentStr, "font-size:") {
		t.Error("Expected font-size field in generated config")
	}

	if !strings.Contains(contentStr, "default:") {
		t.Error("Expected default preset in generated config")
	}

	if !strings.Contains(contentStr, "light-mode:") {
		t.Error("Expected light-mode preset in generated config")
	}

	if !strings.Contains(contentStr, "post-toggle:") {
		t.Error("Expected hooks in generated config")
	}

	// Test duplicate generation
	err = AutoGenerate(plugin)
	if err == nil {
		t.Error("Expected error for duplicate config generation")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %s", err.Error())
	}
}

// TestAutoGenerate_DetectFailure tests auto-generation with detect failure
func TestAutoGenerate_DetectFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	plugin := newMockPlugin()
	plugin.failDetect = true

	err = AutoGenerate(plugin)
	if err == nil {
		t.Error("Expected error when detect fails")
	}

	if !strings.Contains(err.Error(), "failed to detect config path") {
		t.Errorf("Expected detect error, got: %s", err.Error())
	}
}

// TestFieldMeta tests FieldMeta structure
func TestFieldMeta(t *testing.T) {
	field := FieldMeta{
		Type:        "choice",
		Values:      []string{"value1", "value2"},
		Default:     "value1",
		Description: "Test field",
		Path:        "nested.field",
	}

	if field.Type != "choice" {
		t.Errorf("Expected type 'choice', got '%s'", field.Type)
	}

	if len(field.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(field.Values))
	}

	if field.Default != "value1" {
		t.Errorf("Expected default 'value1', got '%v'", field.Default)
	}

	if field.Description != "Test field" {
		t.Errorf("Expected description 'Test field', got '%s'", field.Description)
	}

	if field.Path != "nested.field" {
		t.Errorf("Expected path 'nested.field', got '%s'", field.Path)
	}
}

// TestPreset tests Preset structure
func TestPreset(t *testing.T) {
	preset := Preset{
		Name:        "test-preset",
		Description: "Test preset",
		Values: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	if preset.Name != "test-preset" {
		t.Errorf("Expected name 'test-preset', got '%s'", preset.Name)
	}

	if preset.Description != "Test preset" {
		t.Errorf("Expected description 'Test preset', got '%s'", preset.Description)
	}

	if len(preset.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(preset.Values))
	}

	if preset.Values["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got '%v'", preset.Values["key1"])
	}

	if preset.Values["key2"] != 42 {
		t.Errorf("Expected key2 42, got '%v'", preset.Values["key2"])
	}
}

// TestPlugin_InterfaceCompliance tests that mock plugin implements interface
func TestPlugin_InterfaceCompliance(t *testing.T) {
	plugin := newMockPlugin()

	// Test all interface methods
	if plugin.Name() != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", plugin.Name())
	}

	if plugin.Description() == "" {
		t.Error("Expected non-empty description")
	}

	path, err := plugin.DetectConfigPath()
	if err != nil {
		t.Errorf("Unexpected error from DetectConfigPath: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty config path")
	}

	config, err := plugin.ParseConfig("dummy")
	if err != nil {
		t.Errorf("Unexpected error from ParseConfig: %v", err)
	}
	if len(config) == 0 {
		t.Error("Expected non-empty config")
	}

	err = plugin.WriteConfig("dummy", config)
	if err != nil {
		t.Errorf("Unexpected error from WriteConfig: %v", err)
	}

	fields := plugin.GetFieldMetadata()
	if len(fields) == 0 {
		t.Error("Expected some field metadata")
	}

	presets := plugin.GetPresets()
	if len(presets) == 0 {
		t.Error("Expected some presets")
	}

	hooks := plugin.GetHooks()
	if len(hooks) == 0 {
		t.Error("Expected some hooks")
	}

	err = plugin.ValidateValue("theme", "dark")
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}

	err = plugin.ValidateValue("theme", "invalid")
	if err == nil {
		t.Error("Expected validation error for invalid value")
	}
}

// TestPlugin_ErrorHandling tests plugin error scenarios
func TestPlugin_ErrorHandling(t *testing.T) {
	plugin := newMockPlugin()

	// Test DetectConfigPath failure
	plugin.failDetect = true
	_, err := plugin.DetectConfigPath()
	if err == nil {
		t.Error("Expected error from DetectConfigPath")
	}

	// Test ParseConfig failure
	plugin.failDetect = false
	plugin.failParse = true
	_, err = plugin.ParseConfig("dummy")
	if err == nil {
		t.Error("Expected error from ParseConfig")
	}

	// Test WriteConfig failure
	plugin.failParse = false
	plugin.failWrite = true
	err = plugin.WriteConfig("dummy", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error from WriteConfig")
	}

	// Test ValidateValue failure
	plugin.failWrite = false
	plugin.failValidate = true
	err = plugin.ValidateValue("theme", "dark")
	if err == nil {
		t.Error("Expected validation error")
	}
}

// TestPlugin_FieldValidation tests field validation logic
func TestPlugin_FieldValidation(t *testing.T) {
	plugin := newMockPlugin()

	// Test valid choice values
	validThemes := []string{"dark", "light", "auto"}
	for _, theme := range validThemes {
		if err := plugin.ValidateValue("theme", theme); err != nil {
			t.Errorf("Unexpected error for valid theme '%s': %v", theme, err)
		}
	}

	// Test invalid choice value
	if err := plugin.ValidateValue("theme", "invalid"); err == nil {
		t.Error("Expected error for invalid theme value")
	}

	// Test unknown field (should be allowed)
	if err := plugin.ValidateValue("unknown", "value"); err != nil {
		t.Errorf("Unexpected error for unknown field: %v", err)
	}

	// Test valid number field
	if err := plugin.ValidateValue("font-size", "14"); err != nil {
		t.Errorf("Unexpected error for valid font-size: %v", err)
	}

	// Test boolean field (should pass since we don't validate type strictly)
	if err := plugin.ValidateValue("debug", true); err != nil {
		t.Errorf("Unexpected error for boolean field: %v", err)
	}
}

// BenchmarkRegistry_Register benchmarks plugin registration
func BenchmarkRegistry_Register(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin := newMockPlugin()
		plugin.name = fmt.Sprintf("plugin-%d", i)

		if err := registry.Register(plugin); err != nil {
			b.Fatalf("Failed to register plugin: %v", err)
		}
	}
}

// BenchmarkRegistry_Get benchmarks plugin retrieval
func BenchmarkRegistry_Get(b *testing.B) {
	registry := NewRegistry()

	// Register some plugins
	for i := 0; i < 100; i++ {
		plugin := newMockPlugin()
		plugin.name = fmt.Sprintf("plugin-%d", i)
		registry.Register(plugin)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("plugin-%d", i%100)
		if _, err := registry.Get(name); err != nil {
			b.Fatalf("Failed to get plugin: %v", err)
		}
	}
}

// BenchmarkAutoGenerate benchmarks config auto-generation
func BenchmarkAutoGenerate(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "plugin-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalHome := os.Getenv("HOME")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDir := filepath.Join(tmpDir, fmt.Sprintf("test-%d", i))
		os.Setenv("HOME", testDir)

		plugin := newMockPlugin()
		plugin.name = fmt.Sprintf("bench-plugin-%d", i)

		if err := AutoGenerate(plugin); err != nil {
			b.Fatalf("Failed to auto-generate: %v", err)
		}
	}

	os.Setenv("HOME", originalHome)
}
