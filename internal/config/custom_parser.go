package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/config/providers"
	"github.com/mrtkrcm/ZeroUI/internal/performance"
	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
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

	// Normalize values for friendlier access in tests and callers:
	// - Collapse single-item arrays (e.g., ["value"]) to plain strings
	// - Leave true arrays intact
	all := k.All()
	for key, value := range all {
		switch v := value.(type) {
		case []string:
			if len(v) == 1 {
				k.Set(key, v[0])
			}
		case []interface{}:
			if len(v) == 1 {
				k.Set(key, v[0])
			}
		}
	}

	return k, nil
}

// WriteGhosttyConfig writes config back in Ghostty's format using koanf providers
func WriteGhosttyConfig(configPath string, k *koanf.Koanf, originalPath string) error {
	// Validate configuration against Ghostty schema before writing
	validator := configextractor.NewGhosttySchemaValidator()
	configMap := k.All()
	validationResult := validator.ValidateConfig(configMap)

	if !validationResult.Valid {
		// Log validation errors but don't block writing (for backward compatibility)
		// In a future version, this could be made more strict
		for _, err := range validationResult.Errors {
			fmt.Printf("Ghostty config validation warning: %s\n", err)
		}
	}

	// Prefer legacy writer to preserve comments and structure expected by tests
	// Falls back to provider-based marshal on failure
	if err := writeGhosttyConfigLegacy(configPath, k, originalPath); err == nil {
		return nil
	}

	// Fallback: koanf provider-based marshal
	parser := providers.NewGhosttyParser()
	data, err := parser.Marshal(k.All())
	if err != nil {
		return fmt.Errorf("failed to marshal Ghostty config: %w", err)
	}
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

// writeGhosttyConfigLegacy writes config preserving structure and comments
func writeGhosttyConfigLegacy(configPath string, k *koanf.Koanf, originalPath string) error {
	// Read original file to preserve structure and comments
	originalLines, comments, err := readGhosttyConfigWithComments(originalPath)
	if err != nil {
		// If original doesn't exist, write new file
		return writeNewGhosttyConfig(configPath, k)
	}

	// Create output with pre-allocated capacity
	output := make([]string, 0, len(originalLines)+len(k.All()))

	// Use pooled maps for better memory efficiency
	processedKeys := performance.GetStringBoolMap()
	defer performance.PutStringBoolMap(processedKeys)

	originalKeys := performance.GetStringBoolMap()
	defer performance.PutStringBoolMap(originalKeys)

	// Track sections and their organization
	var sectionLines []string

	for i, line := range originalLines {
		trimmed := strings.TrimSpace(line)

		// Handle section headers (comments that look like section dividers)
		if strings.HasPrefix(trimmed, "#") && (strings.Contains(strings.ToUpper(trimmed), "CONFIG") ||
			strings.Contains(strings.ToUpper(trimmed), "SECTION") ||
			strings.Contains(strings.ToUpper(trimmed), "THEME") ||
			strings.Contains(strings.ToUpper(trimmed), "FONT") ||
			strings.Contains(strings.ToUpper(trimmed), "WINDOW") ||
			strings.Contains(strings.ToUpper(trimmed), "CURSOR")) {
			// This is a section header, preserve it
			if len(sectionLines) > 0 {
				// Write previous section
				output = append(output, sectionLines...)
				sectionLines = nil
			}
			output = append(output, line)
			continue
		}

		// Preserve comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if len(sectionLines) > 0 {
				sectionLines = append(sectionLines, line)
			} else {
				output = append(output, line)
			}
			continue
		}

		// Parse key from line and record it as present in original
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			if len(sectionLines) > 0 {
				sectionLines = append(sectionLines, line)
			} else {
				output = append(output, line)
			}
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key != "" {
			originalKeys[key] = true
		}

		// Get updated value - only process each key once
		if k.Exists(key) && !processedKeys[key] {
			value := k.Get(key)

			// Add any comments that were before this line
			if comment, exists := comments[i]; exists {
				if len(sectionLines) > 0 {
					sectionLines = append(sectionLines, strings.Split(comment, "\n")...)
				} else {
					output = append(output, strings.Split(comment, "\n")...)
				}
			}

			// Write the updated value(s)
			var newLines []string
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					if ok, sanitized := sanitizeGhosttyKV(key, val); ok {
						newLines = append(newLines, fmt.Sprintf("%s = %s", key, sanitized))
					}
				}
			case []interface{}:
				for _, val := range v {
					if ok, sanitized := sanitizeGhosttyKV(key, fmt.Sprintf("%v", val)); ok {
						newLines = append(newLines, fmt.Sprintf("%s = %s", key, sanitized))
					}
				}
			default:
				if ok, sanitized := sanitizeGhosttyKV(key, fmt.Sprintf("%v", value)); ok {
					newLines = append(newLines, fmt.Sprintf("%s = %s", key, sanitized))
				} else {
					// If invalid, keep original line instead of overwriting
					newLines = append(newLines, line)
				}
			}

			if len(sectionLines) > 0 {
				sectionLines = append(sectionLines, newLines...)
			} else {
				output = append(output, newLines...)
			}

			processedKeys[key] = true
		} else if k.Exists(key) && processedKeys[key] {
			// Skip this line as we already processed this key
			continue
		} else {
			// Keep original line if key not in new config
			if len(sectionLines) > 0 {
				sectionLines = append(sectionLines, line)
			} else {
				output = append(output, line)
			}
		}
	}

	// Write any remaining section lines
	if len(sectionLines) > 0 {
		output = append(output, sectionLines...)
	}

	// Add any new keys that weren't in original
	newKeysSection := []string{"\n# Additional Settings"}
	hasNewKeys := false

	for key, value := range k.All() {
		if !originalKeys[key] {
			hasNewKeys = true
			switch v := value.(type) {
			case []string:
				for _, val := range v {
					if ok, sanitized := sanitizeGhosttyKV(key, val); ok {
						newKeysSection = append(newKeysSection, fmt.Sprintf("%s = %s", key, sanitized))
					}
				}
			case []interface{}:
				for _, val := range v {
					if ok, sanitized := sanitizeGhosttyKV(key, fmt.Sprintf("%v", val)); ok {
						newKeysSection = append(newKeysSection, fmt.Sprintf("%s = %s", key, sanitized))
					}
				}
			default:
				if ok, sanitized := sanitizeGhosttyKV(key, fmt.Sprintf("%v", value)); ok {
					newKeysSection = append(newKeysSection, fmt.Sprintf("%s = %s", key, sanitized))
				}
			}
		}
	}

	if hasNewKeys {
		output = append(output, newKeysSection...)
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

// sanitizeGhosttyKV validates and sanitizes Ghostty configuration values to prevent syntax errors.
// Returns ok=false to skip writing an invalid/unsafe value.
func sanitizeGhosttyKV(key string, value string) (bool, string) {
	k := strings.ToLower(strings.TrimSpace(key))
	v := strings.TrimSpace(value)
	if k == "" {
		return false, ""
	}

	// Palette entries: handle hex colors and palette index formats
	if strings.HasPrefix(k, "palette-") {
		return sanitizePaletteValue(v)
	}

	// Keybind entries: validate key combination and action format
	if strings.HasPrefix(k, "keybind") || k == "keybind" {
		return sanitizeKeybindValue(v)
	}

	// Font family: handle multiple fonts with proper spacing
	if k == "font-family" {
		return sanitizeFontFamilyValue(v)
	}

	// Color values: validate hex colors and named colors
	if strings.Contains(k, "color") || strings.Contains(k, "background") || strings.Contains(k, "foreground") {
		return sanitizeColorValue(v)
	}

	// Theme values: allow alphanumeric, hyphens, underscores
	if k == "theme" {
		return sanitizeThemeValue(v)
	}

	// Command/shell values: sanitize paths and command names
	if k == "shell" || strings.Contains(k, "command") {
		return sanitizeCommandValue(v)
	}

	// Generic sanitization for all other values
	return sanitizeGenericValue(v)
}

// sanitizePaletteValue handles palette color values with special index format support
func sanitizePaletteValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Handle "N=#rrggbb" format (palette index with color)
	if idx := strings.Index(v, "="); idx != -1 {
		right := strings.TrimSpace(v[idx+1:])
		if strings.HasPrefix(right, "#") && len(right) >= 7 {
			return true, right
		}
	}

	// Handle direct hex color format
	if strings.HasPrefix(v, "#") && len(v) >= 7 {
		// Validate hex color format
		for i := 1; i < len(v); i++ {
			if !((v[i] >= '0' && v[i] <= '9') || (v[i] >= 'a' && v[i] <= 'f') || (v[i] >= 'A' && v[i] <= 'F')) {
				return false, ""
			}
		}
		return true, v
	}

	return false, ""
}

