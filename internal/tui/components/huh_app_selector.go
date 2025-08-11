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

// HuhAppSelectorModel represents the Huh-based app selection component
type HuhAppSelectorModel struct {
	width        int
	height       int
	statuses     []registry.AppStatus
	form         *huh.Form
	selectedApp  string
	focused      bool
	keyMap       keys.AppKeyMap
	styles       *styles.Styles
	showAll      bool
	lastFormSize int
}

// NewHuhAppSelector creates a new Huh-based app selector
func NewHuhAppSelector() *HuhAppSelectorModel {
	statuses := registry.GetAppStatuses()
	
	model := &HuhAppSelectorModel{
		statuses: statuses,
		focused:  false,
		keyMap:   keys.DefaultKeyMap(),
		styles:   styles.GetStyles(),
		showAll:  false,
	}
	
	model.buildForm()
	return model
}

// buildForm creates the Huh form for app selection
func (m *HuhAppSelectorModel) buildForm() {
	var options []huh.Option[string]
	
	// Show applications - by default show all available apps
	for _, status := range m.statuses {
		// Always include applications in the selector
		// Future: could add filtering based on categories or other criteria
		
		// Create descriptive label with status indicators
		var statusIcons []string
		if status.IsInstalled {
			statusIcons = append(statusIcons, "✓")
		}
		if status.HasConfig {
			statusIcons = append(statusIcons, "⚙️")
		}
		if !status.IsInstalled && !status.HasConfig {
			statusIcons = append(statusIcons, "❌")
		}
		
		statusText := ""
		if len(statusIcons) > 0 {
			statusText = fmt.Sprintf(" [%s]", strings.Join(statusIcons, " "))
		}
		
		label := fmt.Sprintf("%s %s%s - %s", 
			status.Definition.Logo, 
			status.Definition.Name, 
			statusText,
			status.Definition.Category)
		
		options = append(options, huh.NewOption(label, status.Definition.Name))
	}
	
	if len(options) == 0 {
		// Show a message when no apps are available
		options = append(options, huh.NewOption("No applications available", ""))
	}
	
	// Create the select form with modern styling
	selectField := huh.NewSelect[string]().
		Title("Select Application").
		Description("Choose an application to configure").
		Options(options...).
		Value(&m.selectedApp).
		Height(10). // Show more options at once
		WithTheme(huh.ThemeCharm())
	
	m.form = huh.NewForm(
		huh.NewGroup(selectField),
	).
		WithShowHelp(true).
		WithShowErrors(true)
	
	// Store form size for responsive updates
	m.lastFormSize = len(options)
}

// Init initializes the component
func (m *HuhAppSelectorModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles messages with proper Huh form lifecycle integration
func (m *HuhAppSelectorModel) Update(msg tea.Msg) (*HuhAppSelectorModel, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			m.updateFormSize()
		}
		
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		
		// Handle custom key bindings first, before form processes them
		switch {
		case key.Matches(msg, m.keyMap.Toggle):
			// Toggle between showing all apps and only installed/configured ones
			m.showAll = !m.showAll
			m.buildForm()
			return m, m.form.Init()
			
		case key.Matches(msg, m.keyMap.Enter):
			// Check if form is completed after processing
			if m.form != nil && m.form.State == huh.StateCompleted && m.selectedApp != "" {
				// App was selected, send selection message
				return m, SelectAppCmd(m.selectedApp)
			}
		}
	}
	
	// Always update the form with the message - this is critical for Huh integration
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		m.form = form.(*huh.Form)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Check if form state changed to completed
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

