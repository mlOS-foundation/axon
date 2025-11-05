#!/bin/bash

# Download and package actual model files from Hugging Face
# This creates real .axon packages that can be used end-to-end

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REGISTRY_DIR="$SCRIPT_DIR"
PACKAGES_DIR="$REGISTRY_DIR/packages"
TEMP_DIR="$REGISTRY_DIR/tmp/models"

# Python script to download models
PYTHON_SCRIPT="$REGISTRY_DIR/download_hf_models.py"

echo "📦 Downloading Real Models from Hugging Face"
echo "=============================================="
echo ""

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is required to download models"
    echo "   Install Python 3 and try again"
    exit 1
fi

# Check if huggingface_hub is installed
if ! python3 -c "import huggingface_hub" 2>/dev/null; then
    echo "📥 Installing huggingface_hub..."
    pip3 install --user huggingface_hub || {
        echo "⚠️  Failed to install huggingface_hub"
        echo "   Try: pip3 install huggingface_hub"
        exit 1
    }
fi

# Create directories
mkdir -p "$PACKAGES_DIR"
mkdir -p "$TEMP_DIR"

echo "📋 Downloading models (this may take a while)..."
echo ""

# Run Python script to download models
python3 "$PYTHON_SCRIPT" "$REGISTRY_DIR" || {
    echo "⚠️  Model download failed"
    echo "   Continuing with placeholder packages..."
    exit 0
}

echo ""
echo "✅ Model downloads complete!"
echo ""
echo "📊 Next steps:"
echo "   1. Update checksums: go run update-checksums.go ."
echo "   2. Start registry: go run server.go ."
echo "   3. Test: axon install nlp/bert-base-uncased@1.0.0"
echo ""

