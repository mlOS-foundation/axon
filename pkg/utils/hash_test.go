package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeSHA256(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("hello, world")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := ComputeSHA256(testFile)
	if err != nil {
		t.Fatalf("ComputeSHA256() error = %v", err)
	}

	// Expected SHA256 of "hello, world"
	expected := "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b"
	if hash != expected {
		t.Errorf("ComputeSHA256() = %v, want %v", hash, expected)
	}
}

func TestVerifySHA256(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("hello, world")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expectedHash := "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b"

	if err := VerifySHA256(testFile, expectedHash); err != nil {
		t.Errorf("VerifySHA256() error = %v, want nil", err)
	}

	// Test with wrong hash
	if err := VerifySHA256(testFile, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Error("VerifySHA256() should fail with wrong hash")
	}
}

func TestComputeSHA256Bytes(t *testing.T) {
	data := []byte("hello, world")
	hash := ComputeSHA256Bytes(data)

	expected := "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b"
	if hash != expected {
		t.Errorf("ComputeSHA256Bytes() = %v, want %v", hash, expected)
	}
}
