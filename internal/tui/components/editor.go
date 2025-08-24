package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Helper functions for robustness

// safeString safely returns a string value, handling nil and type assertions
func safeString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// safeBounds ensures an index is within bounds
func safeBounds(index, length int) int {
	if length <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

// safeMapAccess safely accesses a map value
func safeMapAccess(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	val, _ := m[key]
	return val
}

// EnhancedConfigModel provides an improved configuration editing experience
type EnhancedConfigModel struct {
	appName string
	fields  []ConfigField
	
	// Enhanced tabs with search and filtering
	tabs          []EnhancedTab
	activeTab     int
	searchInput   textinput.Model
	searchMode    bool
	searchResults []EnhancedConfigItem
	
	// Current view items
	items        []EnhancedConfigItem
	cursor       int
	scrollOffset int
	
	// Enhanced editing with inline preview
	editing      bool
	editingIndex int
	editInput    textinput.Model
	previewValue string
	validation   ValidationResult
	
	// Multi-select for batch operations
	multiSelect  bool
	selected     map[string]bool
	
	// Values and changes
	values         map[string]string
	originalValues map[string]string
	changed        map[string]bool
	history        []ConfigChange
	
	// Enhanced UI elements
	spinner      spinner.Model
	showTooltip  bool
	tooltipText  string
	showPreview  bool
	previewPanel string
	
	// Dimensions
	width  int
	height int
	
	// Key bindings
	keys EnhancedConfigKeyMap
	
	// Pager for viewing original file
	pager          *ConfigPager
	showingSource  bool
	configFilePath string
	configContent  string
}

// EnhancedTab with additional features
type EnhancedTab struct {
	Name        string
	Icon        string
	Count       int
	Filter      func([]ConfigField, map[string]bool) []EnhancedConfigItem
	Description string
	Shortcut    string
}

// EnhancedConfigItem with more metadata
type EnhancedConfigItem struct {
	Field       ConfigField
	Value       string
	IsChanged   bool
	IsSelected  bool
	IsValid     bool
	Tooltip     string
	Preview     string
}

// ConfigChange for undo/redo history
type ConfigChange struct {
	Key       string
	OldValue  string
	NewValue  string
	Timestamp time.Time
}

// ValidationResult for field validation
type ValidationResult struct {
	IsValid bool
	Message string
	Level   string // "error", "warning", "info"
}

// EnhancedConfigKeyMap with more shortcuts
type EnhancedConfigKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	NextTab     key.Binding
	PrevTab     key.Binding
	Edit        key.Binding
	QuickEdit   key.Binding
	Search      key.Binding
	Filter      key.Binding
	MultiSelect key.Binding
	SelectAll   key.Binding
	Save        key.Binding
	Reset       key.Binding
	Undo        key.Binding
	Redo        key.Binding
	Copy        key.Binding
	Paste       key.Binding
	Help        key.Binding
	Quit        key.Binding
	ViewSource  key.Binding
}

// Enhanced styles with better visual hierarchy
var (
	enhancedBaseStyle = lipgloss.NewStyle().
		Padding(1, 2)
	
	enhancedAccentColor = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	enhancedSuccessColor = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	enhancedWarningColor = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FFB700"}
	enhancedErrorColor = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	
	enhancedTabStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(enhancedAccentColor).
		Padding(0, 2)
	
	enhancedActiveTabStyle = enhancedTabStyle.
		Background(enhancedAccentColor).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	enhancedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)
	
	enhancedSelectedStyle = enhancedItemStyle.
		Background(lipgloss.Color("#3C3C3C")).
		Foreground(enhancedAccentColor).
		Bold(true)
	
	enhancedChangedStyle = enhancedItemStyle.
		Foreground(enhancedSuccessColor)
	
	enhancedErrorStyle = enhancedItemStyle.
		Foreground(enhancedErrorColor)
	
	enhancedTooltipStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(0, 1).
		Foreground(lipgloss.Color("#FAFAFA"))
	
	enhancedSearchStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(enhancedAccentColor).
		Padding(0, 1)
	
	enhancedPreviewStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(enhancedAccentColor).
		Padding(1).
		MarginTop(1)
)

