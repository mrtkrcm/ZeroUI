package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathValidator_ValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "path-validator-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validator := NewPathValidator(tmpDir)

	tests := []struct {
		name        string
		path        string
		expectError bool
		description string
	}{
		// Valid paths
		{
			name:        "valid_relative_path",
			path:        "backup.txt",
			expectError: false,
			description: "Simple filename should be valid",
		},
		{
			name:        "valid_nested_path",
			path:        filepath.Join(tmpDir, "subdir", "backup.txt"),
			expectError: false,
			description: "Path within allowed directory should be valid",
		},

		// Directory traversal attacks
		{
			name:        "simple_traversal",
			path:        "../backup.txt",
			expectError: true,
			description: "Simple directory traversal should be blocked",
		},
		{
			name:        "deep_traversal",
			path:        "../../../etc/passwd",
			expectError: true,
			description: "Deep directory traversal should be blocked",
		},
		{
			name:        "encoded_traversal",
			path:        "..%2Fbackup.txt",
			expectError: true,
			description: "URL-encoded traversal should be blocked",
		},
		{
			name:        "mixed_traversal",
			path:        "backup/../../../etc/passwd",
			expectError: true,
			description: "Mixed legitimate path with traversal should be blocked",
		},
		{
			name:        "windows_traversal",
			path:        "..\\backup.txt",
			expectError: true,
			description: "Windows-style directory traversal should be blocked",
		},

		// Null byte injection
		{
			name:        "null_byte_injection",
			path:        "backup.txt\x00.evil",
			expectError: true,
			description: "Null byte injection should be blocked",
		},
		{
			name:        "null_byte_traversal",
			path:        "../../../etc/passwd\x00.txt",
			expectError: true,
			description: "Null byte with traversal should be blocked",
		},

		// Path outside allowed directories
		{
			name:        "outside_allowed_path",
			path:        "/etc/passwd",
			expectError: true,
			description: "Absolute path outside allowed directories should be blocked",
		},
		{
			name:        "home_directory_access",
			path:        "~/../../etc/passwd",
			expectError: true,
			description: "Home directory traversal should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for path %q but got none. %s", tt.path, tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for path %q but got: %v. %s", tt.path, err, tt.description)
			}
		})
	}
}

func TestPathValidator_ValidateBackupName(t *testing.T) {
	validator := NewPathValidator()

	tests := []struct {
		name        string
		backupName  string
		expectError bool
		description string
	}{
		// Valid backup names
		{
			name:        "valid_backup_name",
			backupName:  "app_20230101_120000.backup",
			expectError: false,
			description: "Standard backup name should be valid",
		},
		{
			name:        "valid_with_dashes",
			backupName:  "my-app_backup.bak",
			expectError: false,
			description: "Backup name with dashes should be valid",
		},

		// Invalid backup names
		{
			name:        "empty_name",
			backupName:  "",
			expectError: true,
			description: "Empty backup name should be rejected",
		},
		{
			name:        "path_separator_slash",
			backupName:  "app/backup.txt",
			expectError: true,
			description: "Backup name with forward slash should be rejected",
		},
		{
			name:        "path_separator_backslash",
			backupName:  "app\\backup.txt",
			expectError: true,
			description: "Backup name with backslash should be rejected",
		},
		{
			name:        "hidden_file",
			backupName:  ".hidden_backup",
			expectError: true,
			description: "Hidden file name should be rejected",
		},

		// Windows reserved names
		{
			name:        "reserved_con",
			backupName:  "CON",
			expectError: true,
			description: "Windows reserved name CON should be rejected",
		},
		{
			name:        "reserved_con_with_extension",
			backupName:  "CON.backup",
			expectError: true,
			description: "Windows reserved name with extension should be rejected",
		},
		{
			name:        "reserved_prn",
			backupName:  "prn.txt",
			expectError: true,
			description: "Windows reserved name (case insensitive) should be rejected",
		},
		{
			name:        "reserved_nul",
			backupName:  "NUL.backup",
			expectError: true,
			description: "Windows reserved NUL should be rejected",
		},

		// Directory traversal in backup names
		{
			name:        "traversal_in_name",
			backupName:  "../backup.txt",
			expectError: true,
			description: "Directory traversal in backup name should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBackupName(tt.backupName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for backup name %q but got none. %s", tt.backupName, tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for backup name %q but got: %v. %s", tt.backupName, err, tt.description)
			}
		})
	}
}

func TestPathValidator_SanitizeBackupName(t *testing.T) {
	validator := NewPathValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean_name",
			input:    "app_backup.txt",
			expected: "app_backup.txt",
		},
		{
			name:     "with_slash",
			input:    "app/backup.txt",
			expected: "app_backup.txt",
		},
		{
			name:     "with_backslash",
			input:    "app\\backup.txt",
			expected: "app_backup.txt",
		},
		{
			name:     "with_traversal",
			input:    "../backup.txt",
			expected: "_backup.txt",
		},
		{
			name:     "with_dots",
			input:    "...backup.txt",
			expected: "backup.txt",
		},
		{
			name:     "null_bytes",
			input:    "backup\x00.txt",
			expected: "backup.txt",
		},
		{
			name:     "empty_after_sanitization",
			input:    "./../",
			expected: "backup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeBackupName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeBackupName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPathValidator_WithMultipleAllowedPaths(t *testing.T) {
	tmpDir1, err := os.MkdirTemp("", "validator-test-1")
	if err != nil {
		t.Fatalf("Failed to create temp dir 1: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "validator-test-2")
	if err != nil {
		t.Fatalf("Failed to create temp dir 2: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir2)

	validator := NewPathValidator(tmpDir1, tmpDir2)

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "allowed_path_1",
			path:        filepath.Join(tmpDir1, "backup.txt"),
			expectError: false,
		},
		{
			name:        "allowed_path_2",
			path:        filepath.Join(tmpDir2, "backup.txt"),
			expectError: false,
		},
		{
			name:        "disallowed_path",
			path:        "/tmp/other/backup.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for path %q but got none", tt.path)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for path %q but got: %v", tt.path, err)
			}
		})
	}
}
