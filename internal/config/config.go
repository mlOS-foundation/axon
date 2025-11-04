package config

import (
	"os"
	"path/filepath"
)

// Config represents the axon configuration
// This is the myelin sheath - configuration that optimizes signal transmission
type Config struct {
	// Home directory (~/.axon)
	HomeDir string `yaml:"home_dir"`

	// Cache directory
	CacheDir string `yaml:"cache_dir"`

	// Registry configuration
	Registry RegistryConfig `yaml:"registry"`

	// Download settings
	Download DownloadConfig `yaml:"download"`

	// Logging
	LogLevel string `yaml:"log_level"`
}

// RegistryConfig contains registry settings
type RegistryConfig struct {
	// Primary registry URL
	URL string `yaml:"url"`

	// Mirror URLs (axon terminals - multiple endpoints)
	Mirrors []string `yaml:"mirrors"`

	// Authentication token (future)
	Token string `yaml:"token,omitempty"`

	// Timeout settings
	Timeout int `yaml:"timeout"` // seconds
}

// DownloadConfig contains download settings
type DownloadConfig struct {
	// Parallel downloads (saltatory conduction!)
	Parallel int `yaml:"parallel"`

	// Retry settings
	MaxRetries int `yaml:"max_retries"`

	// Verify checksums
	VerifyChecksums bool `yaml:"verify_checksums"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	axonHome := filepath.Join(homeDir, ".axon")

	return &Config{
		HomeDir:  axonHome,
		CacheDir: filepath.Join(axonHome, "cache"),
		Registry: RegistryConfig{
			URL:     "https://registry.axon.mlos.io",
			Mirrors: []string{},
			Timeout: 300,
		},
		Download: DownloadConfig{
			Parallel:        3,
			MaxRetries:      3,
			VerifyChecksums: true,
		},
		LogLevel: "info",
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".axon", "config.yaml")
}

// Load loads configuration from file
// TODO: Implement YAML loading
func Load() (*Config, error) {
	cfgPath := ConfigPath()

	// Check if config file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// TODO: Load from YAML file
	// For now, return default
	return DefaultConfig(), nil
}

// Save saves configuration to file
// TODO: Implement YAML saving
func (c *Config) Save() error {
	cfgPath := ConfigPath()
	cfgDir := filepath.Dir(cfgPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return err
	}

	// TODO: Save to YAML file
	// For now, just ensure directory exists
	return nil
}

