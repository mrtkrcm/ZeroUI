package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
	"github.com/mrtkrcm/ZeroUI/internal/tui/keybindings"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// ViewState represents the view states for the app
type ViewState int

const (
	ListView ViewState = iota // List-based app selection
	FormView                  // Dynamic forms for configuration
	HelpView                  // Rich markdown help system
	ProgressView              // Progress and loading operations
)

// App represents the TUI application with modern components
type App struct {
	engine     *toggle.Engine
	initialApp string
	program    *tea.Program
	ctx        context.Context
	logger     *logging.CharmLogger
}

// NewApp creates a new TUI application
func NewApp(initialApp string) (*App, error) {
	// Initialize logging first
	logConfig := logging.DefaultConfig()
	logger, err := logging.NewCharmLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize the toggle engine
	engine, err := toggle.NewEngine()
	if err != nil {
		logger.LogError(err, "engine_initialization")
		return nil, fmt.Errorf("failed to create toggle engine: %w", err)
	}

	logger.Info("ZeroUI initialized", 
		"initial_app", initialApp,
		"log_file", logger.GetFileLocation())

	return &App{
		engine:     engine,
		initialApp: initialApp,
		logger:     logger,
	}, nil
}

// Model represents the application state
type Model struct {
	// Core state
	engine *toggle.Engine
	state  ViewState
	width  int
	height int
	err    error
	ctx    context.Context
	logger *logging.CharmLogger

	// Modern components using Charm libraries
	appList     *components.ApplicationListModel
	configForm  *components.HuhConfigFormModel
	helpSystem  *components.GlamourHelpModel

	// UI state
	keyMap      keybindings.AppKeyMap
	styles      *styles.Styles
	theme       *styles.Theme
	currentApp  string
	showingHelp bool
	
	// Performance tracking and optimization
	lastRenderTime time.Time
	frameCount     int
	renderCache    map[ViewState]string
	lastCacheTime  time.Time
	cacheDuration  time.Duration
}

// NewModel creates a new model with modern components
func NewModel(engine *toggle.Engine, initialApp string, logger *logging.CharmLogger) (*Model, error) {
	// Initialize theme
	theme := styles.DefaultTheme()
	styles.SetTheme(theme)

	// Determine initial state
	initialState := ListView
	if initialApp != "" {
		initialState = FormView
	}

	model := &Model{
		engine:     engine,
		state:      initialState,
		currentApp: initialApp,
		keyMap:     keybindings.NewAppKeyMap(),
		styles:     styles.GetStyles(),
		theme:      theme,
		logger:     logger,

		// Initialize modern components
		appList:    components.NewApplicationList(),
		configForm: components.NewHuhConfigForm(initialApp),
		helpSystem: components.NewGlamourHelp(),
		
		// Performance optimization
		renderCache:   make(map[ViewState]string),
		cacheDuration: 50 * time.Millisecond, // 20fps cache for smooth performance
	}

	// Load app configuration if initial app is specified
	if initialApp != "" {
		if err := model.loadAppConfigForForm(initialApp); err != nil {
			logger.LogError(err, "initial_app_config_load", "app", initialApp)
			model.err = err
		}
	}

	logger.LogUIEvent("model_initialized", string(rune(initialState)), 
		"initial_app", initialApp)

	return model, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	startTime := time.Now()
	defer func() {
		m.logger.LogPerformance("model_init", time.Since(startTime))
	}()

	var cmds []tea.Cmd

	// Initialize components
	if cmd := m.appList.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := m.configForm.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := m.helpSystem.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Start initial progress indicator (using existing components)

	// Welcome message
	cmds = append(cmds, func() tea.Msg {
		return util.InfoMsg{
			Msg:  "ZeroUI ready with modern components",
			Type: util.InfoTypeSuccess,
		}
	})

	// Complete initialization
	cmds = append(cmds, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return InitCompleteMsg{}
	}))

	return tea.Batch(cmds...)
}

