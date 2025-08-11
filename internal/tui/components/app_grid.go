package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/performance"
	"github.com/mrtkrcm/ZeroUI/internal/tui/keys"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// AppGridModel represents a high-performance grid of application cards
type AppGridModel struct {
	cards        []*AppCardModel
	statuses     []registry.AppStatus
	selectedIdx  int
	columns      int
	width        int
	height       int
	cardSize     int // Card width
	cardHeight   int // Card height (separate from width for rectangular cards)
	keyMap       keys.AppKeyMap
	styles       *styles.Styles
	showAll      bool
	filterByCategory string
	
	// Performance optimizations
	viewport       viewport.Model
	cachedView     string
	lastCacheTime  time.Time
	cacheDuration  time.Duration
	lastWidth      int
	lastHeight     int
	lastSelection  int
	
	// Responsive design
	minCardSize    int
	maxCardSize    int
	cardSpacing    int
	
	// Visual enhancements
	animationStep  int
	showAnimation  bool
}

// NewAppGrid creates a new high-performance app grid component
func NewAppGrid() *AppGridModel {
	statuses := registry.GetAppStatuses()
	cards := make([]*AppCardModel, len(statuses))
	
	for i, status := range statuses {
		cards[i] = NewAppCard(status)
	}
	
	// Select first card if available and set selectedIdx
	if len(cards) > 0 {
		cards[0].SetSelected(true)
	}
	
	// Viewport disabled - was causing border alignment issues
	vp := viewport.New(80, 24)
	vp.YPosition = 0
	
	return &AppGridModel{
		cards:         cards,
		statuses:      statuses,
		selectedIdx:   0,  // Initialize to 0 to fix double press issue
		columns:       4,  // Default to 4 columns (production standard)
		cardSize:      30, // Width of cards (rectangular)
		cardHeight:    10, // Height of cards (shorter than width for rectangular)
		keyMap:        keys.DefaultKeyMap(),
		styles:        styles.GetStyles(),
		showAll:       false,
		viewport:      vp,
		cacheDuration: 16 * time.Millisecond, // 60fps optimization
		minCardSize:   24,
		maxCardSize:   36, // Reduced max size for better 4-column fit
		cardSpacing:   2, // Optimized spacing for 4 columns
		showAnimation: true,
	}
}

// Init initializes the app grid
func (m *AppGridModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, card := range m.cards {
		cmds = append(cmds, card.Init())
	}
	return tea.Batch(cmds...)
}

