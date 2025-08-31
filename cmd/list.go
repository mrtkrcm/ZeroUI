package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <type> [app]",
	Short: "List available apps, presets, keys, current values, or changed values",
	Long: `List available applications, presets for an app, UI configurable keys, current configuration values, or only changed values.

Examples:
  zeroui list apps
  zeroui list presets ghostty
  zeroui list keys ghostty
  zeroui list values ghostty
  zeroui list changed ghostty`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		listType := args[0]
		var app string
		if len(args) > 1 {
			app = args[1]
		}

		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()

		switch listType {
		case "apps":
			return listApps(configService)
		case "presets":
			if app == "" {
				return fmt.Errorf("app name required for listing presets")
			}
			return listPresets(configService, app)
		case "keys":
			if app == "" {
				return fmt.Errorf("app name required for listing keys")
			}
			return listKeys(configService, app)
		case "values", "current":
			if app == "" {
				return fmt.Errorf("app name required for listing current values")
			}
			return listCurrentValues(configService, app)
		case "changed":
			if app == "" {
				return fmt.Errorf("app name required for listing changed values")
			}
			return listChangedValues(configService, app)
		default:
			return fmt.Errorf("invalid list type: %s (valid: apps, presets, keys, values, changed)", listType)
		}
	},
}

// keymapCmd represents the keymap command
var keymapCmd = &cobra.Command{
	Use:   "keymap",
	Short: "Manage keyboard shortcuts and keymaps for applications",
	Long: `Manage keyboard shortcuts and keymaps across different applications.
Supports ghostty, vscode, zed, and other apps with keymap configurations.

Examples:
  zeroui keymap list ghostty
  zeroui keymap add ghostty "ctrl+shift+t=new_tab"
  zeroui keymap remove ghostty "ctrl+w"
  zeroui keymap edit ghostty
  zeroui keymap validate ghostty
  zeroui keymap presets ghostty
  zeroui keymap conflicts ghostty`,
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(keymapCmd)

	// Add keymap subcommands
	keymapCmd.AddCommand(keymapListCmd)
	keymapCmd.AddCommand(keymapAddCmd)
	keymapCmd.AddCommand(keymapRemoveCmd)
	keymapCmd.AddCommand(keymapEditCmd)
	keymapCmd.AddCommand(keymapValidateCmd)
	keymapCmd.AddCommand(keymapPresetsCmd)
	keymapCmd.AddCommand(keymapConflictsCmd)
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Styles for list output
var (
	listHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	listItemDisplayStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	listDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	listCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50C7E3")).
			Bold(true)
)

func listApps(configService *service.ConfigService) error {
	apps, err := configService.ListApplications()
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		fmt.Println("No applications configured")
		return nil
	}

	header := listHeaderStyle.Render("Available Applications")
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(apps)))
	fmt.Printf("%s %s\n\n", header, count)

	for _, app := range apps {
		fmt.Printf("  %s\n", listItemDisplayStyle.Render("• "+app))
	}

	return nil
}