// Update handles model updates with comprehensive error handling and performance optimizations
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil {
		return m, tea.Quit
	}

	var cmds []tea.Cmd
	renderStart := time.Now()

	// Enhanced panic recovery and error boundaries
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("UI panic recovered", "panic", r, "state", m.state, "msg_type", fmt.Sprintf("%T", msg))
			m.err = fmt.Errorf("UI panic: %v", r)
			// Attempt to reset to stable state
			m.state = ListView
			m.currentApp = ""
		}
		// Performance monitoring with throttling to avoid log spam
		renderDuration := time.Since(renderStart)
		m.lastRenderTime = renderStart
		m.frameCount++
		
		// Only log performance issues if they're significant
		if renderDuration > 16*time.Millisecond && m.frameCount%60 == 0 { // Log every 60 frames if slow
			m.logger.LogPerformance("slow_update", renderDuration, "frame_count", m.frameCount)
		}
	}()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Prevent unnecessary updates for tiny size changes
		if abs(m.width-msg.Width) > 1 || abs(m.height-msg.Height) > 1 {
			m.width = msg.Width
			m.height = msg.Height
			cmds = append(cmds, m.updateComponentSizes())
		}

	case tea.KeyMsg:
		// Rate limiting for key events to prevent UI lag
		if m.frameCount > 0 && time.Since(m.lastRenderTime) < 16*time.Millisecond {
			// Skip processing if we're updating too frequently
			return m, nil
		}

		// Global key handling with error recovery
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			return m, cmd
		}

		// State-specific key handling
		if cmd := m.handleStateKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case components.AppSelectedMsg:
		// Validate app name before processing
		if msg.App == "" {
			m.logger.Error("Empty app name in AppSelectedMsg")
			return m, nil
		}
		
		m.logger.LogAppOperation(msg.App, "selected")
		m.currentApp = msg.App
		m.state = FormView
		m.invalidateCache() // Clear cache on state change

		// Load configuration asynchronously with timeout protection
		cmds = append(cmds, func() tea.Msg {
			// Add timeout to prevent hanging
			done := make(chan tea.Msg, 1)
			go func() {
				if err := m.loadAppConfigForForm(msg.App); err != nil {
					done <- util.InfoMsg{
						Msg:  fmt.Sprintf("Error loading config: %v", err),
						Type: util.InfoTypeError,
					}
				} else {
					done <- ConfigLoadedMsg{App: msg.App}
				}
			}()
			
			select {
			case result := <-done:
				return result
			case <-time.After(5 * time.Second):
				return util.InfoMsg{
					Msg:  fmt.Sprintf("Timeout loading config for %s", msg.App),
					Type: util.InfoTypeError,
				}
			}
		})

	case ConfigLoadedMsg:
		cmds = append(cmds, func() tea.Msg {
			return util.InfoMsg{
				Msg:  fmt.Sprintf("Configuration loaded for %s", msg.App),
				Type: util.InfoTypeSuccess,
			}
		})

	case components.ConfigSavedMsg:
		m.logger.LogAppOperation(msg.AppName, "config_saved", "fields", len(msg.Values))
		cmds = append(cmds, m.saveConfiguration(msg.AppName, msg.Values))

	case InitCompleteMsg:
		// Handle initialization complete

	case util.InfoMsg:
		m.logger.LogUIEvent("info_message", msg.Type.String(), "message", msg.Msg)

	}

	// Update components based on current state with error boundaries
	switch m.state {
	case ListView:
		if updatedList, listCmd := m.safeUpdateComponent(func() (interface{}, tea.Cmd) {
			return m.appList.Update(msg)
		}, "appList"); updatedList != nil {
			if appList, ok := updatedList.(*components.ApplicationListModel); ok {
				m.appList = appList
			}
			if listCmd != nil {
				cmds = append(cmds, listCmd)
			}
		}

	case FormView:
		if updatedForm, formCmd := m.safeUpdateComponent(func() (interface{}, tea.Cmd) {
			return m.configForm.Update(msg)
		}, "configForm"); updatedForm != nil {
			if configForm, ok := updatedForm.(*components.HuhConfigFormModel); ok {
				m.configForm = configForm
			}
			if formCmd != nil {
				cmds = append(cmds, formCmd)
			}
		}

	case HelpView:
		if updatedHelp, helpCmd := m.safeUpdateComponent(func() (interface{}, tea.Cmd) {
			return m.helpSystem.Update(msg)
		}, "helpSystem"); updatedHelp != nil {
			if helpSystem, ok := updatedHelp.(*components.GlamourHelpModel); ok {
				m.helpSystem = helpSystem
			}
			if helpCmd != nil {
				cmds = append(cmds, helpCmd)
			}
		}
	}


	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleGlobalKeys handles global key presses
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Quit, m.keyMap.ForceQuit):
		m.logger.Info("Application quit requested")
		return tea.Quit

	case key.Matches(msg, m.keyMap.Help):
		if m.state == HelpView {
			m.state = ListView
		} else {
			m.state = HelpView
			m.helpSystem.ShowPage("overview")
		}
		m.invalidateCache() // Clear cache on state change
		m.logger.LogUIEvent("help_toggled", "help", "visible", m.state == HelpView)

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	}

	return nil
}

