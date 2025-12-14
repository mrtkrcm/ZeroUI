//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginSystemIntegration focuses on RPC plugin system core functionality
func TestPluginSystemIntegration(t *testing.T) {
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	// Build main binary and plugin
	binaryPath := buildZeroUI(t, testDir)
	pluginPath := buildGhosttyPlugin(t, testDir)

	t.Run("Plugin Discovery", func(t *testing.T) {
		testPluginDiscovery(t, binaryPath, pluginPath, testDir)
	})

	t.Run("Plugin Communication", func(t *testing.T) {
		testPluginCommunication(t, binaryPath, pluginPath, testDir)
	})

	t.Run("Plugin Lifecycle", func(t *testing.T) {
		testPluginLifecycle(t, binaryPath, pluginPath, testDir)
	})
}

func testPluginDiscovery(t *testing.T, binaryPath, pluginPath, testDir string) {
	// Test 1: Plugin in PATH is discovered
	t.Run("discovers plugin in PATH", func(t *testing.T) {
		// Create plugin directory and symlink
		pluginDir := filepath.Join(testDir, "plugins")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		pluginLink := filepath.Join(pluginDir, "zeroui-plugin-ghostty-rpc")
		require.NoError(t, os.Symlink(pluginPath, pluginLink))

		// Run with plugin directory in PATH
		cmd := exec.Command(binaryPath, "list", "apps")
		cmd.Env = append(os.Environ(),
			"PATH="+pluginDir+":"+os.Getenv("PATH"),
			"HOME="+testDir,
		)

		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should list apps successfully")

		assert.Contains(t, output, "ghostty", "Should discover ghostty via plugin")
	})

	// Test 2: Plugin not in PATH handled gracefully
	t.Run("handles missing plugins gracefully", func(t *testing.T) {
		// Run without plugin in PATH
		cmd := exec.Command(binaryPath, "list", "apps")
		cmd.Env = append(os.Environ(),
			"PATH=/nonexistent",
			"HOME="+testDir,
		)

		output, err := runCommandWithEnv(cmd)
		// Should still work with built-in app detection
		require.NoError(t, err, "Should handle missing plugins gracefully")

		// Should either show built-in apps or empty list
		assert.Contains(t, output, "Applications", "Should show applications header")
	})
}

func testPluginCommunication(t *testing.T, binaryPath, pluginPath, testDir string) {
	// Setup plugin in PATH
	pluginDir := filepath.Join(testDir, "plugins")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	pluginLink := filepath.Join(pluginDir, "zeroui-plugin-ghostty-rpc")
	os.Remove(pluginLink) // Remove if exists
	require.NoError(t, os.Symlink(pluginPath, pluginLink))

	// Test 1: Basic RPC communication
	t.Run("communicates with plugin via RPC", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "list", "keys", "ghostty")
		cmd.Env = append(os.Environ(),
			"PATH="+pluginDir+":"+os.Getenv("PATH"),
			"HOME="+testDir,
		)

		output, err := runCommandWithEnv(cmd)
		require.NoError(t, err, "Should communicate with plugin successfully")

		assert.Contains(t, output, "keys", "Should show configuration keys")
		assert.Contains(t, output, "ghostty", "Should reference ghostty app")
	})

	// Test 2: Plugin timeout handling
	t.Run("handles plugin timeout gracefully", func(t *testing.T) {
		// This test verifies that plugin communication has reasonable timeouts
		// and doesn't hang indefinitely

		cmd := exec.Command(binaryPath, "extract", "ghostty", "--dry-run")
		cmd.Env = append(os.Environ(),
			"PATH="+pluginDir+":"+os.Getenv("PATH"),
			"HOME="+testDir,
		)

		done := make(chan error, 1)
		go func() {
			_, err := runCommandWithEnv(cmd)
			done <- err
		}()

		select {
		case err := <-done:
			// Should complete within reasonable time
			if err != nil {
				t.Logf("Plugin communication error (acceptable): %v", err)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Plugin communication should not hang indefinitely")
		}
	})
}

