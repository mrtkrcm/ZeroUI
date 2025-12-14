# ZeroUI - State-of-the-art Go Makefile
.PHONY: help build test lint clean install dev deps security benchmark profile docs check all

# Build information
BINARY_NAME := zeroui
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d " " -f 3)

# Build flags
LDFLAGS := -ldflags "\
	-X 'github.com/mrtkrcm/ZeroUI/internal/version.Version=$(VERSION)' \
	-X 'github.com/mrtkrcm/ZeroUI/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/mrtkrcm/ZeroUI/internal/version.BuildTime=$(BUILD_TIME)' \
	-X 'github.com/mrtkrcm/ZeroUI/internal/version.GoVersion=$(GO_VERSION)' \
	-w -s"

# Directories
BUILD_DIR := build
COVERAGE_DIR := coverage
DOCS_DIR := docs
TOOLS_DIR := tools

# Go environment
export CGO_ENABLED=0
export GOOS ?= $(shell go env GOOS)
export GOARCH ?= $(shell go env GOARCH)

# Colors for pretty output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(CYAN)ZeroUI - State-of-the-art Go Development$(NC)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)Build info:$(NC)"
	@echo "  Version: $(GREEN)$(VERSION)$(NC)"
	@echo "  Commit:  $(GREEN)$(COMMIT)$(NC)"
	@echo "  Go:      $(GREEN)$(GO_VERSION)$(NC)"

## all: Run all checks and build
all: deps lint test security build

## deps: Download and tidy dependencies
deps:
	@echo "$(BLUE)📦 Installing dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@go mod verify

## dev-deps: Install development tools
dev-deps:
	@echo "$(BLUE)🔧 Installing development tools...$(NC)"
	@mkdir -p $(TOOLS_DIR)
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -I {} go install {}
	@echo "$(GREEN)✅ Development tools installed$(NC)"

## air: Start development server with hot reloading
air:
	@echo "$(BLUE)🔥 Starting development server with hot reloading...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)❌ air not found. Installing...$(NC)"; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

## build: Build the binary
build:
	@echo "$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✅ Built $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Build for all platforms
build-all: clean
	@echo "$(BLUE)🔨 Building for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for os in darwin linux windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ] && [ "$$arch" = "arm64" ]; then continue; fi; \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch .; \
			if [ "$$os" = "windows" ]; then \
				mv $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch.exe; \
			fi; \
		done; \
	done
	@echo "$(GREEN)✅ Built all platform binaries$(NC)"

## install: Install the binary to GOPATH/bin
install:
	@echo "$(BLUE)📥 Installing $(BINARY_NAME)...$(NC)"
	@go install $(LDFLAGS) .
	@echo "$(GREEN)✅ Installed $(BINARY_NAME)$(NC)"

## test: Run tests

