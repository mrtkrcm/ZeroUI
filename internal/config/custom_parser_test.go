package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knadh/koanf/v2"
)

// TestParseGhosttyConfig tests parsing Ghostty configuration files
func TestParseGhosttyConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ghostty-parser-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test basic parsing
	t.Run("Basic parsing", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "basic.conf")
		configContent := `# Basic Ghostty config
theme = GruvboxDark
font-family = JetBrains Mono
font-size = 14
background-opacity = 0.9
debug = true
window-width = 120
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write basic config: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse basic config: %v", err)
		}

		// Test string values
		if config.String("theme") != "GruvboxDark" {
			t.Errorf("Expected theme 'GruvboxDark', got '%s'", config.String("theme"))
		}

		if config.String("font-family") != "JetBrains Mono" {
			t.Errorf("Expected font-family 'JetBrains Mono', got '%s'", config.String("font-family"))
		}

		// Test numeric values
		if config.String("font-size") != "14" {
			t.Errorf("Expected font-size '14', got '%s'", config.String("font-size"))
		}

		if config.String("background-opacity") != "0.9" {
			t.Errorf("Expected background-opacity '0.9', got '%s'", config.String("background-opacity"))
		}

		// Test boolean values
		if config.String("debug") != "true" {
			t.Errorf("Expected debug 'true', got '%s'", config.String("debug"))
		}

		// Test integer values
		if config.String("window-width") != "120" {
			t.Errorf("Expected window-width '120', got '%s'", config.String("window-width"))
		}
	})

	// Test multiple values with same key
	t.Run("Multiple values", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "multiple.conf")
		configContent := `# Multiple values for same key
