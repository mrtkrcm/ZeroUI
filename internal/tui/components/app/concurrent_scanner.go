package app

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// ConcurrentScanner performs parallel application scanning
type ConcurrentScanner struct {
	registry *config.AppsRegistry
	results  chan ScanResult
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// ScanResult represents a single scan result
type ScanResult struct {
	App          config.AppDefinition
	ConfigExists bool
	ConfigPath   string
	Error        error
}

// NewConcurrentScanner creates a new concurrent scanner
func NewConcurrentScanner() *ConcurrentScanner {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &ConcurrentScanner{
		results: make(chan ScanResult, 20),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// ScanAll performs concurrent scanning of all apps
func (cs *ConcurrentScanner) ScanAll() tea.Cmd {
	return func() tea.Msg {
		defer cs.cancel()

		// Load registry
		registry, err := config.LoadAppsRegistry()
		if err != nil {
			return scanErrorMsg{error: err}
		}

		cs.registry = registry
		apps := registry.GetAllApps()

		// Start workers
		numWorkers := 5
		workChan := make(chan config.AppDefinition, len(apps))

		// Start worker goroutines
		for i := 0; i < numWorkers; i++ {
			cs.wg.Add(1)
			go cs.worker(workChan)
		}

		// Send work to workers
		for _, app := range apps {
			select {
			case workChan <- app:
			case <-cs.ctx.Done():
				close(workChan)
				return scanErrorMsg{error: cs.ctx.Err()}
			}
		}
		close(workChan)

		// Wait for completion in a separate goroutine
		go func() {
			cs.wg.Wait()
			close(cs.results)
		}()

		// Collect results
		var scanResults []AppInfo
		for result := range cs.results {
			info := AppInfo{
				Name:         result.App.Name,
				Icon:         result.App.Icon,
				Status:       StatusNotConfigured,
				ConfigPath:   result.ConfigPath,
				ConfigExists: result.ConfigExists,
			}

			if result.ConfigExists {
				info.Status = StatusReady
			}

			if result.Error != nil {
				info.Status = StatusError
				info.Error = result.Error
			}

			scanResults = append(scanResults, info)
		}

		return ScanCompleteMsg{Apps: scanResults}
	}
}

// worker processes apps from the work channel
func (cs *ConcurrentScanner) worker(workChan <-chan config.AppDefinition) {
	defer cs.wg.Done()

	for app := range workChan {
		select {
		case <-cs.ctx.Done():
			return
		default:
			result := cs.checkApp(app)

			select {
			case cs.results <- result:
			case <-cs.ctx.Done():
				return
			}
		}
	}
}

// checkApp checks a single app's configuration
func (cs *ConcurrentScanner) checkApp(app config.AppDefinition) ScanResult {
	result := ScanResult{
		App: app,
	}

	// Check if config exists
	exists, path := cs.registry.CheckAppStatus(app.Name)
	result.ConfigExists = exists
	result.ConfigPath = path

	return result
}

// Stop cancels the scanning operation
func (cs *ConcurrentScanner) Stop() {
	if cs.cancel != nil {
		cs.cancel()
	}
}
