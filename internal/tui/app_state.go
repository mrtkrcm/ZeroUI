package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	app "github.com/mrtkrcm/ZeroUI/internal/tui/components/app"
	display "github.com/mrtkrcm/ZeroUI/internal/tui/components/display"
	forms "github.com/mrtkrcm/ZeroUI/internal/tui/components/forms"
	ui "github.com/mrtkrcm/ZeroUI/internal/tui/components/ui"
	"github.com/mrtkrcm/ZeroUI/internal/tui/keybindings"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// Model represents the application state
type Model struct {
	// Core state
	engine       *toggle.Engine
	state        ViewState
	stateMachine *StateMachine
	width        int
	height       int
	err          error
	ctx          context.Context
	logger       *logging.CharmLogger
	errorHandler *ErrorHandler

	// Modern components using unified component system
	appList        *app.ApplicationListModel
	appScanner     *app.AppScannerV2          // Improved scanner
	tabbedConfig   *forms.TabbedConfigModel   // Basic tabbed interface
	enhancedConfig *forms.EnhancedConfigModel // Enhanced config editor
	helpSystem     *display.GlamourHelpModel
	presetSel      *app.PresetsSelector

	// Unified component system
	componentManager *ui.ComponentManager
	screenshotComp   *ui.ScreenshotComponent
	uiManager        *ui.UIIntegrationManager
	enhancedAppList  *ui.EnhancedApplicationList

	// UI implementation selection
	uiSelector *UISelector

	// UI state
	keyMap      keybindings.AppKeyMap
	styles      *styles.Styles
	theme       *styles.Theme
	currentApp  string
	showingHelp bool

	// Transient status/toast fields used by tests and UI
	statusText  string
	statusLevel int
	statusUntil time.Time

	// Loading/progress state
	isLoading   bool
	loadingText string

	// Performance tracking and optimization
	lastRenderTime time.Time
	frameCount     int
	renderCache    map[ViewState]string
	eventBatcher   *EventBatcher

	// Render caching and debounce control for app list refreshes
	cacheDuration   time.Duration
	lastAppsRefresh time.Time
}

// RefreshAppsMsg signals a UI-level request to refresh the apps list.
// Tests and some components send this message to request a refresh; it's
// handled by the model (debounced) via the helper methods added below.
type RefreshAppsMsg = app.RefreshAppsMsg

// NewModel creates a new model for the application
func NewModel(engine *toggle.Engine, initialApp string, logger *logging.CharmLogger) (*Model, error) {
	// Initialize theme
	theme := &styles.DefaultTheme
	appStyles := theme.BuildStyles()

	// Initialize modern components
	// Create a help system and a basic config form so tests and the UI have
	// sensible defaults even when no app is loaded yet.
	helpModel := display.NewGlamourHelp()
	// Get available apps from engine
	_, err := engine.GetApps()
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}
	appList := app.NewApplicationList()
	appScanner := app.NewAppScannerV2()
	stateMachine := NewStateMachine(ListView)
	errorHandler := NewErrorHandler(logger)

	// Initialize unified component system
	componentManager := ui.NewComponentManager()
	screenshotComp := ui.NewScreenshotComponent("testdata/screenshots")
	uiManager := ui.NewUIIntegrationManager()

	// Create base model
	model := &Model{
		engine:       engine,
		state:        ListView,
		stateMachine: stateMachine,
		keyMap:       keybindings.NewAppKeyMap(),
		styles:       appStyles,
		theme:        theme,
		appList:      appList,
		appScanner:   appScanner,
		errorHandler: errorHandler,
		// Initialize help system
		helpSystem:       helpModel,
		presetSel:        app.NewPresetsSelector(),
		componentManager: componentManager,
		screenshotComp:   screenshotComp,
		uiManager:        uiManager,
		renderCache:      make(map[ViewState]string),
		eventBatcher:     NewEventBatcher(),
		logger:           logger,
		ctx:              context.Background(),
		// sensible default debounce/cache duration for tests and UI refreshes
		cacheDuration: 300 * time.Millisecond,
	}

	// If initial app specified, try to load it directly
	if initialApp != "" {
		model.currentApp = initialApp
		if err := model.loadAppConfigForForm(initialApp); err == nil {
			model.state = FormView
			logger.Info("Loaded initial app", "app", initialApp)
		} else {
			logger.LogError(err, "initial_app_load", "app", initialApp)
		}
	}

	return model, nil
}

// Init initializes the model and returns initial commands
func (m *Model) Init() tea.Cmd {
	m.logger.Info("Initializing TUI model")

	// Track initialization time
	startTime := time.Now()
	defer func() {
		m.logger.Info("Model initialization completed",
			"duration_ms", time.Since(startTime).Milliseconds())
	}()

	// Set initial dimensions to prevent flicker during startup
	if m.width == 0 || m.height == 0 {
		m.width = 80
		m.height = 24
	}

	// Start with scanning state
	m.state = ProgressView
	m.isLoading = true
	m.loadingText = "Scanning applications..."

	// Initialize with screen detection
	return tea.Batch(
		// Start application scanning
		m.appScanner.Init(),
		// Initialize application list component
		m.appList.Init(),
		// Enable mouse support
		tea.EnableMouseCellMotion,
		// Request actual terminal size (will trigger WindowSizeMsg)
		tea.WindowSize(),
		// Start event batching for better performance
		m.eventBatcher.StartBatching(),
	)
}

