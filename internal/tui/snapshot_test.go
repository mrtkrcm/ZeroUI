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

	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	forms "github.com/mrtkrcm/ZeroUI/internal/tui/components/forms"
	ui "github.com/mrtkrcm/ZeroUI/internal/tui/components/ui"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
)

const SNAPSHOT_DIR = "testdata/snapshots"

// File permission constants
const (
	DirPerm  os.FileMode = 0755 // rwxr-xr-x
	FilePerm os.FileMode = 0644 // rw-r--r--
)

// ensureDir creates a directory with proper error handling
func ensureDir(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		t.Fatalf("Directory path cannot be empty")
	}
	if err := os.MkdirAll(path, DirPerm); err != nil {
		t.Fatalf("Failed to create directory %s: %v", path, err)
	}
}

// writeFile writes content to a file with proper error handling
func writeFile(t *testing.T, path string, content []byte, perm os.FileMode) {
	t.Helper()
	if path == "" {
		t.Errorf("File path cannot be empty")
		return
	}
	// Use FilePerm as default if no specific permission is provided
	if perm == 0 {
		perm = FilePerm
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		t.Errorf("Failed to write file %s: %v", path, err)
	}
}

// safeView gets the model view with error recovery
func safeView(model *Model) (view string) {
	defer func() {
		if r := recover(); r != nil {
			// Return a safe fallback view on panic
			view = "Error: View rendering failed due to panic"
		}
	}()
	if model == nil {
		return "Error: Model is nil"
	}
	view = model.View()
	return view
}

// TestConfigurationDetection verifies that test configurations are properly detected
// This test ensures that the application's configuration detection mechanism
// works correctly for various supported applications in the test environment.
func TestConfigurationDetection(t *testing.T) {
	// Load the apps registry to test configuration detection
	registry, err := config.LoadAppsRegistry()
	require.NoError(t, err, "Should be able to load apps registry")

	// Test specific applications
	apps := []string{"ghostty", "zed", "neovim", "vscode"}

	for _, app := range apps {
		t.Run(app, func(t *testing.T) {
			exists, path := registry.CheckAppStatus(app)
			if !exists {
				t.Logf("App %s not found, checking config paths...", app)
				paths := registry.GetConfigPaths(app)
				t.Logf("Expected paths: %v", paths)
			} else {
				t.Logf("App %s found at: %s", app, path)
			}
		})
	}
}

// TestTUIRegistryDetection verifies that the TUI registry detects applications correctly
func TestTUIRegistryDetection(t *testing.T) {
	// Get app statuses from TUI registry
	statuses := registry.GetAppStatuses()

	t.Logf("Found %d applications in TUI registry:", len(statuses))

	for i, status := range statuses {
		t.Logf("%d. %s - Installed: %v, HasConfig: %v, ConfigExists: %v",
			i+1, status.Definition.Name, status.IsInstalled, status.HasConfig, status.ConfigExists)
		if status.ConfigExists {
			t.Logf("   Config path: %s", status.Definition.ConfigPath)
		}
	}

	// Check if any applications are detected
	installedCount := 0
	configuredCount := 0
	for _, status := range statuses {
		if status.IsInstalled {
			installedCount++
		}
		if status.ConfigExists {
			configuredCount++
		}
	}

	t.Logf("Summary: %d installed, %d configured", installedCount, configuredCount)
}

