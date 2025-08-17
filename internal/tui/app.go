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
	
	// Performance tracking
	lastRenderTime time.Time
	frameCount     int
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

// Update handles model updates
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m == nil {
		return m, tea.Quit
	}

	var cmds []tea.Cmd
	renderStart := time.Now()

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("UI panic recovered", "panic", r, "state", m.state)
			m.err = fmt.Errorf("UI panic: %v", r)
		}
		m.lastRenderTime = renderStart
		m.frameCount++
	}()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		cmds = append(cmds, m.updateComponentSizes())

	case tea.KeyMsg:
		// Global key handling
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			return m, cmd
		}

		// State-specific key handling
		if cmd := m.handleStateKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case components.AppSelectedMsg:
		m.logger.LogAppOperation(msg.App, "selected")
		m.currentApp = msg.App
		m.state = FormView

		// Load configuration asynchronously
		cmds = append(cmds, func() tea.Msg {
			if err := m.loadAppConfigForForm(msg.App); err != nil {
				return util.InfoMsg{
					Msg:  fmt.Sprintf("Error loading config: %v", err),
					Type: util.InfoTypeError,
				}
			}
			return ConfigLoadedMsg{App: msg.App}
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
		m.logger.LogUIEvent("info_message", string(msg.Type), "message", msg.Msg)

	// case components.OperationCompleteMsg:
	//	m.logger.LogPerformance("operation_"+msg.ID, msg.Duration, "message", msg.Message)
	}

	// Update components based on current state
	switch m.state {
	case ListView:
		if updatedList, listCmd := m.appList.Update(msg); updatedList != nil {
			m.appList = updatedList
			if listCmd != nil {
				cmds = append(cmds, listCmd)
			}
		}

	case FormView:
		var cmd tea.Cmd
		m.configForm, cmd = m.configForm.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case HelpView:
		var cmd tea.Cmd
		m.helpSystem, cmd = m.helpSystem.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
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
	default:
		m.state = ListView
	}

	return nil
}

// View renders the model
func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	// Get the main content based on current state
	var content string

	switch m.state {
	case ListView:
		content = m.appList.View()

	case FormView:
		content = m.configForm.View()

	case HelpView:
		content = m.helpSystem.View()

	case ProgressView:
		content = "Progress view"
	}

	// Add progress overlay if active (placeholder)

	// Add performance indicator in debug mode
	if m.logger.IsLevelEnabled(logging.LevelDebug) {
		debugInfo := m.renderDebugInfo()
		content = lipgloss.JoinVertical(lipgloss.Left, content, debugInfo)
	}

	return content
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

		// Set current value
		if val, exists := currentValues[key]; exists {
			field.Value = val
		} else if fieldConfig.Default != nil {
			field.Value = fieldConfig.Default
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

// Ensure Model implements tea.Model
var _ tea.Model = (*Model)(nil)