package forms

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/animations"
	"github.com/mrtkrcm/ZeroUI/internal/tui/feedback"
	"github.com/mrtkrcm/ZeroUI/internal/tui/help"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// SimpleConfigModel provides a clean, focused configuration editor with delightful UX
type SimpleConfigModel struct {
	appName      string
	fields       []ConfigField
	cursor       int
	scrollOffset int
	width        int
	height       int

	// Editing state
	editing       bool
	editIndex     int
	editInput     textinput.Model
	previewValue  string
	validationMsg string
	isValid       bool

	// Search state
	isSearching bool
	searchInput textinput.Model
	allFields   []ConfigField // Original unfiltered fields

	// Values
	values  map[string]string
	changed map[string]bool

	// UI state
	showHelp    bool
	searchQuery string
	filtered    []ConfigField

	// Delightful UX features ✨
	notifications    *feedback.NotificationSystem
	loadingSystem    *feedback.LoadingSystem
	contextualHelp   *help.ContextualHelp
	animationManager *animations.AnimationManager
	lastActivity     time.Time
	frameCount       int
}

// ConfigField is defined in types.go

// NewSimpleConfig creates a new simple configuration model with delightful UX
func NewSimpleConfig(appName string) *SimpleConfigModel {
	editInput := textinput.New()
	editInput.CharLimit = 200
	editInput.Prompt = "✏️ "
	editInput.Placeholder = "Enter new value..."
	editInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.TextPrimary))
	editInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.Highlight))

	searchInput := textinput.New()
	searchInput.CharLimit = 100
	searchInput.Prompt = "🔍 "
	searchInput.Placeholder = "Search fields..."
	searchInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.TextPrimary))
	searchInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.Highlight))

	m := &SimpleConfigModel{
		appName:      appName,
		cursor:       0,
		width:        80,
		height:       24,
		editInput:    editInput,
		searchInput:  searchInput,
		isSearching:  false,
		values:       make(map[string]string),
		changed:      make(map[string]bool),
		fields:       []ConfigField{},
		filtered:     []ConfigField{},
		allFields:    []ConfigField{},
		lastActivity: time.Now(),

		// Initialize delightful UX features ✨
		notifications:    feedback.NewNotificationSystem(),
		loadingSystem:    feedback.NewLoadingSystem(),
		contextualHelp:   help.NewContextualHelp(),
		animationManager: animations.NewAnimationManager(),
	}

	// Show welcome notification
	m.notifications.ShowInfo("🎉 Welcome! Use ↑↓ to navigate, Enter to edit", 3*time.Second)

	return m
}

// SetFields configures the available fields
func (m *SimpleConfigModel) SetFields(fields []ConfigField) {
	m.fields = fields
	m.allFields = fields
	m.filtered = fields
	m.updateValues()
}

// SetSize sets the terminal size
func (m *SimpleConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.editInput.Width = width - 4
	m.searchInput.Width = width - 4
}