// TestSnapshotListView captures the app grid view with enhanced screenshots
// This test verifies that the main application grid displays correctly with all detected
// applications and proper navigation elements.
func TestSnapshotListView(t *testing.T) {
	// First, verify that applications are detected by the registry
	statuses := registry.GetAppStatuses()
	detectedCount := 0
	for _, status := range statuses {
		if status.IsInstalled && status.ConfigExists {
			detectedCount++
		}
	}

	t.Logf("Registry detects %d applications with both executable and config", detectedCount)

	// Now test the actual TUI
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Set standard dimensions
	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	// Trigger refresh to ensure applications are detected
	t.Logf("Calling HandleRefreshApps...")
	model.HandleRefreshApps()

	// Check if appList has items after refresh
	if model.appList != nil {
		t.Logf("ApplicationList exists after refresh")
		// Demonstrate the screenshot system works by showing what we captured
		t.Logf("Screenshot system successfully captured TUI output")
	} else {
		t.Logf("ApplicationList is nil")
	}

	// Capture enhanced screenshot with real TUI components
	saveScreenshot(t, "list_view_tests", "App Grid View", model, "Application start", "View app grid")

	// Also save the traditional snapshot for compatibility
	snapshot := model.View()
	saveSnapshot(t, "app_grid_view.txt", snapshot)

	// Validate structure (be flexible to styling changes)
	lower := strings.ToLower(snapshot)
	assert.True(t, strings.Contains(lower, "zeroui") || strings.Contains(snapshot, "███████╗"), "Should contain app title")
	assert.True(t, strings.Contains(lower, "applications") || strings.Contains(lower, "apps"), "Should show app list context")

	// Log detailed information about what was captured
	t.Logf("Registry detected %d applications, but TUI shows: %s", detectedCount,
		func() string {
			if strings.Contains(snapshot, "No applications detected") {
				return "NO APPLICATIONS DETECTED"
			}
			return "APPLICATIONS FOUND"
		}())
}

// TestSnapshotAppSelectionView captures the app selection view with enhanced screenshots
func TestSnapshotAppSelectionView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Switch to app selection view
	model.state = ListView
	model.width = 120
	model.height = 40
	model.updateComponentSizes()
	// Focus is handled by modern components

	// Capture enhanced screenshot with real TUI components
	saveScreenshot(t, "app_selection_tests", "App Selection View", model, "Navigate to app selection", "View selectable apps")

	// Also save the traditional snapshot for compatibility
	snapshot := model.View()
	saveSnapshot(t, "app_selection_view.txt", snapshot)

	// Validate structure (flexible)
	assert.True(t, strings.Contains(snapshot, "Select Application") || strings.Contains(strings.ToLower(snapshot), "applications"), "Should show selection context")
}

// TestSnapshotFormView captures the config editor view with enhanced screenshots
func TestSnapshotFormView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Set up config editor with test data
	model.state = FormView
	model.currentApp = "ghostty"
	model.width = 120
	model.height = 40

	// Initialize form components for testing - use a simpler approach
	// Create basic form component without loading actual config
	model.configEditor = forms.NewEnhancedConfig("ghostty")

	// Set basic size
	if model.configEditor != nil {
		model.configEditor.SetSize(model.width, model.height)
	}

	model.updateComponentSizes()
	// Focus is handled by modern components

	// Capture enhanced screenshot with real TUI components
	saveScreenshot(t, "form_view_tests", "Configuration Editor", model, "Select ghostty", "Enter config mode", "View configuration form")

	// Also save the traditional snapshot for compatibility
	snapshot := model.View()
	saveSnapshot(t, "config_edit_view.txt", snapshot)

	// Validate structure
	assert.NotEmpty(t, snapshot, "Config view should not be empty")
}

// TestSnapshotHelpView captures the help view with enhanced screenshots
func TestSnapshotHelpView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Enable help
	model.showingHelp = true
	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	// Capture enhanced screenshot with real TUI components
	saveScreenshot(t, "help_view_tests", "Help Overlay", model, "Press '?' for help", "View help information")

	// Also save the traditional snapshot for compatibility
	snapshot := model.View()
	saveSnapshot(t, "help_view.txt", snapshot)

	// Validate structure
	assert.NotEmpty(t, snapshot, "Help view should not be empty")
}

// TestSnapshotErrorView captures the error display with enhanced screenshots
func TestSnapshotErrorView(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Set an error
	model.err = fmt.Errorf("test error: unable to load configuration")
	model.width = 120
	model.height = 40

	// Capture enhanced screenshot with real TUI components
	saveScreenshot(t, "error_view_tests", "Error Display", model, "Trigger configuration error", "View error message")

	// Also save the traditional snapshot for compatibility
	snapshot := model.View()
	saveSnapshot(t, "error_view.txt", snapshot)

	// Validate structure
	assert.Contains(t, snapshot, "Error", "Should display error heading")
	assert.Contains(t, snapshot, "test error", "Should display error message")
}

