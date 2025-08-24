package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// HuhFancyConfigModel provides the ultimate configuration interface using Huh
type HuhFancyConfigModel struct {
	appName string
	
	// Forms for different views
	mainForm     *huh.Form
	quickForm    *huh.Form
	detailedForm *huh.Form
	
	// State
	currentView ConfigViewMode
	fields      []ConfigField
	changed     map[string]interface{}
	
	// Dimensions
	width  int
	height int
	
	// Styles
	lg     *lipgloss.Renderer
	styles *HuhConfigStyles
	
	// Form values
	selectedMode   string
	quickSettings  map[string]interface{}
	detailSettings map[string]interface{}
}

// ConfigViewMode represents different configuration modes
type ConfigViewMode int

const (
	ModeSelection ConfigViewMode = iota
	QuickSettings
	DetailedSettings
	ReviewChanges
	Completed
)

// HuhConfigStyles contains all styles for the interface
type HuhConfigStyles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help,
	Success lipgloss.Style
}

// NewHuhConfigStyles creates styles for the interface
func NewHuhConfigStyles(lg *lipgloss.Renderer) *HuhConfigStyles {
	s := HuhConfigStyles{}
	
	indigo := lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green := lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	red := lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	
	s.Base = lg.NewStyle().
		Padding(1, 2)
		
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
		
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
		
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
		
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
		
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
		
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
		
	s.Success = lg.NewStyle().
		Foreground(green).
		Bold(true)
		
	return &s
}

// NewHuhFancyConfig creates the ultimate configuration interface
func NewHuhFancyConfig(appName string) *HuhFancyConfigModel {
	lg := lipgloss.DefaultRenderer()
	
	model := &HuhFancyConfigModel{
		appName:        appName,
		lg:             lg,
		styles:         NewHuhConfigStyles(lg),
		changed:        make(map[string]interface{}),
		quickSettings:  make(map[string]interface{}),
		detailSettings: make(map[string]interface{}),
		width:          80,
		height:         24,
	}
	
	model.buildMainForm()
	
	return model
}

// buildMainForm creates the main selection form
func (m *HuhFancyConfigModel) buildMainForm() {
	m.mainForm = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("⚙️  Configure %s", m.appName)).
				Description("Welcome to configuration wizard!\n\nChoose how you'd like to configure:").
				Next(true).
				NextLabel("Start"),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("mode").
				Options(
					huh.NewOption("🚀 Quick Setup", "quick").
						Selected(true),
					huh.NewOption("🔧 Detailed Configuration", "detailed"),
					huh.NewOption("📋 Review Current Settings", "review"),
				).
				Title("Configuration Mode").
				Description("Quick setup for common scenarios or detailed for full control").
				Value(&m.selectedMode),
		),
	).
		WithWidth(60).
		WithShowHelp(false).
		WithTheme(huh.ThemeCharm())
}

// buildQuickForm creates the quick configuration form
func (m *HuhFancyConfigModel) buildQuickForm() {
	// Group common settings
	var groups []*huh.Group
	
	// Appearance group
	appearanceFields := m.filterFieldsByCategory("appearance")
	if len(appearanceFields) > 0 {
		var fields []huh.Field
		
		// Add theme selector if available
		themeField := m.findField("theme")
		if themeField != nil && len(themeField.Options) > 0 {
			fields = append(fields,
				huh.NewSelect[string]().
					Key("theme").
					Options(m.buildOptions(themeField.Options)...).
					Title("🎨 Theme").
					Description("Visual appearance"),
			)
		}
		
		// Add font selector
		fontField := m.findField("font")
		if fontField != nil {
			fields = append(fields,
				huh.NewInput().
					Key("font").
					Title("🔤 Font").
					Placeholder("SF Mono").
					Description("Terminal font family"),
			)
		}
		
		if len(fields) > 0 {
			groups = append(groups, huh.NewGroup(fields...).Title("Appearance"))
		}
	}
	
	// Window settings group
	windowFields := m.filterFieldsByCategory("window")
	if len(windowFields) > 0 {
		groups = append(groups,
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("window_mode").
					Options(
						huh.NewOption("Normal", "normal"),
						huh.NewOption("Maximized", "maximized"),
						huh.NewOption("Fullscreen", "fullscreen"),
					).
					Title("🪟 Window Mode").
					Description("How the window should appear"),
					
				huh.NewConfirm().
					Key("decorations").
					Title("Show window decorations?").
					Affirmative("Yes").
					Negative("No"),
			).Title("Window"),
		)
	}
	
	// Performance group
	groups = append(groups,
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("performance").
				Options(
					huh.NewOption("⚡ High Performance", "high"),
					huh.NewOption("⚖️  Balanced", "balanced"),
					huh.NewOption("🔋 Battery Saver", "battery"),
				).
				Title("Performance Mode").
				Description("Optimize for your use case"),
		).Title("Performance"),
	)
	
	// Confirmation
	groups = append(groups,
		huh.NewGroup(
			huh.NewConfirm().
				Key("apply").
				Title("Apply these settings?").
				Description("Your configuration will be updated").
				Affirmative("Yes, apply").
				Negative("No, go back"),
		),
	)
	
	m.quickForm = huh.NewForm(groups...).
		WithWidth(60).
		WithShowHelp(true).
		WithTheme(huh.ThemeCharm())
}

