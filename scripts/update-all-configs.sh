# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled, skipping destructive actions."
fi

#!/bin/bash

# Automated config updater for all supported applications
# This script extracts and updates configuration references for all apps

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/configs"
BINARY="$PROJECT_ROOT/build/zeroui"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# List of applications to update
APPS=(
    "ghostty"
    "zed"
    "alacritty"
    "wezterm"
    "neovim"
    "tmux"
    "starship"
    "git"
    "mise"
)

echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}       Configuration Auto-Updater${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""

# Build the binary if needed
if [ ! -f "$BINARY" ]; then
    echo -e "${YELLOW}Building zeroui binary...${NC}"
    cd "$PROJECT_ROOT"
    go build -o build/zeroui .
fi

# Create configs directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Track statistics
TOTAL_APPS=${#APPS[@]}
UPDATED=0
FAILED=0
NEW_SETTINGS_TOTAL=0

echo -e "${GREEN}Starting configuration extraction for $TOTAL_APPS applications...${NC}"
echo ""

for APP in "${APPS[@]}"; do
    echo -e "${BLUE}──────────────────────────────────────────${NC}"
    echo -e "${YELLOW}Processing: $APP${NC}"
    
    CONFIG_FILE="$CONFIG_DIR/${APP}.yaml"
    BACKUP_FILE="$CONFIG_DIR/${APP}.yaml.bak"
    
    # Backup existing config if it exists
    if [ -f "$CONFIG_FILE" ]; then
        cp "$CONFIG_FILE" "$BACKUP_FILE"
        echo "  📁 Backed up existing config"
    fi
    
    # Try to extract configuration
    if $BINARY extract-config "$APP" --update 2>/dev/null; then
        echo -e "  ${GREEN}✅ Successfully extracted config for $APP${NC}"
        
        # Count new settings if update mode
        if [ -f "$BACKUP_FILE" ]; then
            OLD_COUNT=$(grep -c "^  [a-z]" "$BACKUP_FILE" 2>/dev/null || echo 0)
            NEW_COUNT=$(grep -c "^  [a-z]" "$CONFIG_FILE" 2>/dev/null || echo 0)
            DIFF=$((NEW_COUNT - OLD_COUNT))
            
            if [ $DIFF -gt 0 ]; then
                echo -e "  ${GREEN}📈 Added $DIFF new settings${NC}"
                NEW_SETTINGS_TOTAL=$((NEW_SETTINGS_TOTAL + DIFF))
            elif [ $DIFF -eq 0 ]; then
                echo "  ℹ️  No new settings found"
            fi
        else
            NEW_COUNT=$(grep -c "^  [a-z]" "$CONFIG_FILE" 2>/dev/null || echo 0)
            echo -e "  ${GREEN}🆕 Created new config with $NEW_COUNT settings${NC}"
            NEW_SETTINGS_TOTAL=$((NEW_SETTINGS_TOTAL + NEW_COUNT))
        fi
        
        UPDATED=$((UPDATED + 1))
    else
        echo -e "  ${RED}❌ Failed to extract config for $APP${NC}"
        
        # Restore backup if extraction failed
        if [ -f "$BACKUP_FILE" ]; then
            mv "$BACKUP_FILE" "$CONFIG_FILE"
            echo "  ↩️  Restored previous config"
        fi
        
        FAILED=$((FAILED + 1))
    fi
    
    # Clean up backup
    rm -f "$BACKUP_FILE"
done

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                    Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  Total apps processed: ${TOTAL_APPS}"
echo -e "  ${GREEN}Successfully updated: ${UPDATED}${NC}"
echo -e "  ${RED}Failed: ${FAILED}${NC}"
echo -e "  ${YELLOW}New settings added: ${NEW_SETTINGS_TOTAL}${NC}"
echo ""

# Validate all configs
echo -e "${BLUE}Running validation on all configs...${NC}"
if $BINARY validate-reference --all 2>&1 | grep -q "validated successfully"; then
    echo -e "${GREEN}✅ All configurations validated successfully!${NC}"
else
    echo -e "${YELLOW}⚠️  Some configurations may have validation issues${NC}"
fi

echo ""
echo -e "${GREEN}Update complete!${NC}"

# Show next steps
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Review the updated configs in: $CONFIG_DIR"
echo "  2. Test the UI with: $BINARY ui"
echo "  3. Commit changes if everything looks good"
