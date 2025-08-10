package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
)

// DemoModel showcases the Huh integration
type DemoModel struct {
	currentView string
	appSelector *components.HuhAppSelectorModel
	configEditor *components.HuhConfigEditorModel
	bubblesDemo *components.EnhancedBubblesModel
	width, height int
}

func NewDemoModel() *DemoModel {
	return &DemoModel{
		currentView:  "selector",
		appSelector:  components.NewHuhAppSelector(),
		configEditor: components.NewHuhConfigEditor("Demo App"),
		bubblesDemo:  components.NewEnhancedBubblesModel(),
	}
}

func (m *DemoModel) Init() tea.Cmd {
	return tea.Batch(
		m.appSelector.Init(),
		m.configEditor.Init(),
		m.bubblesDemo.Init(),
	)
}

func (m *DemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.appSelector.SetSize(m.width, m.height)
		m.configEditor.SetSize(m.width, m.height)
		m.bubblesDemo.SetSize(m.width, m.height)
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.currentView = "selector"
		case "2":
			m.currentView = "config"
		case "3":
			m.currentView = "bubbles"
		}
	
	case components.AppSelectedMsg:
		m.currentView = "config"
		// Set up demo fields for the config editor
		demoFields := []*components.FieldModel{
			components.NewField("theme", "select", "dark", []string{"light", "dark", "auto"}, "Application theme"),
			components.NewField("font_size", "int", "14", []string{"12", "14", "16", "18", "20"}, "Font size in pixels"),
			components.NewField("word_wrap", "boolean", "true", []string{"true", "false"}, "Enable word wrapping"),
			components.NewField("languages", "multiselect", "go,javascript", []string{"go", "javascript", "python", "rust", "typescript"}, "Supported languages"),
		}
		m.configEditor.SetFields(demoFields)
	}
	
	// Update the currently active component
	switch m.currentView {
	case "selector":
		updatedSelector, cmd := m.appSelector.Update(msg)
		m.appSelector = updatedSelector
		return m, cmd
	case "config":
		updatedEditor, cmd := m.configEditor.Update(msg)
		m.configEditor = updatedEditor.(*components.HuhConfigEditorModel)
		return m, cmd
	case "bubbles":
		updatedBubbles, cmd := m.bubblesDemo.Update(msg)
		m.bubblesDemo = updatedBubbles.(*components.EnhancedBubblesModel)
		return m, cmd
	}
	
	return m, nil
}

func (m *DemoModel) View() string {
	if m.width == 0 {
		return "Loading demo..."
	}
	
	// Show instructions at top
	instructions := fmt.Sprintf(`
🎨 ZeroUI Huh Integration Demo

Current View: %s

Controls:
  1 - App Selector (Huh Select)     2 - Config Editor (Huh Forms)     3 - Bubbles Showcase
  q - Quit                         ↑↓ Navigate                       ⏎ Select

`, m.currentView)
	
	var content string
	switch m.currentView {
	case "selector":
		content = m.appSelector.View()
	case "config":
		content = m.configEditor.View()
	case "bubbles":
		content = m.bubblesDemo.View()
	default:
		content = "Unknown view"
	}
	
	return instructions + "\n" + content
}

func main() {
	// Only run demo if called directly
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		fmt.Println("🎨 Starting ZeroUI Huh Integration Demo...")
		fmt.Println("This demo showcases the modern Huh-based UI components.")
		fmt.Println("Note: Demo requires a proper TTY environment to run interactively.")
		
		model := NewDemoModel()
		program := tea.NewProgram(model, tea.WithAltScreen())
		
		if _, err := program.Run(); err != nil {
			log.Fatalf("Demo error: %v", err)
		}
	} else {
		fmt.Println(`
🎨 ZeroUI Huh Integration Demo

This demo showcases the implemented Huh integration:

1. Modern App Selector using Huh Select components
2. Dynamic Configuration Editor with Huh Forms  
3. Complete Bubbles integration showcase

Features demonstrated:
✅ Huh Select/MultiSelect/Input/Confirm components
✅ Elegant Lipgloss styling and centering  
✅ Responsive 4-column layouts
✅ All major Bubbles components integration
✅ Modern form validation and theming

To run the interactive demo:
  go run demo_huh_integration.go demo

Components implemented:
- HuhAppSelectorModel: Modern app selection with status indicators
- HuhConfigEditorModel: Dynamic forms based on field types  
- EnhancedBubblesModel: Complete Bubbles showcase with elegant styling

The main ZeroUI application now defaults to these modern interfaces:
  go run main.go ui              # Modern Huh-based interface
  go run main.go ui --app ghostty # Start with specific app
  
  In-app controls:
  - Ctrl+H: Switch to modern Huh interface
  - Ctrl+L: Switch to legacy interface
  - ↑↓: Navigate, ⏎: Select, ?: Help, q: Quit
		`)
	}
}