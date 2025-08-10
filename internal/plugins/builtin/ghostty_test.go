package builtin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/plugins"
)

// setupGhosttyTest creates a test environment for Ghostty plugin
func setupGhosttyTest(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "ghostty-plugin-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// TestNewGhosttyPlugin tests plugin creation
func TestNewGhosttyPlugin(t *testing.T) {
	plugin := NewGhosttyPlugin()
	if plugin == nil {
		t.Fatal("Expected non-nil plugin")
	}

	// Test interface compliance
	var _ plugins.Plugin = plugin

	if plugin.Name() != "ghostty" {
		t.Errorf("Expected name 'ghostty', got '%s'", plugin.Name())
	}

	if plugin.Description() == "" {
		t.Error("Expected non-empty description")
	}

	if !strings.Contains(plugin.Description(), "Ghostty") {
		t.Error("Expected description to contain 'Ghostty'")
	}
}

// TestGhosttyPlugin_DetectConfigPath tests config path detection
func TestGhosttyPlugin_DetectConfigPath(t *testing.T) {
	tmpDir, cleanup := setupGhosttyTest(t)
	defer cleanup()

	plugin := NewGhosttyPlugin()

	// Set HOME to our test directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test with no existing config (should return default path)
	path, err := plugin.DetectConfigPath()
	if err != nil {
		t.Fatalf("Unexpected error detecting config path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".config", "ghostty", "config")
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}

	// Test with existing Linux config
	linuxConfigDir := filepath.Join(tmpDir, ".config", "ghostty")
	if err := os.MkdirAll(linuxConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create Linux config dir: %v", err)
	}

	linuxConfigPath := filepath.Join(linuxConfigDir, "config")
	if err := ioutil.WriteFile(linuxConfigPath, []byte("theme = GruvboxDark\n"), 0644); err != nil {
		t.Fatalf("Failed to create Linux config file: %v", err)
	}

	path, err = plugin.DetectConfigPath()
	if err != nil {
		t.Fatalf("Unexpected error detecting existing config: %v", err)
	}

	if path != linuxConfigPath {
		t.Errorf("Expected to detect Linux config path '%s', got '%s'", linuxConfigPath, path)
	}

	// Test with macOS config
	os.Remove(linuxConfigPath) // Remove Linux config

	macOSConfigDir := filepath.Join(tmpDir, "Library", "Application Support", "com.mitchellh.ghostty")
	if err := os.MkdirAll(macOSConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create macOS config dir: %v", err)
	}

	macOSConfigPath := filepath.Join(macOSConfigDir, "config")
	if err := ioutil.WriteFile(macOSConfigPath, []byte("theme = GruvboxLight\n"), 0644); err != nil {
		t.Fatalf("Failed to create macOS config file: %v", err)
	}

	path, err = plugin.DetectConfigPath()
	if err != nil {
		t.Fatalf("Unexpected error detecting macOS config: %v", err)
	}

	if path != macOSConfigPath {
		t.Errorf("Expected to detect macOS config path '%s', got '%s'", macOSConfigPath, path)
	}
}

// TestGhosttyPlugin_ParseConfig tests config parsing
func TestGhosttyPlugin_ParseConfig(t *testing.T) {
	tmpDir, cleanup := setupGhosttyTest(t)
	defer cleanup()

	plugin := NewGhosttyPlugin()

	// Test basic parsing
	t.Run("Basic parsing", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "basic.conf")
		configContent := `# Ghostty config
theme = GruvboxDark
font-family = JetBrains Mono
font-size = 14
background-opacity = 0.9
debug = true
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config, err := plugin.ParseConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		expectedValues := map[string]string{
			"theme":              "GruvboxDark",
			"font-family":        "JetBrains Mono",
			"font-size":          "14",
			"background-opacity": "0.9",
			"debug":              "true",
		}

		for key, expected := range expectedValues {
			if actual, exists := config[key]; !exists {
				t.Errorf("Expected key '%s' to exist", key)
			} else if actual != expected {
				t.Errorf("Expected '%s' = '%s', got '%v'", key, expected, actual)
			}
		}
	})

	// Test array values (multiple same keys)
	t.Run("Array values", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "arrays.conf")
		configContent := `font-feature = liga
font-feature = calt
font-feature = kern
keybind = cmd+c=copy
keybind = cmd+v=paste
theme = GruvboxDark
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write array config: %v", err)
		}

		config, err := plugin.ParseConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse array config: %v", err)
		}

		// Check font-feature array
		fontFeatures, exists := config["font-feature"]
		if !exists {
			t.Error("Expected font-feature to exist")
		} else {
			features, ok := fontFeatures.([]string)
			if !ok {
				t.Errorf("Expected font-feature to be []string, got %T", fontFeatures)
			} else {
				expected := []string{"liga", "calt", "kern"}
				if len(features) != len(expected) {
					t.Errorf("Expected %d font features, got %d", len(expected), len(features))
				} else {
					for i, exp := range expected {
						if features[i] != exp {
							t.Errorf("Expected font-feature[%d] = '%s', got '%s'", i, exp, features[i])
						}
					}
				}
			}
		}

		// Check keybind array
		keybinds, exists := config["keybind"]
		if !exists {
			t.Error("Expected keybind to exist")
		} else {
			binds, ok := keybinds.([]string)
			if !ok {
				t.Errorf("Expected keybind to be []string, got %T", keybinds)
			} else if len(binds) != 2 {
				t.Errorf("Expected 2 keybinds, got %d", len(binds))
			}
		}

		// Check single value still works
		if theme := config["theme"]; theme != "GruvboxDark" {
			t.Errorf("Expected theme 'GruvboxDark', got '%v'", theme)
		}
	})

	// Test comments and empty lines
	t.Run("Comments and empty lines", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "comments.conf")
		configContent := `# This is a header comment

# Theme configuration
theme = GruvboxDark

# Font settings
font-family = SF Mono
font-size = 14

# Empty line above, comment below
# window-width = 120
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write comments config: %v", err)
		}

		config, err := plugin.ParseConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse comments config: %v", err)
		}

		// Should only have non-comment lines
		expectedKeys := []string{"theme", "font-family", "font-size"}
		if len(config) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(config))
		}

		for _, key := range expectedKeys {
			if _, exists := config[key]; !exists {
				t.Errorf("Expected key '%s' to exist", key)
			}
		}

		// Commented line should not exist
		if _, exists := config["window-width"]; exists {
			t.Error("Expected commented line to be ignored")
		}
	})

	// Test malformed lines
	t.Run("Malformed lines", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "malformed.conf")
		configContent := `theme = GruvboxDark