// Update handles messages with performance optimization and smooth navigation
func (m *AppGridModel) Update(msg tea.Msg) (*AppGridModel, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			m.updateResponsiveLayout()
			m.invalidateCache()
			
			// Update viewport size
			m.viewport.Width = m.width
			m.viewport.Height = m.height - 12 // Account for header/footer
		}
		return m, nil
		
	case tea.KeyMsg:
		// Handle navigation with smooth animation
		switch {
		case key.Matches(msg, m.keyMap.Up):
			cmd := m.moveSelectionAnimated(-m.columns)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
			
		case key.Matches(msg, m.keyMap.Down):
			cmd := m.moveSelectionAnimated(m.columns)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
			
		case key.Matches(msg, m.keyMap.Left):
			cmd := m.moveSelectionAnimated(-1)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
			
		case key.Matches(msg, m.keyMap.Right):
			cmd := m.moveSelectionAnimated(1)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
			
		case key.Matches(msg, m.keyMap.Enter):
			if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
				app := m.statuses[m.selectedIdx]
				if app.IsInstalled || app.HasConfig {
					// Set loading state for visual feedback
					m.cards[m.selectedIdx].SetLoadingState(true)
					m.invalidateCache()
					return m, SelectAppCmd(app.Definition.Name)
				}
			}
			return m, nil
			
		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.showAll = !m.showAll
			m.updateFilter()
			m.invalidateCache()
			return m, nil
			
		case key.Matches(msg, key.NewBinding(key.WithKeys("i"))):
			// Show only installed apps
			m.filterByCategory = ""
			m.showAll = false
			m.updateFilter()
			m.invalidateCache()
			return m, nil
			
		// Viewport scrolling
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			m.viewport.HalfViewUp()
			m.invalidateCache()
			return m, nil
			
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))):
			m.viewport.HalfViewDown()
			m.invalidateCache()
			return m, nil
		}
	}
	
	// Update all cards for animations and state changes
	for i, card := range m.cards {
		updatedCard, cmd := card.Update(msg)
		m.cards[i] = updatedCard
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	// Update viewport
	vpModel, cmd := m.viewport.Update(msg)
	m.viewport = vpModel
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// View renders the app grid with advanced caching and performance optimization
func (m *AppGridModel) View() string {
	// Check cache for performance (60fps optimization)
	if m.isCacheValid() {
		return m.cachedView
	}
	
	if len(m.cards) == 0 {
		return m.renderEmptyState()
	}
	
	// Render the grid with advanced layout
	renderedGrid := m.renderAdvancedGrid()
	
	// Update cache
	m.cachedView = renderedGrid
	m.lastCacheTime = time.Now()
	
	return renderedGrid
}

// renderAdvancedGrid creates the grid with rectangular cards and responsive design
func (m *AppGridModel) renderAdvancedGrid() string {
	// Build grid rows with rectangular cards optimized for 4-column layout
	var rows []string
	visibleCards := m.getVisibleCards()
	
	if len(visibleCards) == 0 {
		return m.renderEmptyState()
	}
	
	// Calculate optimal spacing for perfect alignment with 10% margins
	sideMarginPercent := 10
	if m.width < 80 {
		sideMarginPercent = 5 // Smaller margins on narrow screens
	}
	
	totalMargins := (m.width * sideMarginPercent * 2) / 100 // 10% on each side
	availableContentWidth := m.width - totalMargins
	
	totalSpacing := (m.columns - 1) * m.cardSpacing
	idealCardWidth := (availableContentWidth - totalSpacing) / m.columns
	
	// Use ideal width but constrain by min/max
	if idealCardWidth > m.maxCardSize {
		m.cardSize = m.maxCardSize
	} else if idealCardWidth < m.minCardSize {
		m.cardSize = m.minCardSize
	} else {
		m.cardSize = idealCardWidth
	}
	
	// Recalculate actual margins based on card size
	totalCardWidth := m.columns * m.cardSize
	actualContentWidth := totalCardWidth + totalSpacing
	leftMargin := (m.width - actualContentWidth) / 2
	
	// Ensure leftMargin is never negative
	if leftMargin < 0 {
		leftMargin = 2 // Minimum padding
	}
	
	visibleLen := len(visibleCards)
	for i := 0; i < visibleLen; i += m.columns {
		var rowCards []string
		
		// Build row with rectangular cards  
		for j := 0; j < m.columns && i+j < visibleLen; j++ {
			idx := i + j
			card := visibleCards[idx]
			
			// Ensure rectangular dimensions (width x height)
			card.SetSize(m.cardSize, m.cardHeight)
			
			// Consistent card rendering without layout-shifting animations
			rowCards = append(rowCards, card.View())
		}
		
		// Add perfect spacing between cards - optimized with string builder
		builder := performance.GetBuilder()
		
		// Pre-calculate total row capacity for efficiency
		estimatedSize := len(rowCards)*m.cardSize + (len(rowCards)-1)*m.cardSpacing + leftMargin
		builder.Grow(estimatedSize)
		
		// Add left margin
		builder.WriteString(performance.GetSpacer(leftMargin))
		
		// Build row with cards and spacing
		for k, card := range rowCards {
			builder.WriteString(card)
			if k < len(rowCards)-1 {
				builder.WriteString(performance.GetSpacer(m.cardSpacing))
			}
		}
		
		row := builder.String()
		performance.PutBuilder(builder)
		rows = append(rows, row)
		
		// Add consistent vertical spacing between rows
		if i+m.columns < visibleLen {
			for s := 0; s < 2; s++ {
				rows = append(rows, "")
			}
		}
	}
	
	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	
	// Calculate responsive layout
	gridHeight := lipgloss.Height(grid)
	headerHeight := m.calculateHeaderHeight()
	footerHeight := 3
	totalContentHeight := headerHeight + gridHeight + footerHeight
	verticalPadding := (m.height - totalContentHeight) / 2
	if verticalPadding < 1 {
		verticalPadding = 1
	}
	
	// Create enhanced header and footer
	header := m.renderEnhancedHeader()
	footer := m.renderEnhancedFooter()
	
	// Build the full screen view with perfect spacing
	var content []string
	
	// Add responsive top padding
	for i := 0; i < verticalPadding; i++ {
		content = append(content, "")
	}
	
	content = append(content, header)
	content = append(content, "")
	content = append(content, grid)
	content = append(content, "")
	content = append(content, footer)
	
	// Removed viewport content update - not using viewport for main grid
	
	// Return the content directly without viewport
	// The viewport was causing alignment issues with borders
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderHeader creates the responsive header with enhanced logo
func (m *AppGridModel) renderHeader() string {
	var logo string
	var logoStyle lipgloss.Style
	
	// Responsive logo selection with enhanced effects
	if m.width < 60 {
		// Minimal text for very small screens
		logo = "ZEROUI"
		logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(1, 0)
	} else if m.width < 80 {
		// Compact logo with gradient
		logo = GetMinimalLogo()
		logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true)
	} else {
		// Full ASCII logo with enhanced styling
		logo = GetASCIILogo()
		logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
	}
	
	// Enhanced subtitle with app count and status
	visibleCount := len(m.getVisibleCards())
	installedCount := 0
	for _, status := range m.statuses {
		if status.IsInstalled {
			installedCount++
		}
	}
	
	var subtitle string
	if m.showAll {
		subtitle = fmt.Sprintf("%d applications • %d installed • showing all", len(m.statuses), installedCount)
	} else {
		subtitle = fmt.Sprintf("%d applications • %d available • installed only", visibleCount, visibleCount)
	}
	
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)
	
	// Center and style the logo
	var logoLines []string
	for _, line := range strings.Split(strings.TrimSpace(logo), "\n") {
		if line != "" {
			centeredLine := lipgloss.NewStyle().
				Width(m.width).
				Align(lipgloss.Center).
				Render(logoStyle.Render(line))
			logoLines = append(logoLines, centeredLine)
		}
	}
	
	// Center the subtitle
	centeredSubtitle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(subtitleStyle.Render(subtitle))
	
	// Join with perfect spacing
	header := lipgloss.JoinVertical(
		lipgloss.Center,
		strings.Join(logoLines, "\n"),
		centeredSubtitle,
	)
	
	return header
}

