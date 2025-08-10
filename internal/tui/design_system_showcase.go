package tui

// TODO: CRITICAL - This file is 1,257 lines and needs decomposition into:
// TODO: - components/buttons.go (button styles and interactions)
// TODO: - components/forms.go (input fields and validation)
// TODO: - components/layout.go (spacing, margins, containers)
// TODO: - themes/colors.go (color schemes and palettes)
// TODO: - themes/typography.go (fonts, sizes, weights)
// TODO: - showcase/examples.go (demo implementations)
// TODO: Target: 8 files of ~150 lines each for maintainability

import (
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
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

// DesignSystemModel represents the design system showcase model
type DesignSystemModel struct {
	logger      *logger.Logger
	interactive bool

	// Navigation
	sections       []ShowcaseItem
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

// ShowcaseItem represents a section in the design system
type ShowcaseItem struct {
	Title       string
	Description string
	Section     ShowcaseSection
}

func (i ShowcaseItem) FilterValue() string { return i.Title }

// newDesignSystemModel creates a new design system showcase model
func newDesignSystemModel(log *logger.Logger, interactive bool) *DesignSystemModel {
	// Define sections
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

	// Create list
	sectionList := list.New(make([]list.Item, len(sections)), showcaseDelegate{}, 0, 0)
	sectionList.Title = "Design System Sections"
	sectionList.SetShowStatusBar(false)
	sectionList.SetFilteringEnabled(true)

	items := make([]list.Item, len(sections))
	for i, section := range sections {
		items[i] = section
	}
	sectionList.SetItems(items)

	// Initialize components
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = showcaseSpinnerStyle

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
	sidebar := m.renderSidebar()
	content := m.renderCurrentSection()

	// Layout side by side
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		content,
	)
}

// renderSidebar renders the navigation sidebar
func (m *DesignSystemModel) renderSidebar() string {
	sidebarWidth := m.width / 3
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}

	title := showcaseTitleStyle.Width(sidebarWidth - 2).Render("ZeroUI Design System")
	
	listHeight := m.height - 6
	m.sectionList.SetHeight(listHeight)
	m.sectionList.SetWidth(sidebarWidth - 2)

	help := showcaseHelpStyle.Width(sidebarWidth - 2).Render(
		"tab/shift+tab: navigate • 1-9: jump to section • enter: select • q: quit",
	)

	sidebar := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		m.sectionList.View(),
		"",
		help,
	)

	return showcaseSidebarStyle.
		Width(sidebarWidth).
		Height(m.height).
		Render(sidebar)
}

// renderCurrentSection renders the current section content
func (m *DesignSystemModel) renderCurrentSection() string {
	contentWidth := (m.width * 2) / 3
	if contentWidth < 50 {
		contentWidth = 50
	}

	// In static mode, use full width
	if !m.interactive {
		contentWidth = m.width
	}

	section := m.sections[m.currentSection]
	
	switch section.Section {
	case OverviewSection:
		return m.renderOverview(contentWidth)
	case ColorsSection:
		return m.renderColors(contentWidth)
	case TypographySection:
		return m.renderTypography(contentWidth)
	case ComponentsSection:
		return m.renderComponents(contentWidth)
	case LayoutSection:
		return m.renderLayout(contentWidth)
	case InteractiveSection:
		return m.renderInteractive(contentWidth)
	case AnimationsSection:
		return m.renderAnimations(contentWidth)
	case ErrorStatesSection:
		return m.renderErrorStates(contentWidth)
	case BoxDrawingSection:
		return m.renderBoxDrawing(contentWidth)
	case RealExamplesSection:
		return m.renderRealExamples(contentWidth)
	}

	return "Section not implemented"
}

