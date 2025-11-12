// Package builtin provides default adapters for Axon.
//
// # ModelScope Adapter
//
// ModelScope (https://www.modelscope.cn) is Alibaba Cloud's model repository,
// popular in Asia-Pacific markets and for multimodal AI applications.
// It provides 5,000+ models with a focus on multimodal AI and enterprise solutions.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// ModelScopeAdapter implements RepositoryAdapter for ModelScope.
// ModelScope is Alibaba Cloud's model repository with 5,000+ models.
type ModelScopeAdapter struct {
	httpClient *core.HTTPClient
	baseURL    string
	token      string
	validator  *core.ModelValidator
}

// NewModelScopeAdapter creates a new ModelScope adapter.
func NewModelScopeAdapter() *ModelScopeAdapter {
	client := core.NewHTTPClient("https://www.modelscope.cn", 5*time.Minute)
	return &ModelScopeAdapter{
		httpClient: client,
		baseURL:    "https://www.modelscope.cn",
		token:      "", // Optional token for private models
		validator:  core.NewModelValidator(),
	}
}

// NewModelScopeAdapterWithToken creates a ModelScope adapter with authentication token.
func NewModelScopeAdapterWithToken(token string) *ModelScopeAdapter {
	adapter := NewModelScopeAdapter()
	adapter.token = token
	adapter.httpClient.SetToken(token)
	return adapter
}

// Name returns the adapter name.
func (m *ModelScopeAdapter) Name() string {
	return "modelscope"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
// ModelScope uses "modelscope" or "ms" namespace.
func (m *ModelScopeAdapter) CanHandle(namespace, name string) bool {
	return namespace == "modelscope" || namespace == "ms"
}

// GetManifest retrieves the manifest for the specified model.
// Model format: modelscope/{owner}/{model_name}@version
// Example: modelscope/damo/cv_resnet50_image-classification@latest
func (m *ModelScopeAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// Parse model specification
	// Format: {owner}/{model_name} (e.g., "damo/cv_resnet50_image-classification")
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid ModelScope model format: %s (expected: owner/model_name)", name)
	}

	owner := parts[0]
	modelName := strings.Join(parts[1:], "/")

	// Construct model URL
	modelURL := fmt.Sprintf("%s/models/%s/%s", m.baseURL, owner, modelName)

	// Validate model exists
	valid, err := m.validator.ValidateModelExists(ctx, modelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate model existence: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
	}

	// Try to fetch model metadata from API
	apiURL := fmt.Sprintf("%s/api/v1/models/%s/%s", m.baseURL, owner, modelName)
	resp, err := m.httpClient.Get(ctx, apiURL)
	if err != nil {
		// If API fails, create basic manifest
		return m.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// API unavailable, create basic manifest
		return m.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
	}

	// Parse API response
	var apiResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Framework   string `json:"framework"`
		License     string `json:"license"`
		Files       []struct {
			Path string `json:"path"`
			Size int64  `json:"size"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		// Parse failed, create basic manifest
		return m.createBasicManifest(namespace, name, version, owner, modelName, modelURL), nil
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
			License:     apiResponse.License,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    apiResponse.Framework,
				Version: "latest",
			},
			Format: types.Format{
				Type:  "modelscope",
				Files: []types.ModelFile{},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       m.baseURL,
				Namespace: "modelscope",
			},
		},
	}

	// Add files from API response
	for _, file := range apiResponse.Files {
		manifest.Spec.Format.Files = append(manifest.Spec.Format.Files, types.ModelFile{
			Path: file.Path,
			Size: file.Size,
		})
	}

	// If no files from API, add default file
	if len(manifest.Spec.Format.Files) == 0 {
		manifest.Spec.Format.Files = []types.ModelFile{
			{
				Path:   "model.tar.gz",
				Size:   0,  // Will be determined during download
				SHA256: "", // Will be computed during download
			},
		}
	}

	return manifest, nil
}

// DownloadPackage downloads a model package to the specified destination path.
func (m *ModelScopeAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
	// Create package builder
	builder, err := core.NewPackageBuilder()
	if err != nil {
		return fmt.Errorf("failed to create package builder: %w", err)
	}
	defer builder.Cleanup()

	// Download model files
	modelURL := manifest.Distribution.Package.URL
	downloadURL := fmt.Sprintf("%s/models/%s/repo?Revision=master&FilePath=", m.baseURL, strings.TrimPrefix(modelURL, m.baseURL+"/models/"))

	// For simplicity, download the main model file
	// In a full implementation, you would download all files from manifest.Spec.Format.Files
	mainFileURL := downloadURL + "model.tar.gz"

	// Create temp file for download
	tempFile := "/tmp/modelscope-download.tar.gz"
	httpClient := &http.Client{Timeout: 10 * time.Minute}

	if err := core.DownloadFile(ctx, httpClient, mainFileURL, tempFile, progress); err != nil {
		// If direct download fails, try alternative approach
		// This is a simplified example - real implementation would handle multiple files
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Add downloaded file to package
	if err := builder.AddFile(tempFile, "model.tar.gz"); err != nil {
		return fmt.Errorf("failed to add file to package: %w", err)
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
// ModelScope provides a search API, but this is a simplified example.
func (m *ModelScopeAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// ModelScope search API endpoint
	searchURL := fmt.Sprintf("%s/api/v1/models?Keyword=%s", m.baseURL, query)

	resp, err := m.httpClient.Get(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []types.SearchResult{}, nil // Return empty on error
	}

	var searchResponse struct {
		Data struct {
			Models []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Owner       string `json:"owner"`
			} `json:"models"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return []types.SearchResult{}, nil
	}

	var results []types.SearchResult
	for _, model := range searchResponse.Data.Models {
		results = append(results, types.SearchResult{
			Namespace:   "modelscope",
			Name:        fmt.Sprintf("%s/%s", model.Owner, model.Name),
			Version:     "latest",
			Description: model.Description,
		})
	}

	return results, nil
}

