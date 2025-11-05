# Local Registry Testing

This directory contains a **simple local registry server** for testing Axon commands with a small set of models.

## Quick Start

### 1. Start the Registry Server

```bash
cd test/registry
go run server.go .
```

The server will start on `http://localhost:8080` and provide:
- 🌐 **Web UI** at `http://localhost:8080` - Browse models in your browser
- 🔍 **Search API** at `http://localhost:8080/api/v1/search?q=<query>`
- 📄 **Manifest API** at `http://localhost:8080/api/v1/models/<namespace>/<name>/<version>/manifest.yaml`
- 📦 **Package API** at `http://localhost:8080/packages/<package-file>.axon`

### 2. Configure Axon to Use Local Registry

```bash
# From the axon directory root
./axon init
./axon registry set default http://localhost:8080
```

### 3. Test Axon Commands

```bash
# Search for models
./axon search bert
./axon search vision

# Get model information
./axon info nlp/bert-base-uncased@1.0.0
./axon info nlp/gpt2@1.0.0

# Install a model
./axon install nlp/gpt2@1.0.0
./axon install vision/resnet50@1.0.0

# List installed models
./axon list
```

## Available Models

The registry includes **10 popular models** for testing:
- BERT, DistilBERT, GPT-2, RoBERTa, T5 (NLP)
- ResNet-50, ViT, YOLOv8 (Vision)
- Whisper (Audio)
- Stable Diffusion (Generation)

## Adding More Models

To add more models for local testing:

```bash
# Generate manifests for additional models
go run create-manifests.go .
```

Edit `create-manifests.go` to add model definitions.

## Note: Direct Install from Hugging Face

**For most use cases**, you don't need a local registry. You can install models directly from Hugging Face:

```bash
# Install directly from Hugging Face (no local registry needed)
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
```

The local registry is only useful for:
- Testing registry server functionality
- Offline access to specific models
- Development and testing

## Hosted Registry

For production hosted registries, a separate pipeline will sync models from Hugging Face to the Axon registry format. This is not part of the core Axon repository.
