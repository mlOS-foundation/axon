//go:build ignore
// +build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mlOS-foundation/axon/internal/manifest"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// Model definitions for top 10 popular models
var models = []struct {
	namespace   string
	name        string
	version     string
	description string
	framework   string
	fwVersion   string
	format      string
	license     string
	tags        []string
}{
	{"nlp", "bert-base-uncased", "1.0.0", "BERT base model (uncased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "transformer"}},
	{"nlp", "gpt2", "1.0.0", "GPT-2 language model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "transformer", "generation"}},
	{"nlp", "distilbert-base-uncased", "1.0.0", "DistilBERT base model (uncased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "transformer", "distilled"}},
	{"nlp", "roberta-base", "1.0.0", "RoBERTa base model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "roberta", "transformer"}},
	{"nlp", "t5-base", "1.0.0", "T5 base model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "t5", "transformer", "seq2seq"}},
	{"vision", "resnet50", "1.0.0", "ResNet-50 image classification model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"vision", "classification", "resnet"}},
	{"vision", "vit-base-patch16-224", "1.0.0", "Vision Transformer base model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "transformer", "vit"}},
	{"vision", "yolov8n", "1.0.0", "YOLOv8 nano object detection model", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}},
	{"audio", "whisper-base", "1.0.0", "Whisper base speech recognition model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "transcription", "whisper"}},
	{"generation", "stable-diffusion-2-1", "2.1.0", "Stable Diffusion 2.1 image generation model", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image", "stable-diffusion"}},
}

func main() {
	registryDir := "."
	if len(os.Args) > 1 {
		registryDir = os.Args[1]
	}

	now := time.Now()

	for _, model := range models {
		m := &types.Manifest{
			APIVersion: "v1",
			Kind:       "Model",
			Metadata: types.Metadata{
				Name:        model.name,
				Namespace:   model.namespace,
				Version:     model.version,
				Description: model.description,
				License:     model.license,
				Created:     now,
				Updated:     now,
				Tags:        model.tags,
				Authors: []types.Author{
					{
						Name:         "Hugging Face",
						Organization: "Hugging Face",
					},
				},
			},
			Spec: types.Spec{
				Framework: types.Framework{
					Name:    model.framework,
					Version: model.fwVersion,
				},
				Format: types.Format{
					Type: model.format,
					Files: []types.ModelFile{
						{
							Path:   "model.bin",
							Size:   1024 * 1024 * 100, // 100MB placeholder
							SHA256: "placeholder-sha256-checksum",
						},
					},
				},
				IO: types.IO{
					Inputs: []types.IOSpec{
						{
							Name:  "input",
							DType: "float32",
							Shape: []int{-1, -1},
						},
					},
					Outputs: []types.IOSpec{
						{
							Name:  "output",
							DType: "float32",
							Shape: []int{-1, -1},
						},
					},
				},
				Requirements: types.Requirements{
					Compute: types.Compute{
						CPU: types.CPURequirement{
							MinCores:         2,
							RecommendedCores: 4,
						},
						Memory: types.MemoryRequirement{
							MinGB:         2.0,
							RecommendedGB: 4.0,
						},
					},
					Storage: types.Storage{
						MinGB:         0.5,
						RecommendedGB: 1.0,
					},
				},
			},
			Distribution: types.Distribution{
				Package: types.PackageInfo{
					URL:    fmt.Sprintf("http://localhost:8080/packages/%s-%s-%s.axon", model.namespace, model.name, model.version),
					Size:   1024 * 1024 * 100, // 100MB placeholder
					SHA256: "placeholder-sha256-checksum",
				},
				Registry: types.RegistryInfo{
					URL:       "http://localhost:8080",
					Namespace: model.namespace,
				},
			},
		}

		// Compute checksum for package file
		packageFilename := fmt.Sprintf("%s-%s-%s.axon", model.namespace, model.name, model.version)
		packagePath := filepath.Join(registryDir, "packages", packageFilename)
		var packageSHA256 string
		if _, err := os.Stat(packagePath); err == nil {
			// Compute SHA256 of package file
			file, err := os.Open(packagePath)
			if err == nil {
				hasher := sha256.New()
				if _, err := io.Copy(hasher, file); err == nil {
					packageSHA256 = hex.EncodeToString(hasher.Sum(nil))
				}
				file.Close()
			}
		}
		if packageSHA256 == "" {
			packageSHA256 = "placeholder-sha256-checksum"
		}

		// Update manifest with correct checksum
		m.Distribution.Package.SHA256 = packageSHA256

		// Create directory structure
		manifestDir := filepath.Join(registryDir, "api/v1/models", model.namespace, model.name, model.version)
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", manifestDir, err)
			continue
		}

		// Write manifest
		manifestPath := filepath.Join(manifestDir, "manifest.yaml")
		if err := manifest.Write(m, manifestPath); err != nil {
			fmt.Printf("Error writing manifest for %s/%s: %v\n", model.namespace, model.name, err)
			continue
		}

		fmt.Printf("✅ Created manifest: %s/%s@%s (checksum: %s)\n", model.namespace, model.name, model.version, packageSHA256[:16]+"...")
	}

	fmt.Printf("\n✨ Created %d manifests\n", len(models))
}

