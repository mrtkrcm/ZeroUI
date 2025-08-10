package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/mrtkrcm/ZeroUI/pkg/reference"
)

var (
	improvedReferenceCmd = &cobra.Command{
		Use:   "ref",
		Short: "Configuration reference (improved)",
		Long: `Improved configuration reference system with clean, reliable data.
Uses curated static configuration files instead of fragile web scraping.`,
	}

	refListCmd = &cobra.Command{
		Use:   "list",
		Short: "List available applications",
		RunE:  runRefList,
	}

	refShowCmd = &cobra.Command{
		Use:   "show [app] [setting]",
		Short: "Show configuration details",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runRefShow,
	}

	refValidateCmd = &cobra.Command{
		Use:   "validate [app] [setting] [value]",
		Short: "Validate a configuration value",
		Args:  cobra.ExactArgs(3),
		RunE:  runRefValidate,
	}

	refSearchCmd = &cobra.Command{
		Use:   "search [app] [query]",
		Short: "Search configuration settings",
		Args:  cobra.ExactArgs(2),
		RunE:  runRefSearch,
	}

	// Styles for improved output
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Margin(0, 0, 1, 0)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	keyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)

	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))
)

func init() {
	rootCmd.AddCommand(improvedReferenceCmd)
	improvedReferenceCmd.AddCommand(refListCmd)
	improvedReferenceCmd.AddCommand(refShowCmd)
	improvedReferenceCmd.AddCommand(refValidateCmd)
	improvedReferenceCmd.AddCommand(refSearchCmd)
}

func setupImprovedManager() *reference.ReferenceManager {
	configDir := "configs" // Relative to project root
	loader := reference.NewStaticConfigLoader(configDir)
	return reference.NewReferenceManager(loader)
}

func runRefList(cmd *cobra.Command, args []string) error {
	manager := setupImprovedManager()
	
	apps, err := manager.ListApps()
	if err != nil {
		return fmt.Errorf("failed to list applications: %w", err)
	}

	if len(apps) == 0 {
		fmt.Println("No applications available.")
		return nil
	}

	fmt.Println(titleStyle.Render("📋 Available Applications"))
	
	for _, app := range apps {
		ref, err := manager.GetReference(app)
		if err != nil {
			continue
		}
		
		fmt.Printf("  %s %s\n", 
			keyStyle.Render("•"),
			formatAppInfo(ref))
	}

	return nil
}

func runRefShow(cmd *cobra.Command, args []string) error {
	appName := args[0]
	manager := setupImprovedManager()
	
	ref, err := manager.GetReference(appName)
	if err != nil {
		return fmt.Errorf("failed to get reference for %s: %w", appName, err)
	}

	if len(args) == 1 {
		// Show all settings for app
		return showAllSettings(ref)
	}

	// Show specific setting
	settingName := args[1]
	setting, exists := ref.Settings[settingName]
	if !exists {
		suggestions := findSimilarSettings(ref, settingName)
		fmt.Printf("%s Setting '%s' not found in %s\n", 
			errorStyle.Render("✗"), settingName, appName)
		if len(suggestions) > 0 {
			fmt.Printf("Did you mean: %s\n", strings.Join(suggestions, ", "))
		}
		return nil
	}

	return showSetting(appName, setting)
}

