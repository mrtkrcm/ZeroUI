// Package version provides build-time version information
package version

import (
	"fmt"
	"runtime"
)

// Build-time variables set via ldflags
var (
	Version   = "dev"     // Version is the semantic version
	Commit    = "unknown" // Commit is the git commit hash
	BuildTime = "unknown" // BuildTime is the build timestamp
	GoVersion = "unknown" // GoVersion is the Go version used to build
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("ZeroUI %s (commit: %s, built: %s, go: %s, %s/%s)",
		i.Version, i.Commit, i.BuildTime, i.GoVersion, i.OS, i.Arch)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("v%s", i.Version)
}

// Print prints version information
func Print() {
	fmt.Println(Get().String())
}

// PrintShort prints short version information
func PrintShort() {
	fmt.Println(Get().Short())
}
