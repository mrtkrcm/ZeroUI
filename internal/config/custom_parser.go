package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/config/providers"
)

// ParseGhosttyConfig parses Ghostty's custom config format using koanf providers
func ParseGhosttyConfig(configPath string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	
	// Use the new Ghostty provider with built-in parser
	provider := providers.NewGhosttyProviderWithParser(configPath)
	
	// Load the config into koanf
	if err := provider.LoadIntoKoanf(k); err != nil {
		return nil, fmt.Errorf("failed to load Ghostty config: %w", err)
	}
	
	return k, nil
}
// WriteGhosttyConfig writes config back in Ghostty's format using koanf providers
func WriteGhosttyConfig(configPath string, k *koanf.Koanf, originalPath string) error {
	// Use the Ghostty parser to marshal the config
	parser := providers.NewGhosttyParser()
	
	// Convert koanf data to Ghostty format
	data, err := parser.Marshal(k.All())
	if err != nil {
		return fmt.Errorf("failed to marshal Ghostty config: %w", err)
	}
	
	// Write to file (simplified version that doesn't preserve comments)
	// For now, we prioritize using the koanf ecosystem over comment preservation
	return writeConfigToFile(configPath, data)
}

// writeConfigToFile writes data to a config file, creating directories if needed
func writeConfigToFile(configPath string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Deprecated: Use WriteGhosttyConfig with koanf providers instead
func writeGhosttyConfigLegacy(configPath string, k *koanf.Koanf, originalPath string) error {
	// Read original file to preserve structure and comments
	originalLines, comments, err := readGhosttyConfigWithComments(originalPath)
	if err != nil {
		// If original doesn't exist, write new file
		return writeNewGhosttyConfig(configPath, k)
	}

	// Create output
	var output []string
	processedKeys := make(map[string]bool)

	for i, line := range originalLines {
		trimmed := strings.TrimSpace(line)

		// Preserve comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			output = append(output, line)
			continue
		}

		// Parse key from line
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			output = append(output, line)
			continue
		}

		key := strings.TrimSpace(parts[0])

		// Get updated value - only process each key once
		if k.Exists(key) && !processedKeys[key] {
			value := k.Get(key)

			// Add any comments that were before this line
			if comment, exists := comments[i]; exists {
				output = append(output, comment)
			}

			// Write the updated value(s)
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					output = append(output, fmt.Sprintf("%s = %s", key, val))
				}
			case []interface{}:
				for _, val := range v {
					output = append(output, fmt.Sprintf("%s = %v", key, val))
				}
			default:
				output = append(output, fmt.Sprintf("%s = %v", key, value))
			}

			processedKeys[key] = true
		} else if k.Exists(key) && processedKeys[key] {
			// Skip this line as we already processed this key
			continue
		} else {
			// Keep original line if key not in new config
			output = append(output, line)
		}
	}

	// Add any new keys that weren't in original
	for key, value := range k.All() {
		if !processedKeys[key] {
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					output = append(output, fmt.Sprintf("%s = %s", key, val))
				}
			case []interface{}:
				for _, val := range v {
					output = append(output, fmt.Sprintf("%s = %v", key, val))
				}
			default:
				output = append(output, fmt.Sprintf("%s = %v", key, value))
			}
		}
	}

	// Write to file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)
	for _, line := range output {
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return writer.Flush()
}

// readGhosttyConfigWithComments reads config preserving comments
func readGhosttyConfigWithComments(configPath string) ([]string, map[int]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	comments := make(map[int]string)
	var lastComment string

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			lastComment = line
		} else if trimmed != "" && lastComment != "" {
			comments[lineNum] = lastComment
			lastComment = ""
		}

		lineNum++
	}

	return lines, comments, scanner.Err()
}

// writeNewGhosttyConfig writes a new Ghostty config file
func writeNewGhosttyConfig(configPath string, k *koanf.Koanf) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() { _ = file.Close() }()

	writer := bufio.NewWriter(file)

	// Write header
	_, _ = writer.WriteString("# Ghostty Configuration\n")
	_, _ = writer.WriteString("# Generated by configtoggle\n\n")

	// Group keys by category
	categories := map[string][]string{
		"Font":       {"font-family", "font-size", "font-thicken", "font-feature"},
		"Theme":      {"theme", "background", "foreground"},
		"Window":     {"window-padding-x", "window-padding-y", "window-height", "window-width", "window-decoration", "window-theme"},
		"Background": {"background-opacity", "background-blur-radius", "unfocused-split-opacity"},
		"Cursor":     {"cursor-style", "cursor-color"},
	}

	// Write categorized settings
	for category, keys := range categories {
		_, _ = writer.WriteString(fmt.Sprintf("# %s Configuration\n", category))

		for _, key := range keys {
			if k.Exists(key) {
				value := k.Get(key)
				switch v := value.(type) {
				case []string:
					for _, val := range v {
						writer.WriteString(fmt.Sprintf("%s = %s\n", key, val))
					}
				case []interface{}:
					for _, val := range v {
						writer.WriteString(fmt.Sprintf("%s = %v\n", key, val))
					}
				default:
					writer.WriteString(fmt.Sprintf("%s = %v\n", key, value))
				}
			}
		}
		writer.WriteString("\n")
	}

	// Write remaining uncategorized settings
	categorizedKeys := make(map[string]bool)
	for _, keys := range categories {
		for _, key := range keys {
			categorizedKeys[key] = true
		}
	}

	writer.WriteString("# Other Settings\n")
	for key, value := range k.All() {
		if !categorizedKeys[key] {
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					writer.WriteString(fmt.Sprintf("%s = %s\n", key, val))
				}
			case []interface{}:
				for _, val := range v {
					writer.WriteString(fmt.Sprintf("%s = %v\n", key, val))
				}
			default:
				writer.WriteString(fmt.Sprintf("%s = %v\n", key, value))
			}
		}
	}

	return writer.Flush()
}