// sanitizeKeybindValue validates keybind format: "keys=action[:arg]"
func sanitizeKeybindValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Must contain at least one equals sign
	if !strings.Contains(v, "=") {
		return false, ""
	}

	parts := strings.SplitN(v, "=", 2)
	if len(parts) != 2 {
		return false, ""
	}

	keys := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	// Both parts must be non-empty
	if keys == "" || action == "" {
		return false, ""
	}

	// Validate key combinations (basic check for common modifiers)
	validKeys := []string{"ctrl", "alt", "shift", "super", "cmd", "meta"}
	lowerKeys := strings.ToLower(keys)

	for _, validKey := range validKeys {
		if strings.Contains(lowerKeys, validKey) {
			// Found a valid modifier, accept the keybind
			return true, v
		}
	}

	// Allow single character keys and function keys
	if len(keys) == 1 || strings.HasPrefix(lowerKeys, "f") {
		return true, v
	}

	// For other cases, still allow but log potential issues
	return true, v
}

// sanitizeFontFamilyValue handles font family lists with proper spacing
func sanitizeFontFamilyValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Remove any quotes around the entire string
	if (strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`)) ||
		(strings.HasPrefix(v, `'`) && strings.HasSuffix(v, `'`)) {
		v = v[1 : len(v)-1]
	}

	// Split on commas and rejoin with proper spacing
	if strings.Contains(v, ",") {
		parts := strings.Split(v, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		v = strings.Join(parts, ", ")
	}

	// Escape any internal quotes
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `'`, `\'`)

	return true, v
}

