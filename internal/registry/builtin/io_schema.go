// Package builtin provides I/O schema extraction utilities for model adapters.
package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlOS-foundation/axon/pkg/types"
)

// ExtractIOSchemaFromConfig extracts I/O schema from Hugging Face model config.json
// This is exported so it can be used by adapters and commands
func ExtractIOSchemaFromConfig(configPath string) ([]types.IOSpec, []types.IOSpec, error) {
	// Read config.json
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	// Parse JSON
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Get model type
	modelType, ok := config["model_type"].(string)
	if !ok {
		modelType = "unknown"
	}

	// Extract inputs based on model type
	inputs := extractInputsForModelType(modelType)

	// Extract outputs (typically logits, pooler_output, etc.)
	outputs := extractOutputsForModelType(modelType)

	return inputs, outputs, nil
}

// extractInputsForModelType returns input specs based on model architecture
func extractInputsForModelType(modelType string) []types.IOSpec {
	modelType = strings.ToLower(modelType)

	switch modelType {
	case "bert", "roberta", "distilbert", "albert", "electra":
		// BERT-family models need 3 inputs
		return []types.IOSpec{
			{
				Name:        "input_ids",
				DType:       "int64",
				Shape:       []int{-1, -1}, // batch_size, sequence_length
				Description: "Token IDs from tokenizer",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: modelType,
				},
			},
			{
				Name:        "attention_mask",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Attention mask",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: modelType,
				},
			},
			{
				Name:        "token_type_ids",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Token type IDs (segment IDs)",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: modelType,
				},
			},
		}
	case "gpt2", "gpt", "gpt-neo", "gpt-j":
		// GPT-family models need 2 inputs
		return []types.IOSpec{
			{
				Name:        "input_ids",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Token IDs from tokenizer",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: "gpt2",
				},
			},
			{
				Name:        "attention_mask",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Attention mask",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: "gpt2",
				},
			},
		}
	case "t5", "mt5", "ul2":
		// T5-family models
		return []types.IOSpec{
			{
				Name:        "input_ids",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Token IDs from tokenizer",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: "t5",
				},
			},
			{
				Name:        "attention_mask",
				DType:       "int64",
				Shape:       []int{-1, -1},
				Description: "Attention mask",
				Preprocessing: &types.PreprocessingSpec{
					Type:          "tokenization",
					Tokenizer:     "tokenizer.json",
					TokenizerType: "t5",
				},
			},
		}
	case "vit", "deit", "swin":
		// Vision Transformer models
		return []types.IOSpec{
			{
				Name:        "pixel_values",
				DType:       "float32",
				Shape:       []int{-1, 3, 224, 224}, // batch, channels, height, width
				Description: "Preprocessed image pixels",
				Preprocessing: &types.PreprocessingSpec{
					Type: "normalization",
					Config: map[string]interface{}{
						"mean":   []float64{0.485, 0.456, 0.406},
						"std":    []float64{0.229, 0.224, 0.225},
						"resize": 224,
					},
				},
			},
		}
	default:
		// Generic fallback
		return []types.IOSpec{
			{
				Name:        "input",
				DType:       "float32",
				Shape:       []int{-1, -1},
				Description: "Model input",
			},
		}
	}
}

// extractOutputsForModelType returns output specs based on model architecture
func extractOutputsForModelType(modelType string) []types.IOSpec {
	modelType = strings.ToLower(modelType)

	switch modelType {
	case "bert", "roberta", "distilbert", "albert", "electra", "gpt2", "gpt", "t5":
		// Language models typically output logits
		return []types.IOSpec{
			{
				Name:        "logits",
				DType:       "float32",
				Shape:       []int{-1, -1, -1}, // batch, sequence, vocab_size
				Description: "Model logits",
			},
		}
	case "vit", "deit", "swin":
		// Vision models output class logits
		return []types.IOSpec{
			{
				Name:        "logits",
				DType:       "float32",
				Shape:       []int{-1, -1}, // batch, num_classes
				Description: "Class logits",
			},
		}
	default:
		// Generic fallback
		return []types.IOSpec{
			{
				Name:        "output",
				DType:       "float32",
				Shape:       []int{-1, -1},
				Description: "Model output",
			},
		}
	}
}

// determineExecutionFormat determines execution format based on available files
func determineExecutionFormat(modelPath string, files []string) string {
	// Check if ONNX file exists
	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".onnx") || file == "model.onnx" {
			return "onnx"
		}
	}

	// Check model path for ONNX file
	if _, err := os.Stat(filepath.Join(modelPath, "model.onnx")); err == nil {
		return "onnx"
	}

	// Check for PyTorch files
	for _, file := range files {
		if strings.Contains(strings.ToLower(file), "pytorch") ||
			strings.Contains(strings.ToLower(file), ".pth") ||
			strings.Contains(strings.ToLower(file), ".pt") {
			return "pytorch"
		}
	}

	// Check for TensorFlow files
	for _, file := range files {
		if strings.Contains(strings.ToLower(file), "tensorflow") ||
			strings.Contains(strings.ToLower(file), "saved_model") ||
			strings.Contains(strings.ToLower(file), ".pb") {
			return "tensorflow"
		}
	}

	// Default to ONNX (most models will be converted)
	return "onnx"
}
