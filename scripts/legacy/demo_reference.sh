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

echo "🚀 ZeroUI Reference System Demo"
echo "======================================"
echo

echo "📋 1. List Available Applications:"
go run . reference list
echo

echo "🔍 2. Scan Ghostty Configuration Reference:"
echo "   (Scanning configuration reference from https://ghostty.org/docs/config/reference)"
go run . reference scan ghostty --include-cli=false | head -20
echo "   ... (truncated - shows first 20 lines of 137+ settings found)"
echo

echo "✅ 3. Validate Configuration Values:"
echo "   Testing font-size validation:"
go run . reference validate ghostty font-size 14
echo
echo "   Testing invalid setting:"
go run . reference validate ghostty nonexistent-setting "value" || echo "   ❌ Validation correctly failed for nonexistent setting"
echo

echo "📊 4. Show Specific Setting Details:"
echo "   Details for font-size setting:"
go run . reference show ghostty font-size
echo

echo "🔧 5. Integration with ZeroUI:"
echo "   The reference system is now integrated into ZeroUI and provides:"
echo "   • Automatic discovery of configuration options"
echo "   • Real-time validation of configuration values" 
echo "   • Intelligent suggestions for similar settings"
echo "   • Support for multiple configuration formats (JSON, TOML, YAML)"
echo "   • CLI and web-based documentation scanning"
echo

echo "🎯 6. Supported Applications:"
echo "   ✓ Ghostty - Terminal emulator (TOML config)"
echo "   ✓ Zed - Code editor (JSON config)" 
echo "   ✓ Mise - Development tool manager (TOML config)"
echo "   • Extensible architecture for adding more applications"
echo

echo "🌟 Demo Complete!"
echo "The ZeroUI reference system is now ready to automatically"
echo "discover, validate, and help manage configuration options across"
echo "multiple applications with intelligent suggestions and validation."
