// Package core provides helper utilities for adapter implementations.
package core

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// HTTPClient provides a configurable HTTP client for adapters.
type HTTPClient struct {
	client    *http.Client
	baseURL   string
	token     string
	userAgent string
}

// NewHTTPClient creates a new HTTP client with default settings.
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:   baseURL,
		userAgent: "Axon-CLI/1.0",
	}
}

// SetToken sets the authentication token.
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// SetUserAgent sets the user agent string.
func (c *HTTPClient) SetUserAgent(ua string) {
	c.userAgent = ua
}

// Do performs an HTTP request with authentication headers.
func (c *HTTPClient) Do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return c.client.Do(req)
}

// Get performs a GET request.
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	return c.Do(ctx, "GET", url, nil)
}

// PackageBuilder helps build .axon package files.
// This provides common functionality for creating tar.gz packages with manifests.
type PackageBuilder struct {
	tempDir string
	files   []string
}

// NewPackageBuilder creates a new package builder.
func NewPackageBuilder() (*PackageBuilder, error) {
	tempDir, err := os.MkdirTemp("", "axon-package-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &PackageBuilder{
		tempDir: tempDir,
		files:   []string{},
	}, nil
}

// AddFile adds a file to the package.
func (pb *PackageBuilder) AddFile(srcPath, destPath string) error {
	destFullPath := filepath.Join(pb.tempDir, destPath)
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destFullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	pb.files = append(pb.files, destPath)
	return nil
}

// AddFileFromReader adds a file to the package from an io.Reader.
func (pb *PackageBuilder) AddFileFromReader(reader io.Reader, destPath string) error {
	destFullPath := filepath.Join(pb.tempDir, destPath)
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(destFullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	pb.files = append(pb.files, destPath)
	return nil
}

// Build creates the final .axon package file.
func (pb *PackageBuilder) Build(destPath string) error {
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
	return filepath.Walk(pb.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(pb.tempDir, path)
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
		defer srcFile.Close()

		_, err = io.Copy(tarWriter, srcFile)
		return err
	})
}

// Cleanup removes the temporary directory.
func (pb *PackageBuilder) Cleanup() error {
	return os.RemoveAll(pb.tempDir)
}

// ComputeChecksum computes the SHA256 checksum of a file.
func ComputeChecksum(filePath string) (string, int64, error) {
	hasher := sha256.New()
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	size, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, size, nil
}

// UpdateManifestWithChecksum updates a manifest with the computed checksum and size.
func UpdateManifestWithChecksum(manifest *types.Manifest, packagePath string) error {
	checksum, size, err := ComputeChecksum(packagePath)
	if err != nil {
		return err
	}

	manifest.Distribution.Package.SHA256 = checksum
	manifest.Distribution.Package.Size = size
	return nil
}

// DownloadFile downloads a file from a URL to a destination path.
func DownloadFile(ctx context.Context, client *http.Client, url, destPath string, progress ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	total := resp.ContentLength
	var current int64

	if progress != nil && total > 0 {
		// Use TeeReader to track progress
		reader := io.TeeReader(resp.Body, &progressWriter{
			writer:   file,
			progress: progress,
			total:    total,
			current:  &current,
		})
		_, err = io.Copy(file, reader)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if progress != nil {
		progress(current, total)
	}

	return nil
}

// progressWriter wraps a writer and reports progress.
type progressWriter struct {
	writer   io.Writer
	progress ProgressCallback
	total    int64
	current  *int64
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	*pw.current += int64(n)
	if pw.progress != nil {
		pw.progress(*pw.current, pw.total)
	}
	return n, err
}
