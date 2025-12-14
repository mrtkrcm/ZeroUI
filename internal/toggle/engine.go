package toggle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/recovery"
	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
)

// ConfigLoader interface to support both basic and reference-enhanced loaders
type ConfigLoader interface {
	LoadAppConfig(appName string) (*config.AppConfig, error)
	ListApps() ([]string, error)
	LoadTargetConfig(appConfig *config.AppConfig) (*koanf.Koanf, error)
	SaveTargetConfig(appConfig *config.AppConfig, k *koanf.Koanf) error
}

// Engine handles configuration toggling operations (refactored for better separation of concerns)
type Engine struct {
	// Core services
	configOp   *ConfigOperator
	fieldVal   *FieldValidator
	valueConv  *ValueConverter
	hookRunner *HookRunner
	logger     logger.Logger
	runtime    RuntimeConfig

	// Legacy interface compatibility
	loader ConfigLoader // Keep for interface compatibility
}

// RuntimeConfig captures CLI-configured execution preferences.
type RuntimeConfig struct {
	Verbose bool
	DryRun  bool
}

// NewEngine creates a new toggle engine (backwards compatibility)
func NewEngine() (*Engine, error) {
	// Use reference-enhanced loader for better config coverage
	enhancedLoader, err := config.NewReferenceEnhancedLoader()
	if err != nil {
		// Fallback to basic loader if reference-enhanced fails
		basicLoader, basicErr := config.NewLoader()
		if basicErr != nil {
			return nil, fmt.Errorf("failed to create config loader: %w", basicErr)
		}
		return NewEngineWithDeps(basicLoader, logger.New(logger.DefaultConfig()), RuntimeConfig{}), nil
	}

	return NewEngineWithDeps(enhancedLoader, logger.New(logger.DefaultConfig()), RuntimeConfig{}), nil
}

// NewEngineWithDeps creates a new toggle engine with injected dependencies
func NewEngineWithDeps(configLoader ConfigLoader, log logger.Logger, runtime RuntimeConfig) *Engine {
	if log == nil {
		log = logger.New(logger.DefaultConfig())
	}

	// Initialize service components
	configOp := NewConfigOperator(configLoader, runtime.DryRun)
	fieldVal := NewFieldValidator()
	valueConv := NewValueConverter()
	hookRunner := NewHookRunner(log, runtime.Verbose)

	return &Engine{
		configOp:   configOp,
		fieldVal:   fieldVal,
		valueConv:  valueConv,
		hookRunner: hookRunner,
		logger:     log,
		runtime:    runtime,
		loader:     configLoader, // Keep for interface compatibility
	}
}

// Toggle sets a specific configuration key to a value (refactored to use services)
func (e *Engine) Toggle(appName, key, value string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if e.runtime.Verbose {
		log.Debug("Starting toggle operation", logger.Field{Key: "value", Value: value})
	}

	// 1. Load and validate app config
	appConfig, err := e.configOp.LoadAppConfig(appName)
	if err != nil {
		return err // Error already wrapped by configOp
	}

	// 2. Validate field exists
	if err := e.fieldVal.ValidateFieldExists(appConfig, key); err != nil {
		return err
	}

	// 3. Validate field value
	if err := e.fieldVal.ValidateFieldValue(appConfig, key, value); err != nil {
		return err
	}

	// 4. Get field config and convert value
	fieldConfig, err := e.fieldVal.GetFieldConfig(appConfig, key)
	if err != nil {
		return err
	}

	convertedValue, err := e.valueConv.ConvertValue(value, fieldConfig)
	if err != nil {
		return errors.Wrap(errors.FieldInvalidType, "failed to convert value", err).
			WithApp(appName).WithField(key).WithValue(value)
	}

	// 5. Load target config and set value
	targetConfig, err := e.configOp.LoadTargetConfig(appConfig)
	if err != nil {
		return err // Error already wrapped by configOp
	}

	// Store original config for diff calculation
	originalConfig := targetConfig.All()

	// Set the value using config operator
	if err := e.configOp.SetConfigValue(targetConfig, key, convertedValue, appName); err != nil {
		return err // Error already wrapped by configOp
	}

	// Calculate diff for preview
	differ := configextractor.NewConfigDiffer()
	newConfig := targetConfig.All()
	diff := differ.DiffConfigurations(originalConfig, newConfig)

	if e.runtime.DryRun {
		log.Info("Would set configuration",
			logger.Field{Key: "converted_value", Value: convertedValue},
			logger.Field{Key: "changes", Value: diff.Summary()},
		)

		// Show diff in verbose mode
		if e.runtime.Verbose {
			fmt.Println("\nConfiguration changes preview:")
			fmt.Print(diff.FormatDiff())
		}
		return nil
	}

	// 6. Save configuration safely
	if err := e.configOp.SaveConfigSafely(appConfig, targetConfig); err != nil {
		return err // Error already wrapped by configOp
	}

	log.Success("Configuration updated", logger.Field{Key: "value", Value: value})

	// 7. Run post-toggle hooks
	return e.hookRunner.RunHooks(appConfig, "post-toggle")
}

