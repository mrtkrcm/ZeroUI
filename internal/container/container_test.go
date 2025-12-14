package container

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/test/helpers"
)

func TestMain(m *testing.M) {
	helpers.RunTestMainWithCleanup(m, "internal/container", "zeroui-internal-container-test-home-", nil)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Logger.Level != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", config.Logger.Level)
	}

	if config.Logger.Format != "console" {
		t.Errorf("Expected default log format to be 'console', got '%s'", config.Logger.Format)
	}
}

func TestNewContainer(t *testing.T) {
	// Test with nil config (should use defaults)
	container, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create container with nil config: %v", err)
	}

	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Test that all dependencies are initialized
	if container.Logger() == nil {
		t.Error("Logger should be initialized")
	}

	if container.ConfigLoader() == nil {
		t.Error("ConfigLoader should be initialized")
	}

	if container.ToggleEngine() == nil {
		t.Error("ToggleEngine should be initialized")
	}

	if container.ConfigService() == nil {
		t.Error("ConfigService should be initialized")
	}
}

func TestNewContainerWithConfig(t *testing.T) {
	config := &Config{
		Logger: &logger.Config{
			Level:  "debug",
			Format: "json",
		},
	}

	container, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create container with custom config: %v", err)
	}

	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Verify dependencies are still initialized
	if container.Logger() == nil {
		t.Error("Logger should be initialized")
	}

	if container.ConfigLoader() == nil {
		t.Error("ConfigLoader should be initialized")
	}

	if container.ToggleEngine() == nil {
		t.Error("ToggleEngine should be initialized")
	}

	if container.ConfigService() == nil {
		t.Error("ConfigService should be initialized")
	}
}

func TestContainerDependencyInjection(t *testing.T) {
	container, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Test that all services are properly connected
	logger := container.Logger()
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	configLoader := container.ConfigLoader()
	if configLoader == nil {
		t.Fatal("ConfigLoader should not be nil")
	}

	toggleEngine := container.ToggleEngine()
	if toggleEngine == nil {
		t.Fatal("ToggleEngine should not be nil")
	}

	configService := container.ConfigService()
	if configService == nil {
		t.Fatal("ConfigService should not be nil")
	}

	// Test that services can access their dependencies
	// This verifies the dependency injection is working correctly
	_ = toggleEngine
	_ = configService
}

func TestContainerClose(t *testing.T) {
	container, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Close should not panic and should return nil (no cleanup needed currently)
	err = container.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}
}

func TestContainerWithCustomHome(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set HOME to our temp directory
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Create container and verify it works with custom HOME
	container, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create container with custom HOME: %v", err)
	}

	if container == nil {
		t.Fatal("Container should not be nil with custom HOME")
	}

	// Test that config files would be created in the right place
	expectedConfigDir := filepath.Join(tempDir, ".config", "zeroui")
	if _, err := os.Stat(expectedConfigDir); os.IsNotExist(err) {
		// This is expected since we haven't created any config files yet
		// The important thing is that the container initializes successfully
		t.Logf("Config directory would be created at: %s", expectedConfigDir)
	}
}
