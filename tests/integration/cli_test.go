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

// TestCLI_BasicCommands tests basic CLI functionality
func TestCLI_BasicCommands(t *testing.T) {
	// Build the binary
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test config
	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	// Set environment variable to use test config
	os.Setenv("HOME", filepath.Dir(configDir))

	// Test list apps
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "list", "apps")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "list", "keys", "test-app")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "list", "presets", "test-app")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "test-app", "theme", "light", "--dry-run")
		output, err := cmd.CombinedOutput()
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

	// Set environment variable to use test config
	os.Setenv("HOME", filepath.Dir(configDir))

	// Test non-existent app
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "nonexistent", "theme", "dark")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "test-app", "nonexistent", "value")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "test-app", "theme", "invalid")
		output, err := cmd.CombinedOutput()
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
	// Build the binary
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Setup test config
	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	// Set environment variable to use test config
	os.Setenv("HOME", filepath.Dir(configDir))

	// Test backup create
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "backup", "create", "test-app")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "backup", "list")
		output, err := cmd.CombinedOutput()
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
		cmd := exec.CommandContext(ctx, binaryPath, "backup", "cleanup", "test-app", "--keep", "3")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to cleanup backups: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Cleaned up") {
			t.Errorf("Expected 'Cleaned up' in output, got: %s", output)
		}
	}
}

// buildBinary builds the zeroui binary for testing
func buildBinary(t testing.TB) string {
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
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build binary: %v", err)
		}
	}

	return binaryPath
}

// setupTestConfig creates a test configuration
func setupTestConfig(t testing.TB) (string, func()) {
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

// TestCLI_CycleCommands tests cycle functionality
func TestCLI_CycleCommands(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Test cycle theme
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "cycle", "test-app", "theme")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to cycle theme: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "theme") {
			t.Errorf("Expected 'theme' in cycle output, got: %s", output)
		}
	}

	// Test dry-run cycle
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "cycle", "test-app", "font-size", "--dry-run")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to cycle (dry-run): %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Would cycle") {
			t.Errorf("Expected 'Would cycle' in dry-run output, got: %s", output)
		}
	}
}

// TestCLI_PresetCommands tests preset functionality
func TestCLI_PresetCommands(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Test apply preset
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "preset", "test-app", "default")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to apply preset: %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Applied preset") {
			t.Errorf("Expected 'Applied preset' in output, got: %s", output)
		}
	}

	// Test dry-run preset
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "preset", "test-app", "default", "--dry-run")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to apply preset (dry-run): %v\nOutput: %s", err, output)
		}
		if !strings.Contains(string(output), "Would apply") {
			t.Errorf("Expected 'Would apply' in dry-run output, got: %s", output)
		}
	}
}

// TestCLI_EndToEndWorkflow tests complete workflow
func TestCLI_EndToEndWorkflow(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Step 1: List available apps
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "list", "apps")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list apps: %v", err)
		}
		if !strings.Contains(string(output), "test-app") {
			t.Errorf("Expected 'test-app' in list, got: %s", output)
		}
	}

	// Step 2: Create a backup before changes
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "backup", "create", "test-app")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}
	}

	// Step 3: Toggle theme to light
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "test-app", "theme", "light")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to toggle theme: %v", err)
		}
	}

	// Step 4: Cycle font size
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "cycle", "test-app", "font-size")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to cycle font-size: %v", err)
		}
	}

	// Step 5: Apply a preset
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "preset", "test-app", "default")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to apply preset: %v", err)
		}
	}

	// Step 6: List backups to ensure they exist
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "backup", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}
		if !strings.Contains(string(output), "test-app") {
			t.Errorf("Expected backup for 'test-app' in list, got: %s", output)
		}
	}
}

