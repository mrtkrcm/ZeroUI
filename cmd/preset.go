package cmd

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/spf13/cobra"
)

// presetCmd represents the preset command
var presetCmd = &cobra.Command{
	Use:   "preset <app> <preset-name>",
	Short: "Apply a preset UI configuration to an application",
	Long: `Apply a predefined preset UI configuration that changes multiple values at once.
Presets are defined in the application's configuration file and provide
quick access to common UI configuration combinations.

Examples:
  zeroui preset ghostty dark-mode
  zeroui preset vscode minimal
  zeroui preset alacritty high-contrast`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app := args[0]
		presetName := args[1]

		engine, err := toggle.NewEngine()
		if err != nil {
			return fmt.Errorf("failed to create toggle engine: %w", err)
		}

		return engine.ApplyPreset(app, presetName)
	},
}

func init() {
	rootCmd.AddCommand(presetCmd)
}
