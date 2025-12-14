# Dependencies Guide

ZeroUI uses modern Go libraries from the Charm ecosystem and essential tools for professional TUI development. This guide provides comprehensive reference for all dependencies with practical examples.

## Core TUI Framework

### Bubble Tea v1.3.6 ⭐

**Repository:** https://github.com/charmbracelet/bubbletea
**Purpose:** Terminal User Interface framework based on The Elm Architecture
**Role:** Core application framework, event handling, component lifecycle

```go
// Model-View-Update Pattern
type Model struct {
    list    list.Model
    state   ViewState
    width   int
    height  int
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        tea.EnterAltScreen,
        tea.SetWindowTitle("ZeroUI"),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
        m.list.SetSize(msg.Width, msg.Height-2)
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return lipgloss.JoinVertical(lipgloss.Left,
        m.list.View(),
        m.helpView(),
    )
}

// Launch application
p := tea.NewProgram(initialModel, tea.WithAltScreen())
p.Run()
```

**Key Messages:**

- `tea.KeyMsg` - Keyboard input
- `tea.WindowSizeMsg` - Terminal resize
- `tea.MouseMsg` - Mouse events
- `tea.BatchMsg` - Multiple commands

**Commands:**

```go
tea.Quit                    // Exit program
tea.Batch(cmd1, cmd2)       // Execute multiple commands
tea.Sequence(cmd1, cmd2)    // Execute commands in sequence
tea.Tick(time.Second)       // Timer command
```

### Bubbles v0.21.0 ⭐

**Repository:** https://github.com/charmbracelet/bubbles
**Purpose:** Collection of common Bubble Tea components
**Role:** List navigation, text input, viewport scrolling, progress indicators

#### List Component

```go
import "github.com/charmbracelet/bubbles/list"

type AppItem struct {
    name        string
    description string
}

func (i AppItem) FilterValue() string { return i.name }
func (i AppItem) Title() string       { return i.name }
func (i AppItem) Description() string { return i.description }

// Create list with custom delegate
delegate := NewCustomDelegate()
items := []list.Item{
    AppItem{name: "Ghostty", description: "Fast terminal emulator"},
    AppItem{name: "VSCode", description: "Code editor"},
}

l := list.New(items, delegate, 50, 20)
l.Title = "Applications"
l.SetShowStatusBar(true)
l.SetFilteringEnabled(true)
```

#### Text Input Component

```go
import "github.com/charmbracelet/bubbles/textinput"

ti := textinput.New()
ti.Placeholder = "Enter configuration value..."
ti.Focus()
ti.CharLimit = 200
ti.Width = 50

// Styling
ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
```

#### Viewport (Scrollable Content)

```go
import "github.com/charmbracelet/bubbles/viewport"

vp := viewport.New(80, 20)
vp.SetContent(longMarkdownContent)

// In Update:
vp, cmd = vp.Update(msg)

// In View:
return vp.View()
```

#### Progress & Spinner

```go
import (
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
)

// Progress bar
prog := progress.New(progress.WithDefaultGradient())
prog.Width = 40
return prog.ViewAs(0.6) // 60% progress

// Spinner
s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
```

### Huh v0.7.0 ⭐

**Repository:** https://github.com/charmbracelet/huh
**Purpose:** Interactive forms and prompts with accessibility support
**Role:** Dynamic configuration editing, form validation, accessibility

```go
var name string
var theme string
var features []string
var fontSize int

form := huh.NewForm(
    huh.NewGroup(
        // Text input with validation
        huh.NewInput().
            Title("Application Name").
            Prompt("? ").
            Value(&name).
            Validate(func(str string) error {
                if len(str) < 2 {
                    return errors.New("name must be at least 2 characters")
                }
                return nil
            }),

        // Select dropdown
        huh.NewSelect[string]().
            Title("Choose theme").
            Options(
                huh.NewOption("Dark Mode", "dark"),
                huh.NewOption("Light Mode", "light"),
                huh.NewOption("Auto", "auto"),
            ).
            Value(&theme),

        // Multi-select with limit
        huh.NewMultiSelect[string]().
            Title("Enable features").
            Options(
                huh.NewOption("Search", "search").Selected(true),
                huh.NewOption("Shortcuts", "shortcuts"),
                huh.NewOption("Themes", "themes"),
            ).
            Limit(3).
            Value(&features),

        // Numeric input
        huh.NewInput().
            Title("Font size").
            Value(&fontSize).
            Validate(func(s string) error {
                if i, err := strconv.Atoi(s); err != nil || i < 8 || i > 72 {
                    return fmt.Errorf("font size must be between 8-72")
                }
                return nil
            }),
    ),
).WithAccessible(os.Getenv("ACCESSIBLE") != "")

// Run standalone or integrate with Bubble Tea
err := form.Run()

// Bubble Tea integration
type Model struct {
    form *huh.Form
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    form, cmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }

    if m.form.State == huh.StateCompleted {
        // Handle form completion
        values := m.form.GetString("name")
    }

    return m, cmd
}
```

