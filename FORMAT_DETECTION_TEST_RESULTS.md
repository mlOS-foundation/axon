# Format Detection Test Results

## Overview
Tested format detection across all supported default repositories to verify `execution_format` is correctly set in manifests.

## Test Models

### 1. PyTorch Hub
- **Model**: `pytorch/vision/resnet18@latest`
- **Type**: `pytorch`
- **ExecutionFormat**: `pytorch` ✅
- **Files**: `resnet18-f37072fd.pth`
- **Status**: Correctly detected PyTorch format

### 2. Hugging Face
- **Model**: `hf/distilbert-base-uncased@latest`
- **Type**: `pytorch`
- **ExecutionFormat**: `pytorch` ✅
- **Files**: `pytorch_model.bin`, `model.safetensors`, `tokenizer.json`, etc.
- **Status**: Correctly detected PyTorch format (ONNX conversion not attempted/failed)

### 3. TensorFlow Hub
- **Model**: `tfhub/google/imagenet/mobilenet_v2_100_224/classification/5@latest`
- **Type**: `saved_model`
- **ExecutionFormat**: `tensorflow` ✅
- **Files**: `model.tar.gz` (contains SavedModel)
- **Status**: Correctly detected TensorFlow format using manifest type hint

### 4. ModelScope
- **Model**: `modelscope/damo/cv_resnet18_image-classification@latest`
- **Type**: `modelscope`
- **ExecutionFormat**: `pytorch` ✅
- **Files**: `model.tar.gz`
- **Status**: Correctly detected PyTorch format using manifest type hint

## Detection Logic

The `UpdateManifestWithExecutionFormat` function uses a multi-step approach:

1. **Check for ONNX file** (highest priority)
   - If `model.onnx` exists → `execution_format: "onnx"`

2. **Check for framework-specific files**
   - PyTorch: `.pth`, `.pt`, `.bin`, or files containing "pytorch"
   - TensorFlow: `.pb`, `.h5`, or files containing "tensorflow"/"saved_model"

3. **Check archived models**
   - If `model.tar.gz` exists and manifest type is `saved_model` → `tensorflow`

4. **Use manifest type as hint** (fallback)
   - `pytorch`/`torch` → `pytorch`
   - `tensorflow`/`saved_model`/`tf` → `tensorflow`
   - `modelscope` → `pytorch` (ModelScope models are typically PyTorch)
   - Default → `onnx`

## Results Summary

✅ **All repositories tested successfully**
- All models have `execution_format` set correctly
- Format detection works for:
  - Direct model files (PyTorch `.pth`, Hugging Face `.bin`)
  - Archived models (TensorFlow Hub `.tar.gz`)
  - Manifest type hints (ModelScope)

## Improvements Made

1. **Enhanced file detection**: Added support for `.bin` (PyTorch) and `.h5` (TensorFlow/Keras)
2. **Archive handling**: Detect TensorFlow SavedModel in `.tar.gz` archives
3. **Manifest type fallback**: Use manifest `Format.Type` as hint when files don't match
4. **ModelScope support**: Default ModelScope models to PyTorch format

## Test Command

To run the test:
```bash
cd axon
./test-format-detection.sh
```

Or test individually:
```bash
axon install <namespace>/<model>@latest
# Check manifest:
cat ~/.axon/cache/models/<namespace>/<model>/latest/manifest.yaml | \
  python3 -c "import sys, json; m=json.load(sys.stdin); \
  print(m['Spec']['Format'].get('execution_format', 'NOT SET'))"
```

