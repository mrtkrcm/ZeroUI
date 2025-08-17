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

/*
ValidatePath validates that a path is safe and within allowed boundaries.

Behavioral changes:
  - Simple filenames (no separators and not absolute) are accepted even when
    allowedPaths are configured. This lets callers pass backup filenames like
    "backup.txt" which are then resolved against the backup directory by callers.
  - Directory traversal tokens (\"..\") and null bytes are still rejected.
  - For non-filename inputs we resolve to absolute path and validate against
    allowedPaths when configured.
*/
func (pv *PathValidator) ValidatePath(inputPath string) error {
	// Clean the path to resolve . and .. components
	cleanPath := filepath.Clean(inputPath)

	// Reject null bytes up-front
	if strings.Contains(inputPath, "\x00") {
		return fmt.Errorf("null byte detected in path: %s", inputPath)
	}

	// Reject explicit traversal markers anywhere in the cleaned path
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal detected in path: %s", inputPath)
	}

	// If the input is a simple filename (no separators and not absolute),
	// accept it as a valid name. Callers that need a full path should join
	// this filename with their configured directory (e.g. backupDir).
	if !strings.ContainsAny(inputPath, "/\\") && !filepath.IsAbs(inputPath) {
		// Ensure it's not empty or hidden (leading dot)
		trimmed := strings.TrimSpace(inputPath)
		if trimmed == "" || strings.HasPrefix(trimmed, ".") {
			return fmt.Errorf("invalid filename: %s", inputPath)
		}
		return nil
	}

	// Convert to absolute path for proper validation for non-filename inputs
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate against allowed paths if configured
	if len(pv.allowedPaths) > 0 {
		var allowed bool
		for _, allowedPath := range pv.allowedPaths {
			// normalize allowed path as absolute as well
			absAllowed, aerr := filepath.Abs(allowedPath)
			if aerr != nil {
				absAllowed = filepath.Clean(allowedPath)
			}
			if strings.HasPrefix(absPath, absAllowed) {
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

// SanitizeBackupName creates a safe backup name from user input.
//
// Rules applied:
//   - Replace backslashes and forward slashes with underscores
//   - Reduce directory traversal markers (\"../\", \"..\\\") to underscores
//   - Remove null bytes
//   - Remove leading dots
//   - Collapse repeated underscores to a single underscore
//   - If the result is empty or contains only underscores fallback to \"backup\"
//   - If the result matches a reserved Windows name, fallback to \"backup\"
func (pv *PathValidator) SanitizeBackupName(input string) string {
	s := input

	// Replace backslashes early
	s = strings.ReplaceAll(s, "\\", "_")

	// Replace traversal markers with underscore (do this before replacing separators)
	s = strings.ReplaceAll(s, "../", "_")
	s = strings.ReplaceAll(s, "..\\", "_")

	// Replace remaining separators
	s = strings.ReplaceAll(s, "/", "_")

	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Remove leading dots (hidden files)
	s = strings.TrimLeft(s, ".")

	// Collapse consecutive underscores into single underscore
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	// If empty or only underscores, return default name
	if s == "" || strings.Trim(s, "_") == "" {
		return "backup"
	}

	// Prevent reserved Windows names; if matched, fallback to 'backup'
	upper := strings.ToUpper(s)
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	for _, r := range reserved {
		if upper == r || strings.HasPrefix(upper, r+".") {
			return "backup"
		}
	}

	// Final safety: ensure no separators remain
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")

	// Ensure we didn't accidentally return an empty string
	if s == "" {
		return "backup"
	}

	return s
}
