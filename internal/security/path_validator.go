package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathValidator provides secure path validation functionality
type PathValidator struct {
	allowedPaths []string
}

// NewPathValidator creates a new path validator with allowed base paths
func NewPathValidator(allowedPaths ...string) *PathValidator {
	var cleanPaths []string
	for _, path := range allowedPaths {
		// Convert to absolute path and clean
		absPath, err := filepath.Abs(path)
		if err != nil {
			// If we can't resolve it, use the cleaned version
			absPath = filepath.Clean(path)
		}
		cleanPaths = append(cleanPaths, absPath)
	}

	return &PathValidator{
		allowedPaths: cleanPaths,
	}
}

// ValidatePath validates that a path is safe and within allowed boundaries
func (pv *PathValidator) ValidatePath(inputPath string) error {
	// Clean the path to resolve . and .. components
	cleanPath := filepath.Clean(inputPath)

	// Convert to absolute path for proper validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal detected in path: %s", inputPath)
	}

	// Check for null bytes (path injection)
	if strings.Contains(inputPath, "\x00") {
		return fmt.Errorf("null byte detected in path: %s", inputPath)
	}

	// Validate against allowed paths
	if len(pv.allowedPaths) > 0 {
		var allowed bool
		for _, allowedPath := range pv.allowedPaths {
			// Check if the absolute path is within an allowed directory
			if strings.HasPrefix(absPath, allowedPath) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("path outside allowed directories: %s", inputPath)
		}
	}

	return nil
}

// ValidateBackupName validates a backup filename for security
func (pv *PathValidator) ValidateBackupName(backupName string) error {
	// Basic filename validation
	if backupName == "" {
		return fmt.Errorf("backup name cannot be empty")
	}

	// Check for directory separators (should not contain path separators)
	if strings.ContainsAny(backupName, "/\\") {
		return fmt.Errorf("backup name contains invalid path separators: %s", backupName)
	}

	// Check for hidden files or special names
	if strings.HasPrefix(backupName, ".") {
		return fmt.Errorf("backup name cannot start with dot: %s", backupName)
	}

	// Check for reserved names on Windows
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(backupName)
	for _, reserved := range reservedNames {
		if upperName == reserved || strings.HasPrefix(upperName, reserved+".") {
			return fmt.Errorf("backup name uses reserved system name: %s", backupName)
		}
	}

	// Validate path after joining with backup directory
	return pv.ValidatePath(backupName)
}

// SanitizeBackupName creates a safe backup name from user input
func (pv *PathValidator) SanitizeBackupName(input string) string {
	// Remove directory separators
	sanitized := strings.ReplaceAll(input, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")

	// Remove dangerous characters
	sanitized = strings.ReplaceAll(sanitized, "..", "_")
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove leading dots
	sanitized = strings.TrimLeft(sanitized, ".")

	// Ensure it's not empty after sanitization
	if sanitized == "" {
		sanitized = "backup"
	}

	return sanitized
}
