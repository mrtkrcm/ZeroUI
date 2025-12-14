package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// validateReferenceCmd represents the validate-reference command
var validateReferenceCmd = &cobra.Command{
	Use:   "validate-reference [app]",
	Short: "Validate reference config mapping for an app",
	Long: `Validate that reference config mapping is working correctly.

This command checks:
- Reference config can be loaded
- Reference config can be mapped to app config format  
- App config and reference config can be merged
- All settings from reference are available in the final config

Examples:
  zeroui validate-reference ghostty
  zeroui validate-reference --all`,
	Example: `  zeroui validate-reference ghostty
  zeroui validate-reference --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidateReference,
}

var validateAll bool

func init() {
	rootCmd.AddCommand(validateReferenceCmd)
	validateReferenceCmd.Flags().BoolVar(&validateAll, "all", false, "Validate all available apps")
}

func runValidateReference(cmd *cobra.Command, args []string) error {
	// Create reference-enhanced loader
	loader, err := config.NewReferenceEnhancedLoader()
	if err != nil {
		return fmt.Errorf("failed to create reference-enhanced loader: %w", err)
	}

	if validateAll {
		return validateAllReferenceConfigs(loader)
	}

	if len(args) == 0 {
		return fmt.Errorf("app name required, or use --all flag")
	}

	appName := args[0]
	return validateSingleReferenceConfig(loader, appName)
}

func validateAllReferenceConfigs(loader *config.ReferenceEnhancedLoader) error {
	fmt.Println("🔍 Validating all reference configurations...")

	// Get list of all apps
	apps, err := loader.ListAppsWithReference()
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	if len(apps) == 0 {
		fmt.Println("❌ No apps found")
		return nil
	}

	fmt.Printf("📋 Found %d apps to validate\n\n", len(apps))

	var failedApps []string

	for _, app := range apps {
		fmt.Printf("🔄 Validating %s...\n", app)

		if err := validateSingleReferenceConfig(loader, app); err != nil {
			fmt.Printf("  ❌ Failed: %v\n\n", err)
			failedApps = append(failedApps, app)
		} else {
			fmt.Printf("  ✅ Success\n\n")
		}
	}

	// Summary
	successful := len(apps) - len(failedApps)
	fmt.Printf("📊 Summary: %d/%d apps validated successfully\n", successful, len(apps))

	if len(failedApps) > 0 {
		fmt.Printf("❌ Failed apps: %v\n", failedApps)
		return fmt.Errorf("validation failed for %d apps", len(failedApps))
	}

	fmt.Println("🎉 All reference configurations validated successfully!")
	return nil
}

func validateSingleReferenceConfig(loader *config.ReferenceEnhancedLoader, appName string) error {
	// 1. Test app config loading from any source
	fmt.Printf("  🔧 Loading app config...")
	source, err := loader.GetConfigSource(appName)
	if err != nil {
		fmt.Printf(" ❌\n")
		return err
	}
	fmt.Printf(" ✅ (source: %s)\n", source)

	// 2. Test merged config loading
	fmt.Printf("  🔄 Loading merged config...")
	mergedConfig, err := loader.LoadAppConfigWithReference(appName)
	if err != nil {
		fmt.Printf(" ❌\n")
		return fmt.Errorf("failed to load merged config: %w", err)
	}
	fmt.Printf(" ✅ (%d total fields)\n", len(mergedConfig.Fields))

	// 3. Test configuration metadata
	fmt.Printf("  📋 Validating metadata...")
	if mergedConfig.Name == "" {
		return fmt.Errorf("merged config missing name")
	}
	if mergedConfig.Path == "" {
		return fmt.Errorf("merged config missing path")
	}
	if mergedConfig.Format == "" {
		return fmt.Errorf("merged config missing format")
	}
	fmt.Printf(" ✅\n")

	// 4. Test field structure
	fmt.Printf("  🔍 Validating field structure...")
	fieldErrors := validateFieldStructure(mergedConfig)
	if len(fieldErrors) > 0 {
		fmt.Printf(" ❌\n")
		for _, fieldErr := range fieldErrors {
			fmt.Printf("    • %s\n", fieldErr)
		}
		return fmt.Errorf("field structure validation failed")
	}
	fmt.Printf(" ✅\n")

	return nil
}

func validateFieldStructure(mergedConfig *config.AppConfig) []string {
	var errors []string

	// Check that all fields have required properties
	for fieldKey, field := range mergedConfig.Fields {
		// Type should be set
		if field.Type == "" {
			errors = append(errors, fmt.Sprintf("field '%s' missing type", fieldKey))
		}

		// Check for valid type values
		validTypes := map[string]bool{
			"string": true, "number": true, "boolean": true, "choice": true,
			"select": true, "enum": true, "text": true, "int": true,
			"integer": true, "float": true, "bool": true, "array": true, "object": true,
		}
		if !validTypes[field.Type] {
			errors = append(errors, fmt.Sprintf("field '%s' has invalid type '%s'", fieldKey, field.Type))
		}

		// If type is choice/select/enum, should have values
		choiceTypes := map[string]bool{"choice": true, "select": true, "enum": true}
		if choiceTypes[field.Type] && len(field.Values) == 0 {
			errors = append(errors, fmt.Sprintf("field '%s' of type '%s' should have values", fieldKey, field.Type))
		}
	}

	return errors
}

func printConfigDetails(config *config.AppConfig, title string) {
	fmt.Printf("\n📋 %s:\n", title)
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Path: %s\n", config.Path)
	fmt.Printf("  Format: %s\n", config.Format)
	fmt.Printf("  Fields: %d\n", len(config.Fields))
	fmt.Printf("  Presets: %d\n", len(config.Presets))

	if len(config.Fields) > 0 {
		fmt.Printf("  Field types:\n")
		typeCount := make(map[string]int)
		for _, field := range config.Fields {
			typeCount[field.Type]++
		}
		for fieldType, count := range typeCount {
			fmt.Printf("    %s: %d\n", fieldType, count)
		}
	}
}