// NewEnhancedConfig creates an improved configuration interface
func NewEnhancedConfig(appName string) *EnhancedConfigModel {
	// Search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Type to search..."
	searchInput.CharLimit = 50
	
	// Edit input
	editInput := textinput.New()
	editInput.Placeholder = "Enter value..."
	editInput.CharLimit = 200
	
	// Spinner for async operations
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(enhancedAccentColor)
	
	m := &EnhancedConfigModel{
		appName:        appName,
		values:         make(map[string]string),
		originalValues: make(map[string]string),
		changed:        make(map[string]bool),
		selected:       make(map[string]bool),
		history:        []ConfigChange{},
		searchInput:    searchInput,
		editInput:      editInput,
		spinner:        s,
		width:          100,
		height:         30,
		keys:           newEnhancedKeyMap(),
		pager:          NewConfigPager(),
		showingSource:  false,
	}
	
	// Enhanced tabs with better organization
	m.tabs = []EnhancedTab{
		{
			Name:        "All",
			Icon:        "[A]",
			Description: "View all configuration options",
			Shortcut:    "1",
			Filter:      m.filterAll,
		},
		{
			Name:        "Modified",
			Icon:        "[M]",
			Description: "Show only modified settings",
			Shortcut:    "2",
			Filter:      m.filterModified,
		},
		{
			Name:        "Favorites",
			Icon:        "[*]",
			Description: "Frequently used settings",
			Shortcut:    "3",
			Filter:      m.filterFavorites,
		},
		{
			Name:        "Advanced",
			Icon:        "[+]",
			Description: "Advanced configuration options",
			Shortcut:    "4",
			Filter:      m.filterAdvanced,
		},
		{
			Name:        "Search",
			Icon:        "[/]",
			Description: "Search results",
			Shortcut:    "5",
			Filter:      m.filterSearch,
		},
	}
	
	return m
}

func newEnhancedKeyMap() EnhancedConfigKeyMap {
	return EnhancedConfigKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
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
			key.WithHelp("enter/e", "edit"),
		),
		QuickEdit: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "quick edit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+f"),
			key.WithHelp("//ctrl+f", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		MultiSelect: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "multi-select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "select all"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset"),
		),
		Undo: key.NewBinding(
			key.WithKeys("ctrl+z", "u"),
			key.WithHelp("ctrl+z/u", "undo"),
		),
		Redo: key.NewBinding(
			key.WithKeys("ctrl+y", "ctrl+r"),
			key.WithHelp("ctrl+y", "redo"),
		),
		Copy: key.NewBinding(
			key.WithKeys("ctrl+c", "y"),
			key.WithHelp("ctrl+c/y", "copy"),
		),
		Paste: key.NewBinding(
			key.WithKeys("ctrl+v", "p"),
			key.WithHelp("ctrl+v/p", "paste"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		ViewSource: key.NewBinding(
			key.WithKeys("v", "ctrl+o"),
			key.WithHelp("v/ctrl+o", "view source"),
		),
	}
}

// Filter functions
func (m *EnhancedConfigModel) filterAll(fields []ConfigField, changed map[string]bool) []EnhancedConfigItem {
	items := make([]EnhancedConfigItem, 0, len(fields))
	for _, f := range fields {
		items = append(items, m.createEnhancedItem(f, changed[f.Key]))
	}
	return items
}

func (m *EnhancedConfigModel) filterModified(fields []ConfigField, changed map[string]bool) []EnhancedConfigItem {
	var items []EnhancedConfigItem
	for _, f := range fields {
		if changed[f.Key] {
			items = append(items, m.createEnhancedItem(f, true))
		}
	}
	return items
}

func (m *EnhancedConfigModel) filterFavorites(fields []ConfigField, changed map[string]bool) []EnhancedConfigItem {
	var items []EnhancedConfigItem
	// Define favorite fields (commonly used)
	favorites := map[string]bool{
		"theme": true, "font": true, "font_size": true,
		"window_mode": true, "opacity": true, "auto_save": true,
	}
	
	for _, f := range fields {
		if favorites[f.Key] {
			items = append(items, m.createEnhancedItem(f, changed[f.Key]))
		}
	}
	return items
}

func (m *EnhancedConfigModel) filterAdvanced(fields []ConfigField, changed map[string]bool) []EnhancedConfigItem {
	var items []EnhancedConfigItem
	// Advanced fields (performance, debugging, etc.)
	for _, f := range fields {
		if strings.Contains(f.Key, "debug") || strings.Contains(f.Key, "performance") ||
			strings.Contains(f.Key, "cache") || strings.Contains(f.Key, "experimental") {
			items = append(items, m.createEnhancedItem(f, changed[f.Key]))
		}
	}
	return items
}

func (m *EnhancedConfigModel) filterSearch(fields []ConfigField, changed map[string]bool) []EnhancedConfigItem {
	if m.searchInput.Value() == "" {
		return []EnhancedConfigItem{}
	}
	
	query := strings.ToLower(m.searchInput.Value())
	var items []EnhancedConfigItem
	
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f.Key), query) ||
			strings.Contains(strings.ToLower(f.Description), query) {
			items = append(items, m.createEnhancedItem(f, changed[f.Key]))
		}
	}
	return items
}

func (m *EnhancedConfigModel) createEnhancedItem(field ConfigField, isChanged bool) EnhancedConfigItem {
	value := m.getValue(field.Key)
	if value == "" && field.Value != nil {
		value = fmt.Sprintf("%v", field.Value)
	}
	
	return EnhancedConfigItem{
		Field:      field,
		Value:      value,
		IsChanged:  isChanged,
		IsSelected: m.selected[field.Key],
		IsValid:    true, // Will be validated
		Tooltip:    field.Description,
		Preview:    m.generatePreview(field, value),
	}
}

