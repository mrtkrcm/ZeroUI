//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoreIntegration focuses on essential ZeroUI functionality
func TestCoreIntegration(t *testing.T) {
	// Setup test environment
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Build ZeroUI binary for testing
	binaryPath := buildZeroUI(t, testDir)

	t.Run("CLI Core Functionality", func(t *testing.T) {
		testCLICommands(t, binaryPath, testDir)
	})

	t.Run("Configuration Engine Core", func(t *testing.T) {
		testConfigurationEngine(t, binaryPath, testDir)
	})

	t.Run("Plugin System Core", func(t *testing.T) {
		testPluginSystem(t, binaryPath, testDir)
	})

	t.Run("Error Handling Core", func(t *testing.T) {
		testErrorHandling(t, binaryPath, testDir)
	})
}

func testCLICommands(t *testing.T, binaryPath, testDir string) {
	// Test 1: List apps command
	t.Run("list apps shows available applications", func(t *testing.T) {
		output, err := runCommand(binaryPath, "list", "apps")
		require.NoError(t, err, "list apps command should succeed")

		assert.Contains(t, output, "ghostty", "Should list ghostty as available app")
		assert.Contains(t, output, "Available Applications", "Should show header")
	})

	// Test 2: Help command
	t.Run("help command displays usage", func(t *testing.T) {
		output, err := runCommand(binaryPath, "--help")
		require.NoError(t, err, "help command should succeed")

		assert.Contains(t, output, "ZeroUI", "Should contain app name")
		assert.Contains(t, output, "toggle", "Should show toggle command")
		assert.Contains(t, output, "ui", "Should show ui command")
	})

	// Test 3: Extract command (core functionality)
	t.Run("extract command processes configuration", func(t *testing.T) {
		output, err := runCommand(binaryPath, "extract", "ghostty", "--dry-run")
		require.NoError(t, err, "extract command should succeed")

		assert.Contains(t, output, "Extracting", "Should show extraction progress")
		assert.Contains(t, output, "settings", "Should show settings count")
	})
}

