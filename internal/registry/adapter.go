package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// RepositoryAdapter defines the interface for different model repositories
// This allows Axon to support multiple sources: Hugging Face, local registry, custom registries, etc.
type RepositoryAdapter interface {
	// Name returns the adapter name (e.g., "huggingface", "local", "custom")
	Name() string

	// Search searches for models in the repository
	Search(ctx context.Context, query string) ([]types.SearchResult, error)

	// GetManifest retrieves a model manifest
	GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error)

	// DownloadPackage downloads a model package to the destination path
	DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error

	// CanHandle checks if this adapter can handle the given model specification
	CanHandle(namespace, name string) bool
}

// HuggingFaceAdapter implements RepositoryAdapter for Hugging Face Hub
type HuggingFaceAdapter struct {
	httpClient *http.Client
	baseURL    string
}

// NewHuggingFaceAdapter creates a new Hugging Face adapter
func NewHuggingFaceAdapter() *HuggingFaceAdapter {
	return &HuggingFaceAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for large downloads
		},
		baseURL: "https://huggingface.co",
	}
}

func (h *HuggingFaceAdapter) Name() string {
	return "huggingface"
}

func (h *HuggingFaceAdapter) CanHandle(namespace, name string) bool {
	// Hugging Face can handle any model - it's a fallback/default
	return true
}

