# Phase 1 Adapter Implementation Plan

## Overview

This document outlines the implementation plan for Phase 1 adapters: **PyTorch Hub**, **ModelScope**, and **TensorFlow Hub**. These adapters will expand Axon's coverage to **80%+ of the ML model user base**.

## Implementation Priority

Based on complexity, API accessibility, and user demand:

1. **PyTorch Hub** (Priority 1) - Simplest API, well-documented, research focus
2. **TensorFlow Hub** (Priority 2) - Standard REST API, production focus
3. **ModelScope** (Priority 3) - More complex API, multimodal focus, growing adoption

## 1. PyTorch Hub Adapter

### Overview
- **Repository**: https://pytorch.org/hub
- **API**: GitHub-based with metadata in `hubconf.py` files
- **Models**: 100+ pre-trained models
- **Coverage**: ~5% of ML practitioners

### Technical Approach

#### Model Discovery
- PyTorch Hub models are hosted on GitHub
- Each model has a `hubconf.py` file defining entry points
- Models are organized by repository (e.g., `pytorch/vision`, `pytorch/text`)

#### Implementation Strategy
1. **Namespace**: `pytorch/` or `torch/`
2. **Model Format**: `pytorch/{repo}/{model_name}@version`
   - Example: `pytorch/vision/resnet50@latest`
3. **API Access**:
   - Use GitHub API to fetch `hubconf.py` files
   - Parse Python config to extract model metadata
   - Download model weights via GitHub releases or direct URLs

#### Key Challenges
- Parsing Python `hubconf.py` files (may need Python interpreter or parser)
- Handling GitHub rate limits
- Extracting model weights from various repository structures

#### Implementation Steps
1. Create `PyTorchHubAdapter` struct
2. Implement `CanHandle()` for `pytorch/` namespace
3. Implement `Search()` using GitHub API
4. Implement `GetManifest()` by parsing `hubconf.py`
5. Implement `DownloadPackage()` to fetch weights and create `.axon` package
6. Add tests with mock GitHub responses

#### Resources
- PyTorch Hub API: https://pytorch.org/docs/stable/hub.html
- GitHub API: https://docs.github.com/en/rest
- Example repos: `pytorch/vision`, `pytorch/text`, `pytorch/audio`

### Estimated Effort
- **Research**: 2-3 days
- **Implementation**: 5-7 days
- **Testing**: 2-3 days
- **Total**: ~2 weeks

---

## 2. TensorFlow Hub Adapter

### Overview
- **Repository**: https://tfhub.dev
- **API**: RESTful API with JSON metadata
- **Models**: 1,000+ pre-trained models
- **Coverage**: ~7% of ML practitioners

### Technical Approach

#### Model Discovery
- TensorFlow Hub provides a REST API
- Models are organized by publisher (e.g., `google`, `tensorflow`)
- Each model has metadata in JSON format

#### Implementation Strategy
1. **Namespace**: `tfhub/` or `tf/`
2. **Model Format**: `tfhub/{publisher}/{model_path}@version`
   - Example: `tfhub/google/imagenet/resnet_v2_50/classification/5`
3. **API Access**:
   - Use TensorFlow Hub REST API: `https://tfhub.dev/api/v1/models`
   - Fetch model metadata: `https://tfhub.dev/{publisher}/{model_path}/{version}`
   - Download model files via direct URLs

#### Key Features
- Well-documented REST API
- Standardized model format (SavedModel or TFLite)
- Version management built-in
- Metadata includes input/output specs

#### Implementation Steps
1. Create `TensorFlowHubAdapter` struct
2. Implement `CanHandle()` for `tfhub/` namespace
3. Implement `Search()` using TF Hub REST API
4. Implement `GetManifest()` by fetching model metadata
5. Implement `DownloadPackage()` to download SavedModel/TFLite and create `.axon` package
6. Add tests with mock API responses

#### Resources
- TensorFlow Hub API: https://www.tensorflow.org/hub/api_docs
- REST API docs: https://tfhub.dev/api/v1/docs
- Model format: https://www.tensorflow.org/hub/model_format

### Estimated Effort
- **Research**: 1-2 days
- **Implementation**: 4-6 days
- **Testing**: 2-3 days
- **Total**: ~1.5 weeks

---

## 3. ModelScope Adapter

### Overview
- **Repository**: https://modelscope.cn (Chinese) / https://modelscope.co (International)
- **API**: REST API with Python SDK available
- **Models**: 5,000+ models (multimodal, Chinese language models)
- **Coverage**: ~8% of ML practitioners (growing)

### Technical Approach

#### Model Discovery
- ModelScope provides REST API and Python SDK
- Models organized by namespace (e.g., `damo`, `ai`, `cv`, `nlp`)
- Rich metadata including model cards, usage examples

#### Implementation Strategy
1. **Namespace**: `modelscope/`
2. **Model Format**: `modelscope/{namespace}/{model_name}@version`
   - Example: `modelscope/damo/nlp_structbert_sentence-similarity_chinese-base@latest`
3. **API Access**:
   - Use ModelScope REST API: `https://modelscope.cn/api/v1/models`
   - Fetch model metadata: `https://modelscope.cn/api/v1/models/{namespace}/{model_name}`
   - Download model files via API endpoints

