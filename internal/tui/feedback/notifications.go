package feedback

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// NotificationPosition defines where notifications appear on screen
type NotificationPosition int

const (
	PositionTopRight NotificationPosition = iota
	PositionTopCenter
	PositionBottomRight
	PositionBottomCenter
)

// NotificationSystem manages user notifications and feedback
type NotificationSystem struct {
	notifications    []Notification
	maxNotifications int
	theme            NotificationTheme
	position         NotificationPosition
	defaultDuration  time.Duration
}

// Notification represents a user notification
type Notification struct {
	ID       string
	Message  string
	Type     NotificationType
	Priority NotificationPriority
	Created  time.Time
	Expires  time.Time
	AutoHide bool
	Actions  []NotificationAction
	Progress *ProgressIndicator
}

// NotificationType defines the type of notification
type NotificationType int

const (
	NotificationTypeInfo NotificationType = iota
	NotificationTypeSuccess
	NotificationTypeWarning
	NotificationTypeError
	NotificationTypeProgress
	NotificationTypeAchievement
)

// NotificationPriority defines notification priority
type NotificationPriority int

const (
	PriorityLow NotificationPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// NotificationAction represents an action button in a notification
type NotificationAction struct {
	Label    string
	Action   string
	Shortcut string
}

// ProgressIndicator shows progress for long-running operations
type ProgressIndicator struct {
	Current int
	Total   int
	Message string
}

// NotificationTheme defines the visual theme for notifications
type NotificationTheme struct {
	InfoColor    string
	SuccessColor string
	WarningColor string
	ErrorColor   string
	BorderColor  string
	TextColor    string
}

// DefaultTheme provides a beautiful default theme
var DefaultTheme = NotificationTheme{
	InfoColor:    "#8be9fd",
	SuccessColor: "#50fa7b",
	WarningColor: "#f1fa8c",
	ErrorColor:   "#ff5555",
	BorderColor:  "#44475a",
	TextColor:    "#f8f8f2",
}

// NewNotificationSystem creates a new notification system
func NewNotificationSystem() *NotificationSystem {
	return &NotificationSystem{
		notifications:    []Notification{},
		maxNotifications: 5,
		theme:            DefaultTheme,
		position:         PositionTopRight,
		defaultDuration:  3 * time.Second,
	}
}

// ShowInfo shows an info notification
func (ns *NotificationSystem) ShowInfo(message string, duration time.Duration) {
	ns.show(NotificationTypeInfo, PriorityNormal, message, duration, nil)
}

// ShowSuccess shows a success notification
func (ns *NotificationSystem) ShowSuccess(message string, duration time.Duration) {
	ns.show(NotificationTypeSuccess, PriorityNormal, message, duration, nil)
}

// ShowWarning shows a warning notification
func (ns *NotificationSystem) ShowWarning(message string, duration time.Duration) {
	ns.show(NotificationTypeWarning, PriorityNormal, message, duration, nil)
}

// ShowError shows an error notification
func (ns *NotificationSystem) ShowError(message string, duration time.Duration) {
	ns.show(NotificationTypeError, PriorityHigh, message, duration, nil)
}

// ShowProgress shows a progress notification
func (ns *NotificationSystem) ShowProgress(message string, current, total int) {
	progress := &ProgressIndicator{
		Current: current,
		Total:   total,
		Message: message,
	}

	notification := Notification{
		ID:       generateID(),
		Message:  message,
		Type:     NotificationTypeProgress,
		Priority: PriorityNormal,
		Created:  time.Now(),
		Progress: progress,
	}

	ns.addNotification(notification)
}

// ShowAchievement shows an achievement notification
func (ns *NotificationSystem) ShowAchievement(message string) {
	ns.show(NotificationTypeAchievement, PriorityHigh, "🏆 "+message, 5*time.Second, nil)
}

// ShowWithActions shows a notification with action buttons
func (ns *NotificationSystem) ShowWithActions(message string, actions []NotificationAction, duration time.Duration) {
	ns.show(NotificationTypeInfo, PriorityNormal, message, duration, actions)
}

// UpdateProgress updates the progress of a notification
func (ns *NotificationSystem) UpdateProgress(id string, current int, message string) {
	for i := range ns.notifications {
		if ns.notifications[i].ID == id && ns.notifications[i].Progress != nil {
			ns.notifications[i].Progress.Current = current
			ns.notifications[i].Progress.Message = message
			ns.notifications[i].Message = message
			break
		}
	}
}

// CompleteProgress marks a progress notification as complete
func (ns *NotificationSystem) CompleteProgress(id string, successMessage string) {
	for i := range ns.notifications {
		if ns.notifications[i].ID == id {
			ns.notifications[i].Type = NotificationTypeSuccess
			ns.notifications[i].Message = successMessage
			ns.notifications[i].Progress = nil
			ns.notifications[i].Expires = time.Now().Add(3 * time.Second)
			break
		}
	}
}

// GetActiveNotifications returns currently active notifications
func (ns *NotificationSystem) GetActiveNotifications() []Notification {
	var active []Notification
	now := time.Now()

	for _, notification := range ns.notifications {
		if notification.Expires.IsZero() || now.Before(notification.Expires) {
			active = append(active, notification)
		}
	}

	return active
}

// Render renders all active notifications
func (ns *NotificationSystem) Render(width, height int) string {
	active := ns.GetActiveNotifications()
	if len(active) == 0 {
		return ""
	}

	content := ns.renderNotifications(active, width)

	var hPos, vPos lipgloss.Position
	switch ns.position {
	case PositionTopRight:
		hPos, vPos = lipgloss.Right, lipgloss.Top
	case PositionTopCenter:
		hPos, vPos = lipgloss.Center, lipgloss.Top
	case PositionBottomRight:
		hPos, vPos = lipgloss.Right, lipgloss.Bottom
	case PositionBottomCenter:
		hPos, vPos = lipgloss.Center, lipgloss.Bottom
	}

	return lipgloss.Place(width, height, hPos, vPos, content)
}

// SetPosition sets the notification display position
func (ns *NotificationSystem) SetPosition(pos NotificationPosition) {
	ns.position = pos
}

// Dismiss dismisses a notification by ID
func (ns *NotificationSystem) Dismiss(id string) {
	for i := range ns.notifications {
		if ns.notifications[i].ID == id {
			ns.notifications[i].Expires = time.Now()
			break
		}
	}
}

// ClearAll clears all notifications
func (ns *NotificationSystem) ClearAll() {
	ns.notifications = []Notification{}
}

// Update updates the notification system (for timer-based operations)
func (ns *NotificationSystem) Update() {
	// Remove expired notifications
	ns.cleanExpired()
}

// renderNotifications renders all active notifications as a single block
func (ns *NotificationSystem) renderNotifications(active []Notification, width int) string {
	var rendered []string
	for _, notification := range active {
		rendered = append(rendered, ns.renderNotification(notification, width))
	}
	return strings.Join(rendered, "\n")
}

// ShowTooltip shows a tooltip-style notification
func (ns *NotificationSystem) ShowTooltip(message string, duration time.Duration) {
	ns.ShowInfo(message, duration)
}

// Private methods

func (ns *NotificationSystem) show(notificationType NotificationType, priority NotificationPriority, message string, duration time.Duration, actions []NotificationAction) {
	notification := Notification{
		ID:       generateID(),
		Message:  message,
		Type:     notificationType,
		Priority: priority,
		Created:  time.Now(),
		AutoHide: duration > 0,
		Actions:  actions,
	}

	if duration > 0 {
		notification.Expires = time.Now().Add(duration)
	}

	ns.addNotification(notification)
}

func (ns *NotificationSystem) addNotification(notification Notification) {
	// Remove expired notifications
	ns.cleanExpired()

	// Add new notification
	ns.notifications = append(ns.notifications, notification)

	// Sort by priority (highest first)
	for i := len(ns.notifications) - 1; i > 0; i-- {
		if ns.notifications[i].Priority > ns.notifications[i-1].Priority {
			ns.notifications[i], ns.notifications[i-1] = ns.notifications[i-1], ns.notifications[i]
		} else {
			break
		}
	}

	// Limit number of notifications
	if len(ns.notifications) > ns.maxNotifications {
		ns.notifications = ns.notifications[:ns.maxNotifications]
	}
}

func (ns *NotificationSystem) cleanExpired() {
	var active []Notification
	now := time.Now()

	for _, notification := range ns.notifications {
		if notification.Expires.IsZero() || now.Before(notification.Expires) {
			active = append(active, notification)
		}
	}

	ns.notifications = active
}

func (ns *NotificationSystem) renderNotification(notification Notification, width int) string {
	var style lipgloss.Style

	// Choose style based on type
	switch notification.Type {
	case NotificationTypeSuccess:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ns.theme.SuccessColor)).
			Background(lipgloss.Color("#1e1e2e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ns.theme.SuccessColor)).
			Padding(0, 1)
	case NotificationTypeError:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ns.theme.ErrorColor)).
			Background(lipgloss.Color("#1e1e2e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ns.theme.ErrorColor)).
			Padding(0, 1)
	case NotificationTypeWarning:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ns.theme.WarningColor)).
			Background(lipgloss.Color("#1e1e2e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ns.theme.WarningColor)).
			Padding(0, 1)
	case NotificationTypeProgress:
		return ns.renderProgressNotification(notification, width)
	case NotificationTypeAchievement:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffd700")).
			Background(lipgloss.Color("#1e1e2e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#ffd700")).
			Padding(0, 1).
			Bold(true)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ns.theme.InfoColor)).
			Background(lipgloss.Color("#1e1e2e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ns.theme.InfoColor)).
			Padding(0, 1)
	}

	// Get icon based on type
	icon := ns.getNotificationIcon(notification.Type)

	// Build message
	message := fmt.Sprintf("%s %s", icon, notification.Message)

	// Add action hints if present
	if len(notification.Actions) > 0 {
		var actionHints []string
		for _, action := range notification.Actions {
			if action.Shortcut != "" {
				actionHints = append(actionHints, fmt.Sprintf("[%s]%s", action.Shortcut, action.Label))
			} else {
				actionHints = append(actionHints, action.Label)
			}
		}
		message += " " + strings.Join(actionHints, " • ")
	}

	return style.Width(width).Render(message)
}

func (ns *NotificationSystem) renderProgressNotification(notification Notification, width int) string {
	if notification.Progress == nil {
		return ""
	}

	progress := notification.Progress
	percentage := float64(progress.Current) / float64(progress.Total) * 100

	// Create progress bar
	barWidth := width - 10 // Leave space for text
	filled := int(float64(barWidth) * percentage / 100)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	message := fmt.Sprintf("⏳ %s (%.1f%%)", progress.Message, percentage)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ns.theme.InfoColor)).
		Background(lipgloss.Color("#1e1e2e")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ns.theme.InfoColor)).
		Padding(0, 1)

	content := message + "\n" + bar
	return style.Width(width).Render(content)
}

