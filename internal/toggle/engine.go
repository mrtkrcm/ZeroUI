package toggle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/recovery"
	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
	"github.com/spf13/viper"
)

// ConfigLoader interface to support both basic and reference-enhanced loaders
type ConfigLoader interface {
	LoadAppConfig(appName string) (*appconfig.AppConfig, error)
	ListApps() ([]string, error)
	LoadTargetConfig(appConfig *appconfig.AppConfig) (*koanf.Koanf, error)
	SaveTargetConfig(appConfig *appconfig.AppConfig, k *koanf.Koanf) error
}

// Engine handles configuration toggling operations
type Engine struct {
	loader    ConfigLoader
	logger    *logger.Logger
	homeDir   string                     // Cache for home directory
	pathCache *lru.Cache[string, string] // LRU cache for expanded paths (prevents memory leak)
	pathMutex sync.RWMutex               // Thread-safe access to pathCache
}

// NewEngine creates a new toggle engine (backwards compatibility)
func NewEngine() (*Engine, error) {
	// Use reference-enhanced loader for better config coverage
	enhancedLoader, err := appconfig.NewReferenceEnhancedLoader()
	if err != nil {
		// Fallback to basic loader if reference-enhanced fails
		basicLoader, basicErr := appconfig.NewLoader()
		if basicErr != nil {
			return nil, fmt.Errorf("failed to create config loader: %w", basicErr)
		}
		// Use basic loader as ConfigLoader interface
		var loader ConfigLoader = basicLoader
		homeDir, _ := os.UserHomeDir()
		pathCache, _ := lru.New[string, string](1000)
		return &Engine{
			loader:    loader,
			logger:    logger.Global(),
			homeDir:   homeDir,
			pathCache: pathCache,
		}, nil
	}

	// Use enhanced loader as ConfigLoader interface
	var loader ConfigLoader = enhancedLoader
	homeDir, _ := os.UserHomeDir()
	pathCache, _ := lru.New[string, string](1000) // 1000 entry limit prevents memory leak
	return &Engine{
		loader:    loader,
		logger:    logger.Global(), // Use global logger for backwards compatibility
		homeDir:   homeDir,
		pathCache: pathCache,
	}, nil
}

// NewEngineWithDeps creates a new toggle engine with injected dependencies
func NewEngineWithDeps(configLoader ConfigLoader, log *logger.Logger) *Engine {
	homeDir, _ := os.UserHomeDir()
	pathCache, _ := lru.New[string, string](1000) // 1000 entry limit prevents memory leak
	return &Engine{
		loader: configLoader,
		logger: func() *logger.Logger {
			if log != nil {
				return log
			}
			return logger.New(logger.DefaultConfig())
		}(),
		homeDir:   homeDir,
		pathCache: pathCache,
	}
}

// Toggle sets a specific configuration key to a value
func (e *Engine) Toggle(appName, key, value string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if viper.GetBool("verbose") {
		log.Debug("Starting toggle operation", map[string]interface{}{
			"value": value,
		})
	}

	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		// Check if it's an app not found error
		apps, _ := e.loader.ListApps()
		return errors.NewAppNotFoundError(appName, apps)
	}

	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		var availableFields []string
		for field := range appConfig.Fields {
			availableFields = append(availableFields, field)
		}
		return errors.NewFieldNotFoundError(appName, key, availableFields)
	}

	// Validate the value if choices are defined
	if len(fieldConfig.Values) > 0 {
		valid := false
		for _, validValue := range fieldConfig.Values {
			if validValue == value {
				valid = true
				break
			}
		}
		if !valid {
			return errors.NewInvalidValueError(appName, key, value, fieldConfig.Values)
		}
	}

	// Convert value to appropriate type
	convertedValue, err := e.convertValue(value, fieldConfig.Type)
	if err != nil {
		return errors.Wrap(errors.FieldInvalidType, "failed to convert value", err).
			WithApp(appName).WithField(key).WithValue(value)
	}

	// Load target config
	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return errors.Wrap(errors.ConfigParseError, "failed to load target config", err).
			WithApp(appName).
			WithSuggestions("Check if the config file exists and is readable")
	}

	// Set the value
	if err := targetConfig.Set(key, convertedValue); err != nil {
		return errors.Wrap(errors.ConfigWriteError, "failed to set config value", err).
			WithApp(appName).WithField(key).WithValue(value)
	}

	if viper.GetBool("dry-run") {
		log.Info("Would set configuration", map[string]interface{}{
			"converted_value": convertedValue,
		})
		return nil
	}

	// Create safe operation with automatic backup
	configPath := e.expandPath(appConfig.Path)

	safeOp, err := recovery.NewSafeOperation(configPath, appName)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to create backup", err).
			WithApp(appName)
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		// Rollback on failure
		if rollbackErr := safeOp.Rollback(); rollbackErr != nil {
			log.Error("Failed to rollback changes", rollbackErr)
		}
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appName).
			WithSuggestions("Check file permissions and disk space", "Configuration has been rolled back")
	}

	// Commit the operation (remove backup)
	if err := safeOp.Commit(); err != nil {
		log.Error("Failed to cleanup backup", err)
	}

	// Cleanup old backups
	if err := safeOp.Cleanup(5); err != nil {
		log.Error("Failed to cleanup old backups", err)
	}

	log.Success("Configuration updated", map[string]interface{}{
		"value": value,
	})

	// Run post-toggle hooks
	return e.runHooks(appConfig, "post-toggle")
}

