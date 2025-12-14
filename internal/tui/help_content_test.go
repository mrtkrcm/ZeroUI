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

func TestHelpContainsPresetsAndChangedOnly(t *testing.T) {
	log := logger.Global()
	configLoader, err := config.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	engine := toggle.NewEngineWithDeps(configLoader, log)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	model.state = HelpView
	model.helpSystem.ShowPage("configuration")

	view := model.View()
	assert.Contains(t, view, "Changed Only", "Help should mention changed-only")
	assert.Contains(t, view, "Presets", "Help should mention presets")
}
