#!/bin/bash
#
# Build release binaries for all platforms
# Creates archives ready for GitHub releases
#

set -e

VERSION="${1:-1.0.0}"
BINARY_NAME="axon"
BUILD_DIR="build"
RELEASE_DIR="release"

echo "🚀 Building Axon v${VERSION} for all platforms..."

# Clean previous builds
rm -rf "${BUILD_DIR}" "${RELEASE_DIR}"
mkdir -p "${BUILD_DIR}" "${RELEASE_DIR}"

# Platforms to build
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%%/*}"
    GOARCH="${PLATFORM##*/}"
    OUTPUT_NAME="${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    ARCHIVE_NAME="${OUTPUT_NAME}.tar.gz"
    
    echo ""
    echo "📦 Building ${GOOS}/${GOARCH}..."
    
    # Build binary
    GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags "-X main.version=v${VERSION}" \
        -o "${BUILD_DIR}/${OUTPUT_NAME}/${BINARY_NAME}" \
        ./cmd/axon
    
    # Create archive
    cd "${BUILD_DIR}/${OUTPUT_NAME}"
    tar -czf "../../${RELEASE_DIR}/${ARCHIVE_NAME}" "${BINARY_NAME}"
    cd ../..
    
    # Create checksum
    cd "${RELEASE_DIR}"
    shasum -a 256 "${ARCHIVE_NAME}" > "${ARCHIVE_NAME}.sha256"
    cd ..
    
    echo "✅ Built ${ARCHIVE_NAME}"
done

echo ""
echo "🎉 Release builds complete!"
echo "📁 Release files in: ${RELEASE_DIR}/"
echo ""
echo "Files created:"
ls -lh "${RELEASE_DIR}/"

