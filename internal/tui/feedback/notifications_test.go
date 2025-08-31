package feedback

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewNotificationSystem(t *testing.T) {
	ns := NewNotificationSystem()
	if ns == nil {
		t.Fatal("NewNotificationSystem returned nil")
	}
	if len(ns.notifications) != 0 {
		t.Errorf("Expected empty notifications list, got %d", len(ns.notifications))
	}
	if ns.maxNotifications != 5 {
		t.Errorf("Expected maxNotifications=5, got %d", ns.maxNotifications)
	}
}

func TestShowInfo(t *testing.T) {
	ns := NewNotificationSystem()
	message := "Test info message"
	duration := 2 * time.Second

	ns.ShowInfo(message, duration)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Message != message {
		t.Errorf("Expected message %q, got %q", message, n.Message)
	}
	if n.Type != NotificationTypeInfo {
		t.Errorf("Expected type %v, got %v", NotificationTypeInfo, n.Type)
	}
	if n.Priority != PriorityNormal {
		t.Errorf("Expected priority %v, got %v", PriorityNormal, n.Priority)
	}
}

func TestShowSuccess(t *testing.T) {
	ns := NewNotificationSystem()
	message := "Operation completed successfully"

	ns.ShowSuccess(message, 1*time.Second)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Type != NotificationTypeSuccess {
		t.Errorf("Expected success type, got %v", n.Type)
	}
	if !strings.Contains(n.Message, message) {
		t.Errorf("Expected message to contain %q", message)
	}
}

func TestShowError(t *testing.T) {
	ns := NewNotificationSystem()
	message := "An error occurred"

	ns.ShowError(message, 1*time.Second)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Type != NotificationTypeError {
		t.Errorf("Expected error type, got %v", n.Type)
	}
	if n.Priority != PriorityHigh {
		t.Errorf("Expected high priority for errors, got %v", n.Priority)
	}
}

func TestShowWarning(t *testing.T) {
	ns := NewNotificationSystem()
	message := "Warning: This action may have consequences"

	ns.ShowWarning(message, 1*time.Second)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Type != NotificationTypeWarning {
		t.Errorf("Expected warning type, got %v", n.Type)
	}
}

func TestShowAchievement(t *testing.T) {
	ns := NewNotificationSystem()
	achievement := "First configuration saved!"

	ns.ShowAchievement(achievement)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Type != NotificationTypeAchievement {
		t.Errorf("Expected achievement type, got %v", n.Type)
	}
	if n.Priority != PriorityHigh {
		t.Errorf("Expected high priority for achievements, got %v", n.Priority)
	}
	if !strings.Contains(n.Message, "🏆") {
		t.Error("Expected achievement message to contain trophy emoji")
	}
}

func TestShowProgress(t *testing.T) {
	ns := NewNotificationSystem()
	message := "Processing files..."
	current, total := 5, 10

	ns.ShowProgress(message, current, total)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(notifications))
	}

	n := notifications[0]
	if n.Type != NotificationTypeProgress {
		t.Errorf("Expected progress type, got %v", n.Type)
	}
	if n.Progress == nil {
		t.Fatal("Expected progress indicator to be set")
	}
	if n.Progress.Current != current {
		t.Errorf("Expected progress current %d, got %d", current, n.Progress.Current)
	}
	if n.Progress.Total != total {
		t.Errorf("Expected progress total %d, got %d", total, n.Progress.Total)
	}
}

func TestUpdateProgress(t *testing.T) {
	ns := NewNotificationSystem()
	ns.ShowProgress("Processing...", 0, 100)

	notifications := ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Fatal("Expected notification to exist")
	}

	initialID := notifications[0].ID

	// Update progress
	ns.UpdateProgress(initialID, 50, "Halfway there...")

	// Check updated notification
	notifications = ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Fatal("Expected notification to still exist")
	}

	n := notifications[0]
	if n.Progress.Current != 50 {
		t.Errorf("Expected progress 50, got %d", n.Progress.Current)
	}
	if n.Progress.Message != "Halfway there..." {
		t.Errorf("Expected message %q, got %q", "Halfway there...", n.Progress.Message)
	}
}

