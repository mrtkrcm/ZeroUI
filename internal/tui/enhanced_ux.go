package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Enhanced UX components for ZeroUI

// ToastLevel represents the severity of a toast notification
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast represents a temporary notification
type Toast struct {
	Message   string
	Level     ToastLevel
	Duration  time.Duration
	CreatedAt time.Time
	Active    bool
}

// ToastTimeoutMsg is sent when a toast expires
type ToastTimeoutMsg struct {
	Message string
}

// SearchModel provides fuzzy search functionality
type SearchModel struct {
	input      textinput.Model
	active     bool
	query      string
	results    []string
	selected   int
	placeholder string
}

// NewSearchModel creates a new search component
func NewSearchModel(placeholder string) SearchModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.Width = 30

	return SearchModel{
		input:       ti,
		placeholder: placeholder,
	}
}

// ActivateSearch enables the search mode
func (s *SearchModel) ActivateSearch() tea.Cmd {
	s.active = true
	s.input.Focus()
	return textinput.Blink
}

// DeactivateSearch disables the search mode
func (s *SearchModel) DeactivateSearch() {
	s.active = false
	s.input.Blur()
	s.query = ""
	s.results = nil
	s.selected = 0
}

// Update handles search input
func (s *SearchModel) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.query = s.input.Value()

	return cmd
}

// View renders the search interface
func (s SearchModel) View() string {
	if !s.active {
		return ""
	}

	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Margin(1, 0)

	return searchStyle.Render(fmt.Sprintf("🔍 %s", s.input.View()))
}

// StatusBar provides contextual information and feedback
type StatusBar struct {
	currentApp    string
	currentView   string
	keyHints      []string
	toast         Toast
	showStatus    bool
	debugMode     bool
}

// NewStatusBar creates a new status bar
func NewStatusBar() StatusBar {
	return StatusBar{
		showStatus: true,
	}
}

// SetCurrentApp updates the current app context
func (s *StatusBar) SetCurrentApp(app string) {
	s.currentApp = app
}

// SetCurrentView updates the current view context
func (s *StatusBar) SetCurrentView(view string) {
	s.currentView = view
}

// SetKeyHints updates the contextual key hints
func (s *StatusBar) SetKeyHints(hints []string) {
	s.keyHints = hints
}

// ShowToast displays a toast notification
func (s *StatusBar) ShowToast(message string, level ToastLevel, duration time.Duration) tea.Cmd {
	s.toast = Toast{
		Message:   message,
		Level:     level,
		Duration:  duration,
		CreatedAt: time.Now(),
		Active:    true,
	}

	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return ToastTimeoutMsg{message}
	})
}

// Update handles status bar updates
func (s *StatusBar) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ToastTimeoutMsg:
		if s.toast.Message == msg.Message {
			s.toast.Active = false
		}
	}
	return nil
}

// View renders the status bar
func (s StatusBar) View(width int) string {
	if !s.showStatus {
		return ""
	}

	// Left side: Current context
	leftContent := ""
	if s.currentApp != "" {
		leftContent = fmt.Sprintf("📱 %s", s.currentApp)
	}
	if s.currentView != "" {
		if leftContent != "" {
			leftContent += " • "
		}
		leftContent += fmt.Sprintf("🎯 %s", s.currentView)
	}

	// Right side: Key hints
	rightContent := ""
	if len(s.keyHints) > 0 {
		rightContent = strings.Join(s.keyHints, " • ")
	}

	// Calculate spacing
	contentLength := len(leftContent) + len(rightContent)
	spacing := width - contentLength - 2
	if spacing < 0 {
		spacing = 0
	}

	statusContent := leftContent + strings.Repeat(" ", spacing) + rightContent

	// Toast overlay
	toastContent := ""
	if s.toast.Active {
		toastContent = s.renderToast()
	}

	// Style the status bar
	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		Width(width)

	if toastContent != "" {
		return toastContent + "\n" + statusStyle.Render(statusContent)
	}

	return statusStyle.Render(statusContent)
}

// renderToast renders the toast notification
func (s StatusBar) renderToast() string {
	if !s.toast.Active {
		return ""
	}

	var (
		icon  string
		color lipgloss.Color
	)

	switch s.toast.Level {
	case ToastInfo:
		icon = "ℹ️"
		color = "39"
	case ToastSuccess:
		icon = "✅"
		color = "34"
	case ToastWarning:
		icon = "⚠️"
		color = "220"
	case ToastError:
		icon = "❌"
		color = "196"
	}

	toastStyle := lipgloss.NewStyle().
		Background(color).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		Margin(0, 1).
		Bold(true)

	return toastStyle.Render(fmt.Sprintf("%s %s", icon, s.toast.Message))
}

// ContextualHelp provides context-aware help
type ContextualHelp struct {
	currentContext string
	helpMaps       map[string][]HelpItem
	visible        bool
}

// HelpItem represents a single help entry
type HelpItem struct {
	Keys        string
	Description string
}

// NewContextualHelp creates a new contextual help system
func NewContextualHelp() ContextualHelp {
	return ContextualHelp{
		helpMaps: make(map[string][]HelpItem),
	}
}

// SetContext updates the current help context
func (h *ContextualHelp) SetContext(context string) {
	h.currentContext = context
}

