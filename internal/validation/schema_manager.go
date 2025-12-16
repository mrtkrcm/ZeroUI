package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadSchema loads a validation schema from a JSON file
func (v *Validator) LoadSchema(schemaPath string) error {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}

	v.optimizeSchema(&schema)
	v.RegisterSchema(schema.Name, &schema)
	return nil
}

// LoadSchemasFromDir loads all validation schemas from a directory
func (v *Validator) LoadSchemasFromDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		schemaPath := filepath.Join(dir, file.Name())
		if err := v.LoadSchema(schemaPath); err != nil {
			return fmt.Errorf("failed to load schema %s: %w", file.Name(), err)
		}
	}

	return nil
}

// optimizeSchema pre-computes and caches expensive operations
func (v *Validator) optimizeSchema(schema *Schema) {
	for _, rule := range schema.Fields {
		// Cache enum values in map for O(1) lookup
		if len(rule.Enum) > 0 {
			rule.enumMap = make(map[string]struct{}, len(rule.Enum))
			for _, val := range rule.Enum {
				rule.enumMap[val] = struct{}{}
			}
		}

		// Pre-compile regex patterns
		if rule.Pattern != "" {
			if regex, err := regexp.Compile(rule.Pattern); err == nil {
				rule.compiledRegex = regex
			}
		}
	}
}

// HasSchema checks if a schema exists
func (v *Validator) HasSchema(name string) bool {
	_, ok := v.schemas[name]
	return ok
}

// ListSchemas returns all registered schema names
func (v *Validator) ListSchemas() []string {
	names := make([]string, 0, len(v.schemas))
	for name := range v.schemas {
		names = append(names, name)
	}
	return names
}

// isSimpleSchema checks if a schema is simple (for performance optimization)
// A "simple" schema is one that doesn't require the full, expensive validation
// path. Historically any presence of Global made a schema complex; tests and
// usage expect a schema that only declares Global.RequiredFields to still be
// considered simple. Relax the criteria to allow Global to be present as long
// as it does not introduce additional constraints (min/max/forbidden).
func (v *Validator) isSimpleSchema(schema *Schema) bool {
	// If there are global constraints beyond RequiredFields, treat as complex.
	if schema.Global != nil {
		if schema.Global.MinFields != nil || schema.Global.MaxFields != nil || len(schema.Global.ForbiddenFields) > 0 {
			return false
		}
		// Global.RequiredFields alone is allowed for the fast path.
	}

	for _, rule := range schema.Fields {
		if rule.Custom != nil {
			return false
		}
		if len(rule.Dependencies) > 0 || len(rule.ConflictsWith) > 0 {
			return false
		}
		if rule.Pattern != "" || rule.Format != "" {
			return false
		}
	}

	return true
}
