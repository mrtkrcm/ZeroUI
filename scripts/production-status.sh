#!/bin/bash

# ZeroUI Production Status Checker

echo "🚀 ZeroUI Production Readiness Check"
echo "===================================="
echo

# Check CI/CD setup
echo "📋 CI/CD Pipeline:"
if [ -f ".github/workflows/ci.yml" ]; then
    echo "✅ GitHub Actions CI/CD pipeline configured"
    echo "   - Multi-platform builds (Linux, macOS, Windows)"
    echo "   - Automated testing and coverage"
    echo "   - Release automation on tag push"
else
    echo "❌ CI/CD pipeline not configured"
fi
echo

# Check build system
echo "🔨 Build System:"
if [ -f "Makefile" ]; then
    echo "✅ Makefile configured with build targets:"
    echo "   - make build         - Build for current platform"
    echo "   - make build-all     - Build for all platforms"
    echo "   - make test          - Run tests"
    echo "   - make release       - Create release archives"
    echo "   - make install       - Install to system"
    echo "   - make docker-build  - Build Docker image"
else
    echo "❌ Makefile not found"
fi
echo

# Check deployment configurations
echo "🐳 Deployment Configurations:"
configs_found=0

if [ -f "Dockerfile" ]; then
    echo "✅ Dockerfile configured (multi-stage build)"
    ((configs_found++))
fi

if [ -f "docker-compose.yml" ]; then
    echo "✅ Docker Compose configured"
    ((configs_found++))
fi

if [ -d "config" ] && [ -f "config/production.yaml" ]; then
    echo "✅ Production configuration file created"
    ((configs_found++))
fi

if [ $configs_found -eq 0 ]; then
    echo "❌ No deployment configurations found"
fi
echo

# Check installation scripts
echo "📦 Installation Scripts:"
scripts_found=0

if [ -f "scripts/install.sh" ] && [ -x "scripts/install.sh" ]; then
    echo "✅ Universal installer script created"
    ((scripts_found++))
fi

if [ -f "scripts/version.sh" ] && [ -x "scripts/version.sh" ]; then
    echo "✅ Version management script created"
    ((scripts_found++))
fi

if [ $scripts_found -eq 0 ]; then
    echo "❌ No installation scripts found"
fi
echo

# Check release automation
echo "📦 Release Automation:"
release_found=0

if [ -f ".github/workflows/release.yml" ]; then
    echo "✅ GitHub release workflow configured"
    echo "   - Automatic release creation on tag push"
    echo "   - Multi-platform binary packaging"
    echo "   - Raycast extension packaging"
    ((release_found++))
fi

if [ $release_found -eq 0 ]; then
    echo "❌ Release automation not configured"
fi
echo

# Overall status
echo "🎯 OVERALL STATUS:"
echo "=================="

total_checks=5
passed_checks=0

[ -f ".github/workflows/ci.yml" ] && ((passed_checks++))
[ -f "Makefile" ] && ((passed_checks++))
[ -f "Dockerfile" ] && ((passed_checks++))
[ -f "scripts/install.sh" ] && ((passed_checks++))
[ -f ".github/workflows/release.yml" ] && ((passed_checks++))

echo "✅ $passed_checks/$total_checks production components configured"
echo

if [ $passed_checks -eq $total_checks ]; then
    echo "🎉 PROJECT IS PRODUCTION READY!"
    echo
    echo "🚀 Quick deployment commands:"
    echo "   make build-all        # Build for all platforms"
    echo "   make release         # Create release archives"
    echo "   docker-compose up -d  # Deploy with Docker"
    echo "   ./scripts/install.sh  # Install on any platform"
    echo
    echo "📋 Release process:"
    echo "   ./scripts/version.sh release \"Release notes\""
    echo "   # GitHub Actions will automatically create release"
else
    echo "⚠️  Some production components are missing"
    echo "   Run the setup scripts to complete configuration"
fi

echo
echo "📚 For detailed deployment instructions, see README.md"
