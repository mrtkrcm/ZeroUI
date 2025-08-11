package performance

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// setupLargeConfigTest creates test environment with large configurations
func setupLargeConfigTest(t testing.TB, numApps, fieldsPerApp int) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-perf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create apps directory
	appsDir := filepath.Join(tmpDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create target configs directory
	targetDir := filepath.Join(tmpDir, "targets")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create targets dir: %v", err)
	}

	// Generate test applications
	for appIdx := 0; appIdx < numApps; appIdx++ {
		appName := fmt.Sprintf("perf-app-%d", appIdx)
		targetPath := filepath.Join(targetDir, fmt.Sprintf("%s.json", appName))

		// Generate app config
		var appConfig strings.Builder
		appConfig.WriteString(fmt.Sprintf("name: %s\n", appName))
		appConfig.WriteString(fmt.Sprintf("path: %s\n", targetPath))
		appConfig.WriteString("format: json\n")
		appConfig.WriteString(fmt.Sprintf("description: Performance test app %d\n\n", appIdx))

		appConfig.WriteString("fields:\n")

		// Generate target config data
		targetConfig := make(map[string]interface{})

		// Add fields
		for fieldIdx := 0; fieldIdx < fieldsPerApp; fieldIdx++ {
			fieldName := fmt.Sprintf("field-%d", fieldIdx)

			// Create different field types for variety
			switch fieldIdx % 4 {
			case 0: // Choice field
				values := []string{"option1", "option2", "option3", "option4", "option5"}
				appConfig.WriteString(fmt.Sprintf("  %s:\n", fieldName))
				appConfig.WriteString("    type: choice\n")
				appConfig.WriteString("    values: [")
				for i, v := range values {
					if i > 0 {
						appConfig.WriteString(", ")
					}
					appConfig.WriteString(fmt.Sprintf("\"%s\"", v))
				}
				appConfig.WriteString("]\n")
				appConfig.WriteString("    default: \"option1\"\n")
				appConfig.WriteString(fmt.Sprintf("    description: \"Choice field %d\"\n", fieldIdx))
				targetConfig[fieldName] = "option1"

			case 1: // Number field
				values := []string{"10", "20", "30", "40", "50"}
				appConfig.WriteString(fmt.Sprintf("  %s:\n", fieldName))
				appConfig.WriteString("    type: number\n")
				appConfig.WriteString("    values: [")
				for i, v := range values {
					if i > 0 {
						appConfig.WriteString(", ")
					}
					appConfig.WriteString(fmt.Sprintf("\"%s\"", v))
				}
				appConfig.WriteString("]\n")
				appConfig.WriteString("    default: 10\n")
				appConfig.WriteString(fmt.Sprintf("    description: \"Number field %d\"\n", fieldIdx))
				targetConfig[fieldName] = 10

			case 2: // Boolean field
				appConfig.WriteString(fmt.Sprintf("  %s:\n", fieldName))
				appConfig.WriteString("    type: boolean\n")
				appConfig.WriteString("    default: false\n")
				appConfig.WriteString(fmt.Sprintf("    description: \"Boolean field %d\"\n", fieldIdx))
				targetConfig[fieldName] = false

			case 3: // String field
				appConfig.WriteString(fmt.Sprintf("  %s:\n", fieldName))
				appConfig.WriteString("    type: string\n")
				appConfig.WriteString("    default: \"default-value\"\n")
				appConfig.WriteString(fmt.Sprintf("    description: \"String field %d\"\n", fieldIdx))
				targetConfig[fieldName] = "default-value"
			}
		}

		// Add presets (every 5th app gets presets)
		if appIdx%5 == 0 {
			appConfig.WriteString("\npresets:\n")
			for presetIdx := 0; presetIdx < 3; presetIdx++ {
				presetName := fmt.Sprintf("preset-%d", presetIdx)
				appConfig.WriteString(fmt.Sprintf("  %s:\n", presetName))
				appConfig.WriteString(fmt.Sprintf("    name: %s\n", presetName))
				appConfig.WriteString(fmt.Sprintf("    description: \"Preset %d for %s\"\n", presetIdx, appName))
				appConfig.WriteString("    values:\n")

				// Set some field values in preset
				for fieldIdx := 0; fieldIdx < min(fieldsPerApp, 5); fieldIdx++ {
					fieldName := fmt.Sprintf("field-%d", fieldIdx)
					switch fieldIdx % 4 {
					case 0:
						appConfig.WriteString(fmt.Sprintf("      %s: option%d\n", fieldName, (presetIdx%3)+1))
					case 1:
						appConfig.WriteString(fmt.Sprintf("      %s: %d\n", fieldName, (presetIdx+1)*10))
					case 2:
						appConfig.WriteString(fmt.Sprintf("      %s: %t\n", fieldName, presetIdx%2 == 1))
					case 3:
						appConfig.WriteString(fmt.Sprintf("      %s: preset-%d-value\n", fieldName, presetIdx))
					}
				}
			}
		}

		// Write app config file
		appConfigPath := filepath.Join(appsDir, fmt.Sprintf("%s.yaml", appName))
		if err := ioutil.WriteFile(appConfigPath, []byte(appConfig.String()), 0644); err != nil {
			t.Fatalf("Failed to write app config: %v", err)
		}

		// Write target config file (JSON format)
		targetJSON := "{\n"
		first := true
		for key, value := range targetConfig {
			if !first {
				targetJSON += ",\n"
			}
			switch v := value.(type) {
			case string:
				targetJSON += fmt.Sprintf("  \"%s\": \"%s\"", key, v)
			case int:
				targetJSON += fmt.Sprintf("  \"%s\": %d", key, v)
			case bool:
				targetJSON += fmt.Sprintf("  \"%s\": %t", key, v)
			default:
				targetJSON += fmt.Sprintf("  \"%s\": \"%v\"", key, v)
			}
			first = false
		}
		targetJSON += "\n}"

		if err := ioutil.WriteFile(targetPath, []byte(targetJSON), 0644); err != nil {
			t.Fatalf("Failed to write target config: %v", err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// setupLargeCustomConfig creates test with large custom format configs
func setupLargeCustomConfig(t testing.TB, numLines int) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-custom-perf")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create large custom format file
	configPath := filepath.Join(tmpDir, "large.conf")
	var content strings.Builder
	content.WriteString("# Large custom configuration file for performance testing\n")
	content.WriteString("# Generated with many settings and comments\n\n")

	// Add many configuration lines
	for i := 0; i < numLines; i++ {
		if i%20 == 0 {
			content.WriteString(fmt.Sprintf("# Section %d\n", i/20))
		}

		switch i % 6 {
		case 0:
			content.WriteString(fmt.Sprintf("theme-%d = GruvboxDark\n", i))
		case 1:
			content.WriteString(fmt.Sprintf("font-family-%d = JetBrains Mono\n", i))
		case 2:
			content.WriteString(fmt.Sprintf("font-size-%d = %d\n", i, 12+(i%8)))
		case 3:
			content.WriteString(fmt.Sprintf("keybind-%d = cmd+%d=action-%d\n", i, i%10, i))
		case 4:
			content.WriteString(fmt.Sprintf("setting-%d = value-%d\n", i, i))
		case 5:
			content.WriteString(fmt.Sprintf("background-opacity-%d = 0.%d\n", i, (i%10)+1))
		}

		// Add some repeated keys to test array handling
		if i%100 == 0 {
			for j := 0; j < 5; j++ {
				content.WriteString(fmt.Sprintf("font-feature = feature-%d-%d\n", i, j))
			}
		}
	}

	if err := ioutil.WriteFile(configPath, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to write large config: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// BenchmarkEngine_LoadManyApps benchmarks loading many applications
func BenchmarkEngine_LoadManyApps(b *testing.B) {
	sizes := []struct {
		numApps      int
		fieldsPerApp int
	}{
		{10, 10},  // Small: 10 apps, 10 fields each
		{50, 20},  // Medium: 50 apps, 20 fields each
		{100, 30}, // Large: 100 apps, 30 fields each
		{200, 50}, // XLarge: 200 apps, 50 fields each
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Apps_%d_Fields_%d", size.numApps, size.fieldsPerApp), func(b *testing.B) {
			_, cleanup := setupLargeConfigTest(b, size.numApps, size.fieldsPerApp)
			defer cleanup()

			// Create engine with custom config dir
			engine, err := toggle.NewEngine()
			if err != nil {
				b.Fatalf("Failed to create engine: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				apps, err := engine.GetApps()
				if err != nil {
					b.Fatalf("Failed to get apps: %v", err)
				}
				if len(apps) != size.numApps {
					b.Fatalf("Expected %d apps, got %d", size.numApps, len(apps))
				}
			}
		})
	}
}

// BenchmarkEngine_ToggleOperations benchmarks toggle operations on large configs
func BenchmarkEngine_ToggleOperations(b *testing.B) {
	_, cleanup := setupLargeConfigTest(b, 10, 50) // 10 apps, 50 fields each
	defer cleanup()

	engine, err := toggle.NewEngine()
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}

	apps, err := engine.GetApps()
	if err != nil {
		b.Fatalf("Failed to get apps: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appName := apps[i%len(apps)]
		fieldName := fmt.Sprintf("field-%d", i%50) // Cycle through fields

		// Alternate between different values
		var value string
		switch (i % 50) % 4 {
		case 0:
			value = fmt.Sprintf("option%d", (i%3)+1)
		case 1:
			value = fmt.Sprintf("%d", 10+((i%5)*10))
		case 2:
			value = fmt.Sprintf("%t", i%2 == 0)
		case 3:
			value = fmt.Sprintf("toggle-value-%d", i)
		}

		err := engine.Toggle(appName, fieldName, value)
		if err != nil {
			b.Fatalf("Failed to toggle %s.%s: %v", appName, fieldName, err)
		}
	}
}

// BenchmarkEngine_PresetOperations benchmarks preset operations
func BenchmarkEngine_PresetOperations(b *testing.B) {
	_, cleanup := setupLargeConfigTest(b, 20, 30) // 20 apps, 30 fields each
	defer cleanup()

	engine, err := toggle.NewEngine()
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}

	// Get apps that have presets (every 5th app)
	appsWithPresets := []string{}
	for i := 0; i < 20; i += 5 {
		appsWithPresets = append(appsWithPresets, fmt.Sprintf("perf-app-%d", i))
	}

	presetNames := []string{"preset-0", "preset-1", "preset-2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appName := appsWithPresets[i%len(appsWithPresets)]
		presetName := presetNames[i%len(presetNames)]

		err := engine.ApplyPreset(appName, presetName)
		if err != nil {
			b.Fatalf("Failed to apply preset %s.%s: %v", appName, presetName, err)
		}
	}
}

// BenchmarkCustomParser_LargeFiles benchmarks custom format parsing with large files
func BenchmarkCustomParser_LargeFiles(b *testing.B) {
	sizes := []int{1000, 5000, 10000, 20000} // Number of lines

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Lines_%d", size), func(b *testing.B) {
			tmpDir, cleanup := setupLargeCustomConfig(b, size)
			defer cleanup()

			configPath := filepath.Join(tmpDir, "large.conf")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := config.ParseGhosttyConfig(configPath)
				if err != nil {
					b.Fatalf("Failed to parse large config: %v", err)
				}
			}
		})
	}
}

