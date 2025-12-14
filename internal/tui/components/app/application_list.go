package app

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ApplicationItem represents an application in the list
type ApplicationItem struct {
	name        string
	description string
	status      string
	configPath  string
}

// FilterValue implements list.Item
func (i ApplicationItem) FilterValue() string {
	return i.name + " " + i.description
}

// Title implements list.DefaultItem
func (i ApplicationItem) Title() string {
	return i.name
}

// Description implements list.DefaultItem
func (i ApplicationItem) Description() string {
	if i.status != "" {
		return fmt.Sprintf("%s • %s", i.description, i.status)
	}
	return i.description
}

// ApplicationDelegate provides styling for application items
type ApplicationDelegate struct {
	styles *styles.Styles
}

// NewApplicationDelegate creates a new application delegate
func NewApplicationDelegate() ApplicationDelegate {
	return ApplicationDelegate{
		styles: styles.GetStyles(),
	}
}

// Height implements list.ItemDelegate
func (d ApplicationDelegate) Height() int {
	return 2
}

// Spacing implements list.ItemDelegate
func (d ApplicationDelegate) Spacing() int {
	return 1
}

// Update implements list.ItemDelegate
func (d ApplicationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render implements list.ItemDelegate
func (d ApplicationDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if appItem, ok := item.(ApplicationItem); ok {
		var (
			title, desc string
			s           = &d.styles.ApplicationList
		)

		isSelected := index == m.Index()

		// Title styling
		if isSelected {
			title = s.SelectedTitle.Render(appItem.Title())
		} else {
			title = s.NormalTitle.Render(appItem.Title())
		}

		// Description styling
		if isSelected {
			desc = s.SelectedDesc.Render(appItem.Description())
		} else {
			desc = s.NormalDesc.Render(appItem.Description())
		}

		// Status indicator
		var statusIndicator string
		switch appItem.status {
		case "configured":
			statusIndicator = s.StatusConfigured.Render("●")
		case "needs_config":
			statusIndicator = s.StatusNeedsConfig.Render("○")
		case "error":
			statusIndicator = s.StatusError.Render("!")
		default:
			statusIndicator = s.StatusUnknown.Render("?")
		}

		// Combine elements
		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusIndicator,
			" ",
			title,
		)

		if m.Width() > 0 {
			textwidth := lipgloss.Width(line)
			if textwidth > m.Width() {
				line = lipgloss.NewStyle().
					Width(m.Width()).
					MaxWidth(m.Width()).
					Render(line)
			}
		}

		fmt.Fprint(w, line)
		if desc != "" {
			// Constrain description to available width as well
			constrainedDesc := desc
			if m.Width() > 0 {
				descWidth := lipgloss.Width(desc)
				if descWidth > m.Width()-2 { // Account for "  " prefix
					constrainedDesc = lipgloss.NewStyle().
						Width(m.Width()-2).
						MaxWidth(m.Width()-2).
						Render(desc)
				}
			}
			fmt.Fprintf(w, "\n  %s", constrainedDesc)
		}
	}
}

// ApplicationListModel represents the modern application list
type ApplicationListModel struct {
	list     list.Model
	keyMap   ApplicationKeyMap
	styles   *styles.Styles
	delegate ApplicationDelegate
	width    int
	height   int
}

// ApplicationKeyMap defines key bindings for the application list
type ApplicationKeyMap struct {
	Select  key.Binding
	Refresh key.Binding
	Filter  key.Binding
	Help    key.Binding
}

// DefaultApplicationKeyMap returns default key bindings
func DefaultApplicationKeyMap() ApplicationKeyMap {
	return ApplicationKeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select application"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "F5"),
			key.WithHelp("r/F5", "refresh list"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter applications"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// ShortHelp implements key.Map
func (k ApplicationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Filter, k.Refresh, k.Help}
}

// FullHelp implements key.Map
func (k ApplicationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Filter},
		{k.Refresh, k.Help},
	}
}

// NewApplicationList creates a new application list
func NewApplicationList() *ApplicationListModel {
	delegate := NewApplicationDelegate()

	// Create list with delegate
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "ZeroUI Applications"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Style the list
	l.Styles.Title = styles.GetStyles().ApplicationList.Title
	l.Styles.PaginationStyle = styles.GetStyles().ApplicationList.Pagination
	l.Styles.HelpStyle = styles.GetStyles().ApplicationList.Help
	l.Styles.FilterPrompt = styles.GetStyles().ApplicationList.FilterPrompt
	l.Styles.FilterCursor = styles.GetStyles().ApplicationList.FilterCursor

	return &ApplicationListModel{
		list:     l,
		keyMap:   DefaultApplicationKeyMap(),
		styles:   styles.GetStyles(),
		delegate: delegate,
	}
}

