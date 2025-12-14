# Huh Cheatsheet

**Repository:** https://github.com/charmbracelet/huh  
**Purpose:** Interactive forms and prompts for terminal applications

## Basic Form Structure

### Simple Form
```go
package main

import (
    "fmt"
    "github.com/charmbracelet/huh"
)

func main() {
    var name string
    var email string
    var subscribe bool

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("What's your name?").
                Value(&name).
                Validate(func(s string) error {
                    if len(s) < 2 {
                        return fmt.Errorf("name must be at least 2 characters")
                    }
                    return nil
                }),

            huh.NewInput().
                Title("Email address").
                Value(&email).
                Validate(func(s string) error {
                    if !strings.Contains(s, "@") {
                        return fmt.Errorf("invalid email format")
                    }
                    return nil
                }),

            huh.NewConfirm().
                Title("Subscribe to newsletter?").
                Value(&subscribe),
        ),
    )

    err := form.Run()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Name: %s\nEmail: %s\nSubscribe: %t\n", name, email, subscribe)
}
```

## Field Types

### Text Input
```go
var name string

input := huh.NewInput().
    Title("Enter your name").
    Description("First and last name").
    Placeholder("John Doe").
    Value(&name).
    CharLimit(50).
    Validate(func(s string) error {
        if len(strings.TrimSpace(s)) == 0 {
            return fmt.Errorf("name is required")
        }
        return nil
    })
```

### Password Input
```go
var password string

passwordField := huh.NewPassword().
    Title("Enter password").
    Description("Minimum 8 characters").
    Value(&password).
    Validate(func(s string) error {
        if len(s) < 8 {
            return fmt.Errorf("password must be at least 8 characters")
        }
        return nil
    })
```

### Confirmation
```go
var confirmed bool

confirm := huh.NewConfirm().
    Title("Are you sure?").
    Description("This action cannot be undone").
    Affirmative("Yes!").
    Negative("No way").
    Value(&confirmed)
```

### Single Select
```go
var theme string

themeSelect := huh.NewSelect[string]().
    Title("Choose a theme").
    Description("Select your preferred theme").
    Options(
        huh.NewOption("Dark Theme", "dark").
            Selected(true), // Default selection
        huh.NewOption("Light Theme", "light"),
        huh.NewOption("Auto", "auto"),
    ).
    Value(&theme)
```

### Multi Select
```go
var features []string

featureSelect := huh.NewMultiSelect[string]().
    Title("Select features").
    Description("Choose features to enable").
    Options(
        huh.NewOption("Syntax Highlighting", "syntax"),
        huh.NewOption("Auto Save", "autosave").
            Selected(true), // Pre-selected
        huh.NewOption("Line Numbers", "numbers"),
        huh.NewOption("Word Wrap", "wrap"),
    ).
    Value(&features).
    Filterable(true). // Enable filtering
    Limit(2)         // Maximum selections
```

### File Picker
```go
var configFile string

filePicker := huh.NewFilePicker().
    Title("Select config file").
    Description("Choose a configuration file").
    AllowedTypes([]string{".yaml", ".yml", ".json"}).
    CurrentDirectory(".").
    Value(&configFile)
```

### Note (Display Only)
```go
note := huh.NewNote().
    Title("Welcome!").
    Description("This wizard will help you configure the application.\n\nPress Enter to continue.")
```

## Advanced Field Options

### Input Field Options
```go
input := huh.NewInput().
    Title("Username").
    Description("Choose a unique username").
    Placeholder("enter username...").
    Prompt("→ ").
    CharLimit(20).
    Value(&username).
    Validate(func(s string) error {
        if len(s) < 3 {
            return fmt.Errorf("username too short")
        }
        if strings.Contains(s, " ") {
            return fmt.Errorf("username cannot contain spaces")
        }
        return nil
    }).
    Inline(true) // Display on single line
```

### Select Options
```go
type Config struct {
    Name string
    Value string
}

var config Config

configSelect := huh.NewSelect[Config]().
    Title("Choose configuration").
    Options(
        huh.NewOption("Development", Config{"dev", "development"}),
        huh.NewOption("Staging", Config{"stage", "staging"}),
        huh.NewOption("Production", Config{"prod", "production"}),
    ).
    Value(&config).
    OptionsFunc(func() []huh.Option[Config] {
        // Dynamic options
        configs := loadConfigs()
        var options []huh.Option[Config]
        for _, c := range configs {
            options = append(options, huh.NewOption(c.Name, c))
        }
        return options
    })
```

## Form Groups and Layout

### Multiple Groups
```go
var (
    // Personal info
    name  string
    email string
    
    // Preferences  
    theme    string
    features []string
    
    // Confirmation
    confirmed bool
)

form := huh.NewForm(
    // Group 1: Personal Information
    huh.NewGroup(
        huh.NewNote().
            Title("Personal Information").
            Description("Please provide your details"),
            
        huh.NewInput().
            Title("Name").
            Value(&name),
            
        huh.NewInput().
            Title("Email").
            Value(&email),
    ).Title("Step 1"),

    // Group 2: Preferences
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Theme").
            Options(
                huh.NewOption("Dark", "dark"),
                huh.NewOption("Light", "light"),
            ).
            Value(&theme),
            
        huh.NewMultiSelect[string]().
            Title("Features").
            Options(
                huh.NewOption("Auto Save", "autosave"),
                huh.NewOption("Syntax Highlighting", "syntax"),
            ).
            Value(&features),
    ).Title("Step 2"),

    // Group 3: Confirmation
    huh.NewGroup(
        huh.NewConfirm().
            Title("Create account?").
            Value(&confirmed),
    ).Title("Step 3"),
)
```

