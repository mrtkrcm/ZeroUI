package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

const visualTestDir = "testdata/visual"

// TestVisualRendering captures all UI states for visual validation
func TestVisualRendering(t *testing.T) {
	// Create test directory
	err := os.MkdirAll(visualTestDir, 0755)
	require.NoError(t, err)

	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	// Test cases for different screen sizes and states
	testCases := []struct {
		name     string
		width    int
		height   int
		state    ViewState
		setup    func(*Model)
		validate func(*testing.T, string)
	}{
		{
			name:   "huh_grid_large_screen",
			width:  120,
			height: 40,
			state:  HuhGridView,
			setup: func(m *Model) {
				// Default setup with 4 columns
			},
			validate: func(t *testing.T, view string) {
				assert.Contains(t, view, "Application Grid", "Should show grid title")
				assert.Contains(t, view, "4 columns", "Should show 4-column layout")
				assert.Contains(t, view, "Navigate Grid", "Should show navigation help")
			},
		},
		{
			name:   "huh_grid_medium_screen",
			width:  100,
			height: 30,
			state:  HuhGridView,
			setup: func(m *Model) {
				// Should adapt to 3 columns
			},
			validate: func(t *testing.T, view string) {
				assert.Contains(t, view, "Application Grid", "Should show grid title")
				assert.NotEmpty(t, view, "Should render content")
			},
		},
		{
			name:   "huh_grid_small_screen",
			width:  80,
			height: 24,
			state:  HuhGridView,
			setup: func(m *Model) {
				// Should adapt to 2 columns
			},
			validate: func(t *testing.T, view string) {
				assert.Contains(t, view, "Application Grid", "Should show grid title")
				assert.NotEmpty(t, view, "Should render content")
			},
		},
		{
			name:   "huh_app_selector",
			width:  100,
			height: 30,
			state:  HuhAppSelectionView,
			setup:  func(m *Model) {},
			validate: func(t *testing.T, view string) {
				assert.Contains(t, view, "Select Application", "Should show selector title")
			},
		},
		{
			name:   "huh_config_editor",
			width:  100,
			height: 30,
			state:  HuhConfigEditView,
			setup: func(m *Model) {
				m.currentApp = "ghostty"
				m.loadAppConfig("ghostty")
			},
			validate: func(t *testing.T, view string) {
				assert.NotEmpty(t, view, "Should render config editor")
			},
		},
		{
			name:   "legacy_app_grid_4_columns",
			width:  140,
			height: 40,
			state:  AppGridView,
			setup:  func(m *Model) {},
			validate: func(t *testing.T, view string) {
				assert.NotEmpty(t, view, "Should render legacy grid")
			},
		},
		{
			name:   "help_view",
			width:  100,
			height: 30,
			state:  HuhGridView,
			setup: func(m *Model) {
				m.showingHelp = true
			},
			validate: func(t *testing.T, view string) {
				assert.NotEmpty(t, view, "Should render help")
			},
		},
		{
			name:   "error_state",
			width:  100,
			height: 30,
			state:  HuhGridView,
			setup: func(m *Model) {
				m.err = fmt.Errorf("test error for visual testing")
			},
			validate: func(t *testing.T, view string) {
				assert.Contains(t, view, "Error", "Should show error")
				assert.Contains(t, view, "test error", "Should show error message")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := NewModel(engine, "")
			require.NoError(t, err)

			// Set up the test case
			model.state = tc.state
			model.width = tc.width
			model.height = tc.height
			
			// Apply custom setup
			if tc.setup != nil {
				tc.setup(model)
			}

			// Update component sizes
			model.updateComponentSizes()
			model.focusCurrentComponent()

			// Render the view
			view := model.View()

			// Save visual snapshot
			filename := fmt.Sprintf("%s_%dx%d.txt", tc.name, tc.width, tc.height)
			saveVisualSnapshot(t, filename, view)

			// Run validation
			tc.validate(t, view)

			// Basic sanity checks
			assert.NotEmpty(t, view, "View should not be empty")
			lines := countLines(view)
			assert.LessOrEqual(t, lines, tc.height+5, "Should not exceed reasonable line count") // Allow some flexibility
		})
	}
}

