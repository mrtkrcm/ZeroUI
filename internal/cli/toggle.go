package cli

import (
	"fmt"
	"os"

	"github.com/mrtkrcm/ZeroUI/internal/container"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/spf13/cobra"
)

func newToggleCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "toggle <app> <key> <value>",
		Short: "Toggle a UI configuration value for an application",
		Long: `Toggle a specific UI configuration key to a new value for a given application.

Examples:
  zeroui toggle ghostty theme dark
  zeroui toggle alacritty font "JetBrains Mono"
  zeroui toggle vscode editor.fontSize 14`,
		Example: `  zeroui toggle ghostty theme dark
  zeroui toggle alacritty font "JetBrains Mono"
  zeroui toggle vscode editor.fontSize 14`,
		Args:         cobra.ExactArgs(3),
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			key := args[1]
			value := args[2]

			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			configService := container.ConfigService()
			if err := configService.ToggleConfiguration(app, key, value); err != nil {
				// Check if it's a ZeroUIError for better user experience
				if ctErr, ok := errors.GetZeroUIError(err); ok {
					fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
					return nil // Don't return error to avoid double printing
				}
				return err
			}

			return nil
		},
	}
}