// Init implements tea.Model
func (m *ApplicationListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model with performance optimizations and error handling
func (m *ApplicationListModel) Update(msg tea.Msg) (*ApplicationListModel, tea.Cmd) {
	if m == nil {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed to avoid unnecessary work
		if m.width != msg.Width || m.height != msg.Height {
			m.width = msg.Width
			m.height = msg.Height
			m.list.SetWidth(msg.Width)
			newHeight := max(msg.Height-2, 1) // Reserve space for status, ensure positive
			m.list.SetHeight(newHeight)
		}
		return m, nil

	case tea.KeyMsg:
		// Handle custom key bindings with validation
		switch {
		case key.Matches(msg, m.keyMap.Select):
			if item := m.list.SelectedItem(); item != nil {
				if appItem, ok := item.(ApplicationItem); ok {
					// Validate app name before sending message
					if appItem.name != "" {
						return m, func() tea.Msg {
							return AppSelectedMsg{App: appItem.name}
						}
					}
				}
			}
		case key.Matches(msg, m.keyMap.Refresh):
			// Refresh the application list from registry
			m.RefreshFromRegistry()
			return m, func() tea.Msg {
				return RefreshAppsMsg{}
			}
		}
	case RefreshAppsMsg:
		// Handle refresh message from main model
		m.RefreshFromRegistry()
		return m, nil
	}

	// Update the list with error boundary
	defer func() {
		if r := recover(); r != nil {
			// Log error but don't crash the application
			// Note: Would need logger passed in for proper logging
		}
	}()

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Batch commands only if we have any to avoid unnecessary allocations
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// View implements tea.Model
func (m *ApplicationListModel) View() string {
	var content string
	if len(m.list.Items()) == 0 {
		empty := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(
			"No applications detected. Press 'r' to refresh.\nAdd schemas to apps/ or install plugins.")
		header := styles.GetStyles().ApplicationList.Title.Render("ZeroUI Applications")
		content = lipgloss.JoinVertical(lipgloss.Left, header, "", empty)
	} else {
		content = m.list.View()
	}

	// Ensure the entire content fits within the set width
	if m.width > 0 {
		content = lipgloss.NewStyle().
			Width(m.width).
			Render(content)
	}

	return content
}

// SetApplications updates the list of applications
func (m *ApplicationListModel) SetApplications(apps []ApplicationInfo) {
	items := make([]list.Item, len(apps))
	for i, app := range apps {
		items[i] = ApplicationItem{
			name:        app.Name,
			description: app.Description,
			status:      app.Status,
			configPath:  app.ConfigPath,
		}
	}
	m.list.SetItems(items)
}

// RefreshFromRegistry loads applications from the TUI registry and updates the list
func (m *ApplicationListModel) RefreshFromRegistry() {
	// Get application statuses from registry
	statuses := registry.GetAppStatuses()

	// Convert to ApplicationInfo format
	apps := make([]ApplicationInfo, 0, len(statuses))
	for _, status := range statuses {
		// Only include applications that are installed and have configs
		if status.IsInstalled && status.ConfigExists {
			apps = append(apps, ApplicationInfo{
				Name:        status.Definition.Name,
				Description: fmt.Sprintf("%s application", status.Definition.Category),
				Status:      getStatusString(status),
				ConfigPath:  status.Definition.ConfigPath,
			})
		}
	}

	// Update the list
	m.SetApplications(apps)
}

// getStatusString converts registry status to display string
func getStatusString(status registry.AppStatus) string {
	if status.IsInstalled && status.ConfigExists {
		return "Ready"
	} else if status.IsInstalled {
		return "No Config"
	} else if status.ConfigExists {
		return "Not Installed"
	}
	return "Unknown"
}

// GetSelectedApp returns the currently selected application
func (m *ApplicationListModel) GetSelectedApp() string {
	if item := m.list.SelectedItem(); item != nil {
		if appItem, ok := item.(ApplicationItem); ok {
			return appItem.name
		}
	}
	return ""
}

// SelectedApp is an alias for GetSelectedApp for consistency
func (m *ApplicationListModel) SelectedApp() string {
	return m.GetSelectedApp()
}

// SetSize updates the component size
func (m *ApplicationListModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	m.list.SetWidth(width)
	m.list.SetHeight(height - 2)
	return nil
}

// GetItemCount returns the total number of items in the list
func (m *ApplicationListModel) GetItemCount() int {
	return len(m.list.Items())
}

// Index returns the currently selected item index
func (m *ApplicationListModel) Index() int {
	return m.list.Index()
}

// SelectedItem returns the currently selected item
func (m *ApplicationListModel) SelectedItem() list.Item {
	return m.list.SelectedItem()
}

// IsFiltered returns whether the list is currently being filtered
func (m *ApplicationListModel) IsFiltered() bool {
	return m.list.IsFiltered()
}

// VisibleItems returns the currently visible items (after filtering)
func (m *ApplicationListModel) VisibleItems() []list.Item {
	return m.list.VisibleItems()
}

// ApplicationInfo represents application information
type ApplicationInfo struct {
	Name        string
	Description string
	Status      string
	ConfigPath  string
}

// RefreshAppsMsg requests application list refresh
type RefreshAppsMsg struct{}
