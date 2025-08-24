#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/dry_run.sh" && dry_run_init

# High-performance config updater using parallel extraction
# This script leverages the fast batch extraction for maximum efficiency

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/configs"
# Respect an externally provided BINARY; otherwise use the default build path
if [ -n "${BINARY:-}" ]; then
  echo "Using externally provided BINARY: $BINARY"
else
  BINARY="$PROJECT_ROOT/build/zeroui"
fi

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        ⚡ Fast Config Extractor v2.0 ⚡                   ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Build if needed (safe)
# If SKIP_BUILD is set (e.g. in DRY_RUN) we will prefer not to build the binary.
if [ -z "${SKIP_BUILD:-}" ] || [ "${SKIP_BUILD}" = "0" ]; then
  if [ ! -f "$BINARY" ] || [ "$1" = "--rebuild" ]; then
    echo -e "${YELLOW}🔨 Building zeroui binary...${NC}"
    cd "$PROJECT_ROOT" || exit 1

    # Prefer to use 'timeout' if available to avoid indefinite builds
    if command -v timeout >/dev/null 2>&1; then
      if timeout 30s go build -o build/zeroui .; then
        echo -e " ${GREEN}Done!${NC}"
      else
        echo -e "${RED}❌ Build failed or timed out${NC}"
        exit 1
      fi
    else
      # Fallback: run build in background and enforce a max wait
      go build -o build/zeroui . &
      BUILD_PID=$!
      MAX_ATTEMPTS=120
      ATTEMPT=0
      # Wait with a bounded loop
      while kill -0 $BUILD_PID 2>/dev/null && [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
        echo -n "."
        sleep 0.5
        ATTEMPT=$((ATTEMPT+1))
      done
      if kill -0 $BUILD_PID 2>/dev/null; then
        echo -e "${RED}❌ Build still running after timeout; killing${NC}"
        kill -9 $BUILD_PID 2>/dev/null || true
        exit 1
      fi
      # Ensure we capture a non-zero build exit
      if ! wait $BUILD_PID; then
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
      fi
      echo -e " ${GREEN}Done!${NC}"
    fi
  fi
else
  echo -e "${YELLOW}🔨 Skipping build due to DRY_RUN/SKIP_BUILD${NC}"
fi

# Determine number of workers
WORKERS=${WORKERS:-$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 8)}
echo -e "${BLUE}🚀 Using $WORKERS parallel workers${NC}"

# Start extraction with timing
START_TIME=$(date +%s%N)

echo -e "${YELLOW}📥 Extracting configurations...${NC}"
echo ""

# Run batch extraction
if $BINARY batch-extract \
    --output-dir "$CONFIG_DIR" \
    --workers "$WORKERS" \
    --update \
    --verbose; then

    END_TIME=$(date +%s%N)
    DURATION=$((($END_TIME - $START_TIME) / 1000000)) # Convert to milliseconds

    echo ""
    echo -e "${GREEN}✨ Extraction complete in ${DURATION}ms${NC}"

    # Quick validation
    echo -e "${BLUE}🔍 Running quick validation...${NC}"

    VALID_COUNT=0
    TOTAL_COUNT=0
    for config in "$CONFIG_DIR"/*.yaml; do
        if [ -f "$config" ]; then
            TOTAL_COUNT=$((TOTAL_COUNT + 1))
            if grep -q "settings:" "$config" 2>/dev/null; then
                VALID_COUNT=$((VALID_COUNT + 1))
            fi
        fi
    done

    echo -e "${GREEN}✅ $VALID_COUNT/$TOTAL_COUNT configs validated${NC}"

    # Performance metrics
    if [ $DURATION -gt 0 ]; then
        CONFIGS_PER_SEC=$(( (TOTAL_COUNT * 1000) / DURATION ))
        echo -e "${CYAN}⚡ Performance: ~$CONFIGS_PER_SEC configs/second${NC}"
    fi

else
    echo -e "${RED}❌ Extraction failed${NC}"
    exit 1
fi

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}🎉 All configurations updated successfully!${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
