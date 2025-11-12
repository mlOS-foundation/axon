# Contributing to Axon

Thank you for contributing to Axon! This document provides guidelines and instructions for contributing.

## Development Workflow

### Before Pushing PR Updates

**⚠️ CRITICAL: Always run local validation before marking a PR as ready for review!**

Always run validation checks before pushing PR updates:

```bash
# Option 1: Use the validation script
./validate-pr.sh

# Option 2: Use Makefile target
make validate-pr

# Option 3: Run checks individually
make fmt-check  # Check formatting
make vet        # Run go vet
make lint       # Run linters
make test       # Run tests
make build      # Build binary
```

### Validation Checks

The validation script runs:
1. **Code Formatting** (`go fmt`) - Ensures code follows Go formatting standards
2. **go vet** - Static analysis for common mistakes
3. **golangci-lint** - Comprehensive linting (if installed)
4. **Tests** - Runs all unit tests with coverage
5. **Build** - Verifies the code compiles successfully

### Quick Commands

```bash
# Format code
make fmt

# Run tests
make test

# Run all CI checks
make ci

# Validate before PR
make validate-pr
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

4. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   gh pr create --title "feat: Add your feature" --body "Description" --draft
   ```

5. **Run validation again** after any fixes
   ```bash
   make validate-pr
   ```

6. **Mark PR as ready** only after all local validation passes
   ```bash
   gh pr ready 18  # Only after validation passes!
   ```

7. **Wait for CI** - GitHub Actions will run the same checks

8. **Address feedback** - If CI fails, fix issues locally, validate, and push again

**Never mark a PR as ready for review if local validation fails!**

## Code Standards

- **Formatting**: Code must be formatted with `go fmt`
- **Linting**: Must pass `golangci-lint` checks
- **Testing**: New features should include tests
- **Documentation**: Update README/docs as needed

## Getting Help

- Open an issue for questions or bugs
- Check existing issues before creating new ones
- Be respectful and constructive in discussions

Thank you for contributing! 🚀