// Cycle moves to the next value in a field's value list (refactored to use services)
func (e *Engine) Cycle(appName, key string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if e.runtime.Verbose {
		log.Debug("Starting cycle operation")
	}

	// 1. Load and validate app config
	appConfig, err := e.configOp.LoadAppConfig(appName)
	if err != nil {
		return err
	}

	// 2. Validate field exists and get field config
	fieldConfig, err := e.fieldVal.GetFieldConfig(appConfig, key)
	if err != nil {
		return err
	}

	if len(fieldConfig.Values) == 0 {
		return errors.New(errors.FieldInvalidType, "field "+key+" has no predefined values to cycle through")
	}

	// 3. Load current config to get current value
	targetConfig, err := e.configOp.LoadTargetConfig(appConfig)
	if err != nil {
		return err
	}

	currentValue := targetConfig.String(key)

	// 4. Get next value using value converter
	nextValue, err := e.valueConv.GetNextValue(fieldConfig, currentValue)
	if err != nil {
		return err
	}

	// 5. Convert value to appropriate type
	convertedValue, err := e.valueConv.ConvertValue(nextValue, fieldConfig)
	if err != nil {
		return errors.Wrap(errors.FieldInvalidType, "failed to convert cycle value", err).
			WithApp(appName).WithField(key)
	}

	// 6. Set the value using config operator
	if err := e.configOp.SetConfigValue(targetConfig, key, convertedValue, appName); err != nil {
		return err
	}

	if e.runtime.DryRun {
		log.Info("Would cycle configuration",
			logger.Field{Key: "from", Value: currentValue},
			logger.Field{Key: "to", Value: nextValue},
		)
		return nil
	}

	// 7. Save configuration safely
	if err := e.configOp.SaveConfigSafely(appConfig, targetConfig); err != nil {
		return err
	}

	log.Success("Configuration cycled",
		logger.Field{Key: "from", Value: currentValue},
		logger.Field{Key: "to", Value: nextValue},
	)

	// 8. Run post-cycle hooks
	return e.hookRunner.RunHooks(appConfig, "post-cycle")
}

