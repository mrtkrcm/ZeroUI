package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// AvailableThemes holds all predefined themes
var AvailableThemes = []Theme{
	ModernTheme,
	DraculaTheme,
	LightTheme,
	NordTheme,
	CatppuccinTheme,
}

// currentThemeIndex tracks the currently active theme
var currentThemeIndex = 0

// Styles holds all the style definitions for the UI
type Styles struct {
	Base     lipgloss.Style
	Header   lipgloss.Style
	Footer   lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Focused  lipgloss.Style
	Disabled lipgloss.Style
	Success  lipgloss.Style
	Error    lipgloss.Style
	Warning  lipgloss.Style
	Info     lipgloss.Style

	// Additional fields for backward compatibility
	Title           lipgloss.Style
	Help            lipgloss.Style
	Muted           lipgloss.Style
	Text            lipgloss.Style
	Subtitle        lipgloss.Style
	ApplicationList ApplicationListStyles
}

// ApplicationListStyles holds styles specific to application lists
type ApplicationListStyles struct {
	Base              lipgloss.Style
	Title             lipgloss.Style
	Pagination        lipgloss.Style
	StatusConfigured  lipgloss.Style
	StatusNeedsConfig lipgloss.Style
	StatusError       lipgloss.Style
	StatusUnknown     lipgloss.Style
	SelectedTitle     lipgloss.Style
	NormalTitle       lipgloss.Style
	SelectedDesc      lipgloss.Style
	NormalDesc        lipgloss.Style
	Help              lipgloss.Style
	FilterPrompt      lipgloss.Style
	FilterCursor      lipgloss.Style
}

// Theme defines the visual design system for ZeroUI
type Theme struct {
	// Base colors
	Background string
	Foreground string
	Accent     string
	Secondary  string

	// Status colors
	Success string
	Warning string
	Error   string
	Info    string

	// UI element colors
	Border       string
	BorderFocus  string
	Surface      string
	SurfaceHover string

	// Text colors
	TextPrimary   string
	TextSecondary string
	TextMuted     string
	TextDisabled  string

	// Interactive colors
	Selection string
	Highlight string
	Link      string
	LinkHover string

	// Additional fields for backward compatibility
	BgSubtle string
	FgMuted  string

	// Test compatibility fields
	Name    string
	Primary *lipgloss.Color
	BgBase  *lipgloss.Color
	FgBase  *lipgloss.Color
}

// ModernTheme provides a beautiful, accessible color scheme
var ModernTheme = Theme{
	Background: "#0c0c0c", // Deep dark
	Foreground: "#f8f8f2", // Off-white
	Accent:     "#bd93f9", // Purple
	Secondary:  "#6272a4", // Blue-gray

	Success: "#50fa7b", // Green
	Warning: "#f1fa8c", // Yellow
	Error:   "#ff5555", // Red
	Info:    "#8be9fd", // Cyan

	Border:       "#44475a", // Dark gray
	BorderFocus:  "#bd93f9", // Purple
	Surface:      "#1e1e2e", // Dark blue-gray
	SurfaceHover: "#282a36", // Slightly lighter

	TextPrimary:   "#f8f8f2", // Off-white
	TextSecondary: "#6272a4", // Blue-gray
	TextMuted:     "#6272a4", // Blue-gray
	TextDisabled:  "#44475a", // Dark gray

	Selection: "#bd93f9", // Purple
	Highlight: "#ffb86c", // Orange
	Link:      "#8be9fd", // Cyan
	LinkHover: "#50fa7b", // Green

	// Backward compatibility fields
	BgSubtle: "#1e1e2e",
	FgMuted:  "#6272a4",

	// Test compatibility fields
	Name:    "Modern",
	Primary: func() *lipgloss.Color { c := lipgloss.Color("#bd93f9"); return &c }(),
	BgBase:  func() *lipgloss.Color { c := lipgloss.Color("#0c0c0c"); return &c }(),
	FgBase:  func() *lipgloss.Color { c := lipgloss.Color("#f8f8f2"); return &c }(),
}

// DraculaTheme - Alternative theme for variety
var DraculaTheme = Theme{
	Background: "#282a36",
	Foreground: "#f8f8f2",
	Accent:     "#bd93f9",
	Secondary:  "#6272a4",

	Success: "#50fa7b",
	Warning: "#f1fa8c",
	Error:   "#ff5555",
	Info:    "#8be9fd",

	Border:       "#44475a",
	BorderFocus:  "#bd93f9",
	Surface:      "#44475a",
	SurfaceHover: "#6272a4",

	TextPrimary:   "#f8f8f2",
	TextSecondary: "#6272a4",
	TextMuted:     "#6272a4",
	TextDisabled:  "#6272a4",

	Selection: "#bd93f9",
	Highlight: "#ffb86c",
	Link:      "#8be9fd",
	LinkHover: "#50fa7b",

	// Backward compatibility fields
	BgSubtle: "#44475a",
	FgMuted:  "#6272a4",

	// Test compatibility fields
	Name:    "Dracula",
	Primary: func() *lipgloss.Color { c := lipgloss.Color("#bd93f9"); return &c }(),
	BgBase:  func() *lipgloss.Color { c := lipgloss.Color("#282a36"); return &c }(),
	FgBase:  func() *lipgloss.Color { c := lipgloss.Color("#f8f8f2"); return &c }(),
}

