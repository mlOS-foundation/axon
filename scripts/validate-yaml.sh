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

# Check for YAML validation tools (try multiple methods)
YAML_AVAILABLE=false
YAML_TOOL=""

# Method 1: Try yamllint (preferred, no Python dependencies)
if command -v yamllint >/dev/null 2>&1; then
    YAML_AVAILABLE=true
    YAML_TOOL="yamllint"
    echo -e "${GREEN}✅ Using yamllint${NC}"
# Method 2: Try Python yaml module
elif python3 -c "import yaml" 2>/dev/null; then
    YAML_AVAILABLE=true
    YAML_TOOL="python-yaml"
    echo -e "${GREEN}✅ Using Python PyYAML${NC}"
# Method 3: Try to install yamllint via brew (macOS)
elif command -v brew >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  YAML validator not found${NC}"
    echo "Attempting to install yamllint via Homebrew..."
    if brew install yamllint >/dev/null 2>&1; then
        YAML_AVAILABLE=true
        YAML_TOOL="yamllint"
        echo -e "${GREEN}✅ yamllint installed via Homebrew${NC}"
    else
        echo -e "${YELLOW}⚠️  Homebrew installation failed or requires confirmation${NC}"
        echo ""
        echo "Please install yamllint manually:"
        echo "  brew install yamllint"
        echo ""
        echo "Or install PyYAML for Python:"
        echo "  pip3 install --user pyyaml"
        echo ""
        echo "After installation, run 'make validate-yaml' again."
        exit 1
    fi
# Method 4: Try to install PyYAML (last resort)
else
    echo -e "${YELLOW}⚠️  YAML validator not found${NC}"
    echo "Attempting to install PyYAML..."
    
    # Try user install
    if python3 -m pip install --user pyyaml >/dev/null 2>&1; then
        YAML_AVAILABLE=true
        YAML_TOOL="python-yaml"
        echo -e "${GREEN}✅ PyYAML installed${NC}"
    else
        echo -e "${RED}❌ Failed to install YAML validator${NC}"
        echo ""
        echo "Please install one of the following:"
        echo ""
        echo "Option 1 (recommended): Install yamllint"
        echo "  brew install yamllint"
        echo ""
        echo "Option 2: Install PyYAML for Python"
        echo "  pip3 install --user pyyaml"
        echo ""
        echo "After installation, run 'make validate-yaml' again."
        exit 1
    fi
fi

# Find all YAML files (exclude .git directory but include .github directory)
YAML_FILES=$(find "$REPO_ROOT" -name "*.yml" -o -name "*.yaml" | grep -v node_modules | grep -v "/\.git/" | sort)

ERRORS=0

for file in $YAML_FILES; do
    rel_path="${file#$REPO_ROOT/}"
    echo -n "Checking $rel_path... "
    
    # GitHub Actions workflow files need special handling (they use ${{ }} expressions)
    IS_WORKFLOW=false
    if echo "$file" | grep -q "\.github/workflows"; then
        IS_WORKFLOW=true
    fi
    
    # Validate YAML syntax based on available tool
    if [ "$YAML_TOOL" = "yamllint" ]; then
        if [ "$IS_WORKFLOW" = true ]; then
            # For workflow files, yamllint has issues with ${{ }} expressions
            # Use Python yaml module instead (it doesn't parse expressions)
            # Check if PyYAML is available, install if needed
            if ! python3 -c "import yaml" 2>/dev/null; then
                echo ""
                echo -e "${YELLOW}⚠️  PyYAML needed for workflow validation${NC}"
                # Try multiple installation methods
                INSTALLED=false
                # Method 1: Try --user install
                if python3 -m pip install --user pyyaml >/dev/null 2>&1; then
                    INSTALLED=true
                # Method 2: Try with --break-system-packages (macOS externally-managed Python)
                elif python3 -m pip install --user --break-system-packages pyyaml >/dev/null 2>&1; then
                    INSTALLED=true
                # Method 3: Try system-wide (may require sudo, but won't fail silently)
                elif python3 -m pip install pyyaml >/dev/null 2>&1; then
                    INSTALLED=true
                fi
                
                if [ "$INSTALLED" = true ]; then
                    echo -e "${GREEN}✅ PyYAML installed${NC}"
                else
                    echo -e "${RED}❌ Failed to install PyYAML${NC}"
                    echo ""
                    echo "Please install PyYAML manually using one of:"
                    echo "  pip3 install --user --break-system-packages pyyaml"
                    echo "  brew install libyaml && pip3 install --user pyyaml"
                    echo ""
                    echo "Or install yamllint (preferred, no Python dependencies):"
                    echo "  brew install yamllint"
                    echo ""
                    echo "After installation, run 'make validate-yaml' again."
                    exit 1
                fi
            fi
            if python3 << EOF 2>/dev/null; then
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
        else
            # Use yamllint for non-workflow files
            if yamllint -d relaxed "$file" >/dev/null 2>&1; then
                echo -e "${GREEN}✅${NC}"
            else
                echo -e "${RED}❌${NC}"
                yamllint -d relaxed "$file" 2>&1 | head -5
                ERRORS=$((ERRORS + 1))
            fi
        fi
    else
        # Use Python yaml module
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

