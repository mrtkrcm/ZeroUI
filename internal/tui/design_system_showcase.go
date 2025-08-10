package tui

// Refactored design system showcase using modular architecture
// Components are now organized across focused modules:
// - themes/colors.go: Color schemes and palettes
// - themes/typography.go: Typography and text styles  
// - components/buttons.go: Button styles and interactions
// - components/forms.go: Input fields and validation
// - components/layout.go: Spacing, layout, and containers
// - showcase/examples.go: Demo implementations
// - showcase/renderer.go: Main orchestration and rendering

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/tui/showcase"
	"github.com/mrtkrcm/ZeroUI/internal/tui/themes"
)

// DesignSystemShowcase represents the design system showcase application
type DesignSystemShowcase struct {
	logger      *logger.Logger
	interactive bool
	program     *tea.Program
}

// NewDesignSystemShowcase creates a new design system showcase
func NewDesignSystemShowcase(log *logger.Logger, interactive bool) *DesignSystemShowcase {
	return &DesignSystemShowcase{
		logger:      log,
		interactive: interactive,
	}
}

// Run starts the design system showcase
func (d *DesignSystemShowcase) Run() error {
	model := newDesignSystemModel(d.logger, d.interactive)
	d.program = tea.NewProgram(model, tea.WithAltScreen())

	if _, err := d.program.Run(); err != nil {
		return fmt.Errorf("design system showcase error: %w", err)
	}

	return nil
}

// Use showcase package types
type ShowcaseSection = showcase.ShowcaseSection

const (
	OverviewSection     = showcase.OverviewSection
	ColorsSection       = showcase.ColorsSection
	TypographySection   = showcase.TypographySection
	ComponentsSection   = showcase.ComponentsSection
	LayoutSection       = showcase.LayoutSection
	InteractiveSection  = showcase.InteractiveSection
	AnimationsSection   = showcase.AnimationsSection
	ErrorStatesSection  = showcase.ErrorStatesSection
	BoxDrawingSection   = showcase.BoxDrawingSection
	RealExamplesSection = showcase.RealExamplesSection
)

// DesignSystemModel represents the design system showcase model
type DesignSystemModel struct {
	logger      *logger.Logger
	interactive bool

	// Modular components
	renderer        *showcase.ShowcaseRenderer
	layoutCalc      *showcase.LayoutCalculator

	// Navigation
	sections       []showcase.ShowcaseItem
	currentSection int
	sectionList    list.Model

	// Components for demonstration
	spinner     spinner.Model
	progress    progress.Model
	textInput   textinput.Model
	viewport    viewport.Model

	// State
	loading     bool
	progressVal float64
	
	// Layout
	width  int
	height int

	// Demo state
	demoIndex   int
	animTicker  time.Time
}

// Use showcase package types
type ShowcaseItem = showcase.ShowcaseItem

// newDesignSystemModel creates a new design system showcase model
func newDesignSystemModel(log *logger.Logger, interactive bool) *DesignSystemModel {
	// Initialize modular components
	renderer := showcase.NewShowcaseRenderer()
	layoutCalc := showcase.NewLayoutCalculator()
	sections := renderer.GetSections()

	// Create list using renderer
	sectionList := renderer.CreateSectionList()

	// Initialize components with proper styling
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.PrimaryColor))

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	ti := textinput.New()
	ti.Placeholder = "Type something here..."
	ti.Width = 30

	vp := viewport.New(0, 0)
	vp.SetContent("This is viewport content that can scroll vertically when the content is longer than the available space.")

	return &DesignSystemModel{
		logger:      log,
		interactive: interactive,
		renderer:    renderer,
		layoutCalc:  layoutCalc,
		sections:    sections,
		sectionList: sectionList,
		spinner:     s,
		progress:    p,
		textInput:   ti,
		viewport:    vp,
		animTicker:  time.Now(),
	}
}

// Init initializes the model
func (m *DesignSystemModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.tickProgress(),
	)
}

