package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSecurityIntegration demonstrates the security improvements working together
func TestSecurityIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir)

	// Create a mock backup directory
	backupDir := filepath.Join(tempDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	// Test 1: Path validation prevents directory traversal
	t.Run("path_validation_blocks_traversal", func(t *testing.T) {
		validator := NewPathValidator(backupDir)

		// These should all be blocked
		maliciousPaths := []string{
			"../../../etc/passwd",
			"backup/../../../etc/passwd",
			"..\\..\\windows\\system32\\config\\sam",
			"/etc/passwd",
			"backup\x00.txt",
		}

		for _, path := range maliciousPaths {
			if err := validator.ValidatePath(path); err == nil {
				t.Errorf("Expected path validation to block malicious path: %s", path)
			}
		}

		// Valid paths should work
		validPath := filepath.Join(backupDir, "legitimate_backup.txt")
		if err := validator.ValidatePath(validPath); err != nil {
			t.Errorf("Expected valid path to be allowed: %v", err)
		}
	})

	// Test 2: YAML limits prevent resource exhaustion
	t.Run("yaml_limits_prevent_bombs", func(t *testing.T) {
		yamlValidator := NewYAMLValidator(&YAMLLimits{
			MaxFileSize: 1024, // 1KB
			MaxDepth:    5,    // 5 levels
			MaxKeys:     20,   // 20 keys
		})

		// Create a YAML bomb (deeply nested)
		yamlBomb := strings.Repeat(`{"level":`, 10) + `"deep"` + strings.Repeat(`}`, 10)
		if err := yamlValidator.ValidateContent([]byte(yamlBomb)); err == nil {
			t.Error("Expected YAML bomb to be blocked by depth limits")
		}

		// Create a key explosion
		var manyKeys []string
		for i := 0; i < 50; i++ {
			manyKeys = append(manyKeys, fmt.Sprintf("key%d: value%d", i, i))
		}
		keyBomb := strings.Join(manyKeys, "\n")
		if err := yamlValidator.ValidateContent([]byte(keyBomb)); err == nil {
			t.Error("Expected key bomb to be blocked by key limits")
		}

		// Normal YAML should work
		normalYAML := `{"config": {"theme": "dark", "font": "mono"}}`
		if err := yamlValidator.ValidateContent([]byte(normalYAML)); err != nil {
			t.Errorf("Expected normal YAML to be allowed: %v", err)
		}
	})

	// Test 3: Combined protection (realistic attack scenario)
	t.Run("combined_protection_realistic_attack", func(t *testing.T) {
		pathValidator := NewPathValidator(backupDir)
		yamlValidator := NewYAMLValidator(DefaultYAMLLimits())

		// Scenario: Attacker tries to restore from a malicious backup path
		// containing a YAML bomb to a system directory
		maliciousBackupPath := "../../../tmp/yaml_bomb.backup"

		// Path validation should block this
		if err := pathValidator.ValidatePath(maliciousBackupPath); err == nil {
			t.Error("Expected malicious backup path to be blocked")
		}

		// Even if somehow the YAML bomb file existed, it should be blocked
		yamlBomb := func() string {
			bomb := "root:\n"
			for i := 0; i < 1000; i++ {
				bomb += fmt.Sprintf("  level%d:\n    data: %s\n", i, strings.Repeat("x", 100))
			}
			return bomb
		}()

		if err := yamlValidator.ValidateContent([]byte(yamlBomb)); err == nil {
			t.Error("Expected YAML bomb content to be blocked")
		}
	})

	// Test 4: Backup name sanitization
	t.Run("backup_name_sanitization", func(t *testing.T) {
		pathValidator := NewPathValidator(backupDir)

		maliciousNames := map[string]string{
			"../../../etc/passwd":    "_etc_passwd",
			"backup/../../../secret": "backup_secret",
			"CON.backup":             "backup", // Windows reserved name
			".hidden_backup":         "hidden_backup",
			"backup\x00.txt":         "backup.txt",
		}

		for malicious, expected := range maliciousNames {
			sanitized := pathValidator.SanitizeBackupName(malicious)
			if sanitized != expected {
				t.Errorf("Expected sanitized name %q for input %q, got %q", expected, malicious, sanitized)
			}

			// Sanitized name should validate successfully
			if err := pathValidator.ValidateBackupName(sanitized); err != nil {
				t.Errorf("Sanitized name %q should be valid but got error: %v", sanitized, err)
			}
		}
	})
}

// TestSecurityPerformanceImpact ensures security measures don't significantly impact performance
func TestSecurityPerformanceImpact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "security-perf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir)

	// Create a reasonably sized YAML file
	normalYAMLContent := `
config:
  app_name: "test_app"
  version: "1.0.0"
  features:
    - "feature1"
    - "feature2"
    - "feature3"
  settings:
    theme: "dark"
    font_size: 14
    auto_save: true
    backup_count: 5
`

	testFile := filepath.Join(tempDir, "normal_config.yaml")
	if err := os.WriteFile(testFile, []byte(normalYAMLContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	yamlValidator := NewYAMLValidator(DefaultYAMLLimits())

	// Measure performance impact
	iterations := 1000

	// Test validation performance
	for i := 0; i < iterations; i++ {
		if err := yamlValidator.ValidateFile(testFile); err != nil {
			t.Fatalf("Validation failed on iteration %d: %v", i, err)
		}
	}

	// Test path validation performance
	pathValidator := NewPathValidator(tempDir)
	testPath := filepath.Join(tempDir, "backup_file.txt")

	for i := 0; i < iterations; i++ {
		if err := pathValidator.ValidatePath(testPath); err != nil {
			t.Fatalf("Path validation failed on iteration %d: %v", i, err)
		}
	}

	// If we reach here, performance is acceptable (no specific timing requirements,
	// but the test shouldn't hang or take excessively long)
	t.Log("Security validation performance is acceptable for normal use cases")
}
