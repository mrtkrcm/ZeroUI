package extractor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Extractor is the main configuration extraction orchestrator
type Extractor struct {
	strategies []Strategy
	cache      Cache
	pool       chan struct{}
	timeout    time.Duration
	client     HTTPClient
}

// Config represents extracted configuration
type Config struct {
	App      string             `yaml:"app_name"`
	Path     string             `yaml:"config_path"`
	Type     string             `yaml:"config_type"`
	Settings map[string]Setting `yaml:"settings"`
}

// Setting represents a configuration option
type Setting struct {
	Name    string      `yaml:"name"`
	Type    string      `yaml:"type"`
	Default interface{} `yaml:"default_value,omitempty"`
	Values  []string    `yaml:"valid_values,omitempty"`
	Desc    string      `yaml:"description,omitempty"`
	Cat     string      `yaml:"category,omitempty"`
}

// HTTPClient interface for HTTP operations
type HTTPClient interface {
	Get(ctx context.Context, url string) (io.ReadCloser, error)
}

// DefaultHTTPClient implements HTTPClient using standard library
type DefaultHTTPClient struct {
	client *http.Client
}

func (c *DefaultHTTPClient) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// Option configures the extractor
type Option func(*Extractor)

// WithCache sets a custom cache
func WithCache(cache Cache) Option {
	return func(e *Extractor) {
		e.cache = cache
	}
}

// WithTimeout sets extraction timeout
func WithTimeout(timeout time.Duration) Option {
	return func(e *Extractor) {
		e.timeout = timeout
	}
}

// WithConcurrency sets max concurrent extractions
func WithConcurrency(n int) Option {
	return func(e *Extractor) {
		e.pool = make(chan struct{}, n)
	}
}

// WithStrategy adds a custom strategy
func WithStrategy(strategy Strategy) Option {
	return func(e *Extractor) {
		e.strategies = append(e.strategies, strategy)
	}
}

// New creates a new extractor with default configuration
func New(opts ...Option) *Extractor {
	// Optimized HTTP client with better connection pooling and timeouts
	transport := &http.Transport{
		MaxIdleConns:          100,              // Increased total idle connections
		MaxIdleConnsPerHost:   20,               // Increased per-host idle connections
		MaxConnsPerHost:       50,               // Limit concurrent connections per host
		IdleConnTimeout:       90 * time.Second, // Keep connections alive longer
		TLSHandshakeTimeout:   10 * time.Second, // Reasonable TLS timeout
		ExpectContinueTimeout: 1 * time.Second,  // Faster expect-continue
		DisableCompression:    false,            // Enable compression for better bandwidth
		ForceAttemptHTTP2:     true,             // Prefer HTTP/2 when available
	}

	httpClient := &DefaultHTTPClient{
		client: &http.Client{
			Timeout:   30 * time.Second, // Increased timeout for large configs
			Transport: transport,
		},
	}

	e := &Extractor{
		cache:   NewLRUCache(100, 24*time.Hour),
		pool:    make(chan struct{}, 16), // Increased worker pool for better parallelism
		timeout: 30 * time.Second,
		client:  httpClient,
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Initialize default strategies if none provided
	if len(e.strategies) == 0 {
		e.strategies = []Strategy{
			NewCLIStrategy(),
			NewLocalStrategy("configs"),
			NewBuiltinStrategy(),
			NewGitHubStrategy(httpClient),
		}
	}

	return e
}

// Extract gets configuration for a single app
func (e *Extractor) Extract(ctx context.Context, app string) (*Config, error) {
	// Check cache first
	if cached, ok := e.cache.Get(app); ok {
		return cached, nil
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Try strategies in parallel
	type result struct {
		config     *Config
		confidence float64
		err        error
	}

	results := make(chan result, len(e.strategies))

	for _, strategy := range e.strategies {
		go func(s Strategy) {
			cfg, err := s.Extract(ctx, app)
			results <- result{
				config:     cfg,
				confidence: s.Confidence(),
				err:        err,
			}
		}(strategy)
	}

	// Collect results and return best one
	var bestResult result
	var bestConfidence float64

	for i := 0; i < len(e.strategies); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-results:
			if r.err == nil && r.config != nil && r.confidence > bestConfidence {
				bestResult = r
				bestConfidence = r.confidence
			}
		}
	}

	if bestResult.config != nil {
		e.cache.Set(app, bestResult.config)
		return bestResult.config, nil
	}

	return nil, fmt.Errorf("no extraction method succeeded for %s", app)
}

// ExtractBatch extracts configurations for multiple apps concurrently
func (e *Extractor) ExtractBatch(ctx context.Context, apps []string) (map[string]*Config, error) {
	results := make(map[string]*Config)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, app := range apps {
		wg.Add(1)

		// Acquire worker slot
		e.pool <- struct{}{}

		go func(appName string) {
			defer func() {
				<-e.pool // Release worker slot
				wg.Done()
			}()

			config, err := e.Extract(ctx, appName)
			if err == nil && config != nil {
				mu.Lock()
				results[appName] = config
				mu.Unlock()
			}
		}(app)
	}

	wg.Wait()
	return results, nil
}

// AddStrategy adds a new extraction strategy
func (e *Extractor) AddStrategy(strategy Strategy) {
	e.strategies = append(e.strategies, strategy)
}

// ClearCache clears the extraction cache
func (e *Extractor) ClearCache() {
	e.cache.Clear()
}

// Helper functions

func inferType(value string) string {
	value = strings.TrimSpace(value)

	switch {
	case value == "true" || value == "false":
		return "boolean"
	case isNumeric(value):
		return "number"
	case strings.HasPrefix(value, "#") || strings.HasPrefix(value, "0x"):
		return "color"
	case strings.Contains(value, ","):
		return "array"
	default:
		return "string"
	}
}

func inferCategory(key string) string {
	key = strings.ToLower(key)

	categories := map[string][]string{
		"font":        {"font", "text"},
		"appearance":  {"color", "theme", "background", "foreground"},
		"window":      {"window", "pane", "split"},
		"keybindings": {"key", "bind", "map", "shortcut"},
		"editor":      {"editor", "cursor", "scroll", "indent"},
		"git":         {"git", "diff", "merge"},
		"terminal":    {"term", "shell", "prompt"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(key, keyword) {
				return category
			}
		}
	}

	return "general"
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	// Handle negative numbers
	if s[0] == '-' {
		s = s[1:]
	}

	dotCount := 0
	for _, r := range s {
		if r == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
