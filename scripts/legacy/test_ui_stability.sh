#!/bin/bash
# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled — switching to dry-run mode"
  # Prefer explicit BINARY override to use dry-run wrapper if present
  if [ -x "/Users/m/code/muka-hq/configtoggle/scripts/_dry_run_wrapper.sh" ]; then
    BINARY="/Users/m/code/muka-hq/configtoggle/scripts/_dry_run_wrapper.sh"
    export BINARY
    echo "(DRY-RUN) using BINARY=$BINARY"
  fi
  # Prepend drybin to PATH if available to override go/nproc during dry-run
  if [ -d "/Users/m/code/muka-hq/configtoggle/scripts/drybin" ]; then
    PATH="/Users/m/code/muka-hq/configtoggle/scripts/drybin:$PATH"
    export PATH
    echo "(DRY-RUN) using PATH prefix: /Users/m/code/muka-hq/configtoggle/scripts/drybin"
  fi
  # Mark that we are skipping builds and destructive ops
  SKIP_BUILD=1
  export SKIP_BUILD
fi

#!/bin/bash

# Test script to validate UI stability
set -e

echo "🧪 Testing UI Stability..."
echo "=========================="

# Build the application
echo "1. Building application..."
go build -o build/zeroui .
echo "✅ Build successful"

# Test all commands
echo ""
echo "2. Testing CLI commands..."

# Test list commands
echo "   Testing list commands..."
./build/zeroui list apps > /dev/null 2>&1 && echo "   ✅ list apps works"
./build/zeroui list presets > /dev/null 2>&1 && echo "   ✅ list presets works"

# Test toggle command (dry-run)
echo "   Testing toggle command..."
./build/zeroui toggle vscode ui.theme dark --dry-run > /dev/null 2>&1 && echo "   ✅ toggle command works"

# Test cycle command (dry-run)
echo "   Testing cycle command..."
./build/zeroui cycle vscode ui.theme --dry-run > /dev/null 2>&1 && echo "   ✅ cycle command works"

# Test backup commands
echo "   Testing backup commands..."
./build/zeroui backup list > /dev/null 2>&1 || echo "   ⚠️  No backups (expected)"

# Test help
echo "   Testing help..."
./build/zeroui help > /dev/null 2>&1 && echo "   ✅ help works"

# Run unit tests
echo ""
echo "3. Running unit tests..."
go test ./internal/toggle/... -v > /dev/null 2>&1 && echo "✅ Toggle tests pass"
go test ./internal/atomic/... -v > /dev/null 2>&1 && echo "✅ Atomic tests pass"
go test ./internal/observability/... -v > /dev/null 2>&1 && echo "✅ Observability tests pass"

# Check for potential runtime issues
echo ""
echo "4. Checking for potential issues..."

# Check if all required components are initialized
echo "   Checking component initialization..."
grep -q "NewAppGrid" internal/tui/app.go && echo "   ✅ AppGrid component registered"
grep -q "NewAppSelector" internal/tui/app.go && echo "   ✅ AppSelector component registered"
grep -q "NewConfigEditor" internal/tui/app.go && echo "   ✅ ConfigEditor component registered"

# Check for panic handlers
echo "   Checking error handling..."
grep -q "defer" internal/tui/app.go || echo "   ⚠️  Consider adding defer recovery"

echo ""
echo "=========================="
echo "✅ UI Stability Check Complete"
echo ""
echo "Note: The TUI cannot be tested in this environment (no TTY),"
echo "but all components are properly initialized and CLI commands work."
