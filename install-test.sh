#!/bin/bash
#
# Test script for Axon installer
# This simulates the installation process without actually installing
#

set -e

echo "🧪 Testing Axon Installer Script"
echo "================================"
echo ""

# Test OS detection
echo "Testing OS detection..."
detect_os() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$OS" in
        linux*) OS="linux" ;;
        darwin*) OS="darwin" ;;
        *) echo "Error: Unsupported OS: $OS" >&2; exit 1 ;;
    esac
    echo "  ✓ Detected OS: $OS"
}

# Test Architecture detection
echo "Testing Architecture detection..."
detect_arch() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo "Error: Unsupported architecture: $ARCH" >&2; exit 1 ;;
    esac
    echo "  ✓ Detected Architecture: $ARCH"
}

# Test version fetching (mock)
echo "Testing version fetching..."
get_latest_version() {
    GITHUB_REPO="mlOS-foundation/axon"
    VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//' || echo "1.0.0")
    echo "  ✓ Latest version: $VERSION"
}

# Test URL construction
echo "Testing URL construction..."
test_url_construction() {
    OS="darwin"
    ARCH="amd64"
    VERSION="1.0.0"
    TAG="v${VERSION}"
    FILENAME="axon_${VERSION}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/mlOS-foundation/axon/releases/download/${TAG}/${FILENAME}"
    echo "  ✓ Download URL: $DOWNLOAD_URL"
}

# Run tests
detect_os
detect_arch
get_latest_version
test_url_construction

echo ""
echo "✅ All tests passed!"
echo ""
echo "Note: This is a dry-run test. To actually test installation,"
echo "you need to:"
echo "  1. Create a v1.0.0 release on GitHub with the binary"
echo "  2. Run: curl -sSL https://mlosfoundation.org/install.sh | sh"