// View renders the component
func (m *HuhAppSelectorModel) View() string {
	if len(m.statuses) == 0 {
		return m.renderEmptyState()
	}
	
	// Create header with logo and app count info
	header := m.renderHeader()
	
	// Get the form view
	formView := m.form.View()
	
	// Create footer with help text
	footer := m.renderFooter()
	
	// Calculate content dimensions
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	availableHeight := m.height - headerHeight - footerHeight - 4 // padding
	
	if availableHeight < 10 {
		availableHeight = 10
	}
	
	// Style the form container with proper centering
	formStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(availableHeight).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center)
	
	styledForm := formStyle.Render(formView)
	
	// Compose the final view
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		styledForm,
		footer,
	)
	
	// Center everything in the available space
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderHeader creates a centered header with logo
func (m *HuhAppSelectorModel) renderHeader() string {
	logo := "🔧 ZeroUI"
	if m.width > 60 {
		logo = `
╔══════════════════════════════════════╗
║            🔧  Z E R O U I            ║
║        Configuration Manager          ║
╚══════════════════════════════════════╝`
	}
	
	// Count applications
	installedCount := 0
	configuredCount := 0
	totalCount := len(m.statuses)
	
	for _, status := range m.statuses {
		if status.IsInstalled {
			installedCount++
		}
		if status.HasConfig {
			configuredCount++
		}
	}
	
	subtitle := fmt.Sprintf("📱 %d apps • ✅ %d installed • ⚙️ %d configured", 
		totalCount, installedCount, configuredCount)
	
	if m.showAll {
		subtitle += " • Showing: All Applications"
	} else {
		subtitle += " • Showing: Available Only"
	}
	
	logoStyle := lipgloss.NewStyle().
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
		logoStyle.Render(logo),
		subtitleStyle.Render(subtitle),
	)
}

// renderFooter creates help text footer
func (m *HuhAppSelectorModel) renderFooter() string {
	var hints []string
	
	hints = append(hints, "↑↓ Navigate")
	hints = append(hints, "⏎ Select")
	
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

// renderEmptyState renders when no apps are found
func (m *HuhAppSelectorModel) renderEmptyState() string {
	emptyMsg := `
🤷 No Applications Found

It looks like no supported applications are installed
or configured on your system.

Supported applications include:
• Terminal emulators (Alacritty, Ghostty, WezTerm)
• Editors (VS Code, Neovim, Zed)
• Development tools (Git, Tmux, Starship)

Install any of these applications and they'll appear here!
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

// updateFormSize updates the form based on available space
func (m *HuhAppSelectorModel) updateFormSize() {
	if m.form == nil {
		return
	}
	
	// Calculate optimal form height
	maxHeight := m.height - 10 // Reserve space for header/footer
	if maxHeight < 5 {
		maxHeight = 5
	} else if maxHeight > 15 {
		maxHeight = 15 // Don't make it too tall
	}
	
	// Rebuild form if needed
	visibleCount := 0
	for _, status := range m.statuses {
		if m.showAll || status.IsInstalled || status.HasConfig {
			visibleCount++
		}
	}
	
	if visibleCount != m.lastFormSize {
		m.buildForm()
	}
}

// Focus implements the focusable interface
func (m *HuhAppSelectorModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Blur implements the focusable interface
func (m *HuhAppSelectorModel) Blur() tea.Cmd {
	m.focused = false
	return nil
}

// IsFocused returns whether the component is focused
func (m *HuhAppSelectorModel) IsFocused() bool {
	return m.focused
}

// SetSize updates the component size
func (m *HuhAppSelectorModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	m.updateFormSize()
	return nil
}

// GetSize returns the current size
func (m *HuhAppSelectorModel) GetSize() (int, int) {
	return m.width, m.height
}

// Bindings returns key bindings for help
func (m *HuhAppSelectorModel) Bindings() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Enter,
		m.keyMap.Toggle,
		m.keyMap.Help,
		m.keyMap.Quit,
	}
}

// GetSelectedApp returns the currently selected app
func (m *HuhAppSelectorModel) GetSelectedApp() string {
	return m.selectedApp
}

// GetApps returns the available apps
func (m *HuhAppSelectorModel) GetApps() []string {
	var apps []string
	for _, status := range m.statuses {
		apps = append(apps, status.Definition.Name)
	}
	return apps
}