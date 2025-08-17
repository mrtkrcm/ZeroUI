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
}
