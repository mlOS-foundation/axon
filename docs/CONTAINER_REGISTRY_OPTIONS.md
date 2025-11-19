# Container Registry Options for Axon

This document explores alternatives to GitHub Container Registry (GHCR) and Docker Hub for publishing container images as part of releases.

## Current Setup

Currently, Axon publishes Docker images to:
- **GitHub Container Registry (GHCR)**: `ghcr.io/mlos-foundation/axon-converter`

## Alternative Options

### 1. OCI Artifacts as Release Assets (Recommended)

**Store container images directly in GitHub Releases as OCI artifacts**

#### Pros
- ✅ No separate registry needed
- ✅ Images stored alongside release binaries
- ✅ No rate limits or storage costs
- ✅ Direct download from releases
- ✅ OCI-compliant format
- ✅ Works with standard Docker/containerd tools

#### Cons
- ⚠️ Larger release artifacts (images can be large)
- ⚠️ GitHub release size limits (2GB per file, 10GB total per release)
- ⚠️ Not optimized for frequent pulls (no CDN)

#### Implementation
```yaml
- name: Save Docker image as OCI artifact
  run: |
    docker save ghcr.io/mlos-foundation/axon-converter:latest | gzip > axon-converter-${VERSION}.tar.gz
    
- name: Upload to release
  uses: softprops/action-gh-release@v1
  with:
    files: axon-converter-${VERSION}.tar.gz
```

#### Tools
- `docker save` / `docker load` - Standard Docker tools
- `skopeo` - OCI image manipulation tool
- `oras` - OCI Registry as Storage (for OCI artifacts)

---

### 2. Quay.io (Red Hat Container Registry)

**Free, OCI-compliant container registry**

#### Pros
- ✅ Free for public repositories
- ✅ OCI-compliant
- ✅ No rate limits for public images
- ✅ Good performance and reliability
- ✅ Security scanning included
- ✅ Well-established (used by many open source projects)

#### Cons
- ⚠️ Requires separate account setup
- ⚠️ Additional registry to manage

#### Implementation
```yaml
- name: Log in to Quay.io
  uses: docker/login-action@v3
  with:
    registry: quay.io
    username: ${{ secrets.QUAY_USERNAME }}
    password: ${{ secrets.QUAY_PASSWORD }}

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    push: true
    tags: quay.io/mlos-foundation/axon-converter:${VERSION}
```

#### Setup
1. Create account at https://quay.io
2. Create repository: `mlos-foundation/axon-converter`
3. Generate robot account with read/write permissions
4. Add secrets: `QUAY_USERNAME`, `QUAY_PASSWORD`

---

### 3. Docker Hub

**Most widely used container registry**

#### Pros
- ✅ Most familiar to users
- ✅ Widely supported
- ✅ Good documentation

#### Cons
- ⚠️ Rate limits (100 pulls per 6 hours for anonymous, 200 for authenticated)
- ⚠️ Requires paid plan for unlimited pulls
- ⚠️ Not ideal for high-traffic open source projects

#### Implementation
```yaml
- name: Log in to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    push: true
    tags: mlosfoundation/axon-converter:${VERSION}
```

---

### 4. OCI Artifacts via ORAS (OCI Registry as Storage)

**Store images as OCI artifacts in any OCI-compliant registry**

#### Pros
- ✅ Standard OCI format
- ✅ Works with any OCI-compliant registry
- ✅ Can use GitHub Releases, Quay.io, Harbor, etc.
- ✅ Flexible storage options

#### Cons
- ⚠️ Requires `oras` CLI tool
- ⚠️ Less familiar to most users

#### Implementation
```yaml
- name: Install ORAS
  uses: oras-project/oras-install@v2.0.0

- name: Push OCI artifact to GitHub Releases
  run: |
    docker save axon-converter:latest | gzip | \
    oras push ghcr.io/mlos-foundation/axon-converter:${VERSION} \
      --artifact-type application/vnd.docker.container.image.v1+json \
      - < image.tar.gz
```

---

### 5. Multi-Registry Publishing

**Publish to multiple registries simultaneously**

#### Pros
- ✅ Redundancy and availability
- ✅ Users can choose preferred registry
- ✅ No single point of failure

#### Cons
- ⚠️ More complex workflow
- ⚠️ Multiple credentials to manage

#### Implementation
```yaml
- name: Build image
  uses: docker/build-push-action@v5
  with:
    push: false
    tags: axon-converter:${VERSION}
    outputs: type=docker,dest=/tmp/image.tar

- name: Push to GHCR
  uses: docker/build-push-action@v5
  with:
    load: true
    push: true
    tags: ghcr.io/mlos-foundation/axon-converter:${VERSION}
    inputs: /tmp/image.tar

- name: Push to Quay.io
  uses: docker/build-push-action@v5
  with:
    load: true
    push: true
    tags: quay.io/mlos-foundation/axon-converter:${VERSION}
    inputs: /tmp/image.tar
```

---

## Comparison Matrix

| Option | Cost | Rate Limits | Setup Complexity | User Familiarity | Recommended For |
|--------|------|-------------|------------------|------------------|-----------------|
| **GHCR** (current) | Free | None | Low | Medium | GitHub projects |
| **Release Assets** | Free | None | Low | Low | Small images, direct downloads |
| **Quay.io** | Free (public) | None | Medium | Medium | Open source projects |
| **Docker Hub** | Free (limited) | Yes | Low | High | General use |
| **Multi-Registry** | Free | Varies | High | Medium | High availability needs |

---

## Recommendations

### Option 1: OCI Artifacts in GitHub Releases (Best for Simplicity)
- **Use case**: Small to medium images (< 1GB), direct distribution
- **Pros**: No separate registry, integrated with releases
- **Implementation**: Save images as `.tar.gz` and attach to releases

### Option 2: Quay.io (Best for Open Source)
- **Use case**: Standard container registry, no rate limits
- **Pros**: Free, reliable, OCI-compliant
- **Implementation**: Add Quay.io credentials, push alongside GHCR

### Option 3: Multi-Registry (Best for Availability)
- **Use case**: High availability, user choice
- **Pros**: Redundancy, multiple options
- **Implementation**: Push to both GHCR and Quay.io

---

## Implementation Steps

### For OCI Artifacts in Releases:

1. Modify `docker-converter.yml` to save image as artifact
2. Integrate with `release.yml` to attach to release
3. Update documentation with download instructions

### For Quay.io:

1. Create Quay.io account and repository
2. Generate robot account credentials
3. Add secrets to GitHub repository
4. Update workflow to push to Quay.io
5. Update documentation with pull instructions

### For Multi-Registry:

1. Implement both GHCR and Quay.io publishing
2. Add conditional logic for registry selection
3. Update documentation with all registry options

---

## References

- [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec)
- [ORAS Project](https://oras.land/)
- [Quay.io Documentation](https://docs.quay.io/)
- [Docker Hub Rate Limits](https://www.docker.com/pricing/resource-consumption-updates)
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases)

