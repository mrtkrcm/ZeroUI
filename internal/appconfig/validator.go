package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Validator provides configuration validation
type Validator struct {
	rules []ValidationRule
}

// ValidationRule defines a validation check
type ValidationRule struct {
	Name     string
	Check    func(interface{}) error
	Required bool
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []string
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Rule    string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	v := &Validator{
		rules: []ValidationRule{},
	}

	// Add default rules
	v.addDefaultRules()

	return v
}

// addDefaultRules adds standard validation rules
func (v *Validator) addDefaultRules() {
	// Path validation
	v.AddRule("path_exists", func(val interface{}) error {
		path, ok := val.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}

		// Expand home directory
		if strings.HasPrefix(path, "~") {
			home, _ := os.UserHomeDir()
			path = strings.Replace(path, "~", home, 1)
		}

		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("path does not exist: %s", path)
		}

		return nil
	}, false)

	// Format validation
	v.AddRule("valid_format", func(val interface{}) error {
		format, ok := val.(string)
		if !ok {
			return fmt.Errorf("format must be a string")
		}

		validFormats := []string{"yaml", "json", "toml", "ini", "custom", "lua", "shell"}
		for _, valid := range validFormats {
			if format == valid {
				return nil
			}
		}

		return fmt.Errorf("invalid format: %s", format)
	}, false)

	// Name validation
	v.AddRule("valid_name", func(val interface{}) error {
		name, ok := val.(string)
		if !ok {
			return fmt.Errorf("name must be a string")
		}

		if len(name) == 0 {
			return fmt.Errorf("name cannot be empty")
		}

		// Check for valid characters (alphanumeric, dash, underscore)
		validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		if !validName.MatchString(name) {
			return fmt.Errorf("name contains invalid characters: %s", name)
		}

		return nil
	}, true)
}

// AddRule adds a validation rule
func (v *Validator) AddRule(name string, check func(interface{}) error, required bool) {
	v.rules = append(v.rules, ValidationRule{
		Name:     name,
		Check:    check,
		Required: required,
	})
}

// ValidateAppDefinition validates an app definition
func (v *Validator) ValidateAppDefinition(app *AppDefinition) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Validate name
	if err := v.validateField("name", app.Name, "valid_name"); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: err.Error(),
			Rule:    "valid_name",
		})
	}

	// Validate display name
	if app.DisplayName == "" {
		result.Warnings = append(result.Warnings, "display_name is empty, using name")
		app.DisplayName = app.Name
	}

	// Validate icon
	if app.Icon == "" {
		app.Icon = "○" // Default icon
		result.Warnings = append(result.Warnings, "icon is empty, using default ○")
	}

	// Validate config paths
	if len(app.ConfigPaths) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "config_paths",
			Message: "at least one config path is required",
			Rule:    "required",
		})
	} else {
		for i, path := range app.ConfigPaths {
			if err := v.validatePath(path); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("config_paths[%d]: %v", i, err))
			}
		}
	}

	// Validate format
	if app.ConfigFormat != "" {
		if err := v.validateField("config_format", app.ConfigFormat, "valid_format"); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("config_format: %v", err))
		}
	}

	return result
}

// ValidateRegistry validates an entire registry
func (v *Validator) ValidateRegistry(registry *AppsRegistry) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Check for duplicate app names
	seen := make(map[string]bool)
	for _, app := range registry.Applications {
		if seen[app.Name] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "applications",
				Message: fmt.Sprintf("duplicate app name: %s", app.Name),
				Rule:    "unique",
			})
		}
		seen[app.Name] = true

		// Validate each app
		appResult := v.ValidateAppDefinition(&app)
		if !appResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, appResult.Errors...)
		}
		result.Warnings = append(result.Warnings, appResult.Warnings...)
	}

	// Validate categories
	seenCat := make(map[string]bool)
	for _, cat := range registry.Categories {
		if seenCat[cat.Name] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "categories",
				Message: fmt.Sprintf("duplicate category name: %s", cat.Name),
				Rule:    "unique",
			})
		}
		seenCat[cat.Name] = true
	}

	// Check that all app categories exist
	for _, app := range registry.Applications {
		if app.Category != "" && !seenCat[app.Category] {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("app %s references undefined category: %s",
					app.Name, app.Category))
		}
	}

	return result
}

// ValidateConfigFile validates a configuration file
func (v *Validator) ValidateConfigFile(path string) error {
	// Check file exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check it's not a directory
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Check file is readable
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	file.Close()

	// Check file size (warn if too large)
	if info.Size() > 10*1024*1024 { // 10MB
		return fmt.Errorf("file is too large: %d bytes", info.Size())
	}

	// Validate format based on extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return v.validateYAMLFile(path)
	case ".json":
		return v.validateJSONFile(path)
	case ".toml":
		return v.validateTOMLFile(path)
	default:
		// Unknown format, just check it's text
		return v.validateTextFile(path)
	}
}

// Private methods

func (v *Validator) validateField(name string, value interface{}, ruleName string) error {
	for _, rule := range v.rules {
		if rule.Name == ruleName {
			return rule.Check(value)
		}
	}
	return nil
}

func (v *Validator) validatePath(path string) error {
	// Don't validate existence, just format
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// Check for common issues
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains relative parent reference")
	}

	return nil
}

func (v *Validator) validateYAMLFile(path string) error {
	// For now, just check if it can be read
	// Could add YAML parsing validation here
	return nil
}

func (v *Validator) validateJSONFile(path string) error {
	// For now, just check if it can be read
	// Could add JSON parsing validation here
	return nil
}

func (v *Validator) validateTOMLFile(path string) error {
	// For now, just check if it can be read
	// Could add TOML parsing validation here
	return nil
}

func (v *Validator) validateTextFile(path string) error {
	// Check if file contains binary data
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Check for null bytes
	for i, b := range data {
		if b == 0 {
			return fmt.Errorf("binary content detected at byte %d", i)
		}
	}

	return nil
}