// buildDetailedForm creates the detailed configuration form
func (m *HuhFancyConfigModel) buildDetailedForm() {
	// Group all fields by category
	categories := m.groupFieldsByCategory()
	var groups []*huh.Group
	
	for category, fields := range categories {
		var formFields []huh.Field
		
		for _, field := range fields {
			formField := m.createFormField(field)
			if formField != nil {
				formFields = append(formFields, formField)
			}
		}
		
		if len(formFields) > 0 {
			group := huh.NewGroup(formFields...).
				Title(m.formatCategoryName(category))
			groups = append(groups, group)
		}
	}
	
	// Add save confirmation
	groups = append(groups,
		huh.NewGroup(
			huh.NewConfirm().
				Key("save").
				Title("Save configuration?").
				Affirmative("Save").
				Negative("Cancel"),
		),
	)
	
	m.detailedForm = huh.NewForm(groups...).
		WithWidth(70).
		WithShowHelp(true).
		WithTheme(huh.ThemeCharm())
}

// createFormField creates a Huh field from a ConfigField
func (m *HuhFancyConfigModel) createFormField(field ConfigField) huh.Field {
	switch field.Type {
	case FieldTypeBool:
		return huh.NewConfirm().
			Key(field.Key).
			Title(m.formatFieldTitle(field)).
			Description(field.Description)
			
	case FieldTypeSelect:
		if len(field.Options) > 0 {
			return huh.NewSelect[string]().
				Key(field.Key).
				Options(m.buildOptions(field.Options)...).
				Title(m.formatFieldTitle(field)).
				Description(field.Description)
		}
		
	case FieldTypeInt, FieldTypeFloat:
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatFieldTitle(field)).
			Description(field.Description)
		
		if field.Value != nil {
			input.Placeholder(fmt.Sprintf("%v", field.Value))
		}
		
		return input
		
	default: // String
		input := huh.NewInput().
			Key(field.Key).
			Title(m.formatFieldTitle(field)).
			Description(field.Description)
		
		if field.Value != nil {
			input.Placeholder(fmt.Sprintf("%v", field.Value))
		}
		
		return input
	}
	
	return nil
}

// SetFields configures the fields
func (m *HuhFancyConfigModel) SetFields(fields []ConfigField) {
	m.fields = fields
	m.buildQuickForm()
	m.buildDetailedForm()
}

// Init initializes the model
func (m *HuhFancyConfigModel) Init() tea.Cmd {
	if m.mainForm != nil {
		return m.mainForm.Init()
	}
	return nil
}

