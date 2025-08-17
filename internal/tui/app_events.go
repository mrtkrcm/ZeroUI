package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Performance tracking
	m.frameCount++

	// Set up panic recovery for update operations
	defer func() {
		if r := recover(); r != nil {
			m.logger.LogPanic(r, "update_panic")
			m.err = fmt.Errorf("update panic: %v", r)
			// Try to recover to a safe state
			m.state = ListView
			m.showingHelp = false
		}
	}()

	// Type switch for message handling
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case util.ErrorMsg:
		return m.handleError(msg)

	case util.SuccessMsg:
		return m.handleSuccess(msg)

	case util.ShowInfoMsg:
		m.logger.Info("Info message", "message", string(msg))
		return m, nil

	case components.AppSelectedMsg:
		return m.handleAppSelection(msg)

	case components.ConfigSavedMsg:
		return m.handleConfigSaved(msg)

	case components.PresetAppliedMsg:
		return m.handlePresetApplied(msg)

	// Explicitly handle refresh messages at the model level so we can
	// debounce and trigger component refresh logic in a controlled way.
	// Tests and other code send `RefreshAppsMsg{}` (alias to components.RefreshAppsMsg)
	// and expect the model to perform the debounced refresh.
	case RefreshAppsMsg:
		m.HandleRefreshApps()
		return m, nil

	default:
		// Delegate to state-specific handlers
		return m.handleStateUpdate(msg)
	}
}

// handleWindowSize handles window resize events
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.logger.Debug("Window resized", "width", msg.Width, "height", msg.Height)

	// Track significant size changes
	significantChange := abs(m.width-msg.Width) > 10 || abs(m.height-msg.Height) > 5

	m.width = msg.Width
	m.height = msg.Height

	// Invalidate cache on significant resize
	if significantChange {
		m.invalidateCache()
	}

	// Update component sizes
	return m, m.updateComponentSizes()
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.logger.Debug("Key pressed", "key", msg.String())

	// Handle global keys first
	if cmd := m.handleGlobalKeys(msg); cmd != nil {
		return m, cmd
	}

	// Handle state-specific keys
	if cmd := m.handleStateKeys(msg); cmd != nil {
		return m, cmd
	}

	// Delegate to component handlers
	return m.handleComponentKeys(msg)
}

// handleGlobalKeys handles keys that work across all states
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		if m.state == FormView && m.configForm != nil && m.configForm.HasUnsavedChanges() {
			m.logger.Info("Quit requested with unsaved changes")
			// TODO: Show confirmation dialog
			return tea.Quit
		}
		m.logger.Info("Quit requested")
		return tea.Quit

	case key.Matches(msg, m.keyMap.Help):
		m.showingHelp = !m.showingHelp
		m.logger.Debug("Help toggled", "showing", m.showingHelp)
		m.invalidateCache()
		if m.showingHelp {
			m.state = HelpView
		} else {
			m.handleBack()
		}
		return nil

	case key.Matches(msg, m.keyMap.Back):
		return m.handleBack()

	case msg.String() == "ctrl+c":
		m.logger.Info("Force quit requested")
		return tea.Quit
	}

	return nil
}

// handleStateKeys handles keys specific to the current state
func (m *Model) handleStateKeys(msg tea.KeyMsg) tea.Cmd {
	switch m.state {
	case ListView:
		switch {
		case key.Matches(msg, m.keyMap.Select):
			if m.appList != nil {
				selected := m.appList.SelectedApp()
				if selected != "" {
					m.logger.Info("App selected", "app", selected)
					return m.handleAppSelected(selected)
				}
			}
		}

	case FormView:
		switch {
		case key.Matches(msg, m.keyMap.Save):
			if m.configForm != nil && m.configForm.IsValid() {
				values := m.configForm.GetValues()
				m.logger.Info("Saving configuration", "app", m.currentApp)
				return m.saveConfiguration(m.currentApp, values)
			}
			return func() tea.Msg {
				return util.ErrorMsg{Err: fmt.Errorf("form validation failed")}
			}

		case msg.String() == "p":
			// Show preset selector if available
			if m.presetSel != nil {
				m.logger.Debug("Opening preset selector")
				// Show presets for the current app and ensure UI updates so the
				// selector becomes visible in tests and interactive runs.
				m.presetSel.Show(m.currentApp)
				m.invalidateCache()
				// If we're not already in the FormView, switch to it so presets
				// are presented in the proper context.
				if m.state != FormView {
					m.state = FormView
				}
				return nil
			}
		}
	}

	return nil
}