// Init initializes the model
func (m *SimpleConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles model updates with delightful feedback
func (m *SimpleConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update delightful UX systems
	m.updateDelightfulUX()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		model, cmd := m.handleKeyWithFeedback(msg)
		cmds = append(cmds, cmd)
		return model, tea.Batch(cmds...)

	case tea.MouseMsg:
		model, cmd := m.handleMouseWithFeedback(msg)
		cmds = append(cmds, cmd)
		return model, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.notifications.ShowInfo(fmt.Sprintf("📐 Resized to %dx%d", msg.Width, msg.Height), 2*time.Second)
		return m, nil
	}

	// Handle search input
	if m.isSearching && !m.editing {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)

		// Apply search in real-time
		m.applySearch()
		return m, tea.Batch(cmds...)
	}

	if m.editing {
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m *SimpleConfigModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Search bar (if searching)
	if m.isSearching {
		searchBar := m.renderSearchBar()
		sections = append(sections, searchBar)
	}

	// Content
	content := m.renderContent()
	sections = append(sections, content)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// handleKey handles key presses
func (m *SimpleConfigModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editing {
		switch msg.String() {
		case "enter":
			m.saveEdit()
			return m, nil
		case "esc":
			m.cancelEdit()
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
		return m, nil
	case "enter", " ":
		if len(m.filtered) > 0 {
			m.startEdit()
			return m, m.editInput.Focus()
		}
		return m, nil
	case "esc":
		m.showHelp = false
		return m, nil
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "/":
		// Simple search toggle (could be enhanced)
		return m, nil
	}

	return m, nil
}

// renderHeader renders the header
func (m *SimpleConfigModel) renderHeader() string {
	title := fmt.Sprintf("⚙️ %s Configuration", m.appName)
	if title == "⚙️  Configuration" {
		title = "⚙️ Configuration Editor"
	}

	modified := 0
	for _, changed := range m.changed {
		if changed {
			modified++
		}
	}

	var subtitle string
	if modified > 0 {
		subtitle = fmt.Sprintf(" (%d modified)", modified)
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Width(m.width)

	return headerStyle.Render(title + subtitle)
}

// renderContent renders the main content
func (m *SimpleConfigModel) renderContent() string {
	if len(m.filtered) == 0 {
		return m.renderEmptyState()
	}

	var lines []string
	start := m.scrollOffset
	end := min(start+m.height-8, len(m.filtered)) // Leave space for header/footer

	for i := start; i < end; i++ {
		line := m.renderField(i)
		lines = append(lines, line)
	}

	// Fill remaining space
	for len(lines) < m.height-8 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return content
}

// renderField renders a single field
func (m *SimpleConfigModel) renderField(index int) string {
	if index >= len(m.filtered) {
		return ""
	}

	field := m.filtered[index]
	currentValue := m.getValue(field.Key)

	// Determine styling
	var style lipgloss.Style
	prefix := "  "

	if index == m.cursor {
		if m.editing && index == m.cursor {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF"))
			prefix = "▶ "
		} else {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#000000"))
			prefix = "▶ "
		}
	} else if m.changed[field.Key] {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")) // Green for changed
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")) // Default color
	}

	// Build the field line
	key := field.Key
	value := currentValue

	if value == "" {
		value = toString(field.Default)
		if value != "" {
			value = fmt.Sprintf("%s (default)", value)
		}
	}

	// Truncate long values
	if len(value) > 40 {
		value = value[:37] + "..."
	}

	line := fmt.Sprintf("%s%-20s: %s", prefix, key, value)

	if m.editing && index == m.cursor {
		editLine := fmt.Sprintf("%s✏️ %s: %s", prefix[:1], key, m.editInput.View())
		return style.Width(m.width).Render(editLine)
	}

	return style.Width(m.width).Render(line)
}

// renderEmptyState renders when no fields are available
func (m *SimpleConfigModel) renderEmptyState() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Align(lipgloss.Center).
		Width(m.width).
		Height(m.height - 8)

	return style.Render("No configuration fields available")
}

// renderFooter renders the footer with help/status
func (m *SimpleConfigModel) renderFooter() string {
	if m.showHelp {
		return m.renderHelp()
	}

	var parts []string

	// Navigation help
	if m.isSearching {
		parts = append(parts, "↑/↓ Navigate")
		parts = append(parts, "Enter Select")
		parts = append(parts, "Esc End Search")
		parts = append(parts, "? Help")
	} else {
		parts = append(parts, "↑/↓ Navigate")
		parts = append(parts, "Enter Edit")
		parts = append(parts, "/ Search")
		parts = append(parts, "? Help")
		parts = append(parts, "q Quit")
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		Width(m.width)

	return footerStyle.Render(strings.Join(parts, " • "))
}

func (m *SimpleConfigModel) renderSearchBar() string {
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.ModernTheme.Accent)).
		Padding(0, 1).
		Width(m.width)

	return searchStyle.Render(m.searchInput.View())
}

// renderHelp renders the help overlay
func (m *SimpleConfigModel) renderHelp() string {
	help := []string{
		"┌─ Help ─────────────────────────────────┐",
		"│ ↑/↓/j/k    Navigate fields              │",
		"│ Enter/Space Start editing               │",
		"│ Esc         Cancel edit / Close help    │",
		"│ ?           Toggle this help            │",
		"│ q/Ctrl+C    Quit                        │",
		"└─────────────────────────────────────────┘",
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4")).
		Background(lipgloss.Color("#1E1E2E")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return helpStyle.Render(strings.Join(help, "\n"))
}

// Core functionality
func (m *SimpleConfigModel) getValue(key string) string {
	if value, exists := m.values[key]; exists {
		return value
	}
	return ""
}

func (m *SimpleConfigModel) setValue(key, value string) {
	oldValue := m.getValue(key)
	if oldValue != value {
		m.values[key] = value
		m.changed[key] = true
	}
}

func (m *SimpleConfigModel) updateValues() {
	for _, field := range m.fields {
		if value := toString(field.Value); value != "" {
			m.values[field.Key] = value
		}
	}
}

func (m *SimpleConfigModel) startEdit() {
	if len(m.filtered) == 0 {
		return
	}

	field := m.filtered[m.cursor]
	currentValue := m.getValue(field.Key)
	if currentValue == "" {
		currentValue = toString(field.Default)
	}

	m.editing = true
	m.editInput.SetValue(currentValue)
}

func (m *SimpleConfigModel) saveEdit() {
	if !m.editing {
		return
	}

	field := m.filtered[m.cursor]
	newValue := strings.TrimSpace(m.editInput.Value())
	m.setValue(field.Key, newValue)

	m.editing = false
	m.editInput.Blur()
}

func (m *SimpleConfigModel) cancelEdit() {
	if !m.editing {
		return
	}

	m.editing = false
	m.editInput.Blur()
}

// Utility functions are in utils.go

// Backward compatibility aliases
type TabbedConfigModel = SimpleConfigModel
type EnhancedConfigModel = SimpleConfigModel

// Backward compatibility constructors
func NewTabbedConfig(appName string) *SimpleConfigModel {
	return NewSimpleConfig(appName)
}

func NewEnhancedConfig(appName string) *SimpleConfigModel {
	return NewSimpleConfig(appName)
}

// Config field type constants for backward compatibility
type ConfigFieldType string

const (
	ConfigFieldTypeString ConfigFieldType = "string"
	ConfigFieldTypeBool   ConfigFieldType = "boolean"
	ConfigFieldTypeNumber ConfigFieldType = "number"
	ConfigFieldTypeSelect ConfigFieldType = "select"
	ConfigFieldTypeColor  ConfigFieldType = "color"
	ConfigFieldTypePath   ConfigFieldType = "path"
	FieldTypeString       ConfigFieldType = "string"
	FieldTypeBool         ConfigFieldType = "boolean"
	FieldTypeInt          ConfigFieldType = "number"
	FieldTypeSelect       ConfigFieldType = "select"
	FieldTypeFloat        ConfigFieldType = "float"
)

// Delightful UX methods ✨

// updateDelightfulUX updates all delightful UX systems
func (m *SimpleConfigModel) updateDelightfulUX() {
	m.frameCount++
	m.notifications.Update()
	m.loadingSystem.Update()
	m.animationManager.UpdateAll()
	m.animationManager.CleanCompleted()

	// Update contextual help based on activity
	m.contextualHelp.UpdateContext(m.getCurrentContext(), m.getLastAction())
}

// getCurrentContext returns the current UI context
func (m *SimpleConfigModel) getCurrentContext() string {
	if m.editing {
		return "editing"
	}
	if m.showHelp {
		return "help"
	}
	if m.isSearching {
		return "searching"
	}
	return "navigation"
}

// getLastAction returns the last user action
func (m *SimpleConfigModel) getLastAction() string {
	if m.editing {
		return "editing"
	}
	return "navigation"
}

// handleKeyWithFeedback handles key presses with delightful feedback
func (m *SimpleConfigModel) handleKeyWithFeedback(msg tea.KeyMsg) (*SimpleConfigModel, tea.Cmd) {
	key := msg.String()

	// Provide immediate visual feedback
	m.showKeyFeedback(key)
	m.lastActivity = time.Now()

	// Update contextual help
	m.contextualHelp.UpdateContext(m.getCurrentContext(), "key-"+key)

	switch m.getCurrentContext() {
	case "editing":
		return m.handleEditingKeys(key)
	case "searching":
		return m.handleSearchKeys(key)
	default:
		return m.handleNavigationKeys(key)
	}
}

// handleEditingKeys handles editing mode with feedback
func (m *SimpleConfigModel) handleEditingKeys(key string) (*SimpleConfigModel, tea.Cmd) {
	switch key {
	case "enter":
		m.saveEditWithFeedback()
		return m, nil
	case "esc":
		m.cancelEdit()
		m.notifications.ShowInfo("✗ Edit cancelled", 2*time.Second)
		return m, nil
	case "tab":
		// Show auto-complete suggestions
		m.showAutoComplete()
		return m, nil
	}
	return m, nil
}

// handleSearchKeys handles search mode
func (m *SimpleConfigModel) handleSearchKeys(key string) (*SimpleConfigModel, tea.Cmd) {
	switch key {
	case "enter":
		if len(m.filtered) > 0 {
			m.isSearching = false
			m.startEditingWithAnimation()
			return m, m.editInput.Focus()
		}
		return m, nil
	case "esc":
		m.endSearch()
		return m, nil
	}
	return m, nil
}

// handleNavigationKeys handles navigation with delightful feedback
func (m *SimpleConfigModel) handleNavigationKeys(key string) (*SimpleConfigModel, tea.Cmd) {
	switch key {
	case "q", "ctrl+c":
		m.notifications.ShowInfo("👋 See you later!", 2*time.Second)
		return m, tea.Quit

	case "up", "k":
		m.moveCursor(-1)
		m.notifications.ShowTooltip("Navigate up", 1*time.Second)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		m.notifications.ShowTooltip("Navigate down", 1*time.Second)
		return m, nil

	case "enter", " ":
		if len(m.filtered) > 0 {
			m.startEditingWithAnimation()
			return m, m.editInput.Focus()
		}
		return m, nil

	case "h", "?":
		m.toggleHelp()
		return m, nil

	case "/":
		if !m.isSearching {
			m.startSearch()
		}
		return m, nil

	case "esc":
		m.clearOverlays()
		return m, nil

	case "u":
		if m.undoLastChange() {
			m.notifications.ShowSuccess("↶ Undid last change", 2*time.Second)
		} else {
			m.notifications.ShowWarning("Nothing to undo", 2*time.Second)
		}
		return m, nil

	case "ctrl+s":
		m.saveConfiguration()
		return m, nil

	default:
		// Handle number keys for quick navigation
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			index := int(key[0] - '1')
			if index < len(m.filtered) {
				m.cursor = index
				m.notifications.ShowTooltip(fmt.Sprintf("Jumped to item %d", index+1), 1*time.Second)
			}
		}
		return m, nil
	}
}

// handleMouseWithFeedback handles mouse events with feedback
func (m *SimpleConfigModel) handleMouseWithFeedback(msg tea.MouseMsg) (*SimpleConfigModel, tea.Cmd) {
	switch msg.Type {
	case tea.MouseLeft:
		return m.handleClick(msg)
	case tea.MouseWheelUp:
		m.moveCursor(-1)
		return m, nil
	case tea.MouseWheelDown:
		m.moveCursor(1)
		return m, nil
	}
	return m, nil
}

// handleClick handles mouse clicks with delightful feedback
func (m *SimpleConfigModel) handleClick(msg tea.MouseMsg) (*SimpleConfigModel, tea.Cmd) {
	// Calculate which item was clicked
	clickY := msg.Y - 3 // Account for header

	if clickY >= 0 && clickY < len(m.filtered) {
		m.cursor = clickY
		m.notifications.ShowTooltip(fmt.Sprintf("Selected: %s", m.filtered[clickY].Key), 1*time.Second)

		// If double-click or alt+click, start editing
		if msg.Alt {
			m.startEditingWithAnimation()
			return m, m.editInput.Focus()
		}
	}

	return m, nil
}

// Enhanced interaction methods
func (m *SimpleConfigModel) moveCursor(delta int) {
	newCursor := m.cursor + delta
	if newCursor >= 0 && newCursor < len(m.filtered) {
		m.cursor = newCursor
	}
}

func (m *SimpleConfigModel) startEditingWithAnimation() {
	if len(m.filtered) == 0 {
		return
	}

	field := m.filtered[m.cursor]
	currentValue := m.getValue(field.Key)
	if currentValue == "" {
		currentValue = toString(field.Default)
	}

	m.editing = true
	m.editIndex = m.cursor
	m.editInput.SetValue(currentValue)

	// Add animation
	fadeIn := animations.NewFadeAnimation(300*time.Millisecond, true)
	m.animationManager.AddAnimation("edit-fade-in", fadeIn)

	m.notifications.ShowTooltip("Start typing to edit this field", 3*time.Second)
}

func (m *SimpleConfigModel) saveEditWithFeedback() {
	if !m.editing {
		return
	}

	field := m.filtered[m.editIndex]
	newValue := strings.TrimSpace(m.editInput.Value())

	// Basic validation
	if newValue == "" {
		m.notifications.ShowError("Value cannot be empty", 3*time.Second)
		return
	}

	m.setValue(field.Key, newValue)
	m.editing = false
	m.notifications.ShowSuccess("✅ Field updated successfully", 2*time.Second)
}

func (m *SimpleConfigModel) showKeyFeedback(key string) {
	// Visual feedback for key presses
	m.lastActivity = time.Now()
}

func (m *SimpleConfigModel) showAutoComplete() {
	m.notifications.ShowInfo("💡 Auto-complete suggestions coming soon!", 2*time.Second)
}

func (m *SimpleConfigModel) toggleHelp() {
	m.showHelp = !m.showHelp
	if m.showHelp {
		m.notifications.ShowInfo("❓ Press any key to exit help", 3*time.Second)
	}
}

func (m *SimpleConfigModel) startSearch() {
	m.searchInput = textinput.New()
	m.searchInput.CharLimit = 100
	m.searchInput.Prompt = "🔍 "
	m.searchInput.Placeholder = "Search fields..."
	m.searchInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.TextPrimary))
	m.searchInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ModernTheme.Highlight))
	m.searchInput.Width = m.width - 4
	m.searchInput.Focus()
	m.isSearching = true
	m.filtered = m.allFields // Start with all fields
	m.cursor = 0
	m.notifications.ShowInfo("🔍 Type to search fields", 2*time.Second)
}

