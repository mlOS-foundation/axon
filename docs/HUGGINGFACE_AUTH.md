# Hugging Face Authentication

## Do You Need HF Keys?

**Short answer: No, for most use cases.**

### Public Models (No Token Required)

Most Hugging Face models are **public** and can be downloaded **without authentication**:

- ✅ **No token needed** for public models
- ✅ Works out of the box
- ✅ Examples: `bert-base-uncased`, `gpt2`, `roberta-base`, etc.

### When You DO Need a Token

You only need a Hugging Face token for:

1. **Gated Models** - Models that require acceptance of terms
   - Example: `meta-llama/Llama-2-7b-hf`
   - Requires accepting model license on Hugging Face

2. **Private Models** - Models in private repositories
   - Your own private models
   - Organization private models

3. **Rate Limits** - Higher rate limits (optional)
   - Public API: ~50 requests/hour
   - With token: Higher limits (depends on account)

## How It Works

### Without Token (Default)

```bash
# Works for public models - no setup needed
axon install hf/bert-base-uncased@latest
axon install hf/gpt2@latest
```

Axon automatically:
- Downloads from Hugging Face public API
- No authentication required
- Works for 99% of models

### With Token (For Gated/Private Models)

1. **Get your HF token**:
   - Go to https://huggingface.co/settings/tokens
   - Create a new token (read access is enough)

2. **Configure Axon**:
   ```bash
   # Set token in config
   axon config set registry.huggingface_token <your-token>
   ```

   Or edit `~/.axon/config.yaml`:
   ```yaml
   registry:
     enable_huggingface: true
     huggingface_token: "hf_xxxxxxxxxxxxxxxxxxxx"
   ```

3. **Install gated/private models**:
   ```bash
   # Now works with gated models
   axon install hf/meta-llama/Llama-2-7b-hf@latest
   ```

## Authentication Flow

```
┌─────────────────┐
│  axon install   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Is token set?  │───No───► Use public API (works for most models)
└────────┬────────┘
         │
        Yes
         │
         ▼
┌─────────────────┐
│  Use token in   │
│  Authorization  │
│     header      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Download from  │
│  HF Hub (with   │
│  authentication)│
└─────────────────┘
```

## Examples

### Example 1: Public Model (No Token)

```bash
# No configuration needed
axon install hf/bert-base-uncased@latest
```

**Output:**
```
Using huggingface adapter for hf/bert-base-uncased
Downloading package...
Downloading... 100.0% (440234567/440234567 bytes)
✓ Successfully propagated hf/bert-base-uncased@latest
```

### Example 2: Gated Model (Token Required)

```bash
# First, set your token
axon config set registry.huggingface_token hf_xxxxxxxxxxxxxxxxxxxx

# Now install gated model
axon install hf/meta-llama/Llama-2-7b-hf@latest
```

### Example 3: Check if Token is Set

```bash
# View current config
axon config get registry.huggingface_token
```

## Security Best Practices

### Token Storage

- ✅ Tokens are stored in `~/.axon/config.yaml`
- ✅ File permissions: `0600` (readable only by owner)
- ✅ Never commit tokens to version control

### Token Scope

- ✅ Use **read** token for model downloads
- ✅ No need for write permissions
- ✅ Can be revoked anytime from HF settings

### Environment Variables (Alternative)

You can also set token via environment variable:

```bash
export HF_TOKEN="hf_xxxxxxxxxxxxxxxxxxxx"
axon install hf/meta-llama/Llama-2-7b-hf@latest
```

## Troubleshooting

### Error: "401 Unauthorized"

**Cause**: Model is gated/private and requires token

**Solution**:
```bash
axon config set registry.huggingface_token <your-token>
```

### Error: "403 Forbidden"

**Cause**: Token doesn't have access to the model

**Solution**:
1. Accept model license on Hugging Face website
2. Ensure token has correct permissions
3. Check if model is in a private org you're not a member of

### Error: "Rate limit exceeded"

**Cause**: Too many requests without authentication

**Solution**:
- Add token for higher rate limits
- Or wait and retry

## Summary

| Scenario | Token Required? |
|----------|----------------|
| Public models (99% of models) | ❌ No |
| Gated models (Llama, etc.) | ✅ Yes |
| Private models | ✅ Yes |
| Higher rate limits | ✅ Optional |

**For most users**: No token needed! Axon works out of the box with public models.

**For advanced users**: Add token only when needed for gated/private models.

