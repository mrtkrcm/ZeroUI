.PHONY: help build test clean install release deps

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)
BINARY_NAME := zeroui
DIST_DIR := dist

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build for current platform
	mkdir -p $(DIST_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME) .

build-all: ## Build for all platforms
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

test: ## Run tests
	go test -v -race -cover ./...

test-fast: ## Run fast unit tests only
	go test -v -short ./...

lint: ## Run linter (if available)
	command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed"

fmt: ## Format code
	go fmt ./...

clean: ## Clean build artifacts
	rm -rf $(DIST_DIR)
	rm -rf releases

install: ## Install to GOPATH/bin
	go install -ldflags="$(LDFLAGS)" .

# Monorepo targets
workspace-deps: ## Download all workspace dependencies
	go work sync
	cd raycast-extension && npm install

workspace-build: ## Build all workspace components
	$(MAKE) build
	$(MAKE) build-plugins
	cd raycast-extension && npm run build

workspace-test: ## Test all workspace components
	$(MAKE) test-fast
	cd raycast-extension && npm run lint

workspace-clean: ## Clean all workspace artifacts
	$(MAKE) clean
	rm -rf raycast-extension/node_modules
	rm -rf raycast-extension/build

build-plugins: ## Build all plugins
	cd plugins/ghostty-rpc && go build -ldflags="$(LDFLAGS)" -o zeroui-plugin-ghostty-rpc .

test-plugins: ## Test all plugins
	cd plugins/ghostty-rpc && go test -v ./...

install-deps: ## Install development dependencies
	go mod download
	go work sync
	command -v npm >/dev/null 2>&1 && cd raycast-extension && npm install || echo "npm not available, skipping raycast extension dependencies"

release: build-all ## Create release archives
	mkdir -p releases
	cd $(DIST_DIR) && tar -czf ../releases/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(DIST_DIR) && tar -czf ../releases/$(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(DIST_DIR) && tar -czf ../releases/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(DIST_DIR) && tar -czf ../releases/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(DIST_DIR) && zip ../releases/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

.DEFAULT_GOAL := help
