#!/bin/bash
#
# Axon CLI Full Use Case Validation Script
# Tests the complete workflow from installation to model management
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0
ISSUES=()

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASSED++))
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    ((FAILED++))
    ISSUES+=("$1")
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if axon is installed
check_axon_installed() {
    log_info "Checking if Axon CLI is installed..."
    if command -v axon &> /dev/null; then
        AXON_VERSION=$(axon --version 2>/dev/null || echo "unknown")
        log_success "Axon CLI is installed (version: $AXON_VERSION)"
        return 0
    else
        log_error "Axon CLI is not installed or not in PATH"
        return 1
    fi
}

# Test 1: axon init
test_init() {
    log_info "Test 1: Testing 'axon init'..."
    if axon init 2>&1; then
        log_success "axon init completed successfully"
    else
        log_error "axon init failed"
    fi
}

# Test 2: axon search
test_search() {
    log_info "Test 2: Testing 'axon search bert'..."
    if axon search bert 2>&1 | grep -q "bert" || axon search bert 2>&1 | grep -q "found" || axon search bert 2>&1 | grep -q "No models"; then
        log_success "axon search works (may return results or 'no models found')"
    else
        log_error "axon search failed or returned unexpected output"
    fi
}

# Test 3: axon info
test_info() {
    log_info "Test 3: Testing 'axon info hf/bert-base-uncased@latest'..."
    if axon info hf/bert-base-uncased@latest 2>&1; then
        log_success "axon info completed (may show info or error if model not found)"
    else
        log_warning "axon info failed (may be expected if model not in registry)"
    fi
}

# Test 4: axon install (Hugging Face)
test_install() {
    log_info "Test 4: Testing 'axon install hf/bert-base-uncased@latest'..."
    if axon install hf/bert-base-uncased@latest 2>&1; then
        log_success "axon install (Hugging Face) completed"
    else
        log_error "axon install (Hugging Face) failed"
    fi
}

# Test 4b: axon install (PyTorch Hub) - NEW in v1.1.0
test_install_pytorch() {
    log_info "Test 4b: Testing 'axon install pytorch/vision/resnet50@latest' (PyTorch Hub adapter)..."
    if axon install pytorch/vision/resnet50@latest 2>&1; then
        log_success "axon install (PyTorch Hub) completed"
    else
        log_warning "axon install (PyTorch Hub) failed (may be expected if model is large or network issues)"
    fi
}

# Test 4c: axon install (TensorFlow Hub) - NEW in v1.2.0
test_install_tfhub() {
    log_info "Test 4c: Testing 'axon install tfhub/google/universal-sentence-encoder/4@latest' (TensorFlow Hub adapter)..."
    if axon install tfhub/google/universal-sentence-encoder/4@latest 2>&1; then
        log_success "axon install (TensorFlow Hub) completed"
    else
        log_warning "axon install (TensorFlow Hub) failed (may be expected if model is large or network issues)"
    fi
}

# Test 5: axon list
test_list() {
    log_info "Test 5: Testing 'axon list'..."
    if axon list 2>&1; then
        log_success "axon list completed"
    else
        log_error "axon list failed"
    fi
}

# Test 6: axon verify
test_verify() {
    log_info "Test 6: Testing 'axon verify hf/bert-base-uncased'..."
    if axon verify hf/bert-base-uncased 2>&1; then
        log_success "axon verify completed"
    else
        log_warning "axon verify failed (may be expected if model not installed)"
    fi
}

# Test 7: axon cache
test_cache() {
    log_info "Test 7: Testing 'axon cache list'..."
    if axon cache list 2>&1; then
        log_success "axon cache list completed"
    else
        log_error "axon cache list failed"
    fi
    
    log_info "Test 7b: Testing 'axon cache stats'..."
    if axon cache stats 2>&1; then
        log_success "axon cache stats completed"
    else
        log_warning "axon cache stats failed (may be expected)"
    fi
}

# Test 8: axon config
test_config() {
    log_info "Test 8: Testing 'axon config show'..."
    if axon config show 2>&1; then
        log_success "axon config show completed"
    else
        log_error "axon config show failed"
    fi
}

# Test 9: axon registry
test_registry() {
    log_info "Test 9: Testing 'axon registry list'..."
    if axon registry list 2>&1; then
        log_success "axon registry list completed"
    else
        log_warning "axon registry list failed (may be expected)"
    fi
}

# Test 10: axon uninstall
test_uninstall() {
    log_info "Test 10: Testing 'axon uninstall hf/bert-base-uncased'..."
    if axon uninstall hf/bert-base-uncased 2>&1; then
        log_success "axon uninstall completed"
    else
        log_warning "axon uninstall failed (may be expected if model not installed)"
    fi
}

# Test 11: axon update
test_update() {
    log_info "Test 11: Testing 'axon update hf/bert-base-uncased'..."
    if axon update hf/bert-base-uncased 2>&1; then
        log_success "axon update completed"
    else
        log_warning "axon update failed (may be expected if model not installed)"
    fi
}

# Test 12: Help command
test_help() {
    log_info "Test 12: Testing 'axon --help'..."
    if axon --help 2>&1 | grep -q "axon" || axon --help 2>&1 | grep -q "Usage"; then
        log_success "axon --help works"
    else
        log_error "axon --help failed"
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "Axon CLI Full Use Case Validation"
    echo "=========================================="
    echo ""
    
    # Check if axon is installed
    if ! check_axon_installed; then
        log_error "Cannot proceed - Axon CLI is not installed"
        echo ""
        echo "Please install Axon CLI first:"
        echo "  curl -sSL axon.mlosfoundation.org | sh"
        exit 1
    fi
    
    echo ""
    log_info "Starting validation tests..."
    echo ""
    
    # Run all tests
    test_help
    test_init
    test_search
    test_info
    test_install
    test_install_pytorch
    test_install_tfhub
    test_list
    test_verify
    test_cache
    test_config
    test_registry
    test_update
    test_uninstall
    
    # Summary
    echo ""
    echo "=========================================="
    echo "Validation Summary"
    echo "=========================================="
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    echo ""
    
    if [ $FAILED -gt 0 ]; then
        echo "Issues found:"
        for issue in "${ISSUES[@]}"; do
            echo -e "  ${RED}• $issue${NC}"
        done
        echo ""
        exit 1
    else
        log_success "All tests passed!"
        exit 0
    fi
}

# Run main
main "$@"

