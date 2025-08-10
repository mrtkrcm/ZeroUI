# ZeroUI Rebranding Guide

This guide provides a comprehensive overview of the ConfigToggle to ZeroUI rebranding process and serves as documentation for repository setup.

## 🎯 Complete Rebranding Summary

The project has been successfully rebranded from **ConfigToggle** to **ZeroUI** with the following changes:

### 1. **Brand Identity**
- **Old**: ConfigToggle - "A state-of-the-art CLI tool for managing application configurations"
- **New**: ZeroUI - "Zero-configuration UI toolkit manager for developers"

### 2. **Repository Migration**
- **Old**: `github.com/m/configtoggle`
- **New**: `github.com/mrtkrcm/ZeroUI`

### 3. **Binary Name**
- **Old**: `configtoggle`/`ctoggle`
- **New**: `zeroui`

### 4. **Configuration Paths**
- **Old**: `~/.config/configtoggle/`
- **New**: `~/.config/zeroui/`

## 📂 Files Updated

### **Core Go Files**
- [x] `go.mod` - Module path changed
- [x] `main.go` - Import paths updated
- [x] All `cmd/*.go` files - Commands, examples, and import paths
- [x] All `internal/*.go` files - Import paths and branding
- [x] All test files - Import paths and assertions

### **Documentation**
- [x] `README.md` - Complete rebranding
- [x] All `*.md` files in docs/ - References and branding
- [x] Design system documentation
- [x] Implementation guides and cheatsheets

### **Build & Deployment**
- [x] `Dockerfile` - Image name, labels, paths
- [x] `Makefile` - Binary name and build flags
- [x] `.github/workflows/ci.yml` - Build artifacts and paths
- [x] Shell scripts - Binary references

### **Configuration**
- [x] Error types and messages
- [x] TUI application titles and help text
- [x] CLI command descriptions and examples
- [x] Docker container paths and metadata

## 🔧 Technical Changes Made

### **Go Module Updates**
```diff
- module github.com/m/configtoggle
+ module github.com/mrtkrcm/ZeroUI
```

### **Import Path Updates**
```diff
- "github.com/m/configtoggle/internal/config"
+ "github.com/mrtkrcm/ZeroUI/internal/config"
```

### **Binary Build**
```diff
- go build -o configtoggle .
+ go build -o zeroui .
```

### **Configuration Paths**
```diff
- ~/.config/configtoggle/
+ ~/.config/zeroui/
```

### **Error Types**
```diff
- ConfigToggleError -> ZeroUIError
- GetConfigToggleError -> GetZeroUIError
```

### **CLI Commands & Examples**
```diff
- configtoggle list apps
+ zeroui list apps

- configtoggle toggle ghostty theme dark
+ zeroui toggle ghostty theme dark
```

## 🚀 Repository Setup Instructions

### 1. **Create New Private Repository**
1. Go to GitHub and create a new private repository: `mrtkrcm/ZeroUI`
2. **Do not** initialize with README, .gitignore, or license (we'll push existing content)

### 2. **Update Remote Origin**
```bash
# Remove old remote
git remote remove origin

# Add new remote
git remote add origin git@github.com:mrtkrcm/ZeroUI.git

# Push all branches and tags
git push -u origin main
git push origin --all
git push origin --tags
```

### 3. **Update GitHub Settings**
- Set repository description: "Zero-configuration UI toolkit manager for developers"
- Add topics: `go`, `cli`, `tui`, `configuration`, `toolkit`, `ui`
- Configure branch protection rules for `main`
- Set up GitHub Actions secrets if needed

### 4. **Update CI/CD Badges**
The README.md already contains updated badge URLs:
- CI/CD Pipeline: `https://github.com/mrtkrcm/ZeroUI/actions/workflows/ci.yml`
- Go Report Card: `https://goreportcard.com/badge/github.com/mrtkrcm/ZeroUI`
- Coverage: `https://codecov.io/gh/mrtkrcm/ZeroUI`

## 🔄 Migration Checklist

- [x] **Go Module Path**: Updated to `github.com/mrtkrcm/ZeroUI`
- [x] **Import Statements**: All internal imports updated
- [x] **Binary Name**: Changed from `configtoggle` to `zeroui`
- [x] **CLI Commands**: All help text and examples updated
- [x] **Configuration Paths**: Updated from `configtoggle` to `zeroui`
- [x] **Error Types**: Renamed error types and functions
- [x] **Documentation**: Complete README and docs rebranding
- [x] **Build Files**: Dockerfile, Makefile, GitHub Actions updated
- [x] **TUI Branding**: All interface titles and messages updated
- [x] **Container Images**: Docker labels and paths updated

## ✅ Quality Assurance

### **Build Verification**
```bash
# Clean build test
go mod tidy
go build -o zeroui .

# Verify binary works
./zeroui --help
./zeroui list apps
```

### **Expected Output**
The `zeroui --help` command should show:
```
ZeroUI is a zero-configuration UI toolkit manager that simplifies managing
UI configurations, themes, and settings across development tools and applications.
Built for speed and simplicity with both CLI and interactive TUI interfaces.

Examples:
  zeroui toggle ghostty theme dark
  zeroui cycle alacritty font
  zeroui ui
  zeroui preset vscode minimal
```

## 🎨 Brand Positioning

### **New Identity - ZeroUI**
- **Vision**: Zero-configuration UI toolkit management
- **Target**: Developers who want fast, simple configuration management
- **USP**: Speed and simplicity with powerful TUI interfaces
- **Position**: The fastest way to manage development tool configurations

### **Key Messages**
- "Zero-configuration UI toolkit manager"
- "Built for speed and simplicity"
- "The fastest way to manage development tool configurations"
- "Intuitive CLI and interactive TUI interfaces"

## 📈 Next Steps

1. **Repository Setup**: Create and configure `mrtkrcm/ZeroUI` repository
2. **Domain/Branding**: Consider domain registration if planning web presence
3. **Package Distribution**: Update package manager distributions (Homebrew, etc.)
4. **Documentation Site**: Consider GitHub Pages or dedicated documentation site
5. **Community**: Update any community links, discussions, or external references

## 🔧 Development Commands

```bash
# Build
make build

# Test
make test

# Run with new branding
./zeroui ui

# Docker build
docker build -t zeroui:latest .
```

---

**Rebranding Status**: ✅ **COMPLETE**

All files have been successfully rebranded from ConfigToggle to ZeroUI while maintaining full functionality. The project is ready for repository migration to `github.com/mrtkrcm/ZeroUI`.