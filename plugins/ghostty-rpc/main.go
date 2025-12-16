package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/mrtkrcm/ZeroUI/internal/plugins/rpc"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// main starts the RPC plugin
func main() {
	logger := log.New(os.Stderr, "[ghostty-rpc] ", log.LstdFlags)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: rpc.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"config": &rpc.ConfigPluginGRPC{
				Impl: &GhosttyRPCPlugin{logger: logger},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// GhosttyRPCPlugin implements the RPC ConfigPlugin interface
type GhosttyRPCPlugin struct {
	logger *log.Logger
}

// GetInfo returns plugin metadata
func (p *GhosttyRPCPlugin) GetInfo(ctx context.Context) (*rpc.PluginInfo, error) {
	return &rpc.PluginInfo{
		Name:        "ghostty-rpc",
		Version:     "1.0.0",
		Description: "Ghostty terminal emulator configuration management (RPC version)",
		Author:      "ConfigToggle",
		Capabilities: []string{
			rpc.CapabilityConfigParsing,
			rpc.CapabilityConfigWriting,
			rpc.CapabilityValidation,
			rpc.CapabilitySchemaExport,
			rpc.CapabilityPresets,
		},
		ApiVersion: rpc.CurrentAPIVersion,
		Metadata: map[string]string{
			"type":   "rpc",
			"format": "ghostty",
		},
	}, nil
}

// DetectConfig attempts to find the Ghostty configuration file
func (p *GhosttyRPCPlugin) DetectConfig(ctx context.Context) (*rpc.ConfigInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &rpc.ConfigInfo{Discovered: false}, err
	}

	// Check common locations
	paths := []string{
		home + "/.config/ghostty/config",
		home + "/Library/Application Support/com.mitchellh.ghostty/config",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			stat, _ := os.Stat(path)
			return &rpc.ConfigInfo{
				Path:         path,
				Format:       "ghostty",
				Discovered:   true,
				LastModified: timestamppb.New(stat.ModTime()),
			}, nil
		}
	}

	// Return default path even if it doesn't exist
	defaultPath := home + "/.config/ghostty/config"
	return &rpc.ConfigInfo{
		Path:       defaultPath,
		Format:     "ghostty",
		Discovered: false,
		Suggestions: []string{
			"Create " + defaultPath,
			"Check ~/.config/ghostty/config",
			"Check ~/Library/Application Support/com.mitchellh.ghostty/config",
		},
	}, nil
}

// ParseConfig reads and parses a configuration file
func (p *GhosttyRPCPlugin) ParseConfig(ctx context.Context, path string) (*rpc.ConfigData, error) {
	config, err := p.parseGhosttyConfig(path)
	if err != nil {
		return nil, err
	}

	// Convert to protobuf format
	protoFields := make(map[string]*anypb.Any)
	for key, value := range config {
		anyValue, err := convertInterfaceToAny(value)
		if err != nil {
			p.logger.Printf("Failed to convert field %s: %v", key, err)
			continue
		}
		protoFields[key] = anyValue
	}

	metadata, err := p.GetSchema(ctx)
	if err != nil {
		return nil, err
	}

	return &rpc.ConfigData{
		Fields:   protoFields,
		Metadata: metadata,
	}, nil
}

// WriteConfig writes configuration data to a file
func (p *GhosttyRPCPlugin) WriteConfig(ctx context.Context, path string, data *rpc.ConfigData) error {
	// Convert from protobuf format
	config := make(map[string]interface{})
	for key, anyValue := range data.Fields {
		value, err := convertAnyToInterface(anyValue)
		if err != nil {
			p.logger.Printf("Failed to convert field %s: %v", key, err)
			continue
		}
		config[key] = value
	}

	return p.writeGhosttyConfig(path, config)
}

// ValidateField validates a single field value
func (p *GhosttyRPCPlugin) ValidateField(ctx context.Context, field string, value interface{}) error {
	return p.validateGhosttyField(field, value)
}

// ValidateConfig validates entire configuration
func (p *GhosttyRPCPlugin) ValidateConfig(ctx context.Context, data *rpc.ConfigData) error {
	for field, anyValue := range data.Fields {
		value, err := convertAnyToInterface(anyValue)
		if err != nil {
			return fmt.Errorf("field %s: invalid value format: %w", field, err)
		}

		if err := p.validateGhosttyField(field, value); err != nil {
			return fmt.Errorf("field %s: %w", field, err)
		}
	}
	return nil
}

