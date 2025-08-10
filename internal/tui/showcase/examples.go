package showcase

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
	"github.com/mrtkrcm/ZeroUI/internal/tui/themes"
)

// ExampleRenderer provides methods to render different showcase sections
type ExampleRenderer struct {
	colorTheme      *themes.DefaultColorTheme
	typographyTheme *themes.DefaultTypographyTheme
	spacing         *components.SpacingSystem
}

// NewExampleRenderer creates a new example renderer
func NewExampleRenderer() *ExampleRenderer {
	return &ExampleRenderer{
		colorTheme:      themes.NewDefaultColorTheme(),
		typographyTheme: themes.NewDefaultTypographyTheme(),
		spacing:         components.NewSpacingSystem(),
	}
}

// RenderOverview renders the overview section example
func (e *ExampleRenderer) RenderOverview(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("🏠 Welcome to ZeroUI Design System")
	
	intro := themes.ContentStyle.Width(width - 4).Render(`
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

	features := themes.ContentStyle.Width(width - 4).Render(`
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

	usage := themes.CodeStyle.Width(width - 4).Render(`
Navigation:
- Use Tab/Shift+Tab to navigate sections
- Press 1-9 to jump to specific sections  
- Press Enter to select highlighted section
- Press Q to quit the showcase
`)

	return components.ContentBoxStyle.
		Width(width).
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

// RenderColors renders the color palette section example
func (e *ExampleRenderer) RenderColors(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("🎨 Color Palette & Themes")

	// Primary colors
	primaryColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		e.colorTheme.Primary().Render("  Primary  "),
		"  ",
		e.colorTheme.Secondary().Render(" Secondary "),
		"  ",
		e.colorTheme.Accent().Render("  Accent  "),
	)

	// Status colors
	statusColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		e.colorTheme.Success().Render("  Success  "),
		"  ",
		e.colorTheme.Warning().Render("  Warning  "),
		"  ",
		e.colorTheme.Error().Render("   Error   "),
		"  ",
		e.colorTheme.Info().Render("   Info    "),
	)

	// Neutral colors
	neutralColors := lipgloss.JoinHorizontal(
		lipgloss.Center,
		e.colorTheme.Light().Render("   Light   "),
		"  ",
		e.colorTheme.Medium().Render("  Medium   "),
		"  ",
		e.colorTheme.Dark().Render("   Dark    "),
		"  ",
		e.colorTheme.Background().Render("Background"),
	)

	// Color codes
	colorInfo := themes.CodeStyle.Width(width - 4).Render(themes.GetColorInfo())

	return components.ContentBoxStyle.
		Width(width).
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

// RenderTypography renders the typography section example
func (e *ExampleRenderer) RenderTypography(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("📝 Typography & Text Styles")

	titleDemo, textDemo, emphasisDemo, codeDemo := themes.GetTypographyDemo()

	return components.ContentBoxStyle.
		Width(width).
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

// RenderComponents renders the UI components section example
func (e *ExampleRenderer) RenderComponents(width int, progressVal float64, progress progress.Model) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("🧩 UI Components Library")

	// List component demo
	listDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("📋 List Component"),
			"",
			"> Item 1 - Selected",
			"  Item 2 - Normal",
			"  Item 3 - Normal",
			"  Item 4 - Normal",
		),
	)

	// Button-like elements
	buttonDemo := components.ComponentBoxStyle.Render(components.GetButtonsExample())

	// Input component
	inputDemo := components.ComponentBoxStyle.Render(components.GetInputExample())

	// Progress bar
	progressDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("📊 Progress Bar"),
			"",
			progress.ViewAs(progressVal),
			fmt.Sprintf("%.0f%% complete", progressVal*100),
		),
	)

	componentsLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, listDemo, " ", buttonDemo),
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, inputDemo, " ", progressDemo),
	)

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				componentsLayout,
			),
		)
}

// RenderLayout renders the layout patterns section example
func (e *ExampleRenderer) RenderLayout(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("📐 Layout Patterns & Spacing")

	// Spacing demo
	spacingDemo := components.LayoutBoxStyle.Render(components.GetSpacingDemo())

	// Alignment demo
	alignmentDemo := components.LayoutBoxStyle.Render(components.GetAlignmentDemo())

	// Container demo
	containerDemo := components.LayoutBoxStyle.Render(components.GetContainerDemo())

	layouts := lipgloss.JoinHorizontal(
		lipgloss.Top,
		spacingDemo,
		" ",
		alignmentDemo,
		" ",
		containerDemo,
	)

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				layouts,
			),
		)
}

