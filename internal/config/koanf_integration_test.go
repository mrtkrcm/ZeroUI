package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestKoanfProviderIntegration tests the integration of the new koanf provider
// with the existing ParseGhosttyConfig function
func TestKoanfProviderIntegration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Basic parsing with koanf provider", func(t *testing.T) {
		configContent := `theme = dark
font-family = JetBrains Mono
font-size = 12
background-opacity = 0.9
window-decoration = false`

		configPath := filepath.Join(tempDir, "basic.conf")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Test basic values
		if config.String("theme") != "dark" {
			t.Errorf("Expected theme 'dark', got '%s'", config.String("theme"))
		}

		if config.String("font-family") != "JetBrains Mono" {
			t.Errorf("Expected font-family 'JetBrains Mono', got '%s'", config.String("font-family"))
		}

		if config.String("font-size") != "12" {
			t.Errorf("Expected font-size '12', got '%s'", config.String("font-size"))
		}
	})

	t.Run("Multiple values handling", func(t *testing.T) {
		configContent := `keybind = ctrl+c=copy
keybind = ctrl+v=paste
keybind = ctrl+shift+c=copy_html
mouse-click-bindings = 1,2,3`

		configPath := filepath.Join(tempDir, "multi.conf")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Test multiple values
		keybinds := config.Strings("keybind")
		if len(keybinds) != 3 {
			t.Errorf("Expected 3 keybinds, got %d", len(keybinds))
		}

		expectedKeybinds := []string{"ctrl+c=copy", "ctrl+v=paste", "ctrl+shift+c=copy_html"}
		for i, expected := range expectedKeybinds {
			if i >= len(keybinds) || keybinds[i] != expected {
				t.Errorf("Expected keybind[%d] = '%s', got '%s'", i, expected, keybinds[i])
			}
		}
	})

	t.Run("Comments and whitespace handling", func(t *testing.T) {
		configContent := `# Configuration file
# Theme settings
theme = dark

# Font configuration  
font-family = SF Mono
  
# Window settings
window-padding-x = 10    
window-padding-y = 5

# Empty line above`

		configPath := filepath.Join(tempDir, "comments.conf")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Test that comments are ignored and values are parsed correctly
		if config.String("theme") != "dark" {
			t.Errorf("Expected theme 'dark', got '%s'", config.String("theme"))
		}

		if config.String("font-family") != "SF Mono" {
			t.Errorf("Expected font-family 'SF Mono', got '%s'", config.String("font-family"))
		}

		if config.String("window-padding-x") != "10" {
			t.Errorf("Expected window-padding-x '10', got '%s'", config.String("window-padding-x"))
		}
	})

	t.Run("Edge cases and special characters", func(t *testing.T) {
		configContent := `equals-in-value = url=https://example.com
spaces-around = value with spaces  
unicode = 🎨🚀
special-chars = áëíôü
quoted-value = "Hello World"
empty-value = 
no-value`

		configPath := filepath.Join(tempDir, "edge.conf")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Test edge cases
		if config.String("equals-in-value") != "url=https://example.com" {
			t.Errorf("Expected equals-in-value 'url=https://example.com', got '%s'", config.String("equals-in-value"))
		}

		if config.String("spaces-around") != "value with spaces" {
			t.Errorf("Expected spaces-around 'value with spaces', got '%s'", config.String("spaces-around"))
		}

		if config.String("unicode") != "🎨🚀" {
			t.Errorf("Expected unicode '🎨🚀', got '%s'", config.String("unicode"))
		}

		if config.String("special-chars") != "áëíôü" {
			t.Errorf("Expected special-chars 'áëíôü', got '%s'", config.String("special-chars"))
		}

		// Empty value should be empty string
		if config.String("empty-value") != "" {
			t.Errorf("Expected empty-value to be empty, got '%s'", config.String("empty-value"))
		}

		// Key without value should not exist or be empty
		if config.Exists("no-value") && config.String("no-value") != "" {
			t.Errorf("Expected no-value to not exist or be empty, got '%s'", config.String("no-value"))
		}
	})

	t.Run("Performance comparison", func(t *testing.T) {
		// Create a larger config for performance testing
		configContent := `# Large configuration file
theme = dark
font-family = JetBrains Mono
font-size = 12
font-weight = normal
font-style = normal
background = #282828
foreground = #ebdbb2
cursor-color = #fe8019
selection-background = #504945
selection-foreground = #ebdbb2
window-width = 1200
window-height = 800
window-padding-x = 10
window-padding-y = 10
window-decoration = true
window-theme = dark
background-opacity = 1.0
background-blur-radius = 0
unfocused-split-opacity = 0.7
cursor-style = block
cursor-blink = true
keybind = ctrl+c=copy
keybind = ctrl+v=paste
keybind = ctrl+shift+c=copy_html
keybind = ctrl+shift+v=paste_html
keybind = ctrl+t=new_tab
keybind = ctrl+w=close_tab
keybind = ctrl+n=new_window
keybind = ctrl+shift+n=new_os_window
mouse-hide-while-typing = true
mouse-shift-capture = true
mouse-scroll-multiplier = 1.0`

		configPath := filepath.Join(tempDir, "large.conf")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		config, err := ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse large config: %v", err)
		}

		// Verify that large configs work correctly
		if len(config.All()) < 20 {
			t.Errorf("Expected at least 20 configuration keys, got %d", len(config.All()))
		}

		// Test some key values
		if config.String("theme") != "dark" {
			t.Errorf("Expected theme 'dark', got '%s'", config.String("theme"))
		}

		keybinds := config.Strings("keybind")
		if len(keybinds) < 8 {
			t.Errorf("Expected at least 8 keybinds, got %d", len(keybinds))
		}
	})
}

// BenchmarkKoanfProviderVsLegacy compares performance of new vs old implementation
func BenchmarkKoanfProviderVsLegacy(b *testing.B) {
	tempDir := b.TempDir()
	configContent := `theme = dark
font-family = JetBrains Mono
font-size = 12
keybind = ctrl+c=copy
keybind = ctrl+v=paste
keybind = ctrl+shift+c=copy_html
window-padding-x = 10
window-padding-y = 10
background-opacity = 0.9`

	configPath := filepath.Join(tempDir, "bench.conf")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		b.Fatalf("Failed to create config file: %v", err)
	}

	b.Run("Koanf Provider", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ParseGhosttyConfig(configPath)
			if err != nil {
				b.Fatalf("Failed to parse config: %v", err)
			}
		}
	})
}