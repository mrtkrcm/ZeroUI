package helpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConfig holds common test configuration
type TestConfig struct {
	PackageName string
	TempDir     string
	Cleanup     func()
}

// SetupTestEnvironment creates a consistent test environment for any package
func SetupTestEnvironment(t *testing.T, packageName string) *TestConfig {
	t.Helper()

	config := &TestConfig{PackageName: packageName}

	// Setup PATH with test binaries
	setupTestPath(t, config)

	// Setup temporary HOME directory
	setupTestHome(t, config)

	// Add cleanup
	t.Cleanup(func() {
		if config.Cleanup != nil {
			config.Cleanup()
		}
	})

	return config
}

// setupTestPath adds testdata/bin to PATH
func setupTestPath(t *testing.T, config *TestConfig) {
	t.Helper()

	origPATH := os.Getenv("PATH")
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Logf("Unable to locate repo root: %v", err)
		return
	}

	testBin := filepath.Join(repoRoot, "testdata", "bin")
	if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
		newPATH := testBin + string(os.PathListSeparator) + origPATH
		if err := os.Setenv("PATH", newPATH); err != nil {
			t.Logf("Failed to set PATH: %v", err)
		} else {
			t.Logf("Prepended %s to PATH", testBin)
		}
	}

	// Restore PATH in cleanup
	config.Cleanup = func() {
		os.Setenv("PATH", origPATH)
	}
}

// setupTestHome creates a temporary HOME directory
func setupTestHome(t *testing.T, config *TestConfig) {
	t.Helper()

	origHOME, hadHOME := os.LookupEnv("HOME")
	tempDir := t.TempDir()

	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Logf("Failed to set HOME: %v", err)
		return
	}

	config.TempDir = tempDir
	t.Logf("Set HOME=%s", tempDir)

	// Add HOME restoration to cleanup
	originalCleanup := config.Cleanup
	config.Cleanup = func() {
		if originalCleanup != nil {
			originalCleanup()
		}
		if hadHOME {
			os.Setenv("HOME", origHOME)
		} else {
			os.Unsetenv("HOME")
		}
	}
}

// AssertNoPanic runs a function and asserts it doesn't panic
func AssertNoPanic(t *testing.T, fn func(), msg string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: function panicked: %v", msg, r)
		}
	}()

	fn()
}

// AssertDuration runs a function and asserts it completes within a time limit
func AssertDuration(t *testing.T, fn func(), maxDuration time.Duration, msg string) {
	t.Helper()

	start := time.Now()
	fn()
	duration := time.Since(start)

	if duration > maxDuration {
		t.Errorf("%s: took %v, expected <= %v", msg, duration, maxDuration)
	}
}

// WaitForCondition waits for a condition to become true with a timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("%s: condition not met within %v", msg, timeout)
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile.Name()
}

// ReadFileSafe reads a file safely, returning empty string if it doesn't exist
func ReadFileSafe(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Logf("File %s does not exist or cannot be read: %v", path, err)
		return ""
	}

	return string(content)
}

// CleanupTempFiles removes temporary files created during tests
func CleanupTempFiles(t *testing.T, paths ...string) {
	t.Helper()

	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to clean up temp file %s: %v", path, err)
		}
	}
}
