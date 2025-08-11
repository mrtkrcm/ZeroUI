package reference

// AppConfig represents the configuration for a single application (copy from config package to avoid import cycle)
type AppConfig struct {
	Name        string                  `yaml:"name"`
	Path        string                  `yaml:"path"`
	Format      string                  `yaml:"format"`
	Description string                  `yaml:"description,omitempty"`
	Fields      map[string]FieldConfig  `yaml:"fields"`
	Presets     map[string]PresetConfig `yaml:"presets"`
	Hooks       map[string]string       `yaml:"hooks,omitempty"`
	Env         map[string]string       `yaml:"env,omitempty"`
}

// FieldConfig represents a configurable field (copy from config package to avoid import cycle)
type FieldConfig struct {
	Type        string      `yaml:"type"` // choice, string, number, boolean
	Values      []string    `yaml:"values,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Path        string      `yaml:"path,omitempty"` // JSON path for nested values
}

// PresetConfig represents a preset configuration (copy from config package to avoid import cycle)
type PresetConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Values      map[string]interface{} `yaml:"values"`
}
