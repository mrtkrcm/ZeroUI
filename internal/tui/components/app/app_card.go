package appcomponents

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// AppCardModel represents a single application card with advanced rendering
type AppCardModel struct {
	Status   registry.AppStatus
	Selected bool
	Focused  bool
	Width    int
	Height   int
	styles   *styles.Styles

	// Render caching for performance
	cachedView    string
	lastCacheTime time.Time
	cacheDuration time.Duration

	// Animation state
	spinner      spinner.Model
	loadingState bool

	// Visual effects
	hoverEffect  bool
	gradientBase lipgloss.Color
}

// NewAppCard creates a new app card component with advanced features
func NewAppCard(status registry.AppStatus) *AppCardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &AppCardModel{
		Status:        status,
		Width:         15, // Compact: 15 wide (half the original size)
		Height:        5,  // Compact: 5 tall (half the original size)
		styles:        styles.GetStyles(),
		cacheDuration: 100 * time.Millisecond, // Cache for 100ms for 60fps
		spinner:       s,
		gradientBase:  lipgloss.Color("#7D56F4"),
	}
}

// Init initializes the app card with spinner
func (m *AppCardModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the app card with performance optimization
func (m *AppCardModel) Update(msg tea.Msg) (*AppCardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Force re-cache on size change
		m.invalidateCache()
		return m, nil

	case tea.KeyMsg:
		if m.Selected && key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
			m.loadingState = true
			m.invalidateCache()
			return m, tea.Batch(SelectAppCmd(m.Status.Definition.Name), m.spinner.Tick)
		}

	case spinner.TickMsg:
		if m.loadingState {
			m.spinner, cmd = m.spinner.Update(msg)
			m.invalidateCache() // Invalidate cache for animation
			return m, cmd
		}

		// Note: Mouse events would be handled here with proper tea.MouseMsg
		// Currently disabled due to API changes
	}

	return m, cmd
}

// View renders the app card with caching and advanced effects
func (m *AppCardModel) View() string {
	// Check cache first for performance (60fps optimization)
	if m.isCacheValid() {
		return m.cachedView
	}

	// Render the card
	renderedCard := m.renderCard()

	// Update cache
	m.cachedView = renderedCard
	m.lastCacheTime = time.Now()

	return renderedCard
}

// renderCard performs the actual rendering with advanced effects
func (m *AppCardModel) renderCard() string {
	// Get advanced card styling for compact cards
	cardStyle := m.getAdvancedCardStyle()

	// Build compact card content with focus on app name only
	var lines []string

	// Loading spinner or app name
	if m.loadingState {
		spinnerLine := lipgloss.NewStyle().
			Width(m.Width - 2).
			Align(lipgloss.Center).
			Render(m.spinner.View())
		lines = append(lines, spinnerLine)
	} else {
		// App name as the primary and only visual element
		nameStyle := m.getCompactNameStyle()
		appName := m.getAdaptiveAppName()
		lines = append(lines, nameStyle.Render(appName))
	}

	// Minimal status indicator (single line)
	if !m.loadingState {
		statusLine := m.buildCompactStatusLine()
		lines = append(lines, statusLine)
	}

	// Join all lines with minimal spacing for compact design
	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

	// Apply compact card styling with exact dimensions and force size
	card := cardStyle.
		Width(m.Width).
		Height(m.Height).
		MaxWidth(m.Width).
		MaxHeight(m.Height).
		Render(content)

	return card
}

// getAdvancedCardStyle returns enhanced styling optimized for compact cards
func (m *AppCardModel) getAdvancedCardStyle() lipgloss.Style {
	// Base style with compact card optimization - consistent sizing
	baseStyle := lipgloss.NewStyle().
		Padding(1, 1). // Consistent padding for stable layout
		Margin(0).     // No margin to prevent shifts
		Border(lipgloss.RoundedBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	if !m.Status.IsInstalled {
		// Dimmed style for uninstalled apps
		return baseStyle.
			BorderForeground(lipgloss.Color("238")).
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("232"))
	}

	if m.Selected {
		// Selected style with bright border and background
		return baseStyle.
			BorderForeground(lipgloss.Color("212")).
			Background(lipgloss.Color("235")).
			Bold(true)
	}

	if m.Focused || m.hoverEffect {
		// Focused/hover style
		return baseStyle.
			BorderForeground(lipgloss.Color("205")).
			Background(lipgloss.Color("234"))
	}

	// Normal style with clean appearance
	return baseStyle.
		BorderForeground(lipgloss.Color("244")).
		Background(lipgloss.Color("233"))
}

// SetSelected sets the selected state
func (m *AppCardModel) SetSelected(selected bool) {
	m.Selected = selected
}

// SetFocused sets the focused state
func (m *AppCardModel) SetFocused(focused bool) {
	m.Focused = focused
}

// SetSize sets the card dimensions as compact (width x height)
func (m *AppCardModel) SetSize(width, height int) {
	// Use provided dimensions directly for compact cards
	m.Width = width
	m.Height = height

	// Ensure minimum viable size for compact design
	if m.Width < 12 {
		m.Width = 12
	}
	if m.Height < 4 {
		m.Height = 4
	}

	// Invalidate cache when size changes
	m.invalidateCache()
}

// Compact design methods

// getCompactNameStyle returns optimized styling for app names in compact cards
func (m *AppCardModel) getCompactNameStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Bold(true).
		Width(m.Width - 2).
		Align(lipgloss.Center)

	if !m.Status.IsInstalled {
		return style.Foreground(lipgloss.Color("240"))
	}

	if m.Selected {
		return style.Foreground(lipgloss.Color("212"))
	}

	return style.Foreground(lipgloss.Color("255"))
}