func (m *EnhancedConfigModel) getValue(key string) string {
	if val, ok := m.values[key]; ok {
		return val
	}
	return ""
}

func (m *EnhancedConfigModel) generatePreview(field ConfigField, value string) string {
	// Generate context-aware preview
	switch field.Key {
	case "theme":
		return fmt.Sprintf("Theme will change to: %s", value)
	case "font_size":
		return fmt.Sprintf("Font size: %spx", value)
	case "opacity":
		return fmt.Sprintf("Window opacity: %s%%", value)
	default:
		return ""
	}
}

// SetConfigFile sets the path and content of the configuration file for viewing
func (m *EnhancedConfigModel) SetConfigFile(filePath, content string) {
	if m == nil {
		return
	}
	m.configFilePath = filePath
	m.configContent = content
	if m.pager != nil {
		m.pager.SetContent(filePath, content)
	}
}

// SetFields configures the fields
func (m *EnhancedConfigModel) SetFields(fields []ConfigField) {
	// Defensive: handle nil fields
	if fields == nil {
		fields = []ConfigField{}
	}
	if m == nil {
		return
	}
	
	m.fields = fields
	
	// Initialize values with nil checks
	if m.originalValues == nil {
		m.originalValues = make(map[string]string)
	}
	if m.values == nil {
		m.values = make(map[string]string)
	}
	
	for _, field := range fields {
		if field.Key == "" {
			continue // Skip invalid fields
		}
		if field.Value != nil {
			value := safeString(field.Value)
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

func (m *EnhancedConfigModel) updateTabCounts() {
	// Defensive checks
	if m == nil || m.tabs == nil || len(m.tabs) < 5 {
		return
	}
	// Update counts for each tab
	m.tabs[0].Count = len(m.fields) // All
	m.tabs[1].Count = len(m.changed) // Modified
	
	// Count favorites
	favoriteCount := 0
	favorites := []string{"theme", "font", "font_size", "window_mode", "opacity", "auto_save"}
	for _, key := range favorites {
		for _, field := range m.fields {
			if field.Key == key {
				favoriteCount++
				break
			}
		}
	}
	m.tabs[2].Count = favoriteCount
	
	// Count advanced
	advancedCount := 0
	for _, field := range m.fields {
		if strings.Contains(field.Key, "debug") || strings.Contains(field.Key, "performance") ||
			strings.Contains(field.Key, "cache") || strings.Contains(field.Key, "experimental") {
			advancedCount++
		}
	}
	m.tabs[3].Count = advancedCount
	
	// Search results count
	if m.searchMode && m.searchInput.Value() != "" {
		m.tabs[4].Count = len(m.filterSearch(m.fields, m.changed))
	} else {
		m.tabs[4].Count = 0
	}
}

func (m *EnhancedConfigModel) refreshItems() {
	// Defensive checks
	if m == nil || m.tabs == nil {
		return
	}
	
	if m.activeTab >= len(m.tabs) || m.activeTab < 0 {
		m.activeTab = 0
	}
	
	if len(m.tabs) > m.activeTab && m.tabs[m.activeTab].Filter != nil {
		m.items = m.tabs[m.activeTab].Filter(m.fields, m.changed)
	} else {
		m.items = []EnhancedConfigItem{}
	}
	
	// Reset cursor if out of bounds
	if m.cursor >= len(m.items) {
		m.cursor = max(0, len(m.items)-1)
	}
	
	// Reset scroll if needed
	m.ensureVisible()
	
	// Update tooltips
	if m.cursor >= 0 && m.cursor < len(m.items) && len(m.items) > 0 {
		m.tooltipText = m.items[m.cursor].Tooltip
	} else {
		m.tooltipText = ""
	}
}

// Init initializes the model
func (m *EnhancedConfigModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
	)
}

// Update handles messages
// SetSize sets the component dimensions
func (m *EnhancedConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ensureVisible()
	if m.pager != nil {
		m.pager.SetSize(width, height)
	}
}

func (m *EnhancedConfigModel) Update(msg tea.Msg) (model *EnhancedConfigModel, cmd tea.Cmd) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Return safe state on panic
			model = m
			cmd = nil
		}
	}()
	// Handle source viewing mode
	if m.showingSource && m.pager != nil {
		// Only allow toggling back to editor with 'v' or 'q'
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if key.Matches(keyMsg, m.keys.ViewSource) || key.Matches(keyMsg, m.keys.Quit) {
				m.showingSource = false
				return m, nil
			}
		}
		// Update pager for other messages
		updatedPager, pagerCmd := m.pager.Update(msg)
		m.pager = updatedPager
		return m, pagerCmd
	}
	
	// Handle search mode
	if m.searchMode {
		return m.handleSearchMode(msg)
	}
	
	// Handle editing mode
	if m.editing {
		return m.handleEditMode(msg)
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible() // Ensure cursor remains visible after resize
		if m.pager != nil {
			m.pager.SetSize(msg.Width, msg.Height)
		}
		
	case tea.MouseMsg:
		return m.handleMouseEvent(msg)
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	
	return m, nil
}

