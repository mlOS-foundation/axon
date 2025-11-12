# ModelScope Promotion Plan

## Overview

This plan outlines the steps to:
1. Move ModelScope adapter from `examples/` to `builtin/` (making it a full supported adapter)
2. Add a new example adapter to replace ModelScope in examples
3. Update all documentation and website to reflect these changes

## Phase 1: Move ModelScope to Builtin

### Code Changes

1. **Move ModelScope files**:
   - `internal/registry/examples/modelscope.go` → `internal/registry/builtin/modelscope.go`
   - `internal/registry/examples/modelscope_test.go` → `internal/registry/builtin/modelscope_test.go`

2. **Update imports**:
   - Change package from `examples` to `builtin`
   - Update all internal references

3. **Update registration**:
   - Add ModelScope to `builtin/register.go`
   - Register it in `RegisterDefaultAdapters()` function
   - Set appropriate priority (after TensorFlow Hub, before Hugging Face fallback)

4. **Update CLI**:
   - ModelScope will automatically be available via `builtin.RegisterDefaultAdapters()`
   - No CLI changes needed (already uses the registry)

### Testing

- [ ] Unit tests pass for ModelScope adapter
- [ ] Integration test: `axon install modelscope/damo/cv_resnet50@latest`
- [ ] Verify ModelScope appears in adapter registry
- [ ] Verify ModelScope `CanHandle()` works correctly

## Phase 2: Add New Example Adapter

### Repository Selection

**Candidate: Replicate**
- Popular for hosted inference APIs
- Well-documented API
- Good example of REST API integration
- Different use case (API-based vs file-based)

**Alternative: Kaggle Models**
- Large repository
- Different authentication model
- Good example of complex integration

**Decision: Replicate** - Better example of API-based adapter with different patterns

### Implementation

1. **Create Replicate adapter**:
   - `internal/registry/examples/replicate.go`
   - `internal/registry/examples/replicate_test.go`
   - Implement `RepositoryAdapter` interface
   - Use core utilities (HTTPClient, ModelValidator, PackageBuilder)

2. **Features to demonstrate**:
   - API-based model access (vs file downloads)
   - Different authentication model
   - REST API integration patterns
   - Error handling for API responses

## Phase 3: Documentation Updates

### Axon Repository

1. **README.md**:
   - Move ModelScope from "Coming in Phase 1" to "Available in v1.4.0+"
   - Update coverage statistics
   - Add ModelScope usage examples

2. **REPOSITORY_ADAPTERS.md**:
   - Move ModelScope from "Example Adapter" to "Supported Adapters" section
   - Add full ModelScope adapter documentation
   - Update adapter priority list
   - Add Replicate to examples section

3. **ADAPTER_ROADMAP.md**:
   - Mark ModelScope as "Completed - v1.4.0"
   - Update Phase 1 status
   - Add Replicate to examples

4. **ADAPTER_FRAMEWORK.md**:
   - Update to show ModelScope as builtin adapter
   - Update examples to use Replicate instead of ModelScope
   - Update architecture diagrams

5. **ADAPTER_DEVELOPMENT.md**:
   - Update examples to reference Replicate
   - Keep ModelScope as reference for builtin adapters

6. **CHANGELOG.md**:
   - Add v1.4.0 entry for ModelScope promotion
   - Document Replicate example adapter

### Website (mlosfoundation.org)

1. **ecosystem.html**:
   - Move ModelScope from "Phase 1" to "Available v1.4.0+"
   - Update adapter list
   - Update coverage statistics
   - Update architecture diagram to show ModelScope in builtin

## Phase 4: Testing & Validation

1. **Code validation**:
   - [ ] All tests pass
   - [ ] Linting passes
   - [ ] Build succeeds

2. **Functional testing**:
   - [ ] ModelScope adapter works end-to-end
   - [ ] Replicate example adapter compiles and has tests
   - [ ] All adapters still work correctly

3. **Documentation validation**:
   - [ ] All links work
   - [ ] Examples are accurate
   - [ ] Version numbers are consistent

## Implementation Order

1. ✅ Create this plan
2. Move ModelScope files and update code
3. Add Replicate example adapter
4. Update all documentation
5. Update website
6. Test everything
7. Create PR

## Files to Modify

### Code Files
- `internal/registry/examples/modelscope.go` → `internal/registry/builtin/modelscope.go`
- `internal/registry/examples/modelscope_test.go` → `internal/registry/builtin/modelscope_test.go`
- `internal/registry/builtin/register.go` (add ModelScope)
- `internal/registry/examples/replicate.go` (new)
- `internal/registry/examples/replicate_test.go` (new)

### Documentation Files
- `README.md`
- `docs/REPOSITORY_ADAPTERS.md`
- `docs/ADAPTER_ROADMAP.md`
- `docs/ADAPTER_FRAMEWORK.md`
- `docs/ADAPTER_DEVELOPMENT.md`
- `CHANGELOG.md`

### Website Files
- `mlosfoundation.org/ecosystem.html`

## Success Criteria

- [ ] ModelScope is a fully supported builtin adapter
- [ ] ModelScope works end-to-end (`axon install modelscope/...`)
- [ ] Replicate example adapter exists and demonstrates API patterns
- [ ] All documentation is updated and consistent
- [ ] Website reflects ModelScope as available
- [ ] All tests pass
- [ ] No broken links or references

