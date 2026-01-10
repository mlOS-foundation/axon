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
	"strings"
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
	allFiles, correctedModelID, err := h.getModelFiles(ctx, hfModelID)
	if err != nil {
		// Fallback to common files if API fails
		// Include both NLP model files and vision/detection model files
		allFiles = []string{
			// NLP/Transformer model files
			"config.json",
			"pytorch_model.bin",
			"model.safetensors",
			"tokenizer.json",
			"tokenizer_config.json",
			"vocab.txt",
			"vocab.json",
			// Vision/Detection model files (YOLO, etc.)
			"model.pt",
			"weights.pt",
			"best.pt",
			// ONNX formats
			"model.onnx",
		}

		// Also try the model name itself as a .pt file (e.g., "yolov8n" -> "yolov8n.pt")
		// This handles cases like ultralytics/YOLOv8 where individual models are files
		modelName := manifest.Metadata.Name
		if !strings.HasSuffix(strings.ToLower(modelName), ".pt") {
			allFiles = append(allFiles, modelName+".pt")
		}
	} else {
		// Use the corrected model ID (in case a variation was found)
		hfModelID = correctedModelID
	}

	// Detect best format and select appropriate files
	// Priority: GGUF > ONNX > SafeTensors > PyTorch (reduces download size and skips conversion)
	formatType, modelFiles := h.detectModelFormat(allFiles)
	if formatType != "unknown" && formatType != "pytorch" {
		fmt.Printf("✓ Detected %s format, selecting optimized file set\n", strings.ToUpper(formatType))
		// Update manifest with detected format
		manifest.Spec.Format.Type = formatType
		manifest.Spec.Format.ExecutionFormat = formatType
	}

	// Ensure tokenizer files are included for non-GGUF formats
	// (GGUF models have tokenizer embedded)
	if formatType != "gguf" {
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
	}

	// Download files from Hugging Face
	httpClient := &http.Client{Timeout: 10 * time.Minute}
	downloadedFiles := []string{}
	failedFiles := []string{}

	for _, file := range modelFiles {
		url := fmt.Sprintf("%s/%s/resolve/main/%s", h.baseURL, hfModelID, file)

		// Create temp file for download
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("axon-hf-%s-%d", filepath.Base(file), time.Now().UnixNano()))

		// Add auth header if token is provided
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			failedFiles = append(failedFiles, file)
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
			failedFiles = append(failedFiles, file)
			continue // Skip missing files
		}

		// Download file
		if err := core.DownloadFile(ctx, httpClient, url, tempFile, progress); err != nil {
			_ = resp.Body.Close()
			failedFiles = append(failedFiles, file)
			continue
		}
		_ = resp.Body.Close()

		// Add to package
		if err := builder.AddFile(tempFile, file); err != nil {
			_ = os.Remove(tempFile)
			failedFiles = append(failedFiles, file)
			continue
		}

		downloadedFiles = append(downloadedFiles, file)
		_ = os.Remove(tempFile) // Clean up temp file
	}

	if len(downloadedFiles) == 0 {
		return fmt.Errorf("no files downloaded from Hugging Face for %s (tried: %s). "+
			"Check that the model exists at https://huggingface.co/%s and verify the correct namespace/name. "+
			"Note: HuggingFace URLs are case-sensitive",
			hfModelID, strings.Join(failedFiles, ", "), hfModelID)
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
// It tries the original model ID first, then case variations if that fails.
// Returns the files and the corrected model ID (in case a variation worked).
func (h *HuggingFaceAdapter) getModelFiles(ctx context.Context, modelID string) ([]string, string, error) {
	// Try original model ID first
	files, err := h.fetchModelFiles(ctx, modelID)
	if err == nil {
		return files, modelID, nil
	}

	// Try case variations (HuggingFace URLs are case-sensitive)
	// Common patterns: namespace is often title-case (e.g., "Ultralytics")
	caseVariations := h.generateCaseVariations(modelID)
	for _, variation := range caseVariations {
		if variation == modelID {
			continue // Already tried
		}
		files, err := h.fetchModelFiles(ctx, variation)
		if err == nil {
			fmt.Printf("✓ Found model at %s (tried: %s)\n", variation, modelID)
			return files, variation, nil
		}
	}

	return nil, "", fmt.Errorf("model not found: %s (also tried case variations)", modelID)
}

// generateCaseVariations generates common case variations for a model ID.
func (h *HuggingFaceAdapter) generateCaseVariations(modelID string) []string {
	parts := strings.SplitN(modelID, "/", 2)
	if len(parts) != 2 {
		return []string{modelID}
	}

	namespace, name := parts[0], parts[1]
	variations := []string{}

	// Generate namespace variations
	namespaceVariations := []string{
		toTitleCase(namespace),
		strings.ToUpper(namespace),
		strings.ToLower(namespace),
	}

	// Generate model name variations
	// Common patterns: YOLOv8, Llama-2, GPT-J, etc.
	nameVariations := []string{
		name,
		toTitleCase(name),
		strings.ToUpper(name),
	}

	// For YOLO models, try extracting the base model name (yolov8n -> YOLOv8)
	if strings.HasPrefix(strings.ToLower(name), "yolov") {
		// Extract base name (yolov8n -> yolov8, yolov8s -> yolov8)
		baseName := extractYOLOBaseName(name)
		if baseName != "" && baseName != name {
			nameVariations = append(nameVariations, baseName)
		}
	}

	// Generate all combinations
	seen := make(map[string]bool)
	for _, ns := range namespaceVariations {
		for _, n := range nameVariations {
			v := fmt.Sprintf("%s/%s", ns, n)
			if !seen[v] && v != modelID {
				seen[v] = true
				variations = append(variations, v)
			}
		}
	}

	return variations
}

// toTitleCase converts a string to title case (first letter uppercase).
func toTitleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// extractYOLOBaseName extracts the base YOLO model name (e.g., "yolov8n" -> "YOLOv8").
func extractYOLOBaseName(name string) string {
	lower := strings.ToLower(name)

	// Match patterns like yolov5n, yolov8n, yolov8s, yolov8m, yolov8l, yolov8x
	// and convert to YOLOv5, YOLOv8, etc.
	for _, version := range []string{"5", "6", "7", "8", "9", "10", "11"} {
		prefix := "yolov" + version
		if strings.HasPrefix(lower, prefix) {
			// Check if there's a variant suffix (n, s, m, l, x)
			remaining := lower[len(prefix):]
			if len(remaining) > 0 && strings.ContainsAny(remaining[:1], "nsmlx") {
				// Return the base model name with proper casing
				return "YOLOv" + version
			}
		}
	}
	return ""
}

// fetchModelFiles makes the actual API call to fetch model files.
func (h *HuggingFaceAdapter) fetchModelFiles(ctx context.Context, modelID string) ([]string, error) {
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

// detectModelFormat analyzes file list and returns the best format to use.
// Priority for Core-compatible formats:
//  1. GGUF - Native LLM format (llama.cpp plugin)
//  2. ONNX - Direct execution (ONNX Runtime plugin)
//  3. SafeTensors/PyTorch - Need ONNX conversion
//
// Returns the format type and list of files to download
func (h *HuggingFaceAdapter) detectModelFormat(files []string) (string, []string) {
	var ggufFiles, onnxFiles, safetensorFiles, pytorchFiles, configFiles []string

	for _, file := range files {
		lower := strings.ToLower(file)
		switch {
		case strings.HasSuffix(lower, ".gguf"):
			ggufFiles = append(ggufFiles, file)
		case strings.HasSuffix(lower, ".onnx"):
			onnxFiles = append(onnxFiles, file)
		case strings.HasSuffix(lower, ".safetensors"):
			safetensorFiles = append(safetensorFiles, file)
		case strings.HasSuffix(lower, ".bin") || strings.HasSuffix(lower, ".pt") || strings.HasSuffix(lower, ".pth"):
			pytorchFiles = append(pytorchFiles, file)
		case strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".txt"):
			configFiles = append(configFiles, file)
		}
	}

	// Priority 1: GGUF - Core has native llama.cpp plugin
	// Best for LLMs, no conversion needed
	if len(ggufFiles) > 0 {
		selected := selectBestGGUF(ggufFiles)
		return "gguf", append([]string{selected}, configFiles...)
	}

	// Priority 2: ONNX - Core has ONNX Runtime plugin
	// Already execution-ready, no conversion needed
	if len(onnxFiles) > 0 {
		return "onnx", append(onnxFiles, configFiles...)
	}

	// Priority 3: SafeTensors - Preferred over PyTorch .bin files
	// Needs ONNX conversion but safer/faster to load than pickle
	if len(safetensorFiles) > 0 {
		return "safetensors", append(safetensorFiles, configFiles...)
	}

	// Priority 4: PyTorch - Traditional format
	// Needs ONNX conversion
	if len(pytorchFiles) > 0 {
		return "pytorch", append(pytorchFiles, configFiles...)
	}

	// Fallback: return all files (will need manual handling)
	return "unknown", files
}

// selectBestGGUF picks the best GGUF file from a list.
// Prefers Q4_K_M (good balance of quality/size), then Q4_K_S, then any Q4, then first available.
func selectBestGGUF(files []string) string {
	preferences := []string{"q4_k_m", "q4_k_s", "q4_0", "q5_k_m", "q8_0"}

	for _, pref := range preferences {
		for _, file := range files {
			if strings.Contains(strings.ToLower(file), pref) {
				return file
			}
		}
	}

	// Return first GGUF file if no preference matched
	return files[0]
}
