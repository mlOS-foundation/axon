// Package builtin provides tests for builtin adapters.
package builtin

import (
	"context"
	"testing"

	"github.com/mlOS-foundation/axon/internal/registry/core"
)

func TestModelScopeAdapter_Name(t *testing.T) {
	adapter := NewModelScopeAdapter()
	if adapter.Name() != "modelscope" {
		t.Errorf("expected name 'modelscope', got '%s'", adapter.Name())
	}
}

func TestModelScopeAdapter_CanHandle(t *testing.T) {
	adapter := NewModelScopeAdapter()

	tests := []struct {
		namespace string
		name      string
		want      bool
	}{
		{"modelscope", "damo/cv_resnet50", true},
		{"ms", "damo/cv_resnet50", true},
		{"hf", "bert-base-uncased", false},
		{"pytorch", "vision/resnet50", false},
		{"tfhub", "google/resnet", false},
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

func TestModelScopeAdapter_GetManifest(t *testing.T) {
	adapter := NewModelScopeAdapter()
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
			namespace: "modelscope",
			modelName: "damo/cv_resnet50",
			version:   "latest",
			wantErr:   false, // Validation now works correctly, model exists
		},
		{
			name:      "invalid format - no owner",
			namespace: "modelscope",
			modelName: "resnet50",
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

func TestModelScopeFactory_Name(t *testing.T) {
	factory := NewModelScopeFactory()
	if factory.Name() != "modelscope" {
		t.Errorf("expected factory name 'modelscope', got '%s'", factory.Name())
	}
}

func TestModelScopeFactory_Create(t *testing.T) {
	factory := NewModelScopeFactory()

	// Test with default config
	config := core.NewAdapterBuilder().Build()
	adapter, err := factory.Create(config)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if adapter == nil {
		t.Fatal("Create() returned nil adapter")
	}
	if adapter.Name() != "modelscope" {
		t.Errorf("Create() adapter name = %v, want 'modelscope'", adapter.Name())
	}

	// Test with custom config
	customConfig := core.NewAdapterBuilder().
		WithBaseURL("https://custom.modelscope.cn").
		WithToken("test-token").
		Build()

	adapter2, err := factory.Create(customConfig)
	if err != nil {
		t.Fatalf("Create() with custom config error = %v", err)
	}
	if adapter2 == nil {
		t.Fatal("Create() with custom config returned nil adapter")
	}
}
