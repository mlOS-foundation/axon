#!/usr/bin/env python3
"""
Hugging Face Model to ONNX Converter
Converts Hugging Face models to ONNX format using multiple strategies.

Usage:
    python3 convert_huggingface.py <model_path> <output_path> <model_id>

Arguments:
    model_path: Path to the model directory (from Axon cache)
    output_path: Where to save the converted ONNX file
    model_id: Axon model identifier (e.g., "hf/distilgpt2@latest")
"""

import sys
import os
import warnings

# Suppress warnings for cleaner output
warnings.filterwarnings('ignore')
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

def extract_hf_model_id(axon_model_id):
    """Extract Hugging Face model ID from Axon format."""
    # Format: "hf/model_name@version" -> "model_name"
    hf_model_id = axon_model_id
    if '/' in axon_model_id:
        hf_model_id = axon_model_id.split('/', 1)[1]
    if '@' in hf_model_id:
        hf_model_id = hf_model_id.split('@')[0]
    return hf_model_id

def try_optimum_export(model_path, output_path, hf_model_id):
    """
    Strategy 1: Use Optimum library (best for transformers models).
    This is the recommended approach for Hugging Face models.
    """
    try:
        from optimum.exporters.onnx import main_export
        from pathlib import Path
        
        print('🔄 Strategy 1: Trying Optimum ONNX export...')
        
        # Optimum expects output directory, not file
        output_dir = os.path.dirname(output_path) or '.'
        
        # Try loading from local path first
        model_name_or_path = model_path if os.path.isdir(model_path) else hf_model_id
        
        main_export(
            model_name_or_path=model_name_or_path,
            output=output_dir,
            task='auto',
            opset=14,
            device='cpu',
            fp16=False,
        )
        
        # Optimum may create model.onnx in output_dir, move if needed
        optimum_output = os.path.join(output_dir, 'model.onnx')
        if os.path.exists(optimum_output) and optimum_output != output_path:
            os.rename(optimum_output, output_path)
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (Optimum export)')
            return True
        
        return False
        
    except ImportError:
        print('⚠️  Optimum not available, skipping...')
        return False
    except Exception as e:
        print(f'⚠️  Optimum export failed: {str(e)}')
        return False