func TestCompleteProgress(t *testing.T) {
	ns := NewNotificationSystem()
	ns.ShowProgress("Processing...", 0, 100)

	notifications := ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Fatal("Expected notification to exist")
	}

	initialID := notifications[0].ID

	// Complete progress
	successMessage := "Process completed successfully!"
	ns.CompleteProgress(initialID, successMessage)

	// Check completion
	notifications = ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Fatal("Expected notification to still exist after completion")
	}

	n := notifications[0]
	if n.Type != NotificationTypeSuccess {
		t.Errorf("Expected success type after completion, got %v", n.Type)
	}
	if n.Progress != nil {
		t.Error("Expected progress indicator to be cleared after completion")
	}
	if n.Message != successMessage {
		t.Errorf("Expected message %q, got %q", successMessage, n.Message)
	}
}

func TestMaxNotifications(t *testing.T) {
	ns := NewNotificationSystem()

	// Add more than max notifications
	for i := 0; i < 8; i++ {
		ns.ShowInfo(fmt.Sprintf("Notification %d", i), 10*time.Second)
	}

	notifications := ns.GetActiveNotifications()
	if len(notifications) > ns.maxNotifications {
		t.Errorf("Expected at most %d notifications, got %d", ns.maxNotifications, len(notifications))
	}
}

func TestDismiss(t *testing.T) {
	ns := NewNotificationSystem()
	ns.ShowInfo("Test message", 10*time.Second)

	notifications := ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Fatal("Expected notification to exist")
	}

	id := notifications[0].ID
	ns.Dismiss(id)

	// After dismissal, notification should be marked for expiration
	notifications = ns.GetActiveNotifications()
	if len(notifications) != 0 {
		// Check if notification has expiration set (immediate cleanup)
		for _, n := range notifications {
			if n.ID == id && n.Expires.IsZero() {
				t.Error("Expected notification to be marked for expiration after dismiss")
			}
		}
	}

	// After cleanup, notification should be gone
	ns.cleanExpired()
	notifications = ns.GetActiveNotifications()
	if len(notifications) != 0 {
		t.Errorf("Expected notification to be cleaned up, got %d", len(notifications))
	}
}

func TestClearAll(t *testing.T) {
	ns := NewNotificationSystem()
	ns.ShowInfo("Test 1", 10*time.Second)
	ns.ShowInfo("Test 2", 10*time.Second)

	notifications := ns.GetActiveNotifications()
	if len(notifications) != 2 {
		t.Fatalf("Expected 2 notifications, got %d", len(notifications))
	}

	ns.ClearAll()

	notifications = ns.GetActiveNotifications()
	if len(notifications) != 0 {
		t.Errorf("Expected no notifications after ClearAll, got %d", len(notifications))
	}
}

func TestPresetNotifications(t *testing.T) {
	ns := NewNotificationSystem()

	// Test preset welcome
	ns.PresetWelcome()
	notifications := ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Error("Expected welcome notification")
	}

	// Clear and test preset saved
	ns.ClearAll()
	ns.PresetSaved()
	notifications = ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Error("Expected saved notification")
	}

	// Clear and test preset error
	ns.ClearAll()
	ns.PresetError("Something went wrong")
	notifications = ns.GetActiveNotifications()
	if len(notifications) == 0 {
		t.Error("Expected error notification")
	}
}

func TestRender(t *testing.T) {
	ns := NewNotificationSystem()
	ns.ShowSuccess("Operation successful", 5*time.Second)

	width := 80
	rendered := ns.Render(width)

	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}

	// Check that rendered output contains expected elements
	if !strings.Contains(rendered, "✅") {
		t.Error("Expected success icon in rendered output")
	}
	if !strings.Contains(rendered, "Operation successful") {
		t.Error("Expected message in rendered output")
	}
}

func TestNotificationExpiration(t *testing.T) {
	ns := NewNotificationSystem()

	// Add notification with short duration
	ns.ShowInfo("Short lived", 100*time.Millisecond)

	// Initially should exist
	notifications := ns.GetActiveNotifications()
	if len(notifications) != 1 {
		t.Fatalf("Expected 1 notification initially, got %d", len(notifications))
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired after cleanup
	ns.cleanExpired()
	notifications = ns.GetActiveNotifications()
	if len(notifications) != 0 {
		t.Errorf("Expected notification to expire, got %d", len(notifications))
	}
}
