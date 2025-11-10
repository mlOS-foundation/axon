# Adapter Framework Migration Guide

This document explains the migration from the old adapter structure to the new refactored framework.

## Overview

The adapter framework has been refactored to:
- Separate core interfaces from implementations
- Provide reusable helper utilities
- Support dynamic adapter registration
- Enable easier extension with new adapters

## New Structure

```
internal/registry/
├── core/              # Core interfaces and utilities
│   ├── adapter.go     # RepositoryAdapter interface, Registry, Factory, Builder
│   ├── validator.go   # ModelValidator for existence checking
│   └── helpers.go     # Common helper functions
├── builtin/           # Default adapters (migrated from adapter.go)
│   ├── huggingface.go
│   ├── pytorch.go
│   ├── tensorflow.go
│   └── local.go
└── examples/          # Example adapters
    └── modelscope.go  # ModelScope adapter example
```

## Migration Steps

### Step 1: Update Imports

**Old:**
```go
import "github.com/mlOS-foundation/axon/internal/registry"
```

**New:**
```go
import (
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/internal/registry/builtin"
)
```

### Step 2: Update Adapter Creation

**Old:**
```go
adapter := registry.NewHuggingFaceAdapter()
registry := registry.NewAdapterRegistry()
registry.Register(adapter)
```

**New:**
```go
adapter := builtin.NewHuggingFaceAdapter()
registry := core.NewAdapterRegistry()
registry.Register(adapter)
```

### Step 3: Update Interface Usage

The `RepositoryAdapter` interface is now in `core` package:

**Old:**
```go
var adapter registry.RepositoryAdapter
```

**New:**
```go
var adapter core.RepositoryAdapter
```

### Step 4: Use Helper Functions

**Old:**
```go
// Manual HTTP client creation
client := &http.Client{Timeout: 5 * time.Minute}
```

**New:**
```go
// Use helper
client := core.NewHTTPClient("https://api.example.com", 5*time.Minute)
```

**Old:**
```go
// Manual package creation
// ... lots of tar/gzip code ...
```

**New:**
```go
// Use helper
builder, _ := core.NewPackageBuilder()
builder.AddFile(src, dest)
builder.Build(destPath)
```

## Breaking Changes

### 1. Package Structure

- `internal/registry/adapter.go` → Split into `core/` and `builtin/`
- `ProgressCallback` moved to `core` package
- `ModelValidator` moved to `core` package

### 2. Interface Location

- `RepositoryAdapter` → `core.RepositoryAdapter`
- `AdapterRegistry` → `core.AdapterRegistry`

### 3. Helper Functions

- All helpers now in `core` package
- Use `core.NewHTTPClient()`, `core.NewPackageBuilder()`, etc.

## Compatibility

The old adapters continue to work but are deprecated. Migration is recommended for:
- Better code organization
- Access to new helper utilities
- Support for dynamic adapter registration
- Easier testing and extension

## Migration Checklist

- [ ] Update imports to use `core` and `builtin` packages
- [ ] Replace direct adapter creation with `builtin` package
- [ ] Update registry creation to use `core.NewAdapterRegistry()`
- [ ] Replace manual HTTP client creation with `core.NewHTTPClient()`
- [ ] Replace manual package building with `core.NewPackageBuilder()`
- [ ] Update tests to use new package structure
- [ ] Update documentation references

## Example Migration

### Before

```go
package main

import (
    "github.com/mlOS-foundation/axon/internal/registry"
)

func main() {
    adapter := registry.NewHuggingFaceAdapter()
    registry := registry.NewAdapterRegistry()
    registry.Register(adapter)
    
    manifest, _ := adapter.GetManifest(ctx, "hf", "bert-base", "latest")
}
```

### After

```go
package main

import (
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/internal/registry/builtin"
)

func main() {
    adapter := builtin.NewHuggingFaceAdapter()
    registry := core.NewAdapterRegistry()
    registry.Register(adapter)
    
    manifest, _ := adapter.GetManifest(ctx, "hf", "bert-base", "latest")
}
```

## Next Steps

1. Review the new structure
2. Migrate existing code
3. Test thoroughly
4. Update documentation

For creating new adapters, see [Adapter Development Guide](./ADAPTER_DEVELOPMENT.md).

