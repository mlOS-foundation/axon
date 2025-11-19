# Axon Converter Image: Universal ONNX Conversion

## Overview

The **Axon Converter Image** (`ghcr.io/mlos-foundation/axon-converter`) is a Docker container image that provides universal ONNX model conversion capabilities for the MLOS ecosystem. It eliminates the need for Python dependencies on the host machine while enabling seamless conversion of models from any repository (Hugging Face, PyTorch Hub, TensorFlow Hub, ModelScope) to ONNX format.

## What It Is

The converter image is a **multi-framework Docker container** that includes:

- **Python 3.11** runtime environment
- **All major ML frameworks** pre-installed:
  - PyTorch 2.9+ (CPU-only to avoid CUDA conflicts)
  - TensorFlow 2.13+ (CPU-only)
  - Transformers 4.30+
  - ModelScope 1.9+
  - ONNX 1.14+
  - ONNX Runtime 1.18+
  - tf2onnx 1.15+
- **Conversion scripts** for each framework:
  - `convert_huggingface.py` - Hugging Face models
  - `convert_pytorch.py` - PyTorch Hub models
  - `convert_tensorflow.py` - TensorFlow Hub models
- **Zero host dependencies** - No Python installation required on your machine

## Why It Exists

### Problem Statement

Traditional ONNX conversion requires:
- Python 3 installed on the host machine
- Framework-specific packages (torch, transformers, tensorflow, etc.)
- Version compatibility management between Python packages and models
- Platform-specific installation issues
- Dependency conflicts between different models' requirements

This creates significant friction:
- Users must install and maintain Python environments
- Version conflicts between different models' requirements
- Platform-specific installation issues (especially on macOS/Windows)
- Maintenance burden for Axon users

### Solution: Docker-Based Conversion

The converter image solves these problems by:

1. **Zero Python Installation**: No Python needed on the host machine
2. **Version Consistency**: Same Python and framework versions across all users
3. **Cross-Platform**: Works on macOS, Linux, Windows (with Docker)
4. **Isolated Environment**: Conversion happens in containers, avoiding conflicts
5. **Reproducibility**: Same conversion environment everywhere

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Host Machine (No Python Required)                      │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Axon CLI (Go binary)                            │  │
│  │  - Downloads model from repository                │  │
│  │  - Detects framework from manifest               │  │
│  │  - Calls Docker for conversion                   │  │
│  └──────────────┬───────────────────────────────────┘  │
│                 │ docker run                            │
│                 ▼                                       │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Volume: ~/.axon/cache → /axon/cache            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│  Docker Container: axon-converter:latest                 │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Python 3.11 + All Frameworks                     │  │
│  │  - PyTorch 2.9+ (CPU-only)                       │  │
│  │  - TensorFlow 2.13+ (CPU-only)                    │  │
│  │  - Transformers 4.30+                            │  │
│  │  - ModelScope 1.9+                               │  │
│  │  - ONNX Runtime (for validation)                  │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Conversion Scripts                               │  │
│  │  - convert_huggingface.py                        │  │
│  │  - convert_pytorch.py                            │  │
│  │  - convert_tensorflow.py                         │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Conversion Flow

1. **Model Download**: Axon downloads the model from the repository (Hugging Face, PyTorch Hub, etc.)
2. **Framework Detection**: Axon detects the framework from the model manifest
3. **Docker Invocation**: Axon calls Docker with the appropriate conversion script
4. **Volume Mapping**: Host cache directory (`~/.axon/cache`) is mapped to `/axon/cache` in container
5. **Conversion**: Python script runs inside container, converting model to ONNX
6. **Output**: ONNX file is written to host cache directory
7. **Verification**: Axon verifies the ONNX file was created successfully

### Volume Mapping Strategy

**Host Path**: `~/.axon/cache` (or `$AXON_CACHE_DIR`)  
**Container Path**: `/axon/cache`

**Benefits**:
- **Version-independent**: Host Axon version doesn't matter
- **Persistent**: Cache survives container restarts
- **Shared**: Multiple Axon instances can share cache
- **No data copying**: Direct file access, no transfer overhead

## Integration with MLOS Core

### The Complete Workflow

