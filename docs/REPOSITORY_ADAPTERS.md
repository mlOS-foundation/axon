# Repository Adapters

Axon supports **pluggable repository adapters** that allow it to work with multiple model repositories in real-time. This enables Axon to download models directly from Hugging Face, local registries, or any custom repository without needing pre-packaged `.axon` files.

## Architecture

Axon uses an **adapter pattern** to support multiple model repositories:

```
┌─────────────────┐
│   axon install  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ AdapterRegistry │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌──────────────┐
│  Local  │ │ Hugging Face │
│ Registry│ │   Adapter    │
│ Adapter │ └──────────────┘
└─────────┘
```

## Supported Adapters

### 1. Local Registry Adapter

**Purpose**: Connect to Axon-compatible registries (local or hosted)

**Usage**:
```bash
# Configure local registry
axon registry set default http://localhost:8080

# Install from local registry
axon install nlp/bert-base-uncased@1.0.0
```

**Features**:
- ✅ Standard Axon registry protocol
- ✅ Manifest-based metadata
- ✅ Pre-packaged `.axon` files
- ✅ Checksum verification
- ✅ Mirror support

### 2. Hugging Face Adapter

**Purpose**: Download models directly from Hugging Face Hub in real-time

**Usage**:
```bash
# Install directly from Hugging Face (no registry needed)
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
axon install hf/microsoft/resnet-50@latest
```

**Features**:
- ✅ Real-time downloads from Hugging Face
- ✅ Automatic package creation (`.axon` format)
- ✅ No pre-packaging required
- ✅ Works with any Hugging Face model
- ✅ Automatic checksum computation
- ✅ Progress tracking

**How it works**:
1. Axon queries Hugging Face API for model metadata
2. Downloads essential files (config.json, model weights, tokenizer, etc.)
3. Creates `.axon` package on-the-fly
4. Computes SHA256 checksum
5. Caches package locally

## Adapter Priority

Adapters are checked in **registration order**:

1. **Local Registry** (if configured) - checked first
2. **Hugging Face** - fallback for any model

The first adapter that `CanHandle()` returns `true` is used.

## Configuration

### Enable/Disable Hugging Face Adapter

```yaml
# ~/.axon/config.yaml
registry:
  url: "http://localhost:8080"
  enable_huggingface: true  # Enable HF adapter
  mirrors: []
  timeout: 300
```

### Disable Hugging Face

```yaml
registry:
  enable_huggingface: false  # Disable HF adapter
```

## Usage Examples

### Example 1: Install from Local Registry

```bash
# Start local registry
cd test/registry
go run server.go .

# Configure Axon
axon registry set default http://localhost:8080

# Install from local registry
axon install nlp/bert-base-uncased@1.0.0
```

### Example 2: Install from Hugging Face (Real-time)

```bash
# No registry configuration needed!
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
axon install hf/microsoft/resnet-50@latest
```

### Example 3: Mixed Usage

```bash
# Configure local registry for curated models
axon registry set default http://localhost:8080

# Install from local registry
axon install nlp/bert-base-uncased@1.0.0

# Install from Hugging Face (adapter automatically handles it)
axon install hf/roberta-base@latest
```

## Creating Custom Adapters

You can create custom adapters for other repositories:

```go
type CustomAdapter struct {
    // Your adapter fields
}

func (c *CustomAdapter) Name() string {
    return "custom"
}

func (c *CustomAdapter) CanHandle(namespace, name string) bool {
    // Return true if this adapter can handle the model
    return namespace == "custom"
}

func (c *CustomAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
    // Fetch/create manifest from your repository
}

func (c *CustomAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
    // Download and package model from your repository
}

func (c *CustomAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
    // Search your repository
}
```

Then register it:
```go
adapterRegistry.Register(&CustomAdapter{})
```

## Benefits

### Real-time Downloads
- ✅ No need to pre-package models
- ✅ Always get latest versions
- ✅ Works with any Hugging Face model

### Flexible Architecture
- ✅ Support multiple repositories
- ✅ Easy to add new adapters
- ✅ Priority-based selection

### Unified Interface
- ✅ Same `axon install` command works for all repositories
- ✅ Consistent manifest format
- ✅ Automatic caching

## Roadmap: Phase 1 Adapters

**Note**: ONNX Model Zoo has been deprecated (July 2025) and models have transitioned to Hugging Face. See [ONNX deprecation notice](https://onnx.ai/models/).

### PyTorch Hub Adapter (Phase 1)

**Status**: In Pipeline  
**Coverage**: ~5% of ML model user base  
**Use Case**: PyTorch pre-trained models for research and experimentation

```bash
# Coming soon in Phase 1
axon install pytorch/resnet50@latest
axon install pytorch/alexnet@latest
axon install pytorch/vgg16@latest
```

### ModelScope Adapter (Phase 1)

**Status**: In Pipeline  
**Coverage**: ~8% of ML model user base (growing rapidly)  
**Use Case**: Multimodal AI models, Chinese language models, enterprise solutions

```bash
# Coming soon in Phase 1
axon install modelscope/damo/nlp_structbert_sentence-similarity_chinese-base@latest
axon install modelscope/ai/modelscope_damo-text-to-video-synthesis@latest
axon install modelscope/cv/resnet50@latest
```

### TensorFlow Hub Adapter (Phase 1)

**Status**: In Pipeline  
**Coverage**: ~7% of ML model user base  
**Use Case**: TensorFlow models for production deployments and Google Cloud ML

```bash
# Coming soon in Phase 1
axon install tfhub/google/imagenet/resnet_v2_50/classification/5@latest
axon install tfhub/google/universal-sentence-encoder/4@latest
axon install tfhub/tensorflow/bert_en_uncased_L-12_H-768_A-12/4@latest
```

### Combined Coverage

- **Hugging Face**: 60%+ of ML practitioners
- **PyTorch Hub**: 5%+ of ML practitioners
- **ModelScope**: 8%+ of ML practitioners (growing)
- **TensorFlow Hub**: 7%+ of ML practitioners
- **Total**: **80%+ of ML model user base**

## Future Adapters (Post-Phase 1)

Potential adapters for future phases:

- **Replicate Adapter**: Hosted inference APIs and model marketplace
- **Kaggle Models Adapter**: Kaggle's model repository and competitions
- **OpenVINO Model Zoo Adapter**: Intel-optimized models for edge deployment
- **PyPI Adapter**: For Python ML packages
- **DagsHub Adapter**: Git-based model versioning and collaboration
- **S3/GCS Adapter**: For cloud-hosted model repositories
- **Private Registry Adapter**: For enterprise/internal registries
- **Git Adapter**: For models stored in Git repositories

## Technical Details

### Hugging Face Implementation

The Hugging Face adapter:
1. Uses Hugging Face API to fetch model metadata
2. Downloads files via direct HTTP (no Python required)
3. Creates tar.gz packages on-the-fly
4. Computes SHA256 checksums
5. Updates manifests with real metadata

### Package Format

All adapters create `.axon` packages:
- Format: `tar.gz` (gzipped tarball)
- Structure: Model files in root directory
- Metadata: Manifest embedded in package

### Error Handling

- If local registry fails, falls back to Hugging Face
- If Hugging Face fails, returns error
- Progress tracking for large downloads
- Automatic retry for transient errors

---

**With repository adapters, Axon becomes a universal model installer that works with any repository!** 🚀