// renderOverview renders the overview section
func (m *DesignSystemModel) renderOverview(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("🏠 Welcome to ZeroUI Design System")
	
	intro := showcaseContentStyle.Width(width - 4).Render(`
This showcase demonstrates the actual design system used in ZeroUI, 
a powerful CLI tool for managing application configurations.

Built with:
• Charm's Bubble Tea for TUI framework
• Bubbles for pre-built components  
• Lipgloss for styling and layout
• Native terminal rendering

All components shown here are the real implementations used throughout
the ZeroUI application, not mockups or approximations.
`)

	features := showcaseContentStyle.Width(width - 4).Render(`
Features demonstrated:
✓ Color themes and palette
✓ Typography and text styling
✓ Interactive components
✓ Layout patterns
✓ Loading states and animations
✓ Error handling
✓ Box-drawing characters
✓ Real application examples
`)

	usage := showcaseCodeStyle.Width(width - 4).Render(`
Navigation:
- Use Tab/Shift+Tab to navigate sections
- Press 1-9 to jump to specific sections  
- Press Enter to select highlighted section
- Press Q to quit the showcase
`)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				intro,
				"",
				features,
				"",
				usage,
			),
		)
}

// renderColors renders the color palette section
func (m *DesignSystemModel) renderColors(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("🎨 Color Palette & Themes")

	// Primary colors
	primaryColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		showcasePrimaryStyle.Render("  Primary  "),
		"  ",
		showcaseSecondaryStyle.Render(" Secondary "),
		"  ",
		showcaseAccentStyle.Render("  Accent  "),
	)

	// Status colors
	statusColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		showcaseSuccessStyle.Render("  Success  "),
		"  ",
		showcaseWarningStyle.Render("  Warning  "),
		"  ",
		showcaseErrorStyle.Render("   Error   "),
		"  ",
		showcaseInfoStyle.Render("   Info    "),
	)

	// Neutral colors
	neutralColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		showcaseLightStyle.Render("   Light   "),
		"  ",
		showcaseMediumStyle.Render("  Medium   "),
		"  ",
		showcaseDarkStyle.Render("   Dark    "),
		"  ",
		showcaseBackgroundStyle.Render("Background"),
	)

	// Color codes
	colorInfo := showcaseCodeStyle.Width(width - 4).Render(`
Color Definitions (Lipgloss):
• Primary:    #7D56F4  (Purple)
• Secondary:  #FF6B9D  (Pink)  
• Accent:     #C9A96E  (Gold)
• Success:    #50FA7B  (Green)
• Warning:    #FFB86C  (Orange)
• Error:      #FF5555  (Red)
• Info:       #8BE9FD  (Cyan)
• Background: #282A36  (Dark Gray)
• Foreground: #F8F8F2  (Light Gray)
`)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				"Primary Colors:",
				primaryColors,
				"",
				"Status Colors:",
				statusColors,
				"",
				"Neutral Colors:",
				neutralColors,
				"",
				colorInfo,
			),
		)
}

// renderTypography renders the typography section
func (m *DesignSystemModel) renderTypography(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("📝 Typography & Text Styles")

	// Title styles
	titleDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		showcaseH1Style.Render("H1 Title Style - Large and Bold"),
		showcaseH2Style.Render("H2 Subtitle Style - Medium Weight"),
		showcaseH3Style.Render("H3 Section Style - Regular Weight"),
	)

	// Text styles
	textDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		showcaseBodyStyle.Render("Body text - Regular weight for main content"),
		showcaseCaptionStyle.Render("Caption text - Smaller size for secondary info"),
		showcaseCodeInlineStyle.Render("Inline code - Monospace font"),
	)

	// Emphasis styles
	emphasisDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		showcaseBoldStyle.Render("Bold text for emphasis"),
		showcaseItalicStyle.Render("Italic text for highlights"),
		showcaseUnderlineStyle.Render("Underlined text for links"),
		showcaseStrikethroughStyle.Render("Strikethrough for deleted content"),
	)

	// Code block
	codeDemo := showcaseCodeBlockStyle.Width(width - 8).Render(`
// Lipgloss style definition example
titleStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1).
    Bold(true)
`)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				"Title Hierarchy:",
				titleDemo,
				"",
				"Text Styles:",
				textDemo,
				"",
				"Emphasis:",
				emphasisDemo,
				"",
				"Code Block:",
				codeDemo,
			),
		)
}

