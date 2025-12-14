package styles

import (
	"testing"
)

func TestThemes(t *testing.T) {
	t.Run("All themes have names", func(t *testing.T) {
		themes := AllThemes()

		for _, theme := range themes {
			if theme.Name == "" {
				t.Error("Theme should have a name")
			}
		}
	})

	t.Run("Theme cycling works", func(t *testing.T) {
		themes := AllThemes()

		// Cycle through all themes
		seenThemes := make(map[string]bool)
		for i := 0; i < len(themes)+1; i++ {
			currentTheme := CycleTheme()
			seenThemes[currentTheme.Name] = true
		}

		// Should have seen all themes
		if len(seenThemes) != len(themes) {
			t.Errorf("Expected to cycle through %d themes, saw %d", len(themes), len(seenThemes))
		}

		// Should be back to the initial theme (or close)
		currentTheme := GetCurrentThemeName()
		if currentTheme == "" {
			t.Error("Current theme name should not be empty")
		}
	})

	t.Run("SetTheme works", func(t *testing.T) {
		themes := AllThemes()
		originalTheme := GetCurrentThemeName()

		// Set to a different theme
		for _, theme := range themes {
			if theme.Name != originalTheme {
				SetTheme(theme)
				if GetCurrentThemeName() != theme.Name {
					t.Errorf("Expected theme %s, got %s", theme.Name, GetCurrentThemeName())
				}
				break
			}
		}
	})

	t.Run("GetThemeNames returns correct names", func(t *testing.T) {
		themes := AllThemes()
		names := GetThemeNames()

		if len(names) != len(themes) {
			t.Errorf("Expected %d theme names, got %d", len(themes), len(names))
		}

		// Check that all theme names are present
		themeMap := make(map[string]bool)
		for _, theme := range themes {
			themeMap[theme.Name] = true
		}

		for _, name := range names {
			if !themeMap[name] {
				t.Errorf("Theme name %s not found in themes list", name)
			}
		}
	})

	t.Run("All themes have required colors", func(t *testing.T) {
		themes := AllThemes()

		for _, theme := range themes {
			if theme.Primary == nil {
				t.Errorf("Theme %s missing Primary color", theme.Name)
			}
			if theme.BgBase == nil {
				t.Errorf("Theme %s missing BgBase color", theme.Name)
			}
			if theme.FgBase == nil {
				t.Errorf("Theme %s missing FgBase color", theme.Name)
			}
		}
	})

	t.Run("Themes build styles correctly", func(t *testing.T) {
		themes := AllThemes()

		for _, theme := range themes {
			styles := theme.BuildStyles()

			if styles == nil {
				t.Errorf("Theme %s failed to build styles", theme.Name)
			}

			// Check that styles have expected fields
			if styles.Title.String() == "" {
				// This is expected as styles don't have content by default
			}
		}
	})

	t.Run("SetThemeByName with exact match", func(t *testing.T) {
		// Save original theme
		originalTheme := GetCurrentThemeName()
		defer func() {
			SetThemeByName(originalTheme)
		}()

		// Set theme by exact name
		theme, ok := SetThemeByName("Modern")
		if !ok {
			t.Error("Expected SetThemeByName to return true for 'Modern'")
		}
		if theme.Name != "Modern" {
			t.Errorf("Expected theme name 'Modern', got '%s'", theme.Name)
		}
		if GetCurrentThemeName() != "Modern" {
			t.Errorf("Expected current theme to be 'Modern', got '%s'", GetCurrentThemeName())
		}
	})

	t.Run("SetThemeByName with case-insensitive match", func(t *testing.T) {
		// Save original theme
		originalTheme := GetCurrentThemeName()
		defer func() {
			SetThemeByName(originalTheme)
		}()

		// Set theme with different case
		testCases := []string{"modern", "MODERN", "MoDeRn", "dracula", "DRACULA"}
		for _, testCase := range testCases {
			theme, ok := SetThemeByName(testCase)
			if !ok {
				t.Errorf("Expected SetThemeByName to return true for '%s'", testCase)
			}
			if theme.Name == "" {
				t.Errorf("Expected theme to have a name for input '%s'", testCase)
			}
			// Verify the theme was actually set
			currentName := GetCurrentThemeName()
			if currentName != theme.Name {
				t.Errorf("Expected current theme to be '%s', got '%s'", theme.Name, currentName)
			}
		}
	})

	t.Run("SetThemeByName with invalid name", func(t *testing.T) {
		// Save original theme
		originalTheme := GetCurrentThemeName()
		defer func() {
			SetThemeByName(originalTheme)
		}()

		// Try to set theme with invalid name
		theme, ok := SetThemeByName("NonExistent")
		if ok {
			t.Error("Expected SetThemeByName to return false for invalid theme name")
		}
		if theme.Name != "" {
			t.Errorf("Expected empty theme for invalid name, got '%s'", theme.Name)
		}
		// Current theme should not have changed
		if GetCurrentThemeName() != originalTheme {
			t.Errorf("Expected current theme to remain '%s', got '%s'", originalTheme, GetCurrentThemeName())
		}
	})

	t.Run("ListAvailableThemes returns all themes", func(t *testing.T) {
		themes := ListAvailableThemes()
		expectedThemes := GetThemeNames()

		if len(themes) != len(expectedThemes) {
			t.Errorf("Expected %d themes, got %d", len(expectedThemes), len(themes))
		}

		// Verify all expected themes are present
		themeMap := make(map[string]bool)
		for _, name := range themes {
			themeMap[name] = true
		}

		for _, expected := range expectedThemes {
			if !themeMap[expected] {
				t.Errorf("Expected theme '%s' not found in ListAvailableThemes", expected)
			}
		}
	})

	t.Run("ListAvailableThemes contains Modern and Dracula", func(t *testing.T) {
		themes := ListAvailableThemes()

		hasModern := false
		hasDracula := false

		for _, name := range themes {
			if name == "Modern" {
				hasModern = true
			}
			if name == "Dracula" {
				hasDracula = true
			}
		}

		if !hasModern {
			t.Error("Expected 'Modern' theme in available themes")
		}
		if !hasDracula {
			t.Error("Expected 'Dracula' theme in available themes")
		}
	})

	t.Run("GetCurrentThemeName returns non-empty string", func(t *testing.T) {
		name := GetCurrentThemeName()
		if name == "" {
			t.Error("Expected GetCurrentThemeName to return non-empty string")
		}
	})

	t.Run("GetCurrentThemeName reflects SetThemeByName", func(t *testing.T) {
		// Save original theme
		originalTheme := GetCurrentThemeName()
		defer func() {
			SetThemeByName(originalTheme)
		}()

		// Set to Modern
		SetThemeByName("Modern")
		if GetCurrentThemeName() != "Modern" {
			t.Errorf("Expected current theme to be 'Modern', got '%s'", GetCurrentThemeName())
		}

		// Set to Dracula
		SetThemeByName("Dracula")
		if GetCurrentThemeName() != "Dracula" {
			t.Errorf("Expected current theme to be 'Dracula', got '%s'", GetCurrentThemeName())
		}
	})

	t.Run("buildThemeRegistry creates correct mapping", func(t *testing.T) {
		registry := buildThemeRegistry()

		if len(registry) != len(AvailableThemes) {
			t.Errorf("Expected registry to have %d entries, got %d", len(AvailableThemes), len(registry))
		}

		// Verify each theme is in the registry
		for _, theme := range AvailableThemes {
			name := GetThemeName(theme)
			if _, ok := registry[name]; !ok {
				t.Errorf("Expected theme '%s' to be in registry", name)
			}
		}
	})

	t.Run("equalsCaseInsensitive works correctly", func(t *testing.T) {
		testCases := []struct {
			a        string
			b        string
			expected bool
		}{
			{"Modern", "Modern", true},
			{"Modern", "modern", true},
			{"MODERN", "modern", true},
			{"MoDeRn", "mOdErN", true},
			{"Dracula", "dracula", true},
			{"Modern", "Dracula", false},
			{"Modern", "ModernX", false},
			{"", "", true},
			{"a", "A", true},
			{"abc", "ab", false},
		}

		for _, tc := range testCases {
			result := equalsCaseInsensitive(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("equalsCaseInsensitive(%q, %q) = %v, expected %v", tc.a, tc.b, result, tc.expected)
			}
		}
	})
}
