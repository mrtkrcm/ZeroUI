package container

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// Container holds all application dependencies
type Container struct {
	logger        *logger.Logger
	configLoader  *appconfig.ReferenceEnhancedLoader
	toggleEngine  *toggle.Engine
	configService *service.ConfigService
}

// New creates a new dependency container
func New() (*Container, error) {
	c := &Container{}

	// Use the global logger which should already be initialized with runtime config
	// This avoids resetting the log level/format that was configured via flags
	c.logger = logger.Global()

	// Initialize enhanced config loader with reference integration
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced config loader: %w", err)
	}
	c.configLoader = configLoader

	// Initialize toggle engine with dependency injection
	c.toggleEngine = toggle.NewEngineWithDeps(configLoader, c.logger)

	// Initialize config service with all dependencies
	c.configService = service.NewConfigService(c.toggleEngine, configLoader, c.logger)

	return c, nil
}

// Logger returns the logger instance
func (c *Container) Logger() *logger.Logger {
	return c.logger
}

// ConfigLoader returns the config loader instance
func (c *Container) ConfigLoader() *appconfig.ReferenceEnhancedLoader {
	return c.configLoader
}

// ToggleEngine returns the toggle engine instance
func (c *Container) ToggleEngine() *toggle.Engine {
	return c.toggleEngine
}

// ConfigService returns the config service instance
func (c *Container) ConfigService() *service.ConfigService {
	return c.configService
}

// Close cleans up resources
func (c *Container) Close() error {
	// Add any cleanup logic here if needed
	return nil
}