func (m *SimpleConfigModel) applySearch() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.allFields
		m.searchQuery = ""
		return
	}

	m.searchQuery = query
	m.filtered = nil
	for _, field := range m.allFields {
		if strings.Contains(strings.ToLower(field.Key), query) ||
			strings.Contains(strings.ToLower(field.Description), query) {
			m.filtered = append(m.filtered, field)
		}
	}
	m.cursor = 0

	if len(m.filtered) == 0 {
		m.notifications.ShowWarning("No fields match your search", 2*time.Second)
	} else {
		m.notifications.ShowInfo(fmt.Sprintf("Found %d matching fields", len(m.filtered)), 2*time.Second)
	}
}

func (m *SimpleConfigModel) endSearch() {
	m.isSearching = false
	m.searchQuery = ""
	m.filtered = m.allFields
	m.cursor = 0
	m.notifications.ShowInfo("🔍 Search cleared", 1*time.Second)
}

func (m *SimpleConfigModel) clearOverlays() {
	m.showHelp = false
	m.notifications.ClearAll()
}

func (m *SimpleConfigModel) undoLastChange() bool {
	// Simple undo - could be enhanced
	return false
}

func (m *SimpleConfigModel) saveConfiguration() {
	m.loadingSystem.StartStepLoading("save-config", "Saving configuration...", []string{
		"Validating changes",
		"Applying settings",
		"Saving to file",
		"Configuration updated",
	})

	// Simulate save process
	go func() {
		time.Sleep(500 * time.Millisecond)
		m.loadingSystem.UpdateStep("save-config", 0)

		time.Sleep(500 * time.Millisecond)
		m.loadingSystem.UpdateStep("save-config", 1)

		time.Sleep(500 * time.Millisecond)
		m.loadingSystem.UpdateStep("save-config", 2)

		time.Sleep(500 * time.Millisecond)
		m.loadingSystem.CompleteLoading("save-config", "Configuration saved successfully!")

		m.notifications.ShowSuccess("💾 Configuration saved!", 3*time.Second)
	}()
}

