#!/usr/bin/env python3
"""
Hugging Face Model to ONNX Converter
Converts Hugging Face models to ONNX format using transformers and torch.

Usage:
    python3 convert_huggingface.py <model_path> <output_path> <model_id>

Arguments:
    model_path: Path to the model directory (from Axon cache)
    output_path: Where to save the converted ONNX file
    model_id: Hugging Face model identifier (e.g., "distilgpt2")
"""

import sys
import os

def convert_huggingface_to_onnx(model_path, output_path, model_id):
    """Convert a Hugging Face model to ONNX format."""
    try:
        from transformers import AutoModel, AutoTokenizer
        import torch
        
        # Create output directory if needed
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        print(f'Loading model: {model_id}')
        
        # Try loading from local path first, then from Hugging Face Hub
        try:
            if os.path.isdir(model_path):
                model = AutoModel.from_pretrained(model_path, local_files_only=True)
                tokenizer = AutoTokenizer.from_pretrained(model_path, local_files_only=True)
                print('Loaded from local path')
            else:
                raise FileNotFoundError('Not a directory')
        except Exception as e:
            print(f'Local load failed, trying Hugging Face Hub: {str(e)}')
            model = AutoModel.from_pretrained(model_id)
            tokenizer = AutoTokenizer.from_pretrained(model_id)
            print('Loaded from Hugging Face Hub')
        
        model.eval()
        
        # Get model config for input shape
        config = model.config
        seq_len = min(128, getattr(config, 'max_position_embeddings', 128))
        vocab_size = getattr(config, 'vocab_size', 30522)
        
        # Create dummy input
        dummy_input = torch.randint(0, vocab_size, (1, seq_len))
        
        # Export to ONNX
        print(f'Exporting to ONNX: {output_path}')
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=['input_ids'],
            output_names=['output'],
            dynamic_axes={'input_ids': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
            opset_version=12,
            do_constant_folding=True
        )
        
        print('SUCCESS')
        return True
        
    except ImportError as e:
        print(f'ERROR: Missing dependency: {str(e)}')
        print('Install with: pip install transformers torch')
        sys.exit(1)
    except Exception as e:
        print(f'ERROR: {str(e)}')
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: convert_huggingface.py <model_path> <output_path> <model_id>")
        sys.exit(1)
    
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    model_id = sys.argv[3]
    
    success = convert_huggingface_to_onnx(model_path, output_path, model_id)
    sys.exit(0 if success else 1)


