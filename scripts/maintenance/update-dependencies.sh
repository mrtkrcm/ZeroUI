#!/bin/bash
# Modern dependency update script with safety checks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 ZeroUI Dependency Update Script${NC}"
echo "======================================"

# Check if we're in git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}❌ Not in a git repository${NC}"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}⚠️  You have uncommitted changes. Please commit or stash them first.${NC}"
    exit 1
fi

# Create backup branch
BACKUP_BRANCH="dependency-update-backup-$(date +%Y%m%d-%H%M%S)"
echo -e "${BLUE}📝 Creating backup branch: $BACKUP_BRANCH${NC}"
git checkout -b "$BACKUP_BRANCH"
git checkout -

# Function to update dependencies safely
update_deps() {
    echo -e "${BLUE}📦 Updating Go dependencies...${NC}"
    
    # Update indirect dependencies first
    go get -u=patch ./...
    
    # Update direct dependencies
    echo -e "${BLUE}📦 Updating direct dependencies...${NC}"
    go get -u github.com/spf13/cobra@latest
    go get -u github.com/spf13/viper@latest
    go get -u github.com/charmbracelet/bubbletea@latest
    go get -u github.com/charmbracelet/huh@latest
    go get -u github.com/charmbracelet/lipgloss@latest
    go get -u github.com/rs/zerolog@latest
    go get -u github.com/stretchr/testify@latest
    
    # Update OpenTelemetry dependencies
    echo -e "${BLUE}📡 Updating OpenTelemetry dependencies...${NC}"
    go get -u go.opentelemetry.io/otel@latest
    go get -u go.opentelemetry.io/otel/metric@latest
    go get -u go.opentelemetry.io/otel/exporters/prometheus@latest
    go get -u go.opentelemetry.io/otel/sdk/metric@latest
    
    # Update security tools
    echo -e "${BLUE}🔒 Updating security dependencies...${NC}"
    go get -u golang.org/x/vuln@latest
    go get -u golang.org/x/tools@latest
    
    # Clean up
    go mod tidy
    go mod verify
}

# Function to run safety checks
safety_checks() {
    echo -e "${BLUE}🔍 Running safety checks...${NC}"
    
    # Check for vulnerabilities
    if command -v govulncheck >/dev/null 2>&1; then
        echo -e "${BLUE}🛡️  Checking for vulnerabilities...${NC}"
        if ! govulncheck ./...; then
            echo -e "${RED}❌ Vulnerability check failed${NC}"
            return 1
        fi
    fi
    
    # Run tests
    echo -e "${BLUE}🧪 Running tests...${NC}"
    if ! go test -race -timeout=5m ./...; then
        echo -e "${RED}❌ Tests failed${NC}"
        return 1
    fi
    
    # Run linting if available
    if command -v golangci-lint >/dev/null 2>&1; then
        echo -e "${BLUE}🔍 Running linter...${NC}"
        if ! golangci-lint run --timeout=5m; then
            echo -e "${RED}❌ Linting failed${NC}"
            return 1
        fi
    fi
    
    # Check build
    echo -e "${BLUE}🔨 Testing build...${NC}"
    if ! go build -o /tmp/zeroui-test .; then
        echo -e "${RED}❌ Build failed${NC}"
        return 1
    fi
    rm -f /tmp/zeroui-test
    
    return 0
}

# Function to show dependency changes
show_changes() {
    echo -e "${BLUE}📊 Dependency changes:${NC}"
    git diff HEAD~1 go.mod | grep "^[+-]" | grep -v "^[+-][+-][+-]" || true
}

# Main execution
main() {
    # Update dependencies
    if ! update_deps; then
        echo -e "${RED}❌ Dependency update failed${NC}"
        exit 1
    fi
    
    # Run safety checks
    if ! safety_checks; then
        echo -e "${RED}❌ Safety checks failed. Rolling back...${NC}"
        git checkout go.mod go.sum
        echo -e "${YELLOW}⚠️  Rolled back to previous versions${NC}"
        exit 1
    fi
    
    # Show changes
    show_changes
    
    # Commit changes
    if git diff --quiet go.mod go.sum; then
        echo -e "${GREEN}✅ No dependency updates needed${NC}"
    else
        echo -e "${BLUE}💾 Committing dependency updates...${NC}"
        git add go.mod go.sum
        git commit -m "chore(deps): update Go dependencies

- Update all dependencies to latest versions
- Verified with tests and vulnerability scan
- Backup branch: $BACKUP_BRANCH"
        
        echo -e "${GREEN}✅ Dependencies updated successfully!${NC}"
        echo -e "${YELLOW}📝 Backup branch created: $BACKUP_BRANCH${NC}"
    fi
}

# Cleanup function
cleanup() {
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Script failed. Cleaning up...${NC}"
        git checkout go.mod go.sum 2>/dev/null || true
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Run main function
main "$@"