**Advanced Features:**

```go
// Dynamic forms with conditional fields
huh.NewInput().
    Title("Dynamic title").
    TitleFunc(func() string { return getDynamicTitle() }, &dependency).
    OptionsFunc(func() []huh.Option[string] {
        return fetchDynamicOptions(dependency)
    }, &dependency)

// Accessibility
form.WithAccessible(true)  // Screen reader compatible mode
```

### Lipgloss v1.1.1 ⭐

**Repository:** https://github.com/charmbracelet/lipgloss
**Purpose:** CSS-like styling for terminal applications
**Role:** Component styling, layout, colors, themes

```go
// Basic styling
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4")).
    Padding(0, 1).
    Margin(1, 0)

styled := titleStyle.Render("ZeroUI Applications")

// Adaptive colors for light/dark terminals
adaptiveStyle := lipgloss.NewStyle().
    Foreground(lipgloss.AdaptiveColor{
        Light: "#000000",
        Dark:  "#FFFFFF",
    })

// Layout helpers
content := lipgloss.JoinVertical(lipgloss.Left,
    titleStyle.Render("Title"),
    bodyStyle.Render("Content"),
    helpStyle.Render("Help text"),
)

// Positioning
centered := lipgloss.PlaceHorizontal(80, lipgloss.Center, content)
```

**Color Systems:**

```go
// Hex colors
lipgloss.Color("#FF0000")

// 256-color mode
lipgloss.Color("205")

// Named colors
lipgloss.Color("red")

// Adaptive colors
lipgloss.AdaptiveColor{
    Light: "#000000",
    Dark:  "#FFFFFF",
}
```

## User Interface & Experience

### Glamour v0.10.0

**Purpose:** Markdown rendering in terminal with syntax highlighting
**Role:** Help system, documentation display, rich text rendering

```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),        // Automatic dark/light detection
    glamour.WithWordWrap(80),       // Wrap at 80 characters
)

markdown := `
# ZeroUI Help

