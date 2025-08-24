#!/bin/bash
# ZeroUI Development Environment Setup

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}ZeroUI Development Environment Setup${NC}"
echo "====================================="
echo ""

# Check Go
echo -e "${BLUE}Checking Go...${NC}"
if command -v go >/dev/null 2>&1; then
    go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
    echo -e "${GREEN}✅ Go $go_version found${NC}"
else
    echo -e "${RED}❌ Go not found${NC}"
    echo "Please install Go 1.24+ from https://golang.org/dl/"
fi

# Check Python
echo -e "${BLUE}Checking Python...${NC}"
if command -v python3 >/dev/null 2>&1; then
    python_version=$(python3 --version | cut -d' ' -f2)
    echo -e "${GREEN}✅ Python $python_version found${NC}"
else
    echo -e "${YELLOW}⚠️  Python 3 not found${NC}"
    echo "Some scripts may not work without Python"
fi

# Download dependencies
echo -e "${BLUE}Downloading Go dependencies...${NC}"
cd "$PROJECT_ROOT"
if go mod download; then
    echo -e "${GREEN}✅ Dependencies downloaded${NC}"
else
    echo -e "${RED}❌ Failed to download dependencies${NC}"
fi

# Test build
echo -e "${BLUE}Testing build...${NC}"
if go build .; then
    echo -e "${GREEN}✅ Build successful${NC}"
    rm -f zeroui
else
    echo -e "${RED}❌ Build failed${NC}"
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Available commands:"
echo "  make help         - Show all available commands"
echo "  make build        - Build the application"
echo "  make test         - Run tests"
echo "  ./scripts/zeroui-scripts.sh help  - Show script commands"
