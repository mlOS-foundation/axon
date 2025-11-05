package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	// Primary registry URL (local Axon registry)
	URL string `yaml:"url"`

	// Mirror URLs (axon terminals - multiple endpoints)
	Mirrors []string `yaml:"mirrors"`

	// Enable Hugging Face adapter (for real-time downloads)
	EnableHuggingFace bool `yaml:"enable_huggingface"`

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
			URL:               "",
			Mirrors:           []string{},
			EnableHuggingFace: true, // Enable HF adapter by default
			Timeout:           300,
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
func Load() (*Config, error) {
	cfgPath := ConfigPath()

	// Check if config file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Load from YAML file
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	cfgPath := ConfigPath()
	cfgDir := filepath.Dir(cfgPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save to YAML file
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
