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
package converter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

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

	// Step 2: Fall back to Python-based conversion (optional, graceful degradation)
	if modelPath == "" || framework == "" || outputPath == "" {
		return false, fmt.Errorf("modelPath, framework, and outputPath are required")
	}

	// Check if Python 3 is available (optional dependency)
	if _, err := exec.LookPath("python3"); err != nil {
		// Python not available - graceful degradation
		fmt.Printf("⚠️  Python3 not found - skipping ONNX conversion\n")
		fmt.Printf("   💡 Model will work with framework-specific plugins\n")
		fmt.Printf("   💡 To enable ONNX conversion, install Python 3 and: pip install transformers torch\n")
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