// getAdaptiveAppName returns app name adapted programmatically for different lengths
func (m *AppCardModel) getAdaptiveAppName() string {
	name := m.Status.Definition.Name
	maxWidth := m.Width - 2 // Account for padding

	// If name fits, return as-is
	if len(name) <= maxWidth {
		return name
	}

	// Programmatic adaptation strategies
	adaptedName := m.adaptNameProgrammatically(name, maxWidth)

	// If still too long, truncate with ellipsis
	if len(adaptedName) > maxWidth && maxWidth > 1 {
		if maxWidth <= 3 {
			return adaptedName[:maxWidth] // No ellipsis for very small spaces
		}
		return adaptedName[:maxWidth-1] + "…"
	}

	return adaptedName
}

// adaptNameProgrammatically applies intelligent shortening strategies
func (m *AppCardModel) adaptNameProgrammatically(name string, maxWidth int) string {
	// Strategy 1: Remove common suffixes
	if words := m.splitIntoWords(name); len(words) > 1 {
		shortened := m.tryRemoveSuffixes(words, maxWidth)
		if len(shortened) <= maxWidth {
			return shortened
		}
	}

	// Strategy 2: Create acronym for multi-word names
	if words := m.splitIntoWords(name); len(words) > 1 {
		acronym := m.createAcronym(words)
		if len(acronym) <= maxWidth {
			return acronym
		}
	}

	// Strategy 3: Remove vowels (except first character)
	shortened := m.removeVowels(name)
	if len(shortened) <= maxWidth {
		return shortened
	}

	// Strategy 4: Truncate to fit
	if maxWidth > 0 {
		return name[:maxWidth]
	}

	return name
}

// splitIntoWords splits a name into words, handling different separators
func (m *AppCardModel) splitIntoWords(name string) []string {
	var words []string
	var current []rune

	for _, r := range name {
		if r == ' ' || r == '-' || r == '_' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		} else {
			current = append(current, r)
		}
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}

// tryRemoveSuffixes attempts to remove common suffixes to shorten names
func (m *AppCardModel) tryRemoveSuffixes(words []string, maxWidth int) string {
	// Common suffixes to try removing
	suffixes := []string{"Code", "Editor", "Terminal", "IDE", "App"}

	for _, suffix := range suffixes {
		if len(words) > 1 && words[len(words)-1] == suffix {
			candidate := ""
			for i, word := range words[:len(words)-1] {
				if i > 0 {
					candidate += " "
				}
				candidate += word
			}
			if len(candidate) <= maxWidth {
				return candidate
			}
		}
	}

	// Return original joined words if no suffix removal helped
	result := ""
	for i, word := range words {
		if i > 0 {
			result += " "
		}
		result += word
	}
	return result
}

// createAcronym creates an acronym from multiple words
func (m *AppCardModel) createAcronym(words []string) string {
	if len(words) <= 1 {
		return ""
	}

	acronym := ""
	for _, word := range words {
		if len(word) > 0 {
			acronym += string(word[0])
		}
	}
	return acronym
}

// removeVowels removes vowels except from the first character
func (m *AppCardModel) removeVowels(name string) string {
	if len(name) <= 1 {
		return name
	}

	vowels := "aeiouAEIOU"
	result := []rune{rune(name[0])} // Keep first character

	for _, r := range name[1:] {
		isVowel := false
		for _, v := range vowels {
			if r == v {
				isVowel = true
				break
			}
		}
		if !isVowel {
			result = append(result, r)
		}
	}

	return string(result)
}

// buildCompactStatusLine creates a minimal status indicator
func (m *AppCardModel) buildCompactStatusLine() string {
	// Single character status indicator
	var status string
	if m.Status.IsInstalled {
		status = "●" // Installed
	} else {
		status = "○" // Not installed
	}

	// Style the status
	statusStyle := lipgloss.NewStyle().
		Width(m.Width - 2).
		Align(lipgloss.Center)

	if m.Status.IsInstalled {
		statusStyle = statusStyle.Foreground(lipgloss.Color("76")) // Green
		if m.Selected {
			statusStyle = statusStyle.Foreground(lipgloss.Color("82")) // Brighter green
		}
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("240")) // Dim
	}

	return statusStyle.Render(status)
}

// Performance and visual enhancement methods

// isCacheValid checks if the cached view is still valid
func (m *AppCardModel) isCacheValid() bool {
	if m.cachedView == "" {
		return false
	}

	return time.Since(m.lastCacheTime) < m.cacheDuration
}

// invalidateCache forces a re-render on next View() call
func (m *AppCardModel) invalidateCache() {
	m.cachedView = ""
	m.lastCacheTime = time.Time{}
}

// getGradientColor returns a color based on selection state
func (m *AppCardModel) getGradientColor() lipgloss.Color {
	if m.hoverEffect {
		return lipgloss.Color("#9A7FFC") // Lighter gradient on hover
	}
	return m.gradientBase
}

// getGradientBackground creates a gradient background effect
func (m *AppCardModel) getGradientBackground() lipgloss.Color {
	if m.hoverEffect {
		return lipgloss.Color("236") // Slightly lighter when hovering
	}
	return lipgloss.Color("235")
}

// SetLoadingState sets the loading animation state
func (m *AppCardModel) SetLoadingState(loading bool) {
	if m.loadingState != loading {
		m.loadingState = loading
		m.invalidateCache()
	}
}

// IsLoading returns the current loading state
func (m *AppCardModel) IsLoading() bool {
	return m.loadingState
}

// SelectAppCmd creates a command to select an app
func SelectAppCmd(appName string) tea.Cmd {
	return func() tea.Msg {
		return AppSelectedMsg{App: appName}
	}
}
