package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlOS-foundation/axon/internal/manifest"
	"github.com/mlOS-foundation/axon/pkg/types"
)

// Top 100 Hugging Face models based on popularity and downloads
// This list includes models from various categories: NLP, Vision, Audio, etc.
var top100Models = []struct {
	namespace   string
	name        string
	version     string
	description string
	framework   string
	fwVersion   string
	format      string
	license     string
	tags        []string
	category    string
}{
	// NLP Models (0-39)
	{"nlp", "bert-base-uncased", "1.0.0", "BERT base model (uncased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "transformer"}, "nlp"},
	{"nlp", "gpt2", "1.0.0", "GPT-2 language model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "transformer", "generation"}, "nlp"},
	{"nlp", "distilbert-base-uncased", "1.0.0", "DistilBERT base model (uncased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "transformer", "distilled"}, "nlp"},
	{"nlp", "roberta-base", "1.0.0", "RoBERTa base model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "roberta", "transformer"}, "nlp"},
	{"nlp", "t5-base", "1.0.0", "T5 base model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "t5", "transformer", "seq2seq"}, "nlp"},
	{"nlp", "bert-large-uncased", "1.0.0", "BERT large model (uncased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "transformer"}, "nlp"},
	{"nlp", "gpt2-large", "1.0.0", "GPT-2 large language model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "transformer", "generation"}, "nlp"},
	{"nlp", "distilgpt2", "1.0.0", "DistilGPT-2 language model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "gpt", "transformer", "distilled"}, "nlp"},
	{"nlp", "albert-base-v2", "1.0.0", "ALBERT base model v2", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "albert", "transformer"}, "nlp"},
	{"nlp", "xlnet-base-cased", "1.0.0", "XLNet base model (cased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "xlnet", "transformer"}, "nlp"},
	{"nlp", "bert-base-multilingual-cased", "1.0.0", "BERT base multilingual (cased)", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "transformer"}, "nlp"},
	{"nlp", "sentence-transformers/all-MiniLM-L6-v2", "1.0.0", "Sentence transformer model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "sentence-transformers", "embeddings"}, "nlp"},
	{"nlp", "sentence-transformers/all-mpnet-base-v2", "1.0.0", "Sentence transformer MPNet model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "sentence-transformers", "embeddings"}, "nlp"},
	{"nlp", "facebook/bart-base", "1.0.0", "BART base model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "bart", "transformer", "seq2seq"}, "nlp"},
	{"nlp", "microsoft/DialoGPT-medium", "1.0.0", "DialoGPT medium model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "chatbot", "dialogue"}, "nlp"},
	{"nlp", "google/flan-t5-base", "1.0.0", "FLAN-T5 base model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "t5", "flan", "instruction"}, "nlp"},
	{"nlp", "google/flan-t5-large", "1.0.0", "FLAN-T5 large model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "t5", "flan", "instruction"}, "nlp"},
	{"nlp", "EleutherAI/gpt-neo-1.3B", "1.0.0", "GPT-Neo 1.3B model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "gpt", "transformer"}, "nlp"},
	{"nlp", "EleutherAI/gpt-neo-2.7B", "1.0.0", "GPT-Neo 2.7B model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "gpt", "transformer"}, "nlp"},
	{"nlp", "microsoft/deberta-base", "1.0.0", "DeBERTa base model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "deberta", "transformer"}, "nlp"},
	{"nlp", "sentence-transformers/paraphrase-mpnet-base-v2", "1.0.0", "Paraphrase MPNet model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "sentence-transformers", "paraphrase"}, "nlp"},
	{"nlp", "sentence-transformers/paraphrase-MiniLM-L6-v2", "1.0.0", "Paraphrase MiniLM model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "sentence-transformers", "paraphrase"}, "nlp"},
	{"nlp", "facebook/bart-large", "1.0.0", "BART large model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "bart", "transformer"}, "nlp"},
	{"nlp", "microsoft/DialoGPT-large", "1.0.0", "DialoGPT large model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "chatbot"}, "nlp"},
	{"nlp", "deepset/roberta-base-squad2", "1.0.0", "RoBERTa for question answering", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "roberta", "qa", "squad"}, "nlp"},
	{"nlp", "deepset/bert-base-cased-squad2", "1.0.0", "BERT for question answering", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "qa", "squad"}, "nlp"},
	{"nlp", "nlptown/bert-base-multilingual-uncased-sentiment", "1.0.0", "BERT multilingual sentiment", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "sentiment"}, "nlp"},
	{"nlp", "cardiffnlp/twitter-roberta-base-sentiment", "1.0.0", "RoBERTa Twitter sentiment", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "roberta", "sentiment", "twitter"}, "nlp"},
	{"nlp", "ProsusAI/finbert", "1.0.0", "FinBERT financial sentiment", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "finance", "sentiment"}, "nlp"},
	{"nlp", "j-hartmann/emotion-english-distilroberta-base", "1.0.0", "Emotion classification model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "emotion", "roberta"}, "nlp"},
	{"nlp", "microsoft/DialoGPT-small", "1.0.0", "DialoGPT small model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "gpt", "chatbot"}, "nlp"},
	{"nlp", "google/electra-base-discriminator", "1.0.0", "ELECTRA base discriminator", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "electra", "transformer"}, "nlp"},
	{"nlp", "allenai/scibert_scivocab_uncased", "1.0.0", "SciBERT scientific text", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "scientific"}, "nlp"},
	{"nlp", "dbmdz/bert-base-turkish-cased", "1.0.0", "BERT Turkish model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "turkish"}, "nlp"},
	{"nlp", "dbmdz/bert-base-german-cased", "1.0.0", "BERT German model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "german"}, "nlp"},
	{"nlp", "dbmdz/bert-base-french-cased", "1.0.0", "BERT French model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "french"}, "nlp"},
	{"nlp", "dbmdz/bert-base-spanish-cased", "1.0.0", "BERT Spanish model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "spanish"}, "nlp"},
	{"nlp", "dbmdz/bert-base-italian-cased", "1.0.0", "BERT Italian model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"nlp", "bert", "multilingual", "italian"}, "nlp"},
	{"nlp", "xlm-roberta-base", "1.0.0", "XLM-RoBERTa base model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "xlm", "roberta", "multilingual"}, "nlp"},
	{"nlp", "xlm-roberta-large", "1.0.0", "XLM-RoBERTa large model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"nlp", "xlm", "roberta", "multilingual"}, "nlp"},

	// Vision Models (40-69)
	{"vision", "resnet50", "1.0.0", "ResNet-50 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "resnet"}, "vision"},
	{"vision", "vit-base-patch16-224", "1.0.0", "Vision Transformer base", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "transformer", "vit"}, "vision"},
	{"vision", "yolov8n", "1.0.0", "YOLOv8 nano object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "resnet18", "1.0.0", "ResNet-18 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "resnet"}, "vision"},
	{"vision", "resnet34", "1.0.0", "ResNet-34 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "resnet"}, "vision"},
	{"vision", "resnet101", "1.0.0", "ResNet-101 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "resnet"}, "vision"},
	{"vision", "resnet152", "1.0.0", "ResNet-152 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "resnet"}, "vision"},
	{"vision", "mobilenet_v2", "1.0.0", "MobileNet V2 image classification", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "classification", "mobilenet"}, "vision"},
	{"vision", "efficientnet-b0", "1.0.0", "EfficientNet-B0 image classification", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "classification", "efficientnet"}, "vision"},
	{"vision", "efficientnet-b4", "1.0.0", "EfficientNet-B4 image classification", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "classification", "efficientnet"}, "vision"},
	{"vision", "densenet121", "1.0.0", "DenseNet-121 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "densenet"}, "vision"},
	{"vision", "densenet169", "1.0.0", "DenseNet-169 image classification", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "classification", "densenet"}, "vision"},
	{"vision", "yolov8s", "1.0.0", "YOLOv8 small object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov8m", "1.0.0", "YOLOv8 medium object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov8l", "1.0.0", "YOLOv8 large object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov8x", "1.0.0", "YOLOv8 extra large detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov5s", "1.0.0", "YOLOv5 small object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov5m", "1.0.0", "YOLOv5 medium object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov5l", "1.0.0", "YOLOv5 large object detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "yolov5x", "1.0.0", "YOLOv5 extra large detection", "PyTorch", "2.0.0", "pytorch", "AGPL-3.0", []string{"vision", "detection", "yolo"}, "vision"},
	{"vision", "detr-resnet-50", "1.0.0", "DETR object detection", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "detection", "detr"}, "vision"},
	{"vision", "detr-resnet-101", "1.0.0", "DETR ResNet-101 detection", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "detection", "detr"}, "vision"},
	{"vision", "faster-rcnn-resnet50-fpn", "1.0.0", "Faster R-CNN detection", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "detection", "faster-rcnn"}, "vision"},
	{"vision", "mask-rcnn-resnet50-fpn", "1.0.0", "Mask R-CNN segmentation", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "segmentation", "mask-rcnn"}, "vision"},
	{"vision", "retinanet-resnet50-fpn", "1.0.0", "RetinaNet object detection", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "detection", "retinanet"}, "vision"},
	{"vision", "deeplabv3-resnet50", "1.0.0", "DeepLabV3 semantic segmentation", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "segmentation", "deeplab"}, "vision"},
	{"vision", "deeplabv3-resnet101", "1.0.0", "DeepLabV3 ResNet-101", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "segmentation", "deeplab"}, "vision"},
	{"vision", "fcn-resnet50", "1.0.0", "FCN semantic segmentation", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "segmentation", "fcn"}, "vision"},
	{"vision", "lraspp-mobilenet-v3", "1.0.0", "LRASPP MobileNet segmentation", "PyTorch", "2.0.0", "pytorch", "BSD-3-Clause", []string{"vision", "segmentation", "mobilenet"}, "vision"},
	{"vision", "vit-large-patch16-224", "1.0.0", "Vision Transformer large", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"vision", "transformer", "vit"}, "vision"},

	// Audio Models (70-84)
	{"audio", "whisper-base", "1.0.0", "Whisper base speech recognition", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "transcription", "whisper"}, "audio"},
	{"audio", "whisper-small", "1.0.0", "Whisper small speech recognition", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "transcription", "whisper"}, "audio"},
	{"audio", "whisper-medium", "1.0.0", "Whisper medium speech recognition", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "transcription", "whisper"}, "audio"},
	{"audio", "whisper-large", "1.0.0", "Whisper large speech recognition", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "transcription", "whisper"}, "audio"},
	{"audio", "wav2vec2-base", "1.0.0", "Wav2Vec2 base speech model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "wav2vec2"}, "audio"},
	{"audio", "wav2vec2-large", "1.0.0", "Wav2Vec2 large speech model", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "wav2vec2"}, "audio"},
	{"audio", "facebook/wav2vec2-base-960h", "1.0.0", "Wav2Vec2 base 960h", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "wav2vec2"}, "audio"},
	{"audio", "facebook/wav2vec2-large-960h", "1.0.0", "Wav2Vec2 large 960h", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "wav2vec2"}, "audio"},
	{"audio", "facebook/wav2vec2-large-960h-lv60", "1.0.0", "Wav2Vec2 large 960h LV60", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "wav2vec2"}, "audio"},
	{"audio", "facebook/wav2vec2-xlsr-53", "1.0.0", "Wav2Vec2 XLSR-53 multilingual", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"audio", "speech", "multilingual"}, "audio"},
	{"audio", "facebook/hubert-base-ls960", "1.0.0", "HuBERT base speech model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "hubert"}, "audio"},
	{"audio", "facebook/hubert-large-ls960", "1.0.0", "HuBERT large speech model", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "hubert"}, "audio"},
	{"audio", "facebook/s2t-small-librispeech-asr", "1.0.0", "S2T small LibriSpeech ASR", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "asr"}, "audio"},
	{"audio", "facebook/s2t-medium-librispeech-asr", "1.0.0", "S2T medium LibriSpeech ASR", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "asr"}, "audio"},
	{"audio", "facebook/s2t-large-librispeech-asr", "1.0.0", "S2T large LibriSpeech ASR", "PyTorch", "2.0.0", "pytorch", "MIT", []string{"audio", "speech", "asr"}, "audio"},

	// Generation Models (85-99)
	{"generation", "stable-diffusion-2-1", "2.1.0", "Stable Diffusion 2.1 image generation", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image", "stable-diffusion"}, "generation"},
	{"generation", "stable-diffusion-v1-5", "1.5.0", "Stable Diffusion v1.5", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image"}, "generation"},
	{"generation", "runwayml/stable-diffusion-v1-5", "1.5.0", "Stable Diffusion v1.5 (Runway)", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image"}, "generation"},
	{"generation", "CompVis/stable-diffusion-v1-4", "1.4.0", "Stable Diffusion v1.4", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image"}, "generation"},
	{"generation", "stabilityai/stable-diffusion-2-1-base", "2.1.0", "Stable Diffusion 2.1 base", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image"}, "generation"},
	{"generation", "stabilityai/stable-diffusion-xl-base-1.0", "1.0.0", "Stable Diffusion XL base", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image", "xl"}, "generation"},
	{"generation", "stabilityai/stable-diffusion-xl-refiner-1.0", "1.0.0", "Stable Diffusion XL refiner", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "image", "xl"}, "generation"},
	{"generation", "runwayml/stable-diffusion-inpainting", "1.5.0", "Stable Diffusion inpainting", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "inpainting"}, "generation"},
	{"generation", "runwayml/stable-diffusion-depth2img", "1.5.0", "Stable Diffusion depth-to-image", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "depth"}, "generation"},
	{"generation", "lambdalabs/sd-image-variations-diffusers", "1.0.0", "SD image variations", "PyTorch", "2.0.0", "pytorch", "CreativeML Open RAIL-M", []string{"generation", "diffusion", "variations"}, "generation"},
	{"generation", "timbrooks/instruct-pix2pix", "1.0.0", "InstructPix2Pix image editing", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"generation", "editing", "pix2pix"}, "generation"},
	{"generation", "lllyasviel/controlnet-sd15", "1.0.0", "ControlNet for Stable Diffusion", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"generation", "controlnet", "diffusion"}, "generation"},
	{"generation", "lllyasviel/controlnet-sd15-canny", "1.0.0", "ControlNet Canny edge", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"generation", "controlnet", "canny"}, "generation"},
	{"generation", "lllyasviel/controlnet-sd15-depth", "1.0.0", "ControlNet depth", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"generation", "controlnet", "depth"}, "generation"},
	{"generation", "lllyasviel/controlnet-sd15-openpose", "1.0.0", "ControlNet OpenPose", "PyTorch", "2.0.0", "pytorch", "Apache-2.0", []string{"generation", "controlnet", "openpose"}, "generation"},
}

func main() {
	registryDir := "."
	if len(os.Args) > 1 {
		registryDir = os.Args[1]
	}

	now := time.Now()
	created := 0
	failed := 0

	for _, model := range top100Models {
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
					Size:   1024 * 1024 * 100,
					SHA256: "placeholder-sha256-checksum",
				},
				Registry: types.RegistryInfo{
					URL:       "http://localhost:8080",
					Namespace: model.namespace,
				},
			},
		}

		// Compute checksum for package file if it exists
		packageFilename := fmt.Sprintf("%s-%s-%s.axon", model.namespace, model.name, model.version)
		packagePath := filepath.Join(registryDir, "packages", packageFilename)
		var packageSHA256 string
		if _, err := os.Stat(packagePath); err == nil {
			file, err := os.Open(packagePath)
			if err == nil {
				hasher := sha256.New()
				if _, err := io.Copy(hasher, file); err == nil {
					packageSHA256 = hex.EncodeToString(hasher.Sum(nil))
				}
				file.Close()
			}
		}
		if packageSHA256 != "" {
			m.Distribution.Package.SHA256 = packageSHA256
		}

		// Create directory structure
		manifestDir := filepath.Join(registryDir, "api/v1/models", model.namespace, model.name, model.version)
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			fmt.Printf("❌ Error creating directory %s: %v\n", manifestDir, err)
			failed++
			continue
		}

		// Write manifest
		manifestPath := filepath.Join(manifestDir, "manifest.yaml")
		if err := manifest.Write(m, manifestPath); err != nil {
			fmt.Printf("❌ Error writing manifest for %s/%s: %v\n", model.namespace, model.name, err)
			failed++
			continue
		}

		created++
		if created%10 == 0 {
			fmt.Printf("✅ Created %d/%d manifests...\n", created, len(top100Models))
		}
	}

	fmt.Printf("\n✨ Generated %d manifests", created)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Printf("\n")
}

