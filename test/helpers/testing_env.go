package helpers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// SetupTestEnv configures a deterministic environment for tests:
//
//   - Prepends the repository-local `testdata/bin` directory to `PATH` if it exists.
//     This allows tests to rely on stub binaries (e.g. `testdata/bin/ghostty`) without
//     depending on system-installed executables.
//   - Creates an isolated temporary `HOME` directory and sets the `HOME` env var to it,
//     so tests that read/write under the user's home do not interfere with the real
//     user environment.
//
// It registers a cleanup with `t.Cleanup` that restores the original `PATH` and `HOME`
// values and removes the temporary HOME directory.
func SetupTestEnv(t *testing.T) {
	t.Helper()

	// Save originals to restore later.
	origPATH := os.Getenv("PATH")
	origHOME, hadHOME := os.LookupEnv("HOME")

	// Try to find the repository root (where go.mod lives).
	repoRoot, err := findRepoRoot()
	if err != nil {
		// Non-fatal: tests can still proceed; log for visibility.
		t.Logf("helpers.SetupTestEnv: unable to locate repo root: %v", err)
	}

	// If repo root found, prepend testdata/bin to PATH (if that dir exists).
	var addedPath string
	if repoRoot != "" {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			addedPath = testBin
			newPATH := addedPath + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				t.Logf("helpers.SetupTestEnv: failed to set PATH: %v", err)
			}
		}
	}

	// Create a temporary HOME directory for the test.
	tmpHome, err := os.MkdirTemp("", "zeroui-test-home-")
	if err != nil {
		t.Logf("helpers.SetupTestEnv: failed to create temp HOME dir: %v", err)
	} else {
		if err := os.Setenv("HOME", tmpHome); err != nil {
			t.Logf("helpers.SetupTestEnv: failed to set HOME: %v", err)
		}
	}

	// Register cleanup to restore environment and remove tmpHome.
	t.Cleanup(func() {
		// Restore PATH
		if addedPath != "" {
			// If original PATH was empty, Unset; otherwise reset to original value.
			if err := os.Setenv("PATH", origPATH); err != nil {
				t.Logf("helpers.SetupTestEnv cleanup: failed to restore PATH: %v", err)
			}
		}

		// Restore HOME
		if hadHOME {
			if err := os.Setenv("HOME", origHOME); err != nil {
				t.Logf("helpers.SetupTestEnv cleanup: failed to restore HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("HOME"); err != nil {
				t.Logf("helpers.SetupTestEnv cleanup: failed to unset HOME: %v", err)
			}
		}

		// Remove temporary HOME
		if tmpHome != "" {
			if err := os.RemoveAll(tmpHome); err != nil {
				t.Logf("helpers.SetupTestEnv cleanup: failed to remove tmp HOME %s: %v", tmpHome, err)
			}
		}
	})
}

// SetupTestEnvWithHome behaves like SetupTestEnv but uses the provided `homeDir`
// instead of creating a temporary directory. The caller is responsible for cleanup
// of that directory if needed.
func SetupTestEnvWithHome(t *testing.T, homeDir string) {
	t.Helper()

	origPATH := os.Getenv("PATH")
	origHOME, hadHOME := os.LookupEnv("HOME")

	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Logf("helpers.SetupTestEnvWithHome: unable to locate repo root: %v", err)
	}

	var addedPath string
	if repoRoot != "" {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			addedPath = testBin
			newPATH := addedPath + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				t.Logf("helpers.SetupTestEnvWithHome: failed to set PATH: %v", err)
			}
		}
	}

	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Logf("helpers.SetupTestEnvWithHome: failed to set HOME: %v", err)
	}

	t.Cleanup(func() {
		if addedPath != "" {
			if err := os.Setenv("PATH", origPATH); err != nil {
				t.Logf("helpers.SetupTestEnvWithHome cleanup: failed to restore PATH: %v", err)
			}
		}
		if hadHOME {
			if err := os.Setenv("HOME", origHOME); err != nil {
				t.Logf("helpers.SetupTestEnvWithHome cleanup: failed to restore HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("HOME"); err != nil {
				t.Logf("helpers.SetupTestEnvWithHome cleanup: failed to unset HOME: %v", err)
			}
		}
	})
}

// RunTestMainWithCleanup provides a shared TestMain implementation for packages
// that need deterministic PATH and HOME setup for tests. This eliminates code
// duplication across multiple TestMain implementations.
//
// Parameters:
//   - packageName: Name of the package (for logging)
//   - tempDirPrefix: Prefix for temporary directory name
//   - clearCache: Function to clear any package-specific caches (optional)
func RunTestMainWithCleanup(m *testing.M, packageName string, tempDirPrefix string, clearCache func()) {
	origPATH := os.Getenv("PATH")
	origHOME, hadHOME := os.LookupEnv("HOME")

	// Attempt to locate repository root by walking up from current working dir.
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Printf("%s: TestMain: unable to locate repo root: %v", packageName, err)
	} else {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			newPATH := testBin + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				log.Printf("%s: TestMain: failed to set PATH: %v", packageName, err)
			} else {
				log.Printf("%s: TestMain: prepended %s to PATH", packageName, testBin)
			}
		}
	}

	// Create temporary HOME for tests to avoid interfering with developer environment.
	var tmpHome string
	if tempDirPrefix != "" {
		tmpHome, err = os.MkdirTemp("", tempDirPrefix)
		if err != nil {
			log.Printf("%s: TestMain: failed to create temp HOME dir: %v", packageName, err)
		} else {
			if err := os.Setenv("HOME", tmpHome); err != nil {
				log.Printf("%s: TestMain: failed to set HOME: %v", packageName, err)
			} else {
				// Clear any caches that might be affected by HOME change
				if clearCache != nil {
					clearCache()
				}
				log.Printf("%s: TestMain: set HOME=%s", packageName, tmpHome)
			}
		}
	}

	// Ensure environment is restored and temp HOME removed after tests.
	defer func() {
		if err := os.Setenv("PATH", origPATH); err != nil {
			log.Printf("%s: TestMain: failed to restore PATH: %v", packageName, err)
		}
		if hadHOME {
			if err := os.Setenv("HOME", origHOME); err != nil {
				log.Printf("%s: TestMain: failed to restore HOME: %v", packageName, err)
			}
		} else {
			if err := os.Unsetenv("HOME"); err != nil {
				log.Printf("%s: TestMain: failed to unset HOME: %v", packageName, err)
			}
		}
		if tmpHome != "" {
			if err := os.RemoveAll(tmpHome); err != nil {
				log.Printf("%s: TestMain: failed to remove tmp HOME %s: %v", packageName, tmpHome, err)
			}
		}
	}()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// findRepoRoot locates the repository root directory by walking up from this file's
// location until it finds a directory containing "go.mod". Returns the absolute path
// to that directory or an error if not found.
func findRepoRoot() (string, error) {
	// Determine path of this source file at runtime.
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}

	dir := filepath.Dir(currentFile)
	// Walk upwards looking for go.mod
	for i := 0; i < 40; i++ { // limit to avoid infinite loops
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
		parent := filepath.Dir(dir)
		// If we've reached filesystem root, stop.
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("repository root (go.mod) not found from %s", currentFile)
}
