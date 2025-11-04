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
