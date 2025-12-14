# Bubble Tea Cheatsheet

**Repository:** https://github.com/charmbracelet/bubbletea  
**Purpose:** Terminal User Interface framework based on The Elm Architecture

## Core Architecture

### The Model-View-Update Pattern
```go
type Model struct {
    // Your application state
    items []string
    cursor int
    selected map[int]struct{}
}

func (m Model) Init() tea.Cmd {
    // Initialize the program
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages and update state
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.items)-1 {
                m.cursor++
            }
        case " ":
            _, ok := m.selected[m.cursor]
            if ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    }
    return m, nil
}

func (m Model) View() string {
    // Render the interface
    s := "Select items:\n\n"
    
    for i, item := range m.items {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        
        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }
        
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, item)
    }
    
    s += "\nPress q to quit.\n"
    return s
}

// Run the program
func main() {
    p := tea.NewProgram(initialModel)
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

## Message Types

### Built-in Messages
```go
type tea.KeyMsg struct {
    Type  tea.KeyType
    Runes []rune
    Alt   bool
}

type tea.WindowSizeMsg struct {
    Width  int
    Height int
}

type tea.MouseMsg struct {
    X      int
    Y      int
    Type   tea.MouseEventType
    Button tea.MouseButton
}

type tea.BatchMsg []tea.Msg
```

### Custom Messages
```go
type TickMsg time.Time
type ErrorMsg error
type DataLoadedMsg []Item

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

func loadDataCmd() tea.Cmd {
    return func() tea.Msg {
        data, err := loadData()
        if err != nil {
            return ErrorMsg(err)
        }
        return DataLoadedMsg(data)
    }
}
```

## Key Handling

### Key Detection
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q", "esc":
            return m, tea.Quit
        case "enter":
            return m, m.processSelection()
        case "up", "k":
            m.cursor = max(0, m.cursor-1)
        case "down", "j":
            m.cursor = min(len(m.items)-1, m.cursor+1)
        case "home":
            m.cursor = 0
        case "end":
            m.cursor = len(m.items) - 1
        case "pgup":
            m.cursor = max(0, m.cursor-10)
        case "pgdown":
            m.cursor = min(len(m.items)-1, m.cursor+10)
        }
    }
    return m, nil
}
```

### Key Types
```go
tea.KeyType constants:
- tea.KeyEnter
- tea.KeyTab
- tea.KeyBackspace
- tea.KeyDelete
- tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight
- tea.KeyHome, tea.KeyEnd
- tea.KeyPgUp, tea.KeyPgDown
- tea.KeyCtrlC, tea.KeyCtrlA, etc.
- tea.KeyF1, tea.KeyF2, etc.
```

## Commands

### Built-in Commands
```go
tea.Quit                          // Exit program
tea.Batch(cmd1, cmd2, cmd3)       // Execute multiple commands
tea.Sequence(cmd1, cmd2)          // Execute commands in sequence
tea.Tick(duration, func)          // Timer command
tea.Every(duration, func)         // Recurring timer
tea.Printf(format, args...)       // Print to stdout
tea.Println(args...)              // Print line to stdout
```

### Program Control
```go
// Alternative modes
tea.WithAltScreen()               // Use alternate screen buffer
tea.WithMouseCellMotion()         // Enable mouse support
tea.WithMouseAllMotion()          // Enable all mouse motion
tea.WithoutSignalHandler()        // Disable signal handling

// Program options
p := tea.NewProgram(model,
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(),
)
```

## Advanced Patterns

### State Management
```go
type ViewState int

const (
    ListView ViewState = iota
    DetailView
    EditView
)

type Model struct {
    state ViewState
    // ... other fields
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.state {
    case ListView:
        return m.updateListView(msg)
    case DetailView:
        return m.updateDetailView(msg)
    case EditView:
        return m.updateEditView(msg)
    }
    return m, nil
}
```

### Async Operations
```go
func (m Model) loadDataCmd() tea.Cmd {
    return func() tea.Msg {
        // This runs in a goroutine
        data, err := fetchFromAPI()
        if err != nil {
            return ErrorMsg{err}
        }
        return DataMsg{data}
    }
}

// In Update:
case DataMsg:
    m.data = msg.data
    m.loading = false
case ErrorMsg:
    m.error = msg.error
    m.loading = false
```

### Sub-models Pattern
```go
type Model struct {
    list     list.Model
    input    textinput.Model
    spinner  spinner.Model
    viewport viewport.Model
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Update all sub-models
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)
    
    m.input, cmd = m.input.Update(msg)
    cmds = append(cmds, cmd)
    
    m.spinner, cmd = m.spinner.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}
```

## Debugging Tips

### Debug Mode
```go
import "log"

// Enable debug logging
func main() {
    if len(os.Args) > 1 && os.Args[1] == "debug" {
        f, err := tea.LogToFile("debug.log", "debug")
        if err != nil {
            fmt.Println("fatal:", err)
            os.Exit(1)
        }
        defer f.Close()
    }
    
    p := tea.NewProgram(initialModel)
    p.Run()
}
```

### Message Logging
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Log all messages in debug mode
    switch msg.(type) {
    case tea.KeyMsg:
        log.Printf("Key: %v", msg)
    case tea.WindowSizeMsg:
        log.Printf("Resize: %v", msg)
    default:
        log.Printf("Message: %T %v", msg, msg)
    }
    
    // ... rest of update logic
}
```

## Common Patterns

### Loading States
```go
type Model struct {
    loading bool
    spinner spinner.Model
    data    []Item
    err     error
}

func (m Model) View() string {
    if m.loading {
        return m.spinner.View() + " Loading..."
    }
    
    if m.err != nil {
        return "Error: " + m.err.Error()
    }
    
    return m.renderData()
}
```

### Navigation Stack
```go
type Model struct {
    stack []ViewModel
}

func (m Model) push(view ViewModel) Model {
    m.stack = append(m.stack, view)
    return m
}

func (m Model) pop() Model {
    if len(m.stack) > 1 {
        m.stack = m.stack[:len(m.stack)-1]
    }
    return m
}

func (m Model) current() ViewModel {
    return m.stack[len(m.stack)-1]
}
```

### Input Validation
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            if err := m.validate(); err != nil {
                m.error = err
                return m, nil
            }
            return m, m.submit()
        }
    }
    return m, nil
}

func (m Model) validate() error {
    if len(strings.TrimSpace(m.input)) == 0 {
        return errors.New("input cannot be empty")
    }
    return nil
}
```