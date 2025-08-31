#!/bin/bash

# ZeroUI Raycast Extension Installation Script

set -e

echo "🚀 Installing ZeroUI Raycast Extension"
echo "======================================"

# Check if we're in the right directory
if [ ! -f "package.json" ] || [ ! -d "src" ]; then
    echo "❌ Error: Please run this script from the raycast-extension directory"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Error: Node.js is required but not installed"
    echo "Please install Node.js from https://nodejs.org/"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is required but not installed"
    echo "Please install npm along with Node.js"
    exit 1
fi

# Check and setup ZeroUI binary
echo "📋 Setting up ZeroUI binary..."

ZEROUIPATH=""
BINARY_FOUND=false

# Check for ZeroUI binary in various locations
if [ -f "./zeroui" ]; then
    ZEROUIPATH="./zeroui"
    BINARY_FOUND=true
    echo "✅ Found ZeroUI binary in extension directory"
elif [ -f "../build/zeroui" ]; then
    ZEROUIPATH="../build/zeroui"
    BINARY_FOUND=true
    echo "✅ Found ZeroUI binary in build directory"
elif [ -f "../zeroui" ]; then
    ZEROUIPATH="../zeroui"
    BINARY_FOUND=true
    echo "✅ Found ZeroUI binary in project root"
elif command -v zeroui &> /dev/null; then
    ZEROUIPATH="$(which zeroui)"
    BINARY_FOUND=true
    echo "✅ Found ZeroUI in system PATH: $ZEROUIPATH"
else
    echo "⚠️  ZeroUI binary not found in common locations"
    echo "Please build ZeroUI first:"
    echo "  cd /Users/m/code/muka-hq/zeroui"
    echo "  go build -o build/zeroui"
    echo ""
    echo "Or download the binary and place it in this directory"
    echo ""
    read -p "Continue without ZeroUI binary? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Copy binary to extension directory if found
if [ "$BINARY_FOUND" = true ] && [ ! -f "./zeroui" ]; then
    echo "📋 Copying ZeroUI binary to extension directory..."
    cp "$ZEROUIPATH" "./zeroui"
    chmod +x "./zeroui"
    echo "✅ ZeroUI binary copied and made executable"
fi

echo "📦 Installing dependencies..."
npm install

echo "🔍 Running linting..."
npm run lint

echo "🔨 Building extension..."
npm run build

# Validate the setup
if [ -f "./zeroui" ] && [ -f "validate.js" ]; then
    echo ""
    echo "🔍 Validating setup..."
    node validate.js
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "🎯 Next steps:"
echo "1. Open Raycast"
echo "2. Go to Extensions → Import Extension"
echo "3. Select this raycast-extension folder"
echo "4. Start using ZeroUI commands in Raycast!"
echo ""
echo "📚 Available commands:"
echo "• zeroui list-apps"
echo "• zeroui toggle-config <app> <key> <value>"
echo "• zeroui list-values <app>"
echo "• zeroui list-changed <app>"
echo "• zeroui keymap-list <app>"
echo "• zeroui manage-presets"
echo ""
echo "Happy configuring! 🎉"
