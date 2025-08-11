package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/keys"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// HuhGridModel represents a Huh-based grid layout for app selection
type HuhGridModel struct {
	width       int
	height      int
	statuses    []registry.AppStatus
	form        *huh.Form
	selectedApp string
	focused     bool
	keyMap      keys.AppKeyMap
	styles      *styles.Styles
	showAll     bool
	columns     int

	// Grid layout settings
	cardWidth   int
	cardHeight  int
	gridSpacing int
}

// NewHuhGrid creates a new Huh-based grid component
func NewHuhGrid() *HuhGridModel {
	statuses := registry.GetAppStatuses()

	model := &HuhGridModel{
		statuses:    statuses,
		focused:     false,
		keyMap:      keys.DefaultKeyMap(),
		styles:      styles.GetStyles(),
		showAll:     false,
		columns:     4,  // Default 4 columns
		cardWidth:   30, // Rectangular cards
		cardHeight:  10,
		gridSpacing: 2,
	}

	model.buildGridForm()
	return model
}

// buildGridForm creates a multi-column Huh form for app selection
func (m *HuhGridModel) buildGridForm() {
	var options []huh.Option[string]

	// Show applications - by default show all, filter only if specifically requested
	for _, status := range m.statuses {
		// Always include the application in the grid
		// The showAll flag can be used for different filtering in the future

		// Create grid-friendly labels with status indicators
		var statusIcons []string
		if status.IsInstalled {
			statusIcons = append(statusIcons, "✓")
		}
		if status.HasConfig {
			statusIcons = append(statusIcons, "⚙️")
		}
		if !status.IsInstalled && !status.HasConfig {
			statusIcons = append(statusIcons, "○")
		}

		// Compact label format for grid layout
		statusText := ""
		if len(statusIcons) > 0 {
			statusText = fmt.Sprintf("[%s]", strings.Join(statusIcons, ""))
		}

		// Multi-line label for better grid presentation with string builder
		var labelBuilder strings.Builder
		labelBuilder.Grow(64) // Pre-allocate space to reduce allocations
		labelBuilder.WriteString(status.Definition.Logo)
		labelBuilder.WriteByte(' ')
		labelBuilder.WriteString(status.Definition.Name)
		labelBuilder.WriteByte('\n')
		labelBuilder.WriteString(statusText)
		labelBuilder.WriteByte(' ')
		labelBuilder.WriteString(status.Definition.Category)
		label := labelBuilder.String()

		options = append(options, huh.NewOption(label, status.Definition.Name))
	}

	if len(options) == 0 {
		// Add fallback option
		options = append(options, huh.NewOption("No applications available", ""))
	}

	// Create multi-column select with grid-like appearance
	selectField := huh.NewSelect[string]().
		Title("Select Application").
		Description("Navigate with arrow keys, select with Enter").
		Options(options...).
		Value(&m.selectedApp).
		Height(calculateGridHeight(len(options), m.columns)). // Dynamic height based on grid
		WithTheme(huh.ThemeCharm())

	m.form = huh.NewForm(
		huh.NewGroup(selectField).Title("Applications Grid"),
	).
		WithShowHelp(true).
		WithShowErrors(true)
}

// calculateGridHeight calculates the optimal height for the grid display
func calculateGridHeight(itemCount, columns int) int {
	rows := (itemCount + columns - 1) / columns // Ceiling division

	// Each row needs space for the card (height + spacing)
	// Add some padding for borders and spacing
	baseHeight := 8 // Minimum height
	rowHeight := 3  // Height per row including spacing
	maxHeight := 15 // Maximum reasonable height

	calculatedHeight := baseHeight + (rows * rowHeight)

	if calculatedHeight > maxHeight {
		return maxHeight
	}
	if calculatedHeight < baseHeight {
		return baseHeight
	}

	return calculatedHeight
}

// Init initializes the component
func (m *HuhGridModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles messages with proper grid navigation
func (m *HuhGridModel) Update(msg tea.Msg) (*HuhGridModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			m.updateGridLayout()
		}

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Handle grid-specific key bindings
		switch {
		case key.Matches(msg, m.keyMap.Toggle):
			// Toggle between showing all apps and only available ones
			m.showAll = !m.showAll
			m.buildGridForm()
			return m, m.form.Init()

		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			// Grid mode toggle (future enhancement)
			// Could switch between different grid sizes
			m.toggleGridSize()
			return m, nil
		}
	}

	// Always update the form for proper Huh integration
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		m.form = form.(*huh.Form)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check if app was selected
		if m.form.State == huh.StateCompleted && m.selectedApp != "" {
			cmds = append(cmds, SelectAppCmd(m.selectedApp))
		}
	}

	// Return with batched commands
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// toggleGridSize cycles through different grid column counts
func (m *HuhGridModel) toggleGridSize() {
	switch m.columns {
	case 2:
		m.columns = 3
	case 3:
		m.columns = 4
	case 4:
		m.columns = 6
	case 6:
		m.columns = 2
	default:
		m.columns = 4 // Default fallback
	}

	// Rebuild form with new grid layout
	m.buildGridForm()
}

