package reference

import (
	"fmt"
	"strings"
)

// GetReference gets a configuration reference (with caching)
func (rm *ReferenceManager) GetReference(appName string) (*ConfigReference, error) {
	// Check cache first
	if ref, exists := rm.cache[appName]; exists {
		return ref, nil
	}

	// Load from source
	ref, err := rm.loader.LoadReference(appName)
	if err != nil {
		return nil, err
	}

	// Cache and return
	rm.cache[appName] = ref
	return ref, nil
}

// ValidateConfiguration validates a configuration value
func (rm *ReferenceManager) ValidateConfiguration(appName, settingName string, value interface{}) (*ValidationResult, error) {
	ref, err := rm.GetReference(appName)
	if err != nil {
		return nil, err
	}

	setting, exists := ref.Settings[settingName]
	if !exists {
		return &ValidationResult{
			Valid:       false,
			Errors:      []string{fmt.Sprintf("Unknown setting: %s", settingName)},
			Suggestions: findSimilarSettings(ref, settingName),
		}, nil
	}

	return validateSetting(setting, value), nil
}

// validateSetting validates a value against a setting definition
func validateSetting(setting ConfigSetting, value interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Type validation
	if !isValidType(setting.Type, value) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Expected type %s, got %T", setting.Type, value))
	}

	// Valid values validation
	if len(setting.ValidValues) > 0 {
		strValue := fmt.Sprintf("%v", value)
		valid := false
		for _, validValue := range setting.ValidValues {
			if validValue == strValue {
				valid = true
				break
			}
		}
		if !valid {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Value must be one of: %v", setting.ValidValues))
		}
	}

	return result
}

// isValidType checks if value matches the expected type
func isValidType(expectedType SettingType, value interface{}) bool {
	switch expectedType {
	case TypeString:
		_, ok := value.(string)
		return ok
	case TypeNumber:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		}
		return false
	case TypeBoolean:
		_, ok := value.(bool)
		return ok
	case TypeArray:
		// Accept slices and arrays
		switch value.(type) {
		case []interface{}, []string, []int, []float64, []bool:
			return true
		}
		return false
	case TypeObject:
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true // Unknown types are considered valid
	}
}

// findSimilarSettings finds settings with similar names
func findSimilarSettings(ref *ConfigReference, settingName string) []string {
	var suggestions []string
	targetLower := strings.ToLower(settingName)

	for name := range ref.Settings {
		nameLower := strings.ToLower(name)

		// Simple similarity: contains or is contained
		if strings.Contains(targetLower, nameLower) || strings.Contains(nameLower, targetLower) {
			suggestions = append(suggestions, name)
			if len(suggestions) >= 3 { // Limit suggestions
				break
			}
		}
	}

	return suggestions
}

// SearchSettings searches for settings matching a query
func (rm *ReferenceManager) SearchSettings(appName, query string) ([]ConfigSetting, error) {
	ref, err := rm.GetReference(appName)
	if err != nil {
		return nil, err
	}

	var results []ConfigSetting
	queryLower := strings.ToLower(query)

	for _, setting := range ref.Settings {
		if strings.Contains(strings.ToLower(setting.Name), queryLower) ||
			strings.Contains(strings.ToLower(setting.Description), queryLower) ||
			strings.Contains(strings.ToLower(setting.Category), queryLower) {
			results = append(results, setting)
		}
	}

	return results, nil
}

// ListApps returns available applications
func (rm *ReferenceManager) ListApps() ([]string, error) {
	var apps []string

	// This would scan the config directory for available files
	// For now, return known apps
	knownApps := []string{"ghostty", "zed", "mise"}

	for _, app := range knownApps {
		if _, err := rm.GetReference(app); err == nil {
			apps = append(apps, app)
		}
	}

	return apps, nil
}
