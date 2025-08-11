package reference

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ReferenceConfigMapper converts reference configs to app config format
type ReferenceConfigMapper struct {
	loader ConfigLoader
}

// NewReferenceConfigMapper creates a new reference config mapper
func NewReferenceConfigMapper(configDir string) *ReferenceConfigMapper {
	return &ReferenceConfigMapper{
		loader: NewStaticConfigLoader(configDir),
	}
}

// MapReferenceToAppConfig converts reference config to AppConfig format
func (m *ReferenceConfigMapper) MapReferenceToAppConfig(appName string) (*AppConfig, error) {
	ref, err := m.loader.LoadReference(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to load reference config for %s: %w", appName, err)
	}

	appConfig := &AppConfig{
		Name:        ref.AppName,
		Path:        ref.ConfigPath,
		Format:      ref.ConfigType,
		Description: fmt.Sprintf("Auto-generated from reference config for %s", ref.AppName),
		Fields:      make(map[string]FieldConfig),
		Presets:     make(map[string]PresetConfig),
	}

	// Convert each reference setting to a field config
	for key, setting := range ref.Settings {
		fieldConfig := FieldConfig{
			Type:        convertSettingTypeToFieldType(setting.Type),
			Description: setting.Description,
			Default:     setting.DefaultValue,
		}

		// Handle valid values
		if len(setting.ValidValues) > 0 {
			fieldConfig.Values = setting.ValidValues
			// If we have valid values, it's typically a choice field
			if fieldConfig.Type == "string" && len(setting.ValidValues) > 1 {
				fieldConfig.Type = "choice"
			}
		}

		// Set JSON path for nested values (use the setting name as default)
		if setting.Name != "" {
			fieldConfig.Path = setting.Name
		}

		appConfig.Fields[key] = fieldConfig
	}

	// Generate some basic presets based on common patterns
	m.generateBasicPresets(appConfig, ref)

	return appConfig, nil
}

// MergeWithAppConfig merges reference config with existing app config
// App config takes precedence over reference config for existing fields
func (m *ReferenceConfigMapper) MergeWithAppConfig(appName string, existingAppConfig *AppConfig) (*AppConfig, error) {
	// Start with existing app config as base
	mergedConfig := *existingAppConfig

	// Load reference config
	ref, err := m.loader.LoadReference(appName)
	if err != nil {
		// If no reference config exists, return existing app config unchanged
		return existingAppConfig, nil
	}

	// Create a copy of existing fields to avoid modification
	if mergedConfig.Fields == nil {
		mergedConfig.Fields = make(map[string]FieldConfig)
	} else {
		// Deep copy fields map
		newFields := make(map[string]FieldConfig)
		for k, v := range existingAppConfig.Fields {
			newFields[k] = v
		}
		mergedConfig.Fields = newFields
	}

	// Add reference fields that don't exist in app config
	for key, setting := range ref.Settings {
		if _, exists := mergedConfig.Fields[key]; !exists {
			fieldConfig := FieldConfig{
				Type:        convertSettingTypeToFieldType(setting.Type),
				Description: setting.Description,
				Default:     setting.DefaultValue,
			}

			// Handle valid values
			if len(setting.ValidValues) > 0 {
				fieldConfig.Values = setting.ValidValues
				// If we have valid values, it's typically a choice field
				if fieldConfig.Type == "string" && len(setting.ValidValues) > 1 {
					fieldConfig.Type = "choice"
				}
			}

			// Set JSON path for nested values
			if setting.Name != "" {
				fieldConfig.Path = setting.Name
			}

			mergedConfig.Fields[key] = fieldConfig
		}
	}

	// Update metadata from reference if not already set
	if mergedConfig.Path == "" {
		mergedConfig.Path = ref.ConfigPath
	}
	if mergedConfig.Format == "" {
		mergedConfig.Format = ref.ConfigType
	}
	if mergedConfig.Description == "" {
		mergedConfig.Description = fmt.Sprintf("Configuration for %s (enhanced with reference)", ref.AppName)
	}

	return &mergedConfig, nil
}

// GetAvailableApps returns list of apps that have reference configs
func (m *ReferenceConfigMapper) GetAvailableApps() ([]string, error) {
	// List all reference config files
	configDir := ""
	if staticLoader, ok := m.loader.(*StaticConfigLoader); ok {
		configDir = staticLoader.configDir
	}

	if configDir == "" {
		return []string{"ghostty", "zed", "mise"}, nil // fallback to known apps
	}

	files, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list reference configs: %w", err)
	}

	var apps []string
	for _, file := range files {
		appName := strings.TrimSuffix(filepath.Base(file), ".yaml")
		apps = append(apps, appName)
	}

	return apps, nil
}

