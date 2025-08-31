package tui

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mrtkrcm/ZeroUI/internal/logging"
)

// ErrorHandler provides centralized error handling for the TUI
type ErrorHandler struct {
	logger     *logging.CharmLogger
	errors     []ErrorInfo
	maxErrors  int
	errorStyle lipgloss.Style
	warnStyle  lipgloss.Style
	infoStyle  lipgloss.Style
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Error     error
	Timestamp time.Time
	Context   string
	Severity  ErrorSeverity
	Recovered bool
}

// ErrorSeverity represents error severity levels
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *logging.CharmLogger) *ErrorHandler {
	return &ErrorHandler{
		logger:    logger,
		errors:    make([]ErrorInfo, 0),
		maxErrors: 10,
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		warnStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),
	}
}

// Handle processes an error with context
func (h *ErrorHandler) Handle(err error, context string) tea.Cmd {
	if err == nil {
		return nil
	}

	info := ErrorInfo{
		Error:     err,
		Timestamp: time.Now(),
		Context:   context,
		Severity:  h.determineSeverity(err),
		Recovered: false,
	}

	h.addError(info)
	h.logError(info)

	// Return a command to show error notification
	return h.showErrorNotification(info)
}

// HandlePanic recovers from panic and logs it
func (h *ErrorHandler) HandlePanic(context string) {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic: %v", r)

		info := ErrorInfo{
			Error:     err,
			Timestamp: time.Now(),
			Context:   context,
			Severity:  SeverityCritical,
			Recovered: true,
		}

		h.addError(info)
		h.logPanic(r, context)
	}
}

// SafeExecute runs a function with panic recovery
func (h *ErrorHandler) SafeExecute(fn func() error, context string) error {
	defer h.HandlePanic(context)
	return fn()
}

// SafeUpdate wraps component updates with error handling
func (h *ErrorHandler) SafeUpdate(
	updateFn func() (interface{}, tea.Cmd),
	context string,
) (result interface{}, cmd tea.Cmd) {
	defer func() {
		if r := recover(); r != nil {
			h.HandlePanic(context)
			result = nil
			cmd = nil
		}
	}()

	return updateFn()
}

// GetErrors returns recent errors
func (h *ErrorHandler) GetErrors() []ErrorInfo {
	return h.errors
}

// GetLastError returns the most recent error
func (h *ErrorHandler) GetLastError() *ErrorInfo {
	if len(h.errors) == 0 {
		return nil
	}
	return &h.errors[len(h.errors)-1]
}

// HasErrors checks if there are any errors
func (h *ErrorHandler) HasErrors() bool {
	return len(h.errors) > 0
}

// ClearErrors clears all stored errors
func (h *ErrorHandler) ClearErrors() {
	h.errors = make([]ErrorInfo, 0)
}

// RenderError renders an error message for display
func (h *ErrorHandler) RenderError(err error) string {
	if err == nil {
		return ""
	}

	return h.errorStyle.Render(fmt.Sprintf("Error: %v", err))
}

// RenderLastError renders the last error for display
func (h *ErrorHandler) RenderLastError() string {
	last := h.GetLastError()
	if last == nil {
		return ""
	}

	var style lipgloss.Style
	switch last.Severity {
	case SeverityCritical, SeverityError:
		style = h.errorStyle
	case SeverityWarning:
		style = h.warnStyle
	default:
		style = h.infoStyle
	}

	prefix := h.getSeverityPrefix(last.Severity)
	return style.Render(fmt.Sprintf("%s %v", prefix, last.Error))
}

// Private methods

func (h *ErrorHandler) addError(info ErrorInfo) {
	h.errors = append(h.errors, info)

	// Keep only recent errors
	if len(h.errors) > h.maxErrors {
		h.errors = h.errors[1:]
	}
}

func (h *ErrorHandler) logError(info ErrorInfo) {
	if h.logger == nil {
		return
	}

	switch info.Severity {
	case SeverityCritical:
		h.logger.LogError(info.Error, "critical_error", "context", info.Context)
	case SeverityError:
		h.logger.LogError(info.Error, "error", "context", info.Context)
	case SeverityWarning:
		h.logger.Warn("Warning occurred", "error", info.Error.Error(), "context", info.Context)
	default:
		h.logger.Info("Info", "message", info.Error.Error(), "context", info.Context)
	}
}

func (h *ErrorHandler) logPanic(r interface{}, context string) {
	if h.logger == nil {
		return
	}

	stack := string(debug.Stack())
	h.logger.LogPanic(r, "panic_recovered", "context", context, "stack", stack)
}

func (h *ErrorHandler) determineSeverity(err error) ErrorSeverity {
	if err == nil {
		return SeverityInfo
	}

	errStr := err.Error()

	// Check for critical patterns
	if contains(errStr, "panic", "fatal", "critical") {
		return SeverityCritical
	}

	// Check for error patterns
	if contains(errStr, "error", "failed", "unable") {
		return SeverityError
	}

	// Check for warning patterns
	if contains(errStr, "warning", "warn", "deprecated") {
		return SeverityWarning
	}

	return SeverityInfo
}

func (h *ErrorHandler) getSeverityPrefix(severity ErrorSeverity) string {
	switch severity {
	case SeverityCritical:
		return "✗"
	case SeverityError:
		return "✗"
	case SeverityWarning:
		return "⚠"
	default:
		return "ℹ"
	}
}

func (h *ErrorHandler) showErrorNotification(info ErrorInfo) tea.Cmd {
	// Return a command that will show a temporary error notification
	return func() tea.Msg {
		return ErrorNotificationMsg{
			Error:    info.Error,
			Severity: info.Severity,
			Context:  info.Context,
		}
	}
}

// ErrorNotificationMsg is sent to show error notifications
type ErrorNotificationMsg struct {
	Error    error
	Severity ErrorSeverity
	Context  string
}

// Helper function to check if string contains any of the patterns
func contains(s string, patterns ...string) bool {
	for _, pattern := range patterns {
		if len(pattern) > 0 && len(s) >= len(pattern) {
			// Case-insensitive contains
			if containsIgnoreCase(s, pattern) {
				return true
			}
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
