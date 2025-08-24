package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ScreenCapture represents a comprehensive screen capture with metadata
type ScreenCapture struct {
	Timestamp   time.Time              `json:"timestamp"`
	TestName    string                 `json:"test_name"`
	Description string                 `json:"description"`
	ScreenSize  ScreenSize             `json:"screen_size"`
	State       string                 `json:"state"`
	CurrentApp  string                 `json:"current_app,omitempty"`
	UserActions []string               `json:"user_actions,omitempty"`
	ScreenText  string                 `json:"screen_text"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ScreenSize represents terminal dimensions
type ScreenSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ScreenCapturer manages screen captures during testing
type ScreenCapturer struct {
	testName    string
	captures    []ScreenCapture
	captureDir  string
	userActions []string
}

// NewScreenCapturer creates a new screen capturer for a test
func NewScreenCapturer(t *testing.T, testName string) *ScreenCapturer {
	captureDir := filepath.Join("testdata", "screenshots", testName)
	return &ScreenCapturer{
		testName:    testName,
		captures:    []ScreenCapture{},
		captureDir:  captureDir,
		userActions: []string{},
	}
}

// AddAction records a user action for the next capture
func (sc *ScreenCapturer) AddAction(action string) {
	sc.userActions = append(sc.userActions, action)
}

// CaptureScreen captures the current screen state with metadata
func (sc *ScreenCapturer) CaptureScreen(t *testing.T, model *Model, description string) {
	t.Helper()

	// Create capture directory
	if err := os.MkdirAll(sc.captureDir, 0755); err != nil {
		t.Fatalf("Failed to create capture directory: %v", err)
	}

	// Get screen content
	screenText := model.View()

	// Determine state name
	stateName := "unknown"
	switch model.state {
	case ListView:
		stateName = "list_view"
	case FormView:
		stateName = "form_view"
	case HelpView:
		stateName = "help_view"
	}

	// Create screen capture
	capture := ScreenCapture{
		Timestamp:   time.Now(),
		TestName:    sc.testName,
		Description: description,
		ScreenSize: ScreenSize{
			Width:  model.width,
			Height: model.height,
		},
		State:       stateName,
		CurrentApp:  model.currentApp,
		UserActions: append([]string(nil), sc.userActions...), // copy
		ScreenText:  screenText,
		Metadata: map[string]interface{}{
			"showingHelp":    model.showingHelp,
			"hasError":       model.err != nil,
			"componentCount": 5, // Fixed count of main components
		},
	}

	// Add error info if present
	if model.err != nil {
		capture.Metadata["error"] = model.err.Error()
	}

	sc.captures = append(sc.captures, capture)
	sc.userActions = []string{} // reset for next capture

	// Save individual files
	sc.saveCapture(t, capture)

	// Save HTML version for better viewing
	sc.saveHTMLCapture(t, capture)
}

// SaveCapture saves a single capture to files
func (sc *ScreenCapturer) saveCapture(t *testing.T, capture ScreenCapture) {
	t.Helper()

	// Save as JSON for structured data
	jsonPath := filepath.Join(sc.captureDir, fmt.Sprintf("%d_%s.json",
		len(sc.captures), strings.ReplaceAll(capture.Description, " ", "_")))

	jsonData, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal capture JSON: %v", err)
		return
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		t.Errorf("Failed to save capture JSON: %v", err)
		return
	}

	// Save screen text for easy viewing
	txtPath := filepath.Join(sc.captureDir, fmt.Sprintf("%d_%s.txt",
		len(sc.captures), strings.ReplaceAll(capture.Description, " ", "_")))

	if err := os.WriteFile(txtPath, []byte(capture.ScreenText), 0644); err != nil {
		t.Errorf("Failed to save capture text: %v", err)
	}

	t.Logf("Screen captured: %s", txtPath)
}

// SaveHTMLCapture creates an HTML version for better viewing
func (sc *ScreenCapturer) saveHTMLCapture(t *testing.T, capture ScreenCapture) {
	t.Helper()

	htmlPath := filepath.Join(sc.captureDir, fmt.Sprintf("%d_%s.html",
		len(sc.captures), strings.ReplaceAll(capture.Description, " ", "_")))

	htmlContent := sc.generateHTML(capture)

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Errorf("Failed to save HTML capture: %v", err)
	}
}

// GenerateHTML creates an HTML representation of the screen capture
func (sc *ScreenCapturer) generateHTML(capture ScreenCapture) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>ZeroUI Screen Capture - %s</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: #1a1a1a;
            color: #f0f0f0;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: #2d2d2d;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .metadata {
            background: #3d3d3d;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            font-size: 14px;
        }
        .screen {
            background: #000000;
            color: #00ff00;
            padding: 20px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.2;
            white-space: pre;
            overflow-x: auto;
            border: 1px solid #555;
        }
        .actions {
            background: #4d4d4d;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .action {
            background: #2d2d2d;
            padding: 5px 10px;
            margin: 5px 0;
            border-radius: 4px;
            font-family: monospace;
        }
        .footer {
            background: #2d2d2d;
            padding: 10px;
            border-radius: 8px;
            margin-top: 20px;
            font-size: 12px;
            color: #ccc;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ZeroUI Screen Capture</h1>
            <h2>%s</h2>
            <p><strong>Test:</strong> %s</p>
            <p><strong>Timestamp:</strong> %s</p>
        </div>

        <div class="metadata">
            <h3>Screen Information</h3>
            <p><strong>Size:</strong> %dx%d</p>
            <p><strong>State:</strong> %s</p>
            <p><strong>Current App:</strong> %s</p>
            <p><strong>Showing Help:</strong> %v</p>
        </div>`,
		capture.Description,
		capture.Description,
		capture.TestName,
		capture.Timestamp.Format("2006-01-02 15:04:05"),
		capture.ScreenSize.Width,
		capture.ScreenSize.Height,
		capture.State,
		capture.CurrentApp,
		func() string {
			if showingHelp, ok := capture.Metadata["showingHelp"].(bool); ok {
				return fmt.Sprintf("%v", showingHelp)
			}
			return "false"
		}())

	if len(capture.UserActions) > 0 {
		html += `
        <div class="actions">
            <h3>User Actions</h3>`
		for _, action := range capture.UserActions {
			html += fmt.Sprintf(`
            <div class="action">%s</div>`, action)
		}
		html += `
        </div>`
	}

	// Clean up the screen text for better HTML display
	cleanScreenText := strings.TrimSpace(capture.ScreenText)
	// Remove excessive empty lines but keep some structure
	lines := strings.Split(cleanScreenText, "\n")
	var cleanedLines []string
	emptyLineCount := 0
	maxEmptyLines := 2 // Allow max 2 consecutive empty lines

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "" {
			emptyLineCount++
			if emptyLineCount <= maxEmptyLines {
				cleanedLines = append(cleanedLines, "")
			}
		} else {
			emptyLineCount = 0
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	// Join back with <br> for HTML
	cleanedHTML := strings.Join(cleanedLines, "<br>")

	html += fmt.Sprintf(`
        <div class="screen">%s</div>

        <div class="footer">
            <p>Generated by ZeroUI test suite</p>
            <p>Test: %s | State: %s</p>
        </div>
    </div>
</body>
</html>`,
		cleanedHTML,
		capture.TestName,
		capture.State)

	return html
}

