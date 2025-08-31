package app

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/tui/registry"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

type TickMsg time.Time
type SparklineMsg []float64

type DelightfulUIModel struct {
	apps          []registry.AppStatus
	selectedIndex int
	width         int
	height        int

	spinner  spinner.Model
	progress progress.Model

	sparklines map[string][]float64
	particles  []Particle
	waves      []Wave

	animationTick int
	showSparkles  bool
	rainbowMode   bool

	styles *styles.Styles
	keyMap key.Binding

	lastInteraction time.Time
	idleAnimations  bool
}

type Particle struct {
	x, y   float64
	vx, vy float64
	life   float64
	char   string
	color  lipgloss.Color
}

type Wave struct {
	offset    int
	speed     float64
	amplitude float64
	frequency float64
}

func NewDelightfulUI() *DelightfulUIModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	apps := registry.GetAppStatuses()
	sparklines := make(map[string][]float64)

	for _, app := range apps {
		sparklines[app.Definition.Name] = generateSparklineData(20)
	}

	return &DelightfulUIModel{
		apps:            apps,
		selectedIndex:   0,
		spinner:         s,
		progress:        p,
		sparklines:      sparklines,
		particles:       []Particle{},
		waves:           generateWaves(),
		styles:          styles.GetStyles(),
		lastInteraction: time.Now(),
		idleAnimations:  false,
	}
}

func (m *DelightfulUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
		sparklineCmd(),
	)
}

func (m *DelightfulUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 40

	case tea.KeyMsg:
		m.lastInteraction = time.Now()
		m.idleAnimations = false

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			m.moveSelection(-1)
			m.createBurstParticles(m.selectedIndex)

		case "down", "j":
			m.moveSelection(1)
			m.createBurstParticles(m.selectedIndex)

		case "left", "h":
			m.moveSelectionHorizontal(-1)
			m.createWaveEffect()

		case "right", "l":
			m.moveSelectionHorizontal(1)
			m.createWaveEffect()

		case "enter", " ":
			if cmd := m.activateApp(); cmd != nil {
				cmds = append(cmds, cmd)
			}
			m.createCelebrationParticles()

		case "r":
			m.rainbowMode = !m.rainbowMode

		case "s":
			m.showSparkles = !m.showSparkles

		case "tab":
			m.cycleTheme()
		}

	case TickMsg:
		m.animationTick++
		m.updateParticles()
		m.updateWaves()

		if time.Since(m.lastInteraction) > 5*time.Second {
			m.idleAnimations = true
			if m.animationTick%30 == 0 {
				m.createIdleParticles()
			}
		}

		cmds = append(cmds, tickCmd())

	case SparklineMsg:
		for name := range m.sparklines {
			m.sparklines[name] = append(m.sparklines[name][1:], rand.Float64())
		}
		cmds = append(cmds, sparklineCmd())

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *DelightfulUIModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")
	b.WriteString(m.renderAppGrid())
	b.WriteString("\n")
	b.WriteString(m.renderSparklines())
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	if m.showSparkles {
		return m.addParticleOverlay(b.String())
	}

	return b.String()
}

func (m *DelightfulUIModel) renderHeader() string {
	title := "✨ ConfigToggle Deluxe ✨"

	if m.rainbowMode {
		title = m.rainbowText(title)
	}

	wave := m.generateWavePattern()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Background(lipgloss.Color("235")).
		Padding(1, 2).
		Width(m.width).
		Align(lipgloss.Center)

	return headerStyle.Render(wave + "\n" + title + "\n" + wave)
}

func (m *DelightfulUIModel) renderAppGrid() string {
	cols := 4
	if m.width < 100 {
		cols = 3
	}
	if m.width < 80 {
		cols = 2
	}

	var rows []string
	var currentRow []string

	for i, app := range m.apps {
		card := m.renderAppCard(app, i == m.selectedIndex)
		currentRow = append(currentRow, card)

		if len(currentRow) >= cols || i == len(m.apps)-1 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *DelightfulUIModel) renderAppCard(app registry.AppStatus, selected bool) string {
	width := (m.width / 4) - 2
	if width < 20 {
		width = 20
	}

	icon := m.getAnimatedIcon(app.Definition.Name)
	status := m.getStatusIndicator(app.HasConfig)

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Height(8).
		Padding(1).
		Margin(1).
		Border(lipgloss.RoundedBorder())

	if selected {
		cardStyle = cardStyle.
			BorderForeground(lipgloss.Color("212")).
			Background(lipgloss.Color("235"))

		if m.animationTick%10 < 5 {
			cardStyle = cardStyle.BorderForeground(lipgloss.Color("213"))
		}
	} else {
		cardStyle = cardStyle.BorderForeground(lipgloss.Color("240"))
	}

	content := fmt.Sprintf("%s %s\n\n%s\n\n%s",
		icon,
		lipgloss.NewStyle().Bold(true).Render(app.Definition.Name),
		status,
		m.renderMiniSparkline(app.Definition.Name))

	return cardStyle.Render(content)
}

func (m *DelightfulUIModel) renderSparklines() string {
	if m.height < 30 {
		return ""
	}

	sparkStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 4)

	selectedApp := m.apps[m.selectedIndex]
	data := m.sparklines[selectedApp.Definition.Name]

	sparkline := m.generateSparkline(data, m.width-10, 5)

	return sparkStyle.Render(fmt.Sprintf("Activity Monitor: %s\n%s",
		selectedApp.Definition.Name, sparkline))
}

