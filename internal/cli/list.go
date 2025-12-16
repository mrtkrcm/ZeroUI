package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/container"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/spf13/cobra"
)

func newListCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <type> [app]",
		Short: "List available apps, presets, keys, values, or changed settings",
		Long: `List available applications, presets for an app, UI configurable keys,
current configuration values, or changed settings.

Examples:
  zeroui list apps
  zeroui list presets ghostty
  zeroui list keys ghostty
  zeroui list values ghostty
  zeroui list changed ghostty`,
		Example: `  zeroui list apps
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

			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
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
			case "values":
				if app == "" {
					return fmt.Errorf("app name required for listing values")
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
	return cmd
}

func newKeymapCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	keymapCmd := &cobra.Command{
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
		Example: `  zeroui keymap list ghostty
  zeroui keymap add ghostty "ctrl+shift+t=new_tab"
  zeroui keymap remove ghostty "ctrl+w"
  zeroui keymap edit ghostty
  zeroui keymap validate ghostty
  zeroui keymap presets ghostty
  zeroui keymap conflicts ghostty`,
		Args: cobra.NoArgs,
	}

	// Add keymap subcommands
	keymapCmd.AddCommand(newKeymapListCmd(getContainer))
	keymapCmd.AddCommand(newKeymapAddCmd(getContainer))
	keymapCmd.AddCommand(newKeymapRemoveCmd(getContainer))
	keymapCmd.AddCommand(newKeymapEditCmd(getContainer))
	keymapCmd.AddCommand(newKeymapValidateCmd(getContainer))
	keymapCmd.AddCommand(newKeymapPresetsCmd(getContainer))
	keymapCmd.AddCommand(newKeymapConflictsCmd(getContainer))
	return keymapCmd
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
func newKeymapAddCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "add <app> <keymap>",
		Short: "Add a new keymap to an application",
		Example: `  zeroui keymap add ghostty "ctrl+shift+t=new_tab"
  zeroui keymap add zed "cmd+b=toggle_sidebar"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			keymap := args[1]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			configService := container.ConfigService()
			return addKeymap(container, app, keymap)
		},
	}
}

func newKeymapConflictsCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "conflicts <app>",
		Short: "Detect and show keymap conflicts",
		Example: `  zeroui keymap conflicts ghostty
  zeroui keymap conflicts vscode`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			return detectKeymapConflicts(container, app)
		},
	}
}

func newKeymapEditCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <app>",
		Short: "Launch interactive keymap editor",
		Example: `  zeroui keymap edit ghostty
  zeroui keymap edit vscode`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}
			return editKeymaps(app)
		},
	}
}

func newKeymapRemoveCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <app> <keys>",
		Short: "Remove a keymap from an application",
		Example: `  zeroui keymap remove ghostty "ctrl+w"
  zeroui keymap remove vscode "ctrl+shift+p"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			keys := args[1]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			return removeKeymap(container, app, keys)
		},
	}
}

func newKeymapValidateCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "validate <app>",
		Short: "Validate all keymaps for an application",
		Example: `  zeroui keymap validate ghostty
  zeroui keymap validate zed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			return validateKeymaps(container, app)
		},
	}
}

func newKeymapPresetsCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "presets <app>",
		Short: "Show available keymap presets for an application",
		Example: `  zeroui keymap presets ghostty
  zeroui keymap presets wezterm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			return showKeymapPresets(container, app)
		},
	}
}

func newKeymapListCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "list <app>",
		Short: "List all keymaps for an application",
		Example: `  zeroui keymap list ghostty
  zeroui keymap list zed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			return listKeymaps(container, app)
		},
	}
}

// Keymap management functions
func listKeymaps(container *container.Container, app string) error {
	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())
	keymaps, err := kmService.GetKeymapsForApp(app)
	if err != nil {
		return err
	}

	if len(keymaps) == 0 {
		fmt.Printf("No keymaps found for %s\n", app)
		return nil
	}

	header := listHeaderStyle.Render(fmt.Sprintf("Keymaps for %s", app))
	count := listCountStyle.Render(fmt.Sprintf("(%d)", len(keymaps)))
	fmt.Printf("%s %s\n\n", header, count)

	for _, km := range keymaps {
		fmt.Printf("  %s → %s\n",
			listItemDisplayStyle.Render(km.Keys),
			listDescriptionStyle.Render(km.Action))
	}

	return nil
}