// Cycle moves to the next value in a field's value list
func (e *Engine) Cycle(appName, key string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if viper.GetBool("verbose") {
		log.Debug("Starting cycle operation")
	}

	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}

	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		return fmt.Errorf("field %s not found for app %s", key, appName)
	}

	if len(fieldConfig.Values) == 0 {
		return fmt.Errorf("field %s has no predefined values to cycle through", key)
	}

	// Load current config to get current value
	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return fmt.Errorf("failed to load target config: %w", err)
	}

	currentValue := targetConfig.String(key)

	// Find current value index
	currentIndex := -1
	for i, value := range fieldConfig.Values {
		if value == currentValue {
			currentIndex = i
			break
		}
	}

	// Get next value (wrap around)
	nextIndex := (currentIndex + 1) % len(fieldConfig.Values)
	nextValue := fieldConfig.Values[nextIndex]

	// Convert value to appropriate type
	convertedValue, err := e.convertValue(nextValue, fieldConfig.Type)
	if err != nil {
		return fmt.Errorf("failed to convert value: %w", err)
	}

	// Set the value
	_ = targetConfig.Set(key, convertedValue)

	if viper.GetBool("dry-run") {
		log.Info("Would cycle configuration", map[string]interface{}{
			"from": currentValue,
			"to":   nextValue,
		})
		return nil
	}

	// Create safe operation with automatic backup
	configPath := e.expandPath(appConfig.Path)

	safeOp, err := recovery.NewSafeOperation(configPath, appName)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to create backup", err).
			WithApp(appName)
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		// Rollback on failure
		if rollbackErr := safeOp.Rollback(); rollbackErr != nil {
			log.Error("Failed to rollback changes", rollbackErr)
		}
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appName).
			WithSuggestions("Check file permissions and disk space", "Configuration has been rolled back")
	}

	// Commit the operation (remove backup)
	if err := safeOp.Commit(); err != nil {
		log.Error("Failed to cleanup backup", err)
	}

	// Cleanup old backups
	if err := safeOp.Cleanup(5); err != nil {
		log.Error("Failed to cleanup old backups", err)
	}

	log.Success("Configuration cycled", map[string]interface{}{
		"from": currentValue,
		"to":   nextValue,
	})

	// Run post-toggle hooks
	return e.runHooks(appConfig, "post-cycle")
}

