package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ConfigFieldType represents the type of configuration field
type ConfigFieldType string

const (
	FieldTypeString ConfigFieldType = "string"
	FieldTypeInt    ConfigFieldType = "int"
	FieldTypeBool   ConfigFieldType = "bool"
	FieldTypeSelect ConfigFieldType = "select"
	FieldTypeFloat  ConfigFieldType = "float"
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
	IsSet       bool   // Whether this field has a value in the config file
	Source      string // Where this value comes from (e.g., "config file", "default", "available")
}

// HuhConfigFormModel provides dynamic configuration editing with Huh forms
type HuhConfigFormModel struct {
	form      *huh.Form
	fields    []ConfigField
	appName   string
	width     int
	height    int
	focused   bool
	submitted bool
	// Track values and changes for precise saving
	values         map[string]interface{}
	stringBindings map[string]*string
	boolBindings   map[string]*bool
	originalValues map[string]string
	defaultValues  map[string]string

	// Changed-only view toggle
	showChangedOnly bool
}

// NewHuhConfigForm creates a new configuration form using Huh
func NewHuhConfigForm(appName string) *HuhConfigFormModel {
	return &HuhConfigFormModel{
		appName:        appName,
		values:         make(map[string]interface{}),
		stringBindings: make(map[string]*string),
		boolBindings:   make(map[string]*bool),
		originalValues: make(map[string]string),
		defaultValues:  make(map[string]string),
		width:          80,
		height:         24,
	}
}

// SetFields configures the form fields
func (m *HuhConfigFormModel) SetFields(fields []ConfigField) {
	m.fields = fields
	// Seed originals and defaults for change detection
	m.originalValues = make(map[string]string)
	m.defaultValues = make(map[string]string)
	m.stringBindings = make(map[string]*string)
	m.boolBindings = make(map[string]*bool)
	for _, f := range fields {
		if f.IsSet {
			if f.Value != nil {
				m.originalValues[f.Key] = fmt.Sprintf("%v", f.Value)
			} else {
				m.originalValues[f.Key] = ""
			}
		} else if f.Value != nil {
			// Treat provided Value for not-set fields as default for the UI
			m.defaultValues[f.Key] = fmt.Sprintf("%v", f.Value)
		}
	}
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
		if len(groupFields) == 0 {
			continue // Skip empty groups
		}

		var fields []huh.Field

		// Add fields to the group
		for _, field := range groupFields {
			if m.showChangedOnly && !m.isFieldChanged(field.Key) {
				continue
			}
			huhField := m.createHuhField(field)
			if huhField != nil {
				fields = append(fields, huhField)
			}
		}

		// Only create group if we have fields
		if len(fields) > 0 {
			group := huh.NewGroup(fields...)

			// Add group title if not the default group
			if groupName != "General" {
				group = group.Title(groupName)
			}

			huhGroups = append(huhGroups, group)
		}
	}

	// Create the form with enhanced styling
	if len(huhGroups) > 0 {
		formWidth, formHeight, _ := m.calculateLayout()
		m.form = huh.NewForm(huhGroups...).
			WithTheme(m.createCustomTheme()).
			WithWidth(formWidth).
			WithHeight(formHeight)
	} else {
		// Create empty form if no groups
		formWidth, formHeight, _ := m.calculateLayout()
		m.form = huh.NewForm().
			WithTheme(m.createCustomTheme()).
			WithWidth(formWidth).
			WithHeight(formHeight)
	}
}

// groupFields organizes fields into logical groups with existing vs available separation
func (m *HuhConfigFormModel) groupFields() map[string][]ConfigField {
	groups := make(map[string][]ConfigField)

	// First, separate existing configurations from available options
	var existingFields []ConfigField
	var availableFields []ConfigField

	for _, field := range m.fields {
		if field.IsSet {
			existingFields = append(existingFields, field)
		} else {
			availableFields = append(availableFields, field)
		}
	}

	// Group existing configurations
	if len(existingFields) > 0 {
		for _, field := range existingFields {
			groupName := "[*] Current Configuration - " + m.determineGroup(field.Key)
			groups[groupName] = append(groups[groupName], field)
		}
	}

	// Group available options
	if len(availableFields) > 0 {
		for _, field := range availableFields {
			groupName := "[+] Available Options - " + m.determineGroup(field.Key)
			groups[groupName] = append(groups[groupName], field)
		}
	}

	// Ensure we have at least one group
	if len(groups) == 0 {
		groups["[*] Configuration"] = m.fields
	}

	return groups
}