func runRefValidate(cmd *cobra.Command, args []string) error {
	appName := args[0]
	settingName := args[1]
	valueStr := args[2]

	manager := setupImprovedManager()
	
	// Parse value based on context
	value := parseValue(valueStr)
	
	result, err := manager.ValidateConfiguration(appName, settingName, value)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if result.Valid {
		fmt.Printf("%s Valid: %s.%s = %s\n", 
			successStyle.Render("✓"), appName, settingName, valueStr)
	} else {
		fmt.Printf("%s Invalid: %s.%s = %s\n", 
			errorStyle.Render("✗"), appName, settingName, valueStr)
		
		for _, error := range result.Errors {
			fmt.Printf("  Error: %s\n", error)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Printf("Suggestions: %s\n", strings.Join(result.Suggestions, ", "))
	}

	return nil
}

func runRefSearch(cmd *cobra.Command, args []string) error {
	appName := args[0]
	query := args[1]

	manager := setupImprovedManager()
	
	results, err := manager.SearchSettings(appName, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No settings found matching '%s' in %s\n", query, appName)
		return nil
	}

	fmt.Println(titleStyle.Render(fmt.Sprintf("🔍 Found %d settings matching '%s'", len(results), query)))
	
	for _, setting := range results {
		showSettingBrief(setting)
		fmt.Println()
	}

	return nil
}

func formatAppInfo(ref *reference.ConfigReference) string {
	return fmt.Sprintf("%s (%s, %d settings)", 
		keyStyle.Render(ref.AppName),
		dimStyle.Render(ref.ConfigType),
		len(ref.Settings))
}

func showAllSettings(ref *reference.ConfigReference) error {
	fmt.Println(titleStyle.Render(fmt.Sprintf("📖 %s Configuration", ref.AppName)))
	fmt.Printf("Config: %s (%s)\n", ref.ConfigPath, ref.ConfigType)
	fmt.Printf("Settings: %d\n\n", len(ref.Settings))

	// Group by category
	categories := make(map[string][]reference.ConfigSetting)
	for _, setting := range ref.Settings {
		category := setting.Category
		if category == "" {
			category = "Other"
		}
		categories[category] = append(categories[category], setting)
	}

	// Sort categories
	var categoryNames []string
	for name := range categories {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	for _, categoryName := range categoryNames {
		settings := categories[categoryName]
		
		fmt.Println(headerStyle.Render(fmt.Sprintf("📁 %s", categoryName)))
		
		// Sort settings by name
		sort.Slice(settings, func(i, j int) bool {
			return settings[i].Name < settings[j].Name
		})
		
		for _, setting := range settings {
			showSettingBrief(setting)
		}
		fmt.Println()
	}

	return nil
}

func showSetting(appName string, setting reference.ConfigSetting) error {
	fmt.Println(titleStyle.Render(fmt.Sprintf("⚙️  %s.%s", appName, setting.Name)))
	
	fmt.Printf("%s %s\n", keyStyle.Render("Type:"), valueStyle.Render(string(setting.Type)))
	
	if setting.Category != "" {
		fmt.Printf("%s %s\n", keyStyle.Render("Category:"), valueStyle.Render(setting.Category))
	}
	
	if setting.Description != "" {
		fmt.Printf("%s %s\n", keyStyle.Render("Description:"), valueStyle.Render(setting.Description))
	}
	
	if setting.DefaultValue != nil {
		fmt.Printf("%s %v\n", keyStyle.Render("Default:"), valueStyle.Render(fmt.Sprintf("%v", setting.DefaultValue)))
	}
	
	if setting.Example != nil {
		fmt.Printf("%s %v\n", keyStyle.Render("Example:"), valueStyle.Render(fmt.Sprintf("%v", setting.Example)))
	}
	
	if len(setting.ValidValues) > 0 {
		fmt.Printf("%s %s\n", keyStyle.Render("Valid values:"), valueStyle.Render(strings.Join(setting.ValidValues, ", ")))
	}

	if setting.Required {
		fmt.Printf("%s %s\n", keyStyle.Render("Required:"), successStyle.Render("Yes"))
	}

	return nil
}

func showSettingBrief(setting reference.ConfigSetting) {
	typeStr := dimStyle.Render(fmt.Sprintf("(%s)", setting.Type))
	
	if setting.DefaultValue != nil {
		typeStr += dimStyle.Render(fmt.Sprintf(" = %v", setting.DefaultValue))
	}
	
	fmt.Printf("  %s %s\n", keyStyle.Render(setting.Name), typeStr)
	
	if setting.Description != "" {
		desc := setting.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		fmt.Printf("    %s\n", dimStyle.Render(desc))
	}
}

// parseValue attempts to parse string value into appropriate type
func parseValue(valueStr string) interface{} {
	// Try boolean
	if valueStr == "true" {
		return true
	}
	if valueStr == "false" {
		return false
	}
	
	// Try integer
	if intVal, err := strconv.Atoi(valueStr); err == nil {
		return intVal
	}
	
	// Try float
	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal
	}
	
	// Default to string
	return valueStr
}

func findSimilarSettings(ref *reference.ConfigReference, target string) []string {
	var suggestions []string
	targetLower := strings.ToLower(target)
	
	for name := range ref.Settings {
		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, targetLower) || strings.Contains(targetLower, nameLower) {
			suggestions = append(suggestions, name)
			if len(suggestions) >= 3 {
				break
			}
		}
	}
	
	return suggestions
}