malformed line without equals
= value without key
font-size = 14
another malformed line
font-family = JetBrains Mono
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write malformed config: %v", err)
		}

		config, err := plugin.ParseConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse malformed config: %v", err)
		}

		// Should only have valid key=value pairs
		expectedKeys := []string{"theme", "font-size", "font-family"}
		if len(config) != len(expectedKeys) {
			t.Errorf("Expected %d keys, got %d keys: %v", len(expectedKeys), len(config), config)
		}
	})

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		_, err := plugin.ParseConfig("/nonexistent/file")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	// Test empty file
	t.Run("Empty file", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "empty.conf")
		if err := ioutil.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write empty config: %v", err)
		}

		config, err := plugin.ParseConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse empty config: %v", err)
		}

		if len(config) != 0 {
			t.Errorf("Expected empty config, got %d keys", len(config))
		}
	})
}

// TestGhosttyPlugin_WriteConfig tests config writing
func TestGhosttyPlugin_WriteConfig(t *testing.T) {
	tmpDir, cleanup := setupGhosttyTest(t)
	defer cleanup()

	plugin := NewGhosttyPlugin()

	// Test writing new config
	t.Run("New config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "new.conf")
		config := map[string]interface{}{
			"theme":              "GruvboxDark",
			"font-family":        "JetBrains Mono",
			"font-size":          14,
			"background-opacity": 0.9,
		}

		err := plugin.WriteConfig(configPath, config)
		if err != nil {
			t.Fatalf("Failed to write new config: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Expected config file to be created")
		}

		// Verify content
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read written config: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected 'theme = GruvboxDark' in written config")
		}
		if !strings.Contains(contentStr, "font-family = JetBrains Mono") {
			t.Error("Expected 'font-family = JetBrains Mono' in written config")
		}
		if !strings.Contains(contentStr, "Generated by configtoggle") {
			t.Error("Expected header comment in new config")
		}
	})

	// Test updating existing config
	t.Run("Update existing config", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "original.conf")
		originalContent := `# My custom Ghostty config
theme = GruvboxLight
font-family = SF Mono
font-size = 16

# Window settings
window-width = 120
window-height = 40
`
		if err := ioutil.WriteFile(originalPath, []byte(originalContent), 0644); err != nil {
			t.Fatalf("Failed to write original config: %v", err)
		}

		updatePath := filepath.Join(tmpDir, "updated.conf")
		updateConfig := map[string]interface{}{
			"theme":       "GruvboxDark",    // Update existing
			"font-family": "JetBrains Mono", // Update existing
			"font-size":   14,               // Update existing
			"new-setting": "new-value",      // Add new
		}

		err := plugin.WriteConfig(updatePath, updateConfig)
		if err != nil {
			t.Fatalf("Failed to write updated config: %v", err)
		}

		// Read result
		content, err := ioutil.ReadFile(updatePath)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		contentStr := string(content)

		// Check updated values
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected updated theme value")
		}
		if !strings.Contains(contentStr, "font-family = JetBrains Mono") {
			t.Error("Expected updated font-family value")
		}
		if !strings.Contains(contentStr, "font-size = 14") {
			t.Error("Expected updated font-size value")
		}

		// Check new value added
		if !strings.Contains(contentStr, "new-setting = new-value") {
			t.Error("Expected new setting to be added")
		}

		// Check preserved values
		if !strings.Contains(contentStr, "window-width = 120") {
			t.Error("Expected unchanged values to be preserved")
		}
	})

	// Test array values
	t.Run("Array values", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "arrays.conf")
		config := map[string]interface{}{
			"font-feature": []string{"liga", "calt", "kern"},
			"keybind":      []string{"cmd+c=copy", "cmd+v=paste"},
			"theme":        "GruvboxDark",
		}

		err := plugin.WriteConfig(configPath, config)
		if err != nil {
			t.Fatalf("Failed to write array config: %v", err)
		}

		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read array config: %v", err)
		}

		contentStr := string(content)

		// Check that array values are written as multiple lines
		ligaCount := strings.Count(contentStr, "font-feature = liga")
		caltCount := strings.Count(contentStr, "font-feature = calt")
		kernCount := strings.Count(contentStr, "font-feature = kern")

		if ligaCount != 1 || caltCount != 1 || kernCount != 1 {
			t.Error("Expected each font-feature value to appear once")
		}

		// Check keybinds
		if !strings.Contains(contentStr, "keybind = cmd+c=copy") {
			t.Error("Expected first keybind")
		}
		if !strings.Contains(contentStr, "keybind = cmd+v=paste") {
			t.Error("Expected second keybind")
		}

		// Check regular value
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected theme value")
		}
	})
}

