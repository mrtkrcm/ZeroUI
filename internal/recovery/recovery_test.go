package recovery

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestBackupManager_CreateBackup tests creating backups
func TestBackupManager_CreateBackup(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recovery-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bm := &BackupManager{
		backupDir: tmpDir,
	}

	// Create test config file
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "test.conf")
	configContent := "theme = dark\nfont-size = 14"
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create backup
	backupPath, err := bm.CreateBackup(configPath, "test-app")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup was created
	if backupPath == "" {
		t.Error("Expected backup path to be returned")
	}

	// Verify backup content
	backupContent, err := ioutil.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(backupContent) != configContent {
		t.Errorf("Expected backup content '%s', got '%s'", configContent, string(backupContent))
	}

	// Test backing up non-existent file
	backupPath2, err := bm.CreateBackup("/nonexistent/file", "test-app")
	if err != nil {
		t.Fatalf("Failed to handle non-existent file: %v", err)
	}

	if backupPath2 != "" {
		t.Error("Expected empty backup path for non-existent file")
	}
}

// TestBackupManager_RestoreBackup tests restoring from backup
func TestBackupManager_RestoreBackup(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recovery-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bm := &BackupManager{
		backupDir: tmpDir,
	}

	// Create backup file
	backupContent := "theme = light\nfont-size = 16"
	backupPath := filepath.Join(tmpDir, "test-backup.backup")
	if err := ioutil.WriteFile(backupPath, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Restore to target location
	targetDir := filepath.Join(tmpDir, "target")
	targetPath := filepath.Join(targetDir, "config.conf")

	err = bm.RestoreBackup(backupPath, targetPath)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restored content
	restoredContent, err := ioutil.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restoredContent) != backupContent {
		t.Errorf("Expected restored content '%s', got '%s'", backupContent, string(restoredContent))
	}

	// Test restoring non-existent backup
	err = bm.RestoreBackup("/nonexistent/backup", targetPath)
	if err == nil {
		t.Error("Expected error for non-existent backup")
	}
}

// TestBackupManager_ListBackups tests listing backups
func TestBackupManager_ListBackups(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recovery-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bm := &BackupManager{
		backupDir: tmpDir,
	}

	// Create test backup files
	backups := []struct {
		name    string
		content string
	}{
		{"app1_20230101_120000.backup", "config 1"},
		{"app1_20230101_130000.backup", "config 2"},
		{"app2_20230101_140000.backup", "config 3"},
	}

	for _, backup := range backups {
		path := filepath.Join(tmpDir, backup.name)
		if err := ioutil.WriteFile(path, []byte(backup.content), 0644); err != nil {
			t.Fatalf("Failed to write backup file %s: %v", backup.name, err)
		}
	}

	// List all backups
	allBackups, err := bm.ListBackups("")
	if err != nil {
		t.Fatalf("Failed to list all backups: %v", err)
	}

	if len(allBackups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(allBackups))
	}

	// List backups for specific app
	app1Backups, err := bm.ListBackups("app1")
	if err != nil {
		t.Fatalf("Failed to list app1 backups: %v", err)
	}

	if len(app1Backups) != 2 {
		t.Errorf("Expected 2 app1 backups, got %d", len(app1Backups))
	}

	// Test empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	emptyBm := &BackupManager{
		backupDir: emptyDir,
	}

	emptyBackups, err := emptyBm.ListBackups("")
	if err != nil {
		t.Fatalf("Failed to list backups from empty dir: %v", err)
	}

	if len(emptyBackups) != 0 {
		t.Errorf("Expected 0 backups from empty dir, got %d", len(emptyBackups))
	}
}

// TestSafeOperation tests safe operations with automatic rollback
func TestSafeOperation(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recovery-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create original config file
	configPath := filepath.Join(tmpDir, "config.conf")
	originalContent := "theme = dark"
	if err := ioutil.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write original config: %v", err)
	}

	// Test successful operation
	safeOp, err := NewSafeOperation(configPath, "test-app")
	if err != nil {
		t.Fatalf("Failed to create safe operation: %v", err)
	}

	// Modify the file (simulating a config change)
	newContent := "theme = light"
	if err := ioutil.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to modify config: %v", err)
	}

	// Commit the operation
	if err := safeOp.Commit(); err != nil {
		t.Fatalf("Failed to commit operation: %v", err)
	}

	// Verify file still has new content
	finalContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read final config: %v", err)
	}

	if string(finalContent) != newContent {
		t.Errorf("Expected final content '%s', got '%s'", newContent, string(finalContent))
	}

	// Test rollback scenario
	safeOp2, err := NewSafeOperation(configPath, "test-app")
	if err != nil {
		t.Fatalf("Failed to create second safe operation: %v", err)
	}

	// Modify the file again
	badContent := "invalid config"
	if err := ioutil.WriteFile(configPath, []byte(badContent), 0644); err != nil {
		t.Fatalf("Failed to write bad config: %v", err)
	}

	// Rollback the operation
	if err := safeOp2.Rollback(); err != nil {
		t.Fatalf("Failed to rollback operation: %v", err)
	}

	// Verify file was restored
	restoredContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(restoredContent) != newContent {
		t.Errorf("Expected restored content '%s', got '%s'", newContent, string(restoredContent))
	}
}

// TestBackupManager_CleanupOldBackups tests cleanup functionality
func TestBackupManager_CleanupOldBackups(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recovery-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bm := &BackupManager{
		backupDir: tmpDir,
	}

	// Create multiple backup files with different timestamps
	backupNames := []string{
		"app1_20230101_120000.backup",
		"app1_20230101_130000.backup", 
		"app1_20230101_140000.backup",
		"app1_20230101_150000.backup",
		"app1_20230101_160000.backup",
	}

	for _, name := range backupNames {
		path := filepath.Join(tmpDir, name)
		if err := ioutil.WriteFile(path, []byte("backup content"), 0644); err != nil {
			t.Fatalf("Failed to write backup file %s: %v", name, err)
		}
		// Sleep to ensure different modification times
		time.Sleep(time.Millisecond * 10)
	}

	// Keep only 3 most recent backups
	err = bm.CleanupOldBackups("app1", 3)
	if err != nil {
		t.Fatalf("Failed to cleanup old backups: %v", err)
	}

	// List remaining backups
	remainingBackups, err := bm.ListBackups("app1")
	if err != nil {
		t.Fatalf("Failed to list remaining backups: %v", err)
	}

	if len(remainingBackups) != 3 {
		t.Errorf("Expected 3 remaining backups, got %d", len(remainingBackups))
	}

	// Test cleanup with no backups to remove
	err = bm.CleanupOldBackups("app1", 5)
	if err != nil {
		t.Fatalf("Failed to cleanup when nothing to remove: %v", err)
	}

	// Test cleanup non-existent app
	err = bm.CleanupOldBackups("nonexistent", 3)
	if err != nil {
		t.Fatalf("Failed to cleanup non-existent app: %v", err)
	}
}

// BenchmarkBackupManager_CreateBackup benchmarks backup creation
func BenchmarkBackupManager_CreateBackup(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "recovery-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bm := &BackupManager{
		backupDir: tmpDir,
	}

	// Create test config file
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := "theme = dark\nfont-size = 14\nother = value"
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bm.CreateBackup(configPath, "test-app")
		if err != nil {
			b.Fatalf("Failed to create backup: %v", err)
		}
	}
}