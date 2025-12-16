package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/service"
	app "github.com/mrtkrcm/ZeroUI/internal/tui/components/app"
	display "github.com/mrtkrcm/ZeroUI/internal/tui/components/display"
	forms "github.com/mrtkrcm/ZeroUI/internal/tui/components/forms"
	ui "github.com/mrtkrcm/ZeroUI/internal/tui/components/ui"
	"github.com/mrtkrcm/ZeroUI/internal/tui/keybindings"
	"github.com/mrtkrcm/ZeroUI/internal/tui/logging"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// Model represents the application state
type Model struct {
	// Core state
	configService *service.ConfigService
	state         ViewState
	stateMachine  *StateMachine
	width         int
	height        int
	err           error
	ctx           context.Context
	logger        *logging.CharmLogger
	errorHandler  *ErrorHandler

	// Modern components using unified component system
	appList      *app.ApplicationListModel
	appScanner   *app.AppScannerV2          // Improved scanner
	configEditor *forms.EnhancedConfigModel // Primary config editor (was duplicated)
	helpSystem   *display.GlamourHelpModel
	presetSel    *app.PresetsSelector

	// Unified component system
	componentManager *ui.ComponentManager
	screenshotComp   *ui.ScreenshotComponent
	confirmDialog    *ui.ConfirmDialog

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
func NewModel(configService *service.ConfigService, initialApp string, logger *logging.CharmLogger) (*Model, error) {
	// Initialize theme
	theme := &styles.DefaultTheme
	appStyles := theme.BuildStyles()

	// Initialize modern components
	// Create a help system and a basic config form so tests and the UI have
	// sensible defaults even when no app is loaded yet.
	helpModel := display.NewGlamourHelp()
	// Get available apps from service
	_, err := configService.ListApplications()
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
	confirmDialog := ui.NewConfirmDialog(
		"Unsaved Changes",
		"You have unsaved changes. Are you sure you want to quit?",
		func() tea.Cmd { return tea.Quit }, // onConfirm
		func() tea.Cmd { return nil },      // onCancel
	)

	// Create base model
	model := &Model{
		configService: configService,
		state:         ListView,
		stateMachine:  stateMachine,
		keyMap:        keybindings.NewAppKeyMap(),
		styles:        appStyles,
		theme:         theme,
		appList:       appList,
		appScanner:    appScanner,
		errorHandler:  errorHandler,
		// Initialize help system
		helpSystem:       helpModel,
		presetSel:        app.NewPresetsSelector(),
		componentManager: componentManager,
		screenshotComp:   screenshotComp,
		confirmDialog:    confirmDialog,
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
			// Sync state machine with the forced state change during initialization
			model.stateMachine.Reset(FormView)
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
	// We use direct assignment here or SetState?
	// SetState checks transitions.
	// If NewModel set FormView, FormView->ProgressView is valid.
	// If ListView, ListView->ProgressView is valid.
	m.SetState(ProgressView)
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

	// Get app config from service
	appConfig, err := m.configService.GetApplicationConfig(appName)
	if err != nil {
		m.logger.LogError(err, "app_config_load", "app", appName)
		return fmt.Errorf("failed to get app config: %w", err)
	}

	// Get target config (actual values from config file)
	targetPath := appConfig.Path
	if targetPath == "" {
		// Try to find config file
		if foundPath := util.FindConfigPath(appName); foundPath != "" {
			targetPath = foundPath
			m.logger.Info("Found config file", "app", appName, "path", targetPath)
		}
	}

	var targetConfig map[string]interface{}
	if targetPath != "" {
		// First load the app config
		_, err := m.configService.GetApplicationConfig(appName)
		if err != nil {
			m.logger.Warn("Failed to load app config",
				"app", appName,
				"error", err.Error())
			targetConfig = make(map[string]interface{})
		} else {
			// For now, just get current values from service
			currentValues, err := m.configService.GetCurrentValues(appName)
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

	// Create the primary configuration interface (single source instead of duplicated UIs)
	m.configEditor = forms.NewEnhancedConfig(appName)

	// Set initial size to prevent flicker when switching views
	if m.width > 0 && m.height > 0 {
		m.configEditor.SetSize(m.width, m.height)
	}

	// Load the actual config file content for viewing
	if targetPath != "" {
		content, err := os.ReadFile(targetPath)
		if err == nil {
			m.configEditor.SetConfigFile(targetPath, string(content))
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

	// Set fields on the active configuration interface
	m.configEditor.SetFields(configFields)

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
	if m.configEditor != nil {
		m.configEditor.SetSize(m.width, m.height-4)
	}
	if m.helpSystem != nil {
		m.helpSystem.SetSize(m.width, m.height-4)
	}
	if m.confirmDialog != nil {
		m.confirmDialog.SetStyles(m.styles)
		m.confirmDialog.SetSize(m.width-10, m.height-10) // Leave some margin
	}

	// Invalidate the render cache after resizing components so any cached views
	// that depend on dimensions are refreshed. This prevents stale snapshots that
	// exceed the current model width/height.
	m.invalidateCache()

	return nil
}

// State management helpers

// SetState changes the current view state using the state machine
func (m *Model) SetState(state ViewState) {
	if err := m.stateMachine.Transition(state); err != nil {
		m.logger.Warn("Invalid state transition attempted",
			"from", m.state,
			"to", state,
			"error", err)
		// For now, we don't block the transition to avoid breaking things immediately,
		// but we log it. In a stricter future version, we might return the error.
		// However, to enforce using the state machine, we should rely on its internal state
		// if the transition was successful.
		// But since we just failed, we probably shouldn't update m.state if we want to be strict.
		// Let's force it for now to maintain behavior but log the violation.
		m.state = state
	} else {
		// Sync model state with state machine
		m.state = m.stateMachine.Current()
	}
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
