package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mrtkrcm/ZeroUI/internal/container"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/runtimeconfig"
	"github.com/mrtkrcm/ZeroUI/internal/tui"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	cfgFile      string
	appContainer *container.Container
	runtimeCfg   *runtimeconfig.Config
	configLoader *runtimeconfig.Loader
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zeroui",
	Short: "Zero-configuration UI toolkit manager for developers",
	Long: `ZeroUI is a zero-configuration UI toolkit manager that simplifies managing
UI configurations, themes, and settings across development tools and applications.
Built for speed and simplicity with both CLI and interactive TUI interfaces.

Examples:
  zeroui                              # Launch interactive app grid
  zeroui toggle ghostty theme dark
  zeroui cycle alacritty font
  zeroui ui ghostty
  zeroui preset vscode minimal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, launch the UI
		if len(args) == 0 && cmd.Flags().NFlag() == 0 {
			// Avoid launching interactive TUI in non-interactive environments (CI/tests)
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprintln(os.Stderr, "Non-interactive environment detected: TUI requires a terminal. Use subcommands (e.g. 'zeroui list apps') or run this command in a terminal to launch the UI. To run non-interactively, use explicit subcommands or flags.")
				// Show help so callers (including CI) won't hang waiting on a UI
				return cmd.Help()
			}

			// Launch the UI without a specific app (show grid)
			// Import the functionality directly instead of calling uiCmd
			tuiApp, err := tui.NewApp("")
			if err != nil {
				return fmt.Errorf("failed to create TUI app: %w", err)
			}
			return tuiApp.RunWithContext(cmd.Context())
		}
		// Show help if arguments are provided but no valid subcommand
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ExecuteWithContext(context.Background())
}

// ExecuteWithContext executes the root command with context support for graceful shutdown
func ExecuteWithContext(ctx context.Context) {
	// Set the context on the root command for propagation to subcommands
	rootCmd.SetContext(ctx)

	err := rootCmd.ExecuteContext(ctx)
	if appContainer != nil {
		_ = appContainer.Close()
	}
	if err != nil {
		// Check if error is due to context cancellation (graceful shutdown)
		if ctx.Err() == context.Canceled {
			logger.Info("Application shutdown requested")
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func init() {
	configLoader = runtimeconfig.NewLoader(viper.GetViper())
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/zeroui/config.yaml)")
	rootCmd.PersistentFlags().String("config-dir", runtimeconfig.DefaultConfigDir(), "directory containing ZeroUI application definitions")
	rootCmd.PersistentFlags().String("log-level", logger.DefaultConfig().Level, "log level (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().String("log-format", logger.DefaultConfig().Format, "log format (console or json)")
	rootCmd.PersistentFlags().String("default-theme", "Modern", "default theme used by the TUI")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "show what would be changed without making changes")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfg, err := configLoader.Load(cfgFile, rootCmd.PersistentFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	runtimeCfg = cfg

	containerConfig := &container.Config{
		LogLevel:  cfg.LogLevel,
		LogFormat: cfg.LogFormat,
	}

	appContainer, err = container.New(containerConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize application container:", err)
		os.Exit(1)
	}

	if _, ok := styles.SetThemeByName(cfg.DefaultTheme); !ok {
		fmt.Fprintf(os.Stderr, "Unknown theme %q; falling back to %s\n", cfg.DefaultTheme, styles.GetCurrentThemeName())
	}
}

// GetContainer returns the application container (for use in subcommands)
func GetContainer() *container.Container {
	return appContainer
}
