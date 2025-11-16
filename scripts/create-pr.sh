#!/bin/bash
# create-pr.sh - Create a PR with automatic validation
# Usage: ./scripts/create-pr.sh [--title "Title"] [--body "Body"] [--base main] [--head branch] [--draft] [--skip-validation]

set -e

# Default values
TITLE=""
BODY=""
BASE="main"
HEAD=""
DRAFT=false
SKIP_VALIDATION=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --title)
            TITLE="$2"
            shift 2
            ;;
        --body)
            BODY="$2"
            shift 2
            ;;
        --base)
            BASE="$2"
            shift 2
            ;;
        --head)
            HEAD="$2"
            shift 2
            ;;
        --draft)
            DRAFT=true
            shift
            ;;
        --skip-validation)
            SKIP_VALIDATION=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--title \"Title\"] [--body \"Body\"] [--base main] [--head branch] [--draft] [--skip-validation]"
            exit 1
            ;;
    esac
done

# Get current branch if HEAD not specified
if [ -z "$HEAD" ]; then
    HEAD=$(git branch --show-current)
    if [ -z "$HEAD" ]; then
        echo "❌ Error: Could not determine current branch. Please specify --head"
        exit 1
    fi
fi

echo "🔍 Creating PR from '$HEAD' to '$BASE'..."
echo ""

# Run validation unless skipped
if [ "$SKIP_VALIDATION" = false ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🔍 Running PR validation checks..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Run validation script
    if ! ./validate-pr.sh; then
        echo ""
        echo "❌ Validation failed! Please fix the issues before creating a PR."
        echo ""
        echo "To skip validation (not recommended):"
        echo "  $0 --skip-validation [other options...]"
        exit 1
    fi
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ All validations passed! Proceeding with PR creation..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
fi

# Build gh pr create command
PR_CMD="gh pr create --base $BASE --head $HEAD"

if [ -n "$TITLE" ]; then
    PR_CMD="$PR_CMD --title \"$TITLE\""
fi

if [ -n "$BODY" ]; then
    PR_CMD="$PR_CMD --body \"$BODY\""
fi

if [ "$DRAFT" = true ]; then
    PR_CMD="$PR_CMD --draft"
fi

# Create PR
echo "📝 Creating PR..."
eval $PR_CMD

# Get PR number
PR_NUMBER=$(gh pr list --head "$HEAD" --base "$BASE" --json number --jq '.[0].number')

if [ -n "$PR_NUMBER" ]; then
    echo ""
    echo "✅ PR #$PR_NUMBER created successfully!"
    echo "🔗 https://github.com/$(gh repo view --json owner,name -q '.owner.login + "/" + .name')/pull/$PR_NUMBER"
    echo ""
    echo "CI workflows will now run. Monitor with:"
    echo "  gh pr checks $PR_NUMBER"
else
    echo "⚠️  PR created but could not determine PR number"
fi

