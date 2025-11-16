# Contributing to Axon

Thank you for contributing to Axon! This document provides guidelines and instructions for contributing.

## Development Workflow

### Before Creating a PR

**⚠️ CRITICAL: Always run local validation before creating a PR!**

We provide automated tools to ensure your code passes all CI checks before creating a PR:

```bash
# Option 1: Use the automated PR creation script (RECOMMENDED)
./scripts/create-pr.sh --title "feat: Add feature" --body "Description"

# Option 2: Use Makefile target
make create-pr TITLE="feat: Add feature" BODY="Description"

# Option 3: Run validation manually, then create PR
make validate-pr
gh pr create --title "feat: Add feature" --body "Description"

# Option 4: Run checks individually
make fmt-check  # Check formatting
make vet        # Run go vet
make lint       # Run linters
make test       # Run tests
make build      # Build binary
```

### Validation Checks

The validation script (`validate-pr.sh`) runs the **exact same checks as CI**:

1. **Go Version Check** - Verifies Go version compatibility
2. **Dependencies** - Downloads and tidies dependencies
3. **Code Formatting** (`go fmt`) - Ensures code follows Go formatting standards
4. **go vet** - Static analysis for common mistakes
5. **golangci-lint** - Comprehensive linting (same version as CI: v1.64.8)
6. **Tests** - Runs all unit tests with coverage
7. **Build** - Verifies the code compiles successfully

### Quick Commands

```bash
# Format code
make fmt

# Run tests
make test

# Run all CI checks
make ci

# Validate before PR (runs all checks)
make validate-pr

# Create PR with automatic validation
make create-pr TITLE="feat: Add feature" BODY="Description"
```

## Pull Request Process

1. **Create a branch** from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit
   ```bash
   git add .
   git commit -m "feat: Add your feature"
   ```

3. **Run validation** before pushing
   ```bash
   make validate-pr
   ```
   **⚠️ Ensure all checks pass before proceeding!**

4. **Create PR with validation** (RECOMMENDED)
   ```bash
   # This automatically runs validation first
   ./scripts/create-pr.sh --title "feat: Add feature" --body "Description"
   ```
   
   Or manually:
   ```bash
   git push origin feature/your-feature-name
   gh pr create --title "feat: Add feature" --body "Description" --draft
   ```

5. **Run validation again** after any fixes
   ```bash
   make validate-pr
   ```

6. **Mark PR as ready** only after all local validation passes
   ```bash
   gh pr ready <PR_NUMBER>  # Only after validation passes!
   ```

7. **Wait for CI** - GitHub Actions will run the same checks

8. **Address feedback** - If CI fails, fix issues locally, validate, and push again

## Validation Script Details

The `validate-pr.sh` script:
- Runs the **exact same checks** as `.github/workflows/pr-validation.yml`
- Uses the **same golangci-lint version** as CI (v1.64.8)
- Provides **clear error messages** with fix suggestions
- **Exits with error code** if any check fails (prevents PR creation)

## Troubleshooting

### Validation fails with "golangci-lint not found"
```bash
make install-tools  # Installs golangci-lint and goimports
```

### Validation fails with "Code is not formatted"
```bash
make fmt  # Auto-formats code
```

### Validation fails with "go.mod or go.sum has uncommitted changes"
```bash
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

### Skip validation (NOT RECOMMENDED)
```bash
./scripts/create-pr.sh --skip-validation --title "..." --body "..."
```

## Best Practices

1. **Always run `make validate-pr` before creating/updating PRs**
2. **Fix formatting issues** with `make fmt` before committing
3. **Run tests locally** before pushing
4. **Keep PRs focused** - one feature/fix per PR
5. **Write clear commit messages** following conventional commits
6. **Update documentation** if adding new features

## CI/CD

All PRs automatically run:
- Code formatting check
- go vet
- golangci-lint (v1.64.8)
- Tests with coverage
- Build verification

**PRs will be blocked from merging if any CI check fails.**