// SaveSummary creates a summary of all captures for the test
func (sc *ScreenCapturer) SaveSummary(t *testing.T) {
	t.Helper()

	summaryPath := filepath.Join(sc.captureDir, "summary.json")

	summary := map[string]interface{}{
		"testName":      sc.testName,
		"totalCaptures": len(sc.captures),
		"captures":      sc.captures,
	}

	summaryData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal summary: %v", err)
		return
	}

	if err := os.WriteFile(summaryPath, summaryData, 0644); err != nil {
		t.Errorf("Failed to save summary: %v", err)
		return
	}

	t.Logf("Screenshot summary saved: %s (%d captures)", summaryPath, len(sc.captures))
}

// GenerateIndex creates an HTML index of all captures
func (sc *ScreenCapturer) GenerateIndex(t *testing.T) {
	t.Helper()

	indexPath := filepath.Join(sc.captureDir, "index.html")

	html := `<!DOCTYPE html>
<html>
<head>
    <title>ZeroUI Screenshots - ` + sc.testName + `</title>
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
            <h1>ZeroUI Screenshots - ` + sc.testName + `</h1>
            <p>Total captures: ` + fmt.Sprintf("%d", len(sc.captures)) + `</p>
        </div>
        <div class="capture-list">`

	for i, capture := range sc.captures {
		html += fmt.Sprintf(`
            <div class="capture-item">
                <h3>%d. %s</h3>
                <div class="capture-meta">
                    <strong>State:</strong> %s |
                    <strong>Size:</strong> %dx%d |
                    <strong>Time:</strong> %s
                </div>
                <div class="capture-actions">
                    <a href="%d_%s.html">View HTML</a>
                    <a href="%d_%s.txt">View Text</a>
                    <a href="%d_%s.json">View JSON</a>
                </div>
            </div>`,
			i+1,
			capture.Description,
			capture.State,
			capture.ScreenSize.Width,
			capture.ScreenSize.Height,
			capture.Timestamp.Format("15:04:05"),
			i+1, strings.ReplaceAll(capture.Description, " ", "_"),
			i+1, strings.ReplaceAll(capture.Description, " ", "_"),
			i+1, strings.ReplaceAll(capture.Description, " ", "_"))
	}

	html += `
        </div>
    </div>
</body>
</html>`

	if err := os.WriteFile(indexPath, []byte(html), 0644); err != nil {
		t.Errorf("Failed to save index: %v", err)
		return
	}

	t.Logf("Screenshot index created: %s", indexPath)
}
