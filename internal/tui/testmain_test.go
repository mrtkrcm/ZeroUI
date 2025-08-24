package tui

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/test/helpers"
	"github.com/mrtkrcm/ZeroUI/internal/performance"
)

// TestMain configures a deterministic environment for the package tests:
//   - Prepends repo-local testdata/bin (if present) to PATH so stub binaries
//     such as testdata/bin/ghostty are preferred.
//   - Creates an isolated temporary HOME directory for the test run.
//
// The original environment is restored after tests complete.
func TestMain(m *testing.M) {
	helpers.RunTestMainWithCleanup(m, "internal/tui", "zeroui-internal-tui-test-home-", performance.ClearHomeDirCache)
}


