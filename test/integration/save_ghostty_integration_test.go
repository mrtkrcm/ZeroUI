//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cfg "github.com/mrtkrcm/ZeroUI/internal/appconfig"
)

// This integration test ensures we only persist modified keys for Ghostty configs,
// and that unknown/schema-missing keys are not written.
func TestIntegration_Ghostty_SaveOnlyChanges(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeroui-integ-ghostty")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Build a minimal apps schema dir
	appsDir := filepath.Join(tmpDir, "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	ghosttySchema := `name: ghostty
path: ` + filepath.Join(tmpDir, "ghostty.conf") + `
format: custom
fields:
  font-family:
    type: string
    description: Font
  font-size:
    type: number
    description: Size
  theme:
    type: string
    description: Theme
`
	require.NoError(t, os.WriteFile(filepath.Join(appsDir, "ghostty.yaml"), []byte(ghosttySchema), 0644))

	// Original target config with only font-family defined
	original := "font-family = SF Mono\n"
	targetPath := filepath.Join(tmpDir, "ghostty.conf")
	require.NoError(t, os.WriteFile(targetPath, []byte(original), 0644))

	// Create a loader rooted at tmpDir
	loader, err := cfg.NewLoader()
	require.NoError(t, err)
	loader.OverrideConfigDir(tmpDir)

	appCfg, err := loader.LoadAppConfig("ghostty")
	require.NoError(t, err)

	// Load target config
	k, err := loader.LoadTargetConfig(appCfg)
	require.NoError(t, err)

	// Mutate values: change font-family, add font-size, leave theme untouched
	k.Set("font-family", "SF Pro")
	k.Set("font-size", 14)
	// theme remains absent (default), so should not be added

	// Save back
	require.NoError(t, loader.SaveTargetConfig(appCfg, k))

	// Read file
	data, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	text := string(data)

	// Should contain updated font-family line
	require.Contains(t, text, "font-family = SF Pro")
	// Should contain added font-size only because it's explicitly set
	require.Contains(t, text, "font-size = 14")
	// Should NOT include theme (not present originally, not changed)
	require.NotContains(t, text, "theme = ")
}

// Additional integration: sanitize palette updates and skip invalid keybind lines when saving
func TestIntegration_Ghostty_SanitizeAndSkipInvalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeroui-integ-ghostty2")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	appsDir := filepath.Join(tmpDir, "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	ghosttySchema := `name: ghostty
path: ` + filepath.Join(tmpDir, "ghostty.conf") + `
format: custom
fields:
  palette-117:
    type: string
  keybind-31:
    type: string
  keybind-33:
    type: string
`
	require.NoError(t, os.WriteFile(filepath.Join(appsDir, "ghostty.yaml"), []byte(ghosttySchema), 0644))

	original := "palette-117 = #87d7d7\nkeybind-31 = super+ctrl+left=resize_split:left\nkeybind-33 = super+ctrl+up=resize_split:up\n"
	targetPath := filepath.Join(tmpDir, "ghostty.conf")
	require.NoError(t, os.WriteFile(targetPath, []byte(original), 0644))

	loader, err := cfg.NewLoader()
	require.NoError(t, err)
	loader.OverrideConfigDir(tmpDir)

	appCfg, err := loader.LoadAppConfig("ghostty")
	require.NoError(t, err)

	// Load
	k, err := loader.LoadTargetConfig(appCfg)
	require.NoError(t, err)

	// Provide palette in form "116=#87d7d7" that should normalize to color only
	k.Set("palette-117", "116=#87d7d7")
	// Provide malformed keybind that should be skipped on write
	k.Set("keybind-33", "10")

	// Save
	require.NoError(t, loader.SaveTargetConfig(appCfg, k))

	// Verify
	data, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	text := string(data)

	require.Contains(t, text, "palette-117 = #87d7d7")
	require.Contains(t, text, "keybind-31 = super+ctrl+left=resize_split:left")
	require.NotContains(t, text, "keybind-33 = 10")
}
