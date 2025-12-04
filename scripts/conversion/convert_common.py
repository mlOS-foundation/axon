#!/usr/bin/env python3
"""
Common utilities for ONNX conversion scripts.
Shared functions for multi-encoder detection and manifest generation.
"""

import os
import json


def find_onnx_files(directory):
    """Find all ONNX files in a directory."""
    onnx_files = []
    if not os.path.isdir(directory):
        return onnx_files
    for f in os.listdir(directory):
        if f.endswith('.onnx'):
            onnx_files.append(os.path.join(directory, f))
    return sorted(onnx_files)


def write_multi_encoder_manifest(output_dir, onnx_files, task=None):
    """
    Write a manifest file for multi-encoder models.
    This helps Core understand how to load and orchestrate multiple ONNX files.
    """
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
        'task': task or 'unknown',
        'components': components,
        'files': file_names
    }
    
    manifest_path = os.path.join(output_dir, 'onnx_manifest.json')
    with open(manifest_path, 'w') as f:
        json.dump(manifest, f, indent=2)
    
    print(f'   Created onnx_manifest.json for {architecture} model')
    return manifest_path

