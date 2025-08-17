package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	// Primary states - stable and tested
	AppGridView ViewState = iota // Main application grid (default)
	ConfigEditView               // Configuration editor
	HelpView                     // Help display
	
	// Secondary states - for advanced features
	HuhAppSelectionView // Modern app selection
	HuhConfigEditView   // Modern config editing
	PresetSelectionView // Preset management
)

// App represents the TUI application
type App struct {
	engine     *toggle.Engine
	initialApp string
	program    *tea.Program
	ctx        context.Context
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
	return a.RunWithContext(context.Background())
}

// RunWithContext starts the TUI application with context support for graceful shutdown
func (a *App) RunWithContext(ctx context.Context) error {
	a.ctx = ctx

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

	// Set context on the model for graceful shutdown handling
	model.ctx = ctx

	a.program = tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx), // Pass context to tea program
	)

	// Start a goroutine to monitor context cancellation
	go func() {
		<-ctx.Done()
		if a.program != nil {
			a.program.Quit()
		}
	}()

	if _, err := a.program.Run(); err != nil {
		// Check if error is due to context cancellation
		if ctx.Err() == context.Canceled {
			return nil // Graceful shutdown, not an error
		}
		return fmt.Errorf("TUI application error: %w", err)
	}

	return nil
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

	// Enhanced UX components
	search              SearchModel
	statusBar           StatusBar
	contextualHelp      ContextualHelp
	loadingIndicator    LoadingIndicator
	enhancedKeyMap      EnhancedKeyMap

	// New Huh-based components (primary)
	huhAppSelector  *components.HuhAppSelectorModel
	huhConfigEditor *components.HuhConfigEditorModel
	huhGrid         *components.HuhGridModel // New grid component

	// Delightful new components
	delightfulUI *components.DelightfulUIModel
	animatedList *components.AnimatedListModel

	// Legacy components (fallback)
	appGrid        *components.AppGridModel
	appSelector    *components.AppSelectorModel
	configEditor   *components.ConfigEditorModel
	statusBarLegacy *components.StatusBarModel
	responsiveHelp *components.ResponsiveHelpModel
	help           help.Model

	// UI state
	keyMap      keys.AppKeyMap
	styles      *styles.Styles
	theme       *styles.Theme
	showingHelp bool
	currentApp  string

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
	// Default to traditional 4-column card grid, with Huh as alternative
	initialState := AppGridView // Start with traditional 4-column grid
	if initialApp != "" {
		initialState = ConfigEditView // Use traditional config editor
	}

	model := &Model{
		engine:     engine,
		state:      initialState,
		currentApp: initialApp,
		keyMap:     keys.DefaultKeyMap(),
		styles:     styles.GetStyles(),
		theme:      theme,
		
		// Initialize enhanced UX components
		search:              NewSearchModel("Search apps..."),
		statusBar:           NewStatusBar(),
		contextualHelp:      NewContextualHelp(),
		loadingIndicator:    NewLoadingIndicator(),
		enhancedKeyMap:      NewEnhancedKeyMap(),
		
		// Initialize new Huh-based components
		huhGrid:         components.NewHuhGrid(),
		huhAppSelector:  components.NewHuhAppSelector(),
		huhConfigEditor: components.NewHuhConfigEditor(""),
		// Initialize delightful components
		delightfulUI: components.NewDelightfulUI(),
		animatedList: components.NewAnimatedList(),
		// Keep legacy components for fallback
		appGrid:        components.NewAppGrid(),
		appSelector:    components.NewAppSelector(apps),
		configEditor:   components.NewConfigEditor(""),
		statusBarLegacy: components.NewStatusBar(),
		responsiveHelp: components.NewResponsiveHelp(),
		help:           help.New(),
	}
	
	// Initialize contextual help maps
	model.contextualHelp.AddHelpMap("app_grid", model.enhancedKeyMap.GetContextualHelp("app_grid"))
	model.contextualHelp.AddHelpMap("config_edit", model.enhancedKeyMap.GetContextualHelp("config_edit"))
	model.contextualHelp.AddHelpMap("search", model.enhancedKeyMap.GetContextualHelp("search"))

	// Set up help
	model.help.ShortSeparator = " • "
	model.help.FullSeparator = "   "
	model.help.Ellipsis = "…"

	// If initial app is specified and exists, start with Huh config view
	if initialApp != "" {
		for _, app := range apps {
			if app == initialApp {
				model.state = ConfigEditView
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
	
	// Initialize status bar with current context
	model.updateStatusBar(model.getCurrentContext())

	return model, nil
}

// Init initializes the model with proper component setup and performance optimization
func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize all components with error handling
	// Priority: New Huh components first
	if cmd := m.huhGrid.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.huhAppSelector.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.huhConfigEditor.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Initialize delightful components
	if cmd := m.delightfulUI.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.animatedList.Init(); cmd != nil {
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

	if cmd := m.statusBarLegacy.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.responsiveHelp.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Initialize legacy status bar with proper app count
	appCount := len(m.appSelector.GetApps())
	m.statusBarLegacy.SetAppCount(appCount)
	m.statusBarLegacy.SetTheme("Default")

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

// Update handles messages with defensive error handling
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Defensive check - ensure model is valid
	if m == nil {
		return m, tea.Quit
	}
	
	var cmds []tea.Cmd

	// Panic recovery with detailed logging
	defer func() {
		if r := recover(); r != nil {
			m.err = fmt.Errorf("UI panic recovered: %v", r)
		}
	}()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed (performance optimization)
		if m.width != msg.Width || m.height != msg.Height {
			oldWidth, oldHeight := m.width, m.height
			m.width = msg.Width
			m.height = msg.Height

			// Only trigger expensive size updates if significant change (>5% or >10 pixels)
			widthDiff := abs(msg.Width - oldWidth)
			heightDiff := abs(msg.Height - oldHeight)

			if widthDiff > 10 || heightDiff > 10 ||
				(oldWidth > 0 && widthDiff*100/oldWidth > 5) ||
				(oldHeight > 0 && heightDiff*100/oldHeight > 5) {
				cmds = append(cmds, m.updateComponentSizes())
			}
		}

	case tea.KeyMsg:
		// Handle enhanced UX features first
		if cmd := m.handleEnhancedUXKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Handle search if active
		if m.search.active {
			if cmd := m.search.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			// Handle keys in priority order for predictability
			if cmd := m.handleGlobalKeys(msg); cmd != nil {
				return m, cmd
			}
			
			if !m.showingHelp {
				if cmd := m.handleStateKeys(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case components.AppSelectedMsg:
		// Handle app selection with loading state
		m.currentApp = msg.App
		// Use legacy config editor for stability
		m.state = ConfigEditView

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
		// Show toast notification for user feedback
		var toastLevel ToastLevel
		switch msg.Type {
		case util.InfoTypeError:
			toastLevel = ToastError
		case util.InfoTypeSuccess:
			toastLevel = ToastSuccess
		case util.InfoTypeWarn:
			toastLevel = ToastWarning
		default:
			toastLevel = ToastInfo
		}
		cmds = append(cmds, m.statusBar.ShowToast(msg.Msg, toastLevel, 3*time.Second))

	case ToastTimeoutMsg:
		cmds = append(cmds, m.statusBar.Update(msg))

	case LoadingTickMsg:
		cmds = append(cmds, m.loadingIndicator.Update(msg))

	case components.AnimationTickMsg:
		// Handle animation updates for smooth UI - batch updates to reduce overhead
		if m.state == AppGridView {
			// Only update animations if component is actually visible and focused
			if m.appGrid != nil {
				updatedGrid, cmd := m.appGrid.Update(msg)
				m.appGrid = updatedGrid
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

		// Reduce animation frequency for better performance
		cmds = append(cmds, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
			return components.AnimationTickMsg{}
		}))
	}

	// Update components with defensive checks
	switch m.state {
	case AppGridView:
		if m.appGrid != nil {
			if updatedGrid, cmd := m.appGrid.Update(msg); updatedGrid != nil {
				m.appGrid = updatedGrid
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case ConfigEditView:
		if m.configEditor != nil {
			if model, cmd := m.configEditor.Update(msg); model != nil {
				if editor, ok := model.(*components.ConfigEditorModel); ok {
					m.configEditor = editor
				}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case HuhAppSelectionView:
		if m.huhAppSelector != nil {
			if updatedSelector, cmd := m.huhAppSelector.Update(msg); updatedSelector != nil {
				m.huhAppSelector = updatedSelector
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case HuhConfigEditView:
		if m.huhConfigEditor != nil {
			if model, cmd := m.huhConfigEditor.Update(msg); model != nil {
				if editor, ok := model.(*components.HuhConfigEditorModel); ok {
					m.huhConfigEditor = editor
				}
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// Return with batched commands for optimal performance
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleEnhancedUXKeys handles enhanced UX features like search, help, etc.
func (m *Model) handleEnhancedUXKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.enhancedKeyMap.Search):
		// Activate search mode
		if !m.search.active {
			m.updateStatusBar("search")
			return m.search.ActivateSearch()
		}
		return nil
		
	case key.Matches(msg, m.enhancedKeyMap.Escape):
		// Deactivate search if active
		if m.search.active {
			m.search.DeactivateSearch()
			m.updateStatusBar(m.getCurrentContext())
			return nil
		}
		return nil
		
	case key.Matches(msg, m.enhancedKeyMap.Help):
		// Toggle contextual help
		m.contextualHelp.Toggle()
		return nil
		
	case key.Matches(msg, m.enhancedKeyMap.Reload):
		// Refresh current view with loading indicator
		return m.loadingIndicator.Start("Refreshing...")
		
	case key.Matches(msg, m.enhancedKeyMap.Save):
		// Save configuration with feedback
		if m.currentApp != "" {
			return func() tea.Msg {
				return util.InfoMsg{
					Msg:  fmt.Sprintf("Configuration saved for %s", m.currentApp),
					Type: util.InfoTypeSuccess,
				}
			}
		}
		return nil
	}
	return nil
}

// handleGlobalKeys handles global key presses with predictable priority
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Quit, m.keyMap.ForceQuit):
		return tea.Quit
	case key.Matches(msg, m.keyMap.Help):
		m.showingHelp = !m.showingHelp
		return nil
	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()
	}
	return nil
}

// handleStateKeys handles state-specific key presses
func (m *Model) handleStateKeys(msg tea.KeyMsg) tea.Cmd {
	// Most keys are handled by individual components
	// Only handle special cases here
	if m.state == PresetSelectionView {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.currentApp != "" {
				return func() tea.Msg {
					presets, err := m.engine.GetPresets(m.currentApp)
					if err != nil || len(presets) == 0 {
						return util.InfoMsg{
							Msg:  "No presets available",
							Type: util.InfoTypeError,
						}
					}
					
					presetName := presets[0]
					if err := m.engine.ApplyPreset(m.currentApp, presetName); err != nil {
						return util.InfoMsg{
							Msg:  fmt.Sprintf("Failed to apply preset: %v", err),
							Type: util.InfoTypeError,
						}
					}
					return util.InfoMsg{
						Msg:  fmt.Sprintf("Applied preset '%s' to %s", presetName, m.currentApp),
						Type: util.InfoTypeSuccess,
					}
				}
			}
		}
	}
	return nil
}

// handleBack handles the back/escape key with predictable transitions
func (m *Model) handleBack() tea.Cmd {
	if m.showingHelp {
		m.showingHelp = false
		return nil
	}

	// Simple state machine - always go back to grid or quit
	switch m.state {
	case AppGridView:
		return tea.Quit
	case ConfigEditView, PresetSelectionView, HuhConfigEditView:
		m.state = AppGridView
		m.currentApp = ""
	case HuhAppSelectionView, HelpView:
		m.state = AppGridView
	default:
		return tea.Quit
	}
	
	m.focusCurrentComponent()
	return nil
}

// focusCurrentComponent focuses the appropriate component for the current state
func (m *Model) focusCurrentComponent() {
	// Blur all components first
	m.huhAppSelector.Blur()
	m.huhConfigEditor.Blur()
	m.huhGrid.Blur()
	m.appSelector.Blur()
	m.configEditor.Blur()

	// Focus the current component based on state
	switch m.state {
	case HuhAppSelectionView:
		if m.huhAppSelector != nil {
			m.huhAppSelector.Focus()
		}
	case HuhConfigEditView:
		if m.huhConfigEditor != nil {
			m.huhConfigEditor.Focus()
		}
	case AppGridView:
		// App grid handles its own internal selection state
	case ConfigEditView:
		if m.configEditor != nil {
			m.configEditor.Focus()
		}
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
	cmds = append(cmds, m.huhGrid.SetSize(contentWidth, contentHeight))

	// Update legacy components
	cmds = append(cmds, m.appSelector.SetSize(contentWidth, contentHeight))
	cmds = append(cmds, m.configEditor.SetSize(contentWidth, contentHeight))
	cmds = append(cmds, m.statusBarLegacy.SetSize(m.width, statusHeight))
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
		if m.huhAppSelector != nil {
			bindings = m.huhAppSelector.Bindings()
		}
	case HuhConfigEditView:
		if m.huhConfigEditor != nil {
			bindings = m.huhConfigEditor.Bindings()
		}
	case ConfigEditView:
		if m.configEditor != nil {
			bindings = m.configEditor.Bindings()
		}
	case PresetSelectionView:
		bindings = []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply preset")),
			key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q/esc", "back")),
		}
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

	// Return content directly for simpler rendering
	if (m.state == HuhAppSelectionView || m.state == HuhConfigEditView || m.state == AppGridView) && !m.showingHelp {
		return content
	}

	return m.wrapWithLayout(content)
}

// updateStatusBar updates the status bar with current context
func (m *Model) updateStatusBar(context string) {
	m.statusBar.SetCurrentView(context)
	m.statusBar.SetCurrentApp(m.currentApp)
	
	// Set contextual key hints
	hints := []string{}
	switch context {
	case "app_grid":
		hints = []string{"↑↓←→: Navigate", "Enter: Select", "/: Search", "?: Help"}
	case "config_edit":
		hints = []string{"↑↓: Navigate", "Enter: Edit", "Ctrl+S: Save", "Esc: Back"}
	case "search":
		hints = []string{"Type: Search", "Enter: Select", "Esc: Cancel"}
	}
	m.statusBar.SetKeyHints(hints)
	
	// Update contextual help context
	m.contextualHelp.SetContext(context)
}

// getCurrentContext returns the current UI context
func (m *Model) getCurrentContext() string {
	switch m.state {
	case AppGridView:
		return "app_grid"
	case ConfigEditView, HuhConfigEditView:
		return "config_edit"
	case HuhAppSelectionView:
		return "app_selection"
	default:
		return "general"
	}
}

// renderMainContent renders the main application content with enhanced UX overlays
func (m *Model) renderMainContent() string {
	var content string
	
	// Render main content based on state
	switch m.state {
	case AppGridView:
		if m.appGrid != nil {
			content = m.appGrid.View()
		}
	case ConfigEditView:
		if m.configEditor != nil {
			content = m.configEditor.View()
		}
	case HuhAppSelectionView:
		if m.huhAppSelector != nil {
			content = m.huhAppSelector.View()
		}
	case HuhConfigEditView:
		if m.huhConfigEditor != nil {
			content = m.huhConfigEditor.View()
		}
	case PresetSelectionView:
		content = m.renderPresetSelection()
	case HelpView:
		content = m.renderHelp()
	default:
		content = m.loadingIndicator.View()
		if content == "" {
			content = "Loading..."
		}
	}
	
	// Add enhanced UX overlays
	overlays := []string{}
	
	// Add search overlay
	if m.search.active {
		overlays = append(overlays, m.search.View())
	}
	
	// Add contextual help overlay
	if m.contextualHelp.visible {
		overlays = append(overlays, m.contextualHelp.View(m.width, m.height))
	}
	
	// Add loading indicator overlay
	if m.loadingIndicator.active {
		overlays = append(overlays, m.loadingIndicator.View())
	}
	
	// Combine content with overlays
	if len(overlays) > 0 {
		allContent := []string{content}
		allContent = append(allContent, overlays...)
		content = lipgloss.JoinVertical(lipgloss.Left, allContent...)
	}
	
	// Add status bar at bottom
	statusBar := m.statusBar.View(m.width)
	if statusBar != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
	}
	
	return content
}

// renderHelp renders the help content
func (m *Model) renderHelp() string {
	var bindings []key.Binding

	switch m.state {
	case HuhAppSelectionView:
		if m.huhAppSelector != nil {
			bindings = m.huhAppSelector.Bindings()
		}
	case HuhConfigEditView:
		if m.huhConfigEditor != nil {
			bindings = m.huhConfigEditor.Bindings()
		}
	case ConfigEditView:
		if m.configEditor != nil {
			bindings = m.configEditor.Bindings()
		}
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
	statusBar := m.statusBar.View(m.width)
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
		titleText = "🔧 ZeroUI - Select Application"
	case HuhConfigEditView:
		titleText = fmt.Sprintf("⚙️ ZeroUI - %s Configuration", m.currentApp)
	case ConfigEditView:
		titleText = fmt.Sprintf("⚙️ ZeroUI - %s Configuration", m.currentApp)
	case AppGridView:
		titleText = "🔧 ZeroUI - Applications"
	case PresetSelectionView:
		titleText = fmt.Sprintf("🎨 ZeroUI - %s Presets", m.currentApp)
	case HelpView:
		titleText = "❓ ZeroUI - Help"
	default:
		titleText = "🔧 ZeroUI"
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

// renderPresetSelection renders the preset selection view
func (m *Model) renderPresetSelection() string {
	title := fmt.Sprintf("Select Preset for %s", m.currentApp)
	
	// Get available presets for the current app
	presets, err := m.engine.GetPresets(m.currentApp)
	if err != nil {
		return m.styles.Error.Render(fmt.Sprintf("Error loading presets: %v", err))
	}
	
	if len(presets) == 0 {
		return m.styles.Muted.Render(fmt.Sprintf("No presets available for %s", m.currentApp))
	}
	
	var content strings.Builder
	content.WriteString(m.styles.Title.Render(title) + "\n\n")
	
	for i, preset := range presets {
		prefix := "  "
		if i == 0 { // Simple selection - first item is selected for now
			prefix = "> "
		}
		content.WriteString(fmt.Sprintf("%s%s\n", prefix, preset))
	}
	
	content.WriteString("\n" + m.styles.Muted.Render("Press Enter to apply, q to go back"))
	
	return content.String()
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
