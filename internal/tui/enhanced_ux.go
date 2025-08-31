package tui

// This file provides backward compatibility by re-exporting the enhanced UX components
// that have been moved to separate files for better organization.

import (
	"github.com/charmbracelet/lipgloss"

	display "github.com/mrtkrcm/ZeroUI/internal/tui/components/display"
)

// Re-export types for backward compatibility
type (
	ToastLevel      = display.ToastLevel
	Toast           = display.Toast
	ToastTimeoutMsg = display.ToastTimeoutMsg
)

// Re-export constants
const (
	ToastInfo    = display.ToastInfo
	ToastSuccess = display.ToastSuccess
	ToastWarning = display.ToastWarning
	ToastError   = display.ToastError
)

// Re-export constructor functions for backward compatibility
var (
	NewToast = display.NewToast
)

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
