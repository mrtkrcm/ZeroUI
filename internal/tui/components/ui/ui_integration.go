package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// UIIntegrationManager provides a unified interface to all UI libraries
type UIIntegrationManager struct {
	styles        *styles.Styles
	theme         *styles.Theme
	width         int
	height        int
	isInitialized bool
}

// NewUIIntegrationManager creates a new UI integration manager
func NewUIIntegrationManager() *UIIntegrationManager {
	theme := &styles.DefaultTheme
	appStyles := theme.BuildStyles()

	return &UIIntegrationManager{
		styles: appStyles,
		theme:  theme,
	}
}

// Initialize sets up the UI integration with proper sizing
func (ui *UIIntegrationManager) Initialize(width, height int) {
	ui.width = width
	ui.height = height
	ui.isInitialized = true
}

// IsInitialized returns whether the UI integration manager has been initialized
func (ui *UIIntegrationManager) IsInitialized() bool {
	return ui.isInitialized
}

// GetBubbleTeaList creates a properly styled Bubble Tea list
func (ui *UIIntegrationManager) GetBubbleTeaList(items []list.Item, delegate list.ItemDelegate) list.Model {
	if !ui.isInitialized {
		// Provide sensible defaults
		ui.Initialize(120, 40)
	}

	l := list.New(items, delegate, ui.width-4, ui.height-8)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return l
}

// GetTextInput creates a properly styled text input
func (ui *UIIntegrationManager) GetTextInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.Width = ui.width - 10

	return ti
}

// GetSpinner creates a properly styled spinner
func (ui *UIIntegrationManager) GetSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.styles.Info

	return s
}

// GetViewport creates a properly styled viewport
func (ui *UIIntegrationManager) GetViewport() viewport.Model {
	vp := viewport.New(ui.width-4, ui.height-6)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1)

	return vp
}

// GetHuhForm creates a properly styled Huh form with consistent theming
func (ui *UIIntegrationManager) GetHuhForm(groups ...*huh.Group) *huh.Form {
	theme := huh.ThemeCharm()
	theme.Focused.Title = ui.styles.Title
	theme.Focused.Base = ui.styles.Base
	theme.Focused.NoteTitle = ui.styles.Subtitle
	theme.Focused.Directory = ui.styles.Text
	theme.Focused.FocusedButton = ui.styles.Selected
	theme.Focused.BlurredButton = ui.styles.Muted

	form := huh.NewForm(groups...).
		WithTheme(theme).
		WithWidth(ui.width - 4).
		WithHeight(ui.height - 4).
		WithShowHelp(true).
		WithShowErrors(true).
		WithAccessible(true)

	return form
}

// Note: EnhancedListDelegate is now available from enhanced_app_list.go
// Use NewApplicationListDelegate from that package for list rendering

// Note: ApplicationListItem, ApplicationData, and ApplicationStatus types
// are defined in enhanced_app_list.go to avoid duplication

// Layout helpers for consistent UI

// CreateHeader creates a consistent header layout
func (ui *UIIntegrationManager) CreateHeader(title, subtitle string) string {
	header := ui.styles.Title.Render(title)
	if subtitle != "" {
		sub := ui.styles.Subtitle.Render(subtitle)
		return lipgloss.JoinVertical(lipgloss.Top, header, sub)
	}
	return header
}

// CreateFooter creates a consistent footer layout
func (ui *UIIntegrationManager) CreateFooter(leftText, centerText, rightText string) string {
	footer := ui.styles.Help.Render(leftText)

	if centerText != "" {
		center := ui.styles.Muted.Render(centerText)
		footer = lipgloss.JoinHorizontal(lipgloss.Top, footer, center)
	}

	if rightText != "" {
		right := ui.styles.Info.Render(rightText)
		footer = lipgloss.JoinHorizontal(lipgloss.Top, footer, right)
	}

	return footer
}

// CreateStatusBar creates a status bar with multiple sections
func (ui *UIIntegrationManager) CreateStatusBar(sections ...string) string {
	var styledSections []string

	for i, section := range sections {
		switch i % 3 {
		case 0:
			styledSections = append(styledSections, ui.styles.Text.Render(section))
		case 1:
			styledSections = append(styledSections, ui.styles.Muted.Render(section))
		case 2:
			styledSections = append(styledSections, ui.styles.Info.Render(section))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, styledSections...)
}

// CreateErrorMessage creates a consistently styled error message
func (ui *UIIntegrationManager) CreateErrorMessage(message string) string {
	return ui.styles.Error.Render(fmt.Sprintf("❌ %s", message))
}

// CreateSuccessMessage creates a consistently styled success message
func (ui *UIIntegrationManager) CreateSuccessMessage(message string) string {
	return ui.styles.Success.Render(fmt.Sprintf("✅ %s", message))
}

// CreateInfoMessage creates a consistently styled info message
func (ui *UIIntegrationManager) CreateInfoMessage(message string) string {
	return ui.styles.Info.Render(fmt.Sprintf("ℹ️ %s", message))
}

// CreateWarningMessage creates a consistently styled warning message
func (ui *UIIntegrationManager) CreateWarningMessage(message string) string {
	return ui.styles.Warning.Render(fmt.Sprintf("⚠️ %s", message))
}

// CreateLoadingMessage creates a consistently styled loading message
func (ui *UIIntegrationManager) CreateLoadingMessage(message string) string {
	spinner := ui.GetSpinner()
	return fmt.Sprintf("%s %s", spinner.View(), ui.styles.Muted.Render(message))
}

// Key binding helpers

// GetNavigationKeyBindings returns standard navigation key bindings
func (ui *UIIntegrationManager) GetNavigationKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
	}
}

// GetCommonKeyBindings returns common application key bindings
func (ui *UIIntegrationManager) GetCommonKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
	}
}

// GetFilterKeyBindings returns filtering key bindings
func (ui *UIIntegrationManager) GetFilterKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
	}
}

// Component state helpers

// IsComponentReady checks if a component is ready for use
func (ui *UIIntegrationManager) IsComponentReady(component tea.Model) bool {
	if component == nil {
		return false
	}

	// Check if component implements a readiness interface
	if readyComponent, ok := component.(interface{ IsReady() bool }); ok {
		return readyComponent.IsReady()
	}

	// Default to ready for components that don't implement the interface
	return true
}

// ValidateComponent validates a component's state
func (ui *UIIntegrationManager) ValidateComponent(component tea.Model) []string {
	var errors []string

	if component == nil {
		errors = append(errors, "component is nil")
		return errors
	}

	// Check if component implements validation
	if validator, ok := component.(interface{ Validate() []string }); ok {
		return validator.Validate()
	}

	return errors
}

// Performance monitoring

// MonitorPerformance tracks component performance
func (ui *UIIntegrationManager) MonitorPerformance(component string, start time.Time, operation string) {
	duration := time.Since(start)

	// Log slow operations (over 50ms)
	if duration > 50*time.Millisecond {
		fmt.Printf("[PERF] %s: %s took %v\n", component, operation, duration)
	}
}

// CreatePerformanceTimer creates a performance timer
func (ui *UIIntegrationManager) CreatePerformanceTimer(component string, operation string) func() {
	start := time.Now()
	return func() {
		ui.MonitorPerformance(component, start, operation)
	}
}
