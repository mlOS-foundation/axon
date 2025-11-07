#!/bin/sh
#
# Axon Installer
# Installs Axon CLI from GitHub releases
#
# Usage: curl -sSL https://mlosfoundation.org/install.sh | sh
# Or: curl -sSL axon.mlosfoundation.org | sh
# Or: curl -sSL install.axon.mlos.org | sh
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AXON_VERSION="${AXON_VERSION:-latest}"
GITHUB_REPO="mlOS-foundation/axon"
INSTALL_DIR="${AXON_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="axon"

# Detect OS and Architecture
detect_os() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$OS" in
        linux*) OS="linux" ;;
        darwin*) OS="darwin" ;;
        *) echo "${RED}Error: Unsupported OS: $OS${NC}" >&2; exit 1 ;;
    esac
}

detect_arch() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo "${RED}Error: Unsupported architecture: $ARCH${NC}" >&2; exit 1 ;;
    esac
}

# Get latest release version from GitHub
get_latest_version() {
    if [ "$AXON_VERSION" = "latest" ]; then
        VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
        if [ -z "$VERSION" ]; then
            echo "${YELLOW}Warning: Could not fetch latest version, using v1.0.0${NC}" >&2
            VERSION="1.0.0"
        fi
    else
        VERSION="$AXON_VERSION"
    fi
    echo "$VERSION"
}

# Download and install Axon
install_axon() {
    echo "${BLUE}🚀 Installing Axon CLI...${NC}"
    
    # Detect OS and architecture
    detect_os
    detect_arch
    
    # Get version
    VERSION=$(get_latest_version)
    echo "${BLUE}📦 Version: ${VERSION}${NC}"
    echo "${BLUE}💻 OS: ${OS}, Architecture: ${ARCH}${NC}"
    
    # Construct download URL
    TAG="v${VERSION}"
    FILENAME="axon_${VERSION}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/${FILENAME}"
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    echo "${BLUE}⬇️  Downloading Axon from GitHub...${NC}"
    
    # Try to download with expected filename first
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME" 2>/dev/null; then
        # Fallback: List release assets and find matching one
        echo "${YELLOW}⚠️  Asset not found with expected name, searching release assets...${NC}" >&2
        ASSETS_JSON=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${TAG}")
        
        # Try to find a matching asset
        MATCHING_ASSET=$(echo "$ASSETS_JSON" | grep -o "\"browser_download_url\":[^,]*" | grep -i "${OS}.*${ARCH}\|${ARCH}.*${OS}" | head -1 | cut -d'"' -f4)
        
        if [ -z "$MATCHING_ASSET" ]; then
            # Last resort: try generic name (for v1.0.0 with broken naming)
            FALLBACK_URL="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/axon___.tar.gz"
            if curl -fsSL "$FALLBACK_URL" -o "$TMP_DIR/$FILENAME" 2>/dev/null; then
                echo "${YELLOW}⚠️  Using fallback asset name${NC}" >&2
                DOWNLOAD_URL="$FALLBACK_URL"
            else
                echo "${RED}❌ Error: Failed to download Axon${NC}" >&2
                echo "${YELLOW}   Tried: $DOWNLOAD_URL${NC}" >&2
                echo "${YELLOW}   Please check if the release exists on GitHub${NC}" >&2
                echo "${YELLOW}   Release: https://github.com/${GITHUB_REPO}/releases/tag/${TAG}${NC}" >&2
                exit 1
            fi
        else
            echo "${YELLOW}⚠️  Using alternative asset: $MATCHING_ASSET${NC}" >&2
            curl -fsSL "$MATCHING_ASSET" -o "$TMP_DIR/$FILENAME"
            DOWNLOAD_URL="$MATCHING_ASSET"
        fi
    fi
    
    echo "${BLUE}📦 Extracting...${NC}"
    tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"
    
    # Find the binary (could be in root or in a subdirectory)
    if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
        BINARY_PATH="$TMP_DIR/$BINARY_NAME"
    elif [ -f "$TMP_DIR/axon_${OS}_${ARCH}/$BINARY_NAME" ]; then
        BINARY_PATH="$TMP_DIR/axon_${OS}_${ARCH}/$BINARY_NAME"
    else
        echo "${RED}❌ Error: Could not find axon binary in archive${NC}" >&2
        exit 1
    fi
    
    # Install binary
    echo "${BLUE}📥 Installing to $INSTALL_DIR...${NC}"
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Verify installation
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        INSTALLED_VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || echo "installed")
        echo "${GREEN}✅ Axon installed successfully!${NC}"
        echo "${GREEN}   Location: $INSTALL_DIR/$BINARY_NAME${NC}"
        echo "${GREEN}   Version: $INSTALLED_VERSION${NC}"
    else
        echo "${RED}❌ Error: Installation failed${NC}" >&2
        exit 1
    fi
    
    # Check if install directory is in PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo "${YELLOW}⚠️  Warning: $INSTALL_DIR is not in your PATH${NC}"
        echo "${YELLOW}   Add it to your PATH by running:${NC}"
        echo "${BLUE}   export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        
        # Detect shell and suggest adding to profile
        SHELL_NAME=$(basename "$SHELL")
        case "$SHELL_NAME" in
            bash)
                echo "${YELLOW}   Or add to ~/.bashrc or ~/.bash_profile:${NC}"
                echo "${BLUE}   echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc${NC}"
                ;;
            zsh)
                echo "${YELLOW}   Or add to ~/.zshrc:${NC}"
                echo "${BLUE}   echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc${NC}"
                ;;
        esac
    else
        echo "${GREEN}✅ Axon is ready to use!${NC}"
        echo "${BLUE}   Try: axon --version${NC}"
    fi
}

# Main execution
main() {
    echo "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║      Axon CLI Installer                ║"
    echo "╚════════════════════════════════════════╝"
    echo "${NC}"
    
    install_axon
    
    echo ""
    echo "${GREEN}🎉 Installation complete!${NC}"
    echo "${BLUE}📖 Next steps:${NC}"
    echo "   ${BLUE}axon init${NC}              # Initialize Axon"
    echo "   ${BLUE}axon install hf/bert-base-uncased@latest${NC}  # Install a model"
    echo "   ${BLUE}axon list${NC}               # List installed models"
    echo ""
}

# Run main function
main