// GetSchema returns the current configuration schema
func (p *GhosttyRPCPlugin) GetSchema(ctx context.Context) (*rpc.ConfigMetadata, error) {
	fieldMeta := p.getGhosttyFieldMetadata()
	presets := p.getGhosttyPresets()

	// Convert to protobuf format
	protoFields := make(map[string]*rpc.FieldMetadata)
	for name, field := range fieldMeta {
		protoFields[name] = &rpc.FieldMetadata{
			Type:        field.Type,
			Description: field.Description,
			Required:    false, // Ghostty fields are optional
			Options:     field.Values,
		}
	}

	protoPresets := make(map[string]*rpc.PresetData)
	for name, preset := range presets {
		protoValues := make(map[string]*anypb.Any)
		for key, value := range preset.Values {
			anyValue, err := convertInterfaceToAny(value)
			if err != nil {
				p.logger.Printf("Failed to convert preset %s field %s: %v", name, key, err)
				continue
			}
			protoValues[key] = anyValue
		}

		protoPresets[name] = &rpc.PresetData{
			Name:        preset.Name,
			Description: preset.Description,
			Values:      protoValues,
		}
	}

	return &rpc.ConfigMetadata{
		Fields:  protoFields,
		Presets: protoPresets,
		Schema: &rpc.SchemaInfo{
			Version:    "1.0.0",
			Compatible: true,
		},
	}, nil
}

// SupportsFeature checks if plugin supports a specific feature
func (p *GhosttyRPCPlugin) SupportsFeature(ctx context.Context, feature string) (bool, error) {
	switch feature {
	case rpc.CapabilityConfigParsing,
		rpc.CapabilityConfigWriting,
		rpc.CapabilityValidation,
		rpc.CapabilitySchemaExport,
		rpc.CapabilityPresets:
		return true, nil
	default:
		return false, nil
	}
}

// Helper functions

// FieldMeta represents field metadata (matching legacy format)
type FieldMeta struct {
	Type        string
	Values      []string
	Default     interface{}
	Description string
}

// Preset represents a configuration preset
type Preset struct {
	Name        string
	Description string
	Values      map[string]interface{}
}

// convertInterfaceToAny converts interface{} to protobuf Any
func convertInterfaceToAny(value interface{}) (*anypb.Any, error) {
	if value == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	any := &anypb.Any{
		TypeUrl: "type.googleapis.com/google.protobuf.Value",
		Value:   jsonData,
	}

	return any, nil
}

// convertAnyToInterface converts protobuf Any to interface{}
func convertAnyToInterface(any *anypb.Any) (interface{}, error) {
	if any == nil {
		return nil, nil
	}

	var value interface{}
	err := json.Unmarshal(any.Value, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Any value: %w", err)
	}

	return value, nil
}

// parseGhosttyConfig parses the Ghostty configuration file
func (p *GhosttyRPCPlugin) parseGhosttyConfig(configPath string) (map[string]interface{}, error) {
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

		if key == "" {
			continue
		}

		// Handle multiple values with same key
		if existing, exists := config[key]; exists {
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

	return config, scanner.Err()
}

// writeGhosttyConfig writes the configuration back to file
func (p *GhosttyRPCPlugin) writeGhosttyConfig(configPath string, config map[string]interface{}) error {
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

// readConfigLines reads the config file lines
func (p *GhosttyRPCPlugin) readConfigLines(configPath string) ([]string, error) {
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
func (p *GhosttyRPCPlugin) updateConfigLines(lines []string, config map[string]interface{}) []string {
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
func (p *GhosttyRPCPlugin) writeNewConfig(configPath string, config map[string]interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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
	writer.WriteString("# Generated by configtoggle ghostty-rpc plugin\n\n")

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

// validateGhosttyField validates a value for a specific field
func (p *GhosttyRPCPlugin) validateGhosttyField(field string, value interface{}) error {
	metadata := p.getGhosttyFieldMetadata()

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
			// Try to parse as number - this is handled elsewhere
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

// getGhosttyFieldMetadata returns metadata about configurable fields
func (p *GhosttyRPCPlugin) getGhosttyFieldMetadata() map[string]FieldMeta {
	return map[string]FieldMeta{
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

// getGhosttyPresets returns available presets
func (p *GhosttyRPCPlugin) getGhosttyPresets() map[string]Preset {
	return map[string]Preset{
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
