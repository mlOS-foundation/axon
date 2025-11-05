#!/bin/bash

echo "🧪 Axon Local Registry Tester"
echo "=============================="
echo ""

# Check if server is running
if curl -s http://localhost:8080/api/v1/search?q=test > /dev/null 2>&1; then
    echo "✅ Registry server is running on http://localhost:8080"
    echo ""
    echo "🌐 Web UI: http://localhost:8080"
    echo "🔍 Test search: curl 'http://localhost:8080/api/v1/search?q=bert'"
    echo ""
else
    echo "❌ Registry server is NOT running"
    echo ""
    echo "Start it with:"
    echo "  cd test/registry"
    echo "  go run server.go ."
    echo ""
    exit 1
fi

# Check if axon is configured
if [ -f ~/.axon/config.yaml ]; then
    echo "✅ Axon is configured"
    REGISTRY_URL=$(grep -A 1 "registry:" ~/.axon/config.yaml | grep "url:" | awk '{print $2}')
    echo "   Registry URL: $REGISTRY_URL"
    echo ""
else
    echo "⚠️  Axon not configured. Run:"
    echo "   ./bin/axon init"
    echo "   ./bin/axon registry set default http://localhost:8080"
    echo ""
fi

echo "📋 Quick Test Commands:"
echo "   ./bin/axon search bert"
echo "   ./bin/axon info nlp/gpt2@1.0.0"
echo "   ./bin/axon install nlp/gpt2@1.0.0"
echo ""
