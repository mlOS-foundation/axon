# Quick Start - Local Registry Testing

## Direct Install (Recommended)

For most use cases, install models directly from Hugging Face:

```bash
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
```

No setup required - manifests and packages are created on-the-fly!

## Local Registry Testing (Optional)

If you want to test with a local registry:

```bash
# 1. Start local registry
cd test/registry
go run server.go .

# 2. Configure Axon
axon registry set default http://localhost:8080

# 3. Install from local registry
axon install nlp/bert-base-uncased@1.0.0
```

That's it! The local registry includes 10 models for testing.

## Adding Models to Local Registry

Edit `create-manifests.go` to add more models, then run:

```bash
go run create-manifests.go .
```

## Hosted Registry

For production hosted registries, a separate sync pipeline will handle model synchronization from Hugging Face. This is not part of the core Axon repository.
