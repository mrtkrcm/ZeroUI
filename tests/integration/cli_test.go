//go:build integration
// +build integration

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
func runCmdCombinedOutputWithHome(ctx context.Context, home string, binary string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	configDir := filepath.Join(home, ".config", "zeroui")
	cmd.Env = append(os.Environ(), "HOME="+home, "ZEROUI_CONFIG_DIR="+configDir)
	return cmd.CombinedOutput()
}

// runCmdRunWithHome runs the given binary with args (using Run) and ensures the
// child process receives HOME and ZEROUI_CONFIG_DIR.
func runCmdRunWithHome(ctx context.Context, home string, binary string, args ...string) error {
	cmd := exec.CommandContext(ctx, binary, args...)
	configDir := filepath.Join(home, ".config", "zeroui")
	cmd.Env = append(os.Environ(), "HOME="+home, "ZEROUI_CONFIG_DIR="+configDir)
	return cmd.Run()
}

// buildBinary builds the zeroui binary in a temporary location and returns its path.
// It uses the repository root as the build context.
func buildBinary(t testing.TB) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "zeroui-test-binary")
	if err != nil {
		t.Fatalf("Failed to create temp dir for binary: %v", err)
	}

	binaryPath := filepath.Join(tmpDir, "zeroui")

	// Change to project root directory
	originalDir, _ := os.Getwd()
	projectRoot := filepath.Join(originalDir, "..", "..")
	_ = os.Chdir(projectRoot)
	defer os.Chdir(originalDir)

	// Build the binary
	{
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
		cmd.Env = os.Environ()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to build binary: %v\nOutput: %s", err, out)
		}
	}

	return binaryPath
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
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	// Create test app config
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
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create test target config file
	targetConfig := `{
   "theme": "dark",
   "font-size": 14
 }`

	targetPath := filepath.Join(tmpDir, "test-config.json")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0644); err != nil {
		t.Fatalf("Failed to write target config: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return configDir, cleanup
}

// TestCLI_BasicCommands tests basic CLI functionality
func TestCLI_BasicCommands(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	// child processes expect HOME to be the parent of .config
	home := filepath.Dir(filepath.Dir(configDir))
	os.Setenv("HOME", home)

	// Test list apps
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "list", "apps")
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "list", "keys", "test-app")
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "list", "presets", "test-app")
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "toggle", "test-app", "theme", "light", "--dry-run")
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
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	home := filepath.Dir(filepath.Dir(configDir))
	os.Setenv("HOME", home)

	// Test non-existent app
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "toggle", "nonexistent", "theme", "dark")
		if err == nil {
			t.Error("Expected error for non-existent app")
		}
		if !strings.Contains(string(output), "not found") {
			t.Errorf("Expected 'not found' in error output, got: %s", output)
		}
		if !strings.Contains(string(output), "Suggestions:") {
			t.Errorf("Expected suggestions in error output, got: %s", output)
		}
	}

	// Test non-existent field
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "toggle", "test-app", "nonexistent", "value")
		if err == nil {
			t.Error("Expected error for non-existent field")
		}
		if !strings.Contains(string(output), "field") && !strings.Contains(string(output), "not found") {
			t.Errorf("Expected field error in output, got: %s", output)
		}
	}

	// Test invalid value
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "toggle", "test-app", "theme", "invalid")
		if err == nil {
			t.Error("Expected error for invalid value")
		}
		if !strings.Contains(string(output), "invalid value") {
			t.Errorf("Expected 'invalid value' in error output, got: %s", output)
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "backup", "create", "test-app")
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "backup", "list")
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
		output, err := runCmdCombinedOutputWithHome(ctx, home, binaryPath, "backup", "cleanup", "test-app", "--keep", "3")
		if err != nil {
			t.Fatalf("Failed to cleanup backups: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Cleaned up") {
			t.Errorf("Expected 'Cleaned up' in output, got: %s", output)
		}
	}
}