```
┌─────────────────────────────────────────────────────────────┐
│  1. Model Installation (Axon)                              │
│     axon install hf/bert-base-uncased@latest               │
│                                                             │
│     ├─ Download model from Hugging Face                   │
│     ├─ Detect framework: Hugging Face / Transformers       │
│     ├─ Convert to ONNX using converter image               │
│     │   └─ docker run axon-converter convert_huggingface.py │
│     └─ Save ONNX file to ~/.axon/cache/.../model.onnx      │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  2. Model Registration (MLOS Core)                         │
│     axon register hf/bert-base-uncased@latest              │
│                                                             │
│     ├─ Read Axon manifest                                  │
│     ├─ Detect ONNX model file                              │
│     ├─ Auto-select ONNX Runtime plugin (built-in)          │
│     └─ Register model with MLOS Core                       │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│  3. Model Execution (MLOS Core)                             │
│     curl -X POST /models/hf/bert-base-uncased/inference    │
│                                                             │
│     ├─ MLOS Core loads ONNX model                          │
│     ├─ ONNX Runtime executes inference                     │
│     └─ Returns results                                     │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**:
   - **Axon**: Model preparation, conversion, packaging
   - **MLOS Core**: Model execution, inference, runtime

2. **Decoupled Architecture**:
   - MLOS Core never needs Python dependencies
   - MLOS Core never performs conversions
   - MLOS Core only executes pre-converted ONNX models

3. **Universal Model Support**:
   - Any model from any repository can be converted
   - ONNX Runtime provides universal execution
   - No framework-specific plugins needed for converted models

## Usage

### Automatic Usage (Recommended)

The converter image is used **automatically** by Axon when Docker is available:

```bash
# Install a model - conversion happens automatically
axon install hf/distilgpt2@latest

# Output:
# 📦 Downloading model from Hugging Face...
# 🐳 Converting model using Docker (ghcr.io/mlos-foundation/axon-converter:latest)...
# ✅ Model converted to ONNX: model.onnx
# ✅ Model installed successfully
```

**First-time setup** (one-time Docker image pull):
```bash
$ axon install hf/distilgpt2@latest
📥 Pulling Docker image: ghcr.io/mlos-foundation/axon-converter:latest (first time only)
# ... downloads ~4GB image (one time only) ...
📦 Downloading model...
🔄 Converting to ONNX...
✅ Model installed successfully
```

### Manual Usage

You can also use the converter image directly:

```bash
# Pull the image
docker pull ghcr.io/mlos-foundation/axon-converter:latest

# Convert a Hugging Face model
docker run --rm \
  -v ~/.axon/cache:/axon/cache \
  -w /axon/cache \
  ghcr.io/mlos-foundation/axon-converter:latest \
  /axon/scripts/convert_huggingface.py \
  model.pth \
  model.onnx \
  bert-base-uncased

# Convert a PyTorch model
docker run --rm \
  -v ~/.axon/cache:/axon/cache \
  -w /axon/cache \
  ghcr.io/mlos-foundation/axon-converter:latest \
  /axon/scripts/convert_pytorch.py \
  model.pth \
  model.onnx \
  pytorch/vision/resnet50
```

### Docker Requirements

- **Docker Desktop** (macOS/Windows) or **Docker Engine** (Linux)
- **Minimum 4GB disk space** for the image
- **Internet connection** for first-time image pull

### Graceful Degradation

If Docker is not available, Axon gracefully falls back:

```bash
$ axon install hf/distilgpt2@latest
⚠️  Docker not available - cannot perform conversion
   💡 To enable ONNX conversion, install Docker: https://docs.docker.com/get-docker/
   💡 Model will work with framework-specific plugins (if available)
✅ Model installed successfully (without ONNX conversion)
```

## Image Details

### Image Specifications

- **Base Image**: `python:3.11-slim`
- **Size**: ~4GB (compressed: ~1.2GB)
- **Platforms**: `linux/amd64`, `linux/arm64`
- **Registry**: GitHub Container Registry (`ghcr.io/mlos-foundation/axon-converter`)
- **Tags**: `latest`, `2.1.0`, `2.1`, `2` (semantic versioning)

### Available Versions

```bash
# Latest version (recommended)
ghcr.io/mlos-foundation/axon-converter:latest

# Specific version
ghcr.io/mlos-foundation/axon-converter:2.1.0

# Major version
ghcr.io/mlos-foundation/axon-converter:2

# Minor version
ghcr.io/mlos-foundation/axon-converter:2.1
```

### OCI Artifacts

The converter image is also available as **OCI artifacts** attached to GitHub Releases:

```bash
# Download OCI artifact from release
wget https://github.com/mlOS-foundation/axon/releases/download/v2.1.0/axon-converter-2.1.0-linux-amd64.tar.gz

