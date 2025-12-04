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

Supported Model Types:
    Vision (Image Classification):
        ResNet, ViT, DeiT, BEiT, Swin, ConvNeXt, EfficientNet, VGG,
        MobileNet, MobileViT, RegNet, DINOv2, PoolFormer, LeViT,
        CvT, FocalNet, BiT, Data2Vec Vision, NAT, PVT, VAN, DenseNet
    
    Vision (Object Detection):
        DETR, YOLOS, Conditional DETR, Deformable DETR, DETA, RT-DETR
    
    Vision (Segmentation):
        SegFormer, MaskFormer, Mask2Former, UperNet, OneFormer
    
    Vision (Depth Estimation):
        DPT, GLPN, Depth Anything
    
    NLP (Text Generation):
        GPT-2, GPT-Neo, GPT-NeoX, GPT-J, Llama, Mistral, Phi-3, OPT, Bloom
    
    NLP (Fill Mask):
        BERT, RoBERTa, DistilBERT, ALBERT, ELECTRA, DeBERTa, XLM-RoBERTa
    
    NLP (Text2Text):
        T5, BART, MT5, MBart, Pegasus
    
    Multimodal:
        CLIP, BLIP, BLIP-2
    
    Audio:
        Wav2Vec2, Whisper, Hubert

Conversion Strategies (tried in order):
    1. Optimum ONNX export (recommended, handles most models)
    2. torch.onnx.export (fallback for unsupported models)
    3. torch.jit.trace + ONNX (last resort)
