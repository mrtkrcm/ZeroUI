# Lipgloss Cheatsheet

**Repository:** https://github.com/charmbracelet/lipgloss  
**Purpose:** CSS-like styling for terminal applications

## Basic Styling

### Simple Style Creation
```go
import "github.com/charmbracelet/lipgloss"

// Basic style
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1).
    Margin(1, 0)

// Apply style
styledText := style.Render("Hello World")
fmt.Println(styledText)
```

### Chaining Methods
```go
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Italic(false).
    Underline(true).
    Strikethrough(false).
    Reverse(false).
    Blink(false).
    Faint(false)
```

## Colors

### Color Formats
```go
// Hex colors
red := lipgloss.Color("#FF0000")
blue := lipgloss.Color("#0000FF")

// 256-color mode
purple := lipgloss.Color("99")   // Purple
green := lipgloss.Color("46")    // Green

// Named colors (terminal dependent)
yellow := lipgloss.Color("yellow")
```

### Adaptive Colors
```go
adaptiveStyle := lipgloss.NewStyle().
    Foreground(lipgloss.AdaptiveColor{
        Light: "#000000", // Black for light backgrounds
        Dark:  "#FFFFFF", // White for dark backgrounds
    }).
    Background(lipgloss.AdaptiveColor{
        Light: "#FFFFFF", // White for light mode
        Dark:  "#000000", // Black for dark mode
    })
```

### Complete Color Palette
```go
// Standard colors
colors := map[string]lipgloss.Color{
    "black":   lipgloss.Color("0"),
    "red":     lipgloss.Color("1"),
    "green":   lipgloss.Color("2"),
    "yellow":  lipgloss.Color("3"),
    "blue":    lipgloss.Color("4"),
    "magenta": lipgloss.Color("5"),
    "cyan":    lipgloss.Color("6"),
    "white":   lipgloss.Color("7"),
}

// Bright colors
brightColors := map[string]lipgloss.Color{
    "bright_black":   lipgloss.Color("8"),
    "bright_red":     lipgloss.Color("9"),
    "bright_green":   lipgloss.Color("10"),
    "bright_yellow":  lipgloss.Color("11"),
    "bright_blue":    lipgloss.Color("12"),
    "bright_magenta": lipgloss.Color("13"),
    "bright_cyan":    lipgloss.Color("14"),
    "bright_white":   lipgloss.Color("15"),
}
```

## Spacing

### Padding and Margin
```go
// Padding: top, right, bottom, left
style := lipgloss.NewStyle().
    Padding(1, 2, 1, 2)    // 1 top/bottom, 2 left/right

// Margin: top, right, bottom, left  
style = style.Margin(0, 1, 0, 1)    // 1 left/right margin

// Individual sides
style = style.
    PaddingTop(1).
    PaddingRight(2).
    PaddingBottom(1).
    PaddingLeft(2).
    MarginTop(0).
    MarginRight(1).
    MarginBottom(0).
    MarginLeft(1)
```

### Uniform Spacing
```go
// Same padding on all sides
uniformPadding := lipgloss.NewStyle().Padding(1)

// Same margin on all sides  
uniformMargin := lipgloss.NewStyle().Margin(1)

// Horizontal and vertical
horizontalPadding := lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)
verticalPadding := lipgloss.NewStyle().PaddingTop(1).PaddingBottom(1)
```

## Dimensions

### Width and Height
```go
fixedWidth := lipgloss.NewStyle().
    Width(20).          // Exact width
    Height(5).          // Exact height
    MaxWidth(50).       // Maximum width
    MaxHeight(10)       // Maximum height

// Auto-sizing
autoStyle := lipgloss.NewStyle().
    Width(lipgloss.Width("content to measure")).
    Height(lipgloss.Height("content\nwith\nmultiple\nlines"))
```

### Alignment within Dimensions
```go
centeredStyle := lipgloss.NewStyle().
    Width(20).
    Align(lipgloss.Center)  // Center text horizontally

// Alignment options
leftAligned := lipgloss.NewStyle().Align(lipgloss.Left)
rightAligned := lipgloss.NewStyle().Align(lipgloss.Right)
centerAligned := lipgloss.NewStyle().Align(lipgloss.Center)
```