// determineGroup determines which group a field belongs to based on its key
func (m *HuhConfigFormModel) determineGroup(key string) string {
	lowerKey := strings.ToLower(key)

	// Theme-related fields
	if strings.Contains(lowerKey, "theme") || strings.Contains(lowerKey, "color") ||
		strings.Contains(lowerKey, "background") || strings.Contains(lowerKey, "foreground") ||
		strings.Contains(lowerKey, "bell") || strings.Contains(lowerKey, "sound") || strings.Contains(lowerKey, "audio") {
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
		ptr := m.getBindingString(field.Key, field.Value)
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field)).
			Description(m.formatDescription(field)).
			Value(ptr)

		if field.Required {
			input = input.Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("Enter a value")
				}
				return nil
			})
		}

		return input

	case FieldTypeInt:
		ptr := m.getBindingString(field.Key, field.Value)
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field)).
			Description(m.formatDescription(field)).
			Value(ptr).
			Validate(func(s string) error {
				if s == "" {
					if field.Required {
						return fmt.Errorf("Enter a number")
					}
					return nil
				}

				val, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("Please enter a valid number")
				}

				if field.Min != nil && float64(val) < *field.Min {
					return fmt.Errorf("Must be at least %.0f", *field.Min)
				}

				if field.Max != nil && float64(val) > *field.Max {
					return fmt.Errorf("Must be at most %.0f", *field.Max)
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
		bptr := m.getBindingBool(field.Key, currentValue)

		return huh.NewConfirm().
			Key(field.Key).
			Title(m.formatTitle(field)).
			Description(m.formatDescription(field)).
			Value(bptr)

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
		sptr := m.getBindingString(field.Key, currentValue)
		return huh.NewSelect[string]().
			Key(field.Key).
			Title(m.formatTitle(field)).
			Description(m.formatDescription(field)).
			Options(options...).
			Value(sptr)

	case FieldTypeFloat:
		ptr := m.getBindingString(field.Key, field.Value)
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatTitle(field)).
			Description(m.formatDescription(field)).
			Value(ptr).
			Validate(func(s string) error {
				if s == "" {
					if field.Required {
						return fmt.Errorf("Enter a number")
					}
					return nil
				}

				val, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return fmt.Errorf("Please enter a valid number")
				}

				if field.Min != nil && val < *field.Min {
					return fmt.Errorf("Must be at least %.1f", *field.Min)
				}

				if field.Max != nil && val > *field.Max {
					return fmt.Errorf("Must be at most %.1f", *field.Max)
				}

				return nil
			})

		return input
	}

	return nil
}

// getStringPointer returns a pointer to the string representation of a value
func (m *HuhConfigFormModel) getBindingString(key string, value interface{}) *string {
	if ptr, ok := m.stringBindings[key]; ok && ptr != nil {
		return ptr
	}
	initial := ""
	if value != nil {
		initial = fmt.Sprintf("%v", value)
	}
	s := initial
	m.stringBindings[key] = &s
	m.values[key] = initial
	return &s
}

func (m *HuhConfigFormModel) getBindingBool(key string, value bool) *bool {
	if ptr, ok := m.boolBindings[key]; ok && ptr != nil {
		return ptr
	}
	b := value
	m.boolBindings[key] = &b
	m.values[key] = b
	return &b
}

