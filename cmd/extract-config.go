package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
	"github.com/mrtkrcm/ZeroUI/pkg/reference"
)

// extractConfigCmd represents the extract-config command
var extractConfigCmd = &cobra.Command{
	Use:   "extract-config <app>",
	Short: "Automatically extract configuration from an application",
	Long: `Automatically extract configuration options from an application using multiple methods:
- CLI help output (if app has +show-config or similar)
- GitHub repository documentation
- Man pages
- Existing config files

Examples:
  zeroui extract-config ghostty
  zeroui extract-config zed --output configs/zed.yaml
  zeroui extract-config alacritty --method cli`,
	Args: cobra.ExactArgs(1),
	RunE: runExtractConfig,
}

var (
	extractOutput string
	extractMethod string
	extractUpdate bool
)

func init() {
	rootCmd.AddCommand(extractConfigCmd)
	extractConfigCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "Output file path (default: configs/<app>.yaml)")
	extractConfigCmd.Flags().StringVarP(&extractMethod, "method", "m", "auto", "Extraction method: auto, cli, github, man, file")
	extractConfigCmd.Flags().BoolVarP(&extractUpdate, "update", "u", false, "Update existing config (merge new settings)")
}

func runExtractConfig(cmd *cobra.Command, args []string) error {
	appName := args[0]
	
	// Determine output path
	if extractOutput == "" {
		extractOutput = filepath.Join("configs", appName+".yaml")
	}
	
	fmt.Printf("🔍 Extracting configuration for %s...\n", appName)
	
	// Create extractor registry
	registry := configextractor.NewExtractorRegistry()
	
	// Extract configuration
	config, err := registry.ExtractConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to extract config: %w", err)
	}
	
	fmt.Printf("✅ Extracted %d settings from %s\n", len(config.Settings), config.Source)
	fmt.Printf("   Confidence: %.0f%%\n", config.Confidence*100)
	
	// Convert to reference format
	ref := convertToReference(config)
	
	// Handle update mode
	if extractUpdate && fileExists(extractOutput) {
		existingRef, err := loadExistingReference(extractOutput)
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
		
		// Merge settings
		mergedCount := mergeSettings(existingRef, ref)
		fmt.Printf("📝 Merged %d new settings into existing config\n", mergedCount)
		ref = existingRef
	}
	
	// Save to file
	if err := saveReference(ref, extractOutput); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	fmt.Printf("💾 Saved configuration to %s\n", extractOutput)
	
	// Show sample settings
	fmt.Println("\n📋 Sample settings extracted:")
	count := 0
	for name, setting := range config.Settings {
		if count >= 5 {
			fmt.Printf("   ... and %d more settings\n", len(config.Settings)-5)
			break
		}
		fmt.Printf("   - %s (%s): %s\n", name, setting.Type, truncate(setting.Description, 50))
		count++
	}
	
	return nil
}

func convertToReference(config *configextractor.ExtractedConfig) *reference.ConfigReference {
	ref := &reference.ConfigReference{
		AppName:    config.AppName,
		ConfigPath: config.ConfigPath,
		ConfigType: config.ConfigType,
		Settings:   make(map[string]reference.ConfigSetting),
	}
	
	// Set default paths if not provided
	if ref.ConfigPath == "" {
		ref.ConfigPath = fmt.Sprintf("~/.config/%s/config", config.AppName)
	}
	if ref.ConfigType == "" {
		ref.ConfigType = "custom"
	}
	
	// Convert settings
	for name, setting := range config.Settings {
		ref.Settings[name] = reference.ConfigSetting{
			Name:         setting.Name,
			Type:         reference.SettingType(setting.Type),
			Description:  setting.Description,
			DefaultValue: setting.DefaultValue,
			ValidValues:  setting.ValidValues,
			Category:     setting.Category,
		}
	}
	
	return ref
}

func loadExistingReference(path string) (*reference.ConfigReference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var ref reference.ConfigReference
	if err := yaml.Unmarshal(data, &ref); err != nil {
		return nil, err
	}
	
	return &ref, nil
}

func mergeSettings(existing, new *reference.ConfigReference) int {
	count := 0
	for name, setting := range new.Settings {
		if _, exists := existing.Settings[name]; !exists {
			existing.Settings[name] = setting
			count++
		}
	}
	return count
}

func saveReference(ref *reference.ConfigReference, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(ref)
	if err != nil {
		return err
	}
	
	// Write file
	return os.WriteFile(path, data, 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}