// handleComponentKeys delegates key handling to components
func (m *Model) handleComponentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ListView:
		if m.appList != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.appList.Update(msg)
				},
				"appList",
			)
			if updated, ok := updatedModel.(*components.ApplicationListModel); ok {
				m.appList = updated
			}
			return m, cmd
		}

	case FormView:
		if m.configForm != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.configForm.Update(msg)
				},
				"configForm",
			)
			if updated, ok := updatedModel.(*components.HuhConfigFormModel); ok {
				m.configForm = updated
			}
			return m, cmd
		}

	case HelpView:
		if m.helpSystem != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.helpSystem.Update(msg)
				},
				"helpSystem",
			)
			if updated, ok := updatedModel.(*components.GlamourHelpModel); ok {
				m.helpSystem = updated
			}
			return m, cmd
		}
	}

	return m, nil
}

// handleStateUpdate handles non-key messages based on state
func (m *Model) handleStateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ListView:
		if m.appList != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.appList.Update(msg)
				},
				"appList",
			)
			if updated, ok := updatedModel.(*components.ApplicationListModel); ok {
				m.appList = updated
			}
			return m, cmd
		}

	case FormView:
		if m.configForm != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.configForm.Update(msg)
				},
				"configForm",
			)
			if updated, ok := updatedModel.(*components.HuhConfigFormModel); ok {
				m.configForm = updated
			}
			return m, cmd
		}

	case HelpView:
		if m.helpSystem != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.helpSystem.Update(msg)
				},
				"helpSystem",
			)
			if updated, ok := updatedModel.(*components.GlamourHelpModel); ok {
				m.helpSystem = updated
			}
			return m, cmd
		}
	}

	return m, nil
}

// Event handlers

// handleBack handles going back to previous state
func (m *Model) handleBack() tea.Cmd {
	m.logger.Debug("Back navigation", "from_state", m.state)

	switch m.state {
	case FormView:
		m.state = ListView
		m.currentApp = ""
		m.configForm = nil
		m.presetSel = nil
		m.invalidateCache()

	case HelpView:
		if m.currentApp != "" {
			m.state = FormView
		} else {
			m.state = ListView
		}
		m.showingHelp = false
		m.invalidateCache()

	case ListView:
		// At root, treat back as a quit request (consistent with typical TUI behavior)
		m.logger.Info("Quit requested")
		return tea.Quit
	}

	return nil
}

// handleAppSelected handles app selection
func (m *Model) handleAppSelected(appName string) tea.Cmd {
	m.logger.Info("Loading app configuration", "app", appName)

	m.currentApp = appName
	if err := m.loadAppConfigForForm(appName); err != nil {
		m.logger.LogError(err, "app_load", "app", appName)
		return func() tea.Msg {
			return util.ErrorMsg{Err: fmt.Errorf("failed to load %s: %w", appName, err)}
		}
	}

	m.state = FormView
	m.invalidateCache()
	return nil
}

// handleAppSelection handles app selection message
func (m *Model) handleAppSelection(msg components.AppSelectedMsg) (tea.Model, tea.Cmd) {
	return m, m.handleAppSelected(msg.App)
}

// handleConfigSaved handles configuration saved message
func (m *Model) handleConfigSaved(msg components.ConfigSavedMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Configuration saved", "app", msg.AppName)
	// Return to app list
	m.state = ListView
	m.currentApp = ""
	m.configForm = nil
	m.invalidateCache()

	return m, func() tea.Msg {
		return util.SuccessMsg{
			Title: "Success",
			Body:  fmt.Sprintf("%s configuration saved", msg.AppName),
		}
	}
}

// handlePresetApplied handles preset application
func (m *Model) handlePresetApplied(msg components.PresetAppliedMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Preset applied", "preset", msg.PresetName)

	if m.configForm != nil {
		// Apply preset values to form
		m.configForm.ApplyPreset(msg.Values)
		m.invalidateCache()
	}

	return m, nil
}

// handleError handles error messages
func (m *Model) handleError(msg util.ErrorMsg) (tea.Model, tea.Cmd) {
	m.logger.LogError(msg.Err, "ui_error")
	m.err = msg.Err
	return m, nil
}

// handleSuccess handles success messages
func (m *Model) handleSuccess(msg util.SuccessMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Success", "title", msg.Title, "body", msg.Body)
	// Could show a toast notification here
	return m, nil
}

// safeUpdateComponent safely updates a component with panic recovery
func (m *Model) safeUpdateComponent(updateFn func() (interface{}, tea.Cmd), componentName string) (interface{}, tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.LogPanic(r, "component_update_panic", "component", componentName)
		}
	}()

	return updateFn()
}
