# Release Context - Reusable Release Process

This document provides a reusable context for creating releases across the MLOS Foundation repositories. It can be invoked via Cursor's context system or referenced directly.

## Quick Release Command

```bash
# 1. Sync with main
git checkout main && git pull origin main

# 2. Update CHANGELOG.md (add new version section)

# 3. Create and push release tag
git tag -a vX.Y.Z -m "Release vX.Y.Z: [Description]

[Features/Fixes/Changes]

See CHANGELOG.md for full details."
git push origin vX.Y.Z

# 4. Verify release workflow runs
gh run watch
```

## Release Checklist

### Pre-Release

- [ ] All tests pass: `make test`
- [ ] Linting passes: `make validate-pr` or `make lint`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated with new version section
- [ ] All PRs merged to main
- [ ] Main branch is stable

### Release

- [ ] Create release tag with proper message
- [ ] Push tag to trigger workflow
- [ ] Monitor release workflow
- [ ] Verify all platform binaries created
- [ ] Verify GitHub Release created

### Post-Release

- [ ] Verify installer script works with new version
- [ ] Update website/documentation if needed
- [ ] Announce release (if major/minor)

## CHANGELOG.md Format

The CHANGELOG.md format directly impacts the GitHub release page. Follow this format for best results:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- **Feature Name**: Detailed description of the feature (#PR_NUMBER)
  - Sub-feature or detail
  - Another detail
  - Usage example: `axon install example/model@latest`
- **Another Feature**: Description (#PR_NUMBER)

### Changed
- Change description (#PR_NUMBER)

### Fixed
- Bug fix description (#PR_NUMBER)

### Deprecated
- Deprecated feature (#PR_NUMBER)

### Removed
- Removed feature (#PR_NUMBER)

### Security
- Security fix (#PR_NUMBER)

### Documentation
- Documentation updates (#PR_NUMBER)
```

**Best Practices for Release Notes**:

1. **Be Descriptive**: Write changelog entries as if they'll appear on the release page (they will!)
2. **Include Examples**: For new features, include usage examples in the "Added" section
3. **Link PRs**: Always include PR numbers in parentheses (e.g., `(#13)`)
4. **Use Bold for Features**: Use `**Feature Name**` for major features to make them stand out
5. **Group Related Changes**: Use sub-bullets to group related changes under a feature
6. **Include Impact**: Mention the impact or benefit of changes (e.g., "Axon now supports 72%+ of ML practitioners")

## Release Tag Message Template

```markdown
Release vX.Y.Z: [Brief Description]

## Features
- Feature 1 (#PR)
- Feature 2 (#PR)

## Fixes
- Fix 1 (#PR)
- Fix 2 (#PR)

## Changes
- Change 1 (#PR)
- Change 2 (#PR)

See CHANGELOG.md for full details.
Installation: curl -sSL axon.mlosfoundation.org | sh
```

## Version Numbering Rules

- **MAJOR (x.0.0)**: Breaking changes
- **MINOR (x.y.0)**: New features, backward compatible
- **PATCH (x.y.z)**: Bug fixes, backward compatible

## Release Workflow

The GitHub Actions workflow (`.github/workflows/release.yml`) automatically:

1. Detects tag push (pattern: `v*.*.*`)
2. Extracts version from tag
3. Builds binaries for all platforms:
   - Linux: amd64, arm64
   - macOS: amd64, arm64
   - Windows: amd64
4. Creates archives with naming: `axon_${VERSION}_${GOOS}_${GOARCH}.tar.gz`
5. Generates SHA256 checksums
6. Creates GitHub Release with:
   - Tag name as release name
   - **Detailed release notes** extracted from CHANGELOG.md:
     - "What's New" section (key features from "Added" section)
     - Full changelog for the version
     - Installation instructions
     - Supported platforms
     - Documentation links
   - All binary assets
   - Checksum files

### Release Page Details

The release workflow automatically generates a comprehensive release page that includes:

1. **Installation Instructions**: One-line installer command
2. **What's New**: Key features extracted from the "Added" section of CHANGELOG.md
3. **Full Changelog**: Complete changelog entry for the version
4. **Supported Platforms**: List of all supported OS/architecture combinations
5. **Documentation Links**: Links to relevant documentation files

**Important**: To ensure detailed release notes appear on the GitHub release page:

- ✅ **Always update CHANGELOG.md** before creating a release tag
- ✅ **Use proper CHANGELOG format** with "Added", "Changed", "Fixed", etc. sections
- ✅ **Include PR numbers** in changelog entries (e.g., `(#13)`)
- ✅ **Be descriptive** in changelog entries - they appear on the release page
- ✅ **Include usage examples** in the "Added" section for new features
- ✅ **Use sub-bullets** for feature details (they will be included in release notes)
- ✅ **Ensure "Added" section has content** - this populates the "What's New" section

The workflow automatically extracts:
- **Features** from the "### Added" section (shown in "🎯 What's New")
  - Includes all bullet points and sub-bullets under "Added"
  - Preserves formatting (bold, code blocks, etc.)
- **Full changelog** entry (shown in "📋 Full Changelog")
  - Includes all sections: Added, Changed, Fixed, Documentation, etc.
  - Preserves all formatting and structure

**Note**: If "What's New" or "Full Changelog" sections appear empty:
1. Check that CHANGELOG.md has the version entry with proper format: `## [X.Y.Z] - YYYY-MM-DD`
2. Verify the "### Added" section exists and has content
3. Check the release workflow logs for extraction errors
4. The workflow includes debug output showing extracted content

## Platform Support

- **Linux**: amd64, arm64
- **macOS**: amd64, arm64

## Installation

One-line installation:
```bash
curl -sSL axon.mlosfoundation.org | sh
```

Or from GitHub:
```bash
curl -sSL https://raw.githubusercontent.com/mlOS-foundation/axon/main/install.sh | sh
```

## Release Types

### Patch Release (x.y.z → x.y.z+1)
- Critical bug fixes
- Security patches
- Documentation updates
- Minor improvements

### Minor Release (x.y.0 → x.y+1.0)
- New adapter implementation
- New features
- Significant enhancements
- Backward-compatible changes

### Major Release (x.0.0 → x+1.0.0)
- Breaking API changes
- Major architectural changes
- Significant milestones
- Complete rewrites

## CHANGELOG.md Location

The CHANGELOG.md file must be:
- Located at repository root
- Accessible at each release tag
- Updated before each release
- Following [Keep a Changelog](https://keepachangelog.com/) format

## Verification Commands

```bash
# Check release exists
gh release view vX.Y.Z

# List release assets
gh release view vX.Y.Z --json assets --jq '.assets[].name'

# Test installer
curl -sSL axon.mlosfoundation.org | sh

# Verify CHANGELOG accessible
curl -s https://raw.githubusercontent.com/mlOS-foundation/axon/vX.Y.Z/CHANGELOG.md | head -50
```

## Common Release Scenarios

### Scenario 1: Patch Release (Bug Fix)

```bash
# 1. Fix is merged to main
git checkout main && git pull

# 2. Update CHANGELOG.md
# Add section: ## [1.1.2] - 2024-11-11
# Add: ### Fixed
# Add: - Bug description (#PR)

# 3. Create and push tag
git tag -a v1.1.2 -m "Release v1.1.2: Bug fix

Fixes:
- Bug description (#PR)

See CHANGELOG.md for details."
git push origin v1.1.2
```

### Scenario 2: Minor Release (New Feature)

```bash
# 1. Feature is merged to main
git checkout main && git pull

# 2. Update CHANGELOG.md
# Add section: ## [1.2.0] - 2024-12-01
# Add: ### Added
# Add: - Feature description (#PR)

# 3. Create and push tag
git tag -a v1.2.0 -m "Release v1.2.0: New Feature

Features:
- Feature description (#PR)

See CHANGELOG.md for details."
git push origin v1.2.0

# 4. Update website/documentation
# 5. Announce release
```

### Scenario 3: Major Release (Breaking Changes)

```bash
# 1. All changes merged to main
git checkout main && git pull

# 2. Update CHANGELOG.md with migration guide
# Add section: ## [2.0.0] - 2025-01-01
# Add: ### Changed (Breaking)
# Add: - Breaking change description (#PR)
# Add: ### Migration Guide
# Add: Migration instructions

# 3. Create and push tag
git tag -a v2.0.0 -m "Release v2.0.0: Major Update

Breaking Changes:
- Change description (#PR)

Migration:
See CHANGELOG.md for migration guide.

See CHANGELOG.md for full details."
git push origin v2.0.0

# 4. Update all documentation
# 5. Create migration guide
# 6. Announce release widely
```

## Context Invocation

### In Cursor

Use this context by referencing:
```
@RELEASE_CONTEXT.md
```

Or invoke via shortcut:
```
call release-context
```

### Manual Reference

When creating a release, reference this document for:
- Checklist items
- Tag message format
- CHANGELOG format
- Verification steps

## Repository-Specific Notes

### Axon

- **Installer URL**: `axon.mlosfoundation.org`
- **Binary naming**: `axon_${VERSION}_${GOOS}_${GOARCH}.tar.gz`
- **Go version**: 1.21
- **Build command**: `go build -ldflags "-X main.version=v${VERSION}"`

### Other MLOS Repositories

Adapt this context for other repositories by:
1. Updating binary naming pattern
2. Updating installer URL
3. Updating build commands
4. Updating platform support

## Troubleshooting

### Release workflow fails

1. Check workflow logs: `gh run view [RUN_ID]`
2. Verify tag format: Must match `v*.*.*`
3. Verify Go version compatibility
4. Check artifact upload permissions

### Assets missing

1. Check workflow completed successfully
2. Verify asset naming matches expected pattern
3. Check GitHub Release page
4. Re-run workflow if needed

### CHANGELOG not accessible

1. Verify CHANGELOG.md exists at tag
2. Check file is committed to repository
3. Verify GitHub raw URL works
4. Update release workflow if needed

## Best Practices

1. **Always update CHANGELOG.md before tagging**
2. **Use descriptive tag messages**
3. **Link to PRs in changelog**
4. **Test installer script after release**
5. **Verify all platforms build successfully**
6. **Keep release notes user-focused**
7. **Follow semantic versioning strictly**
8. **Document breaking changes clearly**

## Related Documents

- [RELEASE_CADENCE.md](RELEASE_CADENCE.md) - Detailed release process
- [CHANGELOG.md](../CHANGELOG.md) - Change history
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines

---

**Signal. Propagate. Myelinate.** 🧠

