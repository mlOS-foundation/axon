# Docker-Based ONNX Conversion Design

## Overview

This document explores design options for using Docker-based execution for Axon's ONNX conversion, eliminating the need for Python dependencies on the host machine.

## Problem Statement

Currently, Axon's ONNX conversion requires:
- Python 3 installed on the host
- Framework-specific packages (torch, transformers, tf2onnx, etc.)
- Version compatibility between Python packages and models

This creates friction:
- Users must install Python and manage dependencies
- Version conflicts between different models' requirements
- Platform-specific installation issues
- Maintenance burden for Axon users

## Solution: Docker-Based Conversion

### Core Concept

Use Docker containers to execute Python-based conversions, with:
- **Version-independent volume mapping**: Host Axon cache mapped to container
- **Docker image controls Python deps**: Pre-configured images with all frameworks
- **Repository-driven dependencies**: Image selection based on supported repositories
- **Zero host Python requirement**: No Python needed on host machine

## Design Options

### Option 1: Single Multi-Framework Docker Image (Recommended)

**Concept**: One Docker image with all Python frameworks pre-installed.

#### Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Host Machine (No Python Required)                      │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Axon CLI (Go binary)                            │  │
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
│  │  - PyTorch 2.0+                                   │  │
│  │  - Transformers 4.30+                            │  │
│  │  - TensorFlow 2.13+                               │  │
│  │  - tf2onnx                                        │  │
│  │  - ONNX Runtime (for validation)                  │  │
│  └──────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Conversion Scripts                               │  │
│  │  - convert_pytorch.py                            │  │
│  │  - convert_tensorflow.py                        │  │
│  │  - convert_huggingface.py                       │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

#### Implementation

**Dockerfile** (`axon/Dockerfile.converter`):
```dockerfile
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install all Python ML frameworks
RUN pip install --no-cache-dir \
    torch>=2.0.0 \
    transformers>=4.30.0 \
    tensorflow>=2.13.0 \
    tf2onnx>=1.15.0 \
    onnxruntime>=1.18.0 \
    onnx>=1.14.0

# Create conversion scripts directory
WORKDIR /axon/scripts

# Copy conversion scripts
COPY scripts/convert_pytorch.py .
COPY scripts/convert_tensorflow.py .
COPY scripts/convert_huggingface.py .

# Set entrypoint
ENTRYPOINT ["python3"]
```

**Go Implementation** (`axon/internal/converter/docker.go`):
```go
package converter

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
)

// ConvertToONNXWithDocker uses Docker for Python-based conversion
func ConvertToONNXWithDocker(ctx context.Context, modelPath, framework, namespace, modelID, outputPath string) (bool, error) {
    // Step 1: Try pure Go download first
    if namespace != "" && modelID != "" {
        downloaded, err := DownloadPreConvertedONNX(ctx, namespace, modelID, outputPath)
        if err == nil && downloaded {
            return true, nil
        }
    }

    // Step 2: Use Docker for conversion
    // Map host cache directory to container
    cacheDir := filepath.Dir(modelPath)
    hostCacheDir := os.Getenv("AXON_CACHE_DIR")
    if hostCacheDir == "" {
        hostCacheDir = filepath.Join(os.Getenv("HOME"), ".axon", "cache")
    }

    // Determine conversion script based on framework
    frameworkLower := strings.ToLower(framework)
    var scriptName string
    switch {
    case frameworkLower == "huggingface" || frameworkLower == "transformers":
        scriptName = "convert_huggingface.py"
    case frameworkLower == "pytorch" || frameworkLower == "torch":
        scriptName = "convert_pytorch.py"
    case frameworkLower == "tensorflow" || frameworkLower == "tf":
        scriptName = "convert_tensorflow.py"
    default:
        return false, fmt.Errorf("unsupported framework: %s", framework)
    }

    // Build Docker command
    dockerCmd := exec.CommandContext(ctx, "docker", "run", "--rm",
        "-v", fmt.Sprintf("%s:/axon/cache", hostCacheDir),
        "-w", "/axon/cache",
        "axon-converter:latest",
        fmt.Sprintf("/axon/scripts/%s", scriptName),
        modelPath,
        outputPath,
        modelID,
    )

    output, err := dockerCmd.CombinedOutput()
    if err != nil {
        return false, fmt.Errorf("docker conversion failed: %w\nOutput: %s", err, string(output))
    }

    // Verify output file exists
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        return false, fmt.Errorf("conversion output file not created: %s", outputPath)
    }

    return true, nil
}
```