// AppendConfiguration adds a value to a list-based configuration
func (e *Engine) AppendConfiguration(appName, key, value string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if viper.GetBool("verbose") {
		log.Debug("Starting append operation", map[string]interface{}{
			"value": value,
		})
	}

	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		apps, _ := e.loader.ListApps()
		return errors.NewAppNotFoundError(appName, apps)
	}

	_, exists := appConfig.Fields[key]
	if !exists {
		// Even if not defined in fields, we might want to allow appending if it's a known list type
		// But for now strict mode:
		var availableFields []string
		for field := range appConfig.Fields {
			availableFields = append(availableFields, field)
		}
		return errors.NewFieldNotFoundError(appName, key, availableFields)
	}

	// Load target config
	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return errors.Wrap(errors.ConfigParseError, "failed to load target config", err).
			WithApp(appName).
			WithSuggestions("Check if the config file exists and is readable")
	}

	// Get current value
	currentVal := targetConfig.Get(key)
	var newList []interface{}

	if currentVal == nil {
		newList = []interface{}{value}
	} else if sliceVal, ok := currentVal.([]interface{}); ok {
		// Check for duplicates
		for _, item := range sliceVal {
			if strItem, ok := item.(string); ok && strItem == value {
				log.Info("Value already exists in configuration", map[string]interface{}{
					"value": value,
				})
				return nil
			}
		}
		newList = append(sliceVal, value)
	} else if strVal, ok := currentVal.(string); ok {
		// Convert single string to list if we are appending
		if strVal == value {
			log.Info("Value already exists in configuration", map[string]interface{}{
				"value": value,
			})
			return nil
		}
		newList = []interface{}{strVal, value}
	} else {
		// Fallback for other types, try to make a list
		newList = []interface{}{currentVal, value}
	}

	// Set the value
	if err := targetConfig.Set(key, newList); err != nil {
		return errors.Wrap(errors.ConfigWriteError, "failed to set config value", err).
			WithApp(appName).WithField(key).WithValue(value)
	}

	if viper.GetBool("dry-run") {
		log.Info("Would append configuration", map[string]interface{}{
			"value": value,
		})
		return nil
	}

	// Create safe operation with automatic backup
	configPath := e.expandPath(appConfig.Path)

	safeOp, err := recovery.NewSafeOperation(configPath, appName)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to create backup", err).
			WithApp(appName)
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		// Rollback on failure
		if rollbackErr := safeOp.Rollback(); rollbackErr != nil {
			log.Error("Failed to rollback changes", rollbackErr)
		}
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appName).
			WithSuggestions("Check file permissions and disk space", "Configuration has been rolled back")
	}

	// Commit the operation (remove backup)
	if err := safeOp.Commit(); err != nil {
		log.Error("Failed to cleanup backup", err)
	}

	// Cleanup old backups
	if err := safeOp.Cleanup(5); err != nil {
		log.Error("Failed to cleanup old backups", err)
	}

	log.Success("Configuration appended", map[string]interface{}{
		"value": value,
	})

	// Run post-toggle hooks (reusing same hook type for now or add new one)
	return e.runHooks(appConfig, "post-toggle")
}

