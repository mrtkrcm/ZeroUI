package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabbedConfigModel provides a clean tab-based configuration interface
type TabbedConfigModel struct {
	appName string
	fields  []ConfigField
	
	// Tabs
	tabs       []Tab
	activeTab  int
	
	// Current view items
	items         []TabbedConfigItem
	cursor        int
	scrollOffset  int
	
	// Editing
	editing       bool
	editingIndex  int
	textInput     textinput.Model
	
	// Changes tracking
	values        map[string]string
	originalValues map[string]string
	changed       map[string]bool
	
	// UI
	width  int
	height int
	
	// Key bindings
	keys TabKeyMap
}

// Tab represents a tab in the interface
type Tab struct {
	Name  string
	Icon  string
	Count int
	Filter func([]ConfigField, map[string]bool) []TabbedConfigItem
}

// TabbedConfigItem for display
type TabbedConfigItem struct {
	Field     ConfigField
	Value     string
	IsChanged bool
}

// TabKeyMap contains key bindings
type TabKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Edit     key.Binding
	Save     key.Binding
	Reset    key.Binding
	Quit     key.Binding
}

// Tab styles
var (
	tabBorderWithBottom = func(left, middle, right string) lipgloss.Border {
		border := lipgloss.RoundedBorder()
		border.BottomLeft = left
		border.Bottom = middle
		border.BottomRight = right
		return border
	}
	
	tabbedInactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	tabbedActiveTabBorder   = tabBorderWithBottom("┘", " ", "└")
	
	tabbedDocStyle          = lipgloss.NewStyle().Padding(1, 2)
	tabbedHighlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	tabbedInactiveTabStyle  = lipgloss.NewStyle().Border(tabbedInactiveTabBorder, true).BorderForeground(tabbedHighlightColor).Padding(0, 1)
	tabbedActiveTabStyle    = tabbedInactiveTabStyle.Border(tabbedActiveTabBorder, true)
	tabbedWindowStyle       = lipgloss.NewStyle().BorderForeground(tabbedHighlightColor).Padding(1, 2).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	
	tabbedSelectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	tabbedChangedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	tabbedNormalStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	tabbedDimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	tabbedStatusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
)

// NewTabbedConfig creates a clean tabbed configuration interface
func NewTabbedConfig(appName string) *TabbedConfigModel {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 100
	
	m := &TabbedConfigModel{
		appName:        appName,
		values:         make(map[string]string),
		originalValues: make(map[string]string),
		changed:        make(map[string]bool),
		textInput:      ti,
		width:          80,
		height:         24,
		keys:           newTabKeyMap(),
	}
	
	// Define tabs with automatic filters
	m.tabs = []Tab{
		{
			Name: "All",
			Icon: "📋",
			Filter: func(fields []ConfigField, changed map[string]bool) []TabbedConfigItem {
				items := make([]TabbedConfigItem, 0, len(fields))
				for _, f := range fields {
					item := TabbedConfigItem{
						Field:     f,
						Value:     m.getValue(f.Key),
						IsChanged: changed[f.Key],
					}
					items = append(items, item)
				}
				return items
			},
		},
		{
			Name: "Changed",
			Icon: "✓",
			Filter: func(fields []ConfigField, changed map[string]bool) []TabbedConfigItem {
				var items []TabbedConfigItem
				for _, f := range fields {
					if changed[f.Key] {
						items = append(items, TabbedConfigItem{
							Field:     f,
							Value:     m.getValue(f.Key),
							IsChanged: true,
						})
					}
				}
				return items
			},
		},
		{
			Name: "Set",
			Icon: "●",
			Filter: func(fields []ConfigField, changed map[string]bool) []TabbedConfigItem {
				var items []TabbedConfigItem
				for _, f := range fields {
					if f.IsSet && !changed[f.Key] {
						items = append(items, TabbedConfigItem{
							Field:     f,
							Value:     m.getValue(f.Key),
							IsChanged: false,
						})
					}
				}
				return items
			},
		},
		{
			Name: "Available",
			Icon: "○",
			Filter: func(fields []ConfigField, changed map[string]bool) []TabbedConfigItem {
				var items []TabbedConfigItem
				for _, f := range fields {
					if !f.IsSet {
						items = append(items, TabbedConfigItem{
							Field:     f,
							Value:     m.getValue(f.Key),
							IsChanged: changed[f.Key],
						})
					}
				}
				return items
			},
		},
	}
	
	return m
}

