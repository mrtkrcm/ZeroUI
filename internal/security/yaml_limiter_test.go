package security

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestYAMLValidator_ValidateFile(t *testing.T) {
	validator := NewYAMLValidator(&YAMLLimits{
		MaxFileSize: 1024, // 1KB limit for testing
	})

	// Create test files
	tempDir, err := os.MkdirTemp("", "yaml-validator-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Small file (should pass)
	smallFile := tempDir + "/small.yaml"
	if err := os.WriteFile(smallFile, []byte("key: value"), 0o644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// Large file (should fail)
	largeContent := strings.Repeat("key: value\n", 200) // > 1KB
	largeFile := tempDir + "/large.yaml"
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "small_file_valid",
			filePath:    smallFile,
			expectError: false,
			description: "Small file should pass validation",
		},
		{
			name:        "large_file_invalid",
			filePath:    largeFile,
			expectError: true,
			description: "Large file should be rejected",
		},
		{
			name:        "nonexistent_file",
			filePath:    tempDir + "/nonexistent.yaml",
			expectError: true,
			description: "Nonexistent file should cause error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFile(tt.filePath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s but got none. %s", tt.filePath, tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s but got: %v. %s", tt.filePath, err, tt.description)
			}
		})
	}
}

func TestYAMLValidator_ValidateContent(t *testing.T) {
	validator := NewYAMLValidator(&YAMLLimits{
		MaxFileSize: 100, // 100 bytes
		MaxDepth:    3,   // 3 levels
		MaxKeys:     5,   // 5 keys max
	})

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "simple_valid_yaml",
			content:     `key1: value1\nkey2: value2`,
			expectError: false,
			description: "Simple YAML should be valid",
		},
		{
			name:        "too_large_content",
			content:     strings.Repeat("key: very long value that exceeds the limit\n", 10),
			expectError: true,
			description: "Content exceeding size limit should be rejected",
		},
		{
			name:        "too_deep_nesting",
			content:     `{"level1": {"level2": {"level3": {"level4": "too deep"}}}}`,
			expectError: true,
			description: "Deeply nested content should be rejected",
		},
		{
			name:        "too_many_keys",
			content:     `key1: 1\nkey2: 2\nkey3: 3\nkey4: 4\nkey5: 5\nkey6: 6\nkey7: 7`,
			expectError: true,
			description: "Content with too many keys should be rejected",
		},
		{
			name:        "valid_nested_content",
			content:     `{"level1": {"level2": {"level3": "ok"}}}`,
			expectError: false,
			description: "Valid nested content should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContent([]byte(tt.content))

			if tt.expectError && err == nil {
				t.Errorf("Expected error for content but got none. %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for content but got: %v. %s", err, tt.description)
			}
		})
	}
}

func TestYAMLValidator_SafeReadFile(t *testing.T) {
	validator := NewYAMLValidator(&YAMLLimits{
		MaxFileSize:  1024, // 1KB
		MaxDepth:     10,
		MaxKeys:      100,
		ParseTimeout: 1 * time.Second, // Short timeout for testing
	})

	tempDir, err := os.MkdirTemp("", "yaml-safe-read-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Valid file
	validFile := tempDir + "/valid.yaml"
	validContent := "key: value\nother: data"
	if err := os.WriteFile(validFile, []byte(validContent), 0o644); err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}

	// Large file
	largeFile := tempDir + "/large.yaml"
	largeContent := strings.Repeat("key: value\n", 200)
	if err := os.WriteFile(largeFile, []byte(largeContent), 0o644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		description string
	}{
		{
			name:        "valid_file_read",
			filePath:    validFile,
			expectError: false,
			description: "Valid file should be read successfully",
		},
		{
			name:        "large_file_rejected",
			filePath:    largeFile,
			expectError: true,
			description: "Large file should be rejected",
		},
		{
			name:        "nonexistent_file",
			filePath:    tempDir + "/nonexistent.yaml",
			expectError: true,
			description: "Nonexistent file should cause error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := validator.SafeReadFile(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none. %s", tt.filePath, tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s but got: %v. %s", tt.filePath, err, tt.description)
				} else if string(data) != validContent {
					t.Errorf("Expected content %q but got %q", validContent, string(data))
				}
			}
		})
	}
}

func TestYAMLValidator_YAMLBombProtection(t *testing.T) {
	validator := NewYAMLValidator(&YAMLLimits{
		MaxFileSize: 10000, // 10KB
		MaxDepth:    5,     // Low depth limit
		MaxKeys:     50,    // Low key limit
	})

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "billion_laughs_protection",
			content:     strings.Repeat(`{"a":`, 10) + "1" + strings.Repeat("}", 10),
			expectError: true,
			description: "Deep nesting should be blocked",
		},
		{
			name: "many_keys_attack",
			content: func() string {
				var keys []string
				for i := 0; i < 100; i++ {
					keys = append(keys, fmt.Sprintf("key%d: value%d", i, i))
				}
				return strings.Join(keys, "\n")
			}(),
			expectError: true,
			description: "Too many keys should be blocked",
		},
		{
			name:        "normal_yaml",
			content:     `{"config": {"theme": "dark", "size": 14}}`,
			expectError: false,
			description: "Normal YAML should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContent([]byte(tt.content))

			if tt.expectError && err == nil {
				t.Errorf("Expected error for YAML bomb but got none. %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for normal YAML but got: %v. %s", err, tt.description)
			}
		})
	}
}

func TestDefaultYAMLLimits(t *testing.T) {
	limits := DefaultYAMLLimits()

	// Verify reasonable defaults
	if limits.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected MaxFileSize to be 10MB, got %d", limits.MaxFileSize)
	}

	if limits.MaxDepth != 50 {
		t.Errorf("Expected MaxDepth to be 50, got %d", limits.MaxDepth)
	}

	if limits.MaxKeys != 10000 {
		t.Errorf("Expected MaxKeys to be 10000, got %d", limits.MaxKeys)
	}

	if limits.ParseTimeout != 30*time.Second {
		t.Errorf("Expected ParseTimeout to be 30s, got %v", limits.ParseTimeout)
	}
}

func TestYAMLValidator_GetLimits(t *testing.T) {
	customLimits := &YAMLLimits{
		MaxFileSize:  2048,
		MaxDepth:     20,
		MaxKeys:      500,
		ParseTimeout: 10 * time.Second,
	}

	validator := NewYAMLValidator(customLimits)
	retrievedLimits := validator.GetLimits()

	if retrievedLimits.MaxFileSize != customLimits.MaxFileSize {
		t.Errorf("Expected MaxFileSize %d, got %d", customLimits.MaxFileSize, retrievedLimits.MaxFileSize)
	}

	if retrievedLimits.MaxDepth != customLimits.MaxDepth {
		t.Errorf("Expected MaxDepth %d, got %d", customLimits.MaxDepth, retrievedLimits.MaxDepth)
	}
}
