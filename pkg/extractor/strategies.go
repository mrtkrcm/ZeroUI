package extractor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Strategy defines the extraction method interface
type Strategy interface {
	Name() string
	Extract(ctx context.Context, app string) (*Config, error)
	Confidence() float64
}

// CLIStrategy extracts config via CLI commands
type CLIStrategy struct {
	commands map[string]string
}

// NewCLIStrategy creates a CLI extraction strategy
func NewCLIStrategy() *CLIStrategy {
	return &CLIStrategy{
		commands: map[string]string{
			"ghostty":   "ghostty +show-config --default --docs",
			"alacritty": "alacritty --print-config",
			"wezterm":   "wezterm show-config",
			"tmux":      "tmux show-options -g",
			"git":       "git config --list --show-origin",
			"neovim":    "nvim --headless -c 'set all' -c 'qa' 2>&1",
			"starship":  "starship config",
		},
	}
}

func (s *CLIStrategy) Name() string { return "CLI" }
func (s *CLIStrategy) Confidence() float64 { return 0.95 }

func (s *CLIStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	cmd, exists := s.commands[app]
	if !exists {
		return nil, fmt.Errorf("no CLI command for %s", app)
	}

	execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	output, err := execCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("CLI failed: %w", err)
	}

	return parseCLIOutput(app, output), nil
}

// LocalStrategy extracts from local config files
type LocalStrategy struct {
	configDir string
}

// NewLocalStrategy creates a local file extraction strategy
func NewLocalStrategy(dir string) *LocalStrategy {
	if dir == "" {
		dir = "configs"
	}
	return &LocalStrategy{configDir: dir}
}

func (s *LocalStrategy) Name() string { return "Local" }
func (s *LocalStrategy) Confidence() float64 { return 0.85 }

func (s *LocalStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	// Try different extensions
	for _, ext := range []string{".yaml", ".yml", ".json", ".toml"} {
		path := filepath.Join(s.configDir, app+ext)
		if file, err := os.Open(path); err == nil {
			defer file.Close()
			return parseConfigFile(app, file), nil
		}
	}
	return nil, fmt.Errorf("no local config for %s", app)
}

// BuiltinStrategy provides fallback configs
type BuiltinStrategy struct {
	configs map[string]*Config
}

// NewBuiltinStrategy creates a builtin fallback strategy
func NewBuiltinStrategy() *BuiltinStrategy {
	return &BuiltinStrategy{
		configs: map[string]*Config{
			"tmux": {
				App:  "tmux",
				Path: "~/.tmux.conf",
				Type: "custom",
				Settings: map[string]Setting{
					"prefix":        {Name: "prefix", Type: "key", Default: "C-b", Cat: "keybindings"},
					"base-index":    {Name: "base-index", Type: "number", Default: 0, Cat: "general"},
					"mouse":         {Name: "mouse", Type: "boolean", Default: false, Cat: "input"},
					"history-limit": {Name: "history-limit", Type: "number", Default: 2000, Cat: "general"},
				},
			},
			"git": {
				App:  "git",
				Path: "~/.gitconfig",
				Type: "ini",
				Settings: map[string]Setting{
					"user.name":          {Name: "user.name", Type: "string", Cat: "user"},
					"user.email":         {Name: "user.email", Type: "string", Cat: "user"},
					"core.editor":        {Name: "core.editor", Type: "string", Default: "vim", Cat: "core"},
					"init.defaultBranch": {Name: "init.defaultBranch", Type: "string", Default: "main", Cat: "init"},
				},
			},
		},
	}
}

func (s *BuiltinStrategy) Name() string { return "Builtin" }
func (s *BuiltinStrategy) Confidence() float64 { return 0.60 }

func (s *BuiltinStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	if config, exists := s.configs[app]; exists {
		return config, nil
	}
	return nil, fmt.Errorf("no builtin config for %s", app)
}

// GitHubStrategy extracts from GitHub repositories
type GitHubStrategy struct {
	client HTTPClient
	repos  map[string]repoInfo
}

type repoInfo struct {
	owner string
	repo  string
	path  string
}

