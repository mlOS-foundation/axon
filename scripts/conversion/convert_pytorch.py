#!/usr/bin/env python3
"""
PyTorch Model to ONNX Converter
Converts PyTorch models to ONNX format using multiple strategies.
Supports single models, multi-encoder models (CLIP), and proper NLP/Vision handling.

Usage:
    python3 convert_pytorch.py <model_path> <output_path> <model_id>

Arguments:
    model_path: Path to the model file or directory
    output_path: Where to save the converted ONNX file
    model_id: Axon model identifier (e.g., "pytorch/resnet50@latest")
"""

import sys
import os
import warnings

warnings.filterwarnings('ignore')

# Import shared utilities for multi-encoder support
try:
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    from convert_common import find_onnx_files, write_multi_encoder_manifest
except ImportError:
    # Fallback if convert_common not available
    def find_onnx_files(directory):
        onnx_files = []
        if os.path.isdir(directory):
            for f in os.listdir(directory):
                if f.endswith('.onnx'):
                    onnx_files.append(os.path.join(directory, f))
        return sorted(onnx_files)
    
    def write_multi_encoder_manifest(output_dir, onnx_files, task=None):
        import json
        file_names = [os.path.basename(f) for f in onnx_files]
        if 'text_model.onnx' in file_names and 'vision_model.onnx' in file_names:
            architecture, encoder_type = 'multi-encoder', 'clip'
            components = {'text_encoder': 'text_model.onnx', 'vision_encoder': 'vision_model.onnx'}
        elif 'encoder_model.onnx' in file_names and 'decoder_model.onnx' in file_names:
            architecture, encoder_type = 'encoder-decoder', 'seq2seq'
            components = {'encoder': 'encoder_model.onnx', 'decoder': 'decoder_model.onnx'}
            if 'decoder_with_past_model.onnx' in file_names:
                components['decoder_with_past'] = 'decoder_with_past_model.onnx'
        else:
            architecture, encoder_type = 'multi-model', 'unknown'
            components = {f'model_{i}': f for i, f in enumerate(file_names)}
        manifest = {'architecture': architecture, 'encoder_type': encoder_type, 'task': task or 'unknown',
                    'components': components, 'files': file_names}
        manifest_path = os.path.join(output_dir, 'onnx_manifest.json')
        with open(manifest_path, 'w') as f:
            json.dump(manifest, f, indent=2)
        print(f'   Created onnx_manifest.json for {architecture} model')
        return manifest_path

def extract_pytorch_model_id(axon_model_id):
    """Extract PyTorch Hub model ID from Axon format."""
    # Format: "pytorch/model_name@version" -> "model_name"
    model_id = axon_model_id
    if '/' in axon_model_id:
        model_id = axon_model_id.split('/', 1)[1]
    if '@' in model_id:
        model_id = model_id.split('@')[0]
    return model_id

def detect_model_type(model, model_name=""):
    """Detect if model is NLP, Vision, or Multi-modal based on structure."""
    import torch
    import torch.nn as nn
    
    model_type = "unknown"
    task = "unknown"
    
    # Check model class name
    class_name = model.__class__.__name__.lower()
    
    # Multi-modal detection (CLIP)
    if 'clip' in class_name or 'clip' in model_name.lower():
        model_type = "multimodal"
        task = "zero-shot-image-classification"
        return model_type, task
    
    # Vision model detection
    vision_indicators = ['resnet', 'vgg', 'alexnet', 'densenet', 'mobilenet', 
                        'efficientnet', 'vit', 'deit', 'convnext', 'swin',
                        'regnet', 'efficientnet', 'mnasnet', 'shufflenet']
    if any(indicator in class_name for indicator in vision_indicators):
        model_type = "vision"
        task = "image-classification"
        return model_type, task
    
    # NLP model detection
    nlp_indicators = ['bert', 'gpt', 'transformer', 'lstm', 'rnn', 'roberta',
                     'distilbert', 'albert', 'electra', 'xlnet', 't5', 'bart']
    if any(indicator in class_name for indicator in nlp_indicators):
        model_type = "nlp"
        # Try to detect specific task
        if 'gpt' in class_name or 'causal' in class_name:
            task = "text-generation"
        elif 'bert' in class_name or 'roberta' in class_name:
            task = "fill-mask"
        elif 't5' in class_name or 'bart' in class_name:
            task = "text2text-generation"
        else:
            task = "feature-extraction"
        return model_type, task
    
    # Check input signature by inspecting forward method
    try:
        import inspect
        sig = inspect.signature(model.forward)
        params = list(sig.parameters.keys())
        
        # Vision models typically take single tensor input
        if len(params) == 1 and params[0] in ['x', 'input', 'img', 'image']:
            model_type = "vision"
            task = "image-classification"
        # NLP models often take multiple inputs (input_ids, attention_mask, etc.)
        elif any('token' in p or 'input_ids' in p or 'attention' in p for p in params):
            model_type = "nlp"
            task = "feature-extraction"
    except:
        pass
    
    return model_type, task


