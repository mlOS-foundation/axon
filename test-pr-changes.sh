#!/bin/bash

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 Axon PR Changes - Comprehensive Test Suite${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

test_result() {
    local test_name="$1"
    local result="$2"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✅ PASS: $test_name${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAIL: $test_name${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test 1: Build Axon
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 1: Build Axon${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if make build 2>&1 | grep -q "Build complete"; then
    test_result "Axon Build" "PASS"
else
    test_result "Axon Build" "FAIL"
    echo -e "${RED}Build failed - stopping tests${NC}"
    exit 1
fi

# Test 2: Go tests
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 2: Go Unit Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if go test ./... 2>&1 | grep -q "PASS"; then
    test_result "Go Unit Tests" "PASS"
else
    test_result "Go Unit Tests" "FAIL"
fi

# Test 3: Docker converter image build
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 3: Docker Converter Image Build${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

echo "Building Docker converter image (this may take 5-10 minutes)..."
if docker build -f docker/Dockerfile.converter -t ghcr.io/mlos-foundation/axon-converter:test . 2>&1 | tee /tmp/docker-build.log | grep -q "Successfully built"; then
    test_result "Docker Image Build" "PASS"
else
    test_result "Docker Image Build" "FAIL"
    echo -e "${YELLOW}Check /tmp/docker-build.log for details${NC}"
fi

# Test 4: Docker image has required dependencies
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 4: Docker Image Dependencies${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

ALL_DEPS_OK=true

# Test PyTorch
if docker run --rm --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import torch; print('PyTorch:', torch.__version__)" 2>&1 | grep -q "PyTorch:"; then
    echo -e "${GREEN}  ✅ PyTorch${NC}"
else
    echo -e "${RED}  ❌ PyTorch${NC}"
    ALL_DEPS_OK=false
fi

# Test Transformers
if docker run --rm --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import transformers; print('Transformers:', transformers.__version__)" 2>&1 | grep -q "Transformers:"; then
    echo -e "${GREEN}  ✅ Transformers${NC}"
else
    echo -e "${RED}  ❌ Transformers${NC}"
    ALL_DEPS_OK=false
fi

# Test TensorFlow
if docker run --rm -e TF_CPP_MIN_LOG_LEVEL=2 --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import tensorflow; print('TensorFlow:', tensorflow.__version__)" 2>&1 | grep -q "TensorFlow:"; then
    echo -e "${GREEN}  ✅ TensorFlow${NC}"
else
    echo -e "${RED}  ❌ TensorFlow${NC}"
    ALL_DEPS_OK=false
fi

# Test ONNX
if docker run --rm --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import onnx; print('ONNX:', onnx.__version__)" 2>&1 | grep -q "ONNX:"; then
    echo -e "${GREEN}  ✅ ONNX${NC}"
else
    echo -e "${RED}  ❌ ONNX${NC}"
    ALL_DEPS_OK=false
fi

# Test Optimum
if docker run --rm --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import optimum; print('Optimum:', optimum.__version__)" 2>&1 | grep -q "Optimum:"; then
    echo -e "${GREEN}  ✅ Optimum${NC}"
else
    echo -e "${RED}  ❌ Optimum${NC}"
    ALL_DEPS_OK=false
fi

# Test conversion scripts
if docker run --rm --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test test -f /axon/scripts/convert_huggingface.py; then
    echo -e "${GREEN}  ✅ Conversion scripts${NC}"
else
    echo -e "${RED}  ❌ Conversion scripts${NC}"
    ALL_DEPS_OK=false
fi

if [ "$ALL_DEPS_OK" = true ]; then
    test_result "Docker Dependencies" "PASS"
else
    test_result "Docker Dependencies" "FAIL"
fi

# Test 5: Hugging Face model conversion (GPT-2)
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 5: Hugging Face Conversion (GPT-2)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Clean cache first
rm -rf ~/.axon/cache/models/hf/distilgpt2/latest

if ./bin/axon install hf/distilgpt2@latest 2>&1 | tee /tmp/gpt2-install.log | grep -q "success\|✅"; then
    # Check if ONNX file was created
    if [ -f ~/.axon/cache/models/hf/distilgpt2/latest/model.onnx ]; then
        # Verify ONNX file is valid
        if docker run --rm -v ~/.axon/cache/models:/models --entrypoint="" ghcr.io/mlos-foundation/axon-converter:test python3 -c "import onnx; model = onnx.load('/models/hf/distilgpt2/latest/model.onnx'); print('Valid ONNX model')" 2>&1 | grep -q "Valid ONNX model"; then
            test_result "GPT-2 Conversion" "PASS"
        else
            test_result "GPT-2 Conversion (ONNX invalid)" "FAIL"
        fi
    else
        test_result "GPT-2 Conversion (no ONNX file)" "FAIL"
    fi
else
    test_result "GPT-2 Conversion" "FAIL"
    echo -e "${YELLOW}Check /tmp/gpt2-install.log for details${NC}"
fi

# Test 6: Namespace routing
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 6: Namespace Routing${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Test that hf/ models use convert_huggingface.py
if grep -q "convert_huggingface.py" internal/converter/docker.go && \
   grep -q "namespace" internal/converter/docker.go; then
    test_result "Namespace Routing Logic" "PASS"
else
    test_result "Namespace Routing Logic" "FAIL"
fi

# Test 7: Docker image naming (lowercase)
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test 7: Docker Image Naming${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if grep -q "mlos-foundation" internal/converter/docker.go && \
   ! grep -q "mlOS-foundation" internal/converter/docker.go; then
    test_result "Docker Image Naming (lowercase)" "PASS"
else
    test_result "Docker Image Naming" "FAIL"
fi

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}📊 Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

echo ""
echo -e "Total Tests: ${TESTS_TOTAL}"
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
else
    echo -e "Failed: ${TESTS_FAILED}"
fi
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}🎉 ALL TESTS PASSED - Ready for PR!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}❌ SOME TESTS FAILED - Fix before PR${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi

