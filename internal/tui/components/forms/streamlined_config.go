package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StreamlinedConfigModel provides a simplified, fast configuration interface
type StreamlinedConfigModel struct {
	appName string
	fields  []ConfigField
	
	// UI state
	currentView  StreamlinedView
	cursor       int
	searchInput  textinput.Model
	filterText   string
	
	// Field organization
	changedFields   []int // Indices of changed fields
	currentFields   []int // Indices of currently set fields
	availableFields []int // Indices of available but unset fields
	filteredFields  []int // Currently visible fields after search
	
	// Values
	values         map[string]string
	originalValues map[string]string
	
	// Display settings
	width     int
	height    int
	scrollPos int
	
	// Styles
	styles StreamlinedStyles
}

// StreamlinedView represents different interface modes
type StreamlinedView int

const (
	StreamMainView StreamlinedView = iota
	StreamSearchView
	StreamEditView
	StreamPreviewView
)

// StreamlinedStyles for the interface
type StreamlinedStyles struct {
	Title          lipgloss.Style
	ActiveField    lipgloss.Style
	InactiveField  lipgloss.Style
	ChangedField   lipgloss.Style
	SearchBar      lipgloss.Style
	ShortcutKey    lipgloss.Style
	ShortcutDesc   lipgloss.Style
	Value          lipgloss.Style
	DefaultValue   lipgloss.Style
	Preview        lipgloss.Style
	StatusBar      lipgloss.Style
}

// NewStreamlinedConfig creates a streamlined configuration interface
func NewStreamlinedConfig(appName string) *StreamlinedConfigModel {
	searchInput := textinput.New()
	searchInput.Placeholder = "Type to filter options..."
	searchInput.CharLimit = 50
	
	return &StreamlinedConfigModel{
		appName:        appName,
		currentView:    StreamMainView,
		searchInput:    searchInput,
		values:         make(map[string]string),
		originalValues: make(map[string]string),
		styles:         createStreamlinedStyles(),
		width:          80,
		height:         24,
	}
}

// SetFields configures the fields and automatically organizes them
func (m *StreamlinedConfigModel) SetFields(fields []ConfigField) {
	m.fields = fields
	m.changedFields = []int{}
	m.currentFields = []int{}
	m.availableFields = []int{}
	
	// Automatically organize fields by their state
	for i, field := range fields {
		// Store original values
		if field.Value != nil {
			m.originalValues[field.Key] = fmt.Sprintf("%v", field.Value)
			m.values[field.Key] = m.originalValues[field.Key]
		}
		
		// Categorize fields
		if field.IsSet {
			m.currentFields = append(m.currentFields, i)
		} else {
			m.availableFields = append(m.availableFields, i)
		}
	}
	
	// Initially show all fields
	m.updateFilteredFields()
}

// updateFilteredFields updates which fields are visible based on search
func (m *StreamlinedConfigModel) updateFilteredFields() {
	m.filteredFields = []int{}
	
	// If no filter, show changed fields first, then current, then available
	if m.filterText == "" {
		// Changed fields first (most important)
		m.filteredFields = append(m.filteredFields, m.changedFields...)
		
		// Then current fields (already configured)
		for _, idx := range m.currentFields {
			if !contains(m.changedFields, idx) {
				m.filteredFields = append(m.filteredFields, idx)
			}
		}
		
		// Then available fields (not yet configured)
		m.filteredFields = append(m.filteredFields, m.availableFields...)
	} else {
		// Filter all fields by search text
		searchLower := strings.ToLower(m.filterText)
		for i, field := range m.fields {
			keyLower := strings.ToLower(field.Key)
			descLower := strings.ToLower(field.Description)
			
			if strings.Contains(keyLower, searchLower) || strings.Contains(descLower, searchLower) {
				m.filteredFields = append(m.filteredFields, i)
			}
		}
	}
}

// Init initializes the model
func (m *StreamlinedConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m *StreamlinedConfigModel) Update(msg tea.Msg) (*StreamlinedConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	
	return m, nil
}

// handleKeyPress manages keyboard input
func (m *StreamlinedConfigModel) handleKeyPress(msg tea.KeyMsg) (*StreamlinedConfigModel, tea.Cmd) {
	switch m.currentView {
	case StreamSearchView:
		return m.handleSearchInput(msg)
	case StreamEditView:
		return m.handleEditInput(msg)
	default:
		return m.handleMainNavigation(msg)
	}
}

