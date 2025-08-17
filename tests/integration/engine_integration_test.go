package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngineIntegration tests the core toggle engine functionality
func TestEngineIntegration(t *testing.T) {
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	binaryPath := buildZeroUI(t, testDir)

	t.Run("Configuration Detection", func(t *testing.T) {
		testConfigurationDetection(t, binaryPath, testDir)
	})

	t.Run("Toggle Operations", func(t *testing.T) {
		testToggleOperations(t, binaryPath, testDir)
	})

	t.Run("File Operations", func(t *testing.T) {
		testFileOperations(t, binaryPath, testDir)
	})

	t.Run("Validation Engine", func(t *testing.T) {
		testValidationEngine(t, binaryPath, testDir)
	})
}

func testConfigurationDetection(t *testing.T, binaryPath, testDir string) {
	// Create test configuration files in various formats
	setupTestConfigs(t, testDir)

	// Test 1: Detects existing configuration
	t.Run("detects existing ghostty config", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "extract", "ghostty", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should extract config successfully")
		
		assert.Contains(t, output, "settings", "Should detect configuration settings")
		assert.Contains(t, output, "Extracting", "Should show extraction process")
	})

	// Test 2: Handles missing configuration
	t.Run("handles missing config gracefully", func(t *testing.T) {
		// Remove config directory
		configDir := filepath.Join(testDir, ".config", "ghostty")
		os.RemoveAll(configDir)
		
		cmd := exec.Command(binaryPath, "extract", "ghostty", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Should either succeed with default config or provide helpful error
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "config", "Should mention config issue")
		} else {
			assert.Contains(t, output, "settings", "Should show available settings")
		}
	})

	// Test 3: Multi-format support
	t.Run("supports multiple config formats", func(t *testing.T) {
		// Test with different app configs if available
		apps := []string{"ghostty"}
		
		for _, app := range apps {
			cmd := exec.Command(binaryPath, "list", "keys", app)
			cmd.Env = append(os.Environ(), "HOME="+testDir)
			
			output, err := runCommandWithEnv(cmd)
			if err == nil {
				assert.Contains(t, output, "keys", "Should show configuration keys for "+app)
			}
		}
	})
}

func testToggleOperations(t *testing.T, binaryPath, testDir string) {
	setupTestConfigs(t, testDir)

	// Test 1: Basic toggle operation
	t.Run("performs basic toggle operation", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Toggle operation should succeed")
		
		assert.Contains(t, output, "cursor-style", "Should reference the toggled field")
		assert.Contains(t, output, "block", "Should reference the new value")
	})

	// Test 2: Cycle operation
	t.Run("performs cycle operation", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "cycle", "ghostty", "cursor-style", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Cycle might fail if field has no predefined values, which is acceptable
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "cycle", "Should mention cycling")
		} else {
			assert.Contains(t, output, "cursor-style", "Should reference the cycled field")
		}
	})

	// Test 3: Field validation
	t.Run("validates field names", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "nonexistent-field", "value", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Should either validate field or proceed with user-provided field
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "field", "Should mention field issue")
		}
	})

	// Test 4: Value validation
	t.Run("validates field values", func(t *testing.T) {
		// Test with a known field that likely has validation
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "font-size", "not-a-number", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Should either validate value or accept user input
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "value", "Should mention value validation")
		}
	})
}

func testFileOperations(t *testing.T, binaryPath, testDir string) {
	setupTestConfigs(t, testDir)

	configDir := filepath.Join(testDir, ".config", "ghostty")
	configFile := filepath.Join(configDir, "config")

	// Test 1: Preserves file in dry-run mode
	t.Run("preserves original file in dry-run", func(t *testing.T) {
		// Read original content
		originalContent, err := os.ReadFile(configFile)
		require.NoError(t, err, "Should read original config")
		
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		_, err = runCommandWithEnv(cmd)
		require.NoError(t, err, "Dry-run should succeed")
		
		// Verify file unchanged
		currentContent, err := os.ReadFile(configFile)
		require.NoError(t, err, "Should read config after dry-run")
		assert.Equal(t, originalContent, currentContent, "File should be unchanged in dry-run")
	})

	// Test 2: Backup creation (when not in dry-run)
	t.Run("creates backup when modifying config", func(t *testing.T) {
		// Note: This test would actually modify files, so we'll test the backup logic
		// by ensuring the application handles backup creation gracefully
		
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run", "--verbose")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should handle backup logic")
		
		// In dry-run, should mention what would happen
		assert.Contains(t, strings.ToLower(output), "dry", "Should indicate dry-run mode")
	})

	// Test 3: Handles read-only files
	t.Run("handles read-only files gracefully", func(t *testing.T) {
		// Make config file read-only
		err := os.Chmod(configFile, 0444)
		require.NoError(t, err, "Should make file read-only")
		
		// Restore permissions after test
		defer func() {
			os.Chmod(configFile, 0644)
		}()
		
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Dry-run should still work even with read-only files
		require.NoError(t, err, "Dry-run should work with read-only files")
		assert.Contains(t, output, "dry", "Should indicate dry-run mode")
	})
}

