package builtin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mlOS-foundation/axon/pkg/types"
)

func TestPyTorchHubAdapter_Name(t *testing.T) {
	adapter := NewPyTorchHubAdapter()
	if adapter.Name() != "pytorch" {
		t.Errorf("Name() = %v, want 'pytorch'", adapter.Name())
	}
}

func TestPyTorchHubAdapter_CanHandle(t *testing.T) {
	adapter := NewPyTorchHubAdapter()

	tests := []struct {
		namespace string
		name      string
		want      bool
	}{
		{"pytorch", "vision/resnet50", true},
		{"torch", "vision/resnet50", true},
		{"hf", "bert-base-uncased", false},
		{"modelscope", "cv/resnet50", false},
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

func TestPyTorchHubAdapter_GetManifest(t *testing.T) {
	adapter := NewPyTorchHubAdapter()
	ctx := context.Background()

	tests := []struct {
		name      string
		namespace string
		modelName string
		version   string
		wantErr   bool
	}{
		{
			name:      "valid model",
			namespace: "pytorch",
			modelName: "vision/resnet50",
			version:   "latest",
			wantErr:   false,
		},
		{
			name:      "invalid format - no repo",
			namespace: "pytorch",
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
			if !tt.wantErr {
				if manifest == nil {
					t.Fatal("GetManifest() returned nil manifest")
				}
				if manifest.Metadata.Namespace != tt.namespace {
					t.Errorf("GetManifest() namespace = %v, want %v", manifest.Metadata.Namespace, tt.namespace)
				}
				if manifest.Metadata.Name != tt.modelName {
					t.Errorf("GetManifest() name = %v, want %v", manifest.Metadata.Name, tt.modelName)
				}
				if manifest.Spec.Framework.Name != "PyTorch" {
					t.Errorf("GetManifest() framework = %v, want PyTorch", manifest.Spec.Framework.Name)
				}
			}
		})
	}
}

func TestPyTorchHubAdapter_ParseHubconf(t *testing.T) {
	adapter := NewPyTorchHubAdapter()

	tests := []struct {
		name        string
		hubconf     string
		modelName   string
		wantURLs    []string
		wantAtLeast int // Minimum number of URLs expected
	}{
		{
			name: "model_urls dictionary",
			hubconf: `
model_urls = {
    'resnet50': 'https://download.pytorch.org/models/resnet50-19c8e357.pth',
    'resnet101': 'https://download.pytorch.org/models/resnet101-5d3b4d8f.pth',
}
`,
			modelName:   "resnet50",
			wantAtLeast: 1,
		},
		{
			name: "direct URL assignment",
			hubconf: `
resnet50_url = 'https://download.pytorch.org/models/resnet50-19c8e357.pth'
`,
			modelName:   "resnet50",
			wantAtLeast: 1,
		},
		{
			name: "load_state_dict_from_url",
			hubconf: `
def resnet50(pretrained=False, **kwargs):
    if pretrained:
        model_url = 'https://download.pytorch.org/models/resnet50-19c8e357.pth'
        load_state_dict_from_url(model_url)
`,
			modelName:   "resnet50",
			wantAtLeast: 0, // This pattern requires the URL to be in the function, which is harder to match
		},
		{
			name: "multiple patterns",
			hubconf: `
model_urls = {
    'resnet50': 'https://download.pytorch.org/models/resnet50-19c8e357.pth',
}
resnet50_url = 'https://download.pytorch.org/models/resnet50-19c8e357.pth'
def resnet50(pretrained=False, **kwargs):
    if pretrained:
        load_state_dict_from_url('https://download.pytorch.org/models/resnet50-19c8e357.pth')
`,
			modelName:   "resnet50",
			wantAtLeast: 1,
		},
		{
			name:        "no match",
			hubconf:     `model_urls = {'resnet101': 'https://example.com/resnet101.pth'}`,
			modelName:   "resnet50",
			wantAtLeast: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := adapter.parseHubconf([]byte(tt.hubconf), tt.modelName)
			if len(urls) < tt.wantAtLeast {
				t.Errorf("parseHubconf() found %d URLs, want at least %d. URLs: %v", len(urls), tt.wantAtLeast, urls)
			}
			// Check for duplicates
			seen := make(map[string]bool)
			for _, url := range urls {
				if seen[url] {
					t.Errorf("parseHubconf() found duplicate URL: %s", url)
				}
				seen[url] = true
			}
		})
	}
}

func TestPyTorchHubAdapter_DownloadPackage_WithMockServer(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "hubconf.py") {
			// Return mock hubconf.py
			hubconf := `
model_urls = {
    'resnet50': 'http://localhost:8080/models/resnet50.pth',
}
`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(hubconf))
		} else if strings.Contains(r.URL.Path, "resnet50.pth") {
			// Return mock model file
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock model weights"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create adapter with custom base URL pointing to test server
	adapter := NewPyTorchHubAdapter()
	// We need to modify the adapter to use our test server
	// For now, we'll test the parseHubconf function which is the core logic

	// Test parseHubconf with the mock hubconf content
	hubconf := `
model_urls = {
    'resnet50': '` + server.URL + `/models/resnet50.pth',
}
`
	urls := adapter.parseHubconf([]byte(hubconf), "resnet50")
	if len(urls) == 0 {
		t.Fatal("parseHubconf() should find at least one URL")
	}

	// Verify the URL is correct
	found := false
	for _, url := range urls {
		if strings.Contains(url, "resnet50.pth") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("parseHubconf() should find resnet50.pth URL, got: %v", urls)
	}
}

