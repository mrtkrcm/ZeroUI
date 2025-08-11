package components

import (
	"strings"
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
		Width:         30, // Rectangular: 30 wide (default)
		Height:        10, // Rectangular: 10 tall (default)
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
	// Get advanced card styling for rectangular cards
	cardStyle := m.getAdvancedCardStyle()

	// Build card content with perfect spacing optimized for rectangular shape
	var lines []string

	// Top spacing - less for rectangular cards
	lines = append(lines, "")

	// Loading spinner or logo
	if m.loadingState {
		spinnerLine := lipgloss.NewStyle().
			Width(m.Width - 4).
			Align(lipgloss.Center).
			Render(m.spinner.View())
		lines = append(lines, spinnerLine)
	} else {
		// Enhanced logo with size scaling for rectangular cards
		logoStyle := m.getLogoStyle().
			Bold(true).
			Width(m.Width - 4).
			Align(lipgloss.Center)

		// Scale logo size based on card width (rectangular optimization)
		logoText := m.Status.Definition.Logo
		if m.Width > 32 {
			logoText = logoText + " " + logoText // Double for larger cards
		}

		lines = append(lines, logoStyle.Render(logoText))
	}

	// App name with gradient effect - more prominent for rectangular cards
	nameStyle := m.getNameStyle().
		Bold(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	// Add gradient for selected cards
	if m.Selected {
		nameStyle = nameStyle.Foreground(m.getGradientColor())
	}

	lines = append(lines, nameStyle.Render(m.Status.Definition.Name))

	// Category with improved styling - compact for rectangular cards
	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Italic(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	if m.hoverEffect {
		categoryStyle = categoryStyle.Foreground(lipgloss.Color("250"))
	}

	lines = append(lines, categoryStyle.Render(m.Status.Definition.Category))

	// Status indicators with icons
	statusLine := m.buildEnhancedStatusLine()
	lines = append(lines, statusLine)

	// Bottom spacing - minimal for rectangular cards
	lines = append(lines, "")

	// Join all lines with optimized spacing for rectangular shape
	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

	// Apply advanced card styling with exact rectangular dimensions
	card := cardStyle.
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return card
}

// getAdvancedCardStyle returns enhanced styling with effects and perfect rectangular dimensions
func (m *AppCardModel) getAdvancedCardStyle() lipgloss.Style {
	// Base style with rectangular card optimization
	baseStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true)

	// Remove inconsistent shadow effects that cause layout shifts
	// Instead, use consistent border highlighting for all states

	if !m.Status.IsInstalled {
		// Dimmed style with subtle gradient for uninstalled apps
		return baseStyle.
			BorderForeground(lipgloss.Color("238")).
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("232"))
	}

	if m.Selected {
		// Selected style with gradient background - use consistent border style
		gradientBg := m.getGradientBackground()
		return baseStyle.
			BorderForeground(lipgloss.Color("212")).
			BorderStyle(lipgloss.RoundedBorder()).
			Background(gradientBg).
			Bold(true)
	}

	if m.Focused || m.hoverEffect {
		// Focused/hover style with subtle animation
		return baseStyle.
			BorderForeground(lipgloss.Color("205")).
			Background(lipgloss.Color("234"))
	}

	// Normal style with clean appearance
	return baseStyle.
		BorderForeground(lipgloss.Color("244")).
		Background(lipgloss.Color("233"))
}

// getLogoStyle returns the style for the logo
func (m *AppCardModel) getLogoStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Bold(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	if !m.Status.IsInstalled {
		return style.Foreground(lipgloss.Color("240"))
	}

	return style
}

// getNameStyle returns the style for the app name
func (m *AppCardModel) getNameStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Bold(true).
		Width(m.Width - 4).
		Align(lipgloss.Center)

	if !m.Status.IsInstalled {
		return style.Foreground(lipgloss.Color("242"))
	}

	if m.Selected {
		return style.Foreground(lipgloss.Color("212"))
	}

	return style.Foreground(lipgloss.Color("255"))
}

// buildEnhancedStatusLine creates advanced status indicators with icons
func (m *AppCardModel) buildEnhancedStatusLine() string {
	var indicators []string

	// Enhanced status indicators with better icons and colors
	if m.Status.IsInstalled {
		indicators = append(indicators, "●")
	} else {
		indicators = append(indicators, "○")
	}

	if m.Status.HasConfig {
		indicators = append(indicators, "⚙")
	} else if m.Status.ConfigExists {
		indicators = append(indicators, "📝")
	}

	// Responsive status styling
	statusStyle := lipgloss.NewStyle().
		Width(m.Width - 4).
		Align(lipgloss.Center)

	// Color coding for status
	if m.Status.IsInstalled {
		statusStyle = statusStyle.Foreground(lipgloss.Color("76")) // Bright green
		if m.Selected {
			statusStyle = statusStyle.Foreground(lipgloss.Color("82")) // Even brighter when selected
		}
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("240"))
	}

	return statusStyle.Render(strings.Join(indicators, "  "))
}

// SetSelected sets the selected state
func (m *AppCardModel) SetSelected(selected bool) {
	m.Selected = selected
}

// SetFocused sets the focused state
func (m *AppCardModel) SetFocused(focused bool) {
	m.Focused = focused
}

// SetSize sets the card dimensions as rectangular (width x height)
func (m *AppCardModel) SetSize(width, height int) {
	// Use provided dimensions directly for rectangular cards
	m.Width = width
	m.Height = height

	// Ensure minimum viable size
	if m.Width < 24 {
		m.Width = 24
	}
	if m.Height < 8 {
		m.Height = 8
	}

	// Invalidate cache when size changes
	m.invalidateCache()
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
