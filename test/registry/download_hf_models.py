#!/usr/bin/env python3
"""
Download and package models from Hugging Face for Axon registry.
This creates actual .axon packages that can be used end-to-end.
"""

import os
import sys
import json
import tarfile
import hashlib
from pathlib import Path
from huggingface_hub import hf_hub_download, snapshot_download
from huggingface_hub.utils import HfHubHTTPError

# Top models to download - real models for e2e experience
# These are actual model files that can be installed and used immediately
MODELS_TO_DOWNLOAD = [
    # Small NLP models (fast download, commonly used)
    {"namespace": "nlp", "name": "bert-base-uncased", "hf_id": "bert-base-uncased", "files": ["config.json", "pytorch_model.bin", "vocab.txt", "tokenizer_config.json"]},
    {"namespace": "nlp", "name": "distilbert-base-uncased", "hf_id": "distilbert-base-uncased", "files": ["config.json", "pytorch_model.bin", "vocab.txt", "tokenizer_config.json"]},
    {"namespace": "nlp", "name": "gpt2", "hf_id": "gpt2", "files": ["config.json", "pytorch_model.bin", "vocab.json", "merges.txt", "tokenizer_config.json"]},
    {"namespace": "nlp", "name": "roberta-base", "hf_id": "roberta-base", "files": ["config.json", "pytorch_model.bin", "vocab.json", "merges.txt"]},
    {"namespace": "nlp", "name": "distilgpt2", "hf_id": "distilgpt2", "files": ["config.json", "pytorch_model.bin", "vocab.json", "merges.txt"]},
    
    # Sentence transformers (useful for embeddings)
    {"namespace": "nlp", "name": "sentence-transformers/all-MiniLM-L6-v2", "hf_id": "sentence-transformers/all-MiniLM-L6-v2", "files": None},  # Download all files
    
    # Vision models
    {"namespace": "vision", "name": "resnet50", "hf_id": "microsoft/resnet-50", "files": None},
    {"namespace": "vision", "name": "resnet18", "hf_id": "microsoft/resnet-18", "files": None},
    
    # Audio models
    {"namespace": "audio", "name": "whisper-base", "hf_id": "openai/whisper-base", "files": None},
    
    # Add more models as needed - start with smaller ones for faster setup
]

def compute_sha256(file_path):
    """Compute SHA256 hash of a file."""
    sha256 = hashlib.sha256()
    with open(file_path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b''):
            sha256.update(chunk)
    return sha256.hexdigest()

def create_axon_package(model_info, registry_dir):
    """Download model from HF and create .axon package."""
    namespace = model_info["namespace"]
    name = model_info["name"]
    hf_id = model_info["hf_id"]
    version = "1.0.0"
    
    package_name = f"{namespace}-{name}-{version}.axon"
    package_path = os.path.join(registry_dir, "packages", package_name)
    temp_dir = os.path.join(registry_dir, "tmp", "models", f"{namespace}-{name}")
    
    print(f"📥 Downloading {namespace}/{name} ({hf_id})...")
    
    try:
        # Create temp directory
        os.makedirs(temp_dir, exist_ok=True)
        
        # Download specific files if specified, otherwise download all
        if "files" in model_info and model_info["files"] and model_info["files"] is not None:
            for file_name in model_info["files"]:
                try:
                    local_path = hf_hub_download(
                        repo_id=hf_id,
                        filename=file_name,
                        local_dir=temp_dir,
                        local_dir_use_symlinks=False
                    )
                    print(f"   ✓ Downloaded {file_name}")
                except HfHubHTTPError as e:
                    print(f"   ⚠️  Could not download {file_name}: {e}")
                    # Try alternative filenames
                    if file_name == "pytorch_model.bin":
                        # Try safetensors or other formats
                        try:
                            hf_hub_download(repo_id=hf_id, filename="model.safetensors", local_dir=temp_dir, local_dir_use_symlinks=False)
                            print(f"   ✓ Downloaded model.safetensors (alternative)")
                        except:
                            pass
        else:
            # Download entire model (skip very large files)
            try:
                snapshot_download(
                    repo_id=hf_id,
                    local_dir=temp_dir,
                    local_dir_use_symlinks=False,
                    ignore_patterns=["*.safetensors", "*.msgpack", "*.h5", "*.ckpt", "*.ot", "*.onnx"]  # Skip large formats
                )
                print(f"   ✓ Downloaded model files")
            except Exception as e:
                print(f"   ⚠️  Full download failed: {e}, trying essential files only...")
                # Fallback: download just essential files
                essential_files = ["config.json", "tokenizer_config.json", "vocab.txt", "vocab.json"]
                for file_name in essential_files:
                    try:
                        hf_hub_download(repo_id=hf_id, filename=file_name, local_dir=temp_dir, local_dir_use_symlinks=False)
                        print(f"   ✓ Downloaded {file_name}")
                    except:
                        pass
        
        # Create .axon package (tar.gz format)
        with tarfile.open(package_path, "w:gz") as tar:
            tar.add(temp_dir, arcname=os.path.basename(temp_dir))
        
        # Compute checksum
        checksum = compute_sha256(package_path)
        package_size = os.path.getsize(package_path)
        
        print(f"   ✓ Created package: {package_name}")
        print(f"   ✓ Size: {package_size / 1024 / 1024:.2f} MB")
        print(f"   ✓ SHA256: {checksum[:16]}...")
        
        # Update manifest with real checksum
        manifest_path = os.path.join(
            registry_dir, "api/v1/models", namespace, name, version, "manifest.yaml"
        )
        
        if os.path.exists(manifest_path):
            # Read manifest
            import yaml
            with open(manifest_path, 'r') as f:
                manifest = yaml.safe_load(f)
            
            # Update checksum and size
            if "distribution" in manifest and "package" in manifest["distribution"]:
                manifest["distribution"]["package"]["sha256"] = checksum
                manifest["distribution"]["package"]["size"] = package_size
                manifest["distribution"]["package"]["url"] = f"http://localhost:8080/packages/{package_name}"
            
            # Write updated manifest
            with open(manifest_path, 'w') as f:
                yaml.dump(manifest, f, default_flow_style=False, sort_keys=False)
            
            print(f"   ✓ Updated manifest with checksum")
        
        # Cleanup temp directory
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)
        
        return True
        
    except Exception as e:
        print(f"   ❌ Error: {e}")
        # Cleanup on error
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)
        return False

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 download_hf_models.py <registry_dir>")
        sys.exit(1)
    
    registry_dir = sys.argv[1]
    packages_dir = os.path.join(registry_dir, "packages")
    
    if not os.path.exists(packages_dir):
        print(f"❌ Packages directory not found: {packages_dir}")
        sys.exit(1)
    
    print(f"📦 Creating real model packages in {packages_dir}")
    print(f"📋 Will download {len(MODELS_TO_DOWNLOAD)} models")
    print("")
    
    success_count = 0
    failed_count = 0
    
    for model_info in MODELS_TO_DOWNLOAD:
        if create_axon_package(model_info, registry_dir):
            success_count += 1
        else:
            failed_count += 1
        print("")
    
    print(f"✅ Successfully packaged {success_count} models")
    if failed_count > 0:
        print(f"⚠️  Failed to package {failed_count} models")
    
    print("")
    print("📝 Next steps:")
    print("   1. Review downloaded packages in packages/ directory")
    print("   2. Start registry: go run server.go .")
    print("   3. Test: axon install nlp/bert-base-uncased@1.0.0")

if __name__ == "__main__":
    main()

