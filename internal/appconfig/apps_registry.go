package appconfig

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AppDefinition represents a known application in the registry
type AppDefinition struct {
	Name         string   `yaml:"name"`
	DisplayName  string   `yaml:"display_name"`
	Icon         string   `yaml:"icon"`
	Description  string   `yaml:"description"`
	Category     string   `yaml:"category"`
	ConfigPaths  []string `yaml:"config_paths"`
	ConfigFormat string   `yaml:"config_format"`
}

// CategoryDefinition represents an app category
type CategoryDefinition struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Icon        string `yaml:"icon"`
}

// AppsRegistry contains all known applications and categories
type AppsRegistry struct {
	Applications []AppDefinition      `yaml:"applications"`
	Categories   []CategoryDefinition `yaml:"categories"`

	// Internal maps for quick lookup
	appsByName     map[string]*AppDefinition
	appsByCategory map[string][]*AppDefinition
}

//go:embed apps_registry.yaml
var defaultRegistry string

// LoadAppsRegistry loads the apps registry from embedded and custom locations
func LoadAppsRegistry() (*AppsRegistry, error) {
	// Start with embedded registry
	registry, err := loadRegistryFromString(defaultRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded registry: %w", err)
	}

	// Try to load custom apps from user config
	home, err := os.UserHomeDir()
	if err == nil {
		// Check for custom apps.yaml to add/override apps
		customAppsPath := filepath.Join(home, ".config", "zeroui", "apps.yaml")
		if _, err := os.Stat(customAppsPath); err == nil {
			if err := registry.MergeFromFile(customAppsPath); err != nil {
				// Log error but don't fail - embedded registry is still valid
				fmt.Fprintf(os.Stderr, "Warning: failed to load custom apps: %v\n", err)
			}
		}

		// Also check for full registry override (advanced users)
		customRegistryPath := filepath.Join(home, ".config", "zeroui", "apps_registry.yaml")
		if _, err := os.Stat(customRegistryPath); err == nil {
			// This completely replaces the embedded registry
			return loadRegistryFromFile(customRegistryPath)
		}
	}

	return registry, nil
}

// LoadAppsRegistryFromFile loads registry from a specific file
func LoadAppsRegistryFromFile(path string) (*AppsRegistry, error) {
	return loadRegistryFromFile(path)
}

// loadRegistryFromFile loads registry from a file
func loadRegistryFromFile(path string) (*AppsRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	return loadRegistryFromBytes(data)
}

// loadRegistryFromString loads registry from a string
func loadRegistryFromString(data string) (*AppsRegistry, error) {
	return loadRegistryFromBytes([]byte(data))
}

// loadRegistryFromBytes loads registry from bytes
func loadRegistryFromBytes(data []byte) (*AppsRegistry, error) {
	var registry AppsRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	// Build internal maps
	registry.appsByName = make(map[string]*AppDefinition)
	registry.appsByCategory = make(map[string][]*AppDefinition)

	for i := range registry.Applications {
		app := &registry.Applications[i]
		registry.appsByName[app.Name] = app

		if app.Category != "" {
			registry.appsByCategory[app.Category] = append(
				registry.appsByCategory[app.Category],
				app,
			)
		}
	}

	return &registry, nil
}

// GetApp returns an app definition by name
func (r *AppsRegistry) GetApp(name string) (*AppDefinition, bool) {
	app, ok := r.appsByName[name]
	return app, ok
}

// GetAppsByCategory returns all apps in a category
func (r *AppsRegistry) GetAppsByCategory(category string) []*AppDefinition {
	return r.appsByCategory[category]
}

// GetCategories returns all categories
func (r *AppsRegistry) GetCategories() []CategoryDefinition {
	return r.Categories
}

// GetAllApps returns all registered applications
func (r *AppsRegistry) GetAllApps() []AppDefinition {
	return r.Applications
}

// FindConfigPath finds the actual config file path for an app
func (r *AppsRegistry) FindConfigPath(appName string) (string, bool) {
	app, ok := r.appsByName[appName]
	if !ok {
		return "", false
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	for _, path := range app.ConfigPaths {
		expandedPath := strings.ReplaceAll(path, "~", home)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, true
		}
	}

	return "", false
}

// CheckAppStatus checks if an app's config exists
func (r *AppsRegistry) CheckAppStatus(appName string) (exists bool, path string) {
	path, exists = r.FindConfigPath(appName)
	return exists, path
}

// GetConfigPaths returns expanded config paths for an app
func (r *AppsRegistry) GetConfigPaths(appName string) []string {
	app, ok := r.appsByName[appName]
	if !ok {
		return nil
	}

	home, _ := os.UserHomeDir()
	paths := make([]string, len(app.ConfigPaths))

	for i, path := range app.ConfigPaths {
		paths[i] = strings.ReplaceAll(path, "~", home)
	}

	return paths
}

// MergeFromFile merges applications from another YAML file
func (r *AppsRegistry) MergeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read custom apps file: %w", err)
	}

	// Parse custom apps
	var custom struct {
		Applications []AppDefinition      `yaml:"applications"`
		Categories   []CategoryDefinition `yaml:"categories,omitempty"`
	}

	if err := yaml.Unmarshal(data, &custom); err != nil {
		return fmt.Errorf("failed to parse custom apps: %w", err)
	}

	// Merge or override applications
	for _, app := range custom.Applications {
		// Check if app already exists
		if existing, ok := r.appsByName[app.Name]; ok {
			// Override existing app
			*existing = app
		} else {
			// Add new app
			r.Applications = append(r.Applications, app)
			r.appsByName[app.Name] = &r.Applications[len(r.Applications)-1]
		}

		// Update category mapping
		if app.Category != "" {
			// Remove from old category if it exists elsewhere
			for cat, apps := range r.appsByCategory {
				for i, a := range apps {
					if a.Name == app.Name && cat != app.Category {
						r.appsByCategory[cat] = append(apps[:i], apps[i+1:]...)
						break
					}
				}
			}
			// Add to new category
			found := false
			for _, a := range r.appsByCategory[app.Category] {
				if a.Name == app.Name {
					found = true
					break
				}
			}
			if !found {
				r.appsByCategory[app.Category] = append(
					r.appsByCategory[app.Category],
					r.appsByName[app.Name],
				)
			}
		}
	}

	// Merge categories if provided
	for _, cat := range custom.Categories {
		found := false
		for i, existing := range r.Categories {
			if existing.Name == cat.Name {
				r.Categories[i] = cat
				found = true
				break
			}
		}
		if !found {
			r.Categories = append(r.Categories, cat)
		}
	}

	return nil
}

// EmbeddedRegistry returns the default embedded registry
func EmbeddedRegistry() string {
	return defaultRegistry
}
