// Package builtin provides default adapters included with Axon.
package builtin

import (
	"context"

	"github.com/mlOS-foundation/axon/internal/registry"
	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// LocalRegistryAdapter implements RepositoryAdapter for local Axon registry.
type LocalRegistryAdapter struct {
	client *registry.Client
}

// NewLocalRegistryAdapter creates a new local registry adapter.
func NewLocalRegistryAdapter(baseURL string, mirrors []string) *LocalRegistryAdapter {
	return &LocalRegistryAdapter{
		client: registry.NewClient(baseURL, mirrors),
	}
}

// Name returns the adapter name.
func (l *LocalRegistryAdapter) Name() string {
	return "local"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
// Local registry can only handle models that are NOT from known adapters.
func (l *LocalRegistryAdapter) CanHandle(namespace, name string) bool {
	// Known adapter namespaces: hf, pytorch, torch, modelscope, tfhub, tf
	if namespace == "hf" || namespace == "pytorch" || namespace == "torch" ||
		namespace == "modelscope" || namespace == "tfhub" || namespace == "tf" {
		return false
	}
	// Local registry can handle models if it's configured and model is not from a known adapter
	return l.client.BaseURL() != ""
}

// Search searches for models matching the query.
func (l *LocalRegistryAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	return l.client.Search(ctx, query)
}

// GetManifest retrieves the manifest for the specified model.
func (l *LocalRegistryAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	return l.client.GetManifest(ctx, namespace, name, version)
}

// DownloadPackage downloads the model package to the specified destination path.
func (l *LocalRegistryAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
	// Convert core.ProgressCallback to registry.ProgressCallback
	var clientProgress registry.ProgressCallback
	if progress != nil {
		clientProgress = func(downloaded, total int64) {
			progress(downloaded, total)
		}
	}
	return l.client.DownloadPackage(ctx, manifest, destPath, clientProgress)
}
