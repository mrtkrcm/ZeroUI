package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AnimatedListModel struct {
	list          list.Model
	items         []list.Item
	width         int
	height        int
	animationTick int
	scrollOffset  float64
	targetOffset  float64
	velocity      float64
	
	// Visual effects
	ripples       []Ripple
	glowIntensity float64
	pulsePhase    float64
	
	// Easter eggs
	konamiCode    []string
	konamiIndex   int
	secretMode    bool
}

type Ripple struct {
	x, y   int
	radius float64
	maxRadius float64
	color  lipgloss.Color
}

type AnimatedItem struct {
	title       string
	description string
	icon        string
	tags        []string
}

func (i AnimatedItem) Title() string       { return i.title }
func (i AnimatedItem) Description() string { return i.description }
func (i AnimatedItem) FilterValue() string { return i.title }

func NewAnimatedList() *AnimatedListModel {
	items := []list.Item{
		AnimatedItem{
			title:       "Visual Studio Code",
			description: "The editor that thinks it's an OS",
			icon:        "📝",
			tags:        []string{"editor", "microsoft", "popular"},
		},
		AnimatedItem{
			title:       "Neovim",
			description: "For those who dream in modal editing",
			icon:        "🌙",
			tags:        []string{"editor", "vim", "terminal"},
		},
		AnimatedItem{
			title:       "Ghostty",
			description: "The spooky-fast terminal",
			icon:        "👻",
			tags:        []string{"terminal", "fast", "new"},
		},
		AnimatedItem{
			title:       "Alacritty",
			description: "GPU-accelerated terminal blazingly fast",
			icon:        "🚀",
			tags:        []string{"terminal", "rust", "gpu"},
		},
		AnimatedItem{
			title:       "Kitty",
			description: "The terminal with nine lives",
			icon:        "🐱",
			tags:        []string{"terminal", "graphics", "features"},
		},
		AnimatedItem{
			title:       "iTerm2",
			description: "macOS terminal on steroids",
			icon:        "🍎",
			tags:        []string{"terminal", "macos", "features"},
		},
		AnimatedItem{
			title:       "Warp",
			description: "The terminal for the 21st century",
			icon:        "⚡",
			tags:        []string{"terminal", "ai", "modern"},
		},
		AnimatedItem{
			title:       "Zed",
			description: "Multiplayer code editor at the speed of thought",
			icon:        "⚡",
			tags:        []string{"editor", "collaborative", "rust"},
		},
	}
	
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("212")).
		Foreground(lipgloss.Color("212")).
		Bold(true).
		Padding(0, 0, 0, 1)
	
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("212")).
		Foreground(lipgloss.Color("245")).
		Padding(0, 0, 0, 1)
	
	l := list.New(items, delegate, 0, 0)
	l.Title = "🎯 App Selector Deluxe"
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 1)
	
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	
	l.KeyMap.CursorUp.SetKeys("up", "k")
	l.KeyMap.CursorDown.SetKeys("down", "j")
	
	return &AnimatedListModel{
		list:        l,
		items:       items,
		konamiCode:  []string{"up", "up", "down", "down", "left", "right", "left", "right", "b", "a"},
		konamiIndex: 0,
	}
}

func (m *AnimatedListModel) Init() tea.Cmd {
	return tea.Batch(
		animationTick(),
		m.list.StartSpinner(),
	)
}

func (m *AnimatedListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		
	case tea.KeyMsg:
		// Check for Konami code
		if m.checkKonamiCode(msg.String()) {
			m.secretMode = true
			m.createCelebration()
		}
		
		// Handle smooth scrolling
		switch msg.String() {
		case "up", "k":
			m.targetOffset -= 1
			m.createRipple(m.width/2, m.height/2)
			
		case "down", "j":
			m.targetOffset += 1
			m.createRipple(m.width/2, m.height/2)
			
		case "enter":
			m.createExplosion()
			
		case "g":
			m.glowIntensity = 1.0
		}
		
		// Update list
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case TickMsg:
		m.animationTick++
		m.updateAnimations()
		cmds = append(cmds, animationTick())
	}
	
	// Update smooth scrolling
	m.updateSmoothScroll()
	
	return m, tea.Batch(cmds...)
}

func (m *AnimatedListModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	
	var b strings.Builder
	
	// Render animated header
	b.WriteString(m.renderAnimatedHeader())
	b.WriteString("\n")
	
	// Render list with effects
	listView := m.list.View()
	
	if m.secretMode {
		listView = m.applySecretEffects(listView)
	}
	
	if m.glowIntensity > 0 {
		listView = m.applyGlowEffect(listView)
	}
	
	b.WriteString(listView)
	
	// Render ripple effects
	if len(m.ripples) > 0 {
		return m.overlayRipples(b.String())
	}
	
	return b.String()
}

