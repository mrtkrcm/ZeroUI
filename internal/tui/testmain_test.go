package tui

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestMain configures a deterministic environment for the package tests:
//   - Prepends repo-local testdata/bin (if present) to PATH so stub binaries
//     such as testdata/bin/ghostty are preferred.
//   - Creates an isolated temporary HOME directory for the test run.
//
// The original environment is restored after tests complete.
func TestMain(m *testing.M) {
	origPATH := os.Getenv("PATH")
	origHOME, hadHOME := os.LookupEnv("HOME")

	// Attempt to locate repository root by walking up from current working dir.
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Printf("internal/tui: TestMain: unable to locate repo root: %v", err)
	} else {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			newPATH := testBin + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				log.Printf("internal/tui: TestMain: failed to set PATH: %v", err)
			} else {
				log.Printf("internal/tui: TestMain: prepended %s to PATH", testBin)
			}
		}
	}

	// Create temporary HOME for tests to avoid interfering with developer environment.
	tmpHome, err := os.MkdirTemp("", "zeroui-internal-tui-test-home-")
	if err != nil {
		log.Printf("internal/tui: TestMain: failed to create temp HOME dir: %v", err)
	} else {
		if err := os.Setenv("HOME", tmpHome); err != nil {
			log.Printf("internal/tui: TestMain: failed to set HOME: %v", err)
		} else {
			log.Printf("internal/tui: TestMain: set HOME=%s", tmpHome)
		}
	}

	// Ensure environment is restored and temp HOME removed after tests.
	defer func() {
		if err := os.Setenv("PATH", origPATH); err != nil {
			log.Printf("internal/tui: TestMain: failed to restore PATH: %v", err)
		}
		if hadHOME {
			if err := os.Setenv("HOME", origHOME); err != nil {
				log.Printf("internal/tui: TestMain: failed to restore HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("HOME"); err != nil {
				log.Printf("internal/tui: TestMain: failed to unset HOME: %v", err)
			}
		}
		if tmpHome != "" {
			if err := os.RemoveAll(tmpHome); err != nil {
				log.Printf("internal/tui: TestMain: failed to remove tmp HOME %s: %v", tmpHome, err)
			}
		}
	}()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// findRepoRoot walks upward from current working directory to find the directory
// containing go.mod. Returns the absolute path to the repo root or an error.
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
