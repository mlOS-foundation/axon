# Changelog

All notable changes to Axon will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- ModelScope adapter (Phase 1)
- TensorFlow Hub adapter (Phase 1)
- Enhanced caching and optimization
- Model versioning and A/B testing
- MLOS Core Runtime integration

## [1.1.1] - 2024-11-10

### Fixed
- **Multi-part model names in `axon list`**: Fixed issue where `axon list` didn't display models with multi-part names (e.g., `pytorch/vision/resnet50@latest`)
- **Cache manager parsing**: Enhanced `ListCachedModels` to correctly handle model paths with more than 3 parts (namespace/repo/model/version)

### Changed
- Improved model path parsing to support complex model namespaces

## [1.1.0] - 2024-11-09

### Added
- **PyTorch Hub Adapter**: Full support for installing models from PyTorch Hub
  - Automatic detection of PyTorch Hub models via `pytorch/` namespace
  - Support for multi-part model names (e.g., `pytorch/vision/resnet50@latest`)
  - Intelligent parsing of `hubconf.py` files to extract model URLs
  - Fallback URL support for common PyTorch models
  - GitHub API integration for downloading from PyTorch repositories
- **Enhanced adapter priority**: PyTorch Hub adapter registered with higher priority than Hugging Face
- **Comprehensive unit tests**: Added test coverage for PyTorch Hub adapter

### Changed
- **Roadmap update**: Removed ONNX Model Zoo (deprecated July 2025), added PyTorch Hub, ModelScope, and TensorFlow Hub as Phase 1 adapters
- **Documentation**: Updated all docs and website pages to reflect new Phase 1 roadmap
- **Release workflow**: Fixed asset naming to include version number in filename

### Fixed
- **Release asset naming**: Fixed issue where release assets were incorrectly named (missing version number)
- **Installer script**: Enhanced fallback logic for handling release assets with various naming conventions

## [1.0.2] - 2024-11-08

### Fixed
- **Linting errors**: Resolved all `golangci-lint` errors including:
  - Package comments for all exported packages
  - Import formatting (`goimports`)
  - Error handling (`errcheck`)
  - Method documentation (`exported`)
  - Config path stutter warning
- **Go version**: Updated `go.mod` to use Go 1.21 (compatible with golangci-lint)
- **PR validation**: Pinned `golangci-lint-action` to v1.64.8 for consistency
- **Local validation**: Added `validate-pr` Makefile target and documentation for local PR validation

### Changed
- **Configuration API**: Renamed `ConfigPath()` to `Path()` to fix stutter warning
- **Contributing guide**: Added instructions for running local PR validation

## [1.0.1] - 2024-11-07

### Fixed
- **Installer script**: Added fallback logic to handle release assets with incorrect naming
- **Release workflow**: Fixed asset naming to ensure proper format (`axon_${VERSION}_${GOOS}_${GOARCH}.tar.gz`)
- **Cloudflare DNS**: Updated documentation for proper DNS and redirect rule setup

### Changed
- **Release process**: Improved release workflow to handle edge cases in asset naming

## [1.0.0] - 2024-11-06

### Added
- **MVP Release**: First stable release with core functionality
- **Universal Model Installer**: Pluggable adapter architecture for installing models from any repository
- **Hugging Face Integration**: Zero-configuration support for Hugging Face models
  - Automatic model discovery
  - Direct download from Hugging Face Hub
  - Support for gated/private models with token authentication
- **Model Lifecycle Management**:
  - `axon install` - Install models from any supported repository
  - `axon list` - List all installed models
  - `axon search` - Search for models across repositories
  - `axon info` - Get detailed model information
  - `axon update` - Update models to latest version
  - `axon uninstall` - Remove installed models
- **Manifest System**: YAML-based manifests with validation
  - Automatic manifest generation on-the-fly
  - Checksum verification
  - Metadata tracking
- **Local Caching**: Intelligent caching for offline access
  - Automatic cache management
  - Integrity verification
  - Cache directory structure
- **Registry Client**: HTTP client for model discovery
- **CLI Interface**: Comprehensive command-line interface
- **Installer Script**: One-line installation via `curl -sSL axon.mlosfoundation.org | sh`
- **Release Infrastructure**: Automated build and release workflow
  - Multi-platform binary builds (Linux/macOS, amd64/arm64)
  - Checksum generation
  - GitHub Releases integration

### Documentation
- Comprehensive README with quick start guide
- Contributing guidelines
- MVP status documentation
- Architecture documentation
- Adapter roadmap

### Supported Platforms
- **macOS**: amd64, arm64
- **Linux**: amd64, arm64

---

## Release Types

- **Major** (x.0.0): Breaking changes, major features, or significant architectural changes
- **Minor** (x.y.0): New features, new adapters, backward-compatible enhancements
- **Patch** (x.y.z): Bug fixes, security patches, documentation updates

## Links

- [GitHub Releases](https://github.com/mlOS-foundation/axon/releases)
- [Installation Guide](README.md#installation)
- [Contributing Guide](CONTRIBUTING.md)
- [Adapter Roadmap](docs/ADAPTER_ROADMAP.md)

---

**Signal. Propagate. Myelinate.** 🧠

