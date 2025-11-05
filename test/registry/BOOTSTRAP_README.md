# Bootstrap Package: Top 100 Hugging Face Models

This package provides a complete bootstrap solution to:
1. Install Axon locally
2. Create a local registry with top 100 Hugging Face models
3. Set up the foundation for a hosted model repository

## Quick Start

```bash
cd test/registry
./bootstrap-top-100.sh
```

This will:
- ✅ Check/build Axon if needed
- ✅ Initialize Axon configuration
- ✅ Create registry structure
- ✅ Generate manifests for 100 popular models
- ✅ Create placeholder packages
- ✅ Compute and update checksums
- ✅ Configure Axon to use local registry

## What Gets Created

### Registry Structure
```
test/registry/
├── api/v1/models/
│   ├── nlp/              # 40 NLP models
│   ├── vision/           # 30 vision models
│   ├── audio/            # 15 audio models
│   └── generation/       # 15 generation models
├── packages/             # 100 placeholder package files
└── server.go             # Registry HTTP server
```

### Models Included

#### NLP Models (40)
- BERT variants (base, large, multilingual)
- GPT models (GPT-2, GPT-Neo)
- Sentence transformers
- Specialized models (sentiment, QA, multilingual)

#### Vision Models (30)
- ResNet family (18, 34, 50, 101, 152)
- Vision Transformers (ViT)
- YOLO models (v5, v8)
- Detection models (DETR, Faster R-CNN, Mask R-CNN)
- Segmentation models (DeepLab, FCN)

#### Audio Models (15)
- Whisper models (base, small, medium, large)
- Wav2Vec2 variants
- HuBERT models
- S2T models

#### Generation Models (15)
- Stable Diffusion variants
- ControlNet models
- Image editing models

## Usage

### 1. Bootstrap the Registry

```bash
./bootstrap-top-100.sh
```

### 2. Start the Registry Server

```bash
go run server.go .
```

### 3. Test with Axon

```bash
# Search for models
axon search bert
axon search vision

# Get model info
axon info nlp/bert-base-uncased@1.0.0

# Install a model
axon install nlp/gpt2@1.0.0
```

### 4. Browse in Browser

Open http://localhost:8080 to see all models with the web UI.

## Customization

### Add More Models

Edit `generate-top-100-manifests.go` and add models to the `top100Models` slice:

```go
{"namespace", "model-name", "1.0.0", "Description", "PyTorch", "2.0.0", "pytorch", "License", []string{"tag1", "tag2"}, "category"},
```

Then regenerate:
```bash
go run generate-top-100-manifests.go .
```

### Update Checksums

After adding real package files:
```bash
go run update-checksums.go .
```

This will:
- Compute SHA256 for all packages
- Update manifest files with correct checksums
- Update package sizes

## Downloading Real Models

### Quick Download (Recommended)

Download a curated set of popular models for immediate use:

```bash
./download-models.sh
```

This downloads:
- BERT, DistilBERT, GPT-2 (NLP)
- ResNet-50 (Vision)
- Additional commonly used models

### Full Download (All 100 Models)

Download all models from the top 100 list:

```bash
./download-all-models.sh
```

**Note**: This downloads ~50-100GB and takes several hours.

### Custom Download

Edit `download_hf_models.py` to customize which models to download.

## Deployment to Hosted Registry

This local registry can be deployed as a hosted model repository:

### 1. Package Structure
```
registry/
├── api/v1/models/        # All manifests
├── packages/              # All model packages (real .axon files)
└── server.go             # HTTP server
```

### 2. Deployment Steps

1. **Download real models** using `download-models.sh` or `download-all-models.sh`
2. **Update checksums** using `update-checksums.go` (done automatically)
3. **Deploy server.go** to your hosting platform
4. **Update registry URLs** in manifests to point to hosted domain
5. **Configure CDN** for package downloads

### 3. Update Manifest URLs

After deployment, update package URLs in manifests:

```bash
# Update localhost:8080 to your domain
sed -i 's|http://localhost:8080|https://registry.axon.mlos.io|g' api/v1/models/**/manifest.yaml
```

### 4. Hosting Options

- **Cloud Storage**: AWS S3, GCS, Azure Blob
- **CDN**: CloudFlare, AWS CloudFront
- **Container**: Docker + Kubernetes
- **Serverless**: AWS Lambda, Vercel, Netlify

## Files

- `bootstrap-top-100.sh` - Main bootstrap script
- `generate-top-100-manifests.go` - Manifest generator for 100 models
- `update-checksums.go` - Checksum updater
- `server.go` - HTTP registry server
- `create-manifests.go` - Original manifest generator (10 models)

## Statistics

After running bootstrap:
- **100 models** across 4 categories
- **100 manifests** in YAML format
- **100 placeholder packages** (ready for real model files)
- **Full metadata** with tags, licenses, descriptions

## Next Steps

1. **Replace placeholders** with actual model files from Hugging Face
2. **Update checksums** automatically
3. **Deploy** to production hosting
4. **Configure** as default Axon registry
5. **Scale** as needed

## Troubleshooting

### Axon not found
```bash
# Build Axon first
cd ../../..
make build
```

### Checksum errors
```bash
# Regenerate checksums
go run update-checksums.go .
```

### Server won't start
```bash
# Check port 8080
lsof -i :8080

# Use different port
PORT=8081 go run server.go .
```

---

**Ready to deploy as a hosted model repository!** 🚀