def create_dummy_input(model, model_type, task):
    """Create appropriate dummy input based on model type."""
    import torch
    
    if model_type == "vision" or task == "image-classification":
        # Standard ImageNet input
        return torch.randn(1, 3, 224, 224), ['pixel_values'], {'pixel_values': {0: 'batch_size'}}
    
    elif model_type == "nlp":
        # NLP input - try to get vocab size from model if available
        vocab_size = 30522  # Default BERT vocab size
        seq_len = 128
        
        # Try to detect from model
        if hasattr(model, 'config'):
            vocab_size = getattr(model.config, 'vocab_size', vocab_size)
            seq_len = min(128, getattr(model.config, 'max_position_embeddings', seq_len))
        elif hasattr(model, 'vocab_size'):
            vocab_size = model.vocab_size
        
        if task == "text-generation":
            # Causal LM - single input_ids
            dummy_input = torch.randint(0, vocab_size, (1, seq_len))
            return dummy_input, ['input_ids'], {'input_ids': {0: 'batch_size', 1: 'sequence_length'}}
        else:
            # Most NLP models use input_ids
            dummy_input = torch.randint(0, vocab_size, (1, seq_len))
            return dummy_input, ['input_ids'], {'input_ids': {0: 'batch_size', 1: 'sequence_length'}}
    
    elif model_type == "multimodal" or task == "zero-shot-image-classification":
        # CLIP-style: both image and text
        pixel_values = torch.randn(1, 3, 224, 224)
        input_ids = torch.randint(0, 49408, (1, 77))  # CLIP vocab size
        attention_mask = torch.ones(1, 77, dtype=torch.long)
        return {
            'pixel_values': pixel_values,
            'input_ids': input_ids,
            'attention_mask': attention_mask
        }, ['pixel_values', 'input_ids', 'attention_mask'], {}
    
    # Default: vision input
    return torch.randn(1, 3, 224, 224), ['input'], {'input': {0: 'batch_size'}}


def try_optimum_export_pytorch(model_path, output_path, model_id):
    """Try Optimum export for PyTorch models (works for CLIP and other transformers-based models)."""
    try:
        from optimum.exporters.onnx import main_export
        
        print('🔄 Strategy 0: Trying Optimum ONNX export (for transformers-based models)...')
        
        output_dir = os.path.dirname(output_path) or '.'
        
        # Try to detect if this is a transformers model
        # Optimum works best with Hugging Face model IDs
        if '/' in model_id and ('pytorch' not in model_id.lower() or 'vision' in model_id.lower()):
            # Might be a Hugging Face model accessed via PyTorch Hub
            # Try using the model_id directly
            try:
                main_export(
                    model_name_or_path=model_id,
                    output=output_dir,
                    task='auto',
                    opset=14,
                    device='cpu',
                    fp16=False,
                )
                
                onnx_files = find_onnx_files(output_dir)
                if onnx_files:
                    if len(onnx_files) > 1:
                        print(f'✅ SUCCESS (Optimum export) - Multi-encoder model')
                        print(f'   Created {len(onnx_files)} ONNX files')
                        write_multi_encoder_manifest(output_dir, onnx_files, 'auto')
                        return True
                    elif len(onnx_files) == 1:
                        # Move to expected location
                        if onnx_files[0] != output_path:
                            os.rename(onnx_files[0], output_path)
                        print('✅ SUCCESS (Optimum export)')
                        return True
            except Exception as e:
                print(f'   Optimum export failed: {str(e)}')
                return False
        
        return False
    except ImportError:
        return False
    except Exception as e:
        print(f'   Optimum not available or failed: {str(e)}')
        return False