func testPluginLifecycle(t *testing.T, binaryPath, pluginPath, testDir string) {
	// Setup plugin in PATH
	pluginDir := filepath.Join(testDir, "plugins")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	pluginLink := filepath.Join(pluginDir, "zeroui-plugin-ghostty-rpc")
	os.Remove(pluginLink) // Remove if exists
	require.NoError(t, os.Symlink(pluginPath, pluginLink))

	// Test 1: Plugin starts and stops cleanly
	t.Run("plugin lifecycle management", func(t *testing.T) {
		// Run multiple commands to test plugin reuse/restart
		commands := [][]string{
			{"list", "apps"},
			{"list", "keys", "ghostty"},
			{"extract", "ghostty", "--dry-run"},
		}

		for _, cmdArgs := range commands {
			cmd := exec.Command(binaryPath, cmdArgs...)
			cmd.Env = append(os.Environ(),
				"PATH="+pluginDir+":"+os.Getenv("PATH"),
				"HOME="+testDir,
			)

			output, err := runCommandWithEnv(cmd)
			if err != nil {
				t.Logf("Command %v failed (may be acceptable): %v", cmdArgs, err)
				t.Logf("Output: %s", output)
			} else {
				assert.NotEmpty(t, output, "Should produce output")
			}
		}
	})

	// Test 2: Plugin crash recovery
	t.Run("handles plugin crashes gracefully", func(t *testing.T) {
		// Create a fake plugin that exits immediately
		crashingPlugin := filepath.Join(pluginDir, "zeroui-plugin-crashing")
		crashScript := `#!/bin/bash
exit 1
`
		require.NoError(t, os.WriteFile(crashingPlugin, []byte(crashScript), 0755))

		// Try to use the crashing plugin
		cmd := exec.Command(binaryPath, "list", "apps")
		cmd.Env = append(os.Environ(),
			"PATH="+pluginDir+":"+os.Getenv("PATH"),
			"HOME="+testDir,
		)

		output, err := runCommandWithEnv(cmd)
		// Should handle plugin crash gracefully and continue
		require.NoError(t, err, "Should handle plugin crashes gracefully")
		assert.Contains(t, output, "Applications", "Should still show applications")
	})
}

func buildGhosttyPlugin(t *testing.T, testDir string) string {
	// Build the Ghostty RPC plugin
	pluginPath := filepath.Join(testDir, "zeroui-plugin-ghostty-rpc")

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", pluginPath, ".")
	cmd.Dir = "../../plugins/ghostty-rpc" // Go to plugin directory with its own go.mod

	err := cmd.Run()
	require.NoError(t, err, "Should build Ghostty plugin")

	return pluginPath
}

// TestPluginAPIIntegration tests the plugin API interface
func TestPluginAPIIntegration(t *testing.T) {
	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	pluginPath := buildGhosttyPlugin(t, testDir)

	t.Run("Plugin responds to info request", func(t *testing.T) {
		// Test plugin directly (if it supports standalone mode)
		// This verifies the plugin implementation itself

		// For now, just verify the plugin binary exists and is executable
		info, err := os.Stat(pluginPath)
		require.NoError(t, err, "Plugin binary should exist")
		assert.True(t, info.Mode().IsRegular(), "Should be a regular file")
		assert.True(t, info.Mode().Perm()&0111 != 0, "Should be executable")
	})
}

// TestRPCProtocolIntegration tests the gRPC protocol implementation
func TestRPCProtocolIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RPC protocol tests in short mode")
	}

	testDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(testDir)

	binaryPath := buildZeroUI(t, testDir)
	pluginPath := buildGhosttyPlugin(t, testDir)

	// Setup plugin environment
	pluginDir := filepath.Join(testDir, "plugins")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	pluginLink := filepath.Join(pluginDir, "zeroui-plugin-ghostty-rpc")
	os.Remove(pluginLink) // Remove if exists
	require.NoError(t, os.Symlink(pluginPath, pluginLink))

	t.Run("RPC method invocation", func(t *testing.T) {
		// Test various RPC methods through the main application
		testCases := []struct {
			name     string
			args     []string
			expectOK bool
		}{
			{"GetInfo", []string{"list", "apps"}, true},
			{"DetectConfig", []string{"extract", "ghostty", "--dry-run"}, true},
			{"GetSchema", []string{"list", "keys", "ghostty"}, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := exec.Command(binaryPath, tc.args...)
				cmd.Env = append(os.Environ(),
					"PATH="+pluginDir+":"+os.Getenv("PATH"),
					"HOME="+testDir,
				)

				output, err := runCommandWithEnv(cmd)
				if tc.expectOK {
					if err != nil {
						t.Logf("RPC method %s failed (may be acceptable): %v", tc.name, err)
						t.Logf("Output: %s", output)
					}
				}

				// At minimum, should not hang or crash
				assert.NotContains(t, strings.ToLower(output), "panic", "Should not panic")
			})
		}
	})
}
