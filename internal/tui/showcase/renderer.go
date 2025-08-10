package showcase

import (
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
	"github.com/mrtkrcm/ZeroUI/internal/tui/themes"
)

// ShowcaseSection represents different sections of the showcase
type ShowcaseSection int

const (
	OverviewSection ShowcaseSection = iota
	ColorsSection
	TypographySection
	ComponentsSection
	LayoutSection
	InteractiveSection
	AnimationsSection
	ErrorStatesSection
	BoxDrawingSection
	RealExamplesSection
)

// ShowcaseItem represents a section in the design system
type ShowcaseItem struct {
	Title       string
	Description string
	Section     ShowcaseSection
}

func (i ShowcaseItem) FilterValue() string { return i.Title }

// ShowcaseRenderer handles rendering of the design system showcase
type ShowcaseRenderer struct {
	exampleRenderer *ExampleRenderer
	sections        []ShowcaseItem
	listDelegate    ShowcaseDelegate
}

// NewShowcaseRenderer creates a new showcase renderer
func NewShowcaseRenderer() *ShowcaseRenderer {
	sections := []ShowcaseItem{
		{Title: "🏠 Overview", Description: "Welcome to ZeroUI Design System", Section: OverviewSection},
		{Title: "🎨 Colors & Themes", Description: "Color palette and theme system", Section: ColorsSection},
		{Title: "📝 Typography", Description: "Text styles and formatting", Section: TypographySection},
		{Title: "🧩 Components", Description: "UI components library", Section: ComponentsSection},
		{Title: "📐 Layout", Description: "Spacing and layout patterns", Section: LayoutSection},
		{Title: "⚡ Interactive", Description: "Interactive elements demo", Section: InteractiveSection},
		{Title: "🎬 Animations", Description: "Loading states and transitions", Section: AnimationsSection},
		{Title: "❌ Error States", Description: "Error handling and feedback", Section: ErrorStatesSection},
		{Title: "📦 Box Drawing", Description: "Borders and decorative elements", Section: BoxDrawingSection},
		{Title: "🚀 Real Examples", Description: "Actual ZeroUI components", Section: RealExamplesSection},
	}

	return &ShowcaseRenderer{
		exampleRenderer: NewExampleRenderer(),
		sections:        sections,
		listDelegate:    ShowcaseDelegate{},
	}
}

// GetSections returns the available showcase sections
func (r *ShowcaseRenderer) GetSections() []ShowcaseItem {
	return r.sections
}

// CreateSectionList creates a list model for section navigation
func (r *ShowcaseRenderer) CreateSectionList() list.Model {
	sectionList := list.New(make([]list.Item, len(r.sections)), r.listDelegate, 0, 0)
	sectionList.Title = "Design System Sections"
	sectionList.SetShowStatusBar(false)
	sectionList.SetFilteringEnabled(true)

	items := make([]list.Item, len(r.sections))
	for i, section := range r.sections {
		items[i] = section
	}
	sectionList.SetItems(items)

	return sectionList
}

// RenderSidebar renders the navigation sidebar
func (r *ShowcaseRenderer) RenderSidebar(sectionList list.Model, width, height int) string {
	sidebarWidth := width / 3
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}

	title := themes.TitleStyle.Width(sidebarWidth - 2).Render("ZeroUI Design System")
	
	listHeight := height - 6
	sectionList.SetHeight(listHeight)
	sectionList.SetWidth(sidebarWidth - 2)

	help := themes.HelpStyle.Width(sidebarWidth - 2).Render(
		"tab/shift+tab: navigate • 1-9: jump to section • enter: select • q: quit",
	)

	sidebar := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		sectionList.View(),
		"",
		help,
	)

	return components.SidebarStyle.
		Width(sidebarWidth).
		Height(height).
		Render(sidebar)
}

// RenderSection renders a specific showcase section
func (r *ShowcaseRenderer) RenderSection(section ShowcaseSection, width int, params RenderParams) string {
	switch section {
	case OverviewSection:
		return r.exampleRenderer.RenderOverview(width)
	case ColorsSection:
		return r.exampleRenderer.RenderColors(width)
	case TypographySection:
		return r.exampleRenderer.RenderTypography(width)
	case ComponentsSection:
		return r.exampleRenderer.RenderComponents(width, params.ProgressVal, params.Progress)
	case LayoutSection:
		return r.exampleRenderer.RenderLayout(width)
	case InteractiveSection:
		return r.exampleRenderer.RenderInteractive(width, params.TextInput)
	case AnimationsSection:
		return r.exampleRenderer.RenderAnimations(width, params.Spinner, params.ProgressVal, params.Progress, params.AnimTicker)
	case ErrorStatesSection:
		return r.exampleRenderer.RenderErrorStates(width)
	case BoxDrawingSection:
		return r.exampleRenderer.RenderBoxDrawing(width)
	case RealExamplesSection:
		return r.exampleRenderer.RenderRealExamples(width)
	}

	return "Section not implemented"
}

// RenderStaticShowcase renders all sections in a static format
func (r *ShowcaseRenderer) RenderStaticShowcase(width int, params RenderParams) string {
	var sections []string

	for _, section := range r.sections {
		content := r.RenderSection(section.Section, width, params)
		sections = append(sections, content)
		sections = append(sections, strings.Repeat("─", width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderParams holds parameters needed for rendering sections
type RenderParams struct {
	ProgressVal  float64
	Progress     progress.Model
	Spinner      spinner.Model  
	TextInput    textinput.Model
	AnimTicker   time.Time
}

// ShowcaseDelegate implements list.ItemDelegate for the section list
type ShowcaseDelegate struct{}

func (d ShowcaseDelegate) Height() int                             { return 2 }
func (d ShowcaseDelegate) Spacing() int                            { return 0 }
func (d ShowcaseDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ShowcaseDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ShowcaseItem)
	if !ok {
		return
	}

	str := lipgloss.JoinVertical(
		lipgloss.Left,
		i.Title,
		themes.HelpStyle.Render(i.Description),
	)

	fn := ListItemStyle.Render
	if index == m.Index() {
		fn = SelectedItemStyle.Render
	}

	_, _ = w.Write([]byte(fn(str)))
}

// List item styles
var (
	ListItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#44475A")).
		Bold(true).
		Padding(0, 1)
)

// LayoutCalculator helps with responsive layout calculations
type LayoutCalculator struct{}

// NewLayoutCalculator creates a new layout calculator
func NewLayoutCalculator() *LayoutCalculator {
	return &LayoutCalculator{}
}

// CalculateSidebarWidth calculates appropriate sidebar width
func (l *LayoutCalculator) CalculateSidebarWidth(totalWidth int) int {
	sidebarWidth := totalWidth / 3
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}
	return sidebarWidth
}

// CalculateContentWidth calculates content area width
func (l *LayoutCalculator) CalculateContentWidth(totalWidth int, interactive bool) int {
	if !interactive {
		return totalWidth
	}
	
	contentWidth := (totalWidth * 2) / 3
	if contentWidth < 50 {
		contentWidth = 50
	}
	return contentWidth
}

// UpdateComponentDimensions updates component dimensions based on content width
func (l *LayoutCalculator) UpdateComponentDimensions(contentWidth int, textInput, progress interface{}) {
	// Update text input width
	if contentWidth > 50 {
		// textInput.Width = contentWidth - 20 (would need proper type assertion)
	}

	// Update progress bar width  
	if contentWidth > 60 {
		// progress.Width = contentWidth - 20 (would need proper type assertion)
	}
}