func testConfigurationEngine(t *testing.T, binaryPath, testDir string) {
	// Create test config file
	configDir := filepath.Join(testDir, ".config", "ghostty")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configFile := filepath.Join(configDir, "config")
	configContent := `# Test Ghostty config
cursor-style = beam
font-size = 12
theme = dark
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Create ghostty app registry entry
	appsDir := filepath.Join(testDir, ".config", "zeroui", "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	ghosttyAppConfig := `name: ghostty
path: ` + configFile + `
format: custom
description: Test Ghostty application

fields:
  cursor-style:
    type: choice
    values: ["beam", "block", "underline"]
    default: "beam"
    description: "Cursor style"

  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 12
    description: "Font size"

  theme:
    type: choice
    values: ["dark", "light", "auto"]
    default: "dark"
    description: "Theme"

presets:
  default:
    name: default
    description: Default settings
    values:
      cursor-style: beam
      font-size: 12
      theme: dark

hooks:
  post-toggle: "echo 'Ghostty config updated'"
`
	ghosttyAppPath := filepath.Join(appsDir, "ghostty.yaml")
	require.NoError(t, os.WriteFile(ghosttyAppPath, []byte(ghosttyAppConfig), 0644))

	// Test 1: Toggle operation with existing config
	t.Run("toggle operation modifies config correctly", func(t *testing.T) {
		// Set HOME and ZEROUI_CONFIG_DIR to test directory for config discovery
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "toggle command should succeed")

		assert.Contains(t, output, "cursor-style", "Should reference the field being toggled")
		assert.Contains(t, output, "Would set", "Should indicate dry-run mode")
	})

	// Test 2: Config file validation
	t.Run("handles invalid config gracefully", func(t *testing.T) {
		// Create invalid config
		invalidConfigFile := filepath.Join(configDir, "invalid_config")
		invalidContent := `invalid syntax here =
		malformed = = =
		`
		require.NoError(t, os.WriteFile(invalidConfigFile, []byte(invalidContent), 0644))

		// Should handle parsing errors gracefully
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, _ := runCommandWithEnv(cmd)
		// Command might succeed but should show available keys regardless
		assert.Contains(t, output, "Configurable Keys", "Should still show available keys")
	})
}

func testPluginSystem(t *testing.T, binaryPath, testDir string) {
	// Test 1: Plugin discovery
	t.Run("discovers available plugins", func(t *testing.T) {
		// Check if plugin registry initializes without error
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "list", "apps")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "should list apps via plugin system")

		// Should show apps detected by plugins
		assert.Contains(t, output, "ghostty", "Should detect ghostty via plugins")
	})

	// Test 2: Plugin communication resilience
	t.Run("handles plugin communication gracefully", func(t *testing.T) {
		// Test with normal operation - plugins should work or fail gracefully
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		// Either succeeds with keys or fails gracefully
		if err != nil {
			assert.Contains(t, output, "error", "Should provide error information")
		} else {
			assert.Contains(t, output, "Configurable Keys", "Should show available keys")
		}
	})
}

func testErrorHandling(t *testing.T, binaryPath, testDir string) {
	// Test 1: Invalid app name
	t.Run("handles invalid app name gracefully", func(t *testing.T) {
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "toggle", "nonexistent-app", "some-field", "value")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, _ := runCommandWithEnv(cmd)
		// Should provide helpful message even if it doesn't error
		assert.Contains(t, strings.ToLower(output), "not found", "Should indicate app not found")
	})

	// Test 2: Missing required arguments
	t.Run("validates required arguments", func(t *testing.T) {
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "toggle")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		assert.Error(t, err, "Should error for missing arguments")
		assert.Contains(t, output, "accepts 3 arg(s), received 0", "Should show argument validation error")
	})

	// Test 3: Invalid field values
	t.Run("validates field values", func(t *testing.T) {
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "invalid-field", "value", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		// Should either succeed in dry-run or provide validation error
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "field", "Should mention field validation")
		}
	})
}

// Helper functions

func setupTestEnvironment(t *testing.T) string {
	testDir, err := os.MkdirTemp("", "zeroui-integration-test-*")
	require.NoError(t, err, "Should create test directory")
	return testDir
}

func cleanupTestEnvironment(testDir string) {
	os.RemoveAll(testDir)
}

func buildZeroUI(t *testing.T, testDir string) string {
	// Build the binary in a temporary location
	binaryPath := filepath.Join(testDir, "zeroui")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to project root

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "Should build ZeroUI binary: %s", stderr.String())

	return binaryPath
}

func runCommand(binary string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return output.String(), err
}

func runCommandWithEnv(cmd *exec.Cmd) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set context for command cancellation
	cmdCtx := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	cmdCtx.Env = cmd.Env
	cmd = cmdCtx

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return output.String(), err
}

// TestTUIIntegration tests the TUI interface core functionality
func TestTUIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TUI integration tests in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	binaryPath := buildZeroUI(t, testDir)

	t.Run("TUI launches without error", func(t *testing.T) {
		// Test TUI launch with immediate exit
		cmd := exec.Command(binaryPath)
		cmd.Env = append(os.Environ(), "HOME="+testDir)

		// Start the command
		err := cmd.Start()
		require.NoError(t, err, "TUI should start")

		// Give it a moment to initialize
		time.Sleep(100 * time.Millisecond)

		// Terminate gracefully
		if cmd.Process != nil {
			cmd.Process.Signal(os.Interrupt)
			// Wait briefly for graceful shutdown
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case err := <-done:
				// Process terminated
				_ = err // TUI might exit with signal, which is expected
			case <-time.After(2 * time.Second):
				// Force kill if doesn't respond to interrupt
				cmd.Process.Kill()
				cmd.Wait()
			}
		}
	})
}

// TestConfigFileOperations tests file system interactions
func TestConfigFileOperations(t *testing.T) {
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	binaryPath := buildZeroUI(t, testDir)

	// Create test config structure
	configDir := filepath.Join(testDir, ".config", "ghostty")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configFile := filepath.Join(configDir, "config")
	originalContent := `# Test config
cursor-style = beam
font-size = 12
`
	require.NoError(t, os.WriteFile(configFile, []byte(originalContent), 0644))

	// Create ghostty app registry entry
	appsDir := filepath.Join(testDir, ".config", "zeroui", "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	ghosttyAppConfig := `name: ghostty
path: ` + configFile + `
format: custom
description: Test Ghostty application

fields:
  cursor-style:
    type: choice
    values: ["beam", "block", "underline"]
    default: "beam"
    description: "Cursor style"

  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 12
    description: "Font size"

presets:
  default:
    name: default
    description: Default settings
    values:
      cursor-style: beam
      font-size: 12

hooks:
  post-toggle: "echo 'Ghostty config updated'"
`
	ghosttyAppPath := filepath.Join(appsDir, "ghostty.yaml")
	require.NoError(t, os.WriteFile(ghosttyAppPath, []byte(ghosttyAppConfig), 0644))

	t.Run("preserves config file structure in dry-run", func(t *testing.T) {
		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "toggle", "ghostty", "cursor-style", "block", "--dry-run")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "dry-run should succeed")

		// Verify original file unchanged
		content, err := os.ReadFile(configFile)
		require.NoError(t, err, "Should read config file")
		assert.Equal(t, originalContent, string(content), "Config file should be unchanged in dry-run")

		assert.Contains(t, output, "Would set", "Should indicate dry-run mode")
	})

	t.Run("handles missing config directory gracefully", func(t *testing.T) {
		// Remove config directory
		require.NoError(t, os.RemoveAll(configDir))

		configDir := filepath.Join(testDir, ".config", "zeroui")
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(), "HOME="+testDir, "ZEROUI_CONFIG_DIR="+configDir)

		output, err := runCommandWithEnv(cmd)
		// Should either succeed with default keys or fail gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(output), "config", "Should mention config issue")
		} else {
			assert.Contains(t, output, "Configurable Keys", "Should show available keys")
		}
	})
}
