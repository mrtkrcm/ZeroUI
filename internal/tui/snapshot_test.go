package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
)

const snapshotDir = "testdata/snapshots"

// TestSnapshotAppGridView captures the app grid view
func TestSnapshotAppGridView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Set standard dimensions
	model.width = 120
	model.height = 40
	model.updateComponentSizes()
	
	// Capture snapshot
	snapshot := model.View()
	saveSnapshot(t, "app_grid_view.txt", snapshot)
	
	// Validate structure
	assert.Contains(t, snapshot, "███████╗", "Should contain ASCII logo")
	assert.Contains(t, snapshot, "applications available", "Should show app count")
	assert.Contains(t, snapshot, "Ghostty", "Should show Ghostty app")
	assert.Contains(t, snapshot, "Navigate", "Should show navigation help")
}

// TestSnapshotAppSelectionView captures the app selection view
func TestSnapshotAppSelectionView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Switch to app selection view
	model.state = AppSelectionView
	model.width = 120
	model.height = 40
	model.updateComponentSizes()
	model.focusCurrentComponent()
	
	// Capture snapshot
	snapshot := model.View()
	saveSnapshot(t, "app_selection_view.txt", snapshot)
	
	// Validate structure
	assert.Contains(t, snapshot, "Select Application", "Should show selection title")
}

// TestSnapshotConfigEditView captures the config editor view
func TestSnapshotConfigEditView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Set up config editor with test data
	model.state = ConfigEditView
	model.currentApp = "ghostty"
	model.width = 120
	model.height = 40
	
	// Load app config
	err = model.loadAppConfig("ghostty")
	if err != nil {
		t.Logf("Could not load ghostty config: %v", err)
		// Create a mock config editor for testing
		model.configEditor = components.NewConfigEditor("ghostty")
	}
	
	model.updateComponentSizes()
	model.focusCurrentComponent()
	
	// Capture snapshot
	snapshot := model.View()
	saveSnapshot(t, "config_edit_view.txt", snapshot)
	
	// Validate structure
	assert.NotEmpty(t, snapshot, "Config view should not be empty")
}

// TestSnapshotHelpView captures the help view
func TestSnapshotHelpView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Enable help
	model.showingHelp = true
	model.width = 120
	model.height = 40
	model.updateComponentSizes()
	
	// Capture snapshot
	snapshot := model.View()
	saveSnapshot(t, "help_view.txt", snapshot)
	
	// Validate structure
	assert.NotEmpty(t, snapshot, "Help view should not be empty")
}

// TestSnapshotErrorView captures the error display
func TestSnapshotErrorView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Set an error
	model.err = fmt.Errorf("test error: unable to load configuration")
	model.width = 120
	model.height = 40
	
	// Capture snapshot
	snapshot := model.View()
	saveSnapshot(t, "error_view.txt", snapshot)
	
	// Validate structure
	assert.Contains(t, snapshot, "Error", "Should display error heading")
	assert.Contains(t, snapshot, "test error", "Should display error message")
}

// TestSnapshotResponsiveSizes tests different terminal sizes
func TestSnapshotResponsiveSizes(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 80, 24},
		{"medium", 100, 30},
		{"large", 120, 40},
		{"wide", 160, 50},
	}
	
	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			model, err := NewModel(engine, "")
			require.NoError(t, err)
			
			// Set size
			model.width = size.width
			model.height = size.height
			resizeMsg := tea.WindowSizeMsg{Width: size.width, Height: size.height}
			updatedModel, _ := model.Update(resizeMsg)
			model = updatedModel.(*Model)
			
			// Capture snapshot
			snapshot := model.View()
			filename := fmt.Sprintf("responsive_%s_%dx%d.txt", size.name, size.width, size.height)
			saveSnapshot(t, filename, snapshot)
			
			// Validate it fits within bounds
			lines := strings.Split(snapshot, "\n")
			assert.LessOrEqual(t, len(lines), size.height, "Should not exceed height")
			
			for i, line := range lines {
				cleanLine := stripAnsiCodes(line)
				if len(cleanLine) > size.width {
					t.Errorf("Line %d exceeds width %d: %d chars", i, size.width, len(cleanLine))
				}
			}
		})
	}
}

// TestSnapshotComponentStates tests different component states
func TestSnapshotComponentStates(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Standard size
	model.width = 120
	model.height = 40
	
	// Test different states
	states := []struct {
		name  string
		setup func(*Model)
	}{
		{
			name: "initial_state",
			setup: func(m *Model) {
				// Default state
			},
		},
		{
			name: "help_overlay",
			setup: func(m *Model) {
				m.showingHelp = true
			},
		},
		{
			name: "app_selected",
			setup: func(m *Model) {
				m.state = ConfigEditView
				m.currentApp = "ghostty"
			},
		},
	}
	
	for _, state := range states {
		t.Run(state.name, func(t *testing.T) {
			// Reset model
			model.showingHelp = false
			model.state = AppGridView
			model.currentApp = ""
			
			// Apply state
			state.setup(model)
			model.updateComponentSizes()
			
			// Capture snapshot
			snapshot := model.View()
			filename := fmt.Sprintf("state_%s.txt", state.name)
			saveSnapshot(t, filename, snapshot)
			
			// Basic validation
			assert.NotEmpty(t, snapshot, "View should not be empty")
		})
	}
}

