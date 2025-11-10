# Release Cadence and Process

This document outlines the release cadence, process, and best practices for Axon releases.

## Release Cadence

### Current Strategy: **Semantic Versioning with Feature-Driven Releases**

- **Major releases (x.0.0)**: Significant milestones, breaking changes, or major architectural updates
- **Minor releases (x.y.0)**: New features, new adapters, backward-compatible enhancements
- **Patch releases (x.y.z)**: Bug fixes, security patches, documentation updates

### Release Frequency

- **Target**: Release when features are ready, not on a fixed schedule
- **Minimum**: At least one release per quarter
- **Maximum**: As needed for critical fixes or major features

## Release Process

### Pre-Release Checklist

Before creating a release, ensure:

- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint` or `make validate-pr`)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated with all changes
- [ ] Version number is updated in code (if needed)
- [ ] Release notes are prepared
- [ ] All PRs are merged and main branch is stable

### Step 1: Update CHANGELOG.md

1. Add a new section for the release version at the top (after `[Unreleased]`)
2. Categorize changes:
   - **Added**: New features
   - **Changed**: Changes in existing functionality
   - **Deprecated**: Soon-to-be removed features
   - **Removed**: Removed features
   - **Fixed**: Bug fixes
   - **Security**: Security vulnerabilities addressed
3. Include links to relevant PRs/issues
4. Follow [Keep a Changelog](https://keepachangelog.com/) format

Example:
```markdown
## [1.2.0] - 2024-12-01

### Added
- ModelScope adapter support (#15)
- Enhanced caching with TTL support (#16)

### Fixed
- Memory leak in cache manager (#17)
```

### Step 2: Create Release Branch (Optional)

For major/minor releases, consider creating a release branch:

```bash
git checkout -b release/v1.2.0
# Make any final adjustments
git push origin release/v1.2.0
```

### Step 3: Create Release Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create annotated tag with release notes
git tag -a v1.2.0 -m "Release v1.2.0: ModelScope adapter and enhanced caching

Features:
- Add ModelScope adapter support
- Enhanced caching with TTL

Fixes:
- Memory leak in cache manager

See CHANGELOG.md for full details."

# Push tag (triggers release workflow)
git push origin v1.2.0
```

### Step 4: Release Workflow

The GitHub Actions workflow (`.github/workflows/release.yml`) automatically:

1. **Builds binaries** for all platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)

2. **Creates GitHub Release** with:
   - Release notes from tag message
   - All binary assets
   - Checksums for verification

3. **Publishes assets** to GitHub Releases

### Step 5: Verify Release

After the workflow completes:

- [ ] Check GitHub Releases page
- [ ] Verify all platform binaries are present
- [ ] Test installer script with new version
- [ ] Verify CHANGELOG.md is accessible at release tag
- [ ] Update website/documentation if needed

### Step 6: Post-Release

1. **Announce release** (if major/minor):
   - Update website
   - Post on social media/community channels
   - Update documentation links

2. **Update Unreleased section** in CHANGELOG.md:
   - Move completed items to new release section
   - Add new planned items

## Release Notes Template

When creating a release tag, use this template:

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

## Documentation
- Doc update 1 (#PR)

See CHANGELOG.md for full details.
Installation: curl -sSL axon.mlosfoundation.org | sh
```

## Version Numbering

### Semantic Versioning (SemVer)

Format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Examples

- `1.0.0` → `1.0.1`: Patch (bug fix)
- `1.0.1` → `1.1.0`: Minor (new adapter)
- `1.1.0` → `2.0.0`: Major (breaking change)

## Release Types

### Patch Release (x.y.z → x.y.z+1)

**When to release:**
- Critical bug fixes
- Security patches
- Documentation corrections
- Minor improvements

**Process:**
1. Fix on main branch
2. Update CHANGELOG.md
3. Create patch tag
4. Push tag

### Minor Release (x.y.0 → x.y+1.0)

**When to release:**
- New adapter implementation
- New features
- Significant enhancements
- Backward-compatible changes

**Process:**
1. Complete feature development
2. Update all documentation
3. Update CHANGELOG.md
4. Create minor release tag
5. Push tag

### Major Release (x.0.0 → x+1.0.0)

**When to release:**
- Breaking API changes
- Major architectural changes
- Significant milestone achievements
- Complete rewrite of major components

**Process:**
1. Plan breaking changes
2. Update migration guide
3. Update all documentation
4. Update CHANGELOG.md with migration notes
5. Create major release tag
6. Push tag
7. Announce release

## CHANGELOG.md Maintenance

### Structure

```markdown
# Changelog

## [Unreleased]
### Added
- Planned feature 1
- Planned feature 2

## [1.2.0] - 2024-12-01
### Added
- Feature 1 (#15)
### Fixed
- Bug fix 1 (#16)

## [1.1.0] - 2024-11-09
...
```

### Best Practices

1. **Keep it updated**: Update CHANGELOG.md as you develop features
2. **Link to PRs**: Include PR numbers for traceability
3. **User-focused**: Write from user perspective, not developer perspective
4. **Categorize**: Use standard categories (Added, Changed, Fixed, etc.)
5. **Date format**: Use YYYY-MM-DD format
6. **Accessibility**: Ensure CHANGELOG.md is accessible at each release tag

## Automation

### GitHub Actions Workflow

The release workflow (`.github/workflows/release.yml`) handles:

- ✅ Multi-platform builds
- ✅ Asset creation and naming
- ✅ Checksum generation
- ✅ GitHub Release creation
- ✅ Asset upload

### Manual Steps Still Required

- ❌ CHANGELOG.md updates
- ❌ Release notes preparation
- ❌ Tag creation
- ❌ Post-release announcements

## Release Context Reuse

See [RELEASE_CONTEXT.md](RELEASE_CONTEXT.md) for a reusable context document that can be invoked across repositories or releases.

## Emergency Releases

For critical security or stability issues:

1. Create hotfix branch from main
2. Apply fix
3. Test thoroughly
4. Create patch release tag
5. Merge hotfix to main
6. Announce immediately

## Release Communication

### Internal (Before Release)

- Review with team
- Test installer script
- Verify documentation

### External (After Release)

- GitHub Release notes (automatic)
- Website updates (if major/minor)
- Community announcements (if major/minor)

## Questions?

- See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines
- See [CHANGELOG.md](../CHANGELOG.md) for change history
- Open an issue for release-related questions

---

**Signal. Propagate. Myelinate.** 🧠