#### Key Features
- Multimodal support (vision, audio, text, video)
- Strong Chinese language model collection
- Enterprise-focused features
- Model cards with detailed metadata

#### Implementation Steps
1. Create `ModelScopeAdapter` struct
2. Implement `CanHandle()` for `modelscope/` namespace
3. Implement `Search()` using ModelScope REST API
4. Implement `GetManifest()` by fetching model metadata
5. Implement `DownloadPackage()` to download model files and create `.axon` package
6. Handle authentication if needed (API keys for some models)
7. Add tests with mock API responses

#### Resources
- ModelScope API: https://modelscope.cn/docs/api
- API Documentation: https://modelscope.cn/docs/api_docs
- Python SDK: https://github.com/modelscope/modelscope

### Estimated Effort
- **Research**: 2-3 days (API documentation may be in Chinese)
- **Implementation**: 6-8 days (more complex API)
- **Testing**: 3-4 days
- **Total**: ~2.5 weeks

---

## Common Implementation Patterns

### Adapter Structure
All adapters will follow the same pattern as `HuggingFaceAdapter`:

```go
type PyTorchHubAdapter struct {
    httpClient *http.Client
    baseURL    string
    // Adapter-specific fields
}

func (p *PyTorchHubAdapter) Name() string { return "pytorch" }
func (p *PyTorchHubAdapter) CanHandle(namespace, name string) bool { ... }
func (p *PyTorchHubAdapter) Search(...) ([]types.SearchResult, error) { ... }
func (p *PyTorchHubAdapter) GetManifest(...) (*types.Manifest, error) { ... }
func (p *PyTorchHubAdapter) DownloadPackage(...) error { ... }
```

### Registration Order
Adapters will be registered in `commands.go` with priority:
1. Local Registry (if configured)
2. PyTorch Hub (if namespace matches)
3. ModelScope (if namespace matches)
4. TensorFlow Hub (if namespace matches)
5. Hugging Face (fallback)

### Testing Strategy
- Unit tests with mocked HTTP responses
- Integration tests with real API calls (optional, rate-limited)
- Test with popular models from each repository
- Verify package creation and manifest generation

---

## Implementation Timeline

### Phase 1.1: PyTorch Hub (Weeks 1-2)
- [ ] Research PyTorch Hub API and model structure
- [ ] Implement `PyTorchHubAdapter`
- [ ] Add tests
- [ ] Update documentation
- [ ] Create PR and merge

### Phase 1.2: TensorFlow Hub (Weeks 3-4)
- [ ] Research TensorFlow Hub REST API
- [ ] Implement `TensorFlowHubAdapter`
- [ ] Add tests
- [ ] Update documentation
- [ ] Create PR and merge

### Phase 1.3: ModelScope (Weeks 5-7)
- [ ] Research ModelScope API (may need translation)
- [ ] Implement `ModelScopeAdapter`
- [ ] Handle authentication if needed
- [ ] Add tests
- [ ] Update documentation
- [ ] Create PR and merge

### Phase 1.4: Documentation & Release (Week 8)
- [ ] Update all documentation
- [ ] Update website
- [ ] Create release notes
- [ ] Release v1.1.0 with Phase 1 adapters

---

## Success Criteria

### Functional Requirements
- ✅ All three adapters can search for models
- ✅ All three adapters can fetch manifests
- ✅ All three adapters can download and package models
- ✅ Models install correctly via `axon install`
- ✅ Models appear in `axon list` after installation

### Quality Requirements
- ✅ All adapters have >80% test coverage
- ✅ All adapters handle errors gracefully
- ✅ All adapters respect rate limits
- ✅ Documentation is complete and accurate

### Performance Requirements
- ✅ Model download completes within reasonable time
- ✅ Progress tracking works for large models
- ✅ Caching works correctly

---

## Risks and Mitigations

### Risk 1: API Changes
- **Risk**: Repository APIs may change
- **Mitigation**: Version API calls, add fallback mechanisms, monitor API status

### Risk 2: Rate Limiting
- **Risk**: API rate limits may block downloads
- **Mitigation**: Implement retry logic with exponential backoff, cache responses

### Risk 3: Authentication Required
- **Risk**: Some models may require authentication
- **Mitigation**: Support API keys in config, provide clear error messages

### Risk 4: Model Format Variations
- **Risk**: Different repositories use different model formats
- **Mitigation**: Create format-specific handlers, validate packages

---

## Next Steps

1. **Review and approve this plan**
2. **Start with PyTorch Hub adapter** (simplest, good learning experience)
3. **Iterate based on learnings** from first adapter
4. **Continue with TensorFlow Hub** (standard REST API)
5. **Finish with ModelScope** (most complex, but high value)

---

## References

- [PyTorch Hub Documentation](https://pytorch.org/docs/stable/hub.html)
- [TensorFlow Hub API](https://www.tensorflow.org/hub/api_docs)
- [ModelScope Documentation](https://modelscope.cn/docs)
- [ONNX Model Zoo Deprecation](https://onnx.ai/models/)
- [Axon Adapter Architecture](./REPOSITORY_ADAPTERS.md)

---

**Status**: Planning Phase  
**Last Updated**: 2025-01-07  
**Next Review**: After PyTorch Hub implementation

