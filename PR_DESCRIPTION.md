# Manifest-First Architecture: Format-Agnostic Model Package Format

## 🎯 Overview

This PR implements the **manifest-first architecture** for Axon, enabling format-agnostic model execution and future-proof design. The manifest becomes the **source of truth** for I/O schema and execution format, allowing MLOS Core to dynamically select plugins and handle preprocessing automatically.

## 🚀 Key Features

### 1. Execution Format in Manifest
- Added `execution_format` field to `Format` struct
- Enables dynamic plugin selection in Core based on manifest
- Supports ONNX, PyTorch, TensorFlow formats
- Auto-detected from available model files

### 2. Complete I/O Schema Extraction
- **Automatic extraction** from model `config.json` files
- Extracts **actual input/output names** (e.g., `input_ids`, `attention_mask`, `token_type_ids` for BERT)
- Supports BERT, GPT, T5, Vision Transformer models
- Falls back to generic schema if extraction fails

### 3. Preprocessing Hints
- Added `PreprocessingSpec` to `IOSpec` struct
- Specifies tokenization requirements, tokenizer paths, normalization parameters
- Enables automatic preprocessing in Core

### 4. Manifest Updates After Installation
- Updates manifest with actual `execution_format` after ONNX conversion
- Extracts I/O schema from `config.json` if available
- Ensures tokenizer files are included in packages

## 📋 Changes

### Core Changes

- **`pkg/types/manifest.go`**
  - Added `ExecutionFormat` field to `Format` struct
  - Added `PreprocessingSpec` struct with tokenization, normalization support
  - Added `Preprocessing` field to `IOSpec` struct

- **`internal/registry/builtin/io_schema.go`** (NEW)
  - I/O schema extraction from Hugging Face model configs
  - Supports BERT, GPT, T5, Vision Transformer architectures
  - Generates preprocessing hints based on model type

- **`internal/registry/builtin/huggingface.go`**
  - Fetches `config.json` during manifest generation
  - Extracts I/O schema using `ExtractIOSchemaFromConfig()`
  - Sets `execution_format: onnx` by default
  - Ensures tokenizer files (`tokenizer.json`, `tokenizer_config.json`, `vocab.txt`) are included

- **`internal/registry/core/helpers.go`**
  - Added `UpdateManifestWithExecutionFormat()` function
  - Detects execution format from available files (ONNX, PyTorch, TensorFlow)
  - Updates manifest after installation

- **`cmd/axon/commands.go`**
  - Updates manifest after ONNX conversion
  - Extracts I/O schema from `config.json` if available
  - Saves updated manifest with complete metadata

### Documentation

- **`docs/MANIFEST_FIRST_ARCHITECTURE.md`** (NEW)
  - Complete architecture guide
  - I/O schema extraction details
  - Execution format detection
  - Integration with MLOS Core

- **`CHANGELOG.md`**
  - Added manifest-first architecture features
  - Documented benefits and use cases

## 🎨 Example Manifest

### Before (Generic)
```yaml
spec:
  format:
    type: pytorch
  io:
    inputs:
      - name: input
        dtype: float32
        shape: [-1, -1]
```

### After (Complete)
```yaml
spec:
  format:
    type: pytorch
    execution_format: onnx  # NEW
  io:
    inputs:
      - name: input_ids      # Actual input name
        dtype: int64
        shape: [-1, -1]
        preprocessing:        # NEW: Preprocessing hints
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
      - name: attention_mask
        dtype: int64
        shape: [-1, -1]
        preprocessing:
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
      - name: token_type_ids
        dtype: int64
        shape: [-1, -1]
        preprocessing:
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
```

## ✅ Benefits

### 1. Format Independence
- Core doesn't need to know execution format upfront
- Models can use ONNX, PyTorch, TensorFlow, or other formats
- Easy format transitions without Core changes

### 2. Future-Proof Architecture
- Adding new formats doesn't require Core changes
- Format transitions are non-breaking
- Multi-format support simultaneously
- **If Axon moves away from ONNX, Core requires zero changes**

### 3. Complete Metadata
- All information needed for execution in manifest
- No need to inspect model files for metadata
- Fast metadata access (YAML read vs model loading)

### 4. Preprocessing Automation
- Preprocessing hints enable automatic preprocessing
- Tokenization, normalization handled automatically
- No model-specific code needed in Core

### 5. Dynamic Plugin Selection
- Core selects plugin based on `execution_format` in manifest
- No hardcoded defaults
- Per-model plugin selection

## 🧪 Testing

- ✅ **Build:** `make build` succeeds
- ✅ **Tests:** All tests pass (`go test ./...`)
- ✅ **Linting:** `go vet ./...` passes
- ✅ **Formatting:** Code formatted (`go fmt ./...`)
- ✅ **Backward Compatible:** Existing manifests continue to work

## 🔄 Migration

**No migration needed** - This is a backward-compatible enhancement:
- Existing manifests continue to work
- New installations get enhanced manifests automatically
- `execution_format` defaults to "onnx" if not specified
- I/O schema extraction is optional (falls back to generic if `config.json` unavailable)

## 🔗 Integration with MLOS Core

After this PR is merged, Core repository will:
1. Read I/O schema from manifest (not ONNX) - **format-agnostic**
2. Implement dynamic plugin selection based on `execution_format`
3. Implement multi-input tensor creation from manifest I/O schema
4. Integrate preprocessing based on manifest hints

This enables Core to be **truly format-agnostic** and support multiple execution formats simultaneously.

## 📚 Related Documentation

- [Manifest-First Architecture Guide](docs/MANIFEST_FIRST_ARCHITECTURE.md)
- [Universal Model Plugin Design](../../core/docs/UNIVERSAL_MODEL_PLUGIN_DESIGN.md)
- [Dynamic Plugin Selection](../../core/docs/DYNAMIC_PLUGIN_SELECTION.md)

## 🎯 Impact

- **Breaking Changes:** None (backward compatible)
- **New Features:** Execution format, preprocessing hints, I/O schema extraction
- **Documentation:** Complete architecture guide added
- **Future-Proof:** Enables format transitions without Core changes

---

**This PR enables the manifest-first architecture that makes Core format-agnostic and future-proof.**

