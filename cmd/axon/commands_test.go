package main

import (
	"testing"
)

func TestSafeTempFileName(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		modelName string
		version   string
		expected  string
	}{
		{
			name:      "simple namespace",
			namespace: "hf",
			modelName: "distilgpt2",
			version:   "latest",
			expected:  "hf-distilgpt2-latest.axon",
		},
		{
			name:      "nested namespace with one slash",
			namespace: "hf/microsoft",
			modelName: "resnet-50",
			version:   "latest",
			expected:  "hf_microsoft-resnet-50-latest.axon",
		},
		{
			name:      "nested namespace with multiple slashes",
			namespace: "hf/google/research",
			modelName: "vit-base",
			version:   "1.0.0",
			expected:  "hf_google_research-vit-base-1.0.0.axon",
		},
		{
			name:      "model name with slash",
			namespace: "pytorch",
			modelName: "vision/resnet50",
			version:   "latest",
			expected:  "pytorch-vision_resnet50-latest.axon",
		},
		{
			name:      "version with slash",
			namespace: "hf",
			modelName: "bert",
			version:   "v1/beta",
			expected:  "hf-bert-v1_beta.axon",
		},
		{
			name:      "backslash handling",
			namespace: "hf\\microsoft",
			modelName: "resnet\\50",
			version:   "latest",
			expected:  "hf_microsoft-resnet_50-latest.axon",
		},
		{
			name:      "empty namespace",
			namespace: "",
			modelName: "model",
			version:   "1.0",
			expected:  "-model-1.0.axon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeTempFileName(tt.namespace, tt.modelName, tt.version)
			if result != tt.expected {
				t.Errorf("safeTempFileName(%q, %q, %q) = %q, expected %q",
					tt.namespace, tt.modelName, tt.version, result, tt.expected)
			}
		})
	}
}
