package main

import (
	"fmt"
	"os"

	"github.com/mlOS-foundation/axon/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "axon",
		Short: "The Neural Pathway for ML Models",
		Long:  "Axon is the transmission layer for ML models in MLOS. Signal. Propagate. Myelinate.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize global state - the axon hillock (initiation point)
			var err error
			cfg, err = config.Load()
			if err != nil {
				cfg = config.DefaultConfig()
			}
		},
	}

	// Add commands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(infoCmd())
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(uninstallCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(cacheCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(registryCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
			fmt.Println("(Registry search not yet implemented)")
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
			fmt.Printf("Fetching info for %s...\n", modelSpec)
			fmt.Println("(Model info not yet implemented)")
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
			fmt.Printf("Propagating %s...\n", modelSpec)
			fmt.Println("(Installation not yet implemented)")
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
			fmt.Printf("Pruning pathway for %s...\n", modelSpec)
			fmt.Println("(Uninstallation not yet implemented)")
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
			fmt.Println("Active pathways:")
			fmt.Println("(No models installed yet)")
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
			fmt.Printf("Verifying signal integrity for %s...\n", modelSpec)
			fmt.Println("(Verification not yet implemented)")
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
			fmt.Println("Cached models:")
			fmt.Println("(Cache listing not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "Clean cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Cleaning myelin cache...")
			fmt.Println("(Cache cleanup not yet implemented)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Cache statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Cache statistics:")
			fmt.Println("(Cache stats not yet implemented)")
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
		Use:   "add [url]",
		Short: "Add registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			fmt.Printf("Adding registry: %s\n", url)
			fmt.Println("(Registry add not yet implemented)")
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
