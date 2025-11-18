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

// NewDockerConverter creates a new Docker-based converter.
func NewDockerConverter() *DockerConverter {
	cacheDir := os.Getenv("AXON_CACHE_DIR")
	if cacheDir == "" {
		homeDir, _ := os.UserHomeDir()
		cacheDir = filepath.Join(homeDir, ".axon", "cache")
	}

	return &DockerConverter{
		imageName: "ghcr.io/mlOS-foundation/axon-converter:latest", // Default multi-framework image
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
func getDockerImageForRepository(namespace string) string {
	// Repository-to-image mapping
	// Can be extended for repository-specific images
	repositoryImageMap := map[string]string{
		"hf":      "ghcr.io/mlOS-foundation/axon-converter:latest", // Hugging Face (use default for now)
		"pytorch": "ghcr.io/mlOS-foundation/axon-converter:latest", // PyTorch Hub
		"tfhub":   "ghcr.io/mlOS-foundation/axon-converter:latest", // TensorFlow Hub
		"ms":      "ghcr.io/mlOS-foundation/axon-converter:latest", // ModelScope
	}

	if image, ok := repositoryImageMap[namespace]; ok {
		return image
	}

	// Default to multi-framework image
	return "ghcr.io/mlOS-foundation/axon-converter:latest"
}

// getConversionScript returns the conversion script name based on framework.
func getConversionScript(framework string) string {
	frameworkLower := strings.ToLower(framework)
	switch {
	case frameworkLower == "huggingface" || frameworkLower == "transformers":
		return "convert_huggingface.py"
	case frameworkLower == "pytorch" || frameworkLower == "torch":
		return "convert_pytorch.py"
	case frameworkLower == "tensorflow" || frameworkLower == "tf":
		return "convert_tensorflow.py"
	default:
		return "convert_huggingface.py" // Default
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

	// Get conversion script based on framework
	scriptName := getConversionScript(framework)

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
	dockerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/axon/cache", absCacheDir),
		"-w", "/axon/cache",
		imageName,
		fmt.Sprintf("/axon/scripts/%s", scriptName),
		relModelPath,  // Model path (relative to cache)
		relOutputPath, // Output path (relative to cache)
		modelID,       // Model ID for repository lookup
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

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return false, fmt.Errorf("conversion output file not created: %s\nConversion output: %s", outputPath, string(output))
	}

	fmt.Printf("✅ Model converted to ONNX using Docker: %s\n", outputPath)
	return true, nil
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
