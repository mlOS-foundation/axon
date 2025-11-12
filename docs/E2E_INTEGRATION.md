# End-to-End Integration: Axon + MLOS Core

This guide explains how Axon and MLOS Core work together to provide a complete model delivery and execution pipeline.

## Overview

**Axon** is the **delivery layer** - it installs and manages ML models from various repositories.  
**MLOS Core** is the **execution layer** - it provides kernel-level model execution via standardized APIs.

Together, they form a complete E2E solution:
1. **Axon** delivers models as standardized packages (Model Package Format - MPF)
2. **MLOS Core** executes models at the kernel level using these packages

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Model Repositories                        │
│  (Hugging Face, PyTorch Hub, TensorFlow Hub, ModelScope)     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                         AXON                                 │
│  • Universal Model Installer                                 │
│  • Standardized Package Format (MPF)                         │
│  • Local Cache Management                                    │
│  • Manifest Generation                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ axon install
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Axon Package (.axon)                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  manifest.yaml  (metadata, framework, resources)   │    │
│  │  model files     (weights, configs, etc.)          │    │
│  │  checksums       (SHA256 verification)             │    │
│  └─────────────────────────────────────────────────────┘    │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ axon register
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                      MLOS CORE                               │
│  • Kernel-Level Execution                                    │
│  • Standard Model Interface (SMI)                           │
│  • Multi-Protocol APIs (HTTP/gRPC/IPC)                      │
│  • Plugin System (PyTorch, TensorFlow, ONNX, etc.)          │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ inference requests
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Applications                              │
│  (Web Services, CLI Tools, Mobile Apps, etc.)                │
└─────────────────────────────────────────────────────────────┘
```

## Step-by-Step Workflow

### 1. Install Model with Axon

```bash
# Install a model from any supported repository
axon install hf/bert-base-uncased@latest

# Axon will:
# - Download model files from Hugging Face
# - Create standardized .axon package
# - Generate manifest.yaml with metadata
# - Store in local cache (~/.axon/cache/)
```

**What Gets Created:**
```
~/.axon/cache/hf/bert-base-uncased/latest/
├── manifest.yaml          # Model metadata, framework, resources
├── model.safetensors      # Model weights
├── config.json           # Model configuration
└── tokenizer.json        # Tokenizer files
```

### 2. Register with MLOS Core

```bash
# Register the installed model with MLOS Core
axon register hf/bert-base-uncased@latest

# Or specify custom MLOS Core endpoint
MLOS_CORE_ENDPOINT=http://localhost:8080 axon register hf/bert-base-uncased@latest
```

**What Happens:**
1. Axon reads the model's `manifest.yaml` from local cache
2. Sends HTTP POST request to MLOS Core's `/models/register` endpoint
3. Includes:
   - Model ID: `hf/bert-base-uncased@latest`
   - Model path: `~/.axon/cache/hf/bert-base-uncased/latest/`
   - Manifest path: `~/.axon/cache/hf/bert-base-uncased/latest/manifest.yaml`
   - Framework, description, and metadata

**MLOS Core Response:**
- Reads the Axon manifest from the provided path
- Extracts metadata (framework, resources, dependencies)
- Prepares the model for kernel-level execution
- Returns registration confirmation with inference endpoint

### 3. Run Inference via MLOS Core

```bash
# Start MLOS Core (if not already running)
mlos_core

# Make inference request via HTTP API
curl -X POST http://localhost:8080/models/hf/bert-base-uncased@latest/inference \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello, world!",
    "input_format": "text"
  }'
```

**What Happens:**
1. MLOS Core receives inference request
2. Loads the model using the path from registration
3. Uses appropriate plugin (e.g., PyTorch plugin for Hugging Face models)
4. Executes inference at kernel level
5. Returns results via API

## Key Design Principles

### 1. Axon as Delivery Layer

- **Axon's Role**: Standardize model delivery from diverse repositories
- **Output**: Consistent `.axon` packages with `manifest.yaml`
- **Benefit**: MLOS Core doesn't need to know about Hugging Face, PyTorch Hub, etc.

### 2. MLOS Core as Execution Layer

- **MLOS Core's Role**: Kernel-level model execution
- **Input**: Standardized Axon packages
- **Benefit**: Plugins work with any model delivered by Axon

### 3. Manifest as Contract

The `manifest.yaml` file is the contract between Axon and MLOS Core:

```yaml
metadata:
  name: bert-base-uncased
  version: latest
  description: BERT base model
