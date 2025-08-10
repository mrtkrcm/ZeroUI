# ZeroUI

[![CI/CD Pipeline](https://github.com/mrtkrcm/ZeroUI/actions/workflows/ci.yml/badge.svg)](https://github.com/mrtkrcm/ZeroUI/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrtkrcm/ZeroUI)](https://goreportcard.com/report/github.com/mrtkrcm/ZeroUI)
[![Coverage Status](https://codecov.io/gh/mrtkrcm/ZeroUI/branch/main/graph/badge.svg)](https://codecov.io/gh/mrtkrcm/ZeroUI)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mrtkrcm/ZeroUI)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> Zero-configuration UI toolkit manager - The fastest way to manage development tool configurations

ZeroUI is a zero-configuration UI toolkit manager that revolutionizes how developers manage UI configurations, themes, and settings across development tools and applications. Built for speed and simplicity with intuitive CLI and interactive TUI interfaces.

## ✨ Features

### 🎯 Core Functionality
- **Multi-format Support**: JSON, YAML, TOML, and custom configuration formats
- **Interactive TUI**: Beautiful terminal user interface built with Bubble Tea
- **Configuration Presets**: Quick application of predefined configuration sets
- **Safe Operations**: Automatic backups with rollback capabilities
- **Validation**: Built-in configuration validation and type checking

### 🏗️ Enterprise Features
- **Observability**: OpenTelemetry metrics and structured logging
- **Security**: Vulnerability scanning and secure coding practices
- **CI/CD Ready**: Comprehensive GitHub Actions pipeline
- **Docker Support**: Multi-stage builds with security best practices
- **Performance**: Benchmarks and profiling tools
- **Testing**: Comprehensive test suite with mocking

### 🔧 Developer Experience
- **Modern Tooling**: golangci-lint, gosec, govulncheck
- **Code Quality**: 95%+ test coverage, strict linting
- **Documentation**: Comprehensive API docs and examples
- **Cross-platform**: Windows, macOS, Linux support

## 🚀 Quick Start

### Installation

```bash
# Using Go (recommended)
go install github.com/mrtkrcm/ZeroUI@latest

# Using Homebrew (macOS/Linux)
brew install zeroui

# Using Docker
docker run --rm -it zeroui/zeroui:latest

# Download binary from releases
curl -L https://github.com/mrtkrcm/ZeroUI/releases/latest/download/zeroui-$(uname -s)-$(uname -m) -o zeroui
chmod +x zeroui
```

### Basic Usage

```bash
# List available applications
zeroui list apps

# Toggle a configuration value
zeroui toggle ghostty theme dark

# Cycle through available values
zeroui cycle alacritty font

# Apply a preset configuration
zeroui preset vscode minimal

# Launch interactive TUI
zeroui ui

# Launch improved TUI with better patterns
zeroui ui-improved
```

## 📖 Documentation

### Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List apps, presets, or configurable keys | `zeroui list apps` |
| `toggle` | Set a specific configuration value | `zeroui toggle app key value` |
| `cycle` | Cycle to next value in a list | `zeroui cycle app key` |
| `preset` | Apply a preset configuration | `zeroui preset app preset-name` |
| `ui` | Launch interactive TUI | `zeroui ui` |
| `backup` | Manage configuration backups | `zeroui backup list` |
| `ref` | Configuration reference system | `zeroui ref show app` |

### Configuration Structure

ZeroUI uses YAML files to define application configurations:

```yaml
# ~/.config/zeroui/apps/ghostty.yaml
name: ghostty
path: ~/.config/ghostty/config
format: custom
description: "Ghostty terminal emulator configuration"

fields:
  theme:
    type: choice
    values: [light, dark, auto]
    default: auto
    description: "Color theme for the terminal"
  
  font-size:
    type: number
    default: 12
    description: "Font size in pixels"

presets:
  minimal:
    name: "Minimal Setup"
    description: "Clean, minimal configuration"
    values:
      theme: light
      font-size: 14
  
  developer:
    name: "Developer Setup"
    description: "Optimized for development"
    values:
      theme: dark
      font-size: 12

hooks:
  post-toggle: "echo 'Configuration updated'"

env:
  GHOSTTY_CONFIG_PATH: "~/.config/ghostty"
```

## 🏗️ Architecture

ZeroUI follows clean architecture principles with clear separation of concerns:

```
cmd/                    # CLI commands and entry points
├── cycle.go           # Cycle command implementation
├── list.go            # List command with styled output
├── toggle.go          # Toggle command implementation
└── ui_improved.go     # Enhanced TUI implementation

internal/
├── config/            # Configuration management
├── container/         # Dependency injection
├── errors/            # Enhanced error handling
├── logger/            # Structured logging
├── observability/     # Metrics and tracing
├── service/           # Business logic layer
├── tui/              # Terminal user interface
├── toggle/           # Core toggle operations
└── version/          # Build information

tools/                 # Development tools
└── tools.go          # Tool dependencies

.github/workflows/     # CI/CD pipelines
├── ci.yml            # Comprehensive CI/CD
└── ...               # Additional workflows
```

## 🧪 Development

### Prerequisites

- Go 1.21 or later
- Make
- Git

### Setup

```bash
# Clone the repository
git clone https://github.com/mrtkrcm/ZeroUI.git
cd ZeroUI

# Install dependencies
make deps

# Install development tools
make dev-deps

# Run all checks
make check
```

### Available Make Targets

```bash
make help           # Show all available commands
make deps           # Install dependencies
make build          # Build the binary
make test           # Run tests with coverage
make lint           # Run linters
make security       # Run security checks
make benchmark      # Run performance benchmarks
make clean          # Clean build artifacts
```

### Testing

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run benchmarks
make benchmark

# Generate coverage report
make test && open coverage/coverage.html
```

### Code Quality

This project maintains high code quality standards:

- **95%+ Test Coverage**: Comprehensive test suite with unit and integration tests
- **Static Analysis**: golangci-lint with 70+ linters enabled
- **Security Scanning**: gosec and govulncheck for vulnerability detection
- **Performance Testing**: Benchmarks and profiling tools
- **Documentation**: GoDoc comments and comprehensive README

## 🔒 Security

ZeroUI takes security seriously:

- **Secure by Default**: No network access, minimal privileges
- **Input Validation**: All user inputs are validated and sanitized
- **Safe File Operations**: Atomic operations with automatic rollback
- **Vulnerability Scanning**: Regular dependency and code security scans
- **Minimal Attack Surface**: Small, focused codebase with few dependencies

## 📊 Observability

Built-in observability features:

- **Structured Logging**: JSON logging with contextual information
- **Metrics**: Prometheus-compatible metrics via OpenTelemetry
- **Health Checks**: Built-in health monitoring
- **Performance Tracking**: Operation timing and success rates

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make changes and add tests
4. Run quality checks: `make check`
5. Commit changes: `git commit -am 'Add amazing feature'`
6. Push to branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Write comprehensive tests (aim for >95% coverage)
- Use structured logging
- Add appropriate documentation
- Follow the existing code style

## 📈 Performance

ZeroUI is designed for performance:

- **Fast Startup**: <10ms cold start time
- **Memory Efficient**: <10MB RSS for typical operations
- **Concurrent Safe**: Thread-safe operations throughout
- **Benchmarked**: Performance regression testing in CI

## 🐳 Docker Usage

```bash
# Run with Docker
docker run --rm -it \
  -v ~/.config/zeroui:/home/appuser/.config/zeroui \
  zeroui/zeroui:latest list apps

# Build your own image
docker build -t zeroui:local .
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh/) for the excellent TUI libraries
- [spf13](https://github.com/spf13) for Cobra CLI framework
- [zerolog](https://github.com/rs/zerolog) for structured logging
- The Go community for inspiration and best practices

---

<p align="center">
  <strong>Made with ❤️ by the ZeroUI team</strong><br>
  <a href="https://github.com/mrtkrcm/ZeroUI">GitHub</a> •
  <a href="https://github.com/mrtkrcm/ZeroUI/issues">Issues</a> •
  <a href="https://github.com/mrtkrcm/ZeroUI/discussions">Discussions</a>
</p>