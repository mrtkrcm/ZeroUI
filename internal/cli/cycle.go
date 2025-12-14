package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cycleCmd represents the cycle command
var cycleCmd = &cobra.Command{
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

		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		configService := container.ConfigService()
		return configService.CycleConfiguration(app, key)
	},
}

func init() {
	rootCmd.AddCommand(cycleCmd)
}
