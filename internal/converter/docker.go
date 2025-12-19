// Package converter provides Docker-based ONNX conversion functionality.
// This enables conversion without requiring Python on the host machine.
package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerConverter handles ONNX conversion using Docker containers.
type DockerConverter struct {
	imageName string
	cacheDir  string
}

// DefaultConverterImage is the default Docker image for ONNX conversion.
const DefaultConverterImage = "ghcr.io/mlos-foundation/axon-converter:latest"

// NewDockerConverter creates a new Docker-based converter.
func NewDockerConverter() *DockerConverter {
	cacheDir := os.Getenv("AXON_CACHE_DIR")
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".axon", "cache")
	}

	// Allow override via environment variable for testing/development
	imageName := os.Getenv("AXON_CONVERTER_IMAGE")
	if imageName == "" {
		imageName = DefaultConverterImage
	}

	return &DockerConverter{
		imageName: imageName,
		cacheDir:  cacheDir,
	}
}

// IsDockerAvailable checks if Docker is installed and running.
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// getDockerImageForRepository returns the Docker image name for a given repository namespace.
// This allows repository-specific images with optimized dependencies.
// Can be overridden via AXON_CONVERTER_IMAGE environment variable for testing.
func getDockerImageForRepository(namespace string) string {
	// Check for environment variable override first (useful for testing/development)
	if override := os.Getenv("AXON_CONVERTER_IMAGE"); override != "" {
		return override
	}

	// Repository-to-image mapping
	// Can be extended for repository-specific images
	repositoryImageMap := map[string]string{
		"hf":      DefaultConverterImage, // Hugging Face (use default for now)
		"pytorch": DefaultConverterImage, // PyTorch Hub
		"tfhub":   DefaultConverterImage, // TensorFlow Hub
		"ms":      DefaultConverterImage, // ModelScope
	}

	if image, ok := repositoryImageMap[namespace]; ok {
		return image
	}

	// Default to multi-framework image
	return DefaultConverterImage
}

// getConversionScript returns the conversion script name based on repository namespace and framework.
// Prioritize namespace over framework for better accuracy.
func getConversionScript(namespace, framework string) string {
	// First, check repository namespace (most reliable)
	namespaceLower := strings.ToLower(namespace)
	switch namespaceLower {
	case "hf", "huggingface":
		return "convert_huggingface.py"
	case "pytorch":
		return "convert_pytorch.py"
	case "tfhub", "tensorflow":
		return "convert_tensorflow.py"
	case "ms", "modelscope":
		return "convert_huggingface.py" // ModelScope uses transformers-like API
	}

	// Fallback: check framework
	frameworkLower := strings.ToLower(framework)
	switch {
	case frameworkLower == "huggingface" || frameworkLower == "transformers":
		return "convert_huggingface.py"
	case frameworkLower == "pytorch" || frameworkLower == "torch":
		return "convert_pytorch.py"
	case frameworkLower == "tensorflow" || frameworkLower == "tf":
		return "convert_tensorflow.py"
	default:
		return "convert_huggingface.py" // Default to Hugging Face (most common)
	}
}

