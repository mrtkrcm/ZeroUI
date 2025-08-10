# Bubbles Cheatsheet

**Repository:** https://github.com/charmbracelet/bubbles  
**Purpose:** Collection of common Bubble Tea components and widgets

## List Component

### Basic List Setup
```go
import "github.com/charmbracelet/bubbles/list"

// Define list item
type Item struct {
    title, desc string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

// Create list
items := []list.Item{
    Item{title: "Item 1", desc: "Description 1"},
    Item{title: "Item 2", desc: "Description 2"},
    Item{title: "Item 3", desc: "Description 3"},
}

delegate := list.NewDefaultDelegate()
l := list.New(items, delegate, 0, 0) // Width/height set later
l.Title = "My List"
l.SetShowStatusBar(false)
l.SetFilteringEnabled(false)
l.Styles.Title = titleStyle
```

### Advanced List Configuration
```go
// Custom delegate
type customDelegate struct{}

func (d customDelegate) Height() int                             { return 1 }
func (d customDelegate) Spacing() int                            { return 0 }
func (d customDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    i, ok := item.(Item)
    if !ok {
        return
    }
    
    str := fmt.Sprintf("%d. %s", index+1, i.Title())
    
    fn := lipgloss.NewStyle().
        PaddingLeft(4).
        Render
    
    if index == m.Index() {
        fn = func(s ...string) string {
            return lipgloss.NewStyle().
                PaddingLeft(2).
                Background(lipgloss.Color("62")).
                Foreground(lipgloss.Color("15")).
                Render("> " + strings.Join(s, " "))
        }
    }
    
    fmt.Fprint(w, fn(str))
}

// Use custom delegate
l := list.New(items, customDelegate{}, width, height)
```

### List in Bubble Tea Model
```go
type Model struct {
    list list.Model
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "enter":
            selected := m.list.SelectedItem().(Item)
            // Handle selection
            return m, nil
        }
    case tea.WindowSizeMsg:
        m.list.SetWidth(msg.Width)
        m.list.SetHeight(msg.Height)
    }
    
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return m.list.View()
}
```

## Text Input Component

### Basic Text Input
```go
import "github.com/charmbracelet/bubbles/textinput"

type Model struct {
    textInput textinput.Model
}

func initialModel() Model {
    ti := textinput.New()
    ti.Placeholder = "Enter text..."
    ti.Focus()
    ti.CharLimit = 156
    ti.Width = 20
    
    return Model{
        textInput: ti,
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "esc":
            return m, tea.Quit
        case "enter":
            // Process input
            value := m.textInput.Value()
            return m, nil
        }
    }
    
    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return fmt.Sprintf(
        "Enter something:\n\n%s\n\n%s",
        m.textInput.View(),
        "(esc to quit)",
    )
}
```

### Advanced Text Input Options
```go
ti := textinput.New()
ti.Placeholder = "Type here..."
ti.Focus()
ti.CharLimit = 50
ti.Width = 30

// Styling
ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
ti.BackgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("240"))
ti.CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

// Validation
ti.Validate = func(s string) error {
    if len(s) < 3 {
        return fmt.Errorf("input too short")
    }
    return nil
}

// Password mode
ti.EchoMode = textinput.EchoPassword
ti.EchoCharacter = '*'
```

### Multiple Text Inputs
```go
type Model struct {
    inputs []textinput.Model
    focused int
}

func initialModel() Model {
    inputs := make([]textinput.Model, 3)
    
    inputs[0] = textinput.New()
    inputs[0].Placeholder = "Name"
    inputs[0].Focus()
    
    inputs[1] = textinput.New()
    inputs[1].Placeholder = "Email"
    
    inputs[2] = textinput.New()
    inputs[2].Placeholder = "Password"
    inputs[2].EchoMode = textinput.EchoPassword
    
    return Model{
        inputs: inputs,
        focused: 0,
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab", "shift+tab", "enter", "up", "down":
            s := msg.String()
            
            if s == "enter" && m.focused == len(m.inputs)-1 {
                // Submit form
                return m, tea.Quit
            }
            
            if s == "up" || s == "shift+tab" {
                m.focused--
            } else {
                m.focused++
            }
            
            if m.focused > len(m.inputs)-1 {
                m.focused = 0
            } else if m.focused < 0 {
                m.focused = len(m.inputs) - 1
            }
            
            for i := 0; i < len(m.inputs); i++ {
                if i == m.focused {
                    m.inputs[i].Focus()
                } else {
                    m.inputs[i].Blur()
                }
            }
        }
    }
    
    var cmd tea.Cmd
    for i := range m.inputs {
        m.inputs[i], cmd = m.inputs[i].Update(msg)
    }
    
    return m, cmd
}
```