// BenchmarkCustomParser_WriteOperations benchmarks writing large custom configs
func BenchmarkCustomParser_WriteOperations(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-write-perf")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create large config data
	largeConfig := make(map[string]interface{})

	// Add many settings
	for i := 0; i < 1000; i++ {
		largeConfig[fmt.Sprintf("setting-%d", i)] = fmt.Sprintf("value-%d", i)
	}

	// Add array values
	fontFeatures := make([]string, 20)
	keybinds := make([]string, 50)

	for i := 0; i < 20; i++ {
		fontFeatures[i] = fmt.Sprintf("feature-%d", i)
	}
	for i := 0; i < 50; i++ {
		keybinds[i] = fmt.Sprintf("cmd+%d=action-%d", i, i)
	}

	largeConfig["font-feature"] = fontFeatures
	largeConfig["keybind"] = keybinds

	// Create koanf config
	k := koanf.New(".")
	for key, value := range largeConfig {
		k.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configPath := filepath.Join(tmpDir, fmt.Sprintf("bench-%d.conf", i))
		err := config.WriteGhosttyConfig(configPath, k, "/nonexistent")
		if err != nil {
			b.Fatalf("Failed to write large config: %v", err)
		}
	}
}

// TestConcurrentOperations tests concurrent access to configurations
func TestConcurrentOperations(t *testing.T) {
	_, cleanup := setupLargeConfigTest(t, 5, 10) // 5 apps, 10 fields each
	defer cleanup()

	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test concurrent toggles on different apps
	t.Run("Concurrent toggles different apps", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(appIdx int) {
				defer wg.Done()
				appName := fmt.Sprintf("perf-app-%d", appIdx)

				for j := 0; j < 10; j++ {
					fieldName := fmt.Sprintf("field-%d", j)
					value := fmt.Sprintf("option%d", (j%3)+1)

					if err := engine.Toggle(appName, fieldName, value); err != nil {
						errChan <- err
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("Concurrent operation error: %v", err)
		}
	})

	// Test concurrent operations on same app (potential race conditions)
	t.Run("Concurrent operations same app", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 20)
		appName := "perf-app-0"

		for i := 0; i < 10; i++ {
			wg.Add(2)

			// Toggle operation
			go func(fieldIdx int) {
				defer wg.Done()
				fieldName := fmt.Sprintf("field-%d", fieldIdx)
				value := fmt.Sprintf("option%d", (fieldIdx%3)+1)

				if err := engine.Toggle(appName, fieldName, value); err != nil {
					errChan <- err
				}
			}(i)

			// Cycle operation
			go func(fieldIdx int) {
				defer wg.Done()
				if fieldIdx < 5 { // Only cycle fields that have predefined values
					fieldName := fmt.Sprintf("field-%d", fieldIdx)
					if err := engine.Cycle(appName, fieldName); err != nil {
						errChan <- err
					}
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("Concurrent same-app operation error: %v", err)
		}
	})
}

// TestMemoryUsage tests memory usage with large configurations
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	_, cleanup := setupLargeConfigTest(t, 100, 100) // 100 apps, 100 fields each
	defer cleanup()

	engine, err := toggle.NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Load all apps to measure memory usage
	t.Run("Load all apps", func(t *testing.T) {
		apps, err := engine.GetApps()
		if err != nil {
			t.Fatalf("Failed to get apps: %v", err)
		}

		if len(apps) != 100 {
			t.Errorf("Expected 100 apps, got %d", len(apps))
		}

		// Load configuration for each app
		for _, appName := range apps {
			_, err := engine.GetAppConfig(appName)
			if err != nil {
				t.Errorf("Failed to load config for %s: %v", appName, err)
			}

			_, err = engine.GetCurrentValues(appName)
			if err != nil {
				t.Errorf("Failed to get current values for %s: %v", appName, err)
			}
		}
	})
}

