package cmd

import (
	"fmt"
	"os"

	"github.com/mrtkrcm/ZeroUI/internal/container"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	appContainer *container.Container
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zeroui",
	Short: "Zero-configuration UI toolkit manager for developers",
	Long: `ZeroUI is a zero-configuration UI toolkit manager that simplifies managing
UI configurations, themes, and settings across development tools and applications.
Built for speed and simplicity with both CLI and interactive TUI interfaces.

Examples:
  zeroui toggle ghostty theme dark
  zeroui cycle alacritty font
  zeroui ui
  zeroui preset vscode minimal`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Initialize dependency container
	containerConfig := &container.Config{
		LogLevel:  "info",
		LogFormat: "console",
	}
	
	var err error
	appContainer, err = container.New(containerConfig)
	if err != nil {
		logger.Fatal("Failed to initialize application container", err)
	}
	
	defer func() {
		if err := appContainer.Close(); err != nil {
			logger.Error("Failed to close application container", err)
		}
	}()

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/zeroui/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "show what would be changed without making changes")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
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
}

// GetContainer returns the application container (for use in subcommands)
func GetContainer() *container.Container {
	return appContainer
}