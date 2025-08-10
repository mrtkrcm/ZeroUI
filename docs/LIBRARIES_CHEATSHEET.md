# Key Libraries Cheatsheet

This cheatsheet provides quick reference for the key libraries used in ZeroUI.

## Bubble Tea (TUI Framework)
**Repository:** https://github.com/charmbracelet/bubbletea  
**Purpose:** Terminal User Interface framework based on The Elm Architecture

### Core Concepts
```go
type Model struct {
    // Your application state
}

func (m Model) Init() tea.Cmd {
    // Initialize the program
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages and update state
    return m, nil
}

func (m Model) View() string {
    // Render the interface
    return "Hello, World!"
}

// Run the program
p := tea.NewProgram(initialModel)
p.Run()
```

### Common Message Types
```go
type tea.KeyMsg           // Keyboard input
type tea.WindowSizeMsg    // Terminal resize
type tea.MouseMsg         // Mouse events
type tea.BatchMsg         // Batch multiple messages
```

### Commands
```go
tea.Quit                  // Exit program
tea.Batch(cmd1, cmd2)     // Execute multiple commands
tea.Sequence(cmd1, cmd2)  // Execute commands in sequence
tea.Tick(time.Second)     // Timer command
```

## Huh (Forms & Prompts)
**Repository:** https://github.com/charmbracelet/huh  
**Purpose:** Interactive forms and prompts for terminal applications

### Basic Form
```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("What's your name?").
            Value(&name),
        
        huh.NewSelect[string]().
            Title("Choose a color").
            Options(
                huh.NewOption("Red", "red"),
                huh.NewOption("Green", "green"),
            ).
            Value(&color),
    ),
)

err := form.Run()
```

### Field Types
```go
huh.NewInput()           // Text input
huh.NewPassword()        // Password input
huh.NewConfirm()         // Yes/No confirmation
huh.NewSelect()          // Single selection
huh.NewMultiSelect()     // Multiple selection
huh.NewFilePicker()      // File selection
huh.NewNote()            // Information display
```

## Lipgloss (Styling)
**Repository:** https://github.com/charmbracelet/lipgloss  
**Purpose:** CSS-like styling for terminal applications

### Basic Styling
```go
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1).
    Margin(1, 0)

styled := style.Render("Hello World")
```

### Colors
```go
lipgloss.Color("#FF0000")    // Hex colors
lipgloss.Color("205")        // 256-color mode
lipgloss.AdaptiveColor{      // Adaptive colors
    Light: "#000000",
    Dark:  "#FFFFFF",
}
```

### Layout
```go
lipgloss.JoinVertical(lipgloss.Left, str1, str2)
lipgloss.JoinHorizontal(lipgloss.Top, str1, str2)
lipgloss.PlaceVertical(height, lipgloss.Center, str)
lipgloss.PlaceHorizontal(width, lipgloss.Center, str)
```

### Common Styles
```go
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Background(lipgloss.Color("62")).
    Padding(0, 1)

errorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("9"))

helpStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("241"))
```

## Glamour (Markdown Rendering)
**Repository:** https://github.com/charmbracelet/glamour  
**Purpose:** Render Markdown in terminal with syntax highlighting

### Basic Usage
```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(80),
)

out, err := renderer.Render(markdown)
fmt.Print(out)
```

### Styles
```go
glamour.WithAutoStyle()           // Automatic dark/light detection
glamour.WithStylePath("style.json") // Custom style file
glamour.WithStyles(glamour.DarkStyleConfig) // Predefined styles
```

### Custom Styles
```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithStyles(glamour.StyleConfig{
        Document: glamour.StyleBlock{
            StylePrimitive: glamour.StylePrimitive{
                BlockPrefix: "\n",
                BlockSuffix: "\n",
            },
            Margin: uintPtr(2),
        },
    }),
)
```

## Bubbles (Common Components)
**Repository:** https://github.com/charmbracelet/bubbles  
**Purpose:** Collection of common Bubble Tea components

### List Component
```go
import "github.com/charmbracelet/bubbles/list"

type item string
func (i item) FilterValue() string { return string(i) }

items := []list.Item{
    item("Item 1"),
    item("Item 2"),
}

l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
l.Title = "My List"
```

### Text Input
```go
import "github.com/charmbracelet/bubbles/textinput"

ti := textinput.New()
ti.Placeholder = "Enter text..."
ti.Focus()

// In Update:
ti, cmd = ti.Update(msg)

// In View:
return ti.View()
```

