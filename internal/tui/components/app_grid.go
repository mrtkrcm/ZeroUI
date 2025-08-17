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
	cards            []*AppCardModel
	statuses         []registry.AppStatus
	selectedIdx      int
	columns          int
	width            int
	height           int
	cardSize         int // Card width
	cardHeight       int // Card height (separate from width for rectangular cards)
	keyMap           keys.AppKeyMap
	styles           *styles.Styles
	showAll          bool
	filterByCategory string

	// Simple state tracking
	viewport viewport.Model

	// Responsive design
	minCardSize int
	maxCardSize int
	cardSpacing int

	// Visual enhancements
	animationStep int
	showAnimation bool
}

// NewAppGrid creates a new high-performance app grid component
func NewAppGrid() *AppGridModel {
	statuses := registry.GetAppStatuses()
	cards := make([]*AppCardModel, len(statuses))

	for i, status := range statuses {
		cards[i] = NewAppCard(status)
	}

	// Initialize with proper selection handling (will be set in NewModel)

	// Viewport disabled - was causing border alignment issues
	vp := viewport.New(80, 24)
	vp.YPosition = 0

	return &AppGridModel{
		cards:         cards,
		statuses:      statuses,
		selectedIdx:   -1, // Will be set properly in updateFilter
		columns:       4,  // Default to 4 columns (production standard)
		cardSize:      15, // Width of compact cards (half original size)
		cardHeight:    5,  // Height of compact cards (half original size)
		keyMap:        keys.DefaultKeyMap(),
		styles:        styles.GetStyles(),
		showAll:       false,
		viewport:      vp,
		minCardSize:   12,
		maxCardSize:   18,
		cardSpacing:   2,
		showAnimation: false, // Disabled for stability
	}
}

// Init initializes the app grid
func (m *AppGridModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, card := range m.cards {
		cmds = append(cmds, card.Init())
	}

	// Initialize selection properly
	m.updateFilter()

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

			// Update viewport size
			m.viewport.Width = m.width
			m.viewport.Height = m.height - 12 // Account for header/footer
		}
		return m, nil

	case tea.KeyMsg:
		// Handle navigation with smooth animation
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.moveSelection(-m.columns)
		case key.Matches(msg, m.keyMap.Down):
			m.moveSelection(m.columns)
		case key.Matches(msg, m.keyMap.Left):
			m.moveSelection(-1)
		case key.Matches(msg, m.keyMap.Right):
			m.moveSelection(1)

		case key.Matches(msg, m.keyMap.Enter):
			if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
				app := m.statuses[m.selectedIdx]
				if app.IsInstalled || app.HasConfig {
					// Set loading state for visual feedback
					m.cards[m.selectedIdx].SetLoadingState(true)
					return m, SelectAppCmd(app.Definition.Name)
				}
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			m.showAll = !m.showAll
			m.updateFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("i"))):
			// Show only installed apps
			m.filterByCategory = ""
			m.showAll = false
			m.updateFilter()
			return m, nil

		// Viewport scrolling
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup"))):
			m.viewport.HalfViewUp()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown"))):
			m.viewport.HalfViewDown()
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

// View renders the app grid with simple, fast rendering
func (m *AppGridModel) View() string {
	if len(m.cards) == 0 {
		return m.renderEmptyState()
	}
	// Use enhanced header/footer layout so tests see logo/title and counts
	return m.renderAdvancedGrid()
}

