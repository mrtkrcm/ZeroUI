package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mrtkrcm/ZeroUI/pkg/reference"
)

// ReferenceEnhancedLoader extends the basic loader with reference config integration
type ReferenceEnhancedLoader struct {
	*Loader         // Embed the base loader
	referenceMapper *reference.ReferenceConfigMapper
	referenceMutex  sync.RWMutex // Thread-safe access to reference configs
}

// NewReferenceEnhancedLoader creates a loader that merges reference and app configs
func NewReferenceEnhancedLoader() (*ReferenceEnhancedLoader, error) {
	baseLoader, err := NewLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create base loader: %w", err)
	}

	// Determine the configs directory (relative to the executable or project root)
	configsDir := "configs"
	if _, err := os.Stat(configsDir); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			"./configs",
			"../configs",
			"../../configs",
			filepath.Join(os.Getenv("CONFIGTOGGLE_ROOT"), "configs"),
		}

		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				configsDir = alt
				found = true
				break
			}
		}

		if !found {
			// Create a basic configs directory as fallback
			if err := os.MkdirAll("configs", 0755); err != nil {
				return nil, fmt.Errorf("failed to create configs directory: %w", err)
			}
			configsDir = "configs"
		}
	}

	referenceMapper := reference.NewReferenceConfigMapper(configsDir)

	return &ReferenceEnhancedLoader{
		Loader:          baseLoader,
		referenceMapper: referenceMapper,
	}, nil
}

// convertReferenceAppConfig converts reference.AppConfig to config.AppConfig
func convertReferenceAppConfig(refConfig *reference.AppConfig) *AppConfig {
	config := &AppConfig{
		Name:        refConfig.Name,
		Path:        refConfig.Path,
		Format:      refConfig.Format,
		Description: refConfig.Description,
		Fields:      make(map[string]FieldConfig),
		Presets:     make(map[string]PresetConfig),
		Hooks:       refConfig.Hooks,
		Env:         refConfig.Env,
	}

	// Convert fields
	for key, refField := range refConfig.Fields {
		config.Fields[key] = FieldConfig{
			Type:        refField.Type,
			Values:      refField.Values,
			Default:     refField.Default,
			Description: refField.Description,
			Path:        refField.Path,
		}
	}

	// Convert presets
	for key, refPreset := range refConfig.Presets {
		config.Presets[key] = PresetConfig{
			Name:        refPreset.Name,
			Description: refPreset.Description,
			Values:      refPreset.Values,
		}
	}

	return config
}

// convertToReferenceAppConfig converts config.AppConfig to reference.AppConfig
func convertToReferenceAppConfig(config *AppConfig) *reference.AppConfig {
	refConfig := &reference.AppConfig{
		Name:        config.Name,
		Path:        config.Path,
		Format:      config.Format,
		Description: config.Description,
		Fields:      make(map[string]reference.FieldConfig),
		Presets:     make(map[string]reference.PresetConfig),
		Hooks:       config.Hooks,
		Env:         config.Env,
	}

	// Convert fields
	for key, field := range config.Fields {
		refConfig.Fields[key] = reference.FieldConfig{
			Type:        field.Type,
			Values:      field.Values,
			Default:     field.Default,
			Description: field.Description,
			Path:        field.Path,
		}
	}

	// Convert presets
	for key, preset := range config.Presets {
		refConfig.Presets[key] = reference.PresetConfig{
			Name:        preset.Name,
			Description: preset.Description,
			Values:      preset.Values,
		}
	}

	return refConfig
}

// LoadAppConfigWithReference loads app config enhanced with reference config data
func (l *ReferenceEnhancedLoader) LoadAppConfigWithReference(appName string) (*AppConfig, error) {
	l.referenceMutex.RLock()
	defer l.referenceMutex.RUnlock()

	// Try to load existing app config first
	appConfig, appErr := l.Loader.LoadAppConfig(appName)

	// If app config doesn't exist, try to generate from reference
	if appErr != nil {
		referenceConfig, refErr := l.referenceMapper.MapReferenceToAppConfig(appName)
		if refErr != nil {
			// Neither app config nor reference config exists
			return nil, fmt.Errorf("no config found for %s (app config error: %v, reference config error: %v)",
				appName, appErr, refErr)
		}

		// Convert reference config to config package format and return
		return convertReferenceAppConfig(referenceConfig), nil
	}

	// Merge existing app config with reference config
	refAppConfig := convertToReferenceAppConfig(appConfig)
	mergedRefConfig, err := l.referenceMapper.MergeWithAppConfig(appName, refAppConfig)
	if err != nil {
		// If merge fails, return original app config
		return appConfig, nil
	}

	// Convert merged config back to config package format
	return convertReferenceAppConfig(mergedRefConfig), nil
}