func (ns *NotificationSystem) getNotificationIcon(notificationType NotificationType) string {
	switch notificationType {
	case NotificationTypeSuccess:
		return "✅"
	case NotificationTypeError:
		return "❌"
	case NotificationTypeWarning:
		return "⚠️"
	case NotificationTypeProgress:
		return "⏳"
	case NotificationTypeAchievement:
		return "🏆"
	default:
		return "ℹ️"
	}
}

// Utility functions
func generateID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

// Preset notifications for common scenarios
func (ns *NotificationSystem) PresetWelcome() {
	ns.ShowInfo("👋 Welcome to ZeroUI! Press ? for help.", 5*time.Second)
}

func (ns *NotificationSystem) PresetSaved() {
	ns.ShowSuccess("💾 Configuration saved successfully!", 3*time.Second)
}

func (ns *NotificationSystem) PresetError(message string) {
	ns.ShowError(fmt.Sprintf("❌ %s", message), 5*time.Second)
}

func (ns *NotificationSystem) PresetLoading(message string) {
	ns.ShowProgress(message, 0, 100)
}

func (ns *NotificationSystem) PresetAchievement(achievement string) {
	ns.ShowAchievement(fmt.Sprintf("Achievement unlocked: %s", achievement))
}

func (ns *NotificationSystem) PresetTip(tip string) {
	ns.ShowInfo(fmt.Sprintf("💡 %s", tip), 4*time.Second)
}

