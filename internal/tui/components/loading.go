package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LoadingIndicator provides visual feedback during operations
type LoadingIndicator struct {
	isActive       bool
	message        string
	currentSpinner int
	spinnerFrames  []string
}

// LoadingTickMsg advances the spinner animation
type LoadingTickMsg struct{}

// NewLoadingIndicator creates a new loading indicator
func NewLoadingIndicator() LoadingIndicator {
	return LoadingIndicator{
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// NewLoadingIndicatorWithFrames creates a loading indicator with custom spinner frames
func NewLoadingIndicatorWithFrames(frames []string) LoadingIndicator {
	if len(frames) == 0 {
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
	return LoadingIndicator{
		spinnerFrames: frames,
	}
}

// Start activates the loading indicator
func (l *LoadingIndicator) Start(message string) tea.Cmd {
	l.isActive = true
	l.message = message
	l.currentSpinner = 0

	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return LoadingTickMsg{}
	})
}

// Stop deactivates the loading indicator
func (l *LoadingIndicator) Stop() {
	l.isActive = false
	l.message = ""
}

// IsActive returns whether the loading indicator is currently active
func (l LoadingIndicator) IsActive() bool {
	return l.isActive
}

// GetMessage returns the current loading message
func (l LoadingIndicator) GetMessage() string {
	return l.message
}

// SetMessage updates the loading message
func (l *LoadingIndicator) SetMessage(message string) {
	l.message = message
}

// Update handles loading indicator updates
func (l *LoadingIndicator) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case LoadingTickMsg:
		if l.isActive {
			l.currentSpinner = (l.currentSpinner + 1) % len(l.spinnerFrames)
			return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return LoadingTickMsg{}
			})
		}
	}
	return nil
}

// View renders the loading indicator
func (l LoadingIndicator) View() string {
	if !l.isActive {
		return ""
	}

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	return loadingStyle.Render(fmt.Sprintf("%s %s", l.spinnerFrames[l.currentSpinner], l.message))
}

// ViewWithStyle renders the loading indicator with custom styling
func (l LoadingIndicator) ViewWithStyle(style lipgloss.Style) string {
	if !l.isActive {
		return ""
	}

	return style.Render(fmt.Sprintf("%s %s", l.spinnerFrames[l.currentSpinner], l.message))
}