func (m *EnhancedConfigModel) handleMouseEvent(msg tea.MouseMsg) (model *EnhancedConfigModel, cmd tea.Cmd) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			model = m
			cmd = nil
		}
	}()
	switch msg.Type {
	case tea.MouseWheelUp:
		if m.scrollOffset > 0 {
			m.scrollOffset--
			// Move cursor if it goes out of view
			viewHeight := m.height - 15
			if m.cursor >= m.scrollOffset+viewHeight {
				m.cursor = m.scrollOffset + viewHeight - 1
			}
		}
		
	case tea.MouseWheelDown:
		viewHeight := m.height - 15
		maxScroll := max(0, len(m.items)-viewHeight)
		if m.scrollOffset < maxScroll {
			m.scrollOffset++
			// Move cursor if it goes out of view
			if m.cursor < m.scrollOffset {
				m.cursor = m.scrollOffset
			}
		}
		
	case tea.MouseLeft:
		// Calculate which item was clicked based on mouse position
		viewHeight := m.height - 15
		headerHeight := 8 // Approximate height of header, tabs, etc.
		
		if msg.Y >= headerHeight && msg.Y < headerHeight+viewHeight {
			clickedIndex := m.scrollOffset + (msg.Y - headerHeight)
			if clickedIndex >= 0 && clickedIndex < len(m.items) {
				m.cursor = clickedIndex
				m.ensureVisible()
			}
		}
	}
	
	return m, nil
}

func (m *EnhancedConfigModel) handleKeyPress(msg tea.KeyMsg) (model *EnhancedConfigModel, cmd tea.Cmd) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			model = m
			cmd = nil
		}
	}()
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
		
	case key.Matches(msg, m.keys.Search):
		m.searchMode = true
		m.searchInput.Focus()
		m.activeTab = 4 // Switch to search tab
		return m, textinput.Blink
		
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
		if m.cursor > 0 && len(m.items) > 0 {
			m.cursor = max(0, m.cursor-1)
			m.ensureVisible()
		}
		
	case key.Matches(msg, m.keys.Down):
		if len(m.items) > 0 && m.cursor < len(m.items)-1 {
			m.cursor = min(len(m.items)-1, m.cursor+1)
			m.ensureVisible()
		}
		
	case key.Matches(msg, m.keys.PageUp):
		if len(m.items) > 0 {
			pageSize := max(1, m.height - 15)
			m.cursor = max(0, m.cursor-pageSize)
			m.ensureVisible()
		}
		
	case key.Matches(msg, m.keys.PageDown):
		if len(m.items) > 0 {
			pageSize := max(1, m.height - 15)
			m.cursor = min(len(m.items)-1, m.cursor+pageSize)
			m.ensureVisible()
		}
		
	case key.Matches(msg, m.keys.Edit):
		if m.cursor < len(m.items) {
			m.startEditing()
			return m, textinput.Blink
		}
		
	case key.Matches(msg, m.keys.QuickEdit):
		if m.cursor < len(m.items) {
			m.quickEdit()
		}
		
	case key.Matches(msg, m.keys.MultiSelect):
		m.multiSelect = !m.multiSelect
		
	case key.Matches(msg, m.keys.SelectAll):
		if m.multiSelect {
			for _, item := range m.items {
				m.selected[item.Field.Key] = true
			}
		}
		
	case key.Matches(msg, m.keys.Reset):
		m.resetSelected()
		
	case key.Matches(msg, m.keys.Undo):
		m.undo()
		
	case key.Matches(msg, m.keys.Redo):
		m.redo()
		
	case key.Matches(msg, m.keys.Save):
		return m.save()
		
	case key.Matches(msg, m.keys.Help):
		m.showTooltip = !m.showTooltip
		
	case key.Matches(msg, m.keys.ViewSource):
		m.showingSource = !m.showingSource
		if m.showingSource && m.pager != nil {
			m.pager.SetSize(m.width, m.height)
		}
		return m, nil
		
	// Number shortcuts for tabs
	case msg.String() >= "1" && msg.String() <= "5":
		tabIndex := int(msg.String()[0] - '1')
		if tabIndex < len(m.tabs) {
			m.activeTab = tabIndex
			m.cursor = 0
			m.scrollOffset = 0
			m.refreshItems()
		}
	}
	
	return m, nil
}

func (m *EnhancedConfigModel) handleSearchMode(msg tea.Msg) (*EnhancedConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.searchMode = false
			m.searchInput.Blur()
			m.refreshItems()
			return m, nil
			
		case "esc":
			m.searchMode = false
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			m.activeTab = 0 // Back to All tab
			m.refreshItems()
			return m, nil
		}
	}
	
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.updateTabCounts()
	m.refreshItems()
	return m, cmd
}

