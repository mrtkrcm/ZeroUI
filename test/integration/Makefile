#!/usr/bin/env make
# Integration test Makefile for ZeroUI

.PHONY: test-integration test-integration-short test-integration-core test-integration-plugin test-integration-engine clean-test help

# Default target
test-integration: test-integration-core

# Core functionality integration tests (fast)
test-integration-core:
	@echo "🧪 Running core integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestCoreIntegration -timeout=5m

# Plugin system integration tests
test-integration-plugin:
	@echo "🔌 Running plugin integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestPluginSystemIntegration -timeout=10m

# Engine integration tests
test-integration-engine:
	@echo "⚙️ Running engine integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestEngineIntegration -timeout=5m

# All integration tests (comprehensive)
test-integration-all:
	@echo "🚀 Running all integration tests..."
	go test -v -tags=integration ./tests/integration/ -timeout=15m

# Short integration tests (essential only)
test-integration-short:
	@echo "⚡ Running essential integration tests..."
	go test -v -tags=integration -short ./tests/integration/ -timeout=3m

# TUI integration tests (interactive components)
test-integration-tui:
	@echo "🖥️ Running TUI integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestTUIIntegration -timeout=10m

# Configuration file operation tests
test-integration-config:
	@echo "📝 Running configuration integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestConfigFileOperations -timeout=5m

# Plugin API tests
test-integration-plugin-api:
	@echo "🔗 Running plugin API integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestPluginAPIIntegration -timeout=5m

# RPC protocol tests
test-integration-rpc:
	@echo "🌐 Running RPC protocol integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestRPCProtocolIntegration -timeout=10m

# Error recovery tests
test-integration-recovery:
	@echo "🛡️ Running error recovery integration tests..."
	go test -v -tags=integration ./tests/integration/ -run TestEngineErrorRecovery -timeout=5m

# Clean test artifacts
clean-test:
	@echo "🧹 Cleaning test artifacts..."
	@rm -rf /tmp/zeroui-integration-test-*
	@rm -rf /tmp/integration-test-session-*.json
	@go clean -testcache

# Test with coverage
test-integration-coverage:
	@echo "📊 Running integration tests with coverage..."
	@mkdir -p coverage
	go test -v -tags=integration -coverprofile=coverage/integration.out ./tests/integration/ -timeout=15m
	go tool cover -html=coverage/integration.out -o coverage/integration.html
	@echo "Coverage report: coverage/integration.html"

# Parallel test execution (faster)
test-integration-parallel:
	@echo "⚡ Running integration tests in parallel..."
	go test -v -tags=integration -parallel=4 ./tests/integration/ -timeout=10m

# Test specific functionality
test-cli:
	go test -v -tags=integration ./tests/integration/ -run TestCoreIntegration/CLI -timeout=3m

test-engine:
	go test -v -tags=integration ./tests/integration/ -run TestEngineIntegration -timeout=5m

test-plugins:
	go test -v -tags=integration ./tests/integration/ -run TestPlugin -timeout=10m

# Development helpers
test-watch:
	@echo "👀 Watching for changes and running tests..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" -not -path "./vendor/*" | entr -r make test-integration-short; \
	else \
		echo "⚠️ entr not found. Install it for file watching."; \
		make test-integration-short; \
	fi

# Verbose test output
test-integration-verbose:
	@echo "📝 Running integration tests with verbose output..."
	go test -v -tags=integration ./tests/integration/ -timeout=15m | tee test-output.log

# Test environment verification
test-env:
	@echo "🔍 Verifying test environment..."
	@echo "Go version: $(shell go version)"
	@echo "Build targets available:"
	@ls -la build/ 2>/dev/null || echo "No build artifacts found"
	@echo "Plugin directory:"
	@ls -la plugins/ 2>/dev/null || echo "No plugins directory found"

# Benchmark integration tests
benchmark-integration:
	@echo "📊 Running integration benchmarks..."
	go test -v -tags=integration -bench=. -benchmem ./tests/integration/ -timeout=30m

# Help target
help:
	@echo "ZeroUI Integration Test Makefile"
	@echo ""
	@echo "Core targets:"
	@echo "  test-integration         Run core integration tests (default)"
	@echo "  test-integration-all     Run all integration tests"
	@echo "  test-integration-short   Run essential tests only"
	@echo ""
	@echo "Specific test suites:"
	@echo "  test-integration-core    Core functionality tests"
	@echo "  test-integration-plugin  Plugin system tests"
	@echo "  test-integration-engine  Engine operation tests"
	@echo "  test-integration-tui     TUI interface tests"
	@echo "  test-integration-config  Configuration file tests"
	@echo ""
	@echo "Advanced targets:"
	@echo "  test-integration-coverage    Run with coverage reporting"
	@echo "  test-integration-parallel    Run tests in parallel"
	@echo "  test-integration-verbose     Run with verbose output"
	@echo "  benchmark-integration        Run performance benchmarks"
	@echo ""
	@echo "Development helpers:"
	@echo "  test-watch              Watch files and run tests"
	@echo "  test-env                Verify test environment"
	@echo "  clean-test              Clean test artifacts"
	@echo ""
	@echo "Quick tests:"
	@echo "  test-cli                Test CLI commands only"
	@echo "  test-engine             Test engine operations only"
	@echo "  test-plugins            Test plugin system only"