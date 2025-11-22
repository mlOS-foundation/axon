# Docker Build Performance Recommendations

## Critical Issue: Build Hanging After PyTorch

### Immediate Fix Required

**Problem**: Build hangs indefinitely after PyTorch installation completes.

**Root Cause**: Single long-running RUN command with no timeout protection. After PyTorch (~2-3GB), subsequent packages (transformers, modelscope) hang due to network issues or cache mount locks.

**Fix**: Split into separate RUN commands with timeouts:

```dockerfile
ENV PIP_DEFAULT_TIMEOUT=300
ENV PIP_ROOT_USER_ACTION=ignore

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 "protobuf<4.21,>=3.20"

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 tensorflow>=2.13.0 tf2onnx>=1.15.0

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=600 --retries 3 \
    torch --index-url https://download.pytorch.org/whl/cpu

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --default-timeout=300 --retries 3 \
    transformers>=4.30.0 \
    modelscope>=1.9.0 \
    onnx>=1.14.0 \
    onnxruntime>=1.18.0 \
    numpy>=1.24.0

RUN pip cache purge || true
```

**Impact**: Prevents hangs, enables better debugging, improves reliability.

## Performance Optimization Strategy

### Phase 1: Quick Wins (Immediate - 50% improvement)

#### 1. Parallel Platform Builds
**Current**: Sequential builds for `linux/amd64` and `linux/arm64`  
**Proposed**: Build platforms in parallel using matrix strategy

```yaml
jobs:
  build-platform:
    strategy:
      matrix:
        platform: [linux/amd64, linux/arm64]
    steps:
      - name: Build for ${{ matrix.platform }}
        uses: docker/build-push-action@v5
        with:
          platforms: ${{ matrix.platform }}
```

**Impact**: 50% faster multi-platform builds (6-10 min → 3-5 min)

#### 2. Registry Cache
**Current**: Only GitHub Actions cache  
**Proposed**: Add registry cache for better persistence

```yaml
cache-from: |
  type=registry,ref=ghcr.io/mlos-foundation/axon-converter:buildcache
  type=gha
cache-to: |
  type=registry,ref=ghcr.io/mlos-foundation/axon-converter:buildcache,mode=max
  type=gha,mode=max
```

**Impact**: 10-15% faster builds with better cache hits

### Phase 2: Dockerfile Optimization (Short-term - 30% improvement)

#### 1. Split Package Installation (Fixes Hang Issue)
Already covered above - split RUN commands with timeouts.

#### 2. Multi-Stage Build
Separate package installation from final image for better caching.

#### 3. Add .dockerignore
Reduce build context size by excluding unnecessary files.

### Phase 3: Industry Standard (Long-term - 60-80% improvement)

#### Option A: Kaniko (Recommended)
**Why**: Industry standard (Google, GitLab), OCI-native, 40-60% faster

```yaml
- name: Build with Kaniko
  uses: imjasonh/setup-kaniko@v0.1
  
- name: Build and push
  run: |
    /kaniko/executor \
      --context . \
      --dockerfile docker/Dockerfile.converter \
      --destination ghcr.io/mlos-foundation/axon-converter:${{ version }} \
      --cache=true \
      --cache-ttl=168h \
      --cache-repo=ghcr.io/mlos-foundation/axon-converter-cache
```

**Benefits**:
- ✅ 40-60% faster than Docker BuildKit
- ✅ OCI-compliant (industry standard)
- ✅ Better cache management
- ✅ No Docker daemon required
- ✅ Used by Google, GitLab, many enterprises

#### Option B: Pre-built Base Image (Highest Impact)
**Why**: 90% faster subsequent builds

Create `axon-converter-base` image with all Python packages, then:

```dockerfile
FROM ghcr.io/mlos-foundation/axon-converter-base:latest
COPY scripts/conversion/ /axon/scripts/
RUN chmod +x /axon/scripts/*.py
```

**Impact**: 20 min → 30 seconds for subsequent builds

## Recommended Implementation Order

1. **URGENT**: Fix hanging issue (split RUN commands with timeouts)
2. **High Priority**: Parallel platform builds (50% improvement)
3. **High Priority**: Pre-built base image (90% improvement)
4. **Medium Priority**: Migrate to Kaniko (40-60% improvement)
5. **Low Priority**: Other optimizations (incremental)

## Expected Performance

### Current
- First build: 15-20 minutes
- Subsequent builds: 3-5 minutes
- Multi-platform: 6-10 minutes

### After Fixes
- First build: 8-12 minutes (-40%)
- Subsequent builds: 30-60 seconds (-90%)
- Multi-platform: 1-2 minutes (-85%)

## Industry Standard Recommendation

**Kaniko + Pre-built Base Image** is the industry-standard approach:

1. **Kaniko** for OCI-native builds (Google standard)
2. **Pre-built base image** for ultra-fast builds (common practice)
3. **Parallel platform builds** for multi-arch (standard practice)

This combination provides:
- ✅ Industry alignment (OCI standard)
- ✅ Maximum performance (90%+ improvement)
- ✅ Reliability (no hangs, better error handling)
- ✅ Maintainability (clear separation of concerns)