#### Pros
- ✅ Simple: One image, one command
- ✅ Works for all frameworks
- ✅ Easy to maintain
- ✅ Fast startup (image pre-built)

#### Cons
- ⚠️ Large image size (~2-3GB with all frameworks)
- ⚠️ May include unused frameworks for specific models

---

### Option 2: Framework-Specific Docker Images

**Concept**: Separate Docker images per framework, selected dynamically.

#### Architecture

```
Host Axon CLI
    │
    ├─→ PyTorch models → axon-converter-pytorch:latest
    ├─→ TensorFlow models → axon-converter-tensorflow:latest
    └─→ HuggingFace models → axon-converter-huggingface:latest
```

#### Implementation

**Image Structure**:
- `axon-converter-pytorch:latest` - PyTorch + transformers
- `axon-converter-tensorflow:latest` - TensorFlow + tf2onnx
- `axon-converter-huggingface:latest` - Transformers + PyTorch (subset)

**Go Implementation**:
```go
func getDockerImageForFramework(framework string) string {
    frameworkLower := strings.ToLower(framework)
    switch {
    case frameworkLower == "pytorch" || frameworkLower == "torch":
        return "axon-converter-pytorch:latest"
    case frameworkLower == "tensorflow" || frameworkLower == "tf":
        return "axon-converter-tensorflow:latest"
    case frameworkLower == "huggingface" || frameworkLower == "transformers":
        return "axon-converter-huggingface:latest"
    default:
        return "axon-converter-pytorch:latest" // Default
    }
}
```

#### Pros
- ✅ Smaller images (~500MB-1GB each)
- ✅ Faster pulls for specific frameworks
- ✅ Better resource usage

#### Cons
- ⚠️ More complex: Multiple images to maintain
- ⚠️ Need to pull correct image per model

---

### Option 3: Repository-Driven Image Selection

**Concept**: Docker images aligned with supported repositories, with dependencies driven by repository requirements.

#### Architecture

```
Supported Repositories → Required Dependencies → Docker Images

Hugging Face (hf/)     → transformers, torch        → axon-converter-hf:latest
PyTorch Hub (pytorch/) → torch                      → axon-converter-pytorch:latest
TensorFlow Hub (tfhub/→ tensorflow, tf2onnx        → axon-converter-tfhub:latest
ModelScope (ms/)       → modelscope, torch         → axon-converter-ms:latest
```

#### Implementation

**Repository-to-Image Mapping**:
```go
// Map repository namespace to Docker image
var repositoryImageMap = map[string]string{
    "hf":       "axon-converter-hf:latest",        // Hugging Face
    "pytorch":  "axon-converter-pytorch:latest",   // PyTorch Hub
    "tfhub":    "axon-converter-tfhub:latest",     // TensorFlow Hub
    "ms":       "axon-converter-ms:latest",       // ModelScope
}

func getDockerImageForRepository(namespace string) string {
    if image, ok := repositoryImageMap[namespace]; ok {
        return image
    }
    // Fallback to multi-framework image
    return "axon-converter:latest"
}
```

**Dockerfile per Repository** (`axon/docker/axon-converter-hf/Dockerfile`):
```dockerfile
FROM python:3.11-slim

# Hugging Face specific dependencies
RUN pip install --no-cache-dir \
    transformers>=4.30.0 \
    torch>=2.0.0 \
    onnxruntime>=1.18.0 \
    onnx>=1.14.0

WORKDIR /axon/scripts
COPY scripts/convert_huggingface.py .

ENTRYPOINT ["python3"]
```