func (m *EnhancedConfigModel) handleEditMode(msg tea.Msg) (*EnhancedConfigModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		
		// Live preview while typing
		if m.cursor < len(m.items) {
			item := m.items[m.cursor]
			m.previewValue = m.editInput.Value()
			m.validation = m.validateField(item.Field, m.previewValue)
			m.previewPanel = m.generatePreview(item.Field, m.previewValue)
		}
		
		return m, cmd
	}
	
	switch keyMsg.String() {
	case "enter":
		if m.validation.IsValid {
			m.applyEdit()
		}
		return m, nil
		
	case "esc":
		m.editing = false
		m.editInput.Blur()
		m.previewValue = ""
		m.showPreview = false
		return m, nil
		
	case "tab":
		// Tab completion for select fields
		if m.cursor < len(m.items) {
			item := m.items[m.cursor]
			if len(item.Field.Options) > 0 {
				m.autoComplete()
			}
		}
	}
	
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	
	// Update preview
	if m.cursor < len(m.items) {
		item := m.items[m.cursor]
		m.previewValue = m.editInput.Value()
		m.validation = m.validateField(item.Field, m.previewValue)
		m.previewPanel = m.generatePreview(item.Field, m.previewValue)
	}
	
	return m, cmd
}

func (m *EnhancedConfigModel) startEditing() {
	if m.cursor >= len(m.items) {
		return
	}
	
	item := m.items[m.cursor]
	
	// Smart editing based on field type
	switch item.Field.Type {
	case FieldTypeBool:
		// Instant toggle for booleans
		m.quickEdit()
		
	case FieldTypeSelect:
		if len(item.Field.Options) > 3 {
			// Show dropdown-like interface for many options
			m.editing = true
			m.editingIndex = m.cursor
			m.editInput.SetValue(item.Value)
			m.editInput.Focus()
			m.showPreview = true
		} else {
			// Quick cycle for few options
			m.quickEdit()
		}
		
	default:
		// Text input with validation and preview
		m.editing = true
		m.editingIndex = m.cursor
		m.editInput.SetValue(item.Value)
		m.editInput.Focus()
		m.editInput.CursorEnd()
		m.showPreview = true
	}
}

func (m *EnhancedConfigModel) quickEdit() {
	if m.cursor >= len(m.items) {
		return
	}
	
	item := m.items[m.cursor]
	
	switch item.Field.Type {
	case FieldTypeBool:
		// Toggle boolean
		current := m.values[item.Field.Key] == "true"
		newValue := fmt.Sprintf("%v", !current)
		m.applyChange(item.Field.Key, newValue)
		
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
			m.applyChange(item.Field.Key, newValue)
		}
	}
}

func (m *EnhancedConfigModel) applyEdit() {
	if m.cursor >= len(m.items) || !m.validation.IsValid {
		return
	}
	
	item := m.items[m.cursor]
	newValue := m.editInput.Value()
	
	m.applyChange(item.Field.Key, newValue)
	
	m.editing = false
	m.editInput.Blur()
	m.showPreview = false
}

func (m *EnhancedConfigModel) applyChange(key, value string) {
	// Record history for undo/redo
	change := ConfigChange{
		Key:       key,
		OldValue:  m.values[key],
		NewValue:  value,
		Timestamp: time.Now(),
	}
	m.history = append(m.history, change)
	
	// Apply change
	m.values[key] = value
	
	if value != m.originalValues[key] {
		m.changed[key] = true
	} else {
		delete(m.changed, key)
	}
	
	m.updateTabCounts()
	m.refreshItems()
}