// LightTheme - GitHub-inspired light theme for daylight usage
var LightTheme = Theme{
	Background: "#ffffff", // Pure white
	Foreground: "#24292e", // Dark gray
	Accent:     "#0366d6", // GitHub blue
	Secondary:  "#6a737d", // Gray

	Success: "#22863a", // Green
	Warning: "#b08800", // Dark yellow
	Error:   "#cb2431", // Red
	Info:    "#0366d6", // Blue

	Border:       "#e1e4e8", // Light gray
	BorderFocus:  "#0366d6", // Blue
	Surface:      "#f6f8fa", // Very light gray
	SurfaceHover: "#e1e4e8", // Light gray

	TextPrimary:   "#24292e", // Dark gray
	TextSecondary: "#586069", // Medium gray
	TextMuted:     "#6a737d", // Light gray
	TextDisabled:  "#959da5", // Very light gray

	Selection: "#0366d6", // Blue
	Highlight: "#ffd33d", // Yellow
	Link:      "#0366d6", // Blue
	LinkHover: "#22863a", // Green

	// Backward compatibility fields
	BgSubtle: "#f6f8fa",
	FgMuted:  "#6a737d",

	// Test compatibility fields
	Name:    "Light",
	Primary: func() *lipgloss.Color { c := lipgloss.Color("#0366d6"); return &c }(),
	BgBase:  func() *lipgloss.Color { c := lipgloss.Color("#ffffff"); return &c }(),
	FgBase:  func() *lipgloss.Color { c := lipgloss.Color("#24292e"); return &c }(),
}

// NordTheme - Arctic, north-bluish color palette
var NordTheme = Theme{
	Background: "#2e3440", // Polar Night
	Foreground: "#eceff4", // Snow Storm
	Accent:     "#88c0d0", // Frost
	Secondary:  "#81a1c1", // Frost blue

	Success: "#a3be8c", // Aurora green
	Warning: "#ebcb8b", // Aurora yellow
	Error:   "#bf616a", // Aurora red
	Info:    "#88c0d0", // Frost cyan

	Border:       "#4c566a", // Polar Night light
	BorderFocus:  "#88c0d0", // Frost
	Surface:      "#3b4252", // Polar Night medium
	SurfaceHover: "#434c5e", // Polar Night lighter

	TextPrimary:   "#eceff4", // Snow Storm bright
	TextSecondary: "#d8dee9", // Snow Storm medium
	TextMuted:     "#81a1c1", // Frost blue
	TextDisabled:  "#4c566a", // Polar Night light

	Selection: "#88c0d0", // Frost
	Highlight: "#ebcb8b", // Aurora yellow
	Link:      "#88c0d0", // Frost
	LinkHover: "#a3be8c", // Aurora green

	// Backward compatibility fields
	BgSubtle: "#3b4252",
	FgMuted:  "#81a1c1",

	// Test compatibility fields
	Name:    "Nord",
	Primary: func() *lipgloss.Color { c := lipgloss.Color("#88c0d0"); return &c }(),
	BgBase:  func() *lipgloss.Color { c := lipgloss.Color("#2e3440"); return &c }(),
	FgBase:  func() *lipgloss.Color { c := lipgloss.Color("#eceff4"); return &c }(),
}

// CatppuccinTheme - Soothing pastel theme (Mocha variant)
var CatppuccinTheme = Theme{
	Background: "#1e1e2e", // Base
	Foreground: "#cdd6f4", // Text
	Accent:     "#cba6f7", // Mauve
	Secondary:  "#9399b2", // Overlay1

	Success: "#a6e3a1", // Green
	Warning: "#f9e2af", // Yellow
	Error:   "#f38ba8", // Red
	Info:    "#89dceb", // Sky

	Border:       "#45475a", // Surface1
	BorderFocus:  "#cba6f7", // Mauve
	Surface:      "#313244", // Surface0
	SurfaceHover: "#45475a", // Surface1

	TextPrimary:   "#cdd6f4", // Text
	TextSecondary: "#a6adc8", // Subtext0
	TextMuted:     "#9399b2", // Overlay1
	TextDisabled:  "#6c7086", // Overlay0

	Selection: "#cba6f7", // Mauve
	Highlight: "#f9e2af", // Yellow
	Link:      "#89b4fa", // Blue
	LinkHover: "#a6e3a1", // Green

	// Backward compatibility fields
	BgSubtle: "#313244",
	FgMuted:  "#9399b2",

	// Test compatibility fields
	Name:    "Catppuccin",
	Primary: func() *lipgloss.Color { c := lipgloss.Color("#cba6f7"); return &c }(),
	BgBase:  func() *lipgloss.Color { c := lipgloss.Color("#1e1e2e"); return &c }(),
	FgBase:  func() *lipgloss.Color { c := lipgloss.Color("#cdd6f4"); return &c }(),
}

