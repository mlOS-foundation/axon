#!/bin/bash
#
# Pre-PR Validation Script
# Runs lint, vet, and tests before pushing PR updates
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track failures
FAILED=0

echo "=========================================="
echo "Pre-PR Validation"
echo "=========================================="
echo ""

# 1. Format check
echo -e "${BLUE}1. Checking code formatting (go fmt)...${NC}"
if [ -n "$(gofmt -s -l . | grep -v vendor | head -20)" ]; then
    echo -e "${RED}❌ Code is not formatted. Run 'go fmt ./...'${NC}"
    gofmt -s -d . | head -50
    FAILED=1
else
    echo -e "${GREEN}✅ Code formatting: OK${NC}"
fi
echo ""

# 2. Run go fmt to auto-fix
echo -e "${BLUE}2. Running go fmt to auto-fix formatting...${NC}"
go fmt ./...
echo -e "${GREEN}✅ Formatting applied${NC}"
echo ""

# 3. Vet check
echo -e "${BLUE}3. Running go vet...${NC}"
if ! go vet ./...; then
    echo -e "${RED}❌ go vet failed${NC}"
    FAILED=1
else
    echo -e "${GREEN}✅ go vet: OK${NC}"
fi
echo ""

# 4. Lint check
echo -e "${BLUE}4. Running golangci-lint...${NC}"
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --timeout=5m; then
        echo -e "${RED}❌ golangci-lint failed${NC}"
        FAILED=1
    else
        echo -e "${GREEN}✅ golangci-lint: OK${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  golangci-lint not installed, skipping${NC}"
    echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi
echo ""

# 5. Run tests
echo -e "${BLUE}5. Running tests...${NC}"
if ! go test -v -coverprofile=coverage.out ./...; then
    echo -e "${RED}❌ Tests failed${NC}"
    FAILED=1
else
    echo -e "${GREEN}✅ Tests: OK${NC}"
    
    # Show coverage if available
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo -e "${BLUE}   Coverage: ${COVERAGE}${NC}"
    fi
fi
echo ""

# 6. Build check
echo -e "${BLUE}6. Building binary...${NC}"
if ! go build -o /tmp/axon-test ./cmd/axon; then
    echo -e "${RED}❌ Build failed${NC}"
    FAILED=1
else
    echo -e "${GREEN}✅ Build: OK${NC}"
    rm -f /tmp/axon-test
fi
echo ""

# Summary
echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! Ready to push PR.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some checks failed. Please fix issues before pushing PR.${NC}"
    exit 1
fi

