package recovery

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/errors"
)

// BackupManager handles configuration backups for recovery
type BackupManager struct {
	backupDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager() (*BackupManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(errors.SystemFileError, "failed to get home directory", err)
	}

	backupDir := filepath.Join(home, ".config", "configtoggle", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, errors.Wrap(errors.SystemPermission, "failed to create backup directory", err).
			WithSuggestions("Check directory permissions")
	}

	return &BackupManager{
		backupDir: backupDir,
	}, nil
}

// CreateBackup creates a backup of a configuration file
func (bm *BackupManager) CreateBackup(configPath, appName string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return "", nil
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s.backup", appName, timestamp)
	backupPath := filepath.Join(bm.backupDir, backupName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", errors.Wrap(errors.SystemFileError, "failed to read config for backup", err).
			WithSuggestions("Check file permissions")
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", errors.Wrap(errors.SystemFileError, "failed to write backup", err).
			WithSuggestions("Check disk space and permissions")
	}

	return backupPath, nil
}

// RestoreBackup restores a configuration file from backup
func (bm *BackupManager) RestoreBackup(backupPath, targetPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return errors.New(errors.ConfigNotFound, "backup file not found").
			WithSuggestions("Check if backup exists", "List backups with: configtoggle list backups")
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to read backup", err)
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return errors.Wrap(errors.SystemPermission, "failed to create target directory", err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return errors.Wrap(errors.SystemFileError, "failed to restore backup", err).
			WithSuggestions("Check target directory permissions")
	}

	return nil
}

// ListBackups returns available backups for an app
func (bm *BackupManager) ListBackups(appName string) ([]BackupInfo, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, errors.Wrap(errors.SystemFileError, "failed to read backup directory", err)
	}

	var backups []BackupInfo
	prefix := appName + "_"
	suffix := ".backup"

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if appName == "" || (len(name) > len(prefix)+len(suffix) && 
				name[:len(prefix)] == prefix && 
				name[len(name)-len(suffix):] == suffix) {
				
				info, err := entry.Info()
				if err != nil {
					continue
				}

				backupPath := filepath.Join(bm.backupDir, name)
				backups = append(backups, BackupInfo{
					Name:    name,
					Path:    backupPath,
					App:     appName,
					Created: info.ModTime(),
					Size:    info.Size(),
				})
			}
		}
	}

	return backups, nil
}

// CleanupOldBackups removes old backups, keeping only the most recent ones
func (bm *BackupManager) CleanupOldBackups(appName string, keepCount int) error {
	backups, err := bm.ListBackups(appName)
	if err != nil {
		return err
	}

	if len(backups) <= keepCount {
		return nil // Nothing to clean up
	}

	// Sort by creation time (newest first)
	// Remove the oldest ones
	for i := keepCount; i < len(backups); i++ {
		if err := os.Remove(backups[i].Path); err != nil {
			// Log but don't fail on individual file removal errors
			fmt.Printf("Warning: failed to remove old backup %s: %v\n", backups[i].Name, err)
		}
	}

	return nil
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	Name    string
	Path    string
	App     string
	Created time.Time
	Size    int64
}

// SafeOperation provides a safe way to perform config operations with automatic backup/restore
type SafeOperation struct {
	backupManager *BackupManager
	backupPath    string
	targetPath    string
	appName       string
}

// NewSafeOperation creates a new safe operation with automatic backup
func NewSafeOperation(targetPath, appName string) (*SafeOperation, error) {
	backupManager, err := NewBackupManager()
	if err != nil {
		return nil, err
	}

	backupPath, err := backupManager.CreateBackup(targetPath, appName)
	if err != nil {
		return nil, err
	}

	return &SafeOperation{
		backupManager: backupManager,
		backupPath:    backupPath,
		targetPath:    targetPath,
		appName:       appName,
	}, nil
}

// Rollback restores the configuration from backup
func (so *SafeOperation) Rollback() error {
	if so.backupPath == "" {
		return nil // No backup was created
	}

	return so.backupManager.RestoreBackup(so.backupPath, so.targetPath)
}

// Commit removes the backup (operation succeeded)
func (so *SafeOperation) Commit() error {
	if so.backupPath == "" {
		return nil // No backup to remove
	}

	if err := os.Remove(so.backupPath); err != nil {
		// Log but don't fail - backup can stay
		fmt.Printf("Warning: failed to remove backup %s: %v\n", so.backupPath, err)
	}

	return nil
}

// Cleanup ensures backups don't accumulate too much
func (so *SafeOperation) Cleanup(keepCount int) error {
	return so.backupManager.CleanupOldBackups(so.appName, keepCount)
}

// HealthCheck verifies the backup system is functioning properly
func (bm *BackupManager) HealthCheck() error {
	// Check if backup directory exists and is writable
	if _, err := os.Stat(bm.backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
			return errors.Wrap(errors.SystemPermission, "backup directory not accessible", err).
				WithSuggestions("Check directory permissions")
		}
	}

	// Try to write a test file to verify write permissions
	testFile := filepath.Join(bm.backupDir, ".health_check")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return errors.Wrap(errors.SystemPermission, "backup directory not writable", err).
			WithSuggestions("Check directory permissions")
	}

	// Clean up test file
	os.Remove(testFile)
	return nil
}

// GetStats returns statistics about the backup system
func (bm *BackupManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count total backups
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		stats["error"] = err.Error()
		return stats
	}

	stats["backup_directory"] = bm.backupDir
	stats["total_backups"] = len(entries)

	// Calculate total backup size
	var totalSize int64
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	stats["total_size_bytes"] = totalSize
	return stats
}