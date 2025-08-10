package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/mrtkrcm/ZeroUI/internal/tui"
)

// uiImprovedCmd represents the improved ui command
var uiImprovedCmd = &cobra.Command{
	Use:   "ui-improved [app]",
	Short: "Launch improved interactive terminal user interface",
	Long: `Launch an improved interactive terminal user interface (TUI) for managing 
application configurations. This version uses better patterns from the Charm libraries.

If an application name is provided, the UI will start directly in 
configuration mode for that application.

Examples:
  zeroui ui-improved
  zeroui ui-improved ghostty`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		var initialApp string
		if len(args) > 0 {
			initialApp = args[0]
		}

		// Create improved TUI app with injected dependencies
		app := tui.NewImprovedApp(container.ConfigService(), container.Logger())
		
		// Set initial app if provided
		if initialApp != "" {
			container.Logger().Debug("Starting TUI with initial app", map[string]interface{}{
				"app": initialApp,
			})
		}

		return app.Run()
	},
}

func init() {
	rootCmd.AddCommand(uiImprovedCmd)
}