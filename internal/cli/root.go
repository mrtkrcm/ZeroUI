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
	"github.com/mrtkrcm/ZeroUI/internal/tui"
	"github.com/mrtkrcm/ZeroUI/internal/tui/styles"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// RootCommand encapsulates the root command and its dependencies.
type RootCommand struct {
	cmd           *cobra.Command
	cfgFile       string
	container     *container.Container
	cleanupHooks  []func()
	cleanupMu     sync.Mutex
	containerOnce sync.Once
}

// NewRootCommand creates a new root command with dependencies.
func NewRootCommand() *RootCommand {
	rc := &RootCommand{}

	rc.cmd = &cobra.Command{
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
				container, err := rc.getContainer()
				if err != nil {
					return fmt.Errorf("failed to get container: %w", err)
				}
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

	cobra.OnInitialize(rc.initConfig)

	// Global flags
	rc.cmd.PersistentFlags().StringVar(&rc.cfgFile, "config", "", "config file (default is $HOME/.config/zeroui/config.yaml)")
	rc.cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rc.cmd.PersistentFlags().BoolP("dry-run", "n", false, "show what would be changed without making changes")

	// Runtime config flags (for future use with runtime config loader)
	rc.cmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rc.cmd.PersistentFlags().String("log-format", "text", "log format (text, json)")
	rc.cmd.PersistentFlags().String("default-theme", "modern", "default theme (modern, dracula)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rc.cmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dry-run", rc.cmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("log-level", rc.cmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-format", rc.cmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("default-theme", rc.cmd.PersistentFlags().Lookup("default-theme"))

	// Add command tracing
	attachCommandTracing(rc.cmd)

	return rc
}

func (rc *RootCommand) Execute(ctx context.Context, args []string) error {
	defer func() {
		if rc.container != nil {
			if err := rc.container.Close(); err != nil {
				logger.Error("Failed to close application container", err)
			}
		}
		rc.runCleanupHooks()
	}()

	// Set the context on the root command for propagation to subcommands
	rc.cmd.SetContext(ctx)
	if args != nil {
		rc.cmd.SetArgs(args)
	}

	err := rc.cmd.ExecuteContext(ctx)
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

// initConfig reads in config file and ENV variables if set.
func (rc *RootCommand) initConfig() {
	// Set default values
	viper.SetDefault("log-level", "info")
	viper.SetDefault("log-format", "text")
	viper.SetDefault("default-theme", "modern")

	if rc.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rc.cfgFile)
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

	// Map "text" format to "console" for logger
	logFormat := viper.GetString("log-format")
	if logFormat == "text" {
		logFormat = "console"
	}

	// Validate and initialize global logger
	logLevel := viper.GetString("log-level")
	if _, err := zerolog.ParseLevel(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid log level %q, falling back to 'info'\n", logLevel)
		logLevel = "info"
	}

	logger.InitGlobal(&logger.Config{
		Level:      logLevel,
		Format:     logFormat,
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
	})

	// Set theme from runtime config
	// Map "default" to "modern" for backward compatibility
	themeName := viper.GetString("default-theme")
	if themeName == "default" {
		themeName = "modern"
	}

	if _, ok := styles.SetThemeByName(themeName); !ok {
		// Fall back to modern theme if the configured theme doesn't exist
		fmt.Fprintf(os.Stderr, "Warning: theme %q not found, using modern theme\n", themeName)
		styles.SetThemeByName("modern")
	}
}

func (rc *RootCommand) getContainer() (*container.Container, error) {
	var err error
	rc.containerOnce.Do(func() {
		rc.container, err = container.New()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application container: %w", err)
	}
	return rc.container, nil
}

// AddSubcommands adds all the subcommands to the root command.
func (rc *RootCommand) AddSubcommands() {
	getContainer := func() (*container.Container, error) {
		return rc.getContainer()
	}

	rc.cmd.AddCommand(
		newUICmd(getContainer),
		newToggleCmd(getContainer),
		newListCmd(getContainer),
		newKeymapCmd(getContainer),
		newBackupCmd(),
		newCompletionCmd(rc.cmd),
		newCycleCmd(getContainer),
		newDesignSystemCmd(getContainer),
		newExtractCmd(),
		newPresetCmd(),
		newReferenceImprovedCmd(),
		newValidateReferenceCmd(),
		newVersionCmd(),
	)
}

// RegisterCleanupHook adds a function to run after command execution completes
func (rc *RootCommand) RegisterCleanupHook(hook func()) {
	if hook == nil {
		return
	}

	rc.cleanupMu.Lock()
	defer rc.cleanupMu.Unlock()
	rc.cleanupHooks = append(rc.cleanupHooks, hook)
}

func (rc *RootCommand) runCleanupHooks() {
	rc.cleanupMu.Lock()
	hooks := rc.cleanupHooks
	rc.cleanupHooks = nil
	rc.cleanupMu.Unlock()

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