// Helper function to save snapshots
func saveSnapshot(t *testing.T, filename, content string) {
	t.Helper()
	
	// Create snapshot directory if it doesn't exist
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshot directory: %v", err)
	}
	
	path := filepath.Join(snapshotDir, filename)
	
	// Save snapshot
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}
	
	t.Logf("Snapshot saved: %s (%d lines)", path, strings.Count(content, "\n"))
}

// Helper function to strip ANSI codes
func stripAnsiCodes(str string) string {
	// Remove ANSI escape sequences
	var result strings.Builder
	inEscape := false
	
	for _, ch := range str {
		if ch == '\x1b' {
			inEscape = true
		} else if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
		} else {
			result.WriteRune(ch)
		}
	}
	
	return result.String()
}

// TestCoreUIFunctionality validates core UI operations
func TestCoreUIFunctionality(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	model, err := NewModel(engine, "")
	require.NoError(t, err)
	
	// Test 1: Initialization
	t.Run("Initialization", func(t *testing.T) {
		cmd := model.Init()
		assert.NotNil(t, cmd, "Init should return a command")
		
		snapshot := model.View()
		saveSnapshot(t, "test_initialization.txt", snapshot)
	})
	
	// Test 2: Window Resize
	t.Run("WindowResize", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		newModel, cmd := model.Update(msg)
		assert.Nil(t, cmd, "Resize should not return a command")
		
		m := newModel.(*Model)
		assert.Equal(t, 100, m.width)
		assert.Equal(t, 30, m.height)
		
		snapshot := m.View()
		saveSnapshot(t, "test_window_resize.txt", snapshot)
	})
	
	// Test 3: Key Navigation
	t.Run("KeyNavigation", func(t *testing.T) {
		// Test help toggle
		helpKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		newModel, _ := model.Update(helpKey)
		m := newModel.(*Model)
		assert.True(t, m.showingHelp, "Help should be shown")
		
		snapshot := m.View()
		saveSnapshot(t, "test_help_toggle.txt", snapshot)
		
		// Toggle help off
		newModel, _ = m.Update(helpKey)
		m = newModel.(*Model)
		assert.False(t, m.showingHelp, "Help should be hidden")
	})
	
	// Test 4: State Transitions
	t.Run("StateTransitions", func(t *testing.T) {
		// AppGrid -> AppSelection
		model.state = AppSelectionView
		model.focusCurrentComponent()
		
		snapshot := model.View()
		saveSnapshot(t, "test_state_app_selection.txt", snapshot)
		
		// AppSelection -> ConfigEdit
		model.state = ConfigEditView
		model.currentApp = "ghostty"
		model.focusCurrentComponent()
		
		snapshot = model.View()
		saveSnapshot(t, "test_state_config_edit.txt", snapshot)
	})
	
	// Test 5: Component Updates
	t.Run("ComponentUpdates", func(t *testing.T) {
		// Update status bar
		model.statusBar.SetAppCount(5)
		model.statusBar.SetTheme("Dark")
		
		snapshot := model.View()
		saveSnapshot(t, "test_component_updates.txt", snapshot)
	})
}

// TestUILayoutCoverage ensures all layout variations are tested
func TestUILayoutCoverage(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)
	
	// Test each view state with proper setup
	viewStates := []struct {
		name     string
		state    ViewState
		setup    func(*Model)
		validate func(*testing.T, string)
	}{
		{
			name:  "AppGridView",
			state: AppGridView,
			setup: func(m *Model) {
				// Default setup
			},
			validate: func(t *testing.T, snapshot string) {
				assert.Contains(t, snapshot, "███", "Should contain logo")
				assert.Contains(t, snapshot, "applications", "Should show app count")
			},
		},
		{
			name:  "AppSelectionView",
			state: AppSelectionView,
			setup: func(m *Model) {
				m.appSelector.Focus()
			},
			validate: func(t *testing.T, snapshot string) {
				assert.NotEmpty(t, snapshot, "Should have content")
			},
		},
		{
			name:  "ConfigEditView",
			state: ConfigEditView,
			setup: func(m *Model) {
				m.currentApp = "test-app"
				m.configEditor.Focus()
			},
			validate: func(t *testing.T, snapshot string) {
				assert.NotEmpty(t, snapshot, "Should have content")
			},
		},
		{
			name:  "HelpView",
			state: HelpView,
			setup: func(m *Model) {
				m.showingHelp = true
			},
			validate: func(t *testing.T, snapshot string) {
				assert.NotEmpty(t, snapshot, "Should have help content")
			},
		},
	}
	
	for _, vs := range viewStates {
		t.Run(vs.name, func(t *testing.T) {
			model, err := NewModel(engine, "")
			require.NoError(t, err)
			
			// Set standard size
			model.width = 120
			model.height = 40
			model.state = vs.state
			
			// Run setup
			vs.setup(model)
			model.updateComponentSizes()
			
			// Capture snapshot
			snapshot := model.View()
			filename := fmt.Sprintf("layout_%s.txt", vs.name)
			saveSnapshot(t, filename, snapshot)
			
			// Validate
			vs.validate(t, snapshot)
		})
	}
}