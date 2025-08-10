package atomic

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/config"
)

// setupAtomicTest creates a test environment for atomic operations
func setupAtomicTest(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "atomic-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// TestNewManager tests creating a new atomic manager
func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.locks == nil {
		t.Error("Expected locks map to be initialized")
	}

	if manager.recovery == nil {
		t.Error("Expected recovery manager to be initialized")
	}
}

// TestOperation_Basic tests basic atomic operations
func TestOperation_Basic(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "test.json")

	// Create initial config file
	initialConfig := `{"setting": "initial_value"}`
	if err := ioutil.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Test atomic write operation
	operation := manager.BeginOperation(configPath)

	// Create backup
	err = operation.CreateBackup("test-app")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Write new config
	appConfig := &config.AppConfig{
		Path:   configPath,
		Format: "json",
	}

	newConfigData := map[string]interface{}{
		"setting": "new_value",
		"added":   "additional_setting",
	}

	err = operation.WriteConfig(appConfig, newConfigData)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Commit operation
	operation.Commit()

	// Verify the file was updated
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read updated config: %v", err)
	}

	if !containsString(string(content), "new_value") {
		t.Error("Expected config to contain new_value")
	}

	if !containsString(string(content), "additional_setting") {
		t.Error("Expected config to contain additional_setting")
	}
}

// TestOperation_Rollback tests operation rollback
func TestOperation_Rollback(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "test.json")

	// Create initial config file
	originalContent := `{"setting": "original_value"}`
	if err := ioutil.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Begin operation
	operation := manager.BeginOperation(configPath)

	// Create backup
	err = operation.CreateBackup("test-app")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Write new config
	appConfig := &config.AppConfig{
		Path:   configPath,
		Format: "json",
	}

	newConfigData := map[string]interface{}{
		"setting": "modified_value",
	}

	err = operation.WriteConfig(appConfig, newConfigData)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Verify the file was modified
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read modified config: %v", err)
	}

	if !containsString(string(content), "modified_value") {
		t.Error("Expected config to contain modified_value")
	}

	// Rollback the operation
	err = operation.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify the file was restored
	content, err = ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if !containsString(string(content), "original_value") {
		t.Error("Expected config to be restored to original_value")
	}

	if containsString(string(content), "modified_value") {
		t.Error("Expected modified_value to be gone after rollback")
	}
}

// TestReadOperation tests read operations
func TestReadOperation(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "test.json")

	// Create config file
	configContent := `{"setting1": "value1", "setting2": 42, "setting3": true}`
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Test read operation
	readOp := manager.BeginReadOperation(configPath)

	appConfig := &config.AppConfig{
		Path:   configPath,
		Format: "json",
	}

	configData, err := readOp.ReadConfig(appConfig)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	readOp.Complete()

	// Verify data was read correctly
	if configData["setting1"] != "value1" {
		t.Errorf("Expected setting1 to be 'value1', got '%v'", configData["setting1"])
	}

	// The JSON unmarshaling will read 42 as float64
	if configData["setting2"] != float64(42) {
		t.Errorf("Expected setting2 to be 42, got '%v'", configData["setting2"])
	}

	if configData["setting3"] != true {
		t.Errorf("Expected setting3 to be true, got '%v'", configData["setting3"])
	}
}

