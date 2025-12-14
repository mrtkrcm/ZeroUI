package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mrtkrcm/ZeroUI/internal/tui"
	"github.com/spf13/cobra"
)

// uiSelectCmd represents the ui-select command
var uiSelectCmd = &cobra.Command{
	Use:   "ui-select",
	Short: "Select and configure UI implementation",
	Long: `Select and configure different UI implementations for the ZeroUI application.

This command allows you to choose between different UI implementations:
- Standard UI: Reliable, well-tested interface
- Enhanced UI: Advanced Bubble Tea integration with professional styling
- Delightful UI: Beautiful animations and visual effects
- Minimal UI: Clean, distraction-free interface

Examples:
  zeroui ui-select                    # Interactive selection
  zeroui ui-select --implementation enhanced
  zeroui ui-select --list             # List available implementations
  zeroui ui-select --current          # Show current implementation`,
	Aliases: []string{"select-ui", "ui-choose", "choose-ui"},
	Example: `  zeroui ui-select
  zeroui ui-select --implementation enhanced
  zeroui ui-select --list
  zeroui ui-select --current`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		implementation, _ := cmd.Flags().GetString("implementation")
		list, _ := cmd.Flags().GetBool("list")
		current, _ := cmd.Flags().GetBool("current")

		// Create UI selector
		selector := tui.NewUISelector()

		// Handle different modes
		if current {
			fmt.Printf("Current UI Implementation: %s\n", selector.GetCurrentImplementation())
			return nil
		}

		if list {
			fmt.Print(selector.GetImplementationSummary())
			return nil
		}

		if implementation != "" {
			// Set specific implementation
			impl := tui.UIImplementation(implementation)
			if err := selector.SetImplementation(impl); err != nil {
				return fmt.Errorf("failed to set UI implementation: %w", err)
			}

			config := selector.GetImplementationConfig(impl)
			if config != nil {
				fmt.Printf("✅ UI implementation set to: %s\n", config.Name)
				fmt.Printf("   %s\n", config.Description)
			}
			return nil
		}

		// Interactive selection
		return runInteractiveSelection(selector)
	},
}

func runInteractiveSelection(selector *tui.UISelector) error {
	fmt.Println("🎨 ZeroUI - UI Implementation Selection")
	fmt.Println("=======================================")
	fmt.Println()

	// Show current implementation
	fmt.Printf("Current: %s\n", selector.GetCurrentImplementation())
	fmt.Println()

	// Show available implementations
	fmt.Println("Available UI Implementations:")
	fmt.Println()

	implementations := selector.GetAvailableImplementations()
	for i, impl := range implementations {
		config := selector.GetImplementationConfig(impl)
		if config != nil {
			current := ""
			if impl == selector.GetCurrentImplementation() {
				current = " (current)"
			}
			fmt.Printf("%d. %s%s\n", i+1, config.Name, current)
			fmt.Printf("   %s\n", config.Description)
			fmt.Printf("   Complexity: %s | Performance: %s\n", config.Complexity, config.Performance)
			fmt.Println()
		}
	}

	// Show recommended
	recommended := selector.GetRecommendedImplementation()
	fmt.Printf("📋 Recommended: %s\n", recommended)
	fmt.Println()

	// Get user input
	fmt.Print("Enter implementation number (or 'q' to quit): ")
	var input string
	fmt.Scanln(&input)

	if strings.ToLower(input) == "q" {
		fmt.Println("Selection cancelled.")
		return nil
	}

	// Parse selection
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(implementations) {
		return fmt.Errorf("invalid selection: %s", input)
	}

	selectedImpl := implementations[index-1]

	// Set the implementation
	if err := selector.SetImplementation(selectedImpl); err != nil {
		return fmt.Errorf("failed to set UI implementation: %w", err)
	}

	// Show detailed information about the selected implementation
	config := selector.GetImplementationConfig(selectedImpl)
	if config != nil {
		fmt.Printf("\n✅ UI Implementation Selected: %s\n", config.Name)
		fmt.Println("=====================================")
		fmt.Printf("Description: %s\n", config.Description)
		fmt.Printf("Complexity: %s\n", config.Complexity)
		fmt.Printf("Performance: %s\n", config.Performance)
		fmt.Println("Features:")
		for _, feature := range config.Features {
			fmt.Printf("  • %s\n", feature)
		}

		// Test if we can create the model
		if _, err := selector.CreateAppModel(""); err != nil {
			fmt.Printf("\n⚠️  Warning: %v\n", err)
		} else {
			fmt.Println("\n✅ Implementation is ready to use!")
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(uiSelectCmd)
	uiSelectCmd.Flags().String("implementation", "", "Set specific UI implementation (standard, enhanced, delightful, minimal)")
	uiSelectCmd.Flags().Bool("list", false, "List all available UI implementations")
	uiSelectCmd.Flags().Bool("current", false, "Show current UI implementation")
}