## Borders

### Border Styles
```go
// Built-in border styles
normalBorder := lipgloss.NewStyle().Border(lipgloss.NormalBorder())
roundedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
blockBorder := lipgloss.NewStyle().Border(lipgloss.BlockBorder())
outerHalfBlockBorder := lipgloss.NewStyle().Border(lipgloss.OuterHalfBlockBorder())
innerHalfBlockBorder := lipgloss.NewStyle().Border(lipgloss.InnerHalfBlockBorder())
thickBorder := lipgloss.NewStyle().Border(lipgloss.ThickBorder())
doubleBorder := lipgloss.NewStyle().Border(lipgloss.DoubleBorder())
hiddenBorder := lipgloss.NewStyle().Border(lipgloss.HiddenBorder())
```

### Custom Borders
```go
customBorder := lipgloss.Border{
    Top:         "─",
    Bottom:      "─",
    Left:        "│",
    Right:       "│",
    TopLeft:     "╭",
    TopRight:    "╮",
    BottomLeft:  "╰",
    BottomRight: "╯",
}

style := lipgloss.NewStyle().Border(customBorder)
```

### Selective Borders
```go
// Individual border sides
partialBorder := lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderTop(true).
    BorderBottom(true).
    BorderLeft(false).
    BorderRight(false)

// Border colors
coloredBorder := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("62"))
```

## Layout and Positioning

### Joining Content
```go
// Vertical joining
content1 := "First line"
content2 := "Second line"
content3 := "Third line"

vertical := lipgloss.JoinVertical(
    lipgloss.Left,    // Alignment
    content1,
    content2,
    content3,
)

// Horizontal joining
horizontal := lipgloss.JoinHorizontal(
    lipgloss.Top,     // Alignment
    "Left",
    "Middle", 
    "Right",
)
```

### Positioning
```go
// Place content vertically within a height
positioned := lipgloss.PlaceVertical(
    10,                    // Total height
    lipgloss.Center,       // Vertical position
    "Centered content",
)

// Place content horizontally within a width
horizontal := lipgloss.PlaceHorizontal(
    20,                    // Total width
    lipgloss.Center,       // Horizontal position
    "Centered",
)

// Position options
lipgloss.Top     // or lipgloss.Left
lipgloss.Center
lipgloss.Bottom  // or lipgloss.Right
```

## Advanced Styling

### Conditional Styling
```go
func getStyle(isError bool, isFocused bool) lipgloss.Style {
    style := lipgloss.NewStyle()
    
    if isError {
        style = style.Foreground(lipgloss.Color("9")) // Red
    } else {
        style = style.Foreground(lipgloss.Color("10")) // Green
    }
    
    if isFocused {
        style = style.Bold(true).Border(lipgloss.RoundedBorder())
    }
    
    return style
}
```

### Style Inheritance
```go
baseStyle := lipgloss.NewStyle().
    Padding(1).
    Margin(0, 1)

// Inherit from base style
titleStyle := baseStyle.Copy().
    Bold(true).
    Foreground(lipgloss.Color("12"))

errorStyle := baseStyle.Copy().
    Foreground(lipgloss.Color("9")).
    Background(lipgloss.Color("1"))
```

### Transformations
```go
// Transform text
upperStyle := lipgloss.NewStyle().Transform(strings.ToUpper)
result := upperStyle.Render("hello world") // "HELLO WORLD"

// Custom transform
reverseStyle := lipgloss.NewStyle().Transform(func(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
})
```

## Measuring and Utilities

### Text Measurements
```go
text := "Hello, World!"

width := lipgloss.Width(text)
height := lipgloss.Height(text)

// Multiline text
multiline := "Line 1\nLine 2\nLine 3"
multiWidth := lipgloss.Width(multiline)   // Width of longest line
multiHeight := lipgloss.Height(multiline) // Number of lines
```

