package providers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/knadh/koanf/v2"
)

func TestGhosttyProvider(t *testing.T) {
	// Create a temporary Ghostty config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	configContent := `# Ghostty configuration
theme = dark
font-family = JetBrains Mono
font-size = 12
window-padding-x = 10
# Comment about keybinds
keybind = ctrl+c=copy
keybind = ctrl+v=paste
background-opacity = 0.9
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("GhosttyProvider reads correctly", func(t *testing.T) {
		provider := NewGhosttyProvider(configPath)

		data, err := provider.Read()
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected non-empty data")
		}

		// Check that it contains expected keys
		content := string(data)
		if !contains(content, "theme") {
			t.Error("Expected theme key in output")
		}
		if !contains(content, "font-family") {
			t.Error("Expected font-family key in output")
		}
	})

	t.Run("GhosttyParser parses correctly", func(t *testing.T) {
		parser := NewGhosttyParser()

		testData := []byte(`theme=dark
font-family=JetBrains Mono
font-size=12
keybind=ctrl+c=copy,ctrl+v=paste`)

		result, err := parser.Unmarshal(testData)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Check parsed values
		if result["theme"] != "dark" {
			t.Errorf("Expected theme=dark, got %v", result["theme"])
		}

		if result["font-family"] != "JetBrains Mono" {
			t.Errorf("Expected font-family=JetBrains Mono, got %v", result["font-family"])
		}

		if result["font-size"] != "12" {
			t.Errorf("Expected font-size=12, got %v", result["font-size"])
		}

		// Check array handling
		keybind, ok := result["keybind"].([]string)
		if !ok {
			t.Errorf("Expected keybind to be []string, got %T", result["keybind"])
		} else {
			expected := []string{"ctrl+c=copy", "ctrl+v=paste"}
			if !reflect.DeepEqual(keybind, expected) {
				t.Errorf("Expected keybind=%v, got %v", expected, keybind)
			}
		}
	})

	t.Run("GhosttyParser marshals correctly", func(t *testing.T) {
		parser := NewGhosttyParser()

		testMap := map[string]interface{}{
			"theme":       "dark",
			"font-family": "JetBrains Mono",
			"font-size":   "12",
			"keybind":     []string{"ctrl+c=copy", "ctrl+v=paste"},
		}

		data, err := parser.Marshal(testMap)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		content := string(data)

		// Check that all keys are present
		if !contains(content, "theme = dark") {
			t.Error("Expected 'theme = dark' in output")
		}
		if !contains(content, "font-family = JetBrains Mono") {
			t.Error("Expected 'font-family = JetBrains Mono' in output")
		}
		if !contains(content, "font-size = 12") {
			t.Error("Expected 'font-size = 12' in output")
		}

		// Check that array values are on separate lines
		if !contains(content, "keybind = ctrl+c=copy") {
			t.Error("Expected 'keybind = ctrl+c=copy' in output")
		}
		if !contains(content, "keybind = ctrl+v=paste") {
			t.Error("Expected 'keybind = ctrl+v=paste' in output")
		}
	})

	t.Run("GhosttyProviderWithParser loads into koanf", func(t *testing.T) {
		provider := NewGhosttyProviderWithParser(configPath)
		k := koanf.New(".")

		err := provider.LoadIntoKoanf(k)
		if err != nil {
			t.Fatalf("Failed to load into koanf: %v", err)
		}

		// Check loaded values
		if k.String("theme") != "dark" {
			t.Errorf("Expected theme=dark, got %s", k.String("theme"))
		}

		if k.String("font-family") != "JetBrains Mono" {
			t.Errorf("Expected font-family=JetBrains Mono, got %s", k.String("font-family"))
		}

		if k.String("font-size") != "12" {
			t.Errorf("Expected font-size=12, got %s", k.String("font-size"))
		}

		// Check array values
		keybinds := k.Strings("keybind")
		if len(keybinds) != 2 {
			t.Errorf("Expected 2 keybinds, got %d", len(keybinds))
		}

		expected := []string{"ctrl+c=copy", "ctrl+v=paste"}
		if !reflect.DeepEqual(keybinds, expected) {
			t.Errorf("Expected keybinds=%v, got %v", expected, keybinds)
		}
	})

	t.Run("Error handling - missing file", func(t *testing.T) {
		provider := NewGhosttyProvider("/nonexistent/file")

		_, err := provider.Read()
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})

	t.Run("Error handling - invalid config", func(t *testing.T) {
		// Create invalid config
		invalidPath := filepath.Join(tempDir, "invalid")
		err := os.WriteFile(invalidPath, []byte("invalid content without equals"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		provider := NewGhosttyProviderWithParser(invalidPath)
		k := koanf.New(".")

		// Should not fail even with invalid lines (they're skipped)
		err = provider.LoadIntoKoanf(k)
		if err != nil {
			t.Errorf("Unexpected error with invalid config: %v", err)
		}

		// Should have no keys loaded
		if len(k.All()) != 0 {
			t.Errorf("Expected no keys loaded from invalid config, got %d", len(k.All()))
		}
	})
}

func TestGhosttyProviderByteOperations(t *testing.T) {
	t.Run("ReadBytes works correctly", func(t *testing.T) {
		provider := NewGhosttyProvider("")

		testData := []byte(`theme = dark
font-family = JetBrains Mono
# comment
keybind = ctrl+c=copy
keybind = ctrl+v=paste`)

		result, err := provider.ReadBytes(testData)
		if err != nil {
			t.Fatalf("Failed to read bytes: %v", err)
		}

		content := string(result)
		if !contains(content, "theme") {
			t.Error("Expected theme in result")
		}
		if !contains(content, "keybind") {
			t.Error("Expected keybind in result")
		}
	})
}

// BenchmarkGhosttyProvider benchmarks the provider performance
func BenchmarkGhosttyProvider(b *testing.B) {
	// Create test config
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "config")

	configContent := `theme = dark
font-family = JetBrains Mono
font-size = 12
window-padding-x = 10
keybind = ctrl+c=copy
keybind = ctrl+v=paste
keybind = ctrl+shift+c=copy
keybind = ctrl+shift+v=paste
background-opacity = 0.9
window-decoration = false
cursor-style = block
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	b.Run("Provider Read", func(b *testing.B) {
		provider := NewGhosttyProvider(configPath)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := provider.Read()
			if err != nil {
				b.Fatalf("Failed to read: %v", err)
			}
		}
	})

	b.Run("Parser Unmarshal", func(b *testing.B) {
		parser := NewGhosttyParser()
		testData := []byte(configContent)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := parser.Unmarshal(testData)
			if err != nil {
				b.Fatalf("Failed to unmarshal: %v", err)
			}
		}
	})

	b.Run("Full LoadIntoKoanf", func(b *testing.B) {
		provider := NewGhosttyProviderWithParser(configPath)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := koanf.New(".")
			err := provider.LoadIntoKoanf(k)
			if err != nil {
				b.Fatalf("Failed to load: %v", err)
			}
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())
}