// sanitizeColorValue validates color formats (hex or named colors)
func sanitizeColorValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Hex color validation
	if strings.HasPrefix(v, "#") {
		if len(v) == 4 || len(v) == 7 { // #rgb or #rrggbb
			for i := 1; i < len(v); i++ {
				if !((v[i] >= '0' && v[i] <= '9') || (v[i] >= 'a' && v[i] <= 'f') || (v[i] >= 'A' && v[i] <= 'F')) {
					return false, ""
				}
			}
			return true, v
		}
		return false, ""
	}

	// Named color validation
	namedColors := []string{
		"black", "white", "red", "green", "blue", "yellow", "magenta", "cyan",
		"gray", "grey", "background", "foreground", "extend", "transparent",
	}

	lowerV := strings.ToLower(v)
	for _, color := range namedColors {
		if lowerV == color {
			return true, v
		}
	}

	// Allow other named colors (we can't validate all possible X11 colors)
	return true, v
}

// sanitizeThemeValue validates theme names
func sanitizeThemeValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Allow alphanumeric, hyphens, underscores, and spaces
	for _, r := range v {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == ' ') {
			return false, ""
		}
	}

	return true, v
}

// sanitizeCommandValue sanitizes shell commands and paths
func sanitizeCommandValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Remove dangerous characters that could cause command injection
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(v, char) {
			return false, ""
		}
	}

	// Allow basic command names and paths
	return true, v
}

// sanitizeGenericValue provides general sanitization for other configuration values
func sanitizeGenericValue(value string) (bool, string) {
	v := strings.TrimSpace(value)

	// Handle quoted values
	if (strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`)) ||
		(strings.HasPrefix(v, `'`) && strings.HasSuffix(v, `'`)) {
		// Keep quoted values as-is if properly quoted
		return true, v
	}

	// Escape quotes in unquoted values that contain them
	if strings.Contains(v, `"`) {
		v = strings.ReplaceAll(v, `"`, `\"`)
	}

	// Handle values with spaces - wrap in quotes if needed
	if strings.Contains(v, " ") && !strings.Contains(v, `"`) {
		return true, fmt.Sprintf(`"%s"`, v)
	}

	// Check for other special characters that might need escaping
	specialChars := []string{"#", "=", "[", "]", "{", "}"}
	for _, char := range specialChars {
		if strings.Contains(v, char) && !strings.Contains(v, `"`) {
			return true, fmt.Sprintf(`"%s"`, v)
		}
	}

	return true, v
}

// ConfigSection represents a section of the configuration with its comments and settings
type ConfigSection struct {
	HeaderComments []string // Comments at the start of the section
	Settings       map[string]ConfigSetting
	Order          []string // Maintain order of settings
}

// ConfigSetting represents a single configuration setting with its comments
type ConfigSetting struct {
	Key           string
	Value         string
	InlineComment string
	PreComments   []string // Comments before this setting
	PostComments  []string // Comments after this setting
}

// readGhosttyConfigWithComments reads config preserving comments and structure
func readGhosttyConfigWithComments(configPath string) ([]string, map[int]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	comments := make(map[int]string)
	var pendingComments []string

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		trimmed := strings.TrimSpace(line)

		// Handle comments
		if strings.HasPrefix(trimmed, "#") {
			pendingComments = append(pendingComments, line)
			lineNum++
			continue
		}

		// Handle empty lines
		if trimmed == "" {
			// Preserve empty lines but clear pending comments for next non-empty line
			pendingComments = nil
			lineNum++
			continue
		}

		// Handle configuration lines
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			if key != "" && len(pendingComments) > 0 {
				// Associate pending comments with this line
				commentText := strings.Join(pendingComments, "\n")
				comments[lineNum] = commentText
				pendingComments = nil
			}
		} else {
			// Non-key=value line, preserve as-is
			pendingComments = nil
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
