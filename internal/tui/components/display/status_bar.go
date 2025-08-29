package display

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/layout"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// StatusBarModel represents a status bar component
type StatusBarModel struct {
	width  int
	height int

	// Status information
	status   string
	appCount int
	theme    string
	message  string
	msgType  util.InfoType

	styles *styles.Styles
}

// NewStatusBar creates a new status bar component
func NewStatusBar() *StatusBarModel {
	return &StatusBarModel{
		height: 1,
		styles: styles.GetStyles(),
		theme:  "Default",
	}
}

// Init implements tea.Model
func (m *StatusBarModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *StatusBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case util.InfoMsg:
		m.message = msg.Msg
		m.msgType = msg.Type
	}
	return m, nil
}

// View implements tea.Model
func (m *StatusBarModel) View() string {
	if m.width == 0 {
		return ""
	}

	// Create status sections
	left := m.renderLeftSection()
	right := m.renderRightSection()

	// Calculate spacing
	usedWidth := lipgloss.Width(left) + lipgloss.Width(right)
	spacing := m.width - usedWidth

	if spacing < 1 {
		spacing = 1
	}

	spacer := strings.Repeat(" ", spacing)

	statusLine := lipgloss.JoinHorizontal(lipgloss.Left, left, spacer, right)

	// Style the status bar
	style := m.styles.Base.
		Width(m.width).
		Background(lipgloss.Color(styles.ColorToHex(styles.GetTheme().BgSubtle))).
		Foreground(lipgloss.Color(styles.ColorToHex(styles.GetTheme().FgMuted)))

	return style.Render(statusLine)
}

// renderLeftSection renders the left side of the status bar
func (m *StatusBarModel) renderLeftSection() string {
	var parts []string

	if m.status != "" {
		parts = append(parts, fmt.Sprintf("Status: %s", m.status))
	} else {
		parts = append(parts, "Status: Ready")
	}

	if m.appCount > 0 {
		parts = append(parts, fmt.Sprintf("Apps: %d", m.appCount))
	}

	if m.theme != "" {
		parts = append(parts, fmt.Sprintf("Theme: %s", m.theme))
	}

	return strings.Join(parts, " • ")
}

// renderRightSection renders the right side of the status bar
func (m *StatusBarModel) renderRightSection() string {
	if m.message != "" {
		var style lipgloss.Style
		switch m.msgType {
		case util.InfoTypeError:
			style = m.styles.Error
		case util.InfoTypeWarn:
			style = m.styles.Warning
		case util.InfoTypeInfo:
			style = m.styles.Info
		default:
			style = m.styles.Text
		}
		return style.Render(m.message)
	}

	return ""
}

// SetSize implements layout.Sizeable
func (m *StatusBarModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable
func (m *StatusBarModel) GetSize() (int, int) {
	return m.width, m.height
}

// SetStatus updates the status text
func (m *StatusBarModel) SetStatus(status string) {
	m.status = status
}

// SetAppCount updates the application count
func (m *StatusBarModel) SetAppCount(count int) {
	m.appCount = count
}

// SetTheme updates the theme name
func (m *StatusBarModel) SetTheme(theme string) {
	m.theme = theme
}

// SetMessage sets a status message
func (m *StatusBarModel) SetMessage(message string, msgType util.InfoType) {
	m.message = message
	m.msgType = msgType
}

// ClearMessage clears the status message
func (m *StatusBarModel) ClearMessage() {
	m.message = ""
}

// Ensure StatusBarModel implements the required interfaces
var _ util.Model = (*StatusBarModel)(nil)
var _ layout.Sizeable = (*StatusBarModel)(nil)
