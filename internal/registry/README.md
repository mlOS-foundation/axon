# Axon Registry Adapter Framework

This package provides a pluggable adapter framework for integrating different model repositories into Axon.

## Architecture

The framework uses several GoF design patterns:

- **Adapter Pattern**: `RepositoryAdapter` interface unifies different repositories
- **Strategy Pattern**: Each adapter implements its own strategy for model access
- **Factory Pattern**: `AdapterFactory` creates adapters dynamically
- **Builder Pattern**: `AdapterBuilder` configures adapters fluently
- **Registry Pattern**: `AdapterRegistry` manages all adapters

## Package Structure

```
internal/registry/
├── core/              # Core interfaces and utilities
│   ├── adapter.go     # RepositoryAdapter, Registry, Factory, Builder
│   ├── validator.go   # ModelValidator for existence checking
│   └── helpers.go     # Common helper functions (HTTP, Package, etc.)
├── builtin/           # Default adapters (Hugging Face, PyTorch, TensorFlow, Local)
│   ├── huggingface.go
│   ├── pytorch.go
│   ├── tensorflow.go
│   └── local.go
└── examples/          # Example adapters for reference
    └── modelscope.go  # ModelScope adapter example
```

## Quick Start

### Using Builtin Adapters

```go
import (
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/internal/registry/builtin"
)

// Create registry
registry := core.NewAdapterRegistry()

// Register builtin adapters
registry.Register(builtin.NewHuggingFaceAdapter())
registry.Register(builtin.NewPyTorchHubAdapter())
registry.Register(builtin.NewTensorFlowHubAdapter())

// Find adapter for a model
adapter, err := registry.FindAdapter("hf", "bert-base-uncased")
if err != nil {
    // Handle error
}

// Get model manifest
manifest, err := adapter.GetManifest(ctx, "hf", "bert-base-uncased", "latest")
```

### Creating a New Adapter

See [Adapter Development Guide](../../docs/ADAPTER_DEVELOPMENT.md) for detailed instructions.

Quick example:

```go
type MyAdapter struct {
    httpClient *core.HTTPClient
    validator  *core.ModelValidator
}

func (a *MyAdapter) Name() string {
    return "my-adapter"
}

func (a *MyAdapter) CanHandle(namespace, name string) bool {
    return namespace == "my"
}

func (a *MyAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
    // Validate model exists
    valid, err := a.validator.ValidateModelExists(ctx, modelURL)
    // ... create manifest
}

func (a *MyAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
    // Use core.NewPackageBuilder() to create package
    // ... download and package files
}

func (a *MyAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
    // Implement search if supported
}
```

## Core Components

### RepositoryAdapter Interface

All adapters must implement this interface:

```go
type RepositoryAdapter interface {
    Name() string
    CanHandle(namespace, name string) bool
    GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error)
    DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error
    Search(ctx context.Context, query string) ([]types.SearchResult, error)
}
```

### AdapterRegistry

Manages adapters and finds the appropriate one:

```go
registry := core.NewAdapterRegistry()
registry.Register(adapter)
adapter, err := registry.FindAdapter("namespace", "model-name")
```

### Helper Utilities

- `core.NewHTTPClient()`: HTTP client with auth support
- `core.NewModelValidator()`: Model existence validation
- `core.NewPackageBuilder()`: Build .axon packages
- `core.DownloadFile()`: Download with progress
- `core.ComputeChecksum()`: SHA256 checksums

## Documentation

- [Adapter Development Guide](../../docs/ADAPTER_DEVELOPMENT.md) - How to create new adapters
- [Adapter Migration Guide](../../docs/ADAPTER_MIGRATION.md) - Migrating from old structure
- [Repository Adapters](../../docs/REPOSITORY_ADAPTERS.md) - Overview of adapters

## Examples

See `examples/modelscope.go` for a complete adapter implementation example.

## Testing

Run tests:

```bash
go test ./internal/registry/...
```

Test coverage:

```bash
go test -cover ./internal/registry/...
```

