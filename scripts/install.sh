#!/bin/bash

# ZeroUI Universal Installer

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Detect platform
detect_platform() {
    case "$(uname -s)" in
        Linux*)     
            if [[ "$(uname -m)" == "aarch64" ]]; then
                echo "linux-arm64"
            else
                echo "linux-amd64"
            fi
            ;;
        Darwin*)
            if [[ "$(uname -m)" == "arm64" ]]; then
                echo "darwin-arm64"
            else
                echo "darwin-amd64"
            fi
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            echo "windows-amd64"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Get latest release
get_latest_release() {
    curl -s "https://api.github.com/repos/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases/latest" | 
    grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install_binary() {
    local platform="$1"
    local version="${2:-latest}"
    
    echo -e "${BLUE}Installing ZeroUI for $platform${NC}"
    
    # Create temp directory
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Download archive
    local archive_name="zeroui-${platform}.tar.gz"
    local download_url="https://github.com/mrtkrcm/zeroui/releases/download/${version}/${archive_name}"
    
    echo -e "${YELLOW}Downloading $download_url${NC}"
    curl -L -o "$archive_name" "$download_url"
    
    # Extract
    tar -xzf "$archive_name"
    
    # Install
    local binary_path="$temp_dir/zeroui-${platform}"
    
    if [[ "$platform" == windows-* ]]; then
        binary_path="$temp_dir/zeroui-${platform}.exe"
        # On Windows, suggest adding to PATH
        echo -e "${GREEN}Installation complete!${NC}"
        echo -e "${YELLOW}Add $binary_path to your PATH or move to a directory in PATH${NC}"
    else
        # On Unix-like systems, install to /usr/local/bin
        if [[ -w "/usr/local/bin" ]] || [[ -w "/usr/local" ]]; then
            sudo cp "$binary_path" "/usr/local/bin/zeroui"
            sudo chmod +x "/usr/local/bin/zeroui"
            echo -e "${GREEN}✅ Installed to /usr/local/bin/zeroui${NC}"
        else
            mkdir -p "$HOME/.local/bin"
            cp "$binary_path" "$HOME/.local/bin/zeroui"
            chmod +x "$HOME/.local/bin/zeroui"
            echo -e "${GREEN}✅ Installed to $HOME/.local/bin/zeroui${NC}"
            echo -e "${YELLOW}Add $HOME/.local/bin to your PATH${NC}"
        fi
    fi
    
    # Test installation
    if command -v zeroui >/dev/null 2>&1; then
        echo -e "${GREEN}✅ ZeroUI installed successfully!${NC}"
        zeroui --version
    else
        echo -e "${RED}❌ Installation may have issues. Try restarting your shell.${NC}"
    fi
    
    # Cleanup
    cd - >/dev/null
    rm -rf "$temp_dir"
}

# Main
main() {
    echo -e "${BLUE}🚀 ZeroUI Universal Installer${NC}"
    echo
    
    local platform=$(detect_platform)
    
    if [[ "$platform" == "unknown" ]]; then
        echo -e "${RED}❌ Unsupported platform: $(uname -s) $(uname -m)${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Detected platform: $platform${NC}"
    
    # Get version
    local version="latest"
    if [[ -n "$1" ]]; then
        version="$1"
    fi
    
    echo -e "${YELLOW}📦 Installing version: $version${NC}"
    
    install_binary "$platform" "$version"
    
    echo
    echo -e "${GREEN}🎉 Installation complete!${NC}"
    echo -e "${BLUE}Run 'zeroui --help' to get started${NC}"
}

main "$@"
