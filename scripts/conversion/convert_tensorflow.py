#!/usr/bin/env python3
"""
TensorFlow Model to ONNX Converter
Converts TensorFlow models to ONNX format using tf2onnx and multiple strategies.

Usage:
    python3 convert_tensorflow.py <model_path> <output_path> <model_id>

Arguments:
    model_path: Path to the model file or directory (SavedModel, H5, etc.)
    output_path: Where to save the converted ONNX file
    model_id: Axon model identifier (e.g., "tfhub/model_name@latest")
"""

import sys
import os
import warnings

warnings.filterwarnings('ignore')
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

def extract_tf_model_id(axon_model_id):
    """Extract TensorFlow Hub model ID from Axon format."""
    # Format: "tfhub/model_name@version" -> "model_name"
    model_id = axon_model_id
    if '/' in axon_model_id:
        model_id = axon_model_id.split('/', 1)[1]
    if '@' in model_id:
        model_id = model_id.split('@')[0]
    return model_id

def try_saved_model_conversion(model_path, output_path):
    """
    Strategy 1: Convert TensorFlow SavedModel format.
    This is the most common format for TensorFlow models.
    """
    try:
        import tf2onnx
        import tensorflow as tf
        
        print('🔄 Strategy 1: Trying SavedModel conversion...')
        
        if not os.path.isdir(model_path):
            print('   Not a directory, skipping SavedModel')
            return False
        
        # Check if it's a valid SavedModel
        if not os.path.exists(os.path.join(model_path, 'saved_model.pb')):
            print('   No saved_model.pb found')
            return False
        
        # Convert using tf2onnx
        print(f'   Converting SavedModel to ONNX...')
        model_proto, external_tensor_storage = tf2onnx.convert.from_saved_model(
            model_path,
            opset=14,
            output_path=output_path,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (SavedModel)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  SavedModel conversion failed: {str(e)}')
        return False

def try_keras_h5_conversion(model_path, output_path):
    """
    Strategy 2: Convert Keras H5 model.
    """
    try:
        import tf2onnx
        import tensorflow as tf
        
        print('🔄 Strategy 2: Trying Keras H5 conversion...')
        
        if not model_path.endswith('.h5') and not model_path.endswith('.hdf5'):
            print('   Not an H5 file, skipping')
            return False
        
        if not os.path.isfile(model_path):
            print('   File not found')
            return False
        
        # Load Keras model
        print(f'   Loading Keras model from {model_path}...')
        model = tf.keras.models.load_model(model_path)
        
        # Convert to ONNX
        print('   Converting to ONNX...')
        spec = (tf.TensorSpec(model.input_shape, tf.float32, name="input"),)
        model_proto, external_tensor_storage = tf2onnx.convert.from_keras(
            model,
            input_signature=spec,
            opset=14,
            output_path=output_path,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (Keras H5)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  Keras H5 conversion failed: {str(e)}')
        return False

def try_keras_model_conversion(model_path, output_path):
    """
    Strategy 3: Convert Keras model directory.
    """
    try:
        import tf2onnx
        import tensorflow as tf
        
        print('🔄 Strategy 3: Trying Keras model directory conversion...')
        
        if not os.path.isdir(model_path):
            print('   Not a directory, skipping')
            return False
        
        # Try loading as Keras model
        print(f'   Loading Keras model from directory...')
        model = tf.keras.models.load_model(model_path)
        
        # Convert to ONNX
        print('   Converting to ONNX...')
        spec = (tf.TensorSpec(model.input_shape, tf.float32, name="input"),)
        model_proto, external_tensor_storage = tf2onnx.convert.from_keras(
            model,
            input_signature=spec,
            opset=14,
            output_path=output_path,
        )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (Keras model directory)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  Keras model directory conversion failed: {str(e)}')
        return False

def try_tfhub_model(model_id, output_path):
    """
    Strategy 4: Load from TensorFlow Hub and convert.
    """
    try:
        import tf2onnx
        import tensorflow as tf
        import tensorflow_hub as hub
        
        print(f'🔄 Strategy 4: Trying TensorFlow Hub: {model_id}')
        
        # Construct TensorFlow Hub URL
        if not model_id.startswith('https://'):
            # Assume it's a short name
            hub_url = f'https://tfhub.dev/{model_id}'
        else:
            hub_url = model_id
        
        print(f'   Loading from TensorFlow Hub: {hub_url}')
        model = hub.load(hub_url)
        
        # Try to get signatures
        if hasattr(model, 'signatures'):
            signature_key = list(model.signatures.keys())[0]
            func = model.signatures[signature_key]
            
            # Convert using tf2onnx
            print('   Converting to ONNX...')
            model_proto, external_tensor_storage = tf2onnx.convert.from_function(
                func,
                opset=14,
                output_path=output_path,
            )
            
            if os.path.exists(output_path):
                print('✅ SUCCESS (TensorFlow Hub)')
                return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  TensorFlow Hub conversion failed: {str(e)}')
        return False

def convert_tensorflow_to_onnx(model_path, output_path, axon_model_id):
    """Convert a TensorFlow model to ONNX using multiple strategies."""
    try:
        import tf2onnx
        import tensorflow as tf
        
        # Create output directory
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        model_id = extract_tf_model_id(axon_model_id)
        print(f'📦 Converting TensorFlow model: {model_id} (Axon ID: {axon_model_id})')
        print(f'   Model path: {model_path}')
        
        # Try strategies in order
        strategies = []
        
        # If model_path exists, try file-based conversions first
        if os.path.exists(model_path):
        if os.path.isdir(model_path):
                strategies.append(lambda: try_saved_model_conversion(model_path, output_path))
                strategies.append(lambda: try_keras_model_conversion(model_path, output_path))
            elif os.path.isfile(model_path):
                strategies.append(lambda: try_keras_h5_conversion(model_path, output_path))
        
        # Try TensorFlow Hub
        strategies.append(lambda: try_tfhub_model(model_id, output_path))
        
        for strategy in strategies:
            if strategy():
                return True
        
        # All strategies failed
        print('❌ ERROR: All conversion strategies failed')
        print('   TensorFlow models require:')
        print('   - SavedModel directory (with saved_model.pb)')
        print('   - Keras H5 file (.h5 or .hdf5)')
        print('   - Keras model directory')
        print('   - TensorFlow Hub model identifier')
        return False
        
    except ImportError as e:
        print(f'❌ ERROR: Missing dependency: {str(e)}')
        print('   Install with: pip install tf2onnx tensorflow tensorflow-hub')
        return False
    except Exception as e:
        print(f'❌ ERROR: {str(e)}')
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: convert_tensorflow.py <model_path> <output_path> <model_id>")
        sys.exit(1)
    
    model_path = sys.argv[1]
    output_path = sys.argv[2]
    axon_model_id = sys.argv[3]
    
    success = convert_tensorflow_to_onnx(model_path, output_path, axon_model_id)
    sys.exit(0 if success else 1)
