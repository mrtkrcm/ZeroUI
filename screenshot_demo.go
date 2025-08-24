package main

import (
	"encoding/json"
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

	// Create index file
	if err := createIndexFile(captureDir, scenarios); err != nil {
		fmt.Printf("❌ Failed to create index: %v\n", err)
	} else {
		fmt.Println("✅ Created index file")
	}

	fmt.Println("\n🎉 Screenshot demo completed!")
	fmt.Printf("📁 Check results in: %s\n", captureDir)
	fmt.Println("\n📋 Generated files:")
	files, _ := filepath.Glob(filepath.Join(captureDir, "*"))
	for _, file := range files {
		fmt.Printf("   📄 %s\n", filepath.Base(file))
	}
}

func captureScreen(captureDir, name, description string, model *MockModel, actions []string) error {
	// Get screen content
	screenText := model.View()

	// Create capture data
	capture := map[string]interface{}{
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"test_name":    "manual_screenshot_demo",
		"description":  description,
		"screen_size":  map[string]int{"width": model.width, "height": model.height},
		"state":        model.state,
		"current_app":  model.currentApp,
		"user_actions": actions,
		"screen_text":  screenText,
		"metadata": map[string]interface{}{
			"showingHelp": model.showingHelp,
			"hasError":    model.err != nil,
		},
	}

	// Add error info if present
	if model.err != nil {
		capture["error"] = model.err.Error()
	}

	// Save as JSON
	jsonPath := filepath.Join(captureDir, fmt.Sprintf("%s.json", name))
	jsonData, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save JSON: %w", err)
	}

	// Save as text
	txtPath := filepath.Join(captureDir, fmt.Sprintf("%s.txt", name))
	if err := os.WriteFile(txtPath, []byte(screenText), 0644); err != nil {
		return fmt.Errorf("failed to save text: %w", err)
	}

	// Save as HTML
	htmlPath := filepath.Join(captureDir, fmt.Sprintf("%s.html", name))
	htmlContent := generateHTML(capture)
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to save HTML: %w", err)
	}

	return nil
}

func generateHTML(capture map[string]interface{}) string {
	description := capture["description"].(string)
	testName := capture["test_name"].(string)
	timestamp := capture["timestamp"].(string)
	screenSize := capture["screen_size"].(map[string]int)
	state := capture["state"].(string)
	currentApp := capture["current_app"].(string)
	metadata := capture["metadata"].(map[string]interface{})
	screenText := capture["screen_text"].(string)
	userActions := capture["user_actions"].([]string)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>ZeroUI Screen Capture - %s</title>
    <style>
        body { font-family: 'Courier New', monospace; background: #1a1a1a; color: #f0f0f0; margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2d2d2d; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
        .screen { background: #000000; color: #00ff00; padding: 20px; border-radius: 8px; font-family: 'Courier New', monospace; font-size: 14px; line-height: 1.2; white-space: pre; overflow-x: auto; border: 1px solid #555; }
        .metadata { background: #3d3d3d; padding: 15px; border-radius: 8px; margin-bottom: 20px; font-size: 14px; }
        .actions { background: #4d4d4d; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
        .action { background: #2d2d2d; padding: 5px 10px; margin: 5px 0; border-radius: 4px; font-family: monospace; }
        .footer { background: #2d2d2d; padding: 10px; border-radius: 8px; margin-top: 20px; font-size: 12px; color: #ccc; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ZeroUI Screen Capture</h1>
            <h2>%s</h2>
            <p><strong>Test:</strong> %s</p>
            <p><strong>Timestamp:</strong> %s</p>
            <p><strong>Size:</strong> %dx%d</p>
            <p><strong>State:</strong> %s</p>
            <p><strong>Current App:</strong> %s</p>
            <p><strong>Showing Help:</strong> %v</p>
        </div>

        <div class="actions">
            <h3>User Actions</h3>`, 
		description, description, testName, timestamp, screenSize["width"], screenSize["height"], state, currentApp, metadata["showingHelp"])

	for _, action := range userActions {
		html += fmt.Sprintf(`
            <div class="action">%s</div>`, action)
	}

	html += fmt.Sprintf(`
        </div>

        <div class="screen">%s</div>

        <div class="footer">
            <p>Generated by ZeroUI screenshot system</p>
            <p>Test: %s | State: %s</p>
        </div>
    </div>
</body>
</html>`,
		strings.ReplaceAll(screenText, "\n", "<br>"),
		testName,
		state)

	return html
}

func createIndexFile(captureDir string, scenarios []struct {
	name        string
	description string
	model       *MockModel
	actions     []string
}) error {
	indexPath := filepath.Join(captureDir, "index.html")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>ZeroUI Screenshots - Manual Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .capture-list { display: grid; gap: 15px; }
        .capture-item { background: white; padding: 15px; border-radius: 8px; border: 1px solid #ddd; }
        .capture-item h3 { margin: 0 0 10px 0; color: #333; }
        .capture-meta { color: #666; font-size: 14px; }
        .capture-actions { margin-top: 10px; }
        .capture-actions a { margin-right: 15px; color: #007bff; text-decoration: none; }
        .capture-actions a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ZeroUI Screenshots - Manual Demo</h1>
            <p>Comprehensive screen capture demonstration</p>
        </div>
        <div class="capture-list">`

	for i, scenario := range scenarios {
		html += fmt.Sprintf(`
            <div class="capture-item">
                <h3>%d. %s</h3>
                <div class="capture-meta">
                    <strong>State:</strong> %s |
                    <strong>Size:</strong> %dx%d |
                    <strong>Time:</strong> %s
                </div>
                <div class="capture-actions">
                    <a href="%s.html">View HTML</a>
                    <a href="%s.txt">View Text</a>
                    <a href="%s.json">View JSON</a>
                </div>
            </div>`,
			i+1,
			scenario.description,
			scenario.model.state,
			scenario.model.width,
			scenario.model.height,
			time.Now().Format("15:04:05"),
			scenario.name,
			scenario.name,
			scenario.name)
	}

	html += `
        </div>
    </div>
</body>
</html>`

	return os.WriteFile(indexPath, []byte(html), 0644)
}
