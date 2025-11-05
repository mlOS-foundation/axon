# Axon Local Registry Testing Guide

This directory contains a local static registry server for testing Axon commands with a set of popular ML models.

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
./bin/axon init
./bin/axon registry set default http://localhost:8080
```

### 3. Test Axon Commands

```bash
# Search for models
./bin/axon search bert
./bin/axon search vision
./bin/axon search gpt

# Get model information
./bin/axon info nlp/bert-base-uncased@1.0.0
./bin/axon info nlp/gpt2@1.0.0

# Install a model
./bin/axon install nlp/gpt2@1.0.0
./bin/axon install vision/resnet50@1.0.0

# List installed models
./bin/axon list
```

## Available Models

The registry includes 10 popular models:

### NLP Models
- `nlp/bert-base-uncased@1.0.0` - BERT base model (uncased)
- `nlp/gpt2@1.0.0` - GPT-2 language model
- `nlp/distilbert-base-uncased@1.0.0` - DistilBERT base model
- `nlp/roberta-base@1.0.0` - RoBERTa base model
- `nlp/t5-base@1.0.0` - T5 base model

### Vision Models
- `vision/resnet50@1.0.0` - ResNet-50 image classification
- `vision/vit-base-patch16-224@1.0.0` - Vision Transformer base
- `vision/yolov8n@1.0.0` - YOLOv8 nano object detection

### Audio & Generation
- `audio/whisper-base@1.0.0` - Whisper speech recognition
- `generation/stable-diffusion-2-1@2.1.0` - Stable Diffusion 2.1

## Browser Testing

### Web UI

Open your browser and navigate to:
```
http://localhost:8080
```

You'll see:
- A search interface to find models
- Model cards with details
- Links to view manifests
- Install buttons (shows CLI command)

### Direct API Access

You can also test the API directly in your browser:

**Search API:**
```
http://localhost:8080/api/v1/search?q=bert
http://localhost:8080/api/v1/search?q=vision
```

**Manifest API:**
```
http://localhost:8080/api/v1/models/nlp/bert-base-uncased/1.0.0/manifest.yaml
http://localhost:8080/api/v1/models/nlp/gpt2/1.0.0/manifest.yaml
```

**View Package:**
```
http://localhost:8080/packages/nlp-gpt2-1.0.0.axon
```

## Regenerating Manifests

If you modify the package files, regenerate manifests with correct checksums:

```bash
cd test/registry
go run create-manifests.go .
```

This will:
- Compute SHA256 checksums for all package files
- Update all manifest files with correct checksums
- Preserve existing metadata

## Directory Structure

```
test/registry/
├── server.go              # HTTP server with web UI
├── create-manifests.go     # Manifest generator
├── README.md              # This file
├── api/
│   └── v1/
│       └── models/
│           ├── nlp/
│           │   ├── bert-base-uncased/
│           │   │   └── 1.0.0/
│           │   │       └── manifest.yaml
│           │   └── ...
│           └── ...
└── packages/
    ├── nlp-bert-base-uncased-1.0.0.axon
    └── ...
```

## Testing Workflow

1. **Start the server** (in one terminal):
   ```bash
   cd test/registry
   go run server.go .
   ```

2. **Configure Axon** (in another terminal):
   ```bash
   cd ../..
   ./bin/axon init
   ./bin/axon registry set default http://localhost:8080
   ```

3. **Test via CLI**:
   ```bash
   ./bin/axon search bert
   ./bin/axon info nlp/gpt2@1.0.0
   ./bin/axon install nlp/gpt2@1.0.0
   ```

4. **Test via Browser**:
   - Open http://localhost:8080
   - Search for models
   - Click "View Manifest" to see YAML
   - Click "Install" to see CLI command

## Troubleshooting

### Server won't start
- Check if port 8080 is already in use: `lsof -i :8080`
- Use a different port: `PORT=8081 go run server.go .`
- Update Axon config: `axon registry set default http://localhost:8081`

### Models not found
- Verify manifests exist: `ls -R api/v1/models/`
- Regenerate manifests: `go run create-manifests.go .`
- Check server logs for errors

### Checksum verification fails
- Regenerate manifests with correct checksums: `go run create-manifests.go .`
- Ensure package files exist in `packages/` directory

### Browser shows empty page
- Check browser console for errors
- Verify server is running: `curl http://localhost:8080/api/v1/search?q=bert`
- Check server logs for errors

## Next Steps

1. **Replace placeholder packages** with actual model files from Hugging Face
2. **Add more models** by creating new entries in `create-manifests.go`
3. **Deploy to production** when ready to host a live registry
4. **Add authentication** for protected models
5. **Implement version resolution** for "latest" tag

---

**Happy Testing! 🚀**

