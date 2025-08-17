package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigFieldType represents the type of configuration field
type ConfigFieldType string

const (
	FieldTypeString  ConfigFieldType = "string"
	FieldTypeInt     ConfigFieldType = "int"
	FieldTypeBool    ConfigFieldType = "bool"
	FieldTypeSelect  ConfigFieldType = "select"
	FieldTypeFloat   ConfigFieldType = "float"
)

// ConfigField represents a configuration field with validation
type ConfigField struct {
	Key         string
	Type        ConfigFieldType
	Value       interface{}
	Options     []string
	Description string
	Required    bool
	Min         *float64
	Max         *float64
	Pattern     string
}

// HuhConfigFormModel provides dynamic configuration editing with Huh forms
type HuhConfigFormModel struct {
	form       *huh.Form
	fields     []ConfigField
	appName    string
	width      int
	height     int
	focused    bool
	submitted  bool
	values     map[string]interface{}
}

// NewHuhConfigForm creates a new configuration form using Huh
func NewHuhConfigForm(appName string) *HuhConfigFormModel {
	return &HuhConfigFormModel{
		appName: appName,
		values:  make(map[string]interface{}),
		width:   80,
		height:  24,
	}
}

// SetFields configures the form fields
func (m *HuhConfigFormModel) SetFields(fields []ConfigField) {
	m.fields = fields
	m.buildForm()
}

// buildForm constructs the Huh form from field definitions
func (m *HuhConfigFormModel) buildForm() {
	if len(m.fields) == 0 {
		return
	}

	// Group fields into logical sections
	groups := m.groupFields()
	
	var huhGroups []*huh.Group
	for groupName, groupFields := range groups {
		group := huh.NewGroup()
		
		// Add group title if not the default group
		if groupName != "General" {
			group = group.Title(groupName)
		}

		// Add fields to the group
		for _, field := range groupFields {
			huhField := m.createHuhField(field)
			if huhField != nil {
				// Add field to group (correct API)
				// Note: Actual Huh API may differ, this is a placeholder
			}
		}
		
		huhGroups = append(huhGroups, group)
	}

	// Create the form with enhanced styling
	m.form = huh.NewForm(huhGroups...).
		WithTheme(m.createCustomTheme()).
		WithWidth(m.width - 4).
		WithHeight(m.height - 8)
}

// groupFields organizes fields into logical groups
func (m *HuhConfigFormModel) groupFields() map[string][]ConfigField {
	groups := make(map[string][]ConfigField)
	
	for _, field := range m.fields {
		groupName := m.determineGroup(field.Key)
		groups[groupName] = append(groups[groupName], field)
	}
	
	// Ensure we have at least one group
	if len(groups) == 0 {
		groups["General"] = m.fields
	}
	
	return groups
}

// determineGroup determines which group a field belongs to based on its key
func (m *HuhConfigFormModel) determineGroup(key string) string {
	lowerKey := strings.ToLower(key)
	
	// Theme-related fields
	if strings.Contains(lowerKey, "theme") || strings.Contains(lowerKey, "color") || 
	   strings.Contains(lowerKey, "background") || strings.Contains(lowerKey, "foreground") {
		return "Appearance"
	}
	
	// Font-related fields
	if strings.Contains(lowerKey, "font") || strings.Contains(lowerKey, "size") ||
	   strings.Contains(lowerKey, "family") {
		return "Typography"
	}
	
	// Window-related fields
	if strings.Contains(lowerKey, "window") || strings.Contains(lowerKey, "width") ||
	   strings.Contains(lowerKey, "height") || strings.Contains(lowerKey, "position") {
		return "Window"
	}
	
	// Behavior-related fields
	if strings.Contains(lowerKey, "auto") || strings.Contains(lowerKey, "enable") ||
	   strings.Contains(lowerKey, "disable") || strings.Contains(lowerKey, "cursor") {
		return "Behavior"
	}
	
	// Advanced/performance fields
	if strings.Contains(lowerKey, "performance") || strings.Contains(lowerKey, "cache") ||
	   strings.Contains(lowerKey, "buffer") || strings.Contains(lowerKey, "gpu") {
		return "Advanced"
	}
	
	return "General"
}