// handleMainNavigation handles navigation in the main view
func (m *StreamlinedConfigModel) handleMainNavigation(msg tea.KeyMsg) (*StreamlinedConfigModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		
	case "down", "j":
		if m.cursor < len(m.filteredFields)-1 {
			m.cursor++
			m.ensureVisible()
		}
		
	case "pgup":
		m.cursor = max(0, m.cursor-10)
		m.ensureVisible()
		
	case "pgdown":
		m.cursor = min(len(m.filteredFields)-1, m.cursor+10)
		m.ensureVisible()
		
	case "/", "ctrl+f":
		// Start search
		m.currentView = StreamSearchView
		m.searchInput.Focus()
		return m, textinput.Blink
		
	case "enter", " ", "e":
		// Edit selected field
		if m.cursor < len(m.filteredFields) {
			m.currentView = StreamEditView
		}
		
	case "r":
		// Reset field to original value
		if m.cursor < len(m.filteredFields) {
			idx := m.filteredFields[m.cursor]
			field := m.fields[idx]
			if original, ok := m.originalValues[field.Key]; ok {
				m.values[field.Key] = original
				m.updateChangedFields()
			}
		}
		
	case "d":
		// Delete/unset field
		if m.cursor < len(m.filteredFields) {
			idx := m.filteredFields[m.cursor]
			field := m.fields[idx]
			delete(m.values, field.Key)
			m.updateChangedFields()
		}
		
	case "c":
		// Toggle showing only changed fields
		if len(m.changedFields) > 0 {
			if len(m.filteredFields) == len(m.changedFields) {
				// Currently showing only changed, show all
				m.filteredFields = nil
				m.updateFilteredFields()
			} else {
				// Show only changed
				m.filteredFields = m.changedFields
			}
			m.cursor = 0
		}
		
	case "a":
		// Show all fields
		m.filterText = ""
		m.updateFilteredFields()
		m.cursor = 0
		
	case "s", "ctrl+s":
		// Save changes
		if len(m.changedFields) > 0 {
			return m, func() tea.Msg {
				return ConfigSavedMsg{
					AppName: m.appName,
					Values:  m.getChangedValues(),
				}
			}
		}
		
	case "p":
		// Preview changes
		if len(m.changedFields) > 0 {
			m.currentView = StreamPreviewView
		}
		
	case "q", "esc":
		return m, tea.Quit
	}
	
	return m, nil
}

// handleSearchInput handles search mode input
func (m *StreamlinedConfigModel) handleSearchInput(msg tea.KeyMsg) (*StreamlinedConfigModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = StreamMainView
		m.filterText = ""
		m.searchInput.SetValue("")
		m.updateFilteredFields()
		m.cursor = 0
		return m, nil
		
	case "enter":
		m.currentView = StreamMainView
		m.filterText = m.searchInput.Value()
		m.updateFilteredFields()
		m.cursor = 0
		return m, nil
	}
	
	// Update search input
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	
	// Live filter as user types
	m.filterText = m.searchInput.Value()
	m.updateFilteredFields()
	
	return m, cmd
}

// handleEditInput handles field editing
func (m *StreamlinedConfigModel) handleEditInput(msg tea.KeyMsg) (*StreamlinedConfigModel, tea.Cmd) {
	if m.cursor >= len(m.filteredFields) {
		m.currentView = StreamMainView
		return m, nil
	}
	
	idx := m.filteredFields[m.cursor]
	field := m.fields[idx]
	
	switch field.Type {
	case FieldTypeBool:
		// Toggle boolean
		current := m.values[field.Key] == "true"
		m.values[field.Key] = fmt.Sprintf("%v", !current)
		m.updateChangedFields()
		m.currentView = StreamMainView
		
	case FieldTypeSelect:
		// Cycle through options
		if len(field.Options) > 0 {
			current := m.values[field.Key]
			nextIdx := 0
			for i, opt := range field.Options {
				if opt == current {
					nextIdx = (i + 1) % len(field.Options)
					break
				}
			}
			m.values[field.Key] = field.Options[nextIdx]
			m.updateChangedFields()
		}
		m.currentView = StreamMainView
		
	default:
		// For string/number fields, we'd need a text input
		// For now, just go back
		m.currentView = StreamMainView
	}
	
	return m, nil
}

// View renders the interface
func (m *StreamlinedConfigModel) View() string {
	switch m.currentView {
	case StreamSearchView:
		return m.renderSearchView()
	case StreamPreviewView:
		return m.renderPreviewView()
	default:
		return m.renderMainView()
	}
}