// Update handles messages
func (m *HuhFancyConfigModel) Update(msg tea.Msg) (*HuhFancyConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == ModeSelection {
				return m, tea.Quit
			}
			// Go back to mode selection
			m.currentView = ModeSelection
			m.buildMainForm()
			return m, m.mainForm.Init()
			
		case "esc":
			if m.currentView != ModeSelection {
				m.currentView = ModeSelection
				m.buildMainForm()
				return m, m.mainForm.Init()
			}
		}
	}
	
	// Update the current form
	var cmd tea.Cmd
	
	switch m.currentView {
	case ModeSelection:
		if m.mainForm != nil {
			f, c := m.mainForm.Update(msg)
			if frm, ok := f.(*huh.Form); ok {
				m.mainForm = frm
				cmd = c
				
				// Check if form is completed
				if m.mainForm.State == huh.StateCompleted {
					switch m.selectedMode {
					case "quick":
						m.currentView = QuickSettings
						m.buildQuickForm()
						return m, m.quickForm.Init()
					case "detailed":
						m.currentView = DetailedSettings
						m.buildDetailedForm()
						return m, m.detailedForm.Init()
					case "review":
						m.currentView = ReviewChanges
					}
				}
			}
		}
		
	case QuickSettings:
		if m.quickForm != nil {
			f, c := m.quickForm.Update(msg)
			if frm, ok := f.(*huh.Form); ok {
				m.quickForm = frm
				cmd = c
				
				if m.quickForm.State == huh.StateCompleted {
					// Process quick settings
					if m.quickForm.GetBool("apply") {
						m.applyQuickSettings()
						m.currentView = Completed
					} else {
						m.currentView = ModeSelection
						m.buildMainForm()
						return m, m.mainForm.Init()
					}
				}
			}
		}
		
	case DetailedSettings:
		if m.detailedForm != nil {
			f, c := m.detailedForm.Update(msg)
			if frm, ok := f.(*huh.Form); ok {
				m.detailedForm = frm
				cmd = c
				
				if m.detailedForm.State == huh.StateCompleted {
					// Process detailed settings
					if m.detailedForm.GetBool("save") {
						m.applyDetailedSettings()
						m.currentView = Completed
					} else {
						m.currentView = ModeSelection
						m.buildMainForm()
						return m, m.mainForm.Init()
					}
				}
			}
		}
	}
	
	return m, cmd
}

// View renders the interface
func (m *HuhFancyConfigModel) View() string {
	s := m.styles
	
	switch m.currentView {
	case ModeSelection:
		if m.mainForm != nil {
			return s.Base.Render(m.mainForm.View())
		}
		
	case QuickSettings:
		if m.quickForm != nil {
			header := m.headerView("Quick Setup")
			form := m.quickForm.View()
			return s.Base.Render(header + "\n\n" + form)
		}
		
	case DetailedSettings:
		if m.detailedForm != nil {
			header := m.headerView("Detailed Configuration")
			form := m.detailedForm.View()
			return s.Base.Render(header + "\n\n" + form)
		}
		
	case ReviewChanges:
		return m.renderReviewView()
		
	case Completed:
		return m.renderCompletedView()
	}
	
	return s.Base.Render("Loading...")
}

// renderReviewView shows current configuration
func (m *HuhFancyConfigModel) renderReviewView() string {
	s := m.styles
	
	var content strings.Builder
	content.WriteString(s.HeaderText.Render(fmt.Sprintf("Current %s Configuration", m.appName)))
	content.WriteString("\n\n")
	
	// Group fields by category
	categories := m.groupFieldsByCategory()
	
	for category, fields := range categories {
		if len(fields) == 0 {
			continue
		}
		
		content.WriteString(s.StatusHeader.Render(m.formatCategoryName(category)))
		content.WriteString("\n")
		
		for _, field := range fields {
			value := "not set"
			if field.Value != nil {
				value = fmt.Sprintf("%v", field.Value)
			}
			
			if field.IsSet {
				content.WriteString(fmt.Sprintf("  • %s: %s\n", 
					field.Key, 
					s.Highlight.Render(value)))
			}
		}
		content.WriteString("\n")
	}
	
	content.WriteString("\n")
	content.WriteString(s.Help.Render("Press ESC to go back to menu"))
	
	return s.Base.Render(content.String())
}

