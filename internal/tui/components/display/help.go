package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ContextualHelp provides context-aware help
type ContextualHelp struct {
	currentContext string
	helpMaps       map[string][]HelpItem
	isVisible      bool
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

// GetContext returns the current help context
func (h ContextualHelp) GetContext() string {
	return h.currentContext
}

// AddHelpMap adds help items for a context
func (h *ContextualHelp) AddHelpMap(context string, items []HelpItem) {
	h.helpMaps[context] = items
}

// GetHelpItems returns help items for the current context
func (h ContextualHelp) GetHelpItems() []HelpItem {
	if items, exists := h.helpMaps[h.currentContext]; exists {
		return items
	}
	return []HelpItem{}
}

// GetHelpItemsForContext returns help items for a specific context
func (h ContextualHelp) GetHelpItemsForContext(context string) []HelpItem {
	if items, exists := h.helpMaps[context]; exists {
		return items
	}
	return []HelpItem{}
}

// Toggle toggles help visibility
func (h *ContextualHelp) Toggle() {
	h.isVisible = !h.isVisible
}

// Show shows the help
func (h *ContextualHelp) Show() {
	h.isVisible = true
}

// Hide hides the help
func (h *ContextualHelp) Hide() {
	h.isVisible = false
}

// IsVisible returns whether help is currently visible
func (h ContextualHelp) IsVisible() bool {
	return h.isVisible
}

// View renders the contextual help
func (h ContextualHelp) View(width, height int) string {
	if !h.isVisible {
		return ""
	}

	items := h.GetHelpItems()
	if len(items) == 0 {
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
