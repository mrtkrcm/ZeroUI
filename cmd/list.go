package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <type> [app]",
	Short: "List available apps, presets, or UI configuration keys",
	Long: `List available applications, presets for an app, or UI configurable keys.

Examples:
  zeroui list apps
  zeroui list presets ghostty
  zeroui list keys ghostty`,
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
		default:
			return fmt.Errorf("invalid list type: %s (valid: apps, presets, keys)", listType)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
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