// renderSimpleGrid creates a simple, fast grid layout
func (m *AppGridModel) renderSimpleGrid() string {
	visibleCards := m.getVisibleCards()
	if len(visibleCards) == 0 {
		return m.renderEmptyState()
	}

	var rows []string
	for i := 0; i < len(visibleCards); i += m.columns {
		var rowCards []string
		for j := 0; j < m.columns && i+j < len(visibleCards); j++ {
			card := visibleCards[i+j]
			card.SetSize(m.cardSize, m.cardHeight)
			rowCards = append(rowCards, card.View())
		}

		row := strings.Join(rowCards, strings.Repeat(" ", m.cardSpacing))
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderAdvancedGrid creates the grid with rectangular cards and responsive design
func (m *AppGridModel) renderAdvancedGrid() string {
	// Build grid rows with rectangular cards optimized for 4-column layout
	var rows []string
	visibleCards := m.getVisibleCards()

	if len(visibleCards) == 0 {
		return m.renderEmptyState()
	}

	// Use fixed card size for consistent layout
	m.cardSize = 16  // Fixed compact width
	m.cardHeight = 5 // Fixed compact height

	// Calculate spacing to center the grid
	totalCardWidth := m.columns * m.cardSize
	totalSpacing := (m.columns - 1) * m.cardSpacing
	actualContentWidth := totalCardWidth + totalSpacing
	leftMargin := (m.width - actualContentWidth) / 2

	// Ensure leftMargin is never negative
	if leftMargin < 0 {
		leftMargin = 1 // Minimum padding
	}

	visibleLen := len(visibleCards)
	for i := 0; i < visibleLen; i += m.columns {
		var rowCards []string

		// Build row with compact cards
		for j := 0; j < m.columns && i+j < visibleLen; j++ {
			idx := i + j
			card := visibleCards[idx]

			// Ensure compact dimensions (width x height)
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

		// Add minimal vertical spacing between rows for compact cards
		if i+m.columns < visibleLen {
			rows = append(rows, "") // Single line spacing for compact design
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

	// Safety check to prevent crashes
	if m.cards == nil || len(m.cards) == 0 {
		return nil
	}

	// Map visible cards to their indices in the main array
	visibleIndices := m.getVisibleCardIndices()

	// Find current visible position
	currentVisiblePos := -1
	for i, idx := range visibleIndices {
		if idx == m.selectedIdx {
			currentVisiblePos = i
			break
		}
	}

	// If no current selection or selection not in visible cards, start at 0
	if currentVisiblePos < 0 {
		currentVisiblePos = 0
	}

	// Clear current selection
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
		m.cards[m.selectedIdx].SetSelected(false)
	}

	// Calculate new visible position with bounds checking
	newVisiblePos := currentVisiblePos + offset

	// Smart wrapping for grid navigation
	if offset == -m.columns || offset == m.columns {
		// Vertical movement
		if newVisiblePos < 0 {
			// Wrap to bottom - for small lists, go to last item
			if len(visibleCards) <= m.columns {
				newVisiblePos = len(visibleCards) - 1
			} else {
				// Normal wrapping for larger grids
				col := currentVisiblePos % m.columns
				lastRowStart := ((len(visibleCards) - 1) / m.columns) * m.columns
				newVisiblePos = lastRowStart + col
				if newVisiblePos >= len(visibleCards) {
					newVisiblePos = len(visibleCards) - 1
				}
			}
		} else if newVisiblePos >= len(visibleCards) {
			// Wrap to top - for small lists, go to first item
			if len(visibleCards) <= m.columns {
				newVisiblePos = 0
			} else {
				// Normal wrapping for larger grids
				col := currentVisiblePos % m.columns
				newVisiblePos = col
				if newVisiblePos >= len(visibleCards) {
					newVisiblePos = 0
				}
			}
		}
	} else {
		// Horizontal movement
		if newVisiblePos < 0 {
			newVisiblePos = len(visibleCards) - 1
		} else if newVisiblePos >= len(visibleCards) {
			newVisiblePos = 0
		}
	}

	// Convert back to main array index
	if newVisiblePos >= 0 && newVisiblePos < len(visibleIndices) {
		m.selectedIdx = visibleIndices[newVisiblePos]
	}

	// Set new selection with animation
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) {
		m.cards[m.selectedIdx].SetSelected(true)
		m.animationStep++
	}

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

// getVisibleCardIndices returns the indices of visible cards in the main array
func (m *AppGridModel) getVisibleCardIndices() []int {
	if m.showAll {
		indices := make([]int, len(m.cards))
		for i := range m.cards {
			indices[i] = i
		}
		return indices
	}

	var indices []int
	for i := range m.cards {
		status := m.statuses[i]
		if status.IsInstalled || status.HasConfig {
			indices = append(indices, i)
		}
	}

	return indices
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

	// Determine optimal layout with compact cards - force smaller sizes
	if m.width < 50 {
		// Very small screens - single column
		m.columns = 1
		m.cardSize = min(availableWidth, 15)
		m.cardHeight = 5
	} else if m.width <= 80 {
		// Small screens - two columns expected by tests
		m.columns = 2
		m.cardSize = min((availableWidth-1*m.cardSpacing)/2, 18)
		m.cardHeight = 5
	} else if m.width <= 120 {
		// Medium screens - three columns expected by tests
		m.columns = 3
		m.cardSize = min((availableWidth-2*m.cardSpacing)/3, 18)
		m.cardHeight = 5
	} else {
		// Large screens - cap at 4 columns expected by tests
		m.columns = 4
		m.cardSize = min((availableWidth-3*m.cardSpacing)/4, 18)
		m.cardHeight = 5 // Consistent compact height
	}

	// Enforce size constraints for compact cards
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

	// Update all card sizes to compact dimensions (width x height)
	for _, card := range m.cards {
		card.SetSize(m.cardSize, m.cardHeight)
	}

	// Update viewport dimensions
	m.viewport.Width = m.width
	m.viewport.Height = m.height - 12
}

// updateFilter updates the visible cards based on filter with animation
func (m *AppGridModel) updateFilter() {
	// Select first visible card using proper index mapping
	visibleIndices := m.getVisibleCardIndices()
	if len(visibleIndices) > 0 {
		// Always select first visible card when filtering
		m.selectedIdx = visibleIndices[0]
	} else {
		m.selectedIdx = -1
	}

	// Update card selection states
	m.updateCardSelection()

	// Update layout after filter change
	m.updateResponsiveLayout()
}

// Performance and visual enhancement methods

// AnimationTickMsg represents animation frame updates
type AnimationTickMsg struct{}

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

// updateCardSelection updates the visual selection state of all cards
func (m *AppGridModel) updateCardSelection() {
	// Safety check
	if m.cards == nil || len(m.cards) == 0 {
		return
	}

	// Clear all selections first
	for _, card := range m.cards {
		if card != nil {
			card.SetSelected(false)
		}
	}

	// Set the current selection if valid
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.cards) && m.cards[m.selectedIdx] != nil {
		m.cards[m.selectedIdx].SetSelected(true)
	}
}

// moveSelection moves selection with simple, predictable logic
func (m *AppGridModel) moveSelection(offset int) {
	if !m.isValidState() {
		return
	}

	visibleIndices := m.getVisibleCardIndices()
	if len(visibleIndices) == 0 {
		return
	}

	currentPos := m.findCurrentPosition(visibleIndices)
	newPos := m.calculateNewPosition(currentPos, offset, len(visibleIndices))

	m.selectedIdx = visibleIndices[newPos]
	m.updateCardSelection()
}

// isValidState checks if the grid is in a valid state
func (m *AppGridModel) isValidState() bool {
	return m.cards != nil && len(m.cards) > 0
}

// findCurrentPosition finds the current position in visible cards
func (m *AppGridModel) findCurrentPosition(visibleIndices []int) int {
	for i, idx := range visibleIndices {
		if idx == m.selectedIdx {
			return i
		}
	}
	return 0 // Default to first position
}

// calculateNewPosition calculates the new position with simple wrapping
func (m *AppGridModel) calculateNewPosition(current, offset, total int) int {
	newPos := current + offset
	if newPos < 0 {
		return total - 1 // Wrap to end
	}
	if newPos >= total {
		return 0 // Wrap to beginning
	}
	return newPos
}

// GetSelectedApp returns the currently selected app name
func (m *AppGridModel) GetSelectedApp() string {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.statuses) {
		return m.statuses[m.selectedIdx].Definition.Name
	}
	return ""
}
