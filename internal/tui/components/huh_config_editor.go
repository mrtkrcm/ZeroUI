package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/keys"
	"github.com/mrtkrcm/ZeroUI/internal/tui/layout"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// HuhConfigEditorModel represents the Huh-based configuration editor
type HuhConfigEditorModel struct {
	width      int
	height     int
	appName    string
	fields     []*FieldModel
	form       *huh.Form
	focused    bool
	keyMap     keys.AppKeyMap
	styles     *styles.Styles
	formValues map[string]interface{}
	hasChanges bool
}

// NewHuhConfigEditor creates a new Huh-based configuration editor
func NewHuhConfigEditor(appName string) *HuhConfigEditorModel {
	return &HuhConfigEditorModel{
		appName:    appName,
		fields:     make([]*FieldModel, 0),
		focused:    false,
		keyMap:     keys.DefaultKeyMap(),
		styles:     styles.GetStyles(),
		formValues: make(map[string]interface{}),
		hasChanges: false,
	}
}

// buildForm creates the Huh form based on field configuration
func (m *HuhConfigEditorModel) buildForm() {
	if len(m.fields) == 0 {
		return
	}

	var groups []*huh.Group
	var currentGroup []*huh.Field

	// Group fields by categories or create logical groups
	for i, field := range m.fields {
		var formField huh.Field

		// Initialize form value if not set
		if _, exists := m.formValues[field.Key]; !exists {
			m.formValues[field.Key] = field.CurrentValue
		}

		switch field.Type {
		case "boolean", "bool":
			// Use confirm for boolean fields
			boolValue := false
			if field.CurrentValue == "true" || field.CurrentValue == "1" || field.CurrentValue == "yes" {
				boolValue = true
			}

			formField = huh.NewConfirm().
				Key(field.Key).
				Title(field.Key).
				Description(field.Description).
				Value(&boolValue).
				Affirmative("Yes").
				Negative("No")

		case "select", "enum":
			if len(field.Values) > 0 {
				// Use select for enumerated values
				var options []huh.Option[string]
				for _, val := range field.Values {
					options = append(options, huh.NewOption(val, val))
				}

				selectValue := field.CurrentValue

				formField = huh.NewSelect[string]().
					Key(field.Key).
					Title(field.Key).
					Description(field.Description).
					Options(options...).
					Value(&selectValue)
			} else {
				// Fallback to input
				inputValue := field.CurrentValue
				formField = huh.NewInput().
					Key(field.Key).
					Title(field.Key).
					Description(field.Description).
					Value(&inputValue).
					Placeholder("Enter " + field.Key)
			}

		case "multiselect":
			if len(field.Values) > 0 {
				// Use multiselect for multiple choices
				var options []huh.Option[string]
				for _, val := range field.Values {
					options = append(options, huh.NewOption(val, val))
				}

				// Parse current value as comma-separated
				var selectedValues []string
				if field.CurrentValue != "" {
					selectedValues = strings.Split(field.CurrentValue, ",")
					for i, val := range selectedValues {
						selectedValues[i] = strings.TrimSpace(val)
					}
				}

				formField = huh.NewMultiSelect[string]().
					Key(field.Key).
					Title(field.Key).
					Description(field.Description).
					Options(options...).
					Value(&selectedValues)
			}

		case "text", "string", "":
		default:
			// Use input for text fields
			inputValue := field.CurrentValue

			input := huh.NewInput().
				Key(field.Key).
				Title(field.Key).
				Description(field.Description).
				Value(&inputValue).
				Placeholder("Enter " + field.Key)

			// Add validation for specific field types with real-time feedback
			switch field.Type {
			case "int", "integer":
				input = input.Validate(func(s string) error {
					if s == "" {
						return nil // Allow empty values
					}
					// Enhanced integer validation
					for i, r := range s {
						if r < '0' || r > '9' {
							if r == '-' && i == 0 && len(s) > 1 {
								continue // Allow negative sign at start
							}
							return fmt.Errorf("must be a valid integer (only digits and optional minus sign)")
						}
					}
					return nil
				})
			case "float", "number":
				input = input.Validate(func(s string) error {
					if s == "" {
						return nil // Allow empty values
					}
					// Float validation
					dotCount := 0
					for i, r := range s {
						if r == '.' {
							dotCount++
							if dotCount > 1 {
								return fmt.Errorf("only one decimal point allowed")
							}
						} else if r < '0' || r > '9' {
							if r == '-' && i == 0 && len(s) > 1 {
								continue // Allow negative sign at start
							}
							return fmt.Errorf("must be a valid number")
						}
					}
					return nil
				})
			case "email":
				input = input.Validate(func(s string) error {
					if s == "" {
						return nil // Allow empty values
					}
					// Basic email validation
					if !strings.Contains(s, "@") {
						return fmt.Errorf("must be a valid email address")
					}
					parts := strings.Split(s, "@")
					if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
						return fmt.Errorf("must be a valid email address")
					}
					if !strings.Contains(parts[1], ".") {
						return fmt.Errorf("email domain must contain a dot")
					}
					return nil
				})
			case "url":
				input = input.Validate(func(s string) error {
					if s == "" {
						return nil // Allow empty values
					}
					// Basic URL validation
					if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
						return fmt.Errorf("URL must start with http:// or https://")
					}
					return nil
				})
			case "path":
				input = input.Validate(func(s string) error {
					if s == "" {
						return nil // Allow empty values
					}
					// Basic path validation - check for invalid characters
					invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
					for _, char := range invalidChars {
						if strings.Contains(s, char) {
							return fmt.Errorf("path contains invalid character: %s", char)
						}
					}
					return nil
				})
			}

			formField = input
		}

		currentGroup = append(currentGroup, &formField)

		// Create groups of 5 fields for better organization
		if len(currentGroup) >= 5 || i == len(m.fields)-1 {
			// Only create group if we have fields
			if len(currentGroup) > 0 {
				groupTitle := "Configuration"
				if len(groups) > 0 {
					groupTitle = fmt.Sprintf("Configuration (Part %d)", len(groups)+1)
				}

				// Convert []*huh.Field to []huh.Field
				var fieldsSlice []huh.Field
				for _, field := range currentGroup {
					if field != nil {
						fieldsSlice = append(fieldsSlice, *field)
					}
				}

				// Only create group if we have valid fields
				if len(fieldsSlice) > 0 {
					// Additional safety check - ensure no nil fields
					var safeFields []huh.Field
					for _, field := range fieldsSlice {
						// Use reflection or type assertion to verify field is valid
						safeFields = append(safeFields, field)
					}

					if len(safeFields) > 0 {
						// Wrap in defer to catch any remaining panics
						func() {
							defer func() {
								if r := recover(); r != nil {
									// Log error but don't crash
									return
								}
							}()
							group := huh.NewGroup(safeFields...).Title(groupTitle)
							groups = append(groups, group)
						}()
					}
				}
				currentGroup = []*huh.Field{}
			}
		}
	}

	// Create the form - only if we have groups
	if len(groups) > 0 {
		m.form = huh.NewForm(groups...).
			WithShowHelp(true).
			WithShowErrors(true).
			WithTheme(huh.ThemeCharm())
	} else {
		// Create empty form if no groups
		m.form = huh.NewForm().
			WithShowHelp(true).
			WithShowErrors(true).
			WithTheme(huh.ThemeCharm())
	}
}

