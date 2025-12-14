package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ConfirmDialog represents a modal confirmation dialog
type ConfirmDialog struct {
	title     string
	message   string
	visible   bool
	focused   int // 0=confirm, 1=cancel
	onConfirm func() tea.Cmd
	onCancel  func() tea.Cmd
	width     int
	height    int
	styles    *styles.Styles
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title, message string, onConfirm, onCancel func() tea.Cmd) *ConfirmDialog {
	return &ConfirmDialog{
		title:     title,
		message:   message,
		visible:   false,
		focused:   0, // Default to confirm button
		onConfirm: onConfirm,
		onCancel:  onCancel,
		width:     50,
		height:    8,
		styles:    nil, // Will be set when shown
	}
}

// Show displays the confirmation dialog
func (d *ConfirmDialog) Show() {
	d.visible = true
	d.focused = 0 // Reset focus to confirm
}

// Hide hides the confirmation dialog
func (d *ConfirmDialog) Hide() {
	d.visible = false
}

// IsVisible returns whether the dialog is currently visible
func (d *ConfirmDialog) IsVisible() bool {
	return d.visible
}

// SetSize sets the dialog size
func (d *ConfirmDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// SetStyles sets the dialog styles
func (d *ConfirmDialog) SetStyles(s *styles.Styles) {
	d.styles = s
}

// Update handles dialog updates
func (d *ConfirmDialog) Update(msg tea.Msg) (*ConfirmDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			d.Hide()
			if d.focused == 0 && d.onConfirm != nil {
				return d, d.onConfirm()
			} else if d.focused == 1 && d.onCancel != nil {
				return d, d.onCancel()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "right", "l"))):
			d.focused = (d.focused + 1) % 2
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "left", "h"))):
			d.focused = (d.focused + 1) % 2
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			d.Hide()
			if d.onCancel != nil {
				return d, d.onCancel()
			}
		}
	}

	return d, nil
}

// View renders the confirmation dialog
func (d *ConfirmDialog) View() string {
	if !d.visible {
		return ""
	}

	// Dialog box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff5555")).
		Padding(1, 2).
		Width(d.width).
		Align(lipgloss.Center)

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#ff5555"))

	// Message style
	messageStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Padding(0, 0, 1, 0)

	// Button styles
	confirmStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	if d.focused == 0 {
		confirmStyle = confirmStyle.
			Background(lipgloss.Color("#ff5555")).
			Foreground(lipgloss.Color("#ffffff"))
	} else {
		cancelStyle = cancelStyle.
			Background(lipgloss.Color("#ff5555")).
			Foreground(lipgloss.Color("#ffffff"))
	}

	// Create buttons
	confirmBtn := confirmStyle.Render("[Confirm]")
	cancelBtn := cancelStyle.Render("[Cancel]")

	// Combine buttons
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirmBtn, cancelBtn)

	// Build dialog content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(d.title),
		"",
		messageStyle.Render(d.message),
		"",
		buttons,
	)

	return boxStyle.Render(content)
}
