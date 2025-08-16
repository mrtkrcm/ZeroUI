# 🚀 ZeroUI Repository Migration Guide

## Complete migration from ConfigToggle to private `mrtkrcm/ZeroUI` repository

---

## 📋 **Pre-Migration Checklist**

- ✅ **Rebranding Complete** - All code updated to ZeroUI
- ✅ **Build Verified** - `go build -o zeroui .` works
- ✅ **Tests Passing** - All functionality maintained
- ✅ **Documentation Updated** - README, docs, help text all rebranded

---

## 🔧 **Step 1: Create Private Repository**

### **Via GitHub Web Interface:**
1. Go to [GitHub New Repository](https://github.com/new)
2. **Repository name**: `ZeroUI`
3. **Owner**: `mrtkrcm`
4. **Visibility**: ✅ **Private**
5. **Description**: `Zero-configuration UI toolkit manager. Fastest way to manage development tool configurations with powerful CLI and interactive TUI.`
6. **DO NOT** initialize with README, .gitignore, or license (we have existing files)
7. Click **"Create repository"**

### **Via GitHub CLI (Alternative):**
```bash
# Install gh CLI if not already installed
brew install gh

# Create private repository
gh repo create mrtkrcm/ZeroUI --private --description "Zero-configuration UI toolkit manager"
```

---

## 🔄 **Step 2: Update Git Remote**

```bash
# From your current configtoggle directory
cd /Users/m/code/muka-hq/configtoggle

# Add new remote (keep old as backup)
git remote add zeroui https://github.com/mrtkrcm/ZeroUI.git

# Verify remotes
git remote -v
# Should show:
# origin    [original-repo-url] (fetch)
# origin    [original-repo-url] (push) 
# zeroui    https://github.com/mrtkrcm/ZeroUI.git (fetch)
# zeroui    https://github.com/mrtkrcm/ZeroUI.git (push)
```

---

## 📤 **Step 3: Push to New Repository**

```bash
# Push all branches and tags to new repository
git push zeroui main
git push zeroui --all
git push zeroui --tags

# Verify push was successful
git ls-remote zeroui
```

---

## 🔧 **Step 4: Switch Primary Remote**

```bash
# Make zeroui the default remote
git remote remove origin
git remote rename zeroui origin

# Verify new setup
git remote -v
# Should show:
# origin    https://github.com/mrtkrcm/ZeroUI.git (fetch)
# origin    https://github.com/mrtkrcm/ZeroUI.git (push)

# Test connection
git push origin main
```

---

## 📁 **Step 5: Update Local Directory Structure**

```bash
# Optional: Rename local directory to match new project
cd /Users/m/code/muka-hq/
mv configtoggle ZeroUI
cd ZeroUI

# Update any local development scripts that reference paths
```

---

## 🔧 **Step 6: Update Go Module Cache**

```bash
# Clear Go module cache to avoid import issues
go clean -modcache

# Update dependencies
go mod download
go mod tidy

# Verify everything builds
go build -o zeroui .

# Test the rebranded application
./zeroui --help
./zeroui design-system  # Should show ZeroUI branding
```

---

## 🚀 **Step 7: Configure Repository Settings**

### **GitHub Repository Settings:**

1. **Go to** `https://github.com/mrtkrcm/ZeroUI/settings`

2. **General Settings:**
   - Description: `Zero-configuration UI toolkit manager. Fastest way to manage development tool configurations with powerful CLI and interactive TUI.`
   - Website: (optional - your personal site or docs)
   - Topics: `cli`, `tui`, `configuration`, `developer-tools`, `go`, `terminal`, `ui-toolkit`

3. **Security Settings:**
   - ✅ Enable private vulnerability reporting
   - ✅ Enable dependency graph
   - ✅ Enable Dependabot alerts

4. **Actions Settings:**
   - ✅ Enable GitHub Actions (if you want CI/CD)
   - Configure secrets if needed

5. **Pages Settings:**
   - Configure if you want to host documentation

---

## 📋 **Step 8: Update Documentation**

The rebranding has already updated:
- ✅ `README.md` - Complete ZeroUI branding
- ✅ All CLI help text and commands
- ✅ Documentation files and guides
- ✅ Code comments and error messages

### **Verify Documentation:**
```bash
# Check README displays correctly
cat README.md | head -20

# Verify CLI help text
./zeroui --help
./zeroui design-system --help
```

---

## 🧪 **Step 9: Final Verification**

### **Build & Test:**
```bash
# Clean build
go clean
go build -o zeroui .

# Test core functionality  
./zeroui list apps
./zeroui design-system
./zeroui --version

# Run tests
go test ./...

# Test coverage
go test -cover ./...
```

### **Verify Imports:**
```bash
# Check all imports resolve correctly
go mod verify
go list -m all
```

---

## ✅ **Post-Migration Checklist**

- [ ] ✅ Private repository `mrtkrcm/ZeroUI` created
- [ ] ✅ All branches and tags pushed successfully
- [ ] ✅ Git remotes updated to new repository
- [ ] ✅ Go module builds without errors
- [ ] ✅ CLI commands work with new branding
- [ ] ✅ Tests pass with new imports
- [ ] ✅ Repository settings configured
- [ ] ✅ Documentation displays correctly
- [ ] ✅ Design system showcase shows ZeroUI branding

---

## 🎯 **Success Criteria**

When migration is complete, you should have:

1. **Private Repository**: `https://github.com/mrtkrcm/ZeroUI` (private access only)
2. **Working CLI**: `./zeroui --help` shows ZeroUI branding
3. **Clean Imports**: All Go imports use `github.com/mrtkrcm/ZeroUI`
4. **Functional Testing**: All commands work normally
5. **Design System**: `./zeroui design-system` shows complete ZeroUI branding

---

## 🚨 **Troubleshooting**

### **Import Path Issues:**
```bash
# If you get import path errors
go mod edit -module=github.com/mrtkrcm/ZeroUI
go mod tidy
```

### **Remote Issues:**
```bash
# If remote push fails
git remote set-url origin https://github.com/mrtkrcm/ZeroUI.git
git push -u origin main
```

### **Build Issues:**
```bash
# If build fails after migration
go clean -cache
go mod download
go build -v .
```

---

## 🎉 **Congratulations!**

Your **ZeroUI** project is now live as a private repository with complete rebranding! 

The transformation from ConfigToggle to ZeroUI maintains all functionality while presenting a clean, modern brand identity focused on zero-configuration simplicity for developers.

**Next Steps:**
- Start using `./zeroui` commands
- Share with collaborators (private repo access)
- Continue development with the new brand identity
- Consider public release when ready

Welcome to **ZeroUI** - *The interface that isn't there!* 🎯✨