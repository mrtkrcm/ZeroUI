package performance

import (
	"sync"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
)

// ParserPool manages pools of reusable parsers for different formats
var ParserPool = struct {
	JSON *sync.Pool
	YAML *sync.Pool
	TOML *sync.Pool
}{
	JSON: &sync.Pool{
		New: func() interface{} {
			return json.Parser()
		},
	},
	YAML: &sync.Pool{
		New: func() interface{} {
			return yaml.Parser()
		},
	},
	TOML: &sync.Pool{
		New: func() interface{} {
			return toml.Parser()
		},
	},
}

// GetJSONParser gets a JSON parser from the pool
func GetJSONParser() koanf.Parser {
	return ParserPool.JSON.Get().(koanf.Parser)
}

// PutJSONParser returns a JSON parser to the pool
func PutJSONParser(p koanf.Parser) {
	if p != nil {
		ParserPool.JSON.Put(p)
	}
}

// GetYAMLParser gets a YAML parser from the pool
func GetYAMLParser() koanf.Parser {
	return ParserPool.YAML.Get().(koanf.Parser)
}

// PutYAMLParser returns a YAML parser to the pool
func PutYAMLParser(p koanf.Parser) {
	if p != nil {
		ParserPool.YAML.Put(p)
	}
}

// GetTOMLParser gets a TOML parser from the pool
func GetTOMLParser() koanf.Parser {
	return ParserPool.TOML.Get().(koanf.Parser)
}

// PutTOMLParser returns a TOML parser to the pool
func PutTOMLParser(p koanf.Parser) {
	if p != nil {
		ParserPool.TOML.Put(p)
	}
}

// GetParserForFormat gets an appropriate parser from the pool based on format
func GetParserForFormat(format string) koanf.Parser {
	switch format {
	case "json":
		return GetJSONParser()
	case "yaml", "yml":
		return GetYAMLParser()
	case "toml":
		return GetTOMLParser()
	default:
		return nil
	}
}

// PutParserForFormat returns a parser to the appropriate pool based on format
func PutParserForFormat(format string, p koanf.Parser) {
	if p == nil {
		return
	}

	switch format {
	case "json":
		PutJSONParser(p)
	case "yaml", "yml":
		PutYAMLParser(p)
	case "toml":
		PutTOMLParser(p)
	}
}
