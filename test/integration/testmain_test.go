package integration

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// TestMain prepares a deterministic environment for integration tests.
//
// It only ensures repository-local test stub binaries under `testdata/bin` are
// preferred by prepending that directory to PATH. We avoid modifying HOME here
// to prevent interference with package-level logic that expects a real HOME.
func TestMain(m *testing.M) {
	origPATH := os.Getenv("PATH")

	// Try to locate the repository root (where go.mod lives).
	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Printf("test/integration: TestMain: unable to locate repo root: %v", err)
	} else {
		testBin := filepath.Join(repoRoot, "testdata", "bin")
		if fi, err := os.Stat(testBin); err == nil && fi.IsDir() {
			newPATH := testBin + string(os.PathListSeparator) + origPATH
			if err := os.Setenv("PATH", newPATH); err != nil {
				log.Printf("test/integration: TestMain: failed to set PATH: %v", err)
			} else {
				log.Printf("test/integration: TestMain: prepended %s to PATH", testBin)
			}
		}
	}

	// Run tests
	code := m.Run()

	// Restore PATH
	if err := os.Setenv("PATH", origPATH); err != nil {
		log.Printf("test/integration: TestMain: failed to restore PATH: %v", err)
	}

	os.Exit(code)
}

// findRepoRoot walks upwards to find go.mod
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
