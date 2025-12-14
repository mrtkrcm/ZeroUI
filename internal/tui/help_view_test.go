package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// TestHelpOverlayHasLiteralTitle validates that the help view contains 'Help'
func TestHelpOverlayHasLiteralTitle(t *testing.T) {
	log := logger.Global()
	configLoader, err := config.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	engine := toggle.NewEngineWithDeps(configLoader, log)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Trigger help
	model.state = HelpView
	model.helpSystem.ShowPage("overview")

	view := model.View()
	assert.Contains(t, view, "Help", "Help overlay should contain literal 'Help' title")
}
