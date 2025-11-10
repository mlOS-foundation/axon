// Package registry provides adapters for different model registries (Hugging Face, local, etc.).
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
	"regexp"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// ModelValidator provides generic model existence validation for adapters
type ModelValidator struct {
	httpClient *http.Client
}

// NewModelValidator creates a new model validator
func NewModelValidator() *ModelValidator {
	return &ModelValidator{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateModelExists checks if a model exists at the given URL
// Returns true if model exists, false if not found, error for validation failures
// This is a generic helper that can be used by all adapters
// Uses GET request with redirect following to handle repositories that don't support HEAD
func (mv *ModelValidator) ValidateModelExists(ctx context.Context, modelURL string) (bool, error) {
	// Use GET request (some repositories like TensorFlow Hub don't support HEAD properly)
	// Limit response size to avoid downloading large files
	req, err := http.NewRequestWithContext(ctx, "GET", modelURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Range header to only request first few bytes (validation only)
	req.Header.Set("Range", "bytes=0-1023")

	// Create client that follows redirects
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		// Network error - can't validate, return error so caller can decide
		return false, fmt.Errorf("network error during validation: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 404 means definitely doesn't exist
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	// 200-299 means got a response (including 206 Partial Content from Range request)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// For HTML responses, check if it's an error/search page
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			// Read a small portion to check for error indicators
			bodyBytes := make([]byte, 2048) // Read first 2KB
			n, _ := resp.Body.Read(bodyBytes)
			bodyStr := strings.ToLower(string(bodyBytes[:n]))

			// Check for common error/search page indicators
			// TensorFlow Hub redirects non-existent models to search page
			errorIndicators := []string{
				"<title>find pre-trained models",
				"<title>search",
				"model not found",
				"does not exist",
				"404",
				"page not found",
			}

			// Check if it looks like a search/error page
			for _, indicator := range errorIndicators {
				if strings.Contains(bodyStr, indicator) {
					// If it's a search page title, likely model doesn't exist
					if strings.Contains(bodyStr, "<title>find pre-trained models") ||
						strings.Contains(bodyStr, "<title>search") {
						return false, nil
					}
				}
			}

			// Check if URL was redirected to search/browse page (but not model page)
			finalURL := resp.Request.URL.String()
			// TensorFlow Hub redirects to Kaggle, but valid models go to model pages
			// Invalid models go to search/browse pages without publisher/model path
			if strings.Contains(finalURL, "/models") && !strings.Contains(finalURL, "/google/") &&
				!strings.Contains(finalURL, "/tensorflow/") && !strings.Contains(finalURL, "/publisher/") {
				// Redirected to general models page - model doesn't exist
				return false, nil
			}
		}
		return true, nil
	}

	// 416 Range Not Satisfiable might mean file exists but is smaller than requested range
	// This is actually a good sign - the file exists
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return true, nil
	}

	// Other status codes (401, 403, 500, etc.) - assume might exist
	// Could be auth required, server error, etc.
	// Return true to allow adapter to handle it
	return true, nil
}

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
	httpClient     *http.Client
	baseURL        string
	token          string // Optional HF token for gated/private models
	modelValidator *ModelValidator
}

// NewHuggingFaceAdapter creates a new Hugging Face adapter
func NewHuggingFaceAdapter() *HuggingFaceAdapter {
	return &HuggingFaceAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for large downloads
		},
		baseURL:        "https://huggingface.co",
		token:          "", // No token by default - works for public models
		modelValidator: NewModelValidator(),
	}
}

// NewHuggingFaceAdapterWithToken creates a Hugging Face adapter with authentication token
func NewHuggingFaceAdapterWithToken(token string) *HuggingFaceAdapter {
	return &HuggingFaceAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://huggingface.co",
		token:          token,
		modelValidator: NewModelValidator(),
	}
}

// SetToken sets the Hugging Face token (for gated/private models)
func (h *HuggingFaceAdapter) SetToken(token string) {
	h.token = token
}

