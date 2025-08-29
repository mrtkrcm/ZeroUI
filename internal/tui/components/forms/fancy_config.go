package forms

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FancyConfigModel provides a beautiful, Charm-style configuration interface
type FancyConfigModel struct {
	appName string
	
	// UI Components
	list        list.Model
	spinner     spinner.Model
	help        help.Model
	textInput   textinput.Model
	
	// State
	activeTab   int
	tabs        []ConfigTab
	showHelp    bool
	editing     bool
	filtering   bool
	
	// Data
	items       []list.Item
	allItems    []list.Item
	changed     map[string]interface{}
	
	// Dimensions
	width  int
	height int
	
	// Keys
	keys FancyKeyMap
}

// ConfigTab represents a tab in the interface
type ConfigTab struct {
	Name  string
	Icon  string
	Count int
}

// FancyKeyMap contains all key bindings
type FancyKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Back     key.Binding
	Filter   key.Binding
	Save     key.Binding
	Reset    key.Binding
	Help     key.Binding
	Quit     key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
}

// ConfigItem represents a configuration item for the list
type ConfigItem struct {
	key         string
	value       string
	description string
	isChanged   bool
	isSet       bool
	fieldType   ConfigFieldType
}

func (i ConfigItem) Title() string {
	status := " "
	if i.isChanged {
		status = "✓"
	} else if i.isSet {
		status = "●"
	}
	
	title := fmt.Sprintf("%s %s", status, i.key)
	if i.value != "" {
		title = fmt.Sprintf("%s = %s", title, i.value)
	}
	return title
}

func (i ConfigItem) Description() string { 
	return i.description 
}

func (i ConfigItem) FilterValue() string { 
	return fmt.Sprintf("%s %s", i.key, i.description) 
}

// Styles for the fancy interface
var (
	// Window and layout
	docStyle = lipgloss.NewStyle().Padding(1, 2)
	
	// Title and header
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(1)
	
	// Tabs
	inactiveTabStyle = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)
	
	activeTabStyle = inactiveTabStyle.
		Bold(true).
		Foreground(lipgloss.Color("#FFF")).
		Background(lipgloss.Color("#7D56F4"))
	
	// List styles
	selectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)
	
	changedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))
	
	// Status messages
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
	
	// Help
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	
	// Tab border
	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}
)

// NewFancyConfig creates a beautiful configuration interface
func NewFancyConfig(appName string) *FancyConfigModel {
	// Initialize components
	items := []list.Item{}
	
	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle.Foreground(lipgloss.Color("#a0a0a0"))
	
	l := list.New(items, delegate, 0, 0)
	l.Title = fmt.Sprintf("Configure %s", appName)
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	
	// Create text input for inline editing
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 100
	
	// Create tabs
	tabs := []ConfigTab{
		{Name: "All", Icon: "📋", Count: 0},
		{Name: "Changed", Icon: "✓", Count: 0},
		{Name: "Set", Icon: "●", Count: 0},
		{Name: "Available", Icon: "○", Count: 0},
	}
	
	return &FancyConfigModel{
		appName:   appName,
		list:      l,
		spinner:   s,
		help:      help.New(),
		textInput: ti,
		tabs:      tabs,
		changed:   make(map[string]interface{}),
		keys:      newFancyKeyMap(),
		width:     80,
		height:    24,
	}
}

func newFancyKeyMap() FancyKeyMap {
	return FancyKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next tab"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
	}
}

// SetFields configures the fields for display
func (m *FancyConfigModel) SetFields(fields []ConfigField) {
	items := make([]list.Item, 0, len(fields))
	changedCount := 0
	setCount := 0
	availableCount := 0
	
	for _, field := range fields {
		item := ConfigItem{
			key:         field.Key,
			description: field.Description,
			isSet:       field.IsSet,
			fieldType:   field.Type,
		}
		
		// Set value
		if field.Value != nil {
			item.value = fmt.Sprintf("%v", field.Value)
		}
		
		// Check if changed
		if _, ok := m.changed[field.Key]; ok {
			item.isChanged = true
			changedCount++
		}
		
		if field.IsSet {
			setCount++
		} else {
			availableCount++
		}
		
		items = append(items, item)
	}
	
	m.allItems = items
	m.items = items
	m.list.SetItems(items)
	
	// Update tab counts
	m.tabs[0].Count = len(items)        // All
	m.tabs[1].Count = changedCount      // Changed
	m.tabs[2].Count = setCount          // Set
	m.tabs[3].Count = availableCount    // Available
}

// Init initializes the model
func (m *FancyConfigModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.list.StartSpinner(),
	)
}

