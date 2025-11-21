#!/usr/bin/env python3
"""
PyTorch Model to ONNX Converter
Converts PyTorch models to ONNX format using multiple strategies.

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

def extract_pytorch_model_id(axon_model_id):
    """Extract PyTorch Hub model ID from Axon format."""
    # Format: "pytorch/model_name@version" -> "model_name"
    model_id = axon_model_id
    if '/' in axon_model_id:
        model_id = axon_model_id.split('/', 1)[1]
    if '@' in model_id:
        model_id = model_id.split('@')[0]
    return model_id

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
                
                # Use appropriate dummy input based on model type
                # Default: image input (224x224)
                dummy_input = torch.randn(1, 3, 224, 224)
                
                torch.onnx.export(
                    model,
                    dummy_input,
                    output_path,
                    input_names=['input'],
                    output_names=['output'],
                    dynamic_axes={'input': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
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
        
        # Default dummy input for vision models
        dummy_input = torch.randn(1, 3, 224, 224)
        
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=['input'],
            output_names=['output'],
            dynamic_axes={'input': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
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
        
        # Default dummy input
        dummy_input = torch.randn(1, 3, 224, 224)
        
        torch.onnx.export(
            model,
            dummy_input,
            output_path,
            input_names=['input'],
            output_names=['output'],
            dynamic_axes={'input': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
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
            
            dummy_input = torch.randn(1, 3, 224, 224)
            
            torch.onnx.export(
                model,
                dummy_input,
                output_path,
                input_names=['input'],
                output_names=['output'],
                dynamic_axes={'input': {0: 'batch_size'}, 'output': {0: 'batch_size'}},
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
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        model_id = extract_pytorch_model_id(axon_model_id)
        print(f'📦 Converting PyTorch model: {model_id} (Axon ID: {axon_model_id})')
        print(f'   Model path: {model_path}')
        
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
        return False
        
    except ImportError as e:
        print(f'❌ ERROR: Missing dependency: {str(e)}')
        print('   Install with: pip install torch torchvision')
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
