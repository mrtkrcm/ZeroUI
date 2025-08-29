package display

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// GlamourHelpModel provides rich markdown-based help system
type GlamourHelpModel struct {
	renderer    *glamour.TermRenderer
	content     map[string]string // Help content keyed by context
	currentPage string
	width       int
	height      int
	visible     bool

	// Navigation
	pages      []string
	currentIdx int
}

// NewGlamourHelp creates a new markdown-based help system
func NewGlamourHelp() *GlamourHelpModel {
	// Create renderer with ZeroUI-compatible styling
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	model := &GlamourHelpModel{
		renderer: renderer,
		content:  make(map[string]string),
		width:    80,
		height:   24,
	}

	// Initialize with built-in help content
	model.loadBuiltinContent()

	return model
}

// loadBuiltinContent loads the default help content
func (m *GlamourHelpModel) loadBuiltinContent() {
	m.content["overview"] = `# ZeroUI Help

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools.

## Quick Start

- **Navigate**: Use arrow keys or hjkl (vim-style)
- **Select**: Press Enter to configure an application
- **Search**: Press / to search for applications
- **Filter**: Press 'a' to toggle between all/installed apps
- **Help**: Press ? to toggle this help system
- **Quit**: Press q or Ctrl+C to exit

## Features

- 🔧 **Zero Configuration**: Works out of the box
- 🎨 **Theme Management**: Consistent themes across tools
- ⚡ **Fast Performance**: Optimized for speed
- 🔍 **Smart Search**: Find apps quickly
- 💅 **Beautiful UI**: Modern terminal interface`

	m.content["navigation"] = `# Navigation Guide

## Basic Movement

| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| ←/h | Move left |
| →/l | Move right |
| Enter | Select/Confirm |
| Esc | Back/Cancel |

## Advanced Navigation

- **Page Up/Down**: Use PgUp/PgDown for faster scrolling
- **Home/End**: Jump to beginning/end of lists
- **Tab**: Navigate between form fields
- **Shift+Tab**: Navigate backwards through form fields

## Search & Filter

- **Search**: Press '/' to start searching
- **Clear Search**: Press Esc while searching
- **Filter Installed**: Press 'i' to show only installed apps
- **Show All**: Press 'a' to show all available apps

## Compact Mode

- **Toggle Compact**: Press 'c' to switch between compact and detailed views
- **Responsive**: UI automatically adapts to terminal size`

	m.content["configuration"] = `# Configuration Guide

## App Configuration

1. **Select Application**: Navigate to an app and press Enter
2. **Edit Fields**: Use Tab to move between configuration fields
3. **Field Types**:
   - **Text**: Type directly into input fields
   - **Numbers**: Enter numeric values (with validation)
   - **Booleans**: Toggle with Enter or Space
   - **Select**: Choose from dropdown options
4. **Save Changes**: Press Ctrl+S to save configuration
5. **Changed Only**: Press 'C' to toggle showing only changed fields
6. **Presets**: Press 'p' in the form to open presets and apply
5. **Reset Field**: Press Ctrl+R to reset a field to default

## Validation

- Real-time validation for all fields
- Clear error messages for invalid input
- Required fields are marked and validated
- Numeric fields support min/max constraints

## Presets

Some applications support configuration presets:
- Navigate to preset selection with 'p'
- Choose from predefined configurations
- Apply instantly to get started quickly

## File Management

- Configurations are saved to app-specific locations
- Automatic backup of existing configurations
- Support for various config formats (YAML, TOML, JSON)`

	m.content["troubleshooting"] = `# Troubleshooting

## Common Issues

### App Not Detected
- **Check Installation**: Ensure the application is installed
- **Path Issues**: Verify the application is in your PATH
- **Refresh**: Press 'r' to refresh the application list

### Configuration Not Saving
- **Permissions**: Check file permissions in config directory
- **Disk Space**: Ensure sufficient disk space
- **File Locks**: Close the application before configuring

### Display Issues
- **Terminal Size**: Ensure terminal is at least 80x24
- **Color Support**: Use a terminal with 256-color support
- **Font**: Use a monospace font for best results

## Performance Tips

- Use compact mode ('c') for better performance with many apps
- Filter to installed apps ('i') to reduce UI complexity
- Keep terminal size reasonable (80-120 columns)

## Getting Help

- **Context Help**: Press '?' for context-specific help
- **Full Help**: Press F1 for complete help system
- **Logs**: Check ~/.local/state/zeroui/zeroui.log for errors
- **Debug Mode**: Run with --debug for detailed logging`

	m.content["keyboard"] = `# Keyboard Shortcuts

## Global Shortcuts

| Shortcut | Action |
|----------|--------|
| q | Quit application |
| Ctrl+C | Force quit |
| ? | Toggle help |
| F1 | Full help system |
| / | Search |
| Esc | Back/Cancel |

## Navigation Shortcuts

| Shortcut | Action |
|----------|--------|
| ↑↓←→ | Navigate |
| hjkl | Vim-style navigation |
| PgUp/PgDn | Page navigation |
| Home/End | Jump to start/end |
| Tab | Next field |
| Shift+Tab | Previous field |

## Application Shortcuts

| Shortcut | Action |
|----------|--------|
| Enter | Select/Configure |
| Space | Quick toggle |
| a | Show all apps |
| i | Show installed only |
| c | Toggle compact mode |
| r | Refresh |

## Configuration Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+S | Save configuration |
| Ctrl+R | Reset field |
| Ctrl+E | Export config |
| Tab | Next field |
| Enter | Edit/Confirm |

## Advanced Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Alt+D | Debug mode |
| Ctrl+Alt+S | Screenshot |
| F5 | Full refresh |`

	// Set up page navigation
	m.pages = []string{"overview", "navigation", "configuration", "keyboard", "troubleshooting"}
	m.currentPage = "overview"
}

