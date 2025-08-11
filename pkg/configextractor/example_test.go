package configextractor_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
)

func ExampleExtractor_Extract() {
	// Create a new extractor with default settings
	extractor := configextractor.New()

	ctx := context.Background()
	config, err := extractor.Extract(ctx, "ghostty")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App: %s\n", config.App)
	fmt.Printf("Format: %s\n", config.Format)
	fmt.Printf("Number of settings: %d\n", len(config.Settings))
	
	// Print first few settings
	count := 0
	for name, setting := range config.Settings {
		if count >= 3 {
			break
		}
		fmt.Printf("Setting: %s (type: %s)\n", name, setting.Type)
		count++
	}
}

func ExampleExtractor_ExtractBatch() {
	// Create extractor with custom configuration
	extractor := configextractor.New(
		configextractor.WithTimeout(15*time.Second),
		configextractor.WithConcurrency(4),
	)

	apps := []string{"ghostty", "zed", "tmux", "git"}
	ctx := context.Background()

	results, err := extractor.ExtractBatch(ctx, apps)
	if err != nil {
		log.Printf("Some extractions failed: %v", err)
	}

	fmt.Printf("Successfully extracted configs for %d apps:\n", len(results))
	for app := range results {
		fmt.Printf("- %s\n", app)
	}
}

func ExampleValidator() {
	// Create a validator with custom rules
	validator := configextractor.NewValidator()

	// Add validation rules
	validator.AddRule("font-size", configextractor.NumberRule(
		false, // not required
		configextractor.Min(8),  // minimum size
		configextractor.Max(72), // maximum size
	))

	validator.AddRule("theme", configextractor.ChoiceRule(
		false, // not required
		"dark", "light", "auto",
	))

	// Validate individual settings
	result := validator.Validate("font-size", 12)
	if result.Valid {
		fmt.Println("font-size: valid")
	}

	result = validator.Validate("font-size", 200) // Too large
	if !result.Valid {
		fmt.Printf("font-size validation failed: %v\n", result.Errors)
	}

	result = validator.Validate("theme", "dark")
	if result.Valid {
		fmt.Println("theme: valid")
	}

	result = validator.Validate("theme", "rainbow") // Invalid choice
	if !result.Valid {
		fmt.Printf("theme validation failed: %v\n", result.Errors)
	}
}

func ExampleConfig() {
	// Example of working with extracted configuration
	extractor := configextractor.New()
	ctx := context.Background()

	config, err := extractor.Extract(ctx, "tmux")
	if err != nil {
		log.Fatal(err)
	}

	// Access configuration settings
	fmt.Printf("Configuration for %s:\n", config.App)

	// Group settings by category
	categories := make(map[string][]configextractor.Setting)
	for _, setting := range config.Settings {
		cat := setting.Cat
		if cat == "" {
			cat = "general"
		}
		categories[cat] = append(categories[cat], setting)
	}

	for category, settings := range categories {
		fmt.Printf("\n%s settings:\n", category)
		for _, setting := range settings {
			fmt.Printf("  %s: %v", setting.Name, setting.Default)
			if setting.Desc != "" {
				fmt.Printf(" (%s)", setting.Desc)
			}
			fmt.Println()
		}
	}

	// Show extraction metadata
	fmt.Printf("\nExtracted from: %s\n", config.Source.Method)
	fmt.Printf("Confidence: %.2f\n", config.Source.Confidence)
	fmt.Printf("Timestamp: %s\n", config.Timestamp.Format(time.RFC3339))
}

// CustomExtractorStrategy demonstrates how to implement and use a custom extraction strategy
type CustomExtractorStrategy struct{}

func (c *CustomExtractorStrategy) Name() string {
	return "custom"
}

func (c *CustomExtractorStrategy) Priority() int {
	return 80 // High priority
}

func (c *CustomExtractorStrategy) CanExtract(app string) bool {
	return app == "my-custom-app"
}

func (c *CustomExtractorStrategy) Extract(ctx context.Context, app string) (*configextractor.Config, error) {
	return &configextractor.Config{
		App:    app,
		Format: "custom",
		Settings: map[string]configextractor.Setting{
			"custom-setting": {
				Name:    "custom-setting",
				Type:    configextractor.TypeString,
				Default: "custom-value",
				Desc:    "A custom setting from custom strategy",
			},
		},
		Source: configextractor.ExtractionSource{
			Method:     "custom",
			Location:   "custom-strategy",
			Confidence: 1.0,
		},
		Timestamp: time.Now(),
	}, nil
}

func Example_customStrategy() {
	// Example of adding a custom extraction strategy

	// Create extractor with custom strategy
	extractor := configextractor.New(
		configextractor.WithStrategy(&CustomExtractorStrategy{}),
	)

	ctx := context.Background()
	config, err := extractor.Extract(ctx, "my-custom-app")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Extracted %s config using %s strategy\n", 
		config.App, config.Source.Method)
}

func Example_minimalUsage() {
	// Simplest possible usage
	extractor := configextractor.New()
	
	config, err := extractor.Extract(context.Background(), "git")
	if err != nil {
		log.Printf("Failed to extract git config: %v", err)
		return
	}

	// Check if a specific setting exists
	if userEmail, exists := config.Settings["user.email"]; exists {
		fmt.Printf("Git user email setting: %s\n", userEmail.Name)
		if userEmail.Default != nil {
			fmt.Printf("Default value: %v\n", userEmail.Default)
		}
	}

	// List all setting names
	fmt.Println("Available git settings:")
	for name := range config.Settings {
		fmt.Printf("- %s\n", name)
	}
}

func Example_performanceOptimized() {
	// Example showing performance optimizations
	extractor := configextractor.New(
		configextractor.WithTimeout(5*time.Second),   // Shorter timeout
		configextractor.WithConcurrency(16),          // Higher concurrency
		// Could add custom cache here with WithCache()
	)

	// Extract multiple apps concurrently
	apps := []string{"ghostty", "zed", "alacritty", "wezterm", "tmux", "git"}
	
	start := time.Now()
	results, _ := extractor.ExtractBatch(context.Background(), apps)
	elapsed := time.Since(start)

	fmt.Printf("Extracted %d configs in %v\n", len(results), elapsed)
	fmt.Printf("Average time per app: %v\n", elapsed/time.Duration(len(results)))

	// Second extraction should be much faster due to caching
	start = time.Now()
	cachedResults, _ := extractor.ExtractBatch(context.Background(), apps)
	cachedElapsed := time.Since(start)

	fmt.Printf("Second extraction (cached): %d configs in %v\n", 
		len(cachedResults), cachedElapsed)
	fmt.Printf("Speedup: %.2fx\n", float64(elapsed)/float64(cachedElapsed))
}