// ApplyPreset applies a preset configuration
func (e *Engine) ApplyPreset(appName, presetName string) error {
	log := e.logger.WithApp(appName).With(logger.Field{Key: "preset", Value: presetName})

	if e.runtime.Verbose {
		log.Debug("Starting preset application")
	}

	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		apps, _ := e.loader.ListApps()
		return errors.NewAppNotFoundError(appName, apps)
	}

	preset, exists := appConfig.Presets[presetName]
	if !exists {
		var availablePresets []string
		for name := range appConfig.Presets {
			availablePresets = append(availablePresets, name)
		}
		return errors.NewPresetNotFoundError(appName, presetName, availablePresets)
	}

	// Load target config
	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return fmt.Errorf("failed to load target config: %w", err)
	}

	// Store original config for diff calculation
	originalConfig := targetConfig.All()

	// Apply all values from the preset
	for key, value := range preset.Values {
		fieldConfig, exists := appConfig.Fields[key]
		if !exists {
			if e.runtime.Verbose {
				log.Warn("Field not found in app config, applying anyway", logger.Field{Key: "field", Value: key})
			}
		}

		// Convert value to appropriate type if field config exists
		convertedValue := value
		if exists {
			convertedValue, err = e.convertValue(fmt.Sprintf("%v", value), fieldConfig.Type)
			if err != nil {
				return fmt.Errorf("failed to convert value for %s: %w", key, err)
			}
		}

		_ = targetConfig.Set(key, convertedValue)
	}

	// Calculate diff for preview
	differ := configextractor.NewConfigDiffer()
	newConfig := targetConfig.All()
	diff := differ.DiffConfigurations(originalConfig, newConfig)

	if e.runtime.DryRun {
		log.Info("Would apply preset",
			logger.Field{Key: "values", Value: preset.Values},
			logger.Field{Key: "changes", Value: diff.Summary()},
		)

		// Show diff in verbose mode
		if e.runtime.Verbose {
			fmt.Println("\nPreset application preview:")
			fmt.Print(diff.FormatDiff())
		}
		return nil
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Success("Preset applied successfully")
	if e.runtime.Verbose {
		log.Debug("Preset values applied", logger.Field{Key: "values", Value: preset.Values})
	}

	// Run post-preset hooks
	return e.runHooks(appConfig, "post-preset")
}

// GetApps returns all available applications for programmatic use
func (e *Engine) GetApps() ([]string, error) {
	return e.loader.ListApps()
}

// ListApps lists all available applications
func (e *Engine) ListApps() error {
	apps, err := e.loader.ListApps()
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		fmt.Println("No applications configured")
		return nil
	}

	fmt.Println("Available applications:")
	for _, app := range apps {
		fmt.Printf("  %s\n", app)
	}

	return nil
}

// ListPresets lists all presets for an application
func (e *Engine) ListPresets(appName string) error {
	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}

	if len(appConfig.Presets) == 0 {
		fmt.Printf("No presets configured for %s\n", appName)
		return nil
	}

	fmt.Printf("Available presets for %s:\n", appName)
	for name, preset := range appConfig.Presets {
		fmt.Printf("  %s", name)
		if preset.Description != "" {
			fmt.Printf(" - %s", preset.Description)
		}
		fmt.Println()
	}

	return nil
}

// GetPresets returns a list of available preset names for an application
func (e *Engine) GetPresets(appName string) ([]string, error) {
	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	var presets []string
	for name := range appConfig.Presets {
		presets = append(presets, name)
	}

	return presets, nil
}

// ListKeys lists all configurable keys for an application
func (e *Engine) ListKeys(appName string) error {
	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}

	if len(appConfig.Fields) == 0 {
		fmt.Printf("No configurable keys for %s\n", appName)
		return nil
	}

	fmt.Printf("Configurable keys for %s:\n", appName)
	for key, field := range appConfig.Fields {
		fmt.Printf("  %s (%s)", key, field.Type)
		if len(field.Values) > 0 {
			fmt.Printf(" - choices: %v", field.Values)
		}
		if field.Description != "" {
			fmt.Printf(" - %s", field.Description)
		}
		fmt.Println()
	}

	return nil
}

// convertValue converts a string value to the appropriate type
// Deprecated: convertValue has been moved to ValueConverter service
func (e *Engine) convertValue(value, fieldType string) (interface{}, error) {
	fieldConfig := &config.FieldConfig{Type: fieldType}
	return e.valueConv.ConvertValue(value, fieldConfig)
}

// Deprecated: runHooks has been moved to HookRunner service
func (e *Engine) runHooks(appConfig *config.AppConfig, hookType string) error {
	return e.hookRunner.RunHooks(appConfig, hookType)
}

// GetAppConfig returns the configuration metadata for an app (for TUI use)
func (e *Engine) GetAppConfig(appName string) (*config.AppConfig, error) {
	return e.loader.LoadAppConfig(appName)
}

// GetCurrentValues returns the current values from the target config file
func (e *Engine) GetCurrentValues(appName string) (map[string]interface{}, error) {
	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load target config: %w", err)
	}

	return targetConfig.All(), nil
}