#### Pros
- ✅ Optimized for each repository
- ✅ Minimal dependencies per image
- ✅ Aligns with Axon's repository architecture
- ✅ Easy to extend for new repositories

#### Cons
- ⚠️ Multiple images to build/maintain
- ⚠️ Need to ensure image availability

---

### Option 4: Hybrid Approach (Recommended for MVP)

**Concept**: Start with single multi-framework image, evolve to repository-specific images.

#### Phases

**Phase 1 (MVP)**: Single image with all frameworks
- Quick to implement
- Works for all repositories
- Validates the concept

**Phase 2 (Optimization)**: Repository-specific images
- Based on usage patterns
- Optimize for most common repositories
- Keep single image as fallback

#### Implementation Strategy

```go
func ConvertToONNX(ctx context.Context, modelPath, framework, namespace, modelID, outputPath string) (bool, error) {
    // Step 1: Pure Go download (always try first)
    if namespace != "" && modelID != "" {
        downloaded, err := DownloadPreConvertedONNX(ctx, namespace, modelID, outputPath)
        if err == nil && downloaded {
            return true, nil
        }
    }

    // Step 2: Docker-based conversion
    // Try repository-specific image first, fallback to multi-framework
    image := getDockerImageForRepository(namespace)
    if image == "" {
        image = "axon-converter:latest" // Fallback
    }

    return convertWithDocker(ctx, image, modelPath, framework, namespace, modelID, outputPath)
}
```

---

## Implementation Details

### Volume Mapping Strategy

**Host Path**: `~/.axon/cache` (or `$AXON_CACHE_DIR`)
**Container Path**: `/axon/cache`

**Benefits**:
- Version-independent: Host Axon version doesn't matter
- Persistent: Cache survives container restarts
- Shared: Multiple Axon instances can share cache

### Docker Image Management

**Build Strategy**:
```bash
# Build all converter images
make docker-build-converters

# Or build specific repository image
make docker-build-converter-hf
make docker-build-converter-pytorch
```

**Image Registry**:
- Option A: GitHub Container Registry (GHCR)
- Option B: Docker Hub
- Option C: Local build (for development)

### Conversion Scripts

**Location**: `axon/scripts/conversion/`

**Structure**:
```
axon/scripts/conversion/
├── convert_huggingface.py
├── convert_pytorch.py
├── convert_tensorflow.py
└── common.py (shared utilities)
```

**Script Interface**:
```python
# convert_huggingface.py
import sys
import os

def convert(model_path, output_path, model_id):
    # Conversion logic
    pass

if __name__ == "__main__":
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    model_id = sys.argv[3]
    convert(model_path, output_path, model_id)
```

### Error Handling

**Docker Availability Check**:
```go
func isDockerAvailable() bool {
    cmd := exec.Command("docker", "version")
    return cmd.Run() == nil
}

func ConvertToONNX(...) (bool, error) {
    // Try pure Go first
    if downloaded, _ := DownloadPreConvertedONNX(...); downloaded {
        return true, nil
    }

    // Check Docker availability
    if !isDockerAvailable() {
        return false, fmt.Errorf("Docker not available - cannot perform conversion")
    }

    // Use Docker conversion
    return ConvertToONNXWithDocker(...)
}
```

### Graceful Degradation

**Fallback Chain**:
1. Pure Go download (pre-converted ONNX)
2. Docker-based conversion (if Docker available)
3. Skip conversion (graceful degradation)
   - Model still works with framework-specific plugins
   - User can convert manually later

---

## Repository Dependency Matrix

| Repository | Namespace | Required Python Packages | Docker Image |
|------------|-----------|-------------------------|--------------|
| Hugging Face | `hf/` | transformers, torch, onnx | `axon-converter-hf:latest` |
| PyTorch Hub | `pytorch/` | torch, onnx | `axon-converter-pytorch:latest` |
| TensorFlow Hub | `tfhub/` | tensorflow, tf2onnx, onnx | `axon-converter-tfhub:latest` |
| ModelScope | `ms/` | modelscope, torch, onnx | `axon-converter-ms:latest` |