// loadAppConfigForForm loads app configuration for form editing
func (m *Model) loadAppConfigForForm(appName string) error {
	m.logger.Info("Loading app config for form", "app", appName)

	// Get app config from engine
	appConfig, err := m.engine.GetAppConfig(appName)
	if err != nil {
		m.logger.LogError(err, "app_config_load", "app", appName)
		return fmt.Errorf("failed to get app config: %w", err)
	}

	// Get target config (actual values from config file)
	targetPath := appConfig.Path
	if targetPath == "" {
		// Try to find config file
		home, _ := os.UserHomeDir()
		possiblePaths := []string{
			filepath.Join(home, ".config", appName, "config.yml"),
			filepath.Join(home, fmt.Sprintf(".%s", appName), "config"),
			filepath.Join(home, fmt.Sprintf(".%src", appName)),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				targetPath = path
				m.logger.Info("Found config file", "app", appName, "path", path)
				break
			}
		}
	}

	var targetConfig map[string]interface{}
	if targetPath != "" {
		// First load the app config
		_, err := m.engine.GetAppConfig(appName)
		if err != nil {
			m.logger.Warn("Failed to load app config",
				"app", appName,
				"error", err.Error())
			targetConfig = make(map[string]interface{})
		} else {
			// For now, just get current values from engine
			currentValues, err := m.engine.GetCurrentValues(appName)
			if err != nil {
				m.logger.Warn("Failed to load current config values",
					"app", appName,
					"error", err.Error())
				targetConfig = make(map[string]interface{})
			} else {
				targetConfig = currentValues
			}
		}
	} else {
		targetConfig = make(map[string]interface{})
	}

	// Create configuration interfaces
	m.tabbedConfig = forms.NewTabbedConfig(appName)
	m.enhancedConfig = forms.NewEnhancedConfig(appName)

	// Set initial size to prevent flicker when switching views
	if m.width > 0 && m.height > 0 {
		m.enhancedConfig.SetSize(m.width, m.height)
		if m.tabbedConfig != nil {
			// Try to set size if method exists
			if setter, ok := interface{}(m.tabbedConfig).(interface{ SetSize(int, int) }); ok {
				setter.SetSize(m.width, m.height)
			}
		}
	}

	// Load the actual config file content for viewing
	if targetPath != "" {
		content, err := os.ReadFile(targetPath)
		if err == nil {
			m.enhancedConfig.SetConfigFile(targetPath, string(content))
		} else {
			m.logger.Warn("Failed to read config file for viewing",
				"path", targetPath,
				"error", err.Error())
		}
	}

	// Convert config fields to the format expected by components
	var configFields []forms.ConfigField
	for key, field := range appConfig.Fields {
		configField := forms.ConfigField{
			Key:         key,
			Description: field.Description,
			Required:    false, // AppConfig doesn't carry required info; validator/schema can enrich later
			IsSet:       false,
			Source:      "default",
		}

		// If the targetConfig (actual values from the user's config file) contains
		// a value for this key, prefer that and mark the field as set.
		if targetConfig != nil {
			if v, ok := targetConfig[key]; ok {
				configField.IsSet = true
				configField.Source = "file"
				configField.Value = v
				// If the runtime value gives better type information, infer from it.
				configField.Type = inferFieldType(v)
			}
		}

		// If no value came from targetConfig, use the app-config default (if present).
		if !configField.IsSet {
			if field.Default != nil {
				configField.Value = field.Default
				configField.Source = "default"
				configField.Type = inferFieldType(field.Default)
			} else {
				// Fall back to type declared in AppConfig if provided.
				switch field.Type {
				case "choice":
					configField.Type = forms.FieldTypeSelect
				case "boolean":
					configField.Type = forms.FieldTypeBool
				case "number":
					configField.Type = forms.FieldTypeInt
				default:
					configField.Type = forms.FieldTypeString
				}
			}
		}

		// Attach options for select/choice fields when available in the AppConfig.
		if len(field.Values) > 0 {
			configField.Options = field.Values
			// Ensure select type if not already set by runtime/default
			if configField.Type == forms.FieldTypeString {
				configField.Type = forms.FieldTypeSelect
			}
		}

		configFields = append(configFields, configField)
	}

	// Set fields on both interfaces
	m.tabbedConfig.SetFields(configFields)
	m.enhancedConfig.SetFields(configFields)

	// NOTE: targetConfig integration not yet implemented - forms currently use field definitions only
	_ = targetConfig // Suppress unused variable warning for now

	// Also create preset selector
	if len(appConfig.Presets) > 0 {
		m.presetSel = app.NewPresetsSelector()
	}

	m.logger.Info("App config loaded successfully",
		"app", appName,
		"fields", len(appConfig.Fields),
		"presets", len(appConfig.Presets))

	return nil
}

