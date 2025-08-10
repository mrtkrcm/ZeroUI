package styles

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/lipgloss"
)

// Theme holds color scheme and styling information
type Theme struct {
	Name   string
	IsDark bool

	// Core colors
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Success   color.Color
	Error     color.Color
	Warning   color.Color
	Info      color.Color

	// Background colors
	BgBase     color.Color
	BgSubtle   color.Color
	BgOverlay  color.Color

	// Foreground colors
	FgBase     color.Color
	FgMuted    color.Color
	FgSubtle   color.Color
	FgSelected color.Color

	// Border colors
	Border      color.Color
	BorderFocus color.Color
}

// DefaultTheme returns the default theme
func DefaultTheme() *Theme {
	return &Theme{
		Name:   "Default",
		IsDark: false,

		Primary:   parseColor("#7D56F4"),
		Secondary: parseColor("#626262"),
		Accent:    parseColor("#FF8C94"),
		Success:   parseColor("#90EE90"),
		Error:     parseColor("#FF6B6B"),
		Warning:   parseColor("#FFD93D"),
		Info:      parseColor("#6BB6FF"),

		BgBase:    parseColor("#FFFFFF"),
		BgSubtle:  parseColor("#F8F8F8"),
		BgOverlay: parseColor("#F0F0F0"),

		FgBase:     parseColor("#000000"),
		FgMuted:    parseColor("#626262"),
		FgSubtle:   parseColor("#A0A0A0"),
		FgSelected: parseColor("#7D56F4"),

		Border:      parseColor("#D9D9D9"),
		BorderFocus: parseColor("#7D56F4"),
	}
}

// DarkTheme returns a dark theme
func DarkTheme() *Theme {
	return &Theme{
		Name:   "Dark",
		IsDark: true,

		Primary:   parseColor("#BB9AF7"),
		Secondary: parseColor("#9CA3AF"),
		Accent:    parseColor("#F7768E"),
		Success:   parseColor("#9ECE6A"),
		Error:     parseColor("#F7768E"),
		Warning:   parseColor("#E0AF68"),
		Info:      parseColor("#7AA2F7"),

		BgBase:    parseColor("#1A1B26"),
		BgSubtle:  parseColor("#24283B"),
		BgOverlay: parseColor("#2F3549"),

		FgBase:     parseColor("#C0CAF5"),
		FgMuted:    parseColor("#9CA3AF"),
		FgSubtle:   parseColor("#6B7280"),
		FgSelected: parseColor("#BB9AF7"),

		Border:      parseColor("#414558"),
		BorderFocus: parseColor("#BB9AF7"),
	}
}

// parseColor converts hex string to color.Color
func parseColor(hex string) color.Color {
	c := lipgloss.Color(hex)
	return c
}

// Styles holds pre-configured lipgloss styles
type Styles struct {
	Base         lipgloss.Style
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Text         lipgloss.Style
	Muted        lipgloss.Style
	Selected     lipgloss.Style
	Success      lipgloss.Style
	Error        lipgloss.Style
	Warning      lipgloss.Style
	Info         lipgloss.Style
	Border       lipgloss.Style
	BorderFocus  lipgloss.Style
	Help         lipgloss.Style
}

// BuildStyles creates lipgloss styles from theme
func (t *Theme) BuildStyles() *Styles {
	return &Styles{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.FgBase))).
			Background(lipgloss.Color(ColorToHex(t.BgBase))),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Primary))).
			Bold(true).
			Padding(0, 1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Secondary))).
			Bold(true),

		Text: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.FgBase))),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.FgSelected))).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Success))).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Error))).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Warning))).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.Info))).
			Bold(true),

		Border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorToHex(t.Border))),

		BorderFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorToHex(t.BorderFocus))),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),
	}
}

// ColorToHex converts color.Color to hex string
func ColorToHex(c color.Color) string {
	if c == nil {
		return "#000000"
	}
	
	// If it's already a lipgloss.Color, return its value
	if lc, ok := c.(lipgloss.Color); ok {
		return string(lc)
	}
	
	// Convert from color.Color to hex
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// Global theme instance
var currentTheme = DefaultTheme()
var currentStyles = currentTheme.BuildStyles()

// GetTheme returns the current theme
func GetTheme() *Theme {
	return currentTheme
}

// GetStyles returns the current styles
func GetStyles() *Styles {
	return currentStyles
}

// SetTheme sets the current theme
func SetTheme(theme *Theme) {
	currentTheme = theme
	currentStyles = theme.BuildStyles()
}