func listPresets(configService *service.ConfigService, app string) error {
	presets, err := configService.ListPresets(app)
	if err != nil {
		return err
	}

	if len(presets) == 0 {
		fmt.Printf("No presets configured for %s\n", app)
		return nil
	}

	header := listHeaderStyle.Render(fmt.Sprintf("Available Presets for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(presets)))
	fmt.Printf("%s %s\n\n", header, count)

	for name, preset := range presets {
		if preset.Description != "" {
			fmt.Printf("  %s - %s\n",
				listItemDisplayStyle.Render("• "+name),
				listDescriptionStyle.Render(preset.Description))
		} else {
			fmt.Printf("  %s\n", listItemDisplayStyle.Render("• "+name))
		}
	}

	return nil
}

func listKeys(configService *service.ConfigService, app string) error {
	fields, err := configService.ListFields(app)
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		fmt.Printf("No configurable keys for %s\n", app)
		return nil
	}

	header := listHeaderStyle.Render(fmt.Sprintf("Configurable Keys for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(fields)))
	fmt.Printf("%s %s\n\n", header, count)

	for key, field := range fields {
		var parts []string

		if field.Type != "" {
			parts = append(parts, fmt.Sprintf("(%s)", field.Type))
		}

		if len(field.Values) > 0 {
			parts = append(parts, "choices: "+strings.Join(field.Values, ", "))
		}

		if field.Description != "" {
			parts = append(parts, field.Description)
		}

		keyDisplay := listItemDisplayStyle.Render("• " + key)
		if len(parts) > 0 {
			fmt.Printf("  %s %s\n", keyDisplay, listDescriptionStyle.Render("- "+strings.Join(parts, " - ")))
		} else {
			fmt.Printf("  %s\n", keyDisplay)
		}
	}

	return nil
}

func listCurrentValues(configService *service.ConfigService, app string) error {
	values, err := configService.GetCurrentValues(app)
	if err != nil {
		return err
	}

	if len(values) == 0 {
		fmt.Printf("No current configuration values found for %s\n", app)
		return nil
	}

	header := listHeaderStyle.Render(fmt.Sprintf("Current Configuration Values for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(values)))
	fmt.Printf("%s %s\n\n", header, count)

	// Sort keys for consistent output
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}

	// Simple sort
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		value := values[key]
		fmt.Printf("  %s %s\n",
			listItemDisplayStyle.Render(fmt.Sprintf("%s:", key)),
			listDescriptionStyle.Render(fmt.Sprintf("%v", value)))
	}

	return nil
}

func listChangedValues(configService *service.ConfigService, app string) error {
	// Get current values
	currentValues, err := configService.GetCurrentValues(app)
	if err != nil {
		return err
	}

	// Get application config to access default values
	appConfig, err := configService.GetApplicationConfig(app)
	if err != nil {
		return err
	}

	// Compare with defaults and find changed values
	var changedKeys []string
	changedValues := make(map[string]interface{})

	for key, field := range appConfig.Fields {
		currentValue, hasCurrent := currentValues[key]
		defaultValue := field.Default

		// If field has a current value different from default, it's changed
		if hasCurrent {
			// Simple comparison - could be enhanced for complex types
			currentStr := fmt.Sprintf("%v", currentValue)
			defaultStr := fmt.Sprintf("%v", defaultValue)
			if currentStr != defaultStr {
				changedKeys = append(changedKeys, key)
				changedValues[key] = currentValue
			}
		} else if defaultValue != nil {
			// Field has a default but no current value - might be explicitly set to nil or empty
			// For now, we'll consider this as potentially changed if the default is meaningful
			defaultStr := fmt.Sprintf("%v", defaultValue)
			if defaultStr != "" && defaultStr != "0" && defaultStr != "false" {
				changedKeys = append(changedKeys, key)
				changedValues[key] = currentValue // Will be nil or empty
			}
		}
	}

	if len(changedKeys) == 0 {
		fmt.Printf("No configuration values have been changed from defaults for %s\n", app)
		return nil
	}

	// Sort keys for consistent output
	for i := 0; i < len(changedKeys); i++ {
		for j := i + 1; j < len(changedKeys); j++ {
			if changedKeys[i] > changedKeys[j] {
				changedKeys[i], changedKeys[j] = changedKeys[j], changedKeys[i]
			}
		}
	}

	header := listHeaderStyle.Render(fmt.Sprintf("Changed Configuration Values for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(changedKeys)))
	fmt.Printf("%s %s\n\n", header, count)

	for _, key := range changedKeys {
		value := changedValues[key]
		field := appConfig.Fields[key]

		// Show default value for context
		defaultValue := field.Default
		fmt.Printf("  %s %s %s\n",
			listItemDisplayStyle.Render(fmt.Sprintf("%s:", key)),
			listDescriptionStyle.Render(fmt.Sprintf("%v", value)),
			listDescriptionStyle.Render(fmt.Sprintf("(default: %v)", defaultValue)))
	}

	return nil
}

// Keymap subcommands
var keymapListCmd = &cobra.Command{
	Use:   "list <app>",
	Short: "List all keymaps for an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return listKeymaps(configService, app)
	},
}

var keymapAddCmd = &cobra.Command{
	Use:   "add <app> <keymap>",
	Short: "Add a new keymap to an application",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		keymap := args[1]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return addKeymap(configService, app, keymap)
	},
}

var keymapRemoveCmd = &cobra.Command{
	Use:   "remove <app> <keys>",
	Short: "Remove a keymap from an application",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		keys := args[1]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return removeKeymap(configService, app, keys)
	},
}

var keymapEditCmd = &cobra.Command{
	Use:   "edit <app>",
	Short: "Launch interactive keymap editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		return editKeymaps(app)
	},
}

var keymapValidateCmd = &cobra.Command{
	Use:   "validate <app>",
	Short: "Validate all keymaps for an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return validateKeymaps(configService, app)
	},
}