// Name returns the name of the adapter.
func (h *HuggingFaceAdapter) Name() string {
	return "huggingface"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
func (h *HuggingFaceAdapter) CanHandle(namespace, name string) bool {
	// Hugging Face can handle any model - it's a fallback/default
	return true
}

// Search searches for models matching the query.
func (h *HuggingFaceAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// Use Hugging Face API to search
	url := fmt.Sprintf("%s/api/models?search=%s", h.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header if token is provided
	if h.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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
	valid, err := h.modelValidator.ValidateModelExists(ctx, modelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate model existence: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("model not found: %s/%s@%s", namespace, name, version)
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
						Size:   0,  // Will be determined during download
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

// DownloadPackage downloads the model package to the specified destination path.
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
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

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

		// Add authentication header if token is provided (needed for gated/private models)
		if h.token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
		}

		resp, err := h.httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			continue // Skip missing files
		}

		filePath := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			_ = resp.Body.Close()
			continue
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			_ = resp.Body.Close()
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
			_ = outFile.Close()
			_ = resp.Body.Close()
			continue
		}

		_ = outFile.Close()
		_ = resp.Body.Close()
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

	// Add authentication header if token is provided
	if h.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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
	defer func() {
		_ = file.Close()
	}()

	gzWriter := gzip.NewWriter(file)
	defer func() {
		_ = gzWriter.Close()
	}()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		_ = tarWriter.Close()
	}()

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
		defer func() {
			_ = srcFile.Close()
		}()

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
	defer func() {
		_ = file.Close()
	}()

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

// Name returns the name of the adapter.
func (l *LocalRegistryAdapter) Name() string {
	return "local"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
func (l *LocalRegistryAdapter) CanHandle(namespace, name string) bool {
	// Local registry can only handle models that are NOT from known adapters
	// Known adapter namespaces: hf, pytorch, torch, modelscope, tfhub, tf
	if namespace == "hf" || namespace == "pytorch" || namespace == "torch" ||
		namespace == "modelscope" || namespace == "tfhub" || namespace == "tf" {
		return false
	}
	// Local registry can handle models if it's configured and model is not from a known adapter
	return l.client.baseURL != ""
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

// PyTorchHubAdapter implements RepositoryAdapter for PyTorch Hub
// PyTorch Hub models are hosted on GitHub repositories (e.g., pytorch/vision, pytorch/text)
// Models are defined via hubconf.py files and can be loaded via torch.hub.load()
type PyTorchHubAdapter struct {
	httpClient     *http.Client
	baseURL        string // GitHub API base URL
	githubToken    string // Optional GitHub token for rate limit increases
	modelValidator *ModelValidator
}

// NewPyTorchHubAdapter creates a new PyTorch Hub adapter
func NewPyTorchHubAdapter() *PyTorchHubAdapter {
	return &PyTorchHubAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://api.github.com",
		githubToken:    "", // No token by default
		modelValidator: NewModelValidator(),
	}
}

// NewPyTorchHubAdapterWithToken creates a PyTorch Hub adapter with GitHub token
func NewPyTorchHubAdapterWithToken(token string) *PyTorchHubAdapter {
	return &PyTorchHubAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://api.github.com",
		githubToken:    token,
		modelValidator: NewModelValidator(),
	}
}

// SetToken sets the GitHub token (for rate limit increases)
func (p *PyTorchHubAdapter) SetToken(token string) {
	p.githubToken = token
}

// Name returns the name of the adapter.
func (p *PyTorchHubAdapter) Name() string {
	return "pytorch"
}

// CanHandle returns true if this adapter can handle the given namespace and name.
func (p *PyTorchHubAdapter) CanHandle(namespace, name string) bool {
	// PyTorch Hub can handle models with "pytorch" or "torch" namespace
	return namespace == "pytorch" || namespace == "torch"
}

// Search searches for models matching the query.
// PyTorch Hub doesn't have a direct search API, so we search GitHub repositories
// that are known to host PyTorch Hub models (pytorch/vision, pytorch/text, etc.)
func (p *PyTorchHubAdapter) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	// PyTorch Hub models are primarily in these repositories:
	// - pytorch/vision (computer vision models)
	// - pytorch/text (NLP models)
	// - pytorch/audio (audio models)

	// For now, return empty results as PyTorch Hub doesn't have a search API
	// In a full implementation, we could:
	// 1. Search GitHub for hubconf.py files containing the query
	// 2. Parse hubconf.py to extract model names
	// 3. Return search results

	var results []types.SearchResult

	// TODO: Implement GitHub-based search for PyTorch Hub models
	// This would involve searching for hubconf.py files in known PyTorch repos

	return results, nil
}