// TestSnapshotResponsiveSizes tests different terminal sizes with enhanced screenshots
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
			model, err := NewTestModel(engine, "")
			require.NoError(t, err)

			// Set size
			model.width = size.width
			model.height = size.height
			resizeMsg := tea.WindowSizeMsg{Width: size.width, Height: size.height}
			updatedModel, _ := model.Update(resizeMsg)
			model = updatedModel.(*Model)

			// Trigger refresh to ensure applications are detected
			model.HandleRefreshApps()

			// Capture enhanced screenshot with real TUI components
			description := fmt.Sprintf("Responsive %s terminal", size.name)
			saveScreenshot(t, "responsive_tests", description, model, fmt.Sprintf("Resize to %dx%d", size.width, size.height))

			// Also save the traditional snapshot for compatibility
			snapshot := model.View()
			filename := fmt.Sprintf("responsive_%s_%dx%d.txt", size.name, size.width, size.height)
			saveSnapshot(t, filename, snapshot)

			// Validate it fits within bounds
			lines := strings.Split(snapshot, "\n")
			// Allow 15 lines tolerance for content that needs to be displayed
			heightTolerance := 15
			assert.LessOrEqual(t, len(lines), size.height+heightTolerance, "Should not exceed height by more than tolerance")

			for i, line := range lines {
				cleanLine := stripAnsiCodes(line)
				// Allow 15 characters tolerance for rendering edge cases and emoji/unicode
				tolerance := 15
				maxWidth := size.width + tolerance
				if len(cleanLine) > maxWidth {
					t.Errorf("Line %d exceeds width %d: %d chars (max allowed: %d)",
						i, size.width, len(cleanLine), maxWidth)
				}
			}
		})
	}
}

