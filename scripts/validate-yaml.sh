#!/bin/bash
# Validate YAML files using Python's yaml module
# This ensures we catch YAML syntax errors locally before pushing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔍 YAML Validation"
echo "=================="
echo ""

# Check if pyyaml is available (try multiple methods)
PYTHON_CMD="python3"
YAML_AVAILABLE=false

# Try to import yaml
if python3 -c "import yaml" 2>/dev/null; then
    YAML_AVAILABLE=true
elif python3 -m pip show pyyaml >/dev/null 2>&1; then
    YAML_AVAILABLE=true
else
    echo -e "${YELLOW}⚠️  PyYAML not found${NC}"
    echo "Attempting to install PyYAML..."
    
    # Try user install first
    if python3 -m pip install --user pyyaml >/dev/null 2>&1; then
        YAML_AVAILABLE=true
        echo -e "${GREEN}✅ PyYAML installed${NC}"
    else
        echo -e "${RED}❌ Failed to install PyYAML${NC}"
        echo "Please install manually: pip3 install --user pyyaml"
        echo "Or use: brew install libyaml && pip3 install --user pyyaml"
        exit 1
    fi
fi

# Find all YAML files
YAML_FILES=$(find "$REPO_ROOT" -name "*.yml" -o -name "*.yaml" | grep -v node_modules | grep -v ".git" | sort)

ERRORS=0

for file in $YAML_FILES; do
    rel_path="${file#$REPO_ROOT/}"
    echo -n "Checking $rel_path... "
    
    # Validate YAML syntax
    if python3 << EOF 2>&1 | grep -q "✅"; then
import yaml
import sys
try:
    with open('$file', 'r') as f:
        data = f.read()
        # Check for trailing newline (GitHub Actions requirement)
        if not data.endswith('\n'):
            print(f"⚠️  Missing trailing newline", file=sys.stderr)
        yaml.safe_load(data)
    print("✅")
    sys.exit(0)
except yaml.YAMLError as e:
    if hasattr(e, 'problem_mark'):
        mark = e.problem_mark
        print(f"❌ YAML Error at line {mark.line + 1}, column {mark.column + 1}: {e}", file=sys.stderr)
    else:
        print(f"❌ YAML Error: {e}", file=sys.stderr)
    sys.exit(1)
except Exception as e:
    print(f"❌ Error: {e}", file=sys.stderr)
    sys.exit(1)
EOF
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
        python3 << EOF 2>&1 | head -3
import yaml
import sys
try:
    with open('$file', 'r') as f:
        data = f.read()
        if not data.endswith('\n'):
            print(f"⚠️  Missing trailing newline")
        yaml.safe_load(data)
except yaml.YAMLError as e:
    if hasattr(e, 'problem_mark'):
        mark = e.problem_mark
        print(f"Line {mark.line + 1}, column {mark.column + 1}: {e}")
    else:
        print(f"YAML Error: {e}")
except Exception as e:
    print(f"Error: {e}")
EOF
        ERRORS=$((ERRORS + 1))
    fi
done

echo ""
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All YAML files are valid${NC}"
    exit 0
else
    echo -e "${RED}❌ Found $ERRORS YAML file(s) with errors${NC}"
    exit 1
fi