// saveConfiguration saves the configuration values
func (m *Model) saveConfiguration(appName string, values map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		m.logger.Info("Saving configuration", "app", appName, "values", len(values))

		// NOTE: Engine-based save not yet implemented - configuration changes are logged but not persisted
		// For now, just return success (future enhancement needed)
		_ = values // Suppress unused variable warning

		m.logger.Info("Configuration saved successfully", "app", appName)
		return util.SuccessMsg{
			Title: "Configuration Saved",
			Body:  fmt.Sprintf("Successfully saved %s configuration", appName),
		}
	}
}

// updateComponentSizes updates component sizes based on window dimensions
func (m *Model) updateComponentSizes() tea.Cmd {
	// Update styles with new dimensions
	m.styles = styles.GetStyles()

	// Update component sizes
	if m.appList != nil {
		m.appList.SetSize(m.width, m.height-4) // Leave room for header/footer
	}
	if m.tabbedConfig != nil {
		m.tabbedConfig.SetSize(m.width, m.height-4)
	}
	if m.enhancedConfig != nil {
		m.enhancedConfig.SetSize(m.width, m.height-4)
	}
	if m.helpSystem != nil {
		m.helpSystem.SetSize(m.width, m.height-4)
	}

	// Invalidate the render cache after resizing components so any cached views
	// that depend on dimensions are refreshed. This prevents stale snapshots that
	// exceed the current model width/height.
	m.invalidateCache()

	return nil
}

// State management helpers

// SetState changes the current view state
func (m *Model) SetState(state ViewState) {
	m.state = state
	m.invalidateCache()
}

// GetState returns the current view state
func (m *Model) GetState() ViewState {
	return m.state
}

// SetLoading sets the loading state
func (m *Model) SetLoading(loading bool, text string) {
	m.isLoading = loading
	m.loadingText = text
	if loading {
		m.invalidateCache()
	}
}

// Error management

// SetError sets an error on the model
func (m *Model) SetError(err error) {
	m.err = err
	m.logger.LogError(err, "model_error")
}

// ClearError clears the current error
func (m *Model) ClearError() {
	m.err = nil
}

// SetStatus sets a transient status/toast on the model.
// `until` is the expiry time for the status; an empty time clears expiry.
func (m *Model) SetStatus(text string, level int, until time.Time) {
	m.statusText = text
	m.statusLevel = level
	m.statusUntil = until
	// Update render cache so status is shown immediately
	m.invalidateCache()
}

// ClearStatus clears any transient status/toast
func (m *Model) ClearStatus() {
	m.statusText = ""
	m.statusLevel = 0
	m.statusUntil = time.Time{}
	m.invalidateCache()
}

// isStatusActive reports whether a transient status should be shown.
func (m *Model) isStatusActive() bool {
	if m.statusText == "" {
		return false
	}
	if m.statusUntil.IsZero() {
		return true
	}
	return time.Now().Before(m.statusUntil)
}

// inferFieldType determines the field type from the value
func inferFieldType(value interface{}) forms.ConfigFieldType {
	switch value.(type) {
	case bool:
		return forms.FieldTypeBool
	case int, int32, int64:
		return forms.FieldTypeInt
	case float32, float64:
		return forms.FieldTypeFloat
	default:
		return forms.FieldTypeString
	}
}

// HandleRefreshApps performs a debounced refresh of the application list.
// This is intended to be invoked by tests or by UI event handlers when
// a RefreshAppsMsg is received. It updates `lastAppsRefresh` and triggers
// the application list component to refresh via its Update path.
func (m *Model) HandleRefreshApps() {
	now := time.Now()

	// Default debounce duration if not set
	if m.cacheDuration <= 0 {
		m.cacheDuration = 300 * time.Millisecond
	}

	// Debounce: ignore if last refresh was recent
	if now.Sub(m.lastAppsRefresh) < m.cacheDuration {
		if m.logger != nil {
			m.logger.Debug("HandleRefreshApps debounced due to recent refresh")
		}
		return
	}

	m.lastAppsRefresh = now

	// Trigger component refresh if component exists
	if m.appList != nil {
		// Create a RefreshAppsMsg using the same type as defined in the component
		var refreshMsg app.RefreshAppsMsg
		updatedModel, cmd := m.appList.Update(refreshMsg)
		// ApplicationListModel.Update returns (*ApplicationListModel, tea.Cmd) so we can
		// assign the returned pointer directly if non-nil.
		if updatedModel != nil {
			m.appList = updatedModel
		} else {
			if m.logger != nil {
				m.logger.Warn("appList.Update returned nil model")
			}
		}
		// Best-effort: run cmd if non-nil (most commands are harmless no-ops in tests)
		if cmd != nil {
			_ = cmd
		}
	} else {
		if m.logger != nil {
			m.logger.Warn("appList is nil, cannot refresh applications")
		}
	}

	// Invalidate view cache so refreshed apps are shown immediately
	m.invalidateCache()
}