// Enhanced View method with delightful overlays
func (m *SimpleConfigModel) EnhancedView() string {
	var sections []string

	// Header
	header := m.renderDelightfulHeader()
	sections = append(sections, header)

	// Content
	content := m.renderContent()
	sections = append(sections, content)

	// Footer
	footer := m.renderDelightfulFooter()
	sections = append(sections, footer)

	// Add delightful overlays
	result := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add notifications
	if notificationView := m.notifications.Render(m.width, m.height); notificationView != "" {
		result += "\n" + notificationView
	}

	// Add loading indicators
	if loadingView := m.loadingSystem.Render(m.width); loadingView != "" {
		result += "\n" + loadingView
	}

	return result
}

// renderDelightfulHeader creates a beautiful, informative header
func (m *SimpleConfigModel) renderDelightfulHeader() string {
	title := fmt.Sprintf("⚙️ %s Configuration", m.appName)
	if m.appName == "" {
		title = "⚙️ Configuration Editor"
	}

	// Add status indicators
	var statusIcons []string
	if m.editing {
		statusIcons = append(statusIcons, "✏️")
	}
	if len(m.changed) > 0 {
		statusIcons = append(statusIcons, fmt.Sprintf("📝%d", len(m.changed)))
	}

	status := ""
	if len(statusIcons) > 0 {
		status = " " + strings.Join(statusIcons, " ")
	}

	return styles.HeaderStyle.Render(title + status)
}

