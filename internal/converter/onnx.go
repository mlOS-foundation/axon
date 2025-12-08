// Package converter provides ONNX model conversion functionality for Axon.
// This enables conversion during 'axon install' so that MLOS Core receives
// pre-converted ONNX files in MPF packages.
//
// Architecture (Patent-Aligned):
// - Axon: Handles model conversion during install (pure Go when possible)
// - MLOS Core: Pure execution layer (uses pre-converted ONNX files)
// - This decouples MLOS Core from Python dependencies and repository logic
//
// Conversion Strategy (Pure Go First):
// 1. Check if repository provides pre-converted ONNX files (download directly)
// 2. If not available, attempt Python-based conversion (optional, graceful degradation)
// 3. If Python unavailable, skip conversion (user can convert manually or use framework-specific plugins)
//
// Multi-Encoder Support:
// Some models (CLIP, T5, BART) export multiple ONNX files. This package handles:
// - CLIP: text_model.onnx + vision_model.onnx
// - T5/BART: encoder_model.onnx + decoder_model.onnx + decoder_with_past_model.onnx
// An onnx_manifest.json file is created to describe multi-encoder architectures.
package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MultiEncoderManifest describes the structure of a multi-encoder model
type MultiEncoderManifest struct {
	Architecture string            `json:"architecture"` // "multi-encoder", "encoder-decoder", "multi-model"
	EncoderType  string            `json:"encoder_type"` // "clip", "seq2seq", "unknown"
	Task         string            `json:"task"`
	Components   map[string]string `json:"components"` // e.g., {"text_encoder": "text_model.onnx"}
	Files        []string          `json:"files"`
}

// ConversionResult contains information about a successful conversion
type ConversionResult struct {
	Success        bool
	IsMultiEncoder bool
	PrimaryFile    string   // model.onnx for single, empty for multi
	AllFiles       []string // All ONNX files created
	ManifestPath   string   // Path to onnx_manifest.json if multi-encoder
	Architecture   string   // "single", "multi-encoder", "encoder-decoder"
}

// DownloadPreConvertedONNX attempts to download a pre-converted ONNX file
// from the repository (e.g., Hugging Face often provides ONNX versions).
// This is the preferred method as it requires no Python dependencies.
//
// Parameters:
//   - ctx: Context for cancellation
//   - namespace: Repository namespace (e.g., "hf", "pytorch")
//   - modelID: Model identifier (e.g., "bert-base-uncased")
//   - outputPath: Where to save the ONNX file
//
// Returns:
//   - bool: true if ONNX file was successfully downloaded, false otherwise
//   - error: error if download failed (not found is not an error)
func DownloadPreConvertedONNX(ctx context.Context, namespace, modelID, outputPath string) (bool, error) {
	// Only Hugging Face currently provides ONNX files directly
	if namespace != "hf" {
		return false, nil
	}

	// Hugging Face ONNX files are typically at:
	// https://huggingface.co/{model_id}/resolve/main/model.onnx
	// or https://huggingface.co/{model_id}/resolve/main/onnx/model.onnx
	baseURL := "https://huggingface.co"
	urls := []string{
		fmt.Sprintf("%s/%s/resolve/main/model.onnx", baseURL, modelID),
		fmt.Sprintf("%s/%s/resolve/main/onnx/model.onnx", baseURL, modelID),
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				_ = resp.Body.Close()
			}
			continue
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		// Create output directory if needed
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return false, fmt.Errorf("failed to create output directory: %w", err)
		}

		// Download file
		outFile, err := os.Create(outputPath)
		if err != nil {
			return false, fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			_ = outFile.Close()
		}()

		if _, err := io.Copy(outFile, resp.Body); err != nil {
			_ = os.Remove(outputPath)
			return false, fmt.Errorf("failed to download ONNX file: %w", err)
		}

		fmt.Printf("✅ Downloaded pre-converted ONNX from repository: %s\n", url)
		return true, nil
	}

	return false, nil // Not found is not an error
}