// GetManifest retrieves the manifest for the specified model.
// Model format: pytorch/{repo}/{model_name}@version
// Example: pytorch/vision/resnet50@latest
func (p *PyTorchHubAdapter) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	// Parse model specification
	// Format: {repo}/{model_name} (e.g., "vision/resnet50")
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid PyTorch Hub model format: %s (expected: repo/model_name)", name)
	}

	repo := parts[0]                          // e.g., "vision", "text", "audio"
	modelName := strings.Join(parts[1:], "/") // e.g., "resnet50"

	// Construct GitHub repo path (PyTorch Hub repos are under pytorch/ organization)
	githubRepo := fmt.Sprintf("pytorch/%s", repo)

	// Validate GitHub repository exists
	repoURL := fmt.Sprintf("https://github.com/%s", githubRepo)
	valid, err := p.modelValidator.ValidateModelExists(ctx, repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to validate repository existence: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("repository not found: %s (model: %s/%s@%s)", githubRepo, namespace, name, version)
	}

	// Validate that the specific model exists in hubconf.py
	hubconfURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/hubconf.py", githubRepo)
	hubconfReq, err := http.NewRequestWithContext(ctx, "GET", hubconfURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create hubconf request: %w", err)
	}
	if p.githubToken != "" {
		hubconfReq.Header.Set("Authorization", fmt.Sprintf("token %s", p.githubToken))
	}

	hubconfResp, err := p.httpClient.Do(hubconfReq)
	if err != nil {
		// If we can't fetch hubconf.py, assume model might exist
		// (could be network issue, not necessarily model doesn't exist)
	} else {
		defer func() {
			_ = hubconfResp.Body.Close()
		}()

		if hubconfResp.StatusCode == http.StatusOK {
			hubconfContent, err := io.ReadAll(hubconfResp.Body)
			if err == nil {
				// Check if model exists in hubconf.py
				modelURLs := p.parseHubconf(hubconfContent, modelName)
				if len(modelURLs) == 0 {
					return nil, fmt.Errorf("model not found in hubconf.py: %s (repository: %s)", modelName, githubRepo)
				}
			}
		}
	}

	// Create manifest
	manifest := &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:        name,
			Namespace:   namespace,
			Version:     version,
			Description: fmt.Sprintf("Model from PyTorch Hub: %s/%s", repo, modelName),
			License:     "BSD-3-Clause", // PyTorch models typically use BSD-3-Clause
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "PyTorch",
				Version: "2.0.0", // Will be determined from actual model
			},
			Format: types.Format{
				Type: "pytorch",
				Files: []types.ModelFile{
					{
						Path:   fmt.Sprintf("%s.pth", modelName), // PyTorch model file
						Size:   0,                                // Will be determined during download
						SHA256: "",                               // Will be computed during download
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
				URL: fmt.Sprintf("https://github.com/%s", githubRepo),
				// Model weights will be downloaded from GitHub releases or direct URLs
			},
			Registry: types.RegistryInfo{
				URL:       "https://pytorch.org/hub",
				Namespace: "pytorch",
			},
		},
	}

	return manifest, nil
}