// Init initializes the component
func (m *HuhConfigEditorModel) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return nil
}

// Update handles messages with proper Huh form lifecycle integration
func (m *HuhConfigEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
		}

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Handle custom key bindings before form processes them
		switch {
		case key.Matches(msg, m.keyMap.Reset):
			// Reset all fields to their default values
			for _, field := range m.fields {
				if len(field.Values) > 0 {
					field.CurrentValue = field.Values[0]
					field.cursor = 0
				}
			}
			m.buildForm()
			if m.form != nil {
				cmds = append(cmds, m.form.Init())
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keyMap.Presets):
			// Send preset selection message
			cmds = append(cmds, func() tea.Msg {
				return OpenPresetsMsg{}
			})
			return m, tea.Batch(cmds...)
		}
	}

	// Always update the form - critical for Huh integration
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		m.form = form.(*huh.Form)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check for field changes and emit change events
		changeCmd := m.checkForChanges()
		if changeCmd != nil {
			cmds = append(cmds, changeCmd)
		}
	}

	// Return with batched commands
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// checkForChanges detects when form values change and emits change events
func (m *HuhConfigEditorModel) checkForChanges() tea.Cmd {
	if m.form == nil {
		return nil
	}

	// Check if the form is completed to mark as having changes
	// Note: This is a simplified approach. In practice, you'd want to track
	// individual field changes more precisely.
	if m.form.State == huh.StateCompleted {
		m.hasChanges = true

		// Emit field change events for each field
		var cmds []tea.Cmd
		for _, field := range m.fields {
			// Check if field value has changed from initial
			if field.CurrentValue != m.formValues[field.Key] {
				cmds = append(cmds, func() tea.Msg {
					return FieldChangedMsg{
						Key:   field.Key,
						Value: field.CurrentValue,
					}
				})
				// Update stored value
				m.formValues[field.Key] = field.CurrentValue
			}
		}

		if len(cmds) > 0 {
			return tea.Batch(cmds...)
		}
	}

	return nil
}

