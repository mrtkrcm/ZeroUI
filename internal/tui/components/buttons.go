package components

import "github.com/charmbracelet/lipgloss"

// ButtonStyle defines the interface for button styling
type ButtonStyle interface {
	Primary() lipgloss.Style
	Secondary() lipgloss.Style
	Disabled() lipgloss.Style
}

// DefaultButtonStyle implements the default button styles
type DefaultButtonStyle struct{}

// NewDefaultButtonStyle creates a new default button style
func NewDefaultButtonStyle() *DefaultButtonStyle {
	return &DefaultButtonStyle{}
}

// Primary button style
func (b *DefaultButtonStyle) Primary() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1).
		Bold(true)
}

// Secondary button style
func (b *DefaultButtonStyle) Secondary() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
}

// Disabled button style
func (b *DefaultButtonStyle) Disabled() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#6272A4")).
		Padding(0, 1)
}

// Global button styles for direct use
var (
	ButtonPrimaryStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1).
		Bold(true)

	ButtonSecondaryStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	ButtonDisabledStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#6272A4")).
		Padding(0, 1)
)

// ButtonComponent represents a button with state
type ButtonComponent struct {
	Text     string
	Style    lipgloss.Style
	Disabled bool
	Active   bool
}

// NewButtonComponent creates a new button component
func NewButtonComponent(text string, variant string) *ButtonComponent {
	var style lipgloss.Style
	
	switch variant {
	case "primary":
		style = ButtonPrimaryStyle
	case "secondary":
		style = ButtonSecondaryStyle
	case "disabled":
		style = ButtonDisabledStyle
	default:
		style = ButtonPrimaryStyle
	}

	return &ButtonComponent{
		Text:     text,
		Style:    style,
		Disabled: variant == "disabled",
	}
}

// Render renders the button
func (b *ButtonComponent) Render() string {
	if b.Disabled {
		return ButtonDisabledStyle.Render(b.Text)
	}
	
	if b.Active {
		// Add visual feedback for active state
		activeStyle := b.Style.Copy().
			Background(lipgloss.Color("#9575FF"))
		return activeStyle.Render(b.Text)
	}
	
	return b.Style.Render(b.Text)
}

// SetActive sets the button active state
func (b *ButtonComponent) SetActive(active bool) {
	b.Active = active
}

// SetDisabled sets the button disabled state
func (b *ButtonComponent) SetDisabled(disabled bool) {
	b.Disabled = disabled
}

// GetButtonDemo returns a demonstration of button states
func GetButtonDemo() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		ButtonPrimaryStyle.Render(" Primary "),
		"  ",
		ButtonSecondaryStyle.Render("Secondary"),
		"  ",
		ButtonDisabledStyle.Render(" Disabled "),
	)
}

// GetButtonsExample renders a complete button examples section
func GetButtonsExample() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true).
			Render("🔘 Button States"),
		"",
		GetButtonDemo(),
	)
}