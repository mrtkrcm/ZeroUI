# Glamour Cheatsheet

**Repository:** https://github.com/charmbracelet/glamour  
**Purpose:** Render Markdown in terminal with syntax highlighting and beautiful formatting

## Basic Usage

### Simple Rendering
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/glamour"
)

func main() {
    markdown := `
# Hello World

This is **bold** and *italic* text.

- Item 1
- Item 2
- Item 3

` + "```go\nfmt.Println(\"Hello, World!\")\n```" + `
`

    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(), // Automatic light/dark detection
        glamour.WithWordWrap(80),
    )
    if err != nil {
        panic(err)
    }

    out, err := renderer.Render(markdown)
    if err != nil {
        panic(err)
    }

    fmt.Print(out)
}
```

### Quick Render
```go
import "github.com/charmbracelet/glamour"

// Quick render with default settings
output, err := glamour.Render(markdown, "dark")
if err != nil {
    return err
}
fmt.Print(output)
```

## Renderer Configuration

### Basic Options
```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),           // Auto light/dark detection
    glamour.WithWordWrap(80),          // Wrap at 80 columns
    glamour.WithPreserveNewLines(),    // Keep original line breaks
    glamour.WithEmoji(),               // Enable emoji rendering
)
```

### Style Options
```go
// Use predefined styles
renderer, err := glamour.NewTermRenderer(
    glamour.WithStyles(glamour.DarkStyleConfig),  // Dark theme
)

renderer, err := glamour.NewTermRenderer(
    glamour.WithStyles(glamour.LightStyleConfig), // Light theme
)

// Load from file
renderer, err := glamour.NewTermRenderer(
    glamour.WithStylePath("custom-style.json"),
)

// Auto style (detects terminal background)
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
)
```

### Environment Detection
```go
// Check terminal capabilities
if glamour.HasDarkBackground() {
    renderer, _ = glamour.NewTermRenderer(
        glamour.WithStyles(glamour.DarkStyleConfig),
    )
} else {
    renderer, _ = glamour.NewTermRenderer(
        glamour.WithStyles(glamour.LightStyleConfig),
    )
}
```

## Built-in Styles

### Available Styles
```go
// Pre-defined style configs
glamour.DarkStyleConfig     // Dark theme
glamour.LightStyleConfig    // Light theme  
glamour.NoTTYStyleConfig    // Plain text output

// Style names for glamour.Render()
"auto"        // Auto-detect based on terminal
"dark"        // Dark theme
"light"       // Light theme
"notty"       // Plain text
"dracula"     // Dracula theme
"ascii"       // ASCII-only characters
```

### Using Named Styles
```go
// Render with named style
darkOutput, err := glamour.Render(markdown, "dark")
lightOutput, err := glamour.Render(markdown, "light")
draculaOutput, err := glamour.Render(markdown, "dracula")
asciiOutput, err := glamour.Render(markdown, "ascii")
```

## Custom Styling

### Style Configuration Structure
```go
customStyle := glamour.StyleConfig{
    Document: glamour.StyleBlock{
        StylePrimitive: glamour.StylePrimitive{
            BlockPrefix: "\n",
            BlockSuffix: "\n",
        },
        Margin: uintPtr(2),
    },
    
    Heading: glamour.StyleBlock{
        StylePrimitive: glamour.StylePrimitive{
            Color:      stringPtr("#7D56F4"),
            Bold:       boolPtr(true),
            BlockPrefix: "\n",
            BlockSuffix: "\n",
        },
        Margin: uintPtr(1),
    },
    
    Paragraph: glamour.StyleBlock{
        StylePrimitive: glamour.StylePrimitive{
            BlockSuffix: "\n",
        },
        Margin: uintPtr(1),
    },
    
    CodeBlock: glamour.StyleCodeBlock{
        StyleBlock: glamour.StyleBlock{
            StylePrimitive: glamour.StylePrimitive{
                Color:       stringPtr("#FAFAFA"),
                Background:  stringPtr("#282828"),
                Padding:     uintPtr(1),
                BlockPrefix: "\n",
                BlockSuffix: "\n",
            },
            Margin: uintPtr(1),
        },
        Chroma: &glamour.Chroma{
            Text:       glamour.StylePrimitive{Color: stringPtr("#FAFAFA")},
            Error:      glamour.StylePrimitive{Color: stringPtr("#F1FA8C")},
            Comment:    glamour.StylePrimitive{Color: stringPtr("#6272A4")},
            Keyword:    glamour.StylePrimitive{Color: stringPtr("#8BE9FD")},
            Name:       glamour.StylePrimitive{Color: stringPtr("#50FA7B")},
            Literal:    glamour.StylePrimitive{Color: stringPtr("#FFB86C")},
            String:     glamour.StylePrimitive{Color: stringPtr("#F1FA8C")},
        },
    },
}