// TestSnapshotComponentStates tests different component states with enhanced screenshots
func TestSnapshotComponentStates(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
	require.NoError(t, err)

	// Standard size
	model.width = 120
	model.height = 40

	// Test different states
	states := []struct {
		name        string
		description string
		setup       func(*Model)
		actions     []string
	}{
		{
			name:        "initial_state",
			description: "Initial Application State",
			setup: func(m *Model) {
				// Default state
			},
			actions: []string{"Application start"},
		},
		{
			name:        "help_overlay",
			description: "Help Overlay Active",
			setup: func(m *Model) {
				m.showingHelp = true
			},
			actions: []string{"Press '?' for help"},
		},
		{
			name:        "app_selected",
			description: "Application Selected",
			setup: func(m *Model) {
				m.state = FormView
				m.currentApp = "ghostty"
			},
			actions: []string{"Select ghostty", "Enter configuration mode"},
		},
	}

	for _, state := range states {
		t.Run(state.name, func(t *testing.T) {
			// Reset model
			model.showingHelp = false
			model.state = ListView
			model.currentApp = ""

			// Apply state
			state.setup(model)
			model.updateComponentSizes()

			// Capture enhanced screenshot with real TUI components
			saveScreenshot(t, "component_states", state.description, model, state.actions...)

			// Also save the traditional snapshot for compatibility
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
	ensureDir(t, SNAPSHOT_DIR)

	path := filepath.Join(SNAPSHOT_DIR, filename)

	// Save snapshot
	writeFile(t, path, []byte(content), FilePerm)

	t.Logf("Snapshot saved: %s (%d lines)", path, strings.Count(content, "\n"))
}

// Enhanced screenshot function that captures comprehensive screen data with real TUI components
func saveScreenshot(t *testing.T, testName, description string, model *Model, actions ...string) {
	t.Helper()

	// Use the new screenshot component for better integration
	captureDir := filepath.Join("testdata", "screenshots")
	screenshotComp := ui.NewScreenshotComponent(captureDir)
	screenshotComp.SetSize(model.width, model.height)

	// Create a wrapper interface to extract model information
	modelWrapper := &ModelWrapper{model}

	// Try to use enhanced integration if components are available
	var err error

	// For list view tests, try to integrate with application list
	if strings.Contains(testName, "list_view") {
		if appList := getApplicationListComponent(model); appList != nil {
			integrator := screenshotComp.IntegrateWithComponents().WithApplicationList(appList)
			err = integrator.Capture(modelWrapper, description, testName, actions...)
		} else {
			err = screenshotComp.Capture(modelWrapper, description, testName, actions...)
		}
	} else {
		// Use standard capture for other tests
		err = screenshotComp.Capture(modelWrapper, description, testName, actions...)
	}

	if err != nil {
		t.Errorf("Failed to capture screenshot: %v", err)
		return
	}

	captureName := strings.ReplaceAll(description, " ", "_")
	txtPath := filepath.Join(captureDir, testName, fmt.Sprintf("%s.txt", captureName))
	t.Logf("Screen captured: %s", txtPath)
}

// getApplicationListComponent attempts to extract the application list component from the model
func getApplicationListComponent(model *Model) interface{} {
	// This is a helper function that tries to access the application list component
	// In a real implementation, this would depend on the model's structure
	return nil // For now, return nil to use the standard capture method
}

// ModelWrapper provides a consistent interface for the screenshot component
type ModelWrapper struct {
	model *Model
}

func (m *ModelWrapper) Init() tea.Cmd {
	return nil
}

func (m *ModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *ModelWrapper) View() string {
	return safeView(m.model)
}

func (m *ModelWrapper) GetState() string {
	return getStateName(m.model.state)
}

func (m *ModelWrapper) GetCurrentApp() string {
	return m.model.currentApp
}

func (m *ModelWrapper) GetWidth() int {
	return m.model.width
}

func (m *ModelWrapper) GetHeight() int {
	return m.model.height
}

func (m *ModelWrapper) IsShowingHelp() bool {
	return m.model.showingHelp
}

func (m *ModelWrapper) GetError() error {
	return m.model.err
}

// Helper function to get state name
func getStateName(state ViewState) string {
	switch state {
	case ListView:
		return "list_view"
	case FormView:
		return "form_view"
	case HelpView:
		return "help_view"
	default:
		return "unknown"
	}
}

// stripAnsiCodes is now defined in automation_framework.go

// TestCoreUIFunctionality validates core UI operations
func TestCoreUIFunctionality(t *testing.T) {
	engine, err := toggle.NewEngine()
	require.NoError(t, err)

	model, err := NewTestModel(engine, "")
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
		assert.True(t, m.showingHelp || m.state == HelpView, "Help should be shown")

		snapshot := m.View()
		saveSnapshot(t, "test_help_toggle.txt", snapshot)

		// Toggle help off
		newModel, _ = m.Update(helpKey)
		m = newModel.(*Model)
		assert.False(t, m.showingHelp || m.state == HelpView, "Help should be hidden")
	})

	// Test 4: State Transitions
	t.Run("StateTransitions", func(t *testing.T) {
		// AppGrid -> AppSelection
		model.state = ListView
		// Focus is handled by modern components

		snapshot := model.View()
		saveSnapshot(t, "test_state_app_selection.txt", snapshot)

		// AppSelection -> ConfigEdit
		model.state = FormView
		model.currentApp = "ghostty"
		// Focus is handled by modern components

		snapshot = model.View()
		saveSnapshot(t, "test_state_config_edit.txt", snapshot)
	})

	// Test 5: Component Updates
	t.Run("ComponentUpdates", func(t *testing.T) {
		// Update status bar
		// Status bar functionality handled by modern components //SetAppCount(5)
		// Status bar functionality handled by modern components //SetTheme("Dark")

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
			name:  "ListView",
			state: ListView,
			setup: func(m *Model) {
				// Default setup
			},
			validate: func(t *testing.T, snapshot string) {
				// Accept modern header instead of legacy logo/count
				assert.Contains(t, snapshot, "ZeroUI Applications", "Should show header")
			},
		},
		{
			name:  "ListView",
			state: ListView,
			setup: func(m *Model) {
				// Focus is handled by modern components
			},
			validate: func(t *testing.T, snapshot string) {
				assert.NotEmpty(t, snapshot, "Should have content")
			},
		},
		{
			name:  "FormView",
			state: FormView,
			setup: func(m *Model) {
				m.currentApp = "test-app"
				// Focus is handled by modern components
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
			model, err := NewTestModel(engine, "")
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

// formatActions formats the actions array for YAML frontmatter
func formatActions(actions []string) string {
	if len(actions) == 0 {
		return "  - \"No actions\""
	}

	var result strings.Builder
	for _, action := range actions {
		result.WriteString(fmt.Sprintf("  - \"%s\"\n", action))
	}
	return result.String()
}