font-feature = liga
font-feature = calt
font-feature = kern
keybind = cmd+c=copy
keybind = cmd+v=paste
keybind = cmd+n=new_window
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write multiple config: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse multiple config: %v", err)
		}

		// Test that multiple values are stored as arrays
		fontFeatures := config.Get("font-feature")
		if fontFeatures == nil {
			t.Error("Expected font-feature to exist")
		}

		// Should be a slice
		switch v := fontFeatures.(type) {
		case []string:
			if len(v) != 3 {
				t.Errorf("Expected 3 font-feature values, got %d", len(v))
			}
			if v[0] != "liga" || v[1] != "calt" || v[2] != "kern" {
				t.Errorf("Expected font-features [liga, calt, kern], got %v", v)
			}
		default:
			t.Errorf("Expected font-feature to be []string, got %T", fontFeatures)
		}

		// Test keybindings
		keybinds := config.Get("keybind")
		switch v := keybinds.(type) {
		case []string:
			if len(v) != 3 {
				t.Errorf("Expected 3 keybind values, got %d", len(v))
			}
		default:
			t.Errorf("Expected keybind to be []string, got %T", keybinds)
		}
	})

	// Test comments and empty lines
	t.Run("Comments and empty lines", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "comments.conf")
		configContent := `# This is a comment
# Another comment

theme = GruvboxDark

# Font configuration
font-family = SF Mono  # Inline comment (not supported, but should not break)
font-size = 14


# More comments
debug = false

`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write comments config: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse comments config: %v", err)
		}

		if config.String("theme") != "GruvboxDark" {
			t.Errorf("Expected theme 'GruvboxDark', got '%s'", config.String("theme"))
		}

		if config.String("font-family") != "SF Mono  # Inline comment (not supported, but should not break)" {
			t.Errorf("Expected font-family to include inline comment, got '%s'", config.String("font-family"))
		}

		if config.String("debug") != "false" {
			t.Errorf("Expected debug 'false', got '%s'", config.String("debug"))
		}
	})

	// Test edge cases
	t.Run("Edge cases", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "edge.conf")
		configContent := `# Edge cases
no-value =
empty-value = 
quoted-value = "Hello World"
spaces-around = value with spaces 
equals-in-value = url=https://example.com
malformed line without equals
= value-without-key
special-chars = áëíôü
unicode = 🎨🚀
`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write edge cases config: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse edge cases config: %v", err)
		}

		// Test empty values
		if config.String("no-value") != "" {
			t.Errorf("Expected no-value to be empty, got '%s'", config.String("no-value"))
		}

		if config.String("empty-value") != "" {
			t.Errorf("Expected empty-value to be empty, got '%s'", config.String("empty-value"))
		}

		// Test quoted values (quotes should be preserved)
		if config.String("quoted-value") != `"Hello World"` {
			t.Errorf("Expected quoted-value to preserve quotes, got '%s'", config.String("quoted-value"))
		}

		// Test values with spaces
		if config.String("spaces-around") != "value with spaces" {
			t.Errorf("Expected spaces-around 'value with spaces', got '%s'", config.String("spaces-around"))
		}

		// Test values with equals signs
		if config.String("equals-in-value") != "url=https://example.com" {
			t.Errorf("Expected equals-in-value to preserve equals, got '%s'", config.String("equals-in-value"))
		}

		// Test special characters
		if config.String("special-chars") != "áëíôü" {
			t.Errorf("Expected special-chars 'áëíôü', got '%s'", config.String("special-chars"))
		}

		// Test unicode
		if config.String("unicode") != "🎨🚀" {
			t.Errorf("Expected unicode '🎨🚀', got '%s'", config.String("unicode"))
		}

		// Malformed lines should be ignored
		if config.Exists("malformed line without equals") {
			t.Error("Expected malformed line to be ignored")
		}

		// Empty key should be ignored or handled gracefully
		if config.String("") != "" {
			t.Logf("Empty key value: '%s'", config.String(""))
		}
	})

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		_, err := ParseGhosttyConfig("/nonexistent/file.conf")
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

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse empty config: %v", err)
		}

		if len(config.All()) != 0 {
			t.Errorf("Expected empty config to have no keys, got %d", len(config.All()))
		}
	})

	// Test file with only comments
	t.Run("Comments only", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "comments-only.conf")
		configContent := `# Only comments
# Another comment
# Yet another comment

# Final comment`
		if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write comments-only config: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse comments-only config: %v", err)
		}

		if len(config.All()) != 0 {
			t.Errorf("Expected comments-only config to have no keys, got %d", len(config.All()))
		}
	})
}

