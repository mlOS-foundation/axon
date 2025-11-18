// Package builtin provides default adapters included with Axon.
// These adapters are registered automatically and provide support for
// popular model repositories.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// HuggingFaceAdapter implements RepositoryAdapter for Hugging Face Hub.
// Hugging Face is the most popular model repository with 100,000+ models.
type HuggingFaceAdapter struct {
	httpClient *core.HTTPClient
	baseURL    string
	token      string
	validator  *core.ModelValidator
}

// NewHuggingFaceAdapter creates a new Hugging Face adapter.
func NewHuggingFaceAdapter() *HuggingFaceAdapter {
	client := core.NewHTTPClient("https://huggingface.co", 5*time.Minute)
	return &HuggingFaceAdapter{
		httpClient: client,
		baseURL:    "https://huggingface.co",
		token:      "",
		validator:  core.NewModelValidator(),
	}
}

// NewHuggingFaceAdapterWithToken creates a Hugging Face adapter with authentication token.
func NewHuggingFaceAdapterWithToken(token string) *HuggingFaceAdapter {
	adapter := NewHuggingFaceAdapter()
	adapter.token = token
	adapter.httpClient.SetToken(token)
	return adapter
}

// SetToken sets the Hugging Face token (for gated/private models).
func (h *HuggingFaceAdapter) SetToken(token string) {
	h.token = token
	h.httpClient.SetToken(token)
}

// Name returns the adapter name.
func (h *HuggingFaceAdapter) Name() string {
	return "huggingface"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
// Hugging Face can handle any model - it's a fallback/default.
func (h *HuggingFaceAdapter) CanHandle(namespace, name string) bool {
	return true
}

// Search searches for models matching the query.
func (h *HuggingFaceAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	url := fmt.Sprintf("%s/api/models?search=%s", h.baseURL, query)

	resp, err := h.httpClient.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse Hugging Face API response
	// This is a simplified version - real implementation would parse HF's JSON format
	var results []types.SearchResult

	// For now, return empty - this would need HF API parsing
	// In real implementation, we'd parse HF's model list response
	return results, nil
}

// GetManifest retrieves the manifest for the specified model.
func (h *HuggingFaceAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// Construct HF model ID and URL
	hfModelID := name
	if namespace != "" && namespace != "hf" {
		hfModelID = fmt.Sprintf("%s/%s", namespace, name)
	}

	// Validate model exists on Hugging Face
	modelURL := fmt.Sprintf("%s/%s", h.baseURL, hfModelID)
	valid, err := h.validator.ValidateModelExists(ctx, modelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate model existence: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
	}

	// Try to fetch config.json to extract I/O schema
	// This is optional - if it fails, we'll use generic I/O schema
	var inputs, outputs []types.IOSpec
	configURL := fmt.Sprintf("%s/%s/resolve/main/config.json", h.baseURL, hfModelID)
	tempConfig := filepath.Join(os.TempDir(), fmt.Sprintf("axon-config-%d.json", time.Now().UnixNano()))

	if resp, err := h.httpClient.Get(ctx, configURL); err == nil && resp.StatusCode == http.StatusOK {
		// Download config.json temporarily
		if file, err := os.Create(tempConfig); err == nil {
			io.Copy(file, resp.Body)
			file.Close()
			resp.Body.Close()

			// Extract I/O schema from config
			if extractedInputs, extractedOutputs, err := ExtractIOSchemaFromConfig(tempConfig); err == nil {
				inputs = extractedInputs
				outputs = extractedOutputs
			}
			os.Remove(tempConfig) // Clean up
		}
	}

	// Fallback to generic I/O schema if extraction failed
	if len(inputs) == 0 {
		inputs = []types.IOSpec{
			{
				Name:  "input",
				DType: "float32",
				Shape: []int{-1, -1},
			},
		}
		outputs = []types.IOSpec{
			{
				Name:  "output",
				DType: "float32",
				Shape: []int{-1, -1},
			},
		}
	}

	// Create manifest with HF download URLs
	manifest := &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: fmt.Sprintf("Model from Hugging Face: %s", hfModelID),
			License:     "Unknown", // Would fetch from HF API
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "PyTorch",
				Version: "2.0.0",
			},
			Format: types.Format{
				Type:            "pytorch",
				ExecutionFormat: "onnx", // Default to ONNX (will be updated after conversion)
				Files: []types.ModelFile{
					{
						Path:   "pytorch_model.bin",
						Size:   0,  // Will be determined during download
						SHA256: "", // Will be computed during download
					},
				},
			},
			IO: types.IO{
				Inputs:  inputs,
				Outputs: outputs,
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
				URL: fmt.Sprintf("%s/%s/resolve/main/pytorch_model.bin", h.baseURL, hfModelID),
			},
			Registry: types.RegistryInfo{
				URL:       h.baseURL,
				Namespace: "huggingface",
			},
		},
	}

	return manifest, nil
}

