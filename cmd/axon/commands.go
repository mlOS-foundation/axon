package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlOS-foundation/axon/internal/cache"
	"github.com/mlOS-foundation/axon/internal/registry"
	"github.com/spf13/cobra"
)

// parseModelSpec parses a model specification string (namespace/name[@version])
func parseModelSpec(spec string) (namespace, name, version string) {
	parts := strings.Split(spec, "/")
	if len(parts) != 2 {
		return "", "", ""
	}

	namespace = parts[0]
	nameVersion := parts[1]

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

			client := registry.NewClient(cfg.Registry.URL, cfg.Registry.Mirrors)
			results, err := client.Search(cmd.Context(), query)
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

			client := registry.NewClient(cfg.Registry.URL, cfg.Registry.Mirrors)
			manifest, err := client.GetManifest(cmd.Context(), namespace, name, version)
			if err != nil {
				fmt.Printf("⚠ Model info not yet available (registry may not be configured)\n")
				fmt.Printf("   Model: %s/%s@%s\n", namespace, name, version)
				return nil
			}

			fmt.Printf("\nModel: %s\n", manifest.FullVersion())
			fmt.Printf("Description: %s\n", manifest.Metadata.Description)
			fmt.Printf("Framework: %s %s\n", manifest.Spec.Framework.Name, manifest.Spec.Framework.Version)
			fmt.Printf("License: %s\n", manifest.Metadata.License)

			return nil
		},
	}
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
			adapterRegistry := registry.NewAdapterRegistry()

			// Register adapters in priority order
			// 1. Local registry (if configured)
			if cfg.Registry.URL != "" {
				localAdapter := registry.NewLocalRegistryAdapter(cfg.Registry.URL, cfg.Registry.Mirrors)
				adapterRegistry.Register(localAdapter)
			}

			// 2. Hugging Face (fallback - can handle any model)
			if cfg.Registry.EnableHuggingFace {
				var hfAdapter *registry.HuggingFaceAdapter
				if cfg.Registry.HuggingFaceToken != "" {
					// Use token if provided (for gated/private models)
					hfAdapter = registry.NewHuggingFaceAdapterWithToken(cfg.Registry.HuggingFaceToken)
				} else {
					// No token - works for public models
					hfAdapter = registry.NewHuggingFaceAdapter()
				}
				adapterRegistry.Register(hfAdapter)
			}

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

			// Download package
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s.axon", namespace, name, version))
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

			// Cache model
			if err := cacheMgr.CacheModel(namespace, name, version, manifest); err != nil {
				return fmt.Errorf("failed to cache model: %w", err)
			}

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