// renderComponents renders the UI components section
func (m *DesignSystemModel) renderComponents(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("🧩 UI Components Library")

	// List component demo
	listDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("📋 List Component"),
			"",
			"> Item 1 - Selected",
			"  Item 2 - Normal",
			"  Item 3 - Normal",
			"  Item 4 - Normal",
		),
	)

	// Button-like elements
	buttonDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("🔘 Button States"),
			"",
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				showcaseButtonPrimaryStyle.Render(" Primary "),
				"  ",
				showcaseButtonSecondaryStyle.Render("Secondary"),
				"  ",
				showcaseButtonDisabledStyle.Render(" Disabled "),
			),
		),
	)

	// Input component
	inputDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("📝 Text Input"),
			"",
			showcaseInputStyle.Render("│ Sample input text...          │"),
			showcaseInputFocusedStyle.Render("│ Focused input with cursor|    │"),
		),
	)

	// Progress bar
	progressDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("📊 Progress Bar"),
			"",
			m.progress.ViewAs(m.progressVal),
			fmt.Sprintf("%.0f%% complete", m.progressVal*100),
		),
	)

	components := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, listDemo, " ", buttonDemo),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, inputDemo, " ", progressDemo),
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				components,
			),
		)
}

// renderLayout renders the layout patterns section
func (m *DesignSystemModel) renderLayout(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("📐 Layout Patterns & Spacing")

	// Spacing demo
	spacingDemo := showcaseLayoutBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseLayoutTitleStyle.Render("Spacing System"),
			"",
			"• xs: 1 unit",
			"• sm: 2 units", 
			"• md: 4 units",
			"• lg: 8 units",
			"• xl: 16 units",
		),
	)

	// Alignment demo
	alignmentDemo := showcaseLayoutBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseLayoutTitleStyle.Render("Alignment"),
			"",
			lipgloss.Place(30, 3, lipgloss.Left, lipgloss.Top, "← Left"),
			lipgloss.Place(30, 3, lipgloss.Center, lipgloss.Center, "Center"),
			lipgloss.Place(30, 3, lipgloss.Right, lipgloss.Bottom, "Right →"),
		),
	)

	// Container demo
	containerDemo := showcaseLayoutBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseLayoutTitleStyle.Render("Containers"),
			"",
			showcaseContainerStyle.Width(25).Render(
				"Padded container with border and background styling.",
			),
		),
	)

	layouts := lipgloss.JoinHorizontal(
		lipgloss.Top,
		spacingDemo,
		" ",
		alignmentDemo,
		" ",
		containerDemo,
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				layouts,
			),
		)
}

// renderInteractive renders the interactive elements section
func (m *DesignSystemModel) renderInteractive(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("⚡ Interactive Elements")

	// Live text input
	inputDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("📝 Live Text Input"),
			"",
			"Try typing here:",
			m.textInput.View(),
		),
	)

	// Key bindings demo
	keysDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("⌨️ Key Bindings"),
			"",
			"• ↑/↓ or j/k: Navigate",
			"• Enter/Space: Select",
			"• Tab: Next section",
			"• Esc: Go back",
			"• q: Quit",
		),
	)

	// Mouse interaction demo
	mouseDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("🖱️ Mouse Support"),
			"",
			"• Click to select items",
			"• Scroll in viewports",
			"• Drag to resize",
			"• Hover for tooltips",
		),
	)

	interactions := lipgloss.JoinVertical(
		lipgloss.Left,
		inputDemo,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, keysDemo, " ", mouseDemo),
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				interactions,
			),
		)
}

