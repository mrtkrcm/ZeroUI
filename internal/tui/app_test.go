package tui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// setupTestTUI creates a test TUI environment
func setupTestTUI(t *testing.T) (*toggle.Engine, string, func()) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-tui-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create apps directory
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app config
	testConfig := `name: tui-test-app
path: ` + filepath.Join(tmpDir, "target", "config.json") + `
format: json
description: Test application for TUI

fields:
  theme:
    type: choice
    values: ["dark", "light", "auto"]
    default: "dark"
    description: "Application theme"
  
  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 14
    description: "Font size"
    
  debug:
    type: boolean
    default: false
    description: "Enable debug mode"

  editor:
    type: string
    default: "vim"
    description: "Default editor"

presets:
  default:
    name: default
    description: Default settings
    values:
      theme: dark
      font-size: 14
      debug: false
      
  light-mode:
    name: light-mode
    description: Light theme setup
    values:
      theme: light
      font-size: 16
      debug: false

  dev-mode:
    name: dev-mode
    description: Development setup
    values:
      theme: dark
      font-size: 12
      debug: true

hooks:
  post-toggle: "echo 'Config updated'"
`

	configPath := filepath.Join(appsDir, "tui-test-app.yaml")
	if err := ioutil.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create target directory and config
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	targetConfigPath := filepath.Join(targetDir, "config.json")
	targetConfig := `{
  "theme": "dark",
  "font-size": 14,
  "debug": false,
  "editor": "vim"
}`

	if err := ioutil.WriteFile(targetConfigPath, []byte(targetConfig), 0644); err != nil {
		t.Fatalf("Failed to write target config: %v", err)
	}

	// Set up engine with custom config dir
	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// We need to set the config directory - this requires adding a method to the engine
	// For now, we'll work with the default directory structure

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return engine, tmpDir, cleanup
}

// TestNewApp tests creating a new TUI application
func TestNewApp(t *testing.T) {
	_, _, cleanup := setupTestTUI(t)
	defer cleanup()

	app, err := NewApp("tui-test-app")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	if app.engine == nil {
		t.Error("Expected engine to be initialized")
	}

	if app.initialApp != "tui-test-app" {
		t.Errorf("Expected initialApp 'tui-test-app', got '%s'", app.initialApp)
	}
}

// TestNewApp_EmptyApp tests creating app without initial app
func TestNewApp_EmptyApp(t *testing.T) {
	app, err := NewApp("")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	if app.initialApp != "" {
		t.Errorf("Expected empty initialApp, got '%s'", app.initialApp)
	}
}

// TestNewModel tests creating a new model
func TestNewModel(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	// Test with no initial app
	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.state != AppSelectionView {
		t.Errorf("Expected initial state AppSelectionView, got %v", model.state)
	}

	if model.engine != engine {
		t.Error("Expected engine to be set")
	}

	if len(model.apps) == 0 {
		t.Error("Expected some apps to be loaded")
	}

	// Test with initial app
	model2, err := NewModel(engine, "tui-test-app")
	if err != nil {
		t.Fatalf("Failed to create model with initial app: %v", err)
	}

	if model2.currentApp != "tui-test-app" {
		t.Errorf("Expected currentApp 'tui-test-app', got '%s'", model2.currentApp)
	}

	if model2.state != ConfigEditView {
		t.Errorf("Expected state ConfigEditView with initial app, got %v", model2.state)
	}
}

