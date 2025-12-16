package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mrtkrcm/ZeroUI/pkg/legacyextractor"
)

var extractCmd = &cobra.Command{
	Use:   "extract [app|--all]",
	Short: "Extract configuration for app(s)",
	Long:  `Extract configuration from applications using parallel methods (CLI, GitHub, local files).`,
	Example: `  zeroui extract ghostty
  zeroui extract --all
  zeroui extract --apps "ghostty,zed" --output configs`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExtract,
}

var (
	extractAll    bool
	extractOutput string
	extractApps   string
)

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().BoolVar(&extractAll, "all", false, "Extract all known apps")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "configs", "Output directory")
	extractCmd.Flags().StringVar(&extractApps, "apps", "", "Comma-separated app list")
}

func runExtract(cmd *cobra.Command, args []string) error {
	start := time.Now()
	ext := legacyextractor.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Determine apps to extract
	var apps []string
	if extractAll {
		apps = []string{"ghostty", "zed", "alacritty", "wezterm", "tmux", "git", "neovim", "starship"}
	} else if extractApps != "" {
		apps = strings.Split(extractApps, ",")
		for i := range apps {
			apps[i] = strings.TrimSpace(apps[i])
		}
	} else if len(args) > 0 {
		apps = []string{args[0]}
	} else {
		return fmt.Errorf("specify app name, --all, or --apps")
	}

	// Create output directory
	if err := os.MkdirAll(extractOutput, 0o755); err != nil {
		return err
	}

	fmt.Printf("Extracting %d apps using %d workers...\n", len(apps), runtime.NumCPU())

	// Extract all in parallel
	configs := make(map[string]*legacyextractor.Config)
	for _, app := range apps {
		cfg, err := ext.Extract(ctx, app)
		if err != nil {
			fmt.Printf("Failed to extract %s: %v\n", app, err)
			continue
		}
		configs[app] = cfg
	}

	// Save results
	saved := 0
	for app, cfg := range configs {
		path := filepath.Join(extractOutput, app+".yaml")
		if err := saveConfig(cfg, path); err != nil {
			fmt.Printf("  ✗ %s: %v\n", app, err)
		} else {
			fmt.Printf("  ✓ %s: %d settings\n", app, len(cfg.Settings))
			saved++
		}
	}

	fmt.Printf("\nExtracted %d/%d apps in %v\n", saved, len(apps), time.Since(start).Round(time.Millisecond))
	return nil
}

func saveConfig(cfg *legacyextractor.Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
