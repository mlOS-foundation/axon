#!/bin/bash
#
# Validation script to check where manifest and package are created during install
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Axon Install Path Validation"
echo "=========================================="
echo ""

# Check if axon is installed
if ! command -v axon &> /dev/null; then
    echo -e "${RED}❌ Axon CLI is not installed${NC}"
    echo "Please install first: curl -sSL axon.mlosfoundation.org | sh"
    exit 1
fi

# Get axon config directory
AXON_HOME="${HOME}/.axon"
CACHE_DIR="${AXON_HOME}/cache"
MODELS_DIR="${CACHE_DIR}/models"

echo -e "${BLUE}Configuration:${NC}"
echo "  Axon Home: $AXON_HOME"
echo "  Cache Dir: $CACHE_DIR"
echo "  Models Dir: $MODELS_DIR"
echo ""

# Check if init has been run
if [ ! -d "$AXON_HOME" ]; then
    echo -e "${YELLOW}⚠️  Axon not initialized. Running 'axon init'...${NC}"
    axon init
    echo ""
fi

# Test model
TEST_MODEL="hf/bert-base-uncased@latest"
TEST_NAMESPACE="hf"
TEST_NAME="bert-base-uncased"
TEST_VERSION="latest"

echo -e "${BLUE}Testing with model: ${TEST_MODEL}${NC}"
echo ""

# Show temp directory
TEMP_DIR=$(dirname $(mktemp -u))
echo -e "${BLUE}Expected locations:${NC}"
echo "  Temp package: ${TEMP_DIR}/<namespace>-<name>-<version>.axon"
echo "  Cache manifest: ${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/manifest.yaml"
echo "  Cache metadata: ${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/.axon_metadata.json"
echo ""

# Check if model is already installed
if [ -f "${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/manifest.yaml" ]; then
    echo -e "${YELLOW}⚠️  Model already installed. Uninstalling first...${NC}"
    axon uninstall "${TEST_MODEL}" 2>/dev/null || true
    echo ""
fi

echo -e "${BLUE}Installing model...${NC}"
echo ""

# Install model and capture output
axon install "${TEST_MODEL}" 2>&1 | tee /tmp/axon-install.log

echo ""
echo "=========================================="
echo "Validation Results"
echo "=========================================="
echo ""

# Check for manifest
MANIFEST_PATH="${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/manifest.yaml"
if [ -f "$MANIFEST_PATH" ]; then
    echo -e "${GREEN}✅ Manifest created:${NC}"
    echo "  $MANIFEST_PATH"
    echo "  Size: $(ls -lh "$MANIFEST_PATH" | awk '{print $5}')"
    echo ""
else
    echo -e "${RED}❌ Manifest NOT found at:${NC}"
    echo "  $MANIFEST_PATH"
    echo ""
fi

# Check for metadata
METADATA_PATH="${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/.axon_metadata.json"
if [ -f "$METADATA_PATH" ]; then
    echo -e "${GREEN}✅ Metadata created:${NC}"
    echo "  $METADATA_PATH"
    echo "  Content:"
    cat "$METADATA_PATH" | jq '.' 2>/dev/null || cat "$METADATA_PATH"
    echo ""
else
    echo -e "${RED}❌ Metadata NOT found at:${NC}"
    echo "  $METADATA_PATH"
    echo ""
fi

# Check for package in temp (might be cleaned up)
echo -e "${BLUE}Checking temp directory for package...${NC}"
TEMP_PACKAGE="${TEMP_DIR}/${TEST_NAMESPACE}-${TEST_NAME}-${TEST_VERSION}.axon"
if [ -f "$TEMP_PACKAGE" ]; then
    echo -e "${YELLOW}⚠️  Package found in temp (should be moved to cache):${NC}"
    echo "  $TEMP_PACKAGE"
    echo "  Size: $(ls -lh "$TEMP_PACKAGE" | awk '{print $5}')"
    echo ""
else
    echo -e "${BLUE}ℹ️  Package not in temp (may have been cleaned up or moved)${NC}"
    echo ""
fi

# Check cache directory structure
echo -e "${BLUE}Cache directory structure:${NC}"
if [ -d "$MODELS_DIR" ]; then
    find "$MODELS_DIR" -type f -name "*.yaml" -o -name "*.json" -o -name "*.axon" 2>/dev/null | head -20
    echo ""
else
    echo -e "${RED}❌ Models directory does not exist: $MODELS_DIR${NC}"
    echo ""
fi

# Check if package is in cache
CACHE_PACKAGE="${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/*.axon"
if ls ${CACHE_PACKAGE} 2>/dev/null; then
    echo -e "${GREEN}✅ Package found in cache:${NC}"
    ls -lh ${CACHE_PACKAGE}
    echo ""
else
    echo -e "${YELLOW}⚠️  Package NOT found in cache directory${NC}"
    echo "  Expected: ${MODELS_DIR}/${TEST_NAMESPACE}/${TEST_NAME}/${TEST_VERSION}/*.axon"
    echo ""
fi

# Summary
echo "=========================================="
echo "Summary"
echo "=========================================="
echo ""

if [ -f "$MANIFEST_PATH" ] && [ -f "$METADATA_PATH" ]; then
    echo -e "${GREEN}✅ Manifest and metadata are created correctly${NC}"
else
    echo -e "${RED}❌ Manifest or metadata creation failed${NC}"
fi

if ls ${CACHE_PACKAGE} 2>/dev/null; then
    echo -e "${GREEN}✅ Package is stored in cache${NC}"
else
    echo -e "${YELLOW}⚠️  Package is NOT in cache (may be stored elsewhere or cleaned up)${NC}"
fi

echo ""
echo "Full cache directory listing:"
ls -laR "$MODELS_DIR" 2>/dev/null | head -50 || echo "No models directory"

