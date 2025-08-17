package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchModel provides fuzzy search functionality
type SearchModel struct {
	input       textinput.Model
	isActive    bool
	query       string
	results     []string
	selected    int
	placeholder string
}

// NewSearchModel creates a new search component
func NewSearchModel(placeholder string) SearchModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.Width = 30

	return SearchModel{
		input:       ti,
		placeholder: placeholder,
	}
}

// ActivateSearch enables the search mode
func (s *SearchModel) ActivateSearch() tea.Cmd {
	s.isActive = true
	s.input.Focus()
	return textinput.Blink
}

// DeactivateSearch disables the search mode
func (s *SearchModel) DeactivateSearch() {
	s.isActive = false
	s.input.Blur()
	s.query = ""
	s.results = nil
	s.selected = 0
}

// IsActive returns whether search is currently active
func (s SearchModel) IsActive() bool {
	return s.isActive
}

// GetQuery returns the current search query
func (s SearchModel) GetQuery() string {
	return s.query
}

// SetResults updates the search results
func (s *SearchModel) SetResults(results []string) {
	s.results = results
	s.selected = 0
}

// GetResults returns the current search results
func (s SearchModel) GetResults() []string {
	return s.results
}

// GetSelected returns the currently selected result index
func (s SearchModel) GetSelected() int {
	return s.selected
}

// MoveSelectionUp moves selection up in results
func (s *SearchModel) MoveSelectionUp() {
	if s.selected > 0 {
		s.selected--
	}
}

// MoveSelectionDown moves selection down in results
func (s *SearchModel) MoveSelectionDown() {
	if s.selected < len(s.results)-1 {
		s.selected++
	}
}

// Update handles search input
func (s *SearchModel) Update(msg tea.Msg) tea.Cmd {
	if !s.isActive {
		return nil
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.query = s.input.Value()

	return cmd
}

// View renders the search interface
func (s SearchModel) View() string {
	if !s.isActive {
		return ""
	}

	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Margin(1, 0)

	return searchStyle.Render(fmt.Sprintf("🔍 %s", s.input.View()))
}
