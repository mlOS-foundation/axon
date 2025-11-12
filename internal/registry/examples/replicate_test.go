// Package examples provides tests for example adapters.
package examples

import (
	"context"
	"testing"
)

func TestReplicateAdapter_Name(t *testing.T) {
	adapter := NewReplicateAdapter()
	if adapter.Name() != "replicate" {
		t.Errorf("expected name 'replicate', got '%s'", adapter.Name())
	}
}

func TestReplicateAdapter_CanHandle(t *testing.T) {
	adapter := NewReplicateAdapter()

	tests := []struct {
		namespace string
		name      string
		want      bool
	}{
		{"replicate", "stability-ai/stable-diffusion", true},
		{"rep", "stability-ai/stable-diffusion", true},
		{"hf", "bert-base-uncased", false},
		{"pytorch", "vision/resnet50", false},
		{"tfhub", "google/resnet", false},
		{"modelscope", "damo/cv_resnet50", false},
		{"", "resnet50", false},
	}

	for _, tt := range tests {
		t.Run(tt.namespace+"/"+tt.name, func(t *testing.T) {
			if got := adapter.CanHandle(tt.namespace, tt.name); got != tt.want {
				t.Errorf("CanHandle(%q, %q) = %v, want %v", tt.namespace, tt.name, got, tt.want)
			}
		})
	}
}

func TestReplicateAdapter_GetManifest(t *testing.T) {
	adapter := NewReplicateAdapter()
	ctx := context.Background()

	tests := []struct {
		name      string
		namespace string
		modelName string
		version   string
		wantErr   bool
	}{
		{
			name:      "valid model format",
			namespace: "replicate",
			modelName: "stability-ai/stable-diffusion",
			version:   "latest",
			wantErr:   false, // May succeed if model exists, or fail if not - both are valid
		},
		{
			name:      "invalid format - no owner",
			namespace: "replicate",
			modelName: "stable-diffusion",
			version:   "latest",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := adapter.GetManifest(ctx, tt.namespace, tt.modelName, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && manifest == nil {
				t.Fatal("GetManifest() returned nil manifest")
			}
			if !tt.wantErr {
				if manifest.Metadata.Namespace != tt.namespace {
					t.Errorf("GetManifest() namespace = %v, want %v", manifest.Metadata.Namespace, tt.namespace)
				}
				if manifest.Metadata.Name != tt.modelName {
					t.Errorf("GetManifest() name = %v, want %v", manifest.Metadata.Name, tt.modelName)
				}
			}
		})
	}
}

func TestReplicateAdapterWithToken(t *testing.T) {
	adapter := NewReplicateAdapterWithToken("test-token")
	if adapter.Name() != "replicate" {
		t.Errorf("expected name 'replicate', got '%s'", adapter.Name())
	}
	if adapter.apiToken != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", adapter.apiToken)
	}
}