func newTabKeyMap() TabKeyMap {
	return TabKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "l"),
			key.WithHelp("tab/l", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "h"),
			key.WithHelp("shift+tab/h", "prev tab"),
		),
		Edit: key.NewBinding(
			key.WithKeys("enter", "e"),
			key.WithHelp("enter", "edit"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}

// SetFields configures the fields
func (m *TabbedConfigModel) SetFields(fields []ConfigField) {
	m.fields = fields
	
	// Initialize values
	for _, field := range fields {
		if field.Value != nil {
			value := fmt.Sprintf("%v", field.Value)
			m.originalValues[field.Key] = value
			m.values[field.Key] = value
		} else {
			m.originalValues[field.Key] = ""
			m.values[field.Key] = ""
		}
	}
	
	m.updateTabCounts()
	m.refreshItems()
}

// getValue gets the current value for a field
func (m *TabbedConfigModel) getValue(key string) string {
	if val, ok := m.values[key]; ok {
		return val
	}
	return ""
}

// updateTabCounts updates the count for each tab
func (m *TabbedConfigModel) updateTabCounts() {
	// All
	m.tabs[0].Count = len(m.fields)
	
	// Changed
	m.tabs[1].Count = len(m.changed)
	
	// Set
	setCount := 0
	for _, f := range m.fields {
		if f.IsSet && !m.changed[f.Key] {
			setCount++
		}
	}
	m.tabs[2].Count = setCount
	
	// Available
	availableCount := 0
	for _, f := range m.fields {
		if !f.IsSet {
			availableCount++
		}
	}
	m.tabs[3].Count = availableCount
}

// refreshItems refreshes the current view items based on active tab
func (m *TabbedConfigModel) refreshItems() {
	if m.activeTab >= len(m.tabs) {
		return
	}
	
	m.items = m.tabs[m.activeTab].Filter(m.fields, m.changed)
	
	// Reset cursor if out of bounds
	if m.cursor >= len(m.items) {
		m.cursor = max(0, len(m.items)-1)
	}
}

// Init initializes the model
func (m *TabbedConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *TabbedConfigModel) Update(msg tea.Msg) (*TabbedConfigModel, tea.Cmd) {
	// Handle editing mode
	if m.editing {
		return m.handleEditMode(msg)
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
			
		case key.Matches(msg, m.keys.NextTab):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			m.cursor = 0
			m.scrollOffset = 0
			m.refreshItems()
			
		case key.Matches(msg, m.keys.PrevTab):
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
			m.cursor = 0
			m.scrollOffset = 0
			m.refreshItems()
			
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
			
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.ensureVisible()
			}
			
		case key.Matches(msg, m.keys.Edit):
			if m.cursor < len(m.items) {
				m.startEditing()
				return m, textinput.Blink
			}
			
		case key.Matches(msg, m.keys.Reset):
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				m.values[item.Field.Key] = m.originalValues[item.Field.Key]
				delete(m.changed, item.Field.Key)
				m.updateTabCounts()
				m.refreshItems()
			}
			
		case key.Matches(msg, m.keys.Save):
			if len(m.changed) > 0 {
				return m, func() tea.Msg {
					values := make(map[string]interface{})
					for key := range m.changed {
						values[key] = m.values[key]
					}
					return ConfigSavedMsg{
						AppName: m.appName,
						Values:  values,
					}
				}
			}
		}
	}
	
	return m, nil
}

