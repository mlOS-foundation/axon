# PR Validation Guide

## Overview

All PRs must pass CI checks before merging. To avoid CI failures, **always test locally before pushing**.

## Quick Validation

Run all CI checks locally:

```bash
make ci
```

This runs:
- ✅ Code formatting check (`fmt-check`)
- ✅ Go vet (`vet`)
- ✅ Linting (`lint`)
- ✅ Tests (`test`)
- ✅ Build (`build`)

## Detailed Validation

For comprehensive validation (matches CI exactly):

```bash
bash validate-pr.sh
```

This script runs:
1. Go version check
2. Dependency download
3. go.mod tidy check
4. Code formatting check
5. go vet
6. golangci-lint
7. Tests with coverage
8. Build verification

## Pre-Push Hook

A pre-push git hook is available to automatically run validation before pushing:

```bash
# Hook is already installed at .git/hooks/pre-push
# It will automatically run validate-pr.sh before each push
```

To bypass the hook (not recommended):

```bash
git push --no-verify
```

## Common Issues and Fixes

### Formatting Issues

**Error**: `Code is not formatted`

**Fix**:
```bash
make fmt
# or
go fmt ./...
```

### Vet Issues

**Error**: `go vet failed`

**Fix**: Check the vet output for specific issues and fix them.

### Test Failures

**Error**: `Tests failed`

**Fix**: Run tests locally to see the failure:
```bash
go test -v ./...
```

### Build Failures

**Error**: `Build failed`

**Fix**: Check for compilation errors:
```bash
go build ./cmd/axon
```

## CI Workflow

The CI runs these checks in parallel:

1. **Validate PR** - Formatting and basic checks
2. **Vet** - go vet and fmt check
3. **Lint** - golangci-lint
4. **Test** - Unit tests with coverage
5. **Build** - Build on multiple platforms (Linux, macOS, Windows)
6. **Security Scan** - Gosec security scanner

## Best Practices

1. **Always run `make ci` before pushing**
2. **Fix formatting issues immediately** (`make fmt`)
3. **Run tests locally** to catch failures early
4. **Check CI status** after pushing
5. **Don't bypass pre-push hook** unless absolutely necessary

## Troubleshooting

### golangci-lint not found

```bash
make install-tools
```

### Tests timing out

Some tests may take longer. CI has a timeout, but local runs don't.

### Coverage too low

Coverage warnings are non-blocking during MVP phase, but should be improved over time.

## Related Commands

- `make fmt` - Format code
- `make fmt-check` - Check formatting without modifying
- `make vet` - Run go vet
- `make lint` - Run golangci-lint
- `make test` - Run tests
- `make build` - Build binary
- `make ci` - Run all CI checks
- `bash validate-pr.sh` - Run comprehensive validation