# Prepare test stubs and ensure test binaries under testdata/bin are executable.
# This target is idempotent and safe to run in CI or locally.
test-setup:
	@echo "$(BLUE)🧪 Preparing test stubs...$(NC)"
	@mkdir -p ./testdata/bin || true
	@chmod +x ./testdata/bin/* 2>/dev/null || true
	@# Ensure git index marks them executable if running in a git workspace (best-effort)
	@git update-index --add --chmod=+x ./testdata/bin/* 2>/dev/null || true

test:
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@$(MAKE) test-setup
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "$(GREEN)✅ Tests completed$(NC)"

## test-fast: Run fast unit tests only (short, no integration/perf)
test-fast:
	@echo "$(BLUE)⚡ Running fast tests (short mode, no integration/perf)...$(NC)"
	@FAST_TUI_TESTS=true go test -short ./...
	@echo "$(GREEN)✅ Fast tests completed$(NC)"

## test-deterministic: Run full test suite in deterministic/fast mode
test-deterministic:
	@echo "$(BLUE)🧪 Running deterministic full test suite...$(NC)"
	@ZEROUI_TEST_MODE=true FAST_TUI_TESTS=true go test ./...
	@echo "$(GREEN)✅ Deterministic full tests completed$(NC)"

## test-update-baselines: Update visual test baselines
test-update-baselines:
	@echo "$(BLUE)🖼️ Updating visual baselines...$(NC)"
	@UPDATE_TUI_BASELINES=true go test ./internal/tui -run TestTUIVisualRegression
	@echo "$(GREEN)✅ Visual baselines updated$(NC)"

## test-clean-snapshots: Clean local snapshot and test artifacts
test-clean-snapshots:
	@echo "$(BLUE)🧹 Cleaning test snapshots and artifacts...$(NC)"
	@rm -rf testdata/screenshots/*
	@rm -rf testdata/snapshots/*.new
	@echo "$(GREEN)✅ Test artifacts cleaned$(NC)"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(BLUE)🧪 Running tests with verbose output...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@gotestsum --format=standard-verbose -- -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@echo "$(GREEN)✅ Verbose tests completed$(NC)"

## test-watch: Run tests in watch mode
test-watch:
	@echo "$(BLUE)👀 Running tests in watch mode...$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r go test ./...; \
	else \
		echo "$(RED)❌ entr not found. Install it for test watching.$(NC)"; \
	fi

## test-integration: Run integration tests
test-integration:
	@echo "$(BLUE)🧪 Running integration tests...$(NC)"
	@go test -v -tags=integration ./test/integration/...
	@echo "$(GREEN)✅ Integration tests completed$(NC)"

## benchmark: Run benchmarks
benchmark:
	@echo "$(BLUE)📊 Running benchmarks...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -bench=. -benchmem -cpuprofile=$(COVERAGE_DIR)/cpu.prof -memprofile=$(COVERAGE_DIR)/mem.prof ./...
	@echo "$(GREEN)✅ Benchmarks completed$(NC)"

## profile: Generate and view CPU profile
profile: benchmark
	@echo "$(BLUE)📊 Opening CPU profile...$(NC)"
	@go tool pprof -http=:8080 $(COVERAGE_DIR)/cpu.prof

## lint: Run linters
lint:
	@echo "$(BLUE)🔍 Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml; \
	else \
		echo "$(RED)❌ golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Linting completed$(NC)"

## fmt: Format code
fmt:
	@echo "$(BLUE)📝 Formatting code...$(NC)"
	@gofmt -s -w .
	@goimports -w .
	@echo "$(GREEN)✅ Code formatted$(NC)"

## security: Run security checks
security:
	@echo "$(BLUE)🔒 Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt sarif -out gosec-report.sarif -stdout ./...; \
	else \
		echo "$(YELLOW)⚠️  gosec not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -deps ./... | nancy sleuth; \
	else \
		echo "$(YELLOW)⚠️  nancy not found. Install it with: go install github.com/sonatypecommunity/nancy@latest$(NC)"; \
	fi
	@echo "$(GREEN)✅ Security checks completed$(NC)"

## vuln: Check for known vulnerabilities
vuln:
	@echo "$(BLUE)🔍 Checking for vulnerabilities...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)⚠️  govulncheck not found. Install it with: go install golang.org/x/vuln/cmd/govulncheck@latest$(NC)"; \
	fi
	@echo "$(GREEN)✅ Vulnerability check completed$(NC)"

## check: Run all quality checks
check: lint test security vuln

## docs: Generate documentation
docs:
	@echo "$(BLUE)📚 Generating documentation...$(NC)"
	@mkdir -p $(DOCS_DIR)
	@go doc -all . > $(DOCS_DIR)/godoc.txt
	@if command -v godoc >/dev/null 2>&1; then \
		echo "$(GREEN)📖 Start godoc server with: godoc -http=:6060$(NC)"; \
	fi
	@echo "$(GREEN)✅ Documentation generated$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)🧹 Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f gosec-report.sarif
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)✅ Cleaned$(NC)"

## version: Show version information
version:
	@echo "$(CYAN)ZeroUI Version Information$(NC)"
	@echo "Version:    $(GREEN)$(VERSION)$(NC)"
	@echo "Commit:     $(GREEN)$(COMMIT)$(NC)"
	@echo "Build Time: $(GREEN)$(BUILD_TIME)$(NC)"
	@echo "Go Version: $(GREEN)$(GO_VERSION)$(NC)"

## run: Run the application in development mode
run: build
	@echo "$(BLUE)🚀 Running $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✅ Docker image built$(NC)"

## docker-run: Run Docker container
docker-run: docker-build
	@echo "$(BLUE)🐳 Running Docker container...$(NC)"
	@docker run --rm -it $(BINARY_NAME):latest

## pre-commit: Run pre-commit checks
pre-commit: fmt lint test security
	@echo "$(GREEN)✅ Pre-commit checks completed$(NC)"

## release: Prepare for release
release: clean all build-all
	@echo "$(GREEN)🎉 Release artifacts prepared$(NC)"
	@ls -la $(BUILD_DIR)/

# Development shortcuts
.PHONY: dev watch
## dev: Start development environment
dev:
	@echo "$(BLUE)🚀 Starting development environment...$(NC)"
	@echo "$(YELLOW)Running with file watching...$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r make run; \
	else \
		echo "$(YELLOW)⚠️  entr not found. Running once...$(NC)"; \
		make run; \
	fi

## watch: Watch for changes and rebuild
watch:
	@echo "$(BLUE)👀 Watching for changes...$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r make build; \
	else \
		echo "$(RED)❌ entr not found. Install it for file watching.$(NC)"; \
	fi
