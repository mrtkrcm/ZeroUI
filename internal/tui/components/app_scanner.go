package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// AppStatus represents the status of an application
type AppStatus int

const (
	StatusUnknown AppStatus = iota
	StatusScanning
	StatusReady
	StatusNotConfigured
	StatusError
)

// AppInfo represents information about a scanned application
type AppInfo struct {
	Name         string
	Status       AppStatus
	ConfigPath   string
	ConfigExists bool
	Error        error
	Icon         string
}

// ScanProgressMsg is sent during scanning
type ScanProgressMsg struct {
	App      string
	Progress float64
}

// ScanCompleteMsg is sent when scanning is complete
type ScanCompleteMsg struct {
	Apps []AppInfo
}

// AppScanner handles application scanning and status checking
type AppScanner struct {
	apps        []AppInfo
	scanning    bool
	currentApp  int
	totalApps   int
	progress    progress.Model
	spinner     spinner.Model
	width       int
	height      int
	startTime   time.Time
	
	// Styles
	titleStyle    lipgloss.Style
	statusStyle   lipgloss.Style
	readyStyle    lipgloss.Style
	notFoundStyle lipgloss.Style
	errorStyle    lipgloss.Style
}

// NewAppScanner creates a new application scanner
func NewAppScanner() *AppScanner {
	p := progress.New(progress.WithDefaultGradient())
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return &AppScanner{
		apps:     []AppInfo{},
		progress: p,
		spinner:  s,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		readyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),
		notFoundStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
	}
}

// Init initializes the scanner
func (s *AppScanner) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		s.startScan(),
	)
}

// Update handles messages
func (s *AppScanner) Update(msg tea.Msg) (*AppScanner, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.progress.Width = msg.Width - 4
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
		
	case progress.FrameMsg:
		progressModel, cmd := s.progress.Update(msg)
		s.progress = progressModel.(progress.Model)
		return s, cmd
		
	case ScanProgressMsg:
		// Update progress
		cmds := []tea.Cmd{}
		if s.scanning {
			cmds = append(cmds, s.progress.SetPercent(msg.Progress))
		}
		return s, tea.Batch(cmds...)
		
	case ScanCompleteMsg:
		s.scanning = false
		s.apps = msg.Apps
		return s, nil
	}
	
	return s, nil
}

// View renders the scanner view
func (s *AppScanner) View() string {
	if s.scanning {
		return s.renderScanning()
	}
	return s.renderResults()
}

// renderScanning shows the scanning progress
func (s *AppScanner) renderScanning() string {
	var b strings.Builder
	
	title := s.titleStyle.Render("🔍 Scanning Applications")
	b.WriteString(title + "\n\n")
	
	// Show spinner and current app
	if s.currentApp < s.totalApps {
		status := fmt.Sprintf("%s Checking app %d of %d", 
			s.spinner.View(), 
			s.currentApp+1, 
			s.totalApps)
		b.WriteString(s.statusStyle.Render(status) + "\n\n")
	}
	
	// Show progress bar
	b.WriteString(s.progress.View() + "\n\n")
	
	// Show elapsed time
	elapsed := time.Since(s.startTime).Round(time.Second)
	b.WriteString(s.statusStyle.Render(fmt.Sprintf("Elapsed: %s", elapsed)))
	
	return b.String()
}

// renderResults shows the scan results
func (s *AppScanner) renderResults() string {
	var b strings.Builder
	
	title := s.titleStyle.Render("📦 Applications")
	b.WriteString(title + "\n\n")
	
	// Group apps by status
	ready := []AppInfo{}
	notConfigured := []AppInfo{}
	errors := []AppInfo{}
	
	for _, app := range s.apps {
		switch app.Status {
		case StatusReady:
			ready = append(ready, app)
		case StatusNotConfigured:
			notConfigured = append(notConfigured, app)
		case StatusError:
			errors = append(errors, app)
		}
	}
	
	// Show ready apps
	if len(ready) > 0 {
		b.WriteString(s.readyStyle.Render("✓ Ready") + "\n")
		for _, app := range ready {
			b.WriteString(fmt.Sprintf("  %s %s\n", app.Icon, app.Name))
		}
		b.WriteString("\n")
	}
	
	// Show not configured apps
	if len(notConfigured) > 0 {
		b.WriteString(s.notFoundStyle.Render("⚠ Not Configured") + "\n")
		for _, app := range notConfigured {
			b.WriteString(fmt.Sprintf("  %s %s\n", app.Icon, app.Name))
		}
		b.WriteString("\n")
	}
	
	// Show errors
	if len(errors) > 0 {
		b.WriteString(s.errorStyle.Render("✗ Errors") + "\n")
		for _, app := range errors {
			b.WriteString(fmt.Sprintf("  %s %s: %v\n", app.Icon, app.Name, app.Error))
		}
	}
	
	return b.String()
}

// startScan initiates the scanning process
func (s *AppScanner) startScan() tea.Cmd {
	s.scanning = true
	s.startTime = time.Now()
	
	// Return the scan command
	return s.performScan()
}