// ConvertToONNX converts a model to ONNX format during Axon install.
// The converted ONNX file will be included in the MPF package.
//
// Strategy:
// 1. First try to download pre-converted ONNX from repository (pure Go, no deps)
// 2. If not available, attempt Python-based conversion (optional, graceful degradation)
//
// Parameters:
//   - ctx: Context for cancellation
//   - modelPath: Path to the model directory (from Axon cache)
//   - framework: Framework name (e.g., "PyTorch", "HuggingFace", "TensorFlow")
//   - namespace: Repository namespace (e.g., "hf", "pytorch")
//   - modelID: Model identifier for repository lookup
//   - outputPath: Where to save the converted ONNX file (typically modelPath/model.onnx)
//
// Returns:
//   - bool: true if ONNX file was created, false if conversion skipped (Python unavailable)
//   - error: nil on success, error on failure
func ConvertToONNX(ctx context.Context, modelPath, framework, namespace, modelID, outputPath string) (bool, error) {
	// Step 1: Try to download pre-converted ONNX from repository (pure Go, no Python needed)
	if namespace != "" && modelID != "" {
		downloaded, err := DownloadPreConvertedONNX(ctx, namespace, modelID, outputPath)
		if err != nil {
			return false, fmt.Errorf("failed to download pre-converted ONNX: %w", err)
		}
		if downloaded {
			return true, nil // Success - pure Go, no Python needed!
		}
	}

	// Step 2: Fall back to conversion (Docker first, then local Python)
	if modelPath == "" || framework == "" || outputPath == "" {
		return false, fmt.Errorf("modelPath, framework, and outputPath are required")
	}

	// Try Docker-based conversion first (no Python needed on host)
	if IsDockerAvailable() {
		// Ensure Docker image is available
		if err := EnsureDockerImage(ctx, namespace); err != nil {
			fmt.Printf("⚠️  Docker image not available: %v\n", err)
			fmt.Printf("   💡 Falling back to local Python (if available)\n")
		} else {
			// Try Docker conversion
			converted, err := ConvertToONNXWithDocker(ctx, modelPath, framework, namespace, modelID, outputPath)
			if err == nil && converted {
				return true, nil // Success with Docker!
			}
			// If Docker conversion fails, fall through to local Python
			if err != nil {
				fmt.Printf("⚠️  Docker conversion failed: %v\n", err)
				fmt.Printf("   💡 Falling back to local Python (if available)\n")
			}
		}
	}

	// Step 3: Fall back to local Python-based conversion (if available)
	// Check if Python 3 is available (optional dependency)
	if _, err := exec.LookPath("python3"); err != nil {
		// Python not available - graceful degradation
		fmt.Printf("⚠️  Python3 not found - skipping ONNX conversion\n")
		fmt.Printf("   💡 Model will work with framework-specific plugins\n")
		if !IsDockerAvailable() {
			fmt.Printf("   💡 To enable ONNX conversion, either:\n")
			fmt.Printf("      - Install Docker: https://docs.docker.com/get-docker/\n")
			fmt.Printf("      - Or install Python 3 and: pip install transformers torch\n")
		}
		return false, nil // Not an error - just skipped
	}

	// Normalize framework name for comparison
	frameworkLower := strings.ToLower(framework)

	// Build Python conversion command based on framework
	var pythonCmd string

	switch {
	case frameworkLower == "huggingface" || frameworkLower == "transformers" ||
		frameworkLower == "pytorch" || frameworkLower == "torch":
		// Hugging Face / PyTorch conversion
		if frameworkLower == "huggingface" || frameworkLower == "transformers" {
			// Extract model name from path for Hugging Face models
			// Path format: .../hf/bert-base-uncased/latest
			modelName := extractModelNameFromPath(modelPath)
			if modelName == "" {
				modelName = "bert-base-uncased" // fallback
			}

			pythonCmd = fmt.Sprintf(`python3 -c "
import sys
import os
try:
    from transformers import AutoModel, AutoTokenizer
    import torch
    model_path = '%s'
    output_path = '%s'
    hf_model_id = '%s'
    os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
    print('Loading model:', hf_model_id)
    # Try loading from local path first, then from Hugging Face Hub
    try:
        if os.path.isdir(model_path):
            model = AutoModel.from_pretrained(model_path, local_files_only=True)
            tokenizer = AutoTokenizer.from_pretrained(model_path, local_files_only=True)
            print('Loaded from local path')
        else:
            raise FileNotFoundError('Not a directory')
    except Exception as e:
        print('Local load failed, trying Hugging Face Hub:', str(e))
        model = AutoModel.from_pretrained(hf_model_id)
        tokenizer = AutoTokenizer.from_pretrained(hf_model_id)
        print('Loaded from Hugging Face Hub')
    model.eval()
    # Get model config for input shape
    config = model.config
    seq_len = min(128, getattr(config, 'max_position_embeddings', 128))
    vocab_size = getattr(config, 'vocab_size', 30522)
    # Create dummy input
    dummy_input = torch.randint(0, vocab_size, (1, seq_len))
    # Export to ONNX
    torch.onnx.export(model, dummy_input, output_path,
        input_names=['input_ids'],
        output_names=['output'],
        dynamic_axes={'input_ids': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
        opset_version=12,
        do_constant_folding=True)
    print('SUCCESS')
except ImportError as e:
    print('ERROR: Missing dependency:', str(e))
    print('Install with: pip install transformers torch')
    sys.exit(1)
except Exception as e:
    print('ERROR:', str(e))
    import traceback
    traceback.print_exc()
    sys.exit(1)
"`, modelPath, outputPath, modelName)
		} else {
			// PyTorch conversion
			pythonCmd = fmt.Sprintf(`python3 -c "
import sys
import torch
import os
try:
    model_path = '%s'
    output_path = '%s'
    os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
    # Try to load as PyTorch model
    if os.path.isdir(model_path):
        print('ERROR: Directory-based PyTorch models need specific loading code')
        sys.exit(1)
    else:
        model = torch.load(model_path, map_location='cpu')
        if isinstance(model, torch.nn.Module):
            model.eval()
            dummy_input = torch.randn(1, 3, 224, 224)
            torch.onnx.export(model, dummy_input, output_path, opset_version=12)
            print('SUCCESS')
        else:
            print('ERROR: Model file is not a PyTorch Module')
            sys.exit(1)
except Exception as e:
    print('ERROR:', str(e))
    import traceback
    traceback.print_exc()
    sys.exit(1)
"`, modelPath, outputPath)
		}

	case frameworkLower == "tensorflow" || frameworkLower == "tf":
		// TensorFlow conversion
		pythonCmd = fmt.Sprintf(`python3 -c "
import sys
try:
    import tf2onnx
    import tensorflow as tf
    import os
    model_path = '%s'
    output_path = '%s'
    os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
    print('ERROR: TensorFlow conversion not fully implemented')
    print('Please use tf2onnx.convert or implement full conversion')
    sys.exit(1)
except ImportError as e:
    print('ERROR: Missing dependency:', str(e))
    print('Install with: pip install tf2onnx tensorflow')
    sys.exit(1)
except Exception as e:
    print('ERROR:', str(e))
    sys.exit(1)
"`, modelPath, outputPath)

	default:
		return false, fmt.Errorf("unsupported framework for ONNX conversion: %s", framework)
	}

	// Execute Python conversion
	fmt.Printf("🔄 Converting model to ONNX format (framework: %s)...\n", framework)
	fmt.Printf("   Source: %s\n", modelPath)
	fmt.Printf("   Target: %s\n", outputPath)

	cmd := exec.Command("sh", "-c", pythonCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("conversion failed: %w\nOutput: %s", err, string(output))
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return false, fmt.Errorf("conversion output file not created: %s\nConversion output: %s", outputPath, string(output))
	}

	fmt.Printf("✅ Model converted to ONNX: %s\n", outputPath)
	return true, nil
}