func (m *DelightfulUIModel) renderFooter() string {
	help := []string{
		"↑↓←→ Navigate",
		"Enter Select",
		"R Rainbow",
		"S Sparkles",
		"Tab Theme",
		"Q Quit",
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(m.width).
		Align(lipgloss.Center).
		Margin(1, 0)

	progressBar := m.progress.ViewAs(float64(m.selectedIndex+1) / float64(len(m.apps)))

	// Add current theme info
	currentTheme := styles.GetCurrentThemeName()
	themeInfo := fmt.Sprintf("Theme: %s", currentTheme)

	return footerStyle.Render(progressBar + "\n" + strings.Join(help, " • ") + "\n" + themeInfo)
}

func (m *DelightfulUIModel) getAnimatedIcon(name string) string {
	icons := map[string][]string{
		"VSCode":    {"📝", "✏️", "🖊️", "📄"},
		"Ghostty":   {"👻", "💻", "🖥️", "⌨️"},
		"Alacritty": {"🚀", "⚡", "💫", "✨"},
		"Kitty":     {"🐱", "😺", "😸", "😻"},
		"iTerm2":    {"🍎", "🖥️", "💻", "📟"},
		"Neovim":    {"📗", "🌙", "✨", "🎯"},
	}

	iconSet, exists := icons[name]
	if !exists {
		iconSet = []string{"🔧", "⚙️", "🛠️", "🔨"}
	}

	return iconSet[(m.animationTick/10)%len(iconSet)]
}

func (m *DelightfulUIModel) getStatusIndicator(configured bool) string {
	if configured {
		if m.animationTick%20 < 10 {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("● Active")
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("83")).Render("◉ Active")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○ Inactive")
}

func (m *DelightfulUIModel) generateSparkline(data []float64, width, height int) string {
	if len(data) == 0 {
		return ""
	}

	chars := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var result strings.Builder

	for h := height - 1; h >= 0; h-- {
		for i := 0; i < len(data) && i < width; i++ {
			level := int(data[i] * float64(height))
			if level > h {
				result.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(fmt.Sprintf("%d", 50+level*20))).
					Render(chars[min(8, level-h)]))
			} else {
				result.WriteString(" ")
			}
		}
		if h > 0 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (m *DelightfulUIModel) renderMiniSparkline(name string) string {
	data := m.sparklines[name]
	if len(data) < 10 {
		return ""
	}

	chars := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇"}
	var result strings.Builder

	for i := len(data) - 10; i < len(data); i++ {
		idx := int(data[i] * 7)
		result.WriteString(chars[idx])
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("105")).Render(result.String())
}

func (m *DelightfulUIModel) generateWavePattern() string {
	chars := []string{"~", "≈", "∼", "〜", "～"}
	var pattern strings.Builder

	for i := 0; i < m.width; i++ {
		wave := m.waves[0]
		y := math.Sin(float64(i)*wave.frequency + float64(m.animationTick)*wave.speed)
		charIdx := int((y + 1) * 2.5)
		if charIdx >= 0 && charIdx < len(chars) {
			pattern.WriteString(chars[charIdx])
		} else {
			pattern.WriteString("~")
		}
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Render(pattern.String())
}

func (m *DelightfulUIModel) rainbowText(text string) string {
	colors := []string{"196", "202", "208", "214", "220", "226", "190", "154", "118", "82", "46", "47", "48", "49", "50", "51", "45", "39", "33", "27", "21", "57", "93", "129", "165", "201"}

	var result strings.Builder
	for i, ch := range text {
		colorIdx := (i + m.animationTick/5) % len(colors)
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIdx])).
			Render(string(ch)))
	}

	return result.String()
}

func (m *DelightfulUIModel) moveSelection(delta int) {
	m.selectedIndex = (m.selectedIndex + delta + len(m.apps)) % len(m.apps)
}

func (m *DelightfulUIModel) moveSelectionHorizontal(delta int) {
	cols := 4
	if m.width < 100 {
		cols = 3
	}
	if m.width < 80 {
		cols = 2
	}

	row := m.selectedIndex / cols
	col := m.selectedIndex % cols

	col = (col + delta + cols) % cols
	newIdx := row*cols + col

	if newIdx < len(m.apps) {
		m.selectedIndex = newIdx
	}
}

