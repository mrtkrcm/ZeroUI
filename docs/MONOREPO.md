# ZeroUI Monorepo

This document describes the monorepo structure and development workflow for ZeroUI.

## 📁 Repository Structure

```
zeroui/
├── .github/workflows/          # GitHub Actions CI/CD
├── cmd/                        # CLI commands
├── docs/                       # Documentation
├── internal/                   # Private Go packages
│   ├── tui/                   # Terminal UI components
│   ├── config/                # Configuration management
│   ├── plugins/               # Plugin system
│   └── ...
├── pkg/                       # Public Go packages
├── plugins/                   # Plugin modules
│   └── ghostty-rpc/          # Ghostty RPC plugin
├── raycast-extension/         # Raycast extension (npm workspace)
├── scripts/                   # Build and utility scripts
├── test/                      # Integration tests
├── testdata/                  # Test fixtures
├── go.work                    # Go workspace configuration
├── go.mod                     # Main Go module
├── package.json               # NPM workspace configuration
├── Makefile                   # Build system
└── README.md                  # Project documentation
```

## 🏗️ Workspace Configuration

### Go Workspace

The repository uses Go workspaces (`go.work`) to manage multiple Go modules:

```go
go 1.24.0

use (
    .
    ./plugins/ghostty-rpc
)
```

This allows:
- Shared dependencies between modules
- Unified build and test commands
- Cross-module development

### NPM Workspace

The Raycast extension is managed as an NPM workspace:

```json
{
  "name": "zeroui-monorepo",
  "workspaces": ["raycast-extension"],
  "scripts": {
    "build": "npm run build --workspaces",
    "test": "npm run test --workspaces"
  }
}
```

## 🚀 Development Workflow

### Setting Up Development Environment

1. **Clone the repository:**
   ```bash
   git clone https://github.com/mrtkrcm/zeroui.git
   cd zeroui
   ```

2. **Install all dependencies:**
   ```bash
   make install-deps
   # or
   npm run install:all
   ```

3. **Set up test environment:**
   ```bash
   make test-setup
   ```

### Building All Components

```bash
# Build everything
make workspace-build
# or
npm run build

# Build individual components
make build          # Main Go application
make build-plugins  # All plugins
npm run build:raycast  # Raycast extension
```

### Running Tests

```bash
# Test everything
make workspace-test
# or
npm run test

# Test individual components
make test-fast      # Go tests
npm run test:main   # Go tests via npm
npm run lint:raycast # Raycast extension linting
```

### Development Mode

```bash
# Start Raycast extension in development mode
npm run dev:raycast

# Watch mode for Go development (requires entr)
make dev
```

## 📦 Components

### 1. Main Application (`/`)

- **Language:** Go
- **Purpose:** Core CLI application
- **Build:** `make build`
- **Test:** `make test-fast`

### 2. Raycast Extension (`raycast-extension/`)

- **Language:** TypeScript/React
- **Purpose:** Raycast integration
- **Build:** `npm run build --workspace=raycast-extension`
- **Dev:** `npm run dev --workspace=raycast-extension`
- **Test:** `npm run lint --workspace=raycast-extension`

### 3. Ghostty RPC Plugin (`plugins/ghostty-rpc/`)

- **Language:** Go
- **Purpose:** Ghostty terminal integration
- **Build:** `make build-plugins`
- **Test:** `make test-plugins`

## 🔧 Build System

### Makefile Targets

```bash
# Core targets
make build          # Build main application
make test           # Run all tests with coverage
make clean          # Clean build artifacts

# Workspace targets
make workspace-build    # Build all components
make workspace-test     # Test all components
make workspace-clean    # Clean all artifacts
make workspace-deps     # Download all dependencies

# Plugin targets
make build-plugins      # Build all plugins
make test-plugins       # Test all plugins

# Development
make install-deps       # Install all development dependencies
make fmt               # Format code
make lint              # Run linters
```

### NPM Scripts

```bash
# Workspace commands
npm run build          # Build all components
npm run test           # Test all components
npm run clean          # Clean all components

# Component-specific
npm run build:main     # Build main Go application
npm run build:raycast  # Build Raycast extension
npm run dev:raycast    # Start Raycast development mode
npm run lint:raycast   # Lint Raycast extension
```

## 🧪 Testing Strategy

### Unit Tests
- **Go:** `go test -v -short ./...`
- **TypeScript:** `npm run lint` (ESLint + Prettier)

### Integration Tests
- Located in `test/` directory
- Use deterministic test stubs in `testdata/bin/`
- Run with `make test`

### Visual Tests
- TUI component snapshots in `internal/tui/testdata/`
- Update with `make test-update-baselines`

## 🚢 Release Process

### Automated Release (GitHub Actions)

1. **Create a git tag:**
   ```bash
   ./scripts/version.sh release "Release notes"
   ```

2. **Push the tag:**
   ```bash
   git push origin v1.0.0
   ```

3. **GitHub Actions will:**
   - Build binaries for all platforms
   - Build Raycast extension
   - Create GitHub release
   - Upload release assets

### Manual Release

```bash
# Build all components
make workspace-build

# Create release archives
make release

# Publish Raycast extension
cd raycast-extension && npm run publish
```

## 🔗 Cross-Component Development

### Sharing Code Between Components

1. **Go Packages:** Use `internal/` for private packages, `pkg/` for public ones
2. **TypeScript:** Shared utilities can be moved to separate packages if needed
3. **Build Dependencies:** Use `go.work` for Go modules, npm workspaces for Node.js

### Dependency Management

```bash
# Update Go dependencies
go mod tidy
go work sync

# Update npm dependencies
npm update

# Update Raycast extension dependencies
cd raycast-extension && npm update
```

## 🐛 Troubleshooting

### Common Issues

1. **Go workspace issues:**
   ```bash
   go work sync
   go mod tidy
   ```

2. **NPM workspace issues:**
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

3. **Raycast extension issues:**
   ```bash
   cd raycast-extension
   rm -rf node_modules
   npm install
   npm run build
   ```

### Development Tips

- Use `make workspace-deps` to ensure all dependencies are installed
- Run `make workspace-test` before committing
- Use `make workspace-clean` to reset the workspace
- Check `.gitignore` for workspace-specific exclusions

## 📋 Contributing

1. **Fork the repository**
2. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature
   ```
3. **Make changes and test:**
   ```bash
   make workspace-test
   ```
4. **Commit changes:**
   ```bash
   git commit -am "feat: add your feature"
   ```
5. **Push and create PR:**
   ```bash
   git push origin feature/your-feature
   ```

## 🔗 Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [QUICKSTART.md](QUICKSTART.md) - Getting started guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [TESTING_VALIDATION.md](TESTING_VALIDATION.md) - Testing strategy
