package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MockModel represents a simple mock model for screenshot testing
type MockModel struct {
	width       int
	height      int
	state       string
	currentApp  string
	showingHelp bool
	err         error
}

// View returns a mock screen representation
func (m *MockModel) View() string {
	switch m.state {
	case "form":
		return fmt.Sprintf(`╔══════════════════════════════════════════════════════════╗
║               Configuration Editor                      ║
╚══════════════════════════════════════════════════════════╝

App: %s
Setting: font-family
Value: JetBrains Mono

[Save] [Cancel] [Help]

════════════════════════════════════════════════════════════`, m.currentApp)
	case "help":
		return `┌─────────────────────────────────────────────────────────────┐
│ Help: Keyboard Shortcuts                                     │
│                                                             │
│ Navigation:                                                 │
│   ↑↓      Move selection                                    │
│   Enter    Select/Confirm                                   │
│   Esc      Cancel/Back                                      │
│   ?        Show this help                                   │
│   q        Quit application                                 │
│                                                             │
│ [Close Help]                                                │
└─────────────────────────────────────────────────────────────┘`
	default:
		return `╔══════════════════════════════════════════════════════════╗
║                 ZeroUI Applications                     ║
╚══════════════════════════════════════════════════════════╝

📋 Available Applications (3)

  • ghostty     Terminal emulator     [Active]
  • zed         Code editor          [Active] 
  • mise        Tool version manager [Active]

📖 Navigation: ↑↓ Select • Enter Choose • ? Help • q Quit

════════════════════════════════════════════════════════════`
	}
}

func main() {
	fmt.Println("🎯 ZeroUI Enhanced Screenshot System Demo")
	fmt.Println("=========================================")

	// Create capture directory
	captureDir := filepath.Join("testdata", "screenshots", "manual_demo")
	if err := os.MkdirAll(captureDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create directory: %v\n", err)
		return
	}

	fmt.Printf("📁 Created directory: %s\n", captureDir)

	// Test scenarios
	scenarios := []struct {
		name        string
		description string
		model       *MockModel
		actions     []string
	}{
		{
			name:        "initial_grid",
			description: "Initial Application Grid",
			model: &MockModel{
				width:      120,
				height:     40,
				state:      "list",
				currentApp: "",
			},
			actions: []string{"Application start"},
		},
		{
			name:        "app_selection",
			description: "App Selection View",
			model: &MockModel{
				width:      120,
				height:     40,
				state:      "list",
				currentApp: "",
			},
			actions: []string{"Navigate to apps", "Select ghostty"},
		},
		{
			name:        "config_editor",
			description: "Configuration Editor",
			model: &MockModel{
				width:       120,
				height:      40,
				state:       "form",
				currentApp:  "ghostty",
				showingHelp: false,
			},
			actions: []string{"Select ghostty", "Enter config mode"},
		},
		{
			name:        "help_overlay",
			description: "Help Overlay",
			model: &MockModel{
				width:       120,
				height:      40,
				state:       "help",
				currentApp:  "",
				showingHelp: true,
			},
			actions: []string{"Press '?' for help"},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n📸 Capturing screen %d: %s\n", i+1, scenario.description)
		
		if err := captureScreen(captureDir, scenario.name, scenario.description, scenario.model, scenario.actions); err != nil {
			fmt.Printf("❌ Failed to capture %s: %v\n", scenario.name, err)
		} else {
			fmt.Printf("✅ Captured: %s\n", scenario.name)
		}
	}

	fmt.Println("\n🎉 Screenshot demo completed!")
	fmt.Printf("📁 Check results in: %s\n", captureDir)
	fmt.Println("\n📋 Generated files:")
	files, _ := filepath.Glob(filepath.Join(captureDir, "*"))
	for _, file := range files {
		fmt.Printf("   📄 %s\n", filepath.Base(file))
	}

	fmt.Println("\n🌐 To view screenshots:")
	fmt.Printf("   📂 Open: http://localhost:8000/screenshot_viewer.html\n")
	fmt.Printf("   📁 Files: %s\n", captureDir)
}

func captureScreen(captureDir, name, description string, model *MockModel, actions []string) error {
	// Get screen content
	screenText := model.View()

	// Create YAML frontmatter
	yamlContent := fmt.Sprintf(`---
current_app: "%s"
description: "%s"
metadata:
  hasError: %v
  showingHelp: %v
screen_size:
  height: %d
  width: %d
state: "%s"
test_name: "manual_screenshot_demo"
timestamp: "%s"
user_actions:
%s
---
%s`,
		model.currentApp,
		description,
		model.err != nil,
		model.showingHelp,
		model.height,
		model.width,
		model.state,
		time.Now().Format("2006-01-02 15:04:05"),
		formatActions(actions),
		screenText,
	)

	// Save as text with YAML frontmatter
	txtPath := filepath.Join(captureDir, fmt.Sprintf("%s.txt", name))
	if err := os.WriteFile(txtPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to save text file: %w", err)
	}

	return nil
}

// formatActions formats the actions array for YAML
func formatActions(actions []string) string {
	if len(actions) == 0 {
		return "  - \"No actions\""
	}

	var result strings.Builder
	for _, action := range actions {
		result.WriteString(fmt.Sprintf("  - \"%s\"\n", action))
	}
	return result.String()
}


