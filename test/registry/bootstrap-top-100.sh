#!/bin/bash

# Bootstrap script to install Axon and set up local registry with top 100 Hugging Face models
# This creates a base registry that can later be deployed as a hosted model repository

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REGISTRY_DIR="$SCRIPT_DIR"
AXON_REPO_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)/axon"

echo "🚀 Axon Registry Bootstrap - Top 100 Models"
echo "==========================================="
echo ""

# Step 1: Check if Axon is installed
echo "1️⃣  Checking Axon installation..."
if command -v axon &> /dev/null; then
    AXON_BIN="axon"
    echo "   ✅ Axon found in PATH"
elif [ -f "$AXON_REPO_DIR/bin/axon" ]; then
    AXON_BIN="$AXON_REPO_DIR/bin/axon"
    echo "   ✅ Axon found at $AXON_BIN"
elif [ -f "$AXON_REPO_DIR/axon" ]; then
    AXON_BIN="$AXON_REPO_DIR/axon"
    echo "   ✅ Axon found at $AXON_BIN"
else
    echo "   ⚠️  Axon not found. Building..."
    if [ -d "$AXON_REPO_DIR" ]; then
        cd "$AXON_REPO_DIR"
        make build
        AXON_BIN="$AXON_REPO_DIR/bin/axon"
        echo "   ✅ Axon built successfully"
    else
        echo "   ❌ Axon repository not found at $AXON_REPO_DIR"
        echo "   Please install Axon first or update AXON_REPO_DIR"
        exit 1
    fi
fi

# Step 2: Initialize Axon
echo ""
echo "2️⃣  Initializing Axon..."
$AXON_BIN init || echo "   (Already initialized)"

# Step 3: Create registry structure
echo ""
echo "3️⃣  Setting up registry structure..."
mkdir -p "$REGISTRY_DIR/api/v1/models"
mkdir -p "$REGISTRY_DIR/packages"
mkdir -p "$REGISTRY_DIR/tmp"

# Step 4: Generate manifests for top 100 models
echo ""
echo "4️⃣  Generating manifests for top 100 Hugging Face models..."
cd "$REGISTRY_DIR"

if [ ! -f "generate-top-100-manifests.go" ]; then
    echo "   ❌ generate-top-100-manifests.go not found"
    echo "   Please ensure the manifest generator exists"
    exit 1
fi

go run generate-top-100-manifests.go "$REGISTRY_DIR" || {
    echo "   ⚠️  Manifest generation failed. Using fallback..."
    # Fallback: Generate manifests for top 10 models
    if [ -f "create-manifests.go" ]; then
        go run create-manifests.go "$REGISTRY_DIR"
    fi
}

# Step 5: Create placeholder packages
echo ""
echo "5️⃣  Creating placeholder package files..."
cd "$REGISTRY_DIR/packages"
PACKAGE_COUNT=$(find "$REGISTRY_DIR/api/v1/models" -name "manifest.yaml" | wc -l | tr -d ' ')
echo "   Creating $PACKAGE_COUNT placeholder packages..."

find "$REGISTRY_DIR/api/v1/models" -name "manifest.yaml" | while read manifest; do
    # Extract namespace, name, version from path
    rel_path=$(echo "$manifest" | sed "s|$REGISTRY_DIR/api/v1/models/||")
    namespace=$(echo "$rel_path" | cut -d'/' -f1)
    name=$(echo "$rel_path" | cut -d'/' -f2)
    version=$(echo "$rel_path" | cut -d'/' -f3)
    
    package_file="${namespace}-${name}-${version}.axon"
    if [ ! -f "$package_file" ]; then
        echo "Placeholder package for ${namespace}/${name}@${version}" > "$package_file"
    fi
done

# Step 6: Update checksums in manifests
echo ""
echo "6️⃣  Computing and updating checksums..."
cd "$REGISTRY_DIR"
if [ -f "update-checksums.go" ]; then
    go run update-checksums.go "$REGISTRY_DIR" || echo "   ⚠️  Checksum update skipped"
else
    # Simple checksum update using existing script
    if [ -f "create-manifests.go" ]; then
        go run create-manifests.go "$REGISTRY_DIR" 2>/dev/null || true
    fi
fi

# Step 7: Configure Axon to use local registry
echo ""
echo "7️⃣  Configuring Axon to use local registry..."
$AXON_BIN registry set default "http://localhost:8080" || echo "   (Already configured)"

# Step 8: Start registry server (optional)
echo ""
echo "8️⃣  Registry setup complete!"
echo ""
echo "📊 Statistics:"
MODEL_COUNT=$(find "$REGISTRY_DIR/api/v1/models" -name "manifest.yaml" | wc -l | tr -d ' ')
PACKAGE_COUNT=$(ls -1 "$REGISTRY_DIR/packages"/*.axon 2>/dev/null | wc -l | tr -d ' ')
echo "   • Models: $MODEL_COUNT"
echo "   • Packages: $PACKAGE_COUNT"
echo ""
echo "🚀 Next steps:"
echo ""
echo "1. Start the registry server:"
echo "   cd $REGISTRY_DIR"
echo "   go run server.go ."
echo ""
echo "2. Test with Axon:"
echo "   $AXON_BIN search bert"
echo "   $AXON_BIN info nlp/bert-base-uncased@1.0.0"
echo ""
echo "3. Browse models in browser:"
echo "   http://localhost:8080"
echo ""
echo "📦 This registry can be deployed as a hosted model repository!"
echo "   All manifests are in: $REGISTRY_DIR/api/v1/models/"
echo "   All packages are in: $REGISTRY_DIR/packages/"
echo ""

