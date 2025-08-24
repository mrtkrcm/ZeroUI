package appcomponents

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// ScannerState represents the scanner's current state
type ScannerState int

const (
	ScannerIdle ScannerState = iota
	ScannerScanning
	ScannerComplete
	ScannerError
)

// AppScannerV2 is an improved, cleaner app scanner
type AppScannerV2 struct {
	// State
	state      ScannerState
	apps       []AppInfo
	errors     []error
	
	// Progress tracking
	current    int
	total      int
	startTime  time.Time
	
	// UI components
	spinner    spinner.Model
	progress   progress.Model
	
	// Dimensions
	width      int
	height     int
	
	// Configuration
	registry   *config.AppsRegistry
	
	// Thread safety
	mu         sync.RWMutex
}

// NewAppScannerV2 creates an improved scanner
func NewAppScannerV2() *AppScannerV2 {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	
	return &AppScannerV2{
		state:    ScannerIdle,
		spinner:  s,
		progress: p,
		apps:     []AppInfo{},
		errors:   []error{},
	}
}

// Init starts the scanner
func (s *AppScannerV2) Init() tea.Cmd {
	s.state = ScannerScanning
	s.startTime = time.Now()
	
	return tea.Batch(
		s.spinner.Tick,
		s.scan(),
	)
}

// Update handles messages
func (s *AppScannerV2) Update(msg tea.Msg) (*AppScannerV2, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.progress.Width = min(msg.Width-4, 60)
		return s, nil
		
	case spinner.TickMsg:
		if s.state == ScannerScanning {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}
		return s, nil
		
	case progress.FrameMsg:
		if s.state == ScannerScanning {
			progressModel, cmd := s.progress.Update(msg)
			s.progress = progressModel.(progress.Model)
			return s, cmd
		}
		return s, nil
		
	case scanTickMsg:
		// Update progress
		s.current = msg.current
		s.total = msg.total
		if s.total > 0 {
			percent := float64(s.current) / float64(s.total)
			return s, s.progress.SetPercent(percent)
		}
		return s, nil
		
	case ScanCompleteMsg:
		s.state = ScannerComplete
		s.apps = msg.Apps
		return s, nil
		
	case scanErrorMsg:
		s.state = ScannerError
		s.errors = append(s.errors, msg.error)
		return s, nil
	}
	
	return s, nil
}

// View renders the scanner
func (s *AppScannerV2) View() string {
	switch s.state {
	case ScannerScanning:
		return s.viewScanning()
	case ScannerComplete:
		return s.viewComplete()
	case ScannerError:
		return s.viewError()
	default:
		return s.viewIdle()
	}
}

// viewScanning shows scanning progress
func (s *AppScannerV2) viewScanning() string {
	var b strings.Builder
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))
	
	b.WriteString(titleStyle.Render("Scanning Applications"))
	b.WriteString("\n\n")
	
	// Spinner and status
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	status := fmt.Sprintf("%s Checking %d/%d applications...", 
		s.spinner.View(),
		s.current,
		s.total,
	)
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n\n")
	
	// Progress bar
	b.WriteString(s.progress.View())
	b.WriteString("\n\n")
	
	// Elapsed time
	elapsed := time.Since(s.startTime).Round(time.Second)
	b.WriteString(statusStyle.Render(fmt.Sprintf("Time: %s", elapsed)))
	
	return b.String()
}

// viewComplete shows scan results
func (s *AppScannerV2) viewComplete() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))
	b.WriteString(titleStyle.Render("Applications"))
	b.WriteString("\n\n")
	
	// Group apps by status
	ready := []AppInfo{}
	notConfigured := []AppInfo{}
	
	for _, app := range s.apps {
		if app.ConfigExists {
			ready = append(ready, app)
		} else {
			notConfigured = append(notConfigured, app)
		}
	}
	
	// Style definitions
	readyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)
	notConfiguredStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	// Show ready apps
	if len(ready) > 0 {
		b.WriteString(readyStyle.Render("● Configured"))
		b.WriteString(fmt.Sprintf(" %s\n", dimStyle.Render(fmt.Sprintf("(%d)", len(ready)))))
		
		for _, app := range ready {
			b.WriteString(fmt.Sprintf("  %s %s\n", 
				app.Icon, 
				app.Name,
			))
		}
		b.WriteString("\n")
	}
	
	// Show not configured apps
	if len(notConfigured) > 0 {
		b.WriteString(notConfiguredStyle.Render("○ Not Configured"))
		b.WriteString(fmt.Sprintf(" %s\n", dimStyle.Render(fmt.Sprintf("(%d)", len(notConfigured)))))
		
		for _, app := range notConfigured {
			b.WriteString(fmt.Sprintf("  %s %s\n",
				app.Icon,
				dimStyle.Render(app.Name),
			))
		}
	}
	
	// Summary
	b.WriteString("\n")
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)
	summary := fmt.Sprintf("Found %d apps, %d configured", 
		len(s.apps), 
		len(ready),
	)
	b.WriteString(summaryStyle.Render(summary))
	
	return b.String()
}

// viewError shows error state
func (s *AppScannerV2) viewError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	
	var b strings.Builder
	b.WriteString(errorStyle.Render("✗ Scan Failed"))
	b.WriteString("\n\n")
	
	for _, err := range s.errors {
		b.WriteString(fmt.Sprintf("  • %v\n", err))
	}
	
	return b.String()
}

// viewIdle shows idle state
func (s *AppScannerV2) viewIdle() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Render("Ready to scan...")
}

// scan performs the actual scanning
func (s *AppScannerV2) scan() tea.Cmd {
	return func() tea.Msg {
		// Load registry
		registry, err := config.LoadAppsRegistry()
		if err != nil {
			return scanErrorMsg{error: fmt.Errorf("failed to load registry: %w", err)}
		}
		
		s.registry = registry
		apps := registry.GetAllApps()
		s.total = len(apps)
		
		results := make([]AppInfo, 0, len(apps))
		
		for i, app := range apps {
			// Send progress update
			s.current = i + 1
			
			// Check config
			exists, path := registry.CheckAppStatus(app.Name)
			
			info := AppInfo{
				Name:         app.Name,
				Icon:         app.Icon,
				Status:       StatusNotConfigured,
				ConfigPath:   path,
				ConfigExists: exists,
			}
			
			if exists {
				info.Status = StatusReady
			}
			
			results = append(results, info)
			
			// Small delay for UI smoothness
			time.Sleep(30 * time.Millisecond)
		}
		
		return ScanCompleteMsg{Apps: results}
	}
}

// GetApps returns scanned applications
func (s *AppScannerV2) GetApps() []AppInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apps
}

// GetState returns current state
func (s *AppScannerV2) GetState() ScannerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// IsScanning checks if scanning is in progress
func (s *AppScannerV2) IsScanning() bool {
	return s.GetState() == ScannerScanning
}

// IsComplete checks if scanning is complete
func (s *AppScannerV2) IsComplete() bool {
	return s.GetState() == ScannerComplete
}

// Reset resets the scanner for a new scan
func (s *AppScannerV2) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.state = ScannerIdle
	s.apps = []AppInfo{}
	s.errors = []error{}
	s.current = 0
	s.total = 0
}

// Messages

type scanTickMsg struct {
	current int
	total   int
}

type scanErrorMsg struct {
	error error
}

// Helper functions are defined in other components