package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mlOS-foundation/axon/internal/manifest"
	"github.com/mlOS-foundation/axon/pkg/types"
	"github.com/mlOS-foundation/axon/pkg/utils"
)

// Client represents a registry HTTP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	mirrors    []string
}

// NewClient creates a new registry client
func NewClient(baseURL string, mirrors []string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		mirrors: mirrors,
	}
}

// Search searches for models in the registry
func (c *Client) Search(ctx context.Context, query string) ([]types.SearchResult, error) {
	url := fmt.Sprintf("%s/api/v1/search?q=%s", c.baseURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []types.SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetManifest retrieves a model manifest from the registry
func (c *Client) GetManifest(ctx context.Context, namespace, name, version string) (*types.Manifest, error) {
	url := fmt.Sprintf("%s/api/v1/models/%s/%s/%s/manifest.yaml", c.baseURL, namespace, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse manifest using the manifest package
	manifest, err := manifest.ParseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest, nil
}

// DownloadPackage downloads a model package
func (c *Client) DownloadPackage(ctx context.Context, manifest *types.Manifest, destPath string, progress ProgressCallback) error {
	urls := []string{manifest.Distribution.Package.URL}
	urls = append(urls, manifest.Distribution.Package.Mirrors...)

	var lastErr error
	for _, url := range urls {
		err := c.downloadFromURL(ctx, url, destPath, manifest.Distribution.Package.SHA256, progress)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to download from all sources: %w", lastErr)
}

func (c *Client) downloadFromURL(ctx context.Context, url, destPath, expectedSHA256 string, progress ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	reader := &progressReader{
		Reader:   resp.Body,
		Total:    resp.ContentLength,
		Callback: progress,
	}

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Verify checksum if provided
	if expectedSHA256 != "" {
		if err := verifyChecksum(destPath, expectedSHA256); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	return nil
}

// ProgressCallback is called during download progress
type ProgressCallback func(downloaded, total int64)

type progressReader struct {
	Reader     io.Reader
	Total      int64
	Downloaded int64
	Callback   ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)
	if pr.Callback != nil {
		pr.Callback(pr.Downloaded, pr.Total)
	}
	return n, err
}

func verifyChecksum(filePath, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return nil // No checksum to verify
	}

	return utils.VerifySHA256(filePath, expectedSHA256)
}
