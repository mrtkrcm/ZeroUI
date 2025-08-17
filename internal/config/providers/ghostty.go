package providers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/knadh/koanf/v2"
)

// GhosttyProvider is a koanf provider for Ghostty's custom configuration format.
// It implements the Provider interface from github.com/knadh/koanf/providers
type GhosttyProvider struct {
	path string
}

// NewGhosttyProvider creates a new Ghostty config provider.
func NewGhosttyProvider(path string) *GhosttyProvider {
	return &GhosttyProvider{path: path}
}

// Read reads the Ghostty config file and returns raw bytes.
// This implements the koanf Provider interface.
func (p *GhosttyProvider) Read() ([]byte, error) {
	file, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Ghostty config file %q: %w", p.path, err)
	}
	defer file.Close()

	// Read and convert Ghostty format to key=value properties format
	// that can be parsed by existing parsers
	return p.convertGhosttyToProperties(file)
}

// ReadBytes reads Ghostty config from a byte slice.
func (p *GhosttyProvider) ReadBytes(b []byte) ([]byte, error) {
	reader := strings.NewReader(string(b))
	return p.convertGhosttyToProperties(reader)
}

// convertGhosttyToProperties converts Ghostty format to Java properties format
func (p *GhosttyProvider) convertGhosttyToProperties(r io.Reader) ([]byte, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(r)
	
	// Track multiple values for the same key
	keyValues := make(map[string][]string)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Skip lines with empty keys
		if key == "" {
			continue
		}
		
		// Collect values for each key
		keyValues[key] = append(keyValues[key], value)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Ghostty config: %w", err)
	}
	
	// Convert to properties format
	for key, values := range keyValues {
		if len(values) == 1 {
			// Single value - simple key=value
			result.WriteString(fmt.Sprintf("%s=%s\n", key, values[0]))
		} else {
			// Multiple values - encode as comma-separated list
			// This preserves the array nature while being parseable
			result.WriteString(fmt.Sprintf("%s=%s\n", key, strings.Join(values, ",")))
		}
	}
	
	return []byte(result.String()), nil
}

// GhosttyParser is a koanf parser for Ghostty configuration format.
// It handles the conversion from Ghostty's key=value format to koanf's map structure.
type GhosttyParser struct{}

// NewGhosttyParser creates a new Ghostty format parser.
func NewGhosttyParser() *GhosttyParser {
	return &GhosttyParser{}
}

// Unmarshal parses Ghostty format data into koanf's map structure.
func (p *GhosttyParser) Unmarshal(b []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if key == "" {
			continue
		}
		
		// Check if value contains comma (indicates multiple values)
		if strings.Contains(value, ",") {
			// Split into array
			values := strings.Split(value, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			result[key] = values
		} else {
			result[key] = value
		}
	}
	
	return result, nil
}

// Marshal converts koanf's map structure back to Ghostty format.
func (p *GhosttyParser) Marshal(m map[string]interface{}) ([]byte, error) {
	var result strings.Builder
	
	for key, value := range m {
		switch v := value.(type) {
		case []string:
			// Multiple values - write each on separate line
			for _, val := range v {
				result.WriteString(fmt.Sprintf("%s = %s\n", key, val))
			}
		case []interface{}:
			// Multiple values of mixed types
			for _, val := range v {
				result.WriteString(fmt.Sprintf("%s = %v\n", key, val))
			}
		case string:
			result.WriteString(fmt.Sprintf("%s = %s\n", key, v))
		default:
			result.WriteString(fmt.Sprintf("%s = %v\n", key, v))
		}
	}
	
	return []byte(result.String()), nil
}

// GhosttyProviderWithParser combines provider and parser for convenience.
type GhosttyProviderWithParser struct {
	provider *GhosttyProvider
	parser   *GhosttyParser
}

// NewGhosttyProviderWithParser creates a provider with built-in parser.
func NewGhosttyProviderWithParser(path string) *GhosttyProviderWithParser {
	return &GhosttyProviderWithParser{
		provider: NewGhosttyProvider(path),
		parser:   NewGhosttyParser(),
	}
}

// LoadIntoKoanf loads Ghostty config directly into a koanf instance.
func (p *GhosttyProviderWithParser) LoadIntoKoanf(k *koanf.Koanf) error {
	// Read the config file using the provider
	data, err := p.provider.Read()
	if err != nil {
		return fmt.Errorf("failed to read Ghostty config: %w", err)
	}
	
	// Parse the data using the parser
	configMap, err := p.parser.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("failed to parse Ghostty config: %w", err)
	}
	
	// Load each key-value pair directly into koanf
	for key, value := range configMap {
		if err := k.Set(key, value); err != nil {
			return fmt.Errorf("failed to set key %s: %w", key, err)
		}
	}
	
	return nil
}