package configextractor

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// SimpleExtractor provides a streamlined config extraction interface
type SimpleExtractor struct {
	cache sync.Map // Thread-safe cache for extracted configs
}

// NewSimpleExtractor creates an optimized extractor
func NewSimpleExtractor() *SimpleExtractor {
	return &SimpleExtractor{}
}

// Extract attempts to extract config using the most efficient method
func (e *SimpleExtractor) Extract(app string) (*Config, error) {
	// Check cache first (performance optimization)
	if cached, ok := e.cache.Load(app); ok {
		return cached.(*Config), nil
	}

	// Try extraction methods in order of speed
	methods := []func(string) (*Config, error){
		e.extractFromCLI,     // Fastest if available
		e.extractFromBuiltin, // Pre-defined configs
		e.extractFromGitHub,  // Network call (slowest)
	}

	for _, method := range methods {
		if config, err := method(app); err == nil && config != nil {
			e.cache.Store(app, config) // Cache successful extraction
			return config, nil
		}
	}

	return nil, fmt.Errorf("no extraction method succeeded for %s", app)
}

// Config represents extracted configuration (simplified)
type Config struct {
	App      string            `yaml:"app"`
	Path     string            `yaml:"path"`
	Settings map[string]Setting `yaml:"settings"`
}

// Setting represents a single config option (minimal fields)
type Setting struct {
	Type    string   `yaml:"type"`
	Default any      `yaml:"default,omitempty"`
	Values  []string `yaml:"values,omitempty"`
	Desc    string   `yaml:"desc,omitempty"`
}

// extractFromCLI uses CLI if available (fastest method)
func (e *SimpleExtractor) extractFromCLI(app string) (*Config, error) {
	commands := map[string]string{
		"ghostty": "ghostty +show-config --default --docs",
		"zed":     "zed --print-config",
	}

	cmd, exists := commands[app]
	if !exists {
		return nil, fmt.Errorf("no CLI command for %s", app)
	}

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return nil, err
	}

	return parseCLIOutput(app, string(output)), nil
}

// extractFromBuiltin returns pre-defined configs (fastest, no external calls)
func (e *SimpleExtractor) extractFromBuiltin(app string) (*Config, error) {
	// Minimal built-in configs for common apps
	configs := map[string]*Config{
		"tmux": {
			App:  "tmux",
			Path: "~/.tmux.conf",
			Settings: map[string]Setting{
				"prefix":         {Type: "key", Default: "C-b"},
				"base-index":     {Type: "number", Default: 0},
				"mouse":          {Type: "boolean", Default: false},
				"history-limit":  {Type: "number", Default: 2000},
				"status":         {Type: "boolean", Default: true},
			},
		},
		"git": {
			App:  "git",
			Path: "~/.gitconfig",
			Settings: map[string]Setting{
				"user.name":     {Type: "string"},
				"user.email":    {Type: "string"},
				"core.editor":   {Type: "string", Default: "vim"},
				"init.defaultBranch": {Type: "string", Default: "main"},
			},
		},
	}

	if config, ok := configs[app]; ok {
		return config, nil
	}
	return nil, fmt.Errorf("no builtin config for %s", app)
}

// extractFromGitHub fetches from GitHub (network call, slowest)
func (e *SimpleExtractor) extractFromGitHub(app string) (*Config, error) {
	// Simplified GitHub extraction
	// Implementation would fetch and parse from known repos
	return nil, fmt.Errorf("GitHub extraction not implemented")
}

// parseCLIOutput parses CLI output into Config (optimized parser)
func parseCLIOutput(app, output string) *Config {
	config := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Settings: make(map[string]Setting),
	}

	lines := strings.Split(output, "\n")
	var currentKey string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Fast parsing with simple checks
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				
				currentKey = key
				config.Settings[key] = Setting{
					Type:    inferType(val),
					Default: parseValue(val),
				}
			}
		} else if strings.HasPrefix(line, "#") && currentKey != "" {
			// Add description to current setting
			if setting, ok := config.Settings[currentKey]; ok {
				setting.Desc = strings.TrimPrefix(line, "# ")
				config.Settings[currentKey] = setting
			}
		}
	}

	return config
}

// inferType quickly determines setting type
func inferType(val string) string {
	switch {
	case val == "true" || val == "false":
		return "boolean"
	case isNumber(val):
		return "number"
	case strings.Contains(val, "#") || strings.HasPrefix(val, "0x"):
		return "color"
	default:
		return "string"
	}
}

// isNumber checks if string is numeric (optimized)
func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if (c < '0' || c > '9') && c != '.' && (i != 0 || c != '-') {
			return false
		}
	}
	return true
}

// parseValue converts string to appropriate type
func parseValue(val string) any {
	switch {
	case val == "true":
		return true
	case val == "false":
		return false
	case isNumber(val):
		// Return as string to avoid precision issues
		return val
	default:
		return val
	}
}