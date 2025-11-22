# Docker Build Hang Fix - After PyTorch Installation

## Problem

The Docker build hangs indefinitely after PyTorch installation completes. The log shows:
```
Successfully installed torch-2.9.1+cpu
WARNING: Running pip as the 'root' user...
```
Then the build appears to hang with no further output.

## Root Cause

The issue is caused by the **single long-running RUN command** that installs all packages sequentially:

```dockerfile
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip && \
    pip install "protobuf<4.21,>=3.20" && \
    pip install tensorflow>=2.13.0 tf2onnx>=1.15.0 && \
    pip install torch --index-url https://download.pytorch.org/whl/cpu && \
    pip install transformers>=4.30.0 modelscope>=1.9.0 onnx>=1.14.0 onnxruntime>=1.18.0 numpy>=1.24.0 && \
    pip cache purge || true
```

### Why It Hangs

1. **No Timeout Protection**: After PyTorch (~2-3GB download), pip continues to the next packages (transformers ~200MB, modelscope ~200MB) without timeout protection. Network issues cause indefinite waits.

2. **BuildKit Cache Mount Lock**: The cache mount may lock during large package installations, causing subsequent pip commands to wait indefinitely.

3. **No Visibility**: All pip installs in one RUN command means if any package hangs, the entire build hangs with no indication of which package is causing the issue.

4. **GitHub Actions Network Variability**: GitHub Actions runners have variable network speeds. After downloading PyTorch (large), network may be slow or interrupted, causing subsequent downloads to hang.

5. **No Retry Logic**: Pip doesn't retry failed downloads automatically, so network hiccups cause permanent hangs.

## Solution: Split RUN Commands with Timeouts

### Fixed Dockerfile

```dockerfile
# Set pip timeout environment variable
ENV PIP_DEFAULT_TIMEOUT=300
ENV PIP_ROOT_USER_ACTION=ignore

# Upgrade pip first (small, fast)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip

# Install protobuf (small, fast, required for TensorFlow)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 "protobuf<4.21,>=3.20"

# Install TensorFlow ecosystem (large but independent)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 \
    tensorflow>=2.13.0 \
    tf2onnx>=1.15.0

# Install PyTorch (largest package, ~2-3GB, needs longer timeout)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=600 \
    --retries 3 \
    torch --index-url https://download.pytorch.org/whl/cpu

# Install ML libraries (smaller but can hang if network issues)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 \
    --retries 3 \
    transformers>=4.30.0 \
    modelscope>=1.9.0 \
    onnx>=1.14.0 \
    onnxruntime>=1.18.0 \
    numpy>=1.24.0

# Clean cache (separate command for clarity)
RUN pip cache purge || true
```

### Key Changes

1. **Separate RUN Commands**: Each package group in its own RUN command
   - Better visibility into which package hangs
   - Independent caching per group
   - Can retry specific groups

2. **Timeout Protection**: `--default-timeout=300` (5 minutes) or `--default-timeout=600` (10 minutes for PyTorch)
   - Prevents indefinite hangs
   - Fails fast on network issues
   - Can be retried

3. **Retry Logic**: `--retries 3` for large packages
   - Automatically retries on network failures
   - Reduces build failures due to transient network issues

4. **Environment Variables**: 
   - `PIP_DEFAULT_TIMEOUT=300`: Global timeout fallback
   - `PIP_ROOT_USER_ACTION=ignore`: Suppresses warning (cleaner logs)

## Benefits

- ✅ **No More Hangs**: Timeout protection prevents indefinite waits
- ✅ **Better Visibility**: Can see exactly which package is installing
- ✅ **Better Caching**: Each package group cached independently
- ✅ **Easier Debugging**: Failed package clearly identified in logs
- ✅ **Resumable**: Can retry specific package groups without rebuilding everything
- ✅ **More Reliable**: Retry logic handles transient network issues

## Performance Impact

- **Build Time**: Same or slightly faster (better cache utilization)
- **Reliability**: Significantly improved (no hangs, automatic retries)
- **Debugging**: Much easier (clear failure points)

## Alternative: Use requirements.txt

For even better maintainability:

```dockerfile
# requirements.txt
protobuf<4.21,>=3.20
tensorflow>=2.13.0
tf2onnx>=1.15.0
torch --index-url https://download.pytorch.org/whl/cpu
transformers>=4.30.0
modelscope>=1.9.0
onnx>=1.14.0
onnxruntime>=1.18.0
numpy>=1.24.0

# Dockerfile
COPY requirements.txt /tmp/requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 --retries 3 -r /tmp/requirements.txt
```

This allows pip to optimize installation order and handle dependencies better.

