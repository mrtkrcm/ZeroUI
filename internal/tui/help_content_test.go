package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

func TestHelpContainsPresetsAndChangedOnly(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	model.state = HelpView
	model.helpSystem.ShowPage("configuration")

	view := model.View()
	assert.Contains(t, view, "Changed Only", "Help should mention changed-only")
	assert.Contains(t, view, "Presets", "Help should mention presets")
}
