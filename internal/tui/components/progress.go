package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ProgressModel represents an enhanced progress component
type ProgressModel struct {
	progress   progress.Model
	styles     *styles.Styles
	label      string
	value      float64
	animated   bool
	showPercent bool
}

// ProgressUpdateMsg represents progress updates
type ProgressUpdateMsg struct {
	Value float64
	Label string
}

// NewProgress creates a new progress component
func NewProgress(width int, animated bool) *ProgressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = width

	return &ProgressModel{
		progress:    p,
		styles:      styles.GetStyles(),
		animated:    animated,
		showPercent: true,
	}
}

// Init implements tea.Model
func (m *ProgressModel) Init() tea.Cmd {
	if m.animated {
		return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return ProgressAnimationMsg{}
		})
	}
	return nil
}

// ProgressAnimationMsg represents animation ticks
type ProgressAnimationMsg struct{}

// Update implements tea.Model
func (m *ProgressModel) Update(msg tea.Msg) (*ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressUpdateMsg:
		m.value = msg.Value
		m.label = msg.Label
		return m, nil

	case ProgressAnimationMsg:
		if m.animated {
			return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
				return ProgressAnimationMsg{}
			})
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 4
		return m, nil
	}

	var cmd tea.Cmd
	progressModel, progressCmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	cmd = progressCmd
	return m, cmd
}

// View implements tea.Model
func (m *ProgressModel) View() string {
	var parts []string

	// Add label if present
	if m.label != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			MarginBottom(1)
		parts = append(parts, labelStyle.Render(m.label))
	}

	// Add progress bar
	parts = append(parts, m.progress.ViewAs(m.value))

	// Add percentage if enabled
	if m.showPercent {
		percentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Align(lipgloss.Right).
			Width(m.progress.Width).
			MarginTop(1)
		
		percent := int(m.value * 100)
		parts = append(parts, percentStyle.Render(fmt.Sprintf("%d%%", percent)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetProgress updates the progress value
func (m *ProgressModel) SetProgress(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	m.value = value
}

// SetLabel updates the progress label
func (m *ProgressModel) SetLabel(label string) {
	m.label = label
}

// SetWidth updates the progress width
func (m *ProgressModel) SetWidth(width int) {
	m.progress.Width = width
}

// GetProgress returns current progress value
func (m *ProgressModel) GetProgress() float64 {
	return m.value
}

// MultiProgressModel represents multiple progress bars
type MultiProgressModel struct {
	progresses []*ProgressModel
	labels     []string
	values     []float64
	styles     *styles.Styles
	width      int
}

// NewMultiProgress creates a new multi-progress component
func NewMultiProgress(count int, width int) *MultiProgressModel {
	progresses := make([]*ProgressModel, count)
	labels := make([]string, count)
	values := make([]float64, count)

	for i := 0; i < count; i++ {
		progresses[i] = NewProgress(width, false)
	}

	return &MultiProgressModel{
		progresses: progresses,
		labels:     labels,
		values:     values,
		styles:     styles.GetStyles(),
		width:      width,
	}
}

// Init implements tea.Model
func (m *MultiProgressModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, p := range m.progresses {
		if cmd := p.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *MultiProgressModel) Update(msg tea.Msg) (*MultiProgressModel, tea.Cmd) {
	var cmds []tea.Cmd

	// Update all progress bars
	for i, p := range m.progresses {
		updated, cmd := p.Update(msg)
		m.progresses[i] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// View implements tea.Model
func (m *MultiProgressModel) View() string {
	var parts []string

	for i, p := range m.progresses {
		if m.labels[i] != "" {
			p.SetLabel(m.labels[i])
		}
		p.SetProgress(m.values[i])
		parts = append(parts, p.View())
		
		// Add spacing between progress bars
		if i < len(m.progresses)-1 {
			parts = append(parts, "")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetProgress updates a specific progress bar
func (m *MultiProgressModel) SetProgress(index int, value float64, label string) {
	if index >= 0 && index < len(m.progresses) {
		m.values[index] = value
		m.labels[index] = label
	}
}

// SetAllProgress updates all progress bars
func (m *MultiProgressModel) SetAllProgress(values []float64, labels []string) {
	for i := 0; i < len(m.progresses) && i < len(values); i++ {
		m.values[i] = values[i]
		if i < len(labels) {
			m.labels[i] = labels[i]
		}
	}
}