// TestConcurrentOperations tests concurrent access
func TestConcurrentOperations(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "concurrent.json")

	// Create initial config
	initialConfig := `{"counter": 0}`
	if err := ioutil.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Test concurrent read operations (should be allowed)
	t.Run("Concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make([]map[string]interface{}, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				readOp := manager.BeginReadOperation(configPath)
				defer readOp.Complete()

				appConfig := &config.AppConfig{
					Path:   configPath,
					Format: "json",
				}

				data, err := readOp.ReadConfig(appConfig)
				if err != nil {
					t.Errorf("Failed to read config in goroutine %d: %v", index, err)
					return
				}

				results[index] = data
			}(i)
		}

		wg.Wait()

		// All reads should succeed and return the same data
		for i, result := range results {
			if result == nil {
				continue // Error case already logged
			}

			if result["counter"] != float64(0) {
				t.Errorf("Read %d: expected counter 0, got %v", i, result["counter"])
			}
		}
	})

	// Test that write operations are exclusive
	t.Run("Exclusive writes", func(t *testing.T) {
		var wg sync.WaitGroup
		writeCount := 0
		var mu sync.Mutex

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				operation := manager.BeginOperation(configPath)

				// Simulate some work time
				time.Sleep(10 * time.Millisecond)

				mu.Lock()
				writeCount++
				currentCount := writeCount
				mu.Unlock()

				// Write the current count
				appConfig := &config.AppConfig{
					Path:   configPath,
					Format: "json",
				}

				configData := map[string]interface{}{
					"counter": currentCount,
					"writer":  index,
				}

				if err := operation.WriteConfig(appConfig, configData); err != nil {
					t.Errorf("Failed to write config in goroutine %d: %v", index, err)
				}

				operation.Commit()
			}(i)
		}

		wg.Wait()

		// Final value should be consistent (last writer wins due to serialization)
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read final config: %v", err)
		}

		if !containsString(string(content), "\"counter\": 3") {
			t.Errorf("Expected counter to be 3 in final result, got: %s", string(content))
		}
	})
}

// TestTransaction tests transaction functionality
func TestTransaction(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create multiple config files
	configPaths := []string{
		filepath.Join(tmpDir, "config1.json"),
		filepath.Join(tmpDir, "config2.json"),
		filepath.Join(tmpDir, "config3.json"),
	}

	appNames := []string{"app1", "app2", "app3"}

	// Create initial files
	for i, configPath := range configPaths {
		content := fmt.Sprintf(`{"id": %d, "value": "initial"}`, i+1)
		if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write initial config %d: %v", i+1, err)
		}
	}

	// Test successful transaction
	t.Run("Successful transaction", func(t *testing.T) {
		tx := manager.BeginTransaction()

		// Add operations
		operations := make([]*Operation, len(configPaths))
		for i, configPath := range configPaths {
			operations[i] = tx.AddOperation(configPath)
		}

		// Create backups
		if err := tx.CreateBackups(appNames); err != nil {
			t.Fatalf("Failed to create backups: %v", err)
		}

		// Write to all files
		for i, op := range operations {
			appConfig := &config.AppConfig{
				Path:   configPaths[i],
				Format: "json",
			}

			configData := map[string]interface{}{
				"id":    i + 1,
				"value": "updated",
			}

			if err := op.WriteConfig(appConfig, configData); err != nil {
				t.Fatalf("Failed to write config %d: %v", i+1, err)
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify all files were updated
		for i, configPath := range configPaths {
			content, err := ioutil.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read updated config %d: %v", i+1, err)
			}

			if !containsString(string(content), "\"value\": \"updated\"") {
				t.Errorf("Config %d was not updated properly", i+1)
			}
		}
	})

	// Test transaction rollback
	t.Run("Transaction rollback", func(t *testing.T) {
		tx := manager.BeginTransaction()

		// Add operations
		operations := make([]*Operation, len(configPaths))
		for i, configPath := range configPaths {
			operations[i] = tx.AddOperation(configPath)
		}

		// Create backups
		if err := tx.CreateBackups(appNames); err != nil {
			t.Fatalf("Failed to create backups: %v", err)
		}

		// Write to some files
		for i := 0; i < 2; i++ {
			appConfig := &config.AppConfig{
				Path:   configPaths[i],
				Format: "json",
			}

			configData := map[string]interface{}{
				"id":    i + 1,
				"value": "rolled_back",
			}

			if err := operations[i].WriteConfig(appConfig, configData); err != nil {
				t.Fatalf("Failed to write config %d: %v", i+1, err)
			}
		}

		// Rollback transaction
		if err := tx.Rollback(); err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// Verify all files were restored to previous state
		for i, configPath := range configPaths {
			content, err := ioutil.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read rolled back config %d: %v", i+1, err)
			}

			// Should still contain "updated" from previous test, not "rolled_back"
			if !containsString(string(content), "\"value\": \"updated\"") {
				t.Errorf("Config %d was not properly rolled back", i+1)
			}

			if containsString(string(content), "rolled_back") {
				t.Errorf("Config %d still contains rolled back data", i+1)
			}
		}
	})
}

