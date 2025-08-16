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

# Test the ZeroUI app grid

echo "Testing ZeroUI App Grid"
echo "========================"
echo ""
echo "1. Testing with no arguments (should show grid):"
./build/zeroui

echo ""
echo "2. Testing with specific app:"
./build/zeroui ui ghostty

echo ""
echo "3. Testing help:"
./build/zeroui --help
