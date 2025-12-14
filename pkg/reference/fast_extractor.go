package reference

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/performance"
)

// Optimized string operations to reduce allocations

// findEqualsByte finds the index of '=' character using byte operations
func findEqualsByte(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return i
		}
	}
	return -1
}

// trimSpaceFast performs fast space trimming without allocations for common cases
func trimSpaceFast(s string) string {
	// Fast path for empty strings
	if len(s) == 0 {
		return s
	}

	// Find start
	start := 0
	for start < len(s) && isSpace(s[start]) {
		start++
	}

	// Find end
	end := len(s)
	for end > start && isSpace(s[end-1]) {
		end--
	}

	// Return substring if trimming is needed
	if start > 0 || end < len(s) {
		return s[start:end]
	}
	return s
}

// isSpace checks if a byte is a space character (optimized for common cases)
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// FastExtractor provides high-performance config extraction
type FastExtractor struct {
	// Performance optimizations
	cache      *ExtractorCache
	httpClient *http.Client
	workerPool chan struct{} // Limit concurrent operations
	bufferPool sync.Pool     // Reuse buffers for parsing
}

// ExtractorCache provides thread-safe caching with TTL
type ExtractorCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	config    *ConfigReference
	timestamp time.Time
}

// NewFastExtractor creates an optimized extractor
func NewFastExtractor() *FastExtractor {
	return &FastExtractor{
		cache: &ExtractorCache{
			entries: make(map[string]*cacheEntry),
			ttl:     24 * time.Hour,
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		workerPool: make(chan struct{}, 8), // Max 8 concurrent operations
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// ExtractBatch extracts configs for multiple apps concurrently
func (e *FastExtractor) ExtractBatch(apps []string) (map[string]*ConfigReference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results := make(map[string]*ConfigReference)
	resultsMu := sync.Mutex{}

	var wg sync.WaitGroup
	errChan := make(chan error, len(apps))

	for _, app := range apps {
		wg.Add(1)
		go func(appName string) {
			defer wg.Done()

			// Acquire worker slot
			e.workerPool <- struct{}{}
			defer func() { <-e.workerPool }()

			config, err := e.ExtractWithContext(ctx, appName)
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", appName, err)
				return
			}

			resultsMu.Lock()
			results[appName] = config
			resultsMu.Unlock()
		}(app)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("extraction failed for %d apps", len(errs))
	}

	return results, nil
}

// ExtractWithContext extracts config with context support for cancellation
func (e *FastExtractor) ExtractWithContext(ctx context.Context, app string) (*ConfigReference, error) {
	// Check cache first
	if cached := e.cache.Get(app); cached != nil {
		return cached, nil
	}

	// Try extraction methods in parallel
	type result struct {
		config *ConfigReference
		err    error
		source string
	}

	resultChan := make(chan result, 3)

	// Launch parallel extraction attempts
	go func() {
		config, err := e.extractFromCLI(ctx, app)
		resultChan <- result{config, err, "CLI"}
	}()

	go func() {
		config, err := e.extractFromGitHub(ctx, app)
		resultChan <- result{config, err, "GitHub"}
	}()

	go func() {
		config, err := e.extractFromLocalFile(ctx, app)
		resultChan <- result{config, err, "LocalFile"}
	}()

	// Wait for first successful result
	var lastErr error
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-resultChan:
			if res.err == nil && res.config != nil {
				e.cache.Set(app, res.config)
				return res.config, nil
			}
			lastErr = res.err
		}
	}

	return nil, fmt.Errorf("all extraction methods failed: %w", lastErr)
}

// extractFromCLI uses optimized CLI extraction
func (e *FastExtractor) extractFromCLI(ctx context.Context, app string) (*ConfigReference, error) {
	commands := map[string]string{
		"ghostty":   "ghostty +show-config --default --docs",
		"alacritty": "alacritty --print-config",
		"wezterm":   "wezterm show-config",
		"tmux":      "tmux show-options -g",
		"git":       "git config --list --show-origin",
	}

	cmd, exists := commands[app]
	if !exists {
		return nil, fmt.Errorf("no CLI command for %s", app)
	}

	// Execute with context
	execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	output, err := execCmd.Output()
	if err != nil {
		return nil, err
	}

	// Use fast streaming parser
	return e.parseStreamingCLI(app, bytes.NewReader(output))
}