func (m *DelightfulUIModel) activateApp() tea.Cmd {
	if len(m.apps) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.apps) {
		return nil
	}

	selectedApp := m.apps[m.selectedIndex]
	return SelectAppCmd(selectedApp.Definition.Name)
}

func (m *DelightfulUIModel) cycleTheme() {
	styles.CycleTheme()
	m.styles = styles.GetStyles()

	// Create theme change particles for visual feedback
	m.createThemeChangeParticles(styles.GetCurrentThemeName())
}

func (m *DelightfulUIModel) createThemeChangeParticles(themeName string) {
	centerX := float64(m.width / 2)
	centerY := float64(m.height / 4)

	// Create particles with theme-appropriate colors
	themeColors := []string{"205", "212", "99", "105", "203", "213", "214"}

	for i := 0; i < 20; i++ {
		angle := float64(i) * (math.Pi * 2 / 20)
		speed := rand.Float64()*2 + 1

		m.particles = append(m.particles, Particle{
			x:     centerX,
			y:     centerY,
			vx:    math.Cos(angle) * speed,
			vy:    math.Sin(angle)*speed - 1,
			life:  1.0,
			char:  "🎨",
			color: lipgloss.Color(themeColors[rand.Intn(len(themeColors))]),
		})
	}
}

func (m *DelightfulUIModel) createBurstParticles(index int) {
	cols := 4
	x := float64((index%cols)*(m.width/cols) + m.width/cols/2)
	y := float64((index/cols)*10 + 5)

	for i := 0; i < 10; i++ {
		angle := float64(i) * (math.Pi * 2 / 10)
		m.particles = append(m.particles, Particle{
			x:     x,
			y:     y,
			vx:    math.Cos(angle) * 2,
			vy:    math.Sin(angle) * 2,
			life:  1.0,
			char:  "✨",
			color: lipgloss.Color(fmt.Sprintf("%d", 50+rand.Intn(200))),
		})
	}
}

func (m *DelightfulUIModel) createWaveEffect() {
	m.waves = append(m.waves, Wave{
		offset:    0,
		speed:     0.1,
		amplitude: 2,
		frequency: 0.1,
	})

	if len(m.waves) > 3 {
		m.waves = m.waves[1:]
	}
}

func (m *DelightfulUIModel) createCelebrationParticles() {
	centerX := float64(m.width / 2)
	centerY := float64(m.height / 2)

	confetti := []string{"🎉", "🎊", "✨", "⭐", "💫", "🌟"}

	for i := 0; i < 30; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := rand.Float64()*3 + 1

		m.particles = append(m.particles, Particle{
			x:     centerX,
			y:     centerY,
			vx:    math.Cos(angle) * speed,
			vy:    math.Sin(angle)*speed - 2,
			life:  1.0,
			char:  confetti[rand.Intn(len(confetti))],
			color: lipgloss.Color(fmt.Sprintf("%d", rand.Intn(256))),
		})
	}
}

func (m *DelightfulUIModel) createIdleParticles() {
	if !m.idleAnimations {
		return
	}

	m.particles = append(m.particles, Particle{
		x:     rand.Float64() * float64(m.width),
		y:     0,
		vx:    (rand.Float64() - 0.5) * 0.5,
		vy:    rand.Float64() + 0.5,
		life:  1.0,
		char:  "✦",
		color: lipgloss.Color("105"),
	})
}

func (m *DelightfulUIModel) updateParticles() {
	var alive []Particle

	for _, p := range m.particles {
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1
		p.life -= 0.02

		if p.life > 0 && p.y < float64(m.height) && p.x >= 0 && p.x < float64(m.width) {
			alive = append(alive, p)
		}
	}

	m.particles = alive
}

func (m *DelightfulUIModel) updateWaves() {
	for i := range m.waves {
		m.waves[i].offset++
	}
}

func (m *DelightfulUIModel) addParticleOverlay(base string) string {
	lines := strings.Split(base, "\n")

	for _, p := range m.particles {
		x := int(p.x)
		y := int(p.y)

		if y >= 0 && y < len(lines) && x >= 0 {
			line := []rune(lines[y])
			if x < len(line) {
				style := lipgloss.NewStyle().Foreground(p.color)
				if p.life < 0.5 {
					style = style.Faint(true)
				}
				line[x] = []rune(style.Render(p.char))[0]
				lines[y] = string(line)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func generateSparklineData(length int) []float64 {
	data := make([]float64, length)
	for i := range data {
		data[i] = rand.Float64()
	}
	return data
}

func generateWaves() []Wave {
	return []Wave{
		{offset: 0, speed: 0.05, amplitude: 1, frequency: 0.1},
		{offset: 10, speed: 0.03, amplitude: 1.5, frequency: 0.08},
		{offset: 20, speed: 0.07, amplitude: 0.8, frequency: 0.12},
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func sparklineCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return SparklineMsg(nil)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
