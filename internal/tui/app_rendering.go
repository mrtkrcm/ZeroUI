package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view
func (m *Model) View() string {
	// Early return for uninitialized state
	if m.width == 0 || m.height == 0 {
		return ""
	}
	
	// Check for cached view first (performance optimization)
	if cached, ok := m.getCachedView(); ok {
		return cached
	}

	// Track render time for performance monitoring
	startTime := time.Now()
	defer func() {
		renderTime := time.Since(startTime)
		if renderTime > 50*time.Millisecond {
			m.logger.Warn("Slow render detected",
				"duration_ms", renderTime.Milliseconds(),
				"state", m.state)
		}
		m.lastRenderTime = startTime
	}()

	// Handle loading state
	if m.isLoading {
		return m.renderLoadingState()
	}

	// Handle error state
	if m.err != nil {
		return m.renderError()
	}

	// Render based on current state with panic recovery
	var content string
	switch m.state {
	case ListView:
		content = m.safeViewRender(m.renderListView, "ListView")
	case FormView:
		content = m.safeViewRender(m.renderFormView, "FormView")
	case HelpView:
		content = m.safeViewRender(m.renderHelpView, "HelpView")
	case ProgressView:
		content = m.safeViewRender(m.renderProgressView, "ProgressView")
	default:
		content = m.renderFallbackView()
	}

	// If a transient status/toast is active, render it as a simple overlay line
	// so tests and callers can observe transient status messages.
	if m.isStatusActive() {
		// Build a concise status line. Include a numeric level prefix when present.
		statusLine := m.statusText
		if m.statusLevel != 0 {
			statusLine = fmt.Sprintf("[%d] %s", m.statusLevel, statusLine)
		}
		// Prepend the status line so it's immediately visible at the top of the view.
		content = statusLine + "\n" + content
	}

	// Return the styled content directly for proper terminal rendering
	// Cache for performance but don't strip ANSI codes as that breaks rendering
	m.cacheView(content)
	return content
}

// renderListView renders the application list view with enhanced UI integration
func (m *Model) renderListView() string {
	// Initialize UI manager if needed
	if m.uiManager != nil && !m.uiManager.IsInitialized() {
		m.uiManager.Initialize(m.width, m.height)
	}

	if m.appList == nil {
		if m.uiManager != nil {
			return m.uiManager.CreateErrorMessage("Application list not initialized")
		}
		return m.styles.Error.Render("Application list not initialized")
	}

	// Get the list view with proper styling
	listView := m.appList.View()

	// Enhanced header with status information using UI manager
	appCount := m.appList.GetItemCount()
	selectedIndex := m.appList.Index()
	var statusInfo string
	if selectedIndex >= 0 && selectedIndex < appCount {
		if item := m.appList.SelectedItem(); item != nil {
			// Extract app name from the item
			if appItem, ok := item.(interface{ Title() string }); ok {
				appName := appItem.Title()
				statusInfo = fmt.Sprintf(" • Selected: %s", appName)
			}
		}
	}

	var header string
	if m.uiManager != nil {
		header = m.uiManager.CreateHeader(
			fmt.Sprintf("🎯 ZeroUI - Application Manager (%d apps)%s", appCount, statusInfo),
			"")
	} else {
		headerText := fmt.Sprintf("🎯 ZeroUI - Application Manager (%d apps)%s", appCount, statusInfo)
		header = m.styles.Title.Render(headerText)
	}

	// Enhanced footer with contextual information using UI manager
	var footer string
	if m.appList.IsFiltered() {
		filteredCount := len(m.appList.VisibleItems())
		footerText := fmt.Sprintf("↑/↓: Navigate • Enter: Select • /: Clear Filter (%d/%d) • ?: Help • q: Quit",
			filteredCount, appCount)
		if m.uiManager != nil {
			footer = m.uiManager.CreateFooter(footerText, "", "")
		} else {
			footer = m.styles.Help.Render(footerText)
		}
	} else {
		footerText := fmt.Sprintf("↑/↓: Navigate • Enter: Select • /: Filter • ?: Help • q: Quit (%d)", appCount)
		if m.uiManager != nil {
			footer = m.uiManager.CreateFooter(footerText, "", "")
		} else {
			footer = m.styles.Help.Render(footerText)
		}
	}

	// Status indicators for screenshot integration
	var statusIndicator string
	if m.screenshotComp != nil && m.screenshotComp.IsCapturing() {
		if m.uiManager != nil {
			statusIndicator = m.uiManager.CreateInfoMessage("Capturing screenshot...")
		} else {
			statusIndicator = m.styles.Info.Render("📸 Capturing screenshot...")
		}
	}

	// Combine all elements with improved spacing and layout
	elements := []string{header, ""}

	if statusIndicator != "" {
		elements = append(elements, statusIndicator, "")
	}

	elements = append(elements, listView, "", footer)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		elements...,
	)

	// Ensure proper width constraints and centering
	finalContent := lipgloss.NewStyle().
		MaxWidth(m.width).
		Align(lipgloss.Left).
		Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Top,
		finalContent,
		lipgloss.WithWhitespaceChars("·"),
		lipgloss.WithWhitespaceForeground(m.styles.Muted.GetForeground()),
	)
}

