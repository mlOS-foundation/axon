// Package builtin provides default adapters included with Axon.
package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

type PyTorchHubAdapter struct {
	httpClient     *http.Client
	baseURL        string // GitHub API base URL
	githubToken    string // Optional GitHub token for rate limit increases
	modelValidator *core.ModelValidator
}

// NewPyTorchHubAdapter creates a new PyTorch Hub adapter
func NewPyTorchHubAdapter() *PyTorchHubAdapter {
	return &PyTorchHubAdapter{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:        "https://api.github.com",
		githubToken:    "", // No token by default
		modelValidator: core.NewModelValidator(),
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
		modelValidator: core.NewModelValidator(),
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
func (p *PyTorchHubAdapter) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress core.ProgressCallback) error {
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
		builder, err := core.NewPackageBuilder()
		if err != nil {
			return fmt.Errorf("failed to create package builder: %w", err)
		}
		defer builder.Cleanup()

		if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(tempDir, path)
			return builder.AddFile(path, relPath)
		}); err != nil {
			return fmt.Errorf("failed to add files to package: %w", err)
		}

		if err := builder.Build(destPath); err != nil {
			return fmt.Errorf("failed to build package: %w", err)
		}

		if err := core.UpdateManifestWithChecksum(manifest, destPath); err != nil {
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
		builder, err := core.NewPackageBuilder()
		if err != nil {
			return fmt.Errorf("failed to create package builder: %w", err)
		}
		defer builder.Cleanup()

		if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(tempDir, path)
			return builder.AddFile(path, relPath)
		}); err != nil {
			return fmt.Errorf("failed to add files to package: %w", err)
		}

		if err := builder.Build(destPath); err != nil {
			return fmt.Errorf("failed to build package: %w", err)
		}

		if err := core.UpdateManifestWithChecksum(manifest, destPath); err != nil {
			fmt.Printf("Warning: failed to update manifest checksum: %v\n", err)
		}
		return nil
	}

	// Create .axon package
	builder, err := core.NewPackageBuilder()
	if err != nil {
		return fmt.Errorf("failed to create package builder: %w", err)
	}
	defer builder.Cleanup()

	if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(tempDir, path)
		return builder.AddFile(path, relPath)
	}); err != nil {
		return fmt.Errorf("failed to add files to package: %w", err)
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

// downloadFromBranch downloads model files from a GitHub branch
// This is a pure Go implementation that:
// 1. Fetches hubconf.py from GitHub
// 2. Parses it to extract model weight URLs
// 3. Downloads weights directly from those URLs
func (p *PyTorchHubAdapter) downloadFromBranch(ctx context.Context, githubRepo, modelName, destDir string, progress core.ProgressCallback) error {
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
func (p *PyTorchHubAdapter) downloadFile(ctx context.Context, url, destPath string, expectedSize int64, progress core.ProgressCallback) error {
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
	if progress != nil {
		// Use TeeReader to track progress
		reader := io.TeeReader(resp.Body, &progressWriter{
			writer:   outFile,
			progress: progress,
			total:    resp.ContentLength,
		})
		_, err = io.Copy(io.Discard, reader)
	} else {
		_, err = io.Copy(outFile, resp.Body)
	}
	return err
}

// progressWriter wraps a writer and reports progress
type progressWriter struct {
	writer   io.Writer
	progress core.ProgressCallback
	total    int64
	current  int64
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.current += int64(n)
	if pw.progress != nil {
		pw.progress(pw.current, pw.total)
	}
	return n, err
}

// TensorFlowHubAdapter implements RepositoryAdapter for TensorFlow Hub
// TensorFlow Hub models are hosted at https://tfhub.dev
// Models are organized by publisher (e.g., google, tensorflow) and can be SavedModel or TFLite format
