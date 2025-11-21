# Axon v3.0.0 - Universal ONNX Conversion

**Release Date**: November 21, 2024

## 🚀 Major Features

### Universal ONNX Conversion
Axon now supports **universal ONNX conversion** across all major ML model repositories with repository-specific conversion strategies.

#### Multi-Framework Support
- **Hugging Face Hub** 🤗: GPT-2, BERT, T5, RoBERTa, DistilBERT
  - 100,000+ models accessible
  - Multi-strategy conversion using `optimum`, `torch.onnx.export`, and `transformers`
  - Automatic handling of complex outputs (tuples, caches)

- **PyTorch Hub** 🔥: ResNet, VGG, AlexNet, and more
  - TorchScript file support
  - PyTorch Hub model loading
  - torchvision models by name

- **TensorFlow Hub** 🧠: Production-ready models
  - SavedModel format support
  - Keras H5 file conversion
  - TensorFlow Hub integration

- **ModelScope** 🎨: Multimodal AI models
  - Automatic framework detection
  - Comprehensive model support

### Enhanced Conversion Features
- ✅ **Smart Repository Routing**: Automatic converter selection based on model namespace
- ✅ **Multi-Strategy Fallbacks**: Multiple conversion methods for maximum compatibility
- ✅ **Complex Model Support**: Handles models with cache, tuples, and complex outputs
- ✅ **Optimized Docker Image**: Single multi-framework image for all conversions

### Docker Converter Enhancements
- Updated to multi-framework image with:
  - `optimum[exporters,onnxruntime]` for Hugging Face
  - `tf2onnx` for TensorFlow conversion
  - `onnxoptimizer` for graph optimization
  - `scikit-learn`, `skl2onnx`, `xgboost` for traditional ML
  - `accelerate` for advanced model loading

## 📝 What Changed

### Files Modified
- `docker/Dockerfile.converter` - Multi-framework Docker image
- `internal/converter/docker.go` - Repository-specific routing
- `scripts/conversion/convert_huggingface.py` - Enhanced HF conversion
- `scripts/conversion/convert_pytorch.py` - PyTorch Hub support
- `scripts/conversion/convert_tensorflow.py` - TensorFlow conversion
- `README.md` - Updated with Universal ONNX Conversion section

### Files Added
- `CHANGES_SUMMARY.md` - Comprehensive change documentation
- `test-pr-changes.sh` - PR validation test suite

## ✅ Testing

Successfully tested with:
- GPT-2 (DistilGPT-2) from Hugging Face
- BERT (base-uncased) from Hugging Face
- Docker image builds successfully across platforms
- All dependencies verified

## 🔧 Technical Details

### Conversion Strategies

#### Hugging Face
```
1. Try optimum (Hugging Face's official ONNX exporter)
2. Fallback to torch.onnx.export with model wrapper
3. Automatic cache disabling for complex models
```

#### PyTorch Hub
```
1. Try PyTorch Hub loading (repo/model format)
2. Try TorchScript file loading (.pt, .pth)
3. Try torchvision models by name
4. Comprehensive error messages
```

#### TensorFlow
```
1. Try SavedModel format
2. Try Keras H5 file (.h5, .keras)
3. Try TensorFlow Hub models
4. Use tf2onnx converter
```

### Namespace Routing
- `hf/` → `convert_huggingface.py`
- `pytorch/` → `convert_pytorch.py`
- `tfhub/` or `tf/` → `convert_tensorflow.py`
- `modelscope/` or `ms/` → Auto-detect framework

## 💥 Breaking Changes

**None** - This release is fully backward compatible.

All existing functionality continues to work as before. The new conversion strategies are additive and enhance existing capabilities.

## 🔗 Integration

This release integrates seamlessly with **MLOS Core v2.0.0-alpha**, which includes:
- Enhanced ONNX plugin with multi-type tensor support
- Named input parsing for multi-input models
- Support for int64, float32, int32, and bool tensors

## 📊 Performance

- **Conversion Speed**: Similar to previous versions
- **Model Compatibility**: Significantly improved across repositories
- **Docker Image Size**: ~500MB increase (acceptable for universal support)
- **ONNX Runtime**: Optimized models with opset 12-14

## 🐛 Bug Fixes

- Fixed Docker image naming from `mlOS-foundation` to `mlos-foundation` (lowercase)
- Fixed conversion script routing to use namespace instead of framework
- Improved error handling in conversion scripts
- Better handling of Hugging Face model IDs (extract from Axon format)

## 📚 Documentation

- Updated README with comprehensive Universal ONNX Conversion section
- Added CHANGES_SUMMARY.md with detailed file-by-file changes
- Created test-pr-changes.sh for PR validation
- Enhanced error messages in conversion scripts

## 🚀 Upgrade Guide

### From v2.x to v3.0.0

1. **No code changes required** - Fully backward compatible
2. Update to latest version:
   ```bash
   # If installed via script
   curl -sSL https://install.mlosfoundation.org/axon | sh
   
   # Or build from source
   git pull origin main
   make clean && make build
   ```
3. Existing models continue to work without changes
4. New repositories are automatically available

### Docker Converter Image

If you're using the Docker converter directly:
```bash
# Pull latest image
docker pull ghcr.io/mlos-foundation/axon-converter:latest

# Or pull specific version
docker pull ghcr.io/mlos-foundation/axon-converter:3.0.0
```

## 🔮 Future Enhancements

- Additional repository adapters (Replicate, Kaggle)
- Enhanced vision model support (automated preprocessing)
- Audio model conversion (Wav2Vec2, Whisper)
- Multi-modal model optimizations (CLIP, BLIP)
- Quantization support (int8, float16)
- Model optimization passes
- Conversion caching for faster subsequent conversions

## 🙏 Acknowledgments

This release represents a major milestone in MLOS's vision of universal model compatibility. Special thanks to the community for testing and feedback.

## 📞 Support

- GitHub Issues: https://github.com/mlOS-foundation/axon/issues
- Documentation: https://mlosfoundation.org
- Discord: https://discord.gg/mlos (coming soon)

---

**Full Changelog**: https://github.com/mlOS-foundation/axon/compare/v2.0.2...v3.0.0

