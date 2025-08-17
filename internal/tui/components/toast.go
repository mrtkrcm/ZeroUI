package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

// NewToast creates a new toast notification
func NewToast(message string, level ToastLevel, duration time.Duration) Toast {
	return Toast{
		Message:   message,
		Level:     level,
		Duration:  duration,
		CreatedAt: time.Now(),
		Active:    true,
	}
}

// CreateTimeoutCmd creates a timeout command for the toast
func (t Toast) CreateTimeoutCmd() tea.Cmd {
	return tea.Tick(t.Duration, func(time.Time) tea.Msg {
		return ToastTimeoutMsg{t.Message}
	})
}

// IsExpired checks if the toast should be hidden
func (t Toast) IsExpired(msg ToastTimeoutMsg) bool {
	return t.Message == msg.Message
}

// Render renders the toast notification
func (t Toast) Render() string {
	if !t.Active {
		return ""
	}

	var (
		icon  string
		color lipgloss.Color
	)

	switch t.Level {
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

	return toastStyle.Render(fmt.Sprintf("%s %s", icon, t.Message))
}
