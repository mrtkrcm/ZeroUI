// Package configextractor provides a Runner abstraction for executing external
// commands. This decouples command execution from platform specifics and makes
// it easy to inject test doubles in unit tests.
package configextractor

import (
	"bytes"
	"context"
	"os/exec"
)

// Runner is an abstraction over running external commands. Implementations
// should honor the provided context for cancellation/timeout behavior.
type Runner interface {
	// Run executes the given command with arguments under the provided context.
	// It returns the captured stdout and stderr (as byte slices) and any error
	// produced while starting or running the command. The error may be of type
	// *exec.ExitError when the process exited non-zero; callers can inspect
	// stdout/stderr for diagnostics.
	Run(ctx context.Context, command string, args ...string) (stdout []byte, stderr []byte, err error)
}

// OSRunner is the default Runner implementation that uses the OS to execute
// commands via os/exec.
type OSRunner struct{}

// NewOSRunner returns a new OSRunner instance.
func NewOSRunner() *OSRunner { return &OSRunner{} }

// Run executes the command using exec.CommandContext, capturing stdout and
// stderr separately. The provided context controls cancellation and timeout.
func (r *OSRunner) Run(ctx context.Context, command string, args ...string) (stdout []byte, stderr []byte, err error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}
