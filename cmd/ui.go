package cmd

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/tui"
	"github.com/spf13/cobra"
)

// uiCmd represents the ui command
var uiCmd = &cobra.Command{
	Use:   "ui [app]",
	Short: "Launch interactive TUI for UI configuration management",
	Long: `Launch an interactive terminal user interface for managing UI configurations.
The TUI provides a visual way to browse, edit, and toggle configuration values
with real-time preview and easy navigation.

Examples:
  zeroui ui
  zeroui ui ghostty
  zeroui ui --app alacritty`,
	Example: `  zeroui ui
  zeroui ui ghostty
  zeroui ui --app alacritty`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var app string
		if len(args) > 0 {
			app = args[0]
		}

		appFlag, err := cmd.Flags().GetString("app")
		if err != nil {
			return err
		}

		if appFlag != "" {
			app = appFlag
		}

		tuiApp, err := tui.NewApp(app)
		if err != nil {
			return fmt.Errorf("failed to create TUI app: %w", err)
		}

		return tuiApp.Run()
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().StringP("app", "a", "", "start with specific app selected")
}