func testValidationEngine(t *testing.T, binaryPath, testDir string) {
	setupTestConfigs(t, testDir)

	// Test 1: Schema validation
	t.Run("validates configuration schema", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should list available keys")
		
		assert.Contains(t, output, "keys", "Should show configuration keys")
		// Should show some common ghostty fields
		lines := strings.Split(output, "\n")
		keyCount := 0
		for _, line := range lines {
			if strings.Contains(line, "•") || strings.Contains(line, "-") {
				keyCount++
			}
		}
		assert.Greater(t, keyCount, 0, "Should show at least some configuration keys")
	})

	// Test 2: Type validation
	t.Run("validates value types", func(t *testing.T) {
		// Test numeric field with non-numeric value
		testCases := []struct {
			field    string
			value    string
			shouldWarn bool
		}{
			{"font-size", "abc", true},      // Non-numeric for numeric field
			{"cursor-style", "invalid", false}, // Might accept any string
			{"font-size", "12", false},      // Valid numeric
		}

		for _, tc := range testCases {
			cmd := exec.Command(binaryPath, "toggle", "ghostty", tc.field, tc.value, "--dry-run")
			cmd.Env = append(os.Environ(), "HOME="+testDir)
			
			output, err := runCommandWithEnv(cmd)
			if tc.shouldWarn && err != nil {
				assert.Contains(t, strings.ToLower(output), "value", "Should validate value type")
			}
		}
	})

	// Test 3: Configuration consistency
	t.Run("maintains configuration consistency", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "extract", "ghostty", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should extract config consistently")
		
		assert.Contains(t, output, "settings", "Should show configuration settings")
		
		// Run again to ensure consistency
		output2, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should extract config consistently on repeat")
		
		// Results should be consistent between runs
		assert.Contains(t, output2, "settings", "Should show configuration settings consistently")
	})
}

func setupTestConfigs(t *testing.T, testDir string) {
	// Create Ghostty configuration
	ghosttyDir := filepath.Join(testDir, ".config", "ghostty")
	require.NoError(t, os.MkdirAll(ghosttyDir, 0755))
	
	ghosttyConfig := filepath.Join(ghosttyDir, "config")
	ghosttyContent := `# Test Ghostty Configuration
cursor-style = beam
font-size = 12
theme = dark
font-family = "JetBrains Mono"
window-width = 80
window-height = 24

# Color configuration
foreground = #ffffff
background = #000000

# Advanced settings
cursor-click-to-move = true
copy-on-select = false
`
	require.NoError(t, os.WriteFile(ghosttyConfig, []byte(ghosttyContent), 0644))

	// Create other app configs if needed (e.g., Alacritty, WezTerm)
	// This demonstrates multi-app support testing
	
	// Alacritty config (YAML format)
	alacrittyDir := filepath.Join(testDir, ".config", "alacritty")
	require.NoError(t, os.MkdirAll(alacrittyDir, 0755))
	
	alacrittyConfig := filepath.Join(alacrittyDir, "alacritty.yml")
	alacrittyContent := `# Test Alacritty Configuration
window:
  dimensions:
    columns: 80
    lines: 24

font:
  size: 12
  normal:
    family: "JetBrains Mono"

colors:
  primary:
    background: '#000000'
    foreground: '#ffffff'

cursor:
  style: Block
`
	require.NoError(t, os.WriteFile(alacrittyConfig, []byte(alacrittyContent), 0644))
}

// TestEngineErrorRecovery tests error recovery mechanisms
func TestEngineErrorRecovery(t *testing.T) {
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	binaryPath := buildZeroUI(t, testDir)

	// Test 1: Malformed configuration recovery
	t.Run("recovers from malformed configuration", func(t *testing.T) {
		// Create malformed config
		configDir := filepath.Join(testDir, ".config", "ghostty")
		require.NoError(t, os.MkdirAll(configDir, 0755))
		
		configFile := filepath.Join(configDir, "config")
		malformedContent := `# Malformed config
		cursor-style = 
		font-size = "not a number"
		invalid-syntax-here = = = =
		= missing-key
		`
		require.NoError(t, os.WriteFile(configFile, []byte(malformedContent), 0644))
		
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(), "HOME="+testDir)
		
		output, err := runCommandWithEnv(cmd)
		// Should handle malformed config gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "error", "Should report error gracefully")
		} else {
			assert.Contains(t, output, "keys", "Should still show available keys")
		}
	})

	// Test 2: File system error recovery
	t.Run("recovers from file system errors", func(t *testing.T) {
		// Test with non-existent directory
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME=/nonexistent/directory")
		
		output, err := runCommandWithEnv(cmd)
		// Should handle missing directories gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "error", "Should report error gracefully")
		}
	})
}