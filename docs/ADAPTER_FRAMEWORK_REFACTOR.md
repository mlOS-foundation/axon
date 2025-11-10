# Adapter Framework Refactoring Summary

## Overview

The adapter framework has been refactored to use GoF design patterns, making it more extensible, maintainable, and easier to use. This refactoring provides a solid foundation for adding new adapters while keeping the existing functionality intact.

## What's New

### 1. Core Framework (`internal/registry/core/`)

**Design Patterns Implemented:**

- **Adapter Pattern**: `RepositoryAdapter` interface unifies different repositories
- **Strategy Pattern**: Each adapter implements its own strategy for model access
- **Factory Pattern**: `AdapterFactory` enables dynamic adapter creation
- **Builder Pattern**: `AdapterBuilder` provides fluent configuration
- **Registry Pattern**: `AdapterRegistry` manages all adapters

**Key Components:**

- `adapter.go`: Core interfaces (RepositoryAdapter, AdapterRegistry, AdapterFactory, AdapterBuilder)
- `validator.go`: ModelValidator for existence checking
- `helpers.go`: Common utilities (HTTPClient, PackageBuilder, DownloadFile, etc.)

### 2. Example Adapter (`internal/registry/examples/`)

**ModelScope Adapter** - Complete example showing:
- REST API integration
- Model validation
- Package creation
- Factory pattern implementation
- Comprehensive tests

### 3. Documentation

- **ADAPTER_DEVELOPMENT.md**: Complete guide for creating new adapters
- **ADAPTER_MIGRATION.md**: Migration guide from old structure
- **internal/registry/README.md**: Package overview and quick start

## Benefits

### For Developers

1. **Easier Extension**: Clear patterns for adding new adapters
2. **Reusable Utilities**: Common helpers reduce boilerplate
3. **Better Testing**: Isolated components are easier to test
4. **Type Safety**: Strong interfaces prevent errors

### For Users

1. **More Adapters**: Easier to add support for new repositories
2. **Consistent Behavior**: All adapters follow the same patterns
3. **Better Error Handling**: Standardized error messages
4. **Extensibility**: Can add custom adapters without modifying core

## Architecture

```
┌─────────────────────────────────────────┐
│         Axon CLI                        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│    AdapterRegistry (Registry Pattern)   │
└──────────────┬──────────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
    ▼          ▼          ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│ Builtin │ │Examples │ │ Custom  │
│Adapters │ │Adapters │ │Adapters │
└─────────┘ └─────────┘ └─────────┘
```

## Design Patterns Explained

### Adapter Pattern

**Purpose**: Allow incompatible interfaces to work together

**Implementation**: `RepositoryAdapter` interface provides a unified way to access different model repositories (Hugging Face, PyTorch Hub, TensorFlow Hub, etc.)

**Example**:
```go
type RepositoryAdapter interface {
    GetManifest(ctx, namespace, name, version) (*Manifest, error)
    DownloadPackage(ctx, manifest, destPath, progress) error
}
```

### Strategy Pattern

**Purpose**: Define a family of algorithms and make them interchangeable

**Implementation**: Each adapter implements its own strategy for:
- Determining if it can handle a model (`CanHandle`)
- Fetching metadata (`GetManifest`)
- Downloading models (`DownloadPackage`)

**Example**:
```go
// Hugging Face strategy
func (h *HuggingFaceAdapter) CanHandle(namespace, name string) bool {
    return true // HF can handle any model
}

// PyTorch Hub strategy
func (p *PyTorchHubAdapter) CanHandle(namespace, name string) bool {
    return namespace == "pytorch" || namespace == "torch"
}
```

### Factory Pattern

**Purpose**: Create objects without specifying the exact class

**Implementation**: `AdapterFactory` creates adapters from configuration

**Example**:
```go
factory := NewModelScopeFactory()
adapter, err := factory.Create(config)
```

### Builder Pattern

**Purpose**: Construct complex objects step by step