// createBasicManifest creates a basic manifest when API metadata is unavailable.
func (m *ModelScopeAdapter) createBasicManifest(namespace, name, version, owner, modelName, modelURL string) *types.Manifest {
	return &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: fmt.Sprintf("Model from ModelScope: %s/%s", owner, modelName),
			License:     "Unknown",
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "PyTorch", // ModelScope primarily uses PyTorch
				Version: "latest",
			},
			Format: types.Format{
				Type: "modelscope",
				Files: []types.ModelFile{
					{
						Path:   "model.tar.gz",
						Size:   0,  // Will be determined during download
						SHA256: "", // Will be computed during download
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       m.baseURL,
				Namespace: "modelscope",
			},
		},
	}
}

// ModelScopeFactory implements AdapterFactory for creating ModelScope adapters.
// This demonstrates the Factory Pattern for dynamic adapter creation.
type ModelScopeFactory struct{}

// NewModelScopeFactory creates a new ModelScope factory.
func NewModelScopeFactory() *ModelScopeFactory {
	return &ModelScopeFactory{}
}

// Name returns the factory name.
func (f *ModelScopeFactory) Name() string {
	return "modelscope"
}

// Create creates a new ModelScope adapter with the given configuration.
func (f *ModelScopeFactory) Create(config core.AdapterConfig) (core.RepositoryAdapter, error) {
	adapter := NewModelScopeAdapter()

	if config.BaseURL != "" {
		adapter.baseURL = config.BaseURL
		adapter.httpClient = core.NewHTTPClient(config.BaseURL, time.Duration(config.Timeout)*time.Second)
	}

	if config.Token != "" {
		adapter.token = config.Token
		adapter.httpClient.SetToken(config.Token)
	}

	return adapter, nil
}
