package container

import (
	"fmt"
	"os"

	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// Container holds all application dependencies
type Container struct {
	logger        *logger.Logger
	configLoader  *config.ReferenceEnhancedLoader
	toggleEngine  *toggle.Engine
	configService *service.ConfigService
}

// Config holds container configuration
type Config struct {
	LogLevel  string
	LogFormat string
}

// DefaultConfig returns default container configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  "info",
		LogFormat: "console",
	}
}

// New creates a new dependency container
func New(cfg *Config) (*Container, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	c := &Container{}

	// Initialize logger first
	loggerConfig := &logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
		Output: os.Stderr, // Use stderr for logging to avoid interfering with TUI
	}
	c.logger = logger.New(loggerConfig)

	// Initialize global logger for convenience
	logger.InitGlobal(loggerConfig)

	// Initialize enhanced config loader with reference integration
	configLoader, err := config.NewReferenceEnhancedLoader()
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
func (c *Container) ConfigLoader() *config.ReferenceEnhancedLoader {
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
