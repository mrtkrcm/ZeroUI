package tui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
)

// ImprovedApp represents the improved TUI application using better patterns
type ImprovedApp struct {
	configService *service.ConfigService
	logger        *logger.Logger
	program       *tea.Program
}

// NewImprovedApp creates a new improved TUI application
func NewImprovedApp(configService *service.ConfigService, log *logger.Logger) *ImprovedApp {
	return &ImprovedApp{
		configService: configService,
		logger:        log,
	}
}

// Run starts the improved TUI application
func (a *ImprovedApp) Run() error {
	model := newImprovedModel(a.configService, a.logger)
	a.program = tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("TUI application error: %w", err)
	}

	return nil
}

// ImprovedModel represents the application state using better patterns
type ImprovedModel struct {
	// Services
	configService *service.ConfigService
	logger        *logger.Logger

	// UI Components
	appList     list.Model
	viewport    viewport.Model
	spinner     spinner.Model
	textInput   textinput.Model

	// State
	state       ViewState
	currentApp  string
	loading     bool
	error       error
	
	// Layout
	width  int
	height int
}

// newImprovedModel creates a new improved model
func newImprovedModel(configService *service.ConfigService, log *logger.Logger) *ImprovedModel {
	// Initialize components
	appList := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	appList.Title = "Applications"
	appList.SetShowStatusBar(false)
	appList.SetFilteringEnabled(false)

	vp := viewport.New(0, 0)
	
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Enter value..."

	return &ImprovedModel{
		configService: configService,
		logger:        log,
		appList:       appList,
		viewport:      vp,
		spinner:       s,
		textInput:     ti,
		state:         AppSelectionView,
	}
}

// Init initializes the model
func (m *ImprovedModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadApps(),
		m.spinner.Tick,
	)
}

// Update handles messages and updates the model
func (m *ImprovedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m.handleEscape()
		}

		// Handle state-specific keys
		switch m.state {
		case AppSelectionView:
			return m.handleAppSelectionKeys(msg)
		case ConfigEditView:
			return m.handleConfigEditKeys(msg)
		case HelpView:
			return m.handleHelpKeys(msg)
		}

	case appsLoadedMsg:
		return m.handleAppsLoaded(msg)
	
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case errorMsg:
		m.error = msg.err
		m.loading = false
		return m, nil
	}

	// Update components
	if m.state == AppSelectionView {
		var cmd tea.Cmd
		m.appList, cmd = m.appList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m *ImprovedModel) View() string {
	if m.error != nil {
		return m.renderError()
	}

	switch m.state {
	case AppSelectionView:
		return m.renderAppSelection()
	case ConfigEditView:
		return m.renderConfigEdit()
	case HelpView:
		return m.renderHelp()
	}

	return ""
}

// Custom messages
type appsLoadedMsg struct {
	apps []string
}

type errorMsg struct {
	err error
}

// Commands
func (m *ImprovedModel) loadApps() tea.Cmd {
	return func() tea.Msg {
		apps, err := m.configService.ListApplications()
		if err != nil {
			return errorMsg{err}
		}
		return appsLoadedMsg{apps}
	}
}

// Event handlers
func (m *ImprovedModel) handleAppsLoaded(msg appsLoadedMsg) (*ImprovedModel, tea.Cmd) {
	items := make([]list.Item, len(msg.apps))
	for i, app := range msg.apps {
		items[i] = appItem(app)
	}
	m.appList.SetItems(items)
	m.loading = false
	return m, nil
}

func (m *ImprovedModel) handleEscape() (*ImprovedModel, tea.Cmd) {
	switch m.state {
	case ConfigEditView, HelpView:
		m.state = AppSelectionView
		return m, nil
	default:
		return m, tea.Quit
	}
}

func (m *ImprovedModel) handleAppSelectionKeys(msg tea.KeyMsg) (*ImprovedModel, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		if selectedItem := m.appList.SelectedItem(); selectedItem != nil {
			if app, ok := selectedItem.(appItem); ok {
				m.currentApp = string(app)
				m.state = ConfigEditView
				return m, m.loadAppConfig()
			}
		}
	case "?":
		m.state = HelpView
		return m, nil
	}
	return m, nil
}

func (m *ImprovedModel) handleConfigEditKeys(msg tea.KeyMsg) (*ImprovedModel, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.state = HelpView
		return m, nil
	}
	return m, nil
}

func (m *ImprovedModel) handleHelpKeys(msg tea.KeyMsg) (*ImprovedModel, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "esc":
		m.state = AppSelectionView
		return m, nil
	}
	return m, nil
}

func (m *ImprovedModel) loadAppConfig() tea.Cmd {
	return func() tea.Msg {
		// This would load the app configuration
		// For now, just return nil
		return nil
	}
}

// Layout management
func (m *ImprovedModel) updateLayout() {
	headerHeight := 3
	footerHeight := 3
	availableHeight := m.height - headerHeight - footerHeight

	m.appList.SetWidth(m.width)
	m.appList.SetHeight(availableHeight)
	
	m.viewport.Width = m.width
	m.viewport.Height = availableHeight
}

// Rendering methods
func (m *ImprovedModel) renderError() string {
	content := improvedErrorStyle.Render(fmt.Sprintf("Error: %s", m.error.Error()))
	help := improvedHelpStyle.Render("Press 'q' to quit or any key to continue.")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		improvedTitleStyle.Render("ZeroUI - Error"),
		"",
		content,
		"",
		help,
	)
}

func (m *ImprovedModel) renderAppSelection() string {
	if m.loading {
		content := lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.spinner.View(),
			" Loading applications...",
		)
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	title := improvedTitleStyle.Render("ZeroUI - Select Application")
	help := improvedHelpStyle.Render("↑/↓: navigate • enter: select • ?: help • q: quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		m.appList.View(),
		"",
		help,
	)
}

func (m *ImprovedModel) renderConfigEdit() string {
	title := improvedTitleStyle.Render(fmt.Sprintf("ZeroUI - %s", m.currentApp))
	content := "Configuration editing would be implemented here"
	help := improvedHelpStyle.Render("↑/↓: navigate • ←/→: change value • esc: back • ?: help")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content,
		"",
		help,
	)
}

func (m *ImprovedModel) renderHelp() string {
	title := improvedTitleStyle.Render("ZeroUI - Help")
	content := `
Key Bindings:

Application Selection:
  ↑/↓ or k/j    Navigate applications
  enter/space   Select application
  q             Quit

Configuration Edit:
  ↑/↓ or k/j    Navigate fields
  ←/→ or h/l    Change field value
  enter/space   Cycle to next value
  esc           Back to app selection

Global:
  ?             Show this help
  esc           Go back/quit
  ctrl+c        Force quit
`
	help := improvedHelpStyle.Render("Press any key to go back")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		"",
		help,
	)
}

// List item implementation
type appItem string

func (i appItem) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(appItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := listItemStyle.Render
	if index == m.Index() {
		fn = selectedItemStyle.Render
		str = "> " + str
	}

	fmt.Fprint(w, fn(str))
}

// Styles for the improved TUI
var (
	improvedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1).
				Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	listItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	improvedHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	improvedErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")).
				Bold(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))
)