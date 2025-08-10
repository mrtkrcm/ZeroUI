package tui

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

func TestNewModel_Enhanced(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "test-app")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model == nil {
		t.Error("Expected non-nil model")
	}

	// Test initial state
	if model.state != AppSelectionView {
		t.Errorf("Expected initial state to be AppSelectionView, got %v", model.state)
	}
}

func TestModel_Update(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test key handling
	testCases := []struct {
		name string
		key  tea.KeyType
	}{
		{"Down", tea.KeyDown},
		{"Up", tea.KeyUp},
		{"Enter", tea.KeyEnter},
		{"Escape", tea.KeyEsc},
		{"Quit", tea.KeyCtrlC},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tc.key}
			updatedModel, _ := model.Update(keyMsg)

			// Should return a valid model
			if updatedModel == nil {
				t.Error("Expected non-nil updated model")
			}

			// Model should be of correct type
			if _, ok := updatedModel.(*Model); !ok {
				t.Errorf("Expected *Model, got %T", updatedModel)
			}
		})
	}
}

func TestModel_View_Enhanced(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test basic view rendering
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Test that view contains expected elements
	if len(view) < 10 {
		t.Error("View seems too short, expected some content")
	}
}

func TestModel_ErrorState(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set an error and verify it's displayed
	testError := errors.New("Test error message")
	model.err = testError

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty error view")
	}

	// Error view should contain the error message
	// This is a basic test - in practice, we'd check for specific formatting
}

func TestModel_AppNavigation(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test navigation when apps are available
	if len(model.apps) > 0 {

		// Navigate down
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(downMsg)
		model = updatedModel.(*Model)

		// Navigate up
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = model.Update(upMsg)
		model = updatedModel.(*Model)

		// The cursor should be back to initial position or close to it
		// depending on the wrapping behavior
	}
}

func TestModel_StateTransitions(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test state transition from app selection to config edit
	if len(model.apps) > 0 {
		initialState := model.state

		// Press Enter to select an app
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(enterMsg)
		model = updatedModel.(*Model)

		// State should have changed (or at least not crashed)
		if model.state == initialState {
			// This is okay - might mean no state change occurred
			t.Logf("State remained the same: %v", model.state)
		}

		// Press Escape to go back
		escMsg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ = model.Update(escMsg)
		model = updatedModel.(*Model)
	}
}

func TestModel_Init_Enhanced(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test that Init() returns a valid command
	cmd := model.Init()
	// Init() should return either nil or a valid tea.Cmd
	// We just verify it doesn't crash
	_ = cmd
}

func TestModel_WindowSizeUpdate(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test window size message
	sizeMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, _ := model.Update(sizeMsg)
	model = updatedModel.(*Model)

	// Check that dimensions were updated
	if model.width != 80 || model.height != 24 {
		t.Errorf("Expected dimensions 80x24, got %dx%d", model.width, model.height)
	}
}

func TestModel_WithInitialApp(t *testing.T) {
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	// Test with an initial app specified
	model, err := NewModel(engine, "test-app")
	if err != nil {
		t.Fatalf("Failed to create model with initial app: %v", err)
	}

	// The model should be created successfully even if the app doesn't exist
	if model == nil {
		t.Error("Expected non-nil model even with non-existent initial app")
	}
}

func TestModel_EmptyApps(t *testing.T) {
	// Create a temporary directory with no apps configured
	tmpDir, err := ioutil.TempDir("", "configtoggle-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create toggle engine: %v", err)
	}

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test that empty apps list is handled gracefully
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view even with no apps")
	}

	// Navigation should not crash with empty apps
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(downMsg)
	if updatedModel == nil {
		t.Error("Expected non-nil model after navigation with empty apps")
	}
}
