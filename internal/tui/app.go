package tui

// TODO: Add field configuration caching to improve TUI performance (Week 2)
// TODO: Linear search on every TUI update causes UI lag with 100+ fields
// TODO: Implement map[string]FieldView cache with invalidation on config change

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// App represents the TUI application
type App struct {
	engine      *toggle.Engine
	initialApp  string
	program     *tea.Program
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
	model, err := NewModel(a.engine, a.initialApp)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	a.program = tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("TUI application error: %w", err)
	}

	return nil
}

// Model represents the application state
type Model struct {
	engine     *toggle.Engine
	state      ViewState
	apps       []string
	currentApp string
	appConfigs map[string]*AppConfigView
	cursor     int
	width      int
	height     int
	err        error
}

// ViewState represents the current view state
type ViewState int

const (
	AppSelectionView ViewState = iota
	ConfigEditView
	PresetSelectionView
	HelpView
)

// AppConfigView holds the view data for an application configuration
type AppConfigView struct {
	Name        string
	Fields      []FieldView
	Presets     []PresetView
	cursor      int
	editMode    bool
	fieldCursor int
}

// FieldView represents a configuration field in the TUI
type FieldView struct {
	Key         string
	Type        string
	CurrentValue string
	Values      []string
	Description string
	cursor      int
}

// PresetView represents a preset in the TUI
type PresetView struct {
	Name        string
	Description string
	Values      map[string]interface{}
}

// NewModel creates a new model
func NewModel(engine *toggle.Engine, initialApp string) (*Model, error) {
	apps, err := engine.GetApps()
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	model := &Model{
		engine:     engine,
		state:      AppSelectionView,
		apps:       apps,
		currentApp: initialApp,
		appConfigs: make(map[string]*AppConfigView),
	}

	// If initial app is specified and exists, start with config view
	if initialApp != "" {
		for _, app := range apps {
			if app == initialApp {
				model.state = ConfigEditView
				if err := model.loadAppConfig(initialApp); err != nil {
					model.err = err
				}
				break
			}
		}
	}

	return model, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == AppSelectionView {
			return m, tea.Quit
		}
		// Go back to app selection from other views
		m.state = AppSelectionView
		return m, nil
		
	case "?":
		m.state = HelpView
		return m, nil
		
	case "esc":
		switch m.state {
		case ConfigEditView, PresetSelectionView, HelpView:
			m.state = AppSelectionView
		default:
			return m, tea.Quit
		}
		return m, nil
	}

	// Handle state-specific key presses
	switch m.state {
	case AppSelectionView:
		return m.handleAppSelectionKeys(msg)
	case ConfigEditView:
		return m.handleConfigEditKeys(msg)
	case PresetSelectionView:
		return m.handlePresetSelectionKeys(msg)
	case HelpView:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleAppSelectionKeys handles keys in app selection view
func (m *Model) handleAppSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.apps)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.apps) > 0 {
			m.currentApp = m.apps[m.cursor]
			m.state = ConfigEditView
			if err := m.loadAppConfig(m.currentApp); err != nil {
				m.err = err
			}
		}
	}
	return m, nil
}

// handleConfigEditKeys handles keys in config edit view
func (m *Model) handleConfigEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.currentApp == "" || m.appConfigs[m.currentApp] == nil {
		return m, nil
	}

	appConfig := m.appConfigs[m.currentApp]

	switch msg.String() {
	case "up", "k":
		if appConfig.cursor > 0 {
			appConfig.cursor--
		}
	case "down", "j":
		if appConfig.cursor < len(appConfig.Fields)-1 {
			appConfig.cursor++
		}
	case "left", "h":
		if appConfig.cursor < len(appConfig.Fields) {
			field := &appConfig.Fields[appConfig.cursor]
			if len(field.Values) > 0 && field.cursor > 0 {
				field.cursor--
				if err := m.applyFieldChange(field); err != nil {
					m.err = err
				}
			}
		}
	case "right", "l":
		if appConfig.cursor < len(appConfig.Fields) {
			field := &appConfig.Fields[appConfig.cursor]
			if len(field.Values) > 0 && field.cursor < len(field.Values)-1 {
				field.cursor++
				if err := m.applyFieldChange(field); err != nil {
					m.err = err
				}
			}
		}
	case "enter", " ":
		if appConfig.cursor < len(appConfig.Fields) {
			field := &appConfig.Fields[appConfig.cursor]
			if len(field.Values) > 0 {
				// Cycle to next value
				field.cursor = (field.cursor + 1) % len(field.Values)
				if err := m.applyFieldChange(field); err != nil {
					m.err = err
				}
			}
		}
	case "p":
		m.state = PresetSelectionView
	}

	return m, nil
}

// handlePresetSelectionKeys handles keys in preset selection view
func (m *Model) handlePresetSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.currentApp == "" || m.appConfigs[m.currentApp] == nil {
		return m, nil
	}

	appConfig := m.appConfigs[m.currentApp]

	switch msg.String() {
	case "up", "k":
		if appConfig.cursor > 0 {
			appConfig.cursor--
		}
	case "down", "j":
		if appConfig.cursor < len(appConfig.Presets)-1 {
			appConfig.cursor++
		}
	case "enter", " ":
		if appConfig.cursor < len(appConfig.Presets) {
			preset := appConfig.Presets[appConfig.cursor]
			if err := m.engine.ApplyPreset(m.currentApp, preset.Name); err != nil {
				m.err = err
			} else {
				// Reload the config to show updated values
				if err := m.loadAppConfig(m.currentApp); err != nil {
					m.err = err
				}
				m.state = ConfigEditView
			}
		}
	}

	return m, nil
}

