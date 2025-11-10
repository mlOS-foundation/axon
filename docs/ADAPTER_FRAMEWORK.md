# Axon Adapter Framework

## Overview

Axon's adapter framework is a **pluggable, extensible architecture** that enables support for any model repository. The framework provides a clean separation between core interfaces and adapter implementations, making it easy to add support for new repositories without modifying core code.

## Framework Architecture

### Core Components

The framework consists of three main packages:

1. **`core/`** - Core interfaces and utilities
2. **`builtin/`** - Default adapters (Hugging Face, PyTorch Hub, TensorFlow Hub, Local)
3. **`examples/`** - Example adapters for reference (ModelScope)

### Package Structure

```
internal/registry/
├── core/              # Core framework
│   ├── adapter.go     # RepositoryAdapter interface, Registry, Factory, Builder
│   ├── validator.go   # ModelValidator for existence checking
│   └── helpers.go     # Common utilities (HTTPClient, PackageBuilder, etc.)
├── builtin/           # Default adapters
│   ├── huggingface.go
│   ├── pytorch.go
│   ├── tensorflow.go
│   ├── local.go
│   └── register.go    # Registration helper
└── examples/          # Example adapters
    └── modelscope.go  # Complete ModelScope implementation
```

## Framework Components

### RepositoryAdapter Interface

The `RepositoryAdapter` interface provides a unified way to access different model repositories:

```go
type RepositoryAdapter interface {
    Name() string
    CanHandle(namespace, name string) bool
    GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error)
    DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error
    Search(ctx context.Context, query string) ([]types.SearchResult, error)
}
```

**Benefits**:
- Unified interface for all repositories
- Easy to add new adapters
- Type-safe implementations

### Adapter Selection

Each adapter implements its own logic for determining which models it can handle:

- **HuggingFace**: Can handle any model (fallback strategy)
- **PyTorch Hub**: Handles `pytorch/` and `torch/` namespaces
- **TensorFlow Hub**: Handles `tfhub/` and `tf/` namespaces
- **Local Registry**: Handles configured registries (excludes known adapters)

**Example**:
```go
// Hugging Face - accepts any model (fallback)
func (h *HuggingFaceAdapter) CanHandle(namespace, name string) bool {
    return true
}

// PyTorch Hub - specific namespace
func (p *PyTorchHubAdapter) CanHandle(namespace, name string) bool {
    return namespace == "pytorch" || namespace == "torch"
}
```

### AdapterRegistry

The `AdapterRegistry` manages all adapters and finds the right one for each model:

```go
registry := core.NewAdapterRegistry()
builtin.RegisterDefaultAdapters(registry, ...)
adapter, _ := registry.FindAdapter("namespace", "model")
```

## Example: ModelScope Adapter

The ModelScope adapter (`examples/modelscope.go`) demonstrates how to implement a new adapter using the framework. Here's how it works:

### Step 1: Define the Adapter Struct

```go
type ModelScopeAdapter struct {
    httpClient *core.HTTPClient
    baseURL    string
    token      string
    validator  *core.ModelValidator
}
```

### Step 2: Implement Required Methods

#### Name()

```go
func (m *ModelScopeAdapter) Name() string {
    return "modelscope"
}
```

#### CanHandle()

```go
func (m *ModelScopeAdapter) CanHandle(namespace, name string) bool {
    return namespace == "modelscope" || namespace == "ms"
}
```

#### GetManifest()

```go
func (m *ModelScopeAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
    // 1. Parse model specification
    parts := strings.Split(name, "/")
    owner := parts[0]
    modelName := strings.Join(parts[1:], "/")
    
    // 2. Validate model exists
    modelURL := fmt.Sprintf("%s/models/%s/%s", m.baseURL, owner, modelName)
    valid, err := m.validator.ValidateModelExists(ctx, modelURL)
    if !valid {
        return nil, fmt.Errorf("model not found")
    }
    
    // 3. Fetch metadata from API
    apiURL := fmt.Sprintf("%s/api/v1/models/%s/%s", m.baseURL, owner, modelName)
    resp, err := m.httpClient.Get(ctx, apiURL)
    // ... parse response and create manifest
    
    return manifest, nil
}
```

#### DownloadPackage()

```go
func (m *ModelScopeAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
    // 1. Create package builder
    builder, err := core.NewPackageBuilder()
    if err != nil {
        return err
    }
    defer builder.Cleanup()
    
    // 2. Download model files
    for _, file := range manifest.Spec.Format.Files {
        // Download each file using core.DownloadFile()
        // Add to package using builder.AddFile()
    }
    
    // 3. Build package
    if err := builder.Build(destPath); err != nil {
        return err
    }
    
    // 4. Update manifest with checksum
    return core.UpdateManifestWithChecksum(manifest, destPath)
}
```

