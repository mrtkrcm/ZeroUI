package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigCategory represents high-level user goals
type ConfigCategory struct {
	ID          string
	Name        string
	Icon        string
	Description string
	Presets     []ConfigPreset
	Fields      []ConfigField
}

// ConfigPreset represents common configurations for quick setup
type ConfigPreset struct {
	ID          string
	Name        string
	Description string
	Values      map[string]interface{}
	Icon        string
}

// IntuitiveConfigModel provides an intuitive, goal-oriented configuration interface
type IntuitiveConfigModel struct {
	appName    string
	categories []ConfigCategory
	
	// Navigation state
	currentView    ConfigView
	selectedCat    int
	selectedPreset int
	selectedField  int
	
	// UI state
	width     int
	height    int
	focused   bool
	showingPreview bool
	
	// Values and changes
	currentValues map[string]interface{}
	pendingChanges map[string]interface{}
	
	// Styles
	styles IntuitiveStyles
}

// ConfigView represents different interface states
type ConfigView int

const (
	CategoryView ConfigView = iota // Main category selection
	PresetView                     // Quick preset selection
	DetailView                     // Detailed field editing
	PreviewView                    // Preview changes
)

// IntuitiveStyles defines the visual styling
type IntuitiveStyles struct {
	CategoryCard     lipgloss.Style
	SelectedCard     lipgloss.Style
	PresetCard       lipgloss.Style
	Header           lipgloss.Style
	Description      lipgloss.Style
	Preview          lipgloss.Style
	QuickAction      lipgloss.Style
	Navigation       lipgloss.Style
}

// NewIntuitiveConfig creates a new intuitive configuration interface
func NewIntuitiveConfig(appName string) *IntuitiveConfigModel {
	return &IntuitiveConfigModel{
		appName:        appName,
		currentView:    CategoryView,
		currentValues:  make(map[string]interface{}),
		pendingChanges: make(map[string]interface{}),
		styles:         createIntuitiveStyles(),
		width:          80,
		height:         24,
	}
}

// SetFields organizes fields into intuitive categories
func (m *IntuitiveConfigModel) SetFields(fields []ConfigField) {
	m.categories = m.organizeByCategoryGoals(fields)
}

// organizeByCategoryGoals groups fields by what users want to accomplish
func (m *IntuitiveConfigModel) organizeByCategoryGoals(fields []ConfigField) []ConfigCategory {
	categories := []ConfigCategory{
		{
			ID:          "appearance",
			Name:        "Appearance & Themes",
			Icon:        "🎨",
			Description: "Colors, fonts, and visual style",
			Presets: []ConfigPreset{
				{
					ID:          "dark",
					Name:        "Dark Theme",
					Description: "Dark background with light text",
					Icon:        "🌙",
					Values: map[string]interface{}{
						"theme":            "dark",
						"background":       "#1e1e1e",
						"foreground":       "#ffffff",
					},
				},
				{
					ID:          "light",
					Name:        "Light Theme", 
					Description: "Light background with dark text",
					Icon:        "☀️",
					Values: map[string]interface{}{
						"theme":            "light",
						"background":       "#ffffff",
						"foreground":       "#000000",
					},
				},
				{
					ID:          "productivity",
					Name:        "Productivity",
					Description: "Focused, distraction-free setup",
					Icon:        "💼",
					Values: map[string]interface{}{
						"theme":                "auto",
						"window-decoration":    false,
						"background-opacity":   1.0,
					},
				},
			},
		},
		{
			ID:          "window",
			Name:        "Window & Layout",
			Icon:        "🪟",
			Description: "Window size, position, and behavior",
			Presets: []ConfigPreset{
				{
					ID:          "fullscreen",
					Name:        "Fullscreen",
					Description: "Take up the entire screen",
					Icon:        "⛶",
					Values: map[string]interface{}{
						"window-height":     "100%",
						"window-width":      "100%",
						"window-position":   "center",
					},
				},
				{
					ID:          "compact",
					Name:        "Compact",
					Description: "Small window for side-by-side use",
					Icon:        "📱",
					Values: map[string]interface{}{
						"window-height":     600,
						"window-width":      400,
						"window-position":   "center",
					},
				},
			},
		},
		{
			ID:          "performance",
			Name:        "Performance & Speed",
			Icon:        "🚀",
			Description: "Optimize for speed and responsiveness",
			Presets: []ConfigPreset{
				{
					ID:          "gaming",
					Name:        "Gaming Mode",
					Description: "Low latency, high performance",
					Icon:        "🎮",
					Values: map[string]interface{}{
						"gpu-renderer":           true,
						"fps-cap":               144,
						"input-latency":         "low",
					},
				},
				{
					ID:          "battery",
					Name:        "Battery Saver",
					Description: "Optimize for battery life",
					Icon:        "🔋",
					Values: map[string]interface{}{
						"gpu-renderer":           false,
						"fps-cap":               30,
						"background-opacity":     0.95,
					},
				},
			},
		},
	}
	
	// Distribute fields into appropriate categories
	for _, field := range fields {
		catID := m.determineFieldCategory(field.Key)
		for i := range categories {
			if categories[i].ID == catID {
				categories[i].Fields = append(categories[i].Fields, field)
				break
			}
		}
	}
	
	return categories
}

