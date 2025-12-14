package styles_test

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
)

// Example demonstrating SetThemeByName
func ExampleSetThemeByName() {
	// Set theme by exact name
	theme, ok := styles.SetThemeByName("Modern")
	if ok {
		fmt.Println("Theme set successfully:", theme.Name)
	}

	// Set theme by case-insensitive name
	theme, ok = styles.SetThemeByName("dracula")
	if ok {
		fmt.Println("Theme set successfully:", theme.Name)
	}

	// Try to set invalid theme
	_, ok = styles.SetThemeByName("NonExistent")
	if !ok {
		fmt.Println("Theme not found")
	}

	// Output:
	// Theme set successfully: Modern
	// Theme set successfully: Dracula
	// Theme not found
}

// Example demonstrating GetCurrentThemeName
func ExampleGetCurrentThemeName() {
	styles.SetThemeByName("Modern")
	fmt.Println("Current theme:", styles.GetCurrentThemeName())

	styles.SetThemeByName("Dracula")
	fmt.Println("Current theme:", styles.GetCurrentThemeName())

	// Output:
	// Current theme: Modern
	// Current theme: Dracula
}

// Example demonstrating ListAvailableThemes
func ExampleListAvailableThemes() {
	themes := styles.ListAvailableThemes()
	fmt.Printf("Available themes: %d\n", len(themes))
	for _, name := range themes {
		fmt.Printf("- %s\n", name)
	}

	// Output:
	// Available themes: 5
	// - Modern
	// - Dracula
	// - Light
	// - Nord
	// - Catppuccin
}
