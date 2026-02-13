package cli

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/container"
	"github.com/spf13/cobra"
)

func newCycleCmd(getContainer func() (*container.Container, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "cycle <app> <key>",
		Short: "Cycle to the next value for a UI configuration key",
		Long: `Cycle through predefined values for a UI configuration key.
The next value in the sequence will be selected, wrapping around to the first
value after the last one.

Examples:
  zeroui cycle ghostty theme
  zeroui cycle alacritty font
  zeroui cycle vscode colorTheme`,
		Example: `  zeroui cycle ghostty theme
  zeroui cycle alacritty font
  zeroui cycle vscode colorTheme`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := args[0]
			key := args[1]

			container, err := getContainer()
			if err != nil {
				return fmt.Errorf("failed to get container: %w", err)
			}
			if container == nil {
				return fmt.Errorf("application container not initialized")
			}

			configService := container.ConfigService()
			return configService.CycleConfiguration(app, key)
		},
	}
}
