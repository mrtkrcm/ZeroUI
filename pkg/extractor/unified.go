// Package extractor provides a unified, minimal configuration extraction system
package extractor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Config represents extracted configuration
type Config struct {
	App      string            `yaml:"app_name"`
	Path     string            `yaml:"config_path"`
	Type     string            `yaml:"config_type"`
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

// Extractor handles all config extraction with minimal code
type Extractor struct {
	cache  sync.Map
	client *http.Client
	pool   chan struct{}
}

// New creates an extractor with sensible defaults
func New() *Extractor {
	return &Extractor{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 10,
			},
		},
		pool: make(chan struct{}, 8),
	}
}

// Extract gets config for an app using the fastest available method
func (e *Extractor) Extract(ctx context.Context, app string) (*Config, error) {
	// Check cache
	if v, ok := e.cache.Load(app); ok {
		return v.(*Config), nil
	}

	// Try methods in parallel
	type result struct {
		cfg *Config
		err error
	}
	
	ch := make(chan result, 3)
	
	// CLI extraction
	go func() {
		cfg, err := e.fromCLI(ctx, app)
		ch <- result{cfg, err}
	}()
	
	// GitHub extraction
	go func() {
		cfg, err := e.fromGitHub(ctx, app)
		ch <- result{cfg, err}
	}()
	
	// Local file extraction
	go func() {
		cfg, err := e.fromFile(ctx, app)
		ch <- result{cfg, err}
	}()
	
	// Return first success
	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-ch:
			if r.err == nil && r.cfg != nil {
				e.cache.Store(app, r.cfg)
				return r.cfg, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no extraction method succeeded for %s", app)
}

// ExtractAll extracts configs for multiple apps concurrently
func (e *Extractor) ExtractAll(ctx context.Context, apps []string) map[string]*Config {
	results := make(map[string]*Config)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	
	for _, app := range apps {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			
			// Rate limit
			e.pool <- struct{}{}
			defer func() { <-e.pool }()
			
			if cfg, err := e.Extract(ctx, a); err == nil {
				mu.Lock()
				results[a] = cfg
				mu.Unlock()
			}
		}(app)
	}
	
	wg.Wait()
	return results
}

// fromCLI extracts using CLI commands
func (e *Extractor) fromCLI(ctx context.Context, app string) (*Config, error) {
	cmds := map[string]string{
		"ghostty": "ghostty +show-config --default --docs",
		"zed":     "zed --print-config",
		"tmux":    "tmux show-options -g",
		"git":     "git config --list --show-origin",
	}
	
	cmd, ok := cmds[app]
	if !ok {
		return nil, fmt.Errorf("no CLI for %s", app)
	}
	
	out, err := exec.CommandContext(ctx, "sh", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}
	
	return parseCLI(app, out), nil
}

// fromGitHub fetches from GitHub repos
func (e *Extractor) fromGitHub(ctx context.Context, app string) (*Config, error) {
	urls := map[string]string{
		"zed":       "https://raw.githubusercontent.com/zed-industries/zed/main/assets/settings/default.json",
		"alacritty": "https://raw.githubusercontent.com/alacritty/alacritty/master/alacritty.yml",
		"wezterm":   "https://raw.githubusercontent.com/wez/wezterm/main/docs/config/lua/config/index.md",
	}
	
	url, ok := urls[app]
	if !ok {
		return nil, fmt.Errorf("no GitHub URL for %s", app)
	}
	
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	return parseStream(app, resp.Body), nil
}

// fromFile reads local config files
func (e *Extractor) fromFile(ctx context.Context, app string) (*Config, error) {
	path := fmt.Sprintf("configs/%s.yaml", app)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	return parseStream(app, f), nil
}

// parseCLI parses CLI output efficiently
func parseCLI(app string, data []byte) *Config {
	cfg := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Type:     "custom",
		Settings: make(map[string]Setting),
	}
	
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var key, desc string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Setting line (key = value)
		if idx := strings.Index(line, "="); idx > 0 {
			if key != "" && desc != "" {
				s := cfg.Settings[key]
				s.Desc = desc
				cfg.Settings[key] = s
				desc = ""
			}
			
			key = strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			
			cfg.Settings[key] = Setting{
				Name:    key,
				Type:    inferType(val),
				Default: val,
				Cat:     inferCat(key),
			}
		} else if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			// Comment/description
			desc += strings.TrimSpace(strings.TrimLeft(line, "#/")) + " "
		}
	}
	
	return cfg
}

// parseStream parses any stream format
func parseStream(app string, r io.Reader) *Config {
	cfg := &Config{
		App:      app,
		Settings: make(map[string]Setting),
	}
	
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Simple key detection
		if strings.Contains(line, ":") || strings.Contains(line, "=") {
			parts := strings.FieldsFunc(line, func(r rune) bool {
				return r == ':' || r == '='
			})
			
			if len(parts) >= 2 {
				key := strings.Trim(parts[0], `"' `)
				val := strings.Trim(parts[1], `"', `)
				
				if key != "" && !strings.HasPrefix(key, "#") && !strings.HasPrefix(key, "//") {
					cfg.Settings[key] = Setting{
						Name: key,
						Type: inferType(val),
						Cat:  inferCat(key),
					}
				}
			}
		}
	}
	
	return cfg
}

// inferType determines setting type from value
func inferType(v string) string {
	switch {
	case v == "true" || v == "false":
		return "boolean"
	case isNum(v):
		return "number"
	case strings.HasPrefix(v, "#") || strings.HasPrefix(v, "0x"):
		return "color"
	default:
		return "string"
	}
}

// inferCat determines category from key name
func inferCat(k string) string {
	prefixes := map[string]string{
		"font":   "font",
		"color":  "appearance",
		"theme":  "appearance",
		"window": "window",
		"key":    "keybindings",
		"bind":   "keybindings",
		"scroll": "editor",
		"cursor": "editor",
		"git":    "git",
	}
	
	for prefix, cat := range prefixes {
		if strings.HasPrefix(k, prefix) {
			return cat
		}
	}
	return "general"
}

// isNum checks if string is numeric
func isNum(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == '.' || (i == 0 && r == '-') {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}