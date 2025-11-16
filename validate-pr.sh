#!/bin/bash
# validate-pr.sh - Run all validations before creating a PR
# This script runs the same checks as CI to catch issues early

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track if any check fails
FAILED=0

echo "🔍 Running PR validation checks..."
echo ""

# Function to print success
success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print error
error() {
    echo -e "${RED}❌ $1${NC}"
    FAILED=1
}

# Function to print warning
warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 1. Check Go version
echo "1️⃣  Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}')
echo "   Go version: $GO_VERSION"
if [[ "$GO_VERSION" == go1.21* ]] || [[ "$GO_VERSION" == go1.22* ]] || [[ "$GO_VERSION" == go1.23* ]]; then
    success "Go version check"
else
    warning "Go version $GO_VERSION may not match CI (expects 1.21+)"
fi
echo ""

# 2. Download dependencies
echo "2️⃣  Downloading dependencies..."
if go mod download; then
    success "Dependencies downloaded"
else
    error "Failed to download dependencies"
    exit 1
fi
echo ""

# 3. Tidy go.mod
echo "3️⃣  Tidying go.mod..."
if go mod tidy; then
    success "go.mod tidied"
else
    error "Failed to tidy go.mod"
    exit 1
fi

# Check if go.mod or go.sum changed
if ! git diff --quiet go.mod go.sum 2>/dev/null; then
    warning "go.mod or go.sum has uncommitted changes"
    echo "   Run: git add go.mod go.sum && git commit -m 'chore: update dependencies'"
    FAILED=1
fi
echo ""

# 4. Check code formatting
echo "4️⃣  Checking code formatting..."
UNFORMATTED=$(gofmt -s -l . | grep -v vendor | head -20)
if [ -z "$UNFORMATTED" ]; then
    success "Code formatting: OK"
else
    error "Code is not formatted. Run 'go fmt ./...' or 'make fmt'"
    echo "$UNFORMATTED" | while read -r line; do
        echo "   $line"
    done
    FAILED=1
fi
echo ""

# 5. Run go vet
echo "5️⃣  Running go vet..."
if go vet ./...; then
    success "go vet: OK"
else
    error "go vet failed"
    FAILED=1
fi
echo ""

# 6. Run golangci-lint
echo "6️⃣  Running golangci-lint..."
# Ensure GOPATH/bin is in PATH
GOPATH_BIN="$(go env GOPATH)/bin"
export PATH="${GOPATH_BIN}:${PATH}"

if ! command -v golangci-lint > /dev/null; then
    warning "golangci-lint not found. Installing..."
    if go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8; then
        success "golangci-lint installed"
        # Re-export PATH after installation
        export PATH="${GOPATH_BIN}:${PATH}"
    else
        error "Failed to install golangci-lint"
        echo "   Run: make install-tools"
        FAILED=1
        echo ""
    fi
fi

# Try to run golangci-lint if available
if command -v golangci-lint > /dev/null; then
    if golangci-lint run --timeout=5m ./...; then
        success "golangci-lint: OK"
    else
        error "golangci-lint failed"
        FAILED=1
    fi
else
    warning "golangci-lint not available - skipping (install with: make install-tools)"
fi
echo ""

# 7. Run tests
echo "7️⃣  Running tests..."
if go test -v -coverprofile=coverage.out ./...; then
    success "Tests: OK"
    
    # Check coverage if file exists
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo "   Coverage: $COVERAGE"
    fi
else
    error "Tests failed"
    FAILED=1
fi
echo ""

# 8. Build
echo "8️⃣  Building binary..."
if go build -o /tmp/axon ./cmd/axon; then
    success "Build: OK"
    rm -f /tmp/axon
else
    error "Build failed"
    FAILED=1
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All PR validation checks passed!${NC}"
    echo ""
    echo "You can now create your PR with confidence."
    exit 0
else
    echo -e "${RED}❌ Some PR validation checks failed${NC}"
    echo ""
    echo "Please fix the issues above before creating a PR."
    echo ""
    echo "Quick fixes:"
    echo "  - Format code: make fmt"
    echo "  - Run tests: make test"
    echo "  - Run all checks: make ci"
    exit 1
fi
