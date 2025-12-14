package appconfig

import (
	"testing"

	"github.com/mrtkrcm/ZeroUI/internal/performance"
	"github.com/mrtkrcm/ZeroUI/test/helpers"
)

// TestMain configures a deterministic environment for the package tests.
// It prepends the repository-local testdata/bin (if present) to PATH so
// stub binaries (e.g. testdata/bin/ghostty) are preferred, and it sets HOME
// to an isolated temporary directory for the duration of the tests.
//
// The original PATH and HOME are restored after the test run.
func TestMain(m *testing.M) {
	helpers.RunTestMainWithCleanup(m, "internal/appconfig", "zeroui-internal-config-test-home-", performance.ClearHomeDirCache)
}