### Viewport (Scrollable Content)
```go
import "github.com/charmbracelet/bubbles/viewport"

vp := viewport.New(width, height)
vp.SetContent(longString)

// In Update:
vp, cmd = vp.Update(msg)

// In View:
return vp.View()
```

### Progress Bar
```go
import "github.com/charmbracelet/bubbles/progress"

prog := progress.New(progress.WithDefaultGradient())
prog.Width = 40

// In View:
return prog.ViewAs(0.6) // 60% progress
```

### Spinner
```go
import "github.com/charmbracelet/bubbles/spinner"

s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

// In Update:
s, cmd = s.Update(msg)

// In View:
return s.View()
```

## Log (Structured Logging)
**Repository:** https://github.com/charmbracelet/log  
**Purpose:** Beautiful, human-readable logging

### Basic Usage
```go
import "github.com/charmbracelet/log"

log.Info("Starting application", "version", "1.0.0")
log.Warn("Configuration missing", "file", "config.yaml")
log.Error("Database connection failed", "err", err)
```

### Logger Configuration
```go
logger := log.NewWithOptions(os.Stderr, log.Options{
    ReportCaller:    true,
    ReportTimestamp: true,
    TimeFormat:      time.Kitchen,
    Prefix:          "ZeroUI",
})

logger.SetLevel(log.DebugLevel)
```

### Structured Fields
```go
logger.With("user", "john", "session", "abc123").Info("User logged in")

// Sub-loggers
userLogger := logger.With("user", "john")
userLogger.Info("Action performed")
userLogger.Error("Action failed")
```

### Custom Styles
```go
styles := log.DefaultStyles()
styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
    SetString("ERROR").
    Padding(0, 1, 0, 1).
    Background(lipgloss.Color("204")).
    Foreground(lipgloss.Color("0"))

logger.SetStyles(styles)
```

## ZeroUI Integration Patterns

### TUI Pattern
```go
type Model struct {
    list     list.Model
    textInput textinput.Model
    spinner   spinner.Model
    viewport  viewport.Model
    
    state    ViewState
    err      error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.state {
        case ListView:
            m.list, cmd = m.list.Update(msg)
        case InputView:
            m.textInput, cmd = m.textInput.Update(msg)
        }
    }
    return m, cmd
}
```

### Form Integration
```go
func promptForConfig() error {
    var theme string
    var fontSize int
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Choose theme").
                Options(
                    huh.NewOption("Dark", "dark"),
                    huh.NewOption("Light", "light"),
                ).
                Value(&theme),
                
            huh.NewInput().
                Title("Font size").
                Value(&fontSize).
                Validate(func(s string) error {
                    if i, err := strconv.Atoi(s); err != nil || i < 8 || i > 72 {
                        return fmt.Errorf("invalid font size")
                    }
                    return nil
                }),
        ),
    )
    
    return form.Run()
}
```

### Error Display Pattern
```go
errorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("9")).
    Bold(true)

warningStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("11"))

func renderError(err error) string {
    if ctErr, ok := err.(*errors.ZeroUIError); ok {
        return errorStyle.Render("Error: ") + ctErr.Message + "\n" +
               warningStyle.Render("Suggestion: ") + strings.Join(ctErr.Suggestions, ", ")
    }
    return errorStyle.Render("Error: ") + err.Error()
}
```

### Logging Integration
```go
func setupLogger() *log.Logger {
    logger := log.NewWithOptions(os.Stderr, log.Options{
        ReportCaller:    false,
        ReportTimestamp: true,
        TimeFormat:      "15:04:05",
        Prefix:          "zeroui",
    })
    
    // Custom styles for ZeroUI
    styles := log.DefaultStyles()
    styles.Key = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
    styles.Value = lipgloss.NewStyle().Foreground(lipgloss.Color("37"))
    
    logger.SetStyles(styles)
    return logger
}
```

## Quick Reference Commands

### Development
```bash
# Run with debug logging
go run . --log-level debug

# Test TUI components
go run . ui test-app

# Generate documentation
glamour README.md

# Format and style check
gofmt -s -w .
```

### Common Patterns
- Use `tea.Batch()` for multiple commands
- Always handle `tea.KeyMsg` for navigation
- Use `lipgloss.JoinVertical()` for layout
- Implement `list.Item` interface for custom list items
- Use `viewport` for scrollable content
- Add `spinner` for loading states
- Use structured logging with contextual fields