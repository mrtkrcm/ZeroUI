package legacyextractor

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

func (s *CLIStrategy) Name() string        { return "CLI" }
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

func (s *LocalStrategy) Name() string        { return "Local" }
func (s *LocalStrategy) Confidence() float64 { return 0.85 }

func (s *LocalStrategy) Extract(ctx context.Context, app string) (*Config, error) {
	// Try different extensions
	for _, ext := range []string{".yaml", ".yml", ".json", ".toml"} {
		path := filepath.Join(s.configDir, app+ext)
		if file, err := os.Open(path); err == nil {
			defer func() { _ = file.Close() }()
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
			"lazygit": {
				App:  "lazygit",
				Path: "~/.config/lazygit/config.yml",
				Type: "yaml",
				Settings: map[string]Setting{
					"gui.theme.activeBorderColor":   {Name: "gui.theme.activeBorderColor", Type: "array", Default: []string{"green", "bold"}, Cat: "appearance"},
					"gui.theme.inactiveBorderColor": {Name: "gui.theme.inactiveBorderColor", Type: "array", Default: []string{"white"}, Cat: "appearance"},
					"gui.showIcons":                 {Name: "gui.showIcons", Type: "boolean", Default: false, Cat: "appearance"},
					"git.paging.colorArg":           {Name: "git.paging.colorArg", Type: "string", Default: "always", Cat: "git"},
				},
			},
			"bat": {
				App:  "bat",
				Path: "~/.config/bat/config",
				Type: "flags",
				Settings: map[string]Setting{
					"theme":       {Name: "theme", Type: "string", Default: "TwoDark", Cat: "appearance"},
					"style":       {Name: "style", Type: "string", Default: "default", Cat: "appearance"},
					"italic-text": {Name: "italic-text", Type: "string", Default: "always", Cat: "appearance"},
					"paging":      {Name: "paging", Type: "string", Default: "auto", Cat: "general"},
				},
			},
			"ripgrep": {
				App:  "ripgrep",
				Path: "~/.ripgreprc",
				Type: "flags",
				Settings: map[string]Setting{
					"smart-case": {Name: "smart-case", Type: "boolean", Default: true, Cat: "search"},
					"hidden":     {Name: "hidden", Type: "boolean", Default: false, Cat: "search"},
					"colors":     {Name: "colors", Type: "string", Default: "line:none", Cat: "appearance"},
					"glob":       {Name: "glob", Type: "array", Default: []string{"!*.git"}, Cat: "search"},
				},
			},
			"rg": {
				App:  "ripgrep",
				Path: "~/.ripgreprc",
				Type: "flags",
				Settings: map[string]Setting{
					"smart-case": {Name: "smart-case", Type: "boolean", Default: true, Cat: "search"},
					"hidden":     {Name: "hidden", Type: "boolean", Default: false, Cat: "search"},
					"colors":     {Name: "colors", Type: "string", Default: "line:none", Cat: "appearance"},
					"glob":       {Name: "glob", Type: "array", Default: []string{"!*.git"}, Cat: "search"},
				},
			},
		},
	}
}

func (s *BuiltinStrategy) Name() string        { return "Builtin" }
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

func (s *GitHubStrategy) Name() string        { return "GitHub" }
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
	defer func() { _ = body.Close() }()

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
	format := detectFormat(app)
	if format == "flags" {
		return parseFlagFile(app, r)
	}

	cfg := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Type:     format,
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

func parseFlagFile(app string, r io.Reader) *Config {
	cfg := &Config{
		App:      app,
		Path:     fmt.Sprintf("~/.config/%s/config", app),
		Type:     "flags",
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "--") {
			line = strings.TrimPrefix(line, "--")
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			var value interface{}
			rawVal := "true"
			if len(parts) > 1 {
				rawVal = parts[1]
			}
			value = parseValue(rawVal)

			if existing, ok := cfg.Settings[key]; ok {
				// Handle repeated flags by converting to array
				// If existing default is already a slice, append
				// If not, create a slice
				var newValues []string

				// Convert existing value to string for slice
				existingStr := fmt.Sprintf("%v", existing.Default)

				// If the type was already array, we need to handle it properly
				// But parseValue returns a single value unless it had commas.
				// Here we are handling repeated keys.

				if existing.Type == "array" {
					if slice, ok := existing.Default.([]string); ok {
						newValues = append(slice, rawVal)
					} else {
						// Should not happen if we maintain type consistency,
						// but let's be safe.
						newValues = []string{existingStr, rawVal}
					}
				} else {
					newValues = []string{existingStr, rawVal}
				}

				cfg.Settings[key] = Setting{
					Name:    key,
					Type:    "array",
					Default: newValues,
					Cat:     existing.Cat,
				}
			} else {
				cfg.Settings[key] = Setting{
					Name:    key,
					Type:    inferType(rawVal),
					Default: value,
					Cat:     inferCategory(key),
				}
			}
		}
	}

	return cfg
}

func detectFormat(app string) string {
	switch app {
	case "zed", "vscode":
		return "json"
	case "alacritty", "starship", "lazygit":
		return "yaml"
	case "wezterm":
		return "lua"
	case "git":
		return "ini"
	case "bat", "ripgrep", "rg":
		return "flags"
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