func (h *HuggingFaceAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// Use Hugging Face API to search
	url := fmt.Sprintf("%s/api/models?search=%s", h.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
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

func (h *HuggingFaceAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// For Hugging Face, we generate a manifest on-the-fly
	// In production, this would fetch model metadata from HF API

	// Construct HF model ID
	hfModelID := name
	if namespace != "" && namespace != "hf" {
		hfModelID = fmt.Sprintf("%s/%s", namespace, name)
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
				Type: "pytorch",
				Files: []types.ModelFile{
					{
						Path:   "pytorch_model.bin",
						Size:   0, // Will be determined during download
						SHA256: "", // Will be computed during download
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
				URL: fmt.Sprintf("%s/%s/resolve/main/pytorch_model.bin", h.baseURL, hfModelID),
				// In production, would fetch from HF API and get actual URLs
			},
			Registry: types.RegistryInfo{
				URL:       h.baseURL,
				Namespace: "huggingface",
			},
		},
	}

	return manifest, nil
}

func (h *HuggingFaceAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
	// For Hugging Face, we download model files in real-time and create a package
	hfModelID := manifest.Metadata.Name
	if manifest.Metadata.Namespace != "" && manifest.Metadata.Namespace != "hf" {
		hfModelID = fmt.Sprintf("%s/%s", manifest.Metadata.Namespace, manifest.Metadata.Name)
	}

	// Create temp directory for model files
	tempDir := filepath.Join(filepath.Dir(destPath), "tmp", fmt.Sprintf("hf-%s-%d", hfModelID, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// First, try to get model file list from Hugging Face API
	modelFiles, err := h.getModelFiles(ctx, hfModelID)
	if err != nil {
		// Fallback to common files if API fails
		modelFiles = []string{"config.json", "pytorch_model.bin", "tokenizer_config.json", "vocab.txt", "vocab.json"}
	}

	// Download files from Hugging Face
	downloadedFiles := []string{}
	for _, file := range modelFiles {
		url := fmt.Sprintf("%s/%s/resolve/main/%s", h.baseURL, hfModelID, file)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := h.httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue // Skip missing files
		}

		filePath := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			resp.Body.Close()
			continue
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			continue
		}

		// Copy with progress tracking
		reader := &progressReader{
			Reader:     resp.Body,
			Total:      resp.ContentLength,
			Downloaded: 0,
			Callback:   progress,
		}

		if _, err := io.Copy(outFile, reader); err != nil {
			outFile.Close()
			resp.Body.Close()
			continue
		}

		outFile.Close()
		resp.Body.Close()
		downloadedFiles = append(downloadedFiles, file)
	}

	if len(downloadedFiles) == 0 {
		return fmt.Errorf("no files downloaded from Hugging Face for %s", hfModelID)
	}

	// Create .axon package (tar.gz)
	if err := h.createAxonPackage(tempDir, destPath); err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	// Update manifest with checksum and size
	if err := h.updateManifestWithChecksum(manifest, destPath); err != nil {
		// Non-fatal - package is created, just checksum update failed
		fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
	}

	return nil
}

// getModelFiles fetches the list of files from Hugging Face API
func (h *HuggingFaceAdapter) getModelFiles(ctx context.Context, modelID string) ([]string, error) {
	// Use Hugging Face API to get file list
	url := fmt.Sprintf("%s/api/models/%s", h.baseURL, modelID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
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

	files := make([]string, 0, len(modelInfo.Siblings))
	for _, sibling := range modelInfo.Siblings {
		// Filter to essential model files
		if strings.HasSuffix(sibling.RFileName, ".json") ||
			strings.HasSuffix(sibling.RFileName, ".bin") ||
			strings.HasSuffix(sibling.RFileName, ".txt") ||
			strings.HasSuffix(sibling.RFileName, ".safetensors") {
			files = append(files, sibling.RFileName)
		}
	}

	return files, nil
}

// createAxonPackage creates a tar.gz package from a directory
func (h *HuggingFaceAdapter) createAxonPackage(srcDir, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk directory and add files to tar
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Copy file content
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if _, err := io.Copy(tarWriter, srcFile); err != nil {
			return err
		}

		return nil
	})
}

// UpdateManifestWithChecksum updates manifest with computed checksum
func (h *HuggingFaceAdapter) updateManifestWithChecksum(manifest *types.Manifest, packagePath string) error {
	hasher := sha256.New()
	file, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	stat, err := os.Stat(packagePath)
	if err != nil {
		return err
	}

	manifest.Distribution.Package.SHA256 = checksum
	manifest.Distribution.Package.Size = stat.Size()
	return nil
}

// LocalRegistryAdapter implements RepositoryAdapter for local Axon registry
type LocalRegistryAdapter struct {
	client *Client
}

// NewLocalRegistryAdapter creates a new local registry adapter
func NewLocalRegistryAdapter(baseURL string, mirrors []string) *LocalRegistryAdapter {
	return &LocalRegistryAdapter{
		client: NewClient(baseURL, mirrors),
	}
}

func (l *LocalRegistryAdapter) Name() string {
	return "local"
}

func (l *LocalRegistryAdapter) CanHandle(namespace, name string) bool {
	// Local registry can handle any model if it's configured
	return l.client.baseURL != ""
}

func (l *LocalRegistryAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	return l.client.Search(ctx, query)
}

func (l *LocalRegistryAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	return l.client.GetManifest(ctx, namespace, name, version)
}

func (l *LocalRegistryAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
	return l.client.DownloadPackage(ctx, manifest, destPath, progress)
}

// AdapterRegistry manages multiple repository adapters
type AdapterRegistry struct {
	adapters []RepositoryAdapter
}

// NewAdapterRegistry creates a new adapter registry
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: []RepositoryAdapter{},
	}
}

// Register adds a new adapter
func (ar *AdapterRegistry) Register(adapter RepositoryAdapter) {
	ar.adapters = append(ar.adapters, adapter)
}

// FindAdapter finds the best adapter for a given model specification
func (ar *AdapterRegistry) FindAdapter(namespace, name string) (RepositoryAdapter, error) {
	// Try adapters in order - first match wins
	for _, adapter := range ar.adapters {
		if adapter.CanHandle(namespace, name) {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("no adapter found for %s/%s", namespace, name)
}

// GetAllAdapters returns all registered adapters
func (ar *AdapterRegistry) GetAllAdapters() []RepositoryAdapter {
	return ar.adapters
}

