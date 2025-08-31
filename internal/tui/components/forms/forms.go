package forms

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// InputStyle defines the interface for input styling
type InputStyle interface {
	Normal() lipgloss.Style
	Focused() lipgloss.Style
	Error() lipgloss.Style
	Success() lipgloss.Style
}

// DefaultInputStyle implements the default input styles
type DefaultInputStyle struct{}

// NewDefaultInputStyle creates a new default input style
func NewDefaultInputStyle() *DefaultInputStyle {
	return &DefaultInputStyle{}
}

// Normal input style
func (i *DefaultInputStyle) Normal() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Foreground(lipgloss.Color("#F8F8F2"))
}

// Focused input style
func (i *DefaultInputStyle) Focused() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2"))
}

// Error input style
func (i *DefaultInputStyle) Error() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FF5555")).
		Foreground(lipgloss.Color("#F8F8F2"))
}

// Success input style
func (i *DefaultInputStyle) Success() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#50FA7B")).
		Foreground(lipgloss.Color("#F8F8F2"))
}

// Global input styles for direct use
var (
	InputNormalStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#6272A4")).
				Foreground(lipgloss.Color("#F8F8F2"))

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#F8F8F2"))

	InputErrorStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#FF5555")).
			Foreground(lipgloss.Color("#F8F8F2"))

	InputSuccessStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#50FA7B")).
				Foreground(lipgloss.Color("#F8F8F2"))
)

// InputComponent represents an input field with validation
type InputComponent struct {
	Input      textinput.Model
	Label      string
	Error      string
	Success    string
	Required   bool
	Focused    bool
	HasError   bool
	HasSuccess bool
}

// NewInputComponent creates a new input component
func NewInputComponent(label, placeholder string, width int) *InputComponent {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Width = width

	return &InputComponent{
		Input: input,
		Label: label,
	}
}

// SetError sets an error message
func (i *InputComponent) SetError(err string) {
	i.Error = err
	i.HasError = true
	i.HasSuccess = false
}

// SetSuccess sets a success message
func (i *InputComponent) SetSuccess(msg string) {
	i.Success = msg
	i.HasSuccess = true
	i.HasError = false
}

// ClearValidation clears validation messages
func (i *InputComponent) ClearValidation() {
	i.Error = ""
	i.Success = ""
	i.HasError = false
	i.HasSuccess = false
}

// Focus sets focus on the input
func (i *InputComponent) Focus() {
	i.Focused = true
	i.Input.Focus()
}

// Blur removes focus from the input
func (i *InputComponent) Blur() {
	i.Focused = false
	i.Input.Blur()
}

// Render renders the input component
func (i *InputComponent) Render() string {
	var style lipgloss.Style

	if i.HasError {
		style = InputErrorStyle
	} else if i.HasSuccess {
		style = InputSuccessStyle
	} else if i.Focused {
		style = InputFocusedStyle
	} else {
		style = InputNormalStyle
	}

	// Render input with appropriate style
	input := style.Render(i.Input.View())

	// Add label if present
	if i.Label != "" {
		input = i.Label + "\n" + input
	}

	// Add validation message if present
	if i.HasError && i.Error != "" {
		input += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Render("❌ "+i.Error)
	} else if i.HasSuccess && i.Success != "" {
		input += "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Render("✓ "+i.Success)
	}

	return input
}

// GetInputDemo returns a demonstration of input states
func GetInputDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		InputNormalStyle.Render("│ Sample input text...          │"),
		InputFocusedStyle.Render("│ Focused input with cursor|    │"),
	)
}

// GetInputExample renders a complete input examples section
func GetInputExample() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true).
			Render("📝 Text Input"),
		"",
		GetInputDemo(),
	)
}

// ValidationState represents different validation states
type ValidationState int

const (
	ValidationNone ValidationState = iota
	ValidationValid
	ValidationInvalid
	ValidationPending
)

// GetValidationDemo returns a demonstration of validation states
func GetValidationDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Render("✓ Valid input"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Render("✗ Invalid format"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Render("⏳ Validating..."),
	)
}

// GetValidationExample renders a complete validation examples section
func GetValidationExample() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true).
			Render("🔍 Validation"),
		"",
		GetValidationDemo(),
	)
}

// GetButtonsDemo renders various button examples
func GetButtonsDemo() string {
	primaryBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Bold(true).
		Render("[ Primary ]")

	secondaryBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Render("[ Secondary ]")

	successBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#50FA7B")).
		Foreground(lipgloss.Color("#282A36")).
		Padding(0, 2).
		Bold(true).
		Render("[ Success ]")

	dangerBtn := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF5555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Render("[ Danger ]")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		primaryBtn,
		secondaryBtn,
		successBtn,
		dangerBtn,
	)
}

// GetButtonsExample renders a complete buttons examples section
func GetButtonsExample() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true).
			Render("🔘 Buttons"),
		"",
		GetButtonsDemo(),
	)
}