// SetContent sets custom help content for a specific page
func (m *GlamourHelpModel) SetContent(page, content string) {
	m.content[page] = content
}

// ShowPage displays a specific help page
func (m *GlamourHelpModel) ShowPage(page string) {
	if _, exists := m.content[page]; exists {
		m.currentPage = page
		m.visible = true

		// Update current index for navigation
		for i, p := range m.pages {
			if p == page {
				m.currentIdx = i
				break
			}
		}
	}
}

// Toggle toggles the help system visibility
func (m *GlamourHelpModel) Toggle() {
	m.visible = !m.visible
}

// IsVisible returns whether the help system is currently visible
func (m *GlamourHelpModel) IsVisible() bool {
	return m.visible
}

// Init initializes the help system
func (m *GlamourHelpModel) Init() tea.Cmd {
	return nil
}

// Update handles help system updates
func (m *GlamourHelpModel) Update(msg tea.Msg) (*GlamourHelpModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update renderer word wrap
		if m.renderer != nil {
			m.renderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.width-8),
			)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc":
			m.visible = false
			return m, nil

		case "left", "h":
			if m.currentIdx > 0 {
				m.currentIdx--
				m.currentPage = m.pages[m.currentIdx]
			}

		case "right", "l":
			if m.currentIdx < len(m.pages)-1 {
				m.currentIdx++
				m.currentPage = m.pages[m.currentIdx]
			}

		case "1":
			m.ShowPage("overview")
		case "2":
			m.ShowPage("navigation")
		case "3":
			m.ShowPage("configuration")
		case "4":
			m.ShowPage("keyboard")
		case "5":
			m.ShowPage("troubleshooting")
		}
	}

	return m, nil
}

// View renders the help system
func (m *GlamourHelpModel) View() string {
	if !m.visible {
		return ""
	}

	// Get the markdown content for current page
	content, exists := m.content[m.currentPage]
	if !exists {
		content = "# Page Not Found\n\nThe requested help page could not be found."
	}

	// Render the markdown with a timeout to avoid blocking/hanging tests.
	// Some renderer calls may block (environmental or library issues). Run
	// the render in a separate goroutine and select on a timeout.
	type renderResult struct {
		out string
		err error
	}
	// Declare `rendered` up-front so it's available to the select branch assignments.
	var rendered string
	renderCh := make(chan renderResult, 1)
	go func() {
		r, e := m.renderer.Render(content)
		renderCh <- renderResult{out: r, err: e}
	}()

	select {
	case res := <-renderCh:
		if res.err != nil {
			rendered = fmt.Sprintf("Error rendering help: %v", res.err)
		} else {
			rendered = res.out
		}
	case <-time.After(200 * time.Millisecond):
		// If rendering takes too long, return a safe message so tests don't hang.
		rendered = "Error: help rendering timed out"
	}

	// Create navigation tabs
	tabs := m.renderTabs()

	// Prominent Help header (literal word for clarity/tests)
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("Help")

	// Create help footer
	footer := m.renderFooter()

	// Combine everything with proper styling
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(m.width - 4).
		Height(m.height - 4)

	content = lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		"",
		strings.TrimSpace(rendered),
		"",
		footer,
	)

	return helpStyle.Render(content)
}

// renderTabs creates navigation tabs for different help pages
func (m *GlamourHelpModel) renderTabs() string {
	var tabs []string

	for i, page := range m.pages {
		title := strings.Title(page)
		number := fmt.Sprintf("%d", i+1)

		var tabStyle lipgloss.Style
		if page == m.currentPage {
			tabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Background(lipgloss.Color("235")).
				Bold(true).
				Padding(0, 1)
		} else {
			tabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Padding(0, 1)
		}

		tab := tabStyle.Render(fmt.Sprintf("%s:%s", number, title))
		tabs = append(tabs, tab)
	}

	return strings.Join(tabs, " ")
}

// renderFooter creates the help footer with navigation instructions
func (m *GlamourHelpModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true)

	return footerStyle.Render("←→/hl: Navigate pages • 1-5: Jump to page • ?: Close help")
}

// SetSize updates the help system dimensions
func (m *GlamourHelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update renderer
	if m.renderer != nil {
		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width-8),
		)
	}
}

// GetCurrentPage returns the currently displayed page
func (m *GlamourHelpModel) GetCurrentPage() string {
	return m.currentPage
}

// GetAvailablePages returns all available help pages
func (m *GlamourHelpModel) GetAvailablePages() []string {
	return m.pages
}

// AddPage adds a new help page
func (m *GlamourHelpModel) AddPage(page, content string) {
	m.content[page] = content

	// Add to pages list if not already present
	for _, p := range m.pages {
		if p == page {
			return
		}
	}
	m.pages = append(m.pages, page)
}
