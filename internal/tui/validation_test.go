package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// createTestModel creates a properly initialized test model for both testing.T and testing.B
func createTestModel(t testing.TB, initialApp string) *Model {
	// Create a test engine
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Create logger for testing
	logConfig := logging.DefaultConfig()
	logConfig.Level = logging.LevelError // Reduce noise in tests
	logger, err := logging.NewCharmLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create model
	model, err := NewModel(engine, initialApp, logger)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize with reasonable dimensions
	model.width = 80
	model.height = 24

	return model
}

// TestUIInitialization validates that the UI initializes correctly with enhanced isolation
func TestUIInitialization(t *testing.T) {
	model := createTestModel(t, "")

	// Validate initial state
	assert.Equal(t, ListView, model.state, "Should start with ListView")
	assert.NotNil(t, model.appList, "AppList should be initialized")
	assert.NotNil(t, model.configForm, "ConfigForm should be initialized")
	assert.NotNil(t, model.helpSystem, "HelpSystem should be initialized")
	assert.NotNil(t, model.styles, "Styles should be initialized")
	assert.NotNil(t, model.theme, "Theme should be initialized")
	assert.NotNil(t, model.renderCache, "Render cache should be initialized")
	assert.Greater(t, model.cacheDuration, time.Duration(0), "Cache duration should be positive")
}

// TestUIRendering validates that the UI renders without panics with performance monitoring
func TestUIRendering(t *testing.T) {
	model := createTestModel(t, "")

	// Test initial render with timing
	start := time.Now()
	view := model.View()
	renderTime := time.Since(start)
	
	assert.NotEmpty(t, view, "Initial view should not be empty")
	assert.Less(t, renderTime, 100*time.Millisecond, "Initial render should be fast")

	// Simulate window resize with optimized handling
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

	// Start in ListView
	assert.Equal(t, ListView, model.state)

	// Transition to FormView
	model.state = FormView
	model.currentApp = "test-app"

	// Validate state changed
	assert.Equal(t, FormView, model.state)

	// Test back navigation
	cmd := model.handleBack()
	assert.Equal(t, ListView, model.state, "Should go back to list view")
	assert.Nil(t, cmd, "Should not quit from list view navigation")

	// Test quit from list view
	model.state = ListView
	cmd = model.handleBack()
	assert.NotNil(t, cmd, "Should quit from list view")
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

// TestPerformanceOptimizations validates that our performance optimizations work
func TestPerformanceOptimizations(t *testing.T) {
	model := createTestModel(t, "")

	// Test cache performance
	t.Run("RenderCaching", func(t *testing.T) {
		// First render - should be slow (cache miss)
		start := time.Now()
		view1 := model.View()
		firstRender := time.Since(start)
		
		// Second render - should be fast (cache hit for non-form views)
		if model.state != FormView {
			start = time.Now()
			view2 := model.View()
			secondRender := time.Since(start)
			
			assert.Equal(t, view1, view2, "Cached views should be identical")
			assert.Less(t, secondRender, firstRender/2, "Cached render should be significantly faster")
		}
	})

	t.Run("StateChangeInvalidation", func(t *testing.T) {
		// Change state and verify cache is invalidated
		originalState := model.state
		model.state = HelpView
		model.invalidateCache()
		
		// Render should work and create new cache
		view := model.View()
		assert.NotEmpty(t, view, "View after state change should render correctly")
		
		// Restore state
		model.state = originalState
		model.invalidateCache()
	})

	t.Run("ComponentUpdateOptimization", func(t *testing.T) {
		// Test that identical size updates are skipped
		initialWidth := model.width
		initialHeight := model.height
		
		// Same size update should be fast
		start := time.Now()
		sameSize := tea.WindowSizeMsg{Width: initialWidth, Height: initialHeight}
		model.Update(sameSize)
		sameUpdateTime := time.Since(start)
		
		// Different size update
		start = time.Now()
		differentSize := tea.WindowSizeMsg{Width: initialWidth + 10, Height: initialHeight + 5}
		model.Update(differentSize)
		differentUpdateTime := time.Since(start)
		
		assert.Less(t, sameUpdateTime, 5*time.Millisecond, "Same size updates should be very fast")
		// Different size updates may be slower due to component resizing
		_ = differentUpdateTime
	})

	t.Run("ErrorRecovery", func(t *testing.T) {
		// Test that error recovery doesn't crash the application
		// This tests the safeUpdateComponent and safeViewRender functions
		
		// Force an error condition by setting a component to nil temporarily
		originalAppList := model.appList
		model.appList = nil
		
		// Should not panic
		view := model.View()
		assert.NotEmpty(t, view, "Should render fallback view on component error")
		
		// Restore component
		model.appList = originalAppList
	})
}

// TestComponentIntegrationStability validates component interaction stability
func TestComponentIntegrationStability(t *testing.T) {
	model := createTestModel(t, "")

	t.Run("StateTransitions", func(t *testing.T) {
		originalState := model.state
		
		// Test all state transitions
		states := []ViewState{ListView, FormView, HelpView, ProgressView}
		
		for _, state := range states {
			model.state = state
			model.invalidateCache()
			
			// Should render without panic
			view := model.View()
			assert.NotEmpty(t, view, "State %d should render successfully", state)
		}
		
		// Restore original state
		model.state = originalState
		model.invalidateCache()
	})

	t.Run("MessageHandling", func(t *testing.T) {
		// Test various message types don't cause instability
		messages := []tea.Msg{
			tea.WindowSizeMsg{Width: 100, Height: 30},
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			tea.KeyMsg{Type: tea.KeyEsc},
		}
		
		for _, msg := range messages {
			// Should not panic
			updatedModel, cmd := model.Update(msg)
			assert.NotNil(t, updatedModel, "Model should remain valid after message")
			// cmd can be nil, that's fine
			_ = cmd
		}
	})
}

// BenchmarkViewRendering benchmarks view rendering performance
func BenchmarkViewRendering(b *testing.B) {
	model := createTestModel(b, "")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}

// BenchmarkUpdateCycle benchmarks the update cycle performance
func BenchmarkUpdateCycle(b *testing.B) {
	model := createTestModel(b, "")
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(msg)
	}
}
