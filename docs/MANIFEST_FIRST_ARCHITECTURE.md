# Manifest-First Architecture
## Format-Agnostic Model Package Format (MPF)

## Overview

Axon implements a **manifest-first architecture** where the manifest (`manifest.yaml`) is the **source of truth** for model metadata, including I/O schema and execution format. This enables format-agnostic model execution and future-proof architecture.

## Key Principles

### 1. Manifest as Source of Truth

The manifest contains **all metadata needed for model execution**:
- I/O schema (inputs/outputs with names, shapes, types)
- Execution format (onnx, pytorch, tensorflow)
- Preprocessing requirements (tokenization, normalization)
- Resource requirements
- Framework information

### 2. Format-Agnostic Execution

**Execution format is specified in manifest**, not hardcoded:
- Models can use ONNX, PyTorch, TensorFlow, or other formats
- Core can select appropriate plugin based on manifest
- Easy format transitions without code changes

### 3. Complete I/O Schema

**Manifest includes actual I/O schema**, not generic placeholders:
- Real input names (e.g., `input_ids`, `attention_mask`, `token_type_ids` for BERT)
- Actual shapes and data types
- Preprocessing hints for automatic preprocessing

## Manifest Structure

### Enhanced Format Section

```yaml
spec:
  format:
    type: pytorch              # Original format from repository
    execution_format: onnx     # Execution format (onnx, pytorch, tensorflow)
    files:
      - path: model.onnx
        size: 1234567890
        sha256: "..."
```

### Enhanced I/O Schema

```yaml
spec:
  io:
    inputs:
      - name: input_ids
        dtype: int64
        shape: [-1, -1]  # batch_size, sequence_length
        description: "Token IDs from tokenizer"
        preprocessing:
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
      - name: attention_mask
        dtype: int64
        shape: [-1, -1]
        description: "Attention mask"
        preprocessing:
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
      - name: token_type_ids
        dtype: int64
        shape: [-1, -1]
        description: "Token type IDs"
        preprocessing:
          type: tokenization
          tokenizer: tokenizer.json
          tokenizer_type: bert
    outputs:
      - name: logits
        dtype: float32
        shape: [-1, -1, 30522]  # batch, sequence, vocab_size
        description: "Model logits"
```

## I/O Schema Extraction

Axon automatically extracts I/O schema from model configs:

### Hugging Face Models

For Hugging Face models, Axon:
1. Fetches `config.json` during manifest generation
2. Extracts model type (bert, gpt2, t5, etc.)
3. Generates appropriate I/O schema based on model architecture
4. Adds preprocessing hints for tokenization

**Supported Model Types:**
- BERT-family: `bert`, `roberta`, `distilbert`, `albert`, `electra`
- GPT-family: `gpt2`, `gpt`, `gpt-neo`, `gpt-j`
- T5-family: `t5`, `mt5`, `ul2`
- Vision: `vit`, `deit`, `swin`

### Automatic Detection

Axon detects model type from `config.json`:
```json
{
  "model_type": "bert",
  "vocab_size": 30522,
  ...
}
```

Based on `model_type`, Axon generates:
- Appropriate input names and shapes
- Preprocessing requirements
- Output specifications

## Execution Format Detection

Axon automatically sets `execution_format` based on available files:

1. **ONNX file exists** → `execution_format: onnx`
2. **PyTorch files** → `execution_format: pytorch` (or `onnx` if converted)
3. **TensorFlow files** → `execution_format: tensorflow` (or `onnx` if converted)
4. **Default** → `execution_format: onnx` (most models converted to ONNX)

## Benefits

### 1. Format Independence

- Core doesn't need to know execution format upfront
- Models can use any format specified in manifest
- Easy format migrations

### 2. Future-Proof

- Adding new formats doesn't require Core changes
- Format transitions are non-breaking
- Multi-format support simultaneously

### 3. Complete Metadata

- All information needed for execution in manifest
- No need to inspect model files for metadata
- Fast metadata access (YAML read vs model loading)

### 4. Preprocessing Automation

- Preprocessing hints enable automatic preprocessing
- Tokenization, normalization handled automatically
- No model-specific code needed

## Implementation Details

### I/O Schema Extraction

**Location:** `axon/internal/registry/builtin/io_schema.go`

**Function:** `ExtractIOSchemaFromConfig(configPath string)`

**Process:**
1. Read `config.json` from model
2. Extract `model_type`
3. Generate I/O schema based on model architecture
4. Add preprocessing hints

### Manifest Generation

**Location:** `axon/internal/registry/builtin/huggingface.go`

**Process:**
1. Fetch `config.json` during `GetManifest()`
2. Extract I/O schema using `ExtractIOSchemaFromConfig()`
3. Set `execution_format: onnx` (default)
4. After download/conversion, update manifest with actual format

### Manifest Updates

**After Installation:**
1. Extract package to cache
2. Convert to ONNX (if needed)
3. Update manifest with actual `execution_format`
4. Extract I/O schema from `config.json` (if available)
5. Save updated manifest

## Usage

### Installation

```bash
# Install model - manifest generated automatically
axon install hf/bert-base-uncased@latest

# Manifest includes:
# - Complete I/O schema (input_ids, attention_mask, token_type_ids)
# - Preprocessing hints (tokenization)
# - Execution format (onnx after conversion)
```

### Manifest Inspection

```bash
# View manifest
cat ~/.axon/cache/models/hf/bert-base-uncased/latest/manifest.yaml

# Check I/O schema
yq '.spec.io.inputs' ~/.axon/cache/models/hf/bert-base-uncased/latest/manifest.yaml

# Check execution format
yq '.spec.format.execution_format' ~/.axon/cache/models/hf/bert-base-uncased/latest/manifest.yaml
```

## Integration with MLOS Core

MLOS Core reads the manifest to:
1. **Determine execution format** → Select appropriate plugin
2. **Read I/O schema** → Create proper input tensors
3. **Apply preprocessing** → Use preprocessing hints for tokenization
4. **Validate inputs** → Check against I/O schema

**Benefits:**
- Core is format-agnostic
- No hardcoded format assumptions
- Dynamic plugin selection
- Automatic preprocessing

## Future Enhancements

1. **More Model Types**: Support for additional architectures
2. **Custom Preprocessing**: User-defined preprocessing pipelines
3. **Multi-Format Packages**: Support multiple formats in same package
4. **Schema Validation**: Validate I/O schema against actual model

## References

- [Universal Model Plugin Design](../../core/docs/UNIVERSAL_MODEL_PLUGIN_DESIGN.md)
- [Manifest-First Architecture](../../core/docs/MANIFEST_FIRST_ARCHITECTURE.md)
- [Dynamic Plugin Selection](../../core/docs/DYNAMIC_PLUGIN_SELECTION.md)

