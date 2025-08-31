package forms

// Essential types for TUI forms

// ConfigField represents a configuration field
type ConfigField struct {
	Key         string          // Field identifier
	Value       interface{}     // Current value
	Type        ConfigFieldType // Field type (string, bool, number)
	Description string          // Human-readable description
	Default     interface{}     // Default value

	// Backward compatibility fields
	Required bool     // Whether field is required
	IsSet    bool     // Whether field has been set
	Source   string   // Source of the field value
	Options  []string // Available options for select fields
}

// ValidationResult represents the result of field validation
type ValidationResult struct {
	Valid   bool     // Whether validation passed
	Message string   // Validation message
	Errors  []string // List of validation errors
}

// KeyMap defines key bindings for forms
type KeyMap struct {
	Up     []string // Keys for moving up
	Down   []string // Keys for moving down
	Enter  []string // Keys for selecting/editing
	Escape []string // Keys for canceling
	Help   []string // Keys for showing help
	Quit   []string // Keys for quitting
}
