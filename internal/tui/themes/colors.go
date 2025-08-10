package themes

import "github.com/charmbracelet/lipgloss"

// ColorTheme defines the interface for color themes
type ColorTheme interface {
	Primary() lipgloss.Style
	Secondary() lipgloss.Style
	Accent() lipgloss.Style
	Success() lipgloss.Style
	Warning() lipgloss.Style
	Error() lipgloss.Style
	Info() lipgloss.Style
	Light() lipgloss.Style
	Medium() lipgloss.Style
	Dark() lipgloss.Style
	Background() lipgloss.Style
}

// DefaultColorTheme implements the default ZeroUI color scheme
type DefaultColorTheme struct{}

// NewDefaultColorTheme creates a new default color theme
func NewDefaultColorTheme() *DefaultColorTheme {
	return &DefaultColorTheme{}
}

// Color palette styles - primary colors
func (t *DefaultColorTheme) Primary() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)
}

func (t *DefaultColorTheme) Secondary() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B9D")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)
}

func (t *DefaultColorTheme) Accent() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#C9A96E")).
		Foreground(lipgloss.Color("#282A36")).
		Padding(0, 1)
}

// Status colors
func (t *DefaultColorTheme) Success() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))
}

func (t *DefaultColorTheme) Warning() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB86C"))
}

func (t *DefaultColorTheme) Error() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))
}

func (t *DefaultColorTheme) Info() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD"))
}

// Neutral colors
func (t *DefaultColorTheme) Light() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#F8F8F2")).
		Foreground(lipgloss.Color("#282A36")).
		Padding(0, 1)
}

func (t *DefaultColorTheme) Medium() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#6272A4")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)
}

func (t *DefaultColorTheme) Dark() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#282A36")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)
}

func (t *DefaultColorTheme) Background() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(0, 1)
}

// Color constants for direct use
const (
	PrimaryColor    = "#7D56F4"  // Purple
	SecondaryColor  = "#FF6B9D"  // Pink
	AccentColor     = "#C9A96E"  // Gold
	SuccessColor    = "#50FA7B"  // Green
	WarningColor    = "#FFB86C"  // Orange
	ErrorColor      = "#FF5555"  // Red
	InfoColor       = "#8BE9FD"  // Cyan
	BackgroundColor = "#282A36"  // Dark Gray
	ForegroundColor = "#F8F8F2"  // Light Gray
	BorderColor     = "#44475A"  // Medium Gray
	CaptionColor    = "#6272A4"  // Blue Gray
)

// Validation state colors
var (
	ValidStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(SuccessColor))

	InvalidStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ErrorColor))

	PendingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(WarningColor))
)

// GetColorInfo returns detailed color information for display
func GetColorInfo() string {
	return `Color Definitions (Lipgloss):
• Primary:    #7D56F4  (Purple)
• Secondary:  #FF6B9D  (Pink)  
• Accent:     #C9A96E  (Gold)
• Success:    #50FA7B  (Green)
• Warning:    #FFB86C  (Orange)
• Error:      #FF5555  (Red)
• Info:       #8BE9FD  (Cyan)
• Background: #282A36  (Dark Gray)
• Foreground: #F8F8F2  (Light Gray)`
}