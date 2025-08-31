package feedback

import (
	"strings"
	"testing"
	"time"
)

func TestNewLoadingSystem(t *testing.T) {
	ls := NewLoadingSystem()
	if ls == nil {
		t.Fatal("NewLoadingSystem returned nil")
	}
	if len(ls.activeLoaders) != 0 {
		t.Errorf("Expected empty activeLoaders map, got %d", len(ls.activeLoaders))
	}
	if len(ls.spinners) != 0 {
		t.Errorf("Expected empty spinners map, got %d", len(ls.spinners))
	}
}

func TestStartLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-loading"
	message := "Loading test data..."

	ls.StartLoading(id, message)

	if len(ls.activeLoaders) != 1 {
		t.Fatalf("Expected 1 active loader, got %d", len(ls.activeLoaders))
	}

	if len(ls.spinners) != 1 {
		t.Fatalf("Expected 1 spinner, got %d", len(ls.spinners))
	}

	loader, exists := ls.activeLoaders[id]
	if !exists {
		t.Fatal("Expected loader to exist")
	}

	if loader.Message != message {
		t.Errorf("Expected message %q, got %q", message, loader.Message)
	}
	if loader.ID != id {
		t.Errorf("Expected ID %q, got %q", id, loader.ID)
	}
	if loader.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
}

func TestStartStepLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-step-loading"
	message := "Processing files..."
	steps := []string{"Reading files", "Processing data", "Saving results"}

	ls.StartStepLoading(id, message, steps)

	loader, exists := ls.activeLoaders[id]
	if !exists {
		t.Fatal("Expected loader to exist")
	}

	if len(loader.Steps) != len(steps) {
		t.Errorf("Expected %d steps, got %d", len(steps), len(loader.Steps))
	}

	for i, step := range steps {
		if loader.Steps[i].Name != step {
			t.Errorf("Expected step %d name %q, got %q", i, step, loader.Steps[i].Name)
		}
		if loader.Steps[i].Completed {
			t.Errorf("Expected step %d to not be completed initially", i)
		}
	}
}

func TestUpdateStep(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-update-step"
	steps := []string{"Step 1", "Step 2", "Step 3"}

	ls.StartStepLoading(id, "Processing...", steps)

	// Update to step 1 (index 0)
	ls.UpdateStep(id, 0)

	loader := ls.activeLoaders[id]
	if loader.CurrentStep != 0 {
		t.Errorf("Expected current step 0, got %d", loader.CurrentStep)
	}
	if !loader.Steps[0].Completed {
		t.Error("Expected step 0 to be completed")
	}
	if loader.Steps[1].Completed {
		t.Error("Expected step 1 to not be completed yet")
	}

	// Update to step 2 (index 1)
	ls.UpdateStep(id, 1)

	if loader.CurrentStep != 1 {
		t.Errorf("Expected current step 1, got %d", loader.CurrentStep)
	}
	if !loader.Steps[1].Completed {
		t.Error("Expected step 1 to be completed")
	}
}

func TestUpdateLoadingProgress(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-progress"
	message := "Processing items..."

	ls.StartLoading(id, message)

	// Update progress
	current, total := 5, 10
	progressMessage := "Halfway there..."
	ls.UpdateProgress(id, current, total, progressMessage)

	loader := ls.activeLoaders[id]
	if loader.Progress == nil {
		t.Fatal("Expected progress to be set")
	}

	if loader.Progress.Current != current {
		t.Errorf("Expected progress current %d, got %d", current, loader.Progress.Current)
	}
	if loader.Progress.Total != total {
		t.Errorf("Expected progress total %d, got %d", total, loader.Progress.Total)
	}
	if loader.Progress.Message != progressMessage {
		t.Errorf("Expected progress message %q, got %q", progressMessage, loader.Progress.Message)
	}
}

func TestCompleteLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-complete"
	message := "Loading data..."

	ls.StartLoading(id, message)

	// Complete loading
	successMessage := "Data loaded successfully!"
	ls.CompleteLoading(id, successMessage)

	loader, exists := ls.activeLoaders[id]
	if !exists {
		t.Fatal("Expected loader to still exist after completion")
	}

	if loader.Message != successMessage {
		t.Errorf("Expected message %q, got %q", successMessage, loader.Message)
	}

	// Simulate time passing for auto-cleanup
	time.Sleep(10 * time.Millisecond)
	ls.Update()

	// After update, the loader should still exist (cleanup happens after 2 seconds)
	_, exists = ls.activeLoaders[id]
	if !exists {
		t.Error("Expected loader to still exist before cleanup delay")
	}
}

func TestFailLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-fail"
	message := "Processing data..."

	ls.StartLoading(id, message)

	// Fail loading
	errorMessage := "Failed to process data"
	ls.FailLoading(id, errorMessage)

	loader := ls.activeLoaders[id]
	if loader.Message != errorMessage {
		t.Errorf("Expected message %q, got %q", errorMessage, loader.Message)
	}
}

func TestCancelLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-cancel"
	message := "Loading data..."

	ls.StartLoading(id, message)

	if len(ls.activeLoaders) != 1 {
		t.Fatalf("Expected 1 active loader, got %d", len(ls.activeLoaders))
	}

	ls.CancelLoading(id)

	if len(ls.activeLoaders) != 0 {
		t.Errorf("Expected 0 active loaders after cancel, got %d", len(ls.activeLoaders))
	}

	if len(ls.spinners) != 0 {
		t.Errorf("Expected 0 spinners after cancel, got %d", len(ls.spinners))
	}
}

func TestIsLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-is-loading"

	if ls.IsLoading(id) {
		t.Error("Expected not loading initially")
	}

	ls.StartLoading(id, "Loading...")

	if !ls.IsLoading(id) {
		t.Error("Expected loading after start")
	}

	ls.CancelLoading(id)

	if ls.IsLoading(id) {
		t.Error("Expected not loading after cancel")
	}
}

func TestGetActiveLoaders(t *testing.T) {
	ls := NewLoadingSystem()

	// Initially empty
	loaders := ls.GetActiveLoaders()
	if len(loaders) != 0 {
		t.Errorf("Expected 0 active loaders initially, got %d", len(loaders))
	}

	// Add loaders
	ls.StartLoading("loader1", "Loading 1...")
	ls.StartLoading("loader2", "Loading 2...")

	loaders = ls.GetActiveLoaders()
	if len(loaders) != 2 {
		t.Errorf("Expected 2 active loaders, got %d", len(loaders))
	}

	// Verify both loaders exist
	_, exists1 := loaders["loader1"]
	_, exists2 := loaders["loader2"]

	if !exists1 {
		t.Error("Expected loader1 to exist")
	}
	if !exists2 {
		t.Error("Expected loader2 to exist")
	}
}

func TestRenderLoading(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-render"
	message := "Rendering test..."

	ls.StartLoading(id, message)

	width := 80
	rendered := ls.Render(width)

	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}

	// Check that rendered output contains expected elements
	if !strings.Contains(rendered, message) {
		t.Errorf("Expected rendered output to contain message %q", message)
	}
}

func TestRenderProgress(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-render-progress"

	ls.StartLoading(id, "Processing...")
	ls.UpdateProgress(id, 5, 10, "Halfway done")

	width := 80
	rendered := ls.Render(width)

	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}

	// Check for progress indication
	if !strings.Contains(rendered, "50") { // 50% or 5/10
		t.Error("Expected rendered output to contain progress information")
	}
}

func TestRenderMultiStep(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-multi-step"
	steps := []string{"Reading", "Processing", "Saving"}

	ls.StartStepLoading(id, "Multi-step operation...", steps)
	ls.UpdateStep(id, 0) // Complete first step

	width := 80
	rendered := ls.Render(width)

	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}

	// Check for step indicators
	if !strings.Contains(rendered, "✅") { // Completed step
		t.Error("Expected rendered output to contain completed step indicator")
	}
	if !strings.Contains(rendered, "⏸️") { // Pending step
		t.Error("Expected rendered output to contain pending step indicator")
	}
}

