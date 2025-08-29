package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/tui/components/core"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// EnhancedApplicationList is a fully integrated application list component
type EnhancedApplicationList struct {
	*core.BaseComponent
	styles        *styles.Styles
	list          list.Model
	applications  []ApplicationData
	filteredApps  []ApplicationData
	selectedIndex int
	isActive      bool
	isLoading     bool
	error         error
	onSelect      func(app ApplicationData) tea.Cmd
}

// ApplicationData represents application information
type ApplicationData struct {
	Name        string
	DisplayName string
	Description string
	Category    string
	Status      ApplicationStatus
	ConfigPath  string
	Icon        string
}

// ApplicationStatus represents the status of an application
type ApplicationStatus string

const (
	AppStatusReady        ApplicationStatus = "ready"
	AppStatusNotInstalled ApplicationStatus = "not_installed"
	AppStatusNoConfig     ApplicationStatus = "no_config"
	AppStatusError        ApplicationStatus = "error"
)

// NewEnhancedApplicationList creates a new enhanced application list
func NewEnhancedApplicationList() *EnhancedApplicationList {
	styles := styles.GetStyles()

	// Create the list model
	listModel := list.New([]list.Item{}, NewApplicationListDelegate(styles), 0, 0)
	listModel.SetShowStatusBar(false)
	listModel.SetShowHelp(false)
	listModel.SetFilteringEnabled(true)
	listModel.SetShowTitle(false)

	return &EnhancedApplicationList{
		BaseComponent: core.NewBaseComponent("app_list"),
		styles:        styles,
		list:          listModel,
		applications:  []ApplicationData{},
		filteredApps:  []ApplicationData{},
		selectedIndex: 0,
		isActive:      false,
		isLoading:     false,
	}
}

// ID implements UnifiedComponent
func (e *EnhancedApplicationList) ID() string {
	return "app_list"
}

// Title implements UnifiedComponent
func (e *EnhancedApplicationList) Title() string {
	return "Application Manager"
}

// Description implements UnifiedComponent
func (e *EnhancedApplicationList) Description() string {
	return "Manage and configure applications"
}

// IsActive implements UnifiedComponent
func (e *EnhancedApplicationList) IsActive() bool {
	return e.isActive
}

// SetActive implements UnifiedComponent
func (e *EnhancedApplicationList) SetActive(active bool) tea.Cmd {
	e.isActive = active
	if active {
		return e.list.NewStatusMessage("Application list activated")
	}
	return nil
}

// GetState implements UnifiedComponent
func (e *EnhancedApplicationList) GetState() interface{} {
	return map[string]interface{}{
		"selectedIndex": e.selectedIndex,
		"totalCount":    len(e.applications),
		"filteredCount": len(e.filteredApps),
		"isLoading":     e.isLoading,
		"hasError":      e.error != nil,
	}
}

// SetState implements UnifiedComponent
func (e *EnhancedApplicationList) SetState(state interface{}) tea.Cmd {
	if stateMap, ok := state.(map[string]interface{}); ok {
		if selectedIndex, ok := stateMap["selectedIndex"].(int); ok {
			e.selectedIndex = selectedIndex
		}
	}
	return nil
}

// GetData implements UnifiedComponent
func (e *EnhancedApplicationList) GetData() interface{} {
	return e.applications
}

// SetData implements UnifiedComponent
func (e *EnhancedApplicationList) SetData(data interface{}) tea.Cmd {
	if apps, ok := data.([]ApplicationData); ok {
		e.applications = apps
		e.updateListItems()
	}
	return nil
}

// IsValid implements UnifiedComponent
func (e *EnhancedApplicationList) IsValid() bool {
	return e.error == nil && len(e.applications) > 0
}

// Validate implements UnifiedComponent
func (e *EnhancedApplicationList) Validate() []ValidationError {
	var errors []ValidationError

	if e.error != nil {
		errors = append(errors, ValidationError{
			Field:   "applications",
			Message: e.error.Error(),
			Type:    ErrorTypeCustom,
		})
	}

	if len(e.applications) == 0 {
		errors = append(errors, ValidationError{
			Field:   "applications",
			Message: "No applications available",
			Type:    ErrorTypeRequired,
		})
	}

	return errors
}

