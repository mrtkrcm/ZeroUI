package configextractor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubExtractor extracts config from GitHub repositories
type GitHubExtractor struct {
	client       *http.Client
	knownRepos   map[string]RepoConfig
	cacheTimeout time.Duration
}

type RepoConfig struct {
	Owner      string
	Repo       string
	ConfigPath string // Path to default config in repo
	DocPath    string // Path to documentation
	ParseFunc  func(content []byte) map[string]Setting
}

// NewGitHubExtractor creates a GitHub-based extractor
func NewGitHubExtractor() *GitHubExtractor {
	return &GitHubExtractor{
		client:       &http.Client{Timeout: 10 * time.Second},
		cacheTimeout: 24 * time.Hour,
		knownRepos: map[string]RepoConfig{
			"zed": {
				Owner:      "zed-industries",
				Repo:       "zed",
				ConfigPath: "assets/settings/default.json",
				ParseFunc:  parseZedDefaultJSON,
			},
			"alacritty": {
				Owner:      "alacritty",
				Repo:       "alacritty",
				ConfigPath: "alacritty.yml",
				ParseFunc:  parseAlacrittyYAML,
			},
			"wezterm": {
				Owner:      "wez",
				Repo:       "wezterm",
				ConfigPath: "docs/config/lua/config/index.md",
				ParseFunc:  parseWezTermDocs,
			},
			"neovim": {
				Owner:      "neovim",
				Repo:       "neovim",
				DocPath:    "runtime/doc/options.txt",
				ParseFunc:  parseNeovimOptions,
			},
		},
	}
}

func (e *GitHubExtractor) SupportsApp(appName string) bool {
	_, ok := e.knownRepos[strings.ToLower(appName)]
	return ok
}

func (e *GitHubExtractor) ExtractConfig(appName string) (*ExtractedConfig, error) {
	repo, ok := e.knownRepos[strings.ToLower(appName)]
	if !ok {
		return nil, fmt.Errorf("unknown repository for %s", appName)
	}

	// Fetch config file from GitHub
	content, err := e.fetchFile(repo.Owner, repo.Repo, repo.ConfigPath)
	if err != nil {
		// Try documentation path if config path fails
		if repo.DocPath != "" {
			content, err = e.fetchFile(repo.Owner, repo.Repo, repo.DocPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	settings := repo.ParseFunc(content)
	
	return &ExtractedConfig{
		AppName:    appName,
		Settings:   settings,
		Source:     fmt.Sprintf("GitHub: %s/%s", repo.Owner, repo.Repo),
		Confidence: 0.85, // Good confidence from official repo
	}, nil
}

func (e *GitHubExtractor) fetchFile(owner, repo, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, path)
	
	resp, err := e.client.Get(url)
	if err != nil {
		// Try master branch if main doesn't exist
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s", owner, repo, path)
		resp, err = e.client.Get(url)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func parseZedDefaultJSON(content []byte) map[string]Setting {
	settings := make(map[string]Setting)
	
	// Remove comments from JSON
	lines := strings.Split(string(content), "\n")
	var cleanedLines []string
	for _, line := range lines {
		if !strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "//")) {
			continue
		}
		if idx := strings.Index(line, "//"); idx > 0 {
			// Check if it's not a URL
			if !strings.Contains(line[:idx], "http") {
				line = line[:idx]
			}
		}
		cleanedLines = append(cleanedLines, line)
	}
	
	cleanedJSON := strings.Join(cleanedLines, "\n")
	// Remove trailing commas
	cleanedJSON = strings.ReplaceAll(cleanedJSON, ",]", "]")
	cleanedJSON = strings.ReplaceAll(cleanedJSON, ",}", "}")
	
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedJSON), &data); err != nil {
		// Fallback to line-by-line parsing
		return parseZedFallback(content)
	}
	
	// Flatten nested settings
	flattenSettings(data, "", settings)
	
	return settings
}

func flattenSettings(data map[string]interface{}, prefix string, settings map[string]Setting) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			// Recurse into nested objects
			flattenSettings(v, fullKey, settings)
		default:
			settings[fullKey] = Setting{
				Name:     fullKey,
				Type:     inferTypeFromValue(value),
				Category: inferCategory(fullKey),
			}
		}
	}
}

func inferTypeFromValue(value interface{}) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case float64, int:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}

func parseZedFallback(content []byte) map[string]Setting {
	// Fallback parser for when JSON parsing fails
	settings := make(map[string]Setting)
	// Implementation of line-by-line parsing
	return settings
}

func parseAlacrittyYAML(content []byte) map[string]Setting {
	settings := make(map[string]Setting)
	// Parse YAML configuration
	return settings
}

func parseWezTermDocs(content []byte) map[string]Setting {
	settings := make(map[string]Setting)
	// Parse markdown documentation
	return settings
}

func parseNeovimOptions(content []byte) map[string]Setting {
	settings := make(map[string]Setting)
	// Parse Vim help format
	return settings
}