// DownloadPackage downloads the model package to the specified destination path.
// PyTorch Hub models are typically loaded via torch.hub.load() which downloads
// pre-trained weights. We'll need to download these weights directly from GitHub.
func (p *PyTorchHubAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
	// Parse model specification
	parts := strings.Split(manifest.Metadata.Name, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid PyTorch Hub model format: %s", manifest.Metadata.Name)
	}

	repo := parts[0]
	modelName := strings.Join(parts[1:], "/")
	githubRepo := fmt.Sprintf("pytorch/%s", repo)

	// Create temp directory for model files
	tempDir := filepath.Join(filepath.Dir(destPath), "tmp", fmt.Sprintf("pytorch-%s-%s-%d", repo, modelName, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// PyTorch Hub models are typically loaded via torch.hub.load() which:
	// 1. Downloads model weights from GitHub releases or cache
	// 2. Loads the model architecture from the repository

	// For now, we'll try to download from GitHub releases
	// In a full implementation, we might need to:
	// 1. Parse hubconf.py to understand model structure
	// 2. Download weights from GitHub releases
	// 3. Download model architecture files if needed

	// Try to get latest release from GitHub
	releaseURL := fmt.Sprintf("%s/repos/%s/releases/latest", p.baseURL, githubRepo)
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if provided
	if p.githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", p.githubToken))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		// If no release found, try to download from main branch
		// PyTorch Hub models might not have releases
		if err := p.downloadFromBranch(ctx, githubRepo, modelName, tempDir, progress); err != nil {
			return err
		}
		// Create .axon package after downloading from branch
		if err := p.createAxonPackage(tempDir, destPath); err != nil {
			return fmt.Errorf("failed to create package: %w", err)
		}
		// Update manifest with checksum
		if err := p.updateManifestWithChecksum(manifest, destPath); err != nil {
			fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
		}
		return nil
	}

	// Parse release response
	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release response: %w", err)
	}

	// Download model files from release assets
	downloadedFiles := []string{}
	for _, asset := range release.Assets {
		// Filter for model files (.pth, .pt, .pkl, etc.)
		if strings.HasSuffix(asset.Name, ".pth") ||
			strings.HasSuffix(asset.Name, ".pt") ||
			strings.HasSuffix(asset.Name, ".pkl") {

			if err := p.downloadFile(ctx, asset.BrowserDownloadURL, filepath.Join(tempDir, asset.Name), asset.Size, progress); err != nil {
				continue // Skip failed downloads
			}
			downloadedFiles = append(downloadedFiles, asset.Name)
		}
	}

	if len(downloadedFiles) == 0 {
		// Fallback: try downloading from branch
		if err := p.downloadFromBranch(ctx, githubRepo, modelName, tempDir, progress); err != nil {
			return err
		}
		// Create .axon package after downloading from branch
		if err := p.createAxonPackage(tempDir, destPath); err != nil {
			return fmt.Errorf("failed to create package: %w", err)
		}
		// Update manifest with checksum
		if err := p.updateManifestWithChecksum(manifest, destPath); err != nil {
			fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
		}
		return nil
	}

	// Create .axon package
	if err := p.createAxonPackage(tempDir, destPath); err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	// Update manifest with checksum
	if err := p.updateManifestWithChecksum(manifest, destPath); err != nil {
		fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
	}

	return nil
}

// downloadFromBranch downloads model files from a GitHub branch
// This is a pure Go implementation that:
// 1. Fetches hubconf.py from GitHub
// 2. Parses it to extract model weight URLs
// 3. Downloads weights directly from those URLs
func (p *PyTorchHubAdapter) downloadFromBranch(ctx context.Context, githubRepo, modelName, destDir string, progress ProgressCallback) error {
	// Fetch hubconf.py from GitHub
	hubconfURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/hubconf.py", githubRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", hubconfURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if p.githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", p.githubToken))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch hubconf.py: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		// Try alternative branch names
		for _, branch := range []string{"master", "main"} {
			altURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/hubconf.py", githubRepo, branch)
			altReq, err := http.NewRequestWithContext(ctx, "GET", altURL, nil)
			if err != nil {
				continue
			}
			if p.githubToken != "" {
				altReq.Header.Set("Authorization", fmt.Sprintf("token %s", p.githubToken))
			}
			altResp, err := p.httpClient.Do(altReq)
			if err == nil && altResp.StatusCode == http.StatusOK {
				resp = altResp
				break
			}
			if altResp != nil {
				_ = altResp.Body.Close()
			}
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch hubconf.py: status %d", resp.StatusCode)
		}
	}

	hubconfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read hubconf.py: %w", err)
	}

	// Parse hubconf.py to extract model URLs
	// hubconf.py typically contains model_urls dictionary like:
	// model_urls = {
	//     'resnet50': 'https://download.pytorch.org/models/resnet50-19c8e357.pth',
	//     ...
	// }
	modelURLs := p.parseHubconf(hubconfContent, modelName)

	if len(modelURLs) == 0 {
		return fmt.Errorf("no model URLs found in hubconf.py for model: %s", modelName)
	}

	// Download model weights from extracted URLs
	downloadedFiles := []string{}
	for _, url := range modelURLs {
		// Extract filename from URL
		filename := filepath.Base(url)
		// Remove query parameters if any
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}

		destPath := filepath.Join(destDir, filename)
		if err := p.downloadFile(ctx, url, destPath, 0, progress); err != nil {
			fmt.Printf("Warning: failed to download %s: %v\n", url, err)
			continue
		}
		downloadedFiles = append(downloadedFiles, filename)
	}

	if len(downloadedFiles) == 0 {
		return fmt.Errorf("failed to download any model files for %s", modelName)
	}

	return nil
}

