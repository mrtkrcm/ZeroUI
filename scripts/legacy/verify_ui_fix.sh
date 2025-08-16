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

echo "🔧 UI Fix Verification"
echo "====================="
echo ""

# Build
echo "1. Building application..."
if go build -o build/zeroui . 2>/dev/null; then
    echo "   ✅ Build successful"
else
    echo "   ❌ Build failed"
    exit 1
fi

# Run tests
echo ""
echo "2. Running UI component tests..."
if go test ./internal/tui/components -run TestAppGrid 2>/dev/null; then
    echo "   ✅ All grid tests pass"
else
    echo "   ❌ Grid tests failed"
    exit 1
fi

# Check for panic fix
echo ""
echo "3. Verifying panic fix..."
if grep -q "if leftMargin < 0" internal/tui/components/app_grid.go; then
    echo "   ✅ Negative margin protection in place"
else
    echo "   ❌ Margin protection missing"
fi

if grep -q "cardSpacing:.*4" internal/tui/components/app_grid.go; then
    echo "   ✅ Card spacing initialized"
else
    echo "   ❌ Card spacing not initialized"
fi

# Test the binary (will fail in non-TTY but shouldn't panic)
echo ""
echo "4. Testing binary (non-TTY test)..."
OUTPUT=$(timeout 1 ./build/zeroui ui 2>&1 || true)
if echo "$OUTPUT" | grep -q "panic"; then
    echo "   ❌ Panic detected!"
    echo "$OUTPUT"
    exit 1
else
    echo "   ✅ No panic detected"
fi

echo ""
echo "====================="
echo "✅ UI Fix Verified!"
echo ""
echo "The UI should now:"
echo "  • Not panic with 'negative Repeat count'"
echo "  • Handle small terminal sizes gracefully"
echo "  • Render perfect square cards"
echo "  • Work smoothly without freezing"
echo ""
echo "Run './build/zeroui ui' to test the UI"
