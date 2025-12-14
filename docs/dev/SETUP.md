# Development Setup Guide

This guide will help you set up a complete development environment for ZeroUI.

## Prerequisites

### Required
- **Go 1.24+**: [Download Go](https://golang.org/dl/)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on macOS/Linux

### Recommended
- **VS Code**: [Download VS Code](https://code.visualstudio.com/)
- **direnv**: [Install direnv](https://direnv.net/docs/installation.html)
- **Docker**: [Install Docker](https://docs.docker.com/get-docker/)

## Quick Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/mrtkrcm/ZeroUI.git
   cd ZeroUI
   ```

2. **Install development tools**:
   ```bash
   make dev-deps
   ```

3. **Set up environment** (optional but recommended):
   ```bash
   # If you have direnv installed
   direnv allow
   
   # Or manually set environment variables
   export ZEROUI_DEV=true
   export ZEROUI_LOG_LEVEL=debug
   ```

4. **Run tests to verify setup**:
   ```bash
   make test
   ```

5. **Build and run**:
   ```bash
   make run
   ```

## Development Tools

### Core Tools
- **golangci-lint**: Comprehensive linting
- **gosec**: Security analysis
- **govulncheck**: Vulnerability scanning
- **air**: Hot reloading for development
- **gotestsum**: Better test output

### Installation
All tools are automatically installed via:
```bash
make dev-deps
```

## Development Workflow

### 1. Hot Reloading Development
```bash
make air
# or
air
```

### 2. Running Tests
```bash
# Basic tests
make test

# Verbose output
make test-verbose

# Watch mode
make test-watch

# Integration tests
make test-integration

# Benchmarks
make benchmark
```

### 3. Code Quality
```bash
# Linting
make lint

# Security checks
make security

# Vulnerability scan
make vuln

# All quality checks
make check
```

### 4. Building
```bash
# Development build
make build

# All platforms
make build-all

# Install locally
make install
```

## VS Code Configuration

The project includes comprehensive VS Code configuration:

- **Auto-formatting** on save
- **Integrated linting** with golangci-lint
- **Test coverage** visualization
- **Debug configurations** for main app and tests
- **Task definitions** for common operations

### Recommended Extensions
Install the recommended extensions when prompted, or manually install:
- Go (official)
- YAML Support
- TOML Support
- GitLens
- Test Explorer

## Environment Configuration

### direnv (Recommended)
If you have direnv installed, the `.envrc` file will automatically set up your development environment.

### Manual Setup
```bash
export GO111MODULE=on
export CGO_ENABLED=0
export ZEROUI_DEV=true
export ZEROUI_LOG_LEVEL=debug
export ZEROUI_CONFIG_DIR="./configs"
```

## Git Workflow

### Pre-commit Hooks
Install pre-commit hooks for automatic quality checks:
```bash
# Install pre-commit (if not already installed)
pip install pre-commit

# Install hooks
pre-commit install
```

### Commit Convention
We use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

## Dependency Management

### Updating Dependencies
Use the automated script for safe dependency updates:
```bash
./scripts/maintenance/update-dependencies.sh
```

### Manual Updates
```bash
# Update all dependencies
go get -u ./...

# Update specific dependencies
go get -u github.com/spf13/cobra@latest

# Clean up
go mod tidy
go mod verify
```

## Troubleshooting

### Common Issues

1. **Build fails with missing tools**:
   ```bash
   make dev-deps
   ```

2. **Tests fail with race conditions**:
   ```bash
   go test -race ./...
   ```

3. **Linter not found**:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

4. **VS Code Go extension issues**:
   - Restart VS Code
   - Run "Go: Install/Update Tools" from command palette

### Getting Help

- Check existing [GitHub Issues](https://github.com/mrtkrcm/ZeroUI/issues)
- Create a new issue with the `question` label
- Join our community discussions

## Performance Optimization

### Development Builds
```bash
# Fast builds for development
export GOGC=off
export GOCACHE="$(pwd)/.cache/go-build"
```

### Profiling
```bash
# CPU profiling
make profile

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Security Considerations

### Regular Security Checks
```bash
# Vulnerability scanning
make vuln

# Security linting
make security

# Dependency audit
go list -json -deps ./... | nancy sleuth
```

### Secrets Management
- Never commit secrets to the repository
- Use environment variables for sensitive data
- Use `.env.local` for local development secrets (gitignored)

## CI/CD Integration

The project includes comprehensive GitHub Actions workflows:
- **Quality checks**: Linting, formatting, security
- **Testing**: Multiple Go versions and platforms
- **Build**: Cross-platform binary builds
- **Security**: Vulnerability scanning, SARIF reports

All checks must pass before merging to main branch.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following our coding standards
4. Run quality checks (`make check`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.