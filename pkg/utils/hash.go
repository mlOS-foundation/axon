package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ComputeSHA256 computes the SHA256 hash of a file
func ComputeSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// VerifySHA256 verifies that a file matches the expected SHA256 checksum
func VerifySHA256(filePath, expectedSHA256 string) error {
	actual, err := ComputeSHA256(filePath)
	if err != nil {
		return err
	}

	expected := fmt.Sprintf("%064s", expectedSHA256)
	actual = fmt.Sprintf("%064s", actual)

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actual)
	}

	return nil
}

// ComputeSHA256Bytes computes the SHA256 hash of byte data
func ComputeSHA256Bytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