// TestLargeConfigOperations tests operations on very large individual configs
func TestLargeConfigOperations(t *testing.T) {
	tmpDir, cleanup := setupLargeCustomConfig(t, 10000) // 10,000 lines
	defer cleanup()

	configPath := filepath.Join(tmpDir, "large.conf")

	t.Run("Parse large config", func(t *testing.T) {
		start := time.Now()
		config, err := config.ParseGhosttyConfig(configPath)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to parse large config: %v", err)
		}

		t.Logf("Parsed config with %d keys in %v", len(config), duration)

		if len(config) == 0 {
			t.Error("Expected non-empty config")
		}

		// Verify some expected keys exist
		expectedKeys := []string{"theme-0", "font-family-1", "font-size-2"}
		for _, key := range expectedKeys {
			if _, exists := config[key]; !exists {
				t.Errorf("Expected key '%s' not found", key)
			}
		}
	})

	t.Run("Write large config", func(t *testing.T) {
		// First parse the config
		config, err := config.ParseGhosttyConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		// Create koanf config
		k := koanf.New(".")
		for key, value := range config {
			k.Set(key, value)
		}

		// Add some modifications
		k.Set("new-setting-1", "new-value-1")
		k.Set("new-setting-2", "new-value-2")

		outputPath := filepath.Join(tmpDir, "large-output.conf")

		start := time.Now()
		err = config.WriteGhosttyConfig(outputPath, k, configPath)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to write large config: %v", err)
		}

		t.Logf("Wrote large config in %v", duration)

		// Verify output file exists and has content
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Expected output file to be created")
		}

		content, err := ioutil.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		if len(content) == 0 {
			t.Error("Expected non-empty output file")
		}

		// Verify new settings were added
		contentStr := string(content)
		if !strings.Contains(contentStr, "new-setting-1 = new-value-1") {
			t.Error("Expected new setting 1 in output")
		}
		if !strings.Contains(contentStr, "new-setting-2 = new-value-2") {
			t.Error("Expected new setting 2 in output")
		}
	})
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