// formatTitle converts a field key to a human-readable title with visual indicators
func (m *HuhConfigFormModel) formatTitle(field ConfigField) string {
	// Convert snake_case and kebab-case to title case
	words := strings.FieldsFunc(field.Key, func(c rune) bool {
		return c == '_' || c == '-' || c == '.'
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	title := strings.Join(words, " ")

	// Add visual indicators based on field status
	if field.IsSet {
		if field.Value != nil && field.Value != "" {
			return fmt.Sprintf("[*] %s (current: %v)", title, field.Value)
		} else {
			return fmt.Sprintf("[!] %s (set but empty)", title)
		}
	} else {
		return fmt.Sprintf("[ ] %s (available)", title)
	}
}

// formatDescription enhances the description with source information
func (m *HuhConfigFormModel) formatDescription(field ConfigField) string {
	baseDesc := field.Description
	if baseDesc == "" {
		baseDesc = fmt.Sprintf("Configure %s", field.Key)
	}

	// Add source information
	if field.IsSet {
		if field.Source != "" {
			return fmt.Sprintf("%s\n>> Source: %s", baseDesc, field.Source)
		} else {
			return fmt.Sprintf("%s\n>> Currently configured", baseDesc)
		}
	} else {
		return fmt.Sprintf("%s\n>> Available option - not currently set", baseDesc)
	}
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

// Update handles form updates with enhanced error handling and performance
func (m *HuhConfigFormModel) Update(msg tea.Msg) (*HuhConfigFormModel, tea.Cmd) {
	if m == nil || m.form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			if m.form != nil {
				newWidth := max(m.width-4, 20)  // Ensure minimum width
				newHeight := max(m.height-8, 5) // Ensure minimum height
				m.form = m.form.WithWidth(newWidth).WithHeight(newHeight)
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Save the form with validation
			if m.form != nil && m.appName != "" {
				values := m.getFormValues()
				if len(values) > 0 {
					return m, func() tea.Msg {
						return ConfigSavedMsg{AppName: m.appName, Values: values}
					}
				}
			}
		case "c", "C":
			m.showChangedOnly = !m.showChangedOnly
			m.buildForm()
			return m, nil
		}
	}

	// Update form with error recovery
	defer func() {
		if r := recover(); r != nil {
			// Log error but don't crash
			// Would need logger for proper error handling
		}
	}()

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

	// Add a header and breadcrumb
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		MarginBottom(1).
		Render(fmt.Sprintf("Apps › %s", m.appName))

	// Add help text / footer hints
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginTop(1).
		Render("tab/shift+tab navigate • enter select • ctrl+s save • C changed-only • p presets • esc back")

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
		formWidth, formHeight, _ := m.calculateLayout()
		m.form = m.form.WithWidth(formWidth).WithHeight(formHeight)
	}
	return nil
}

// getFormValues extracts all form values
func (m *HuhConfigFormModel) getFormValues() map[string]interface{} {
	if m.form == nil {
		return make(map[string]interface{})
	}

	values := make(map[string]interface{})

	for _, field := range m.fields {
		key := field.Key
		// Determine current value by type
		var curStr string
		if ptr, ok := m.stringBindings[key]; ok && ptr != nil {
			curStr = *ptr
		} else if bptr, ok := m.boolBindings[key]; ok && bptr != nil {
			if *bptr {
				curStr = "true"
			} else {
				curStr = "false"
			}
		} else {
			// No binding; skip
			continue
		}

		orig, hadOrig := m.originalValues[key]
		def, hadDef := m.defaultValues[key]

		include := false
		if hadOrig {
			include = (curStr != orig)
		} else if hadDef {
			include = (curStr != def)
		} else {
			// No origin and no default; include only if non-empty (user-provided)
			include = (strings.TrimSpace(curStr) != "")
		}

		if include {
			values[key] = curStr
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

// calculateLayout determines adaptive width/height and column count
func (m *HuhConfigFormModel) calculateLayout() (width, height, columns int) {
	// Respect padding: keep some breathing room
	targetWidth := m.width - 6
	targetHeight := m.height - 8

	// Minimum sizes for comfortable interaction
	width = max(targetWidth, 40)
	height = max(targetHeight, 10)

	// Simple column heuristic for future use
	columns = 1
	if width > 120 {
		columns = 2
	}
	return
}

// isFieldChanged reports whether a given field key differs from original/default
func (m *HuhConfigFormModel) isFieldChanged(key string) bool {
	var cur string
	if ptr, ok := m.stringBindings[key]; ok && ptr != nil {
		cur = *ptr
	} else if bptr, ok := m.boolBindings[key]; ok && bptr != nil {
		if *bptr {
			cur = "true"
		} else {
			cur = "false"
		}
	} else {
		return false
	}
	if orig, ok := m.originalValues[key]; ok {
		return cur != orig
	}
	if def, ok := m.defaultValues[key]; ok {
		return cur != def
	}
	return strings.TrimSpace(cur) != ""
}

// IsValid returns true if the form is valid
func (m *HuhConfigFormModel) IsValid() bool {
	// Check if form exists and has no errors
	if m.form == nil {
		return false
	}
	// For now, assume form is valid if it's been submitted successfully
	return m.submitted || m.form.State == huh.StateCompleted
}

// GetValues returns the current form values
func (m *HuhConfigFormModel) GetValues() map[string]interface{} {
	// Return the values that have changed
	values := make(map[string]interface{})

	for key, ptr := range m.stringBindings {
		if ptr != nil && m.isFieldChanged(key) {
			values[key] = *ptr
		}
	}

	for key, ptr := range m.boolBindings {
		if ptr != nil && m.isFieldChanged(key) {
			values[key] = *ptr
		}
	}

	return values
}

// HasUnsavedChanges returns true if there are unsaved changes
func (m *HuhConfigFormModel) HasUnsavedChanges() bool {
	for key := range m.stringBindings {
		if m.isFieldChanged(key) {
			return true
		}
	}
	for key := range m.boolBindings {
		if m.isFieldChanged(key) {
			return true
		}
	}
	return false
}

// ApplyPreset applies preset values to the form
func (m *HuhConfigFormModel) ApplyPreset(values map[string]interface{}) {
	for key, value := range values {
		// Update string bindings
		if ptr, ok := m.stringBindings[key]; ok && ptr != nil {
			switch v := value.(type) {
			case string:
				*ptr = v
			case bool:
				if v {
					*ptr = "true"
				} else {
					*ptr = "false"
				}
			default:
				*ptr = fmt.Sprintf("%v", value)
			}
		}

		// Update bool bindings
		if ptr, ok := m.boolBindings[key]; ok && ptr != nil {
			switch v := value.(type) {
			case bool:
				*ptr = v
			case string:
				*ptr = (v == "true" || v == "yes" || v == "1")
			default:
				*ptr = false
			}
		}

		// Update values map
		m.values[key] = value
	}
}