// handleHelpKeys handles keys in help view
func (m *Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "esc":
		m.state = AppSelectionView
	}
	return m, nil
}

// applyFieldChange applies a field value change
func (m *Model) applyFieldChange(field *FieldView) error {
	if len(field.Values) == 0 || field.cursor >= len(field.Values) {
		return nil
	}

	newValue := field.Values[field.cursor]
	if err := m.engine.Toggle(m.currentApp, field.Key, newValue); err != nil {
		return err
	}

	field.CurrentValue = newValue
	return nil
}

// View renders the current view
func (m *Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	switch m.state {
	case AppSelectionView:
		return m.renderAppSelection()
	case ConfigEditView:
		return m.renderConfigEdit()
	case PresetSelectionView:
		return m.renderPresetSelection()
	case HelpView:
		return m.renderHelp()
	}

	return ""
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
)

// renderError renders error messages
func (m *Model) renderError() string {
	return fmt.Sprintf("Error: %s\n\nPress 'q' to quit or any key to continue.", m.err.Error())
}

// renderAppSelection renders the application selection view
func (m *Model) renderAppSelection() string {
	title := titleStyle.Render("ZeroUI - Select Application")
	content := "\n"

	if len(m.apps) == 0 {
		content += "No applications configured.\n"
		content += "Add app configurations to ~/.config/zeroui/apps/\n"
	} else {
		for i, app := range m.apps {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				app = selectedStyle.Render(app)
			}
			content += fmt.Sprintf("%s %s\n", cursor, app)
		}
	}

	help := "\n" + helpStyle.Render("↑/↓: navigate • enter: select • ?: help • q: quit")

	return title + content + help
}

// renderConfigEdit renders the configuration editing view
func (m *Model) renderConfigEdit() string {
	if m.currentApp == "" || m.appConfigs[m.currentApp] == nil {
		return "No application selected"
	}

	appConfig := m.appConfigs[m.currentApp]
	title := titleStyle.Render(fmt.Sprintf("ZeroUI - %s", m.currentApp))
	content := "\n"

	for i, field := range appConfig.Fields {
		cursor := " "
		if appConfig.cursor == i {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %s: %s", cursor, field.Key, field.CurrentValue)
		if appConfig.cursor == i {
			line = selectedStyle.Render(line)
		}

		if len(field.Values) > 0 && appConfig.cursor == i {
			line += " " + helpStyle.Render(fmt.Sprintf("[%s]", fmt.Sprintf("%v", field.Values)))
		}

		content += line + "\n"
	}

	help := "\n" + helpStyle.Render("↑/↓: navigate • ←/→: change value • enter/space: cycle • p: presets • esc: back • q: quit")

	return title + content + help
}

// renderPresetSelection renders the preset selection view
func (m *Model) renderPresetSelection() string {
	if m.currentApp == "" || m.appConfigs[m.currentApp] == nil {
		return "No application selected"
	}

	appConfig := m.appConfigs[m.currentApp]
	title := titleStyle.Render(fmt.Sprintf("ZeroUI - %s Presets", m.currentApp))
	content := "\n"

	if len(appConfig.Presets) == 0 {
		content += "No presets configured for this application.\n"
	} else {
		for i, preset := range appConfig.Presets {
			cursor := " "
			name := preset.Name
			if appConfig.cursor == i {
				cursor = ">"
				name = selectedStyle.Render(name)
			}
			
			line := fmt.Sprintf("%s %s", cursor, name)
			if preset.Description != "" {
				line += " - " + preset.Description
			}
			content += line + "\n"
		}
	}

	help := "\n" + helpStyle.Render("↑/↓: navigate • enter: apply preset • esc: back")

	return title + content + help
}

// renderHelp renders the help view
func (m *Model) renderHelp() string {
	title := titleStyle.Render("ZeroUI - Help")
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
  p             Open presets
  esc           Back to app selection

Preset Selection:
  ↑/↓ or k/j    Navigate presets
  enter/space   Apply preset
  esc           Back to config edit

Global:
  ?             Show this help
  esc           Go back/quit
  ctrl+c        Force quit

`

	help := helpStyle.Render("Press any key to go back")
	return title + content + "\n" + help
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

	// Convert fields to TUI format
	var fields []FieldView
	for key, fieldConfig := range appConfig.Fields {
		currentValue := ""
		if val, exists := currentValues[key]; exists {
			currentValue = fmt.Sprintf("%v", val)
		} else if fieldConfig.Default != nil {
			currentValue = fmt.Sprintf("%v", fieldConfig.Default)
		}

		// Find cursor position for current value
		cursor := 0
		if len(fieldConfig.Values) > 0 {
			for i, value := range fieldConfig.Values {
				if value == currentValue {
					cursor = i
					break
				}
			}
		}

		fields = append(fields, FieldView{
			Key:          key,
			Type:         fieldConfig.Type,
			CurrentValue: currentValue,
			Values:       fieldConfig.Values,
			Description:  fieldConfig.Description,
			cursor:       cursor,
		})
	}

	// Convert presets to TUI format
	var presets []PresetView
	for name, presetConfig := range appConfig.Presets {
		presets = append(presets, PresetView{
			Name:        name,
			Description: presetConfig.Description,
			Values:      presetConfig.Values,
		})
	}

	m.appConfigs[appName] = &AppConfigView{
		Name:    appName,
		Fields:  fields,
		Presets: presets,
	}

	return nil
}