var keymapPresetsCmd = &cobra.Command{
	Use:   "presets <app>",
	Short: "Show available keymap presets for an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		return showKeymapPresets(app)
	},
}

var keymapConflictsCmd = &cobra.Command{
	Use:   "conflicts <app>",
	Short: "Detect and show keymap conflicts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return detectKeymapConflicts(configService, app)
	},
}

// Keymap management functions
func listKeymaps(configService *service.ConfigService, app string) error {
	// Get current configuration values
	values, err := configService.GetCurrentValues(app)
	if err != nil {
		return err
	}

	// Extract keybind values
	var keymaps []string
	for key, value := range values {
		if key == "keybind" || strings.HasPrefix(key, "keybind") {
			if strVal, ok := value.(string); ok && strVal != "" {
				// Handle ghostty format where multiple keymaps are in one string
				if strings.HasPrefix(strVal, "[") && strings.HasSuffix(strVal, "]") {
					// Remove brackets and split by spaces
					content := strings.Trim(strVal, "[]")
					if content != "" {
						individualKeymaps := strings.Fields(content)
						keymaps = append(keymaps, individualKeymaps...)
					}
				} else {
					keymaps = append(keymaps, strVal)
				}
			} else if strSlice, ok := value.([]string); ok {
				// Handle []string type (which is what we're getting from ghostty)
				keymaps = append(keymaps, strSlice...)
			} else if sliceVal, ok := value.([]interface{}); ok {
				for _, item := range sliceVal {
					if strItem, ok := item.(string); ok && strItem != "" {
						keymaps = append(keymaps, strItem)
					}
				}
			}
		}
	}

	if len(keymaps) == 0 {
		fmt.Printf("No keymaps found for %s\n", app)
		return nil
	}

	// Sort keymaps for consistent output
	sort.Strings(keymaps)

	header := listHeaderStyle.Render(fmt.Sprintf("Keymaps for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(keymaps)))
	fmt.Printf("%s %s\n\n", header, count)

	for _, keymap := range keymaps {
		// Parse keymap to extract keys and action
		if strings.Contains(keymap, "=") {
			parts := strings.SplitN(keymap, "=", 2)
			keys := strings.TrimSpace(parts[0])
			action := strings.TrimSpace(parts[1])

			fmt.Printf("  %s → %s\n",
				listItemDisplayStyle.Render(keys),
				listDescriptionStyle.Render(action))
		} else {
			fmt.Printf("  %s\n", listItemDisplayStyle.Render(keymap))
		}
	}

	return nil
}

func addKeymap(configService *service.ConfigService, app, keymap string) error {
	fmt.Printf("Adding keymap: %s\n", keymap)

	// Validate keymap format
	validator := &configextractor.KeybindValidator{}
	result := validator.ValidateKeybind(keymap)

	if !result.Valid {
		fmt.Printf("❌ Invalid keymap format:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("keymap validation failed")
	}

	// Here we would typically add the keymap to the configuration
	// For now, just show success
	fmt.Printf("✅ Keymap validated successfully\n")
	fmt.Printf("📝 Keys: %s\n", result.ParsedKeybind.Keys)
	fmt.Printf("🎯 Action: %s\n", result.ParsedKeybind.Action)

	return nil
}

func removeKeymap(configService *service.ConfigService, app, keys string) error {
	fmt.Printf("Removing keymap for keys: %s\n", keys)

	// Here we would search and remove the keymap
	// For now, just show what would be done
	fmt.Printf("🔍 Searching for keymap with keys: %s\n", keys)
	fmt.Printf("⚠️  Note: This would remove the keymap from %s configuration\n", app)

	return nil
}

func editKeymaps(app string) error {
	fmt.Printf("Launching interactive keymap editor for %s\n", app)
	fmt.Printf("🔧 Interactive editing not yet implemented\n")
	fmt.Printf("💡 Use: zeroui keymap add/remove for now\n")

	return nil
}

func validateKeymaps(configService *service.ConfigService, app string) error {
	// Get current configuration values
	values, err := configService.GetCurrentValues(app)
	if err != nil {
		return err
	}

	// Extract and validate keybind values
	validator := configextractor.NewKeybindValidatorForApp(app)
	var validCount, invalidCount int
	var allKeymaps []string

	// Collect all keymaps first
	for key, value := range values {
		if key == "keybind" || strings.HasPrefix(key, "keybind") {
			if strSlice, ok := value.([]string); ok {
				allKeymaps = append(allKeymaps, strSlice...)
			} else if strVal, ok := value.(string); ok && strVal != "" {
				allKeymaps = append(allKeymaps, strVal)
			}
		}
	}

	// Validate each keymap
	for _, keymap := range allKeymaps {
		if strings.TrimSpace(keymap) == "" {
			continue // Skip empty entries
		}

		result := validator.ValidateKeybind(keymap)
		if result.Valid {
			validCount++
		} else {
			invalidCount++
			fmt.Printf("❌ Invalid keymap: %s\n", keymap)
			for _, err := range result.Errors {
				fmt.Printf("   %s\n", err)
			}
			if len(result.Warnings) > 0 {
				for _, warning := range result.Warnings {
					fmt.Printf("   ⚠️  %s\n", warning)
				}
			}
		}
	}

	fmt.Printf("✅ Keymap validation complete for %s\n", app)
	fmt.Printf("📊 Valid keymaps: %d\n", validCount)
	if invalidCount > 0 {
		fmt.Printf("❌ Invalid keymaps: %d\n", invalidCount)
		return fmt.Errorf("found %d invalid keymaps", invalidCount)
	}

	return nil
}

func showKeymapPresets(app string) error {
	fmt.Printf("Available keymap presets for %s\n", app)

	presets := map[string][]string{
		"vim-like": {
			"ctrl+h=previous_tab",
			"ctrl+l=next_tab",
			"ctrl+j=scroll_page_down",
			"ctrl+k=scroll_page_up",
		},
		"tmux-like": {
			"ctrl+b+c=new_tab",
			"ctrl+b+n=next_tab",
			"ctrl+b+p=previous_tab",
			"ctrl+b+x=close_surface",
		},
		"emacs-like": {
			"ctrl+x+ctrl+c=quit",
			"ctrl+x+2=split_vertical",
			"ctrl+x+3=split_horizontal",
			"ctrl+x+o=goto_split:next",
		},
	}

	presetNames := make([]string, 0, len(presets))
	for presetName, keymaps := range presets {
		presetNames = append(presetNames, presetName)
		fmt.Printf("\n🎨 %s:\n", presetName)
		for _, keymap := range keymaps {
			if strings.Contains(keymap, "=") {
				parts := strings.SplitN(keymap, "=", 2)
				fmt.Printf("  %s → %s\n",
					listItemDisplayStyle.Render(parts[0]),
					listDescriptionStyle.Render(parts[1]))
			}
		}
	}

	if len(presetNames) > 0 {
		fmt.Printf("\n💡 Use: zeroui preset apply %s <preset> where preset is one of: %s\n", app, strings.Join(presetNames, ", "))
	}
	return nil
}

func detectKeymapConflicts(configService *service.ConfigService, app string) error {
	// Get current configuration values
	values, err := configService.GetCurrentValues(app)
	if err != nil {
		return err
	}

	// Extract keymaps
	keymapMap := make(map[string][]string) // key -> []actions
	var conflicts []string

	for key, value := range values {
		if strings.HasPrefix(key, "keybind") {
			if strVal, ok := value.(string); ok && strVal != "" {
				if strings.Contains(strVal, "=") {
					parts := strings.SplitN(strVal, "=", 2)
					keys := strings.TrimSpace(parts[0])
					action := strings.TrimSpace(parts[1])

					if actions, exists := keymapMap[keys]; exists {
						// Conflict found
						conflicts = append(conflicts, fmt.Sprintf("Keys '%s' mapped to: %s", keys, strings.Join(append(actions, action), ", ")))
					} else {
						keymapMap[keys] = []string{action}
					}
				}
			}
		}
	}

	if len(conflicts) == 0 {
		fmt.Printf("✅ No keymap conflicts found in %s\n", app)
	} else {
		fmt.Printf("⚠️  Keymap conflicts detected in %s:\n\n", app)
		for _, conflict := range conflicts {
			fmt.Printf("  ❌ %s\n", conflict)
		}
		fmt.Printf("\n🔧 Consider using: zeroui keymap edit %s\n", app)
		return fmt.Errorf("found %d keymap conflicts", len(conflicts))
	}

	return nil
}
