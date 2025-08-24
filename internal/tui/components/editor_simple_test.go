package components

import (
	"testing"
)

func TestEditorPagerSimple(t *testing.T) {
	// Create enhanced config editor
	editor := NewEnhancedConfig("testapp")
	
	// Verify pager is created
	if editor.pager == nil {
		t.Fatal("Pager should be initialized")
	}
	
	// Test toggling showingSource directly
	editor.showingSource = false
	
	// Toggle on
	editor.showingSource = true
	if !editor.showingSource {
		t.Error("showingSource should be true")
	}
	
	// Set config file
	editor.SetConfigFile("/test/config.toml", "# Test Config\nkey = value")
	
	// Verify config is set
	if editor.configFilePath != "/test/config.toml" {
		t.Error("Config file path not set")
	}
	
	if editor.configContent != "# Test Config\nkey = value" {
		t.Error("Config content not set")
	}
	
	// Test View when showing source
	editor.showingSource = true
	editor.width = 80
	editor.height = 24
	editor.pager.SetSize(80, 24)
	
	view := editor.View()
	
	// The view should call pager.View() when showingSource is true
	// We can't easily test the exact output, but we can verify it doesn't panic
	if view == "" {
		t.Error("View should return something when showing source")
	}
	
	// Test View when not showing source
	editor.showingSource = false
	editor.SetFields([]ConfigField{
		{Key: "test", Value: "value"},
	})
	
	view = editor.View()
	if view == "" {
		t.Error("View should return something when showing editor")
	}
}

func TestViewSourceKeyBinding(t *testing.T) {
	editor := NewEnhancedConfig("testapp")
	
	// Check that ViewSource key binding is defined
	binding := editor.keys.ViewSource
	
	// Check the keys
	keys := binding.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys for ViewSource, got %d", len(keys))
	}
	
	foundV := false
	foundCtrlO := false
	for _, k := range keys {
		if k == "v" {
			foundV = true
		}
		if k == "ctrl+o" {
			foundCtrlO = true
		}
	}
	
	if !foundV {
		t.Error("ViewSource should include 'v' key")
	}
	
	if !foundCtrlO {
		t.Error("ViewSource should include 'ctrl+o' key")
	}
}