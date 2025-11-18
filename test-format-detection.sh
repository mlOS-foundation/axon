#!/bin/bash
# Test format detection for models from different repositories

set -e

AXON_BIN="${HOME}/.local/bin/axon"
CACHE_DIR="${HOME}/.axon/cache/models"

echo "🧪 Testing format detection across repositories"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_model() {
    local namespace=$1
    local name=$2
    local version=${3:-latest}
    local expected_format=$4
    local model_spec="${namespace}/${name}@${version}"
    
    echo -e "${YELLOW}Testing: ${model_spec}${NC}"
    echo "Expected format: ${expected_format}"
    
    # Uninstall if exists
    $AXON_BIN uninstall "${model_spec}" 2>/dev/null || true
    
    # Install model
    if $AXON_BIN install "${model_spec}" > /tmp/axon_install.log 2>&1; then
        echo "✓ Installation successful"
        
        # Extract cache path
        cache_path="${CACHE_DIR}/${namespace}/${name}/${version}"
        
        # Check manifest for execution_format
        if [ -f "${cache_path}/manifest.yaml" ]; then
            execution_format=$(python3 -c "
import sys, json
try:
    with open('${cache_path}/manifest.yaml') as f:
        m = json.load(f)
        ef = m.get('Spec', {}).get('Format', {}).get('execution_format', 'NOT SET')
        print(ef)
except Exception as e:
    print('ERROR: ' + str(e))
" 2>/dev/null)
            
            echo "Detected execution_format: ${execution_format}"
            
            if [ "${execution_format}" != "NOT SET" ] && [ "${execution_format}" != "" ]; then
                echo -e "${GREEN}✓ execution_format is set: ${execution_format}${NC}"
                
                # Check if it matches expected (if provided)
                if [ -n "${expected_format}" ] && [ "${execution_format}" != "${expected_format}" ]; then
                    echo -e "${YELLOW}⚠ Warning: Expected ${expected_format}, got ${execution_format}${NC}"
                fi
            else
                echo -e "${RED}✗ execution_format is NOT SET${NC}"
            fi
            
            # Show format type
            format_type=$(python3 -c "
import sys, json
try:
    with open('${cache_path}/manifest.yaml') as f:
        m = json.load(f)
        ft = m.get('Spec', {}).get('Format', {}).get('type', 'NOT SET')
        print(ft)
except:
    print('NOT SET')
" 2>/dev/null)
            echo "Format type: ${format_type}"
            
        else
            echo -e "${RED}✗ Manifest file not found${NC}"
        fi
        
        # List files in cache
        echo "Files in cache:"
        ls -lh "${cache_path}" | grep -E "\.(onnx|pth|pt|pb|h5|saved_model)" || echo "  (no model files found)"
        
    else
        echo -e "${RED}✗ Installation failed${NC}"
        echo "Last 10 lines of install log:"
        tail -10 /tmp/axon_install.log
    fi
    
    echo ""
}

# Test models from each repository
echo "1. PyTorch Hub (PyTorch format)"
test_model "pytorch" "vision/resnet18" "latest" "pytorch"

echo "2. Hugging Face (should convert to ONNX or stay PyTorch)"
test_model "hf" "distilbert-base-uncased" "latest" "onnx"

echo "3. TensorFlow Hub (TensorFlow format)"
# Try a small TensorFlow Hub model
test_model "tfhub" "google/imagenet/mobilenet_v2_100_224/classification/5" "5" "tensorflow"

echo "4. ModelScope (PyTorch format typically)"
test_model "modelscope" "damo/cv_resnet18_image-classification" "latest" "pytorch"

echo ""
echo "================================================"
echo "✅ Format detection test complete!"
echo ""
echo "Summary:"
echo "- Check execution_format is set for all models"
echo "- Verify format detection works correctly"
echo "- Ensure manifest reflects actual model files"

