package main

import (
	"context"
	"os"
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/plugins/rpc"
)

func TestGhosttyRPCPlugin(t *testing.T) {
	// Create a test plugin instance
	p := &GhosttyRPCPlugin{
		logger: nil, // Use nil for tests to avoid log output
	}

	ctx := context.Background()

	t.Run("GetInfo", func(t *testing.T) {
		info, err := p.GetInfo(ctx)
		if err != nil {
			t.Fatalf("GetInfo failed: %v", err)
		}

		if info.Name != "ghostty-rpc" {
			t.Errorf("Expected name 'ghostty-rpc', got '%s'", info.Name)
		}

		if info.ApiVersion != rpc.CurrentAPIVersion {
			t.Errorf("Expected API version '%s', got '%s'", rpc.CurrentAPIVersion, info.ApiVersion)
		}

		expectedCapabilities := []string{
			rpc.CapabilityConfigParsing,
			rpc.CapabilityConfigWriting,
			rpc.CapabilityValidation,
			rpc.CapabilitySchemaExport,
			rpc.CapabilityPresets,
		}

		if len(info.Capabilities) != len(expectedCapabilities) {
			t.Errorf("Expected %d capabilities, got %d", len(expectedCapabilities), len(info.Capabilities))
		}
	})

	t.Run("DetectConfig", func(t *testing.T) {
		config, err := p.DetectConfig(ctx)
		if err != nil {
			t.Fatalf("DetectConfig failed: %v", err)
		}

		if config.Format != "ghostty" {
			t.Errorf("Expected format 'ghostty', got '%s'", config.Format)
		}

		// Should return a valid path even if file doesn't exist
		if config.Path == "" {
			t.Error("Expected non-empty path")
		}
	})

	t.Run("GetSchema", func(t *testing.T) {
		schema, err := p.GetSchema(ctx)
		if err != nil {
			t.Fatalf("GetSchema failed: %v", err)
		}

		if len(schema.Fields) == 0 {
			t.Error("Expected field metadata")
		}

		// Check for a known field
		themeField, exists := schema.Fields["theme"]
		if !exists {
			t.Error("Expected 'theme' field in schema")
		} else {
			if themeField.Type != "choice" {
				t.Errorf("Expected theme field type 'choice', got '%s'", themeField.Type)
			}
		}

		if len(schema.Presets) == 0 {
			t.Error("Expected preset data")
		}

		// Check for known preset
		_, exists = schema.Presets["dark-mode"]
		if !exists {
			t.Error("Expected 'dark-mode' preset")
		}
	})

	t.Run("SupportsFeature", func(t *testing.T) {
		// Test supported feature
		supported, err := p.SupportsFeature(ctx, rpc.CapabilityConfigParsing)
		if err != nil {
			t.Errorf("SupportsFeature failed: %v", err)
		}
		if !supported {
			t.Error("Expected config parsing to be supported")
		}

		// Test unsupported feature
		supported, err = p.SupportsFeature(ctx, "unsupported-feature")
		if err != nil {
			t.Errorf("SupportsFeature failed: %v", err)
		}
		if supported {
			t.Error("Expected unsupported feature to return false")
		}
	})

	t.Run("ValidateField", func(t *testing.T) {
		// Test valid choice
		err := p.ValidateField(ctx, "theme", "GruvboxDark")
		if err != nil {
			t.Errorf("Field validation failed: %v", err)
		}

		// Test invalid choice
		err = p.ValidateField(ctx, "theme", "invalid-theme")
		if err == nil {
			t.Error("Expected validation error for invalid theme")
		}

		// Test unknown field (should pass)
		err = p.ValidateField(ctx, "unknown-field", "any-value")
		if err != nil {
			t.Errorf("Unknown field validation should pass: %v", err)
		}
	})
}

func TestGhosttyConfigParsing(t *testing.T) {
	// Create temporary config file for testing
	content := `# Test Ghostty config
theme = GruvboxDark
font-family = JetBrains Mono
font-size = 14
background-opacity = 0.9

# Comment line
window-theme = dark
`

	tmpfile, err := os.CreateTemp("", "ghostty-test-*.config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	p := &GhosttyRPCPlugin{logger: nil}
	ctx := context.Background()

	t.Run("ParseConfig", func(t *testing.T) {
		data, err := p.ParseConfig(ctx, tmpfile.Name())
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		if len(data.Fields) == 0 {
			t.Error("Expected parsed fields")
		}

		// Check if specific fields were parsed
		themeAny, exists := data.Fields["theme"]
		if !exists {
			t.Error("Expected 'theme' field in parsed config")
		} else {
			theme, err := convertAnyToInterface(themeAny)
			if err != nil {
				t.Fatalf("Failed to convert theme value: %v", err)
			}
			if theme != "GruvboxDark" {
				t.Errorf("Expected theme 'GruvboxDark', got '%v'", theme)
			}
		}
	})

	t.Run("WriteConfig", func(t *testing.T) {
		// First parse the existing config
		data, err := p.ParseConfig(ctx, tmpfile.Name())
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		// Modify a field
		newTheme, err := convertInterfaceToAny("nord")
		if err != nil {
			t.Fatalf("Failed to convert new theme: %v", err)
		}
		data.Fields["theme"] = newTheme

		// Write it back
		err = p.WriteConfig(ctx, tmpfile.Name(), data)
		if err != nil {
			t.Fatalf("WriteConfig failed: %v", err)
		}

		// Parse again to verify the change
		newData, err := p.ParseConfig(ctx, tmpfile.Name())
		if err != nil {
			t.Fatalf("ParseConfig after write failed: %v", err)
		}

		themeAny, exists := newData.Fields["theme"]
		if !exists {
			t.Error("Expected 'theme' field after write")
		} else {
			theme, err := convertAnyToInterface(themeAny)
			if err != nil {
				t.Fatalf("Failed to convert updated theme: %v", err)
			}
			if theme != "nord" {
				t.Errorf("Expected updated theme 'nord', got '%v'", theme)
			}
		}
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		data, err := p.ParseConfig(ctx, tmpfile.Name())
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		// Validate the parsed config
		err = p.ValidateConfig(ctx, data)
		if err != nil {
			t.Errorf("Config validation failed: %v", err)
		}

		// Add an invalid field and test validation failure
		invalidTheme, err := convertInterfaceToAny("invalid-theme")
		if err != nil {
			t.Fatalf("Failed to convert invalid theme: %v", err)
		}
		data.Fields["theme"] = invalidTheme

		err = p.ValidateConfig(ctx, data)
		if err == nil {
			t.Error("Expected validation error for invalid config")
		}
	})
}