### Conditional Fields
```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("User type").
            Options(
                huh.NewOption("Regular", "regular"),
                huh.NewOption("Admin", "admin"),
            ).
            Value(&userType),
    ),
    
    // Conditional group - only show if admin selected
    huh.NewGroup(
        huh.NewInput().
            Title("Admin password").
            Value(&adminPassword),
    ).WithHideFunc(func() bool {
        return userType != "admin"
    }),
)
```

## Validation Patterns

### Common Validators
```go
// Required field
func required(s string) error {
    if len(strings.TrimSpace(s)) == 0 {
        return fmt.Errorf("field is required")
    }
    return nil
}

// Email validation
func validateEmail(email string) error {
    if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
        return fmt.Errorf("invalid email format")
    }
    return nil
}

// Number validation
func validatePort(s string) error {
    port, err := strconv.Atoi(s)
    if err != nil {
        return fmt.Errorf("must be a number")
    }
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
}

// File exists validation
func validateFile(path string) error {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("file does not exist")
    }
    return nil
}
```

### Using Validators
```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("Name").
            Value(&name).
            Validate(required),
            
        huh.NewInput().
            Title("Email").
            Value(&email).
            Validate(func(s string) error {
                if err := required(s); err != nil {
                    return err
                }
                return validateEmail(s)
            }),
            
        huh.NewInput().
            Title("Port").
            Value(&portStr).
            Validate(validatePort),
    ),
)
```

## Styling and Theming

### Custom Styling
```go
import "github.com/charmbracelet/lipgloss"

theme := huh.ThemeCharm()
theme.Focused.Title = lipgloss.NewStyle().
    Foreground(lipgloss.Color("212")).
    Bold(true)

theme.Focused.Description = lipgloss.NewStyle().
    Foreground(lipgloss.Color("244"))

form := huh.NewForm(groups...).
    WithTheme(theme)
```

### Built-in Themes
```go
// Available themes
huh.ThemeBase()      // Minimal theme
huh.ThemeDracula()   // Dracula colors
huh.ThemeCatppuccin() // Catppuccin colors  
huh.ThemeCharm()     // Charm colors (default)
```

## Integration Patterns

### With Bubble Tea
```go
type Model struct {
    form     *huh.Form
    complete bool
    result   FormData
}

func (m Model) Init() tea.Cmd {
    return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    // Handle form updates
    form, formCmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }
    cmd = tea.Batch(cmd, formCmd)
    
    // Check if form is complete
    if m.form.State == huh.StateCompleted {
        m.complete = true
        m.result = FormData{
            Name:  m.form.GetString("name"),
            Email: m.form.GetString("email"),
        }
    }
    
    return m, cmd
}

func (m Model) View() string {
    if m.complete {
        return "Form submitted successfully!"
    }
    return m.form.View()
}
```

### Error Handling
```go
form := huh.NewForm(groups...)

err := form.Run()
if err != nil {
    switch {
    case errors.Is(err, huh.ErrUserAborted):
        fmt.Println("Form was cancelled")
    case errors.Is(err, huh.ErrValidationFailed):
        fmt.Println("Validation failed")
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return
}
```

## ZeroUI Integration Examples

### Configuration Selection
```go
func selectConfig() (string, error) {
    var configPath string
    
    // Get available configs
    configs, err := listConfigs()
    if err != nil {
        return "", err
    }
    
    var options []huh.Option[string]
    for _, config := range configs {
        options = append(options, huh.NewOption(config.Name, config.Path))
    }
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select configuration").
                Description("Choose a config file to modify").
                Options(options...).
                Value(&configPath),
        ),
    )
    
    if err := form.Run(); err != nil {
        return "", err
    }
    
    return configPath, nil
}
```

### Setting Values
```go
func promptForSettings(app string) (map[string]interface{}, error) {
    var theme string
    var fontSize int
    var enableFeatures []string
    var autoSave bool
    
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title(fmt.Sprintf("Configure %s", app)).
                Description("Modify application settings"),
                
            huh.NewSelect[string]().
                Title("Theme").
                Options(
                    huh.NewOption("Dark", "dark"),
                    huh.NewOption("Light", "light"),
                    huh.NewOption("Auto", "auto"),
                ).
                Value(&theme),
                
            huh.NewInput().
                Title("Font Size").
                Description("Font size in pixels").
                Value(&fontSize).
                Validate(func(s string) error {
                    size, err := strconv.Atoi(s)
                    if err != nil {
                        return fmt.Errorf("must be a number")
                    }
                    if size < 8 || size > 72 {
                        return fmt.Errorf("size must be between 8 and 72")
                    }
                    return nil
                }),
                
            huh.NewMultiSelect[string]().
                Title("Features").
                Options(
                    huh.NewOption("Syntax Highlighting", "syntax"),
                    huh.NewOption("Line Numbers", "numbers"),
                    huh.NewOption("Word Wrap", "wrap"),
                ).
                Value(&enableFeatures),
                
            huh.NewConfirm().
                Title("Auto Save").
                Description("Automatically save changes").
                Value(&autoSave),
        ),
    )
    
    if err := form.Run(); err != nil {
        return nil, err
    }
    
    settings := map[string]interface{}{
        "theme":    theme,
        "fontSize": fontSize,
        "features": enableFeatures,
        "autoSave": autoSave,
    }
    
    return settings, nil
}
```