func (m *EnhancedConfigModel) validateField(field ConfigField, value string) ValidationResult {
	// Sanitize input
	value = strings.TrimSpace(value)
	
	// Check for max length
	if len(value) > 1000 {
		return ValidationResult{
			IsValid: false,
			Message: "Value too long (max 1000 characters)",
			Level:   "error",
		}
	}
	
	// Basic validation based on field type
	switch field.Type {
	case FieldTypeInt:
		// Validate integer
		if value == "" {
			return ValidationResult{IsValid: true}
		}
		// Check if valid integer
		for _, ch := range value {
			if ch != '-' && ch != '+' && (ch < '0' || ch > '9') {
				return ValidationResult{
					IsValid: false,
					Message: "Must be a valid integer",
					Level:   "error",
				}
			}
		}
		
	case FieldTypeFloat:
		// Validate float
		if value == "" {
			return ValidationResult{IsValid: true}
		}
		// Check if valid float
		dotCount := 0
		for _, ch := range value {
			if ch == '.' {
				dotCount++
				if dotCount > 1 {
					return ValidationResult{
						IsValid: false,
						Message: "Must be a valid decimal number",
						Level:   "error",
					}
				}
			} else if ch != '-' && ch != '+' && (ch < '0' || ch > '9') {
				return ValidationResult{
					IsValid: false,
					Message: "Must be a valid decimal number",
					Level:   "error",
				}
			}
		}
		
	case FieldTypeBool:
		// Validate boolean
		lower := strings.ToLower(value)
		if value != "" && lower != "true" && lower != "false" && lower != "yes" && lower != "no" && lower != "1" && lower != "0" {
			return ValidationResult{
				IsValid: false,
				Message: "Must be true/false, yes/no, or 1/0",
				Level:   "error",
			}
		}
		
	case FieldTypeSelect:
		// Validate against options
		if len(field.Options) > 0 {
			valid := false
			for _, opt := range field.Options {
				if opt == value {
					valid = true
					break
				}
			}
			if !valid && value != "" {
				return ValidationResult{
					IsValid: false,
					Message: fmt.Sprintf("Must be one of: %s", strings.Join(field.Options, ", ")),
					Level:   "error",
				}
			}
		}
		
	default:
		// String validation - check for dangerous characters
		if strings.ContainsAny(value, "\x00\r\n") {
			return ValidationResult{
				IsValid: false,
				Message: "Value contains invalid characters",
				Level:   "error",
			}
		}
	}
	
	return ValidationResult{IsValid: true}
}

func (m *EnhancedConfigModel) autoComplete() {
	if m.cursor >= len(m.items) {
		return
	}
	
	item := m.items[m.cursor]
	if len(item.Field.Options) == 0 {
		return
	}
	
	current := m.editInput.Value()
	
	// Find matching option
	for _, opt := range item.Field.Options {
		if strings.HasPrefix(strings.ToLower(opt), strings.ToLower(current)) {
			m.editInput.SetValue(opt)
			return
		}
	}
}

func (m *EnhancedConfigModel) ensureVisible() {
	// Defensive checks
	if m == nil || len(m.items) == 0 {
		return
	}
	
	viewHeight := max(1, m.height - 15)
	
	// Ensure cursor is within bounds
	m.cursor = safeBounds(m.cursor, len(m.items))
	
	// If cursor is above the visible area, scroll up
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	
	// If cursor is below the visible area, scroll down
	if m.cursor >= m.scrollOffset+viewHeight {
		m.scrollOffset = max(0, m.cursor - viewHeight + 1)
	}
	
	// Ensure scroll offset is within bounds
	maxScroll := max(0, len(m.items)-viewHeight)
	m.scrollOffset = max(0, min(m.scrollOffset, maxScroll))
}

func (m *EnhancedConfigModel) resetSelected() {
	if m.multiSelect {
		// Reset selected items
		for key := range m.selected {
			m.values[key] = m.originalValues[key]
			delete(m.changed, key)
		}
		m.selected = make(map[string]bool)
	} else if m.cursor < len(m.items) {
		// Reset current item
		item := m.items[m.cursor]
		m.values[item.Field.Key] = m.originalValues[item.Field.Key]
		delete(m.changed, item.Field.Key)
	}
	
	m.updateTabCounts()
	m.refreshItems()
}

func (m *EnhancedConfigModel) undo() {
	if len(m.history) == 0 {
		return
	}
	
	// Get last change
	change := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	
	// Revert change
	m.values[change.Key] = change.OldValue
	
	if change.OldValue != m.originalValues[change.Key] {
		m.changed[change.Key] = true
	} else {
		delete(m.changed, change.Key)
	}
	
	m.updateTabCounts()
	m.refreshItems()
}

func (m *EnhancedConfigModel) redo() {
	// Implement redo functionality
	// Would need a separate redo stack
}

func (m *EnhancedConfigModel) save() (*EnhancedConfigModel, tea.Cmd) {
	if len(m.changed) == 0 {
		return m, nil
	}
	
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

// View renders the enhanced interface
func (m *EnhancedConfigModel) View() string {
	// Panic recovery for rendering
	defer func() {
		if r := recover(); r != nil {
			// Return error message on render failure
			return
		}
	}()
	
	// Defensive checks
	if m == nil {
		return "Error: Configuration model not initialized"
	}
	
	// Early return if dimensions not set
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}
	
	// Show pager when viewing source
	if m.showingSource && m.pager != nil {
		return m.pager.View()
	}
	
	if m.editing && m.showPreview {
		return m.renderEditWithPreview()
	}
	
	if m.editing {
		return m.renderEditView()
	}
	
	if m.searchMode {
		return m.renderSearchView()
	}
	
	return m.renderMainView()
}

