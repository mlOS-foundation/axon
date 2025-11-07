# Axon v1.0.0 Release Notes

## 🎉 MVP Complete - First Stable Release

Axon v1.0.0 marks the completion of the MVP (Minimum Viable Product) with core functionality ready for production use.

## What's New

### Core Features
- ✅ **Universal Model Installer** - Install models from any repository via pluggable adapters
- ✅ **Hugging Face Integration** - Zero-configuration support for Hugging Face models
- ✅ **Model Lifecycle Management** - Install, list, search, update, and uninstall models
- ✅ **Manifest System** - YAML-based manifests with validation
- ✅ **Local Caching** - Intelligent caching for offline access
- ✅ **Registry Client** - Model discovery across repositories
- ✅ **CLI Interface** - Comprehensive command-line interface

### Installation

One-line installation:

```bash
curl -sSL axon.mlosfoundation.org | sh
```

Or from GitHub:

```bash
curl -sSL https://raw.githubusercontent.com/mlOS-foundation/axon/main/install.sh | sh
```

### Quick Start

```bash
# Install a model
axon install hf/bert-base-uncased@latest

# List installed models
axon list

# Search for models
axon search resnet

# Get model information
axon info hf/bert-base-uncased@latest
```

## Supported Platforms

- **macOS**: amd64, arm64
- **Linux**: amd64, arm64

## Download

Download binaries from the [Releases](https://github.com/mlOS-foundation/axon/releases/tag/v1.0.0) page:

- `axon_1.0.0_darwin_amd64.tar.gz`
- `axon_1.0.0_darwin_arm64.tar.gz`
- `axon_1.0.0_linux_amd64.tar.gz`
- `axon_1.0.0_linux_arm64.tar.gz`

## What's Next

- Phase 1: MLOS Core Runtime integration
- Additional repository adapters (ONNX Model Zoo, PyTorch Hub)
- Enhanced caching and optimization
- Model versioning and A/B testing

## Documentation

- [README](README.md) - Project overview and quick start
- [Contributing](CONTRIBUTING.md) - How to contribute
- [MVP Status](MVP_STATUS.md) - Current implementation status

## Thank You

Thank you for using Axon! We're excited to see what you build with it.

---

**Signal. Propagate. Myelinate.** 🧠

