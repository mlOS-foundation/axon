// Package core provides validation utilities for adapters.
package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ModelValidator provides generic model existence validation for adapters.
// This is a shared utility that all adapters can use to validate model URLs.
type ModelValidator struct {
	httpClient *http.Client
}

// NewModelValidator creates a new model validator with default settings.
func NewModelValidator() *ModelValidator {
	return &ModelValidator{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateModelExists checks if a model exists at the given URL.
// Returns true if model exists, false if not found, error for validation failures.
//
// This is a generic helper that can be used by all adapters.
// Uses GET request with redirect following to handle repositories that don't support HEAD.
func (mv *ModelValidator) ValidateModelExists(ctx context.Context, modelURL string) (bool, error) {
	// Use GET request (some repositories like TensorFlow Hub don't support HEAD properly)
	// Limit response size to avoid downloading large files
	req, err := http.NewRequestWithContext(ctx, "GET", modelURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Range header to only request first few bytes (validation only)
	req.Header.Set("Range", "bytes=0-1023")
	req.Header.Set("User-Agent", "Axon-CLI/1.0")

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
