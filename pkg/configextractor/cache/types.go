package cache

import "time"

// Config represents a cached configuration
// This is a local type to avoid import cycles
type Config struct {
	App        string                 `json:"app"`
	ConfigPath string                 `json:"config_path"`
	Format     string                 `json:"format"`
	Settings   map[string]interface{} `json:"settings"`
	Source     string                 `json:"source"`
	Timestamp  time.Time              `json:"timestamp"`
}