// handleEditMode handles editing mode
func (m *TabbedConfigModel) handleEditMode(msg tea.Msg) (*TabbedConfigModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		// Update text input for non-key messages
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	
	// Handle key messages
	switch keyMsg.String() {
	case "enter":
		// Save the value
		newValue := m.textInput.Value()
		item := m.items[m.cursor]
		
		if newValue != m.originalValues[item.Field.Key] {
			m.values[item.Field.Key] = newValue
			m.changed[item.Field.Key] = true
		} else {
			m.values[item.Field.Key] = newValue
			delete(m.changed, item.Field.Key)
		}
		
		m.editing = false
		m.textInput.Blur()
		m.updateTabCounts()
		m.refreshItems()
		return m, nil
		
	case "esc":
		// Cancel editing
		m.editing = false
		m.textInput.Blur()
		return m, nil
	}
	
	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// startEditing starts editing the selected item
func (m *TabbedConfigModel) startEditing() {
	if m.cursor >= len(m.items) {
		return
	}
	
	item := m.items[m.cursor]
	
	// Handle different field types
	switch item.Field.Type {
	case FieldTypeBool:
		// Toggle boolean directly
		current := m.values[item.Field.Key] == "true"
		newValue := fmt.Sprintf("%v", !current)
		m.values[item.Field.Key] = newValue
		
		if newValue != m.originalValues[item.Field.Key] {
			m.changed[item.Field.Key] = true
		} else {
			delete(m.changed, item.Field.Key)
		}
		
		m.updateTabCounts()
		m.refreshItems()
		
	case FieldTypeSelect:
		// Cycle through options
		if len(item.Field.Options) > 0 {
			currentValue := m.values[item.Field.Key]
			nextIndex := 0
			
			for i, opt := range item.Field.Options {
				if opt == currentValue {
					nextIndex = (i + 1) % len(item.Field.Options)
					break
				}
			}
			
			newValue := item.Field.Options[nextIndex]
			m.values[item.Field.Key] = newValue
			
			if newValue != m.originalValues[item.Field.Key] {
				m.changed[item.Field.Key] = true
			} else {
				delete(m.changed, item.Field.Key)
			}
			
			m.updateTabCounts()
			m.refreshItems()
		}
		
	default:
		// Text input for other types
		m.editing = true
		m.editingIndex = m.cursor
		m.textInput.SetValue(item.Value)
		m.textInput.Focus()
		m.textInput.CursorEnd()
	}
}

// ensureVisible ensures the cursor is visible
func (m *TabbedConfigModel) ensureVisible() {
	viewHeight := m.height - 10 // Account for header, tabs, footer
	
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+viewHeight {
		m.scrollOffset = m.cursor - viewHeight + 1
	}
}

// View renders the interface
func (m *TabbedConfigModel) View() string {
	if m.editing {
		return m.renderEditView()
	}
	
	// Render tabs
	tabs := m.renderTabs()
	
	// Render items
	content := m.renderItems()
	
	// Status bar
	status := m.renderStatus()
	
	// Combine
	doc := strings.Builder{}
	doc.WriteString(tabs)
	doc.WriteString("\n")
	doc.WriteString(tabbedWindowStyle.Width(m.width - 4).Height(m.height - 8).Render(content))
	doc.WriteString("\n")
	doc.WriteString(status)
	
	return tabbedDocStyle.Render(doc.String())
}

// renderTabs renders the tab bar
func (m *TabbedConfigModel) renderTabs() string {
	var renderedTabs []string
	
	for i, tab := range m.tabs {
		var style lipgloss.Style
		isFirst := i == 0
		isLast := i == len(m.tabs)-1
		isActive := i == m.activeTab
		
		if isActive {
			style = tabbedActiveTabStyle
		} else {
			style = tabbedInactiveTabStyle
		}
		
		// Adjust borders for first/last tabs
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		
		// Tab text
		text := fmt.Sprintf("%s %s", tab.Icon, tab.Name)
		if tab.Count > 0 {
			text = fmt.Sprintf("%s (%d)", text, tab.Count)
		}
		
		renderedTabs = append(renderedTabs, style.Render(text))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

// renderItems renders the configuration items
func (m *TabbedConfigModel) renderItems() string {
	if len(m.items) == 0 {
		return tabbedDimStyle.Render("No items in this category")
	}
	
	var lines []string
	viewHeight := m.height - 10
	
	for i := m.scrollOffset; i < min(m.scrollOffset+viewHeight, len(m.items)); i++ {
		item := m.items[i]
		
		// Status indicator
		status := "  "
		if item.IsChanged {
			status = "✓ "
		} else if item.Field.IsSet {
			status = "● "
		} else {
			status = "○ "
		}
		
		// Format line
		key := item.Field.Key
		value := item.Value
		if value == "" && item.Field.Value != nil {
			value = fmt.Sprintf("%v", item.Field.Value)
		}
		if value == "" {
			value = "(not set)"
		}
		
		line := fmt.Sprintf("%s%-25s = %s", status, key, value)
		
		// Apply style
		var style lipgloss.Style
		if i == m.cursor {
			style = tabbedSelectedStyle
		} else if item.IsChanged {
			style = tabbedChangedStyle
		} else {
			style = tabbedNormalStyle
		}
		
		lines = append(lines, style.Render(line))
	}
	
	return strings.Join(lines, "\n")
}

// renderStatus renders the status bar
func (m *TabbedConfigModel) renderStatus() string {
	var parts []string
	
	// Changed count
	if len(m.changed) > 0 {
		parts = append(parts, tabbedStatusStyle.Render(fmt.Sprintf("✓ %d changes", len(m.changed))))
	}
	
	// Help
	help := "↑↓ navigate • tab switch • enter edit • r reset • ctrl+s save • q quit"
	parts = append(parts, tabbedDimStyle.Render(help))
	
	return strings.Join(parts, " • ")
}

// renderEditView renders the editing interface
func (m *TabbedConfigModel) renderEditView() string {
	if m.cursor >= len(m.items) {
		return ""
	}
	
	item := m.items[m.cursor]
	
	editBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tabbedHighlightColor).
		Padding(1).
		Width(60).
		Render(
			fmt.Sprintf("Editing: %s\n\n%s\n\nenter: save • esc: cancel",
				tabbedSelectedStyle.Render(item.Field.Key),
				m.textInput.View(),
			),
		)
	
	return tabbedDocStyle.Render(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			editBox,
		),
	)
}

// Interface methods

func (m *TabbedConfigModel) Focus() {}
func (m *TabbedConfigModel) Blur() {}

func (m *TabbedConfigModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

func (m *TabbedConfigModel) IsValid() bool {
	return true
}

func (m *TabbedConfigModel) GetValues() map[string]interface{} {
	values := make(map[string]interface{})
	for key := range m.changed {
		values[key] = m.values[key]
	}
	return values
}

func (m *TabbedConfigModel) HasUnsavedChanges() bool {
	return len(m.changed) > 0
}