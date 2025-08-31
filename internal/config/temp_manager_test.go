package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTempFileManager(t *testing.T) {
	// Create temp manager with custom options for testing
	opts := &TempFileOptions{
		TempDir:    t.TempDir(),
		MaxBackups: 3,
		MaxTempAge: 1 * time.Hour,
		BufferSize: 4096,
	}
	manager, err := NewTempFileManagerWithOptions(opts)
	if err != nil {
		t.Fatalf("Failed to create temp manager: %v", err)
	}
	defer manager.Close()

	// Create a test file
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.json")
	testContent := `{"key": "value"}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("CreateTempCopy", func(t *testing.T) {
		tempFile, err := manager.CreateTempCopy(testFile)
		if err != nil {
			t.Fatalf("Failed to create temp copy: %v", err)
		}
		defer manager.Rollback(tempFile) // Clean up after test

		// Verify temp file exists
		if _, err := os.Stat(tempFile.TempPath); err != nil {
			t.Errorf("Temp file does not exist: %v", err)
		}

		// Verify content matches
		content, err := os.ReadFile(tempFile.TempPath)
		if err != nil {
			t.Errorf("Failed to read temp file: %v", err)
		}
		if string(content) != testContent {
			t.Errorf("Content mismatch: got %s, want %s", content, testContent)
		}

		// Verify lock file exists
		if _, err := os.Stat(tempFile.LockFile); err != nil {
			t.Errorf("Lock file does not exist: %v", err)
		}
	})

	t.Run("ValidateTemp", func(t *testing.T) {
		// Create a separate test file for this test
		validateFile := filepath.Join(testDir, "validate.json")
		if err := os.WriteFile(validateFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create validate test file: %v", err)
		}

		tempFile, err := manager.CreateTempCopy(validateFile)
		if err != nil {
			t.Fatalf("Failed to create temp copy: %v", err)
		}
		defer manager.Rollback(tempFile)

		// Should pass validation
		if err := manager.ValidateTemp(tempFile); err != nil {
			t.Errorf("Validation failed: %v", err)
		}

		// Delete temp file and validate again
		os.Remove(tempFile.TempPath)
		if err := manager.ValidateTemp(tempFile); err == nil {
			t.Error("Expected validation to fail for missing file")
		}
	})

	t.Run("CommitTemp", func(t *testing.T) {
		// Create a new test file
		commitFile := filepath.Join(testDir, "commit.json")
		if err := os.WriteFile(commitFile, []byte(`{"old": "data"}`), 0644); err != nil {
			t.Fatalf("Failed to create commit test file: %v", err)
		}

		tempFile, err := manager.CreateTempCopy(commitFile)
		if err != nil {
			t.Fatalf("Failed to create temp copy: %v", err)
		}

		// Modify temp file
		newContent := `{"new": "data"}`
		if err := os.WriteFile(tempFile.TempPath, []byte(newContent), 0644); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		// Commit changes
		if err := manager.CommitTemp(tempFile); err != nil {
			t.Errorf("Failed to commit: %v", err)
		}

		// Verify original file has new content
		content, err := os.ReadFile(commitFile)
		if err != nil {
			t.Errorf("Failed to read committed file: %v", err)
		}
		if string(content) != newContent {
			t.Errorf("Content not updated: got %s, want %s", content, newContent)
		}

		// Verify backup exists
		if _, err := os.Stat(tempFile.BackupPath); err != nil {
			t.Errorf("Backup file does not exist: %v", err)
		}

		// Verify temp file is cleaned up
		if _, err := os.Stat(tempFile.TempPath); err == nil {
			t.Error("Temp file not cleaned up after commit")
		}
	})

	t.Run("Rollback", func(t *testing.T) {
		// Create a test file
		rollbackFile := filepath.Join(testDir, "rollback.json")
		originalContent := `{"original": "data"}`
		if err := os.WriteFile(rollbackFile, []byte(originalContent), 0644); err != nil {
			t.Fatalf("Failed to create rollback test file: %v", err)
		}

		tempFile, err := manager.CreateTempCopy(rollbackFile)
		if err != nil {
			t.Fatalf("Failed to create temp copy: %v", err)
		}

		// Modify temp file
		if err := os.WriteFile(tempFile.TempPath, []byte(`{"modified": "data"}`), 0644); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		// Rollback changes
		if err := manager.Rollback(tempFile); err != nil {
			t.Errorf("Failed to rollback: %v", err)
		}

		// Verify temp file is cleaned up
		if _, err := os.Stat(tempFile.TempPath); err == nil {
			t.Error("Temp file not cleaned up after rollback")
		}

		// Verify original file unchanged
		content, err := os.ReadFile(rollbackFile)
		if err != nil {
			t.Errorf("Failed to read original file: %v", err)
		}
		if string(content) != originalContent {
			t.Errorf("Original file modified: got %s, want %s", content, originalContent)
		}
	})

	t.Run("ConcurrentEditPrevention", func(t *testing.T) {
		// Create a separate test file for this test
		concurrentFile := filepath.Join(testDir, "concurrent.json")
		if err := os.WriteFile(concurrentFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create concurrent test file: %v", err)
		}

		// Create temp copy
		tempFile1, err := manager.CreateTempCopy(concurrentFile)
		if err != nil {
			t.Fatalf("First create failed: %v", err)
		}

		// Try to create another copy (should fail)
		_, err = manager.CreateTempCopy(concurrentFile)
		if err == nil {
			t.Error("Expected error for concurrent edit, got nil")
		}

		// Clean up first copy
		manager.Rollback(tempFile1)

		// Now should be able to create again
		tempFile2, err := manager.CreateTempCopy(concurrentFile)
		if err != nil {
			t.Errorf("Failed to create after cleanup: %v", err)
		}
		manager.Rollback(tempFile2)
	})

	t.Run("CleanupStale", func(t *testing.T) {
		// Create a separate test file for this test
		staleFile := filepath.Join(testDir, "stale.json")
		if err := os.WriteFile(staleFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create stale test file: %v", err)
		}

		// Create a temp file
		_, err := manager.CreateTempCopy(staleFile)
		if err != nil {
			t.Fatalf("Failed to create temp copy: %v", err)
		}

		// Make it stale by modifying creation time in the manager's internal map
		// We need to directly modify the manager's map since CreatedAt is part of internal state
		manager.mu.Lock()
		if tf, exists := manager.tempFiles[staleFile]; exists {
			tf.CreatedAt = time.Now().Add(-2 * time.Hour)
		}
		manager.mu.Unlock()

		// Clean up stale files older than 1 hour
		if err := manager.CleanupStale(1 * time.Hour); err != nil {
			t.Errorf("Failed to cleanup stale: %v", err)
		}

		// Verify temp file is gone
		if _, exists := manager.GetTempFile(staleFile); exists {
			t.Error("Stale temp file not cleaned up")
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		metrics := manager.GetMetrics()

		// Check that metrics are present
		if _, ok := metrics["operations"]; !ok {
			t.Error("Missing operations metric")
		}

		if _, ok := metrics["errors"]; !ok {
			t.Error("Missing errors metric")
		}

		if _, ok := metrics["temp_files"]; !ok {
			t.Error("Missing temp_files metric")
		}

		// Operations should be greater than 0 after all the tests
		if ops, ok := metrics["operations"].(uint64); ok && ops == 0 {
			t.Error("Expected operations count to be greater than 0")
		}
	})
}

func TestIntegrityChecker(t *testing.T) {
	checker := NewIntegrityChecker()
	testDir := t.TempDir()

	t.Run("CalculateChecksum", func(t *testing.T) {
		testFile := filepath.Join(testDir, "checksum.txt")
		content := "test content"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		checksum1, err := checker.CalculateChecksum(testFile)
		if err != nil {
			t.Errorf("Failed to calculate checksum: %v", err)
		}

		if checksum1 == "" {
			t.Error("Checksum is empty")
		}

		// Same content should produce same checksum
		checksum2, err := checker.CalculateChecksum(testFile)
		if err != nil {
			t.Errorf("Failed to calculate checksum: %v", err)
		}

		if checksum1 != checksum2 {
			t.Errorf("Checksums don't match: %s != %s", checksum1, checksum2)
		}
	})

	t.Run("ValidateFormat", func(t *testing.T) {
		tests := []struct {
			name    string
			ext     string
			content string
			valid   bool
		}{
			{"Valid JSON", ".json", `{"key": "value"}`, true},
			{"Invalid JSON", ".json", `{invalid json}`, false},
			{"Valid YAML", ".yaml", `key: value`, true},
			{"Invalid YAML", ".yaml", `[unclosed`, false},
			{"Valid TOML", ".toml", `key = "value"`, true},
			{"Text file", ".conf", `some config`, true},
			{"Binary file", ".conf", "\x00binary\x00", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				testFile := filepath.Join(testDir, "test"+tt.ext)
				if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				err := checker.ValidateFormat(testFile)
				if tt.valid && err != nil {
					t.Errorf("Expected valid, got error: %v", err)
				}
				if !tt.valid && err == nil {
					t.Error("Expected invalid, got no error")
				}
			})
		}
	})

	t.Run("CompareFiles", func(t *testing.T) {
		file1 := filepath.Join(testDir, "file1.txt")
		file2 := filepath.Join(testDir, "file2.txt")
		file3 := filepath.Join(testDir, "file3.txt")

		content1 := "same content"
		content2 := "different content"

		os.WriteFile(file1, []byte(content1), 0644)
		os.WriteFile(file2, []byte(content1), 0644)
		os.WriteFile(file3, []byte(content2), 0644)

		// Same content
		same, err := checker.CompareFiles(file1, file2)
		if err != nil {
			t.Errorf("Failed to compare: %v", err)
		}
		if !same {
			t.Error("Expected files to be the same")
		}

		// Different content
		same, err = checker.CompareFiles(file1, file3)
		if err != nil {
			t.Errorf("Failed to compare: %v", err)
		}
		if same {
			t.Error("Expected files to be different")
		}
	})

	t.Run("IntegrityReport", func(t *testing.T) {
		testFile := filepath.Join(testDir, "report.json")
		if err := os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		report, err := checker.CreateIntegrityReport(testFile)
		if err != nil {
			t.Errorf("Failed to create report: %v", err)
		}

		if !report.Valid {
			t.Errorf("Expected valid report, got: %+v", report)
		}

		if report.Checksum == "" {
			t.Error("Report missing checksum")
		}

		if !report.FormatValid {
			t.Error("Expected format to be valid")
		}
	})
}