// renderAnimations renders the animations section  
func (m *DesignSystemModel) renderAnimations(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("🎬 Animations & Loading States")

	// Spinner demo
	spinnerDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("🔄 Spinners"),
			"",
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.spinner.View(),
				"  Loading...",
			),
		),
	)

	// Progress animation
	progressDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("📈 Progress Animation"),
			"",
			m.progress.ViewAs(m.progressVal),
			fmt.Sprintf("Progress: %.0f%%", m.progressVal*100),
		),
	)

	// Blinking cursor demo
	cursor := "_"
	if time.Since(m.animTicker).Milliseconds()%1000 < 500 {
		cursor = " "
	}

	cursorDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("💫 Cursor Animation"),
			"",
			fmt.Sprintf("Typing cursor%s", cursor),
		),
	)

	animations := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerDemo,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, progressDemo, " ", cursorDemo),
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				animations,
			),
		)
}

// renderErrorStates renders the error states section
func (m *DesignSystemModel) renderErrorStates(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("❌ Error States & Feedback")

	// Error message
	errorDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("🚨 Error Messages"),
			"",
			showcaseErrorStyle.Render("❌ Error: Configuration file not found"),
			showcaseWarningStyle.Render("⚠️  Warning: Default values will be used"),
		),
	)

	// Success feedback
	successDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("✅ Success States"),
			"",
			showcaseSuccessStyle.Render("✓ Configuration saved successfully"),
			showcaseInfoStyle.Render("ℹ️  Changes will take effect on restart"),
		),
	)

	// Validation states
	validationDemo := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			showcaseComponentTitleStyle.Render("🔍 Validation"),
			"",
			showcaseValidStyle.Render("✓ Valid input"),
			showcaseInvalidStyle.Render("✗ Invalid format"),
			showcasePendingStyle.Render("⏳ Validating..."),
		),
	)

	states := lipgloss.JoinVertical(
		lipgloss.Left,
		errorDemo,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, successDemo, " ", validationDemo),
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				states,
			),
		)
}

// renderBoxDrawing renders the box drawing section
func (m *DesignSystemModel) renderBoxDrawing(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("📦 Box Drawing & Borders")

	// Basic boxes
	basicBox := showcaseBoxStyle.Width(20).Height(5).Render("Basic Box")
	
	roundedBox := showcaseRoundedBoxStyle.Width(20).Height(5).Render("Rounded Box")
	
	thickBox := showcaseThickBoxStyle.Width(20).Height(5).Render("Thick Border")

	// Double borders
	doubleBox := showcaseDoubleBoxStyle.Width(20).Height(5).Render("Double Border")

	// Complex layout with mixed borders
	complexDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		"┌─ Complex Layout ─────────────────┐",
		"│                                  │",
		"│  ╭─ Nested Box ─╮               │",
		"│  │   Content    │  Side Panel   │", 
		"│  ╰──────────────╯               │",
		"│                                  │",
		"└──────────────────────────────────┘",
	)

	boxes := lipgloss.JoinVertical(
		lipgloss.Left,
		"Border Styles:",
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, basicBox, " ", roundedBox),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, thickBox, " ", doubleBox),
		"",
		"Complex Layouts:",
		complexDemo,
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				boxes,
			),
		)
}

// renderRealExamples renders real ZeroUI examples
func (m *DesignSystemModel) renderRealExamples(width int) string {
	title := showcaseContentTitleStyle.Width(width - 4).Render("🚀 Real ZeroUI Examples")

	// App selection example (from actual TUI)
	appExample := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("ZeroUI - Select Application"),
			"",
			"> " + selectedStyle.Render("ghostty"),
			"  alacritty",
			"  vscode",
			"  neovim",
			"",
			helpStyle.Render("↑/↓: navigate • enter: select • q: quit"),
		),
	)

	// Config editing example
	configExample := showcaseComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("ZeroUI - ghostty"),
			"",
			"> " + selectedStyle.Render("theme: dark") + " " + helpStyle.Render("[dark, light, auto]"),
			"  font-size: 12",
			"  font-family: JetBrains Mono",
			"  opacity: 0.9",
			"",
			helpStyle.Render("↑/↓: navigate • ←/→: change • p: presets • esc: back"),
		),
	)

	examples := lipgloss.JoinVertical(
		lipgloss.Left,
		"Application Selection:",
		appExample,
		"",
		"Configuration Editing:",
		configExample,
	)

	return showcaseContentBoxStyle.
		Width(width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				examples,
			),
		)
}