// parseHubconf parses hubconf.py to extract model weight URLs
// This is a pure Go implementation using regex to extract URLs from Python code
func (p *PyTorchHubAdapter) parseHubconf(hubconfContent []byte, modelName string) []string {
	content := string(hubconfContent)
	var urls []string

	// Pattern 1: Look for model_urls dictionary
	// Example: model_urls = {'resnet50': 'https://...', ...}
	modelURLsPattern := regexp.MustCompile(`model_urls\s*=\s*\{([^}]+)\}`)
	matches := modelURLsPattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		// Parse dictionary entries
		dictContent := matches[1]
		// Match key-value pairs: 'key': 'url' or "key": "url"
		entryPattern := regexp.MustCompile(`['"](\w+)['"]\s*:\s*['"](https?://[^'"]+)['"]`)
		entries := entryPattern.FindAllStringSubmatch(dictContent, -1)

		for _, entry := range entries {
			if len(entry) >= 3 {
				key := entry[1]
				url := entry[2]
				// Check if this entry matches our model name
				if key == modelName || strings.Contains(modelName, key) {
					urls = append(urls, url)
				}
			}
		}
	}

	// Pattern 2: Look for direct URL assignments
	// Example: resnet50_url = 'https://...'
	directURLPattern := regexp.MustCompile(fmt.Sprintf(`%s_url\s*=\s*['"](https?://[^'"]+)['"]`, regexp.QuoteMeta(modelName)))
	directMatches := directURLPattern.FindAllStringSubmatch(content, -1)
	for _, match := range directMatches {
		if len(match) >= 2 {
			urls = append(urls, match[1])
		}
	}

	// Pattern 3: Look for WeightsEnum patterns (torchvision style)
	// Example: class ResNet50_Weights(WeightsEnum): url="https://..."
	weightsPattern := regexp.MustCompile(fmt.Sprintf(`(?i)class\s+%s.*?Weights.*?url\s*=\s*['"](https?://[^'"]+)['"]`, regexp.QuoteMeta(modelName)))
	weightsMatches := weightsPattern.FindAllStringSubmatch(content, -1)
	for _, match := range weightsMatches {
		if len(match) >= 2 {
			urls = append(urls, match[1])
		}
	}

	// Pattern 4: Look for URLs in function definitions that load the model
	// Example: def resnet50(pretrained=True, ...): ... load_state_dict_from_url('https://...')
	loadURLPattern := regexp.MustCompile(`load_state_dict_from_url\(['"](https?://[^'"]+)['"]`)
	loadMatches := loadURLPattern.FindAllStringSubmatch(content, -1)
	for _, match := range loadMatches {
		if len(match) >= 2 {
			// Check if this is in a function related to our model
			// Find the function name before this URL
			funcPattern := regexp.MustCompile(fmt.Sprintf(`def\s+%s[^:]*:.*?load_state_dict_from_url\(['"](https?://[^'"]+)['"]`, regexp.QuoteMeta(modelName)))
			if funcPattern.MatchString(content) {
				urls = append(urls, match[1])
			}
		}
	}

	// Pattern 5: Try to fetch from model source file if hubconf.py doesn't have URLs
	// This is a fallback for torchvision models that use WeightsEnum
	if len(urls) == 0 {
		// Try to extract repo name from content or use known patterns
		// For now, we'll use a fallback URL pattern for common models
		urls = p.getFallbackURLs(modelName)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueURLs := []string{}
	for _, url := range urls {
		if !seen[url] {
			seen[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}

	return uniqueURLs
}

// getFallbackURLs returns known PyTorch model weight URLs as a fallback
// This handles cases where hubconf.py doesn't contain direct URLs (e.g., torchvision models)
func (p *PyTorchHubAdapter) getFallbackURLs(modelName string) []string {
	// Known PyTorch model URLs (common models from torchvision)
	knownURLs := map[string]string{
		"resnet50":     "https://download.pytorch.org/models/resnet50-0676ba61.pth",
		"resnet101":    "https://download.pytorch.org/models/resnet101-63fe2227.pth",
		"resnet152":    "https://download.pytorch.org/models/resnet152-394f9c45.pth",
		"resnet18":     "https://download.pytorch.org/models/resnet18-f37072fd.pth",
		"resnet34":     "https://download.pytorch.org/models/resnet34-b627a593.pth",
		"alexnet":      "https://download.pytorch.org/models/alexnet-owt-7be5be79.pth",
		"vgg16":        "https://download.pytorch.org/models/vgg16-397923af.pth",
		"vgg19":        "https://download.pytorch.org/models/vgg19-dcbb9e9d.pth",
		"mobilenet_v2": "https://download.pytorch.org/models/mobilenet_v2-7ebf99e0.pth",
	}

	// Try exact match first
	if url, ok := knownURLs[modelName]; ok {
		return []string{url}
	}

	// Try partial match (e.g., "vision/resnet50" -> "resnet50")
	parts := strings.Split(modelName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if url, ok := knownURLs[lastPart]; ok {
			return []string{url}
		}
	}

	return []string{}
}

// downloadFile downloads a file from a URL
func (p *PyTorchHubAdapter) downloadFile(ctx context.Context, url, destPath string, expectedSize int64, progress ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if p.githubToken != "" && strings.Contains(url, "github.com") {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", p.githubToken))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = outFile.Close()
	}()

	// Copy with progress tracking
	reader := &progressReader{
		Reader:     resp.Body,
		Total:      resp.ContentLength,
		Downloaded: 0,
		Callback:   progress,
	}

	_, err = io.Copy(outFile, reader)
	return err
}

// createAxonPackage creates a tar.gz package from a directory
func (p *PyTorchHubAdapter) createAxonPackage(srcDir, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzWriter := gzip.NewWriter(file)
	defer func() {
		_ = gzWriter.Close()
	}()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		_ = tarWriter.Close()
	}()

	// Walk directory and add files to tar
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = srcFile.Close()
		}()

		_, err = io.Copy(tarWriter, srcFile)
		return err
	})
}

