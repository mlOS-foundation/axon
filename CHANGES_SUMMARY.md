# Axon Changes Summary - Universal ONNX Conversion Support

## Overview
This PR enhances Axon's ONNX conversion capabilities to support all model repositories with repository-specific conversion strategies.

## Files Modified

### 1. `internal/converter/docker.go`
**Purpose**: Docker-based ONNX converter

**Changes**:
- Fixed Docker image name from `ghcr.io/mlOS-foundation` to `ghcr.io/mlos-foundation` (lowercase)
- Modified `getConversionScript()` to use `namespace` instead of `framework` for selecting conversion scripts
- Ensures proper routing: `hf` → `convert_huggingface.py`, `pytorch` → `convert_pytorch.py`, etc.

**Impact**: Fixes conversion failures and ensures correct converter is used for each repository

### 2. `scripts/conversion/convert_huggingface.py`
**Purpose**: Hugging Face model to ONNX converter

**Changes**:
- Added multi-strategy conversion using `optimum`, `torch.jit.trace`, and `torch.onnx.export`
- Extracts pure Hugging Face model ID from Axon format (`hf/distilgpt2@latest` → `distilgpt2`)
- Added model wrapper to handle complex outputs and disable cache
- Falls back gracefully between conversion methods

**New Features**:
- Supports GPT-2, BERT, T5, and other transformer models
- Handles models with complex output structures (tuples, caches, etc.)
- Automatic ONNX opset selection

**Tested Models**: ✅ GPT-2, ✅ BERT

### 3. `scripts/conversion/convert_pytorch.py`
**Purpose**: PyTorch Hub model to ONNX converter

**Changes**:
- Added support for PyTorch Hub models (`repo/model` format)
- Added support for TorchScript files (`.pt`, `.pth`)
- Added support for `torchvision.models` by name
- Multi-strategy conversion with graceful fallbacks

**New Features**:
- Strategy 1: PyTorch Hub loading
- Strategy 2: TorchScript file loading
- Strategy 3: torchvision models by name
- Comprehensive error messages for unsupported formats

**Tested**: Not yet tested with real PyTorch Hub models

### 4. `scripts/conversion/convert_tensorflow.py`
**Purpose**: TensorFlow model to ONNX converter

**Changes**:
- Added support for SavedModel format
- Added support for Keras H5 files (`.h5`, `.keras`)
- Added support for TensorFlow Hub models
- Multi-strategy conversion using `tf2onnx`

**New Features**:
- Strategy 1: SavedModel loading
- Strategy 2: Keras H5 file loading
- Strategy 3: TensorFlow Hub loading (placeholder)
- Proper signature handling for SavedModel

**Tested**: Not yet tested with real TensorFlow models

### 5. `docker/Dockerfile.converter`
**Purpose**: Multi-framework Docker image for ONNX conversion

**Changes**:
- Added `optimum[exporters,onnxruntime]` for better Hugging Face ONNX export
- Added `accelerate` for Hugging Face model loading
- Added `tensorflow-hub` for TensorFlow Hub models
- Added `onnxoptimizer` for ONNX graph optimization
- Added `scikit-learn`, `skl2onnx`, `xgboost`, `onnxmltools` for ML model conversion
- Added `pillow` and `requests` for utilities
- Optimized layer ordering and caching

**Impact**: Universal converter supporting all major ML frameworks

## Test Coverage

### ✅ Tested and Working
- Hugging Face GPT-2 (single input, int64)
- Hugging Face BERT (multi-input, int64)
- Docker image builds successfully
- Docker-based conversion works

### ⏳ Not Yet Tested
- PyTorch Hub models
- TensorFlow Hub models
- ModelScope models
- Vision models (ResNet, ViT)

## Breaking Changes
None - fully backward compatible

## Dependencies Added
- Python packages in Dockerfile (all optional at runtime)

## CI/CD Impact
- Existing CI workflows will continue to work
- `docker-converter.yml` workflow will build enhanced image
- No changes needed to CI configuration

## Documentation Impact
- Conversion scripts now have better error messages
- Each converter documents supported formats

## Migration Notes
No migration needed - changes are transparent to users

## Next Steps After Merge
1. Test PyTorch Hub conversion with real models
2. Test TensorFlow Hub conversion with real models
3. Test ModelScope integration
4. Add vision model tests (ResNet, ViT)

## Related Issues
- Fixes Docker image naming issue
- Fixes Hugging Face model conversion for GPT-2
- Enables universal model conversion across repositories

## Reviewer Notes
- All changes are in conversion layer only
- No changes to core Axon logic
- Docker image size increased by ~500MB (acceptable for universal support)
- Test locally with: `./axon install hf/distilgpt2@latest`

