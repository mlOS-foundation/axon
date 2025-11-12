# Axon - The Neural Pathway for ML Models

**Tagline**: "Signal. Propagate. Myelinate."

Axon is the transmission layer for ML models in MLOS. It's not just a package manager—it's the neural pathway that connects models to your operating system, optimized at the kernel level for maximum performance.

## Vision

Just as axons transmit signals between neurons to form neural networks, **Axon** provides the transmission layer for ML models in MLOS. It enables seamless model distribution, versioning, and lifecycle management with kernel-level optimizations.

### The Neural Metaphor

- **Neurons** = Individual models
- **Axons** = Transmission pathways (Axon CLI)
- **Myelin sheaths** = Kernel-level optimizations
- **Neural networks** = Model ecosystems
- **MLOS** = The substrate that makes it all work

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/mlOS-foundation/axon.git
cd axon

# Build from source
make build

# Or install directly
make install
```

### Basic Usage

```bash
# Initialize the axon pathway
axon init

# Install any model directly from Hugging Face (no setup needed!)
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
axon install hf/roberta-base@latest

# Install PyTorch Hub models (v1.1.0+)
axon install pytorch/vision/resnet50@latest
axon install pytorch/vision/alexnet@latest

# Install TensorFlow Hub models (v1.2.0+)
axon install tfhub/google/imagenet/resnet_v2_50/classification/5@latest
axon install tfhub/google/universal-sentence-encoder/4@latest

# Install ModelScope models (v1.4.0+)
axon install modelscope/damo/cv_resnet50_image-classification@latest
axon install modelscope/ai/modelscope_damo-text-to-video-synthesis@latest

# Or use local registry (optional)
axon registry set default http://localhost:8080
axon install nlp/bert-base-uncased@1.0.0

# Search for models (discover neurons)
axon search resnet
axon search "image classification"

# Get model info (inspect the neuron)
axon info hf/bert-base-uncased@latest
axon info vision/resnet50@1.0.0

# List installed (active pathways)
axon list

# Update model (strengthen the pathway)
axon update vision/resnet50

# Remove model (prune the pathway)
axon uninstall vision/resnet50
```

## Universal Model Installer

Axon uses a **pluggable adapter architecture** that enables installation from any model repository:

- ✅ **Hugging Face Hub** - Available now (100,000+ models, 60%+ of ML practitioners)
- ✅ **PyTorch Hub** - Available in v1.1.0+ (5%+ coverage, research focus)
- ✅ **TensorFlow Hub** - Available in v1.2.0+ (7%+ coverage, production deployments)
- ✅ **ModelScope** - Available in v1.4.0+ (8%+ coverage, multimodal & enterprise)

**Note**: ONNX Model Zoo has been deprecated (July 2025) and models have transitioned to Hugging Face. See [ONNX deprecation notice](https://onnx.ai/models/).

**Coverage**: According to industry data, Hugging Face alone hosts models used by **60%+ of ML practitioners**, with PyTorch Hub, ModelScope, and TensorFlow Hub covering additional **20%+**, for a total of **80%+ of the ML model user base**. This makes Axon a universal installer that works with virtually any model without vendor lock-in.

### Plug-and-Play Architecture

```bash
# Adapters are automatically selected based on model namespace
axon install hf/model-name@latest           # → Hugging Face adapter
axon install pytorch/vision/resnet50@latest # → PyTorch Hub adapter (v1.1.0+)
axon install tfhub/model-name@latest        # → TensorFlow Hub adapter (v1.2.0+)
axon install modelscope/model-name@latest    # → ModelScope adapter (v1.4.0+)
```

No configuration needed - Axon automatically detects and uses the right adapter!

## Architecture

```
axon/
├── cmd/axon/          # CLI entry point
├── internal/          # Internal packages
│   ├── cache/         # Local cache management
│   ├── config/        # Configuration management
│   ├── manifest/      # Manifest parsing & validation
│   ├── registry/      # Registry HTTP client
│   ├── model/         # Model package handling
│   └── ui/            # CLI UI components
├── pkg/               # Public packages
│   ├── types/         # Public types
│   └── utils/         # Utility functions
└── test/              # Tests
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build scripts)

### Build

```bash
make build          # Build the binary
make test           # Run tests
make lint           # Run linters
make install        # Install to $GOPATH/bin
```

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to Axon.

## Project Status

✅ **MVP Complete** - Core functionality implemented and tested.

The project is in active development. See [MVP_STATUS.md](MVP_STATUS.md) for detailed status.

## License

GNU Affero General Public License v3.0 (AGPL-3.0)

Copyright (C) 2025 MLOS Foundation

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

## Links

- **Repository**: https://github.com/mlOS-foundation/axon
- **Organization**: https://github.com/mlOS-foundation
- **Issues**: https://github.com/mlOS-foundation/axon/issues
- **Discussions**: https://github.com/mlOS-foundation/axon/discussions

---

**Part of the [MLOS Foundation](https://mlos.foundation)** - Building the operating system for intelligence.

