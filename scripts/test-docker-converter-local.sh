#!/bin/bash
# Local test script for Docker converter image
# This validates the Docker image before pushing to PR

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

IMAGE_NAME="axon-converter:test-local"
DOCKERFILE="docker/Dockerfile.converter"

echo -e "${YELLOW}🧪 Testing Docker converter image locally...${NC}"
echo ""

# Build the image
echo "1️⃣  Building Docker image..."
if docker build -f "$DOCKERFILE" -t "$IMAGE_NAME" . > /tmp/docker-build.log 2>&1; then
    echo -e "${GREEN}✅ Docker image built successfully${NC}"
else
    echo -e "${RED}❌ Docker image build failed${NC}"
    echo "Build log:"
    tail -50 /tmp/docker-build.log
    exit 1
fi
echo ""

# Test Python version
echo "2️⃣  Testing Python version..."
if docker run --rm --entrypoint="" "$IMAGE_NAME" python3 --version 2>&1; then
    echo -e "${GREEN}✅ Python version check passed${NC}"
else
    echo -e "${RED}❌ Python version check failed${NC}"
    exit 1
fi
echo ""

# Test PyTorch
echo "3️⃣  Testing PyTorch..."
if docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import torch; print('✅ PyTorch:', torch.__version__)" 2>&1 | grep -q "PyTorch:"; then
    echo -e "${GREEN}✅ PyTorch check passed${NC}"
else
    echo -e "${RED}❌ PyTorch check failed${NC}"
    docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import torch; print('PyTorch:', torch.__version__)" 2>&1
    exit 1
fi
echo ""

# Test Transformers
echo "4️⃣  Testing Transformers..."
if docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import transformers; print('✅ Transformers:', transformers.__version__)" 2>&1 | grep -q "Transformers:"; then
    echo -e "${GREEN}✅ Transformers check passed${NC}"
else
    echo -e "${RED}❌ Transformers check failed${NC}"
    docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import transformers; print('Transformers:', transformers.__version__)" 2>&1
    exit 1
fi
echo ""

# Test TensorFlow (suppress CUDA warnings)
echo "5️⃣  Testing TensorFlow..."
# Suppress CUDA warnings by setting environment variables
if docker run --rm -e TF_CPP_MIN_LOG_LEVEL=2 --entrypoint="" "$IMAGE_NAME" python3 -c "import tensorflow as tf; print('✅ TensorFlow:', tf.__version__)" 2>&1 | grep -q "TensorFlow:"; then
    echo -e "${GREEN}✅ TensorFlow check passed${NC}"
else
    echo -e "${RED}❌ TensorFlow check failed${NC}"
    echo "Full output:"
    docker run --rm -e TF_CPP_MIN_LOG_LEVEL=2 --entrypoint="" "$IMAGE_NAME" python3 -c "import tensorflow as tf; print('TensorFlow:', tf.__version__)" 2>&1
    exit 1
fi
echo ""

# Test ONNX
echo "6️⃣  Testing ONNX..."
if docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import onnx; print('✅ ONNX:', onnx.__version__)" 2>&1 | grep -q "ONNX:"; then
    echo -e "${GREEN}✅ ONNX check passed${NC}"
else
    echo -e "${RED}❌ ONNX check failed${NC}"
    docker run --rm --entrypoint="" "$IMAGE_NAME" python3 -c "import onnx; print('ONNX:', onnx.__version__)" 2>&1
    exit 1
fi
echo ""

# Test conversion scripts
echo "7️⃣  Testing conversion scripts..."
if docker run --rm --entrypoint="" "$IMAGE_NAME" test -f /axon/scripts/convert_huggingface.py && \
   docker run --rm --entrypoint="" "$IMAGE_NAME" test -f /axon/scripts/convert_pytorch.py && \
   docker run --rm --entrypoint="" "$IMAGE_NAME" test -f /axon/scripts/convert_tensorflow.py; then
    echo -e "${GREEN}✅ Conversion scripts check passed${NC}"
else
    echo -e "${RED}❌ Conversion scripts check failed${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ All tests passed! Docker image is ready.${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

