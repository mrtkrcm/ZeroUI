package reference

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// StaticConfigLoader loads from embedded YAML/JSON files
type StaticConfigLoader struct {
	configDir string
}

// NewStaticConfigLoader creates a loader for static config files
func NewStaticConfigLoader(configDir string) *StaticConfigLoader {
	return &StaticConfigLoader{configDir: configDir}
}

// LoadReference loads configuration reference from static files
func (s *StaticConfigLoader) LoadReference(appName string) (*ConfigReference, error) {
	// Try different file formats
	extensions := []string{".yaml", ".yml", ".json"}
	
	for _, ext := range extensions {
		filename := filepath.Join(s.configDir, appName+ext)
		if data, err := os.ReadFile(filename); err == nil {
			return s.parseConfigFile(appName, filename, data)
		}
	}
	
	return nil, fmt.Errorf("no configuration file found for %s", appName)
}

// parseConfigFile parses configuration from different file formats
func (s *StaticConfigLoader) parseConfigFile(appName, filename string, data []byte) (*ConfigReference, error) {
	var ref ConfigReference
	var err error
	
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &ref)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &ref)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
	}
	
	// Set metadata
	ref.AppName = appName
	ref.LastUpdated = time.Now()
	
	return &ref, nil
}