// TestResponsiveLayout tests that all layouts adapt correctly to different sizes
func TestResponsiveLayout(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	sizes := []struct {
		name   string
		width  int
		height int
		expectedColumns int
	}{
		{"Mobile", 60, 20, 1},
		{"Tablet", 90, 25, 2},
		{"Laptop", 120, 30, 3},
		{"Desktop", 140, 40, 4},
		{"Ultrawide", 180, 50, 4},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			model, err := NewModel(engine, "")
			require.NoError(t, err)

			// Test both modern and legacy grids
			states := []ViewState{HuhGridView, AppGridView}
			
			for _, state := range states {
				stateStr := "huh"
				if state == AppGridView {
					stateStr = "legacy"
				}
				
				t.Run(stateStr, func(t *testing.T) {
					model.state = state
					model.width = size.width
					model.height = size.height
					
					// Trigger resize
					resizeMsg := tea.WindowSizeMsg{Width: size.width, Height: size.height}
					model.Update(resizeMsg)
					
					// Render
					view := model.View()
					
					// Save snapshot
					filename := fmt.Sprintf("responsive_%s_%s_%dx%d.txt", 
						stateStr, size.name, size.width, size.height)
					saveVisualSnapshot(t, filename, view)
					
					// Validate responsive behavior
					assert.NotEmpty(t, view, "Should render at any size")
					
					// Check that it doesn't exceed screen bounds
					lines := countLines(view)
					assert.LessOrEqual(t, lines, size.height+10, "Should not far exceed screen height")
				})
			}
		})
	}
}

// TestAnimationsAndTransitions tests dynamic UI elements
func TestAnimationsAndTransitions(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewModel(engine, "")
	require.NoError(t, err)

	model.width = 120
	model.height = 40
	model.state = HuhGridView

	// Test state transitions
	transitions := []struct {
		name      string
		fromState ViewState
		toState   ViewState
		action    func(*Model)
	}{
		{
			name:      "grid_to_selector",
			fromState: HuhGridView,
			toState:   HuhAppSelectionView,
			action: func(m *Model) {
				m.state = HuhAppSelectionView
				m.focusCurrentComponent()
			},
		},
		{
			name:      "selector_to_config",
			fromState: HuhAppSelectionView,
			toState:   HuhConfigEditView,
			action: func(m *Model) {
				m.state = HuhConfigEditView
				m.currentApp = "ghostty"
				m.focusCurrentComponent()
			},
		},
		{
			name:      "modern_to_legacy",
			fromState: HuhGridView,
			toState:   AppGridView,
			action: func(m *Model) {
				m.state = AppGridView
				m.focusCurrentComponent()
			},
		},
	}

	for _, transition := range transitions {
		t.Run(transition.name, func(t *testing.T) {
			// Start state
			model.state = transition.fromState
			model.focusCurrentComponent()
			startView := model.View()
			
			// Apply transition
			transition.action(model)
			endView := model.View()
			
			// Save both states
			saveVisualSnapshot(t, fmt.Sprintf("transition_%s_start.txt", transition.name), startView)
			saveVisualSnapshot(t, fmt.Sprintf("transition_%s_end.txt", transition.name), endView)
			
			// Validate both states render
			assert.NotEmpty(t, startView, "Start state should render")
			assert.NotEmpty(t, endView, "End state should render")
			assert.NotEqual(t, startView, endView, "States should be visually different")
		})
	}
}

// TestPerformance benchmarks UI rendering performance
func TestPerformance(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewModel(engine, "")
	require.NoError(t, err)

	model.width = 120
	model.height = 40
	model.state = HuhGridView

	// Warm up
	for i := 0; i < 10; i++ {
		model.View()
	}

	// Benchmark rendering
	start := time.Now()
	iterations := 100
	
	for i := 0; i < iterations; i++ {
		view := model.View()
		assert.NotEmpty(t, view)
	}
	
	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)
	
	t.Logf("Average render time: %v", avgTime)
	t.Logf("Renders per second: %.0f", float64(time.Second)/float64(avgTime))
	
	// Performance requirements
	assert.Less(t, avgTime, 16*time.Millisecond, "Should render in <16ms for 60fps")
	
	// Save performance snapshot
	finalView := model.View()
	saveVisualSnapshot(t, "performance_final.txt", finalView)
}