def try_pytorch_hub_load(model_id, output_path):
    """
    Strategy 1: Load from PyTorch Hub and export.
    Format: "repo/model" (e.g., "pytorch/vision:v0.10.0/resnet50")
    """
    try:
        import torch
        
        print(f'🔄 Strategy 1: Trying PyTorch Hub load: {model_id}')
        
        # Parse model_id for PyTorch Hub
        # Expected format: "owner/repo:tag/model" or "owner/repo/model"
        if '/' in model_id:
            parts = model_id.split('/')
            if len(parts) >= 2:
                repo = '/'.join(parts[:-1])
                model_name = parts[-1]
                
                print(f'   Loading from PyTorch Hub: repo={repo}, model={model_name}')
                model = torch.hub.load(repo, model_name, pretrained=True)
                model.eval()
                
                # Detect model type
                model_type, task = detect_model_type(model, model_name)
                print(f'   Detected model type: {model_type}, task: {task}')
                
                # Create appropriate dummy input
                dummy_input, input_names, dynamic_axes = create_dummy_input(model, model_type, task)
                
                # Handle multi-input models (dict)
                if isinstance(dummy_input, dict):
                    # For CLIP and other multi-input models, we need special handling
                    # Try Optimum first if it's a transformers model
                    if model_type == "multimodal":
                        print('   Multi-modal model detected - attempting Optimum export...')
                        if try_optimum_export_pytorch(None, output_path, model_id):
                            return True
                        print('   Optimum failed, trying manual export...')
                    
                    # Manual export for multi-input
                    output_dir = os.path.dirname(output_path) or '.'
                    # This is complex - for now, skip multi-input manual export
                    print('   ⚠️  Multi-input models require Optimum or specialized handling')
                    return False
                
                torch.onnx.export(
                    model,
                    dummy_input,
                    output_path,
                    input_names=input_names,
                    output_names=['output'],
                    dynamic_axes=dynamic_axes,
                    opset_version=14,
                    do_constant_folding=True,
                )
                
                if os.path.exists(output_path):
                    print('✅ SUCCESS (PyTorch Hub)')
                    return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  PyTorch Hub load failed: {str(e)}')
        return False

def try_load_torch_script(model_path, output_path):
    """
    Strategy 2: Load TorchScript model and export.
    """
    try:
        import torch
        
        print('🔄 Strategy 2: Trying TorchScript load...')
        
        model = torch.jit.load(model_path)
        model.eval()
        
        # Try to detect model type from path/name
        model_name = os.path.basename(model_path)
        model_type, task = detect_model_type(model, model_name)
        print(f'   Detected model type: {model_type}, task: {task}')
        
        # Create appropriate dummy input
        dummy_input, input_names, dynamic_axes = create_dummy_input(model, model_type, task)
        
        if isinstance(dummy_input, dict):
            print('   ⚠️  Multi-input TorchScript models need special handling')
            return False
        
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=input_names,
            output_names=['output'],
            dynamic_axes=dynamic_axes,
            opset_version=14,
            do_constant_folding=True,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (TorchScript)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  TorchScript load failed: {str(e)}')
        return False

def try_load_state_dict(model_path, output_path):
    """
    Strategy 3: Load model from state dict or checkpoint.
    """
    try:
        import torch
        
        print('🔄 Strategy 3: Trying state dict load...')
        
        # Load checkpoint
        checkpoint = torch.load(model_path, map_location='cpu')
        
        # Try to extract model from checkpoint
        if isinstance(checkpoint, torch.nn.Module):
            model = checkpoint
        elif isinstance(checkpoint, dict):
            if 'model' in checkpoint:
                model = checkpoint['model']
            elif 'state_dict' in checkpoint:
                # Need model architecture (not available without code)
                print('   State dict found but model architecture needed')
                return False
            else:
                print('   Unknown checkpoint format')
                return False
        else:
            print('   Unknown model format')
            return False
        
        model.eval()
        
        # Detect model type
        model_name = os.path.basename(model_path)
        model_type, task = detect_model_type(model, model_name)
        print(f'   Detected model type: {model_type}, task: {task}')
        
        # Create appropriate dummy input
        dummy_input, input_names, dynamic_axes = create_dummy_input(model, model_type, task)
        
        if isinstance(dummy_input, dict):
            print('   ⚠️  Multi-input models from checkpoint need special handling')
            return False
        
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=input_names,
            output_names=['output'],
            dynamic_axes=dynamic_axes,
            opset_version=14,
            do_constant_folding=True,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (state dict)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  State dict load failed: {str(e)}')
        return False

