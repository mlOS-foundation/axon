#!/bin/bash

# Test script for Axon commands with local registry

set -e

REGISTRY_DIR="$(cd "$(dirname "$0")" && pwd)"
REGISTRY_URL="http://localhost:8080"

echo "🧪 Testing Axon Commands with Local Registry"
echo "============================================"
echo ""

# Check if axon binary exists
AXON_BIN=""
if [ -f "../../bin/axon" ]; then
    AXON_BIN="../../bin/axon"
elif [ -f "../../axon" ]; then
    AXON_BIN="../../axon"
elif command -v axon &> /dev/null; then
    AXON_BIN="axon"
else
    echo "❌ Error: axon binary not found"
    echo "   Build it first: cd ../../ && make build"
    exit 1
fi

echo "Using axon binary: $AXON_BIN"
echo ""

# Initialize axon
echo "1️⃣  Initializing Axon..."
$AXON_BIN init || echo "   (init may have already been done)"
echo ""

# Configure registry
echo "2️⃣  Configuring registry to use local server..."
$AXON_BIN registry set default "$REGISTRY_URL" || echo "   (registry may already be configured)"
echo ""

# Test search
echo "3️⃣  Testing search command..."
echo "   Searching for 'bert'..."
$AXON_BIN search bert || echo "   ⚠️  Search failed (this is expected if registry is not running)"
echo ""

echo "4️⃣  Testing info command..."
echo "   Getting info for nlp/bert-base-uncased..."
$AXON_BIN info nlp/bert-base-uncased || echo "   ⚠️  Info failed (this is expected if registry is not running)"
echo ""

echo "5️⃣  Testing list command..."
$AXON_BIN list
echo ""

echo "✅ Test script completed!"
echo ""
echo "To start the registry server, run:"
echo "   cd $REGISTRY_DIR"
echo "   go run server.go ."
echo ""
echo "Then run this script again to test with the live server."

