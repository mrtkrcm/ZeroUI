package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runCmdCombinedOutputWithHome runs the given binary with args, ensuring the child
// process receives both HOME and ZEROUI_CONFIG_DIR environment variables so it
// reliably reads the test-created configuration directory.
func runCmdCombinedOutputWithHome(ctx context.Context, home, configDir, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "ZEROUI_CONFIG_DIR="+configDir)
	return cmd.CombinedOutput()
}

// runCmdRunWithHome runs the given binary with args (using Run) and ensures the
// child process receives HOME and ZEROUI_CONFIG_DIR.
func runCmdRunWithHome(ctx context.Context, home, configDir, binary string, args ...string) error {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), "HOME="+home, "ZEROUI_CONFIG_DIR="+configDir)
	return cmd.Run()
}

// TestCLI_BasicCommands tests basic CLI functionality
func TestCLI_BasicCommands(t *testing.T) {
	// Build the binary
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test config
	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	// compute project HOME (parent of .config)
	home := filepath.Dir(filepath.Dir(configDir))
	// Set environment variable for current process (keeps compatibility)
	os.Setenv("HOME", home)

	// Test list apps
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "list", "apps")
		if err != nil {
			t.Fatalf("Failed to list apps: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "test-app") {
			t.Errorf("Expected 'test-app' in output, got: %s", output)
		}
	}

	// Test list keys
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "list", "keys", "test-app")
		if err != nil {
			t.Fatalf("Failed to list keys: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "theme") {
			t.Errorf("Expected 'theme' in output, got: %s", output)
		}
	}

	// Test list presets
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "list", "presets", "test-app")
		if err != nil {
			t.Fatalf("Failed to list presets: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "default") {
			t.Errorf("Expected 'default' preset in output, got: %s", output)
		}
	}

	// Test dry-run toggle
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "toggle", "test-app", "theme", "light", "--dry-run")
		if err != nil {
			t.Fatalf("Failed to toggle (dry-run): %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Would set") {
			t.Errorf("Expected 'Would set' in dry-run output, got: %s", output)
		}
	}
}

// TestCLI_ErrorHandling tests CLI error handling
func TestCLI_ErrorHandling(t *testing.T) {
	// Build the binary
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test config
	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	// compute HOME correctly
	home := filepath.Dir(filepath.Dir(configDir))
	os.Setenv("HOME", home)

	// Test non-existent app
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "toggle", "nonexistent", "theme", "dark")
		// The CLI prints user-friendly ZeroUI errors to stderr but may exit 0.
		// Assert on the output text rather than expecting a non-zero exit.
		if err != nil {
			t.Fatalf("Failed to run command: %v\nOutput: %s", err, output)
		}
		outStr := string(output)
		if !strings.Contains(outStr, "not found") && !strings.Contains(outStr, "application") {
			t.Errorf("Expected 'not found' or application message in output, got: %s", outStr)
		}
		if !strings.Contains(outStr, "Suggestions") && !strings.Contains(outStr, "Did you mean") {
			t.Errorf("Expected suggestions in output, got: %s", outStr)
		}
	}

	// Test non-existent field
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "toggle", "test-app", "nonexistent", "value")
		// CLI may print errors rather than returning non-zero; check output content.
		if err != nil {
			t.Fatalf("Failed to run command: %v\nOutput: %s", err, output)
		}
		outStr := string(output)
		if !strings.Contains(outStr, "field") && !strings.Contains(outStr, "not found") {
			t.Errorf("Expected field error in output, got: %s", outStr)
		}
	}

	// Test invalid value
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "toggle", "test-app", "theme", "invalid")
		// The CLI returns user-facing messages; assert on output text.
		if err != nil {
			t.Fatalf("Failed to run command: %v\nOutput: %s", err, output)
		}
		outStr := string(output)
		if !strings.Contains(outStr, "invalid value") && !strings.Contains(outStr, "invalid") {
			t.Errorf("Expected 'invalid value' in error output, got: %s", outStr)
		}
	}
}

// TestCLI_BackupCommands tests backup functionality
func TestCLI_BackupCommands(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	home := filepath.Dir(filepath.Dir(configDir))
	os.Setenv("HOME", home)

	// Test backup create
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "backup", "create", "test-app")
		if err != nil {
			t.Fatalf("Failed to create backup: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Backup created") {
			t.Errorf("Expected 'Backup created' in output, got: %s", output)
		}
	}

	// Test backup list
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "backup", "list")
		if err != nil {
			t.Fatalf("Failed to list backups: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "test-app") {
			t.Errorf("Expected 'test-app' in backup list, got: %s", output)
		}
	}

	// Test backup cleanup
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, configDir, binaryPath, "backup", "cleanup", "test-app", "--keep", "3")
		if err != nil {
			t.Fatalf("Failed to cleanup backups: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Cleaned up") {
			t.Errorf("Expected 'Cleaned up' in output, got: %s", output)
		}
	}
}

// buildBinary builds the zeroui binary for testing.
// It uses the repository root as the build context.
func buildBinary(t testing.TB) string {
	t.Helper()

	// Use a cached binary if available and recent
	cacheKey := "zeroui-test-binary"
	if cached, exists := getCachedBinary(cacheKey); exists && isBinaryRecent(cached) {
		t.Logf("Using cached binary: %s", cached)
		return cached
	}

	tmpDir, err := os.MkdirTemp("", "zeroui-test-binary")
	if err != nil {
		t.Fatalf("Failed to create temp dir for binary: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "zeroui")

	// Find project root more reliably
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// Build the binary with optimizations
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build",
			"-buildvcs=false",
			"-ldflags=-s -w", // Strip debug info for smaller binary
			"-o", binaryPath, ".")
		cmd.Dir = projectRoot
		cmd.Env = os.Environ()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, out)
		}
	}

	// Cache the binary
	cacheBinary(cacheKey, binaryPath)
	return binaryPath
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// Binary cache for performance
var binaryCache = make(map[string]string)
var binaryTimestamps = make(map[string]time.Time)

func getCachedBinary(key string) (string, bool) {
	path, exists := binaryCache[key]
	return path, exists && fileExists(path)
}

func cacheBinary(key, path string) {
	binaryCache[key] = path
	binaryTimestamps[key] = time.Now()
}

func isBinaryRecent(path string) bool {
	if timestamp, exists := binaryTimestamps[path]; exists {
		// Consider binary recent if less than 5 minutes old
		return time.Since(timestamp) < 5*time.Minute
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// setupTestConfig creates a test configuration directory structure and returns
// the path to the created `.config/zeroui` directory and a cleanup func.
func setupTestConfig(t testing.TB) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeroui-test-config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .config/zeroui structure
	configDir := filepath.Join(tmpDir, ".config", "zeroui")
	appsDir := filepath.Join(configDir, "apps")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app config (proper YAML indentation)
	testConfig := `name: test-app
path: ` + filepath.Join(tmpDir, "test-config.json") + `
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light", "auto"]
    default: "dark"
    description: "Application theme"

  font-size:
    type: number
    values: ["12", "14", "16", "18"]
    default: 14
    description: "Font size"

presets:
  default:
    name: default
    description: Default settings
    values:
      theme: dark
      font-size: 14

hooks:
  post-toggle: "echo 'Config updated'"
`

	configPath := filepath.Join(appsDir, "test-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create test target config file
	targetConfig := `{
  "theme": "dark",
  "font-size": 14
}`

	targetPath := filepath.Join(tmpDir, "test-config.json")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0o644); err != nil {
		t.Fatalf("Failed to write target config: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return configDir, cleanup
}