// AddHelpMap adds help items for a context
func (h *ContextualHelp) AddHelpMap(context string, items []HelpItem) {
	h.helpMaps[context] = items
}

// Toggle toggles help visibility
func (h *ContextualHelp) Toggle() {
	h.visible = !h.visible
}

// View renders the contextual help
func (h ContextualHelp) View(width, height int) string {
	if !h.visible {
		return ""
	}

	items, exists := h.helpMaps[h.currentContext]
	if !exists || len(items) == 0 {
		return ""
	}

	var helpLines []string
	helpLines = append(helpLines, fmt.Sprintf("Help - %s", h.currentContext))
	helpLines = append(helpLines, strings.Repeat("─", 40))

	for _, item := range items {
		line := fmt.Sprintf("%-15s %s", item.Keys, item.Description)
		helpLines = append(helpLines, line)
	}

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Margin(1).
		Background(lipgloss.Color("235")).
		MaxWidth(width - 4).
		MaxHeight(height - 4)

	return helpStyle.Render(strings.Join(helpLines, "\n"))
}

// LoadingIndicator provides visual feedback during operations
type LoadingIndicator struct {
	active   bool
	message  string
	spinner  int
	spinners []string
}

// NewLoadingIndicator creates a new loading indicator
func NewLoadingIndicator() LoadingIndicator {
	return LoadingIndicator{
		spinners: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start activates the loading indicator
func (l *LoadingIndicator) Start(message string) tea.Cmd {
	l.active = true
	l.message = message
	l.spinner = 0

	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return LoadingTickMsg{}
	})
}

// Stop deactivates the loading indicator
func (l *LoadingIndicator) Stop() {
	l.active = false
	l.message = ""
}

// LoadingTickMsg advances the spinner animation
type LoadingTickMsg struct{}

// Update handles loading indicator updates
func (l *LoadingIndicator) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case LoadingTickMsg:
		if l.active {
			l.spinner = (l.spinner + 1) % len(l.spinners)
			return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return LoadingTickMsg{}
			})
		}
	}
	return nil
}

// View renders the loading indicator
func (l LoadingIndicator) View() string {
	if !l.active {
		return ""
	}

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	return loadingStyle.Render(fmt.Sprintf("%s %s", l.spinners[l.spinner], l.message))
}

// Enhanced key bindings with better UX
type EnhancedKeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Selection and actions
	Enter   key.Binding
	Space   key.Binding
	Tab     key.Binding
	Escape  key.Binding

	// Application specific
	Toggle    key.Binding
	Cycle     key.Binding
	Reset     key.Binding
	Save      key.Binding
	Reload    key.Binding

	// Interface
	Search   key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
	FullHelp key.Binding

	// Advanced
	Debug       key.Binding
	Screenshot  key.Binding
	Export      key.Binding
}

// NewEnhancedKeyMap creates an enhanced key mapping with better UX
func NewEnhancedKeyMap() EnhancedKeyMap {
	return EnhancedKeyMap{
		// Navigation - follows vim/standard conventions
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("PgUp", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("PgDn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("Home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("End/G", "go to bottom"),
		),

		// Selection and actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "select/confirm"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "quick toggle"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "next field"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "back/cancel"),
		),

		// Application specific
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle value"),
		),
		Cycle: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cycle values"),
		),
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("Ctrl+R", "reset to default"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("Ctrl+S", "save changes"),
		),
		Reload: key.NewBinding(
			key.WithKeys("r", "f5"),
			key.WithHelp("r/F5", "refresh"),
		),

		// Interface
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+f"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		FullHelp: key.NewBinding(
			key.WithKeys("F1"),
			key.WithHelp("F1", "full help"),
		),

		// Advanced
		Debug: key.NewBinding(
			key.WithKeys("ctrl+alt+d"),
			key.WithHelp("Ctrl+Alt+D", "debug mode"),
		),
		Screenshot: key.NewBinding(
			key.WithKeys("ctrl+alt+s"),
			key.WithHelp("Ctrl+Alt+S", "screenshot"),
		),
		Export: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("Ctrl+E", "export config"),
		),
	}
}

// GetContextualHelp returns help items for the current context
func (k EnhancedKeyMap) GetContextualHelp(context string) []HelpItem {
	switch context {
	case "app_grid":
		return []HelpItem{
			{"↑↓←→/hjkl", "Navigate apps"},
			{"Enter", "Open app config"},
			{"Space", "Quick toggle"},
			{"/", "Search apps"},
			{"f", "Filter by status"},
			{"r", "Refresh"},
			{"?", "Toggle help"},
			{"q", "Quit"},
		}
	case "config_edit":
		return []HelpItem{
			{"↑↓/jk", "Navigate fields"},
			{"Enter", "Edit field"},
			{"Space", "Quick toggle"},
			{"Tab", "Next field"},
			{"Ctrl+S", "Save config"},
			{"Ctrl+R", "Reset field"},
			{"Esc", "Back to grid"},
			{"?", "Toggle help"},
		}
	case "search":
		return []HelpItem{
			{"Type", "Search query"},
			{"↑↓/jk", "Navigate results"},
			{"Enter", "Select result"},
			{"Esc", "Cancel search"},
		}
	default:
		return []HelpItem{
			{"?", "Show help"},
			{"q", "Quit"},
		}
	}
}