// Package examples provides example adapter implementations for reference.
// This package demonstrates how to create new adapters using the adapter framework.
//
// Example: Replicate Adapter
//
// Replicate (https://replicate.com) is a platform for running ML models via APIs.
// It provides hosted inference for thousands of models, making it a good example
// of an API-based adapter (vs file-based adapters like Hugging Face).
//
// This adapter demonstrates:
//   - API-based model access (vs direct file downloads)
//   - REST API integration with authentication
//   - Different model access patterns
//   - Error handling for API responses
package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// ReplicateAdapter implements RepositoryAdapter for Replicate.
// Replicate is a platform for running ML models via hosted APIs.
type ReplicateAdapter struct {
	httpClient *core.HTTPClient
	baseURL    string
	apiToken   string
	validator  *core.ModelValidator
}

// NewReplicateAdapter creates a new Replicate adapter.
func NewReplicateAdapter() *ReplicateAdapter {
	client := core.NewHTTPClient("https://api.replicate.com", 5*time.Minute)
	return &ReplicateAdapter{
		httpClient: client,
		baseURL:    "https://api.replicate.com",
		apiToken:   "", // Optional API token
		validator:  core.NewModelValidator(),
	}
}

// NewReplicateAdapterWithToken creates a Replicate adapter with API token.
func NewReplicateAdapterWithToken(token string) *ReplicateAdapter {
	adapter := NewReplicateAdapter()
	adapter.apiToken = token
	adapter.httpClient.SetToken(token)
	return adapter
}

// Name returns the adapter name.
func (r *ReplicateAdapter) Name() string {
	return "replicate"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
// Replicate uses "replicate" or "rep" namespace.
func (r *ReplicateAdapter) CanHandle(namespace, name string) bool {
	return namespace == "replicate" || namespace == "rep"
}

// GetManifest retrieves the manifest for the specified model.
// Model format: replicate/{owner}/{model_name}@version
// Example: replicate/stability-ai/stable-diffusion@latest
func (r *ReplicateAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// Parse model specification
	// Format: {owner}/{model_name} (e.g., "stability-ai/stable-diffusion")
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid Replicate model format: %s (expected: owner/model_name)", name)
	}

	owner := parts[0]
	modelName := strings.Join(parts[1:], "/")

	// Construct model URL for validation
	modelURL := fmt.Sprintf("https://replicate.com/%s/%s", owner, modelName)

	// Validate model exists
	valid, err := r.validator.ValidateModelExists(ctx, modelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate model existence: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
	}

	// Fetch model metadata from Replicate API
	apiURL := fmt.Sprintf("%s/v1/models/%s/%s", r.baseURL, owner, modelName)
	resp, err := r.httpClient.Get(ctx, apiURL)
	if err != nil {
		// If API fails, create basic manifest
		return r.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		// API unavailable, create basic manifest
		return r.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
	}

	// Parse API response
	var apiResponse struct {
		URL         string `json:"url"`
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		LatestVersion struct {
			ID      string `json:"id"`
			Created string `json:"created_at"`
		} `json:"latest_version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		// Parse failed, create basic manifest
		return r.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
	}

	// Create manifest with API metadata
	manifest := &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: apiResponse.Description,
			License:     "Unknown", // Replicate doesn't expose license in API
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "Replicate", // Replicate hosts models from various frameworks
				Version: "latest",
			},
			Format: types.Format{
				Type: "replicate",
				Files: []types.ModelFile{
					{
						Path:   "model.api", // Replicate models are accessed via API
						Size:   0,
						SHA256: "",
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       r.baseURL,
				Namespace: "replicate",
			},
		},
	}

	return manifest, nil
}

// DownloadPackage downloads a model package to the specified destination path.
// Note: Replicate models are API-based, so this creates a metadata package
// rather than downloading actual model files.
func (r *ReplicateAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
	// Create package builder
	builder, err := core.NewPackageBuilder()
	if err != nil {
		return fmt.Errorf("failed to create package builder: %w", err)
	}
	defer builder.Cleanup()

	// For Replicate, we create a metadata package since models are API-based
	// In a real implementation, you might download model weights if available
	metadata := fmt.Sprintf(`{
  "adapter": "replicate",
  "model": "%s/%s",
  "version": "%s",
  "api_url": "%s",
  "description": "%s"
}`, manifest.Metadata.Namespace, manifest.Metadata.Name, manifest.Metadata.Version,
		manifest.Distribution.Package.URL, manifest.Metadata.Description)

	// Write metadata to temp file and add to package
	tempFile, err := os.CreateTemp("", "replicate-metadata-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()
	
	if _, err := tempFile.WriteString(metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	
	if err := builder.AddFile(tempFile.Name(), "metadata.json"); err != nil {
		return fmt.Errorf("failed to add metadata: %w", err)
	}

	// Build package
	if err := builder.Build(destPath); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}

	// Update manifest with checksum
	if err := core.UpdateManifestWithChecksum(manifest, destPath); err != nil {
		return fmt.Errorf("failed to update manifest checksum: %w", err)
	}

	return nil
}

// Search searches for models matching the query.
func (r *ReplicateAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// Replicate search API endpoint
	searchURL := fmt.Sprintf("%s/v1/models/search?q=%s", r.baseURL, query)

	resp, err := r.httpClient.Get(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return []types.SearchResult{}, nil // Return empty on error
	}

	var searchResponse struct {
		Results []struct {
			URL         string `json:"url"`
			Owner       string `json:"owner"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return []types.SearchResult{}, nil
	}

	var results []types.SearchResult
	for _, model := range searchResponse.Results {
		results = append(results, types.SearchResult{
			Namespace:   "replicate",
			Name:        fmt.Sprintf("%s/%s", model.Owner, model.Name),
			Version:     "latest",
			Description: model.Description,
		})
	}

	return results, nil
}

// createBasicManifest creates a basic manifest when API metadata is unavailable.
func (r *ReplicateAdapter) createBasicManifest(namespace, name, version, owner, modelName, modelURL string) *types.Manifest {
	return &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: fmt.Sprintf("Model from Replicate: %s/%s", owner, modelName),
			License:     "Unknown",
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "Replicate",
				Version: "latest",
			},
			Format: types.Format{
				Type: "replicate",
				Files: []types.ModelFile{
					{
						Path:   "model.api",
						Size:   0,
						SHA256: "",
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       r.baseURL,
				Namespace: "replicate",
			},
		},
	}
}

