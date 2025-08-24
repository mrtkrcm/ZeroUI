package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/layout"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// ResponsiveHelpModel provides context-aware help that adapts to screen size
type ResponsiveHelpModel struct {
	width  int
	height int

	bindings []key.Binding
	styles   *styles.Styles
}

// NewResponsiveHelp creates a new responsive help component
func NewResponsiveHelp() *ResponsiveHelpModel {
	return &ResponsiveHelpModel{
		height: 1,
		styles: styles.GetStyles(),
	}
}

// Init implements tea.Model
func (m *ResponsiveHelpModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ResponsiveHelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model
func (m *ResponsiveHelpModel) View() string {
	if len(m.bindings) == 0 {
		return ""
	}

	helpText := m.renderHelp()

	// Style the help text
	style := m.styles.Help.
		Width(m.width).
		Align(lipgloss.Left)

	return style.Render(helpText)
}

// renderHelp creates responsive help text based on available width
func (m *ResponsiveHelpModel) renderHelp() string {
	if m.width == 0 {
		return ""
	}

	// Create help items from bindings
	var helpItems []string
	for _, binding := range m.bindings {
		if binding.Help().Key != "" && binding.Help().Desc != "" {
			helpItems = append(helpItems, binding.Help().Key+" "+binding.Help().Desc)
		}
	}

	if len(helpItems) == 0 {
		return ""
	}

	// Choose separator based on available width
	separator := " • "
	if m.width < 60 {
		separator = " | " // Shorter separator for narrow screens
	}

	// Try to fit all help items
	fullHelp := strings.Join(helpItems, separator)
	if lipgloss.Width(fullHelp) <= m.width {
		return fullHelp
	}

	// If too long, use progressive truncation
	return m.truncateHelp(helpItems, separator)
}

// truncateHelp progressively removes help items to fit width
func (m *ResponsiveHelpModel) truncateHelp(items []string, separator string) string {
	if len(items) == 0 {
		return ""
	}

	// Prioritize the most important help items
	priority := m.prioritizeHelpItems(items)

	// Build help text by adding items until we run out of space
	var selectedItems []string
	var currentWidth int

	for _, item := range priority {
		testWidth := currentWidth
		if len(selectedItems) > 0 {
			testWidth += lipgloss.Width(separator)
		}
		testWidth += lipgloss.Width(item)

		if testWidth <= m.width {
			selectedItems = append(selectedItems, item)
			currentWidth = testWidth
		} else {
			break
		}
	}

	// If we can't fit anything, show just the most important
	if len(selectedItems) == 0 && len(priority) > 0 {
		item := priority[0]
		if lipgloss.Width(item) <= m.width {
			selectedItems = append(selectedItems, item)
		}
	}

	// Add ellipsis if we truncated items
	result := strings.Join(selectedItems, separator)
	if len(selectedItems) < len(items) && len(result) < m.width-3 {
		result += " …"
	}

	return result
}

// prioritizeHelpItems sorts help items by importance
func (m *ResponsiveHelpModel) prioritizeHelpItems(items []string) []string {
	// Create a priority map for common actions
	priorities := map[string]int{
		"q quit":               1, // Most important - always show quit
		"? help":               2, // Help is very important
		"? toggle help":        2,
		"enter":                3, // Action keys are important
		"enter select":         3,
		"enter choose":         3,
		"enter select/confirm": 3,
		"↑↓":                   4, // Navigation is common
		"↑/k move up":          4,
		"↓/j move down":        4,
		"←/h":                  5,
		"→/l":                  5,
		"space":                6,
		"tab":                  7,
	}

	// Sort items by priority (lower number = higher priority)
	prioritized := make([]string, len(items))
	copy(prioritized, items)

	// Simple priority-based sorting
	for i := 0; i < len(prioritized)-1; i++ {
		for j := i + 1; j < len(prioritized); j++ {
			iPriority := m.getItemPriority(prioritized[i], priorities)
			jPriority := m.getItemPriority(prioritized[j], priorities)

			if iPriority > jPriority {
				prioritized[i], prioritized[j] = prioritized[j], prioritized[i]
			}
		}
	}

	return prioritized
}

// getItemPriority returns the priority of a help item
func (m *ResponsiveHelpModel) getItemPriority(item string, priorities map[string]int) int {
	// Check for exact matches first
	for key, priority := range priorities {
		if strings.Contains(item, key) {
			return priority
		}
	}

	// Default priority for unknown items
	return 999
}

// SetSize implements layout.Sizeable
func (m *ResponsiveHelpModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable
func (m *ResponsiveHelpModel) GetSize() (int, int) {
	return m.width, m.height
}

// SetBindings updates the key bindings to display
func (m *ResponsiveHelpModel) SetBindings(bindings []key.Binding) {
	m.bindings = bindings
}

// GetBindings returns the current key bindings
func (m *ResponsiveHelpModel) GetBindings() []key.Binding {
	return m.bindings
}

// Ensure ResponsiveHelpModel implements the required interfaces
var _ util.Model = (*ResponsiveHelpModel)(nil)
var _ layout.Sizeable = (*ResponsiveHelpModel)(nil)
