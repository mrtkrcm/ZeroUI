# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled, skipping destructive actions."
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
