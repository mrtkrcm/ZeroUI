package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	app "github.com/mrtkrcm/ZeroUI/internal/tui/components/app"
)

func TestPresetsSelectorFlow(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "ghostty")
	require.NoError(t, err)

	// Simulate showing presets for ghostty
	model.state = HelpView
	model.presetSel.Show("ghostty")
	model.presetSel.SetPresets([]string{"minimal", "dark", "light"})

	// Move selection down and select
	updated, _ := model.presetSel.Update(app.PresetSelectedMsg{App: "ghostty", Name: "minimal"})
	assert.NotNil(t, updated)
}
