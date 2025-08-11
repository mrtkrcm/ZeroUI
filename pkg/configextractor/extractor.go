package configextractor

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/performance"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Extractor is the main configuration extractor
type Extractor struct {
	strategies []Strategy
	parsers    map[string]Parser
	cache      Cache
	registry   AppRegistry

	// Performance optimizations
	timeout          time.Duration
	concurrency      int
	concurrentLoader *performance.ConcurrentConfigLoader
}

// New creates a new extractor with sensible defaults
func New(opts ...Option) *Extractor {
	e := &Extractor{
		strategies:       make([]Strategy, 0),
		parsers:          make(map[string]Parser),
		cache:            NewLRUCache(100, 24*time.Hour), // 100 entries, 24h TTL
		timeout:          30 * time.Second,
		concurrency:      8,
		concurrentLoader: performance.NewConcurrentLoader(8), // Match concurrency
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Register default strategies (sorted by priority)
	e.registerDefaultStrategies()

	// Register default parsers
	e.registerDefaultParsers()

	// Register default app definitions
	e.registerDefaultApps()

	return e
}

// Extract gets configuration for a single app
func (e *Extractor) Extract(ctx context.Context, app string) (*Config, error) {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Check cache first
	cacheKey := fmt.Sprintf("config:%s", app)
	if cached, ok := e.cache.Get(cacheKey); ok {
		return cached, nil
	}

	// Get applicable strategies for this app
	strategies := e.getStrategiesForApp(app)
	if len(strategies) == 0 {
		return nil, fmt.Errorf("no extraction strategies available for app: %s", app)
	}

	// Try strategies in parallel using worker pool to prevent goroutine leaks
	type result struct {
		config *Config
		err    error
		source string
	}

	resultChan := make(chan result, len(strategies))
	workerPool := make(chan struct{}, min(len(strategies), 5)) // Limit concurrent goroutines
	var wg sync.WaitGroup

	// Launch extraction attempts with worker pool
	for _, strategy := range strategies {
		wg.Add(1)
		go func(s Strategy) {
			defer wg.Done()

			// Acquire worker slot
			workerPool <- struct{}{}
			defer func() { <-workerPool }()

			// Check context before expensive operation
			select {
			case <-ctx.Done():
				resultChan <- result{nil, ctx.Err(), s.Name()}
				return
			default:
			}

			config, err := s.Extract(ctx, app)
			resultChan <- result{config, err, s.Name()}
		}(strategy)
	}

	// Close result channel when all workers complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results, return best one
	var bestConfig *Config
	var lastErr error

	// Process results as they arrive
	for res := range resultChan {
		if res.err == nil && res.config != nil {
			// Success! Cache and return
			e.cache.Set(cacheKey, res.config)
			return res.config, nil
		}
		lastErr = res.err

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	if bestConfig != nil {
		e.cache.Set(cacheKey, bestConfig)
		return bestConfig, nil
	}

	return nil, fmt.Errorf("all extraction strategies failed for %s: %w", app, lastErr)
}

// ExtractBatch processes multiple apps concurrently
func (e *Extractor) ExtractBatch(ctx context.Context, apps []string) (map[string]*Config, error) {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	results := make(map[string]*Config)
	resultsMu := sync.Mutex{}
	errors := make(map[string]error)
	errorsMu := sync.Mutex{}

	// Use worker pool for concurrency control
	work := make(chan string, len(apps))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < e.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for app := range work {
				config, err := e.Extract(ctx, app)
				if err != nil {
					errorsMu.Lock()
					errors[app] = err
					errorsMu.Unlock()
				} else {
					resultsMu.Lock()
					results[app] = config
					resultsMu.Unlock()
				}
			}
		}()
	}

	// Send work
	for _, app := range apps {
		work <- app
	}
	close(work)

	// Wait for completion
	wg.Wait()

	// Return results (partial success is OK)
	return results, nil
}

