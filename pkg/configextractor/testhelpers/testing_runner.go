package testhelpers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mrtkrcm/ZeroUI/pkg/configextractor"
)

// TestRunner is an in-memory Runner implementation for tests.
// Tests can register expected command invocations and their responses.
// It is safe for concurrent use.
type TestRunner struct {
	mu       sync.RWMutex
	entries  map[string]runnerEntry
	fallback runnerEntry // optional fallback used when no exact match exists
}

type runnerEntry struct {
	stdout []byte
	stderr []byte
	err    error
	delay  time.Duration
	// optional dynamic handler (takes ctx, command, args) that overrides static stdout/stderr
	handler func(ctx context.Context, command string, args ...string) (stdout []byte, stderr []byte, err error)
}

// New returns an initialized TestRunner.
func New() *TestRunner {
	return &TestRunner{
		entries: make(map[string]runnerEntry),
	}
}

// keyFor builds a simple key from command and args to allow exact-match registration.
// It joins command and args with ASCII 0 delimiter to avoid accidental collisions.
func keyFor(command string, args ...string) string {
	if len(args) == 0 {
		return command
	}
	return command + "\x00" + strings.Join(args, "\x00")
}

// Register registers a static response for a specific command+args.
// Use args=nil or omitted to register for the command with any args (use exact key).
func (tr *TestRunner) Register(command string, stdout, stderr []byte, err error, delay time.Duration, args ...string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.entries[keyFor(command, args...)] = runnerEntry{
		stdout: stdout,
		stderr: stderr,
		err:    err,
		delay:  delay,
	}
}

// RegisterString is a convenience overload to register string output.
func (tr *TestRunner) RegisterString(command, stdout, stderr string, err error, delay time.Duration, args ...string) {
	tr.Register(command, []byte(stdout), []byte(stderr), err, delay, args...)
}

// RegisterHandler registers a handler function for a specific command+args that will be invoked
// when Run is called. The handler may inspect args and ctx and return dynamic stdout/stderr/err.
func (tr *TestRunner) RegisterHandler(command string, handler func(ctx context.Context, command string, args ...string) (stdout []byte, stderr []byte, err error), delay time.Duration, args ...string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.entries[keyFor(command, args...)] = runnerEntry{
		handler: handler,
		delay:   delay,
	}
}

// SetFallback registers a fallback response used when no exact match is found.
func (tr *TestRunner) SetFallback(stdout, stderr []byte, err error, delay time.Duration) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.fallback = runnerEntry{
		stdout: stdout,
		stderr: stderr,
		err:    err,
		delay:  delay,
	}
}

// Clear removes all registered entries and fallback.
func (tr *TestRunner) Clear() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.entries = make(map[string]runnerEntry)
	tr.fallback = runnerEntry{}
}

// findEntry looks for an exact match, then for a match on command only (no args),
// and finally returns the fallback (if any) and a boolean indicating whether a match was found.
func (tr *TestRunner) findEntry(command string, args ...string) (runnerEntry, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	// Try exact match
	if e, ok := tr.entries[keyFor(command, args...)]; ok {
		return e, true
	}
	// Try command-only match (no args)
	if e, ok := tr.entries[keyFor(command)]; ok {
		return e, true
	}
	// Use fallback if set (non-zero)
	if tr.fallback.stdout != nil || tr.fallback.stderr != nil || tr.fallback.err != nil || tr.fallback.handler != nil {
		return tr.fallback, true
	}
	return runnerEntry{}, false
}

// Run implements configextractor.Runner.
func (tr *TestRunner) Run(ctx context.Context, command string, args ...string) (stdout []byte, stderr []byte, err error) {
	entry, found := tr.findEntry(command, args...)
	if !found {
		// No registered entry; return a descriptive error so tests fail explicitly.
		return nil, nil, fmt.Errorf("test runner: no entry registered for command %q (args=%v)", command, args)
	}

	// If configured delay is non-zero, wait but respect context cancellation.
	if entry.delay > 0 {
		select {
		case <-time.After(entry.delay):
			// continue
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}

	// If handler present, call it
	if entry.handler != nil {
		return entry.handler(ctx, command, args...)
	}
	// Otherwise return static data
	return entry.stdout, entry.stderr, entry.err
}

// Ensure TestRunner implements the Runner interface
var _ configextractor.Runner = (*TestRunner)(nil)