// renderFormView renders the configuration form view
func (m *Model) renderFormView() string {
	// Use the enhanced interface as primary
	if m.enhancedConfig != nil {
		return m.renderEnhancedConfigView()
	}
	// Fall back to tabbed interface
	if m.tabbedConfig != nil {
		return m.renderTabbedConfigView()
	}
	
	return m.styles.Error.Render("Configuration not initialized")
}

// renderTabbedConfigView renders the tabbed configuration interface
func (m *Model) renderTabbedConfigView() string {
	if m.tabbedConfig == nil {
		return m.styles.Error.Render("Tabbed configuration not initialized")
	}

	// Get the tabbed config view
	configView := m.tabbedConfig.View()

	// The tabbed view handles its own layout
	return configView
}

// renderEnhancedConfigView renders the enhanced configuration interface
func (m *Model) renderEnhancedConfigView() string {
	if m.enhancedConfig == nil {
		return m.styles.Error.Render("Enhanced configuration not initialized")
	}

	// Get the enhanced config view
	configView := m.enhancedConfig.View()

	// The enhanced view handles its own layout
	return configView
}

// renderHelpView renders the help view
func (m *Model) renderHelpView() string {
	if m.helpSystem == nil {
		return m.styles.Error.Render("Help system not initialized")
	}

	// Basic help content available if needed
	// NOTE: Dynamic help content integration not yet implemented - currently uses static help system

	// Render the help view
	helpView := m.helpSystem.View()

	// Add header
	header := m.styles.Title.Render("📚 Help")

	// Add footer
	footer := m.styles.Help.Render("Esc: Back • q: Quit")

	// Combine
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		"",
		helpView,
		"",
		footer,
	)

	// Ensure help content is constrained to terminal width before placement
	wrapped := lipgloss.NewStyle().MaxWidth(m.width).Render(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		wrapped,
	)
}

// renderProgressView renders the progress/loading view
func (m *Model) renderProgressView() string {
	// If we have a scanner, show its progress
	if m.appScanner != nil && m.appScanner.IsScanning() {
		return m.renderScannerProgress()
	}
	
	// Simple progress view for other loading states
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := (m.frameCount / 2) % len(spinner)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		spinner[idx],
		"",
		m.loadingText,
	)

	// Wrap progress content to avoid lines exceeding terminal width
	styled := m.styles.Help.Render(content)
	wrapped := lipgloss.NewStyle().MaxWidth(m.width).Render(styled)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		wrapped,
	)
}

