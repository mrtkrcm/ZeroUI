package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// TestUIInitialization validates that the UI initializes correctly
func TestUIInitialization(t *testing.T) {
	// Create a test engine
	engine, err := toggle.NewEngine()
	require.NoError(t, err, "Failed to create engine")

	// Create model without initial app (should show grid view)
	model, err := NewTestModel(engine, "")
	require.NoError(t, err, "Failed to create model")

	// Validate initial state
	assert.Equal(t, AppGridView, model.state, "Should start with AppGridView")
	assert.NotNil(t, model.appGrid, "AppGrid should be initialized")
	assert.NotNil(t, model.appSelector, "AppSelector should be initialized")
	assert.NotNil(t, model.configEditor, "ConfigEditor should be initialized")
	assert.NotNil(t, model.statusBar, "StatusBar should be initialized")
	assert.NotNil(t, model.responsiveHelp, "ResponsiveHelp should be initialized")
	assert.NotNil(t, model.styles, "Styles should be initialized")
	assert.NotNil(t, model.theme, "Theme should be initialized")
}

// TestUIRendering validates that the UI renders without panics
func TestUIRendering(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Test initial render
	view := model.View()
	assert.NotEmpty(t, view, "Initial view should not be empty")

	// Simulate window resize
	model.width = 120
	model.height = 40
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(resizeMsg)
	model = updatedModel.(*Model)

	// Test render after resize
	view = model.View()
	assert.NotEmpty(t, view, "View after resize should not be empty")

	// Check that it contains expected elements for AppGridView
	// The logo uses Unicode box drawing characters
	assert.Contains(t, view, "███████╗███████╗██████╗", "Should contain app title/logo")
}

// TestFullscreenLayout validates fullscreen rendering
func TestFullscreenLayout(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Set fullscreen dimensions
	model.width = 120
	model.height = 40

	// Update component sizes
	model.updateComponentSizes()

	// Render view
	view := model.View()

	// Validate dimensions
	lines := strings.Split(view, "\n")
	assert.LessOrEqual(t, len(lines), 40, "Should not exceed terminal height")

	for _, line := range lines {
		// Strip ANSI codes for length check
		cleanLine := stripAnsi(line)
		assert.LessOrEqual(t, len(cleanLine), 120, "Lines should not exceed terminal width")
	}
}

// TestKeyboardNavigation validates keyboard input handling
func TestKeyboardNavigation(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Test navigation keys
	testCases := []struct {
		name     string
		key      tea.KeyMsg
		expected func(*Model) bool
	}{
		{
			name: "Toggle help",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			expected: func(m *Model) bool {
				return m.showingHelp
			},
		},
		{
			name: "Quit key",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			expected: func(m *Model) bool {
				// This would trigger tea.Quit
				return true
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset help state
			model.showingHelp = false

			// Send key message
			updatedModel, cmd := model.Update(tc.key)

			// Check if it's a quit command
			if cmd != nil {
				// Check if it's a quit command by type
				_, isQuit := cmd().(tea.QuitMsg)
				if tc.key.Runes[0] == 'q' {
					assert.True(t, isQuit, "Should return quit command for 'q' key")
				}
			} else {
				// Check model state
				m := updatedModel.(*Model)
				assert.True(t, tc.expected(m), "Model state should match expected")
			}
		})
	}
}

// TestComponentInteraction validates component message handling
func TestComponentInteraction(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Test window size message propagation
	resizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ := model.Update(resizeMsg)
	model = updatedModel.(*Model)

	assert.Equal(t, 100, model.width, "Width should be updated")
	assert.Equal(t, 30, model.height, "Height should be updated")
}

// TestStateTransitions validates view state transitions
func TestStateTransitions(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Start in AppGridView
	assert.Equal(t, AppGridView, model.state)

	// Transition to HuhAppSelectionView
	model.state = HuhAppSelectionView
	model.focusCurrentComponent()

	// Validate focus changed
	assert.Equal(t, HuhAppSelectionView, model.state)

	// Test back navigation
	cmd := model.handleBack()
	assert.Equal(t, AppGridView, model.state, "Should go back to grid view")
	assert.Nil(t, cmd, "Should not quit from grid view navigation")

	// Test quit from grid view
	model.state = AppGridView
	cmd = model.handleBack()
	assert.NotNil(t, cmd, "Should quit from grid view")
}

// Helper function to strip ANSI codes
func stripAnsi(str string) string {
	// Simple ANSI code stripping for testing
	result := str
	for {
		start := strings.Index(result, "\x1b[")
		if start == -1 {
			break
		}
		end := strings.IndexByte(result[start:], 'm')
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

// TestInitCommand validates initialization commands
func TestInitCommand(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Test Init command
	cmd := model.Init()
	assert.NotNil(t, cmd, "Init should return a command")
}

// TestErrorHandling validates error display
func TestErrorHandling(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Set an error
	model.err = assert.AnError

	// Render with error
	view := model.View()
	assert.Contains(t, view, "Error", "Should display error message")
}

// TestHelpView validates help rendering
func TestHelpView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Enable help
	model.showingHelp = true

	// Render help view
	view := model.View()
	assert.NotEmpty(t, view, "Help view should not be empty")

	// Help view should be wrapped with layout
	lines := strings.Split(view, "\n")
	assert.Greater(t, len(lines), 3, "Help view should have multiple lines")
}
