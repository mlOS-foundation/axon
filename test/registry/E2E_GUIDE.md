# End-to-End Usage Guide

This guide shows how to set up a **complete e2e registry** with real model packages that can be used immediately, without needing to visit Hugging Face or other external tools.

## Why Real Packages?

**Placeholder packages** are just for testing the registry infrastructure. For **actual usage**, you need **real model files** that:

- ✅ Can be installed and used immediately
- ✅ Work with MLOS Core for inference
- ✅ Provide complete e2e experience
- ✅ Don't require external tools or websites
- ✅ Are ready for production deployment

## Complete Setup (Recommended)

### Step 1: Bootstrap Registry

```bash
cd test/registry
./bootstrap-top-100.sh
```

This creates:
- Registry structure
- 100 model manifests
- Placeholder packages (for testing infrastructure)

### Step 2: Download Real Models

**Option A: Quick Download (Recommended for testing)**

Downloads a curated set of popular models (~500MB):

```bash
./download-models.sh
```

Downloads:
- BERT, DistilBERT, GPT-2, RoBERTa (NLP)
- ResNet-18, ResNet-50 (Vision)
- Whisper base (Audio)
- Sentence transformers

**Option B: Full Download (Complete registry)**

Downloads all 100 models (~50-100GB, takes hours):

```bash
./download-all-models.sh
```

### Step 3: Start Registry

```bash
go run server.go .
```

### Step 4: Use with Axon

```bash
# Configure Axon
axon registry set default http://localhost:8080

# Search for models
axon search bert

# Install and use (real models!)
axon install nlp/bert-base-uncased@1.0.0
axon install nlp/gpt2@1.0.0
axon install vision/resnet50@1.0.0

# List installed
axon list

# Use models with MLOS Core or other inference engines
```

## What You Get

### Real Model Packages

Each `.axon` package contains:
- ✅ **Model weights** (pytorch_model.bin or model.safetensors)
- ✅ **Configuration files** (config.json)
- ✅ **Tokenizer files** (vocab.txt, tokenizer_config.json)
- ✅ **Complete model** ready for inference

### Package Structure

```
packages/
├── nlp-bert-base-uncased-1.0.0.axon    # Real BERT model
├── nlp-gpt2-1.0.0.axon                 # Real GPT-2 model
├── vision-resnet50-1.0.0.axon         # Real ResNet model
└── ...
```

Each `.axon` file is a gzipped tarball containing the complete model.

### Updated Manifests

Manifests are automatically updated with:
- ✅ Real SHA256 checksums
- ✅ Actual package sizes
- ✅ Correct file listings
- ✅ Valid download URLs

## E2E Workflow Example

```bash
# 1. Setup
cd test/registry
./bootstrap-top-100.sh
./download-models.sh

# 2. Start registry
go run server.go . &

# 3. Configure Axon
axon registry set default http://localhost:8080

# 4. Install model (real download from local registry)
axon install nlp/bert-base-uncased@1.0.0

# 5. Use with MLOS Core or other tools
# Model is now available at ~/.axon/cache/nlp/bert-base-uncased/1.0.0/
```

## Benefits

### Complete E2E Experience
- ✅ No external tools needed
- ✅ Everything works locally
- ✅ Models ready to use immediately
- ✅ No manual downloads from Hugging Face

### Production Ready
- ✅ Real model files
- ✅ Proper checksums
- ✅ Valid packages
- ✅ Ready for deployment

### Development Ready
- ✅ Test with real models
- ✅ Verify installation works
- ✅ Test inference with MLOS Core
- ✅ Complete development workflow

## Customization

### Download Specific Models

Edit `download_hf_models.py`:

```python
MODELS_TO_DOWNLOAD = [
    {"namespace": "nlp", "name": "my-model", "hf_id": "username/model-name", "files": None},
    # Add your models
]
```

Then run:
```bash
python3 download_hf_models.py .
```

### Download More Models

To add more models to the download list, edit `download_hf_models.py` and add entries to `MODELS_TO_DOWNLOAD`.

## Troubleshooting

### Download Fails

If a model download fails:
1. Check internet connection
2. Verify model ID exists on Hugging Face
3. Check disk space (models can be large)
4. Try downloading specific files instead of full model

### Package Too Large

Some models are very large. To download only essential files:

Edit `download_hf_models.py` and specify files:
```python
{"namespace": "nlp", "name": "large-model", "hf_id": "model-id", 
 "files": ["config.json", "tokenizer_config.json"]}  # Skip weights
```

### Update Checksums

After downloading, update checksums:
```bash
go run update-checksums.go .
```

## Next Steps

1. **Use with MLOS Core**: Register downloaded models with MLOS Core for inference
2. **Deploy as Hosted Registry**: Upload packages to cloud storage and deploy server
3. **Add More Models**: Expand the registry with additional models
4. **Share with Team**: Use as internal model registry for your organization

---

**You now have a complete, functional model registry with real packages!** 🎉

