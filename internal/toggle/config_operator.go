package toggle

import (
	"os"
	"path/filepath"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/spf13/viper"
)

// ConfigOperator handles core config read/write operations
type ConfigOperator struct {
	loader    ConfigLoader
	homeDir   string
	pathCache *lru.Cache[string, string]
	pathMutex sync.RWMutex
}

// NewConfigOperator creates a new config operator
func NewConfigOperator(loader ConfigLoader) *ConfigOperator {
	homeDir, _ := os.UserHomeDir()
	pathCache, _ := lru.New[string, string](1000)

	return &ConfigOperator{
		loader:    loader,
		homeDir:   homeDir,
		pathCache: pathCache,
	}
}

// LoadAppConfig loads an application configuration
func (co *ConfigOperator) LoadAppConfig(appName string) (*config.AppConfig, error) {
	appConfig, err := co.loader.LoadAppConfig(appName)
	if err != nil {
		// Check if it's an app not found error
		apps, _ := co.loader.ListApps()
		return nil, errors.NewAppNotFoundError(appName, apps)
	}
	return appConfig, nil
}

// LoadTargetConfig loads the target configuration file
func (co *ConfigOperator) LoadTargetConfig(appConfig *config.AppConfig) (*koanf.Koanf, error) {
	targetConfig, err := co.loader.LoadTargetConfig(appConfig)
	if err != nil {
		return nil, errors.Wrap(errors.ConfigParseError, "failed to load target config", err).
			WithApp(appConfig.Name).
			WithSuggestions("Check if the config file exists and is readable")
	}
	return targetConfig, nil
}

// SetConfigValue sets a value in the configuration
func (co *ConfigOperator) SetConfigValue(targetConfig *koanf.Koanf, key string, value interface{}, appName string) error {
	if err := targetConfig.Set(key, value); err != nil {
		return errors.Wrap(errors.ConfigWriteError, "failed to set config value", err).
			WithApp(appName).WithField(key)
	}
	return nil
}

// SaveConfigSafely saves configuration with backup and validation
func (co *ConfigOperator) SaveConfigSafely(appConfig *config.AppConfig, targetConfig *koanf.Koanf) error {
	if viper.GetBool("dry-run") {
		return nil // Don't actually save in dry-run mode
	}

	// For now, use simple save without advanced validation to avoid circular dependency
	// TODO: Refactor the validator adapter to remove circular dependency
	if err := co.loader.SaveTargetConfig(appConfig, targetConfig); err != nil {
		return errors.Wrap(errors.ConfigWriteError, "failed to save config", err).
			WithApp(appConfig.Name).
			WithSuggestions("Check file permissions and disk space")
	}

	return nil
}

// expandPath expands ~ to home directory with caching
func (co *ConfigOperator) expandPath(configPath string) string {
	// Check cache first for performance
	co.pathMutex.RLock()
	if expanded, found := co.pathCache.Get(configPath); found {
		co.pathMutex.RUnlock()
		return expanded
	}
	co.pathMutex.RUnlock()

	// Expand the path
	var expanded string
	if configPath[:1] == "~" {
		expanded = filepath.Join(co.homeDir, configPath[1:])
	} else {
		expanded = configPath
	}

	// Cache the result
	co.pathMutex.Lock()
	co.pathCache.Add(configPath, expanded)
	co.pathMutex.Unlock()

	return expanded
}
