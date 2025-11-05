# Testing the Adapter System

## Quick Test: Install from Hugging Face

Test the real-time download and packaging:

```bash
# Build Axon
cd axon
go build ./cmd/axon

# Initialize (creates ~/.axon directory)
./axon init

# Test install from Hugging Face (small model for quick test)
./axon install hf/bert-base-uncased@latest
```

Expected output:
```
Propagating hf/bert-base-uncased@latest...
Using huggingface adapter for hf/bert-base-uncased
Downloading package...
Downloading... 100.0% (440234567/440234567 bytes)
✓ Successfully propagated hf/bert-base-uncased@latest
```

## Test Local Registry (Optional)

If you want to test the local registry:

```bash
cd test/registry

# Start local registry server
go run server.go . &

# Configure Axon to use local registry
axon registry set default http://localhost:8080

# Install from local registry
axon install nlp/bert-base-uncased@1.0.0
```

## Test Adapter Selection

Test that adapters are selected correctly:

```bash
# Should use Hugging Face adapter (no local registry configured)
axon install hf/gpt2@latest

# Configure local registry
axon registry set default http://localhost:8080

# Should try local registry first, then fall back to HF
axon install hf/roberta-base@latest
```

## Test Token Support (Optional)

For gated models, test token support:

```bash
# Set token (if you have one)
axon config set registry.huggingface_token hf_xxxxxxxxxxxxxxxxxxxx

# Install gated model (e.g., Llama)
axon install hf/meta-llama/Llama-2-7b-hf@latest
```

## Verify Package Creation

Check that packages are created correctly:

```bash
# Install a model
axon install hf/bert-base-uncased@latest

# Check cache directory
ls -lh ~/.axon/cache/hf/bert-base-uncased/

# Verify package exists
find ~/.axon/cache -name "*.axon" -ls
```

## Troubleshooting

### Error: "no adapter found"
- Check that `enable_huggingface: true` in config
- Verify network connection

### Error: "401 Unauthorized"
- Model is gated - need HF token
- Set token: `axon config set registry.huggingface_token <token>`

### Error: "Download failed"
- Check network connection
- Verify model name is correct
- Try with explicit version: `@main` instead of `@latest`