// TestWriteGhosttyConfig tests writing Ghostty configuration files
func TestWriteGhosttyConfig(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ghostty-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test writing new config file
	t.Run("New config file", func(t *testing.T) {
		outputPath := filepath.Join(tmpDir, "new.conf")

		// Create config data
		k := koanf.New(".")
		k.Set("theme", "GruvboxDark")
		k.Set("font-family", "JetBrains Mono")
		k.Set("font-size", "14")
		k.Set("debug", "true")

		err := WriteGhosttyConfig(outputPath, k, "/nonexistent/original")
		if err != nil {
			t.Fatalf("Failed to write new config: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Expected config file to be created")
		}

		// Verify content
		content, err := ioutil.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read written config: %v", err)
		}

		contentStr := string(content)

		// Check header
		if !strings.Contains(contentStr, "# Ghostty Configuration") {
			t.Error("Expected header in new config file")
		}

		// Check values
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected 'theme = GruvboxDark' in output")
		}

		if !strings.Contains(contentStr, "font-family = JetBrains Mono") {
			t.Error("Expected 'font-family = JetBrains Mono' in output")
		}

		if !strings.Contains(contentStr, "font-size = 14") {
			t.Error("Expected 'font-size = 14' in output")
		}
	})

	// Test updating existing config file
	t.Run("Update existing config", func(t *testing.T) {
		// Create original config
		originalPath := filepath.Join(tmpDir, "original.conf")
		originalContent := `# My custom config
theme = GruvboxLight
font-family = SF Mono
font-size = 16

# Window settings
window-width = 120
window-height = 40

# Debug mode
debug = false
`
		if err := ioutil.WriteFile(originalPath, []byte(originalContent), 0644); err != nil {
			t.Fatalf("Failed to write original config: %v", err)
		}

		outputPath := filepath.Join(tmpDir, "updated.conf")

		// Create updated config data
		k := koanf.New(".")
		k.Set("theme", "GruvboxDark")          // Update existing
		k.Set("font-family", "JetBrains Mono") // Update existing
		k.Set("font-size", "14")               // Update existing
		k.Set("debug", "true")                 // Update existing
		k.Set("new-setting", "new-value")      // Add new

		err := WriteGhosttyConfig(outputPath, k, originalPath)
		if err != nil {
			t.Fatalf("Failed to write updated config: %v", err)
		}

		// Verify content
		content, err := ioutil.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		contentStr := string(content)

		// Check that original structure and comments are preserved
		if !strings.Contains(contentStr, "# My custom config") {
			t.Error("Expected original comment to be preserved")
		}

		if !strings.Contains(contentStr, "# Window settings") {
			t.Error("Expected original section comment to be preserved")
		}

		// Check updated values
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected updated theme value")
		}

		if !strings.Contains(contentStr, "font-family = JetBrains Mono") {
			t.Error("Expected updated font-family value")
		}

		if !strings.Contains(contentStr, "debug = true") {
			t.Error("Expected updated debug value")
		}

		// Check that unchanged values are preserved
		if !strings.Contains(contentStr, "window-width = 120") {
			t.Error("Expected unchanged window-width to be preserved")
		}

		if !strings.Contains(contentStr, "window-height = 40") {
			t.Error("Expected unchanged window-height to be preserved")
		}

		// Check new value is added
		if !strings.Contains(contentStr, "new-setting = new-value") {
			t.Error("Expected new setting to be added")
		}
	})

	// Test array values
	t.Run("Array values", func(t *testing.T) {
		outputPath := filepath.Join(tmpDir, "arrays.conf")

		k := koanf.New(".")
		k.Set("font-feature", []string{"liga", "calt", "kern"})
		k.Set("keybind", []string{"cmd+c=copy", "cmd+v=paste", "cmd+n=new_window"})
		k.Set("theme", "GruvboxDark")

		err := WriteGhosttyConfig(outputPath, k, "/nonexistent")
		if err != nil {
			t.Fatalf("Failed to write array config: %v", err)
		}

		content, err := ioutil.ReadFile(outputPath)
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
		if !strings.Contains(contentStr, "keybind = cmd+n=new_window") {
			t.Error("Expected third keybind")
		}

		// Regular value should still work
		if !strings.Contains(contentStr, "theme = GruvboxDark") {
			t.Error("Expected regular theme value")
		}
	})

	// Test error cases
	t.Run("Error cases", func(t *testing.T) {
		k := koanf.New(".")
		k.Set("test", "value")

		// Try to write to invalid path
		err := WriteGhosttyConfig("/invalid/path/file.conf", k, "/nonexistent")
		if err == nil {
			t.Error("Expected error for invalid output path")
		}
	})
}