// LoadAppConfig overrides base method to use reference-enhanced loading
func (l *ReferenceEnhancedLoader) LoadAppConfig(appName string) (*AppConfig, error) {
	return l.LoadAppConfigWithReference(appName)
}

// ListAppsWithReference returns apps from both sources
func (l *ReferenceEnhancedLoader) ListAppsWithReference() ([]string, error) {
	l.referenceMutex.RLock()
	defer l.referenceMutex.RUnlock()

	// Get apps from base loader
	appConfigApps, err := l.Loader.ListApps()
	if err != nil {
		appConfigApps = []string{} // Continue with empty list
	}

	// Get apps from reference configs
	referenceApps, err := l.referenceMapper.GetAvailableApps()
	if err != nil {
		referenceApps = []string{} // Continue with empty list
	}

	// Merge and deduplicate
	appSet := make(map[string]bool)
	var allApps []string

	// Add app config apps
	for _, app := range appConfigApps {
		if !appSet[app] {
			appSet[app] = true
			allApps = append(allApps, app)
		}
	}

	// Add reference config apps
	for _, app := range referenceApps {
		if !appSet[app] {
			appSet[app] = true
			allApps = append(allApps, app)
		}
	}

	return allApps, nil
}

// ListApps overrides base method to use reference-enhanced listing
func (l *ReferenceEnhancedLoader) ListApps() ([]string, error) {
	return l.ListAppsWithReference()
}

// GetReferenceMapper returns the reference mapper for direct access
func (l *ReferenceEnhancedLoader) GetReferenceMapper() *reference.ReferenceConfigMapper {
	return l.referenceMapper
}

// ValidateReferenceIntegration validates that reference integration is working
func (l *ReferenceEnhancedLoader) ValidateReferenceIntegration() error {
	apps, err := l.ListAppsWithReference()
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	if len(apps) == 0 {
		return fmt.Errorf("no apps found from either source")
	}

	// Test loading one app to ensure integration works
	for _, app := range apps {
		_, err := l.LoadAppConfigWithReference(app)
		if err != nil {
			continue // Try next app
		}

		// Successfully loaded at least one app
		return nil
	}

	return fmt.Errorf("failed to load any app config with reference integration")
}

// RefreshReferenceCache clears reference-related caches
func (l *ReferenceEnhancedLoader) RefreshReferenceCache() {
	l.referenceMutex.Lock()
	defer l.referenceMutex.Unlock()

	// Clear the base loader cache
	l.ClearCache()

	// Note: The reference mapper doesn't have its own cache currently,
	// but if it did, we would clear it here as well
}

// GetConfigSource returns information about where the config came from
func (l *ReferenceEnhancedLoader) GetConfigSource(appName string) (string, error) {
	// Check if app config exists
	_, appErr := l.Loader.LoadAppConfig(appName)

	// Check if reference config exists
	_, refErr := l.referenceMapper.MapReferenceToAppConfig(appName)

	if appErr == nil && refErr == nil {
		return "merged", nil // Both sources available
	} else if appErr == nil {
		return "app_config", nil // Only app config
	} else if refErr == nil {
		return "reference", nil // Only reference config
	} else {
		return "none", fmt.Errorf("no config source found for %s", appName)
	}
}

// GetReferenceConfigInfo returns metadata about the reference config for an app
func (l *ReferenceEnhancedLoader) GetReferenceConfigInfo(appName string) (map[string]interface{}, error) {
	l.referenceMutex.RLock()
	defer l.referenceMutex.RUnlock()

	// This would require extending the reference mapper to expose more metadata
	// For now, return basic info
	info := make(map[string]interface{})

	source, err := l.GetConfigSource(appName)
	if err != nil {
		return nil, err
	}

	info["config_source"] = source
	info["app_name"] = appName

	return info, nil
}
