package rpc

import "context"

// ConfigPlugin defines the RPC interface for configuration plugins
type ConfigPlugin interface {
	GetInfo(ctx context.Context) (*PluginInfo, error)
	DetectConfig(ctx context.Context) (*ConfigInfo, error)
	ParseConfig(ctx context.Context, path string) (*ConfigData, error)
	WriteConfig(ctx context.Context, path string, data *ConfigData) error
	ValidateField(ctx context.Context, field string, value interface{}) error
	ValidateConfig(ctx context.Context, data *ConfigData) error
	GetSchema(ctx context.Context) (*ConfigMetadata, error)
	SupportsFeature(ctx context.Context, feature string) (bool, error)
}

// Core capabilities
const (
	CapabilityConfigParsing = "config.parsing"
	CapabilityConfigWriting = "config.writing"
	CapabilityValidation    = "validation"
	CapabilitySchemaExport  = "schema.export"
	CapabilityPresets       = "presets"
)

const CurrentAPIVersion = "v1.0.0"