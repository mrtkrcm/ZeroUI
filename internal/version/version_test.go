package version

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	info := Get()

	// Test that we get some values (they may be defaults)
	assert.NotEmpty(t, info.Version)
	assert.NotEmpty(t, info.Commit)
	assert.NotEmpty(t, info.BuildTime)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)

	// Test runtime values
	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Arch)
}

func TestString(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origBuildTime := BuildTime
	origGoVersion := GoVersion

	// Set test values
	Version = "1.2.3"
	Commit = "abc123"
	BuildTime = "2023-12-01T10:00:00Z"
	GoVersion = "go1.21.0"

	// Restore original values after test
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildTime = origBuildTime
		GoVersion = origGoVersion
	}()

	info := Get()
	str := info.String()

	// Test that string contains expected values
	assert.Contains(t, str, "1.2.3")
	assert.Contains(t, str, "abc123")
	assert.Contains(t, str, "2023-12-01T10:00:00Z")
	assert.Contains(t, str, "go1.21.0")
}

func TestInfoFields(t *testing.T) {
	// Test that the Info struct has the expected fields
	info := Info{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildTime: "2023-01-01T00:00:00Z",
		GoVersion: "go1.20",
		OS:        "linux",
		Arch:      "amd64",
	}

	// Test field access
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "abc123", info.Commit)
	assert.Equal(t, "2023-01-01T00:00:00Z", info.BuildTime)
	assert.Equal(t, "go1.20", info.GoVersion)
	assert.Equal(t, "linux", info.OS)
	assert.Equal(t, "amd64", info.Arch)
}

func TestDefaultValues(t *testing.T) {
	// Test the default values set in the package
	// These should be the build-time defaults
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "unknown", Commit)
	assert.Equal(t, "unknown", BuildTime)
	assert.Equal(t, "unknown", GoVersion)
}
