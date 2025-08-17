package integration

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestMain prepares a deterministic environment for integration tests.
//
// It ensures that any repository-local test stub binaries under `testdata/bin` are
// preferred by prepending that directory to PATH, and it also sets HOME to an
// isolated temporary directory for the duration of the integration test run.
//
// The original PATH and HOME are restored after the tests complete.
func TestMain(m *testing.M) {
	origPATH := os.Getenv("PATH")
	origHOME, hadHOME := os.LookupEnv("HOME")

	// Try to locate the repository root (where go.mod lives).
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Printf("tests/integration: TestMain: unable to locate repo root: %v", err)
	} else {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			newPATH := testBin + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				log.Printf("tests/integration: TestMain: failed to set PATH: %v", err)
			} else {
				log.Printf("tests/integration: TestMain: prepended %s to PATH", testBin)
			}
		}
	}

	// Create a temporary HOME directory for integration tests.
	tmpHome, err := os.MkdirTemp("", "zeroui-integration-test-home-")
	if err != nil {
		log.Printf("tests/integration: TestMain: failed to create temp HOME dir: %v", err)
	} else {
		if err := os.Setenv("HOME", tmpHome); err != nil {
			log.Printf("tests/integration: TestMain: failed to set HOME: %v", err)
		} else {
			log.Printf("tests/integration: TestMain: set HOME=%s", tmpHome)
		}
	}

	// Run tests
	code := m.Run()

	// Restore environment and cleanup
	if err := os.Setenv("PATH", origPATH); err != nil {
		log.Printf("tests/integration: TestMain: failed to restore PATH: %v", err)
	}
	if hadHOME {
		if err := os.Setenv("HOME", origHOME); err != nil {
			log.Printf("tests/integration: TestMain: failed to restore HOME: %v", err)
		}
	} else {
		if err := os.Unsetenv("HOME"); err != nil {
			log.Printf("tests/integration: TestMain: failed to unset HOME: %v", err)
		}
	}
	if tmpHome != "" {
		if err := os.RemoveAll(tmpHome); err != nil {
			log.Printf("tests/integration: TestMain: failed to remove tmp HOME %s: %v", tmpHome, err)
		}
	}

	os.Exit(code)
}

// findRepoRoot walks upward from the current working directory to find the
// repository root (the first directory containing go.mod). Returns the absolute
// path to the repository root or an error if not found.
func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 40; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}
