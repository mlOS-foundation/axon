#!/bin/bash
# Quick validation script for docker-converter.yml workflow
# Checks configuration without building the full image

set -e

echo "🔍 Quick Docker Workflow Validation"
echo "===================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0

# Check Dockerfile exists
echo "1️⃣  Checking Dockerfile..."
if [ -f "docker/Dockerfile.converter" ]; then
    echo -e "${GREEN}✅ Dockerfile exists${NC}"
else
    echo -e "${RED}❌ Dockerfile not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check conversion scripts
echo ""
echo "2️⃣  Checking conversion scripts..."
if [ -d "scripts/conversion" ]; then
    SCRIPT_COUNT=$(find scripts/conversion -name "*.py" | wc -l | tr -d ' ')
    echo -e "${GREEN}✅ Found $SCRIPT_COUNT conversion script(s)${NC}"
    find scripts/conversion -name "*.py" | while read script; do
        echo "   - $script"
    done
else
    echo -e "${RED}❌ Conversion scripts directory not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check workflow file
echo ""
echo "3️⃣  Checking workflow file..."
if [ -f ".github/workflows/docker-converter.yml" ]; then
    echo -e "${GREEN}✅ Workflow file exists${NC}"
    
    # Check for lowercase image name
    if grep -q "ghcr.io/mlos-foundation/axon-converter" .github/workflows/docker-converter.yml; then
        echo -e "${GREEN}✅ Image name is lowercase (correct)${NC}"
    elif grep -q "ghcr.io/mlOS-foundation/axon-converter" .github/workflows/docker-converter.yml; then
        echo -e "${RED}❌ Image name contains uppercase (mlOS) - should be mlos${NC}"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "${YELLOW}⚠️  Could not verify image name format${NC}"
    fi
else
    echo -e "${RED}❌ Workflow file not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

# Check code references
echo ""
echo "4️⃣  Checking code references..."
if grep -r "ghcr.io/mlos-foundation/axon-converter" internal/converter/docker.go > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Code uses lowercase image name${NC}"
elif grep -r "ghcr.io/mlOS-foundation/axon-converter" internal/converter/docker.go > /dev/null 2>&1; then
    echo -e "${RED}❌ Code uses uppercase image name${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${YELLOW}⚠️  Could not verify code image name${NC}"
fi

# Summary
echo ""
echo "===================================="
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All quick checks passed!${NC}"
    echo ""
    echo "To fully validate (builds image - takes 5-10 minutes):"
    echo "  ./scripts/validate-docker-workflow.sh"
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS error(s)${NC}"
    exit 1
fi

