package appcomponents

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/tui/keys"
	"github.com/mrtkrcm/ZeroUI/internal/tui/layout"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// AppSelectedMsg is sent when an app is selected
type AppSelectedMsg struct {
	App string
}

// AppSelectorModel represents the app selection component
type AppSelectorModel struct {
	width  int
	height int

	apps    []string
	cursor  int
	focused bool
	keyMap  keys.AppKeyMap
	styles  *styles.Styles
}

// NewAppSelector creates a new app selector component
func NewAppSelector(apps []string) *AppSelectorModel {
	return &AppSelectorModel{
		apps:   apps,
		cursor: 0,
		keyMap: keys.DefaultKeyMap(),
		styles: styles.GetStyles(),
	}
}

// Init implements tea.Model
func (m *AppSelectorModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *AppSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.cursor < len(m.apps)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keyMap.Enter, m.keyMap.Space):
			if len(m.apps) > 0 {
				return m, func() tea.Msg {
					return AppSelectedMsg{App: m.apps[m.cursor]}
				}
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *AppSelectorModel) View() string {
	if len(m.apps) == 0 {
		return m.styles.Muted.Render("No applications configured.\nAdd app configurations to ~/.config/zeroui/apps/")
	}

	var b strings.Builder

	for i, app := range m.apps {
		var line string
		cursor := " "

		if i == m.cursor {
			cursor = ">"
			if m.focused {
				line = m.styles.Selected.Render(fmt.Sprintf("%s %s", cursor, app))
			} else {
				line = m.styles.Text.Render(fmt.Sprintf("%s %s", cursor, app))
			}
		} else {
			line = m.styles.Text.Render(fmt.Sprintf("%s %s", cursor, app))
		}

		b.WriteString(line)
		if i < len(m.apps)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Focus implements layout.Focusable
func (m *AppSelectorModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur implements layout.Focusable
func (m *AppSelectorModel) Blur() tea.Cmd {
	m.focused = false
	return nil
}

// IsFocused implements layout.Focusable
func (m *AppSelectorModel) IsFocused() bool {
	return m.focused
}

// SetSize implements layout.Sizeable
func (m *AppSelectorModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable
func (m *AppSelectorModel) GetSize() (int, int) {
	return m.width, m.height
}

// Bindings implements layout.Help
func (m *AppSelectorModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Enter,
	}
}

// SetApps updates the list of applications
func (m *AppSelectorModel) SetApps(apps []string) {
	m.apps = apps
	if m.cursor >= len(apps) {
		m.cursor = len(apps) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// GetSelectedApp returns the currently selected app
func (m *AppSelectorModel) GetSelectedApp() string {
	if len(m.apps) == 0 || m.cursor < 0 || m.cursor >= len(m.apps) {
		return ""
	}
	return m.apps[m.cursor]
}

// GetApps returns all available apps
func (m *AppSelectorModel) GetApps() []string {
	return m.apps
}

// Ensure AppSelectorModel implements the required interfaces
var _ util.Model = (*AppSelectorModel)(nil)
var _ layout.Focusable = (*AppSelectorModel)(nil)
var _ layout.Sizeable = (*AppSelectorModel)(nil)
var _ layout.Help = (*AppSelectorModel)(nil)