// renderFooter creates the grid footer
func (m *AppGridModel) renderFooter() string {
	var hints []string
	
	hints = append(hints, "↑↓←→ Navigate")
	hints = append(hints, "⏎ Select")
	
	if m.showAll {
		hints = append(hints, "a Show Installed")
	} else {
		hints = append(hints, "a Show All")
	}
	
	hints = append(hints, "q Quit")
	
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	return hintStyle.Render(strings.Join(hints, "  •  "))
}

// renderEmptyState renders when no apps are available
func (m *AppGridModel) renderEmptyState() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Align(lipgloss.Center).
		MarginTop(5)
	
	return emptyStyle.Render("No applications found")
}

// moveSelectionAnimated moves selection with smooth animation effects
func (m *AppGridModel) moveSelectionAnimated(offset int) tea.Cmd {
	visibleCards := m.getVisibleCards()
	if len(visibleCards) == 0 {
		return nil
	}
	
	// Clear current selection
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
		m.cards[m.selectedIdx].SetSelected(false)
	}
	
	// Initialize selectedIdx if it's unset (fixes double press issue)
	if m.selectedIdx < 0 && len(visibleCards) > 0 {
		m.selectedIdx = 0
	}
	
	// Calculate new index with bounds checking
	newIdx := m.selectedIdx + offset
	
	// Smart wrapping for grid navigation
	if offset == -m.columns || offset == m.columns {
		// Vertical movement
		if newIdx < 0 {
			// Wrap to bottom row, same column
			col := m.selectedIdx % m.columns
			lastRowStart := ((len(visibleCards)-1)/m.columns) * m.columns
			newIdx = lastRowStart + col
			if newIdx >= len(visibleCards) {
				newIdx = len(visibleCards) - 1
			}
		} else if newIdx >= len(visibleCards) {
			// Wrap to top row, same column
			col := m.selectedIdx % m.columns
			newIdx = col
		}
	} else {
		// Horizontal movement
		if newIdx < 0 {
			newIdx = len(visibleCards) - 1
		} else if newIdx >= len(visibleCards) {
			newIdx = 0
		}
	}
	
	m.selectedIdx = newIdx
	
	// Set new selection with animation
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
		m.cards[m.selectedIdx].SetSelected(true)
		m.animationStep++
	}
	
	// Invalidate cache for smooth updates
	m.invalidateCache()
	
	// Return animation command
	if m.showAnimation {
		return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return AnimationTickMsg{}
		})
	}
	
	return nil
}

// getVisibleCards returns cards based on current filter
func (m *AppGridModel) getVisibleCards() []*AppCardModel {
	if m.showAll {
		return m.cards
	}
	
	var visible []*AppCardModel
	for i, card := range m.cards {
		status := m.statuses[i]
		if status.IsInstalled || status.HasConfig {
			visible = append(visible, card)
		}
	}
	
	return visible
}