### Step 3: Use Helper Utilities

The framework provides several helpers:

- **`core.NewHTTPClient()`**: HTTP client with authentication
- **`core.NewModelValidator()`**: Model existence validation
- **`core.NewPackageBuilder()`**: Build .axon packages
- **`core.DownloadFile()`**: Download files with progress
- **`core.ComputeChecksum()`**: SHA256 checksums

### Complete ModelScope Implementation Flow

```
1. User runs: axon install modelscope/damo/cv_resnet50@latest
                    │
                    ▼
2. CLI calls: adapterRegistry.FindAdapter("modelscope", "damo/cv_resnet50")
                    │
                    ▼
3. Registry checks: ModelScopeAdapter.CanHandle("modelscope", "damo/cv_resnet50")
                    │ Returns: true
                    ▼
4. Registry returns: ModelScopeAdapter instance
                    │
                    ▼
5. CLI calls: adapter.GetManifest(ctx, "modelscope", "damo/cv_resnet50", "latest")
                    │
                    ├─► Validates model exists (ModelValidator)
                    ├─► Fetches metadata from ModelScope API
                    └─► Creates manifest
                    │
                    ▼
6. CLI calls: adapter.DownloadPackage(ctx, manifest, destPath, progress)
                    │
                    ├─► Creates PackageBuilder
                    ├─► Downloads files (core.DownloadFile)
                    ├─► Adds files to package (builder.AddFile)
                    ├─► Builds package (builder.Build)
                    └─► Updates checksum (core.UpdateManifestWithChecksum)
                    │
                    ▼
7. Package cached locally
```

## Helper Utilities

### HTTPClient

Provides HTTP requests with authentication:

```go
client := core.NewHTTPClient("https://api.example.com", 5*time.Minute)
client.SetToken("my-token")
resp, err := client.Get(ctx, "/models")
```

### ModelValidator

Validates model existence:

```go
validator := core.NewModelValidator()
valid, err := validator.ValidateModelExists(ctx, modelURL)
```

### PackageBuilder

Builds .axon packages:

```go
builder, _ := core.NewPackageBuilder()
builder.AddFile(srcPath, destPath)
builder.Build("package.axon")
builder.Cleanup()
```

### DownloadFile

Downloads files with progress tracking:

```go
progress := func(current, total int64) {
    fmt.Printf("Progress: %d/%d\n", current, total)
}
err := core.DownloadFile(ctx, httpClient, url, destPath, progress)
```

## Creating a New Adapter

See [ADAPTER_DEVELOPMENT.md](./ADAPTER_DEVELOPMENT.md) for a complete guide on creating new adapters.

### Quick Checklist

1. ✅ Create adapter struct implementing `RepositoryAdapter`
2. ✅ Implement `Name()`, `CanHandle()`, `GetManifest()`, `DownloadPackage()`, `Search()`
3. ✅ Use `core.ModelValidator` for model validation
4. ✅ Use `core.PackageBuilder` for package creation
5. ✅ Use `core.DownloadFile` for downloads
6. ✅ Add tests
7. ✅ Register adapter in `builtin/register.go` or create factory

## Benefits of the Framework

### For Developers

- **Clear Patterns**: Well-defined design patterns make code predictable
- **Reusable Utilities**: Common helpers reduce boilerplate
- **Type Safety**: Strong interfaces prevent errors
- **Easy Testing**: Isolated components are easier to test

### For Users

- **More Adapters**: Easier to add support for new repositories
- **Consistent Behavior**: All adapters follow the same patterns
- **Better Error Handling**: Standardized error messages
- **Extensibility**: Can add custom adapters without modifying core

## Migration from Old Framework

The old adapter framework has been completely migrated to the new structure. All existing adapters (Hugging Face, PyTorch Hub, TensorFlow Hub, Local) now use the new framework.

See [ADAPTER_MIGRATION_STATUS.md](./ADAPTER_MIGRATION_STATUS.md) for migration details.

## References

- [Adapter Development Guide](./ADAPTER_DEVELOPMENT.md) - Complete guide for creating new adapters
- [Repository Adapters](./REPOSITORY_ADAPTERS.md) - Overview of all adapters
- [Adapter Roadmap](./ADAPTER_ROADMAP.md) - Future adapter plans

