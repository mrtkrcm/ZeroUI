package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validateUniqueness performs various "uniqueness" and reserved-value checks.
// signature is intentionally unexported to match existing test usage:
//
//	validator.validateUniqueness(field, value, args)
func (v *Validator) validateUniqueness(field string, value string, args map[string]interface{}) error {
	// Empty values should not be validated (per tests)
	if strings.TrimSpace(value) == "" {
		return nil
	}

	// Determine scope and target field from args (with defaults)
	scope := "global"
	if s, ok := args["scope"].(string); ok && s != "" {
		scope = strings.ToLower(s)
	}
	targetField := field
	if f, ok := args["field"].(string); ok && f != "" {
		targetField = f
	}

	switch scope {
	case "global":
		return v.validateGlobalUniqueness(targetField, value)
	case "local":
		return v.validateLocalUniqueness(targetField, value)
	case "app":
		return v.validateAppUniqueness(targetField, value)
	default:
		return fmt.Errorf("unknown uniqueness scope: %s", scope)
	}
}

// validateGlobalUniqueness implements rules for the "global" scope.
func (v *Validator) validateGlobalUniqueness(field, value string) error {
	switch field {
	case "name":
		if isReservedName(value) {
			return errors.New("name is reserved")
		}
		// allow otherwise
		return nil
	case "port":
		if isWellKnownPort(value) {
			return errors.New("port conflicts with well-known port")
		}
		return nil
	case "path":
		if isSystemPath(value) {
			return errors.New("path is reserved/system path")
		}
		return nil
	default:
		// For unknown target fields in global scope, be conservative and allow
		return nil
	}
}

// validateLocalUniqueness implements rules for the "local" (per-user / per-project) scope.
func (v *Validator) validateLocalUniqueness(field, value string) error {
	switch field {
	case "name":
		// Example rules:
		// - too short names are rejected
		// - names matching common test/dev patterns are rejected
		if isTooShort(value, 2) {
			return errors.New("value too short for local uniqueness")
		}
		if matchesCommonPattern(value) {
			return errors.New("value matches a common (non-unique) pattern")
		}
		return nil
	default:
		return nil
	}
}

// validateAppUniqueness implements rules for the "app" scope.
func (v *Validator) validateAppUniqueness(field, value string) error {
	switch field {
	case "name":
		// Disallow extremely generic app names
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "app" || lower == "application" || lower == "service" {
			return errors.New("app name is too generic")
		}
		return nil
	default:
		return nil
	}
}

// Helper predicates

func isReservedName(name string) bool {
	l := strings.ToLower(strings.TrimSpace(name))
	if l == "" {
		return false
	}
	reserved := []string{
		"admin", "root", "system", "null", "none", "default",
	}
	for _, r := range reserved {
		if l == r {
			return true
		}
	}
	return false
}

func isWellKnownPort(port string) bool {
	known := map[string]struct{}{
		"20":   {},
		"21":   {},
		"22":   {}, // ssh
		"23":   {},
		"25":   {},
		"53":   {},
		"80":   {}, // http
		"110":  {},
		"143":  {},
		"443":  {}, // https
		"587":  {},
		"3306": {},
		"5432": {},
	}
	port = strings.TrimSpace(port)
	_, ok := known[port]
	return ok
}

func isSystemPath(p string) bool {
	// treat typical Unix system paths as reserved
	p = strings.TrimSpace(p)
	if p == "" {
		return false
	}
	// simple checks: prefixes commonly used for system-level configs
	systemPrefixes := []string{"/etc/", "/usr/", "/var/", "/bin/", "/sbin/"}
	for _, pref := range systemPrefixes {
		if strings.HasPrefix(p, pref) {
			return true
		}
	}
	// also reject root config paths like "/etc" exactly
	if p == "/etc" || p == "/usr" || p == "/var" {
		return true
	}
	return false
}

func isTooShort(s string, min int) bool {
	return len(strings.TrimSpace(s)) < min
}

var commonPatternRE = regexp.MustCompile(`(?i)\b(test|demo|sample|example|temp|tmp)\b`)

func matchesCommonPattern(s string) bool {
	return commonPatternRE.MatchString(s)
}

// RegisterUniqueValidation registers a struct-tag level validation for the underlying
// go-playground validator instance if callers wish to use a `unique` tag.
//
// Note: NewValidator currently creates a validator instance but does not automatically
// register this tag. Callers that want to enable tag-based `unique` validation can call
// this helper after creating the Validator instance.
func (v *Validator) RegisterUniqueValidation() error {
	// safe-guard: v.validate may be nil
	if v == nil || v.validate == nil {
		return errors.New("validator instance not initialized")
	}
	// register tag 'unique' which will call validateUniqueness with
	// a default/global scope using the field name.
	return v.validate.RegisterValidation("unique", func(fl validator.FieldLevel) bool {
		// FieldLevel provides field info and top-level struct; we only have the field value here.
		val := strings.TrimSpace(fmt.Sprintf("%v", fl.Field().Interface()))
		// use field name from fl.FieldName() if available; fallback to empty
		fieldName := strings.ToLower(fl.FieldName())
		// call our method with defaults (empty args -> defaults to global/name)
		err := v.validateUniqueness(fieldName, val, map[string]interface{}{})
		return err == nil
	})
}
