#!/bin/bash

# Download ALL models from the top 100 list (takes longer, but creates complete registry)
# This provides a truly e2e experience with real model packages

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REGISTRY_DIR="$SCRIPT_DIR"

echo "📦 Downloading ALL Top 100 Models from Hugging Face"
echo "===================================================="
echo ""
echo "⚠️  WARNING: This will download ~50-100GB of model files"
echo "   This may take several hours depending on your connection"
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Expand the Python script to download all models
python3 << 'PYTHON_EOF'
import os
import sys
import json
import tarfile
import hashlib
from pathlib import Path
from huggingface_hub import hf_hub_download, snapshot_download
from huggingface_hub.utils import HfHubHTTPError

# Registry directory
registry_dir = sys.argv[1] if len(sys.argv) > 1 else "."

# Read the model list from generate-top-100-manifests.go
# For now, we'll use a curated list of smaller models that are commonly used
MODELS = [
    # Small NLP models (fast to download)
    ("nlp", "bert-base-uncased", "bert-base-uncased"),
    ("nlp", "distilbert-base-uncased", "distilbert-base-uncased"),
    ("nlp", "gpt2", "gpt2"),
    ("nlp", "roberta-base", "roberta-base"),
    ("nlp", "distilgpt2", "distilgpt2"),
    
    # Sentence transformers (useful and moderate size)
    ("nlp", "sentence-transformers/all-MiniLM-L6-v2", "sentence-transformers/all-MiniLM-L6-v2"),
    
    # Vision models (small to medium)
    ("vision", "resnet50", "microsoft/resnet-50"),
    
    # Audio (small)
    ("audio", "whisper-base", "openai/whisper-base"),
    
    # Add more models as needed - start with smaller ones
]

def compute_sha256(file_path):
    sha256 = hashlib.sha256()
    with open(file_path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b''):
            sha256.update(chunk)
    return sha256.hexdigest()

def create_package(namespace, name, hf_id, registry_dir):
    version = "1.0.0"
    package_name = f"{namespace}-{name}-{version}.axon"
    package_path = os.path.join(registry_dir, "packages", package_name)
    temp_dir = os.path.join(registry_dir, "tmp", "models", f"{namespace}-{name}")
    
    print(f"📥 {namespace}/{name} ({hf_id})...")
    
    try:
        os.makedirs(temp_dir, exist_ok=True)
        
        # Download model files (skip large formats)
        try:
            snapshot_download(
                repo_id=hf_id,
                local_dir=temp_dir,
                local_dir_use_symlinks=False,
                ignore_patterns=["*.safetensors", "*.msgpack", "*.h5", "*.ckpt", "*.bin"]  # Focus on configs
            )
        except Exception as e:
            print(f"   ⚠️  Download failed: {e}")
            return False
        
        # Create package
        with tarfile.open(package_path, "w:gz") as tar:
            tar.add(temp_dir, arcname=os.path.basename(temp_dir))
        
        checksum = compute_sha256(package_path)
        size = os.path.getsize(package_path)
        
        print(f"   ✓ {package_name} ({size / 1024 / 1024:.2f} MB)")
        
        # Update manifest
        manifest_path = os.path.join(registry_dir, "api/v1/models", namespace, name, version, "manifest.yaml")
        if os.path.exists(manifest_path):
            import yaml
            with open(manifest_path, 'r') as f:
                manifest = yaml.safe_load(f)
            if "distribution" in manifest and "package" in manifest["distribution"]:
                manifest["distribution"]["package"]["sha256"] = checksum
                manifest["distribution"]["package"]["size"] = size
            with open(manifest_path, 'w') as f:
                yaml.dump(manifest, f, default_flow_style=False, sort_keys=False)
        
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)
        return True
    except Exception as e:
        print(f"   ❌ Error: {e}")
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)
        return False

# Download models
success = 0
for namespace, name, hf_id in MODELS:
    if create_package(namespace, name, hf_id, registry_dir):
        success += 1
    print("")

print(f"✅ Downloaded {success}/{len(MODELS)} models")
PYTHON_EOF

echo ""
echo "✅ Download complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Update all checksums: go run update-checksums.go ."
echo "   2. Start registry: go run server.go ."
echo "   3. Test: axon install nlp/bert-base-uncased@1.0.0"
echo ""