## Viewport Component

### Basic Viewport
```go
import "github.com/charmbracelet/bubbles/viewport"

type Model struct {
    viewport viewport.Model
    content  string
}

func initialModel() Model {
    vp := viewport.New(30, 10)
    vp.SetContent("Very long content that needs scrolling...\n" +
                  "Line 2\nLine 3\nLine 4\n...")
    
    return Model{
        viewport: vp,
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    }
    
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return m.viewport.View()
}
```

### Dynamic Content Viewport
```go
type Model struct {
    viewport viewport.Model
    content  []string
}

func (m *Model) addLine(line string) {
    m.content = append(m.content, line)
    m.viewport.SetContent(strings.Join(m.content, "\n"))
    m.viewport.GotoBottom()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "a":
            m.addLine(fmt.Sprintf("New line %d", len(m.content)+1))
        case "home":
            m.viewport.GotoTop()
        case "end":
            m.viewport.GotoBottom()
        case "pgup":
            m.viewport.LineUp(10)
        case "pgdown":
            m.viewport.LineDown(10)
        }
    }
    
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}
```

## Progress Bar Component

### Basic Progress Bar
```go
import "github.com/charmbracelet/bubbles/progress"

type Model struct {
    progress progress.Model
    percent  float64
}

func initialModel() Model {
    return Model{
        progress: progress.New(
            progress.WithDefaultGradient(),
            progress.WithWidth(40),
            progress.WithoutPercentage(),
        ),
        percent: 0.0,
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "+", "=":
            m.percent = math.Min(1.0, m.percent+0.1)
        case "-", "_":
            m.percent = math.Max(0.0, m.percent-0.1)
        }
    case progress.FrameMsg:
        progressModel, cmd := m.progress.Update(msg)
        m.progress = progressModel.(progress.Model)
        return m, cmd
    }
    
    return m, nil
}

func (m Model) View() string {
    return fmt.Sprintf(
        "Progress: %.0f%%\n%s\n",
        m.percent*100,
        m.progress.ViewAs(m.percent),
    )
}
```

### Custom Progress Bar Styling
```go
prog := progress.New(
    progress.WithScaledGradient("#FF7CCB", "#FDFF8C"), // Custom gradient
    progress.WithWidth(50),
    progress.WithoutPercentage(),
)

// Custom colors
prog := progress.New(
    progress.WithSolidFill("#7D56F4"), // Solid color
    progress.WithWidth(40),
)

// Full customization
prog := progress.New(
    progress.WithGradient(
        lipgloss.Color("#FF0000"), // Start color
        lipgloss.Color("#00FF00"), // End color
    ),
    progress.WithWidth(60),
    progress.WithSpring(),          // Smooth animations
    progress.WithoutPercentage(),   // Hide percentage
)
```

## Spinner Component

### Basic Spinner
```go
import "github.com/charmbracelet/bubbles/spinner"

type Model struct {
    spinner  spinner.Model
    loading  bool
}

func initialModel() Model {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    
    return Model{
        spinner: s,
        loading: true,
    }
}

func (m Model) Init() tea.Cmd {
    return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case " ":
            m.loading = !m.loading
        }
    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }
    
    return m, nil
}

func (m Model) View() string {
    if m.loading {
        return fmt.Sprintf("%s Loading...", m.spinner.View())
    }
    return "Done! (space to toggle, q to quit)"
}
```

### Spinner Types and Styling
```go
// Available spinner types
spinner.Line        // |/-\
spinner.Dot         // ⣾⣽⣻⢿⡿⣟⣯⣷
spinner.MiniDot     // ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏
spinner.Jump        // ⢄⢂⢁⡁⡈⡐⡠
spinner.Pulse       // █▉▊▋▌▍▎▏▎▍▌▋▊▉█
spinner.Points      // ∙∘○●○∘
spinner.Globe       // 🌍🌎🌏
spinner.Moon        // 🌑🌒🌓🌔🌕🌖🌗🌘
spinner.Monkey      // 🙈🙈🙉🙊

// Custom styling
s := spinner.New()
s.Spinner = spinner.Dot
s.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("12")).
    Background(lipgloss.Color("240")).
    Padding(0, 1)
```

## Table Component

### Basic Table
```go
import "github.com/charmbracelet/bubbles/table"

func createTable() table.Model {
    columns := []table.Column{
        {Title: "Rank", Width: 4},
        {Title: "City", Width: 10},
        {Title: "Country", Width: 10},
        {Title: "Population", Width: 10},
    }
    
    rows := []table.Row{
        {"1", "Tokyo", "Japan", "37,400,068"},
        {"2", "Delhi", "India", "28,514,000"},
        {"3", "Shanghai", "China", "25,582,000"},
    }
    
    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(7),
    )
    
    // Styling
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(false)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(false)
    t.SetStyles(s)
    
    return t
}
```

