// Package cache provides functionality for managing the local model cache.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// Manager manages the local model cache
type Manager struct {
	cacheDir string
}

// NewManager creates a new cache manager
func NewManager(cacheDir string) *Manager {
	return &Manager{
		cacheDir: cacheDir,
	}
}

// GetModelPath returns the cached path for a model
func (cm *Manager) GetModelPath(namespace, name, version string) string {
	return filepath.Join(cm.cacheDir, "models", namespace, name, version)
}

// IsModelCached checks if a model is already cached
func (cm *Manager) IsModelCached(namespace, name, version string) bool {
	path := cm.GetModelPath(namespace, name, version)
	manifestPath := filepath.Join(path, "manifest.yaml")
	_, err := os.Stat(manifestPath)
	return err == nil
}

// GetCachedManifest retrieves the manifest for a cached model
func (cm *Manager) GetCachedManifest(namespace, name, version string) (*types.Manifest, error) {
	if !cm.IsModelCached(namespace, name, version) {
		return nil, fmt.Errorf("model %s/%s@%s is not cached", namespace, name, version)
	}

	path := cm.GetModelPath(namespace, name, version)
	manifestPath := filepath.Join(path, "manifest.yaml")

	// Import here to avoid circular dependency
	// We'll use a simple approach for now
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse using the manifest package
	// For now, return error - this will be implemented when we wire everything together
	_ = data
	return nil, fmt.Errorf("manifest parsing not yet integrated")
}

// CacheModel caches a model package
func (cm *Manager) CacheModel(namespace, name, version string, manifest *types.Manifest) error {
	path := cm.GetModelPath(namespace, name, version)

	// Create directory structure
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Save manifest as YAML (matches parser expectations)
	manifestPath := filepath.Join(path, "manifest.yaml")
	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Save metadata
	metadataPath := filepath.Join(path, ".axon_metadata.json")
	metadata := map[string]interface{}{
		"installed_at": time.Now().Format(time.RFC3339),
		"namespace":    namespace,
		"name":         name,
		"version":      version,
	}

	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// RemoveModel removes a cached model
func (cm *Manager) RemoveModel(namespace, name, version string) error {
	path := cm.GetModelPath(namespace, name, version)
	return os.RemoveAll(path)
}

// GetCacheSize returns total cache size in bytes
func (cm *Manager) GetCacheSize() (int64, error) {
	var size int64
	err := filepath.Walk(cm.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// ListCachedModels lists all cached models
func (cm *Manager) ListCachedModels() ([]CachedModel, error) {
	modelsDir := filepath.Join(cm.cacheDir, "models")
	var models []CachedModel

	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		return models, nil
	}

	err := filepath.Walk(modelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only process files
		if info.IsDir() {
			return nil
		}

		// Look for metadata files
		if info.Name() == ".axon_metadata.json" {
			// Get relative path from modelsDir
			relPath, err := filepath.Rel(modelsDir, filepath.Dir(path))
			if err != nil {
				return err
			}

			// Split path by filepath separator (works cross-platform)
			// Expected structure: namespace/name/version or namespace/repo/model/version (multi-part names)
			parts := []string{}
			dir := relPath
			for dir != "." && dir != "" {
				base := filepath.Base(dir)
				if base != "" {
					parts = append([]string{base}, parts...)
				}
				dir = filepath.Dir(dir)
			}

			// Need at least 3 parts: namespace, name, version
			// For multi-part names (e.g., pytorch/vision/resnet50/latest):
			// - First part is namespace
			// - Last part is version
			// - Everything in between is the name (joined with /)
			if len(parts) >= 3 {
				namespace := parts[0]
				version := parts[len(parts)-1]
				// Join all parts between namespace and version as the name
				name := filepath.Join(parts[1 : len(parts)-1]...)

				models = append(models, CachedModel{
					Namespace: namespace,
					Name:      name,
					Version:   version,
					Path:      filepath.Dir(path),
				})
			}
		}

		return nil
	})

	return models, err
}

// CachedModel represents a cached model
type CachedModel struct {
	Namespace string
	Name      string
	Version   string
	Path      string
}

// CleanPolicy defines cache cleanup policies
type CleanPolicy struct {
	MaxSizeGB   float64
	MaxAgeHours int
	KeepLatest  int
}

// CleanCache removes unused models based on policy
func (cm *Manager) CleanCache(policy CleanPolicy) error {
	// TODO: Implement cleanup logic
	// - LRU eviction
	// - Size-based cleanup
	// - Age-based cleanup
	return fmt.Errorf("cache cleanup not yet implemented")
}
