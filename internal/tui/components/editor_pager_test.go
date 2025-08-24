package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditorPagerToggle(t *testing.T) {
	// Create enhanced config editor
	editor := NewEnhancedConfig("testapp")
	
	// Set some test fields
	fields := []ConfigField{
		{
			Key:         "theme",
			Description: "Color theme",
			Type:        FieldTypeSelect,
			Options:     []string{"dark", "light"},
			Value:       "dark",
		},
		{
			Key:         "font_size",
			Description: "Font size",
			Type:        FieldTypeInt,
			Value:       14,
		},
	}
	editor.SetFields(fields)
	
	// Set config file content
	configContent := `# Test Configuration
theme = dark
font_size = 14

[advanced]
debug = false`
	
	editor.SetConfigFile("/test/config.toml", configContent)
	
	// Initialize with proper size
	updatedEditor, _ := editor.Update(tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	})
	editor = updatedEditor
	
	// Test initial state (should show editor)
	view := editor.View()
	if !strings.Contains(view, "All") || !strings.Contains(view, "Modified") {
		t.Error("Initial view should show editor tabs")
	}
	
	// Toggle to source view using the actual key binding
	// We need to trigger handleKeyPress which checks key.Matches
	editor, _ = editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	
	// Check if we're showing source
	if !editor.showingSource {
		t.Error("showingSource should be true after pressing 'v'")
	}
	
	view = editor.View()
	
	// Should now show pager
	if !strings.Contains(view, "Viewing: /test/config.toml") {
		t.Errorf("After toggle, should show pager with file path. Got:\n%s", view)
	}
	
	if !strings.Contains(view, "# Test Configuration") {
		t.Errorf("Pager should display config content. Got:\n%s", view)
	}
	
	// Toggle back to editor
	editor, _ = editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	view = editor.View()
	
	// Should show editor again
	if !strings.Contains(view, "All") || !strings.Contains(view, "Modified") {
		t.Error("After toggling back, should show editor tabs")
	}
}

func TestEditorPagerKeyBindings(t *testing.T) {
	editor := NewEnhancedConfig("testapp")
	
	// Verify ViewSource key binding exists
	if editor.keys.ViewSource.Keys() == nil {
		t.Error("ViewSource key binding not configured")
	}
	
	// Check key binding includes "v"
	keys := editor.keys.ViewSource.Keys()
	hasV := false
	hasCtrlO := false
	for _, k := range keys {
		if k == "v" {
			hasV = true
		}
		if k == "ctrl+o" {
			hasCtrlO = true
		}
	}
	
	if !hasV {
		t.Error("ViewSource should be bound to 'v'")
	}
	
	if !hasCtrlO {
		t.Error("ViewSource should be bound to 'ctrl+o'")
	}
	
	// Check help text
	help := editor.keys.ViewSource.Help()
	if help.Key != "v/ctrl+o" || help.Desc != "view source" {
		t.Errorf("ViewSource help text incorrect: %+v", help)
	}
}