// TestModel_Init tests model initialization
func TestModel_Init(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	cmd := model.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

// TestModel_Update_WindowSize tests window size updates
func TestModel_Update_WindowSize(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedModel, cmd := model.Update(msg)
	if cmd != nil {
		t.Error("Expected no command from window size update")
	}

	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m.width != 80 {
		t.Errorf("Expected width 80, got %d", m.width)
	}

	if m.height != 24 {
		t.Errorf("Expected height 24, got %d", m.height)
	}
}

// TestModel_HandleKeyPress_Global tests global key handling
func TestModel_HandleKeyPress_Global(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test help key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := model.Update(msg)
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m.state != HelpView {
		t.Errorf("Expected HelpView after '?', got %v", m.state)
	}

	// Test escape from help
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel2, _ := m.Update(escMsg)
	m2, ok := updatedModel2.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m2.state != AppSelectionView {
		t.Errorf("Expected AppSelectionView after escape from help, got %v", m2.state)
	}
}

// TestModel_AppSelectionKeys tests app selection key handling
func TestModel_AppSelectionKeys(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Ensure we have apps
	if len(model.apps) == 0 {
		t.Skip("No apps available for testing")
	}

	initialCursor := model.cursor

	// Test down key
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(downMsg)
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	expectedCursor := initialCursor
	if expectedCursor < len(model.apps)-1 {
		expectedCursor++
	}
	if m.cursor != expectedCursor {
		t.Errorf("Expected cursor %d after down, got %d", expectedCursor, m.cursor)
	}

	// Test up key
	upMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updatedModel2, _ := m.Update(upMsg)
	m2, ok := updatedModel2.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m2.cursor != initialCursor {
		t.Errorf("Expected cursor back to %d after up, got %d", initialCursor, m2.cursor)
	}

	// Test enter key
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel3, _ := m2.Update(enterMsg)
	m3, ok := updatedModel3.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m3.state != ConfigEditView {
		t.Errorf("Expected ConfigEditView after enter, got %v", m3.state)
	}

	if m3.currentApp == "" {
		t.Error("Expected currentApp to be set after selection")
	}
}

// TestModel_ConfigEditKeys tests config edit key handling
func TestModel_ConfigEditKeys(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "tui-test-app")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Ensure we're in config edit mode
	if model.state != ConfigEditView {
		t.Fatalf("Expected ConfigEditView, got %v", model.state)
	}

	appConfig := model.appConfigs[model.currentApp]
	if appConfig == nil {
		t.Fatal("Expected app config to be loaded")
	}

	if len(appConfig.Fields) == 0 {
		t.Fatal("Expected some fields to be loaded")
	}

	initialCursor := appConfig.cursor

	// Test down navigation
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(downMsg)
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	expectedCursor := initialCursor
	if expectedCursor < len(appConfig.Fields)-1 {
		expectedCursor++
	}
	if m.appConfigs[m.currentApp].cursor != expectedCursor {
		t.Errorf("Expected cursor %d after down, got %d", expectedCursor, m.appConfigs[m.currentApp].cursor)
	}

	// Test preset key
	presetMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updatedModel2, _ := m.Update(presetMsg)
	m2, ok := updatedModel2.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	if m2.state != PresetSelectionView {
		t.Errorf("Expected PresetSelectionView after 'p', got %v", m2.state)
	}
}

// TestModel_PresetSelectionKeys tests preset selection key handling
func TestModel_PresetSelectionKeys(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "tui-test-app")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Switch to preset selection
	model.state = PresetSelectionView

	appConfig := model.appConfigs[model.currentApp]
	if appConfig == nil {
		t.Fatal("Expected app config to be loaded")
	}

	if len(appConfig.Presets) == 0 {
		t.Fatal("Expected some presets to be loaded")
	}

	initialCursor := appConfig.cursor

	// Test down navigation
	downMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updatedModel, _ := model.Update(downMsg)
	m, ok := updatedModel.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	expectedCursor := initialCursor
	if expectedCursor < len(appConfig.Presets)-1 {
		expectedCursor++
	}
	if m.appConfigs[m.currentApp].cursor != expectedCursor {
		t.Errorf("Expected cursor %d after down, got %d", expectedCursor, m.appConfigs[m.currentApp].cursor)
	}

	// Test enter to apply preset
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel2, _ := m.Update(enterMsg)
	m2, ok := updatedModel2.(*Model)
	if !ok {
		t.Fatal("Expected model to be *Model")
	}

	// Should go back to config edit after applying preset
	if m2.state != ConfigEditView {
		t.Errorf("Expected ConfigEditView after applying preset, got %v", m2.state)
	}
}

// TestModel_View tests view rendering
func TestModel_View(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test app selection view
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !containsString(view, "ZeroUI") {
		t.Error("Expected view to contain 'ZeroUI'")
	}

	// Test config edit view
	model.state = ConfigEditView
	model.currentApp = "tui-test-app"
	if err := model.loadAppConfig("tui-test-app"); err != nil {
		t.Fatalf("Failed to load app config: %v", err)
	}

	configView := model.View()
	if configView == "" {
		t.Error("Expected non-empty config view")
	}

	// Test preset selection view
	model.state = PresetSelectionView
	presetView := model.View()
	if presetView == "" {
		t.Error("Expected non-empty preset view")
	}

	// Test help view
	model.state = HelpView
	helpView := model.View()
	if helpView == "" {
		t.Error("Expected non-empty help view")
	}

	if !containsString(helpView, "Key Bindings") {
		t.Error("Expected help view to contain 'Key Bindings'")
	}
}

// TestModel_ErrorHandling tests error display
func TestModel_ErrorHandling(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Set an error
	testError := fmt.Errorf("test error message")
	model.err = testError

	errorView := model.View()
	if errorView == "" {
		t.Error("Expected non-empty error view")
	}

	if !containsString(errorView, "test error message") {
		t.Error("Expected error view to contain error message")
	}

	if !containsString(errorView, "Error:") {
		t.Error("Expected error view to contain 'Error:'")
	}
}

