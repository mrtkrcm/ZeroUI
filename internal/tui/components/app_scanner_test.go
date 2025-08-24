package components

import (
	"testing"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/config"
)

func TestAppScanner(t *testing.T) {
	// Create a new scanner
	scanner := NewAppScanner()
	
	// Initialize it
	cmd := scanner.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
	
	// Test window resize
	scanner.Update(tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	})
	
	// Verify initial state
	if !scanner.IsScanning() {
		t.Error("Expected scanner to be scanning after init")
	}
	
	// Test the view renders without panic
	view := scanner.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	
	// Simulate scan completion
	scanner.Update(ScanCompleteMsg{
		Apps: []AppInfo{
			{
				Name:         "zeroui",
				Icon:         "○",
				Status:       StatusReady,
				ConfigPath:   "~/.config/zeroui/config.yml",
				ConfigExists: true,
			},
			{
				Name:   "ghostty",
				Icon:   "◉",
				Status: StatusNotConfigured,
			},
		},
	})
	
	// Verify scanning stopped
	if scanner.IsScanning() {
		t.Error("Expected scanner to stop scanning after completion")
	}
	
	// Check apps were stored
	apps := scanner.GetApps()
	if len(apps) != 2 {
		t.Errorf("Expected 2 apps, got %d", len(apps))
	}
	
	// Test FindApp
	if app, ok := scanner.FindApp("zeroui"); !ok {
		t.Error("Expected to find zeroui app")
	} else if app.Status != StatusReady {
		t.Error("Expected zeroui to have Ready status")
	}
	
	// Test results view
	view = scanner.View()
	if view == "" {
		t.Error("Expected non-empty results view")
	}
}

func TestAppRegistry(t *testing.T) {
	// Load the registry
	registry, err := config.LoadAppsRegistry()
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}
	
	// Check we have apps
	apps := registry.GetAllApps()
	if len(apps) == 0 {
		t.Error("Expected registry to contain apps")
	}
	
	// Check ZeroUI is first
	if len(apps) > 0 && apps[0].Name != "zeroui" {
		t.Errorf("Expected first app to be zeroui, got %s", apps[0].Name)
	}
	
	// Check categories exist
	categories := registry.GetCategories()
	if len(categories) == 0 {
		t.Error("Expected registry to contain categories")
	}
	
	// Test GetApp
	if app, ok := registry.GetApp("ghostty"); !ok {
		t.Error("Expected to find ghostty in registry")
	} else {
		if app.Icon != "◉" {
			t.Errorf("Expected ghostty icon to be ◉, got %s", app.Icon)
		}
	}
}

func TestScanPerformance(t *testing.T) {
	scanner := NewAppScanner()
	
	start := time.Now()
	cmd := scanner.Init()
	
	// Execute the scan command
	if cmd != nil {
		// In a real app, this would be handled by the Bubble Tea runtime
		// Here we just check it returns quickly
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			t.Errorf("Init took too long: %v", elapsed)
		}
	}
}