// OnMount implements UnifiedComponent
func (e *EnhancedApplicationList) OnMount() tea.Cmd {
	return e.loadApplications()
}

// OnUnmount implements UnifiedComponent
func (e *EnhancedApplicationList) OnUnmount() tea.Cmd {
	// Clean up resources
	e.applications = nil
	e.filteredApps = nil
	return nil
}

// Update implements tea.Model
func (e *EnhancedApplicationList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.SetSize(msg.Width, msg.Height)
		e.list.SetSize(msg.Width, msg.Height-4) // Leave space for header/footer

	case tea.KeyMsg:
		if !e.isActive {
			return e, nil
		}

		switch msg.String() {
		case "enter":
			if e.selectedIndex >= 0 && e.selectedIndex < len(e.filteredApps) {
				selectedApp := e.filteredApps[e.selectedIndex]
				if e.onSelect != nil {
					cmd = e.onSelect(selectedApp)
				}
			}
		}

		// Update the underlying list
		var listCmd tea.Cmd
		e.list, listCmd = e.list.Update(msg)
		cmd = tea.Batch(cmd, listCmd)

		// Update our selected index
		e.selectedIndex = e.list.Index()

	case ApplicationsLoadedMsg:
		e.applications = msg.Applications
		e.isLoading = false
		e.updateListItems()

	case ApplicationErrorMsg:
		e.error = msg.Error
		e.isLoading = false
	}

	return e, cmd
}

// View implements tea.Model
func (e *EnhancedApplicationList) View() string {
	if e.error != nil {
		return e.renderError()
	}

	if e.isLoading {
		return e.renderLoading()
	}

	if len(e.applications) == 0 {
		return e.renderEmpty()
	}

	// Render the list with header and footer
	header := e.styles.Title.Render("🎯 ZeroUI - Application Manager")
	listView := e.list.View()
	footer := e.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		"",
		listView,
		"",
		footer,
	)
}

// KeyBindings implements UnifiedComponent
func (e *EnhancedApplicationList) KeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
	}
}

// HandleKey implements UnifiedComponent
func (e *EnhancedApplicationList) HandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return e.Update(msg)
}

// SetSize implements core.Sizeable
func (e *EnhancedApplicationList) SetSize(width, height int) tea.Cmd {
	e.BaseComponent.SetSize(width, height)
	e.list.SetSize(width, height-4)
	return nil
}

// SetOnSelect sets the callback for when an application is selected
func (e *EnhancedApplicationList) SetOnSelect(callback func(app ApplicationData) tea.Cmd) {
	e.onSelect = callback
}

// LoadApplications loads applications from the registry
func (e *EnhancedApplicationList) LoadApplications() tea.Cmd {
	return e.loadApplications()
}

// GetSelectedApplication returns the currently selected application
func (e *EnhancedApplicationList) GetSelectedApplication() (ApplicationData, bool) {
	if e.selectedIndex >= 0 && e.selectedIndex < len(e.filteredApps) {
		return e.filteredApps[e.selectedIndex], true
	}
	return ApplicationData{}, false
}

// GetApplicationCount returns the total number of applications
func (e *EnhancedApplicationList) GetApplicationCount() int {
	return len(e.applications)
}

// GetFilteredCount returns the number of filtered applications
func (e *EnhancedApplicationList) GetFilteredCount() int {
	return len(e.filteredApps)
}

// Helper methods

func (e *EnhancedApplicationList) loadApplications() tea.Cmd {
	e.isLoading = true

	return func() tea.Msg {
		// Get applications from registry
		statuses := registry.GetAppStatuses()

		applications := make([]ApplicationData, 0, len(statuses))

		for _, status := range statuses {
			app := ApplicationData{
				Name:        status.Definition.Name,
				DisplayName: status.Definition.Name,
				Description: fmt.Sprintf("%s application", status.Definition.Category),
				Category:    status.Definition.Category,
				ConfigPath:  status.Definition.ConfigPath,
				Icon:        status.Definition.Logo,
			}

			// Determine status
			if status.IsInstalled && status.ConfigExists {
				app.Status = AppStatusReady
			} else if status.IsInstalled {
				app.Status = AppStatusNoConfig
			} else {
				app.Status = AppStatusNotInstalled
			}

			applications = append(applications, app)
		}

		return ApplicationsLoadedMsg{Applications: applications}
	}
}