func TestGetElapsedTime(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-elapsed"

	// Before starting
	elapsed := ls.GetElapsedTime(id)
	if elapsed != 0 {
		t.Errorf("Expected 0 elapsed time before starting, got %v", elapsed)
	}

	ls.StartLoading(id, "Loading...")

	// Immediately after starting (should be very small)
	elapsed = ls.GetElapsedTime(id)
	if elapsed < 0 {
		t.Error("Expected non-negative elapsed time")
	}

	// After a short delay
	time.Sleep(10 * time.Millisecond)
	elapsed = ls.GetElapsedTime(id)
	if elapsed <= 0 {
		t.Errorf("Expected positive elapsed time after delay, got %v", elapsed)
	}
}

func TestGetProgress(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-get-progress"

	// Before starting
	current, total, message := ls.GetProgress(id)
	if current != 0 || total != 0 || message != "" {
		t.Errorf("Expected (0, 0, \"\") before starting, got (%d, %d, %q)", current, total, message)
	}

	ls.StartLoading(id, "Loading...")
	ls.UpdateProgress(id, 3, 7, "Processing...")

	current, total, message = ls.GetProgress(id)
	if current != 3 {
		t.Errorf("Expected current 3, got %d", current)
	}
	if total != 7 {
		t.Errorf("Expected total 7, got %d", total)
	}
	if message != "Processing..." {
		t.Errorf("Expected message %q, got %q", "Processing...", message)
	}
}

func TestPresetMethods(t *testing.T) {
	ls := NewLoadingSystem()

	// Test preset save
	ls.StartConfigSave()
	if !ls.IsLoading("config-save") {
		t.Error("Expected config-save loading to start")
	}

	// Test preset file load
	ls.StartFileLoad()
	if !ls.IsLoading("file-load") {
		t.Error("Expected file-load loading to start")
	}

	// Test preset validation
	ls.StartValidation()
	if !ls.IsLoading("validation") {
		t.Error("Expected validation loading to start")
	}

	// Test preset backup
	ls.StartBackup()
	if !ls.IsLoading("backup") {
		t.Error("Expected backup loading to start")
	}
}

func TestPerformanceStats(t *testing.T) {
	ls := NewLoadingSystem()

	// Initially empty
	stats := ls.GetPerformanceStats()
	if stats["active_loaders"].(int) != 0 {
		t.Errorf("Expected 0 active loaders in stats, got %v", stats["active_loaders"])
	}

	// Add some loaders
	ls.StartLoading("loader1", "Loading 1...")
	ls.StartLoading("loader2", "Loading 2...")

	stats = ls.GetPerformanceStats()
	if stats["active_loaders"].(int) != 2 {
		t.Errorf("Expected 2 active loaders in stats, got %v", stats["active_loaders"])
	}

	// Verify stats contain expected keys
	expectedKeys := []string{"active_loaders", "total_operations", "average_duration"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected stats to contain key %q", key)
		}
	}
}

func TestMultipleLoaders(t *testing.T) {
	ls := NewLoadingSystem()

	// Start multiple loaders
	ids := []string{"loader1", "loader2", "loader3"}
	messages := []string{"Loading A", "Loading B", "Loading C"}

	for i, id := range ids {
		ls.StartLoading(id, messages[i])
	}

	// Verify all loaders are active
	for _, id := range ids {
		if !ls.IsLoading(id) {
			t.Errorf("Expected loader %s to be active", id)
		}
	}

	// Render should include all loaders
	width := 80
	rendered := ls.Render(width)

	for _, message := range messages {
		if !strings.Contains(rendered, message) {
			t.Errorf("Expected rendered output to contain message %q", message)
		}
	}
}

func TestLoaderCleanup(t *testing.T) {
	ls := NewLoadingSystem()
	id := "test-cleanup"

	ls.StartLoading(id, "Loading...")

	// Verify loader exists
	if !ls.IsLoading(id) {
		t.Fatal("Expected loader to exist")
	}

	// Complete and wait for cleanup
	ls.CompleteLoading(id, "Done")

	// Force cleanup by waiting longer than the 2-second delay
	time.Sleep(2100 * time.Millisecond)
	ls.Update()

	// Loader should be cleaned up
	if ls.IsLoading(id) {
		t.Error("Expected loader to be cleaned up after completion delay")
	}
}