spec:
  framework:
    name: pytorch
    version: "2.0.0"
  format:
    files:
      - model.safetensors
      - config.json
  resources:
    cpu:
      min_cores: 2
    memory:
      min_gb: 4
```

MLOS Core reads this manifest to:
- Determine which plugin to use (PyTorch, TensorFlow, etc.)
- Allocate resources (CPU, memory, GPU)
- Validate model files
- Prepare execution environment

## Example: Complete E2E Flow

```bash
# 1. Install model
axon install pytorch/vision/resnet50@latest

# 2. Verify installation
axon list
# Output: pytorch/vision/resnet50@latest

# 3. Get model info
axon info pytorch/vision/resnet50@latest

# 4. Register with MLOS Core
axon register pytorch/vision/resnet50@latest
# Output:
# ✅ Model registered with MLOS Core
#    Model ID: pytorch/vision/resnet50@latest
#    Framework: PyTorch
#    Ready for kernel-level execution

# 5. Start MLOS Core
mlos_core &
# Output:
# 🌟 MLOS Core is running!
# 📡 HTTP API: http://localhost:8080

# 6. Run inference
curl -X POST http://localhost:8080/models/pytorch/vision/resnet50@latest/inference \
  -H "Content-Type: application/json" \
  -d '{"input": "path/to/image.jpg", "input_format": "image"}'

# 7. Check model status
curl http://localhost:8080/models/pytorch/vision/resnet50@latest/status
```

## Benefits

### For Developers

1. **Single Command Installation**: `axon install` works for any repository
2. **Standardized Format**: All models become `.axon` packages
3. **Kernel-Level Performance**: MLOS Core executes at optimal speed
4. **Multi-Protocol APIs**: HTTP, gRPC, or IPC - choose what fits

### For System Administrators

1. **Centralized Management**: All models in one cache location
2. **Resource Awareness**: Manifests specify resource requirements
3. **Security**: Checksums and validation at every step
4. **Scalability**: MLOS Core handles concurrent requests efficiently

### For ML Engineers

1. **Repository Agnostic**: Use models from any source
2. **Framework Flexibility**: PyTorch, TensorFlow, ONNX all supported
3. **Version Control**: Pin specific model versions
4. **Reproducibility**: Standardized packages ensure consistency

## Troubleshooting

### Model Registration Fails

```bash
# Check if MLOS Core is running
curl http://localhost:8080/health

# Check endpoint configuration
echo $MLOS_CORE_ENDPOINT

# Verify model is installed
axon list | grep <model-name>

# Check manifest exists
cat ~/.axon/cache/<namespace>/<name>/<version>/manifest.yaml
```

### MLOS Core Can't Find Manifest

- Ensure `axon register` was run (not just `axon install`)
- Verify manifest path in registration response
- Check file permissions on cache directory

### Inference Fails After Registration

- Verify model files exist at registered path
- Check MLOS Core logs for plugin errors
- Ensure correct framework plugin is loaded
- Validate input format matches model expectations

## Related Documentation

- [Axon-MLOS Integration](https://github.com/mlOS-foundation/core/docs/AXON_MLOS_INTEGRATION.md)
- [Axon-MLOS Architecture](https://github.com/mlOS-foundation/core/docs/AXON_MLOS_ARCHITECTURE.md)
- [MLOS Core API Documentation](https://github.com/mlOS-foundation/core/README.md)
- [Axon Repository Adapters](docs/REPOSITORY_ADAPTERS.md)

## Next Steps

1. **Install Axon**: `curl -sSL axon.mlosfoundation.org | sh`
2. **Install MLOS Core**: See [MLOS Core README](https://github.com/mlOS-foundation/core)
3. **Try the E2E Demo**: See [demo script](https://github.com/mlOS-foundation/core/examples/e2e_axon_mlos_demo.sh)
4. **Build Your Application**: Use MLOS Core APIs for inference

