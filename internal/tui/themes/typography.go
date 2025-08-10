package themes

import "github.com/charmbracelet/lipgloss"

// TypographyTheme defines the interface for typography themes
type TypographyTheme interface {
	H1() lipgloss.Style
	H2() lipgloss.Style
	H3() lipgloss.Style
	Body() lipgloss.Style
	Caption() lipgloss.Style
	CodeInline() lipgloss.Style
	CodeBlock() lipgloss.Style
	Bold() lipgloss.Style
	Italic() lipgloss.Style
	Underline() lipgloss.Style
	Strikethrough() lipgloss.Style
}

// DefaultTypographyTheme implements the default ZeroUI typography
type DefaultTypographyTheme struct{}

// NewDefaultTypographyTheme creates a new default typography theme
func NewDefaultTypographyTheme() *DefaultTypographyTheme {
	return &DefaultTypographyTheme{}
}

// Title hierarchy styles
func (t *DefaultTypographyTheme) H1() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true).
		MarginBottom(1)
}

func (t *DefaultTypographyTheme) H2() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true)
}

func (t *DefaultTypographyTheme) H3() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))
}

// Text styles
func (t *DefaultTypographyTheme) Body() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))
}

func (t *DefaultTypographyTheme) Caption() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4"))
}

// Code styles
func (t *DefaultTypographyTheme) CodeInline() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Background(lipgloss.Color("#44475A")).
		Padding(0, 1)
}

func (t *DefaultTypographyTheme) CodeBlock() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Background(lipgloss.Color("#282A36")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1)
}

// Emphasis styles
func (t *DefaultTypographyTheme) Bold() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Bold(true)
}

func (t *DefaultTypographyTheme) Italic() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Italic(true)
}

func (t *DefaultTypographyTheme) Underline() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Underline(true)
}

func (t *DefaultTypographyTheme) Strikethrough() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Strikethrough(true)
}

// Global typography styles for common use
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	ContentTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				MarginBottom(1)

	ComponentTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF79C6")).
				Bold(true)

	LayoutTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8BE9FD")).
				Bold(true)

	// Common text styles
	ContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	// Code styles
	CodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))
)

// Typography demonstration content
func GetTypographyDemo() (string, string, string, string) {
	titleDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		TitleStyle.Render("H1 Title Style - Large and Bold"),
		ContentTitleStyle.Render("H2 Subtitle Style - Medium Weight"),
		ComponentTitleStyle.Render("H3 Section Style - Regular Weight"),
	)

	textDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		ContentStyle.Render("Body text - Regular weight for main content"),
		HelpStyle.Render("Caption text - Smaller size for secondary info"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Background(lipgloss.Color("#44475A")).
			Padding(0, 1).
			Render("Inline code - Monospace font"),
	)

	emphasisDemo := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Bold(true).Render("Bold text for emphasis"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Italic(true).Render("Italic text for highlights"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2")).Underline(true).Render("Underlined text for links"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Strikethrough(true).Render("Strikethrough for deleted content"),
	)

	codeDemo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Background(lipgloss.Color("#282A36")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1).
		Render(`// Lipgloss style definition example
titleStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1).
    Bold(true)`)

	return titleDemo, textDemo, emphasisDemo, codeDemo
}