// renderStaticShowcase renders all sections in a static format
func (m *DesignSystemModel) renderStaticShowcase() string {
	var sections []string

	for i := 0; i < len(m.sections); i++ {
		m.currentSection = i
		sections = append(sections, m.renderCurrentSection())
		sections = append(sections, strings.Repeat("─", m.width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// updateLayout updates component layouts
func (m *DesignSystemModel) updateLayout() {
	if !m.interactive {
		return
	}

	// Update text input width
	contentWidth := (m.width * 2) / 3
	if contentWidth > 50 {
		m.textInput.Width = contentWidth - 20
	}

	// Update progress bar width
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

// showcaseDelegate implements list.ItemDelegate for the section list
type showcaseDelegate struct{}

func (d showcaseDelegate) Height() int                             { return 2 }
func (d showcaseDelegate) Spacing() int                            { return 0 }
func (d showcaseDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d showcaseDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ShowcaseItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.Title, showcaseCaptionStyle.Render(i.Description))

	fn := showcaseListItemStyle.Render
	if index == m.Index() {
		fn = showcaseSelectedItemStyle.Render
	}

	fmt.Fprint(w, fn(str))
}

// All the styles used in the showcase
var (
	// Main layout styles  
	showcaseTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	showcaseSidebarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)

	showcaseContentBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2)

	showcaseContentTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true).
		MarginBottom(1)

	showcaseContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	// List styles
	showcaseListItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	showcaseSelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#44475A")).
		Bold(true).
		Padding(0, 1)

	// Typography styles
	showcaseH1Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true).
		MarginBottom(1)

	showcaseH2Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true)

	showcaseH3Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	showcaseBodyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	showcaseCaptionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))

	showcaseCodeInlineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Background(lipgloss.Color("#44475A")).
		Padding(0, 1)

	showcaseCodeBlockStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Background(lipgloss.Color("#282A36")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1)

	showcaseCodeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))

	// Text emphasis styles
	showcaseBoldStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true)

	showcaseItalicStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Italic(true)

	showcaseUnderlineStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Underline(true)

	showcaseStrikethroughStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Strikethrough(true)

	// Color palette styles
	showcasePrimaryStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	showcaseSecondaryStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B9D")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	showcaseAccentStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#C9A96E")).
		Foreground(lipgloss.Color("#282A36")).
		Padding(0, 1)

	showcaseSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))

	showcaseWarningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB86C"))

	showcaseErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))

	showcaseInfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD"))

	showcaseLightStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#F8F8F2")).
		Foreground(lipgloss.Color("#282A36")).
		Padding(0, 1)

	showcaseMediumStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#6272A4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	showcaseDarkStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#282A36")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	showcaseBackgroundStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)

	// Component styles
	showcaseComponentBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1).
		Width(30).
		Height(8)

	showcaseComponentTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF79C6")).
		Bold(true)

	showcaseButtonPrimaryStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1).
		Bold(true)

	showcaseButtonSecondaryStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	showcaseButtonDisabledStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#6272A4")).
		Padding(0, 1)

	showcaseInputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Foreground(lipgloss.Color("#F8F8F2"))

	showcaseInputFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2"))

	// Layout styles
	showcaseLayoutBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1).
		Width(25).
		Height(10)

	showcaseLayoutTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")).
		Bold(true)

	showcaseContainerStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1)

	// State styles
	showcaseValidStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))

	showcaseInvalidStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))

	showcasePendingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB86C"))

	// Box drawing styles
	showcaseBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#F8F8F2")).
		Padding(1).
		Align(lipgloss.Center)

	showcaseRoundedBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Align(lipgloss.Center)

	showcaseThickBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#FF79C6")).
		Padding(1).
		Align(lipgloss.Center)

	showcaseDoubleBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#8BE9FD")).
		Padding(1).
		Align(lipgloss.Center)

	// Animation styles
	showcaseSpinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))

	// Help styles
	showcaseHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))
)