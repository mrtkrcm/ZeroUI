package configextractor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CLI strategy extracts configuration from command-line tools
type CLI struct {
	commands map[string]CLICommand
	// Runner allows injecting a command runner (for tests or alternate execution).
	// If nil, NewCLI initializes it to the OS runner.
	Runner Runner
}

// CLICommand defines how to extract config from a CLI tool
type CLICommand struct {
	Command    string                           // Command to execute
	Args       []string                         // Command arguments
	Parser     func(app, output string) *Config // Custom parser
	Timeout    time.Duration                    // Execution timeout
	Confidence float64                          // Confidence score (0-1)
}

// NewCLI creates a new CLI extraction strategy
func NewCLI() *CLI {
	return &CLI{
		commands: map[string]CLICommand{
			"ghostty": {
				Command:    "ghostty",
				Args:       []string{"+show-config", "--default", "--docs"},
				Parser:     parseGhosttyOutput,
				Timeout:    10 * time.Second,
				Confidence: 0.95, // High confidence - official CLI
			},
			"zed": {
				Command:    "zed",
				Args:       []string{"--print-config"},
				Parser:     parseJSONOutput,
				Timeout:    5 * time.Second,
				Confidence: 0.90,
			},
			"wezterm": {
				Command:    "wezterm",
				Args:       []string{"show-config"},
				Parser:     parseLuaOutput,
				Timeout:    5 * time.Second,
				Confidence: 0.90,
			},
			"tmux": {
				Command:    "tmux",
				Args:       []string{"show-options", "-g"},
				Parser:     parseTmuxOutput,
				Timeout:    3 * time.Second,
				Confidence: 0.85,
			},
			"git": {
				Command:    "git",
				Args:       []string{"config", "--list", "--show-origin"},
				Parser:     parseGitOutput,
				Timeout:    3 * time.Second,
				Confidence: 0.90,
			},
		},
		// Default runner uses the OS; tests may inject a fake runner by setting CLI.Runner.
		Runner: NewOSRunner(),
	}
}

// Name returns strategy identifier
func (c *CLI) Name() string {
	return "cli"
}

// CanExtract checks if CLI strategy can extract config for app
func (c *CLI) CanExtract(app string) bool {
	_, exists := c.commands[app]
	return exists
}

// Extract performs CLI-based config extraction
func (c *CLI) Extract(ctx context.Context, app string) (*Config, error) {
	cmd, exists := c.commands[app]
	if !exists {
		return nil, fmt.Errorf("no CLI command defined for %s", app)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	// Execute command (resolve binary path; allow fallback by searching upward for repo-local test binaries)
	resolved := cmd.Command
	if _, err := exec.LookPath(cmd.Command); err != nil {
		// Walk up from the current working directory toward filesystem root.
		// At each directory check for common test locations such as:
		//   - <dir>/testdata/bin/{cmd}
		//   - <dir>/tools/plugins/{cmd}[.sh]
		//   - <dir>/tools/plugins/{cmd}-rpc/{cmd}[.sh]
		// Stop when we reach filesystem root or find a go.mod (repo root heuristic).
		dir, _ := os.Getwd()
		for {
			tryPaths := []string{
				filepath.Join(dir, "testdata", "bin", cmd.Command),
				filepath.Join(dir, "testdata", "bin", cmd.Command+".sh"),
				filepath.Join(dir, "tools", "plugins", cmd.Command),
				filepath.Join(dir, "tools", "plugins", cmd.Command+".sh"),
				filepath.Join(dir, "tools", "plugins", cmd.Command+"-rpc", cmd.Command),
				filepath.Join(dir, "tools", "plugins", cmd.Command+"-rpc", cmd.Command+".sh"),
			}

			found := false
			for _, p := range tryPaths {
				if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
					resolved = p
					found = true
					break
				}
			}
			if found {
				break
			}

			// If current directory contains go.mod, treat it as repo root and stop searching upward.
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir || parent == "" {
				// reached filesystem root
				break
			}
			dir = parent
		}
	}

	// Use injected Runner if available; otherwise use default OS runner.
	runner := c.Runner
	if runner == nil {
		runner = NewOSRunner()
	}
	stdout, stderr, err := runner.Run(execCtx, resolved, cmd.Args...)
	// Combine stdout/stderr for parsers that expect combined output
	output := append([]byte{}, stdout...)
	if len(stderr) > 0 {
		if len(output) > 0 {
			output = append(output, '\n')
		}
		output = append(output, stderr...)
	}
	if err != nil {
		return nil, fmt.Errorf("CLI command failed for %s: %w", app, err)
	}

	// Parse output using custom parser
	config := cmd.Parser(app, string(output))
	if config == nil {
		return nil, fmt.Errorf("failed to parse CLI output for %s", app)
	}

	// Set extraction source information
	config.Source = ExtractionSource{
		Method:     "cli",
		Location:   fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " ")),
		Confidence: cmd.Confidence,
	}
	config.Timestamp = time.Now()

	return config, nil
}

// Priority returns strategy priority (CLI is highest priority when available)
func (c *CLI) Priority() int {
	return 100 // Highest priority
}

// Parser functions for different CLI outputs

