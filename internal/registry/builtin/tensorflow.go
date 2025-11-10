// Package builtin provides default adapters included with Axon.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

type TensorFlowHubAdapter struct {
	httpClient     *http.Client
	baseURL        string // TensorFlow Hub base URL
	modelValidator *core.ModelValidator
}

// NewTensorFlowHubAdapter creates a new TensorFlow Hub adapter
func NewTensorFlowHubAdapter() *TensorFlowHubAdapter {
	return &TensorFlowHubAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://tfhub.dev",
		modelValidator: core.NewModelValidator(),
	}
}

// Name returns the name of the adapter.
func (t *TensorFlowHubAdapter) Name() string {
	return "tensorflow-hub"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
func (t *TensorFlowHubAdapter) CanHandle(namespace, name string) bool {
	// TensorFlow Hub can handle models with "tfhub" or "tf" namespace
	return namespace == "tfhub" || namespace == "tf"
}

// Search searches for models matching the query.
// TensorFlow Hub provides a REST API for searching models
func (t *TensorFlowHubAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// TensorFlow Hub search API: https://tfhub.dev/api/v1/models?q={query}
	searchURL := fmt.Sprintf("%s/api/v1/models?q=%s", t.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search models: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		// If search API is not available, return empty results
		return []types.SearchResult{}, nil
	}

	var searchResponse struct {
		Results []struct {
			Publisher   string `json:"publisher"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		// If decode fails, return empty results
		return []types.SearchResult{}, nil
	}

	var results []types.SearchResult
	for _, model := range searchResponse.Results {
		results = append(results, types.SearchResult{
			Namespace:   "tfhub",
			Name:        fmt.Sprintf("%s/%s", model.Publisher, model.Name),
			Version:     model.Version,
			Description: model.Description,
		})
	}

	return results, nil
}

// GetManifest retrieves the manifest for the specified model.
// Model format: tfhub/{publisher}/{model_path}@version
// Example: tfhub/google/imagenet/resnet_v2_50/classification/5
func (t *TensorFlowHubAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// Parse model specification
	// Format: {publisher}/{model_path} (e.g., "google/imagenet/resnet_v2_50/classification/5")
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid TensorFlow Hub model format: %s (expected: publisher/model_path)", name)
	}

	publisher := parts[0]
	modelPath := strings.Join(parts[1:], "/")

	// Construct model URL
	modelURL := fmt.Sprintf("%s/%s/%s", t.baseURL, publisher, modelPath)
	if version != "latest" && version != "" {
		modelURL = fmt.Sprintf("%s/%s", modelURL, version)
	}

	// Try to fetch model metadata (optional - if API is available)
	metadataURL := fmt.Sprintf("%s?format=json", modelURL)
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		// Network error - validate model exists before creating manifest
		valid, err := t.modelValidator.ValidateModelExists(ctx, modelURL)
		if err != nil {
			return nil, fmt.Errorf("failed to validate model existence: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
		}
		// Model exists but network error - create basic manifest
		return t.createBasicManifest(namespace, name, version, publisher, modelPath, modelURL), nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check if model exists - 404 means model not found
	if resp.StatusCode == http.StatusNotFound {
		// Metadata API returned 404 - validate model page exists
		valid, err := t.modelValidator.ValidateModelExists(ctx, modelURL)
		if err != nil {
			return nil, fmt.Errorf("failed to validate model existence: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
		}
		// Model page exists but metadata API unavailable - create basic manifest
		return t.createBasicManifest(namespace, name, version, publisher, modelPath, modelURL), nil
	}

	// Other non-200 status codes - validate model exists
	if resp.StatusCode != http.StatusOK {
		// Validate model exists by checking the base model URL
		valid, err := t.modelValidator.ValidateModelExists(ctx, modelURL)
		if err != nil {
			return nil, fmt.Errorf("failed to validate model existence: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
		}
		// Model exists but metadata unavailable - create basic manifest
		return t.createBasicManifest(namespace, name, version, publisher, modelPath, modelURL), nil
	}

	var metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
		Format      string `json:"format"` // "saved_model" or "tflite"
		Inputs      []struct {
			Name  string `json:"name"`
			DType string `json:"dtype"`
			Shape []int  `json:"shape"`
		} `json:"inputs"`
		Outputs []struct {
			Name  string `json:"name"`
			DType string `json:"dtype"`
			Shape []int  `json:"shape"`
		} `json:"outputs"`
	}

	// Decode metadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		// If decode fails, validate model exists before creating basic manifest
		valid, err := t.modelValidator.ValidateModelExists(ctx, modelURL)
		if err != nil {
			return nil, fmt.Errorf("failed to validate model existence: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
		}
		// Model exists but metadata decode failed - create basic manifest
		return t.createBasicManifest(namespace, name, version, publisher, modelPath, modelURL), nil
	}

	// Determine file extension based on format
	fileExt := ".tar.gz" // SavedModel format is typically tar.gz
	if metadata.Format == "tflite" {
		fileExt = ".tflite"
	}

	// Create manifest with metadata
	manifest := &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: metadata.Description,
			License:     "Apache-2.0", // TensorFlow models typically use Apache-2.0
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "TensorFlow",
				Version: "2.0.0", // Will be determined from actual model
			},
			Format: types.Format{
				Type: metadata.Format,
				Files: []types.ModelFile{
					{
						Path:   fmt.Sprintf("model%s", fileExt),
						Size:   0,  // Will be determined during download
						SHA256: "", // Will be computed during download
					},
				},
			},
			IO: types.IO{
				Inputs:  []types.IOSpec{},
				Outputs: []types.IOSpec{},
			},
			Requirements: types.Requirements{
				Compute: types.Compute{
					CPU: types.CPURequirement{
						MinCores:         2,
						RecommendedCores: 4,
					},
					Memory: types.MemoryRequirement{
						MinGB:         2.0,
						RecommendedGB: 4.0,
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       t.baseURL,
				Namespace: "tfhub",
			},
		},
	}

	// Add input/output specs if available
	for _, input := range metadata.Inputs {
		manifest.Spec.IO.Inputs = append(manifest.Spec.IO.Inputs, types.IOSpec{
			Name:  input.Name,
			DType: input.DType,
			Shape: input.Shape,
		})
	}

	for _, output := range metadata.Outputs {
		manifest.Spec.IO.Outputs = append(manifest.Spec.IO.Outputs, types.IOSpec{
			Name:  output.Name,
			DType: output.DType,
			Shape: output.Shape,
		})
	}

	return manifest, nil
}

// createBasicManifest creates a basic manifest when metadata is not available
func (t *TensorFlowHubAdapter) createBasicManifest(namespace, name, version, publisher, modelPath, modelURL string) *types.Manifest {
	return &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: fmt.Sprintf("TensorFlow Hub model: %s/%s", publisher, modelPath),
			License:     "Apache-2.0",
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "TensorFlow",
				Version: "2.0.0",
			},
			Format: types.Format{
				Type: "saved_model",
				Files: []types.ModelFile{
					{
						Path:   "model.tar.gz",
						Size:   0,
						SHA256: "",
					},
				},
			},
			IO: types.IO{
				Inputs: []types.IOSpec{
					{
						Name:  "input",
						DType: "float32",
						Shape: []int{-1, -1},
					},
				},
				Outputs: []types.IOSpec{
					{
						Name:  "output",
						DType: "float32",
						Shape: []int{-1, -1},
					},
				},
			},
			Requirements: types.Requirements{
				Compute: types.Compute{
					CPU: types.CPURequirement{
						MinCores:         2,
						RecommendedCores: 4,
					},
					Memory: types.MemoryRequirement{
						MinGB:         2.0,
						RecommendedGB: 4.0,
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: modelURL,
			},
			Registry: types.RegistryInfo{
				URL:       t.baseURL,
				Namespace: "tfhub",
			},
		},
	}
}

// DownloadPackage downloads the model package to the specified destination path.
// TensorFlow Hub models are typically SavedModel format (tar.gz) or TFLite format
func (t *TensorFlowHubAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
	// Parse model specification
	parts := strings.Split(manifest.Metadata.Name, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid TensorFlow Hub model format: %s", manifest.Metadata.Name)
	}

	publisher := parts[0]
	modelPath := strings.Join(parts[1:], "/")
	version := manifest.Metadata.Version

	// Construct model URL
	modelURL := fmt.Sprintf("%s/%s/%s", t.baseURL, publisher, modelPath)
	if version != "latest" && version != "" {
		modelURL = fmt.Sprintf("%s/%s", modelURL, version)
	}

	// TensorFlow Hub models are downloaded as tar.gz files
	// URL format: https://tfhub.dev/{publisher}/{model_path}/{version}?tf-hub-format=compressed
	downloadURL := fmt.Sprintf("%s?tf-hub-format=compressed", modelURL)

	// Create temp directory for model files
	tempDir := filepath.Join(filepath.Dir(destPath), "tmp", fmt.Sprintf("tfhub-%s-%s-%d", publisher, strings.ReplaceAll(modelPath, "/", "-"), time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Download the model
	modelFile := filepath.Join(tempDir, "model.tar.gz")
	if err := core.DownloadFile(ctx, t.httpClient, downloadURL, modelFile, progress); err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Create .axon package
	builder, err := core.NewPackageBuilder()
	if err != nil {
		return fmt.Errorf("failed to create package builder: %w", err)
	}
	defer builder.Cleanup()

	if err := builder.AddFile(modelFile, "model.tar.gz"); err != nil {
		return fmt.Errorf("failed to add file to package: %w", err)
	}

	if err := builder.Build(destPath); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}

	// Update manifest with checksum
	if err := core.UpdateManifestWithChecksum(manifest, destPath); err != nil {
		fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
	}

	return nil
}
