# Changelog

All notable changes to Axon will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Enhanced caching and optimization
- Model versioning and A/B testing
- MLOS Core Runtime integration

## [1.4.0] - 2024-11-10

### Added
- **ModelScope Adapter**: Full support for installing models from ModelScope (Alibaba Cloud's model repository) (#17)
  - Automatic detection of ModelScope models via `modelscope/` and `ms/` namespaces
  - Support for multimodal AI models (vision, audio, text)
  - REST API integration for model discovery and metadata fetching
  - Automatic manifest creation with model metadata
  - On-the-fly package generation from ModelScope downloads
  - Comprehensive unit tests
  - Promoted from example adapter to builtin adapter
- **Replicate Example Adapter**: New example adapter demonstrating API-based model access patterns
  - Shows how to implement adapters for API-based repositories
  - Demonstrates different patterns from file-based adapters
  - Complete implementation with tests in `internal/registry/examples/`

### Changed
- **ModelScope Promotion**: ModelScope adapter moved from `examples/` to `builtin/` package
  - Now registered automatically with `RegisterDefaultAdapters()`
  - Available by default in all Axon installations
  - Updated adapter priority: Local → PyTorch Hub → TensorFlow Hub → ModelScope → Hugging Face
- **Adapter Registration**: Updated `builtin/register.go` to include ModelScope adapter
- **Coverage Statistics**: Axon now supports 80%+ of ML practitioners with ModelScope included

### Documentation
- Updated `README.md` with ModelScope as available adapter (v1.4.0+)
- Updated `REPOSITORY_ADAPTERS.md` with full ModelScope adapter documentation
- Updated `ADAPTER_ROADMAP.md` marking ModelScope as completed (v1.4.0)
- Updated `ADAPTER_FRAMEWORK.md` to use Replicate as example instead of ModelScope
- Updated website (`ecosystem.html`) to show ModelScope as available

## [1.3.0] - 2024-11-10

### Added
- **Adapter Framework Refactoring**: Complete refactoring of adapter architecture using GoF design patterns (#16)
  - **Core Framework** (`internal/registry/core/`): New package with interfaces and utilities
    - `RepositoryAdapter` interface (Adapter Pattern) - Unified interface for all repositories
    - `AdapterRegistry` (Registry Pattern) - Centralized adapter management
    - `AdapterFactory` (Factory Pattern) - Dynamic adapter creation
    - `AdapterBuilder` (Builder Pattern) - Fluent configuration
    - `ModelValidator` - Generic model existence validation helper
    - `HTTPClient` - HTTP requests with authentication support
    - `PackageBuilder` - Create .axon packages with progress tracking
    - `DownloadFile` - Download files with progress callbacks
    - `ComputeChecksum` - SHA256 checksum computation
  - **Builtin Adapters** (`internal/registry/builtin/`): Migrated all adapters to new framework
    - HuggingFace adapter using core helpers
    - PyTorch Hub adapter using core helpers
    - TensorFlow Hub adapter using core helpers
    - Local Registry adapter
    - Centralized registration via `RegisterDefaultAdapters()`
  - **Example Adapter** (`internal/registry/examples/`): ModelScope adapter as reference implementation
    - Complete implementation demonstrating framework usage
    - Comprehensive unit tests
    - Shows best practices for adapter development
  - **Comprehensive Documentation**:
    - `docs/ADAPTER_FRAMEWORK.md` - Complete framework overview with design patterns
    - `docs/ADAPTER_DEVELOPMENT.md` - Step-by-step guide for creating new adapters
    - `docs/ADAPTER_FRAMEWORK_REFACTOR.md` - Refactoring summary and benefits
    - `docs/ADAPTER_MIGRATION_STATUS.md` - Migration tracking
    - `internal/registry/README.md` - Package overview

### Changed
- **CLI Commands**: Updated to use new `core.AdapterRegistry` and `builtin.RegisterDefaultAdapters()`
  - `axon install` now uses new framework
  - `axon info` now uses new framework
  - `axon search` now uses new framework
- **Adapter Structure**: All adapters now use core helper utilities for common operations
  - Reduced code duplication
  - Consistent error handling
  - Standardized validation
- **Code Organization**: Removed old `internal/registry/adapter.go` (1830 lines) in favor of modular structure

### Benefits
- **Extensibility**: Easier to add new adapters with clear patterns and reusable utilities
- **Maintainability**: Clean separation between core interfaces and implementations
- **Testability**: Isolated components are easier to test
- **Type Safety**: Strong interfaces prevent errors
- **Documentation**: Comprehensive guides for adapter development

### Migration
- **No breaking changes**: All existing functionality preserved
- **Backward compatible**: Existing adapters work identically
- **Internal refactoring**: Changes are transparent to users

## [1.2.2] - 2024-11-10

### Fixed
- **Generic model validation**: Added `ModelValidator` helper for consistent model existence validation across all adapters (#15)
  - TensorFlow Hub adapter now correctly rejects non-existent models
  - Hugging Face adapter validates model existence before creating manifest
  - PyTorch Hub adapter validates both repository and model in hubconf.py
  - Fixes issue where `axon info` showed information for non-existent models
  - Enhanced validation to detect HTML error/search pages for better accuracy

### Added
- **Auto-install CI tools**: Added `make install-tools` target and automatic tool installation in validation scripts (#15)
  - Automatically installs `golangci-lint` v1.64.8 (matching CI) if missing
  - Automatically installs `goimports` if missing
  - Ensures `GOPATH/bin` is in PATH for tool discovery
  - Updated `validate-pr.sh` to install tools before running checks
  - Updated `lint` and `fmt` Makefile targets to depend on `install-tools`

### Changed
- **Validation script**: Enhanced `validate-pr.sh` to automatically install missing CI tools
- **Makefile**: Added tool version variables and `install-tools` target
- **Documentation**: Updated `CONTRIBUTING.md` with auto-install tool instructions

## [1.2.1] - 2024-11-10

### Fixed
- **Info command adapter support**: Fixed `axon info` command to use adapter registry system instead of old registry client (#14)
  - Now works with all adapters (Hugging Face, PyTorch Hub, TensorFlow Hub)
  - Displays comprehensive model information including files, sizes, and metadata
  - Previously showed "Model info not yet available" for all models
  - Added `formatBytes()` helper for human-readable file sizes

### Changed
- **Info command display**: Enhanced model information output with better formatting and file details

## [1.2.0] - 2024-11-10

### Added
- **TensorFlow Hub Adapter**: Full support for installing models from TensorFlow Hub (tfhub.dev) (#13)
  - Automatic detection of TensorFlow Hub models via `tfhub/` and `tf/` namespaces
  - Support for SavedModel and TFLite formats
  - REST API integration for model discovery and metadata fetching
  - Automatic manifest creation with model metadata
  - On-the-fly package generation from TensorFlow Hub downloads
  - Comprehensive unit tests with mock server support
- **Enhanced adapter priority**: TensorFlow Hub adapter registered with higher priority than Hugging Face
- **Validation script updates**: Added test case for TensorFlow Hub installation in `validate-use-case.sh`

### Changed
- **Adapter registration order**: Updated to check TensorFlow Hub adapter before Hugging Face fallback
- **Local registry adapter**: Updated `CanHandle()` to exclude TensorFlow Hub namespaces (`tfhub`, `tf`)
- **Coverage statistics**: Axon now supports 72%+ of ML practitioners (Hugging Face: 60%+, PyTorch Hub: 5%+, TensorFlow Hub: 7%+)

### Documentation
- Updated `README.md` with TensorFlow Hub examples and usage
- Updated `ADAPTER_ROADMAP.md` marking TensorFlow Hub as available (v1.2.0+)
- Updated `REPOSITORY_ADAPTERS.md` with detailed TensorFlow Hub adapter section
- Updated website (`ecosystem.html`) to show TensorFlow Hub as available

## [1.1.2] - 2024-11-10

### Added
- **CHANGELOG.md**: Comprehensive changelog following [Keep a Changelog](https://keepachangelog.com/) format with all versions documented (#12)
- **Release Cadence Documentation**: Added `docs/RELEASE_CADENCE.md` with detailed release process, checklists, and best practices (#12)
- **Reusable Release Context**: Added `docs/RELEASE_CONTEXT.md` for reusable release context across MLOS Foundation repositories (#12)

### Changed
- **Release Workflow**: Enhanced `.github/workflows/release.yml` to automatically extract release notes from CHANGELOG.md and include them in GitHub Releases (#12)
- **Release Process**: Established standardized release cadence and process documentation (#12)

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