// performScan does the actual scanning work - returns a tea.Cmd
func (s *AppScanner) performScan() tea.Cmd {
	return func() tea.Msg {
		// Load apps registry
		registry, err := loadAppsRegistry()
		if err != nil {
			// Fall back to a minimal set if registry fails
			return s.performFallbackScanMsg()
		}
		
		knownApps := registry.GetAllApps()
		s.totalApps = len(knownApps)
		results := []AppInfo{}
		
		for i, app := range knownApps {
			s.currentApp = i
			
			// Progress will be shown via the spinner/progress bar
			progress := float64(i) / float64(s.totalApps)
			_ = progress // We'll handle progress differently
			
			// Check for config file
			info := AppInfo{
				Name:   app.Name,
				Icon:   app.Icon,
				Status: StatusNotConfigured,
			}
			
			// Check if config exists
			configExists, configPath := registry.CheckAppStatus(app.Name)
			if configExists {
				info.Status = StatusReady
				info.ConfigPath = configPath
				info.ConfigExists = true
			}
			
			results = append(results, info)
			
			// Small delay for smooth UI updates
			time.Sleep(50 * time.Millisecond)
		}
		
		// Return completion message
		return ScanCompleteMsg{
			Apps: results,
		}
	}
}

// performFallbackScanMsg performs a minimal scan if registry fails
func (s *AppScanner) performFallbackScanMsg() tea.Msg {
	// Minimal fallback set
	knownApps := []struct {
		name string
		icon string
		configPaths []string
	}{
		{
			name: "zeroui",
			icon: "○",
			configPaths: []string{
				"~/.config/zeroui/config.yml",
				"~/.config/zeroui/config.yaml",
			},
		},
		{
			name: "ghostty",
			icon: "◉",
			configPaths: []string{
				"~/.config/ghostty/config",
				"~/.ghostty/config",
			},
		},
	}
	
	s.totalApps = len(knownApps)
	results := []AppInfo{}
	home, _ := os.UserHomeDir()
	
	for i, app := range knownApps {
		s.currentApp = i
		
		// Check for config file
		info := AppInfo{
			Name:   app.name,
			Icon:   app.icon,
			Status: StatusNotConfigured,
		}
		
		for _, path := range app.configPaths {
			expandedPath := strings.ReplaceAll(path, "~", home)
			if _, err := os.Stat(expandedPath); err == nil {
				info.Status = StatusReady
				info.ConfigPath = expandedPath
				info.ConfigExists = true
				break
			}
		}
		
		results = append(results, info)
		time.Sleep(50 * time.Millisecond)
	}
	
	// Return completion message
	return ScanCompleteMsg{
		Apps: results,
	}
}

// loadAppsRegistry loads the apps registry
func loadAppsRegistry() (*config.AppsRegistry, error) {
	return config.LoadAppsRegistry()
}

// GetApps returns the scanned applications
func (s *AppScanner) GetApps() []AppInfo {
	return s.apps
}

// IsScanning returns whether scanning is in progress
func (s *AppScanner) IsScanning() bool {
	return s.scanning
}

// FindApp finds an app by name
func (s *AppScanner) FindApp(name string) (*AppInfo, bool) {
	for _, app := range s.apps {
		if app.Name == name {
			return &app, true
		}
	}
	return nil, false
}

// RescanApp rescans a specific application
func (s *AppScanner) RescanApp(name string) tea.Cmd {
	return func() tea.Msg {
		// Check config for specific app
		home, _ := os.UserHomeDir()
		
		// Update the app info
		for i, app := range s.apps {
			if app.Name == name {
				// Check common config paths
				configPaths := getConfigPaths(name, home)
				for _, path := range configPaths {
					if _, err := os.Stat(path); err == nil {
						s.apps[i].Status = StatusReady
						s.apps[i].ConfigPath = path
						s.apps[i].ConfigExists = true
						break
					}
				}
			}
		}
		
		return ScanCompleteMsg{Apps: s.apps}
	}
}

// getConfigPaths returns possible config paths for an app
func getConfigPaths(appName, home string) []string {
	switch appName {
	case "ghostty":
		return []string{
			filepath.Join(home, ".config", "ghostty", "config"),
			filepath.Join(home, ".ghostty", "config"),
		}
	case "alacritty":
		return []string{
			filepath.Join(home, ".config", "alacritty", "alacritty.yml"),
			filepath.Join(home, ".config", "alacritty", "alacritty.yaml"),
			filepath.Join(home, ".config", "alacritty", "alacritty.toml"),
		}
	case "kitty":
		return []string{
			filepath.Join(home, ".config", "kitty", "kitty.conf"),
		}
	case "wezterm":
		return []string{
			filepath.Join(home, ".config", "wezterm", "wezterm.lua"),
			filepath.Join(home, ".wezterm.lua"),
		}
	case "zed":
		return []string{
			filepath.Join(home, ".config", "zed", "settings.json"),
		}
	case "neovim":
		return []string{
			filepath.Join(home, ".config", "nvim", "init.lua"),
			filepath.Join(home, ".config", "nvim", "init.vim"),
		}
	case "vscode":
		return []string{
			filepath.Join(home, ".config", "Code", "User", "settings.json"),
			filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"),
		}
	default:
		return []string{}
	}
}