// renderMainView renders the main configuration list
func (m *StreamlinedConfigModel) renderMainView() string {
	var content strings.Builder
	
	// Title bar
	title := fmt.Sprintf("⚙️  %s Configuration", m.appName)
	if len(m.changedFields) > 0 {
		title += fmt.Sprintf(" (%d changes)", len(m.changedFields))
	}
	content.WriteString(m.styles.Title.Render(title))
	content.WriteString("\n\n")
	
	// Calculate visible area
	headerHeight := 3
	footerHeight := 3
	listHeight := m.height - headerHeight - footerHeight
	
	// Render visible fields
	visibleStart := m.scrollPos
	visibleEnd := min(visibleStart+listHeight, len(m.filteredFields))
	
	if len(m.filteredFields) == 0 {
		content.WriteString(m.styles.InactiveField.Render("No matching options. Press / to search or A to show all."))
	} else {
		for i := visibleStart; i < visibleEnd; i++ {
			idx := m.filteredFields[i]
			field := m.fields[idx]
			
			// Determine field style
			style := m.styles.InactiveField
			if i == m.cursor {
				style = m.styles.ActiveField
			} else if contains(m.changedFields, idx) {
				style = m.styles.ChangedField
			}
			
			// Format field display
			value := m.values[field.Key]
			if value == "" && field.Value != nil {
				value = fmt.Sprintf("%v", field.Value)
			}
			
			// Status indicator
			status := " "
			if field.IsSet {
				status = "●"
			}
			if contains(m.changedFields, idx) {
				status = "✓"
			}
			
			// Build field line
			line := fmt.Sprintf("%s %-20s = %-15s", status, truncate(field.Key, 20), truncate(value, 15))
			if field.Description != "" && i == m.cursor {
				line += fmt.Sprintf("  # %s", truncate(field.Description, 30))
			}
			
			content.WriteString(style.Render(line))
			content.WriteString("\n")
		}
	}
	
	// Status bar with shortcuts
	content.WriteString("\n")
	shortcuts := []string{
		"↑↓ Navigate",
		"/ Search",
		"Enter Edit",
		"C Changed",
		"R Reset",
		"S Save",
		"Q Quit",
	}
	
	statusBar := strings.Join(shortcuts, " • ")
	content.WriteString(m.styles.StatusBar.Render(statusBar))
	
	return content.String()
}

// renderSearchView renders the search interface
func (m *StreamlinedConfigModel) renderSearchView() string {
	var content strings.Builder
	
	content.WriteString(m.styles.Title.Render("🔍 Search Configuration"))
	content.WriteString("\n\n")
	content.WriteString(m.searchInput.View())
	content.WriteString("\n\n")
	
	// Show live results count
	if m.filterText != "" {
		resultCount := len(m.filteredFields)
		content.WriteString(fmt.Sprintf("Found %d matching options\n", resultCount))
	}
	
	content.WriteString("\nEnter: Apply Filter • Esc: Cancel")
	
	return content.String()
}

// renderPreviewView shows pending changes
func (m *StreamlinedConfigModel) renderPreviewView() string {
	var content strings.Builder
	
	content.WriteString(m.styles.Title.Render("📋 Preview Changes"))
	content.WriteString("\n\n")
	
	if len(m.changedFields) == 0 {
		content.WriteString("No changes to preview.\n")
	} else {
		for _, idx := range m.changedFields {
			field := m.fields[idx]
			oldVal := m.originalValues[field.Key]
			newVal := m.values[field.Key]
			
			if oldVal == "" {
				oldVal = "(unset)"
			}
			
			change := fmt.Sprintf("• %s: %s → %s\n", field.Key, oldVal, newVal)
			content.WriteString(m.styles.Preview.Render(change))
		}
	}
	
	content.WriteString("\nS: Save • Esc: Back")
	
	return content.String()
}

// Helper functions

func (m *StreamlinedConfigModel) ensureVisible() {
	if m.cursor < m.scrollPos {
		m.scrollPos = m.cursor
	} else if m.cursor >= m.scrollPos+m.height-6 {
		m.scrollPos = m.cursor - m.height + 7
	}
}

func (m *StreamlinedConfigModel) updateChangedFields() {
	m.changedFields = []int{}
	
	for i, field := range m.fields {
		currentVal := m.values[field.Key]
		originalVal := m.originalValues[field.Key]
		
		if currentVal != originalVal {
			m.changedFields = append(m.changedFields, i)
		}
	}
	
	m.updateFilteredFields()
}

func (m *StreamlinedConfigModel) getChangedValues() map[string]interface{} {
	changed := make(map[string]interface{})
	
	for _, idx := range m.changedFields {
		field := m.fields[idx]
		changed[field.Key] = m.values[field.Key]
	}
	
	return changed
}

// Utility functions

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func createStreamlinedStyles() StreamlinedStyles {
	return StreamlinedStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		
		ActiveField: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
		
		InactiveField: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		
		ChangedField: lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			Bold(true),
		
		SearchBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
		
		ShortcutKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),
		
		ShortcutDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")),
		
		DefaultValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),
		
		Preview: lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")),
		
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true),
	}
}

// Interface methods for compatibility

func (m *StreamlinedConfigModel) Focus() {
	// No-op for compatibility
}

func (m *StreamlinedConfigModel) Blur() {
	// No-op for compatibility
}

func (m *StreamlinedConfigModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

func (m *StreamlinedConfigModel) Bindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save"),
		),
	}
}

func (m *StreamlinedConfigModel) IsValid() bool {
	return true // Simplified validation
}

func (m *StreamlinedConfigModel) GetValues() map[string]interface{} {
	return m.getChangedValues()
}

func (m *StreamlinedConfigModel) HasUnsavedChanges() bool {
	return len(m.changedFields) > 0
}