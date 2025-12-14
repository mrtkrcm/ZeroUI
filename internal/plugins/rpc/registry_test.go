package rpc

import (
	"os"
	"testing"
)

func TestRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	registry := NewRegistry(tempDir)
	defer registry.Shutdown()

	t.Run("Discovery", func(t *testing.T) {
		plugins, err := registry.DiscoverPlugins()
		if err == nil {
			t.Logf("Discovered %d plugins: %v", len(plugins), plugins)
		} else {
			t.Logf("Discovery failed (expected): %v", err)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		stats := registry.GetStats()
		if loadedCount, ok := stats["loaded_plugins"].(int); ok {
			if loadedCount != 0 {
				t.Errorf("Expected 0 loaded plugins, got %d", loadedCount)
			}
		}
	})

	t.Run("ListEmpty", func(t *testing.T) {
		plugins := registry.ListPlugins()
		if len(plugins) != 0 {
			t.Errorf("Expected 0 plugins, got %d", len(plugins))
		}
	})

	t.Run("LoadNonExistent", func(t *testing.T) {
		_, err := registry.GetPlugin("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent plugin")
		}
	})
}
