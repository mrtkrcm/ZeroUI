package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ValidationFunc represents a validation function
type ValidationFunc func(string) error

// FormField represents a form field with validation
type FormField struct {
	Input       textinput.Model
	Label       string
	Placeholder string
	Required    bool
	Validators  []ValidationFunc
	Error       string
	Valid       bool
}

// EnhancedFormModel represents a form with validation and styling
type EnhancedFormModel struct {
	fields      []*FormField
	focusedIdx  int
	styles      *styles.Styles
	width       int
	height      int
	title       string
	submitLabel string
	cancelLabel string
	showHelp    bool
}

// FormSubmitMsg represents form submission
type FormSubmitMsg struct {
	Values map[string]string
}

// FormCancelMsg represents form cancellation
type FormCancelMsg struct{}

// NewEnhancedForm creates a new enhanced form
func NewEnhancedForm(title string) *EnhancedFormModel {
	return &EnhancedFormModel{
		fields:      []*FormField{},
		styles:      styles.GetStyles(),
		title:       title,
		submitLabel: "Submit",
		cancelLabel: "Cancel",
		showHelp:    true,
	}
}

// AddField adds a new field to the form
func (m *EnhancedFormModel) AddField(label, placeholder string, required bool, validators ...ValidationFunc) {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 200
	input.Width = 50

	// Style the input
	input.Prompt = "│ "
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	field := &FormField{
		Input:       input,
		Label:       label,
		Placeholder: placeholder,
		Required:    required,
		Validators:  validators,
		Valid:       !required, // Non-required fields are valid by default
	}

	m.fields = append(m.fields, field)

	// Focus first field
	if len(m.fields) == 1 {
		field.Input.Focus()
		m.focusedIdx = 0
	}
}

// Init implements tea.Model
func (m *EnhancedFormModel) Init() tea.Cmd {
	if len(m.fields) > 0 {
		return m.fields[0].Input.Cursor.BlinkCmd()
	}
	return nil
}

// Update implements tea.Model
func (m *EnhancedFormModel) Update(msg tea.Msg) (*EnhancedFormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update input widths
		inputWidth := msg.Width - 20
		if inputWidth < 20 {
			inputWidth = 20
		}

		for _, field := range m.fields {
			field.Input.Width = inputWidth
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return m, m.focusNext()
		case "shift+tab", "up":
			return m, m.focusPrev()
		case "enter":
			if m.isValid() {
				return m, func() tea.Msg {
					return FormSubmitMsg{Values: m.getValues()}
				}
			}
			// If not valid, focus first invalid field
			m.focusFirstInvalid()
			return m, nil
		case "esc":
			return m, func() tea.Msg {
				return FormCancelMsg{}
			}
		case "ctrl+c":
			return m, tea.Quit
		}

		// Update focused field
		if m.focusedIdx >= 0 && m.focusedIdx < len(m.fields) {
			field := m.fields[m.focusedIdx]
			var cmd tea.Cmd
			field.Input, cmd = field.Input.Update(msg)

			// Validate field on change
			m.validateField(m.focusedIdx)

			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *EnhancedFormModel) View() string {
	var sections []string

	// Title
	if m.title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(1, 0).
			Width(m.width).
			Align(lipgloss.Center)
		sections = append(sections, titleStyle.Render(m.title))
		sections = append(sections, "")
	}

	// Form fields
	for i, field := range m.fields {
		sections = append(sections, m.renderField(field, i == m.focusedIdx))
		sections = append(sections, "")
	}

	// Form status
	if !m.isValid() {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Italic(true).
			Padding(0, 2)
		sections = append(sections, errorStyle.Render("⚠ Please fix validation errors before submitting"))
		sections = append(sections, "")
	}

	// Help text
	if m.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2)

		helpText := []string{
			"Tab/↓: Next field",
			"Shift+Tab/↑: Previous field",
			"Enter: Submit",
			"Esc: Cancel",
		}
		sections = append(sections, helpStyle.Render(strings.Join(helpText, " • ")))
	}

	// Container
	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Width(m.width - 4)

	return container.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// renderField renders a single form field