func (m *EnhancedConfigModel) renderMainView() string {
	var sections []string
	
	// Ensure we have valid dimensions
	if m.width <= 0 || m.height <= 0 {
		return "Loading..."
	}
	
	// Header with app name
	header := enhancedBaseStyle.
		Foreground(enhancedAccentColor).
		Bold(true).
		Render(fmt.Sprintf("⚙️  %s Configuration", m.appName))
	sections = append(sections, header)
	
	// Tabs
	sections = append(sections, m.renderTabs())
	
	// Content area with items
	content := m.renderItems()
	
	// Calculate safe dimensions
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}
	contentHeight := m.height - 12
	if contentHeight < 5 {
		contentHeight = 5
	}
	
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(enhancedAccentColor).
		Width(contentWidth).
		Height(contentHeight).
		Padding(1).
		Render(content)
	sections = append(sections, contentBox)
	
	// Status bar
	sections = append(sections, m.renderStatusBar())
	
	// Tooltip (if enabled)
	if m.showTooltip && m.tooltipText != "" {
		sections = append(sections, m.renderTooltip())
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *EnhancedConfigModel) renderTabs() string {
	var tabs []string
	
	for i, tab := range m.tabs {
		label := fmt.Sprintf("%s %s", tab.Icon, tab.Name)
		if tab.Count > 0 {
			label = fmt.Sprintf("%s (%d)", label, tab.Count)
		}
		if tab.Shortcut != "" {
			label = fmt.Sprintf("[%s] %s", tab.Shortcut, label)
		}
		
		style := enhancedTabStyle
		if i == m.activeTab {
			style = enhancedActiveTabStyle
		}
		
		tabs = append(tabs, style.Render(label))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *EnhancedConfigModel) renderItems() string {
	if len(m.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Render("No items to display")
	}
	
	// Calculate visible range for performance
	visibleHeight := m.height - 14 // Account for header, tabs, borders, status
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	
	var lines []string
	viewHeight := max(1, m.height - 15) // Ensure minimum height
	
	// Ensure scrollOffset is valid
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	
	endIndex := min(m.scrollOffset+viewHeight, len(m.items))
	for i := m.scrollOffset; i < endIndex; i++ {
		if i >= 0 && i < len(m.items) {
			item := m.items[i]
			lines = append(lines, m.renderItem(item, i == m.cursor))
		}
	}
	
	// Scrollbar indicator
	if len(m.items) > viewHeight {
		scrollPercent := float64(m.scrollOffset) / float64(len(m.items)-viewHeight)
		scrollBar := m.renderScrollBar(viewHeight, scrollPercent)
		
		// Combine content with scrollbar
		contentLines := strings.Split(strings.Join(lines, "\n"), "\n")
		scrollLines := strings.Split(scrollBar, "\n")
		
		var combined []string
		for i := 0; i < len(contentLines) && i < len(scrollLines); i++ {
			combined = append(combined, contentLines[i]+" "+scrollLines[i])
		}
		
		return strings.Join(combined, "\n")
	}
	
	return strings.Join(lines, "\n")
}

func (m *EnhancedConfigModel) renderItem(item EnhancedConfigItem, selected bool) string {
	// Status indicators
	var status string
	if item.IsChanged {
		status = "M"
	} else if item.Field.IsSet {
		status = "*"
	} else {
		status = "o"
	}
	
	// Multi-select indicator
	if m.multiSelect && item.IsSelected {
		status = "[x] " + status
	}
	
	// Field name with proper formatting
	fieldName := m.formatFieldName(item.Field.Key)
	
	// Value display
	value := item.Value
	if value == "" {
		value = "(not set)"
	}
	
	// Type indicator
	var typeIcon string
	switch item.Field.Type {
	case FieldTypeBool:
		typeIcon = "[B]"
	case FieldTypeSelect:
		typeIcon = "[S]"
	case FieldTypeInt, FieldTypeFloat:
		typeIcon = "[#]"
	default:
		typeIcon = "[T]"
	}
	
	// Build line
	line := fmt.Sprintf("%s %s %-30s %s = %s", 
		status, typeIcon, fieldName, 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("->"),
		value)
	
	// Apply style
	style := enhancedItemStyle
	if selected {
		style = enhancedSelectedStyle
	} else if item.IsChanged {
		style = enhancedChangedStyle
	} else if !item.IsValid {
		style = enhancedErrorStyle
	}
	
	return style.Render(line)
}

func (m *EnhancedConfigModel) renderScrollBar(height int, position float64) string {
	var lines []string
	
	barPosition := int(float64(height) * position)
	
	for i := 0; i < height; i++ {
		if i == barPosition {
			lines = append(lines, "#")
		} else {
			lines = append(lines, "|")
		}
	}
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3C3C3C")).
		Render(strings.Join(lines, "\n"))
}

func (m *EnhancedConfigModel) renderStatusBar() string {
	var parts []string
	
	// Mode indicator
	if m.multiSelect {
		parts = append(parts, "[Multi-Select]")
	}
	
	// Changed count
	if len(m.changed) > 0 {
		parts = append(parts, 
			lipgloss.NewStyle().Foreground(enhancedSuccessColor).Render(fmt.Sprintf("[%d modified]", len(m.changed))))
	}
	
	// History indicator
	if len(m.history) > 0 {
		parts = append(parts, fmt.Sprintf("[H:%d]", len(m.history)))
	}
	
	// Help text
	help := "? help • / search • enter edit • space quick • ctrl+s save"
	parts = append(parts, 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(help))
	
	return lipgloss.JoinHorizontal(lipgloss.Top, 
		strings.Join(parts[:len(parts)-1], " • ")+"  ", parts[len(parts)-1])
}

func (m *EnhancedConfigModel) renderTooltip() string {
	if m.tooltipText == "" {
		return ""
	}
	
	return enhancedTooltipStyle.
		Width(m.width - 4).
		Render("[i] " + m.tooltipText)
}

func (m *EnhancedConfigModel) renderSearchView() string {
	var sections []string
	
	// Search header
	searchHeader := enhancedSearchStyle.
		Width(m.width - 4).
		Render(fmt.Sprintf("[/] Search: %s", m.searchInput.View()))
	sections = append(sections, searchHeader)
	
	// Results count
	if m.searchInput.Value() != "" {
		count := len(m.filterSearch(m.fields, m.changed))
		resultText := fmt.Sprintf("Found %d results", count)
		sections = append(sections, 
			lipgloss.NewStyle().
				Foreground(enhancedAccentColor).
				Render(resultText))
	}
	
	// Search results
	sections = append(sections, m.renderItems())
	
	// Help
	sections = append(sections,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("enter: apply search • esc: cancel"))
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *EnhancedConfigModel) renderEditView() string {
	if m.cursor >= len(m.items) {
		return ""
	}
	
	item := m.items[m.cursor]
	
	var sections []string
	
	// Edit header
	header := fmt.Sprintf("[E] Editing: %s", m.formatFieldName(item.Field.Key))
	sections = append(sections, 
		lipgloss.NewStyle().
			Foreground(enhancedAccentColor).
			Bold(true).
			Render(header))
	
	// Description
	if item.Field.Description != "" {
		sections = append(sections,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Render(item.Field.Description))
	}
	
	// Current value
	sections = append(sections,
		fmt.Sprintf("Current: %s", item.Value))
	
	// Edit input
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(enhancedAccentColor).
		Padding(0, 1).
		Width(60).
		Render(m.editInput.View())
	sections = append(sections, inputBox)
	
	// Validation feedback
	if !m.validation.IsValid {
		var validationStyle lipgloss.Style
		switch m.validation.Level {
		case "error":
			validationStyle = lipgloss.NewStyle().Foreground(enhancedErrorColor)
		case "warning":
			validationStyle = lipgloss.NewStyle().Foreground(enhancedWarningColor)
		default:
			validationStyle = lipgloss.NewStyle().Foreground(enhancedAccentColor)
		}
		sections = append(sections,
			validationStyle.Render("[!] "+m.validation.Message))
	}
	
	// Options (for select fields)
	if len(item.Field.Options) > 0 {
		optionsList := strings.Join(item.Field.Options, ", ")
		sections = append(sections,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Render("Options: "+optionsList))
	}
	
	// Help
	sections = append(sections,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("enter: save • esc: cancel • tab: autocomplete"))
	
	editBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(enhancedAccentColor).
		Padding(1).
		Width(m.width - 4).
		Render(strings.Join(sections, "\n\n"))
	
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		editBox,
	)
}