// validateHookCommand validates that the hook command is safe to execute
func (e *Engine) validateHookCommand(hookCmd string) error {
	// Trim whitespace and check for empty command
	hookCmd = strings.TrimSpace(hookCmd)
	if hookCmd == "" {
		return fmt.Errorf("empty hook command")
	}

	// Check for dangerous characters and patterns - expanded list
	dangerousPatterns := []string{
		"|", "&&", "||", ";", "`", "$", "$(", "${", // Shell operators and command substitution
		"rm -rf", "rm -f", ">/dev/null", "2>&1", // Dangerous operations
		"curl", "wget", "nc", "telnet", "ssh", "scp", // Network operations
		"sudo", "su -", "chmod +x", "chown", "setuid", // Privilege escalation
		"../", "./", "~", "/etc/", "/usr/", "/var/", // Path traversal attempts
		"eval", "exec", "source", "bash -c", "sh -c", // Code execution
		">&", "<&", ">>", "<<", // Redirection operators
		"*", "?", "[", "]", // Glob patterns that could be dangerous
	}

	lowerCmd := strings.ToLower(hookCmd)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return fmt.Errorf("hook command contains dangerous pattern: %s", pattern)
		}
	}

	// Additional validation: no control characters
	for _, r := range hookCmd {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, newline, carriage return
			return fmt.Errorf("hook command contains control character")
		}
	}

	// Allow-list approach: only allow certain safe commands
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty hook command")
	}

	// Strict allow-list - only essential safe commands
	allowedCommands := []string{
		"echo", "printf", "cat", "head", "tail", "wc",
		"grep", "sed", "awk", "sort", "uniq",
		"touch", "mkdir", "ls", "pwd",
		"notify-send", "osascript", // Notification commands
		"date", "sleep", // Time-related safe commands
	}

	// Extract just the command name, handle absolute paths
	command := filepath.Base(parts[0])
	// Also check if it's trying to use a path
	if strings.Contains(parts[0], "/") && parts[0] != command {
		return fmt.Errorf("hook command cannot use absolute or relative paths")
	}

	for _, allowed := range allowedCommands {
		if command == allowed {
			return nil
		}
	}

	return fmt.Errorf("hook command '%s' is not in the allowed list", command)
}

// setEnvironmentVariables safely sets environment variables
func (e *Engine) setEnvironmentVariables(envVars map[string]string) error {
	// Prevent setting dangerous environment variables
	dangerousVars := []string{
		"PATH", "LD_LIBRARY_PATH", "LD_PRELOAD",
		"HOME", "USER", "SHELL", "IFS",
	}

	for key, value := range envVars {
		// Validate environment variable name
		if key == "" {
			return fmt.Errorf("empty environment variable name")
		}

		// Check against dangerous variables
		for _, dangerous := range dangerousVars {
			if strings.ToUpper(key) == dangerous {
				return fmt.Errorf("setting dangerous environment variable '%s' is not allowed", key)
			}
		}

		// Validate value doesn't contain dangerous characters
		if strings.Contains(value, "`") || strings.Contains(value, "$") {
			return fmt.Errorf("environment variable value contains dangerous characters")
		}

		// Set the environment variable (scoped to this process)
		_ = os.Setenv(key, value)
	}

	return nil
}

// Deprecated: expandPath has been moved to ConfigOperator service
func (e *Engine) expandPath(path string) string {
	return e.configOp.expandPath(path)
}

// ghosttyValidatorAdapter adapts the ghostty validation system to the recovery ConfigValidator interface
type ghosttyValidatorAdapter struct {
	appName string
	engine  *Engine
}

// ValidateConfig validates the configuration using ghostty-specific validation rules
func (v *ghosttyValidatorAdapter) ValidateConfig(config map[string]interface{}) recovery.ValidationResult {
	// For ghostty configs, use the schema validator
	if strings.ToLower(v.appName) == "ghostty" {
		validator := configextractor.NewGhosttySchemaValidator()
		result := validator.ValidateConfig(config)

		// Convert the result format
		var errors []string
		if !result.Valid {
			errors = result.Errors
		}

		return recovery.ValidationResult{
			Valid:  result.Valid,
			Errors: errors,
		}
	}

	// For other apps, use basic validation (could be extended)
	return recovery.ValidationResult{Valid: true}
}
