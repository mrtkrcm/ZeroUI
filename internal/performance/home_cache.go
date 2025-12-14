package performance

import (
	"os"
	"sync"
)

var (
	// Cache only successful home dir lookups. Do not cache errors so callers can retry.
	homeDir       string
	homeDirCached bool
	homeDirMu     sync.RWMutex
)

// GetHomeDir returns the user's home directory.
//
// Behavior:
//   - First consults environment variables ($HOME, then USERPROFILE) to support tests/CI.
//   - Only successful lookups are cached. If os.UserHomeDir() returns an error the error
//     is not cached to allow retries (and to avoid permanently shadowing a later-correct env).
func GetHomeDir() (string, error) {
	// Prefer explicit environment overrides (useful for tests and CI)
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	}

	// Return cached positive result if present
	homeDirMu.RLock()
	if homeDirCached {
		dir := homeDir
		homeDirMu.RUnlock()
		return dir, nil
	}
	homeDirMu.RUnlock()

	// Attempt to detect home directory
	dir, err := os.UserHomeDir()
	if err != nil {
		// Do not cache errors; return so caller can decide how to proceed / retry
		return "", err
	}

	// Cache successful discovery
	homeDirMu.Lock()
	homeDir = dir
	homeDirCached = true
	homeDirMu.Unlock()

	return dir, nil
}

// MustGetHomeDir returns the home directory or panics if there's an error
// Use this only in initialization code where the home directory is required
func MustGetHomeDir() string {
	dir, err := GetHomeDir()
	if err != nil {
		panic("failed to get home directory: " + err.Error())
	}
	return dir
}

// ClearHomeDirCache clears any cached home directory result.
//
// This is intended as a test helper so tests can reset the cached value when
// they manipulate environment variables such as HOME. Tests should call this
// after setting or unsetting HOME to ensure subsequent calls to GetHomeDir()
// reflect the updated environment.
//
// Note: exported so package tests (or other packages) can call it. This should
// only be used in test code.
func ClearHomeDirCache() {
	homeDirMu.Lock()
	homeDir = ""
	homeDirCached = false
	homeDirMu.Unlock()
}
