package service_test

import (
	"os"
	"path/filepath"
	"testing"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/mrtkrcm/ZeroUI/internal/config"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

func TestConfigService_Integration(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := ioutil.TempDir("", "zeroui-service-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test config structure
	appsDir := filepath.Join(tmpDir, ".config", "zeroui", "apps")
	require.NoError(t, os.MkdirAll(appsDir, 0755))

	// Create test app config
	testAppConfig := `name: test-app
path: ` + filepath.Join(tmpDir, "test-config.json") + `
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"
    description: "Application theme"
`

	configPath := filepath.Join(appsDir, "test-app.yaml")
	require.NoError(t, ioutil.WriteFile(configPath, []byte(testAppConfig), 0644))

	// Create target config file
	targetConfig := `{"theme": "dark"}`
	targetPath := filepath.Join(tmpDir, "test-config.json")
	require.NoError(t, ioutil.WriteFile(targetPath, []byte(targetConfig), 0644))

	// Set up service components
	configLoader, err := config.NewLoader()
	require.NoError(t, err)
	configLoader.SetConfigDir(filepath.Join(tmpDir, ".config", "zeroui"))
	
	testLogger := logger.New(logger.DefaultConfig())
	toggleEngine := toggle.NewEngineWithDeps(configLoader, testLogger)
	
	service := service.NewConfigService(toggleEngine, configLoader, testLogger)

	t.Run("ListApplications", func(t *testing.T) {
		apps, err := service.ListApplications()
		require.NoError(t, err)
		assert.Contains(t, apps, "test-app")
	})

	t.Run("ToggleConfiguration", func(t *testing.T) {
		err := service.ToggleConfiguration("test-app", "theme", "light")
		require.NoError(t, err)
		
		// Verify the change was applied
		content, err := ioutil.ReadFile(targetPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "light")
	})

	t.Run("CycleConfiguration", func(t *testing.T) {
		// Reset to known state
		err := service.ToggleConfiguration("test-app", "theme", "dark")
		require.NoError(t, err)
		
		// Cycle should change to next value
		err = service.CycleConfiguration("test-app", "theme")
		require.NoError(t, err)
		
		// Verify the change was applied
		content, err := ioutil.ReadFile(targetPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "light")
	})
}

func TestConfigService_ErrorHandling(t *testing.T) {
	// Create empty service for error testing
	tmpDir, err := ioutil.TempDir("", "zeroui-service-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configLoader, err := config.NewLoader()
	require.NoError(t, err)
	configLoader.SetConfigDir(filepath.Join(tmpDir, ".config", "zeroui"))
	
	testLogger := logger.New(logger.DefaultConfig())
	toggleEngine := toggle.NewEngineWithDeps(configLoader, testLogger)
	
	service := service.NewConfigService(toggleEngine, configLoader, testLogger)

	t.Run("NonexistentApp", func(t *testing.T) {
		err := service.ToggleConfiguration("nonexistent-app", "theme", "dark")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("EmptyAppName", func(t *testing.T) {
		err := service.ToggleConfiguration("", "key", "value")
		assert.Error(t, err)
	})

	t.Run("EmptyKey", func(t *testing.T) {
		err := service.CycleConfiguration("app", "")
		assert.Error(t, err)
	})
}

// Benchmark tests for performance
func BenchmarkConfigService_ListApplications(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "zeroui-service-bench")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	appsDir := filepath.Join(tmpDir, ".config", "zeroui", "apps")
	require.NoError(b, os.MkdirAll(appsDir, 0755))

	// Create multiple test apps
	for i := 0; i < 10; i++ {
		testAppConfig := `name: test-app-` + string(rune('0'+i)) + `
path: /tmp/test-config.json
format: json
description: Test application

fields:
  theme:
    type: choice
    values: ["dark", "light"]
    default: "dark"
`
		configPath := filepath.Join(appsDir, "test-app-"+string(rune('0'+i))+".yaml")
		require.NoError(b, ioutil.WriteFile(configPath, []byte(testAppConfig), 0644))
	}

	configLoader, err := config.NewLoader()
	require.NoError(b, err)
	configLoader.SetConfigDir(filepath.Join(tmpDir, ".config", "zeroui"))
	
	testLogger := logger.New(logger.DefaultConfig())
	toggleEngine := toggle.NewEngineWithDeps(configLoader, testLogger)
	
	service := service.NewConfigService(toggleEngine, configLoader, testLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ListApplications()
	}
}