// RemoveConfiguration removes a value from a list-based configuration
func (e *Engine) RemoveConfiguration(appName, key, value string) error {
	log := e.logger.WithApp(appName).WithField(key)

	if viper.GetBool("verbose") {
		log.Debug("Starting remove operation", map[string]interface{}{
			"value": value,
		})
	}

	appConfig, err := e.loader.LoadAppConfig(appName)
	if err != nil {
		apps, _ := e.loader.ListApps()
		return errors.NewAppNotFoundError(appName, apps)
	}

	// Load target config
	targetConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return errors.Wrap(errors.ConfigParseError, "failed to load target config", err).
			WithApp(appName).
			WithSuggestions("Check if the config file exists and is readable")
	}

	// Get current value
	currentVal := targetConfig.Get(key)
	if currentVal == nil {
		log.Warn("Configuration key not found", nil)
		return nil
	}

	var newList []interface{}
	found := false

	if sliceVal, ok := currentVal.([]interface{}); ok {
		for _, item := range sliceVal {
			strItem := fmt.Sprintf("%v", item)
			// Check if this item is the one we want to remove
			// We might need fuzzy matching? For now exact match or specific key match
			// For keybinds: "ctrl+c=copy" vs "ctrl+c"
			// If user passes "ctrl+c", we should remove "ctrl+c=..." or "ctrl+c"
			if strItem == value || strings.HasPrefix(strItem, value+"=") {
				found = true
				continue
			}
			newList = append(newList, item)
		}
	} else if strVal, ok := currentVal.(string); ok {
		if strVal == value || strings.HasPrefix(strVal, value+"=") {
			found = true
			// List becomes empty, or nil? koanf might handle nil differently
			// keeping newList empty
		} else {
			newList = []interface{}{strVal}
		}
	}

	if !found {
		log.Warn("Value not found in configuration", map[string]interface{}{
			"value": value,
		})
		return nil
	}

	// Set the value
	if len(newList) == 0 {
		// Remove key if empty? Or set to empty list?
		// Setting to empty slice is safer for list types
		if err := targetConfig.Set(key, []interface{}{}); err != nil {
			return errors.Wrap(errors.ConfigWriteError, "failed to set config value", err).
				WithApp(appName).WithField(key)
		}
	} else {
		if err := targetConfig.Set(key, newList); err != nil {
			return errors.Wrap(errors.ConfigWriteError, "failed to set config value", err).
				WithApp(appName).WithField(key)
		}
	}

	if viper.GetBool("dry-run") {
		log.Info("Would remove configuration", map[string]interface{}{
			"value": value,
		})
		return nil
	}

	// Create safe operation with automatic backup
	configPath := e.expandPath(appConfig.Path)

	safeOp, err := recovery.NewSafeOperation(configPath, appName)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to create backup", err).
			WithApp(appName)
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		// Rollback on failure
		if rollbackErr := safeOp.Rollback(); rollbackErr != nil {
			log.Error("Failed to rollback changes", rollbackErr)
		}
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appName).
			WithSuggestions("Check file permissions and disk space", "Configuration has been rolled back")
	}

	// Commit the operation (remove backup)
	if err := safeOp.Commit(); err != nil {
		log.Error("Failed to cleanup backup", err)
	}

	// Cleanup old backups
	if err := safeOp.Cleanup(5); err != nil {
		log.Error("Failed to cleanup old backups", err)
	}

	log.Success("Configuration removed", map[string]interface{}{
		"value": value,
	})

	// Run post-toggle hooks
	return e.runHooks(appConfig, "post-toggle")
}

