package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
	"github.com/mrtkrcm/ZeroUI/internal/tui/keys"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// ViewState represents the current view state
type ViewState int

const (
	HuhAppSelectionView ViewState = iota // New Huh-based app selection (primary view)
	HuhConfigEditView                    // New Huh-based config editing
	AppGridView                          // Legacy grid view (fallback)
	AppSelectionView                     // Legacy app selection (fallback) 
	ConfigEditView                       // Legacy config editing (fallback)
	PresetSelectionView
	HelpView
)

// App represents the TUI application
type App struct {
	engine     *toggle.Engine
	initialApp string
	program    *tea.Program
}

// NewApp creates a new TUI application
func NewApp(initialApp string) (*App, error) {
	engine, err := toggle.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create toggle engine: %w", err)
	}

	return &App{
		engine:     engine,
		initialApp: initialApp,
	}, nil
}

// Run starts the TUI application
func (a *App) Run() error {
	// Add panic recovery for UI stability
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			err := fmt.Errorf("UI panic recovered: %v", r)
			// Log the error (will be captured by observability if enabled)
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}()

	model, err := NewModel(a.engine, a.initialApp)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	a.program = tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("TUI application error: %w", err)
	}

	return nil
}

// Model represents the application state
type Model struct {
	// Core state
	engine     *toggle.Engine
	state      ViewState
	width      int
	height     int
	err        error

	// New Huh-based components (primary)
	huhAppSelector  *components.HuhAppSelectorModel
	huhConfigEditor *components.HuhConfigEditorModel
	
	// Legacy components (fallback)
	appGrid        *components.AppGridModel      
	appSelector    *components.AppSelectorModel
	configEditor   *components.ConfigEditorModel
	statusBar      *components.StatusBarModel
	responsiveHelp *components.ResponsiveHelpModel
	help           help.Model
	
	// UI state
	keyMap        keys.AppKeyMap
	styles        *styles.Styles
	theme         *styles.Theme
	showingHelp   bool
	currentApp    string
	
	// Message handling
	lastMessage util.InfoMsg
}

// NewModel creates a new model
func NewModel(engine *toggle.Engine, initialApp string) (*Model, error) {
	apps, err := engine.GetApps()
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	// Initialize theme
	theme := styles.DefaultTheme()
	styles.SetTheme(theme)

	// Determine initial state based on whether an app was specified
	// Always start with the modern Huh-based interface
	initialState := HuhAppSelectionView
	if initialApp != "" {
		initialState = HuhConfigEditView
	}

	model := &Model{
		engine:          engine,
		state:           initialState,
		currentApp:      initialApp,
		keyMap:          keys.DefaultKeyMap(),
		styles:          styles.GetStyles(),
		theme:           theme,
		// Initialize new Huh-based components
		huhAppSelector:  components.NewHuhAppSelector(),
		huhConfigEditor: components.NewHuhConfigEditor(""),
		// Keep legacy components for fallback
		appGrid:         components.NewAppGrid(),
		appSelector:     components.NewAppSelector(apps),
		configEditor:    components.NewConfigEditor(""),
		statusBar:       components.NewStatusBar(),
		responsiveHelp:  components.NewResponsiveHelp(),
		help:            help.New(),
	}

	// Set up help
	model.help.ShortSeparator = " • "
	model.help.FullSeparator = "   "
	model.help.Ellipsis = "…"

	// If initial app is specified and exists, start with Huh config view
	if initialApp != "" {
		for _, app := range apps {
			if app == initialApp {
				model.state = HuhConfigEditView
				model.currentApp = initialApp
				if err := model.loadAppConfig(initialApp); err != nil {
					model.err = err
				}
				break
			}
		}
	}

	// Focus the appropriate component
	model.focusCurrentComponent()

	return model, nil
}

