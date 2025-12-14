package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	app "github.com/mrtkrcm/ZeroUI/internal/tui/components/app"
	core "github.com/mrtkrcm/ZeroUI/internal/tui/components/core"
	display "github.com/mrtkrcm/ZeroUI/internal/tui/components/display"
	forms "github.com/mrtkrcm/ZeroUI/internal/tui/components/forms"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/mrtkrcm/ZeroUI/internal/tui/util"
)

// EventBatchMsg represents a batch of events to be processed
type EventBatchMsg struct {
	Events []tea.Msg
}

// EventBatcher handles batching of events for better performance
type EventBatcher struct {
	events      chan tea.Msg
	batchSize   int
	timeout     time.Duration
	maxWaitTime time.Duration
	mu          sync.RWMutex
}

// NewEventBatcher creates a new event batcher
func NewEventBatcher() *EventBatcher {
	return &EventBatcher{
		events:      make(chan tea.Msg, 100), // Buffer for 100 events
		batchSize:   10,                      // Process in batches of 10
		timeout:     50 * time.Millisecond,   // 50ms batching window
		maxWaitTime: 200 * time.Millisecond,  // Max wait time
	}
}

// ProcessEvent adds an event to the batch queue
func (eb *EventBatcher) ProcessEvent(event tea.Msg) {
	select {
	case eb.events <- event:
		// Event queued successfully
	default:
		// Queue is full, process immediately
		// This prevents blocking when batcher is overwhelmed
	}
}

// ProcessEvents returns a command that processes batched events
func (eb *EventBatcher) ProcessEvents() tea.Cmd {
	return func() tea.Msg {
		var batch []tea.Msg
		timer := time.NewTimer(eb.timeout)
		defer timer.Stop()

		// Collect events up to batch size or timeout
		for len(batch) < eb.batchSize {
			select {
			case event := <-eb.events:
				batch = append(batch, event)
			case <-timer.C:
				goto processBatch
			}
		}

	processBatch:
		if len(batch) == 0 {
			return nil // No events to process
		}

		if len(batch) == 1 {
			// Single event, return directly
			return batch[0]
		}

		// Return batch for processing
		return EventBatchMsg{Events: batch}
	}
}