## Navigation
- **↑/↓** Navigate items
- **Enter** Select item
- **/** Filter/search
- **?** Toggle help

## Configuration
ZeroUI supports multiple configuration formats:
` + "`" + `yaml` + "`" + `, ` + "`" + `toml` + "`" + `, ` + "`" + `json` + "`" + `
`

out, err := renderer.Render(markdown)
fmt.Print(out)

// Custom styles
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

### Log v0.4.2

**Purpose:** Beautiful, human-readable structured logging
**Role:** Application logging, debugging, performance tracking

```go
import "github.com/charmbracelet/log"

// Basic usage
log.Info("Starting ZeroUI", "version", "1.0.0", "port", 8080)
log.Warn("Configuration file not found", "path", "config.yaml")
log.Error("Database connection failed", "error", err, "retries", 3)

// Logger configuration
logger := log.NewWithOptions(os.Stderr, log.Options{
    ReportCaller:    true,
    ReportTimestamp: true,
    TimeFormat:      time.Kitchen,
    Prefix:          "zeroui",
})

logger.SetLevel(log.DebugLevel)

// Structured fields and sub-loggers
userLogger := logger.With("user", "john", "session", "abc123")
userLogger.Info("User logged in")
userLogger.Error("Action failed", "action", "save_config")

// Custom styles
styles := log.DefaultStyles()
styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
    SetString("ERROR").
    Padding(0, 1, 0, 1).
    Background(lipgloss.Color("204")).
    Foreground(lipgloss.Color("0"))

logger.SetStyles(styles)

// Performance logging
start := time.Now()
// ... operation ...
logger.Debug("Operation completed",
    "duration", time.Since(start),
    "operation", "config_save",
)
```

## Configuration Management

### Koanf v2.0.1

**Purpose:** Configuration management with multiple sources
**Role:** Loading app configurations, YAML/TOML/JSON parsing

```go
import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/parsers/toml"
    "github.com/knadh/koanf/parsers/json"
)

k := koanf.New(".")

// Load from multiple sources
k.Load(file.Provider("config.yaml"), yaml.Parser())
k.Load(file.Provider("config.toml"), toml.Parser())
k.Load(file.Provider("config.json"), json.Parser())

// Access configuration
appName := k.String("app.name")
port := k.Int("server.port")
features := k.Strings("app.features")

// Unmarshal to struct
type Config struct {
    App struct {
        Name     string   `koanf:"name"`
        Features []string `koanf:"features"`
    } `koanf:"app"`
}

var cfg Config
k.Unmarshal("", &cfg)
```

### Cobra v1.9.1

**Purpose:** CLI framework for command-line applications
**Role:** Command structure, argument parsing, help generation

```go
var rootCmd = &cobra.Command{
    Use:   "zeroui",
    Short: "Zero-configuration UI toolkit manager",
    Long:  "ZeroUI simplifies managing UI configurations across development tools",
    Run: func(cmd *cobra.Command, args []string) {
        // Launch interactive TUI
        app, _ := tui.NewApp("")
        app.Run()
    },
}

var toggleCmd = &cobra.Command{
    Use:   "toggle [app] [key] [value]",
    Short: "Toggle a UI configuration value",
    Args:  cobra.ExactArgs(3),
    Run: func(cmd *cobra.Command, args []string) {
        engine, _ := toggle.NewEngine()
        engine.Toggle(args[0], args[1], args[2])
    },
}

func init() {
    rootCmd.AddCommand(toggleCmd)
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
    rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "show changes without applying")
}
```

## Plugin Architecture

### Go-Plugin v1.7.0 + gRPC v1.74.2

**Purpose:** RPC-based plugin system for cross-language compatibility
**Role:** Plugin discovery, RPC communication, process isolation

```go
// Plugin interface
type ConfigPlugin interface {
    GetInfo(ctx context.Context) (*PluginInfo, error)
    DetectConfig(ctx context.Context) (*ConfigInfo, error)
    ParseConfig(ctx context.Context, path string) (*ConfigData, error)
    WriteConfig(ctx context.Context, path string, data *ConfigData) error
    ValidateField(ctx context.Context, field string, value interface{}) error
}

// Plugin registry
type Registry struct {
    pluginDir string
    plugins   map[string]*plugin.Client
}

func (r *Registry) LoadPlugin(name string) (ConfigPlugin, error) {
    pluginPath := filepath.Join(r.pluginDir, "zeroui-plugin-"+name)

    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig: handshakeConfig,
        Plugins:         pluginMap,
        Cmd:            exec.Command(pluginPath),
        AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
    })

    rpcClient, err := client.Client()
    if err != nil {
        return nil, err
    }

    raw, err := rpcClient.Dispense("config")
    if err != nil {
        return nil, err
    }

    return raw.(ConfigPlugin), nil
}

// Plugin implementation
func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: handshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "config": &ConfigPluginImpl{},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

## Development & Quality

### Air v1.62.0

**Purpose:** Live reloading for Go applications during development
**Role:** Hot reloading, automatic rebuilds

```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

### Testing & Quality

```go
// Testify assertions
import "github.com/stretchr/testify/assert"

func TestApplicationList(t *testing.T) {
    model := NewApplicationList()

    // Set test data
    apps := []ApplicationInfo{
        {Name: "ghostty", Status: "configured"},
        {Name: "vscode", Status: "needs_config"},
    }
    model.SetApplications(apps)

    // Test selection
    assert.Equal(t, "", model.GetSelectedApp())

    // Simulate key press
    model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
    assert.NotNil(t, cmd)
}

// Go Mock generation
//go:generate mockgen -source=plugin.go -destination=mocks/plugin_mock.go
```

## Performance & Observability

### OpenTelemetry v1.37.0

**Purpose:** Observability framework for metrics and tracing
**Role:** Performance monitoring, metrics collection

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
    "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics setup
meter := otel.Meter("zeroui")

// Counter for UI operations
uiOperations, _ := meter.Int64Counter(
    "ui_operations_total",
    metric.WithDescription("Total UI operations"),
)

// Histogram for operation duration
operationDuration, _ := meter.Float64Histogram(
    "operation_duration_seconds",
    metric.WithDescription("Operation duration in seconds"),
)

// Record metrics
ctx := context.Background()
uiOperations.Add(ctx, 1, metric.WithAttributes(
    attribute.String("operation", "config_save"),
    attribute.String("app", "ghostty"),
))

start := time.Now()
// ... operation ...
operationDuration.Record(ctx, time.Since(start).Seconds(),
    metric.WithAttributes(
        attribute.String("operation", "config_save"),
    ))
```

## Key Architecture Patterns

### 1. Component Integration Pattern

```go
type Model struct {
    // Core Bubbles components
    list     list.Model
    textInput textinput.Model
    viewport  viewport.Model
    progress  progress.Model
    spinner   spinner.Model

    // Huh forms
    form     *huh.Form

    // App state
    state    ViewState
    keyMap   KeyMap
    styles   *Styles
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Update components
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)

    if m.form != nil {
        form, cmd := m.form.Update(msg)
        if f, ok := form.(*huh.Form); ok {
            m.form = f
        }
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}
```

### 2. Key Binding Pattern

```go
import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Up    key.Binding
    Down  key.Binding
    Enter key.Binding
    Help  key.Binding
    Quit  key.Binding
}

func NewKeyMap() KeyMap {
    return KeyMap{
        Up: key.NewBinding(
            key.WithKeys("up", "k"),
            key.WithHelp("↑/k", "move up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("down", "j"),
            key.WithHelp("↓/j", "move down"),
        ),
        // ...
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keyMap.Up):
            // Handle up navigation
        case key.Matches(msg, m.keyMap.Down):
            // Handle down navigation
        }
    }
    return m, nil
}
```

### 3. Accessibility Pattern

```go
// Environment-based accessibility detection
opts := accessibility.DetectAccessibilityNeeds()

// Configure components for accessibility
form := huh.NewForm(...)
    .WithAccessible(opts.GetHuhAccessibilityMode())

// Adaptive styling
style := lipgloss.NewStyle()
if opts.GetColorMode() == accessibility.ColorModeHighContrast {
    style = style.
        Foreground(lipgloss.Color("#FFFFFF")).
        Background(lipgloss.Color("#000000"))
}

// Screen reader friendly descriptions
title := opts.GetAccessibleTitle("Applications", "ZeroUI Configuration")
help := opts.GetAccessibleHelp("Select app", "Use arrow keys to navigate applications, press Enter to configure selected application")
```

## Development Commands

```bash
# Development with hot reload
make dev

# Build and test
make build
make test

# Code quality
make lint
make fmt
make security

# Dependency management
go mod tidy
go mod download

# Plugin development
cd plugins/ghostty-rpc
go build -o zeroui-plugin-ghostty-rpc
```

## Best Practices

1. **Component Lifecycle**: Always implement Init(), Update(), and View() for Bubble Tea components
2. **Error Handling**: Use structured logging with context for debugging
3. **Performance**: Batch commands with `tea.Batch()` for efficiency
4. **Accessibility**: Always check environment variables for accessibility needs
5. **Type Safety**: Use Huh's generic types for type-safe forms
6. **Plugin Architecture**: Follow naming convention `zeroui-plugin-{name}` for discovery
7. **Testing**: Use Testify for assertions and GoMock for interface mocking
8. **Observability**: Instrument critical paths with OpenTelemetry metrics

## Version Compatibility

All dependencies are pinned to specific versions for reproducible builds. Core Charm dependencies (Bubble Tea, Bubbles, Huh, Lipgloss) are kept in sync for compatibility.

| Library    | Version | Compatibility Notes                                   |
| ---------- | ------- | ----------------------------------------------------- |
| Bubble Tea | v1.3.6  | Core framework - pin to exact version                 |
| Bubbles    | v0.21.0 | Component library - compatible with Bubble Tea v1.3.x |
| Huh        | v0.7.0  | Forms library - requires Bubble Tea v1.3+             |
| Lipgloss   | v1.1.1  | Styling library - stable API                          |
| Glamour    | v0.10.0 | Markdown rendering - stable                           |
| Log        | v0.4.2  | Logging - stable API                                  |

Run `go mod tidy` to ensure dependencies are current and `make check` to validate compatibility across all dependencies.
