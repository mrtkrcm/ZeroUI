package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

func TestOpenPresetsAndSelect(t *testing.T) {
	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	engine := toggle.NewEngineWithDeps(configLoader, log)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "ghostty")
	require.NoError(t, err)

	// Enter form view for ghostty
	model.state = FormView
	model.currentApp = "ghostty"

	// Press 'p' to open selector
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m := updated.(*Model)

	assert.True(t, m.presetSel.IsVisible(), "Preset selector should be visible")

	// Simulate pressing Enter if any preset exists (engine provides names)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = updated
}
