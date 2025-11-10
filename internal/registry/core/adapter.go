// Package core provides the foundational interfaces and base implementations
// for Axon's pluggable adapter architecture.
//
// This package implements several GoF design patterns:
//   - Adapter Pattern: RepositoryAdapter interface adapts different model repositories
//   - Strategy Pattern: Different adapters implement different strategies for model access
//   - Factory Pattern: AdapterFactory creates adapter instances
//   - Builder Pattern: AdapterBuilder configures adapters
//
// Adapters can be registered dynamically, allowing Axon to support any model repository
// without code changes to the core system.
package core

import (
	"context"
	"fmt"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// ProgressCallback is a function type for reporting download progress.
// It receives the current number of bytes downloaded and the total size.
// If total is 0, the size is unknown.
type ProgressCallback func(current, total int64)

// RepositoryAdapter is the core interface that all model repository adapters must implement.
// This follows the Adapter Pattern, allowing different repositories to be accessed
// through a unified interface.
//
// Implementations should:
//   - Validate model existence before creating manifests
//   - Handle errors gracefully
//   - Support cancellation via context
//   - Report progress for long-running operations
type RepositoryAdapter interface {
	// Name returns a unique identifier for this adapter (e.g., "huggingface", "pytorch-hub").
	// This is used for logging, debugging, and adapter selection.
	Name() string

	// CanHandle determines if this adapter can handle a given model specification.
	// This is called by the registry to find the appropriate adapter.
	// Namespace and name are parsed from the model spec (e.g., "hf/bert-base-uncased").
	//
	// Returns true if this adapter can handle the model, false otherwise.
	// The first adapter that returns true will be used.
	CanHandle(namespace, name string) bool

	// GetManifest retrieves the manifest for a specific model.
	// This should validate that the model exists before creating a manifest.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - namespace: Model namespace (e.g., "hf", "pytorch", "tfhub")
	//   - name: Model name (e.g., "bert-base-uncased", "vision/resnet50")
	//   - version: Model version (e.g., "latest", "1.0.0")
	//
	// Returns:
	//   - manifest: Model manifest with metadata and file information
	//   - error: Error if model not found or request failed
	GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error)

	// DownloadPackage downloads a model package to the specified destination.
	// The manifest should have been obtained via GetManifest first.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - manifest: Model manifest (from GetManifest)
	//   - destPath: Destination path for the .axon package file
	//   - progress: Optional callback for progress reporting (can be nil)
	//
	// Returns:
	//   - error: Error if download failed
	DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error

	// Search searches for models matching a query string.
	// This is optional - adapters can return an empty list if search is not supported.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - query: Search query string
	//
	// Returns:
	//   - results: List of matching models
	//   - error: Error if search failed
	Search(ctx context.Context, query string) ([]types.SearchResult, error)
}

// AdapterConfig holds configuration options for adapters.
// This follows the Builder Pattern for flexible adapter configuration.
type AdapterConfig struct {
	// BaseURL is the base URL for the repository API
	BaseURL string

	// Token is an optional authentication token
	Token string

	// Timeout is the HTTP client timeout (default: 5 minutes)
	Timeout int // in seconds

	// Additional configuration options (adapter-specific)
	Options map[string]interface{}
}

// AdapterBuilder is a builder for creating and configuring adapters.
// This implements the Builder Pattern for flexible adapter construction.
type AdapterBuilder struct {
	config AdapterConfig
}

// NewAdapterBuilder creates a new adapter builder with default configuration.
func NewAdapterBuilder() *AdapterBuilder {
	return &AdapterBuilder{
		config: AdapterConfig{
			Timeout: 300, // 5 minutes default
			Options: make(map[string]interface{}),
		},
	}
}

// WithBaseURL sets the base URL for the adapter.
func (b *AdapterBuilder) WithBaseURL(url string) *AdapterBuilder {
	b.config.BaseURL = url
	return b
}

// WithToken sets the authentication token.
func (b *AdapterBuilder) WithToken(token string) *AdapterBuilder {
	b.config.Token = token
	return b
}

// WithTimeout sets the HTTP client timeout in seconds.
func (b *AdapterBuilder) WithTimeout(seconds int) *AdapterBuilder {
	b.config.Timeout = seconds
	return b
}

// WithOption sets an adapter-specific option.
func (b *AdapterBuilder) WithOption(key string, value interface{}) *AdapterBuilder {
	if b.config.Options == nil {
		b.config.Options = make(map[string]interface{})
	}
	b.config.Options[key] = value
	return b
}

// Build returns the adapter configuration.
func (b *AdapterBuilder) Build() AdapterConfig {
	return b.config
}

// AdapterFactory is a factory for creating adapter instances.
// This implements the Factory Pattern for adapter creation.
type AdapterFactory interface {
	// Create creates a new adapter instance with the given configuration.
	Create(config AdapterConfig) (RepositoryAdapter, error)

	// Name returns the name of adapters created by this factory.
	Name() string
}

// AdapterRegistry manages multiple repository adapters.
// This implements the Registry Pattern for adapter management.
type AdapterRegistry struct {
	adapters  []RepositoryAdapter
	factories map[string]AdapterFactory
}

// NewAdapterRegistry creates a new adapter registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters:  []RepositoryAdapter{},
		factories: make(map[string]AdapterFactory),
	}
}

// Register adds an adapter to the registry.
// Adapters are checked in registration order (first match wins).
func (r *AdapterRegistry) Register(adapter RepositoryAdapter) {
	r.adapters = append(r.adapters, adapter)
}

// RegisterFactory registers an adapter factory.
// This allows adapters to be created dynamically from configuration.
func (r *AdapterRegistry) RegisterFactory(factory AdapterFactory) {
	r.factories[factory.Name()] = factory
}

// CreateAdapter creates an adapter using a registered factory.
func (r *AdapterRegistry) CreateAdapter(factoryName string, config AdapterConfig) (RepositoryAdapter, error) {
	factory, ok := r.factories[factoryName]
	if !ok {
		return nil, fmt.Errorf("factory not found: %s", factoryName)
	}
	return factory.Create(config)
}

// FindAdapter finds the first adapter that can handle the given model specification.
// This uses the Strategy Pattern - each adapter implements its own strategy for
// determining if it can handle a model.
func (r *AdapterRegistry) FindAdapter(namespace, name string) (RepositoryAdapter, error) {
	for _, adapter := range r.adapters {
		if adapter.CanHandle(namespace, name) {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("no adapter found for %s/%s", namespace, name)
}

// GetAllAdapters returns all registered adapters.
func (r *AdapterRegistry) GetAllAdapters() []RepositoryAdapter {
	return r.adapters
}

// GetAdapterByName returns an adapter by its name.
func (r *AdapterRegistry) GetAdapterByName(name string) (RepositoryAdapter, error) {
	for _, adapter := range r.adapters {
		if adapter.Name() == name {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("adapter not found: %s", name)
}
