# Adapter Development Guide

This guide explains how to create new repository adapters for Axon using the refactored adapter framework. The framework uses several GoF design patterns to make adapter development straightforward and extensible.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Design Patterns Used](#design-patterns-used)
3. [Creating a New Adapter](#creating-a-new-adapter)
4. [Example: Replicate Adapter](#example-replicate-adapter)
5. [Best Practices](#best-practices)
6. [Testing Your Adapter](#testing-your-adapter)
7. [Registering Your Adapter](#registering-your-adapter)

## Architecture Overview

Axon's adapter framework follows a clean, extensible architecture:

```
┌─────────────────────────────────────────────────────────┐
│                    Axon CLI                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              AdapterRegistry (Registry Pattern)          │
│  - Manages all adapters                                   │
│  - Finds appropriate adapter for model spec              │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Builtin   │ │  Examples   │ │   Custom    │
│  Adapters   │ │  Adapters   │ │  Adapters   │
└─────────────┘ └─────────────┘ └─────────────┘
```

### Core Components

- **`core/adapter.go`**: Defines the `RepositoryAdapter` interface and registry
- **`core/validator.go`**: Provides model validation utilities
- **`core/helpers.go`**: Common helper functions for adapters
- **`builtin/`**: Default adapters (Hugging Face, PyTorch Hub, TensorFlow Hub, Local)
- **`examples/`**: Example adapters for reference

## Design Patterns Used

### 1. Adapter Pattern

The `RepositoryAdapter` interface allows different model repositories to be accessed through a unified interface:

```go
type RepositoryAdapter interface {
    Name() string
    CanHandle(namespace, name string) bool
    GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error)
    DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error
    Search(ctx context.Context, query string) ([]types.SearchResult, error)
}
```

### 2. Strategy Pattern

Each adapter implements its own strategy for:
- Determining if it can handle a model (`CanHandle`)
- Fetching model metadata (`GetManifest`)
- Downloading models (`DownloadPackage`)

### 3. Factory Pattern

Adapters can be created via factories for dynamic configuration:

```go
type AdapterFactory interface {
    Create(config AdapterConfig) (RepositoryAdapter, error)
    Name() string
}
```

### 4. Builder Pattern

The `AdapterBuilder` provides a fluent interface for configuring adapters:

```go
builder := core.NewAdapterBuilder().
    WithBaseURL("https://api.example.com").
    WithToken("my-token").
    WithTimeout(300)
config := builder.Build()
```

### 5. Registry Pattern

The `AdapterRegistry` manages all adapters and finds the appropriate one:

```go
registry := core.NewAdapterRegistry()
registry.Register(myAdapter)
adapter, err := registry.FindAdapter("namespace", "model-name")
```

## Creating a New Adapter

### Step 1: Create Adapter Struct

Create a new file in your package (e.g., `examples/replicate.go`):

```go
package examples

import (
    "context"
    "net/http"
    "time"
    
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/pkg/types"
)

// ReplicateAdapter implements RepositoryAdapter for Replicate
type ReplicateAdapter struct {
    httpClient    *core.HTTPClient
    baseURL       string
    token         string
    validator     *core.ModelValidator
}

// NewModelScopeAdapter creates a new ModelScope adapter
func NewModelScopeAdapter() *ModelScopeAdapter {
    return &ModelScopeAdapter{
        httpClient: core.NewHTTPClient("https://www.modelscope.cn", 5*time.Minute),
        baseURL:    "https://www.modelscope.cn",
        validator:  core.NewModelValidator(),
    }
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
    // ModelScope uses "modelscope" or "ms" namespace
    return namespace == "modelscope" || namespace == "ms"
}
```

#### GetManifest()

```go
func (m *ModelScopeAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
    // 1. Validate model exists
    modelURL := fmt.Sprintf("%s/models/%s", m.baseURL, name)
    valid, err := m.validator.ValidateModelExists(ctx, modelURL)
    if err != nil {
        return nil, fmt.Errorf("failed to validate model: %w", err)
    }
    if !valid {
        return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
    }
    
    // 2. Fetch model metadata from API
    // ... API call logic ...
    
    // 3. Create manifest
    manifest := &types.Manifest{
        APIVersion: "v1",
        Kind:       "Model",
        Metadata: types.Metadata{
            Name:        name,
            Namespace:   namespace,
            Version:     version,
            Description: "Model from ModelScope",
            // ... more metadata ...
        },
        // ... rest of manifest ...
    }
    
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
        // Download each file
        // ... download logic ...
        builder.AddFile(tempPath, file.Path)
    }
    
    // 3. Build package
    if err := builder.Build(destPath); err != nil {
        return err
    }
    
    // 4. Update manifest with checksum
    return core.UpdateManifestWithChecksum(manifest, destPath)
}
```

#### Search()

```go
func (m *ModelScopeAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
    // Implement search if API supports it
    // Otherwise return empty list
    return []types.SearchResult{}, nil
}
```

### Step 3: Use Helper Utilities

The framework provides several helpers:

- **`core.NewHTTPClient()`**: HTTP client with authentication
- **`core.NewModelValidator()`**: Model existence validation
- **`core.NewPackageBuilder()`**: Build .axon packages
- **`core.DownloadFile()`**: Download files with progress
- **`core.ComputeChecksum()`**: Compute SHA256 checksums

## Example: Replicate Adapter

See `internal/registry/examples/replicate.go` for a complete implementation example.

### Replicate Overview

- **Base URL**: https://api.replicate.com
- **API**: REST API for hosted inference models
- **Namespace**: `replicate` or `rep`
- **Format**: `replicate/{owner}/{model_name}@version`

### Key Features

1. **Model Validation**: Validates model exists before creating manifest
2. **Metadata Fetching**: Uses Replicate API to get model information
3. **Package Creation**: Creates .axon packages with metadata (API-based adapter)
4. **Progress Tracking**: Reports download progress

## Best Practices

### 1. Always Validate Model Existence

```go
valid, err := adapter.validator.ValidateModelExists(ctx, modelURL)
if !valid {
    return nil, fmt.Errorf("model not found")
}
```

### 2. Handle Errors Gracefully

```go
if err != nil {
    return nil, fmt.Errorf("failed to fetch model: %w", err)
}
```

### 3. Support Context Cancellation

```go
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### 4. Report Progress

```go
progress(currentBytes, totalBytes)
```

### 5. Clean Up Resources

```go
defer resp.Body.Close()
defer builder.Cleanup()
```

### 6. Use Helper Functions

Leverage the provided helpers instead of reimplementing:
- `core.DownloadFile()` for downloads
- `core.NewPackageBuilder()` for packages
- `core.ComputeChecksum()` for checksums

## Testing Your Adapter

### Unit Tests

Create test files (e.g., `examples/modelscope_test.go`):

```go
func TestModelScopeAdapter_Name(t *testing.T) {
    adapter := NewModelScopeAdapter()
    if adapter.Name() != "modelscope" {
        t.Errorf("expected name 'modelscope', got '%s'", adapter.Name())
    }
}

func TestModelScopeAdapter_CanHandle(t *testing.T) {
    adapter := NewModelScopeAdapter()
    
    tests := []struct {
        namespace string
        name      string
        want      bool
    }{
        {"modelscope", "cv/resnet50", true},
        {"ms", "cv/resnet50", true},
        {"hf", "bert-base", false},
    }
    
    for _, tt := range tests {
        if got := adapter.CanHandle(tt.namespace, tt.name); got != tt.want {
            t.Errorf("CanHandle(%q, %q) = %v, want %v", tt.namespace, tt.name, got, tt.want)
        }
    }
}
```

### Integration Tests

Test with real API (use test models):

```go
func TestModelScopeAdapter_GetManifest_RealAPI(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    adapter := NewModelScopeAdapter()
    ctx := context.Background()
    
    manifest, err := adapter.GetManifest(ctx, "modelscope", "cv/resnet50", "latest")
    if err != nil {
        t.Fatalf("GetManifest failed: %v", err)
    }
    
    if manifest.Metadata.Name == "" {
        t.Error("manifest name is empty")
    }
}
```

## Registering Your Adapter

### In Builtin Package

For adapters included with Axon, register in `builtin/register.go`:

```go
package builtin

import (
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/internal/registry/builtin"
)

func RegisterDefaultAdapters(registry *core.AdapterRegistry) {
    registry.Register(builtin.NewHuggingFaceAdapter())
    registry.Register(builtin.NewPyTorchHubAdapter())
    registry.Register(builtin.NewTensorFlowHubAdapter())
    registry.Register(builtin.NewLocalRegistryAdapter("", nil))
}
```

### In CLI

Update `cmd/axon/commands.go`:

```go
// Register adapters
adapterRegistry := core.NewAdapterRegistry()
builtin.RegisterDefaultAdapters(adapterRegistry)

// Add custom adapters
adapterRegistry.Register(examples.NewModelScopeAdapter())
```

### Via Configuration

For dynamically loaded adapters:

```go
factory := examples.NewModelScopeFactory()
registry.RegisterFactory(factory)

config := core.NewAdapterBuilder().
    WithBaseURL("https://api.modelscope.cn").
    WithToken("my-token").
    Build()

adapter, err := registry.CreateAdapter("modelscope", config)
```

## Common Patterns

### Pattern 1: Simple REST API Adapter

```go
// 1. Fetch metadata from REST API
resp, err := httpClient.Get(ctx, apiURL)
// 2. Parse JSON response
// 3. Create manifest
// 4. Download files
```

### Pattern 2: GitHub-Based Adapter

```go
// 1. Parse GitHub repo from model spec
// 2. Fetch files via GitHub API
// 3. Download from GitHub releases or raw content
```

### Pattern 3: File-Based Adapter

```go
// 1. List files in repository
// 2. Download each file
// 3. Package into .axon format
```

## Troubleshooting

### Issue: Adapter not found

**Solution**: Ensure adapter is registered before use:
```go
registry.Register(myAdapter)
```

### Issue: Model validation fails

**Solution**: Check URL format and use `ModelValidator`:
```go
validator := core.NewModelValidator()
valid, err := validator.ValidateModelExists(ctx, modelURL)
```

### Issue: Download fails

**Solution**: Use `core.DownloadFile()` helper:
```go
err := core.DownloadFile(ctx, httpClient, url, destPath, progress)
```

## Next Steps

1. Review the example ModelScope adapter
2. Study existing builtin adapters
3. Create your adapter following the patterns
4. Write comprehensive tests
5. Submit a PR to add your adapter

## Resources

- [Adapter Interface Documentation](../internal/registry/core/adapter.go)
- [Helper Functions](../internal/registry/core/helpers.go)
- [Example: Replicate Adapter](../internal/registry/examples/replicate.go)
- [Builtin Adapters](../internal/registry/builtin/)

