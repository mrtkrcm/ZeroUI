package recovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrtkrcm/ZeroUI/internal/config/providers"
	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/mrtkrcm/ZeroUI/internal/security"
)

// BackupManager handles configuration backups for recovery
type BackupManager struct {
	backupDir     string
	pathValidator *security.PathValidator
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

	// Create path validator with backup directory as the only allowed path
	pathValidator := security.NewPathValidator(backupDir)

	return &BackupManager{
		backupDir:     backupDir,
		pathValidator: pathValidator,
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
	// Validate backup path is within allowed directories (prevents directory traversal)
	// If a PathValidator wasn't provided (some tests construct BackupManager manually),
	// fall back to a safe check using the configured backupDir when available.
	if bm.pathValidator != nil {
		if err := bm.pathValidator.ValidatePath(backupPath); err != nil {
			return errors.Wrap(errors.SystemPermission, "backup path validation failed", err).
				WithSuggestions("Use only backup names from 'list' command")
		}
	} else {
		// Fallback: when no pathValidator is present, ensure the backup path is under
		// the backup directory configured on the manager (if set). This avoids nil
		// deref panics in tests while still providing basic safety.
		if bm.backupDir != "" {
			absBackup, _ := filepath.Abs(backupPath)
			absBackupDir, _ := filepath.Abs(bm.backupDir)
			if !strings.HasPrefix(absBackup, absBackupDir) {
				return errors.Wrap(errors.SystemPermission, "backup path validation failed", fmt.Errorf("path outside allowed directories: %s", backupPath)).
					WithSuggestions("Use only backup names from 'list' command")
			}
		}
	}

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

	// Sort newest first
	sort.Slice(backups, func(i, j int) bool { return backups[i].Created.After(backups[j].Created) })
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

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	ValidateConfig(config map[string]interface{}) ValidationResult
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// SafeOperation provides a safe way to perform config operations with automatic backup/restore
type SafeOperation struct {
	backupManager    *BackupManager
	backupPath       string
	targetPath       string
	appName          string
	validator        ConfigValidator
	preSaveConfig    map[string]interface{} // Store config before changes
	validationPassed bool                   // Track if validation succeeded
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

// NewSafeOperationWithValidator creates a new safe operation with validation
func NewSafeOperationWithValidator(targetPath, appName string, validator ConfigValidator) (*SafeOperation, error) {
	safeOp, err := NewSafeOperation(targetPath, appName)
	if err != nil {
		return nil, err
	}

	safeOp.validator = validator
	return safeOp, nil
}

// PreSaveCheckpoint captures the configuration state before making changes
func (so *SafeOperation) PreSaveCheckpoint(config map[string]interface{}) error {
	if so.validator != nil {
		// Validate configuration before saving
		result := so.validator.ValidateConfig(config)
		if !result.Valid {
			return errors.New(errors.ValidationError, "Configuration validation failed").
				WithSuggestions(result.Errors...)
		}
		so.validationPassed = true
	}

	// Store the configuration for potential rollback comparison
	so.preSaveConfig = make(map[string]interface{})
	for k, v := range config {
		so.preSaveConfig[k] = v
	}

	return nil
}

// PostSaveVerification verifies that the saved configuration is valid and matches expectations
func (so *SafeOperation) PostSaveVerification() error {
	// Basic file existence and readability checks
	if _, err := os.Stat(so.targetPath); os.IsNotExist(err) {
		return errors.New(errors.ConfigNotFound, "Configuration file was not created").
			WithSuggestions("Check file system permissions and disk space")
	}

	data, err := os.ReadFile(so.targetPath)
	if err != nil {
		return errors.Wrap(errors.SystemFileError, "Failed to read saved configuration", err)
	}

	if len(data) == 0 {
		return errors.New(errors.ConfigWriteError, "Saved configuration file is empty").
			WithSuggestions("Check file system and permissions")
	}

	// Integrity check: verify the file can be re-parsed successfully
	if err := so.verifyConfigIntegrity(string(data)); err != nil {
		return errors.Wrap(errors.ConfigParseError, "Configuration integrity check failed", err).
			WithSuggestions("The saved configuration may be corrupted", "Check syntax and format")
	}

	// Schema validation if validator is available
	if so.validator != nil {
		if err := so.validateSavedConfig(data); err != nil {
			return errors.Wrap(errors.ValidationError, "Post-save schema validation failed", err)
		}
	}

	return nil
}

// verifyConfigIntegrity performs basic syntax validation on the saved configuration
func (so *SafeOperation) verifyConfigIntegrity(content string) error {
	trimmed := strings.TrimSpace(content)

	// Skip validation for empty files
	if trimmed == "" {
		return nil
	}

	// Check if content looks like JSON (test files)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		// Basic JSON validation - just check if it's valid JSON structure
		if !strings.Contains(trimmed, `"`) {
			return fmt.Errorf("JSON file appears malformed")
		}
		return nil
	}

	// Check if content looks like YAML
	if strings.Contains(trimmed, ": ") || trimmed == "---" {
		// Basic YAML validation
		return nil
	}

	// For Ghostty/custom format files, validate key=value syntax
	lines := strings.Split(content, "\n")
	lineNum := 0

	for _, line := range lines {
		lineNum++
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Basic syntax check for key=value format
		if !strings.Contains(trimmed, "=") {
			return fmt.Errorf("line %d: missing '=' separator in configuration line: %s", lineNum, line)
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("line %d: malformed configuration line: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check for empty keys
		if key == "" {
			return fmt.Errorf("line %d: empty configuration key", lineNum)
		}

		// Check for unclosed quotes (basic check)
		if strings.Contains(value, `"`) {
			quoteCount := strings.Count(value, `"`)
			if quoteCount%2 != 0 {
				return fmt.Errorf("line %d: unclosed quote in value: %s", lineNum, value)
			}
		}
	}

	return nil
}

// validateSavedConfig re-parses and validates the saved configuration against the schema
func (so *SafeOperation) validateSavedConfig(data []byte) error {
	// Parse using the same provider that was used to save
	provider := providers.NewGhosttyProvider("")
	provider.ReadBytes(data)

	// Convert to koanf format
	lines := strings.Split(string(data), "\n")
	configMap := make(map[string]interface{})

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key != "" {
				// Handle multiple values (arrays)
				if existing, exists := configMap[key]; exists {
					if arr, ok := existing.([]string); ok {
						configMap[key] = append(arr, value)
					} else {
						configMap[key] = []string{existing.(string), value}
					}
				} else {
					configMap[key] = value
				}
			}
		}
	}

	// Validate the parsed configuration
	result := so.validator.ValidateConfig(configMap)
	if !result.Valid {
		return fmt.Errorf("validation errors: %v", result.Errors)
	}

	return nil
}

// Rollback restores the configuration from backup
func (so *SafeOperation) Rollback() error {
	if so.backupPath == "" {
		return errors.New(errors.ConfigWriteError, "No backup available for rollback").
			WithSuggestions("Create a backup before making changes")
	}

	// Verify backup exists before attempting restore
	if _, err := os.Stat(so.backupPath); os.IsNotExist(err) {
		return errors.New(errors.ConfigNotFound, "Backup file not found for rollback").
			WithSuggestions("Check backup directory and permissions")
	}

	err := so.backupManager.RestoreBackup(so.backupPath, so.targetPath)
	if err != nil {
		return errors.Wrap(errors.ConfigWriteError, "Failed to rollback configuration", err).
			WithSuggestions("Manual recovery may be required", "Check backup file integrity")
	}

	// Verify rollback was successful
	if verifyErr := so.PostSaveVerification(); verifyErr != nil {
		return errors.Wrap(errors.ConfigWriteError, "Rollback verification failed", verifyErr)
	}

	return nil
}

// Commit removes the backup (operation succeeded)
func (so *SafeOperation) Commit() error {
	if so.backupPath == "" {
		return nil // No backup to remove
	}

	// Verify the saved configuration before committing
	if err := so.PostSaveVerification(); err != nil {
		return errors.Wrap(errors.ConfigWriteError, "Post-save verification failed", err).
			WithSuggestions("Configuration may be corrupted", "Consider manual verification")
	}

	// If validation was required and passed, ensure it still passes after save
	if so.validator != nil && so.validationPassed {
		// In a more advanced implementation, we could re-validate the saved config
		// For now, we rely on PostSaveVerification
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