// TestSafeOperation tests safe operation wrapper
func TestSafeOperation(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "safe.json")

	// Create initial config
	initialConfig := `{"status": "original"}`
	if err := ioutil.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Test successful safe operation
	t.Run("Successful operation", func(t *testing.T) {
		safeOp := manager.NewSafeOperation(configPath)

		err := safeOp.Execute("test-app", func(op *Operation) error {
			appConfig := &config.AppConfig{
				Path:   configPath,
				Format: "json",
			}

			configData := map[string]interface{}{
				"status": "success",
			}

			return op.WriteConfig(appConfig, configData)
		})

		if err != nil {
			t.Fatalf("Safe operation failed: %v", err)
		}

		// Verify file was updated
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		if !containsString(string(content), "\"status\": \"success\"") {
			t.Error("Expected config to contain success status")
		}
	})

	// Test failed safe operation (should rollback automatically)
	t.Run("Failed operation with auto rollback", func(t *testing.T) {
		safeOp := manager.NewSafeOperation(configPath)

		err := safeOp.Execute("test-app", func(op *Operation) error {
			appConfig := &config.AppConfig{
				Path:   configPath,
				Format: "json",
			}

			configData := map[string]interface{}{
				"status": "failed",
			}

			// Write the config first
			if err := op.WriteConfig(appConfig, configData); err != nil {
				return err
			}

			// Then return an error to trigger rollback
			return fmt.Errorf("simulated failure")
		})

		if err == nil {
			t.Fatal("Expected error from failed operation")
		}

		// Verify file was rolled back
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read rolled back config: %v", err)
		}

		if !containsString(string(content), "\"status\": \"success\"") {
			t.Error("Expected config to be rolled back to success status")
		}

		if containsString(string(content), "failed") {
			t.Error("Expected failed data to be rolled back")
		}
	})
}

// TestLockManager tests high-level lock manager
func TestLockManager(t *testing.T) {
	tmpDir, cleanup := setupAtomicTest(t)
	defer cleanup()

	lockManager, err := NewLockManager()
	if err != nil {
		t.Fatalf("Failed to create lock manager: %v", err)
	}

	configPath := filepath.Join(tmpDir, "locktest.json")

	// Create initial config
	initialConfig := `{"counter": 0}`
	if err := ioutil.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Test read lock
	t.Run("Read lock", func(t *testing.T) {
		var readValue interface{}

		err := lockManager.WithReadLock(configPath, func() error {
			// Read the file
			content, err := ioutil.ReadFile(configPath)
			if err != nil {
				return err
			}

			// Simple parsing to get counter value
			if containsString(string(content), "\"counter\": 0") {
				readValue = 0
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Read lock operation failed: %v", err)
		}

		if readValue != 0 {
			t.Errorf("Expected to read counter value 0, got %v", readValue)
		}
	})

	// Test write lock
	t.Run("Write lock", func(t *testing.T) {
		err := lockManager.WithWriteLock(configPath, "test-app", func(op *Operation) error {
			appConfig := &config.AppConfig{
				Path:   configPath,
				Format: "json",
			}

			configData := map[string]interface{}{
				"counter": 1,
			}

			return op.WriteConfig(appConfig, configData)
		})

		if err != nil {
			t.Fatalf("Write lock operation failed: %v", err)
		}

		// Verify file was updated
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read updated config: %v", err)
		}

		if !containsString(string(content), "\"counter\": 1") {
			t.Error("Expected counter to be updated to 1")
		}
	})
}

// TestHealthCheck tests health check functionality
func TestHealthCheck(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	err = manager.HealthCheck()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	stats := manager.Stats()
	if stats == nil {
		t.Error("Expected stats to be returned")
	}

	if _, exists := stats["active_locks"]; !exists {
		t.Error("Expected active_locks in stats")
	}

	if _, exists := stats["backup_stats"]; !exists {
		t.Error("Expected backup_stats in stats")
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