func (e *Engine) ApplyPreset(appName, presetName string) error {
	log := e.logger.WithApp(appName).WithContext(map[string]interface{}{
		"preset": presetName,
	})

	if viper.GetBool("verbose") {
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

	// Apply all values from the preset
	for key, value := range preset.Values {
		fieldConfig, exists := appConfig.Fields[key]
		if !exists {
			if viper.GetBool("verbose") {
				log.Warn("Field not found in app config, applying anyway", nil, map[string]interface{}{
					"field": key,
				})
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

	if viper.GetBool("dry-run") {
		log.Info("Would apply preset", map[string]interface{}{
			"values": preset.Values,
		})
		return nil
	}

	// Create safe operation with automatic backup
	configPath := e.expandPath(appConfig.Path)

	safeOp, err := recovery.NewSafeOperation(configPath, appName)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to create backup", err).
			WithApp(appName)
	}

	// Save the config
	if err := e.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		// Rollback on failure
		if rollbackErr := safeOp.Rollback(); rollbackErr != nil {
			log.Error("Failed to rollback changes", rollbackErr)
		}
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appName).
			WithSuggestions("Check file permissions and disk space", "Configuration has been rolled back")
	}

	// Commit the operation (remove backup)
	if err := safeOp.Commit(); err != nil {
		log.Error("Failed to cleanup backup", err)
	}

	// Cleanup old backups
	if err := safeOp.Cleanup(5); err != nil {
		log.Error("Failed to cleanup old backups", err)
	}

	log.Success("Preset applied successfully")
	if viper.GetBool("verbose") {
		log.Debug("Preset values applied", map[string]interface{}{
			"values": preset.Values,
		})
	}

	// Run post-preset hooks
	return e.runHooks(appConfig, "post-preset")
}

// ShowPresetDiff shows the configuration changes that would be made by applying a preset
func (e *Engine) ShowPresetDiff(appName, presetName string) error {
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

	// Load current config
	currentConfig, err := e.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return fmt.Errorf("failed to load current config: %w", err)
	}

	// Create a copy of current config and apply preset changes
	modifiedConfig := *currentConfig
	for key, value := range preset.Values {
		fieldConfig, exists := appConfig.Fields[key]
		if !exists {
			if viper.GetBool("verbose") {
				fmt.Printf("Warning: Field %s not found in app config, applying anyway\n", key)
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

		_ = modifiedConfig.Set(key, convertedValue)
	}

	// Calculate diff
	differ := configextractor.NewConfigDiffer()
	diff := differ.DiffConfigurations(currentConfig.All(), modifiedConfig.All())

	if diff.HasChanges() {
		fmt.Printf("Configuration changes for preset '%s' on app '%s':\n", presetName, appName)
		fmt.Println(diff.FormatDiff())
		fmt.Printf("Summary: %s\n", diff.Summary())
	} else {
		fmt.Printf("No changes would be made by applying preset '%s' to app '%s'\n", presetName, appName)
	}

	return nil
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
func (e *Engine) convertValue(value, fieldType string) (interface{}, error) {
	switch fieldType {
	case "boolean":
		return strconv.ParseBool(value)
	case "number":
		// Try int first, then float
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal, nil
		}
		return strconv.ParseFloat(value, 64)
	case "string", "choice":
		return value, nil
	default:
		// Default to string
		return value, nil
	}
}

// runHooks executes post-action hooks
func (e *Engine) runHooks(appConfig *appconfig.AppConfig, hookType string) error {
	hookCmd, exists := appConfig.Hooks[hookType]
	if !exists {
		return nil
	}

	log := e.logger.WithApp(appConfig.Name).WithContext(map[string]interface{}{
		"hook_type": hookType,
		"command":   hookCmd,
	})

	if viper.GetBool("verbose") {
		log.Debug("Running hook")
	}

	// Set environment variables safely
	if err := e.setEnvironmentVariables(appConfig.Env); err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}

	// Execute the hook command
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return nil
	}

	// Security validation: Check if command is allowed
	if err := e.validateHookCommand(hookCmd); err != nil {
		log.Error("Hook command validation failed", err)
		return fmt.Errorf("hook validation failed: %w", err)
	}

	// Use filepath.Clean to prevent path traversal
	commandPath := filepath.Clean(parts[0])

	// Create command with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandPath, parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set working directory to a safe location
	if homeDir, err := os.UserHomeDir(); err == nil {
		cmd.Dir = homeDir
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %s failed: %w", hookType, err)
	}

	return nil
}

// GetAppConfig returns the configuration metadata for an app (for TUI use)
func (e *Engine) GetAppConfig(appName string) (*appconfig.AppConfig, error) {
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

	// Check for dangerous characters and patterns
	dangerousPatterns := []string{
		"|", "&&", "||", ";", "`", "$", // Shell operators
		"rm -rf", "rm -f", ">/dev/null", // Dangerous operations
		"curl", "wget", "nc", "telnet", // Network operations
		"sudo", "su -", "chmod +x", // Privilege escalation
	}

	lowerCmd := strings.ToLower(hookCmd)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return fmt.Errorf("hook command contains dangerous pattern: %s", pattern)
		}
	}

	// Allow-list approach: only allow certain safe commands
	parts := strings.Fields(hookCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty hook command")
	}

	allowedCommands := []string{
		"echo", "printf", "cat", "head", "tail", "wc",
		"grep", "sed", "awk", "sort", "uniq",
		"touch", "mkdir", "cp", "mv", "ls",
		"notify-send", "osascript", // Notification commands
	}

	command := filepath.Base(parts[0])
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

// expandPath efficiently expands ~ to home directory with thread-safe LRU caching
func (e *Engine) expandPath(path string) string {
	// Check cache first with read lock
	e.pathMutex.RLock()
	if expanded, exists := e.pathCache.Get(path); exists {
		e.pathMutex.RUnlock()
		return expanded
	}
	e.pathMutex.RUnlock()

	// Expand path
	var expanded string
	if strings.HasPrefix(path, "~") {
		expanded = strings.Replace(path, "~", e.homeDir, 1)
	} else {
		expanded = path
	}

	// Cache the result with write lock
	e.pathMutex.Lock()
	e.pathCache.Add(path, expanded)
	e.pathMutex.Unlock()

	return expanded
}
