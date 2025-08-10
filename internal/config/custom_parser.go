package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
)

// ParseGhosttyConfig parses Ghostty's custom config format
func ParseGhosttyConfig(configPath string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

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
		
		// Handle multiple values with same key (like keybind)
		if existing := k.Get(key); existing != nil {
			// Convert to slice if not already
			switch v := existing.(type) {
			case []string:
				k.Set(key, append(v, value))
			case string:
				k.Set(key, []string{v, value})
			default:
				k.Set(key, []string{fmt.Sprint(v), value})
			}
		} else {
			k.Set(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return k, nil
}

// WriteGhosttyConfig writes config back in Ghostty's format
func WriteGhosttyConfig(configPath string, k *koanf.Koanf, originalPath string) error {
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
		
		// Get updated value
		if k.Exists(key) {
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
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range output {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
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
	defer file.Close()

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
	defer file.Close()

	writer := bufio.NewWriter(file)
	
	// Write header
	writer.WriteString("# Ghostty Configuration\n")
	writer.WriteString("# Generated by configtoggle\n\n")

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
		writer.WriteString(fmt.Sprintf("# %s Configuration\n", category))
		
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