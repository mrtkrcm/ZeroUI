package builtin

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/plugins"
)

// GhosttyPlugin handles Ghostty terminal configuration
type GhosttyPlugin struct{}

// NewGhosttyPlugin creates a new Ghostty plugin
func NewGhosttyPlugin() *GhosttyPlugin {
	return &GhosttyPlugin{}
}

// Name returns the plugin name
func (p *GhosttyPlugin) Name() string {
	return "ghostty"
}

// Description returns the plugin description
func (p *GhosttyPlugin) Description() string {
	return "Ghostty terminal emulator configuration management"
}

// DetectConfigPath attempts to find the Ghostty configuration file
func (p *GhosttyPlugin) DetectConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check common locations
	paths := []string{
		filepath.Join(home, ".config", "ghostty", "config"),
		filepath.Join(home, "Library", "Application Support", "com.mitchellh.ghostty", "config"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Return default path even if it doesn't exist
	return filepath.Join(home, ".config", "ghostty", "config"), nil
}

// ParseConfig parses the Ghostty configuration file
func (p *GhosttyPlugin) ParseConfig(configPath string) (map[string]interface{}, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Handle multiple values with same key (like font-feature)
		if existing, exists := config[key]; exists {
			// Convert to slice if not already
			switch v := existing.(type) {
			case []string:
				config[key] = append(v, value)
			case string:
				config[key] = []string{v, value}
			}
		} else {
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return config, nil
}

// WriteConfig writes the configuration back to file
func (p *GhosttyPlugin) WriteConfig(configPath string, config map[string]interface{}) error {
	// Read existing file to preserve comments and structure
	originalLines, err := p.readConfigLines(configPath)
	if err != nil {
		// If file doesn't exist, create new one
		return p.writeNewConfig(configPath, config)
	}

	// Update existing config while preserving structure
	updatedLines := p.updateConfigLines(originalLines, config)

	// Write back to file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range updatedLines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	return writer.Flush()
}

// GetFieldMetadata returns metadata about configurable fields
func (p *GhosttyPlugin) GetFieldMetadata() map[string]plugins.FieldMeta {
	return map[string]plugins.FieldMeta{
		"theme": {
			Type: "choice",
			Values: []string{
				"GruvboxLight", "GruvboxDark", "catppuccin-mocha",
				"Everforest Dark - Hard", "nord", "tokyonight",
			},
			Default:     "GruvboxLight",
			Description: "Color theme",
		},
		"font-family": {
			Type: "choice",
			Values: []string{
				"Lekton Nerd Font Mono", "Hurmit Nerd Font Mono",
				"JetBrains Mono", "SF Mono", "Monaco", "Menlo",
				"Fira Code", "Cascadia Mono", "Iosevka Nerd Font Mono",
				"Hack Nerd Font Mono", "Andale Mono",
			},
			Default:     "SF Mono",
			Description: "Font family",
		},
		"font-size": {
			Type:        "number",
			Values:      []string{"10", "11", "12", "13", "14", "15", "16", "18", "20"},
			Default:     14,
			Description: "Font size",
		},
		"background-opacity": {
			Type:        "number",
			Values:      []string{"0.5", "0.6", "0.7", "0.75", "0.8", "0.85", "0.87", "0.9", "0.95", "1.0"},
			Default:     0.9,
			Description: "Background transparency",
		},
		"background-blur-radius": {
			Type:        "number",
			Values:      []string{"0", "10", "20", "30", "40", "50"},
			Default:     30,
			Description: "Background blur radius",
		},
		"cursor-style": {
			Type:        "choice",
			Values:      []string{"block", "bar", "underline"},
			Default:     "block",
			Description: "Cursor style",
		},
		"window-theme": {
			Type:        "choice",
			Values:      []string{"auto", "light", "dark"},
			Default:     "auto",
			Description: "Window theme",
		},
	}
}

// GetPresets returns available presets
func (p *GhosttyPlugin) GetPresets() map[string]plugins.Preset {
	return map[string]plugins.Preset{
		"dark-mode": {
			Name:        "dark-mode",
			Description: "Dark theme with JetBrains Mono",
			Values: map[string]interface{}{
				"theme":                  "GruvboxDark",
				"font-family":            "JetBrains Mono",
				"font-size":              14,
				"background-opacity":     0.9,
				"background-blur-radius": 30,
				"window-theme":           "dark",
			},
		},
		"light-mode": {
			Name:        "light-mode",
			Description: "Light theme with Lekton",
			Values: map[string]interface{}{
				"theme":                  "GruvboxLight",
				"font-family":            "Lekton Nerd Font Mono",
				"font-size":              14,
				"background-opacity":     0.95,
				"background-blur-radius": 20,
				"window-theme":           "light",
			},
		},
		"cyberpunk": {
			Name:        "cyberpunk",
			Description: "Tokyo Night with Hack font",
			Values: map[string]interface{}{
				"theme":                  "tokyonight",
				"font-family":            "Hack Nerd Font Mono",
				"font-size":              13,
				"background-opacity":     0.85,
				"background-blur-radius": 40,
				"cursor-style":           "bar",
				"window-theme":           "dark",
			},
		},
		"minimal": {
			Name:        "minimal",
			Description: "Minimal setup with no transparency",
			Values: map[string]interface{}{
				"theme":                  "nord",
				"font-family":            "SF Mono",
				"font-size":              14,
				"background-opacity":     1.0,
				"background-blur-radius": 0,
				"window-theme":           "auto",
			},
		},
	}
}

// ValidateValue validates a value for a specific field
func (p *GhosttyPlugin) ValidateValue(field string, value interface{}) error {
	metadata := p.GetFieldMetadata()

	fieldMeta, exists := metadata[field]
	if !exists {
		// Allow unknown fields
		return nil
	}

	// Type checking
	switch fieldMeta.Type {
	case "number":
		switch value.(type) {
		case float64, int, int64:
			// Valid number types
		case string:
			// Try to parse as number
			// This is handled elsewhere
		default:
			return fmt.Errorf("field %s expects a number, got %T", field, value)
		}
	case "choice":
		strValue := fmt.Sprintf("%v", value)
		if len(fieldMeta.Values) > 0 {
			valid := false
			for _, validValue := range fieldMeta.Values {
				if validValue == strValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid value %s for field %s", strValue, field)
			}
		}
	}

	return nil
}

// GetHooks returns hooks to run for various events
func (p *GhosttyPlugin) GetHooks() map[string]string {
	return map[string]string{
		"post-toggle": "osascript -e 'tell application \"System Events\" to keystroke \",\" using {command down, shift down}'",
		"post-preset": "osascript -e 'tell application \"System Events\" to keystroke \",\" using {command down, shift down}'",
	}
}

// readConfigLines reads the config file lines
func (p *GhosttyPlugin) readConfigLines(configPath string) ([]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// updateConfigLines updates config lines with new values
func (p *GhosttyPlugin) updateConfigLines(lines []string, config map[string]interface{}) []string {
	updated := make(map[string]bool)
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			result = append(result, line)
			continue
		}

		// Check if this is a key=value line
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			result = append(result, line)
			continue
		}

		key := strings.TrimSpace(parts[0])

		// If we have a new value for this key, use it
		if newValue, exists := config[key]; exists {
			switch v := newValue.(type) {
			case []string:
				// Handle multiple values
				for _, val := range v {
					result = append(result, fmt.Sprintf("%s = %s", key, val))
				}
			default:
				result = append(result, fmt.Sprintf("%s = %v", key, newValue))
			}
			updated[key] = true
		} else {
			// Keep the original line
			result = append(result, line)
		}
	}

	// Add any new keys that weren't in the original file
	for key, value := range config {
		if !updated[key] {
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					result = append(result, fmt.Sprintf("%s = %s", key, val))
				}
			default:
				result = append(result, fmt.Sprintf("%s = %v", key, value))
			}
		}
	}

	return result
}

// writeNewConfig writes a completely new config file
func (p *GhosttyPlugin) writeNewConfig(configPath string, config map[string]interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	writer.WriteString("# Ghostty Configuration\n")
	writer.WriteString("# Generated by configtoggle\n\n")

	// Write config values
	for key, value := range config {
		switch v := value.(type) {
		case []string:
			for _, val := range v {
				writer.WriteString(fmt.Sprintf("%s = %s\n", key, val))
			}
		default:
			writer.WriteString(fmt.Sprintf("%s = %v\n", key, value))
		}
	}

	return writer.Flush()
}