// Batch operations
func (ns *NotificationSystem) BatchSave() {
	ns.ShowProgress("Saving configuration...", 0, 100)

	// Simulate progress
	go func() {
		steps := []string{
			"Validating configuration...",
			"Applying changes...",
			"Saving to file...",
			"Configuration saved!",
		}

		for i, step := range steps {
			time.Sleep(500 * time.Millisecond)
			if i < len(steps)-1 {
				ns.UpdateProgress(ns.notifications[len(ns.notifications)-1].ID, (i+1)*25, step)
			} else {
				ns.CompleteProgress(ns.notifications[len(ns.notifications)-1].ID, step)
			}
		}
	}()
}

// Smart notifications based on user behavior
func (ns *NotificationSystem) SmartNotify(action string, context string) {
	switch {
	case action == "saved" && context == "first-time":
		ns.ShowAchievement("First configuration saved!")
	case action == "searched" && context == "empty":
		ns.PresetTip("Try searching for 'font', 'color', or 'theme'")
	case action == "edited" && context == "advanced":
		ns.PresetTip("Great! You found an advanced setting")
	case strings.Contains(action, "error"):
		ns.PresetError("Something went wrong. Check your input and try again.")
	}
}

// Accessibility features
func (ns *NotificationSystem) SetHighContrast(enabled bool) {
	if enabled {
		ns.theme = NotificationTheme{
			InfoColor:    "#ffffff",
			SuccessColor: "#00ff00",
			WarningColor: "#ffff00",
			ErrorColor:   "#ff0000",
			BorderColor:  "#ffffff",
			TextColor:    "#ffffff",
		}
	} else {
		ns.theme = DefaultTheme
	}
}

func (ns *NotificationSystem) EnableScreenReader() {
	// Add screen reader announcements
	for _, notification := range ns.notifications {
		// In a real implementation, this would announce to screen readers
		_ = notification
	}
}