// GetTheme returns the current theme (could be configurable)
func GetTheme() Theme {
	return AvailableThemes[currentThemeIndex]
}

// AllThemes returns all available themes
func AllThemes() []Theme {
	return AvailableThemes
}

// GetThemeNames returns the names of all available themes
func GetThemeNames() []string {
	names := make([]string, len(AvailableThemes))
	for i, theme := range AvailableThemes {
		names[i] = GetThemeName(theme)
	}
	return names
}

// GetThemeName returns a human-readable name for a theme
func GetThemeName(theme Theme) string {
	// Use the Name field if set
	if theme.Name != "" {
		return theme.Name
	}
	// Fallback to background color matching
	switch theme.Background {
	case ModernTheme.Background:
		return "Modern"
	case DraculaTheme.Background:
		return "Dracula"
	case LightTheme.Background:
		return "Light"
	case NordTheme.Background:
		return "Nord"
	case CatppuccinTheme.Background:
		return "Catppuccin"
	}
	return "Unknown"
}

// SetTheme sets the current theme
func SetTheme(theme Theme) {
	for i, t := range AvailableThemes {
		if GetThemeName(t) == GetThemeName(theme) {
			currentThemeIndex = i
			break
		}
	}
}

// CycleTheme cycles to the next theme and returns it
func CycleTheme() Theme {
	currentThemeIndex = (currentThemeIndex + 1) % len(AvailableThemes)
	return GetTheme()
}

// Style definitions for common UI elements
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Foreground)).
			Background(lipgloss.Color(ModernTheme.Background))

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ModernTheme.Accent)).
			Background(lipgloss.Color(ModernTheme.Surface)).
			Padding(0, 2).
			MarginBottom(1)

	HeaderTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ModernTheme.TextPrimary)).
				Background(lipgloss.Color(ModernTheme.Surface))

	HeaderSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.TextSecondary)).
				Background(lipgloss.Color(ModernTheme.Surface))

	// Content styles
	ContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.TextPrimary)).
			Padding(0, 1)

	// Selection styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Background)).
			Background(lipgloss.Color(ModernTheme.Selection)).
			Bold(true)

	SelectedEditingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Background)).
				Background(lipgloss.Color(ModernTheme.Highlight)).
				Bold(true)

	// Field styles
	FieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.TextPrimary))

	FieldChangedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Success)).
				Bold(true)

	FieldErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Error))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Success)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Warning))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Error)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Info))

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ModernTheme.Border)).
			Padding(0, 1)

	BorderFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ModernTheme.BorderFocus)).
				Padding(0, 1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Background)).
			Background(lipgloss.Color(ModernTheme.Accent)).
			Padding(0, 2).
			Margin(0, 1)

	ButtonHoverStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Background)).
				Background(lipgloss.Color(ModernTheme.Highlight)).
				Padding(0, 2).
				Margin(0, 1)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.TextSecondary)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ModernTheme.Border)).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.Accent)).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ModernTheme.TextPrimary))

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Background)).
				Background(lipgloss.Color(ModernTheme.Accent))

	ProgressFillStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Accent)).
				Background(lipgloss.Color(ModernTheme.Background))

	// Notification styles
	NotificationSuccessStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ModernTheme.Success)).
					Background(lipgloss.Color(ModernTheme.Surface)).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color(ModernTheme.Success)).
					Padding(0, 1)

	NotificationErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Error)).
				Background(lipgloss.Color(ModernTheme.Surface)).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ModernTheme.Error)).
				Padding(0, 1)

	NotificationInfoStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ModernTheme.Info)).
				Background(lipgloss.Color(ModernTheme.Surface)).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ModernTheme.Info)).
				Padding(0, 1)
)