func (m *AnimatedListModel) renderAnimatedHeader() string {
	// Create animated wave pattern
	wave := make([]string, m.width)
	for i := 0; i < m.width; i++ {
		height := math.Sin(float64(i)*0.1 + float64(m.animationTick)*0.05)
		if height > 0.5 {
			wave[i] = "═"
		} else if height > 0 {
			wave[i] = "─"
		} else if height > -0.5 {
			wave[i] = "╌"
		} else {
			wave[i] = "·"
		}
	}
	
	waveStr := strings.Join(wave, "")
	color := fmt.Sprintf("%d", 50 + int(math.Sin(float64(m.animationTick)*0.02)*50+50))
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(waveStr)
}

func (m *AnimatedListModel) checkKonamiCode(key string) bool {
	if key == m.konamiCode[m.konamiIndex] {
		m.konamiIndex++
		if m.konamiIndex >= len(m.konamiCode) {
			m.konamiIndex = 0
			return true
		}
	} else {
		m.konamiIndex = 0
	}
	return false
}

func (m *AnimatedListModel) createRipple(x, y int) {
	m.ripples = append(m.ripples, Ripple{
		x:         x,
		y:         y,
		radius:    0,
		maxRadius: 20,
		color:     lipgloss.Color(fmt.Sprintf("%d", 50+m.animationTick%200)),
	})
}

func (m *AnimatedListModel) createExplosion() {
	centerX := m.width / 2
	centerY := m.height / 2
	
	colors := []string{"196", "202", "208", "214", "220", "226"}
	for i, color := range colors {
		m.ripples = append(m.ripples, Ripple{
			x:         centerX,
			y:         centerY,
			radius:    float64(i * 2),
			maxRadius: 30,
			color:     lipgloss.Color(color),
		})
	}
}

func (m *AnimatedListModel) createCelebration() {
	for i := 0; i < 5; i++ {
		x := (i + 1) * m.width / 6
		y := m.height / 2
		m.createRipple(x, y)
	}
}

func (m *AnimatedListModel) updateAnimations() {
	// Update ripples
	var activeRipples []Ripple
	for _, r := range m.ripples {
		r.radius += 0.5
		if r.radius < r.maxRadius {
			activeRipples = append(activeRipples, r)
		}
	}
	m.ripples = activeRipples
	
	// Update glow
	if m.glowIntensity > 0 {
		m.glowIntensity -= 0.02
	}
	
	// Update pulse
	m.pulsePhase += 0.1
}

func (m *AnimatedListModel) updateSmoothScroll() {
	diff := m.targetOffset - m.scrollOffset
	m.velocity = m.velocity*0.8 + diff*0.2
	m.scrollOffset += m.velocity
	
	if math.Abs(m.velocity) < 0.01 {
		m.scrollOffset = m.targetOffset
		m.velocity = 0
	}
}

func (m *AnimatedListModel) applySecretEffects(view string) string {
	lines := strings.Split(view, "\n")
	
	for i := range lines {
		if i%2 == m.animationTick%2 {
			// Apply rainbow effect to alternating lines
			lines[i] = m.rainbowLine(lines[i])
		}
	}
	
	return strings.Join(lines, "\n")
}

func (m *AnimatedListModel) applyGlowEffect(view string) string {
	intensity := int(m.glowIntensity * 255)
	glowColor := fmt.Sprintf("#%02x%02xff", intensity, intensity)
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(glowColor)).
		Render(view)
}

func (m *AnimatedListModel) overlayRipples(base string) string {
	lines := strings.Split(base, "\n")
	
	for _, ripple := range m.ripples {
		for y := -int(ripple.radius); y <= int(ripple.radius); y++ {
			for x := -int(ripple.radius); x <= int(ripple.radius); x++ {
				dist := math.Sqrt(float64(x*x + y*y))
				if math.Abs(dist-ripple.radius) < 1.0 {
					lineY := ripple.y + y
					lineX := ripple.x + x
					
					if lineY >= 0 && lineY < len(lines) && lineX >= 0 {
						line := []rune(lines[lineY])
						if lineX < len(line) {
							opacity := 1.0 - (ripple.radius / ripple.maxRadius)
							if opacity > 0.3 {
								line[lineX] = '○'
							}
						}
						lines[lineY] = string(line)
					}
				}
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

func (m *AnimatedListModel) rainbowLine(line string) string {
	colors := []string{"196", "202", "208", "214", "220", "226", "190", "154", "118", "82"}
	
	var result strings.Builder
	for i, ch := range line {
		colorIdx := (i + m.animationTick/2) % len(colors)
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors[colorIdx])).
			Render(string(ch)))
	}
	
	return result.String()
}

func animationTick() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}