**Implementation**: `AdapterBuilder` provides fluent configuration

**Example**:
```go
config := core.NewAdapterBuilder().
    WithBaseURL("https://api.example.com").
    WithToken("my-token").
    WithTimeout(300).
    Build()
```

### Registry Pattern

**Purpose**: Centralize object creation and management

**Implementation**: `AdapterRegistry` manages all adapters

**Example**:
```go
registry := core.NewAdapterRegistry()
registry.Register(adapter1)
registry.Register(adapter2)
adapter, _ := registry.FindAdapter("namespace", "model")
```

## Usage Examples

### Basic Usage

```go
import (
    "github.com/mlOS-foundation/axon/internal/registry/core"
    "github.com/mlOS-foundation/axon/internal/registry/builtin"
)

// Create registry
registry := core.NewAdapterRegistry()

// Register adapters
registry.Register(builtin.NewHuggingFaceAdapter())
registry.Register(builtin.NewPyTorchHubAdapter())

// Find and use adapter
adapter, _ := registry.FindAdapter("hf", "bert-base-uncased")
manifest, _ := adapter.GetManifest(ctx, "hf", "bert-base-uncased", "latest")
```

### Using Helpers

```go
// HTTP client with auth
client := core.NewHTTPClient("https://api.example.com", 5*time.Minute)
client.SetToken("my-token")
resp, _ := client.Get(ctx, "/models")

// Package builder
builder, _ := core.NewPackageBuilder()
builder.AddFile("model.pth", "model.pth")
builder.Build("package.axon")

// Model validation
validator := core.NewModelValidator()
valid, _ := validator.ValidateModelExists(ctx, modelURL)
```

### Creating Custom Adapter

```go
type MyAdapter struct {
    httpClient *core.HTTPClient
    validator  *core.ModelValidator
}

func (a *MyAdapter) Name() string { return "my-adapter" }
func (a *MyAdapter) CanHandle(ns, name string) bool { return ns == "my" }
// ... implement other methods
```

## Migration Status

### Completed ✅

- [x] Core framework with GoF patterns
- [x] Helper utilities (HTTP, Package, Validation)
- [x] Example ModelScope adapter
- [x] Comprehensive documentation
- [x] Tests for example adapter

### Pending 🔄

- [ ] Migrate existing adapters to `builtin/` package
- [ ] Update CLI to use new framework
- [ ] Update all tests
- [ ] Full integration testing

## Next Steps

1. **Migrate Builtin Adapters**: Move Hugging Face, PyTorch Hub, TensorFlow Hub, and Local adapters to `builtin/` package
2. **Update CLI**: Modify `cmd/axon/commands.go` to use new framework
3. **Update Tests**: Migrate existing tests to new structure
4. **Integration Testing**: Ensure all adapters work with new framework

## Files Created

### Core Framework
- `internal/registry/core/adapter.go` - Core interfaces and registry
- `internal/registry/core/validator.go` - Model validation
- `internal/registry/core/helpers.go` - Common utilities

### Examples
- `internal/registry/examples/modelscope.go` - ModelScope adapter example
- `internal/registry/examples/modelscope_test.go` - Tests

### Documentation
- `docs/ADAPTER_DEVELOPMENT.md` - Development guide
- `docs/ADAPTER_MIGRATION.md` - Migration guide
- `internal/registry/README.md` - Package overview
- `docs/ADAPTER_FRAMEWORK_REFACTOR.md` - This document

## References

- [GoF Design Patterns](https://en.wikipedia.org/wiki/Design_Patterns)
- [Adapter Pattern](https://refactoring.guru/design-patterns/adapter)
- [Strategy Pattern](https://refactoring.guru/design-patterns/strategy)
- [Factory Pattern](https://refactoring.guru/design-patterns/factory-method)
- [Builder Pattern](https://refactoring.guru/design-patterns/builder)
- [Registry Pattern](https://martinfowler.com/eaaCatalog/registry.html)