### Style Information
```go
style := lipgloss.NewStyle().
    Width(20).
    Height(5).
    Padding(1, 2)

// Get style properties
horizontalMargin := style.GetHorizontalFrameSize()  // Padding + border + margin
verticalMargin := style.GetVerticalFrameSize()      // Padding + border + margin
```

## Common Patterns

### Card Layout
```go
func RenderCard(title, content string) string {
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("12")).
        Background(lipgloss.Color("236")).
        Padding(0, 1).
        Width(30)
    
    contentStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("238")).
        Padding(1).
        Width(30)
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        titleStyle.Render(title),
        contentStyle.Render(content),
    )
}
```

### List Items
```go
func RenderListItem(text string, selected bool) string {
    style := lipgloss.NewStyle().Padding(0, 1)
    
    if selected {
        style = style.
            Background(lipgloss.Color("62")).
            Foreground(lipgloss.Color("15")).
            Bold(true)
    } else {
        style = style.Foreground(lipgloss.Color("7"))
    }
    
    prefix := " "
    if selected {
        prefix = ">"
    }
    
    return style.Render(prefix + " " + text)
}
```

### Progress Bar
```go
func RenderProgressBar(progress float64, width int) string {
    filled := int(float64(width) * progress)
    bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
    
    style := lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("12"))
    
    return style.Render(bar)
}
```

### Status Messages
```go
var (
    successStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("10")).
        Bold(true).
        Render

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("9")).
        Bold(true).
        Render

    warningStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("11")).
        Bold(true).
        Render

    infoStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("12")).
        Bold(true).
        Render
)

func RenderStatus(level, message string) string {
    switch level {
    case "success":
        return successStyle("✓ " + message)
    case "error":
        return errorStyle("✗ " + message)
    case "warning":
        return warningStyle("⚠ " + message)
    case "info":
        return infoStyle("ℹ " + message)
    default:
        return message
    }
}
```

## ZeroUI Theme Examples

### App Theme
```go
// ZeroUI color scheme
var (
    primaryColor   = lipgloss.Color("#7D56F4")  // Purple
    secondaryColor = lipgloss.Color("#F25D94")  // Pink
    accentColor    = lipgloss.Color("#04B575")  // Green
    errorColor     = lipgloss.Color("#ED567A")  // Red
    warningColor   = lipgloss.Color("#FFBD39")  // Yellow
    textColor      = lipgloss.Color("#FAFAFA")  // Light gray
    mutedColor     = lipgloss.Color("#6C7086")  // Muted
)

// Base styles
baseStyle := lipgloss.NewStyle().
    Foreground(textColor)

titleStyle := baseStyle.Copy().
    Bold(true).
    Foreground(primaryColor).
    Border(lipgloss.RoundedBorder(), false, false, true, false).
    BorderForeground(primaryColor).
    Padding(0, 1)

errorStyle := baseStyle.Copy().
    Foreground(errorColor).
    Bold(true)

successStyle := baseStyle.Copy().
    Foreground(accentColor).
    Bold(true)
```

### Configuration Display
```go
func RenderConfigItem(key, value string, isModified bool) string {
    keyStyle := lipgloss.NewStyle().
        Foreground(primaryColor).
        Bold(true).
        Width(20).
        Align(lipgloss.Right)
    
    valueStyle := lipgloss.NewStyle().
        Foreground(textColor)
    
    if isModified {
        valueStyle = valueStyle.
            Foreground(accentColor).
            Bold(true)
    }
    
    return lipgloss.JoinHorizontal(
        lipgloss.Top,
        keyStyle.Render(key+":"),
        " ",
        valueStyle.Render(value),
    )
}
```

### Interactive Elements
```go
func RenderButton(text string, active bool) string {
    style := lipgloss.NewStyle().
        Padding(0, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(primaryColor)
    
    if active {
        style = style.
            Background(primaryColor).
            Foreground(lipgloss.Color("0")).
            Bold(true)
    } else {
        style = style.
            Foreground(primaryColor)
    }
    
    return style.Render(text)
}
```