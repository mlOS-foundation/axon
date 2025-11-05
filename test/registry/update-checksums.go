package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlOS-foundation/axon/internal/manifest"
	"github.com/mlOS-foundation/axon/pkg/types"
	"gopkg.in/yaml.v3"
)

func main() {
	registryDir := "."
	if len(os.Args) > 1 {
		registryDir = os.Args[1]
	}

	manifestsDir := filepath.Join(registryDir, "api/v1/models")
	packagesDir := filepath.Join(registryDir, "packages")

	updated := 0
	skipped := 0

	err := filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, "manifest.yaml") {
			return nil
		}

		// Read manifest
		m, err := manifest.Parse(path)
		if err != nil {
			fmt.Printf("⚠️  Failed to parse %s: %v\n", path, err)
			return nil
		}

		// Find corresponding package file
		packageFilename := fmt.Sprintf("%s-%s-%s.axon", m.Metadata.Namespace, m.Metadata.Name, m.Metadata.Version)
		packagePath := filepath.Join(packagesDir, packageFilename)

		// Compute checksum if package exists
		if _, err := os.Stat(packagePath); os.IsNotExist(err) {
			skipped++
			return nil
		}

		file, err := os.Open(packagePath)
		if err != nil {
			skipped++
			return nil
		}

		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			file.Close()
			skipped++
			return nil
		}
		file.Close()

		packageSHA256 := hex.EncodeToString(hasher.Sum(nil))
		
		// Update manifest if checksum changed
		if m.Distribution.Package.SHA256 != packageSHA256 {
			m.Distribution.Package.SHA256 = packageSHA256
			
			// Also update file size
			if stat, err := os.Stat(packagePath); err == nil {
				m.Distribution.Package.Size = stat.Size()
			}

			// Write updated manifest
			if err := manifest.Write(m, path); err != nil {
				fmt.Printf("⚠️  Failed to update %s: %v\n", path, err)
				return nil
			}

			updated++
		}

		return nil
	})

	if err != nil {
		fmt.Printf("❌ Error walking manifests: %v\n", err)
		return
	}

	fmt.Printf("✅ Updated %d manifests with checksums", updated)
	if skipped > 0 {
		fmt.Printf(" (%d skipped - no package file)", skipped)
	}
	fmt.Printf("\n")
}