// TestGhosttyConfigRoundTrip tests parsing and writing back preserves data
func TestGhosttyConfigRoundTrip(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ghostty-roundtrip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create original config
	originalPath := filepath.Join(tmpDir, "original.conf")
	originalContent := `# Ghostty Configuration
theme = GruvboxDark
font-family = JetBrains Mono
font-size = 14
font-feature = liga
font-feature = calt
background-opacity = 0.9
debug = true
keybind = cmd+c=copy
keybind = cmd+v=paste
`

	if err := ioutil.WriteFile(originalPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write original config: %v", err)
	}

	// Parse config
	parsedConfig, err := ParseGhosttyConfig(originalPath)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// Write it back
	outputPath := filepath.Join(tmpDir, "output.conf")
	err = WriteGhosttyConfig(outputPath, parsedConfig, originalPath)
	if err != nil {
		t.Fatalf("Failed to write config back: %v", err)
	}

	// Parse the output
	reparsedConfig, err := ParseGhosttyConfig(outputPath)
	if err != nil {
		t.Fatalf("Failed to reparse config: %v", err)
	}

	// Compare key values
	originalData := parsedConfig.All()
	reparsedData := reparsedConfig.All()

	for key, originalValue := range originalData {
		reparsedValue, exists := reparsedData[key]
		if !exists {
			t.Errorf("Key '%s' missing in reparsed config", key)
			continue
		}

		// Compare values (handling both single values and arrays)
		switch ov := originalValue.(type) {
		case string:
			if rv, ok := reparsedValue.(string); ok {
				if ov != rv {
					t.Errorf("Value mismatch for key '%s': original '%s', reparsed '%s'", key, ov, rv)
				}
			} else if rva, ok := reparsedValue.([]string); ok {
				if len(rva) != 1 || rva[0] != ov {
					t.Errorf("Value type mismatch for key '%s': original string '%s', reparsed array %v", key, ov, rva)
				}
			} else {
				t.Errorf("Value type mismatch for key '%s': original %T, reparsed %T", key, originalValue, reparsedValue)
			}
		case []string:
			if rv, ok := reparsedValue.([]string); ok {
				if len(ov) != len(rv) {
					t.Errorf("Array length mismatch for key '%s': original %d, reparsed %d", key, len(ov), len(rv))
				} else {
					for i, originalItem := range ov {
						if i >= len(rv) || originalItem != rv[i] {
							t.Errorf("Array item mismatch for key '%s'[%d]: original '%s', reparsed '%s'", key, i, originalItem, rv[i])
						}
					}
				}
			} else {
				t.Errorf("Value type mismatch for key '%s': original []string, reparsed %T", key, reparsedValue)
			}
		}
	}

	// Check for extra keys in reparsed
	for key := range reparsedData {
		if _, exists := originalData[key]; !exists {
			t.Errorf("Extra key '%s' in reparsed config", key)
		}
	}
}

// BenchmarkParseGhosttyConfig benchmarks config parsing
func BenchmarkParseGhosttyConfig(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "ghostty-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large config file
	configPath := filepath.Join(tmpDir, "bench.conf")
	var content strings.Builder
	content.WriteString("# Large Ghostty config for benchmarking\n")

	// Add many entries
	for i := 0; i < 100; i++ {
		content.WriteString(fmt.Sprintf("setting-%d = value-%d\n", i, i))
		if i%10 == 0 {
			content.WriteString(fmt.Sprintf("# Comment %d\n", i))
		}
	}

	// Add some arrays
	for i := 0; i < 20; i++ {
		content.WriteString("font-feature = liga\n")
		content.WriteString("font-feature = calt\n")
		content.WriteString(fmt.Sprintf("keybind = cmd+%d=action-%d\n", i, i))
	}

	if err := ioutil.WriteFile(configPath, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to write bench config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseGhosttyConfig(configPath)
		if err != nil {
			b.Fatalf("Failed to parse config: %v", err)
		}
	}
}

// BenchmarkWriteGhosttyConfig benchmarks config writing
func BenchmarkWriteGhosttyConfig(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "ghostty-write-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config data
	k := koanf.New(".")
	for i := 0; i < 50; i++ {
		k.Set(fmt.Sprintf("setting-%d", i), fmt.Sprintf("value-%d", i))
	}
	k.Set("font-feature", []string{"liga", "calt", "kern", "ss01", "ss02"})
	k.Set("keybind", []string{"cmd+c=copy", "cmd+v=paste", "cmd+x=cut", "cmd+z=undo"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, fmt.Sprintf("bench-%d.conf", i))
		err := WriteGhosttyConfig(outputPath, k, "/nonexistent")
		if err != nil {
			b.Fatalf("Failed to write config: %v", err)
		}
	}
}
