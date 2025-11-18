package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mlOS-foundation/axon/internal/config"
)

var (
	cfg *config.Config

	// Version information set via ldflags during build
	version   = "dev"     // Version (e.g., "1.7.0")
	buildDate = "unknown" // Build date (ISO 8601 format)
	gitCommit = "unknown" // Git commit hash
	buildType = "local"   // Build type: "local" or "release"
)

func init() {
	// Detect if this is an installed version (in PATH) vs local build
	if buildType == "local" {
		// Check if binary is in ~/.local/bin (local build) or system PATH (installed)
		execPath, err := os.Executable()
		if err == nil {
			home := os.Getenv("HOME")
			if home != "" && !strings.Contains(execPath, home+"/.local/bin") {
				buildType = "installed"
			}
		}
	}
}

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
	rootCmd.AddCommand(publishCmd())
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(cacheCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(registryCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