// DownloadPackage downloads the model package to the specified destination path.
func (h *HuggingFaceAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
	// For Hugging Face, we download model files in real-time and create a package
	hfModelID := manifest.Metadata.Name
	if manifest.Metadata.Namespace != "" && manifest.Metadata.Namespace != "hf" {
		hfModelID = fmt.Sprintf("%s/%s", manifest.Metadata.Namespace, manifest.Metadata.Name)
	}

	// Create package builder
	builder, err := core.NewPackageBuilder()
	if err != nil {
		return fmt.Errorf("failed to create package builder: %w", err)
	}
	defer builder.Cleanup()

	// Get model file list from Hugging Face API
	modelFiles, err := h.getModelFiles(ctx, hfModelID)
	if err != nil {
		// Fallback to common files if API fails
		modelFiles = []string{"config.json", "pytorch_model.bin", "tokenizer.json", "tokenizer_config.json", "vocab.txt", "vocab.json"}
	}

	// Ensure tokenizer files are included
	tokenizerFiles := []string{"tokenizer.json", "tokenizer_config.json", "vocab.txt", "vocab.json"}
	for _, tokenizerFile := range tokenizerFiles {
		// Check if already in list
		found := false
		for _, file := range modelFiles {
			if file == tokenizerFile {
				found = true
				break
			}
		}
		if !found {
			// Try to add tokenizer file (will be skipped if not available)
			modelFiles = append(modelFiles, tokenizerFile)
		}
	}

	// Download files from Hugging Face
	httpClient := &http.Client{Timeout: 10 * time.Minute}
	downloadedFiles := []string{}

	for _, file := range modelFiles {
		url := fmt.Sprintf("%s/%s/resolve/main/%s", h.baseURL, hfModelID, file)

		// Create temp file for download
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("axon-hf-%s-%d", file, time.Now().UnixNano()))

		// Add auth header if token is provided
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		if h.token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
		}

		resp, err := httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				_ = resp.Body.Close()
			}
			continue // Skip missing files
		}

		// Download file
		if err := core.DownloadFile(ctx, httpClient, url, tempFile, progress); err != nil {
			_ = resp.Body.Close()
			continue
		}
		_ = resp.Body.Close()

		// Add to package
		if err := builder.AddFile(tempFile, file); err != nil {
			_ = os.Remove(tempFile)
			continue
		}

		downloadedFiles = append(downloadedFiles, file)
		_ = os.Remove(tempFile) // Clean up temp file
	}

	if len(downloadedFiles) == 0 {
		return fmt.Errorf("no files downloaded from Hugging Face for %s", hfModelID)
	}

	// Build package
	if err := builder.Build(destPath); err != nil {
		return fmt.Errorf("failed to build package: %w", err)
	}

	// Update manifest with checksum
	if err := core.UpdateManifestWithChecksum(manifest, destPath); err != nil {
		fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
	}

	return nil
}

// getModelFiles fetches the list of files from Hugging Face API.
func (h *HuggingFaceAdapter) getModelFiles(ctx context.Context, modelID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/models/%s", h.baseURL, modelID)

	resp, err := h.httpClient.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modelInfo struct {
		Siblings []struct {
			RFileName string `json:"rfilename"`
		} `json:"siblings"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelInfo); err != nil {
		return nil, err
	}

	var files []string
	for _, sibling := range modelInfo.Siblings {
		files = append(files, sibling.RFileName)
	}

	return files, nil
}
