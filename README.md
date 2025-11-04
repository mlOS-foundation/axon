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

# Search for models (discover neurons)
axon search resnet
axon search "image classification"

# Get model info (inspect the neuron)
axon info vision/resnet50
axon info vision/resnet50@1.0.0

# Install models (propagate signals)
axon install vision/resnet50          # Latest version
axon install vision/resnet50@2.0.0    # Specific version

# List installed (active pathways)
axon list
axon list --namespace vision

# Update model (strengthen the pathway)
axon update vision/resnet50

# Remove model (prune the pathway)
axon uninstall vision/resnet50
```

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

🚧 **Work in Progress** - This is the initial implementation phase.

Current focus: Foundation and core types (Week 1-2 of MVP roadmap)

## License

Apache 2.0

## Links

- **Documentation**: https://axon.mlos.io/docs (coming soon)
- **Registry**: https://registry.axon.mlos.io (coming soon)
- **Discord**: https://discord.gg/mlos
- **Organization**: https://github.com/mlOS-foundation

---

**Part of the [MLOS Foundation](https://mlos.foundation)** - Building the operating system for intelligence.