func addKeymap(container *container.Container, app, keymap string) error {
	fmt.Printf("Adding keymap: %s\n", keymap)

	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())
	if err := kmService.AddKeymap(app, keymap); err != nil {
		return err
	}

	// Parse keys and action for display
	parts := strings.SplitN(keymap, "=", 2)
	keys := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	fmt.Printf("Keymap added successfully\n")
	fmt.Printf("Keys: %s\n", keys)
	fmt.Printf("Action: %s\n", action)

	return nil
}

func removeKeymap(container *container.Container, app, keys string) error {
	fmt.Printf("Removing keymap for keys: %s\n", keys)

	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())
	if err := kmService.RemoveKeymap(app, keys); err != nil {
		return err
	}

	fmt.Printf("Keymap removed successfully from %s configuration\n", app)
	return nil
}

func editKeymaps(app string) error {
	fmt.Printf("Launching interactive keymap editor for %s\n", app)
	fmt.Printf("Interactive editing not yet implemented\n")
	fmt.Printf("Use: zeroui keymap add/remove for now\n")

	return nil
}

func validateKeymaps(container *container.Container, app string) error {
	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())

	validCount, invalidCount, errors, err := kmService.ValidateAllKeymaps(app)
	if err != nil {
		return err
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			fmt.Printf("  %s\n", errMsg)
		}
	}

	fmt.Printf("Keymap validation complete for %s\n", app)
	fmt.Printf("Valid keymaps: %d\n", validCount)
	if invalidCount > 0 {
		fmt.Printf("Invalid keymaps: %d\n", invalidCount)
		return fmt.Errorf("found %d invalid keymaps", invalidCount)
	}

	return nil
}

func showKeymapPresets(container *container.Container, app string) error {
	fmt.Printf("Available keymap presets for %s\n", app)

	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())
	presets := kmService.GetKeymapPresets(app)

	// Convert map to sorted slice of names for consistent output
	var presetNames []string
	for name := range presets {
		presetNames = append(presetNames, name)
	}
	sort.Strings(presetNames)

	for _, presetName := range presetNames {
		preset := presets[presetName]

		fmt.Printf("\n%s", presetName)
		if preset.Description != "" {
			fmt.Printf(" - %s", preset.Description)
		}
		fmt.Println(":")

		for _, keymap := range preset.Keymaps {
			if strings.Contains(keymap, "=") {
				parts := strings.SplitN(keymap, "=", 2)
				fmt.Printf("  %s -> %s\n",
					listItemDisplayStyle.Render(parts[0]),
					listDescriptionStyle.Render(parts[1]))
			} else {
				fmt.Printf("  %s\n", listItemDisplayStyle.Render(keymap))
			}
		}
	}

	if len(presetNames) > 0 {
		fmt.Printf("\nUse: zeroui preset apply %s <preset> where preset is one of: %s\n", app, strings.Join(presetNames, ", "))
	}
	return nil
}

func detectKeymapConflicts(container *container.Container, app string) error {
	kmService := service.NewKeymapService(container.ConfigService(), container.Logger())

	conflicts, err := kmService.DetectConflicts(app)
	if err != nil {
		return err
	}

	if len(conflicts) == 0 {
		fmt.Printf("No keymap conflicts found in %s\n", app)
	} else {
		fmt.Printf("Keymap conflicts detected in %s:\n\n", app)
		for _, conflict := range conflicts {
			fmt.Printf("  - %s\n", conflict)
		}
		fmt.Printf("\nConsider using: zeroui keymap edit %s\n", app)
		return fmt.Errorf("found %d keymap conflicts", len(conflicts))
	}

	return nil
}
