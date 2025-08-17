package interfaces

import (
	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// ConfigLoader defines the interface for configuration loading operations
type ConfigLoader interface {
	LoadAppConfig(appName string) (*config.AppConfig, error)
	LoadTargetConfig(app *config.AppConfig) (*koanf.Koanf, error)
	SaveTargetConfig(app *config.AppConfig, target *koanf.Koanf) error
	ListApps() ([]string, error)
	SetConfigDir(dir string)
}

// ToggleEngine defines the interface for configuration toggle operations
type ToggleEngine interface {
	Toggle(appName, key, value string) error
	Cycle(appName, key string) error
	ApplyPreset(appName, presetName string) error
	GetAppConfig(appName string) (*config.AppConfig, error)
	GetCurrentValues(appName string) (map[string]interface{}, error)
}

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Debug(msg string, fields ...map[string]interface{})
}

