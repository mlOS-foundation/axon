# Quick Start Guide

## New Way: Direct Install (Recommended) ⚡

With the adapter system, you can install models **directly from Hugging Face** without any setup:

```bash
# Install any model directly from Hugging Face
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
axon install hf/roberta-base@latest

# No manifest generation needed!
# No package pre-creation needed!
# Everything happens on-the-fly!
```

### How It Works

1. **Axon detects** you want a Hugging Face model (`hf/` namespace)
2. **Downloads files** directly from Hugging Face Hub
3. **Creates manifest** on-the-fly with real metadata
4. **Packages model** into `.axon` format automatically
5. **Caches locally** for future use

### Benefits

- ✅ **No setup required** - works out of the box
- ✅ **Any HF model** - not limited to pre-configured list
- ✅ **Always latest** - get latest versions automatically
- ✅ **Real-time** - manifests and packages created on-demand

## Old Way: Local Registry (Optional)

If you want a **local registry** with pre-packaged models:

```bash
cd test/registry
./bootstrap-top-100.sh  # Creates 100 model manifests (optional)
go run server.go .       # Start local registry
```

Then install from local registry:
```bash
axon registry set default http://localhost:8080
axon install nlp/bert-base-uncased@1.0.0
```

**Note**: This is only needed if you want:
- Local testing of registry server
- Curated model collection
- Offline access to specific models
- Hosted registry deployment

## Comparison

| Feature | Direct Install (New) | Local Registry (Old) |
|---------|---------------------|---------------------|
| Setup Required | ❌ None | ✅ Bootstrap script |
| Model Selection | ✅ Any HF model | ⚠️ Pre-configured list |
| Manifest Creation | ✅ On-the-fly | ⚠️ Pre-generated |
| Package Creation | ✅ On-the-fly | ⚠️ Pre-created |
| Always Latest | ✅ Yes | ❌ Fixed versions |
| Works Offline | ❌ No | ✅ Yes (after download) |

## Recommendation

**For most users**: Use direct install:
```bash
axon install hf/model-name@latest
```

**For advanced users**: Set up local registry only if you need:
- Offline access
- Curated model collection
- Hosted registry deployment

## Examples

### Install from Hugging Face (Recommended)

```bash
# Install BERT
axon install hf/bert-base-uncased@latest

# Install GPT-2
axon install hf/gpt2@latest

# Install any model
axon install hf/microsoft/resnet-50@latest
```

### Install from Local Registry (Optional)

```bash
# Setup local registry (one time)
cd test/registry
./bootstrap-top-100.sh
go run server.go . &

# Configure Axon
axon registry set default http://localhost:8080

# Install from local registry
axon install nlp/bert-base-uncased@1.0.0
```

---

**Bottom line**: Use `axon install hf/model-name@latest` - it's simpler and works with any model! 🚀