// NewGitHubStrategy creates a GitHub extraction strategy
func NewGitHubStrategy(client HTTPClient) *GitHubStrategy {
	return &GitHubStrategy{
		client: client,
		repos: map[string]repoInfo{
			"zed":       {"zed-industries", "zed", "assets/settings/default.json"},
			"alacritty": {"alacritty", "alacritty", "alacritty.yml"},
			"wezterm":   {"wez", "wezterm", "docs/config/lua/config/index.md"},
			"neovim":    {"neovim", "neovim", "runtime/doc/options.txt"},
		},
	}
}

func (s *GitHubStrategy) Name() string { return "GitHub" }
func (s *GitHubStrategy) Confidence() float64 { return 0.80 }

func (s *GitHubStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	repo, exists := s.repos[app]
	if !exists {
		return nil, fmt.Errorf("no GitHub repo for %s", app)
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s",
		repo.owner, repo.repo, repo.path)

	body, err := s.client.Get(ctx, url)
	if err != nil {
		// Try master branch
		url = strings.Replace(url, "/main/", "/master/", 1)
		body, err = s.client.Get(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("GitHub fetch failed: %w", err)
		}
	}
	defer body.Close()

	return parseConfigFile(app, body), nil
}

// Parsing functions

func parseCLIOutput(app string, data []byte) *Config {
	cfg := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Type:     "custom",
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var currentKey string
	var descBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check for setting line (key = value or key: value)
		if idx := strings.Index(line, "="); idx > 0 {
			processSettingLine(cfg, line, idx, &currentKey, &descBuilder)
		} else if idx := strings.Index(line, ":"); idx > 0 && !strings.Contains(line[:idx], " ") {
			processSettingLine(cfg, line, idx, &currentKey, &descBuilder)
		} else if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			// Accumulate description
			desc := strings.TrimSpace(strings.TrimLeft(line, "#/"))
			if desc != "" {
				if descBuilder.Len() > 0 {
					descBuilder.WriteString(" ")
				}
				descBuilder.WriteString(desc)
			}
		}
	}

	// Save last description if any
	if currentKey != "" && descBuilder.Len() > 0 {
		if setting, ok := cfg.Settings[currentKey]; ok {
			setting.Desc = descBuilder.String()
			cfg.Settings[currentKey] = setting
		}
	}

	return cfg
}

func processSettingLine(cfg *Config, line string, sepIdx int, currentKey *string, descBuilder *strings.Builder) {
	// Save previous description
	if *currentKey != "" && descBuilder.Len() > 0 {
		if setting, ok := cfg.Settings[*currentKey]; ok {
			setting.Desc = descBuilder.String()
			cfg.Settings[*currentKey] = setting
		}
		descBuilder.Reset()
	}

	key := strings.TrimSpace(line[:sepIdx])
	value := strings.TrimSpace(line[sepIdx+1:])

	*currentKey = key
	cfg.Settings[key] = Setting{
		Name:    key,
		Type:    inferType(value),
		Default: parseValue(value),
		Cat:     inferCategory(key),
	}
}

func parseConfigFile(app string, r io.Reader) *Config {
	cfg := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Type:     detectFormat(app),
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse key-value pairs
		for _, sep := range []string{":", "=", " "} {
			if idx := strings.Index(line, sep); idx > 0 {
				key := strings.Trim(line[:idx], `"' `)
				value := strings.Trim(line[idx+len(sep):], `"', `)
				
				if key != "" && !strings.ContainsAny(key, "{}[]") {
					cfg.Settings[key] = Setting{
						Name:    key,
						Type:    inferType(value),
						Default: parseValue(value),
						Cat:     inferCategory(key),
					}
				}
				break
			}
		}
	}

	return cfg
}

func detectFormat(app string) string {
	switch app {
	case "zed", "vscode":
		return "json"
	case "alacritty", "starship":
		return "yaml"
	case "wezterm":
		return "lua"
	case "git":
		return "ini"
	default:
		return "custom"
	}
}

func parseValue(val string) interface{} {
	val = strings.Trim(val, `"'`)
	
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}
	if isNumeric(val) {
		return val // Keep as string to avoid precision issues
	}
	return val
}