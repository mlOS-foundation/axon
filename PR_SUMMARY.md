# PR Summary: Manifest-First Architecture Implementation

## Overview

This PR implements the **manifest-first architecture** for Axon, enabling format-agnostic model execution and future-proof design. The manifest becomes the source of truth for I/O schema and execution format, allowing Core to dynamically select plugins and handle preprocessing automatically.

## Changes Summary

### Core Changes

1. **Enhanced Manifest Types** (`pkg/types/manifest.go`)
   - Added `ExecutionFormat` field to `Format` struct
   - Added `PreprocessingSpec` struct for preprocessing hints
   - Added `Preprocessing` field to `IOSpec` struct

2. **I/O Schema Extraction** (`internal/registry/builtin/io_schema.go`)
   - New file with I/O schema extraction logic
   - Supports BERT, GPT, T5, Vision Transformer models
   - Extracts actual input/output names, shapes, types
   - Adds preprocessing hints based on model type

3. **Hugging Face Adapter** (`internal/registry/builtin/huggingface.go`)
   - Fetches `config.json` during manifest generation
   - Extracts I/O schema from model config
   - Sets `execution_format: onnx` by default
   - Ensures tokenizer files are included in packages

4. **Manifest Updates** (`internal/registry/core/helpers.go`)
   - Added `UpdateManifestWithExecutionFormat()` function
   - Detects execution format from available files
   - Updates manifest after installation

5. **Install Command** (`cmd/axon/commands.go`)
   - Updates manifest after ONNX conversion
   - Extracts I/O schema from config.json if available
   - Saves updated manifest with complete metadata

### Documentation

1. **New Documentation** (`docs/MANIFEST_FIRST_ARCHITECTURE.md`)
   - Complete guide to manifest-first architecture
   - I/O schema extraction details
   - Execution format detection
   - Integration with MLOS Core

2. **Updated CHANGELOG** (`CHANGELOG.md`)
   - Added manifest-first architecture features
   - Documented benefits and use cases

## Key Features

### 1. Execution Format in Manifest

```yaml
spec:
  format:
    type: pytorch
    execution_format: onnx  # NEW: Specifies execution format
```

**Benefit:** Core can dynamically select plugin based on format

### 2. Complete I/O Schema

```yaml
spec:
  io:
    inputs:
      - name: input_ids        # Actual input name (not generic "input")
        dtype: int64
        shape: [-1, -1]
        preprocessing:          # NEW: Preprocessing hints
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
```

**Benefit:** Core knows exactly what inputs model needs and how to create them

### 3. Automatic I/O Schema Extraction

- Extracts from `config.json` during manifest generation
- Supports BERT, GPT, T5, Vision models
- Falls back to generic schema if extraction fails

**Benefit:** No manual I/O schema specification needed

### 4. Preprocessing Hints

- Tokenization requirements
- Image normalization parameters
- Tokenizer file paths

**Benefit:** Enables automatic preprocessing in Core

## Architecture Benefits

### Format Independence

- Core doesn't need to know execution format upfront
- Models can use ONNX, PyTorch, TensorFlow, or other formats
- Easy format transitions without Core changes

### Future-Proof

- Adding new formats doesn't require Core changes
- Format transitions are non-breaking
- Multi-format support simultaneously

### Complete Metadata

- All information needed for execution in manifest
- No need to inspect model files for metadata
- Fast metadata access (YAML read vs model loading)

## Testing

- ✅ Code compiles successfully
- ✅ No linter errors
- ✅ Backward compatible (existing manifests still work)

## Migration Notes

**No migration needed** - This is a new feature that enhances existing manifests:
- Existing manifests continue to work
- New installations get enhanced manifests automatically
- `execution_format` defaults to "onnx" if not specified
- I/O schema extraction is optional (falls back to generic if config.json unavailable)

## Next Steps (Core Repository)

After this PR is merged, Core repository will:
1. Read I/O schema from manifest (not ONNX)
2. Implement dynamic plugin selection based on `execution_format`
3. Implement multi-input tensor creation from manifest I/O schema
4. Integrate preprocessing based on manifest hints

## Files Changed

- `pkg/types/manifest.go` - Enhanced types
- `internal/registry/builtin/io_schema.go` - NEW: I/O schema extraction
- `internal/registry/builtin/huggingface.go` - I/O schema extraction integration
- `internal/registry/core/helpers.go` - Manifest update helpers
- `cmd/axon/commands.go` - Manifest update after installation
- `docs/MANIFEST_FIRST_ARCHITECTURE.md` - NEW: Architecture documentation
- `CHANGELOG.md` - Updated with new features

## Impact

- **Breaking Changes:** None (backward compatible)
- **New Features:** Execution format, preprocessing hints, I/O schema extraction
- **Documentation:** Complete architecture guide added

