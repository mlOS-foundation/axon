package manifest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// Validate validates a manifest against the schema
func Validate(m *types.Manifest) error {
	if err := validateAPIVersion(m); err != nil {
		return fmt.Errorf("apiVersion: %w", err)
	}

	if err := validateMetadata(m.Metadata); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}

	if err := validateSpec(m.Spec); err != nil {
		return fmt.Errorf("spec: %w", err)
	}

	if err := validateDistribution(m.Distribution); err != nil {
		return fmt.Errorf("distribution: %w", err)
	}

	return nil
}

func validateAPIVersion(m *types.Manifest) error {
	if m.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}

	if m.APIVersion != "axon.mlos.io/v1" {
		return fmt.Errorf("unsupported apiVersion: %s (expected: axon.mlos.io/v1)", m.APIVersion)
	}

	if m.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	if m.Kind != "Model" {
		return fmt.Errorf("unsupported kind: %s (expected: Model)", m.Kind)
	}

	return nil
}

func validateMetadata(meta types.Metadata) error {
	if meta.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Name should be lowercase alphanumeric + hyphens
	nameRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !nameRegex.MatchString(meta.Name) {
		return fmt.Errorf("name must be lowercase alphanumeric with hyphens only")
	}

	if meta.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	// Namespace should be lowercase alphanumeric + hyphens
	if !nameRegex.MatchString(meta.Namespace) {
		return fmt.Errorf("namespace must be lowercase alphanumeric with hyphens only")
	}

	if meta.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Validate semantic version
	if _, err := semver.NewVersion(meta.Version); err != nil {
		return fmt.Errorf("version must be valid semver: %w", err)
	}

	if meta.Description == "" {
		return fmt.Errorf("description is required")
	}

	if meta.License == "" {
		return fmt.Errorf("license is required")
	}

	return nil
}

func validateSpec(spec types.Spec) error {
	if spec.Framework.Name == "" {
		return fmt.Errorf("framework.name is required")
	}

	validFrameworks := map[string]bool{
		"pytorch":      true,
		"tensorflow":   true,
		"onnx":         true,
		"jax":          true,
		"scikit-learn": true,
		"keras":        true,
		"tflite":       true,
	}

	if !validFrameworks[strings.ToLower(spec.Framework.Name)] {
		return fmt.Errorf("unsupported framework: %s", spec.Framework.Name)
	}

	if spec.Framework.Version == "" {
		return fmt.Errorf("framework.version is required")
	}

	if spec.Format.Type == "" {
		return fmt.Errorf("format.type is required")
	}

	if len(spec.Format.Files) == 0 {
		return fmt.Errorf("format.files cannot be empty")
	}

	// Validate file entries
	for i, file := range spec.Format.Files {
		if file.Path == "" {
			return fmt.Errorf("format.files[%d].path is required", i)
		}

		// Check for path traversal
		if strings.Contains(file.Path, "..") {
			return fmt.Errorf("format.files[%d].path contains path traversal: %s", i, file.Path)
		}

		// Check for absolute paths
		if strings.HasPrefix(file.Path, "/") {
			return fmt.Errorf("format.files[%d].path must be relative: %s", i, file.Path)
		}

		if file.Size <= 0 {
			return fmt.Errorf("format.files[%d].size must be positive", i)
		}

		if file.SHA256 == "" {
			return fmt.Errorf("format.files[%d].sha256 is required", i)
		}

		// Validate SHA256 format (64 hex characters)
		sha256Regex := regexp.MustCompile(`^[a-f0-9]{64}$`)
		if !sha256Regex.MatchString(strings.ToLower(file.SHA256)) {
			return fmt.Errorf("format.files[%d].sha256 must be valid SHA256 hex string", i)
		}
	}

	if len(spec.IO.Inputs) == 0 {
		return fmt.Errorf("io.inputs cannot be empty")
	}

	if len(spec.IO.Outputs) == 0 {
		return fmt.Errorf("io.outputs cannot be empty")
	}

	// Validate compute requirements
	if spec.Requirements.Compute.CPU.MinCores <= 0 {
		return fmt.Errorf("requirements.compute.cpu.min_cores must be positive")
	}

	if spec.Requirements.Compute.Memory.MinGB <= 0 {
		return fmt.Errorf("requirements.compute.memory.min_gb must be positive")
	}

	return nil
}

func validateDistribution(dist types.Distribution) error {
	if dist.Package.URL == "" {
		return fmt.Errorf("distribution.package.url is required")
	}

	if dist.Package.Size <= 0 {
		return fmt.Errorf("distribution.package.size must be positive")
	}

	if dist.Package.SHA256 == "" {
		return fmt.Errorf("distribution.package.sha256 is required")
	}

	// Validate SHA256 format
	sha256Regex := regexp.MustCompile(`^[a-f0-9]{64}$`)
	if !sha256Regex.MatchString(strings.ToLower(dist.Package.SHA256)) {
		return fmt.Errorf("distribution.package.sha256 must be valid SHA256 hex string")
	}

	if dist.Registry.URL == "" {
		return fmt.Errorf("distribution.registry.url is required")
	}

	if dist.Registry.Namespace == "" {
		return fmt.Errorf("distribution.registry.namespace is required")
	}

	return nil
}