// StartBatching returns a command to start the batching process
func (eb *EventBatcher) StartBatching() tea.Cmd {
	return eb.ProcessEvents()
}

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

	case app.AppSelectedMsg:
		return m.handleAppSelection(msg)

	case core.ConfigSavedMsg:
		return m.handleConfigSaved(msg)

	case core.PresetAppliedMsg:
		return m.handlePresetApplied(msg)

	case app.ScanProgressMsg:
		return m.handleScanProgress(msg)

	case app.ScanCompleteMsg:
		return m.handleScanComplete(msg)

	// Handle event batching for better performance
	case EventBatchMsg:
		return m.handleEventBatch(msg)

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
	// Ignore invalid sizes
	if msg.Width <= 0 || msg.Height <= 0 {
		return m, nil
	}

	m.logger.Debug("Window resized", "width", msg.Width, "height", msg.Height)

	// Track significant size changes
	significantChange := abs(m.width-msg.Width) > 10 || abs(m.height-msg.Height) > 5

	prevWidth := m.width
	prevHeight := m.height
	m.width = msg.Width
	m.height = msg.Height

	// Only invalidate cache and update if there was a real change
	if prevWidth != m.width || prevHeight != m.height {
		if significantChange {
			m.invalidateCache()
		}
		// Update component sizes
		return m, m.updateComponentSizes()
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.logger.Debug("Key pressed", "key", msg.String())

	// Handle dialog keys first if dialog is visible
	if m.confirmDialog != nil && m.confirmDialog.IsVisible() {
		updatedDialog, cmd := m.confirmDialog.Update(msg)
		m.confirmDialog = updatedDialog
		return m, cmd
	}

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
		hasUnsavedChanges := false
		if m.state == FormView && m.configEditor != nil {
			hasUnsavedChanges = m.configEditor.HasUnsavedChanges()
		}

		if hasUnsavedChanges {
			m.logger.Info("Quit requested with unsaved changes")
			m.confirmDialog.Show()
			return nil
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

	case key.Matches(msg, m.keyMap.ThemeCycle):
		// Cycle through available themes
		newTheme := styles.CycleTheme()
		m.theme = &newTheme
		m.styles = m.theme.BuildStyles()
		m.invalidateCache()
		m.logger.Info("Theme cycled", "new_theme", styles.GetCurrentThemeName())
		return nil

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
			if m.configEditor != nil && m.configEditor.IsValid() {
				values := m.configEditor.GetValues()
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
			if updated, ok := updatedModel.(*app.ApplicationListModel); ok {
				m.appList = updated
			}
			return m, cmd
		}

	case FormView:
		if m.configEditor != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.configEditor.Update(msg)
				},
				"configEditor",
			)
			if updated, ok := updatedModel.(*forms.EnhancedConfigModel); ok {
				m.configEditor = updated
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
			if updated, ok := updatedModel.(*display.GlamourHelpModel); ok {
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
			if updated, ok := updatedModel.(*app.ApplicationListModel); ok {
				m.appList = updated
			}
			return m, cmd
		}

	case FormView:
		if m.configEditor != nil {
			updatedModel, cmd := m.safeUpdateComponent(
				func() (interface{}, tea.Cmd) {
					return m.configEditor.Update(msg)
				},
				"configEditor",
			)
			if updated, ok := updatedModel.(*forms.EnhancedConfigModel); ok {
				m.configEditor = updated
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
			if updated, ok := updatedModel.(*display.GlamourHelpModel); ok {
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

	// Clear any error state when navigating
	m.err = nil

	switch m.state {
	case FormView:
		// Clean transition back to list
		m.state = ListView
		m.currentApp = ""
		m.configEditor = nil
		m.presetSel = nil
		m.invalidateCache()
		// Ensure list is refreshed
		if m.appList != nil {
			// Trigger refresh via message
			m.HandleRefreshApps()
		}

	case HelpView:
		// Navigate to appropriate previous state
		if m.currentApp != "" {
			m.state = FormView
		} else {
			m.state = ListView
		}
		m.showingHelp = false
		m.invalidateCache()

	case ProgressView:
		// Cancel any ongoing operation and return to list
		m.state = ListView
		m.isLoading = false
		m.invalidateCache()

	case ListView:
		// At root, treat back as a quit request (consistent with typical TUI behavior)
		m.logger.Info("Quit requested")
		return tea.Quit

	default:
		// Fallback to list view for any undefined state
		m.state = ListView
		m.invalidateCache()
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
func (m *Model) handleAppSelection(msg app.AppSelectedMsg) (tea.Model, tea.Cmd) {
	return m, m.handleAppSelected(msg.App)
}

// handleConfigSaved handles configuration saved message
func (m *Model) handleConfigSaved(msg core.ConfigSavedMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Configuration saved", "app", msg.AppName)
	// Return to app list
	m.state = ListView
	m.currentApp = ""
	m.configEditor = nil
	m.invalidateCache()

	return m, func() tea.Msg {
		return util.SuccessMsg{
			Title: "Success",
			Body:  fmt.Sprintf("%s configuration saved", msg.AppName),
		}
	}
}

// handlePresetApplied handles preset application
func (m *Model) handlePresetApplied(msg core.PresetAppliedMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Preset applied", "preset", msg.PresetName)

	// For now, presets are not integrated with tabbed config
	// This could be enhanced later if needed
	m.invalidateCache()

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

// handleScanProgress handles scan progress updates
func (m *Model) handleScanProgress(msg app.ScanProgressMsg) (tea.Model, tea.Cmd) {
	// Update scanner
	if m.appScanner != nil {
		updated, cmd := m.appScanner.Update(msg)
		m.appScanner = updated
		return m, cmd
	}
	return m, nil
}

// handleScanComplete handles scan completion
func (m *Model) handleScanComplete(msg app.ScanCompleteMsg) (tea.Model, tea.Cmd) {
	m.logger.Info("Application scan complete", "apps", len(msg.Apps))

	// Update scanner
	if m.appScanner != nil {
		updated, _ := m.appScanner.Update(msg)
		m.appScanner = updated
	}

	// Update app list with scan results
	if m.appList != nil {
		// NOTE: App list status updates not yet implemented - scan completion is handled via ScanCompleteMsg
		// For now, just mark loading as complete
	}

	// Transition to list view
	m.state = ListView
	m.isLoading = false
	m.invalidateCache()

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

// handleEventBatch processes a batch of events for better performance
func (m *Model) handleEventBatch(msg EventBatchMsg) (tea.Model, tea.Cmd) {
	var lastCmd tea.Cmd

	m.logger.Debug("Processing event batch", "count", len(msg.Events))

	// Process each event in the batch
	for i, event := range msg.Events {
		isLastEvent := i == len(msg.Events)-1

		// Process the event (recursive call to Update, but without batching to prevent infinite loop)
		updatedModel, cmd := m.processSingleEvent(event)

		// Update model state
		if updated, ok := updatedModel.(*Model); ok {
			*m = *updated
		}

		// Keep the last command
		if isLastEvent {
			lastCmd = cmd
		}
	}

	// Return the last command and continue batching
	return m, tea.Batch(lastCmd, m.eventBatcher.StartBatching())
}

// processSingleEvent handles a single event without batching (helper for handleEventBatch)
func (m *Model) processSingleEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case app.AppSelectedMsg:
		return m.handleAppSelection(msg)
	case core.ConfigSavedMsg:
		return m.handleConfigSaved(msg)
	case core.PresetAppliedMsg:
		return m.handlePresetApplied(msg)
	case app.ScanProgressMsg:
		return m.handleScanProgress(msg)
	case app.ScanCompleteMsg:
		return m.handleScanComplete(msg)
	case RefreshAppsMsg:
		m.HandleRefreshApps()
		return m, nil
	default:
		// Delegate to state-specific handlers
		return m.handleStateUpdate(msg)
	}
}

// ValidateEventBatching returns true if event batching is properly initialized
func (m *Model) ValidateEventBatching() bool {
	return m.eventBatcher != nil
}