// Init initializes the model with proper component setup and performance optimization
func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	
	// Initialize all components with error handling
	// Priority: New Huh components first
	if cmd := m.huhAppSelector.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := m.huhConfigEditor.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Legacy components for fallback
	if cmd := m.appGrid.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := m.appSelector.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := m.configEditor.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := m.statusBar.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if cmd := m.responsiveHelp.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Initialize status bar with proper app count
	appCount := len(m.appSelector.GetApps())
	m.statusBar.SetAppCount(appCount)
	m.statusBar.SetTheme("Default")
	
	// Initialize with welcome message
	cmds = append(cmds, func() tea.Msg {
		return util.InfoMsg{
			Msg:  fmt.Sprintf("ZeroUI initialized with %d applications", appCount),
			Type: util.InfoTypeSuccess,
		}
	})
	
	// Add initial animation for smooth startup
	cmds = append(cmds, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return components.AnimationTickMsg{}
	}))
	
	return tea.Batch(cmds...)
}

// Update handles messages with performance optimization and non-blocking patterns
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle message with proper error recovery
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error for graceful handling
			m.err = fmt.Errorf("UI panic recovered: %v", r)
		}
	}()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed (performance optimization)
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			cmds = append(cmds, m.updateComponentSizes())
		}

	case tea.KeyMsg:
		// Handle global keys first with proper key mapping
		switch {
		case key.Matches(msg, m.keyMap.Help):
			m.showingHelp = !m.showingHelp
			return m, nil
		case key.Matches(msg, m.keyMap.Quit, m.keyMap.ForceQuit):
			// Graceful shutdown
			return m, tea.Sequence(
				tea.Printf("Shutting down ZeroUI..."),
				tea.Quit,
			)
		case key.Matches(msg, m.keyMap.Back):
			return m, m.handleBack()
		}

		// Handle state-specific keys (non-blocking)
		if !m.showingHelp {
			cmd := m.handleStateKeys(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
		// Handle view switching keys
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+h"))):
			// Switch to Huh interface
			if m.state == AppGridView {
				m.state = HuhAppSelectionView
			} else if m.state == ConfigEditView {
				m.state = HuhConfigEditView
			}
			m.focusCurrentComponent()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+l"))):
			// Switch to legacy interface
			if m.state == HuhAppSelectionView {
				m.state = AppGridView
			} else if m.state == HuhConfigEditView {
				m.state = ConfigEditView
			}
			m.focusCurrentComponent()
			return m, nil
		}

	case components.AppSelectedMsg:
		// Handle app selection with loading state
		m.currentApp = msg.App
		// Use modern Huh config editor by default
		m.state = HuhConfigEditView
		
		// Load config asynchronously to prevent UI freezing
		cmds = append(cmds, func() tea.Msg {
			if err := m.loadAppConfig(msg.App); err != nil {
				return util.InfoMsg{Msg: fmt.Sprintf("Error loading config: %v", err), Type: util.InfoTypeError}
			}
			return util.InfoMsg{Msg: fmt.Sprintf("Loaded configuration for %s", msg.App), Type: util.InfoTypeSuccess}
		})
		
		m.focusCurrentComponent()

	case components.FieldChangedMsg:
		// Handle field changes asynchronously
		cmds = append(cmds, func() tea.Msg {
			if err := m.engine.Toggle(m.currentApp, msg.Key, msg.Value); err != nil {
				return util.InfoMsg{Msg: fmt.Sprintf("Error updating %s: %v", msg.Key, err), Type: util.InfoTypeError}
			}
			return util.InfoMsg{Msg: fmt.Sprintf("Updated %s", msg.Key), Type: util.InfoTypeSuccess}
		})

	case components.OpenPresetsMsg:
		m.state = PresetSelectionView
		m.focusCurrentComponent()

	case util.InfoMsg:
		m.lastMessage = msg
		
	case components.AnimationTickMsg:
		// Handle animation updates for smooth UI
		if m.state == AppGridView {
			// Update app grid animations
			updatedGrid, cmd := m.appGrid.Update(msg)
			m.appGrid = updatedGrid
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Update components based on current state (performance optimized)
	switch m.state {
	case HuhAppSelectionView:
		// Update Huh app selector with enhanced styling
		updatedHuhSelector, cmd := m.huhAppSelector.Update(msg)
		m.huhAppSelector = updatedHuhSelector
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
	case HuhConfigEditView:
		// Update Huh config editor with forms
		if model, cmd := m.huhConfigEditor.Update(msg); cmd != nil || model != m.huhConfigEditor {
			m.huhConfigEditor = model.(*components.HuhConfigEditorModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
	case AppGridView:
		// Update legacy app grid component 
		updatedGrid, cmd := m.appGrid.Update(msg)
		m.appGrid = updatedGrid
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
	case AppSelectionView:
		// Update legacy app selector
		if model, cmd := m.appSelector.Update(msg); cmd != nil || model != m.appSelector {
			m.appSelector = model.(*components.AppSelectorModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
	case ConfigEditView:
		// Update legacy config editor
		if model, cmd := m.configEditor.Update(msg); cmd != nil || model != m.configEditor {
			m.configEditor = model.(*components.ConfigEditorModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Always update status bar and help components
	if statusModel, cmd := m.statusBar.Update(msg); cmd != nil {
		m.statusBar = statusModel.(*components.StatusBarModel)
		cmds = append(cmds, cmd)
	}
	
	if helpModel, cmd := m.responsiveHelp.Update(msg); cmd != nil {
		m.responsiveHelp = helpModel.(*components.ResponsiveHelpModel)
		cmds = append(cmds, cmd)
	}

	// Return with batched commands for optimal performance
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleStateKeys handles state-specific key presses
func (m *Model) handleStateKeys(msg tea.KeyMsg) tea.Cmd {
	switch m.state {
	case HuhAppSelectionView:
		// Keys are handled by the Huh app selector component
		return nil
	case HuhConfigEditView:
		// Keys are handled by the Huh config editor component
		return nil
	case AppSelectionView:
		// Keys are handled by the legacy app selector component
		return nil
	case ConfigEditView:
		// Keys are handled by the legacy config editor component
		return nil
	case AppGridView:
		// Keys are handled by the app grid component
		return nil
	case PresetSelectionView:
		// TODO: Implement preset selection
		return nil
	case HelpView:
		// Help view doesn't need special key handling
		return nil
	}
	return nil
}

// handleBack handles the back/escape key
func (m *Model) handleBack() tea.Cmd {
	if m.showingHelp {
		m.showingHelp = false
		return nil
	}

	switch m.state {
	case HuhAppSelectionView:
		return tea.Quit
	case HuhConfigEditView:
		if m.currentApp == "" {
			m.state = HuhAppSelectionView
		} else {
			m.state = HuhAppSelectionView
		}
		m.focusCurrentComponent()
	case AppGridView:
		return tea.Quit
	case AppSelectionView:
		if m.currentApp == "" {
			m.state = AppGridView
		} else {
			return tea.Quit
		}
	case ConfigEditView, PresetSelectionView:
		if m.currentApp == "" {
			m.state = AppGridView
		} else {
			m.state = AppSelectionView
		}
		m.focusCurrentComponent()
	case HelpView:
		m.state = HuhAppSelectionView
		m.focusCurrentComponent()
	}
	return nil
}

// focusCurrentComponent focuses the appropriate component for the current state
func (m *Model) focusCurrentComponent() {
	// Blur all components first
	m.huhAppSelector.Blur()
	m.huhConfigEditor.Blur()
	m.appSelector.Blur()
	m.configEditor.Blur()

	// Focus the current component based on state
	switch m.state {
	case HuhAppSelectionView:
		m.huhAppSelector.Focus()
	case HuhConfigEditView:
		m.huhConfigEditor.Focus()
	case AppSelectionView:
		m.appSelector.Focus()
	case ConfigEditView:
		m.configEditor.Focus()
	}
}

// updateComponentSizes updates the size of all components
func (m *Model) updateComponentSizes() tea.Cmd {
	// Calculate available space for content
	titleHeight := 1
	statusHeight := 1
	helpHeight := 1
	padding := 2
	
	contentHeight := m.height - titleHeight - statusHeight - helpHeight - padding
	if contentHeight < 3 {
		contentHeight = 3
	}
	
	contentWidth := m.width - 4 // Account for padding

	var cmds []tea.Cmd
	
	// Update Huh components first
	cmds = append(cmds, m.huhAppSelector.SetSize(contentWidth, contentHeight))
	cmds = append(cmds, m.huhConfigEditor.SetSize(contentWidth, contentHeight))
	
	// Update legacy components
	cmds = append(cmds, m.appSelector.SetSize(contentWidth, contentHeight))
	cmds = append(cmds, m.configEditor.SetSize(contentWidth, contentHeight))
	cmds = append(cmds, m.statusBar.SetSize(m.width, statusHeight))
	cmds = append(cmds, m.responsiveHelp.SetSize(m.width, helpHeight))
	
	// Update help bindings based on current state
	m.updateHelpBindings()
	
	return tea.Batch(cmds...)
}

// updateHelpBindings updates the help component with current context bindings
func (m *Model) updateHelpBindings() {
	var bindings []key.Binding

	switch m.state {
	case HuhAppSelectionView:
		bindings = m.huhAppSelector.Bindings()
	case HuhConfigEditView:
		bindings = m.huhConfigEditor.Bindings()
	case AppSelectionView:
		bindings = m.appSelector.Bindings()
	case ConfigEditView:
		bindings = m.configEditor.Bindings()
	case PresetSelectionView:
		// TODO: Add preset selection bindings when implemented
		bindings = []key.Binding{}
	}

	// Add global bindings
	bindings = append(bindings, m.keyMap.Help, m.keyMap.Back, m.keyMap.Quit)
	
	// Add view switching bindings
	bindings = append(bindings, 
		key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "switch to Huh UI")),
		key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "switch to legacy UI")),
	)

	// Update responsive help
	m.responsiveHelp.SetBindings(bindings)
}

// View renders the current view
func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	var content string

	if m.showingHelp {
		content = m.renderHelp()
	} else {
		content = m.renderMainContent()
	}

	// For modern Huh views, return content directly (they handle their own layout)
	if (m.state == HuhAppSelectionView || m.state == HuhConfigEditView) && !m.showingHelp {
		return content
	}
	
	// For legacy AppGridView, return content directly without layout wrapping
	if m.state == AppGridView && !m.showingHelp {
		return content
	}

	return m.wrapWithLayout(content)
}

// renderMainContent renders the main application content
func (m *Model) renderMainContent() string {
	switch m.state {
	case HuhAppSelectionView:
		// Render the modern Huh app selector with elegant styling
		return m.huhAppSelector.View()
	case HuhConfigEditView:
		// Render the modern Huh config editor with forms
		return m.huhConfigEditor.View()
	case AppGridView:
		// Legacy grid view (fallback)
		return m.appGrid.View()
	case AppSelectionView:
		// Legacy list selector (fallback)
		return m.appSelector.View()
	case ConfigEditView:
		// Legacy config editor (fallback)
		return m.configEditor.View()
	case PresetSelectionView:
		return m.styles.Muted.Render("Preset selection coming soon...")
	case HelpView:
		return m.renderHelp()
	}
	return ""
}

// renderHelp renders the help content
func (m *Model) renderHelp() string {
	var bindings []key.Binding

	switch m.state {
	case HuhAppSelectionView:
		bindings = m.huhAppSelector.Bindings()
	case HuhConfigEditView:
		bindings = m.huhConfigEditor.Bindings()
	case AppSelectionView:
		bindings = m.appSelector.Bindings()
	case ConfigEditView:
		bindings = m.configEditor.Bindings()
	}

	// Add global bindings
	bindings = append(bindings, m.keyMap.Help, m.keyMap.Back, m.keyMap.Quit)
	
	// Add view switching bindings
	bindings = append(bindings, 
		key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "switch to modern Huh UI")),
		key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "switch to legacy UI")),
	)

	return m.help.View(&keys.AppKeyMap{})
}

