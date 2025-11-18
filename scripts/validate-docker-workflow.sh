#!/bin/bash
# Local validation script for docker-converter.yml workflow
# This mimics what the GitHub Actions workflow does

set -e

echo "🔍 Validating Docker Converter Workflow Locally"
echo "================================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="ghcr.io/mlos-foundation/axon-converter:latest"
LOCAL_IMAGE_NAME="axon-converter:test"
DOCKERFILE="docker/Dockerfile.converter"

# Check if Docker is available
echo "1️⃣  Checking Docker availability..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker daemon is not running${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is available${NC}"
echo ""

# Check if Dockerfile exists
echo "2️⃣  Checking Dockerfile..."
if [ ! -f "$DOCKERFILE" ]; then
    echo -e "${RED}❌ Dockerfile not found: $DOCKERFILE${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Dockerfile found: $DOCKERFILE${NC}"
echo ""

# Check if conversion scripts exist
echo "3️⃣  Checking conversion scripts..."
if [ ! -d "scripts/conversion" ]; then
    echo -e "${RED}❌ Conversion scripts directory not found: scripts/conversion${NC}"
    exit 1
fi

SCRIPT_COUNT=$(find scripts/conversion -name "*.py" | wc -l | tr -d ' ')
if [ "$SCRIPT_COUNT" -eq 0 ]; then
    echo -e "${RED}❌ No Python conversion scripts found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Found $SCRIPT_COUNT conversion script(s)${NC}"
echo ""

# Build Docker image
echo "4️⃣  Building Docker image..."
echo "   Image: $LOCAL_IMAGE_NAME"
echo "   Dockerfile: $DOCKERFILE"
echo ""

if docker build -f "$DOCKERFILE" -t "$LOCAL_IMAGE_NAME" .; then
    echo -e "${GREEN}✅ Docker image built successfully${NC}"
else
    echo -e "${RED}❌ Docker image build failed${NC}"
    exit 1
fi
echo ""

# Test Docker image (mimics workflow test steps)
echo "5️⃣  Testing Docker image..."
echo ""

# Test Python version
echo "   Testing Python version..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" python3 --version; then
    echo -e "${GREEN}   ✅ Python version check passed${NC}"
else
    echo -e "${RED}   ❌ Python version check failed${NC}"
    exit 1
fi
echo ""

# Test PyTorch
echo "   Testing PyTorch..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" python3 -c "import torch; print('PyTorch:', torch.__version__)" 2>&1; then
    echo -e "${GREEN}   ✅ PyTorch check passed${NC}"
else
    echo -e "${RED}   ❌ PyTorch check failed${NC}"
    exit 1
fi
echo ""

# Test Transformers
echo "   Testing Transformers..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" python3 -c "import transformers; print('Transformers:', transformers.__version__)" 2>&1; then
    echo -e "${GREEN}   ✅ Transformers check passed${NC}"
else
    echo -e "${RED}   ❌ Transformers check failed${NC}"
    exit 1
fi
echo ""

# Test TensorFlow
echo "   Testing TensorFlow..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" python3 -c "import tensorflow as tf; print('TensorFlow:', tf.__version__)" 2>&1; then
    echo -e "${GREEN}   ✅ TensorFlow check passed${NC}"
else
    echo -e "${RED}   ❌ TensorFlow check failed${NC}"
    exit 1
fi
echo ""

# Test ONNX
echo "   Testing ONNX..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" python3 -c "import onnx; print('ONNX:', onnx.__version__)" 2>&1; then
    echo -e "${GREEN}   ✅ ONNX check passed${NC}"
else
    echo -e "${RED}   ❌ ONNX check failed${NC}"
    exit 1
fi
echo ""

# Test conversion scripts exist
echo "   Testing conversion scripts..."
if docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" test -f /axon/scripts/convert_huggingface.py && \
   docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" test -f /axon/scripts/convert_pytorch.py && \
   docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" test -f /axon/scripts/convert_tensorflow.py; then
    echo -e "${GREEN}   ✅ Conversion scripts present${NC}"
else
    echo -e "${RED}   ❌ Conversion scripts missing${NC}"
    exit 1
fi
echo ""

# List scripts
echo "   Listing conversion scripts:"
docker run --rm --entrypoint="" "$LOCAL_IMAGE_NAME" ls -la /axon/scripts/ 2>&1 | grep -E "\.py$" || true
echo ""

# Validate image name format (lowercase check)
echo "6️⃣  Validating image name format..."
if [[ "$IMAGE_NAME" =~ [A-Z] ]]; then
    echo -e "${RED}❌ Image name contains uppercase letters: $IMAGE_NAME${NC}"
    echo "   Docker requires lowercase repository names"
    exit 1
fi
echo -e "${GREEN}✅ Image name format is valid (lowercase)${NC}"
echo ""

# Summary
echo "================================================"
echo -e "${GREEN}✅ All Docker workflow validation checks passed!${NC}"
echo ""
echo "Image built: $LOCAL_IMAGE_NAME"
echo "Image name for registry: $IMAGE_NAME"
echo ""
echo "To push to registry (requires authentication):"
echo "  docker tag $LOCAL_IMAGE_NAME $IMAGE_NAME"
echo "  docker push $IMAGE_NAME"
echo ""