renderer, err := glamour.NewTermRenderer(
    glamour.WithStyles(customStyle),
)
```

### Helper Functions for Custom Styles
```go
// Helper functions for pointer types
func stringPtr(s string) *string {
    return &s
}

func boolPtr(b bool) *bool {
    return &b
}

func uintPtr(u uint) *uint {
    return &u
}

func intPtr(i int) *int {
    return &i
}
```

## Advanced Features

### Code Syntax Highlighting
```go
// Customize code block highlighting
codeStyle := glamour.StyleCodeBlock{
    StyleBlock: glamour.StyleBlock{
        StylePrimitive: glamour.StylePrimitive{
            Background:  stringPtr("#1E1E1E"),
            Color:       stringPtr("#FAFAFA"),
            Padding:     uintPtr(1),
            BlockPrefix: "\n",
            BlockSuffix: "\n",
        },
    },
    // Chroma syntax highlighting
    Chroma: &glamour.Chroma{
        Text:       glamour.StylePrimitive{Color: stringPtr("#FAFAFA")},
        Error:      glamour.StylePrimitive{Color: stringPtr("#FF5555")},
        Comment:    glamour.StylePrimitive{Color: stringPtr("#6272A4")},
        Keyword:    glamour.StylePrimitive{Color: stringPtr("#FF79C6")},
        Name:       glamour.StylePrimitive{Color: stringPtr("#BD93F9")},
        NameClass:  glamour.StylePrimitive{Color: stringPtr("#8BE9FD")},
        String:     glamour.StylePrimitive{Color: stringPtr("#F1FA8C")},
        Number:     glamour.StylePrimitive{Color: stringPtr("#BD93F9")},
    },
}
```

### Word Wrapping
```go
// Different wrapping strategies
renderer, err := glamour.NewTermRenderer(
    glamour.WithWordWrap(80),          // Hard wrap at 80 columns
    glamour.WithPreserveNewLines(),    // Preserve original line breaks
)

// Dynamic width based on terminal
import "golang.org/x/term"

width, _, err := term.GetSize(int(os.Stdout.Fd()))
if err != nil {
    width = 80 // Default fallback
}

renderer, err := glamour.NewTermRenderer(
    glamour.WithWordWrap(width-4), // Leave margin
)
```

### Table Rendering
```go
// Custom table styling
tableStyle := glamour.StyleTable{
    StyleBlock: glamour.StyleBlock{
        StylePrimitive: glamour.StylePrimitive{
            BlockSuffix: "\n",
        },
    },
    CenterSeparator: stringPtr("┼"),
    ColumnSeparator: stringPtr("│"),
    RowSeparator:    stringPtr("─"),
}
```

## Markdown Elements Support

### Supported Elements
```markdown
# Headers (H1-H6)
**Bold** and *italic* text
~~Strikethrough~~
`inline code`

```go
// Code blocks with syntax highlighting
func main() {
    fmt.Println("Hello, World!")
}
```

> Blockquotes
> With multiple lines

- Unordered lists
- With bullets

1. Ordered lists
2. With numbers

[Links](https://example.com)

| Tables | Are | Supported |
|--------|-----|-----------|
| Cell   | 1   | 2         |

---

Horizontal rules
```

### Complex Example
```go
complexMarkdown := `
# Configuration Guide

This guide explains how to configure **ZeroUI**.

## Installation

` + "```bash\ngo install github.com/user/zeroui\n```" + `

## Usage

Use ZeroUI to manage configurations:

1. **List applications**: ` + "`zeroui list`" + `
2. **Toggle settings**: ` + "`zeroui toggle <app> <setting>`" + `
3. **View status**: ` + "`zeroui status`" + `

### Available Applications

| App | Config Path | Status |
|-----|-------------|--------|
| VSCode | ~/.vscode/settings.json | ✅ Configured |
| Terminal | ~/.zshrc | ⚠️ Partial |

> **Note**: Always backup your configurations before making changes.

For more information, visit the [documentation](https://github.com/user/zeroui).
`

renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(80),
    glamour.WithEmoji(),
)

output, err := renderer.Render(complexMarkdown)
```

## Integration Patterns

### With Bubble Tea
```go
type Model struct {
    content  string
    rendered string
    renderer *glamour.TermRenderer
    viewport viewport.Model
}