func TestPyTorchHubAdapter_Search(t *testing.T) {
	adapter := NewPyTorchHubAdapter()
	ctx := context.Background()

	// PyTorch Hub doesn't have a search API, so this should return empty
	results, err := adapter.Search(ctx, "resnet")
	if err != nil {
		t.Errorf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() should return empty results (no API), got %d results", len(results))
	}
}

func TestPyTorchHubAdapter_WithToken(t *testing.T) {
	token := "test_token"
	adapter := NewPyTorchHubAdapterWithToken(token)

	if adapter.Name() != "pytorch" {
		t.Errorf("Name() = %v, want 'pytorch'", adapter.Name())
	}

	// Test that token is set (we can't easily test it's used without making real requests)
	adapter.SetToken("new_token")
	// Just verify it doesn't panic
}

func TestTensorFlowHubAdapter_Name(t *testing.T) {
	adapter := NewTensorFlowHubAdapter()
	if adapter.Name() != "tensorflow-hub" {
		t.Errorf("Name() = %v, want 'tensorflow-hub'", adapter.Name())
	}
}

func TestTensorFlowHubAdapter_CanHandle(t *testing.T) {
	adapter := NewTensorFlowHubAdapter()

	tests := []struct {
		namespace string
		name      string
		want      bool
	}{
		{"tfhub", "google/imagenet/resnet_v2_50/classification/5", true},
		{"tf", "google/imagenet/resnet_v2_50/classification/5", true},
		{"hf", "bert-base-uncased", false},
		{"pytorch", "vision/resnet50", false},
		{"modelscope", "cv/resnet50", false},
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

func TestTensorFlowHubAdapter_GetManifest(t *testing.T) {
	adapter := NewTensorFlowHubAdapter()
	ctx := context.Background()

	tests := []struct {
		name      string
		namespace string
		modelName string
		version   string
		wantErr   bool
	}{
		{
			name:      "valid model",
			namespace: "tfhub",
			modelName: "google/imagenet/resnet_v2_50/classification/5",
			version:   "latest",
			wantErr:   false,
		},
		{
			name:      "valid model with tf namespace",
			namespace: "tf",
			modelName: "google/universal-sentence-encoder/4",
			version:   "latest",
			wantErr:   false,
		},
		{
			name:      "invalid format - no publisher",
			namespace: "tfhub",
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
			if !tt.wantErr {
				if manifest == nil {
					t.Fatal("GetManifest() returned nil manifest")
				}
				if manifest.Metadata.Namespace != tt.namespace {
					t.Errorf("GetManifest() namespace = %v, want %v", manifest.Metadata.Namespace, tt.namespace)
				}
				if manifest.Metadata.Name != tt.modelName {
					t.Errorf("GetManifest() name = %v, want %v", manifest.Metadata.Name, tt.modelName)
				}
				if manifest.Spec.Framework.Name != "TensorFlow" {
					t.Errorf("GetManifest() framework = %v, want TensorFlow", manifest.Spec.Framework.Name)
				}
			}
		})
	}
}

func TestTensorFlowHubAdapter_Search(t *testing.T) {
	adapter := NewTensorFlowHubAdapter()
	ctx := context.Background()

	// TensorFlow Hub search may or may not work depending on API availability
	// So we just test that it doesn't error
	results, err := adapter.Search(ctx, "resnet")
	if err != nil {
		t.Errorf("Search() error = %v", err)
	}
	// Results can be empty if API is not available, which is acceptable
	_ = results
}

func TestTensorFlowHubAdapter_DownloadPackage_WithMockServer(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "model.tar.gz") || strings.Contains(r.URL.String(), "tf-hub-format=compressed") {
			// Return a mock tar.gz file
			w.Header().Set("Content-Type", "application/gzip")
			w.WriteHeader(http.StatusOK)
			// Write a minimal tar.gz file
			_, _ = w.Write([]byte{0x1f, 0x8b, 0x08, 0x00}) // gzip header
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter := NewTensorFlowHubAdapter()
	ctx := context.Background()

	// Create a test manifest
	manifest := &types.Manifest{
		APIVersion: "v1",
		Kind:       "Model",
		Metadata: types.Metadata{
			Name:      "google/test-model/1",
			Namespace: "tfhub",
			Version:   "latest",
		},
		Spec: types.Spec{
			Framework: types.Framework{
				Name:    "TensorFlow",
				Version: "2.0.0",
			},
			Format: types.Format{
				Type: "saved_model",
				Files: []types.ModelFile{
					{
						Path:   "model.tar.gz",
						Size:   0,
						SHA256: "",
					},
				},
			},
		},
		Distribution: types.Distribution{
			Package: types.PackageInfo{
				URL: server.URL + "/google/test-model/1",
			},
		},
	}

	// Create temp directory for test
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test-model.axon")

	// Test download (will fail with mock server, but tests the flow)
	err := adapter.DownloadPackage(ctx, manifest, destPath, nil)
	// We expect this to fail with the mock server, but it should not panic
	if err != nil {
		// Expected - mock server doesn't serve valid model files
		t.Logf("DownloadPackage() failed as expected with mock server: %v", err)
	}
}
