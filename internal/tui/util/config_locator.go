package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindConfigPath attempts to find the configuration file for the given application.
// It searches in common configuration locations.
func FindConfigPath(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	possiblePaths := []string{
		filepath.Join(home, ".config", appName, "config.yml"),
		filepath.Join(home, fmt.Sprintf(".%s", appName), "config"),
		filepath.Join(home, fmt.Sprintf(".%src", appName)),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