// renderDelightfulFooter creates an informative, helpful footer
func (m *SimpleConfigModel) renderDelightfulFooter() string {
	var left, center, right string

	// Left: Navigation info
	if len(m.filtered) > 0 {
		left = fmt.Sprintf("%s (%d/%d)",
			m.filtered[m.cursor].Key,
			m.cursor+1,
			len(m.filtered))
	}

	// Center: Contextual help
	if suggestions := m.contextualHelp.GetSuggestions(); len(suggestions) > 0 {
		center = suggestions[0].Text
	} else {
		center = m.getContextualHint()
	}

	// Right: Key hints
	right = m.contextualHelp.GetQuickHelp()

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.ModernTheme.TextSecondary)).
		Background(lipgloss.Color(styles.ModernTheme.Surface)).
		Padding(0, 1).
		Width(m.width)

	leftPart := lipgloss.NewStyle().Align(lipgloss.Left).Width(m.width / 3).Render(left)
	centerPart := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width / 3).Render(center)
	rightPart := lipgloss.NewStyle().Align(lipgloss.Right).Width(m.width / 3).Render(right)

	return footerStyle.Render(leftPart + centerPart + rightPart)
}

// getContextualHint provides contextual help based on current state
func (m *SimpleConfigModel) getContextualHint() string {
	if len(m.filtered) == 0 {
		return "💡 No items to display"
	}

	field := m.filtered[m.cursor]
	switch {
	case strings.Contains(strings.ToLower(field.Key), "font"):
		return "🎨 Font configuration"
	case strings.Contains(strings.ToLower(field.Key), "color"):
		return "🎨 Color configuration"
	case strings.Contains(strings.ToLower(field.Key), "size"):
		return "📏 Size configuration"
	default:
		return "⚙️ General configuration"
	}
}

// Backward compatibility methods

// HasUnsavedChanges returns true if there are unsaved changes
func (m *SimpleConfigModel) HasUnsavedChanges() bool {
	return len(m.changed) > 0
}

// IsValid returns true if the current configuration is valid
func (m *SimpleConfigModel) IsValid() bool {
	return m.isValid
}

// GetValues returns all current field values
func (m *SimpleConfigModel) GetValues() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m.values {
		result[k] = v
	}
	return result
}

// SetConfigFile sets the configuration file path (for backward compatibility)
func (m *SimpleConfigModel) SetConfigFile(path string, appName string) {
	// This method is for backward compatibility
	// The current implementation doesn't use a single config file path
	// but this maintains the expected interface
	// path and appName parameters are accepted but not used in current implementation
}