func (e *EnhancedApplicationList) updateListItems() {
	items := make([]list.Item, len(e.applications))

	for i, app := range e.applications {
		items[i] = ApplicationListItem{
			data:   app,
			styles: e.styles,
		}
	}

	e.list.SetItems(items)
	e.filteredApps = e.applications
}

func (e *EnhancedApplicationList) renderError() string {
	errorMsg := e.styles.Error.Render(fmt.Sprintf("❌ Error: %v", e.error))
	footer := e.styles.Help.Render("Press 'r' to retry or 'q' to quit")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		errorMsg,
		"",
		footer,
	)
}

func (e *EnhancedApplicationList) renderLoading() string {
	loadingMsg := e.styles.Info.Render("⏳ Loading applications...")
	footer := e.styles.Help.Render("Please wait...")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		loadingMsg,
		"",
		footer,
	)
}

func (e *EnhancedApplicationList) renderEmpty() string {
	emptyMsg := e.styles.Muted.Render("📭 No applications found")
	footer := e.styles.Help.Render("Press 'r' to refresh or 'q' to quit")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		emptyMsg,
		"",
		footer,
	)
}

func (e *EnhancedApplicationList) renderFooter() string {
	var status string
	if e.list.IsFiltered() {
		status = fmt.Sprintf("↑/↓: Navigate • Enter: Select • /: Filter (%d/%d) • ?: Help • q: Quit",
			len(e.filteredApps), len(e.applications))
	} else {
		status = fmt.Sprintf("↑/↓: Navigate • Enter: Select • /: Filter • ?: Help • q: Quit (%d)",
			len(e.applications))
	}

	return e.styles.Help.Render(status)
}

// ApplicationListItem represents an application in the list
type ApplicationListItem struct {
	data   ApplicationData
	styles *styles.Styles
}

// FilterValue implements list.Item
func (a ApplicationListItem) FilterValue() string {
	return a.data.Name + " " + a.data.Description + " " + a.data.Category
}

// Title implements list.Item
func (a ApplicationListItem) Title() string {
	return fmt.Sprintf("%s %s", a.data.Icon, a.data.Name)
}

// Description implements list.Item
func (a ApplicationListItem) Description() string {
	var status string
	switch a.data.Status {
	case AppStatusReady:
		status = a.styles.ApplicationList.StatusConfigured.Render("Ready")
	case AppStatusNoConfig:
		status = a.styles.ApplicationList.StatusNeedsConfig.Render("Needs Config")
	case AppStatusNotInstalled:
		status = a.styles.ApplicationList.StatusUnknown.Render("Not Installed")
	case AppStatusError:
		status = a.styles.ApplicationList.StatusError.Render("Error")
	}

	return fmt.Sprintf("%s • %s", a.data.Description, status)
}

// Messages

type ApplicationsLoadedMsg struct {
	Applications []ApplicationData
}

type ApplicationErrorMsg struct {
	Error error
}

// ApplicationListDelegate provides styling for the application list
type ApplicationListDelegate struct {
	styles *styles.Styles
}

func NewApplicationListDelegate(styles *styles.Styles) ApplicationListDelegate {
	return ApplicationListDelegate{styles: styles}
}

func (d ApplicationListDelegate) Height() int {
	return 2
}

func (d ApplicationListDelegate) Spacing() int {
	return 1
}

func (d ApplicationListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ApplicationListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if appItem, ok := item.(ApplicationListItem); ok {
		var title, desc string
		s := &d.styles.ApplicationList

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

		fmt.Fprintf(w, "%s\n%s", title, desc)
	}
}
