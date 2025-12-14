package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// Test that applying a preset updates the target config file for a known app (ghostty)
func TestApplyPresetUpdatesConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal ghostty app schema with a preset
	appsDir := filepath.Join(tmpDir, "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	schema := `name: ghostty
path: ` + filepath.Join(tmpDir, "ghostty.conf") + `
format: custom
fields:
  theme:
    type: string
    description: Theme
presets:
  minimal:
    name: minimal
    description: Minimal preset
    values:
      theme: Dracula
`
	require.NoError(t, os.WriteFile(filepath.Join(appsDir, "ghostty.yaml"), []byte(schema), 0644))

	// Ensure target config exists and is empty
	target := filepath.Join(tmpDir, "ghostty.conf")
	require.NoError(t, os.WriteFile(target, []byte(""), 0644))

	// Wire engine to temp dir
	loader, err := config.NewLoader()
	require.NoError(t, err)
	loader.SetConfigDir(tmpDir)
	
	log := logger.Global()
	engine := toggle.NewEngineWithDeps(loader, log)
	configService := service.NewConfigService(engine, loader, log)

	_, err = NewTestModel(configService, "ghostty")
	require.NoError(t, err)

	// Apply preset directly via engine to ensure write path is exercised
	err = engine.ApplyPreset("ghostty", "minimal")
	require.NoError(t, err)

	// Read target config and assert theme set
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Contains(t, string(data), "theme = Dracula")
}