// Update handles messages
func (m *FancyConfigModel) Update(msg tea.Msg) (*FancyConfigModel, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle window resize
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4) // Leave room for tabs
		
	case tea.KeyMsg:
		// Don't process keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}
		
		if m.editing {
			return m.handleEditMode(msg)
		}
		
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
			
		case key.Matches(msg, m.keys.NextTab), key.Matches(msg, m.keys.Right):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			m.filterByTab()
			return m, nil
			
		case key.Matches(msg, m.keys.PrevTab), key.Matches(msg, m.keys.Left):
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = len(m.tabs) - 1
			}
			m.filterByTab()
			return m, nil
			
		case key.Matches(msg, m.keys.Enter):
			// Start editing
			if item, ok := m.list.SelectedItem().(ConfigItem); ok {
				m.editing = true
				m.textInput.SetValue(item.value)
				m.textInput.Focus()
				return m, textinput.Blink
			}
			
		case key.Matches(msg, m.keys.Reset):
			// Reset selected item
			if item, ok := m.list.SelectedItem().(ConfigItem); ok {
				delete(m.changed, item.key)
				m.updateItemInList(item.key)
			}
			
		case key.Matches(msg, m.keys.Save):
			// Save changes
			if len(m.changed) > 0 {
				return m, func() tea.Msg {
					return ConfigSavedMsg{
						AppName: m.appName,
						Values:  m.changed,
					}
				}
			}
			
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			m.list.SetShowHelp(m.showHelp)
			return m, nil
		}
	}
	
	// Update spinner
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	
	// Update list
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	
	return m, tea.Batch(cmds...)
}

// handleEditMode handles input when editing a value
func (m *FancyConfigModel) handleEditMode(msg tea.KeyMsg) (*FancyConfigModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Save the value
		if item, ok := m.list.SelectedItem().(ConfigItem); ok {
			newValue := m.textInput.Value()
			if newValue != item.value {
				m.changed[item.key] = newValue
				m.updateItemInList(item.key)
			}
		}
		m.editing = false
		m.textInput.Blur()
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

// filterByTab filters items based on active tab
func (m *FancyConfigModel) filterByTab() {
	var filtered []list.Item
	
	switch m.activeTab {
	case 0: // All
		filtered = m.allItems
		
	case 1: // Changed
		for _, item := range m.allItems {
			if ci, ok := item.(ConfigItem); ok && ci.isChanged {
				filtered = append(filtered, item)
			}
		}
		
	case 2: // Set
		for _, item := range m.allItems {
			if ci, ok := item.(ConfigItem); ok && ci.isSet && !ci.isChanged {
				filtered = append(filtered, item)
			}
		}
		
	case 3: // Available
		for _, item := range m.allItems {
			if ci, ok := item.(ConfigItem); ok && !ci.isSet {
				filtered = append(filtered, item)
			}
		}
	}
	
	m.items = filtered
	m.list.SetItems(filtered)
}

// updateItemInList updates a specific item in the list
func (m *FancyConfigModel) updateItemInList(key string) {
	// Update counts
	m.tabs[1].Count = len(m.changed)
	
	// Refresh the list
	m.filterByTab()
}

// View renders the interface
func (m *FancyConfigModel) View() string {
	if m.editing {
		return m.renderEditView()
	}
	
	// Render tabs
	var renderedTabs []string
	for i, tab := range m.tabs {
		style := inactiveTabStyle
		if i == m.activeTab {
			style = activeTabStyle
		}
		
		tabText := fmt.Sprintf("%s %s", tab.Icon, tab.Name)
		if tab.Count > 0 {
			tabText = fmt.Sprintf("%s (%d)", tabText, tab.Count)
		}
		
		renderedTabs = append(renderedTabs, style.Render(tabText))
	}
	
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	// Status line
	status := ""
	if len(m.changed) > 0 {
		status = statusStyle.Render(fmt.Sprintf("✓ %d changes pending", len(m.changed)))
	}
	
	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		m.list.View(),
		status,
	)
	
	// Help at bottom
	if m.showHelp {
		helpView := m.help.View(m.keys)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			helpView,
		)
	} else {
		shortHelp := helpStyle.Render("? help • / filter • enter edit • ctrl+s save • q quit")
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			shortHelp,
		)
	}
	
	return docStyle.Render(content)
}

// renderEditView renders the inline edit interface
func (m *FancyConfigModel) renderEditView() string {
	item, _ := m.list.SelectedItem().(ConfigItem)
	
	editBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(60).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Editing: %s", item.key)),
				"",
				m.textInput.View(),
				"",
				helpStyle.Render("enter: save • esc: cancel"),
			),
		)
	
	return docStyle.Render(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			editBox,
		),
	)
}

// Interface methods for compatibility

func (m *FancyConfigModel) Focus() {}
func (m *FancyConfigModel) Blur() {}

func (m *FancyConfigModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	h, v := docStyle.GetFrameSize()
	m.list.SetSize(width-h, height-v-4)
	return nil
}

func (m *FancyConfigModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keys.Up,
		m.keys.Down,
		m.keys.Enter,
		m.keys.Filter,
		m.keys.Save,
		m.keys.Help,
		m.keys.Quit,
	}
}

func (m *FancyConfigModel) IsValid() bool {
	return true
}

func (m *FancyConfigModel) GetValues() map[string]interface{} {
	return m.changed
}

func (m *FancyConfigModel) HasUnsavedChanges() bool {
	return len(m.changed) > 0
}

// ShortHelp returns short help
func (k FancyKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns full help
func (k FancyKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Save, k.Reset, k.Filter},
		{k.Help, k.Quit},
	}
}