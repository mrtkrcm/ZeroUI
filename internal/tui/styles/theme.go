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
	BgBase    color.Color
	BgSubtle  color.Color
	BgOverlay color.Color

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
	Base        lipgloss.Style
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Text        lipgloss.Style
	Muted       lipgloss.Style
	Selected    lipgloss.Style
	Success     lipgloss.Style
	Error       lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style
	Border      lipgloss.Style
	BorderFocus lipgloss.Style
	Help        lipgloss.Style

	// Application List styles
	ApplicationList ApplicationListStyles
}

// ApplicationListStyles holds styles for the application list component
type ApplicationListStyles struct {
	Title            lipgloss.Style
	NormalTitle      lipgloss.Style
	SelectedTitle    lipgloss.Style
	NormalDesc       lipgloss.Style
	SelectedDesc     lipgloss.Style
	StatusConfigured lipgloss.Style
	StatusNeedsConfig lipgloss.Style
	StatusError      lipgloss.Style
	StatusUnknown    lipgloss.Style
	Pagination       lipgloss.Style
	Help             lipgloss.Style
	FilterPrompt     lipgloss.Style
	FilterCursor     lipgloss.Style
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

		ApplicationList: ApplicationListStyles{
			Title: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Primary))).
				Bold(true).
				Padding(0, 1),

			NormalTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgBase))).
				Bold(true),

			SelectedTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgSelected))).
				Bold(true),

			NormalDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),

			SelectedDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgSelected))),

			StatusConfigured: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Success))),

			StatusNeedsConfig: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Warning))),

			StatusError: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Error))),

			StatusUnknown: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),

			Pagination: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),

			Help: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.FgMuted))),

			FilterPrompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Primary))),

			FilterCursor: lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorToHex(t.Primary))),
		},
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

// CyberpunkTheme returns a vibrant cyberpunk theme
func CyberpunkTheme() *Theme {
	return &Theme{
		Name:   "Cyberpunk",
		IsDark: true,

		Primary:   parseColor("#FF0080"),
		Secondary: parseColor("#00FFFF"),
		Accent:    parseColor("#FFFF00"),
		Success:   parseColor("#00FF41"),
		Error:     parseColor("#FF073A"),
		Warning:   parseColor("#FFA500"),
		Info:      parseColor("#00D4FF"),

		BgBase:    parseColor("#0D1117"),
		BgSubtle:  parseColor("#161B22"),
		BgOverlay: parseColor("#21262D"),

		FgBase:     parseColor("#00FFFF"),
		FgMuted:    parseColor("#7C3AED"),
		FgSubtle:   parseColor("#58A6FF"),
		FgSelected: parseColor("#FF0080"),

		Border:      parseColor("#30363D"),
		BorderFocus: parseColor("#FF0080"),
	}
}

// OceanTheme returns a calming ocean theme
func OceanTheme() *Theme {
	return &Theme{
		Name:   "Ocean",
		IsDark: false,

		Primary:   parseColor("#006A96"),
		Secondary: parseColor("#52B2CF"),
		Accent:    parseColor("#B8E0D2"),
		Success:   parseColor("#70A288"),
		Error:     parseColor("#D64570"),
		Warning:   parseColor("#EAC435"),
		Info:      parseColor("#345995"),

		BgBase:    parseColor("#F7FBFC"),
		BgSubtle:  parseColor("#E8F4F8"),
		BgOverlay: parseColor("#D6EDF6"),

		FgBase:     parseColor("#003459"),
		FgMuted:    parseColor("#007EA7"),
		FgSubtle:   parseColor("#52B2CF"),
		FgSelected: parseColor("#006A96"),

		Border:      parseColor("#B8E0D2"),
		BorderFocus: parseColor("#006A96"),
	}
}

// SunsetTheme returns a warm sunset theme
func SunsetTheme() *Theme {
	return &Theme{
		Name:   "Sunset",
		IsDark: false,

		Primary:   parseColor("#FF6B35"),
		Secondary: parseColor("#F7931E"),
		Accent:    parseColor("#FFD23F"),
		Success:   parseColor("#6A994E"),
		Error:     parseColor("#BC4749"),
		Warning:   parseColor("#F2CC8F"),
		Info:      parseColor("#81B29A"),

		BgBase:    parseColor("#FFF8F0"),
		BgSubtle:  parseColor("#FFE8D6"),
		BgOverlay: parseColor("#FFD9C4"),

		FgBase:     parseColor("#2F1B14"),
		FgMuted:    parseColor("#A0522D"),
		FgSubtle:   parseColor("#CD853F"),
		FgSelected: parseColor("#FF6B35"),

		Border:      parseColor("#F4A261"),
		BorderFocus: parseColor("#FF6B35"),
	}
}

// AllThemes returns all available themes
func AllThemes() []*Theme {
	return []*Theme{
		DefaultTheme(),
		DarkTheme(),
		CyberpunkTheme(),
		OceanTheme(),
		SunsetTheme(),
	}
}

// Global theme instance
var currentTheme = DefaultTheme()
var currentStyles = currentTheme.BuildStyles()
var currentThemeIndex = 0

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
	
	// Update the current theme index
	themes := AllThemes()
	for i, t := range themes {
		if t.Name == theme.Name {
			currentThemeIndex = i
			break
		}
	}
}

// CycleTheme cycles to the next theme
func CycleTheme() *Theme {
	themes := AllThemes()
	currentThemeIndex = (currentThemeIndex + 1) % len(themes)
	nextTheme := themes[currentThemeIndex]
	
	currentTheme = nextTheme
	currentStyles = nextTheme.BuildStyles()
	
	return nextTheme
}

// GetCurrentThemeName returns the name of the current theme
func GetCurrentThemeName() string {
	return currentTheme.Name
}

// GetThemeNames returns the names of all available themes
func GetThemeNames() []string {
	themes := AllThemes()
	names := make([]string, len(themes))
	for i, theme := range themes {
		names[i] = theme.Name
	}
	return names
}
