package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EnhancedStatusBar provides contextual information and feedback
type EnhancedStatusBar struct {
	currentApp  string
	currentView string
	keyHints    []string
	toast       Toast
	isVisible   bool
	debugMode   bool
}

// NewEnhancedStatusBar creates a new enhanced status bar
func NewEnhancedStatusBar() EnhancedStatusBar {
	return EnhancedStatusBar{
		isVisible: true,
	}
}

// SetCurrentApp updates the current app context
func (s *EnhancedStatusBar) SetCurrentApp(app string) {
	s.currentApp = app
}

// SetCurrentView updates the current view context
func (s *EnhancedStatusBar) SetCurrentView(view string) {
	s.currentView = view
}

// SetKeyHints updates the contextual key hints
func (s *EnhancedStatusBar) SetKeyHints(hints []string) {
	s.keyHints = hints
}

// SetVisible controls status bar visibility
func (s *EnhancedStatusBar) SetVisible(visible bool) {
	s.isVisible = visible
}

// IsVisible returns whether the status bar is visible
func (s EnhancedStatusBar) IsVisible() bool {
	return s.isVisible
}

// SetDebugMode enables or disables debug mode
func (s *EnhancedStatusBar) SetDebugMode(enabled bool) {
	s.debugMode = enabled
}

// ShowToast displays a toast notification
func (s *EnhancedStatusBar) ShowToast(message string, level ToastLevel, duration time.Duration) tea.Cmd {
	s.toast = NewToast(message, level, duration)
	return s.toast.CreateTimeoutCmd()
}

// Update handles status bar updates
func (s *EnhancedStatusBar) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ToastTimeoutMsg:
		if s.toast.IsExpired(msg) {
			s.toast.Active = false
		}
	}
	return nil
}

// View renders the status bar
func (s EnhancedStatusBar) View(width int) string {
	if !s.isVisible {
		return ""
	}

	// Left side: Current context
	leftContent := s.buildLeftContent()

	// Right side: Key hints
	rightContent := s.buildRightContent()

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
		toastContent = s.toast.Render()
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

// buildLeftContent constructs the left side content
func (s EnhancedStatusBar) buildLeftContent() string {
	var parts []string

	if s.currentApp != "" {
		parts = append(parts, fmt.Sprintf("📱 %s", s.currentApp))
	}

	if s.currentView != "" {
		parts = append(parts, fmt.Sprintf("🎯 %s", s.currentView))
	}

	return strings.Join(parts, " • ")
}

// buildRightContent constructs the right side content
func (s EnhancedStatusBar) buildRightContent() string {
	if len(s.keyHints) > 0 {
		return strings.Join(s.keyHints, " • ")
	}
	return ""
}