"""

import sys
import os
import json
import warnings

# Suppress warnings for cleaner output
warnings.filterwarnings('ignore')
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'

# Task mapping from model architecture/config to Optimum task
# This is used when task='auto' fails (especially for local directories)
# Comprehensive mapping for vision, NLP, audio, and multimodal models
TASK_MAPPING = {
    # ═══════════════════════════════════════════════════════════════════════
    # VISION MODELS - Image Classification
    # ═══════════════════════════════════════════════════════════════════════
    # ResNet family
    'ResNetConfig': 'image-classification',
    'ResNetForImageClassification': 'image-classification',
    
    # Vision Transformer (ViT) family
    'ViTConfig': 'image-classification',
    'ViTForImageClassification': 'image-classification',
    'ViTMAEConfig': 'image-classification',
    'ViTMSNConfig': 'image-classification',
    
    # DeiT (Data-efficient Image Transformer)
    'DeiTConfig': 'image-classification',
    'DeiTForImageClassification': 'image-classification',
    
    # BEiT (BERT pre-training of Image Transformers)
    'BeitConfig': 'image-classification',
    'BeitForImageClassification': 'image-classification',
    
    # Swin Transformer family
    'SwinConfig': 'image-classification',
    'SwinForImageClassification': 'image-classification',
    'Swinv2Config': 'image-classification',
    'Swinv2ForImageClassification': 'image-classification',
    
    # ConvNeXt family
    'ConvNextConfig': 'image-classification',
    'ConvNextForImageClassification': 'image-classification',
    'ConvNextV2Config': 'image-classification',
    'ConvNextV2ForImageClassification': 'image-classification',
    
    # EfficientNet
    'EfficientNetConfig': 'image-classification',
    'EfficientNetForImageClassification': 'image-classification',
    
    # VGG
    'VGGConfig': 'image-classification',
    
    # MobileNet family
    'MobileNetV1Config': 'image-classification',
    'MobileNetV2Config': 'image-classification',
    'MobileNetV2ForImageClassification': 'image-classification',
    'MobileViTConfig': 'image-classification',
    'MobileViTForImageClassification': 'image-classification',
    'MobileViTV2Config': 'image-classification',
    
    # RegNet
    'RegNetConfig': 'image-classification',
    'RegNetForImageClassification': 'image-classification',
    
    # DINOv2
    'Dinov2Config': 'image-classification',
    'Dinov2ForImageClassification': 'image-classification',
    
    # PoolFormer
    'PoolFormerConfig': 'image-classification',
    'PoolFormerForImageClassification': 'image-classification',
    
    # LeViT
    'LevitConfig': 'image-classification',
    'LevitForImageClassification': 'image-classification',
    
    # CvT (Convolutional Vision Transformer)
    'CvtConfig': 'image-classification',
    'CvtForImageClassification': 'image-classification',
    
    # FocalNet
    'FocalNetConfig': 'image-classification',
    
    # BiT (Big Transfer)
    'BitConfig': 'image-classification',
    'BitForImageClassification': 'image-classification',
    
    # Data2Vec Vision
    'Data2VecVisionConfig': 'image-classification',
    'Data2VecVisionForImageClassification': 'image-classification',
    
    # NAT (Neighborhood Attention Transformer)
    'NatConfig': 'image-classification',
    'NatForImageClassification': 'image-classification',
    'DinatConfig': 'image-classification',
    
    # PVT (Pyramid Vision Transformer)
    'PvtConfig': 'image-classification',
    'PvtV2Config': 'image-classification',
    
    # VAN (Visual Attention Network)
    'VanConfig': 'image-classification',
    
    # DenseNet
    'DenseNetConfig': 'image-classification',
    
    # Inception
    'InceptionV3Config': 'image-classification',
    
    # ═══════════════════════════════════════════════════════════════════════
    # VISION MODELS - Object Detection
    # ═══════════════════════════════════════════════════════════════════════
    'DetrConfig': 'object-detection',
    'DetrForObjectDetection': 'object-detection',
    'YolosConfig': 'object-detection',
    'YolosForObjectDetection': 'object-detection',
    'ConditionalDetrConfig': 'object-detection',
    'DeformableDetrConfig': 'object-detection',
    'DetaConfig': 'object-detection',
    'GroundingDinoConfig': 'object-detection',
    'RTDetrConfig': 'object-detection',
    
    # ═══════════════════════════════════════════════════════════════════════
    # VISION MODELS - Semantic/Instance Segmentation
    # ═══════════════════════════════════════════════════════════════════════
    'SegformerConfig': 'semantic-segmentation',
    'SegformerForSemanticSegmentation': 'semantic-segmentation',
    'MaskFormerConfig': 'semantic-segmentation',
    'Mask2FormerConfig': 'semantic-segmentation',
    'UperNetConfig': 'semantic-segmentation',
    'OneFormerConfig': 'semantic-segmentation',
    
    # ═══════════════════════════════════════════════════════════════════════
    # VISION MODELS - Depth Estimation
    # ═══════════════════════════════════════════════════════════════════════
    'DPTConfig': 'depth-estimation',
    'GLPNConfig': 'depth-estimation',
    'DepthAnythingConfig': 'depth-estimation',
    
    # NLP - Text Generation
    'GPT2Config': 'text-generation',
    'GPTNeoConfig': 'text-generation',
    'GPTNeoXConfig': 'text-generation',
    'GPTJConfig': 'text-generation',
    'LlamaConfig': 'text-generation',
    'MistralConfig': 'text-generation',
    'Phi3Config': 'text-generation',
    'OPTConfig': 'text-generation',
    'BloomConfig': 'text-generation',
    
    # NLP - Fill Mask (MLM)
    'BertConfig': 'fill-mask',
    'RobertaConfig': 'fill-mask',
    'DistilBertConfig': 'fill-mask',
    'AlbertConfig': 'fill-mask',
    'ElectraConfig': 'fill-mask',
    'CamembertConfig': 'fill-mask',
    'XLMRobertaConfig': 'fill-mask',
    'DebertaConfig': 'fill-mask',
    'DebertaV2Config': 'fill-mask',
    
    # NLP - Text2Text Generation (Encoder-Decoder)
    'T5Config': 'text2text-generation',
    'BartConfig': 'text2text-generation',
    'MT5Config': 'text2text-generation',
    'MBartConfig': 'text2text-generation',
    'PegasusConfig': 'text2text-generation',
    
    # NLP - Sequence Classification
    'XLNetConfig': 'text-classification',
    
    # NLP - Feature Extraction (generic)
    'MPNetConfig': 'feature-extraction',
    'SentenceTransformersConfig': 'feature-extraction',
    
    # Multi-Modal
    'CLIPConfig': 'zero-shot-image-classification',
    'CLIPModel': 'zero-shot-image-classification',  # Architecture name
    'CLIPTextConfig': 'feature-extraction',
    'CLIPVisionConfig': 'image-classification',
    'BlipConfig': 'image-to-text',
    'Blip2Config': 'image-to-text',
    
    # Audio
    'Wav2Vec2Config': 'automatic-speech-recognition',
    'WhisperConfig': 'automatic-speech-recognition',
    'HubertConfig': 'automatic-speech-recognition',
}

# Model class mapping for proper loading
# Maps Optimum task names to the correct AutoModel class
MODEL_CLASS_MAPPING = {
    # ═══════════════════════════════════════════════════════════════════════
    # Vision models - need specific classes, not AutoModel
    # ═══════════════════════════════════════════════════════════════════════
    'image-classification': 'AutoModelForImageClassification',
    'object-detection': 'AutoModelForObjectDetection',
    'semantic-segmentation': 'AutoModelForSemanticSegmentation',
    'depth-estimation': 'AutoModelForDepthEstimation',
    'zero-shot-image-classification': 'AutoModel',
    'image-to-text': 'AutoModelForVision2Seq',
    
    # ═══════════════════════════════════════════════════════════════════════
    # NLP models
    # ═══════════════════════════════════════════════════════════════════════
    'text-generation': 'AutoModelForCausalLM',
    'fill-mask': 'AutoModelForMaskedLM',
    'text2text-generation': 'AutoModelForSeq2SeqLM',
    'text-classification': 'AutoModelForSequenceClassification',
    'token-classification': 'AutoModelForTokenClassification',
    'question-answering': 'AutoModelForQuestionAnswering',
    'feature-extraction': 'AutoModel',
    
    # ═══════════════════════════════════════════════════════════════════════
    # Audio models
    # ═══════════════════════════════════════════════════════════════════════
    'automatic-speech-recognition': 'AutoModelForSpeechSeq2Seq',
    'audio-classification': 'AutoModelForAudioClassification',
    
    # Default fallback
    'default': 'AutoModel',
}


def extract_hf_model_id(axon_model_id):
    """Extract Hugging Face model ID from Axon format."""
    # The Go code already strips the 'hf/' prefix, so we receive:
    # - "microsoft/resnet-50" (org/model format)
    # - "distilgpt2" (simple model name)
    # We should only strip 'hf/' prefix if present (legacy format)
    hf_model_id = axon_model_id
    
    # Only strip if it explicitly starts with 'hf/'
    if hf_model_id.startswith('hf/'):
        hf_model_id = hf_model_id[3:]  # Remove 'hf/' prefix
    
    # Strip version if present
    if '@' in hf_model_id:
        hf_model_id = hf_model_id.split('@')[0]
    
    return hf_model_id


def detect_task_from_config(model_path):
    """
    Detect the appropriate task from model config file.
    Returns task string or None if cannot be determined.
    """
    config_path = os.path.join(model_path, 'config.json')
    
    if not os.path.exists(config_path):
        print(f'   No config.json found at {model_path}')
        return None
    
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)
        
        # Check for architectures field first (most reliable)
        architectures = config.get('architectures', [])
        for arch in architectures:
            if arch in TASK_MAPPING:
                task = TASK_MAPPING[arch]
                print(f'   Detected task from architecture: {arch} → {task}')
                return task
        
        # Check for model_type field
        model_type = config.get('model_type', '')
        # Handle special cases (CLIP, etc.) that need uppercase
        if model_type.lower() == 'clip':
            config_class_name = 'CLIPConfig'
        else:
            config_class_name = f"{model_type.title().replace('-', '').replace('_', '')}Config"
        
        if config_class_name in TASK_MAPPING:
            task = TASK_MAPPING[config_class_name]
            print(f'   Detected task from model_type: {model_type} → {task}')
            return task
        
        # Check for auto_map which may have the class name
        auto_map = config.get('auto_map', {})
        for key, value in auto_map.items():
            # value might be like "modeling_mymodel.MyModelForImageClassification"
            class_name = value.split('.')[-1] if '.' in value else value
            if class_name in TASK_MAPPING:
                task = TASK_MAPPING[class_name]
                print(f'   Detected task from auto_map: {class_name} → {task}')
                return task
        
        print(f'   Could not detect task from config (architectures: {architectures}, model_type: {model_type})')
        return None
        
    except Exception as e:
        print(f'   Error reading config.json: {e}')
        return None


def get_model_class(task):
    """Get the appropriate model class for a given task."""
    class_name = MODEL_CLASS_MAPPING.get(task, MODEL_CLASS_MAPPING['default'])
    return class_name


def try_optimum_export(model_path, output_path, hf_model_id, task=None):
    """
    Strategy 1: Use Optimum library (best for transformers models).
    This is the recommended approach for Hugging Face models.
    
    Handles multi-encoder models (CLIP, T5, etc.) that export multiple ONNX files:
    - CLIP: text_model.onnx, vision_model.onnx
    - T5/BART: encoder_model.onnx, decoder_model.onnx, decoder_with_past_model.onnx
    """
    try:
        from optimum.exporters.onnx import main_export
        from pathlib import Path
        
        print('🔄 Strategy 1: Trying Optimum ONNX export...')
        
        # Optimum expects output directory, not file
        output_dir = os.path.dirname(output_path) or '.'
        
        # Try loading from local path first
        model_name_or_path = model_path if os.path.isdir(model_path) else hf_model_id
        
        # Detect task if not provided
        if task is None or task == 'auto':
            detected_task = detect_task_from_config(model_path)
            if detected_task:
                task = detected_task
            else:
                # For CLIP and other multi-modal models, try to infer from model ID
                if 'clip' in hf_model_id.lower():
                    task = 'zero-shot-image-classification'
                    print(f'   Inferred CLIP task from model ID: {task}')
                else:
                    # Fallback to 'auto' and let Optimum try
                    task = 'auto'
        
        # For Optimum export with specific tasks, prefer using Hugging Face model ID
        # Optimum can't always infer task from local directories, especially for multi-modal models
        if task != 'auto' and task is not None:
            # For multi-modal models like CLIP, always use model ID (Optimum requirement)
            if task in ['zero-shot-image-classification', 'image-to-text', 'image-text-to-text']:
                print(f'   Using Hugging Face model ID for {task} (Optimum requirement for multi-modal)')
                model_name_or_path = hf_model_id
            # For other tasks, try local path first, but Optimum may still need model ID
            # We'll let Optimum handle it and fall back if needed
        
        print(f'   Using task: {task}')
        print(f'   Model path: {model_name_or_path}')
        
        main_export(
            model_name_or_path=model_name_or_path,
            output=output_dir,
            task=task,
            opset=14,
            device='cpu',
            fp16=False,
        )
        
        # Check what ONNX files were created
        onnx_files = find_onnx_files(output_dir)
        
        if not onnx_files:
            print('⚠️  No ONNX files created by Optimum')
            return False
        
        # Handle multi-encoder models
        if len(onnx_files) > 1:
            print(f'✅ SUCCESS (Optimum export) - Multi-encoder model')
            print(f'   Created {len(onnx_files)} ONNX files:')
            for f in onnx_files:
                print(f'     - {os.path.basename(f)}')
            # Write model manifest for multi-encoder
            write_multi_encoder_manifest(output_dir, onnx_files, task)
            return True
        
        # Single model - move to expected location if needed
        if len(onnx_files) == 1:
            created_file = onnx_files[0]
            if created_file != output_path:
                # Rename to model.onnx if it's not already
                if os.path.basename(created_file) != 'model.onnx':
                    new_path = os.path.join(output_dir, 'model.onnx')
                    os.rename(created_file, new_path)
                    created_file = new_path
                if created_file != output_path:
                    os.rename(created_file, output_path)
            print('✅ SUCCESS (Optimum export)')
            return True
        
        return False
        
    except ImportError:
        print('⚠️  Optimum not available, skipping...')
        return False
    except Exception as e:
        print(f'⚠️  Optimum export failed: {str(e)}')
        return False


def find_onnx_files(directory):
    """Find all ONNX files in a directory."""
    onnx_files = []
    for f in os.listdir(directory):
        if f.endswith('.onnx'):
            onnx_files.append(os.path.join(directory, f))
    return sorted(onnx_files)


def write_multi_encoder_manifest(output_dir, onnx_files, task):
    """
    Write a manifest file for multi-encoder models.
    This helps Core understand how to load and orchestrate multiple ONNX files.
    """
    import json
    
    # Determine model architecture type
    file_names = [os.path.basename(f) for f in onnx_files]
    
    if 'text_model.onnx' in file_names and 'vision_model.onnx' in file_names:
        architecture = 'multi-encoder'
        encoder_type = 'clip'
        components = {
            'text_encoder': 'text_model.onnx',
            'vision_encoder': 'vision_model.onnx'
        }
    elif 'encoder_model.onnx' in file_names and 'decoder_model.onnx' in file_names:
        architecture = 'encoder-decoder'
        encoder_type = 'seq2seq'
        components = {
            'encoder': 'encoder_model.onnx',
            'decoder': 'decoder_model.onnx'
        }
        if 'decoder_with_past_model.onnx' in file_names:
            components['decoder_with_past'] = 'decoder_with_past_model.onnx'
    else:
        architecture = 'multi-model'
        encoder_type = 'unknown'
        components = {f'model_{i}': f for i, f in enumerate(file_names)}
    
    manifest = {
        'architecture': architecture,
        'encoder_type': encoder_type,
        'task': task,
        'components': components,
        'files': file_names
    }
    
    manifest_path = os.path.join(output_dir, 'onnx_manifest.json')
    with open(manifest_path, 'w') as f:
        json.dump(manifest, f, indent=2)
    
    print(f'   Created onnx_manifest.json for {architecture} model')


def load_model_and_tokenizer(model_path, hf_model_id, task=None):
    """
    Load model and tokenizer/processor with proper class based on task.
    Includes fallback for safetensors UTF-8 issues.
    """
    from transformers import AutoTokenizer, AutoImageProcessor
    import transformers
    
    # Detect task if not provided
    if task is None:
        task = detect_task_from_config(model_path)
    
    # Get appropriate model class
    model_class_name = get_model_class(task)
    print(f'   Using model class: {model_class_name} (task: {task})')
    
    # Get the model class from transformers
    model_class = getattr(transformers, model_class_name, None)
    if model_class is None:
        print(f'   Model class {model_class_name} not found, falling back to AutoModel')
        from transformers import AutoModel
        model_class = AutoModel
    
    # Determine if this is a vision model
    is_vision_task = task in [
        'image-classification', 
        'object-detection', 
        'semantic-segmentation',
        'depth-estimation',
        'zero-shot-image-classification',
        'image-to-text',
    ]
    
    # Try loading with different configurations
    load_configs = [
        {'local_files_only': True},  # Try local first
        {'local_files_only': True, 'use_safetensors': False},  # Try without safetensors
        {},  # Try from Hub
        {'use_safetensors': False},  # Try from Hub without safetensors
    ]
    
    model = None
    last_error = None
    
    for config in load_configs:
        try:
            # First try local path
            load_path = model_path if os.path.isdir(model_path) else hf_model_id
            if 'local_files_only' not in config and os.path.isdir(model_path):
                load_path = hf_model_id  # Force Hub for non-local attempts
            
            print(f'   Trying to load from: {load_path} with config: {config}')
            model = model_class.from_pretrained(load_path, **config)
            print(f'   ✅ Model loaded successfully')
            break
        except Exception as e:
            last_error = e
            error_str = str(e).lower()
            if 'utf-8' in error_str or 'safetensor' in error_str:
                print(f'   ⚠️  SafeTensors UTF-8 error, will try without safetensors...')
            elif 'local' in error_str or 'not found' in error_str:
                print(f'   ⚠️  Local load failed, will try Hub...')
            else:
                print(f'   ⚠️  Load failed: {str(e)[:100]}...')
    
    if model is None:
        raise RuntimeError(f'Failed to load model after all attempts: {last_error}')
    
    # Load tokenizer or processor
    processor = None
    try:
        if is_vision_task:
            # Try image processor first for vision models
            try:
                processor = AutoImageProcessor.from_pretrained(
                    model_path if os.path.isdir(model_path) else hf_model_id,
                    local_files_only=os.path.isdir(model_path)
                )
                print('   ✅ Image processor loaded')
            except:
                print('   ⚠️  No image processor found')
        else:
            # Load tokenizer for NLP models
            processor = AutoTokenizer.from_pretrained(
                model_path if os.path.isdir(model_path) else hf_model_id,
                local_files_only=os.path.isdir(model_path)
            )
            print('   ✅ Tokenizer loaded')
    except Exception as e:
        print(f'   ⚠️  Processor/Tokenizer load failed: {str(e)[:100]}')
    
    return model, processor, task


def create_dummy_input(model, task, processor=None):
    """Create appropriate dummy input based on task type."""
    import torch
    
    config = model.config
    
    # All vision tasks that need image tensors
    vision_tasks = [
        'image-classification', 
        'object-detection', 
        'semantic-segmentation',
        'depth-estimation',
        'image-to-text',
    ]
    
    if task in vision_tasks:
        # Standard ImageNet input size, but check config for model-specific sizes
        batch_size = 1
        channels = 3
        
        # Try to get image size from config (different models use different attribute names)
        height = getattr(config, 'image_size', None) or \
                 getattr(config, 'img_size', None) or \
                 getattr(config, 'sample_size', None) or \
                 224
        width = height
        
        # Handle tuple/list image sizes
        if isinstance(height, (list, tuple)):
            height, width = height[0], height[-1]
        
        # Ensure integer values
        height = int(height)
        width = int(width)
        
        print(f'   Creating vision input: ({batch_size}, {channels}, {height}, {width})')
        dummy_input = torch.randn(batch_size, channels, height, width)
        input_names = ['pixel_values']
        
        # Output names vary by task
        if task == 'object-detection':
            output_names = ['logits', 'pred_boxes']
        elif task in ['semantic-segmentation', 'depth-estimation']:
            output_names = ['logits']
        else:
            output_names = ['logits']
        
        dynamic_axes = {
            'pixel_values': {0: 'batch_size'},
        }
        for name in output_names:
            dynamic_axes[name] = {0: 'batch_size'}
        
        return dummy_input, input_names, dynamic_axes
    
    elif task == 'zero-shot-image-classification':
        # CLIP-style models need both image and text
        batch_size = 1
        height = getattr(config, 'vision_config', config).get('image_size', 224) if hasattr(config, 'vision_config') else 224
        
        pixel_values = torch.randn(batch_size, 3, height, height)
        input_ids = torch.randint(0, 1000, (batch_size, 77))
        attention_mask = torch.ones(batch_size, 77, dtype=torch.long)
        
        print(f'   Creating multimodal input: image ({batch_size}, 3, {height}, {height}) + text ({batch_size}, 77)')
        return {
            'pixel_values': pixel_values,
            'input_ids': input_ids,
            'attention_mask': attention_mask
        }, ['pixel_values', 'input_ids', 'attention_mask'], {}
    
    else:
        # NLP models - text input
        seq_len = min(128, getattr(config, 'max_position_embeddings', 128))
        vocab_size = getattr(config, 'vocab_size', 30522)
        
        print(f'   Creating text input: (1, {seq_len}), vocab_size={vocab_size}')
        dummy_input = torch.randint(0, vocab_size, (1, seq_len))
        input_names = ['input_ids']
        dynamic_axes = {
            'input_ids': {0: 'batch_size', 1: 'sequence_length'},
            'output': {0: 'batch_size'}
        }
        return dummy_input, input_names, dynamic_axes


def try_torch_export(model, dummy_input, output_path, input_names, dynamic_axes, task):
    """
    Strategy 2: Use torch.onnx.export with proper handling for different model types.
    """
    try:
        import torch
        
        print('🔄 Strategy 2: Trying torch.onnx.export...')
        
        model.eval()
        
        # Handle dict inputs (multimodal)
        if isinstance(dummy_input, dict):
            # Create tuple of inputs in order
            input_tuple = tuple(dummy_input.values())
            
            torch.onnx.export(
                model,
                input_tuple,
                output_path,
                input_names=input_names,
                output_names=['logits'],
                dynamic_axes=dynamic_axes,
                opset_version=14,
                do_constant_folding=True,
                export_params=True,
                verbose=False,
            )
        else:
            # For vision models, wrap input properly
            vision_tasks = [
                'image-classification', 'object-detection', 'semantic-segmentation',
                'depth-estimation', 'image-to-text'
            ]
            if task in vision_tasks:
                output_names = ['logits']
            else:
                output_names = ['output']
            
            torch.onnx.export(
                model,
                dummy_input,
                output_path,
                input_names=input_names,
                output_names=output_names,
                dynamic_axes=dynamic_axes,
                opset_version=14,
                do_constant_folding=True,
                export_params=True,
                verbose=False,
            )
        
        if os.path.exists(output_path):
            print('✅ SUCCESS (torch.onnx.export)')
            return True
        
        return False
        
    except Exception as e:
        print(f'⚠️  torch.onnx.export failed: {str(e)}')
        return False


def try_torch_jit_trace(model, dummy_input, output_path, input_names, dynamic_axes):
    """
    Strategy 3: Use torch.jit.trace + ONNX export.
    Works well for models without dynamic control flow.
    """
    try:
        import torch
        
        print('🔄 Strategy 3: Trying torch.jit.trace + ONNX export...')
        
        # Skip for dict inputs
        if isinstance(dummy_input, dict):
            print('   Skipping JIT trace for dict inputs')
            return False
        
        # Trace the model
        traced_model = torch.jit.trace(model, dummy_input)
        
        # Export traced model to ONNX
        torch.onnx.export(
            traced_model,
            dummy_input,
            output_path,
            input_names=input_names,
            output_names=['output'],
            dynamic_axes=dynamic_axes,
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


def convert_huggingface_to_onnx(model_path, output_path, axon_model_id):
    """Convert a Hugging Face model to ONNX using multiple strategies."""
    try:
        import torch
        
        # Create output directory
        os.makedirs(os.path.dirname(output_path) if os.path.dirname(output_path) else '.', exist_ok=True)
        
        # Extract HF model ID
        hf_model_id = extract_hf_model_id(axon_model_id)
        print(f'📦 Converting model: {hf_model_id} (Axon ID: {axon_model_id})')
        
        # Detect task first
        detected_task = detect_task_from_config(model_path)
        print(f'   Detected task: {detected_task}')
        
        # Strategy 1: Try Optimum first (doesn't need model loading)
        if try_optimum_export(model_path, output_path, hf_model_id, task=detected_task):
            return True
        
        # Load model for other strategies
        print('📥 Loading model with transformers...')
        model, processor, task = load_model_and_tokenizer(model_path, hf_model_id, detected_task)
        model.eval()
        
        # Create appropriate dummy input
        dummy_input, input_names, dynamic_axes = create_dummy_input(model, task, processor)
        
        # Try remaining strategies in order
        strategies = [
            lambda: try_torch_export(model, dummy_input, output_path, input_names, dynamic_axes, task),
            lambda: try_torch_jit_trace(model, dummy_input, output_path, input_names, dynamic_axes),
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
