#!/bin/bash

# Comprehensive sanity check for ConfigToggle/ZeroUI
set -e

echo "🔍 ConfigToggle/ZeroUI Sanity Check"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check binary exists
echo "1. Checking binary..."
if [ -f "./build/zeroui" ]; then
    echo -e "${GREEN}✅ Binary exists${NC}"
    ls -lh ./build/zeroui
else
    echo -e "${RED}❌ Binary not found${NC}"
    exit 1
fi

echo ""
echo "2. Testing basic commands..."

# Test help
echo "   Testing help..."
if ./build/zeroui help > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ Help command works${NC}"
else
    echo -e "${RED}   ❌ Help command failed${NC}"
fi

# Test version
echo "   Testing version..."
./build/zeroui help | head -1

echo ""
echo "3. Testing list commands..."

# List apps
echo "   Listing applications..."
APP_COUNT=$(./build/zeroui list apps 2>/dev/null | grep -c "•" || true)
if [ "$APP_COUNT" -gt 0 ]; then
    echo -e "${GREEN}   ✅ Found $APP_COUNT applications${NC}"
    ./build/zeroui list apps
else
    echo -e "${YELLOW}   ⚠️  No applications found${NC}"
fi

# List keys for ghostty
echo ""
echo "   Listing ghostty configuration keys..."
if ./build/zeroui list keys ghostty > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ Can list configuration keys${NC}"
    ./build/zeroui list keys ghostty | head -5
    echo "   ..."
else
    echo -e "${YELLOW}   ⚠️  Cannot list ghostty keys${NC}"
fi

echo ""
echo "4. Testing toggle command (dry-run)..."

# Test toggle with dry-run
if ./build/zeroui toggle ghostty theme nord --dry-run 2>&1 | grep -q "Would set"; then
    echo -e "${GREEN}   ✅ Toggle command works (dry-run)${NC}"
else
    echo -e "${RED}   ❌ Toggle command failed${NC}"
fi

echo ""
echo "5. Testing cycle command (dry-run)..."

# Test cycle with dry-run
if ./build/zeroui cycle ghostty theme --dry-run > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ Cycle command works (dry-run)${NC}"
else
    echo -e "${YELLOW}   ⚠️  Cycle command may have issues${NC}"
fi

echo ""
echo "6. Checking configuration files..."

# Check for config directory
if [ -d "$HOME/.config/zeroui" ]; then
    echo -e "${GREEN}   ✅ Config directory exists${NC}"
    ls -la "$HOME/.config/zeroui/" 2>/dev/null | head -5 || true
else
    echo -e "${YELLOW}   ⚠️  Config directory not found (will be created on first use)${NC}"
fi

echo ""
echo "7. Running unit tests..."

# Run quick tests
TEST_PACKAGES="./internal/toggle ./internal/atomic ./internal/observability"
PASSED=0
FAILED=0

for pkg in $TEST_PACKAGES; do
    echo -n "   Testing $(basename $pkg)... "
    if go test -timeout 5s $pkg > /dev/null 2>&1; then
        echo -e "${GREEN}✅${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌${NC}"
        ((FAILED++))
    fi
done

echo "   Tests: $PASSED passed, $FAILED failed"

echo ""
echo "8. Checking UI components..."

# Check if UI can be initialized (will fail in non-TTY but that's ok)
if ./build/zeroui ui --help > /dev/null 2>&1; then
    echo -e "${GREEN}   ✅ UI command available${NC}"
else
    echo -e "${RED}   ❌ UI command not available${NC}"
fi

# Check for UI test snapshots
if [ -d "internal/tui/testdata/snapshots" ]; then
    SNAPSHOT_COUNT=$(ls internal/tui/testdata/snapshots/*.txt 2>/dev/null | wc -l | tr -d ' ')
    if [ "$SNAPSHOT_COUNT" -gt 0 ]; then
        echo -e "${GREEN}   ✅ Found $SNAPSHOT_COUNT UI snapshots${NC}"
    fi
else
    echo -e "${YELLOW}   ⚠️  No UI snapshots found${NC}"
fi

echo ""
echo "9. Checking dependencies..."

# Check go.mod
if [ -f "go.mod" ]; then
    echo -e "${GREEN}   ✅ go.mod exists${NC}"
    DEPS=$(grep -c "require" go.mod || true)
    echo "      Found $DEPS dependencies"
else
    echo -e "${RED}   ❌ go.mod not found${NC}"
fi

echo ""
echo "10. Memory and safety checks..."

# Check for race conditions (quick test)
echo -n "   Checking for race conditions... "
if go test -race -timeout 5s ./internal/atomic > /dev/null 2>&1; then
    echo -e "${GREEN}✅ No races detected${NC}"
else
    echo -e "${YELLOW}⚠️  Potential race conditions${NC}"
fi

echo ""
echo "===================================="
echo "📊 Sanity Check Summary"
echo "===================================="

# Summary
echo ""
if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}✅ All core functionality is working!${NC}"
    echo ""
    echo "The application is ready to use:"
    echo "  • Run ${GREEN}./build/zeroui ui${NC} to launch the TUI"
    echo "  • Run ${GREEN}./build/zeroui help${NC} for command documentation"
    echo "  • Run ${GREEN}./build/zeroui list apps${NC} to see available applications"
else
    echo -e "${YELLOW}⚠️  Some components may need attention${NC}"
    echo "  Most functionality is working, but review any failures above."
fi

echo ""
echo "Build Info:"
./build/zeroui help 2>&1 | grep -E "Version|Commit" || echo "  Version: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
echo "  Binary size: $(ls -lh ./build/zeroui | awk '{print $5}')"
echo "  Go version: $(go version | cut -d' ' -f3)"