// TestUIKeyboardNavigation tests navigation behavior in visual context
func TestUIKeyboardNavigation(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewModel(engine, "")
	require.NoError(t, err)

	model.width = 120
	model.height = 40
	model.state = HuhGridView
	model.focusCurrentComponent()

	// Test navigation keys
	navigationTests := []struct {
		name string
		key  tea.KeyMsg
		desc string
	}{
		{"up", tea.KeyMsg{Type: tea.KeyUp}, "Up arrow"},
		{"down", tea.KeyMsg{Type: tea.KeyDown}, "Down arrow"},
		{"left", tea.KeyMsg{Type: tea.KeyLeft}, "Left arrow"},
		{"right", tea.KeyMsg{Type: tea.KeyRight}, "Right arrow"},
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, "Enter key"},
		{"help", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, "Help toggle"},
		{"quit", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, "Quit key"},
	}

	for _, nav := range navigationTests {
		t.Run(nav.name, func(t *testing.T) {
			// Capture state before
			beforeView := model.View()
			
			// Send key
			updatedModel, cmd := model.Update(nav.key)
			model = updatedModel.(*Model)
			
			// Capture state after
			afterView := model.View()
			
			// Save snapshots
			saveVisualSnapshot(t, fmt.Sprintf("nav_%s_before.txt", nav.name), beforeView)
			saveVisualSnapshot(t, fmt.Sprintf("nav_%s_after.txt", nav.name), afterView)
			
			// Basic validation
			assert.NotEmpty(t, beforeView, "Before state should render")
			assert.NotEmpty(t, afterView, "After state should render")
			
			// Check for quit command
			if nav.name == "quit" {
				// Should return quit command
				assert.NotNil(t, cmd, "Quit should return command")
			}
		})
	}
}

// TestErrorRecovery tests error handling and recovery
func TestErrorRecovery(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewModel(engine, "")
	require.NoError(t, err)

	errorCases := []struct {
		name  string
		setup func(*Model)
		check func(*testing.T, string)
	}{
		{
			name: "config_load_error",
			setup: func(m *Model) {
				m.err = fmt.Errorf("failed to load configuration")
				m.state = HuhConfigEditView
			},
			check: func(t *testing.T, view string) {
				assert.Contains(t, view, "Error", "Should show error")
				assert.Contains(t, view, "configuration", "Should mention config error")
			},
		},
		{
			name: "app_not_found",
			setup: func(m *Model) {
				m.currentApp = "nonexistent-app"
				m.state = HuhConfigEditView
			},
			check: func(t *testing.T, view string) {
				assert.NotEmpty(t, view, "Should still render something")
			},
		},
		{
			name: "invalid_state",
			setup: func(m *Model) {
				m.state = ViewState(999) // Invalid state
			},
			check: func(t *testing.T, view string) {
				assert.NotEmpty(t, view, "Should handle invalid state gracefully")
			},
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset model
			model.err = nil
			model.state = HuhGridView
			
			// Apply error condition
			tc.setup(model)
			
			// Render and validate
			view := model.View()
			saveVisualSnapshot(t, fmt.Sprintf("error_%s.txt", tc.name), view)
			
			tc.check(t, view)
		})
	}
}

// Helper functions

func saveVisualSnapshot(t *testing.T, filename, content string) {
	t.Helper()
	
	path := filepath.Join(visualTestDir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Logf("Warning: Could not save visual snapshot %s: %v", filename, err)
	} else {
		t.Logf("Saved visual snapshot: %s (%d bytes)", path, len(content))
	}
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := 1
	for _, char := range s {
		if char == '\n' {
			count++
		}
	}
	return count
}

// TestVisualRegression compares current output with reference snapshots
func TestVisualRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual regression tests in short mode")
	}
	
	// This test would compare current output with saved reference images
	// For now, it just ensures we can generate all the visual outputs
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewModel(engine, "")
	require.NoError(t, err)

	// Generate reference snapshots for all major states
	states := []ViewState{
		HuhGridView,
		HuhAppSelectionView,
		HuhConfigEditView,
		AppGridView,
		AppSelectionView,
		ConfigEditView,
	}
	
	for _, state := range states {
		model.state = state
		model.width = 120
		model.height = 40
		model.focusCurrentComponent()
		
		view := model.View()
		assert.NotEmpty(t, view, fmt.Sprintf("State %d should render", int(state)))
		
		// Save as reference
		filename := fmt.Sprintf("reference_state_%d.txt", int(state))
		saveVisualSnapshot(t, filename, view)
	}
}