package components

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

// FieldChangedMsg is sent when a field value changes
type FieldChangedMsg struct {
	Key   string
	Value string
}

// OpenPresetsMsg is sent when presets should be opened
type OpenPresetsMsg struct{}

// FieldModel represents a configuration field
type FieldModel struct {
	Key          string
	Type         string
	CurrentValue string
	Values       []string
	Description  string
	cursor       int
	valueLookup  map[string]int // O(1) lookup optimization
}

// NewField creates a new field model
func NewField(key, fieldType, currentValue string, values []string, description string) *FieldModel {
	// Build O(1) lookup map
	valueLookup := make(map[string]int)
	for i, value := range values {
		valueLookup[value] = i
	}

	// Find current cursor position
	cursor := 0
	if idx, exists := valueLookup[currentValue]; exists {
		cursor = idx
	}

	return &FieldModel{
		Key:          key,
		Type:         fieldType,
		CurrentValue: currentValue,
		Values:       values,
		Description:  description,
		cursor:       cursor,
		valueLookup:  valueLookup,
	}
}

// GetValueIndex returns the index of a value using O(1) lookup
func (f *FieldModel) GetValueIndex(value string) (int, bool) {
	if f.valueLookup == nil {
		// Fallback to linear search
		for i, v := range f.Values {
			if v == value {
				return i, true
			}
		}
		return 0, false
	}

	idx, exists := f.valueLookup[value]
	return idx, exists
}

// ConfigEditorModel represents the configuration editor component
type ConfigEditorModel struct {
	width  int
	height int

	appName string
	fields  []*FieldModel
	cursor  int
	focused bool
	keyMap  keys.AppKeyMap
	styles  *styles.Styles
}

// NewConfigEditor creates a new config editor component
func NewConfigEditor(appName string) *ConfigEditorModel {
	return &ConfigEditorModel{
		appName: appName,
		fields:  make([]*FieldModel, 0),
		keyMap:  keys.DefaultKeyMap(),
		styles:  styles.GetStyles(),
	}
}

// Init implements tea.Model
func (m *ConfigEditorModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ConfigEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.fields)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keyMap.Left):
			if m.cursor < len(m.fields) {
				field := m.fields[m.cursor]
				if len(field.Values) > 0 && field.cursor > 0 {
					field.cursor--
					newValue := field.Values[field.cursor]
					field.CurrentValue = newValue
					return m, func() tea.Msg {
						return FieldChangedMsg{Key: field.Key, Value: newValue}
					}
				}
			}
		case key.Matches(msg, m.keyMap.Right):
			if m.cursor < len(m.fields) {
				field := m.fields[m.cursor]
				if len(field.Values) > 0 && field.cursor < len(field.Values)-1 {
					field.cursor++
					newValue := field.Values[field.cursor]
					field.CurrentValue = newValue
					return m, func() tea.Msg {
						return FieldChangedMsg{Key: field.Key, Value: newValue}
					}
				}
			}
		case key.Matches(msg, m.keyMap.Enter, m.keyMap.Space, m.keyMap.Toggle, m.keyMap.Cycle):
			if m.cursor < len(m.fields) {
				field := m.fields[m.cursor]
				if len(field.Values) > 0 {
					// Cycle to next value
					field.cursor = (field.cursor + 1) % len(field.Values)
					newValue := field.Values[field.cursor]
					field.CurrentValue = newValue
					return m, func() tea.Msg {
						return FieldChangedMsg{Key: field.Key, Value: newValue}
					}
				}
			}
		case key.Matches(msg, m.keyMap.Presets):
			return m, func() tea.Msg {
				return OpenPresetsMsg{}
			}
		case key.Matches(msg, m.keyMap.Reset):
			// Reset current field to first value
			if m.cursor < len(m.fields) {
				field := m.fields[m.cursor]
				if len(field.Values) > 0 {
					field.cursor = 0
					newValue := field.Values[0]
					field.CurrentValue = newValue
					return m, func() tea.Msg {
						return FieldChangedMsg{Key: field.Key, Value: newValue}
					}
				}
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *ConfigEditorModel) View() string {
	if len(m.fields) == 0 {
		return m.styles.Muted.Render("No configuration fields available")
	}

	var b strings.Builder

	for i, field := range m.fields {
		var line string
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		// Build field display
		fieldDisplay := fmt.Sprintf("%s %s: %s", cursor, field.Key, field.CurrentValue)

		if i == m.cursor && m.focused {
			line = m.styles.Selected.Render(fieldDisplay)

			// Show available values for selected field
			if len(field.Values) > 1 {
				valueDisplay := fmt.Sprintf(" [%s]", strings.Join(field.Values, ", "))
				line += m.styles.Muted.Render(valueDisplay)
			}

			// Show description if available
			if field.Description != "" {
				line += "\n" + strings.Repeat(" ", 2) + m.styles.Info.Render(field.Description)
			}
		} else {
			line = m.styles.Text.Render(fieldDisplay)
		}

		b.WriteString(line)
		if i < len(m.fields)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Focus implements layout.Focusable
func (m *ConfigEditorModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur implements layout.Focusable
func (m *ConfigEditorModel) Blur() tea.Cmd {
	m.focused = false
	return nil
}

// IsFocused implements layout.Focusable
func (m *ConfigEditorModel) IsFocused() bool {
	return m.focused
}

// SetSize implements layout.Sizeable
func (m *ConfigEditorModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable
func (m *ConfigEditorModel) GetSize() (int, int) {
	return m.width, m.height
}

// Bindings implements layout.Help
func (m *ConfigEditorModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Left,
		m.keyMap.Right,
		m.keyMap.Enter,
		m.keyMap.Presets,
		m.keyMap.Reset,
	}
}

// SetFields updates the configuration fields
func (m *ConfigEditorModel) SetFields(fields []*FieldModel) {
	m.fields = fields
	if m.cursor >= len(fields) {
		m.cursor = len(fields) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// GetField returns the field at the given index
func (m *ConfigEditorModel) GetField(index int) *FieldModel {
	if index < 0 || index >= len(m.fields) {
		return nil
	}
	return m.fields[index]
}

// GetCurrentField returns the currently selected field
func (m *ConfigEditorModel) GetCurrentField() *FieldModel {
	return m.GetField(m.cursor)
}

// UpdateField updates a field's current value
func (m *ConfigEditorModel) UpdateField(key, value string) {
	for _, field := range m.fields {
		if field.Key == key {
			field.CurrentValue = value
			if idx, exists := field.GetValueIndex(value); exists {
				field.cursor = idx
			}
			break
		}
	}
}

// SetAppName sets the application name
func (m *ConfigEditorModel) SetAppName(appName string) {
	m.appName = appName
}

// GetAppName returns the application name
func (m *ConfigEditorModel) GetAppName() string {
	return m.appName
}

// Ensure ConfigEditorModel implements the required interfaces
var _ util.Model = (*ConfigEditorModel)(nil)
var _ layout.Focusable = (*ConfigEditorModel)(nil)
var _ layout.Sizeable = (*ConfigEditorModel)(nil)
var _ layout.Help = (*ConfigEditorModel)(nil)
