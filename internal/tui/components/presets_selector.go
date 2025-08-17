package components

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PresetsSelector is a lightweight placeholder for selecting presets
type PresetsSelector struct {
	visible bool
	appName string
	width   int
	height  int
	presets []string
	index   int
}

func NewPresetsSelector() *PresetsSelector { return &PresetsSelector{} }

func (p *PresetsSelector) Show(app string)  { p.visible = true; p.appName = app }
func (p *PresetsSelector) Hide()            { p.visible = false }
func (p *PresetsSelector) IsVisible() bool  { return p.visible }
func (p *PresetsSelector) SetSize(w, h int) { p.width, p.height = w, h }
func (p *PresetsSelector) SetPresets(list []string) {
	p.presets = list
	if p.index >= len(list) {
		p.index = 0
	}
}

func (p *PresetsSelector) Update(msg tea.Msg) (*PresetsSelector, tea.Cmd) {
	if !p.visible {
		return p, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.Hide()
			return p, nil
		case "up":
			if p.index > 0 {
				p.index--
			}
		case "down":
			if p.index < len(p.presets)-1 {
				p.index++
			}
		case "enter":
			if len(p.presets) == 0 {
				return p, nil
			}
			name := p.presets[p.index]
			app := p.appName
			p.Hide()
			return p, func() tea.Msg { return PresetSelectedMsg{App: app, Name: name} }
		}
	}
	return p, nil
}

func (p *PresetsSelector) View() string {
	if !p.visible {
		return ""
	}
	title := lipgloss.NewStyle().Bold(true).Render("Presets")
	if len(p.presets) == 0 {
		body := lipgloss.NewStyle().Render(fmt.Sprintf("No presets for %s", p.appName))
		return lipgloss.JoinVertical(lipgloss.Left, title, body)
	}
	var lines []string
	for i, name := range p.presets {
		marker := "  "
		if i == p.index {
			marker = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s", marker, name))
	}
	list := lipgloss.NewStyle().Render(fmt.Sprintf("%s", lipgloss.JoinVertical(lipgloss.Left, lines...)))
	return lipgloss.JoinVertical(lipgloss.Left, title, list)
}

// PresetSelectedMsg emitted when a preset is selected
type PresetSelectedMsg struct{ App, Name string }
