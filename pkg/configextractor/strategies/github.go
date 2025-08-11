package strategies

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
)

// GitHub strategy extracts configuration from GitHub repositories
type GitHub struct {
	client *http.Client
	repos  map[string]RepoInfo
}

// RepoInfo defines GitHub repository information for config extraction
type RepoInfo struct {
	Owner      string   // Repository owner
	Repo       string   // Repository name  
	Paths      []string // Potential config file paths to try
	Branch     string   // Branch to use (defaults to main/master)
	Format     string   // Expected config format
	Confidence float64  // Confidence score for this source
}

// NewGitHub creates a new GitHub extraction strategy
func NewGitHub() *GitHub {
	return &GitHub{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		repos: map[string]RepoInfo{
			"zed": {
				Owner:  "zed-industries",
				Repo:   "zed",
				Paths:  []string{"assets/settings/default.json", "crates/editor/src/settings.rs"},
				Format: "json",
				Confidence: 0.85,
			},
			"alacritty": {
				Owner:  "alacritty",
				Repo:   "alacritty",
				Paths:  []string{"alacritty.yml", "extra/alacritty.yml"},
				Format: "yaml",
				Confidence: 0.80,
			},
			"wezterm": {
				Owner:  "wez",
				Repo:   "wezterm",
				Paths:  []string{"docs/config/lua/config/index.md", "config/src/lib.rs"},
				Format: "markdown",
				Confidence: 0.75,
			},
			"neovim": {
				Owner:  "neovim",
				Repo:   "neovim",
				Paths:  []string{"runtime/doc/options.txt", "src/nvim/options.lua"},
				Format: "vimdoc",
				Confidence: 0.70,
			},
			"tmux": {
				Owner:  "tmux",
				Repo:   "tmux",
				Paths:  []string{"options-table.c", "tmux.1"},
				Format: "manpage",
				Confidence: 0.65,
			},
		},
	}
}

// Name returns strategy identifier
func (g *GitHub) Name() string {
	return "github"
}

// CanExtract checks if GitHub strategy can extract config for app
func (g *GitHub) CanExtract(app string) bool {
	_, exists := g.repos[app]
	return exists
}

// Extract performs GitHub-based config extraction
func (g *GitHub) Extract(ctx context.Context, app string) (*configextractor.Config, error) {
	repo, exists := g.repos[app]
	if !exists {
		return nil, fmt.Errorf("no GitHub repository configured for %s", app)
	}

	// Try each configured path until one succeeds
	var lastErr error
	for _, path := range repo.Paths {
		config, err := g.extractFromPath(ctx, app, repo, path)
		if err == nil && config != nil {
			return config, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to extract from GitHub for %s: %w", app, lastErr)
}

// Priority returns strategy priority (GitHub is medium priority)
func (g *GitHub) Priority() int {
	return 50 // Medium priority - network dependent
}

// extractFromPath attempts extraction from a specific GitHub file path
func (g *GitHub) extractFromPath(ctx context.Context, app string, repo RepoInfo, path string) (*configextractor.Config, error) {
	// Try main branch first, then master
	branches := []string{"main", "master"}
	if repo.Branch != "" {
		branches = []string{repo.Branch}
	}

	for _, branch := range branches {
		content, err := g.fetchFile(ctx, repo.Owner, repo.Repo, branch, path)
		if err == nil {
			return g.parseContent(app, repo, path, content)
		}
	}

	return nil, fmt.Errorf("file not found: %s", path)
}

// fetchFile retrieves a file from GitHub's raw content API
func (g *GitHub) fetchFile(ctx context.Context, owner, repo, branch, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, path)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent for better rate limiting
	req.Header.Set("User-Agent", "ConfigToggle-Extractor/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body with size limit for safety
	const maxSize = 10 << 20 // 10MB limit
	body := make([]byte, 0, 8192)
	buf := make([]byte, 8192)
	
	for len(body) < maxSize {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return body, nil
}

// parseContent parses the fetched content based on format and path
func (g *GitHub) parseContent(app string, repo RepoInfo, path string, content []byte) (*configextractor.Config, error) {
	config := &configextractor.Config{
		App:      app,
		Format:   repo.Format,
		Settings: make(map[string]configextractor.Setting),
		Source: configextractor.ExtractionSource{
			Method:     "github",
			Location:   fmt.Sprintf("%s/%s/%s", repo.Owner, repo.Repo, path),
			Confidence: repo.Confidence,
		},
		Timestamp: time.Now(),
	}

	// Parse based on file extension and format
	switch {
	case strings.HasSuffix(path, ".json"):
		return g.parseJSON(config, content)
	case strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml"):
		return g.parseYAML(config, content)
	case strings.HasSuffix(path, ".md"):
		return g.parseMarkdown(config, content)
	case strings.HasSuffix(path, ".rs"):
		return g.parseRustSource(config, content)
	case strings.HasSuffix(path, ".c"):
		return g.parseCSource(config, content)
	case strings.HasSuffix(path, ".txt"):
		return g.parseText(config, content)
	default:
		// Try to detect format from content
		return g.parseGeneric(config, content)
	}
}

// parseJSON parses JSON configuration files
func (g *GitHub) parseJSON(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	// Fast line-by-line JSON parsing to avoid full unmarshaling
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}
		
		// Look for key-value pairs: "key": value
		if strings.Contains(line, ":") && strings.Contains(line, `"`) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(strings.TrimSpace(parts[0]), `"`)
				value := strings.TrimSpace(strings.Trim(parts[1], `,`))
				
				if key != "" && !strings.HasPrefix(key, "_") && !strings.HasPrefix(key, "$") {
					config.Settings[key] = configextractor.Setting{
						Name: key,
						Type: inferTypeFromJSON(value),
						Cat:  inferCategory(key),
					}
				}
			}
		}
	}

	return config, nil
}