// parseStreamingCLI uses streaming parser for better memory efficiency
func (e *FastExtractor) parseStreamingCLI(app string, r io.Reader) (*ConfigReference, error) {
	config := &ConfigReference{
		AppName:    app,
		ConfigPath: fmt.Sprintf("~/.config/%s/config", app),
		ConfigType: "custom",
		Settings:   make(map[string]ConfigSetting),
	}

	scanner := bufio.NewScanner(r)
	// Use buffer pool to reduce allocations
	buf := e.bufferPool.Get().([]byte)
	defer func() {
		// Reset buffer before returning to pool
		buf = buf[:0]
		e.bufferPool.Put(&buf)
	}()

	scanner.Buffer(buf[:cap(buf)], 2*1024*1024) // Use pooled buffer, 2MB max for large configs

	var currentKey string
	descBuilder := performance.GetBuilder()
	defer performance.PutBuilder(descBuilder)

	for scanner.Scan() {
		line := scanner.Text()

		// Fast parsing with single-pass byte operations
		if idx := findEqualsByte(line); idx > 0 {
			// Found a setting
			if currentKey != "" && descBuilder.Len() > 0 {
				// Save previous setting
				config.Settings[currentKey] = ConfigSetting{
					Name:        currentKey,
					Type:        inferTypeQuick(config.Settings[currentKey].DefaultValue),
					Description: descBuilder.String(),
				}
				descBuilder.Reset()
			}

			// Optimized trimming with byte-level operations
			currentKey = trimSpaceFast(line[:idx])
			value := trimSpaceFast(line[idx+1:])

			config.Settings[currentKey] = ConfigSetting{
				Name:         currentKey,
				DefaultValue: value,
				Category:     inferCategoryQuick(currentKey),
			}

		} else if len(line) > 0 && (line[0] == '#' || (len(line) > 1 && line[0:2] == "//")) {
			// Description line - optimized prefix check
			if descBuilder.Len() > 0 {
				descBuilder.WriteByte(' ')
			}
			// Skip comment prefix and trim
			var content string
			if line[0] == '#' {
				content = trimSpaceFast(line[1:])
			} else {
				content = trimSpaceFast(line[2:])
			}
			descBuilder.WriteString(content)
		}
	}

	// Save last setting
	if currentKey != "" && descBuilder.Len() > 0 {
		setting := config.Settings[currentKey]
		setting.Description = descBuilder.String()
		config.Settings[currentKey] = setting
	}

	return config, scanner.Err()
}

// extractFromGitHub with connection pooling
func (e *FastExtractor) extractFromGitHub(ctx context.Context, app string) (*ConfigReference, error) {
	repos := map[string]struct {
		owner string
		repo  string
		path  string
	}{
		"zed": {
			owner: "zed-industries",
			repo:  "zed",
			path:  "assets/settings/default.json",
		},
		"alacritty": {
			owner: "alacritty",
			repo:  "alacritty",
			path:  "alacritty.yml",
		},
		"wezterm": {
			owner: "wez",
			repo:  "wezterm",
			path:  "wezterm.lua.in",
		},
	}

	repoInfo, exists := repos[app]
	if !exists {
		return nil, fmt.Errorf("unknown GitHub repo for %s", app)
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s",
		repoInfo.owner, repoInfo.repo, repoInfo.path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned %d", resp.StatusCode)
	}

	// Stream parse directly from response
	return e.parseGitHubStream(app, resp.Body)
}

// parseGitHubStream parses GitHub content without loading all into memory
func (e *FastExtractor) parseGitHubStream(app string, r io.Reader) (*ConfigReference, error) {
	config := &ConfigReference{
		AppName:    app,
		ConfigPath: fmt.Sprintf("~/.config/%s/settings.json", app),
		ConfigType: "json",
		Settings:   make(map[string]ConfigSetting),
	}

	// Use buffered reader for efficiency
	br := bufio.NewReader(r)

	// Simple line-by-line parsing (avoiding full JSON unmarshal for speed)
	lineNum := 0
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		lineNum++
		line = strings.TrimSpace(line)

		// Quick pattern matching
		if strings.Contains(line, `"`) && strings.Contains(line, `:`) {
			// Potential setting line
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `" ,`)
				value := strings.Trim(parts[1], `, `)

				if key != "" && !strings.HasPrefix(key, "//") {
					config.Settings[key] = ConfigSetting{
						Name:         key,
						DefaultValue: parseJSONValue(value),
						Type:         inferTypeFromJSON(value),
						Category:     inferCategoryQuick(key),
					}
				}
			}
		}
	}

	return config, nil
}

