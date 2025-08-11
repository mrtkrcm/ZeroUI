package reference

import "time"

// ConfigReference represents application configuration metadata
type ConfigReference struct {
	AppName     string                   `json:"app_name" yaml:"app_name"`
	ConfigPath  string                   `json:"config_path" yaml:"config_path"`
	ConfigType  string                   `json:"config_type" yaml:"config_type"` // json, toml, yaml, ini
	LastUpdated time.Time                `json:"last_updated" yaml:"last_updated"`
	Settings    map[string]ConfigSetting `json:"settings" yaml:"settings"`
}

// ConfigSetting represents a single configuration option
type ConfigSetting struct {
	Name         string      `json:"name" yaml:"name"`
	Type         SettingType `json:"type" yaml:"type"`
	Description  string      `json:"description" yaml:"description"`
	DefaultValue interface{} `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	Example      interface{} `json:"example,omitempty" yaml:"example,omitempty"`
	ValidValues  []string    `json:"valid_values,omitempty" yaml:"valid_values,omitempty"`
	Required     bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Category     string      `json:"category,omitempty" yaml:"category,omitempty"`
}

// SettingType simplified to essential types only
type SettingType string

const (
	TypeString  SettingType = "string"
	TypeNumber  SettingType = "number"
	TypeBoolean SettingType = "boolean"
	TypeArray   SettingType = "array"
	TypeObject  SettingType = "object"
)

// ValidationResult for configuration validation
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// ConfigLoader interface for loading configuration references
type ConfigLoader interface {
	LoadReference(appName string) (*ConfigReference, error)
}

// ReferenceManager simplified manager
type ReferenceManager struct {
	loader ConfigLoader
	cache  map[string]*ConfigReference
}

// NewReferenceManager creates a simplified reference manager
func NewReferenceManager(loader ConfigLoader) *ReferenceManager {
	return &ReferenceManager{
		loader: loader,
		cache:  make(map[string]*ConfigReference),
	}
}