// SupportedApps returns list of supported applications
func (e *Extractor) SupportedApps() []string {
	if e.registry != nil {
		return e.registry.ListApps()
	}

	// Fallback: collect from strategies
	appSet := make(map[string]bool)
	for range e.strategies {
		// This would need to be enhanced based on strategy implementation
		// For now, return common apps
	}

	apps := []string{"ghostty", "zed", "alacritty", "wezterm", "tmux", "git", "neovim"}
	result := make([]string, 0, len(apps))
	for app := range appSet {
		result = append(result, app)
	}
	if len(result) == 0 {
		return apps // fallback
	}

	sort.Strings(result)
	return result
}

// getStrategiesForApp returns strategies that can handle the app, sorted by priority
func (e *Extractor) getStrategiesForApp(app string) []Strategy {
	var applicable []Strategy

	for _, strategy := range e.strategies {
		if strategy.CanExtract(app) {
			applicable = append(applicable, strategy)
		}
	}

	// Sort by priority (higher first)
	sort.Slice(applicable, func(i, j int) bool {
		return applicable[i].Priority() > applicable[j].Priority()
	})

	return applicable
}

// registerDefaultStrategies adds built-in extraction strategies
func (e *Extractor) registerDefaultStrategies() {
	// Order matters - higher priority strategies first
	e.strategies = []Strategy{
		NewCLI(), // Fastest, most reliable
		// NewLocal(),    // Fast, cached locally
		// NewBuiltin(),  // Always available fallback
		NewGitHub(), // Network dependent, slower
	}
}

// registerDefaultParsers adds built-in format parsers
func (e *Extractor) registerDefaultParsers() {
	// Register parsers by format
	// Implementation would be added here
}

// registerDefaultApps adds built-in app definitions
func (e *Extractor) registerDefaultApps() {
	if e.registry == nil {
		return
	}

	// Register common applications
	apps := []*AppDef{
		{
			Name:       "ghostty",
			ConfigPath: "~/.config/ghostty/config",
			Format:     "custom",
			Strategies: []string{"cli", "local", "builtin"},
			CLICommand: "ghostty +show-config --default --docs",
		},
		{
			Name:       "zed",
			Aliases:    []string{"zed-editor"},
			ConfigPath: "~/.config/zed/settings.json",
			Format:     "json",
			Strategies: []string{"github", "local", "cli"},
			CLICommand: "zed --print-config",
			GitHubRepo: "zed-industries/zed",
		},
		{
			Name:       "alacritty",
			ConfigPath: "~/.config/alacritty/alacritty.yml",
			Format:     "yaml",
			Strategies: []string{"github", "local", "builtin"},
			GitHubRepo: "alacritty/alacritty",
		},
		{
			Name:       "wezterm",
			ConfigPath: "~/.config/wezterm/wezterm.lua",
			Format:     "lua",
			Strategies: []string{"cli", "github", "local"},
			CLICommand: "wezterm show-config",
			GitHubRepo: "wez/wezterm",
		},
	}

	for _, app := range apps {
		e.registry.Register(app)
	}
}

// Option provides functional options for extractor configuration
type Option func(*Extractor)

// WithTimeout sets extraction timeout
func WithTimeout(timeout time.Duration) Option {
	return func(e *Extractor) {
		e.timeout = timeout
	}
}

// WithConcurrency sets max concurrent extractions
func WithConcurrency(n int) Option {
	return func(e *Extractor) {
		if n > 0 {
			e.concurrency = n
		}
	}
}

// WithCache sets custom cache implementation
func WithCache(cache Cache) Option {
	return func(e *Extractor) {
		e.cache = cache
	}
}

// WithRegistry sets custom app registry
func WithRegistry(registry AppRegistry) Option {
	return func(e *Extractor) {
		e.registry = registry
	}
}

// WithStrategy adds a custom extraction strategy
func WithStrategy(strategy Strategy) Option {
	return func(e *Extractor) {
		e.strategies = append(e.strategies, strategy)
		// Re-sort by priority
		sort.Slice(e.strategies, func(i, j int) bool {
			return e.strategies[i].Priority() > e.strategies[j].Priority()
		})
	}
}