// updateResponsiveLayout updates the grid layout with responsive design
func (m *AppGridModel) updateResponsiveLayout() {
	// Ensure minimum spacing
	if m.cardSpacing < 2 {
		m.cardSpacing = 2
	}
	
	// Calculate optimal card size and columns based on screen size
	availableWidth := m.width - 20 // Account for margins
	if availableWidth < 40 {
		availableWidth = 40 // Minimum working width
	}
	
	// Determine optimal layout - prioritize 4 columns (production standard)
	if m.width < 60 {
		// Very small screens - single column
		m.columns = 1
		m.cardSize = min(availableWidth, 30)
		m.cardHeight = 10
	} else if m.width < 80 {
		// Small screens - two columns
		m.columns = 2
		m.cardSize = (availableWidth - m.cardSpacing) / 2
		m.cardHeight = 10
	} else if m.width < 120 {
		// Medium screens - three columns
		m.columns = 3
		m.cardSize = (availableWidth - 2*m.cardSpacing) / 3
		m.cardHeight = 10
	} else {
		// Standard and large screens - four columns (default and preferred)
		m.columns = 4
		m.cardSize = min((availableWidth - 3*m.cardSpacing) / 4, 36)
		m.cardHeight = 10 // Consistent rectangular height
	}
	
	// Enforce size constraints
	if m.cardSize < m.minCardSize {
		m.cardSize = m.minCardSize
		// Recalculate columns with minimum size
		m.columns = availableWidth / (m.cardSize + m.cardSpacing)
		if m.columns < 1 {
			m.columns = 1
		}
	} else if m.cardSize > m.maxCardSize {
		m.cardSize = m.maxCardSize
	}
	
	// Update all card sizes to rectangular dimensions (width x height)
	for _, card := range m.cards {
		card.SetSize(m.cardSize, m.cardHeight)
	}
	
	// Update viewport dimensions
	m.viewport.Width = m.width
	m.viewport.Height = m.height - 12
}

// updateFilter updates the visible cards based on filter with animation
func (m *AppGridModel) updateFilter() {
	// Reset selection with animation
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
		m.cards[m.selectedIdx].SetSelected(false)
	}
	
	// Select first visible card
	visible := m.getVisibleCards()
	if len(visible) > 0 {
		// Find the index of the first visible card in the main array
		for i, card := range m.cards {
			if card == visible[0] {
				m.selectedIdx = i
				m.cards[i].SetSelected(true)
				break
			}
		}
	} else {
		m.selectedIdx = -1
	}
	
	// Update layout after filter change
	m.updateResponsiveLayout()
}

// Performance and visual enhancement methods

// AnimationTickMsg represents animation frame updates
type AnimationTickMsg struct{}

// isCacheValid checks if the cached view is still valid
func (m *AppGridModel) isCacheValid() bool {
	if m.cachedView == "" {
		return false
	}
	
	// Invalidate cache if layout changed
	if m.lastWidth != m.width || m.lastHeight != m.height || m.lastSelection != m.selectedIdx {
		return false
	}
	
	return time.Since(m.lastCacheTime) < m.cacheDuration
}

// invalidateCache forces a re-render on next View() call
func (m *AppGridModel) invalidateCache() {
	m.cachedView = ""
	m.lastCacheTime = time.Time{}
	m.lastWidth = m.width
	m.lastHeight = m.height
	m.lastSelection = m.selectedIdx
}

// calculateHeaderHeight calculates dynamic header height based on screen size
func (m *AppGridModel) calculateHeaderHeight() int {
	if m.width < 60 {
		return 3 // Minimal header for small screens
	} else if m.width < 80 {
		return 6 // Medium header
	}
	return 9 // Full ASCII logo header
}

// renderEnhancedHeader creates responsive header with logo
func (m *AppGridModel) renderEnhancedHeader() string {
	header := m.renderHeader()
	
	// Add gradient effect to header
	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("212")).
		Bold(true)
	
	return headerStyle.Render(header)
}

// renderEnhancedFooter creates footer with enhanced styling
func (m *AppGridModel) renderEnhancedFooter() string {
	footer := m.renderFooter()
	
	// Enhanced footer styling
	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)
	
	return footerStyle.Render(footer)
}

// GetSelectedApp returns the currently selected app name
func (m *AppGridModel) GetSelectedApp() string {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.statuses) {
		return m.statuses[m.selectedIdx].Definition.Name
	}
	return ""
}