func (m *EnhancedFormModel) renderField(field *FormField, focused bool) string {
	var parts []string

	// Label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if focused {
		labelStyle = labelStyle.Foreground(lipgloss.Color("212"))
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("255"))
	}

	label := field.Label
	if field.Required {
		label += " *"
	}
	parts = append(parts, labelStyle.Render(label))

	// Input
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Padding(0, 1)

	if focused {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("212"))
	} else if !field.Valid {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("196"))
	} else {
		inputStyle = inputStyle.BorderForeground(lipgloss.Color("238"))
	}

	parts = append(parts, inputStyle.Render(field.Input.View()))

	// Error message
	if !field.Valid && field.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Italic(true).
			Padding(0, 1)
		parts = append(parts, errorStyle.Render("↳ "+field.Error))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// focusNext focuses the next field
func (m *EnhancedFormModel) focusNext() tea.Cmd {
	if len(m.fields) == 0 {
		return nil
	}

	// Blur current field
	m.fields[m.focusedIdx].Input.Blur()

	// Move to next field
	m.focusedIdx = (m.focusedIdx + 1) % len(m.fields)

	// Focus new field
	m.fields[m.focusedIdx].Input.Focus()
	return m.fields[m.focusedIdx].Input.Cursor.BlinkCmd()
}

// focusPrev focuses the previous field
func (m *EnhancedFormModel) focusPrev() tea.Cmd {
	if len(m.fields) == 0 {
		return nil
	}

	// Blur current field
	m.fields[m.focusedIdx].Input.Blur()

	// Move to previous field
	m.focusedIdx = (m.focusedIdx - 1 + len(m.fields)) % len(m.fields)

	// Focus new field
	m.fields[m.focusedIdx].Input.Focus()
	return m.fields[m.focusedIdx].Input.Cursor.BlinkCmd()
}

// focusFirstInvalid focuses the first invalid field
func (m *EnhancedFormModel) focusFirstInvalid() tea.Cmd {
	for i, field := range m.fields {
		if !field.Valid {
			// Blur current field
			if m.focusedIdx >= 0 && m.focusedIdx < len(m.fields) {
				m.fields[m.focusedIdx].Input.Blur()
			}

			// Focus invalid field
			m.focusedIdx = i
			field.Input.Focus()
			return field.Input.Cursor.BlinkCmd()
		}
	}
	return nil
}

// validateField validates a specific field
func (m *EnhancedFormModel) validateField(idx int) {
	if idx < 0 || idx >= len(m.fields) {
		return
	}

	field := m.fields[idx]
	value := field.Input.Value()

	// Check if required field is empty
	if field.Required && strings.TrimSpace(value) == "" {
		field.Valid = false
		field.Error = "This field is required"
		return
	}

	// Run custom validators
	for _, validator := range field.Validators {
		if err := validator(value); err != nil {
			field.Valid = false
			field.Error = err.Error()
			return
		}
	}

	// Field is valid
	field.Valid = true
	field.Error = ""
}

// isValid checks if all fields are valid
func (m *EnhancedFormModel) isValid() bool {
	// Validate all fields first
	for i := range m.fields {
		m.validateField(i)
	}

	// Check if all fields are valid
	for _, field := range m.fields {
		if !field.Valid {
			return false
		}
	}
	return true
}

// getValues returns all field values
func (m *EnhancedFormModel) getValues() map[string]string {
	values := make(map[string]string)
	for _, field := range m.fields {
		values[field.Label] = field.Input.Value()
	}
	return values
}

// SetValue sets a field value by label
func (m *EnhancedFormModel) SetValue(label, value string) {
	for _, field := range m.fields {
		if field.Label == label {
			field.Input.SetValue(value)
			break
		}
	}
}

// Common validators
func ValidateNotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("field cannot be empty")
	}
	return nil
}

func ValidateMinLength(min int) ValidationFunc {
	return func(value string) error {
		if len(strings.TrimSpace(value)) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

func ValidateMaxLength(max int) ValidationFunc {
	return func(value string) error {
		if len(value) > max {
			return fmt.Errorf("cannot exceed %d characters", max)
		}
		return nil
	}
}