// TestGhosttyPlugin_GetFieldMetadata tests field metadata
func TestGhosttyPlugin_GetFieldMetadata(t *testing.T) {
	plugin := NewGhosttyPlugin()
	fields := plugin.GetFieldMetadata()

	if len(fields) == 0 {
		t.Error("Expected some field metadata")
	}

	// Test specific fields
	expectedFields := map[string]string{
		"theme":                  "choice",
		"font-family":            "choice",
		"font-size":              "number",
		"background-opacity":     "number",
		"background-blur-radius": "number",
		"cursor-style":           "choice",
		"window-theme":           "choice",
	}

	for fieldName, expectedType := range expectedFields {
		field, exists := fields[fieldName]
		if !exists {
			t.Errorf("Expected field '%s' to exist", fieldName)
			continue
		}

		if field.Type != expectedType {
			t.Errorf("Expected field '%s' type '%s', got '%s'", fieldName, expectedType, field.Type)
		}

		if field.Description == "" {
			t.Errorf("Expected field '%s' to have description", fieldName)
		}

		if field.Default == nil {
			t.Errorf("Expected field '%s' to have default value", fieldName)
		}

		if expectedType == "choice" && len(field.Values) == 0 {
			t.Errorf("Expected choice field '%s' to have values", fieldName)
		}
	}

	// Test theme field specifically
	themeField := fields["theme"]
	if !contains(themeField.Values, "GruvboxLight") {
		t.Error("Expected theme field to contain 'GruvboxLight'")
	}
	if !contains(themeField.Values, "GruvboxDark") {
		t.Error("Expected theme field to contain 'GruvboxDark'")
	}

	// Test font-family field
	fontField := fields["font-family"]
	if !contains(fontField.Values, "JetBrains Mono") {
		t.Error("Expected font-family field to contain 'JetBrains Mono'")
	}
}