// wrapWithLayout wraps content with title, status, and help
func (m *Model) wrapWithLayout(content string) string {
	title := m.renderTitle()
	statusBar := m.statusBar.View()
	helpText := m.responsiveHelp.View()

	// Calculate available height for content
	titleHeight := lipgloss.Height(title)
	statusHeight := lipgloss.Height(statusBar)
	helpHeight := lipgloss.Height(helpText)
	
	// Reserve space for padding and borders
	reservedHeight := titleHeight + statusHeight + helpHeight + 2 // +2 for padding
	contentHeight := m.height - reservedHeight
	
	// Ensure minimum content height
	if contentHeight < 3 {
		contentHeight = 3
	}

	// Create content area with better spacing
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Padding(1, 2)

	styledContent := contentStyle.Render(content)

	// Add section separator if status bar has content
	sections := []string{title, styledContent}
	
	if statusBar != "" {
		sections = append(sections, statusBar)
	}
	
	if helpText != "" {
		sections = append(sections, helpText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTitle renders the application title
func (m *Model) renderTitle() string {
	var titleText string
	
	switch m.state {
	case HuhAppSelectionView:
		titleText = "🔧 ZeroUI - Select Application (Modern)"
	case HuhConfigEditView:
		titleText = fmt.Sprintf("⚙️ ZeroUI - %s Configuration (Huh Forms)", m.currentApp)
	case AppSelectionView:
		titleText = "ZeroUI - Select Application (Legacy)"
	case ConfigEditView:
		titleText = fmt.Sprintf("ZeroUI - %s Configuration (Legacy)", m.currentApp)
	case AppGridView:
		titleText = "ZeroUI - Application Grid"
	case PresetSelectionView:
		titleText = fmt.Sprintf("ZeroUI - %s Presets", m.currentApp)
	case HelpView:
		titleText = "ZeroUI - Help"
	}

	titleStyle := m.styles.Title.
		Width(m.width).
		Align(lipgloss.Left).
		Background(lipgloss.Color(styles.ColorToHex(styles.GetTheme().Primary)))

	return titleStyle.Render(titleText)
}

// renderStatus renders the status bar
func (m *Model) renderStatus() string {
	if m.lastMessage.Msg != "" {
		style := m.styles.Info
		switch m.lastMessage.Type {
		case util.InfoTypeWarn:
			style = m.styles.Warning
		case util.InfoTypeError:
			style = m.styles.Error
		}
		return style.Render(m.lastMessage.Msg)
	}
	return ""
}

// renderQuickHelp renders the quick help line
func (m *Model) renderQuickHelp() string {
	if m.showingHelp {
		return m.styles.Help.Render("Press ? to close help")
	}

	var bindings []key.Binding
	
	switch m.state {
	case AppSelectionView:
		bindings = []key.Binding{m.keyMap.Up, m.keyMap.Down, m.keyMap.Enter}
	case ConfigEditView:
		bindings = []key.Binding{m.keyMap.Up, m.keyMap.Down, m.keyMap.Left, m.keyMap.Right, m.keyMap.Enter}
	}
	
	bindings = append(bindings, m.keyMap.Help, m.keyMap.Quit)

	helpStyle := m.styles.Help.Width(m.width)
	return helpStyle.Render(m.help.ShortHelpView(bindings))
}

// renderError renders error messages
func (m *Model) renderError() string {
	errorMsg := fmt.Sprintf("Error: %s\n\nPress 'q' to quit or esc to continue.", m.err.Error())
	return m.styles.Error.Render(errorMsg)
}

// loadAppConfig loads configuration data for display in the TUI
func (m *Model) loadAppConfig(appName string) error {
	// Load the app configuration metadata
	appConfig, err := m.engine.GetAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}

	// Load current values from the target config file
	currentValues, err := m.engine.GetCurrentValues(appName)
	if err != nil {
		// If target config doesn't exist, use defaults
		currentValues = make(map[string]interface{})
	}

	// Convert fields to component format
	var fields []*components.FieldModel
	for key, fieldConfig := range appConfig.Fields {
		currentValue := ""
		if val, exists := currentValues[key]; exists {
			currentValue = fmt.Sprintf("%v", val)
		} else if fieldConfig.Default != nil {
			currentValue = fmt.Sprintf("%v", fieldConfig.Default)
		}

		field := components.NewField(key, fieldConfig.Type, currentValue, fieldConfig.Values, fieldConfig.Description)
		fields = append(fields, field)
	}

	// Update both legacy and modern config editors
	m.configEditor.SetAppName(appName)
	m.configEditor.SetFields(fields)
	
	m.huhConfigEditor.SetAppName(appName)
	m.huhConfigEditor.SetFields(fields)

	return nil
}

// Ensure Model implements tea.Model
var _ tea.Model = (*Model)(nil)