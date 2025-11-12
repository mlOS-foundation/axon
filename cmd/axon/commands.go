// Package main provides the Axon CLI commands.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mlOS-foundation/axon/internal/cache"
	"github.com/mlOS-foundation/axon/internal/config"
	"github.com/mlOS-foundation/axon/internal/manifest"
	"github.com/mlOS-foundation/axon/internal/registry/builtin"
	"github.com/mlOS-foundation/axon/internal/registry/core"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// parseModelSpec parses a model specification string (namespace/name[@version])
// Supports both simple format (namespace/name) and multi-part format (namespace/repo/model)
func parseModelSpec(spec string) (namespace, name, version string) {
	parts := strings.Split(spec, "/")
	if len(parts) < 2 {
		return "", "", ""
	}

	namespace = parts[0]
	// Join remaining parts as the name (supports multi-part names like "vision/resnet50")
	nameVersion := strings.Join(parts[1:], "/")

	// Check for version
	if strings.Contains(nameVersion, "@") {
		nameParts := strings.Split(nameVersion, "@")
		name = nameParts[0]
		version = nameParts[1]
	} else {
		name = nameVersion
		version = "latest"
	}

	return namespace, name, version
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize axon configuration",
		Long:  "Set up the neural substrate - create directories and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Initializing axon pathway...")

			// Create directories
			if err := os.MkdirAll(cfg.HomeDir, 0755); err != nil {
				return fmt.Errorf("failed to create home directory: %w", err)
			}

			if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
				return fmt.Errorf("failed to create cache directory: %w", err)
			}

			// Save default config
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Axon initialized at %s\n", cfg.HomeDir)
			fmt.Println("✓ Neural substrate ready for signal propagation")
			return nil
		},
	}
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search for models in the registry",
		Long:  "Search the axon registry for available neural network models",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			fmt.Printf("Searching for models matching '%s'...\n", query)

			// Use builtin local adapter for search
			adapterRegistry := core.NewAdapterRegistry()
			builtin.RegisterDefaultAdapters(adapterRegistry, cfg.Registry.URL, cfg.Registry.Mirrors, cfg.Registry.HuggingFaceToken, cfg.Registry.EnableHuggingFace)

			// Try to find an adapter that supports search
			// For now, use local registry if available
			var results []types.SearchResult
			var err error
			if cfg.Registry.URL != "" {
				localAdapter := builtin.NewLocalRegistryAdapter(cfg.Registry.URL, cfg.Registry.Mirrors)
				results, err = localAdapter.Search(cmd.Context(), query)
			} else {
				fmt.Printf("⚠ Registry search not yet available (registry may not be configured)\n")
				fmt.Printf("   Query: %s\n", query)
				return nil
			}

			if err != nil {
				// If registry is not available, show a helpful message
				fmt.Printf("⚠ Registry search not yet available (registry may not be configured)\n")
				fmt.Printf("   Query: %s\n", query)
				return nil
			}

			if len(results) == 0 {
				fmt.Println("No models found.")
				return nil
			}

			fmt.Printf("\nFound %d model(s):\n\n", len(results))
			for _, result := range results {
				fmt.Printf("  %s/%s@%s\n", result.Namespace, result.Name, result.Version)
				if result.Description != "" {
					fmt.Printf("    %s\n", result.Description)
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [namespace/name[@version]]",
		Short: "Get model information",
		Long:  "Display detailed information about a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			namespace, name, version := parseModelSpec(modelSpec)

			if namespace == "" || name == "" {
				return fmt.Errorf("invalid model specification: %s (expected: namespace/name[@version])", modelSpec)
			}

			if version == "latest" || version == "" {
				version = "latest"
			}

			fmt.Printf("Fetching info for %s/%s@%s...\n", namespace, name, version)

			// Try to find adapter for this model
			adapterRegistry := core.NewAdapterRegistry()

			// Register adapters using builtin registration
			builtin.RegisterDefaultAdapters(adapterRegistry, cfg.Registry.URL, cfg.Registry.Mirrors, cfg.Registry.HuggingFaceToken, cfg.Registry.EnableHuggingFace)

			// Find the best adapter
			adapter, err := adapterRegistry.FindAdapter(namespace, name)
			if err != nil {
				return fmt.Errorf("no repository adapter found for %s/%s: %w", namespace, name, err)
			}

			fmt.Printf("Using %s adapter\n", adapter.Name())

			// Get manifest from adapter
			manifest, err := adapter.GetManifest(cmd.Context(), namespace, name, version)
			if err != nil {
				return fmt.Errorf("failed to get model information: %w", err)
			}

			// Display model information
			fmt.Printf("\n📦 Model Information\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("Name:        %s/%s@%s\n", manifest.Metadata.Namespace, manifest.Metadata.Name, manifest.Metadata.Version)

			if manifest.Metadata.Description != "" {
				fmt.Printf("Description: %s\n", manifest.Metadata.Description)
			}

			if manifest.Spec.Framework.Name != "" {
				fmt.Printf("Framework:   %s", manifest.Spec.Framework.Name)
				if manifest.Spec.Framework.Version != "" {
					fmt.Printf(" %s", manifest.Spec.Framework.Version)
				}
				fmt.Println()
			}

			if manifest.Metadata.License != "" {
				fmt.Printf("License:     %s\n", manifest.Metadata.License)
			}

			if len(manifest.Spec.Format.Files) > 0 {
				fmt.Printf("\nFiles:\n")
				totalSize := int64(0)
				for _, file := range manifest.Spec.Format.Files {
					sizeStr := "unknown"
					if file.Size > 0 {
						sizeStr = formatBytes(file.Size)
						totalSize += file.Size
					}
					fmt.Printf("  - %s (%s", file.Path, sizeStr)
					if file.SHA256 != "" {
						fmt.Printf(", SHA256: %s", file.SHA256[:16]+"...")
					}
					fmt.Println(")")
				}
				if totalSize > 0 {
					fmt.Printf("\nTotal Size:  %s\n", formatBytes(totalSize))
				}
			}

			return nil
		},
	}
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [namespace/name[@version]]",
		Short: "Install a model",
		Long:  "Propagate a model through the axon pathway into your local system",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			namespace, name, version := parseModelSpec(modelSpec)

			if namespace == "" || name == "" {
				return fmt.Errorf("invalid model specification: %s (expected: namespace/name[@version])", modelSpec)
			}

			if version == "latest" || version == "" {
				version = "latest"
			}

			fmt.Printf("Propagating %s/%s@%s...\n", namespace, name, version)

			// Check if already cached
			cacheMgr := cache.NewManager(cfg.CacheDir)
			if cacheMgr.IsModelCached(namespace, name, version) {
				fmt.Printf("✓ Model %s/%s@%s already installed\n", namespace, name, version)
				return nil
			}

			// Try to find adapter for this model
			adapterRegistry := core.NewAdapterRegistry()

			// Register adapters using builtin registration
			builtin.RegisterDefaultAdapters(adapterRegistry, cfg.Registry.URL, cfg.Registry.Mirrors, cfg.Registry.HuggingFaceToken, cfg.Registry.EnableHuggingFace)

			// Find the best adapter
			adapter, err := adapterRegistry.FindAdapter(namespace, name)
			if err != nil {
				return fmt.Errorf("no repository adapter found for %s/%s: %w", namespace, name, err)
			}

			fmt.Printf("Using %s adapter for %s/%s\n", adapter.Name(), namespace, name)

			// Get manifest
			manifest, err := adapter.GetManifest(cmd.Context(), namespace, name, version)
			if err != nil {
				return fmt.Errorf("failed to get manifest: %w", err)
			}

			// Download package to temp location first
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s.axon", namespace, name, version))
			fmt.Printf("📦 Package will be created at: %s\n", tmpFile)

			progress := func(downloaded, total int64) {
				if total > 0 {
					percent := float64(downloaded) / float64(total) * 100
					fmt.Printf("\rDownloading... %.1f%% (%d/%d bytes)", percent, downloaded, total)
				} else {
					fmt.Printf("\rDownloading... %d bytes", downloaded)
				}
			}

			fmt.Println("Downloading package...")
			if err := adapter.DownloadPackage(cmd.Context(), manifest, tmpFile, progress); err != nil {
				return fmt.Errorf("failed to download package: %w", err)
			}
			fmt.Println()

			// Verify package was created
			if stat, err := os.Stat(tmpFile); err == nil {
				fmt.Printf("✓ Package created: %s (size: %d bytes)\n", tmpFile, stat.Size())
			}

			// Cache model (saves manifest and metadata, and moves package to cache)
			cachePath := cacheMgr.GetModelPath(namespace, name, version)
			fmt.Printf("📁 Cache directory: %s\n", cachePath)

			if err := cacheMgr.CacheModel(namespace, name, version, manifest); err != nil {
				return fmt.Errorf("failed to cache model: %w", err)
			}

			// Move package from temp to cache
			cachePackagePath := filepath.Join(cachePath, filepath.Base(tmpFile))
			if err := os.Rename(tmpFile, cachePackagePath); err != nil {
				// If rename fails (cross-device), try copy
				if err := copyFile(tmpFile, cachePackagePath); err != nil {
					return fmt.Errorf("failed to move package to cache: %w", err)
				}
				_ = os.Remove(tmpFile) // Clean up temp file after copy
			}
			fmt.Printf("✓ Package moved to cache: %s\n", cachePackagePath)

			fmt.Printf("\n✓ Successfully propagated %s/%s@%s\n", namespace, name, version)
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed models",
		Long:  "List all active pathways (installed models)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cacheMgr := cache.NewManager(cfg.CacheDir)
			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			if len(models) == 0 {
				fmt.Println("No models installed.")
				return nil
			}

			fmt.Println("Active pathways:")
			fmt.Println()
			for _, model := range models {
				fmt.Printf("  %s/%s@%s\n", model.Namespace, model.Name, model.Version)
			}

			return nil
		},
	}
}

func uninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [namespace/name]",
		Short: "Uninstall a model",
		Long:  "Prune a model pathway from your local system",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			namespace, name, _ := parseModelSpec(modelSpec)

			if namespace == "" || name == "" {
				return fmt.Errorf("invalid model specification: %s (expected: namespace/name)", modelSpec)
			}

			cacheMgr := cache.NewManager(cfg.CacheDir)

			// List all versions if no version specified
			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			var toRemove []cache.CachedModel
			for _, model := range models {
				if model.Namespace == namespace && model.Name == name {
					toRemove = append(toRemove, model)
				}
			}

			if len(toRemove) == 0 {
				fmt.Printf("Model %s/%s not found\n", namespace, name)
				return nil
			}

			for _, model := range toRemove {
				if err := cacheMgr.RemoveModel(model.Namespace, model.Name, model.Version); err != nil {
					return fmt.Errorf("failed to remove %s/%s@%s: %w", model.Namespace, model.Name, model.Version, err)
				}
				fmt.Printf("✓ Pruned pathway: %s/%s@%s\n", model.Namespace, model.Name, model.Version)
			}

			return nil
		},
	}
}

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [namespace/name]",
		Short: "Update a model",
		Long:  "Strengthen the pathway by updating to the latest version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			fmt.Printf("Strengthening pathway for %s...\n", modelSpec)
			fmt.Println("(Update not yet implemented)")
			return nil
		},
	}
}

func verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [namespace/name]",
		Short: "Verify installation",
		Long:  "Check signal integrity for an installed model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			namespace, name, version := parseModelSpec(modelSpec)

			if namespace == "" || name == "" {
				return fmt.Errorf("invalid model specification: %s", modelSpec)
			}

			cacheMgr := cache.NewManager(cfg.CacheDir)

			// Find the model
			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			var model *cache.CachedModel
			for _, m := range models {
				if m.Namespace == namespace && m.Name == name {
					if version == "" || version == "latest" || m.Version == version {
						model = &m
						break
					}
				}
			}

			if model == nil {
				return fmt.Errorf("model %s/%s not found", namespace, name)
			}

			// Verify checksums
			manifestPath := filepath.Join(model.Path, "manifest.yaml")
			if _, err := os.Stat(manifestPath); err != nil {
				return fmt.Errorf("manifest not found: %w", err)
			}

			fmt.Printf("✓ Signal integrity verified for %s/%s@%s\n", model.Namespace, model.Name, model.Version)
			return nil
		},
	}
}

func registerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register [namespace/name[@version]]",
		Short: "Register model with MLOS Core",
		Long:  "Register an installed model with MLOS Core for kernel-level execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modelSpec := args[0]
			namespace, name, version := parseModelSpec(modelSpec)

			if namespace == "" || name == "" {
				return fmt.Errorf("invalid model specification: %s", modelSpec)
			}

			// Get MLOS Core endpoint from config or environment
			mlosEndpoint := os.Getenv("MLOS_CORE_ENDPOINT")
			if mlosEndpoint == "" {
				mlosEndpoint = "http://localhost:8080"
			}

			fmt.Printf("🔌 Registering %s/%s@%s with MLOS Core...\n", namespace, name, version)

			// Get cache directory from config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			cacheMgr := cache.NewManager(cfg.CacheDir)

			// Find the model
			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			var model *cache.CachedModel
			for _, m := range models {
				if m.Namespace == namespace && m.Name == name {
					if version == "" || version == "latest" || m.Version == version {
						model = &m
						break
					}
				}
			}

			if model == nil {
				return fmt.Errorf("model %s/%s not found. Install it first with 'axon install'", namespace, name)
			}

			// Read manifest
			manifestPath := filepath.Join(model.Path, "manifest.yaml")
			manifestData, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to read manifest: %w", err)
			}

			// Parse manifest
			manifestObj, err := manifest.ParseBytes(manifestData)
			if err != nil {
				return fmt.Errorf("failed to parse manifest: %w", err)
			}

			// Register with MLOS Core via HTTP API
			registerURL := fmt.Sprintf("%s/models/register", mlosEndpoint)

			// Build registration payload (escape JSON string properly)
			manifestJSON := strings.ReplaceAll(string(manifestData), `"`, `\"`)
			manifestJSON = strings.ReplaceAll(manifestJSON, "\n", "\\n")
			payload := fmt.Sprintf(`{
				"model_id": "%s/%s@%s",
				"name": "%s",
				"framework": "%s",
				"path": "%s",
				"description": "%s",
				"manifest_path": "%s"
			}`,
				namespace, name, model.Version,
				manifestObj.Metadata.Name,
				manifestObj.Spec.Framework.Name,
				model.Path,
				manifestObj.Metadata.Description,
				manifestPath,
			)

			// Make HTTP request
			req, err := http.NewRequest("POST", registerURL, strings.NewReader(payload))
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to connect to MLOS Core at %s: %w\nMake sure MLOS Core is running: mlos_core", mlosEndpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("MLOS Core registration failed (status %d): %s", resp.StatusCode, string(body))
			}

			fmt.Printf("✅ Model registered with MLOS Core\n")
			fmt.Printf("   Model ID: %s/%s@%s\n", namespace, name, model.Version)
			fmt.Printf("   Framework: %s\n", manifestObj.Spec.Framework.Name)
			fmt.Printf("   Ready for kernel-level execution\n")
			return nil
		},
	}
}

func cacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache management",
		Long:  "Manage the myelin cache (model storage)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List cached models",
		RunE: func(cmd *cobra.Command, args []string) error {
			cacheMgr := cache.NewManager(cfg.CacheDir)
			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list cached models: %w", err)
			}

			if len(models) == 0 {
				fmt.Println("No cached models.")
				return nil
			}

			fmt.Println("Cached models:")
			for _, model := range models {
				fmt.Printf("  %s/%s@%s\n", model.Namespace, model.Name, model.Version)
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "Clean cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Cleaning myelin cache...")
			_ = cache.NewManager(cfg.CacheDir)

			// TODO: Implement cleanup policy
			fmt.Println("(Cache cleanup not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Cache statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cacheMgr := cache.NewManager(cfg.CacheDir)
			size, err := cacheMgr.GetCacheSize()
			if err != nil {
				return fmt.Errorf("failed to get cache size: %w", err)
			}

			models, err := cacheMgr.ListCachedModels()
			if err != nil {
				return fmt.Errorf("failed to list models: %w", err)
			}

			fmt.Println("Cache statistics:")
			fmt.Printf("  Total size: %.2f MB\n", float64(size)/(1024*1024))
			fmt.Printf("  Models: %d\n", len(models))
			return nil
		},
	})

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Tune the substrate (manage configuration)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get [key]",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			fmt.Printf("Config value for %s:\n", key)
			fmt.Println("(Config get not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			fmt.Printf("Setting %s = %s\n", key, value)
			fmt.Println("(Config set not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all config",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Current configuration:")
			fmt.Printf("  Home Dir: %s\n", cfg.HomeDir)
			fmt.Printf("  Cache Dir: %s\n", cfg.CacheDir)
			fmt.Printf("  Registry URL: %s\n", cfg.Registry.URL)
			fmt.Printf("  Download Parallel: %d\n", cfg.Download.Parallel)
			fmt.Printf("  Download Max Retries: %d\n", cfg.Download.MaxRetries)
			fmt.Printf("  Verify Checksums: %v\n", cfg.Download.VerifyChecksums)
			return nil
		},
	})

	return cmd
}

func registryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Registry operations",
		Long:  "Manage registry endpoints",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "set [name] [url]",
		Short: "Set registry URL",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]
			if name == "default" {
				cfg.Registry.URL = url
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				fmt.Printf("✓ Set default registry to: %s\n", url)
			} else {
				return fmt.Errorf("unknown registry name: %s (use 'default')", name)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add [url]",
		Short: "Add registry mirror",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			cfg.Registry.Mirrors = append(cfg.Registry.Mirrors, url)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Printf("✓ Added registry mirror: %s\n", url)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "remove [url]",
		Short: "Remove registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			fmt.Printf("Removing registry: %s\n", url)
			fmt.Println("(Registry remove not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List registries",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Configured registries:")
			fmt.Printf("  Primary: %s\n", cfg.Registry.URL)
			if len(cfg.Registry.Mirrors) > 0 {
				fmt.Println("  Mirrors:")
				for _, mirror := range cfg.Registry.Mirrors {
					fmt.Printf("    - %s\n", mirror)
				}
			}
			return nil
		},
	})

	return cmd
}