// ConvertToONNXWithDocker converts a model to ONNX using Docker.
// This eliminates the need for Python on the host machine.
func ConvertToONNXWithDocker(ctx context.Context, modelPath, framework, namespace, modelID, outputPath string) (bool, error) {
	// Check Docker availability
	if !IsDockerAvailable() {
		return false, fmt.Errorf("Docker is not available - cannot perform conversion")
	}

	// Get appropriate Docker image for this repository
	imageName := getDockerImageForRepository(namespace)

	// Get conversion script based on namespace and framework
	scriptName := getConversionScript(namespace, framework)

	// Resolve absolute paths for volume mounting
	absCacheDir, err := filepath.Abs(filepath.Dir(modelPath))
	if err != nil {
		return false, fmt.Errorf("failed to resolve cache directory: %w", err)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Get relative path from cache directory for model path
	relModelPath, err := filepath.Rel(filepath.Dir(modelPath), modelPath)
	if err != nil {
		relModelPath = filepath.Base(modelPath)
	}

	// Get relative path for output
	relOutputPath, err := filepath.Rel(filepath.Dir(modelPath), absOutputPath)
	if err != nil {
		relOutputPath = filepath.Base(outputPath)
	}

	// Build Docker command
	// Volume mapping: host cache dir -> /axon/cache in container
	// Working directory: /axon/cache (so relative paths work)
	// IMPORTANT: Use absolute container paths to avoid Optimum/HuggingFace
	// misinterpreting relative paths like "latest" as model IDs
	containerModelPath := "/axon/cache/" + relModelPath
	containerOutputPath := "/axon/cache/" + relOutputPath
	dockerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/axon/cache", absCacheDir),
		"-w", "/axon/cache",
		imageName,
		fmt.Sprintf("/axon/scripts/%s", scriptName),
		containerModelPath,  // Absolute container path to model
		containerOutputPath, // Absolute container path for output
		modelID,             // Model ID for repository lookup (e.g., "microsoft/resnet-50")
	}

	fmt.Printf("🐳 Converting model using Docker (%s)...\n", imageName)
	fmt.Printf("   Image: %s\n", imageName)
	fmt.Printf("   Script: %s\n", scriptName)
	fmt.Printf("   Model: %s\n", modelPath)
	fmt.Printf("   Output: %s\n", outputPath)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if image needs to be pulled
		if strings.Contains(string(output), "Unable to find image") {
			fmt.Printf("📥 Pulling Docker image: %s (first time only, this may take a few minutes)...\n", imageName)
			pullCmd := exec.CommandContext(ctx, "docker", "pull", imageName)
			if pullErr := pullCmd.Run(); pullErr != nil {
				return false, fmt.Errorf("failed to pull Docker image: %w", pullErr)
			}
			// Retry conversion after pulling
			cmd = exec.CommandContext(ctx, "docker", dockerArgs...)
			output, err = cmd.CombinedOutput()
		}

		if err != nil {
			return false, fmt.Errorf("docker conversion failed: %w\nOutput: %s", err, string(output))
		}
	}

	// Verify output file was created and is valid
	fileInfo, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("conversion output file not created: %s\nConversion output: %s", outputPath, string(output))
	}

	// Check minimum file size (a valid ONNX file should be at least 1KB)
	if fileInfo.Size() < 1024 {
		return false, fmt.Errorf("conversion output file too small (%d bytes), likely corrupted: %s", fileInfo.Size(), outputPath)
	}

	// Validate ONNX file magic bytes (protobuf starts with valid field tags)
	if !ValidateONNXFile(outputPath) {
		return false, fmt.Errorf("conversion output file appears corrupted (invalid ONNX format): %s", outputPath)
	}

	fmt.Printf("✅ Model converted to ONNX using Docker: %s (%d bytes)\n", outputPath, fileInfo.Size())
	return true, nil
}

// ValidateONNXFile performs basic validation of an ONNX file.
// Checks protobuf structure without fully parsing the model.
func ValidateONNXFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	// Read first few bytes to check protobuf structure
	// ONNX files are Protocol Buffer format
	// A valid protobuf starts with field tags (wire type + field number)
	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 16 {
		return false
	}

	// Check for valid protobuf wire types in the first bytes
	// Wire types: 0=varint, 1=64-bit, 2=length-delimited, 5=32-bit
	// First byte should be a valid field tag (field_number << 3 | wire_type)
	// Common valid first bytes for ONNX: 0x08 (field 1, varint), 0x0a (field 1, length-delimited)
	firstByte := header[0]
	wireType := firstByte & 0x07
	if wireType > 5 {
		// Invalid wire type
		return false
	}

	// Note: Common ONNX patterns are 0x08, 0x0a, 0x10, 0x12
	// but other valid protobuf field tags are also acceptable
	// File passed basic checks
	return true
}

// EnsureDockerImage ensures the Docker image is available locally.
// If not, it attempts to pull it.
func EnsureDockerImage(ctx context.Context, namespace string) error {
	imageName := getDockerImageForRepository(namespace)

	// Check if image exists locally
	checkCmd := exec.CommandContext(ctx, "docker", "images", "-q", imageName)
	output, err := checkCmd.Output()
	if err == nil && len(output) > 0 {
		// Image exists locally
		return nil
	}

	// Image not found, try to pull
	fmt.Printf("📥 Pulling Docker image: %s (first time only)...\n", imageName)
	pullCmd := exec.CommandContext(ctx, "docker", "pull", imageName)
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull Docker image %s: %w\nMake sure Docker is running and you have internet access", imageName, err)
	}

	return nil
}
