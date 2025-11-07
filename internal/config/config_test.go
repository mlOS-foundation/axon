package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.HomeDir == "" {
		t.Error("DefaultConfig() HomeDir should not be empty")
	}

	if cfg.CacheDir == "" {
		t.Error("DefaultConfig() CacheDir should not be empty")
	}

	// Registry.URL is now empty by default (HF adapter is primary, local registry is optional)
	// This is intentional - users can install directly from HF without configuring a registry

	if cfg.Download.Parallel <= 0 {
		t.Error("DefaultConfig() Download.Parallel should be positive")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("ConfigPath() should not be empty")
	}

	// Should contain .axon
	if !filepath.IsAbs(path) {
		t.Error("ConfigPath() should return absolute path")
	}
}

func TestLoad(t *testing.T) {
	// Test loading when config doesn't exist (should return default)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() should return a config")
	}

	// Verify it's a default config (URL is empty by default)
	// Just verify that HF adapter is enabled
	if !cfg.Registry.EnableHuggingFace {
		t.Error("Load() Registry.EnableHuggingFace should be true by default")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome, _ := os.UserHomeDir()

	// Override home directory for test
	_ = os.Setenv("HOME", tmpDir)
	defer func() {
		_ = os.Setenv("HOME", oldHome)
	}()

	cfg := DefaultConfig()
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify directory was created
	configDir := filepath.Join(tmpDir, ".axon")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Save() should create .axon directory")
	}
}
