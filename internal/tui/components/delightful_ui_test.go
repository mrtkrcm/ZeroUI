package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

func TestDelightfulUI_AppActivation(t *testing.T) {
	// Create a DelightfulUI model with test apps
	model := NewDelightfulUI()

	// Set some test apps
	model.apps = []registry.AppStatus{
		{
			Definition: registry.AppDefinition{
				Name: "test-app1",
				Logo: "🧪",
			},
			IsInstalled: true,
			HasConfig:   true,
		},
		{
			Definition: registry.AppDefinition{
				Name: "test-app2",
				Logo: "🔧",
			},
			IsInstalled: true,
			HasConfig:   false,
		},
	}

	// Test app activation with valid selection
	t.Run("Valid app activation", func(t *testing.T) {
		model.selectedIndex = 0
		cmd := model.activateApp()

		if cmd == nil {
			t.Fatal("Expected command from activateApp, got nil")
		}

		// Execute the command to get the message
		msg := cmd()
		appMsg, ok := msg.(AppSelectedMsg)
		if !ok {
			t.Fatalf("Expected AppSelectedMsg, got %T", msg)
		}

		if appMsg.App != "test-app1" {
			t.Errorf("Expected app 'test-app1', got '%s'", appMsg.App)
		}
	})

	// Test app activation with empty apps
	t.Run("Empty apps list", func(t *testing.T) {
		model.apps = []registry.AppStatus{}
		cmd := model.activateApp()

		if cmd != nil {
			t.Error("Expected nil command for empty apps list")
		}
	})

	// Test app activation with invalid index
	t.Run("Invalid index", func(t *testing.T) {
		model.apps = []registry.AppStatus{
			{Definition: registry.AppDefinition{Name: "test-app"}},
		}
		model.selectedIndex = 5 // Out of bounds
		cmd := model.activateApp()

		if cmd != nil {
			t.Error("Expected nil command for invalid index")
		}
	})
}

func TestDelightfulUI_KeyHandling(t *testing.T) {
	model := NewDelightfulUI()
	model.apps = []registry.AppStatus{
		{Definition: registry.AppDefinition{Name: "test-app"}},
	}

	// Test enter key activation
	t.Run("Enter key activation", func(t *testing.T) {
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Error("Expected command from enter key press")
		}
	})

	// Test space key activation
	t.Run("Space key activation", func(t *testing.T) {
		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})

		if cmd == nil {
			t.Error("Expected command from space key press")
		}
	})
}

func TestDelightfulUI_Navigation(t *testing.T) {
	model := NewDelightfulUI()
	model.apps = []registry.AppStatus{
		{Definition: registry.AppDefinition{Name: "app1"}},
		{Definition: registry.AppDefinition{Name: "app2"}},
		{Definition: registry.AppDefinition{Name: "app3"}},
		{Definition: registry.AppDefinition{Name: "app4"}},
	}

	// Test vertical navigation
	t.Run("Down navigation", func(t *testing.T) {
		initialIndex := model.selectedIndex
		model.Update(tea.KeyMsg{Type: tea.KeyDown})

		if model.selectedIndex == initialIndex {
			t.Error("Expected selectedIndex to change after down navigation")
		}
	})

	t.Run("Up navigation", func(t *testing.T) {
		model.selectedIndex = 1
		model.Update(tea.KeyMsg{Type: tea.KeyUp})

		if model.selectedIndex != 0 {
			t.Errorf("Expected selectedIndex 0, got %d", model.selectedIndex)
		}
	})

	// Test navigation wrapping
	t.Run("Navigation wrapping", func(t *testing.T) {
		model.selectedIndex = 0
		model.Update(tea.KeyMsg{Type: tea.KeyUp}) // Should wrap to last item

		expectedIndex := len(model.apps) - 1
		if model.selectedIndex != expectedIndex {
			t.Errorf("Expected selectedIndex %d, got %d", expectedIndex, model.selectedIndex)
		}
	})
}

func TestDelightfulUI_ThemeCycling(t *testing.T) {
	model := NewDelightfulUI()

	// Get initial theme
	initialTheme := model.styles

	t.Run("Theme cycling works", func(t *testing.T) {
		// Cycle theme
		model.cycleTheme()

		// Styles should be updated
		if model.styles == initialTheme {
			t.Error("Expected styles to change after theme cycling")
		}
	})

	t.Run("Tab key triggers theme cycling", func(t *testing.T) {
		initialThemeName := styles.GetCurrentThemeName()

		model.Update(tea.KeyMsg{Type: tea.KeyTab})

		newThemeName := styles.GetCurrentThemeName()
		if newThemeName == initialThemeName {
			t.Error("Expected theme to change after Tab key press")
		}
	})

	t.Run("All themes cycle correctly", func(t *testing.T) {
		themeNames := styles.GetThemeNames()
		seenThemes := make(map[string]bool)

		// Cycle through all themes
		for i := 0; i < len(themeNames)+1; i++ {
			currentTheme := styles.GetCurrentThemeName()
			seenThemes[currentTheme] = true
			model.cycleTheme()
		}

		// Should have seen all themes
		if len(seenThemes) != len(themeNames) {
			t.Errorf("Expected to see %d themes, saw %d", len(themeNames), len(seenThemes))
		}
	})
}