def try_torchvision_model(model_id, output_path):
    """
    Strategy 4: Load from torchvision.models.
    """
    try:
        import torch
        import torchvision.models as models
        
        print(f'🔄 Strategy 4: Trying torchvision.models: {model_id}')
        
        # Get model from torchvision
        if hasattr(models, model_id):
            model = getattr(models, model_id)(pretrained=True)
            model.eval()
            
            # Torchvision models are all vision models
            model_type, task = "vision", "image-classification"
            dummy_input, input_names, dynamic_axes = create_dummy_input(model, model_type, task)
            
            torch.onnx.export(
                model,
                dummy_input,
                output_path,
                input_names=input_names,
                output_names=['output'],
                dynamic_axes=dynamic_axes,
                opset_version=14,
                do_constant_folding=True,
            )
            
            if os.path.exists(output_path):
                print('✅ SUCCESS (torchvision)')
                return True
        
        return False
                
    except Exception as e:
        print(f'⚠️  torchvision load failed: {str(e)}')
        return False

def convert_pytorch_to_onnx(model_path, output_path, axon_model_id):
    """Convert a PyTorch model to ONNX using multiple strategies."""
    try:
        import torch
        
        # Create output directory
        output_dir = os.path.dirname(output_path) or '.'
        os.makedirs(output_dir, exist_ok=True)
        
        model_id = extract_pytorch_model_id(axon_model_id)
        print(f'📦 Converting PyTorch model: {model_id} (Axon ID: {axon_model_id})')
        print(f'   Model path: {model_path}')
        
        # Strategy 0: Try Optimum first (for transformers-based models like CLIP)
        # This handles multi-encoder models automatically
        if try_optimum_export_pytorch(model_path, output_path, model_id):
            # Check if multi-encoder files were created
            onnx_files = find_onnx_files(output_dir)
            if len(onnx_files) > 1:
                # Multi-encoder model - manifest already created by Optimum
                return True
            elif len(onnx_files) == 1 and os.path.exists(output_path):
                return True
        
        # Try strategies in order
        strategies = [
            lambda: try_pytorch_hub_load(model_id, output_path),
            lambda: try_torchvision_model(model_id, output_path),
        ]
        
        # If model_path exists, try file-based loading
        if os.path.exists(model_path):
            if os.path.isfile(model_path):
                # Try TorchScript or state dict
                strategies.insert(0, lambda: try_load_torch_script(model_path, output_path))
                strategies.insert(1, lambda: try_load_state_dict(model_path, output_path))
        
        for strategy in strategies:
            if strategy():
                return True
        
        # All strategies failed
        print('❌ ERROR: All conversion strategies failed')
        print('   PyTorch models require either:')
        print('   - PyTorch Hub identifier (repo/model)')
        print('   - TorchScript file (.pt)')
        print('   - Model checkpoint with architecture')
        print('   - torchvision model name')
        print('   - Transformers-based models (via Optimum)')
        return False
        
    except ImportError as e:
        print(f'❌ ERROR: Missing dependency: {str(e)}')
        print('   Install with: pip install torch torchvision optimum')
        return False
    except Exception as e:
        print(f'❌ ERROR: {str(e)}')
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: convert_pytorch.py <model_path> <output_path> <model_id>")
        sys.exit(1)
    
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    axon_model_id = sys.argv[3]
    
    success = convert_pytorch_to_onnx(model_path, output_path, axon_model_id)
    sys.exit(0 if success else 1)