// convertSettingTypeToFieldType converts reference SettingType to config field type
func convertSettingTypeToFieldType(settingType SettingType) string {
	switch settingType {
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeBoolean:
		return "boolean"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	default:
		return "string" // default fallback
	}
}

// generateBasicPresets creates some basic presets based on reference config
func (m *ReferenceConfigMapper) generateBasicPresets(appConfig *AppConfig, ref *ConfigReference) {
	if appConfig.Presets == nil {
		appConfig.Presets = make(map[string]PresetConfig)
	}

	// Generate a "default" preset with all default values
	defaultValues := make(map[string]interface{})
	for key, setting := range ref.Settings {
		if setting.DefaultValue != nil {
			defaultValues[key] = setting.DefaultValue
		}
	}

	if len(defaultValues) > 0 {
		appConfig.Presets["default"] = PresetConfig{
			Name:        "Default Settings",
			Description: "Default configuration values from reference",
			Values:      defaultValues,
		}
	}

	// App-specific presets based on app name
	switch strings.ToLower(ref.AppName) {
	case "ghostty":
		m.generateGhosttyPresets(appConfig, ref)
	case "zed":
		m.generateZedPresets(appConfig, ref)
	}
}

// generateGhosttyPresets creates Ghostty-specific presets
func (m *ReferenceConfigMapper) generateGhosttyPresets(appConfig *AppConfig, ref *ConfigReference) {
	// Minimal terminal preset
	minimalValues := map[string]interface{}{
		"window-decoration":  false,
		"window-padding":     0,
		"cursor-blink":       false,
		"background-opacity": 1.0,
	}

	appConfig.Presets["minimal"] = PresetConfig{
		Name:        "Minimal Terminal",
		Description: "Clean, distraction-free terminal setup",
		Values:      minimalValues,
	}

	// Developer preset
	devValues := map[string]interface{}{
		"font-family":           "JetBrains Mono",
		"font-size":             14,
		"theme":                 "dark",
		"shell-integration":     true,
		"scrollback-limit":      50000,
		"copy-on-select":        true,
		"confirm-close-surface": true,
	}

	appConfig.Presets["developer"] = PresetConfig{
		Name:        "Developer Setup",
		Description: "Optimized for development work",
		Values:      devValues,
	}
}

// generateZedPresets creates Zed-specific presets
func (m *ReferenceConfigMapper) generateZedPresets(appConfig *AppConfig, ref *ConfigReference) {
	// VS Code-like preset
	vscodeValues := map[string]interface{}{
		"base_keymap":      "VSCode",
		"buffer_font_size": 14,
		"ui_font_size":     14,
		"tab_size":         4,
		"hard_tabs":        false,
		"vim_mode":         false,
		"autosave":         "on_focus_change",
		"format_on_save":   true,
	}

	appConfig.Presets["vscode"] = PresetConfig{
		Name:        "VS Code Style",
		Description: "Configuration similar to VS Code defaults",
		Values:      vscodeValues,
	}

	// Vim-like preset
	vimValues := map[string]interface{}{
		"vim_mode":         true,
		"base_keymap":      "None",
		"buffer_font_size": 12,
		"hard_tabs":        true,
		"tab_size":         8,
		"autosave":         "off",
		"show_whitespaces": "selection",
	}

	appConfig.Presets["vim"] = PresetConfig{
		Name:        "Vim Style",
		Description: "Configuration for Vim users",
		Values:      vimValues,
	}
}

// ValidateReferenceMapping validates that reference config mapping is correct
func (m *ReferenceConfigMapper) ValidateReferenceMapping(appName string) error {
	ref, err := m.loader.LoadReference(appName)
	if err != nil {
		return fmt.Errorf("validation failed - cannot load reference: %w", err)
	}

	appConfig, err := m.MapReferenceToAppConfig(appName)
	if err != nil {
		return fmt.Errorf("validation failed - cannot map to app config: %w", err)
	}

	// Validate that all reference settings are mapped
	for key, setting := range ref.Settings {
		fieldConfig, exists := appConfig.Fields[key]
		if !exists {
			return fmt.Errorf("validation failed - setting %s not mapped", key)
		}

		// Validate type conversion
		expectedType := convertSettingTypeToFieldType(setting.Type)
		if fieldConfig.Type != expectedType && fieldConfig.Type != "choice" {
			return fmt.Errorf("validation failed - setting %s type mismatch: expected %s, got %s",
				key, expectedType, fieldConfig.Type)
		}
	}

	return nil
}
