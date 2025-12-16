package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/validation"
	app "github.com/mrtkrcm/ZeroUI/internal/tui/components/app"
)

func TestPresetsSelectorFlow(t *testing.T) {
	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "ghostty")
	require.NoError(t, err)

	// Simulate showing presets for ghostty
	model.state = HelpView
	model.presetSel.Show("ghostty")
	model.presetSel.SetPresets([]string{"minimal", "dark", "light"})

	// Move selection down and select
	updated, _ := model.presetSel.Update(app.PresetSelectedMsg{App: "ghostty", Name: "minimal"})
	assert.NotNil(t, updated)
}
