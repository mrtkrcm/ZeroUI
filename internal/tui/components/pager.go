package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigPager displays the original configuration file with syntax highlighting
type ConfigPager struct {
	viewport    viewport.Model
	content     string
	filePath    string
	fileContent string
	ready       bool
	width       int
	height      int
}

// Styles for the pager
var (
	pagerTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	
	pagerInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	
	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D7D7D")).
			MarginRight(1)
	
	syntaxKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))
	
	syntaxValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))
	
	syntaxCommentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A5A6")).
			Italic(true)
	
	syntaxSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12")).
			Bold(true)
)

// NewConfigPager creates a new configuration file pager
func NewConfigPager() *ConfigPager {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD"))
	
	return &ConfigPager{
		viewport: vp,
		ready:    false,
	}
}

// SetContent sets the file content to display
func (p *ConfigPager) SetContent(filePath, content string) {
	p.filePath = filePath
	p.fileContent = content
	p.updateContent()
}

// updateContent processes the content with line numbers and syntax highlighting
func (p *ConfigPager) updateContent() {
	if p.fileContent == "" {
		p.content = "No configuration file content available"
		p.viewport.SetContent(p.content)
		return
	}
	
	lines := strings.Split(p.fileContent, "\n")
	var processedLines []string
	
	for i, line := range lines {
		lineNum := lineNumberStyle.Render(fmt.Sprintf("%4d", i+1))
		highlightedLine := p.highlightLine(line)
		processedLines = append(processedLines, fmt.Sprintf("%s %s", lineNum, highlightedLine))
	}
	
	p.content = strings.Join(processedLines, "\n")
	p.viewport.SetContent(p.content)
}

// highlightLine applies syntax highlighting to a configuration line
func (p *ConfigPager) highlightLine(line string) string {
	trimmed := strings.TrimSpace(line)
	
	// Comments
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
		return syntaxCommentStyle.Render(line)
	}
	
	// Section headers (e.g., [section])
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return syntaxSectionStyle.Render(line)
	}
	
	// Key-value pairs
	if strings.Contains(line, "=") || strings.Contains(line, ":") {
		separator := "="
		if strings.Contains(line, ":") && !strings.Contains(line, "=") {
			separator = ":"
		}
		
		parts := strings.SplitN(line, separator, 2)
		if len(parts) == 2 {
			key := syntaxKeyStyle.Render(parts[0])
			value := syntaxValueStyle.Render(separator + parts[1])
			return key + value
		}
	}
	
	return line
}

// Init initializes the pager
func (p *ConfigPager) Init() tea.Cmd {
	return nil
}

// Update handles messages for the pager
func (p *ConfigPager) Update(msg tea.Msg) (*ConfigPager, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		
		headerHeight := 4
		footerHeight := 3
		verticalMarginHeight := headerHeight + footerHeight
		
		if !p.ready {
			p.viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			p.viewport.Style = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD"))
			p.viewport.YPosition = headerHeight
			p.viewport.SetContent(p.content)
			p.ready = true
		} else {
			p.viewport.Width = msg.Width - 4
			p.viewport.Height = msg.Height - verticalMarginHeight
		}
		
	case tea.KeyMsg:
		// Additional key bindings for the pager
		switch msg.String() {
		case "g", "home":
			p.viewport.GotoTop()
		case "G", "end":
			p.viewport.GotoBottom()
		}
	}
	
	// Handle viewport updates
	p.viewport, cmd = p.viewport.Update(msg)
	cmds = append(cmds, cmd)
	
	return p, tea.Batch(cmds...)
}

// View renders the pager
func (p *ConfigPager) View() string {
	if !p.ready {
		return "\n  Initializing..."
	}
	
	// Header
	headerText := fmt.Sprintf("  Viewing: %s", p.filePath)
	header := pagerTitleStyle.Render(headerText)
	
	// Footer with scroll info
	percent := p.viewport.ScrollPercent()
	footerText := fmt.Sprintf("  %3.f%%", percent*100)
	
	helpText := "  ↑/↓: scroll • g/G: top/bottom • q: back to editor"
	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		pagerInfoStyle.Render(helpText),
		strings.Repeat(" ", max(0, p.viewport.Width-len(helpText)-len(footerText))),
		pagerInfoStyle.Render(footerText),
	)
	
	// Combine all parts
	return fmt.Sprintf("%s\n\n%s\n\n%s", header, p.viewport.View(), footer)
}

// SetSize updates the pager dimensions
func (p *ConfigPager) SetSize(width, height int) {
	p.width = width
	p.height = height
	
	headerHeight := 4
	footerHeight := 3
	verticalMarginHeight := headerHeight + footerHeight
	
	p.viewport.Width = width - 4
	p.viewport.Height = height - verticalMarginHeight
}

// Focus sets focus on the pager
func (p *ConfigPager) Focus() {
	// No-op for now
}

// Blur removes focus from the pager
func (p *ConfigPager) Blur() {
	// No-op for now
}