// renderScannerProgress renders the scanner progress view
func (m *Model) renderScannerProgress() string {
	if m.appScanner == nil {
		return "Initializing scanner..."
	}
	
	// Get scanner view
	scannerView := m.appScanner.View()
	
	// Center it on screen
	wrapped := lipgloss.NewStyle().MaxWidth(m.width).Render(scannerView)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		wrapped,
	)
}

// renderLoadingState renders a loading indicator
func (m *Model) renderLoadingState() string {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := (m.frameCount / 2) % len(spinner)

	loading := fmt.Sprintf("%s %s", spinner[idx], m.loadingText)

	// Wrap loading indicator text as well
	styled := m.styles.Success.Render(loading)
	wrapped := lipgloss.NewStyle().MaxWidth(m.width).Render(styled)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		wrapped,
	)
}

// View caching for performance

// getCachedView returns a cached view if available and valid.
// Avoid returning cached views when transient UI state (like a status/toast)
// is active so that temporary messages are always visible to callers.
func (m *Model) getCachedView() (string, bool) {
	// Don't cache during animations, loading, or when a transient status is active.
	if m.isLoading || m.isStatusActive() {
		return "", false
	}

	// If we never rendered before, there's no valid cache.
	if m.lastRenderTime.IsZero() {
		return "", false
	}

	// Check if we have a cached view for current state
	if cached, ok := m.renderCache[m.state]; ok {
		// Tighten cache validity to reduce risk of returning stale snapshots.
		// Use a smaller window for cache validity (50ms) so tests and callers
		// observe recent changes (e.g. transient toasts) reliably.
		if time.Since(m.lastRenderTime) <= 50*time.Millisecond {
			// Ensure the cached content is non-empty before returning it.
			if len(cached) > 0 {
				return cached, true
			}
		}
	}

	return "", false
}

// cacheView stores the rendered view in cache and records the render timestamp.
// Setting `lastRenderTime` here ensures cache validity is measured from the
// moment the view was actually produced and cached.
func (m *Model) cacheView(content string) {
	if m.renderCache == nil {
		m.renderCache = make(map[ViewState]string)
	}
	m.renderCache[m.state] = content
	// Record the time we generated the cached snapshot so cache validity checks
	// can be performed reliably.
	m.lastRenderTime = time.Now()
}

// invalidateCache clears the render cache
func (m *Model) invalidateCache() {
	// Clear all cached views
	for k := range m.renderCache {
		delete(m.renderCache, k)
	}
}

// Error and fallback rendering

// safeViewRender safely renders a view with panic recovery
func (m *Model) safeViewRender(renderFn func() string, componentName string) string {
	defer func() {
		if r := recover(); r != nil {
			m.logger.LogPanic(r, "render_panic", "component", componentName)
			// Return a safe fallback view
			m.err = fmt.Errorf("render panic in %s: %v", componentName, r)
		}
	}()

	// Call the render function
	result := renderFn()

	return result
}

// renderFallbackView renders a fallback view when something goes wrong
func (m *Model) renderFallbackView() string {
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		m.styles.Error.Render("Something went wrong. Press 'q' to quit or 'r' to restart."),
	)
}

// renderError renders an error view
func (m *Model) renderError() string {
	errorMsg := fmt.Sprintf("❌ Error: %v", m.err)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		m.styles.Error.Render(errorMsg),
		"",
		m.styles.Help.Render("Press 'q' to quit or 'Esc' to go back"),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderDebugInfo renders debug information (development mode)
func (m *Model) renderDebugInfo() string {
	info := []string{
		fmt.Sprintf("State: %v", m.state),
		fmt.Sprintf("Size: %dx%d", m.width, m.height),
		fmt.Sprintf("Frame: %d", m.frameCount),
		fmt.Sprintf("Render: %.2fms", m.lastRenderTime.Sub(time.Now()).Seconds()*1000),
		fmt.Sprintf("App: %s", m.currentApp),
	}

	return m.styles.Help.Render(strings.Join(info, " | "))
}