// handleStateKeys handles state-specific key presses
func (m *Model) handleStateKeys(msg tea.KeyMsg) tea.Cmd {
	// Help system gets priority
	if m.state == HelpView {
		return nil // Let helpSystem handle all keys
	}

	// State-specific handling
	switch m.state {
	case ListView:
		// Handled by component updates
		return nil

	case FormView:
		switch msg.String() {
		case "esc":
			m.state = ListView
			m.currentApp = ""
		}
	}

	return nil
}

// handleBack handles the back/escape action
func (m *Model) handleBack() tea.Cmd {
	if m.showingHelp || m.state == HelpView {
		m.showingHelp = false
		m.state = ListView
		return nil
	}

	switch m.state {
	case ListView:
		return tea.Quit
	case FormView:
		m.state = ListView
		m.currentApp = ""
		m.invalidateCache() // Clear cache on state change
	default:
		m.state = ListView
		m.invalidateCache() // Clear cache on state change
	}

	return nil
}

// View renders the model with performance optimizations and intelligent caching
func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	// Check cache for non-interactive views
	if cachedContent, valid := m.getCachedView(); valid {
		return cachedContent
	}

	// Get the main content based on current state with error protection
	var content string

	switch m.state {
	case ListView:
		content = m.safeViewRender(func() string { return m.appList.View() }, "appList")

	case FormView:
		content = m.safeViewRender(func() string { return m.configForm.View() }, "configForm")

	case HelpView:
		content = m.safeViewRender(func() string { return m.helpSystem.View() }, "helpSystem")

	case ProgressView:
		content = "Progress view"
		
	default:
		content = m.renderFallbackView()
	}

	// Add performance indicator in debug mode (throttled to reduce overhead)
	if m.logger.IsLevelEnabled(logging.LevelDebug) && m.frameCount%30 == 0 {
		debugInfo := m.renderDebugInfo()
		content = lipgloss.JoinVertical(lipgloss.Left, content, debugInfo)
	}

	// Cache the result for static views
	m.cacheView(content)

	return content
}

// getCachedView returns cached content if valid
func (m *Model) getCachedView() (string, bool) {
	if m.renderCache == nil {
		return "", false
	}
	
	// Only cache non-form views to avoid stale form state
	if m.state == FormView {
		return "", false
	}
	
	if cached, exists := m.renderCache[m.state]; exists {
		if time.Since(m.lastCacheTime) < m.cacheDuration {
			return cached, true
		}
	}
	
	return "", false
}

// cacheView stores the rendered content
func (m *Model) cacheView(content string) {
	if m.renderCache == nil {
		m.renderCache = make(map[ViewState]string)
	}
	
	// Don't cache form views to avoid stale form state
	if m.state != FormView {
		m.renderCache[m.state] = content
		m.lastCacheTime = time.Now()
	}
}

// invalidateCache clears the render cache
func (m *Model) invalidateCache() {
	if m.renderCache != nil {
		for k := range m.renderCache {
			delete(m.renderCache, k)
		}
	}
}

// safeViewRender wraps view rendering with error recovery
func (m *Model) safeViewRender(renderFn func() string, componentName string) string {
	var result string
	var panicOccurred bool
	
	func() {
		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("View render panic recovered", "component", componentName, "panic", r)
				panicOccurred = true
			}
		}()
		
		result = renderFn()
	}()
	
	// Return fallback if error occurred or result is empty
	if panicOccurred || result == "" {
		return m.renderFallbackView()
	}
	return result
}

// renderFallbackView provides a safe fallback when components fail
func (m *Model) renderFallbackView() string {
	fallbackStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214"))

	return fallbackStyle.Render("ZeroUI - Safe Mode\n\nPress 'q' to quit or Esc to return to main view.")
}