// TestGhosttyPlugin_GetPresets tests preset definitions
func TestGhosttyPlugin_GetPresets(t *testing.T) {
	plugin := NewGhosttyPlugin()
	presets := plugin.GetPresets()

	if len(presets) == 0 {
		t.Error("Expected some presets")
	}

	expectedPresets := []string{"dark-mode", "light-mode", "cyberpunk", "minimal"}

	for _, presetName := range expectedPresets {
		preset, exists := presets[presetName]
		if !exists {
			t.Errorf("Expected preset '%s' to exist", presetName)
			continue
		}

		if preset.Name != presetName {
			t.Errorf("Expected preset name '%s', got '%s'", presetName, preset.Name)
		}

		if preset.Description == "" {
			t.Errorf("Expected preset '%s' to have description", presetName)
		}

		if len(preset.Values) == 0 {
			t.Errorf("Expected preset '%s' to have values", presetName)
		}

		// Check that preset has theme setting
		if _, hasTheme := preset.Values["theme"]; !hasTheme {
			t.Errorf("Expected preset '%s' to have theme setting", presetName)
		}
	}

	// Test specific preset values
	darkMode := presets["dark-mode"]
	if darkMode.Values["theme"] != "GruvboxDark" {
		t.Errorf("Expected dark-mode theme 'GruvboxDark', got '%v'", darkMode.Values["theme"])
	}
	if darkMode.Values["window-theme"] != "dark" {
		t.Errorf("Expected dark-mode window-theme 'dark', got '%v'", darkMode.Values["window-theme"])
	}

	lightMode := presets["light-mode"]
	if lightMode.Values["theme"] != "GruvboxLight" {
		t.Errorf("Expected light-mode theme 'GruvboxLight', got '%v'", lightMode.Values["theme"])
	}
	if lightMode.Values["window-theme"] != "light" {
		t.Errorf("Expected light-mode window-theme 'light', got '%v'", lightMode.Values["window-theme"])
	}
}

