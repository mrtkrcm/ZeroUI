package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

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
	cfgFile       string
	appContainer  *container.Container
	containerOnce sync.Once
	cleanupMu     sync.Mutex
	cleanupHooks  []func()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zeroui",
	Short: "Zero-configuration UI toolkit manager for developers",
	Long: `ZeroUI is a zero-configuration UI toolkit manager that simplifies managing
UI configurations, themes, and settings across development tools and applications.
Built for speed and simplicity with both CLI and interactive TUI interfaces.`,
	Example: `  zeroui                              # Launch interactive app grid
  zeroui toggle ghostty theme dark    # Direct configuration toggle
  zeroui cycle alacritty font         # Cycle through font options
  zeroui ui ghostty                   # Launch app-specific UI
  zeroui preset vscode minimal        # Apply configuration preset`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: false,
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
			container := GetContainer()
			tuiApp, err := tui.NewApp(container, "")
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
func ExecuteWithContext(ctx context.Context) error {
	return ExecuteContext(ctx, nil)
}

// ExecuteContext executes the root command with optional arguments
func ExecuteContext(ctx context.Context, args []string) error {
	defer func() {
		if appContainer != nil {
			if err := appContainer.Close(); err != nil {
				logger.Error("Failed to close application container", err)
			}
		}
		runCleanupHooks()
	}()

	// Set the context on the root command for propagation to subcommands
	rootCmd.SetContext(ctx)
	if args != nil {
		rootCmd.SetArgs(args)
	}

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		// Check if error is due to context cancellation (graceful shutdown)
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			logger.Info("Application shutdown requested")
			return context.Canceled
		}
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/zeroui/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "show what would be changed without making changes")

	// Runtime config flags (for future use with runtime config loader)
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "log format (text, json)")
	rootCmd.PersistentFlags().String("default-theme", "modern", "default theme (modern, dracula)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("default-theme", rootCmd.PersistentFlags().Lookup("default-theme"))

	// Add command tracing
	attachCommandTracing(rootCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".zeroui" (without extension).
		viper.AddConfigPath(home + "/.config/zeroui")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	// Load runtime configuration
	loader := runtimeconfig.NewLoader(viper.GetViper())
	cfg, err := loader.Load(cfgFile, rootCmd.PersistentFlags())
	if err != nil {
		// If runtime config loading fails, fall back to defaults
		// This ensures backward compatibility
		logger.Global().Warn("Failed to load runtime config, falling back to defaults", logger.Field{Key: "error", Value: err})
		cfg = &runtimeconfig.Config{
			LogLevel:     "info",
			LogFormat:    "text",
			DefaultTheme: "modern",
		}
	}

	// Map "text" format to "console" for logger
	logFormat := cfg.LogFormat
	if logFormat == "text" {
		logFormat = "console"
	}

	// Initialize global logger with runtime config settings
	logger.InitGlobal(&logger.Config{
		Level:      cfg.LogLevel,
		Format:     logFormat,
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
	})

	// Set theme from runtime config
	// Map "default" to "modern" for backward compatibility
	themeName := cfg.DefaultTheme
	if themeName == "default" {
		themeName = "modern"
	}

	if _, ok := styles.SetThemeByName(themeName); !ok {
		// Fall back to modern theme if the configured theme doesn't exist
		logger.Global().Warn("Theme not found, falling back to modern theme",
			logger.Field{Key: "theme", Value: cfg.DefaultTheme},
		)
		styles.SetThemeByName("modern")
	}
}

// GetContainer returns the application container (for use in subcommands)
// The container is lazily initialized on first access to ensure it's created
// after initConfig has run and the global logger is properly configured.
func GetContainer() *container.Container {
	containerOnce.Do(func() {
		var err error
		appContainer, err = container.New(nil)
		if err != nil {
			logger.Fatal("Failed to initialize application container", err)
		}
	})
	return appContainer
}

// RegisterCleanupHook adds a function to run after command execution completes
func RegisterCleanupHook(hook func()) {
	if hook == nil {
		return
	}

	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	cleanupHooks = append(cleanupHooks, hook)
}

func runCleanupHooks() {
	cleanupMu.Lock()
	hooks := cleanupHooks
	cleanupHooks = nil
	cleanupMu.Unlock()

	for _, hook := range hooks {
		hook()
	}
}

// attachCommandTracing adds execution tracing to commands for observability
func attachCommandTracing(cmd *cobra.Command) {
	// Store the original PersistentPreRunE if it exists
	originalPreRunE := cmd.PersistentPreRunE

	// Wrap PersistentPreRunE to log command start
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Create logger with request context - done here after initConfig has run
		// so the logger uses the correct format from runtime config
		requestID := fmt.Sprintf("cmd-%d", time.Now().UnixNano())
		cmdLogger := logger.Global().WithRequest(requestID)

		// Update the command context with the logger
		ctx := logger.ContextWithLogger(cmd.Context(), cmdLogger)
		cmd.SetContext(ctx)

		// Log command execution start (debug level to avoid cluttering output)
		cmdLogger.Debug("Command execution started",
			logger.Field{Key: "command", Value: cmd.CommandPath()},
			logger.Field{Key: "args", Value: args},
		)

		// Call original PreRunE if it exists
		if originalPreRunE != nil {
			return originalPreRunE(cmd, args)
		}
		return nil
	}

	// Store the original PersistentPostRunE if it exists
	originalPostRunE := cmd.PersistentPostRunE

	// Wrap PersistentPostRunE to log command completion
	cmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		// Get logger from context
		cmdLogger := logger.FromContext(cmd.Context())

		// Log command execution end (debug level to avoid cluttering output)
		cmdLogger.Debug("Command execution completed",
			logger.Field{Key: "command", Value: cmd.CommandPath()},
		)

		// Call original PostRunE if it exists
		if originalPostRunE != nil {
			return originalPostRunE(cmd, args)
		}
		return nil
	}

	// Recursively attach tracing to all subcommands
	for _, subCmd := range cmd.Commands() {
		attachCommandTracing(subCmd)
	}
}
