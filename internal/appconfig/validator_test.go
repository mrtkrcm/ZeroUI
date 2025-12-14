package appconfig

import (
	"testing"
)

func TestValidator(t *testing.T) {
	v := NewValidator()

	t.Run("ValidateAppDefinition", func(t *testing.T) {
		app := &AppDefinition{
			Name:         "test-app",
			DisplayName:  "Test App",
			Icon:         "○",
			Description:  "Test application",
			Category:     "test",
			ConfigPaths:  []string{"~/.config/test/config.yaml"},
			ConfigFormat: "yaml",
		}

		result := v.ValidateAppDefinition(app)
		if !result.Valid {
			t.Errorf("Expected valid app, got errors: %v", result.Errors)
		}
	})

	t.Run("ValidateInvalidName", func(t *testing.T) {
		app := &AppDefinition{
			Name:        "test app", // Space is invalid
			ConfigPaths: []string{"~/.config/test/config.yaml"},
		}

		result := v.ValidateAppDefinition(app)
		if result.Valid {
			t.Error("Expected invalid due to space in name")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
	})

	t.Run("ValidateMissingConfigPaths", func(t *testing.T) {
		app := &AppDefinition{
			Name: "test-app",
		}

		result := v.ValidateAppDefinition(app)
		if result.Valid {
			t.Error("Expected invalid due to missing config paths")
		}
	})

	t.Run("ValidateEmptyIcon", func(t *testing.T) {
		app := &AppDefinition{
			Name:        "test-app",
			Icon:        "", // Empty icon
			ConfigPaths: []string{"~/.config/test/config.yaml"},
		}

		result := v.ValidateAppDefinition(app)
		if !result.Valid {
			t.Error("Should be valid with default icon")
		}

		if app.Icon != "○" {
			t.Errorf("Expected default icon ○, got %s", app.Icon)
		}

		if len(result.Warnings) == 0 {
			t.Error("Expected warning about empty icon")
		}
	})
}

func TestValidateRegistry(t *testing.T) {
	v := NewValidator()

	t.Run("ValidRegistry", func(t *testing.T) {
		registry := &AppsRegistry{
			Applications: []AppDefinition{
				{
					Name:        "app1",
					ConfigPaths: []string{"~/.config/app1/config"},
					Category:    "tools",
				},
				{
					Name:        "app2",
					ConfigPaths: []string{"~/.config/app2/config"},
					Category:    "tools",
				},
			},
			Categories: []CategoryDefinition{
				{Name: "tools", DisplayName: "Tools"},
			},
		}

		// Initialize maps
		registry.appsByName = make(map[string]*AppDefinition)
		registry.appsByCategory = make(map[string][]*AppDefinition)

		result := v.ValidateRegistry(registry)
		if !result.Valid {
			t.Errorf("Expected valid registry, got errors: %v", result.Errors)
		}
	})

	t.Run("DuplicateAppNames", func(t *testing.T) {
		registry := &AppsRegistry{
			Applications: []AppDefinition{
				{
					Name:        "app1",
					ConfigPaths: []string{"~/.config/app1/config"},
				},
				{
					Name:        "app1", // Duplicate
					ConfigPaths: []string{"~/.config/app1-alt/config"},
				},
			},
		}

		result := v.ValidateRegistry(registry)
		if result.Valid {
			t.Error("Expected invalid due to duplicate app names")
		}

		found := false
		for _, err := range result.Errors {
			if err.Rule == "unique" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected unique constraint error")
		}
	})

	t.Run("UndefinedCategory", func(t *testing.T) {
		registry := &AppsRegistry{
			Applications: []AppDefinition{
				{
					Name:        "app1",
					ConfigPaths: []string{"~/.config/app1/config"},
					Category:    "undefined-category",
				},
			},
			Categories: []CategoryDefinition{
				{Name: "tools", DisplayName: "Tools"},
			},
		}

		// Initialize maps
		registry.appsByName = make(map[string]*AppDefinition)
		registry.appsByCategory = make(map[string][]*AppDefinition)

		result := v.ValidateRegistry(registry)
		// Should still be valid but with warning
		if !result.Valid {
			t.Error("Should be valid with warning about undefined category")
		}

		if len(result.Warnings) == 0 {
			t.Error("Expected warning about undefined category")
		}
	})
}
