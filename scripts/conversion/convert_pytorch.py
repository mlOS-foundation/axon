#!/usr/bin/env python3
"""
PyTorch Model to ONNX Converter
Converts PyTorch models to ONNX format.

Usage:
    python3 convert_pytorch.py <model_path> <output_path> <model_id>
"""

import sys
import os

def convert_pytorch_to_onnx(model_path, output_path, model_id):
    """Convert a PyTorch model to ONNX format."""
    try:
        import torch
        
        # Create output directory if needed
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        print(f'Loading PyTorch model: {model_path}')
        
        # Try to load as PyTorch model
        if os.path.isdir(model_path):
            print('ERROR: Directory-based PyTorch models need specific loading code')
            sys.exit(1)
        else:
            model = torch.load(model_path, map_location='cpu')
            if isinstance(model, torch.nn.Module):
                model.eval()
                dummy_input = torch.randn(1, 3, 224, 224)
                torch.onnx.export(model, dummy_input, output_path, opset_version=12)
                print('SUCCESS')
                return True
            else:
                print('ERROR: Model file is not a PyTorch Module')
                sys.exit(1)
                
    except Exception as e:
        print(f'ERROR: {str(e)}')
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: convert_pytorch.py <model_path> <output_path> <model_id>")
        sys.exit(1)
    
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    model_id = sys.argv[3]
    
    success = convert_pytorch_to_onnx(model_path, output_path, model_id)
    sys.exit(0 if success else 1)