---

## User Experience

### Before (Current)
```bash
$ axon install hf/distilgpt2@latest
⚠️  Python3 not found - skipping ONNX conversion
   💡 To enable ONNX conversion, install Python 3 and: pip install transformers torch
```

### After (Docker-Based)
```bash
$ axon install hf/distilgpt2@latest
📦 Downloading model...
🔄 Converting to ONNX using Docker (axon-converter-hf:latest)...
✅ Model converted to ONNX: model.onnx
✅ Model installed successfully
```

**First-time setup** (one-time):
```bash
$ axon install hf/distilgpt2@latest
🐳 Pulling Docker image: axon-converter-hf:latest (first time only)
📦 Downloading model...
🔄 Converting to ONNX...
✅ Model installed successfully
```

---

## Implementation Plan

### Phase 1: Single Multi-Framework Image (MVP)

1. **Create Dockerfile** (`axon/Dockerfile.converter`)
   - Python 3.11 base
   - All frameworks installed
   - Conversion scripts

2. **Implement Docker converter** (`axon/internal/converter/docker.go`)
   - Docker availability check
   - Volume mapping
   - Command execution

3. **Update ConvertToONNX** (`axon/internal/converter/onnx.go`)
   - Try Docker before local Python
   - Fallback to local Python if Docker unavailable

4. **Add conversion scripts** (`axon/scripts/conversion/`)
   - Python scripts for each framework
   - Standardized interface

5. **Update Makefile**
   - `make docker-build-converter` target
   - Image tagging and versioning

### Phase 2: Repository-Specific Images (Optimization)

1. **Create repository-specific Dockerfiles**
2. **Implement image selection logic**
3. **Update build process**
4. **Add image caching strategy**

---

## Benefits

### For Users
- ✅ **Zero Python installation**: No need to install Python on host
- ✅ **No dependency management**: Docker handles all Python packages
- ✅ **Version consistency**: Same Python versions across all users
- ✅ **Cross-platform**: Works on macOS, Linux, Windows (with Docker)

### For Axon
- ✅ **Simplified distribution**: No Python dependency documentation
- ✅ **Better UX**: Seamless conversion without user setup
- ✅ **Easier maintenance**: Python deps managed in Docker images
- ✅ **Version control**: Can update Python packages independently

### For MLOS Ecosystem
- ✅ **Cleaner separation**: Host machine stays Python-free
- ✅ **Better isolation**: Conversion in containers
- ✅ **Reproducibility**: Same conversion environment everywhere

---

## Considerations

### Docker Requirement
- **Requirement**: Docker must be installed
- **Mitigation**: Graceful degradation if Docker unavailable
- **User experience**: Clear error messages with installation instructions

### Image Size
- **Multi-framework image**: ~2-3GB
- **Repository-specific**: ~500MB-1GB each
- **Mitigation**: Use multi-stage builds, layer caching

### Performance
- **First pull**: Slower (downloads image)
- **Subsequent**: Fast (uses cached image)
- **Conversion**: Same speed as local Python

### Security
- **Image source**: Use trusted registries (GHCR, Docker Hub)
- **Image signing**: Sign images for verification
- **Content scanning**: Scan for vulnerabilities

---

## Next Steps

1. **Prototype**: Implement Option 1 (single image) as MVP
2. **Test**: Validate with multiple repositories
3. **Optimize**: Move to repository-specific images if needed
4. **Document**: Update user documentation
5. **Release**: Include in next Axon release

---

## References

- [Docker Volume Mounting](https://docs.docker.com/storage/volumes/)
- [Python Docker Images](https://hub.docker.com/_/python)
- [ONNX Conversion Best Practices](https://onnx.ai/onnx/intro/concepts.html)

