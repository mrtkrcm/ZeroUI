package cli

import (
	"fmt"

	"github.com/mrtkrcm/ZeroUI/internal/tui"
	"github.com/spf13/cobra"
)

// designSystemCmd represents the design-system command
var designSystemCmd = &cobra.Command{
	Use:   "design-system",
	Short: "Launch native design system showcase",
	Long: `Launch an interactive showcase of the ZeroUI design system.
This demonstrates all TUI components, styles, and patterns used in the application
including Bubble Tea components, Lipgloss styles, color schemes, typography,
and interactive elements as they appear in the actual terminal.

The showcase includes:
- Color palette and themes
- Typography and text styles  
- Interactive components (lists, forms, inputs)
- Layout patterns and spacing
- Box-drawing characters and borders
- Loading states and animations
- Error and success states

Examples:
  zeroui design-system
  zeroui design-system --interactive
  zeroui showcase`,
	Aliases: []string{"showcase", "ds", "demo"},
	Example: `  zeroui design-system
  zeroui design-system --interactive
  zeroui showcase`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		container := GetContainer()
		if container == nil {
			return fmt.Errorf("application container not initialized")
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			return err
		}

		// Create the design system showcase TUI
		showcase := tui.NewDesignSystemShowcase(container.Logger(), interactive)

		container.Logger().Info("Starting design system showcase", map[string]interface{}{
			"interactive": interactive,
		})

		return showcase.Run()
	},
}

func init() {
	rootCmd.AddCommand(designSystemCmd)
	designSystemCmd.Flags().BoolP("interactive", "i", true, "enable interactive mode (default: true)")
}
