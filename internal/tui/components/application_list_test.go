package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationList_NewApplicationList(t *testing.T) {
	model := NewApplicationList()

	require.NotNil(t, model)
	assert.NotNil(t, model.list)
	assert.Equal(t, "ZeroUI Applications", model.list.Title)
	assert.True(t, model.list.FilteringEnabled())
}

func TestApplicationList_SetApplications(t *testing.T) {
	model := NewApplicationList()

	apps := []ApplicationInfo{
		{Name: "ghostty", Description: "Fast terminal emulator", Status: "configured", ConfigPath: "/path/to/ghostty"},
		{Name: "vscode", Description: "Code editor", Status: "needs_config", ConfigPath: "/path/to/vscode"},
		{Name: "alacritty", Description: "Terminal emulator", Status: "error", ConfigPath: "/path/to/alacritty"},
	}

	model.SetApplications(apps)

	// Verify items were set
	items := model.list.Items()
	assert.Len(t, items, 3)

	// Verify first item
	if len(items) > 0 {
		appItem, ok := items[0].(ApplicationItem)
		require.True(t, ok)
		assert.Equal(t, "ghostty", appItem.name)
		assert.Equal(t, "Fast terminal emulator", appItem.description)
		assert.Equal(t, "configured", appItem.status)
	}
}

func TestApplicationList_GetSelectedApp(t *testing.T) {
	model := NewApplicationList()

	// Initially no selection
	assert.Equal(t, "", model.GetSelectedApp())

	// Set applications
	apps := []ApplicationInfo{
		{Name: "ghostty", Description: "Fast terminal", Status: "configured"},
		{Name: "vscode", Description: "Code editor", Status: "needs_config"},
	}
	model.SetApplications(apps)

	// After setting items, the list should have a default selection (first item)
	if len(model.list.Items()) > 0 {
		selectedApp := model.GetSelectedApp()
		// The list component automatically selects the first item
		assert.Equal(t, "ghostty", selectedApp)
	}
}

func TestApplicationList_Update(t *testing.T) {
	model := NewApplicationList()

	// Test window size message
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(sizeMsg)

	require.NotNil(t, updatedModel)
	assert.Equal(t, 100, updatedModel.width)
	assert.Equal(t, 50, updatedModel.height)
	assert.Equal(t, 100, updatedModel.list.Width())
	assert.Equal(t, 48, updatedModel.list.Height()) // Height - 2 for status
	assert.Nil(t, cmd)
}

func TestApplicationList_KeyNavigation(t *testing.T) {
	model := NewApplicationList()

	// Set test applications
	apps := []ApplicationInfo{
		{Name: "app1", Description: "First app", Status: "configured"},
		{Name: "app2", Description: "Second app", Status: "needs_config"},
	}
	model.SetApplications(apps)

	// Test refresh key
	refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := model.Update(refreshMsg)

	require.NotNil(t, updatedModel)
	require.NotNil(t, cmd)

	// Execute the command to get the message
	if cmd != nil {
		msg := cmd()
		refreshMsg, ok := msg.(RefreshAppsMsg)
		assert.True(t, ok)
		assert.IsType(t, RefreshAppsMsg{}, refreshMsg)
	}
}

func TestApplicationList_SelectAction(t *testing.T) {
	model := NewApplicationList()

	// Set test applications
	apps := []ApplicationInfo{
		{Name: "test-app", Description: "Test application", Status: "configured"},
	}
	model.SetApplications(apps)

	// Test enter key (select)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(enterMsg)

	require.NotNil(t, updatedModel)

	if cmd != nil {
		msg := cmd()
		if appSelectedMsg, ok := msg.(AppSelectedMsg); ok {
			assert.Equal(t, "test-app", appSelectedMsg.App)
		}
	}
}

func TestApplicationList_SetSize(t *testing.T) {
	model := NewApplicationList()

	cmd := model.SetSize(120, 60)

	assert.Equal(t, 120, model.width)
	assert.Equal(t, 60, model.height)
	assert.Equal(t, 120, model.list.Width())
	assert.Equal(t, 58, model.list.Height())
	assert.Nil(t, cmd) // SetSize returns nil command
}

func TestApplicationItem_Interface(t *testing.T) {
	item := ApplicationItem{
		name:        "test-app",
		description: "Test application for unit testing",
		status:      "configured",
		configPath:  "/test/path",
	}

	// Test list.Item interface
	assert.Equal(t, "test-app Test application for unit testing", item.FilterValue())

	// Test list.DefaultItem interface methods
	assert.Equal(t, "test-app", item.Title())
	assert.Equal(t, "Test application for unit testing • configured", item.Description())
}

func TestApplicationItem_DescriptionWithoutStatus(t *testing.T) {
	item := ApplicationItem{
		name:        "test-app",
		description: "Test application",
		status:      "", // No status
		configPath:  "/test/path",
	}

	assert.Equal(t, "Test application", item.Description())
}

func TestApplicationDelegate_Properties(t *testing.T) {
	delegate := NewApplicationDelegate()

	// Test delegate properties
	assert.Equal(t, 2, delegate.Height())
	assert.Equal(t, 1, delegate.Spacing())

	// Test Update returns nil (no special handling needed)
	cmd := delegate.Update(tea.KeyMsg{}, nil)
	assert.Nil(t, cmd)
}

func TestApplicationKeyMap_ShortHelp(t *testing.T) {
	keyMap := DefaultApplicationKeyMap()

	shortHelp := keyMap.ShortHelp()
	assert.Len(t, shortHelp, 4) // select, refresh, filter, help

	// Verify key bindings exist
	assert.NotNil(t, keyMap.Select)
	assert.NotNil(t, keyMap.Refresh)
	assert.NotNil(t, keyMap.Filter)
	assert.NotNil(t, keyMap.Help)
}

func TestApplicationKeyMap_FullHelp(t *testing.T) {
	keyMap := DefaultApplicationKeyMap()

	fullHelp := keyMap.FullHelp()
	assert.Len(t, fullHelp, 2) // Two rows of help

	// Each row should have key bindings
	for _, row := range fullHelp {
		assert.NotEmpty(t, row)
	}
}
