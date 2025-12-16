package service

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// ConfigLoader interface to support different loader types
type ConfigLoader interface {
	LoadAppConfig(appName string) (*appconfig.AppConfig, error)
	ListApps() ([]string, error)
}

// ConfigService provides high-level configuration management operations
type ConfigService struct {
	engine *toggle.Engine
	loader ConfigLoader
	logger *logger.Logger
}

// NewConfigService creates a new config service
func NewConfigService(engine *toggle.Engine, loader ConfigLoader, log *logger.Logger) *ConfigService {
	return &ConfigService{
		engine: engine,
		loader: loader,
		logger: log,
	}
}

// ToggleConfiguration sets a configuration value
func (s *ConfigService) ToggleConfiguration(app, key, value string) error {
	log := s.logger.WithApp(app).WithField(key)
	log.InfoStructured("Toggling configuration", logger.Field{Key: "value", Value: value})

	return s.engine.Toggle(app, key, value)
}

// CycleConfiguration cycles to the next value for a configuration key
func (s *ConfigService) CycleConfiguration(app, key string) error {
	log := s.logger.WithApp(app).WithField(key)
	log.InfoStructured("Cycling configuration")

	return s.engine.Cycle(app, key)
}

// ApplyPreset applies a preset configuration
func (s *ConfigService) ApplyPreset(app, presetName string) error {
	log := s.logger.WithApp(app)
	log.InfoStructured("Applying preset", logger.Field{Key: "preset", Value: presetName})

	return s.engine.ApplyPreset(app, presetName)
}

// AppendConfiguration adds a value to a list-based configuration
func (s *ConfigService) AppendConfiguration(app, key, value string) error {
	log := s.logger.WithApp(app).WithField(key)
	log.InfoStructured("Appending configuration", logger.Field{Key: "value", Value: value})

	return s.engine.AppendConfiguration(app, key, value)
}

// RemoveConfiguration removes a value from a list-based configuration
func (s *ConfigService) RemoveConfiguration(app, key, value string) error {
	log := s.logger.WithApp(app).WithField(key)
	log.InfoStructured("Removing configuration", logger.Field{Key: "value", Value: value})

	return s.engine.RemoveConfiguration(app, key, value)
}

// ListApplications returns all available applications
func (s *ConfigService) ListApplications() ([]string, error) {
	s.logger.DebugStructured("Listing applications")

	apps, err := s.loader.ListApps()
	if err != nil {
		s.logger.ErrorStructured("Failed to list applications", err)
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	s.logger.DebugStructured("Found applications",
		logger.Field{Key: "count", Value: len(apps)},
		logger.Field{Key: "apps", Value: apps},
	)

	return apps, nil
}

// GetApplicationConfig returns configuration metadata for an app
func (s *ConfigService) GetApplicationConfig(app string) (*appconfig.AppConfig, error) {
	log := s.logger.WithApp(app)
	log.DebugStructured("Getting application configuration")

	appConfig, err := s.loader.LoadAppConfig(app)
	if err != nil {
		if contains(err.Error(), "not found") {
			apps, _ := s.ListApplications()
			err = errors.NewAppNotFoundError(app, apps)
			log.ErrorStructured("Application not found", err)
			return nil, err
		}
		log.ErrorStructured("Failed to load app config", err)
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	log.DebugStructured("Loaded application configuration",
		logger.Field{Key: "fields_count", Value: len(appConfig.Fields)},
		logger.Field{Key: "presets_count", Value: len(appConfig.Presets)},
	)

	return appConfig, nil
}

// GetCurrentValues returns the current configuration values for an app
func (s *ConfigService) GetCurrentValues(app string) (map[string]interface{}, error) {
	log := s.logger.WithApp(app)
	log.DebugStructured("Getting current configuration values")

	return s.engine.GetCurrentValues(app)
}

// ValidateConfiguration validates that a configuration change is valid
func (s *ConfigService) ValidateConfiguration(app, key, value string) error {
	log := s.logger.WithApp(app).WithField(key)
	log.DebugStructured("Validating configuration", logger.Field{Key: "value", Value: value})

	appConfig, err := s.GetApplicationConfig(app)
	if err != nil {
		return err
	}

	fieldConfig, exists := appConfig.Fields[key]
	if !exists {
		var availableFields []string
		for field := range appConfig.Fields {
			availableFields = append(availableFields, field)
		}
		return errors.NewFieldNotFoundError(app, key, availableFields)
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
			return errors.NewInvalidValueError(app, key, value, fieldConfig.Values)
		}
	}

	log.DebugStructured("Configuration validation passed")
	return nil
}

// ListPresets returns all available presets for an application
func (s *ConfigService) ListPresets(app string) (map[string]appconfig.PresetConfig, error) {
	log := s.logger.WithApp(app)
	log.DebugStructured("Listing presets")

	appConfig, err := s.GetApplicationConfig(app)
	if err != nil {
		return nil, err
	}

	log.DebugStructured("Found presets", logger.Field{Key: "count", Value: len(appConfig.Presets)})

	return appConfig.Presets, nil
}

// ListFields returns all configurable fields for an application
func (s *ConfigService) ListFields(app string) (map[string]appconfig.FieldConfig, error) {
	log := s.logger.WithApp(app)
	log.DebugStructured("Listing fields")

	appConfig, err := s.GetApplicationConfig(app)
	if err != nil {
		return nil, err
	}

	log.DebugStructured("Found fields", logger.Field{Key: "count", Value: len(appConfig.Fields)})

	return appConfig.Fields, nil
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
