package validation

import (
	"testing"
)

func TestUniquenessValidation(t *testing.T) {
	validator := NewValidator()

	t.Run("Global uniqueness - reserved names", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "global",
			"field": "name",
		}

		// Test reserved name
		err := validator.validateUniqueness("name", "admin", args)
		if err == nil {
			t.Error("Expected error for reserved name 'admin'")
		}

		// Test allowed name
		err = validator.validateUniqueness("name", "myapp", args)
		if err != nil {
			t.Errorf("Unexpected error for valid name: %v", err)
		}
	})

	t.Run("Global uniqueness - port validation", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "global",
			"field": "port",
		}

		// Test well-known port
		err := validator.validateUniqueness("port", "80", args)
		if err == nil {
			t.Error("Expected error for well-known port '80'")
		}

		// Test custom port
		err = validator.validateUniqueness("port", "8080", args)
		if err != nil {
			t.Errorf("Unexpected error for custom port: %v", err)
		}
	})

	t.Run("Global uniqueness - path validation", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "global",
			"field": "path",
		}

		// Test system path
		err := validator.validateUniqueness("path", "/etc/config", args)
		if err == nil {
			t.Error("Expected error for system path '/etc/config'")
		}

		// Test user path
		err = validator.validateUniqueness("path", "/home/user/myapp", args)
		if err != nil {
			t.Errorf("Unexpected error for user path: %v", err)
		}
	})

	t.Run("Local uniqueness validation", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "local",
			"field": "name",
		}

		// Test too short
		err := validator.validateUniqueness("name", "a", args)
		if err == nil {
			t.Error("Expected error for too short value")
		}

		// Test common pattern
		err = validator.validateUniqueness("name", "test-app", args)
		if err == nil {
			t.Error("Expected error for common pattern 'test'")
		}

		// Test unique value
		err = validator.validateUniqueness("name", "unique-identifier", args)
		if err != nil {
			t.Errorf("Unexpected error for unique value: %v", err)
		}
	})

	t.Run("App uniqueness validation", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "app",
			"field": "name",
		}

		// Test generic name
		err := validator.validateUniqueness("name", "app", args)
		if err == nil {
			t.Error("Expected error for generic name 'app'")
		}

		// Test specific name
		err = validator.validateUniqueness("name", "myspecificapp", args)
		if err != nil {
			t.Errorf("Unexpected error for specific name: %v", err)
		}
	})

	t.Run("Empty values", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "global",
			"field": "name",
		}

		// Empty values should not be validated
		err := validator.validateUniqueness("name", "", args)
		if err != nil {
			t.Errorf("Unexpected error for empty value: %v", err)
		}
	})

	t.Run("Unknown scope", func(t *testing.T) {
		args := map[string]interface{}{
			"scope": "unknown",
			"field": "name",
		}

		err := validator.validateUniqueness("name", "value", args)
		if err == nil {
			t.Error("Expected error for unknown scope")
		}
	})

	t.Run("Default scope and field", func(t *testing.T) {
		// No args provided - should use defaults
		args := map[string]interface{}{}

		err := validator.validateUniqueness("name", "admin", args)
		if err == nil {
			t.Error("Expected error for reserved name with default scope")
		}
	})
}

func TestCustomValidationWithUniqueness(t *testing.T) {
	validator := NewValidator()

	t.Run("Custom validation with uniqueness rule", func(t *testing.T) {
		custom := &CustomRule{
			Function: "unique",
			Args: map[string]interface{}{
				"scope": "global",
				"field": "name",
			},
			Message: "Value must be unique",
		}

		// Test reserved value
		err := validator.validateCustom("root", custom)
		if err == nil {
			t.Error("Expected error for reserved value in custom validation")
		}

		// Test valid value
		err = validator.validateCustom("uniquename", custom)
		if err != nil {
			t.Errorf("Unexpected error for valid value: %v", err)
		}
	})
}
