package tui

// This file provides backward compatibility by re-exporting the enhanced UX components
// that have been moved to separate files for better organization.

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
)

// Re-export types for backward compatibility
type (
	ToastLevel       = components.ToastLevel
	Toast            = components.Toast
	ToastTimeoutMsg  = components.ToastTimeoutMsg
	SearchModel      = components.SearchModel
	StatusBar        = components.EnhancedStatusBar
	ContextualHelp   = components.ContextualHelp
	HelpItem         = components.HelpItem
	LoadingIndicator = components.LoadingIndicator
	LoadingTickMsg   = components.LoadingTickMsg
	EnhancedKeyMap   = components.EnhancedKeyMap
)

// Re-export constants
const (
	ToastInfo    = components.ToastInfo
	ToastSuccess = components.ToastSuccess
	ToastWarning = components.ToastWarning
	ToastError   = components.ToastError
)

// Re-export constructor functions for backward compatibility
var (
	NewSearchModel      = components.NewSearchModel
	NewStatusBar        = components.NewEnhancedStatusBar
	NewContextualHelp   = components.NewContextualHelp
	NewLoadingIndicator = components.NewLoadingIndicator
	NewEnhancedKeyMap   = components.NewEnhancedKeyMap
)

// Legacy wrapper functions for any existing code that might depend on them

// ShowToast creates and displays a toast notification (legacy wrapper)
func ShowToast(statusBar *StatusBar, message string, level ToastLevel, duration time.Duration) tea.Cmd {
	return statusBar.ShowToast(message, level, duration)
}

// ActivateSearch enables search mode (legacy wrapper)
func ActivateSearch(searchModel *SearchModel) tea.Cmd {
	return searchModel.ActivateSearch()
}

// StartLoading starts the loading indicator (legacy wrapper)
func StartLoading(loader *LoadingIndicator, message string) tea.Cmd {
	return loader.Start(message)
}

// RenderHelp renders contextual help (legacy wrapper)
func RenderHelp(help *ContextualHelp, width, height int) string {
	return help.View(width, height)
}

// CreateEnhancedStyles creates enhanced styling for the UX components
func CreateEnhancedStyles() map[string]lipgloss.Style {
	return map[string]lipgloss.Style{
		"toast_info": lipgloss.NewStyle().
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true),
		"toast_success": lipgloss.NewStyle().
			Background(lipgloss.Color("34")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true),
		"toast_warning": lipgloss.NewStyle().
			Background(lipgloss.Color("220")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true),
		"toast_error": lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Bold(true),
		"search_active": lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Margin(1, 0),
		"status_bar": lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1),
		"loading": lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true),
		"help_overlay": lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Margin(1).
			Background(lipgloss.Color("235")),
	}
}

// GetDefaultKeyBindings returns the default enhanced key bindings
func GetDefaultKeyBindings() EnhancedKeyMap {
	return NewEnhancedKeyMap()
}

// GetContextualKeyHelp returns contextual help for key bindings
func GetContextualKeyHelp(keyMap EnhancedKeyMap, context string) []HelpItem {
	return keyMap.GetContextualHelp(context)
}
