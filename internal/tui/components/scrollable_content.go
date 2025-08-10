package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ScrollableContentModel wraps content in a scrollable viewport
type ScrollableContentModel struct {
	viewport viewport.Model
	content  string
	styles   *styles.Styles
	title    string
}

// NewScrollableContent creates a new scrollable content component
func NewScrollableContent(width, height int, title string) *ScrollableContentModel {
	vp := viewport.New(width, height)
	vp.HighPerformanceRendering = true

	return &ScrollableContentModel{
		viewport: vp,
		styles:   styles.GetStyles(),
		title:    title,
	}
}

// Init implements tea.Model
func (m *ScrollableContentModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *ScrollableContentModel) Update(msg tea.Msg) (*ScrollableContentModel, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 0
		if m.title != "" {
			headerHeight = 3
		}
		
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight
		return m, nil
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m *ScrollableContentModel) View() string {
	var sections []string

	// Add title if present
	if m.title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(1, 0).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("238")).
			Width(m.viewport.Width)
		
		sections = append(sections, titleStyle.Render(m.title))
	}

	// Add viewport content
	sections = append(sections, m.viewport.View())

	// Add scrollbar indicator if needed
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Right).
			Width(m.viewport.Width).
			Render(fmt.Sprintf("%.0f%%", m.viewport.ScrollPercent()*100))
		
		sections = append(sections, scrollInfo)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// SetContent updates the viewport content
func (m *ScrollableContentModel) SetContent(content string) {
	m.content = content
	m.viewport.SetContent(content)
}

// GetContent returns the current content
func (m *ScrollableContentModel) GetContent() string {
	return m.content
}

// SetSize updates the viewport size
func (m *ScrollableContentModel) SetSize(width, height int) {
	headerHeight := 0
	if m.title != "" {
		headerHeight = 3
	}
	
	m.viewport.Width = width
	m.viewport.Height = height - headerHeight
}

// ScrollToTop scrolls to the top of the content
func (m *ScrollableContentModel) ScrollToTop() {
	m.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom of the content
func (m *ScrollableContentModel) ScrollToBottom() {
	m.viewport.GotoBottom()
}

// GetViewport returns the underlying viewport for advanced operations
func (m *ScrollableContentModel) GetViewport() *viewport.Model {
	return &m.viewport
}