# Changelog

All notable changes to Axon will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Enhanced caching and optimization
- Model versioning and A/B testing

## [2.1.0] - 2024-11-19

### Added
- **OCI Artifacts in Releases**: Docker images now available as OCI artifacts attached to GitHub Releases (#36)
  - Docker images saved as `.tar.gz` files for both `linux/amd64` and `linux/arm64` platforms
  - Users can download images directly from releases without needing a container registry
  - Images available alongside release binaries for convenient distribution
  - Usage: `docker load < axon-converter-2.1.0-linux-amd64.tar.gz`
- **Docker Build Optimization**: Significant build time improvements (#36)
  - BuildKit cache mounts for pip cache (70-80% faster subsequent builds)
  - Removed `--no-cache-dir` flag to allow pip caching during build
  - Cache persists across GitHub Actions runs using BuildKit cache
  - First build: ~15-20 min, Subsequent builds: ~3-5 min
- **Local Docker Testing**: New test script for validating Docker builds locally (#36)
  - `scripts/test-docker-converter-local.sh` for pre-push validation
  - Tests all dependencies (PyTorch, TensorFlow, Transformers, ONNX)
  - Validates conversion scripts are present and executable

### Fixed
- **TensorFlow Installation**: Resolved TensorFlow import errors and CUDA conflicts (#36)
  - Pinned protobuf to `<4.21,>=3.20` for TensorFlow 2.13+ compatibility
  - Installed TensorFlow before PyTorch to avoid CUDA library conflicts
  - Installed PyTorch CPU-only to prevent CUDA conflicts with TensorFlow
  - Split pip install into separate steps for better error handling
  - Added `TF_CPP_MIN_LOG_LEVEL=2` to suppress CUDA warnings in tests
- **Docker Workflow Tests**: Improved test output with better error messages (#36)
  - Added checkmarks (✅) to test output for better visibility
  - Suppressed TensorFlow CUDA warnings in CI tests

### Changed
- **Docker Image**: Updated Docker image with optimized dependency installation (#36)
  - Better installation order to prevent conflicts
  - CPU-only builds for both PyTorch and TensorFlow to avoid CUDA issues
  - Improved error handling and build reliability

### Documentation
- **Container Registry Options**: New documentation exploring alternatives to GHCR/Docker Hub (#36)
  - `docs/CONTAINER_REGISTRY_OPTIONS.md` - Comprehensive guide to registry options
  - Explores OCI artifacts, Quay.io, Docker Hub, and multi-registry strategies
- **Docker Build Optimization**: Documentation explaining build optimizations (#36)
  - `docs/DOCKER_BUILD_OPTIMIZATION.md` - Detailed explanation of build improvements
  - Performance metrics and optimization strategies

## [2.0.2] - 2024-11-19

### Fixed
- **Docker Workflow YAML Syntax**: Fixed YAML syntax error on line 89 in docker-converter workflow (#33, #34)
  - Changed single-line `run:` command to multi-line format for GitHub Actions expressions
  - Fixes "mapping values are not allowed here" error
  - Workflow now validates correctly in GitHub Actions
- **Docker Workflow Triggers**: Optimized Docker image build triggers (#34)
  - Removed push/pull_request triggers (no longer builds on every Dockerfile change)
  - Only triggers on release creation/publication
  - Can still be manually triggered via workflow_dispatch
  - Reduces CI costs and ensures images are only published for stable releases
- **YAML Validation**: Added comprehensive local YAML validation (#32)
  - New `scripts/validate-yaml.sh` script for local validation
  - Integrated into `make validate-yaml` and `make ci`
  - Supports both yamllint and PyYAML validation
  - Auto-installs yamllint via Homebrew on macOS
  - Catches YAML errors locally before pushing

### Changed
- **Docker Workflow**: Improved workflow efficiency and cost management
  - Images only built and pushed on releases (not on every change)
  - Better versioning strategy (semver tags for releases)
  - Cleaner workflow logic (removed unnecessary conditionals)

## [2.0.1] - 2024-11-18

### Fixed
- **Docker Image Name**: Fixed incorrect Docker image name for ONNX converter (#26)
  - Changed from `axon-converter:latest` to `ghcr.io/mlOS-foundation/axon-converter:latest`
  - Docker-based ONNX conversion now works correctly
  - Users can convert models to ONNX without local Python dependencies
  - Fixes error: "failed to pull Docker image axon-converter:latest"

## [2.0.0] - 2024-11-18

### ⚠️ BREAKING CHANGES

This release introduces the **Manifest-First Architecture**, a fundamental architectural change that enables format-agnostic model execution. While backward compatible for basic operations, this change requires Core to read `execution_format` from manifests for proper plugin selection.

**Migration Notes:**
- Existing manifests will be automatically updated with `execution_format` on next `axon install`
- Core implementations must read `execution_format` from manifest for plugin selection
- Manifest structure now includes `execution_format` field (required for Core v2.0.0+)

### Added
- **Manifest-First Architecture**: Format-agnostic model package format (#25)
  - Added `execution_format` to Format struct for dynamic plugin selection
  - Added `PreprocessingSpec` to IOSpec for preprocessing hints
  - Automatic I/O schema extraction from model configs (BERT, GPT, T5, Vision models)
  - Preprocessing hints for tokenization, normalization, image preprocessing
  - Execution format detection based on available model files
  - Manifest updates after installation with actual I/O schema and execution format
  - Ensures tokenizer files are included in packages
  - Enables format-agnostic Core execution
  - Future-proof architecture for format transitions
- **Version Command**: New `axon version` command (#25)
  - Shows Axon version, build type (local vs installed), git commit, build date
  - Displays Go version and OS/Architecture
  - Auto-detects build type based on binary location
  - Helps distinguish between installed and local builds
- **Enhanced Format Detection**: Improved execution format detection (#25)
  - Support for archived models (TensorFlow Hub `.tar.gz`, ModelScope packages)
  - Manifest type fallback when files don't match expected patterns
  - Enhanced file detection for `.bin` (PyTorch), `.h5` (TensorFlow/Keras)
  - Automatic format detection for all supported repositories
- **PR Validation**: Local CI checks before pushing (#25)
  - Pre-push git hook for automatic validation
  - `make ci` target for running all CI checks locally
  - Comprehensive `validate-pr.sh` script matching CI exactly
  - Documentation guide in `docs/PR_VALIDATION.md`

### Changed
- **Manifest Structure**: Enhanced Format and IOSpec structures
  - `Format.ExecutionFormat` field added (required for Core compatibility)
  - `IOSpec.Preprocessing` field added for preprocessing hints
  - Manifest saved as JSON (matching cache manager format)
- **ONNX Conversion Flow**: Improved integration with manifest-first architecture
  - ONNX conversion remains preferred path
  - Graceful degradation when conversion fails
  - Manifest reflects actual execution format after conversion attempts
  - Detailed documentation in `docs/MANIFEST_FIRST_ARCHITECTURE.md`

### Benefits
- **Format Independence**: Core can support multiple execution formats simultaneously
- **Future-Proof**: Easy format transitions (ONNX → PyTorch → etc.) without Core changes
- **Complete Metadata**: All execution information in manifest (I/O schema, preprocessing, format)
- **Preprocessing Automation**: Automatic tokenization and preprocessing based on manifest hints
- **Dynamic Plugin Selection**: Core selects plugin based on `execution_format` in manifest
- **Better Developer Experience**: Local validation prevents CI failures
- **Improved Debugging**: Version command helps identify build type and version

### Documentation
- Added `docs/MANIFEST_FIRST_ARCHITECTURE.md`: Comprehensive architecture documentation
- Added `docs/PR_VALIDATION.md`: Guide for PR validation and local testing
- Added `FORMAT_DETECTION_TEST_RESULTS.md`: Test results for all repositories
- Updated website (`mlosfoundation.org/ecosystem.html`) with manifest-first features

### Testing
- ✅ Format detection tested across all repositories:
  - PyTorch Hub: `pytorch/vision/resnet18` → `execution_format: pytorch`
  - Hugging Face: `hf/distilbert-base-uncased` → `execution_format: pytorch`
  - TensorFlow Hub: `tfhub/google/imagenet/mobilenet_v2_100_224/classification/5` → `execution_format: tensorflow`
  - ModelScope: `modelscope/damo/cv_resnet18_image-classification` → `execution_format: pytorch`
- ✅ All CI checks passing
- ✅ Pre-push validation working

## [1.7.0] - 2024-11-17

### Added
- **List Command Format Flag**: Added `--format` flag to `axon list` command (#24)
  - `--format default`: Original format with header and indentation (default behavior)
  - `--format names`: Simple namespace/name format (one per line, no version) - perfect for piping
  - `--format json`: JSON array output for programmatic use
  - Enables easy uninstall of all models: `axon list --format names | xargs -I {} axon uninstall {}`
  - JSON output for scripting and automation
  - Backward compatible (default format unchanged)
- **Build-Local Target**: Added `make build-local` target for local installation without sudo
  - Installs binaries to `~/.local/bin` (no sudo required)
  - Automatically creates `~/.local/bin` if needed
  - Warns if `~/.local/bin` is not in PATH
  - Provides instructions for adding to PATH
  - Enables easy local testing and validation

### Benefits
- **Easy model management**: Pipe list output directly to uninstall command
- **Scripting support**: JSON output enables programmatic use
- **Local development**: Build and test without system-wide installation
- **User-friendly**: No sudo required for local installation

## [1.6.0] - 2024-11-16

### Added
- **Docker-Based ONNX Conversion**: Zero Python dependencies on host machine (#23)
  - Docker containers handle all Python ML framework dependencies
  - Multi-framework Docker image with PyTorch, TensorFlow, Transformers, and ONNX pre-installed
  - Version-independent volume mapping: Host Axon cache mapped to container
  - Repository-driven dependencies: Docker image selection based on supported repositories
  - Graceful degradation: Falls back to local Python if Docker unavailable
  - Automated CI/CD: Docker image automatically built and published on merge
  - E2E testing: Comprehensive workflow tests for Docker integration
- **Docker Converter Module**: New `internal/converter/docker.go` package
  - `ConvertToONNXWithDocker()`: Docker-based conversion implementation
  - `IsDockerAvailable()`: Docker availability checking
  - `EnsureDockerImage()`: Automatic Docker image pulling
  - `getDockerImageForRepository()`: Repository-specific image selection
- **Conversion Scripts**: Python scripts for framework-specific conversion
  - `convert_huggingface.py`: Hugging Face model conversion
  - `convert_pytorch.py`: PyTorch Hub model conversion
  - `convert_tensorflow.py`: TensorFlow Hub model conversion
- **Docker Image**: `ghcr.io/mlOS-foundation/axon-converter:latest`
  - Pre-configured with all ML frameworks
  - Multi-architecture support (linux/amd64, linux/arm64)
  - Automated build and publish workflow
- **CI/CD Integration**: Automated Docker image publishing
  - Builds and publishes on merge to main
  - E2E testing workflow for Docker converter
  - Multi-architecture builds

### Changed
- **ONNX Conversion Flow**: Enhanced to try Docker first, then local Python
  - Step 1: Download pre-converted ONNX (pure Go, no dependencies)
  - Step 2: Try Docker-based conversion (zero Python on host)
  - Step 3: Fall back to local Python conversion (if available)
  - Step 4: Graceful skip if all methods unavailable
- **User Experience**: Simplified distribution with zero Python installation required
  - Users can run `axon install` without any Python setup
  - Docker image automatically pulled on first use
  - Better isolation and reproducibility

### Benefits
- **Zero Python on Host**: No Python installation required for ONNX conversion
- **Simplified Distribution**: Easier for users, no dependency management
- **Better Isolation**: Docker containers provide clean, reproducible environments
- **Automated Maintenance**: CI/CD handles Docker image updates automatically
- **Multi-Architecture**: Supports both amd64 and arm64 platforms

## [1.5.0] - 2024-11-16

### Added
- **ONNX Conversion During Install**: Automatic ONNX conversion during `axon install` for MLOS Core compatibility (#21)
  - Pure Go-first approach: Downloads pre-converted ONNX files from repositories when available (no Python required)
  - Optional Python conversion: Falls back to Python-based conversion if pre-converted ONNX not available
  - Graceful degradation: Skips conversion if Python unavailable (models still work with framework-specific plugins)
  - Zero mandatory dependencies: Axon works without Python, conversion is optional enhancement
  - ONNX files included in MPF packages: `model.onnx` automatically added to `.axon` packages
  - Package extraction and rebuilding: Extracts packages, adds ONNX files, and rebuilds packages seamlessly
- **ONNX Converter Module**: New `internal/converter/onnx.go` package
  - `DownloadPreConvertedONNX()`: Pure Go HTTP download of pre-converted ONNX from repositories
  - `ConvertToONNX()`: Unified conversion interface with pure Go first, Python fallback
  - `CanConvert()`: Framework compatibility checking
  - Supports Hugging Face, PyTorch, and TensorFlow models
- **Package Extraction**: Added `extractPackage()` function to extract `.axon` tar.gz packages
- **Package Rebuilding**: Added `rebuildPackageWithONNX()` to rebuild packages with ONNX files included

### Changed
- **Install Flow**: Enhanced `axon install` to automatically:
  1. Download and cache model package
  2. Extract package to cache directory
  3. Attempt ONNX conversion (pure Go first, Python optional)
  4. Rebuild package with ONNX file included
  5. Store complete MPF package with ONNX for MLOS Core
- **Architecture Alignment**: Aligned with patent US-63/861,527 architecture
  - Axon: Model delivery + conversion layer (handles ONNX conversion during install)
  - MLOS Core: Pure execution layer (uses pre-converted ONNX files from MPF packages)
  - Complete separation of concerns: MLOS Core has zero Python dependencies

### Benefits
- **Zero Mandatory Dependencies**: Axon works without Python (pure Go when repositories provide ONNX)
- **Better User Experience**: Automatic ONNX conversion during install, no manual steps required
- **MLOS Core Compatibility**: MPF packages include ONNX files ready for MLOS Core execution
- **Graceful Degradation**: Models work even without ONNX conversion (framework-specific plugins available)
- **Patent-Aligned**: Complete decoupling of MLOS Core from conversion logic and Python dependencies

## [1.4.1] - 2024-11-12

### Fixed
- **ModelScope Validation**: Fixed incorrect model validation that was rejecting valid ModelScope models (#18)
  - Made TensorFlow Hub validation check domain-specific (only applies to `tfhub.dev` and `kaggle.com`)
  - ModelScope URLs are now validated correctly
  - Fixes issue where `axon install modelscope/damo/cv_resnet50_image-classification@latest` was incorrectly reported as "model not found"
- **ModelScope Package Creation**: Fixed package creation failure due to missing destination directory
  - Added directory creation before building packages
  - Ensures destination path exists before creating `.axon` package file

### Documentation
- Updated `CONTRIBUTING.md` to emphasize running local validation before marking PRs ready for review
- Added explicit warning: "Never mark a PR as ready if local validation fails!"

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