// TestCLI_MultipleFormats tests different config formats
func TestCLI_MultipleFormats(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, yamlCleanup := setupYAMLTestConfig(t)
	defer yamlCleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Test YAML format
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "yaml-app", "theme", "light")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to toggle YAML app: %v\nOutput: %s", err, output)
		}
	}

	// Verify YAML file was updated
	yamlPath := filepath.Join(filepath.Dir(configDir), "yaml-config.yaml")
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Failed to read YAML config: %v", err)
	}

	if !strings.Contains(string(content), "theme: light") {
		t.Errorf("Expected 'theme: light' in YAML config, got: %s", content)
	}
}

// TestCLI_CustomFormat tests custom format (Ghostty)
func TestCLI_CustomFormat(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, customCleanup := setupCustomTestConfig(t)
	defer customCleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Test custom format
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "custom-app", "theme", "GruvboxLight")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to toggle custom app: %v\nOutput: %s", err, output)
		}
	}

	// Verify custom file was updated
	customPath := filepath.Join(filepath.Dir(configDir), "custom-config.conf")
	content, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("Failed to read custom config: %v", err)
	}

	if !strings.Contains(string(content), "theme = GruvboxLight") {
		t.Errorf("Expected 'theme = GruvboxLight' in custom config, got: %s", content)
	}
}

// TestCLI_HelpCommands tests help functionality
func TestCLI_HelpCommands(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Test root help
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get help: %v", err)
		}
		if !strings.Contains(string(output), "ZeroUI") {
			t.Errorf("Expected 'ZeroUI' in help output, got: %s", output)
		}
		if !strings.Contains(string(output), "toggle") {
			t.Errorf("Expected 'toggle' command in help output, got: %s", output)
		}
	}

	// Test subcommand help
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to get toggle help: %v", err)
		}
		if !strings.Contains(string(output), "Toggle") && !strings.Contains(string(output), "toggle") {
			t.Errorf("Expected toggle help content, got: %s", output)
		}
	}
}

// TestCLI_ConcurrentOperations tests concurrent CLI operations
func TestCLI_ConcurrentOperations(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	configDir, cleanup := setupTestConfig(t)
	defer cleanup()

	os.Setenv("HOME", filepath.Dir(configDir))

	// Run multiple operations concurrently
	done := make(chan error, 3)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "toggle", "test-app", "theme", "light")
		done <- cmd.Run()
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "cycle", "test-app", "font-size")
		done <- cmd.Run()
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binaryPath, "list", "apps")
		done <- cmd.Run()
	}()

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

// setupYAMLTestConfig creates a YAML test configuration
func setupYAMLTestConfig(t testing.TB) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "zeroui-yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".config", "zeroui")
	appsDir := filepath.Join(configDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	testConfig := `name: yaml-app
path: ` + filepath.Join(tmpDir, "yaml-config.yaml") + `
format: yaml
description: YAML Test application

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"
    description: "Application theme"`

	configPath := filepath.Join(appsDir, "yaml-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write YAML app config: %v", err)
	}

	targetConfig := `theme: dark
font-size: 14`
	targetPath := filepath.Join(tmpDir, "yaml-config.yaml")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0644); err != nil {
		t.Fatalf("Failed to write YAML target config: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return configDir, cleanup
}

// setupCustomTestConfig creates a custom format test configuration
func setupCustomTestConfig(t testing.TB) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "zeroui-custom-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".config", "zeroui")
	appsDir := filepath.Join(configDir, "apps")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatalf("Failed to create apps dir: %v", err)
	}

	testConfig := `name: custom-app
path: ` + filepath.Join(tmpDir, "custom-config.conf") + `
format: custom
description: Custom format test application

fields:
  theme:
    type: choice
    values: ["GruvboxDark", "GruvboxLight"]
    default: "GruvboxDark"
    description: "Application theme"`

	configPath := filepath.Join(appsDir, "custom-app.yaml")
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write custom app config: %v", err)
	}

	targetConfig := `# Custom config
theme = GruvboxDark
font-family = JetBrains Mono`
	targetPath := filepath.Join(tmpDir, "custom-config.conf")
	if err := os.WriteFile(targetPath, []byte(targetConfig), 0644); err != nil {
		t.Fatalf("Failed to write custom target config: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return configDir, cleanup
}
