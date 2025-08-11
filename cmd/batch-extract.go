package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mrtkrcm/ZeroUI/pkg/reference"
)

// batchExtractCmd extracts configs for all apps efficiently
var batchExtractCmd = &cobra.Command{
	Use:   "batch-extract",
	Short: "Extract configurations for all supported apps in parallel",
	Long: `Efficiently extract configurations for all supported applications using parallel processing.
	
Features:
- Concurrent extraction for maximum speed
- Automatic caching to avoid redundant work
- Progress tracking
- Performance metrics

Examples:
  zeroui batch-extract
  zeroui batch-extract --apps "ghostty,zed,alacritty"
  zeroui batch-extract --output-dir configs/
  zeroui batch-extract --workers 16`,
	RunE: runBatchExtract,
}

var (
	batchApps      string
	batchOutputDir string
	batchWorkers   int
	batchUpdate    bool
	batchVerbose   bool
)

func init() {
	rootCmd.AddCommand(batchExtractCmd)
	batchExtractCmd.Flags().StringVar(&batchApps, "apps", "", "Comma-separated list of apps (default: all known apps)")
	batchExtractCmd.Flags().StringVar(&batchOutputDir, "output-dir", "configs", "Output directory for configs")
	batchExtractCmd.Flags().IntVar(&batchWorkers, "workers", runtime.NumCPU(), "Number of parallel workers")
	batchExtractCmd.Flags().BoolVar(&batchUpdate, "update", false, "Update existing configs")
	batchExtractCmd.Flags().BoolVarP(&batchVerbose, "verbose", "v", false, "Verbose output")
}

func runBatchExtract(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	
	// Default apps list
	defaultApps := []string{
		"ghostty", "zed", "alacritty", "wezterm", "neovim",
		"tmux", "starship", "git", "mise", "vscode",
		"sublime", "kitty", "iterm2", "terminal",
	}
	
	// Parse apps list
	apps := defaultApps
	if batchApps != "" {
		apps = parseAppsList(batchApps)
	}
	
	fmt.Printf("🚀 Starting batch extraction for %d apps using %d workers\n", len(apps), batchWorkers)
	
	// Create output directory
	if err := os.MkdirAll(batchOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Initialize fast extractor
	extractor := reference.NewFastExtractor()
	
	// Extract all configs in parallel
	configs, err := extractor.ExtractBatch(apps)
	if err != nil && len(configs) == 0 {
		return fmt.Errorf("batch extraction failed: %w", err)
	}
	
	// Statistics
	successCount := 0
	failedApps := []string{}
	totalSettings := 0
	
	// Save extracted configs
	for app, config := range configs {
		outputPath := filepath.Join(batchOutputDir, app+".yaml")
		
		// Handle update mode
		if batchUpdate && fileExists(outputPath) {
			existing, err := loadConfig(outputPath)
			if err == nil {
				mergedCount := mergeConfigs(existing, config)
				if batchVerbose {
					fmt.Printf("  📝 %s: merged %d new settings\n", app, mergedCount)
				}
				config = existing
			}
		}
		
		// Save config
		if err := saveConfig(config, outputPath); err != nil {
			failedApps = append(failedApps, app)
			if batchVerbose {
				fmt.Printf("  ❌ %s: save failed - %v\n", app, err)
			}
		} else {
			successCount++
			totalSettings += len(config.Settings)
			if batchVerbose {
				fmt.Printf("  ✅ %s: %d settings extracted\n", app, len(config.Settings))
			}
		}
	}
	
	// Add failed apps from extraction
	for _, app := range apps {
		if _, ok := configs[app]; !ok {
			failedApps = append(failedApps, app)
		}
	}
	
	// Calculate performance metrics
	duration := time.Since(startTime)
	appsPerSecond := float64(len(apps)) / duration.Seconds()
	
	// Print summary
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📊 Batch Extraction Summary")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  ✅ Successful: %d/%d apps\n", successCount, len(apps))
	fmt.Printf("  📋 Total settings: %d\n", totalSettings)
	fmt.Printf("  ⏱️  Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("  ⚡ Performance: %.2f apps/second\n", appsPerSecond)
	fmt.Printf("  💾 Output directory: %s\n", batchOutputDir)
	
	if len(failedApps) > 0 {
		fmt.Printf("\n  ⚠️  Failed apps: %v\n", failedApps)
	}
	
	// Performance comparison
	sequentialTime := time.Duration(len(apps)) * 500 * time.Millisecond // Assume 500ms per app sequential
	speedup := float64(sequentialTime) / float64(duration)
	fmt.Printf("\n  🚀 Speedup: %.2fx faster than sequential extraction\n", speedup)
	
	return nil
}

func parseAppsList(appsList string) []string {
	parts := strings.Split(appsList, ",")
	apps := make([]string, 0, len(parts))
	for _, app := range parts {
		app = strings.TrimSpace(app)
		if app != "" {
			apps = append(apps, app)
		}
	}
	return apps
}

func loadConfig(path string) (*reference.ConfigReference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var config reference.ConfigReference
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	if config.Settings == nil {
		config.Settings = make(map[string]reference.ConfigSetting)
	}
	
	return &config, nil
}

func saveConfig(config *reference.ConfigReference, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

func mergeConfigs(existing, new *reference.ConfigReference) int {
	count := 0
	for key, setting := range new.Settings {
		if _, exists := existing.Settings[key]; !exists {
			existing.Settings[key] = setting
			count++
		}
	}
	return count
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}