// updateManifestWithChecksum updates manifest with computed checksum
func (p *PyTorchHubAdapter) updateManifestWithChecksum(manifest *types.Manifest, packagePath string) error {
	hasher := sha256.New()
	file, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

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

// TensorFlowHubAdapter implements RepositoryAdapter for TensorFlow Hub
// TensorFlow Hub models are hosted at https://tfhub.dev
// Models are organized by publisher (e.g., google, tensorflow) and can be SavedModel or TFLite format
type TensorFlowHubAdapter struct {
	httpClient     *http.Client
	baseURL        string // TensorFlow Hub base URL
	modelValidator *ModelValidator
}

// NewTensorFlowHubAdapter creates a new TensorFlow Hub adapter
func NewTensorFlowHubAdapter() *TensorFlowHubAdapter {
	return &TensorFlowHubAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://tfhub.dev",
		modelValidator: NewModelValidator(),
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
func (t *TensorFlowHubAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
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
	if err := t.downloadFile(ctx, downloadURL, modelFile, 0, progress); err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	// Create .axon package
	if err := t.createAxonPackage(tempDir, destPath); err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	// Update manifest with checksum
	if err := t.updateManifestWithChecksum(manifest, destPath); err != nil {
		fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
	}

	return nil
}

// downloadFile downloads a file from URL to destination with progress tracking
func (t *TensorFlowHubAdapter) downloadFile(ctx context.Context, url, destPath string, expectedSize int64, progress ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	// Copy with progress tracking
	reader := &progressReader{
		Reader:     resp.Body,
		Total:      resp.ContentLength,
		Downloaded: 0,
		Callback:   progress,
	}

	_, err = io.Copy(outFile, reader)
	return err
}

// createAxonPackage creates a tar.gz package from a directory
func (t *TensorFlowHubAdapter) createAxonPackage(srcDir, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzWriter := gzip.NewWriter(file)
	defer func() {
		_ = gzWriter.Close()
	}()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		_ = tarWriter.Close()
	}()

	// Walk directory and add files to tar
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = srcFile.Close()
		}()

		_, err = io.Copy(tarWriter, srcFile)
		return err
	})
}

// updateManifestWithChecksum updates manifest with computed checksum
func (t *TensorFlowHubAdapter) updateManifestWithChecksum(manifest *types.Manifest, packagePath string) error {
	hasher := sha256.New()
	file, err := os.Open(packagePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

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
