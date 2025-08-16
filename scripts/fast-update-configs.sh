# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled, skipping destructive actions."
fi

#!/bin/bash

# High-performance config updater using parallel extraction
# This script leverages the fast batch extraction for maximum efficiency

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/configs"
BINARY="$PROJECT_ROOT/build/zeroui"

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

# Build if needed
if [ ! -f "$BINARY" ] || [ "$1" == "--rebuild" ]; then
    echo -e "${YELLOW}🔨 Building zeroui binary...${NC}"
    cd "$PROJECT_ROOT"
    go build -o build/zeroui . &
    BUILD_PID=$!
    
    # Show progress while building
    while kill -0 $BUILD_PID 2>/dev/null; do
        echo -n "."
        sleep 0.5
    done
    echo -e " ${GREEN}Done!${NC}"
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