// TestModel_LoadAppConfig tests loading app configuration
func TestModel_LoadAppConfig(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	err = model.loadAppConfig("tui-test-app")
	if err != nil {
		t.Fatalf("Failed to load app config: %v", err)
	}

	appConfig := model.appConfigs["tui-test-app"]
	if appConfig == nil {
		t.Error("Expected app config to be loaded")
	}

	if appConfig.Name != "tui-test-app" {
		t.Errorf("Expected app name 'tui-test-app', got '%s'", appConfig.Name)
	}

	if len(appConfig.Fields) == 0 {
		t.Error("Expected some fields to be loaded")
	}

	if len(appConfig.Presets) == 0 {
		t.Error("Expected some presets to be loaded")
	}

	// Check that field values are properly loaded
	themeField := findField(appConfig.Fields, "theme")
	if themeField == nil {
		t.Error("Expected to find theme field")
	} else {
		if themeField.CurrentValue == "" {
			t.Error("Expected theme field to have a current value")
		}

		if len(themeField.Values) == 0 {
			t.Error("Expected theme field to have possible values")
		}
	}
}

// TestModel_ApplyFieldChange tests applying field changes
func TestModel_ApplyFieldChange(t *testing.T) {
	engine, _, cleanup := setupTestTUI(t)
	defer cleanup()

	model, err := NewModel(engine, "tui-test-app")
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	appConfig := model.appConfigs[model.currentApp]
	if appConfig == nil {
		t.Fatal("Expected app config to be loaded")
	}

	// Find a field with values
	var testField *FieldView
	for i := range appConfig.Fields {
		if len(appConfig.Fields[i].Values) > 1 {
			testField = &appConfig.Fields[i]
			break
		}
	}

	if testField == nil {
		t.Fatal("Expected to find a field with multiple values")
	}

	originalValue := testField.CurrentValue
	testField.cursor = 1 // Change to second value

	err = model.applyFieldChange(testField)
	if err != nil {
		t.Fatalf("Failed to apply field change: %v", err)
	}

	if testField.CurrentValue == originalValue {
		t.Error("Expected current value to change after applying field change")
	}

	if testField.CurrentValue != testField.Values[1] {
		t.Errorf("Expected current value to be '%s', got '%s'", testField.Values[1], testField.CurrentValue)
	}
}

// TestViewState tests ViewState constants
func TestViewState(t *testing.T) {
	if AppSelectionView != 0 {
		t.Errorf("Expected AppSelectionView to be 0, got %d", AppSelectionView)
	}

	if ConfigEditView != 1 {
		t.Errorf("Expected ConfigEditView to be 1, got %d", ConfigEditView)
	}

	if PresetSelectionView != 2 {
		t.Errorf("Expected PresetSelectionView to be 2, got %d", PresetSelectionView)
	}

	if HelpView != 3 {
		t.Errorf("Expected HelpView to be 3, got %d", HelpView)
	}
}

// TestAppConfigView tests AppConfigView structure
func TestAppConfigView(t *testing.T) {
	appConfig := &AppConfigView{
		Name:    "test",
		Fields:  []FieldView{},
		Presets: []PresetView{},
	}

	if appConfig.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", appConfig.Name)
	}

	if len(appConfig.Fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(appConfig.Fields))
	}

	if len(appConfig.Presets) != 0 {
		t.Errorf("Expected 0 presets, got %d", len(appConfig.Presets))
	}
}

// TestFieldView tests FieldView structure
func TestFieldView(t *testing.T) {
	field := FieldView{
		Key:          "test-key",
		Type:         "choice",
		CurrentValue: "value1",
		Values:       []string{"value1", "value2"},
		Description:  "Test field",
	}

	if field.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", field.Key)
	}

	if field.Type != "choice" {
		t.Errorf("Expected type 'choice', got '%s'", field.Type)
	}

	if len(field.Values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(field.Values))
	}
}

// TestPresetView tests PresetView structure
func TestPresetView(t *testing.T) {
	preset := PresetView{
		Name:        "test-preset",
		Description: "Test preset",
		Values:      map[string]interface{}{"key": "value"},
	}

	if preset.Name != "test-preset" {
		t.Errorf("Expected name 'test-preset', got '%s'", preset.Name)
	}

	if len(preset.Values) != 1 {
		t.Errorf("Expected 1 value, got %d", len(preset.Values))
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
			len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func findField(fields []FieldView, key string) *FieldView {
	for i := range fields {
		if fields[i].Key == key {
			return &fields[i]
		}
	}
	return nil
}