// Helper functions for dynamic styling
func GetFieldStyle(isSelected, isEditing, hasChanged bool, hasError bool) lipgloss.Style {
	if hasError {
		return FieldErrorStyle
	}
	if isSelected && isEditing {
		return SelectedEditingStyle
	}
	if isSelected {
		return SelectedStyle
	}
	if hasChanged {
		return FieldChangedStyle
	}
	return FieldStyle
}

func GetNotificationStyle(notificationType string) lipgloss.Style {
	switch notificationType {
	case "success":
		return NotificationSuccessStyle
	case "error":
		return NotificationErrorStyle
	case "info", "warning":
		return NotificationInfoStyle
	default:
		return NotificationInfoStyle
	}
}

func GetStatusIcon(status string) string {
	switch status {
	case "success":
		return "✅"
	case "error":
		return "❌"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	case "loading":
		return "⏳"
	case "saving":
		return "💾"
	default:
		return "•"
	}
}

// GetStyles returns the default style definitions
func GetStyles() *Styles {
	theme := GetTheme()

	return &Styles{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextPrimary)).
			Background(lipgloss.Color(theme.Background)),

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Accent)).
			Background(lipgloss.Color(theme.Surface)).
			Padding(0, 2).
			MarginBottom(1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSecondary)).
			Background(lipgloss.Color(theme.Surface)).
			Padding(0, 1),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(0, 1),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Background)).
			Background(lipgloss.Color(theme.Selection)).
			Bold(true),

		Focused: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Background)).
			Background(lipgloss.Color(theme.Accent)).
			Bold(true),

		Disabled: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextDisabled)),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Success)).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Error)).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Warning)),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Info)),

		// Backward compatibility styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Accent)),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSecondary)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Border)).
			Padding(0, 1),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextMuted)),

		Text: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextPrimary)),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSecondary)).
			Italic(true),

		ApplicationList: ApplicationListStyles{
			Base: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextPrimary)).
				Background(lipgloss.Color(theme.Background)),

			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(theme.Accent)),

			Pagination: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextSecondary)),

			StatusConfigured: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Success)).
				Bold(true),

			StatusNeedsConfig: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Warning)),

			StatusError: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Error)).
				Bold(true),

			StatusUnknown: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextMuted)),

			SelectedTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Background)).
				Background(lipgloss.Color(theme.Selection)).
				Bold(true),

			NormalTitle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextPrimary)),

			SelectedDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Background)).
				Background(lipgloss.Color(theme.Selection)),

			NormalDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextSecondary)),

			Help: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextSecondary)).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(theme.Border)).
				Padding(0, 1),

			FilterPrompt: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true),

			FilterCursor: lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Highlight)),
		},
	}
}

// DefaultTheme provides the default theme (alias for backward compatibility)
var DefaultTheme = ModernTheme

// Color utility functions
func HexToRGB(hex string) (int, int, int) {
	// Simple hex to RGB conversion (could be enhanced)
	if len(hex) != 7 || hex[0] != '#' {
		return 255, 255, 255 // Default to white
	}

	// This is a simplified version - in production you'd want proper hex parsing
	return 255, 255, 255
}

func IsLightColor(hex string) bool {
	r, g, b := HexToRGB(hex)
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
	return luminance > 0.5
}

// ColorToHex converts a lipgloss.Color to hex (for backward compatibility)
func ColorToHex(color lipgloss.Color) string {
	// This is a simplified implementation
	// In a real implementation, you'd need proper color conversion
	return "#ffffff"
}

// GetCurrentThemeName returns the name of the current theme
func GetCurrentThemeName() string {
	return GetThemeName(GetTheme())
}

// buildThemeRegistry creates a map of theme names to themes for quick lookup
func buildThemeRegistry() map[string]Theme {
	registry := make(map[string]Theme)
	for _, theme := range AvailableThemes {
		name := GetThemeName(theme)
		registry[name] = theme
	}
	return registry
}

// SetThemeByName sets the current theme by name (case-insensitive)
// Returns the theme and true if found, zero theme and false otherwise
func SetThemeByName(name string) (Theme, bool) {
	registry := buildThemeRegistry()

	// Try exact match first
	if theme, ok := registry[name]; ok {
		SetTheme(theme)
		return theme, true
	}

	// Try case-insensitive match
	for registeredName, theme := range registry {
		if equalsCaseInsensitive(registeredName, name) {
			SetTheme(theme)
			return theme, true
		}
	}

	// Theme not found
	return Theme{}, false
}

// ListAvailableThemes returns all registered theme names
func ListAvailableThemes() []string {
	return GetThemeNames()
}

// equalsCaseInsensitive compares two strings case-insensitively
func equalsCaseInsensitive(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		// Convert to lowercase for comparison
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// BuildStyles builds the complete style set from the theme (for backward compatibility)
func (t Theme) BuildStyles() *Styles {
	return GetStyles()
}