// extractModelNameFromPath extracts model name from Axon cache path
// Example: /path/to/hf/bert-base-uncased/latest -> bert-base-uncased
func extractModelNameFromPath(path string) string {
	// Remove trailing slashes
	path = strings.TrimSuffix(path, "/")

	// Split by path separator
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) < 2 {
		return ""
	}

	// Get second-to-last part (model name)
	// Path structure: .../namespace/model-name/version
	// We want model-name
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}

	return ""
}

// CanConvert checks if a framework supports ONNX conversion
func CanConvert(framework string) bool {
	if framework == "" {
		return false
	}

	frameworkLower := strings.ToLower(framework)

	// Frameworks that can be converted to ONNX
	supported := []string{
		"pytorch", "torch",
		"tensorflow", "tf",
		"huggingface", "transformers",
		"onnx", // Already ONNX
	}

	for _, fw := range supported {
		if frameworkLower == fw {
			return true
		}
	}

	return false
}

// IsExecutionReady checks if a format is already execution-ready (no conversion needed)
// These formats can be used directly by MLOS Core without ONNX conversion
func IsExecutionReady(format string) bool {
	if format == "" {
		return false
	}

	formatLower := strings.ToLower(format)

	// Formats that Core can execute directly (has runtime plugins)
	// Note: Only formats with actual Core runtime plugins should be here
	ready := []string{
		"gguf", // Native LLM format (llama.cpp) - Core GGUF plugin
		"onnx", // ONNX format - Core ONNX Runtime plugin
		// "safetensors" - Not yet supported by Core (no plugin)
		// "pytorch" - Needs ONNX conversion
		// "tensorflow" - Needs ONNX conversion
	}

	for _, f := range ready {
		if formatLower == f {
			return true
		}
	}

	return false
}

// FindONNXFiles finds all ONNX files in a directory (including onnx/ subdirectory).
// Optimum creates multi-encoder model files (T5, CLIP, etc.) in an onnx/ subdirectory.
func FindONNXFiles(dir string) ([]string, error) {
	var onnxFiles []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Search root directory
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".onnx") {
			onnxFiles = append(onnxFiles, filepath.Join(dir, entry.Name()))
		}
	}

	// Also check onnx/ subdirectory (Optimum creates files here for multi-encoder models)
	onnxSubdir := filepath.Join(dir, "onnx")
	if subdirEntries, err := os.ReadDir(onnxSubdir); err == nil {
		for _, entry := range subdirEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".onnx") {
				onnxFiles = append(onnxFiles, filepath.Join(onnxSubdir, entry.Name()))
			}
		}
	}

	return onnxFiles, nil
}