// parseYAML parses YAML configuration files
func (g *GitHub) parseYAML(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	var currentSection string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Look for key-value pairs
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Check if this is a section header (no value)
				if value == "" {
					currentSection = key
					continue
				}
				
				// Combine section and key if in a section
				fullKey := key
				if currentSection != "" {
					fullKey = currentSection + "." + key
				}
				
				config.Settings[fullKey] = configextractor.Setting{
					Name: fullKey,
					Type: inferType(value),
					Cat:  inferCategory(fullKey),
				}
			}
		}
	}

	return config, nil
}

// parseMarkdown parses markdown documentation for config options
func (g *GitHub) parseMarkdown(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	var inCodeBlock bool
	var currentSetting string
	var description strings.Builder
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Track code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		
		// Skip content inside code blocks unless it's config
		if inCodeBlock && !strings.Contains(line, "config.") {
			continue
		}
		
		// Look for config options in headers or code
		if strings.Contains(line, "config.") || strings.HasPrefix(line, "##") {
			// Extract config key from various formats
			var key string
			if strings.Contains(line, "config.") {
				// Extract from "config.key" format
				parts := strings.Split(line, "config.")
				if len(parts) > 1 {
					key = strings.Fields(parts[1])[0]
					key = strings.Trim(key, "` ()")
				}
			} else if strings.HasPrefix(line, "##") {
				// Header might be a setting name
				key = strings.TrimSpace(strings.TrimPrefix(line, "##"))
				key = strings.Fields(key)[0] // Take first word
			}
			
			if key != "" {
				// Save previous setting
				if currentSetting != "" && description.Len() > 0 {
					if setting, exists := config.Settings[currentSetting]; exists {
						setting.Desc = description.String()
						config.Settings[currentSetting] = setting
					}
				}
				
				currentSetting = key
				config.Settings[key] = configextractor.Setting{
					Name: key,
					Type: configextractor.TypeString, // Default type
					Cat:  inferCategory(key),
				}
				description.Reset()
			}
		} else if currentSetting != "" && line != "" {
			// Collect description text
			if description.Len() > 0 {
				description.WriteString(" ")
			}
			description.WriteString(line)
		}
	}
	
	// Save last setting
	if currentSetting != "" && description.Len() > 0 {
		if setting, exists := config.Settings[currentSetting]; exists {
			setting.Desc = description.String()
			config.Settings[currentSetting] = setting
		}
	}

	return config, nil
}

// parseRustSource parses Rust source files for config definitions
func (g *GitHub) parseRustSource(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Look for struct fields or string literals that might be config keys
		if strings.Contains(line, "pub") && strings.Contains(line, ":") {
			// Parse struct fields: pub field_name: Type,
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "pub" && i+1 < len(parts) {
					fieldName := strings.TrimSuffix(parts[i+1], ":")
					if fieldName != "" {
						config.Settings[fieldName] = configextractor.Setting{
							Name: fieldName,
							Type: configextractor.TypeString,
							Cat:  inferCategory(fieldName),
						}
					}
					break
				}
			}
		}
	}

	return config, nil
}

// parseCSource parses C source files for config definitions
func (g *GitHub) parseCSource(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Look for option definitions in C code
		if strings.Contains(line, "{") && strings.Contains(line, `"`) {
			// Parse option table entries: {"option-name", ...}
			start := strings.Index(line, `"`)
			if start >= 0 {
				end := strings.Index(line[start+1:], `"`)
				if end >= 0 {
					optionName := line[start+1 : start+1+end]
					if optionName != "" && !strings.Contains(optionName, " ") {
						config.Settings[optionName] = configextractor.Setting{
							Name: optionName,
							Type: configextractor.TypeString,
							Cat:  inferCategory(optionName),
						}
					}
				}
			}
		}
	}

	return config, nil
}

// parseText parses generic text files (like man pages or help text)
func (g *GitHub) parseText(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Look for option patterns in text
		// Pattern 1: -option, --option
		if strings.Contains(line, "-") && (strings.Contains(line, "--") || len(line) > 10) {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasPrefix(field, "--") {
					optionName := strings.TrimPrefix(field, "--")
					optionName = strings.Trim(optionName, ".,;")
					if optionName != "" {
						config.Settings[optionName] = configextractor.Setting{
							Name: optionName,
							Type: configextractor.TypeString,
							Cat:  inferCategory(optionName),
						}
					}
				}
			}
		}
	}

	return config, nil
}

// parseGeneric attempts to parse content with unknown format
func (g *GitHub) parseGeneric(config *configextractor.Config, content []byte) (*configextractor.Config, error) {
	// Try different parsing strategies
	contentStr := string(content)
	
	// Check if it looks like JSON
	if strings.Contains(contentStr, "{") && strings.Contains(contentStr, ":") {
		return g.parseJSON(config, content)
	}
	
	// Check if it looks like YAML
	if strings.Contains(contentStr, ":") && !strings.Contains(contentStr, "{") {
		return g.parseYAML(config, content)
	}
	
	// Default to text parsing
	return g.parseText(config, content)
}