// View renders the component
func (m *HuhConfigEditorModel) View() string {
	if len(m.fields) == 0 {
		return m.renderEmptyState()
	}

	if m.form == nil {
		return m.styles.Muted.Render("Building form...")
	}

	// Create header
	header := m.renderHeader()

	// Get form view
	formView := m.form.View()

	// Create footer with actions
	footer := m.renderFooter()

	// Calculate dimensions
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	availableHeight := m.height - headerHeight - footerHeight - 4

	if availableHeight < 10 {
		availableHeight = 10
	}

	// Style the form container
	formStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(availableHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0"))

	styledForm := formStyle.Render(formView)

	// Compose the view
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledForm,
		footer,
	)

	return content
}

// renderHeader creates the editor header
func (m *HuhConfigEditorModel) renderHeader() string {
	title := fmt.Sprintf("⚙️  %s Configuration", m.appName)

	// Add change indicator
	changeIndicator := ""
	if m.hasChanges {
		changeIndicator = " 🔄 (Modified)"
	}

	subtitle := fmt.Sprintf("%d configuration options%s", len(m.fields), changeIndicator)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Padding(0, 1, 1, 1).
		Width(m.width - 4)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Italic(true).
		Padding(0, 1)

	return titleStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			subtitleStyle.Render(subtitle),
		),
	)
}

// renderFooter creates the action footer
func (m *HuhConfigEditorModel) renderFooter() string {
	var actions []string

	// Form navigation hints
	if m.form != nil {
		switch m.form.State {
		case huh.StateNormal:
			actions = append(actions, "↑↓ Navigate", "⏎ Select", "Tab Next Field")
		case huh.StateCompleted:
			actions = append(actions, "✅ Configuration Complete")
		}
	}

	// General actions
	actions = append(actions, "r Reset", "p Presets", "esc Back", "q Quit")

	helpText := strings.Join(actions, "  •  ")

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Background(lipgloss.Color("#F8FAFC")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Width(m.width - 4)

	return style.Render(helpText)
}

// renderEmptyState shows when no fields are available
func (m *HuhConfigEditorModel) renderEmptyState() string {
	emptyMsg := fmt.Sprintf(`
🤷 No Configuration Available

No configuration fields are available for %s.

This could mean:
• The application doesn't have a ZeroUI configuration file
• The configuration file couldn't be loaded
• The application doesn't support configuration management

Try selecting a different application.
`, m.appName)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Padding(2, 4).
		Width(60)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		style.Render(emptyMsg),
	)
}

// Focus implements layout.Focusable
func (m *HuhConfigEditorModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur implements layout.Focusable
func (m *HuhConfigEditorModel) Blur() tea.Cmd {
	m.focused = false
	return nil
}

// IsFocused implements layout.Focusable
func (m *HuhConfigEditorModel) IsFocused() bool {
	return m.focused
}

// SetSize implements layout.Sizeable
func (m *HuhConfigEditorModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable
func (m *HuhConfigEditorModel) GetSize() (int, int) {
	return m.width, m.height
}

// Bindings implements layout.Help
func (m *HuhConfigEditorModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Enter,
		m.keyMap.Tab,
		m.keyMap.Reset,
		m.keyMap.Presets,
		m.keyMap.Back,
		m.keyMap.Quit,
	}
}

// SetFields updates the configuration fields and rebuilds the form
func (m *HuhConfigEditorModel) SetFields(fields []*FieldModel) {
	m.fields = fields
	m.hasChanges = false
	m.buildForm()
}

// SetAppName sets the application name
func (m *HuhConfigEditorModel) SetAppName(appName string) {
	m.appName = appName
}

// GetAppName returns the application name
func (m *HuhConfigEditorModel) GetAppName() string {
	return m.appName
}

// GetField returns a field by index
func (m *HuhConfigEditorModel) GetField(index int) *FieldModel {
	if index < 0 || index >= len(m.fields) {
		return nil
	}
	return m.fields[index]
}

// UpdateField updates a field's current value
func (m *HuhConfigEditorModel) UpdateField(key, value string) {
	for _, field := range m.fields {
		if field.Key == key {
			field.CurrentValue = value
			if idx, exists := field.GetValueIndex(value); exists {
				field.cursor = idx
			}
			break
		}
	}
	m.hasChanges = true
}

// Ensure HuhConfigEditorModel implements required interfaces
var _ util.Model = (*HuhConfigEditorModel)(nil)
var _ layout.Focusable = (*HuhConfigEditorModel)(nil)
var _ layout.Sizeable = (*HuhConfigEditorModel)(nil)
var _ layout.Help = (*HuhConfigEditorModel)(nil)