func NewModel() Model {
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    
    return Model{
        renderer: renderer,
        viewport: viewport.New(80, 24),
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Update renderer width on terminal resize
        m.renderer, _ = glamour.NewTermRenderer(
            glamour.WithAutoStyle(),
            glamour.WithWordWrap(msg.Width-4),
        )
        m.rendered, _ = m.renderer.Render(m.content)
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height
        m.viewport.SetContent(m.rendered)
    }
    
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return m.viewport.View()
}
```

### File Reading and Rendering
```go
func RenderMarkdownFile(filename string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }
    
    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    if err != nil {
        return "", err
    }
    
    return renderer.Render(string(content))
}
```

### Help System Integration
```go
func ShowHelp(command string) error {
    helpContent := map[string]string{
        "toggle": `
# Toggle Command

Toggle configuration values for applications.

## Usage

` + "```bash\nzeroui toggle <app> <key> [value]\n```" + `

## Examples

Toggle theme:
` + "```bash\nzeroui toggle vscode theme\n```" + `

Set specific value:
` + "```bash\nzeroui toggle terminal font-size 14\n```" + `
        `,
        "list": `
# List Command

List available applications and their configurations.

## Usage

` + "```bash\nzeroui list [app]\n```" + `

## Examples

List all apps:
` + "```bash\nzeroui list\n```" + `

List app details:
` + "```bash\nzeroui list vscode\n```" + `
        `,
    }
    
    content, exists := helpContent[command]
    if !exists {
        content = "Help not available for command: " + command
    }
    
    rendered, err := glamour.Render(content, "auto")
    if err != nil {
        return err
    }
    
    fmt.Print(rendered)
    return nil
}
```

## ZeroUI Specific Examples

### Configuration Documentation
```go
func GenerateConfigDocs(apps []App) string {
    var md strings.Builder
    
    md.WriteString("# ZeroUI Applications\n\n")
    md.WriteString("Available applications and their configurations:\n\n")
    
    for _, app := range apps {
        md.WriteString(fmt.Sprintf("## %s\n\n", app.Name))
        md.WriteString(fmt.Sprintf("**Path**: `%s`\n\n", app.ConfigPath))
        
        if len(app.ToggleableFields) > 0 {
            md.WriteString("### Toggleable Fields\n\n")
            md.WriteString("| Field | Type | Description |\n")
            md.WriteString("|-------|------|-------------|\n")
            
            for _, field := range app.ToggleableFields {
                md.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", 
                    field.Name, field.Type, field.Description))
            }
            md.WriteString("\n")
        }
        
        if len(app.Examples) > 0 {
            md.WriteString("### Examples\n\n")
            for _, example := range app.Examples {
                md.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", example))
            }
        }
    }
    
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(100),
    )
    
    rendered, _ := renderer.Render(md.String())
    return rendered
}
```

### Status Reporting
```go
func RenderStatusReport(status ConfigStatus) string {
    var md strings.Builder
    
    md.WriteString("# Configuration Status\n\n")
    
    if len(status.Errors) > 0 {
        md.WriteString("## ❌ Errors\n\n")
        for _, err := range status.Errors {
            md.WriteString(fmt.Sprintf("- **%s**: %s\n", err.App, err.Message))
        }
        md.WriteString("\n")
    }
    
    if len(status.Warnings) > 0 {
        md.WriteString("## ⚠️ Warnings\n\n")
        for _, warn := range status.Warnings {
            md.WriteString(fmt.Sprintf("- **%s**: %s\n", warn.App, warn.Message))
        }
        md.WriteString("\n")
    }
    
    md.WriteString("## ✅ Configured Applications\n\n")
    md.WriteString("| Application | Status | Last Modified |\n")
    md.WriteString("|-------------|--------|---------------|\n")
    
    for _, app := range status.Apps {
        statusIcon := "✅"
        if !app.Valid {
            statusIcon = "❌"
        }
        
        md.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
            app.Name, statusIcon, app.LastModified.Format("2006-01-02 15:04")))
    }
    
    renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
    rendered, _ := renderer.Render(md.String())
    return rendered
}
```

### Command Output Formatting
```go
func FormatCommandOutput(cmd, output, description string) string {
    md := fmt.Sprintf(`
## Command: %s

%s

### Output

` + "```\n%s\n```" + `
`, cmd, description, output)
    
    rendered, _ := glamour.Render(md, "auto")
    return rendered
}
```