### Interactive Table
```go
type Model struct {
    table table.Model
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            if m.table.Focused() {
                m.table.Blur()
            } else {
                m.table.Focus()
            }
        case "q", "ctrl+c":
            return m, tea.Quit
        case "enter":
            // Handle row selection
            selectedRow := m.table.SelectedRow()
            return m, tea.Printf("Selected: %s", selectedRow[1])
        }
    }
    
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}
```

## Combining Components

### Multi-Component Layout
```go
type Model struct {
    list      list.Model
    textInput textinput.Model
    viewport  viewport.Model
    progress  progress.Model
    spinner   spinner.Model
    
    activePanel int
    loading     bool
}

const (
    listPanel = iota
    inputPanel
    viewPanel
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.activePanel = (m.activePanel + 1) % 3
        case "shift+tab":
            m.activePanel = (m.activePanel - 1 + 3) % 3
        }
    }
    
    // Update active component
    switch m.activePanel {
    case listPanel:
        m.list, cmd = m.list.Update(msg)
    case inputPanel:
        m.textInput, cmd = m.textInput.Update(msg)
    case viewPanel:
        m.viewport, cmd = m.viewport.Update(msg)
    }
    cmds = append(cmds, cmd)
    
    // Always update spinner if loading
    if m.loading {
        m.spinner, cmd = m.spinner.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    left := lipgloss.JoinVertical(lipgloss.Left,
        m.list.View(),
        m.textInput.View(),
    )
    
    right := m.viewport.View()
    
    if m.loading {
        right = m.spinner.View() + " " + right
    }
    
    return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
```

## ZeroUI Integration Examples

### Application Selector
```go
type AppItem struct {
    name   string
    path   string
    status string
}

func (i AppItem) Title() string       { return i.name }
func (i AppItem) Description() string { return fmt.Sprintf("%s • %s", i.path, i.status) }
func (i AppItem) FilterValue() string { return i.name }

func createAppList(apps []ConfigApp) list.Model {
    items := make([]list.Item, len(apps))
    for i, app := range apps {
        status := "✅ Configured"
        if !app.Valid {
            status = "❌ Invalid"
        }
        
        items[i] = AppItem{
            name:   app.Name,
            path:   app.ConfigPath,
            status: status,
        }
    }
    
    delegate := list.NewDefaultDelegate()
    delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
        Background(lipgloss.Color("62")).
        Foreground(lipgloss.Color("15"))
    
    l := list.New(items, delegate, 0, 0)
    l.Title = "ZeroUI Applications"
    l.Styles.Title = lipgloss.NewStyle().
        Background(lipgloss.Color("62")).
        Foreground(lipgloss.Color("15")).
        Padding(0, 1)
    
    return l
}
```

### Configuration Input Form
```go
type ConfigForm struct {
    inputs   []textinput.Model
    labels   []string
    focused  int
    values   map[string]string
}

func NewConfigForm(fields []ConfigField) ConfigForm {
    inputs := make([]textinput.Model, len(fields))
    labels := make([]string, len(fields))
    
    for i, field := range fields {
        ti := textinput.New()
        ti.Placeholder = field.DefaultValue
        ti.CharLimit = 100
        ti.Width = 30
        
        if i == 0 {
            ti.Focus()
        }
        
        inputs[i] = ti
        labels[i] = field.Name
    }
    
    return ConfigForm{
        inputs: inputs,
        labels: labels,
        values: make(map[string]string),
    }
}

func (f ConfigForm) View() string {
    var b strings.Builder
    
    for i, input := range f.inputs {
        b.WriteString(fmt.Sprintf("%s:\n", f.labels[i]))
        b.WriteString(input.View())
        b.WriteString("\n\n")
    }
    
    return b.String()
}
```

### Status Display with Progress
```go
type StatusModel struct {
    viewport viewport.Model
    progress progress.Model
    spinner  spinner.Model
    
    operations []Operation
    current    int
    loading    bool
}

func (m StatusModel) View() string {
    header := "Configuration Status\n\n"
    
    if m.loading {
        header += m.spinner.View() + " Processing...\n\n"
        
        if len(m.operations) > 0 {
            progress := float64(m.current) / float64(len(m.operations))
            header += fmt.Sprintf("Progress: %s\n\n", m.progress.ViewAs(progress))
        }
    }
    
    content := m.viewport.View()
    
    return header + content
}
```