package manifest

import (
	"fmt"
	"os"

	"github.com/mlOS-foundation/axon/pkg/types"
	"gopkg.in/yaml.v3"
)

// Parse parses a YAML manifest file
func Parse(filePath string) (*types.Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest types.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &manifest, nil
}

// ParseBytes parses manifest from byte data
func ParseBytes(data []byte) (*types.Manifest, error) {
	var manifest types.Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &manifest, nil
}

// Write writes a manifest to a YAML file
func Write(manifest *types.Manifest, filePath string) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}