func (m *EnhancedConfigModel) renderEditWithPreview() string {
	if m.cursor >= len(m.items) {
		return ""
	}
	
	// Split view: edit on left, preview on right
	editView := m.renderEditView()
	
	// Preview panel
	previewContent := []string{
		lipgloss.NewStyle().
			Foreground(enhancedAccentColor).
			Bold(true).
			Render("[Preview]"),
		"",
	}
	
	if m.previewPanel != "" {
		previewContent = append(previewContent, m.previewPanel)
	}
	
	if m.previewValue != "" {
		previewContent = append(previewContent,
			fmt.Sprintf("New value: %s", m.previewValue))
	}
	
	previewBox := enhancedPreviewStyle.
		Width(m.width/2 - 2).
		Height(m.height - 4).
		Render(strings.Join(previewContent, "\n"))
	
	// Combine edit and preview
	return lipgloss.JoinHorizontal(lipgloss.Top, editView, previewBox)
}

func (m *EnhancedConfigModel) formatFieldName(key string) string {
	// Convert snake_case to Title Case
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

// Interface methods

func (m *EnhancedConfigModel) Focus() {}
func (m *EnhancedConfigModel) Blur() {}

func (m *EnhancedConfigModel) IsValid() bool {
	return true
}

func (m *EnhancedConfigModel) GetValues() map[string]interface{} {
	values := make(map[string]interface{})
	for key := range m.changed {
		values[key] = m.values[key]
	}
	return values
}

func (m *EnhancedConfigModel) HasUnsavedChanges() bool {
	return len(m.changed) > 0
}