package manifest

import (
	"testing"

	"github.com/mlOS-foundation/axon/pkg/types"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		manifest *types.Manifest
		wantErr  bool
	}{
		{
			name: "valid manifest",
			manifest: &types.Manifest{
				APIVersion: "axon.mlos.io/v1",
				Kind:       "Model",
				Metadata: types.Metadata{
					Name:        "test-model",
					Namespace:   "test",
					Version:     "1.0.0",
					Description: "Test model",
					License:     "Apache-2.0",
				},
				Spec: types.Spec{
					Framework: types.Framework{
						Name:    "pytorch",
						Version: "2.0.0",
					},
					Format: types.Format{
						Type: "checkpoint",
						Files: []types.ModelFile{
							{
								Path:   "model.pth",
								Size:   1024,
								SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							},
						},
					},
					IO: types.IO{
						Inputs: []types.IOSpec{
							{Name: "input", DType: "float32", Shape: []int{-1, 224, 224, 3}},
						},
						Outputs: []types.IOSpec{
							{Name: "output", DType: "float32", Shape: []int{-1, 1000}},
						},
					},
					Requirements: types.Requirements{
						Compute: types.Compute{
							CPU: types.CPURequirement{
								MinCores:         2,
								RecommendedCores: 4,
							},
							Memory: types.MemoryRequirement{
								MinGB:         4.0,
								RecommendedGB: 8.0,
							},
						},
					},
				},
				Distribution: types.Distribution{
					Package: types.PackageInfo{
						URL:    "https://example.com/model.axon",
						Size:   1024,
						SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					},
					Registry: types.RegistryInfo{
						URL:       "https://registry.example.com",
						Namespace: "test",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			manifest: &types.Manifest{
				Kind: "Model",
			},
			wantErr: true,
		},
		{
			name: "invalid version",
			manifest: &types.Manifest{
				APIVersion: "axon.mlos.io/v1",
				Kind:       "Model",
				Metadata: types.Metadata{
					Name:        "test",
					Namespace:   "test",
					Version:     "invalid",
					Description: "Test",
					License:     "Apache-2.0",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.manifest)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