// extractFromLocalFile reads from local configs directory
func (e *FastExtractor) extractFromLocalFile(ctx context.Context, app string) (*ConfigReference, error) {
	path := fmt.Sprintf("resources/configs/%s.yaml", app)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// Check context before processing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stream parse YAML
	return e.parseYAMLStream(app, file)
}

// parseYAMLStream efficiently parses YAML without full unmarshal
func (e *FastExtractor) parseYAMLStream(app string, r io.Reader) (*ConfigReference, error) {
	config := &ConfigReference{
		AppName:  app,
		Settings: make(map[string]ConfigSetting),
	}

	scanner := bufio.NewScanner(r)
	var currentSetting string
	var currentIndent int

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		line = strings.TrimSpace(line)

		// Simple YAML parsing
		if indent == 0 && strings.HasSuffix(line, ":") {
			// Top level key
			key := strings.TrimSuffix(line, ":")
			if key == "settings" {
				currentIndent = indent
			}
		} else if indent > currentIndent && currentIndent >= 0 {
			// Setting entry
			if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
				currentSetting = strings.TrimSuffix(line, ":")
				config.Settings[currentSetting] = ConfigSetting{Name: currentSetting}
			} else if currentSetting != "" && strings.Contains(line, ":") {
				// Setting property
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.Trim(parts[1], ` "`)

					setting := config.Settings[currentSetting]
					switch key {
					case "type":
						setting.Type = SettingType(value)
					case "description", "desc":
						setting.Description = value
					case "default", "default_value":
						setting.DefaultValue = value
					case "category":
						setting.Category = value
					}
					config.Settings[currentSetting] = setting
				}
			}
		}
	}

	return config, scanner.Err()
}

// Cache methods

func (c *ExtractorCache) Get(key string) *ConfigReference {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(entry.timestamp) > c.ttl {
		return nil
	}

	return entry.config
}

func (c *ExtractorCache) Set(key string, config *ConfigReference) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		config:    config,
		timestamp: time.Now(),
	}
}

// Helper functions (optimized)

func inferTypeQuick(value interface{}) SettingType {
	switch v := value.(type) {
	case bool:
		return TypeBoolean
	case int, int32, int64, float32, float64:
		return TypeNumber
	case string:
		if v == "true" || v == "false" {
			return TypeBoolean
		}
		if isNumericQuick(v) {
			return TypeNumber
		}
		return TypeString
	default:
		return TypeString
	}
}

func inferCategoryQuick(key string) string {
	// Use prefix matching for speed
	switch {
	case strings.HasPrefix(key, "font"):
		return "font"
	case strings.HasPrefix(key, "color") || strings.HasPrefix(key, "theme"):
		return "appearance"
	case strings.HasPrefix(key, "window"):
		return "window"
	case strings.HasPrefix(key, "key") || strings.HasPrefix(key, "bind"):
		return "keybindings"
	default:
		return "general"
	}
}

func isNumericQuick(s string) bool {
	if s == "" {
		return false
	}
	// Quick check first char
	if s[0] != '-' && (s[0] < '0' || s[0] > '9') {
		return false
	}
	dotCount := 0
	for i := 1; i < len(s); i++ {
		if s[i] == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func parseJSONValue(s string) interface{} {
	s = strings.Trim(s, `, `)
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`)
	}
	if isNumericQuick(s) {
		return s
	}
	return s
}

func inferTypeFromJSON(value string) SettingType {
	value = strings.TrimSpace(value)
	if value == "true" || value == "false" {
		return TypeBoolean
	}
	if isNumericQuick(strings.Trim(value, `",`)) {
		return TypeNumber
	}
	if strings.HasPrefix(value, "[") {
		return TypeArray
	}
	if strings.HasPrefix(value, "{") {
		return TypeObject
	}
	return TypeString
}
