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

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/validation"
	forms "github.com/mrtkrcm/ZeroUI/internal/tui/components/forms"
)

// Only run snapshot tests in CI or when explicitly requested
var runSnapshots = os.Getenv("RUN_SNAPSHOTS") == "true" || os.Getenv("CI") == "true"

const SNAPSHOT_DIR = "testdata/snapshots"

// File permission constants
const (
	DirPerm  os.FileMode = 0o755 // rwxr-xr-x
	FilePerm os.FileMode = 0o644 // rw-r--r--
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
func TestConfigurationDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping configuration detection test in short mode")
	}

	// Load the apps registry to test configuration detection
	registry, err := appconfig.LoadAppsRegistry()
	require.NoError(t, err, "Should be able to load apps registry")

	// Test specific applications
	apps := []string{"ghostty", "zed"}

	for _, app := range apps {
		t.Run(app, func(t *testing.T) {
			exists, _ := registry.CheckAppStatus(app)
			assert.True(t, exists, "App %s should be detected", app)
		})
	}
}

// TestSnapshotListView captures the app grid view
func TestSnapshotListView(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Set standard dimensions
	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	// Trigger refresh to ensure applications are detected
	model.HandleRefreshApps()

	// Save snapshot
	snapshot := model.View()
	saveSnapshot(t, "app_grid_view.txt", snapshot)

	// Validate structure
	lower := strings.ToLower(snapshot)
	assert.True(t, strings.Contains(lower, "zeroui") || strings.Contains(snapshot, "███████╗"), "Should contain app title")
	assert.True(t, strings.Contains(lower, "applications") || strings.Contains(lower, "apps"), "Should show app list context")
}

// TestSnapshotAppSelectionView captures the app selection view
func TestSnapshotAppSelectionView(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Switch to app selection view
	model.state = ListView
	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	snapshot := model.View()
	saveSnapshot(t, "app_selection_view.txt", snapshot)

	// Validate structure
	assert.True(t, strings.Contains(snapshot, "Select Application") || strings.Contains(strings.ToLower(snapshot), "applications"), "Should show selection context")
}

// TestSnapshotFormView captures the config editor view
func TestSnapshotFormView(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Set up config editor with test data
	model.state = FormView
	model.currentApp = "ghostty"
	model.width = 120
	model.height = 40

	// Initialize form components for testing
	model.configEditor = forms.NewEnhancedConfig("ghostty")
	if model.configEditor != nil {
		model.configEditor.SetSize(model.width, model.height)
	}

	model.updateComponentSizes()

	snapshot := model.View()
	saveSnapshot(t, "config_edit_view.txt", snapshot)

	// Validate structure
	assert.NotEmpty(t, snapshot, "Config view should not be empty")
}

// TestSnapshotHelpView captures the help view
func TestSnapshotHelpView(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Enable help
	model.showingHelp = true
	model.width = 120
	model.height = 40
	model.updateComponentSizes()

	snapshot := model.View()
	saveSnapshot(t, "help_view.txt", snapshot)

	// Validate structure
	assert.NotEmpty(t, snapshot, "Help view should not be empty")
}

// TestSnapshotErrorView captures the error display
func TestSnapshotErrorView(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// Set an error
	model.err = fmt.Errorf("test error: unable to load configuration")
	model.width = 120
	model.height = 40

	snapshot := model.View()
	saveSnapshot(t, "error_view.txt", snapshot)

	// Validate structure
	assert.Contains(t, snapshot, "Error", "Should display error heading")
	assert.Contains(t, snapshot, "test error", "Should display error message")
}

// TestSnapshotResponsiveSizes tests different terminal sizes
func TestSnapshotResponsiveSizes(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 80, 24},
		{"medium", 100, 30},
		{"large", 120, 40},
	}

	for _, size := range sizes {
		t.Run(size.name, func(t *testing.T) {
			model, err := NewTestModel(configService, "")
			require.NoError(t, err)

			// Set size
			model.width = size.width
			model.height = size.height
			resizeMsg := tea.WindowSizeMsg{Width: size.width, Height: size.height}
			updatedModel, _ := model.Update(resizeMsg)
			model = updatedModel.(*Model)

			// Trigger refresh to ensure applications are detected
			model.HandleRefreshApps()

			// Save snapshot
			snapshot := model.View()
			filename := fmt.Sprintf("responsive_%s_%dx%d.txt", size.name, size.width, size.height)
			saveSnapshot(t, filename, snapshot)

			// Validate it fits within bounds
			lines := strings.Split(snapshot, "\n")
			heightTolerance := 10
			assert.LessOrEqual(t, len(lines), size.height+heightTolerance, "Should not exceed height by more than tolerance")

			for i, line := range lines {
				cleanLine := stripAnsiCodes(line)
				// Allow 20 characters tolerance for rendering edge cases and emoji/unicode
				tolerance := 20
				maxWidth := size.width + tolerance
				if len(cleanLine) > maxWidth {
					t.Errorf("Line %d exceeds width %d: %d chars (max allowed: %d)",
						i, size.width, len(cleanLine), maxWidth)
				}
			}
		})
	}
}

// TestSnapshotComponentStates tests different component states
func TestSnapshotComponentStates(t *testing.T) {
	if !runSnapshots {
		t.Skip("Skipping snapshot test - set RUN_SNAPSHOTS=true to run")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
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
				m.state = FormView
				m.currentApp = "ghostty"
			},
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

			// Save snapshot
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

	// Only log in verbose mode
	if testing.Verbose() {
		t.Logf("Snapshot saved: %s (%d lines)", path, strings.Count(content, "\n"))
	}
}

// TestCoreUIFunctionality validates core UI operations
func TestCoreUIFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping UI functionality test in short mode")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
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
}

// TestUILayoutCoverage ensures all layout variations are tested
func TestUILayoutCoverage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping layout coverage test in short mode")
	}

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

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
				assert.Contains(t, snapshot, "ZeroUI", "Should show header")
			},
		},
		{
			name:  "FormView",
			state: FormView,
			setup: func(m *Model) {
				m.currentApp = "ghostty"
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
			model, err := NewTestModel(configService, "")
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