// renderCompletedView shows success message
func (m *HuhFancyConfigModel) renderCompletedView() string {
	s := m.styles
	
	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2).
		Width(50).
		Render(
			s.Success.Render("✓ Configuration Saved!\n\n") +
			fmt.Sprintf("Your %s configuration has been updated.\n\n", m.appName) +
			s.Help.Render("Press Q to quit or ESC for main menu"),
		)
	
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		successBox,
	)
}

// Helper methods

func (m *HuhFancyConfigModel) headerView(title string) string {
	return lipgloss.PlaceHorizontal(
		m.width-4,
		lipgloss.Left,
		m.styles.HeaderText.Render(title),
		lipgloss.WithWhitespaceChars("─"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("238")),
	)
}

func (m *HuhFancyConfigModel) filterFieldsByCategory(category string) []ConfigField {
	var result []ConfigField
	for _, field := range m.fields {
		if m.getFieldCategory(field.Key) == category {
			result = append(result, field)
		}
	}
	return result
}

func (m *HuhFancyConfigModel) findField(key string) *ConfigField {
	for _, field := range m.fields {
		if field.Key == key || strings.Contains(strings.ToLower(field.Key), key) {
			return &field
		}
	}
	return nil
}

func (m *HuhFancyConfigModel) groupFieldsByCategory() map[string][]ConfigField {
	groups := make(map[string][]ConfigField)
	for _, field := range m.fields {
		category := m.getFieldCategory(field.Key)
		groups[category] = append(groups[category], field)
	}
	return groups
}

func (m *HuhFancyConfigModel) getFieldCategory(key string) string {
	lowerKey := strings.ToLower(key)
	
	if strings.Contains(lowerKey, "theme") || strings.Contains(lowerKey, "color") ||
		strings.Contains(lowerKey, "font") || strings.Contains(lowerKey, "opacity") {
		return "appearance"
	}
	
	if strings.Contains(lowerKey, "window") || strings.Contains(lowerKey, "size") ||
		strings.Contains(lowerKey, "position") {
		return "window"
	}
	
	if strings.Contains(lowerKey, "performance") || strings.Contains(lowerKey, "gpu") ||
		strings.Contains(lowerKey, "cache") {
		return "performance"
	}
	
	return "general"
}

func (m *HuhFancyConfigModel) formatCategoryName(category string) string {
	switch category {
	case "appearance":
		return "🎨 Appearance"
	case "window":
		return "🪟 Window"
	case "performance":
		return "🚀 Performance"
	default:
		return "⚙️  General"
	}
}

func (m *HuhFancyConfigModel) formatFieldTitle(field ConfigField) string {
	// Convert snake_case to Title Case
	words := strings.FieldsFunc(field.Key, func(c rune) bool {
		return c == '_' || c == '-' || c == '.'
	})
	
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	
	title := strings.Join(words, " ")
	
	if field.IsSet {
		title = "● " + title
	}
	
	return title
}

func (m *HuhFancyConfigModel) buildOptions(options []string) []huh.Option[string] {
	var opts []huh.Option[string]
	for _, opt := range options {
		opts = append(opts, huh.NewOption(opt, opt))
	}
	return opts
}

func (m *HuhFancyConfigModel) applyQuickSettings() {
	// Process quick form values
	if m.quickForm != nil {
		// Extract and save values
		for key := range m.quickSettings {
			if val := m.quickForm.Get(key); val != nil {
				m.changed[key] = val
			}
		}
	}
}

func (m *HuhFancyConfigModel) applyDetailedSettings() {
	// Process detailed form values
	if m.detailedForm != nil {
		for _, field := range m.fields {
			if val := m.detailedForm.Get(field.Key); val != nil {
				m.changed[field.Key] = val
			}
		}
	}
}

// Interface methods

func (m *HuhFancyConfigModel) Focus() {}
func (m *HuhFancyConfigModel) Blur() {}

func (m *HuhFancyConfigModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

func (m *HuhFancyConfigModel) IsValid() bool {
	return len(m.changed) > 0 || m.currentView == Completed
}

func (m *HuhFancyConfigModel) GetValues() map[string]interface{} {
	return m.changed
}

func (m *HuhFancyConfigModel) HasUnsavedChanges() bool {
	return len(m.changed) > 0 && m.currentView != Completed
}