// renderError renders error messages
func (m *Model) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196"))

	errorMsg := fmt.Sprintf("Error: %s\n\nPress 'q' to quit or Esc to continue.", m.err.Error())
	return errorStyle.Render(errorMsg)
}

// renderDebugInfo renders debug information
func (m *Model) renderDebugInfo() string {
	debugStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true)

	info := fmt.Sprintf("Debug: %dx%d | State: %d | Frames: %d | Last render: %v",
		m.width, m.height, m.state, m.frameCount, m.lastRenderTime)

	return debugStyle.Render(info)
}

// updateComponentSizes updates all component sizes
func (m *Model) updateComponentSizes() tea.Cmd {
	var cmds []tea.Cmd

	// Update components
	if cmd := m.appList.SetSize(m.width, m.height); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if cmd := m.configForm.SetSize(m.width, m.height); cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.helpSystem.SetSize(m.width, m.height)

	return tea.Batch(cmds...)
}

// loadAppConfigForForm loads configuration for the form
func (m *Model) loadAppConfigForForm(appName string) error {
	appConfig, err := m.engine.GetAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}

	currentValues, err := m.engine.GetCurrentValues(appName)
	if err != nil {
		currentValues = make(map[string]interface{})
	}

	// Convert to Huh form fields
	var fields []components.ConfigField
	for key, fieldConfig := range appConfig.Fields {
		field := components.ConfigField{
			Key:         key,
			Description: fieldConfig.Description,
			Required:    false, // Default to false since Required field may not exist
		}

		// Set current value and track if it's configured
		if val, exists := currentValues[key]; exists {
			field.Value = val
			field.IsSet = true
			field.Source = "config file"
		} else if fieldConfig.Default != nil {
			field.Value = fieldConfig.Default
			field.IsSet = false
			field.Source = "default value"
		} else {
			field.IsSet = false
			field.Source = "available option"
		}

		// Map field type
		switch fieldConfig.Type {
		case "string":
			field.Type = components.FieldTypeString
		case "int":
			field.Type = components.FieldTypeInt
		case "bool":
			field.Type = components.FieldTypeBool
		case "select":
			field.Type = components.FieldTypeSelect
			field.Options = fieldConfig.Values
		case "float":
			field.Type = components.FieldTypeFloat
		}

		fields = append(fields, field)
	}

	m.configForm.SetFields(fields)
	return nil
}

// saveConfiguration saves the configuration values
func (m *Model) saveConfiguration(appName string, values map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		for key, value := range values {
			// Convert interface{} to string for Toggle method
			var strValue string
			if value != nil {
				strValue = fmt.Sprintf("%v", value)
			}
			
			if err := m.engine.Toggle(appName, key, strValue); err != nil {
				m.logger.LogError(err, "config_save", "app", appName, "key", key)
				return util.InfoMsg{
					Msg:  fmt.Sprintf("Error saving %s: %v", key, err),
					Type: util.InfoTypeError,
				}
			}
			m.logger.LogConfigChange(appName, key, "", strValue)
		}

		return util.InfoMsg{
			Msg:  fmt.Sprintf("Configuration saved for %s", appName),
			Type: util.InfoTypeSuccess,
		}
	}
}

// Run starts the TUI application
func (app *App) Run() error {
	return app.RunWithContext(context.Background())
}

// RunWithContext starts the TUI application with context
func (app *App) RunWithContext(ctx context.Context) error {
	app.ctx = ctx

	defer func() {
		if r := recover(); r != nil {
			app.logger.Fatal("Application panic", "panic", r)
		}
		app.logger.Close()
	}()

	model, err := NewModel(app.engine, app.initialApp, app.logger)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	model.ctx = ctx

	app.program = tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	go func() {
		<-ctx.Done()
		if app.program != nil {
			app.program.Quit()
		}
	}()

	app.logger.Info("Starting ZeroUI application")

	if _, err := app.program.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		return fmt.Errorf("TUI application error: %w", err)
	}

	return nil
}

// Message types
type InitCompleteMsg struct{}
type ConfigLoadedMsg struct {
	App string
}

// safeUpdateComponent wraps component updates with error recovery
func (m *Model) safeUpdateComponent(updateFn func() (interface{}, tea.Cmd), componentName string) (interface{}, tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("Component update panic recovered", "component", componentName, "panic", r)
		}
	}()
	
	return updateFn()
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Ensure Model implements tea.Model
var _ tea.Model = (*Model)(nil)