# Load into Docker
docker load < axon-converter-2.1.0-linux-amd64.tar.gz
```

## Benefits

### For Users

- ✅ **Zero Python installation**: No need to install Python on host
- ✅ **No dependency management**: Docker handles all Python packages
- ✅ **Version consistency**: Same Python versions across all users
- ✅ **Cross-platform**: Works on macOS, Linux, Windows (with Docker)
- ✅ **Isolated environment**: No conflicts with system Python

### For Axon

- ✅ **Simplified distribution**: No Python dependency documentation
- ✅ **Better UX**: Seamless conversion without user setup
- ✅ **Easier maintenance**: Python deps managed in Docker images
- ✅ **Version control**: Can update Python packages independently

### For MLOS Ecosystem

- ✅ **Cleaner separation**: Host machine stays Python-free
- ✅ **Better isolation**: Conversion in containers
- ✅ **Reproducibility**: Same conversion environment everywhere
- ✅ **Universal execution**: ONNX Runtime enables all models

## Technical Details

### Conversion Scripts

All conversion scripts follow the same interface:

```python
# Script signature
convert_<framework>.py <model_path> <output_path> <model_id>

# Example
convert_huggingface.py model.pth model.onnx bert-base-uncased
```

**Scripts handle**:
- Loading models from their native format
- Converting to ONNX using framework-specific APIs
- Validating ONNX output
- Error handling and reporting

### Framework Support

| Framework | Script | Conversion Method |
|-----------|--------|-------------------|
| Hugging Face | `convert_huggingface.py` | `transformers.onnx.export()` |
| PyTorch | `convert_pytorch.py` | `torch.onnx.export()` |
| TensorFlow | `convert_tensorflow.py` | `tf2onnx.convert()` |
| ModelScope | `convert_huggingface.py` | Uses Transformers API |

### Image Build Process

The converter image is built using:

- **Dockerfile**: `axon/docker/Dockerfile.converter`
- **BuildKit**: Advanced caching for faster builds
- **Multi-platform**: Built for both `linux/amd64` and `linux/arm64`
- **CI/CD**: Automated builds on every release

**Build optimizations**:
- BuildKit cache mounts for pip cache
- GitHub Actions cache for Docker layers
- Parallel platform builds
- Registry cache for better persistence

## Troubleshooting

### Image Not Found

```bash
# Error: Unable to find image 'ghcr.io/mlos-foundation/axon-converter:latest'
# Solution: Pull the image manually
docker pull ghcr.io/mlos-foundation/axon-converter:latest
```

### Conversion Fails

```bash
# Check Docker is running
docker ps

# Check image is available
docker images | grep axon-converter

# Run conversion manually to see errors
docker run --rm -v ~/.axon/cache:/axon/cache \
  ghcr.io/mlos-foundation/axon-converter:latest \
  /axon/scripts/convert_huggingface.py model.pth model.onnx model-id
```

### Out of Disk Space

```bash
# Check Docker disk usage
docker system df

# Clean up unused images
docker image prune -a

# The converter image is ~4GB - ensure you have at least 5GB free
```

### Permission Issues

```bash
# Ensure Docker has permission to access cache directory
chmod -R 755 ~/.axon/cache

# Or set AXON_CACHE_DIR to a writable location
export AXON_CACHE_DIR=/tmp/axon-cache
```

## Future Enhancements

### Planned Improvements

1. **Repository-Specific Images**: Smaller images per repository (PyTorch-only, TensorFlow-only, etc.)
2. **GPU Support**: CUDA-enabled images for faster conversions
3. **Pre-built Base Image**: Ultra-fast builds using pre-built base layers
4. **Kaniko Builds**: Migrate to Kaniko for faster, OCI-native builds

### Performance Optimizations

- **Current**: ~15-20 min first build, ~3-5 min subsequent builds
- **Target**: ~8-12 min first build, ~30-60 sec subsequent builds
- **Strategy**: Pre-built base image + Kaniko builds

## References

- [Docker Documentation](https://docs.docker.com/)
- [ONNX Conversion Guide](https://onnx.ai/onnx/intro/concepts.html)
- [Axon Documentation](../README.md)
- [MLOS Core ONNX Integration](../../core/docs/ONNX_RUNTIME_INTEGRATION.md)

