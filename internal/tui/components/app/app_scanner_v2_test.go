package appcomponents

import (
	"fmt"
	"strings"
	"testing"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

func TestAppScannerV2(t *testing.T) {
	t.Run("Initialization", func(t *testing.T) {
		scanner := NewAppScannerV2()
		
		if scanner.GetState() != ScannerIdle {
			t.Errorf("Expected initial state to be Idle, got %v", scanner.GetState())
		}
		
		cmd := scanner.Init()
		if cmd == nil {
			t.Error("Expected Init to return a command")
		}
		
		if scanner.GetState() != ScannerScanning {
			t.Error("Expected state to be Scanning after Init")
		}
	})
	
	t.Run("WindowResize", func(t *testing.T) {
		scanner := NewAppScannerV2()
		
		scanner.Update(tea.WindowSizeMsg{
			Width:  100,
			Height: 30,
		})
		
		if scanner.width != 100 || scanner.height != 30 {
			t.Error("Window size not updated correctly")
		}
	})
	
	t.Run("ScanComplete", func(t *testing.T) {
		scanner := NewAppScannerV2()
		scanner.Init()
		
		// Simulate scan completion
		scanner.Update(ScanCompleteMsg{
			Apps: []AppInfo{
				{Name: "app1", ConfigExists: true},
				{Name: "app2", ConfigExists: false},
			},
		})
		
		if !scanner.IsComplete() {
			t.Error("Expected scanner to be complete")
		}
		
		apps := scanner.GetApps()
		if len(apps) != 2 {
			t.Errorf("Expected 2 apps, got %d", len(apps))
		}
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		scanner := NewAppScannerV2()
		scanner.Init()
		
		// Simulate error
		scanner.Update(scanErrorMsg{
			error: fmt.Errorf("test error"),
		})
		
		if scanner.GetState() != ScannerError {
			t.Error("Expected scanner to be in error state")
		}
		
		view := scanner.View()
		if !strings.Contains(view, "Scan Failed") {
			t.Error("Error view should show failure message")
		}
	})
	
	t.Run("ViewRendering", func(t *testing.T) {
		scanner := NewAppScannerV2()
		
		// Test idle view
		view := scanner.View()
		if view == "" {
			t.Error("Expected non-empty idle view")
		}
		
		// Test scanning view
		scanner.Init()
		view = scanner.View()
		if !strings.Contains(view, "Scanning") {
			t.Error("Scanning view should show scanning message")
		}
		
		// Test complete view
		scanner.Update(ScanCompleteMsg{
			Apps: []AppInfo{
				{Name: "zeroui", Icon: "○", ConfigExists: true},
				{Name: "ghostty", Icon: "◉", ConfigExists: false},
			},
		})
		view = scanner.View()
		if !strings.Contains(view, "Configured") {
			t.Error("Complete view should show configured apps")
		}
	})
	
	t.Run("Reset", func(t *testing.T) {
		scanner := NewAppScannerV2()
		scanner.Init()
		
		scanner.Update(ScanCompleteMsg{
			Apps: []AppInfo{{Name: "test"}},
		})
		
		scanner.Reset()
		
		if scanner.GetState() != ScannerIdle {
			t.Error("Expected state to be Idle after reset")
		}
		
		if len(scanner.GetApps()) != 0 {
			t.Error("Expected apps to be cleared after reset")
		}
	})
}

func TestConcurrentScanner(t *testing.T) {
	t.Run("ConcurrentScanning", func(t *testing.T) {
		scanner := NewConcurrentScanner()
		
		// Start scan
		cmd := scanner.ScanAll()
		if cmd == nil {
			t.Error("Expected ScanAll to return a command")
		}
		
		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)
		
		// Stop the scanner
		scanner.Stop()
	})
	
	t.Run("Performance", func(t *testing.T) {
		scanner := NewConcurrentScanner()
		
		start := time.Now()
		cmd := scanner.ScanAll()
		
		// Execute in a goroutine
		go func() {
			if cmd != nil {
				_ = cmd()
			}
		}()
		
		// Wait a bit
		time.Sleep(100 * time.Millisecond)
		scanner.Stop()
		
		elapsed := time.Since(start)
		if elapsed > 5*time.Second {
			t.Errorf("Scanning took too long: %v", elapsed)
		}
	})
}

