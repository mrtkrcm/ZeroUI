package components

import "github.com/charmbracelet/lipgloss"

// SpacingSystem defines standard spacing units
type SpacingSystem struct {
	XS int // 1 unit
	SM int // 2 units
	MD int // 4 units
	LG int // 8 units
	XL int // 16 units
}

// NewSpacingSystem creates a new spacing system
func NewSpacingSystem() *SpacingSystem {
	return &SpacingSystem{
		XS: 1,
		SM: 2,
		MD: 4,
		LG: 8,
		XL: 16,
	}
}

// LayoutStyles defines the interface for layout styling
type LayoutStyles interface {
	Sidebar() lipgloss.Style
	ContentBox() lipgloss.Style
	ComponentBox() lipgloss.Style
	Container() lipgloss.Style
}

// DefaultLayoutStyles implements the default layout styles
type DefaultLayoutStyles struct{}

// NewDefaultLayoutStyles creates a new default layout styles
func NewDefaultLayoutStyles() *DefaultLayoutStyles {
	return &DefaultLayoutStyles{}
}

// Sidebar style
func (l *DefaultLayoutStyles) Sidebar() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)
}

// ContentBox style
func (l *DefaultLayoutStyles) ContentBox() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2)
}

// ComponentBox style
func (l *DefaultLayoutStyles) ComponentBox() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1).
		Width(30).
		Height(8)
}

// Container style
func (l *DefaultLayoutStyles) Container() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1)
}

// Global layout styles for direct use
var (
	// Sidebar and navigation styles
	SidebarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)

	// Content area styles
	ContentBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(2)

	// Component container styles
	ComponentBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1).
		Width(30).
		Height(8)

	LayoutBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#44475A")).
		Padding(1).
		Width(25).
		Height(10)

	// Container styles
	ContainerStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1)
)

// BoxDrawingStyles defines different border styles
var (
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#F8F8F2")).
		Padding(1).
		Align(lipgloss.Center)

	RoundedBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Align(lipgloss.Center)

	ThickBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#FF79C6")).
		Padding(1).
		Align(lipgloss.Center)

	DoubleBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#8BE9FD")).
		Padding(1).
		Align(lipgloss.Center)
)

// LayoutComponent represents a layout container
type LayoutComponent struct {
	Content string
	Style   lipgloss.Style
	Width   int
	Height  int
	Padding int
	Margin  int
}

// NewLayoutComponent creates a new layout component
func NewLayoutComponent(content string, style lipgloss.Style) *LayoutComponent {
	return &LayoutComponent{
		Content: content,
		Style:   style,
	}
}

// SetDimensions sets width and height
func (l *LayoutComponent) SetDimensions(width, height int) {
	l.Width = width
	l.Height = height
}

// SetSpacing sets padding and margin
func (l *LayoutComponent) SetSpacing(padding, margin int) {
	l.Padding = padding
	l.Margin = margin
}

// Render renders the layout component
func (l *LayoutComponent) Render() string {
	style := l.Style
	
	if l.Width > 0 {
		style = style.Width(l.Width)
	}
	if l.Height > 0 {
		style = style.Height(l.Height)
	}
	if l.Padding > 0 {
		style = style.Padding(l.Padding)
	}
	if l.Margin > 0 {
		style = style.Margin(l.Margin)
	}
	
	return style.Render(l.Content)
}

// GetSpacingDemo returns a demonstration of the spacing system
func GetSpacingDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true).
			Render("Spacing System"),
		"",
		"• xs: 1 unit",
		"• sm: 2 units",
		"• md: 4 units",
		"• lg: 8 units",
		"• xl: 16 units",
	)
}

// GetAlignmentDemo returns a demonstration of alignment options
func GetAlignmentDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true).
			Render("Alignment"),
		"",
		lipgloss.Place(30, 3, lipgloss.Left, lipgloss.Top, "← Left"),
		lipgloss.Place(30, 3, lipgloss.Center, lipgloss.Center, "Center"),
		lipgloss.Place(30, 3, lipgloss.Right, lipgloss.Bottom, "Right →"),
	)
}

// GetContainerDemo returns a demonstration of container styling
func GetContainerDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true).
			Render("Containers"),
		"",
		ContainerStyle.Width(25).Render(
			"Padded container with border and background styling.",
		),
	)
}

// GetBoxDrawingDemo returns a demonstration of border styles
func GetBoxDrawingDemo() (string, string, string, string) {
	basicBox := BoxStyle.Width(20).Height(5).Render("Basic Box")
	roundedBox := RoundedBoxStyle.Width(20).Height(5).Render("Rounded Box")
	thickBox := ThickBoxStyle.Width(20).Height(5).Render("Thick Border")
	doubleBox := DoubleBoxStyle.Width(20).Height(5).Render("Double Border")
	
	return basicBox, roundedBox, thickBox, doubleBox
}

// GetComplexLayoutDemo returns a demonstration of complex layouts
func GetComplexLayoutDemo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"┌─ Complex Layout ─────────────────┐",
		"│                                  │",
		"│  ╭─ Nested Box ─╮               │",
		"│  │   Content    │  Side Panel   │",
		"│  ╰──────────────╯               │",
		"│                                  │",
		"└──────────────────────────────────┘",
	)
}

// Grid represents a simple grid layout system
type Grid struct {
	Columns int
	Gap     int
	Items   []string
}

// NewGrid creates a new grid layout
func NewGrid(columns, gap int) *Grid {
	return &Grid{
		Columns: columns,
		Gap:     gap,
		Items:   make([]string, 0),
	}
}

// AddItem adds an item to the grid
func (g *Grid) AddItem(item string) {
	g.Items = append(g.Items, item)
}

// Render renders the grid layout
func (g *Grid) Render() string {
	if len(g.Items) == 0 {
		return ""
	}
	
	var rows []string
	for i := 0; i < len(g.Items); i += g.Columns {
		end := i + g.Columns
		if end > len(g.Items) {
			end = len(g.Items)
		}
		
		rowItems := g.Items[i:end]
		row := lipgloss.JoinHorizontal(lipgloss.Top, rowItems...)
		rows = append(rows, row)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}