// Update handles messages and updates the model
func (m *DesignSystemModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.interactive {
				m.currentSection = (m.currentSection + 1) % len(m.sections)
				m.sectionList.Select(m.currentSection)
			}
		case "shift+tab":
			if m.interactive {
				m.currentSection = (m.currentSection - 1 + len(m.sections)) % len(m.sections)
				m.sectionList.Select(m.currentSection)
			}
		case "enter", " ":
			if m.interactive {
				if selectedItem := m.sectionList.SelectedItem(); selectedItem != nil {
					if section, ok := selectedItem.(ShowcaseItem); ok {
						m.currentSection = int(section.Section)
					}
				}
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			if m.interactive {
				sectionNum := int(msg.String()[0] - '0')
				if sectionNum >= 1 && sectionNum <= len(m.sections) {
					m.currentSection = sectionNum - 1
					m.sectionList.Select(m.currentSection)
				}
			}
		}

		// Handle section-specific input
		if m.sections[m.currentSection].Section == InteractiveSection {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progressTickMsg:
		m.progressVal += 0.02
		if m.progressVal > 1.0 {
			m.progressVal = 0.0
		}
		cmds = append(cmds, m.tickProgress())

	case tea.MouseMsg:
		// Handle mouse interactions if needed
		if m.interactive && msg.Type == tea.MouseLeft {
			// Could add click-to-navigate functionality here
		}
	}

	// Update components based on current section
	if m.interactive {
		var cmd tea.Cmd
		m.sectionList, cmd = m.sectionList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m *DesignSystemModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if !m.interactive {
		return m.renderStaticShowcase()
	}

	// Interactive mode with sidebar navigation  
	sidebar := m.renderer.RenderSidebar(m.sectionList, m.width, m.height)
	content := m.renderCurrentSection()

	// Layout side by side
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		content,
	)
}

// renderSidebar is now handled by the showcase renderer
// Kept for backward compatibility but delegates to renderer
func (m *DesignSystemModel) renderSidebar() string {
	return m.renderer.RenderSidebar(m.sectionList, m.width, m.height)
}

// renderCurrentSection renders the current section content
func (m *DesignSystemModel) renderCurrentSection() string {
	contentWidth := m.layoutCalc.CalculateContentWidth(m.width, m.interactive)
	section := m.sections[m.currentSection]

	// Prepare render parameters
	params := showcase.RenderParams{
		ProgressVal: m.progressVal,
		Progress:    m.progress,
		Spinner:     m.spinner,
		TextInput:   m.textInput,
		AnimTicker:  m.animTicker,
	}

	return m.renderer.RenderSection(section.Section, contentWidth, params)
}

// renderStaticShowcase renders all sections in a static format
func (m *DesignSystemModel) renderStaticShowcase() string {
	// Prepare render parameters
	params := showcase.RenderParams{
		ProgressVal: m.progressVal,
		Progress:    m.progress,
		Spinner:     m.spinner,
		TextInput:   m.textInput,
		AnimTicker:  m.animTicker,
	}

	return m.renderer.RenderStaticShowcase(m.width, params)
}

// updateLayout updates component layouts
func (m *DesignSystemModel) updateLayout() {
	if !m.interactive {
		return
	}

	// Use layout calculator for responsive layout
	contentWidth := m.layoutCalc.CalculateContentWidth(m.width, m.interactive)
	m.layoutCalc.UpdateComponentDimensions(contentWidth, m.textInput, m.progress)

	// Direct updates for components that need specific handling
	if contentWidth > 50 {
		m.textInput.Width = contentWidth - 20
	}
	if contentWidth > 60 {
		m.progress.Width = contentWidth - 20
	}
}

// Custom messages
type progressTickMsg struct{}

func (m *DesignSystemModel) tickProgress() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return progressTickMsg{}
	})
}

// showcaseDelegate is now handled by the showcase renderer
// Kept for backward compatibility but delegates to renderer
type showcaseDelegate = showcase.ShowcaseDelegate

// All styles have been moved to modular components:
// - themes/colors.go: Color schemes and status colors
// - themes/typography.go: Text styles and emphasis 
// - components/buttons.go: Button styling
// - components/forms.go: Input and validation styles
// - components/layout.go: Layout, spacing, and containers
// - showcase/renderer.go: List and navigation styles

// This file now serves as the main orchestrator, delegating
// all styling and rendering to the appropriate modules.