// ReadMultiEncoderManifest reads the onnx_manifest.json file if present
func ReadMultiEncoderManifest(dir string) (*MultiEncoderManifest, error) {
	manifestPath := filepath.Join(dir, "onnx_manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var manifest MultiEncoderManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// CheckForMultiEncoderManifest checks if a directory contains a multi-encoder manifest
// and returns the manifest if found
func CheckForMultiEncoderManifest(dir string) (*MultiEncoderManifest, bool) {
	manifest, err := ReadMultiEncoderManifest(dir)
	if err != nil {
		return nil, false
	}
	return manifest, true
}

// CheckConversionResult checks what was produced after conversion
// and returns a ConversionResult with details about all created files
// This function is called after conversion to detect single vs multi-encoder models
func CheckConversionResult(modelDir, expectedOutput string) *ConversionResult {
	result := &ConversionResult{
		Success:      false,
		Architecture: "single",
	}

	// First check for multi-encoder manifest (created by Python converters)
	manifest, err := ReadMultiEncoderManifest(modelDir)
	if err == nil && manifest != nil {
		// Multi-encoder model detected via manifest
		result.Success = true
		result.IsMultiEncoder = true
		result.Architecture = manifest.Architecture
		result.ManifestPath = filepath.Join(modelDir, "onnx_manifest.json")

		// Get all ONNX files listed in manifest
		for _, fileName := range manifest.Files {
			fullPath := filepath.Join(modelDir, fileName)
			if _, err := os.Stat(fullPath); err == nil {
				result.AllFiles = append(result.AllFiles, fullPath)
			}
		}

		// If no files found in manifest, try to find them
		if len(result.AllFiles) == 0 {
			onnxFiles, _ := FindONNXFiles(modelDir)
			result.AllFiles = onnxFiles
		}

		return result
	}

	// Check for single model.onnx (standard single-model output)
	if _, err := os.Stat(expectedOutput); err == nil {
		result.Success = true
		result.PrimaryFile = expectedOutput
		result.AllFiles = []string{expectedOutput}
		return result
	}

	// Fallback: Check for any ONNX files (might be multi-encoder without manifest)
	onnxFiles, err := FindONNXFiles(modelDir)
	if err == nil && len(onnxFiles) > 0 {
		result.Success = true
		if len(onnxFiles) == 1 {
			result.PrimaryFile = onnxFiles[0]
			result.AllFiles = []string{onnxFiles[0]}
		} else {
			// Multiple files but no manifest - treat as multi-model
			result.IsMultiEncoder = true
			result.Architecture = "multi-model"
			result.AllFiles = onnxFiles

			// Try to auto-detect architecture from file names
			fileNames := make([]string, len(onnxFiles))
			for i, f := range onnxFiles {
				fileNames[i] = filepath.Base(f)
			}

			// Check for CLIP pattern
			hasText := false
			hasVision := false
			for _, name := range fileNames {
				if name == "text_model.onnx" {
					hasText = true
				}
				if name == "vision_model.onnx" {
					hasVision = true
				}
			}
			if hasText && hasVision {
				result.Architecture = "multi-encoder"
				// Create manifest for future use
				manifest := &MultiEncoderManifest{
					Architecture: "multi-encoder",
					EncoderType:  "clip",
					Task:         "zero-shot-image-classification",
					Components: map[string]string{
						"text_encoder":   "text_model.onnx",
						"vision_encoder": "vision_model.onnx",
					},
					Files: fileNames,
				}
				// Save manifest
				manifestData, _ := json.MarshalIndent(manifest, "", "  ")
				manifestPath := filepath.Join(modelDir, "onnx_manifest.json")
				_ = os.WriteFile(manifestPath, manifestData, 0644)
				result.ManifestPath = manifestPath
			}
		}
	}

	return result
}

// ConvertToONNXWithResult converts a model and returns detailed results
// including information about multi-encoder models
func ConvertToONNXWithResult(ctx context.Context, modelPath, framework, namespace, modelID, outputPath string) (*ConversionResult, error) {
	// Run the standard conversion
	converted, err := ConvertToONNX(ctx, modelPath, framework, namespace, modelID, outputPath)
	if err != nil {
		return &ConversionResult{Success: false}, err
	}

	if !converted {
		return &ConversionResult{Success: false}, nil
	}

	// Check what was actually created
	modelDir := filepath.Dir(outputPath)
	return CheckConversionResult(modelDir, outputPath), nil
}
