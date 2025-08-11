package configextractor

import (
	"context"
	"time"
)

// ConfigExtractor provides the main extraction interface
type ConfigExtractor interface {
	// Extract gets config for a single app with context support
	Extract(ctx context.Context, app string) (*Config, error)
	
	// ExtractBatch processes multiple apps concurrently
	ExtractBatch(ctx context.Context, apps []string) (map[string]*Config, error)
	
	// SupportedApps returns list of supported applications
	SupportedApps() []string
}

// Config represents extracted configuration (minimal, focused design)
type Config struct {
	App        string            `json:"app"`
	ConfigPath string            `json:"config_path"`
	Format     string            `json:"format"` // json, yaml, toml, custom
	Settings   map[string]Setting `json:"settings"`
	Source     ExtractionSource   `json:"source"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Setting represents a configuration option (streamlined)
type Setting struct {
	Name    string      `json:"name"`
	Type    SettingType `json:"type"`
	Default interface{} `json:"default,omitempty"`
	Values  []string    `json:"values,omitempty"`  // For enum/choice types
	Desc    string      `json:"description,omitempty"`
	Cat     string      `json:"category,omitempty"`
}

// SettingType simplified enum
type SettingType string

const (
	TypeString  SettingType = "string"
	TypeNumber  SettingType = "number" 
	TypeBoolean SettingType = "boolean"
	TypeChoice  SettingType = "choice"
	TypeArray   SettingType = "array"
)

// ExtractionSource tracks where config was extracted from
type ExtractionSource struct {
	Method     string `json:"method"`     // cli, github, local, builtin
	Location   string `json:"location"`   // specific source location
	Confidence float64 `json:"confidence"` // 0.0-1.0 confidence score
}

// Strategy defines extraction strategy interface
type Strategy interface {
	// Name returns strategy identifier
	Name() string
	
	// CanExtract checks if strategy supports the app
	CanExtract(app string) bool
	
	// Extract performs the extraction
	Extract(ctx context.Context, app string) (*Config, error)
	
	// Priority returns strategy priority (higher = preferred)
	Priority() int
}

// Parser handles format-specific parsing
type Parser interface {
	// Parse converts raw data to Config
	Parse(app string, data []byte) (*Config, error)
	
	// Supports checks if parser handles the format
	Supports(format string) bool
}

// Cache provides intelligent caching with TTL
type Cache interface {
	// Get retrieves cached config
	Get(key string) (*Config, bool)
	
	// Set stores config with TTL
	Set(key string, config *Config)
	
	// Clear removes expired entries
	Clear()
}

// AppRegistry manages application definitions
type AppRegistry interface {
	// GetApp returns app metadata
	GetApp(name string) (*AppDef, bool)
	
	// ListApps returns all registered apps
	ListApps() []string
	
	// Register adds new app definition
	Register(def *AppDef)
}


// AppDef defines application extraction metadata
type AppDef struct {
	Name        string            `json:"name"`
	Aliases     []string          `json:"aliases,omitempty"`
	ConfigPath  string            `json:"config_path"`
	Format      string            `json:"format"`
	Strategies  []string          `json:"strategies"`         // Preferred extraction methods
	CLICommand  string            `json:"cli_command,omitempty"`
	GitHubRepo  string            `json:"github_repo,omitempty"`
	Categories  map[string]string `json:"categories,omitempty"` // Field prefix -> category mapping
}