// determineFieldCategory maps field keys to user-goal categories
func (m *IntuitiveConfigModel) determineFieldCategory(key string) string {
	lowerKey := strings.ToLower(key)
	
	// Appearance-related
	if strings.Contains(lowerKey, "theme") || strings.Contains(lowerKey, "color") ||
		strings.Contains(lowerKey, "background") || strings.Contains(lowerKey, "foreground") ||
		strings.Contains(lowerKey, "font") || strings.Contains(lowerKey, "opacity") {
		return "appearance"
	}
	
	// Window-related  
	if strings.Contains(lowerKey, "window") || strings.Contains(lowerKey, "size") ||
		strings.Contains(lowerKey, "position") || strings.Contains(lowerKey, "decoration") {
		return "window"
	}
	
	// Performance-related
	if strings.Contains(lowerKey, "gpu") || strings.Contains(lowerKey, "fps") ||
		strings.Contains(lowerKey, "renderer") || strings.Contains(lowerKey, "performance") ||
		strings.Contains(lowerKey, "latency") {
		return "performance"
	}
	
	return "appearance" // Default category
}

// Init initializes the model
func (m *IntuitiveConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles user interactions
func (m *IntuitiveConfigModel) Update(msg tea.Msg) (*IntuitiveConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	
	return m, nil
}

// handleKeyPress manages navigation and interaction
func (m *IntuitiveConfigModel) handleKeyPress(msg tea.KeyMsg) (*IntuitiveConfigModel, tea.Cmd) {
	switch m.currentView {
	case CategoryView:
		return m.handleCategoryNavigation(msg)
	case PresetView:
		return m.handlePresetNavigation(msg)
	case DetailView:
		return m.handleDetailNavigation(msg)
	case PreviewView:
		return m.handlePreviewNavigation(msg)
	}
	return m, nil
}

// handleCategoryNavigation handles category selection
func (m *IntuitiveConfigModel) handleCategoryNavigation(msg tea.KeyMsg) (*IntuitiveConfigModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedCat > 0 {
			m.selectedCat--
		}
	case "down", "j":
		if m.selectedCat < len(m.categories)-1 {
			m.selectedCat++
		}
	case "enter", " ":
		// Enter the selected category's preset view
		m.currentView = PresetView
		m.selectedPreset = 0
	case "d":
		// Go directly to detailed configuration
		m.currentView = DetailView
		m.selectedField = 0
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

// handlePresetNavigation handles preset selection
func (m *IntuitiveConfigModel) handlePresetNavigation(msg tea.KeyMsg) (*IntuitiveConfigModel, tea.Cmd) {
	if m.selectedCat >= len(m.categories) {
		return m, nil
	}
	
	presets := m.categories[m.selectedCat].Presets
	
	switch msg.String() {
	case "up", "k":
		if m.selectedPreset > 0 {
			m.selectedPreset--
		}
	case "down", "j":
		if m.selectedPreset < len(presets)-1 {
			m.selectedPreset++
		}
	case "enter", " ":
		// Apply the selected preset
		if m.selectedPreset < len(presets) {
			preset := presets[m.selectedPreset]
			m.applyPreset(preset)
			m.currentView = PreviewView
		}
	case "d":
		// Go to detailed view for this category
		m.currentView = DetailView
		m.selectedField = 0
	case "esc":
		m.currentView = CategoryView
	}
	return m, nil
}

// handleDetailNavigation handles detailed field editing
func (m *IntuitiveConfigModel) handleDetailNavigation(msg tea.KeyMsg) (*IntuitiveConfigModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = CategoryView
	case "p":
		m.currentView = PreviewView
	}
	return m, nil
}

// handlePreviewNavigation handles preview and save
func (m *IntuitiveConfigModel) handlePreviewNavigation(msg tea.KeyMsg) (*IntuitiveConfigModel, tea.Cmd) {
	switch msg.String() {
	case "s", "enter":
		// Save the changes
		return m, func() tea.Msg {
			return ConfigSavedMsg{
				AppName: m.appName,
				Values:  m.pendingChanges,
			}
		}
	case "esc":
		// Cancel changes
		m.pendingChanges = make(map[string]interface{})
		m.currentView = CategoryView
	}
	return m, nil
}

// applyPreset applies preset values to pending changes
func (m *IntuitiveConfigModel) applyPreset(preset ConfigPreset) {
	for key, value := range preset.Values {
		m.pendingChanges[key] = value
	}
}

// View renders the current interface
func (m *IntuitiveConfigModel) View() string {
	switch m.currentView {
	case CategoryView:
		return m.renderCategoryView()
	case PresetView:
		return m.renderPresetView()
	case DetailView:
		return m.renderDetailView()
	case PreviewView:
		return m.renderPreviewView()
	}
	return ""
}

// renderCategoryView shows the main category selection
func (m *IntuitiveConfigModel) renderCategoryView() string {
	var content strings.Builder
	
	// Header
	header := m.styles.Header.Render(fmt.Sprintf("⚙️  Configure %s", m.appName))
	content.WriteString(header)
	content.WriteString("\n\nWhat would you like to configure?\n\n")
	
	// Category cards
	for i, cat := range m.categories {
		style := m.styles.CategoryCard
		if i == m.selectedCat {
			style = m.styles.SelectedCard
		}
		
		card := style.Render(fmt.Sprintf("%s %s\n%s", cat.Icon, cat.Name, cat.Description))
		content.WriteString(card)
		content.WriteString("\n")
	}
	
	// Help text
	help := m.styles.Navigation.Render("↑/↓: Navigate • Enter: Quick Setup • D: Detailed • Q: Quit")
	content.WriteString("\n")
	content.WriteString(help)
	
	return content.String()
}

// renderPresetView shows quick preset options
func (m *IntuitiveConfigModel) renderPresetView() string {
	if m.selectedCat >= len(m.categories) {
		return "Invalid category"
	}
	
	cat := m.categories[m.selectedCat]
	var content strings.Builder
	
	// Breadcrumb
	breadcrumb := m.styles.Header.Render(fmt.Sprintf("Configure › %s %s", cat.Icon, cat.Name))
	content.WriteString(breadcrumb)
	content.WriteString("\n\nChoose a quick setup:\n\n")
	
	// Preset cards
	for i, preset := range cat.Presets {
		style := m.styles.PresetCard
		if i == m.selectedPreset {
			style = m.styles.SelectedCard
		}
		
		card := style.Render(fmt.Sprintf("%s %s\n%s", preset.Icon, preset.Name, preset.Description))
		content.WriteString(card)
		content.WriteString("\n")
	}
	
	// Help text
	help := m.styles.Navigation.Render("↑/↓: Navigate • Enter: Apply • D: Custom • Esc: Back")
	content.WriteString("\n")
	content.WriteString(help)
	
	return content.String()
}

// renderDetailView shows detailed field editing (simplified for now)
func (m *IntuitiveConfigModel) renderDetailView() string {
	if m.selectedCat >= len(m.categories) {
		return "Invalid category"
	}
	
	cat := m.categories[m.selectedCat]
	var content strings.Builder
	
	// Breadcrumb
	breadcrumb := m.styles.Header.Render(fmt.Sprintf("Configure › %s %s › Custom", cat.Icon, cat.Name))
	content.WriteString(breadcrumb)
	content.WriteString("\n\n")
	
	// Show fields for this category
	if len(cat.Fields) == 0 {
		content.WriteString("No detailed options available for this category.\n")
		content.WriteString("Try the quick presets instead!\n")
	} else {
		content.WriteString("Detailed configuration options:\n\n")
		for _, field := range cat.Fields {
			line := fmt.Sprintf("• %s: %v\n", field.Key, field.Value)
			content.WriteString(line)
		}
	}
	
	// Help text
	help := m.styles.Navigation.Render("P: Preview • Esc: Back")
	content.WriteString("\n")
	content.WriteString(help)
	
	return content.String()
}

// renderPreviewView shows changes before saving
func (m *IntuitiveConfigModel) renderPreviewView() string {
	var content strings.Builder
	
	// Header
	header := m.styles.Header.Render("📋 Preview Changes")
	content.WriteString(header)
	content.WriteString("\n\n")
	
	if len(m.pendingChanges) == 0 {
		content.WriteString("No changes to preview.\n")
	} else {
		content.WriteString("The following settings will be updated:\n\n")
		for key, value := range m.pendingChanges {
			line := m.styles.Preview.Render(fmt.Sprintf("✓ %s → %v", key, value))
			content.WriteString(line)
			content.WriteString("\n")
		}
	}
	
	// Help text
	help := m.styles.Navigation.Render("S/Enter: Save • Esc: Cancel")
	content.WriteString("\n")
	content.WriteString(help)
	
	return content.String()
}

// createIntuitiveStyles creates the visual styling
func createIntuitiveStyles() IntuitiveStyles {
	return IntuitiveStyles{
		CategoryCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Margin(0, 1).
			Width(30),
		
		SelectedCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("212")).
			Background(lipgloss.Color("235")).
			Padding(1, 2).
			Margin(0, 1).
			Width(30).
			Bold(true),
		
		PresetCard: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Margin(0, 1).
			Width(35),
		
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1),
		
		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true),
		
		Preview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			MarginLeft(2),
		
		QuickAction: lipgloss.NewStyle().
			Background(lipgloss.Color("212")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true),
		
		Navigation: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true),
	}
}

// Focus sets focus on the component
func (m *IntuitiveConfigModel) Focus() {
	m.focused = true
}

// Blur removes focus from the component
func (m *IntuitiveConfigModel) Blur() {
	m.focused = false
}

// SetSize updates component dimensions
func (m *IntuitiveConfigModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// Bindings returns key bindings for help
func (m *IntuitiveConfigModel) Bindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "navigate up"),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "navigate down"),
		),
		key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "select"),
		),
		key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "detailed options"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
	}
}

// IsValid returns whether the current state is valid
func (m *IntuitiveConfigModel) IsValid() bool {
	return true // Simplified for now
}

// GetValues returns pending changes
func (m *IntuitiveConfigModel) GetValues() map[string]interface{} {
	return m.pendingChanges
}

// HasUnsavedChanges returns whether there are pending changes
func (m *IntuitiveConfigModel) HasUnsavedChanges() bool {
	return len(m.pendingChanges) > 0
}