// RenderInteractive renders the interactive elements section example
func (e *ExampleRenderer) RenderInteractive(width int, textInput textinput.Model) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("⚡ Interactive Elements")

	// Live text input
	inputDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("📝 Live Text Input"),
			"",
			"Try typing here:",
			textInput.View(),
		),
	)

	// Key bindings demo
	keysDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("⌨️ Key Bindings"),
			"",
			"• ↑/↓ or j/k: Navigate",
			"• Enter/Space: Select",
			"• Tab: Next section",
			"• Esc: Go back",
			"• q: Quit",
		),
	)

	// Mouse interaction demo
	mouseDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("🖱️ Mouse Support"),
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

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				interactions,
			),
		)
}

// RenderAnimations renders the animations section example
func (e *ExampleRenderer) RenderAnimations(width int, spinner spinner.Model, progressVal float64, progress progress.Model, animTicker time.Time) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("🎬 Animations & Loading States")

	// Spinner demo
	spinnerDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("🔄 Spinners"),
			"",
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				spinner.View(),
				"  Loading...",
			),
		),
	)

	// Progress animation
	progressDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("📈 Progress Animation"),
			"",
			progress.ViewAs(progressVal),
			fmt.Sprintf("Progress: %.0f%%", progressVal*100),
		),
	)

	// Blinking cursor demo
	cursor := "_"
	if time.Since(animTicker).Milliseconds()%1000 < 500 {
		cursor = " "
	}

	cursorDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("💫 Cursor Animation"),
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

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				animations,
			),
		)
}

// RenderErrorStates renders the error states section example
func (e *ExampleRenderer) RenderErrorStates(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("❌ Error States & Feedback")

	// Error message
	errorDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("🚨 Error Messages"),
			"",
			e.colorTheme.Error().Render("❌ Error: Configuration file not found"),
			e.colorTheme.Warning().Render("⚠️  Warning: Default values will be used"),
		),
	)

	// Success feedback
	successDemo := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.ComponentTitleStyle.Render("✅ Success States"),
			"",
			e.colorTheme.Success().Render("✓ Configuration saved successfully"),
			e.colorTheme.Info().Render("ℹ️  Changes will take effect on restart"),
		),
	)

	// Validation states
	validationDemo := components.ComponentBoxStyle.Render(components.GetValidationExample())

	states := lipgloss.JoinVertical(
		lipgloss.Left,
		errorDemo,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, successDemo, " ", validationDemo),
	)

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				states,
			),
		)
}

// RenderBoxDrawing renders the box drawing section example
func (e *ExampleRenderer) RenderBoxDrawing(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("📦 Box Drawing & Borders")

	// Basic boxes
	basicBox, roundedBox, thickBox, doubleBox := components.GetBoxDrawingDemo()

	// Complex layout with mixed borders
	complexDemo := components.GetComplexLayoutDemo()

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

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				boxes,
			),
		)
}

// RenderRealExamples renders real ZeroUI examples
func (e *ExampleRenderer) RenderRealExamples(width int) string {
	title := themes.ContentTitleStyle.Width(width - 4).Render("🚀 Real ZeroUI Examples")

	// App selection example (from actual TUI)
	appExample := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.TitleStyle.Render("ZeroUI - Select Application"),
			"",
			"> " + themes.SelectedStyle.Render("ghostty"),
			"  alacritty",
			"  vscode",
			"  neovim",
			"",
			themes.HelpStyle.Render("↑/↓: navigate • enter: select • q: quit"),
		),
	)

	// Config editing example
	configExample := components.ComponentBoxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			themes.TitleStyle.Render("ZeroUI - ghostty"),
			"",
			"> " + themes.SelectedStyle.Render("theme: dark") + " " + themes.HelpStyle.Render("[dark, light, auto]"),
			"  font-size: 12",
			"  font-family: JetBrains Mono",
			"  opacity: 0.9",
			"",
			themes.HelpStyle.Render("↑/↓: navigate • ←/→: change • p: presets • esc: back"),
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

	return components.ContentBoxStyle.
		Width(width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				examples,
			),
		)
}