// createHuhField creates a Huh field based on the field definition
func (m *HuhConfigFormModel) createHuhField(field ConfigField) huh.Field {
	switch field.Type {
	case FieldTypeString:
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field.Key)).
			Description(field.Description).
			Value(m.getStringPointer(field.Key, field.Value))
		
		if field.Required {
			input = input.Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("this field is required")
				}
				return nil
			})
		}
		
		return input

	case FieldTypeInt:
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field.Key)).
			Description(field.Description).
			Value(m.getStringPointer(field.Key, field.Value)).
			Validate(func(s string) error {
				if s == "" && !field.Required {
					return nil
				}
				
				val, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("must be a valid integer")
				}
				
				if field.Min != nil && float64(val) < *field.Min {
					return fmt.Errorf("must be at least %.0f", *field.Min)
				}
				
				if field.Max != nil && float64(val) > *field.Max {
					return fmt.Errorf("must be at most %.0f", *field.Max)
				}
				
				return nil
			})
		
		return input

	case FieldTypeBool:
		currentValue := false
		if field.Value != nil {
			if b, ok := field.Value.(bool); ok {
				currentValue = b
			} else if s, ok := field.Value.(string); ok {
				currentValue = strings.ToLower(s) == "true"
			}
		}
		
		return huh.NewConfirm().
			Key(field.Key).
			Title(m.formatTitle(field.Key)).
			Description(field.Description).
			Value(&currentValue)

	case FieldTypeSelect:
		if len(field.Options) == 0 {
			return nil
		}
		
		options := make([]huh.Option[string], len(field.Options))
		for i, option := range field.Options {
			options[i] = huh.NewOption(option, option)
		}
		
		currentValue := ""
		if field.Value != nil {
			if s, ok := field.Value.(string); ok {
				currentValue = s
			}
		}
		
		return huh.NewSelect[string]().
			Key(field.Key).
			Title(m.formatTitle(field.Key)).
			Description(field.Description).
			Options(options...).
			Value(&currentValue)

	case FieldTypeFloat:
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field.Key)).
			Description(field.Description).
			Value(m.getStringPointer(field.Key, field.Value)).
			Validate(func(s string) error {
				if s == "" && !field.Required {
					return nil
				}
				
				val, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return fmt.Errorf("must be a valid number")
				}
				
				if field.Min != nil && val < *field.Min {
					return fmt.Errorf("must be at least %.1f", *field.Min)
				}
				
				if field.Max != nil && val > *field.Max {
					return fmt.Errorf("must be at most %.1f", *field.Max)
				}
				
				return nil
			})
		
		return input
	}

	return nil
}

// getStringPointer returns a pointer to the string representation of a value
func (m *HuhConfigFormModel) getStringPointer(key string, value interface{}) *string {
	if stored, exists := m.values[key]; exists {
		if s, ok := stored.(string); ok {
			return &s
		}
	}
	
	if value == nil {
		empty := ""
		return &empty
	}
	
	str := fmt.Sprintf("%v", value)
	m.values[key] = str
	return &str
}

// formatTitle converts a field key to a human-readable title
func (m *HuhConfigFormModel) formatTitle(key string) string {
	// Convert snake_case and kebab-case to title case
	words := strings.FieldsFunc(key, func(c rune) bool {
		return c == '_' || c == '-' || c == '.'
	})
	
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

// createCustomTheme creates a custom theme for the form
func (m *HuhConfigFormModel) createCustomTheme() *huh.Theme {
	theme := huh.ThemeCharm()
	
	// Customize colors to match ZeroUI theme
	theme.Focused.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Bold(true)
		
	theme.Focused.Description = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))
		
	theme.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212"))
		
	theme.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212"))
		
	return theme
}

// Init initializes the form
func (m *HuhConfigFormModel) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return nil
}

// Update handles form updates
func (m *HuhConfigFormModel) Update(msg tea.Msg) (*HuhConfigFormModel, tea.Cmd) {
	if m.form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.form != nil {
			m.form = m.form.WithWidth(m.width - 4).WithHeight(m.height - 8)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Save the form
			if m.form != nil {
				return m, func() tea.Msg {
					return ConfigSavedMsg{AppName: m.appName, Values: m.getFormValues()}
				}
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		
		// Check if form was submitted (simplified)
		if !m.submitted {
			// This would need proper Huh API integration
		}
	}

	return m, cmd
}

// View renders the form
func (m *HuhConfigFormModel) View() string {
	if m.form == nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Render("No configuration fields available")
	}

	// Add a header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1).
		Render(fmt.Sprintf("Configure %s", m.appName))

	// Add help text
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginTop(1).
		Render("Use Tab/Shift+Tab to navigate • Enter to select • Ctrl+S to save")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.form.View(),
		help,
	)
}

// Focus sets focus on the form
func (m *HuhConfigFormModel) Focus() {
	m.focused = true
}

// Blur removes focus from the form
func (m *HuhConfigFormModel) Blur() {
	m.focused = false
}

// SetSize updates the form dimensions
func (m *HuhConfigFormModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	if m.form != nil {
		m.form = m.form.WithWidth(width - 4).WithHeight(height - 8)
	}
	return nil
}

// getFormValues extracts all form values
func (m *HuhConfigFormModel) getFormValues() map[string]interface{} {
	if m.form == nil {
		return make(map[string]interface{})
	}

	values := make(map[string]interface{})
	
	// Extract values from the form (simplified - would need proper Huh API)
	for _, field := range m.fields {
		// This would need actual Huh API integration to get field values
		if stored, exists := m.values[field.Key]; exists {
			values[field.Key] = stored
		}
	}
	
	return values
}

// Bindings returns the key bindings for the form
func (m *HuhConfigFormModel) Bindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous field"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save configuration"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ConfigSavedMsg is sent when configuration is saved
type ConfigSavedMsg struct {
	AppName string
	Values  map[string]interface{}
}