// TestGhosttyPlugin_ValidateValue tests value validation
func TestGhosttyPlugin_ValidateValue(t *testing.T) {
	plugin := NewGhosttyPlugin()

	// Test valid theme values
	validThemes := []string{"GruvboxLight", "GruvboxDark", "catppuccin-mocha", "nord", "tokyonight"}
	for _, theme := range validThemes {
		if err := plugin.ValidateValue("theme", theme); err != nil {
			t.Errorf("Unexpected error for valid theme '%s': %v", theme, err)
		}
	}

	// Test invalid theme value
	if err := plugin.ValidateValue("theme", "invalid-theme"); err == nil {
		t.Error("Expected error for invalid theme value")
	}

	// Test valid font-family values
	validFonts := []string{"JetBrains Mono", "SF Mono", "Fira Code"}
	for _, font := range validFonts {
		if err := plugin.ValidateValue("font-family", font); err != nil {
			t.Errorf("Unexpected error for valid font '%s': %v", font, err)
		}
	}

	// Test invalid font-family value
	if err := plugin.ValidateValue("font-family", "Invalid Font"); err == nil {
		t.Error("Expected error for invalid font-family value")
	}

	// Test number field
	if err := plugin.ValidateValue("font-size", 14); err != nil {
		t.Errorf("Unexpected error for number field: %v", err)
	}

	if err := plugin.ValidateValue("font-size", "16"); err != nil {
		t.Errorf("Unexpected error for string number: %v", err)
	}

	// Test unknown field (should be allowed)
	if err := plugin.ValidateValue("unknown-field", "any-value"); err != nil {
		t.Errorf("Unexpected error for unknown field: %v", err)
	}

	// Test cursor-style validation
	validCursors := []string{"block", "bar", "underline"}
	for _, cursor := range validCursors {
		if err := plugin.ValidateValue("cursor-style", cursor); err != nil {
			t.Errorf("Unexpected error for valid cursor '%s': %v", cursor, err)
		}
	}

	if err := plugin.ValidateValue("cursor-style", "invalid"); err == nil {
		t.Error("Expected error for invalid cursor style")
	}
}

// TestGhosttyPlugin_GetHooks tests hook definitions
func TestGhosttyPlugin_GetHooks(t *testing.T) {
	plugin := NewGhosttyPlugin()
	hooks := plugin.GetHooks()

	if len(hooks) == 0 {
		t.Error("Expected some hooks")
	}

	expectedHooks := []string{"post-toggle", "post-preset"}
	for _, hookName := range expectedHooks {
		command, exists := hooks[hookName]
		if !exists {
			t.Errorf("Expected hook '%s' to exist", hookName)
			continue
		}

		if command == "" {
			t.Errorf("Expected hook '%s' to have command", hookName)
		}

		// Check that it's an osascript command (macOS specific)
		if !strings.Contains(command, "osascript") {
			t.Errorf("Expected hook '%s' to contain osascript command", hookName)
		}
	}
}

// BenchmarkGhosttyPlugin_ParseConfig benchmarks config parsing
func BenchmarkGhosttyPlugin_ParseConfig(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "ghostty-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large config file
	configPath := filepath.Join(tmpDir, "bench.conf")
	var content strings.Builder
	content.WriteString("# Large Ghostty config for benchmarking\n")

	// Add many settings
	for i := 0; i < 100; i++ {
		content.WriteString(fmt.Sprintf("setting-%d = value-%d\n", i, i))
		if i%10 == 0 {
			content.WriteString("# Comment line\n")
		}
	}

	// Add array values
	for i := 0; i < 20; i++ {
		content.WriteString("font-feature = liga\n")
		content.WriteString("font-feature = calt\n")
		content.WriteString(fmt.Sprintf("keybind = cmd+%d=action-%d\n", i, i))
	}

	if err := ioutil.WriteFile(configPath, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to write bench config: %v", err)
	}

	plugin := NewGhosttyPlugin()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := plugin.ParseConfig(configPath)
		if err != nil {
			b.Fatalf("Failed to parse config: %v", err)
		}
	}
}

// BenchmarkGhosttyPlugin_WriteConfig benchmarks config writing
func BenchmarkGhosttyPlugin_WriteConfig(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "ghostty-write-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	plugin := NewGhosttyPlugin()

	// Create config data
	config := map[string]interface{}{
		"theme":              "GruvboxDark",
		"font-family":        "JetBrains Mono",
		"font-size":          14,
		"background-opacity": 0.9,
		"font-feature":       []string{"liga", "calt", "kern"},
		"keybind":            []string{"cmd+c=copy", "cmd+v=paste"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configPath := filepath.Join(tmpDir, fmt.Sprintf("bench-%d.conf", i))
		err := plugin.WriteConfig(configPath, config)
		if err != nil {
			b.Fatalf("Failed to write config: %v", err)
		}
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