// updateGridLayout adjusts grid parameters based on screen size
func (m *HuhGridModel) updateGridLayout() {
	if m.width < 60 {
		m.columns = 1
		m.cardWidth = 40
	} else if m.width < 90 {
		m.columns = 2
		m.cardWidth = 35
	} else if m.width < 115 {
		m.columns = 3
		m.cardWidth = 32
	} else {
		m.columns = 4 // Production standard for 115+ width
		m.cardWidth = 30
	}

	// Rebuild form with updated layout
	m.buildGridForm()
}

// View renders the grid component
func (m *HuhGridModel) View() string {
	if len(m.statuses) == 0 {
		return m.renderEmptyState()
	}

	// Create header with grid information
	header := m.renderGridHeader()

	// Get the form view
	var formView string
	if m.form != nil {
		formView = m.form.View()
	} else {
		// Form not initialized - rebuild it
		m.buildGridForm()
		if m.form != nil {
			formView = m.form.View()
		} else {
			formView = "Form initialization failed"
		}
	}

	// Create footer with grid-specific help
	footer := m.renderGridFooter()

	// Calculate content dimensions
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	availableHeight := m.height - headerHeight - footerHeight - 4

	if availableHeight < 10 {
		availableHeight = 10
	}

	// Style the grid container
	gridStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(availableHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Align(lipgloss.Center, lipgloss.Top)

	styledGrid := gridStyle.Render(formView)

	// Compose the final view
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		styledGrid,
		footer,
	)

	// Center everything in available space
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderGridHeader creates header with grid information
func (m *HuhGridModel) renderGridHeader() string {
	title := fmt.Sprintf("🔧 Application Grid (%d columns)", m.columns)

	// Count applications
	installedCount := 0
	configuredCount := 0
	totalCount := len(m.statuses)
	visibleCount := 0

	for _, status := range m.statuses {
		if status.IsInstalled {
			installedCount++
		}
		if status.HasConfig {
			configuredCount++
		}
		if m.showAll || status.IsInstalled || status.HasConfig {
			visibleCount++
		}
	}

	subtitle := fmt.Sprintf("📱 %d apps • ✅ %d installed • ⚙️ %d configured • 👁️ %d visible",
		totalCount, installedCount, configuredCount, visibleCount)

	viewMode := "Available Only"
	if m.showAll {
		viewMode = "All Applications"
	}
	subtitle += fmt.Sprintf(" • Mode: %s", viewMode)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(1)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		subtitleStyle.Render(subtitle),
	)
}

// renderGridFooter creates help text for grid navigation
func (m *HuhGridModel) renderGridFooter() string {
	var hints []string

	hints = append(hints, "↑↓←→ Navigate Grid")
	hints = append(hints, "⏎ Select App")
	hints = append(hints, "g Toggle Grid Size")

	if m.showAll {
		hints = append(hints, "t Show Available Only")
	} else {
		hints = append(hints, "t Show All Apps")
	}

	hints = append(hints, "? Help")
	hints = append(hints, "q Quit")

	helpText := strings.Join(hints, "  •  ")

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Padding(0, 2).
		MarginTop(1)

	return style.Render(helpText)
}

// renderEmptyState shows when no apps are found
func (m *HuhGridModel) renderEmptyState() string {
	emptyMsg := `
🤷 No Applications Found

It looks like no supported applications are installed
or configured on your system.

Supported applications include:
• Terminal emulators (Alacritty, Ghostty, WezTerm)
• Editors (VS Code, Neovim, Zed)
• Development tools (Git, Tmux, Starship)

Install any of these applications and they'll appear in the grid!
`

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E2E8F0")).
		Padding(2, 4).
		Width(60)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		style.Render(emptyMsg),
	)
}

// Focus implements the focusable interface
func (m *HuhGridModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur implements the focusable interface
func (m *HuhGridModel) Blur() tea.Cmd {
	m.focused = false
	return nil
}

// IsFocused returns whether the component is focused
func (m *HuhGridModel) IsFocused() bool {
	return m.focused
}

// SetSize updates the component size
func (m *HuhGridModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	m.updateGridLayout()
	return nil
}

// GetSize returns the current size
func (m *HuhGridModel) GetSize() (int, int) {
	return m.width, m.height
}

// Bindings returns key bindings for help
func (m *HuhGridModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Left,
		m.keyMap.Right,
		m.keyMap.Enter,
		m.keyMap.Toggle,
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "toggle grid size")),
		m.keyMap.Help,
		m.keyMap.Quit,
	}
}

// GetSelectedApp returns the currently selected app
func (m *HuhGridModel) GetSelectedApp() string {
	return m.selectedApp
}

// GetColumns returns the current number of columns
func (m *HuhGridModel) GetColumns() int {
	return m.columns
}

// SetColumns updates the grid column count
func (m *HuhGridModel) SetColumns(columns int) {
	if columns < 1 {
		columns = 1
	} else if columns > 8 {
		columns = 8
	}

	m.columns = columns
	m.buildGridForm()
}