def try_torch_jit_trace(model, dummy_input, output_path):
    """
    Strategy 2: Use torch.jit.trace + ONNX export.
    Works well for models without dynamic control flow.
    """
    try:
        import torch
        
        print('🔄 Strategy 2: Trying torch.jit.trace + ONNX export...')
        
        # Trace the model
        traced_model = torch.jit.trace(model, dummy_input)
        
        # Export traced model to ONNX
        torch.onnx.export(
            traced_model,
            dummy_input,
            output_path,
            input_names=['input_ids'],
            output_names=['output'],
            dynamic_axes={'input_ids': {0: 'batch_size', 1: 'sequence_length'}, 
                         'output': {0: 'batch_size'}},
            opset_version=14,
            do_constant_folding=True,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (torch.jit.trace)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  torch.jit.trace failed: {str(e)}')
        return False

def try_legacy_torch_export(model, dummy_input, output_path, config):
    """
    Strategy 3: Use legacy torch.onnx.export with model wrapper.
    Handles models with complex outputs (tuples, dicts).
    """
    try:
        import torch
        
        print('🔄 Strategy 3: Trying legacy torch.onnx.export...')
        
        # Wrap model to handle complex outputs
        class ModelWrapper(torch.nn.Module):
            def __init__(self, model, config):
                super().__init__()
                self.model = model
                self.config = config
            
            def forward(self, input_ids):
                # Try different forward signatures based on model type
                try:
                    # For decoder models (GPT-2, etc.) - disable cache
                    if hasattr(self.config, 'is_decoder') and self.config.is_decoder:
                        output = self.model(input_ids, use_cache=False, return_dict=False)
                    else:
                        # For encoder models (BERT, etc.)
                        output = self.model(input_ids, return_dict=False)
                    
                    # Extract first element if tuple
                    if isinstance(output, tuple):
                        return output[0]
                    return output
                except:
                    # Fallback: simplest call
                    output = self.model(input_ids)
                    if isinstance(output, tuple):
                        return output[0]
                    return output
        
        wrapped_model = ModelWrapper(model, config)
        
        # Use legacy exporter (more compatible)
        torch.onnx.export(
            wrapped_model,
            dummy_input,
            output_path,
            input_names=['input_ids'],
            output_names=['logits'],
            dynamic_axes={'input_ids': {0: 'batch_size', 1: 'sequence_length'}, 
                         'logits': {0: 'batch_size'}},
            opset_version=14,
            do_constant_folding=True,
            export_params=True,
            verbose=False,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (legacy torch.onnx.export)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  Legacy export failed: {str(e)}')
        return False

def try_direct_onnx_export(model, dummy_input, output_path):
    """
    Strategy 4: Direct torch.onnx.export without wrapper.
    Simple approach for basic models.
    """
    try:
        import torch
        
        print('🔄 Strategy 4: Trying direct torch.onnx.export...')
        
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=['input_ids'],
            output_names=['output'],
            dynamic_axes={'input_ids': {0: 'batch_size', 1: 'sequence_length'}, 
                         'output': {0: 'batch_size'}},
            opset_version=14,
            do_constant_folding=True,
            export_params=True,
            verbose=False,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (direct export)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  Direct export failed: {str(e)}')
        return False

def convert_huggingface_to_onnx(model_path, output_path, axon_model_id):
    """Convert a Hugging Face model to ONNX using multiple strategies."""
    try:
        from transformers import AutoModel, AutoTokenizer
        import torch
        
        # Create output directory
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        # Extract HF model ID
        hf_model_id = extract_hf_model_id(axon_model_id)
        print(f'📦 Loading model: {hf_model_id} (Axon ID: {axon_model_id})')
        
        # Strategy 1: Try Optimum first (doesn't need model loading)
        if try_optimum_export(model_path, output_path, hf_model_id):
            return True
        
        # Load model for other strategies
        print('📥 Loading model with transformers...')
        try:
            if os.path.isdir(model_path):
                model = AutoModel.from_pretrained(model_path, local_files_only=True)
                tokenizer = AutoTokenizer.from_pretrained(model_path, local_files_only=True)
                print('   Loaded from local path')
            else:
                raise FileNotFoundError('Not a directory')
        except:
            print(f'   Loading from Hugging Face Hub: {hf_model_id}')
            model = AutoModel.from_pretrained(hf_model_id)
            tokenizer = AutoTokenizer.from_pretrained(hf_model_id)
        
        model.eval()
        
        # Prepare dummy input
        config = model.config
        seq_len = min(128, getattr(config, 'max_position_embeddings', 128))
        vocab_size = getattr(config, 'vocab_size', 30522)
        dummy_input = torch.randint(0, vocab_size, (1, seq_len))
        
        # Try remaining strategies in order
        strategies = [
            lambda: try_torch_jit_trace(model, dummy_input, output_path),
            lambda: try_legacy_torch_export(model, dummy_input, output_path, config),
            lambda: try_direct_onnx_export(model, dummy_input, output_path),
        ]
        
        for strategy in strategies:
            if strategy():
                return True
        
        # All strategies failed
        print('❌ ERROR: All conversion strategies failed')
        return False
        
    except ImportError as e:
        print(f'❌ ERROR: Missing dependency: {str(e)}')
        print('   Install with: pip install transformers torch optimum')
        return False
    except Exception as e:
        print(f'❌ ERROR: {str(e)}')
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: convert_huggingface.py <model_path> <output_path> <model_id>")
        sys.exit(1)
    
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    axon_model_id = sys.argv[3]
    
    success = convert_huggingface_to_onnx(model_path, output_path, axon_model_id)
    sys.exit(0 if success else 1)
