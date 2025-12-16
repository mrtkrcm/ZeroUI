package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components/core"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// ScreenshotComponent represents a component for capturing and managing TUI screenshots
type ScreenshotComponent struct {
	*core.BaseComponent
	styles         *styles.Styles
	captureDir     string
	currentCapture *ScreenshotData
	isCapturing    bool
}

// ScreenshotData represents a single screenshot capture
type ScreenshotData struct {
	Timestamp   string                 `json:"timestamp"`
	Description string                 `json:"description"`
	TestName    string                 `json:"test_name"`
	ScreenSize  map[string]int         `json:"screen_size"`
	State       string                 `json:"state"`
	CurrentApp  string                 `json:"current_app"`
	UserActions []string               `json:"user_actions"`
	ScreenText  string                 `json:"screen_text"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
}

// NewScreenshotComponent creates a new screenshot component
func NewScreenshotComponent(captureDir string) *ScreenshotComponent {
	component := &ScreenshotComponent{
		BaseComponent: core.NewBaseComponent("screenshot"),
		styles:        styles.GetStyles(),
		captureDir:    captureDir,
		isCapturing:   false,
	}

	// Ensure capture directory exists
	if err := os.MkdirAll(captureDir, 0o755); err != nil {
		// Log error but don't fail - component can still function
		fmt.Printf("Warning: Failed to create screenshot directory: %v\n", err)
	}

	return component
}

// SetSize implements core.Sizeable
func (s *ScreenshotComponent) SetSize(width, height int) tea.Cmd {
	s.BaseComponent.SetSize(width, height)
	return nil
}

// Capture takes a screenshot of the provided model
func (s *ScreenshotComponent) Capture(model tea.Model, description, testName string, actions ...string) error {
	s.isCapturing = true
	defer func() { s.isCapturing = false }()

	// Get screen content safely
	var screenText string
	if viewModel, ok := model.(interface{ View() string }); ok {
		screenText = viewModel.View()
	} else {
		return fmt.Errorf("model does not implement View() method")
	}

	// Extract state information from the model if possible
	state := "unknown"
	currentApp := ""
	var metadata map[string]interface{}

	// Try to extract information from a ZeroUI model
	if zeroUIModel, ok := model.(interface {
		GetState() string
		GetCurrentApp() string
		GetWidth() int
		GetHeight() int
		IsShowingHelp() bool
		GetError() error
	}); ok {
		state = zeroUIModel.GetState()
		currentApp = zeroUIModel.GetCurrentApp()
		metadata = map[string]interface{}{
			"showingHelp": zeroUIModel.IsShowingHelp(),
			"hasError":    zeroUIModel.GetError() != nil,
			"currentApp":  currentApp,
			"state":       state,
		}
	}

	// Create screenshot data
	width, height := s.GetSize()
	if width == 0 || height == 0 {
		width, height = 120, 40 // Default size
	}

	screenshot := &ScreenshotData{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		Description: description,
		TestName:    testName,
		ScreenSize: map[string]int{
			"width":  width,
			"height": height,
		},
		State:       state,
		CurrentApp:  currentApp,
		UserActions: actions,
		ScreenText:  screenText,
		Metadata:    metadata,
	}

	// Capture any error from the model
	if errorModel, ok := model.(interface{ GetError() error }); ok {
		if err := errorModel.GetError(); err != nil {
			screenshot.Error = err.Error()
		}
	}

	s.currentCapture = screenshot

	// Save the screenshot
	return s.saveScreenshot(screenshot)
}

// saveScreenshot saves the screenshot data to files
func (s *ScreenshotComponent) saveScreenshot(data *ScreenshotData) error {
	// Create test directory
	testDir := filepath.Join(s.captureDir, data.TestName)
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Create YAML frontmatter content
	yamlContent := s.formatYAMLFrontmatter(data)

	// Save as text with YAML frontmatter
	captureName := strings.ReplaceAll(data.Description, " ", "_")
	txtPath := filepath.Join(testDir, fmt.Sprintf("%s.txt", captureName))

	if err := os.WriteFile(txtPath, []byte(yamlContent), 0o644); err != nil {
		return fmt.Errorf("failed to save text file: %w", err)
	}

	// Save as JSON for structured data
	jsonPath := filepath.Join(testDir, fmt.Sprintf("%s.json", captureName))
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to save JSON file: %w", err)
	}

	return nil
}

// formatYAMLFrontmatter creates YAML frontmatter for the screenshot
func (s *ScreenshotComponent) formatYAMLFrontmatter(data *ScreenshotData) string {
	var yaml strings.Builder

	yaml.WriteString("---\n")
	yaml.WriteString(fmt.Sprintf("current_app: \"%s\"\n", data.CurrentApp))
	yaml.WriteString(fmt.Sprintf("description: \"%s\"\n", data.Description))

	if data.Metadata != nil {
		yaml.WriteString("metadata:\n")
		for key, value := range data.Metadata {
			yaml.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	yaml.WriteString(fmt.Sprintf("screen_size:\n"))
	yaml.WriteString(fmt.Sprintf("  height: %d\n", data.ScreenSize["height"]))
	yaml.WriteString(fmt.Sprintf("  width: %d\n", data.ScreenSize["width"]))
	yaml.WriteString(fmt.Sprintf("state: \"%s\"\n", data.State))
	yaml.WriteString(fmt.Sprintf("test_name: \"%s\"\n", data.TestName))
	yaml.WriteString(fmt.Sprintf("timestamp: \"%s\"\n", data.Timestamp))

	if len(data.UserActions) > 0 {
		yaml.WriteString("user_actions:\n")
		for _, action := range data.UserActions {
			yaml.WriteString(fmt.Sprintf("  - \"%s\"\n", action))
		}
	}

	yaml.WriteString("---\n")
	yaml.WriteString(data.ScreenText)

	return yaml.String()
}

// GetCurrentCapture returns the last captured screenshot data
func (s *ScreenshotComponent) GetCurrentCapture() *ScreenshotData {
	return s.currentCapture
}

// IsCapturing returns whether a capture is in progress
func (s *ScreenshotComponent) IsCapturing() bool {
	return s.isCapturing
}

// GetCaptureDir returns the capture directory
func (s *ScreenshotComponent) GetCaptureDir() string {
	return s.captureDir
}

// CaptureWithContext captures a screenshot with additional context from existing components
func (s *ScreenshotComponent) CaptureWithContext(model tea.Model, description, testName string, actions []string, context map[string]interface{}) error {
	s.isCapturing = true
	defer func() { s.isCapturing = false }()

	// Get screen content safely
	var screenText string
	if viewModel, ok := model.(interface{ View() string }); ok {
		screenText = viewModel.View()
	} else {
		return fmt.Errorf("model does not implement View() method")
	}

	// Extract state information from the model if possible
	state := "unknown"
	currentApp := ""
	var metadata map[string]interface{}

	// Try to extract information from a ZeroUI model
	if zeroUIModel, ok := model.(interface {
		GetState() string
		GetCurrentApp() string
		GetWidth() int
		GetHeight() int
		IsShowingHelp() bool
		GetError() error
	}); ok {
		state = zeroUIModel.GetState()
		currentApp = zeroUIModel.GetCurrentApp()
		metadata = map[string]interface{}{
			"showingHelp": zeroUIModel.IsShowingHelp(),
			"hasError":    zeroUIModel.GetError() != nil,
			"currentApp":  currentApp,
			"state":       state,
		}
	}

	// Merge with provided context
	if context != nil {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}
		for key, value := range context {
			metadata[key] = value
		}
	}

	// Get size from component or model
	width, height := s.GetSize()
	if width == 0 || height == 0 {
		if sizeModel, ok := model.(interface {
			GetWidth() int
			GetHeight() int
		}); ok {
			width = sizeModel.GetWidth()
			height = sizeModel.GetHeight()
		} else {
			width, height = 120, 40 // Default size
		}
	}

	screenshot := &ScreenshotData{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		Description: description,
		TestName:    testName,
		ScreenSize: map[string]int{
			"width":  width,
			"height": height,
		},
		State:       state,
		CurrentApp:  currentApp,
		UserActions: actions,
		ScreenText:  screenText,
		Metadata:    metadata,
	}

	// Capture any error from the model
	if errorModel, ok := model.(interface{ GetError() error }); ok {
		if err := errorModel.GetError(); err != nil {
			screenshot.Error = err.Error()
		}
	}

	s.currentCapture = screenshot

	// Save the screenshot
	return s.saveScreenshot(screenshot)
}

// CaptureFormState captures a screenshot with form-specific context
func (s *ScreenshotComponent) CaptureFormState(model tea.Model, formComponent interface{}, description, testName string, actions []string) error {
	context := make(map[string]interface{})

	// Extract form-specific information
	if form, ok := formComponent.(interface{ GetCurrentField() string }); ok {
		context["currentField"] = form.GetCurrentField()
	}

	if form, ok := formComponent.(interface{ GetFieldCount() int }); ok {
		context["fieldCount"] = form.GetFieldCount()
	}

	if form, ok := formComponent.(interface{ IsValid() bool }); ok {
		context["formValid"] = form.IsValid()
	}

	context["componentType"] = "form"

	return s.CaptureWithContext(model, description, testName, actions, context)
}

// CaptureHelpState captures a screenshot with help-specific context
func (s *ScreenshotComponent) CaptureHelpState(model tea.Model, helpComponent interface{}, description, testName string, actions []string) error {
	context := make(map[string]interface{})

	// Extract help-specific information
	if help, ok := helpComponent.(interface{ GetContext() string }); ok {
		context["helpContext"] = help.GetContext()
	}

	if help, ok := helpComponent.(interface{ IsVisible() bool }); ok {
		context["helpVisible"] = help.IsVisible()
	}

	if help, ok := helpComponent.(interface{ GetHelpItemCount() int }); ok {
		context["helpItemCount"] = help.GetHelpItemCount()
	}

	context["componentType"] = "help"

	return s.CaptureWithContext(model, description, testName, actions, context)
}

// CaptureAppListState captures a screenshot with application list context
func (s *ScreenshotComponent) CaptureAppListState(model tea.Model, appListComponent interface{}, description, testName string, actions []string) error {
	context := make(map[string]interface{})

	// Extract application list information
	if appList, ok := appListComponent.(interface{ GetItemCount() int }); ok {
		context["applicationCount"] = appList.GetItemCount()
	}

	if appList, ok := appListComponent.(interface{ GetSelectedIndex() int }); ok {
		context["selectedIndex"] = appList.GetSelectedIndex()
	}

	if appList, ok := appListComponent.(interface{ GetSelectedApp() string }); ok {
		context["selectedApp"] = appList.GetSelectedApp()
	}

	if appList, ok := appListComponent.(interface{ GetFilterText() string }); ok {
		context["filterText"] = appList.GetFilterText()
	}

	context["componentType"] = "app_list"

	return s.CaptureWithContext(model, description, testName, actions, context)
}

// IntegrateWithComponents provides a helper for integrating with existing ZeroUI components
func (s *ScreenshotComponent) IntegrateWithComponents() *ComponentIntegrator {
	return &ComponentIntegrator{
		screenshot: s,
	}
}

// ComponentIntegrator provides integration helpers for existing components
type ComponentIntegrator struct {
	screenshot *ScreenshotComponent
}

// WithApplicationList integrates with an application list component
func (ci *ComponentIntegrator) WithApplicationList(appList interface{}) *ApplicationListIntegrator {
	return &ApplicationListIntegrator{
		integrator: ci,
		appList:    appList,
	}
}

// WithForm integrates with a form component
func (ci *ComponentIntegrator) WithForm(form interface{}) *FormIntegrator {
	return &FormIntegrator{
		integrator: ci,
		form:       form,
	}
}

// WithHelp integrates with a help component
func (ci *ComponentIntegrator) WithHelp(help interface{}) *HelpIntegrator {
	return &HelpIntegrator{
		integrator: ci,
		help:       help,
	}
}

// ApplicationListIntegrator provides application list specific integration
type ApplicationListIntegrator struct {
	integrator *ComponentIntegrator
	appList    interface{}
}

// Capture captures a screenshot with application list context
func (ali *ApplicationListIntegrator) Capture(model tea.Model, description, testName string, actions ...string) error {
	return ali.integrator.screenshot.CaptureAppListState(model, ali.appList, description, testName, actions)
}

// FormIntegrator provides form specific integration
type FormIntegrator struct {
	integrator *ComponentIntegrator
	form       interface{}
}

// Capture captures a screenshot with form context
func (fi *FormIntegrator) Capture(model tea.Model, description, testName string, actions ...string) error {
	return fi.integrator.screenshot.CaptureFormState(model, fi.form, description, testName, actions)
}

// HelpIntegrator provides help specific integration
type HelpIntegrator struct {
	integrator *ComponentIntegrator
	help       interface{}
}

// Capture captures a screenshot with help context
func (hi *HelpIntegrator) Capture(model tea.Model, description, testName string, actions ...string) error {
	return hi.integrator.screenshot.CaptureHelpState(model, hi.help, description, testName, actions)
}

// Update implements tea.Model
func (s *ScreenshotComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View implements tea.Model
func (s *ScreenshotComponent) View() string {
	if s.isCapturing {
		return s.styles.Info.Render("📸 Capturing screenshot...")
	}
	return ""
}

// KeyBindings returns key bindings for the screenshot component
func (s *ScreenshotComponent) KeyBindings() []key.Binding {
	return []key.Binding{}
}

// HandleKey handles key messages
func (s *ScreenshotComponent) HandleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return s, nil
}
