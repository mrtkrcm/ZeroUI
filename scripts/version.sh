#!/bin/bash

# ZeroUI Version Management Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get current version
get_current_version() {
    git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev"
}

# Show current version
current() {
    echo -e "${BLUE}Current version:${NC} $(get_current_version)"
}

# Create release
release() {
    local message="$1"
    
    if [ -z "$message" ]; then
        echo -e "${RED}Error: Release message required${NC}"
        echo "Usage: $0 release \"Release message\""
        exit 1
    fi
    
    local version=$(get_current_version)
    
    echo -e "${BLUE}Creating release $version${NC}"
    echo -e "${YELLOW}Message: $message${NC}"
    
    read -p "Proceed? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -a "$version" -m "$message"
        git push origin "$version"
        echo -e "${GREEN}✅ Release $version created!${NC}"
        echo -e "${YELLOW}📋 GitHub Actions will build and publish automatically${NC}"
    fi
}

# Show usage
usage() {
    echo "ZeroUI Version Management"
    echo ""
    echo "Usage: $0 <command> [message]"
    echo ""
    echo "Commands:"
    echo "  current         Show current version"
    echo "  release <msg>   Create release with message"
    echo "  help           Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 current"
    echo "  $0 release \"Add awesome new features\""
}

# Main
case "${1:-help}" in
    "current") current ;;
    "release") release "$2" ;;
    "help"|*) usage ;;
esac
