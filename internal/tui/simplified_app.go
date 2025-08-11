package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
)

// SimpleApp is a streamlined TUI application
type SimpleApp struct {
	engine  *toggle.Engine
	model   tea.Model
	program *tea.Program
}

// NewSimpleApp creates a simplified TUI app
func NewSimpleApp(initialApp string) (*SimpleApp, error) {
	engine, err := toggle.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("engine init failed: %w", err)
	}

	model := NewSimpleModel(engine, initialApp)
	
	return &SimpleApp{
		engine: engine,
		model:  model,
	}, nil
}

// Run starts the TUI
func (a *SimpleApp) Run() error {
	a.program = tea.NewProgram(a.model, tea.WithAltScreen())
	_, err := a.program.Run()
	return err
}

// SimpleModel is a streamlined model with reduced complexity
type SimpleModel struct {
	engine   *toggle.Engine
	state    ViewState
	size     Size
	app      string
	grid     *components.AppGridModel
	editor   *components.ConfigEditorModel
	err      error
}

// Size holds terminal dimensions
type Size struct {
	Width  int
	Height int
}

// NewSimpleModel creates a simplified model
func NewSimpleModel(engine *toggle.Engine, app string) *SimpleModel {
	return &SimpleModel{
		engine: engine,
		state:  AppGridView,
		app:    app,
		grid:   components.NewAppGrid(),
		editor: components.NewConfigEditor(""),
	}
}

// Init initializes the model
func (m *SimpleModel) Init() tea.Cmd {
	return tea.Batch(
		m.grid.Init(),
		m.editor.Init(),
	)
}

// Update handles messages efficiently
func (m *SimpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	
	case tea.KeyMsg:
		return m.handleKey(msg)
	
	case components.AppSelectedMsg:
		return m.handleAppSelect(msg)
	
	default:
		return m.updateCurrentView(msg)
	}
}

// handleResize updates component sizes
func (m *SimpleModel) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.size = Size{msg.Width, msg.Height}
	
	// Update component sizes
	cmds := []tea.Cmd{
		m.grid.SetSize(msg.Width, msg.Height-4), // Reserve space for chrome
		m.editor.SetSize(msg.Width, msg.Height-4),
	}
	
	return m, tea.Batch(cmds...)
}

// handleKey processes keyboard input
func (m *SimpleModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	
	case "esc":
		if m.state == ConfigEditView {
			m.state = AppGridView
			return m, nil
		}
		return m, tea.Quit
	
	case "l":
		// Toggle between grid and list view
		if m.state == AppGridView {
			m.state = HuhAppSelectionView
		} else if m.state == HuhAppSelectionView {
			m.state = AppGridView
		}
		return m, nil
	
	default:
		return m.updateCurrentView(msg)
	}
}

// handleAppSelect processes app selection
func (m *SimpleModel) handleAppSelect(msg components.AppSelectedMsg) (tea.Model, tea.Cmd) {
	m.app = msg.App
	m.state = ConfigEditView
	
	// Load config asynchronously
	return m, func() tea.Msg {
		config, err := m.engine.GetAppConfig(msg.App)
		if err != nil {
			return ErrorMsg{err}
		}
		return ConfigLoadedMsg{config}
	}
}

// updateCurrentView updates the active view component
func (m *SimpleModel) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case AppGridView:
		grid, cmd := m.grid.Update(msg)
		m.grid = grid
		return m, cmd
	
	case ConfigEditView:
		editor, cmd := m.editor.Update(msg)
		m.editor = editor.(*components.ConfigEditorModel)
		return m, cmd
	
	default:
		return m, nil
	}
}

// View renders the current view
func (m *SimpleModel) View() string {
	if m.err != nil {
		return renderError(m.err)
	}
	
	var content string
	switch m.state {
	case AppGridView:
		content = m.grid.View()
	case ConfigEditView:
		content = m.editor.View()
	default:
		content = "Unknown view"
	}
	
	// Simple layout without complex chrome
	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderTitle(m.state, m.app),
		content,
		renderStatus(m.size),
	)
}

// Helper functions (extracted for clarity)

func renderTitle(state ViewState, app string) string {
	title := "ZeroUI"
	if state == ConfigEditView && app != "" {
		title = fmt.Sprintf("ZeroUI - %s", app)
	}
	
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Width(80).
		Align(lipgloss.Center).
		Render(title)
}

func renderStatus(size Size) string {
	status := fmt.Sprintf("%dx%d | Press q to quit | l to toggle view", 
		size.Width, size.Height)
	
	return lipgloss.NewStyle().
		Faint(true).
		Width(size.Width).
		Align(lipgloss.Center).
		Render(status)
}

func renderError(err error) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		Render(fmt.Sprintf("Error: %v", err))
}

// Message types (simplified)

type ErrorMsg struct {
	error
}

type ConfigLoadedMsg struct {
	Config interface{}
}