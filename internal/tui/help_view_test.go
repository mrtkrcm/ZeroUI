package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// TestHelpOverlayHasLiteralTitle validates that the help view contains 'Help'
func TestHelpOverlayHasLiteralTitle(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Trigger help
	model.state = HelpView
	model.helpSystem.ShowPage("overview")

	view := model.View()
	assert.Contains(t, view, "Help", "Help overlay should contain literal 'Help' title")
}