// parseGhosttyOutput parses Ghostty's config output format
func parseGhosttyOutput(app, output string) *Config {
	config := &Config{
		App:      app,
		Format:   "custom",
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentSetting string
	var description strings.Builder

	// Pattern to match setting lines: "setting-name = value"
	settingPattern := regexp.MustCompile(`^([a-z-_]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Comment lines (descriptions)
		if strings.HasPrefix(line, "#") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if desc != "" {
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(desc)
			}
			continue
		}

		// Setting definition line
		if matches := settingPattern.FindStringSubmatch(line); matches != nil {
			// Save previous setting with accumulated description
			if currentSetting != "" && description.Len() > 0 {
				if setting, exists := config.Settings[currentSetting]; exists {
					setting.Desc = description.String()
					config.Settings[currentSetting] = setting
				}
			}

			// Start new setting
			currentSetting = matches[1]
			value := strings.TrimSpace(matches[2])

			config.Settings[currentSetting] = Setting{
				Name:    currentSetting,
				Type:    inferType(value),
				Default: parseValue(value),
				Cat:     inferCategory(currentSetting),
			}

			description.Reset()
		}
	}

	// Handle last setting
	if currentSetting != "" && description.Len() > 0 {
		if setting, exists := config.Settings[currentSetting]; exists {
			setting.Desc = description.String()
			config.Settings[currentSetting] = setting
		}
	}

	return config
}

// parseJSONOutput parses JSON configuration output
func parseJSONOutput(app, output string) *Config {
	// Simple JSON parsing without full unmarshaling for performance
	config := &Config{
		App:      app,
		Format:   "json",
		Settings: make(map[string]Setting),
	}

	// Line-by-line parsing to avoid full JSON unmarshaling
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for key-value pairs: "key": value
		if strings.Contains(line, ":") && strings.Contains(line, `"`) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(strings.TrimSpace(parts[0]), `"`)
				value := strings.TrimSpace(strings.Trim(parts[1], `,`))

				if key != "" && !strings.HasPrefix(key, "_") {
					config.Settings[key] = Setting{
						Name: key,
						Type: inferTypeFromJSON(value),
						Cat:  inferCategory(key),
					}
				}
			}
		}
	}

	return config
}

// parseLuaOutput parses Lua configuration output
func parseLuaOutput(app, output string) *Config {
	config := &Config{
		App:      app,
		Format:   "lua",
		Settings: make(map[string]Setting),
	}

	// Simple Lua config parsing
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for assignments: config.key = value
		if strings.Contains(line, "=") && strings.Contains(line, "config.") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				keyPart := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Extract key from config.key format
				if strings.HasPrefix(keyPart, "config.") {
					key := strings.TrimPrefix(keyPart, "config.")

					config.Settings[key] = Setting{
						Name: key,
						Type: inferTypeFromLua(value),
						Cat:  inferCategory(key),
					}
				}
			}
		}
	}

	return config
}

// parseTmuxOutput parses tmux show-options output
func parseTmuxOutput(app, output string) *Config {
	config := &Config{
		App:      app,
		Format:   "custom",
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// tmux format: "option value"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			config.Settings[key] = Setting{
				Name:    key,
				Type:    inferType(value),
				Default: parseValue(value),
				Cat:     inferCategory(key),
			}
		}
	}

	return config
}

// parseGitOutput parses git config output
func parseGitOutput(app, output string) *Config {
	config := &Config{
		App:      app,
		Format:   "gitconfig",
		Settings: make(map[string]Setting),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Git config format: "file:key=value"
		if strings.Contains(line, "=") {
			// Remove file prefix if present
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					line = parts[1]
				}
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				config.Settings[key] = Setting{
					Name:    key,
					Type:    inferType(value),
					Default: parseValue(value),
					Cat:     inferCategory(key),
				}
			}
		}
	}

	return config
}

// Helper functions for type inference and parsing

// inferType determines setting type from value string
func inferType(value string) SettingType {
	value = strings.TrimSpace(value)

	switch {
	case value == "true" || value == "false":
		return TypeBoolean
	case isNumeric(value):
		return TypeNumber
	case strings.HasPrefix(value, "[") || strings.HasPrefix(value, "{"):
		return TypeArray
	default:
		return TypeString
	}
}

// inferTypeFromJSON determines type from JSON value
func inferTypeFromJSON(value string) SettingType {
	value = strings.TrimSpace(strings.Trim(value, `,`))

	switch {
	case value == "true" || value == "false":
		return TypeBoolean
	case strings.HasPrefix(value, `"`):
		return TypeString
	case strings.HasPrefix(value, "["):
		return TypeArray
	case strings.HasPrefix(value, "{"):
		return TypeString // Simplified
	case isNumeric(value):
		return TypeNumber
	default:
		return TypeString
	}
}

// inferTypeFromLua determines type from Lua value
func inferTypeFromLua(value string) SettingType {
	value = strings.TrimSpace(value)

	switch {
	case value == "true" || value == "false":
		return TypeBoolean
	case strings.HasPrefix(value, `"`):
		return TypeString
	case strings.HasPrefix(value, "{"):
		return TypeArray
	case isNumeric(value):
		return TypeNumber
	default:
		return TypeString
	}
}

// isNumeric checks if string represents a number
func isNumeric(s string) bool {
	if s == "" || s == "-" {
		return false
	}

	dotCount := 0
	for i, r := range s {
		if r == '.' {
			dotCount++
			if dotCount > 1 {
				return false
			}
		} else if r == '-' && i != 0 {
			return false
		} else if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parseValue converts string to appropriate Go type
func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	switch {
	case value == "true":
		return true
	case value == "false":
		return false
	case isNumeric(value):
		return value // Keep as string to avoid precision loss
	default:
		return value
	}
}

// inferCategory determines setting category from name
func inferCategory(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "font"):
		return "font"
	case strings.Contains(name, "color") || strings.Contains(name, "theme"):
		return "appearance"
	case strings.Contains(name, "window"):
		return "window"
	case strings.Contains(name, "key") || strings.Contains(name, "bind"):
		return "keybindings"
	case strings.Contains(name, "cursor"):
		return "editor"
	case strings.Contains(name, "scroll"):
		return "scrolling"
	case strings.Contains(name, "shell"):
		return "terminal"
	default:
		return "general"
	}
}
