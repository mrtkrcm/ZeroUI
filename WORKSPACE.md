# ZeroUI Monorepo Workspace

## 🚀 Quick Start

```bash
# Install all dependencies
make install-deps

# Build everything
make workspace-build

# Run tests
make workspace-test

# Start development
npm run dev:raycast  # Raycast extension
make dev            # Go application (requires entr)
```

## 📦 Components

This monorepo contains:

- **Main CLI Application** (`/`) - Go-based configuration manager
- **Raycast Extension** (`raycast-extension/`) - Mac desktop integration
- **Ghostty RPC Plugin** (`plugins/ghostty-rpc/`) - Terminal integration

## 🛠️ Development Commands

### Build All Components
```bash
make workspace-build
# or
npm run build
```

### Test All Components
```bash
make workspace-test
# or
npm run test
```

### Clean Everything
```bash
make workspace-clean
# or
npm run clean
```

### Install Dependencies
```bash
make workspace-deps
# or
npm run install:all
```

## 🏗️ Workspace Structure

- **Go Workspace**: Managed by `go.work`
- **NPM Workspace**: Managed by root `package.json`
- **Unified CI/CD**: GitHub Actions with multi-platform builds
- **Cross-platform**: Linux, macOS, Windows support

## 📚 Documentation

- [Full Monorepo Guide](docs/MONOREPO.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Quick Start](docs/QUICKSTART.md)
- [Contributing](docs/CONTRIBUTING.md)

## 🔗 Useful Links

- [GitHub Repository](https://github.com/mrtkrcm/zeroui)
- [Issues](https://github.com/mrtkrcm/zeroui/issues)
- [Releases](https://github.com/mrtkrcm/zeroui/releases)
