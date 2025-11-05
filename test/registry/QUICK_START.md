# Quick Start: Testing Axon Locally

## 🚀 In 3 Steps

### Step 1: Start the Registry Server

Open a terminal and run:

```bash
cd axon/test/registry
go run server.go .
```

You should see:
```
🚀 Starting local registry server on http://localhost:8080
📁 Registry directory: .
🌐 Web UI: http://localhost:8080
🔍 API: http://localhost:8080/api/v1/search?q=<query>
```

**Keep this terminal open!**

### Step 2: Configure Axon

Open a **new terminal** and run:

```bash
cd axon
./bin/axon init
./bin/axon registry set default http://localhost:8080
```

### Step 3: Test!

**Via Browser:**
- Open http://localhost:8080 in your browser
- Browse models, search, view manifests
- Click "View Manifest" to see YAML
- Click "Install" to see the CLI command

**Via CLI:**
```bash
# Search for models
./bin/axon search bert
./bin/axon search vision

# Get model info
./bin/axon info nlp/gpt2@1.0.0

# Install a model
./bin/axon install nlp/gpt2@1.0.0

# List installed models
./bin/axon list
```

## 🧪 Test Examples

### Browser Testing

1. **Search Interface:**
   - Go to http://localhost:8080
   - Type "bert" in the search box
   - See filtered results appear

2. **Direct API Access:**
   - Search: http://localhost:8080/api/v1/search?q=bert
   - Manifest: http://localhost:8080/api/v1/models/nlp/bert-base-uncased/1.0.0/manifest.yaml
   - Package: http://localhost:8080/packages/nlp-bert-base-uncased-1.0.0.axon

### CLI Testing

```bash
# Search
./bin/axon search bert
# Output: Found 3 models (bert-base-uncased, distilbert-base-uncased, roberta-base)

# Info
./bin/axon info nlp/gpt2@1.0.0
# Output: Model details, framework, license

# Install
./bin/axon install nlp/gpt2@1.0.0
# Output: Downloads package, verifies checksum, installs

# List
./bin/axon list
# Output: Shows installed models
```

## 🛠️ Troubleshooting

**Server won't start?**
- Check if port 8080 is in use: `lsof -i :8080`
- Use different port: `PORT=8081 go run server.go .`
- Then update config: `axon registry set default http://localhost:8081`

**Browser shows empty page?**
- Check server is running: `curl http://localhost:8080/api/v1/search?q=bert`
- Check browser console for errors
- Verify manifests exist: `ls -R api/v1/models/`

**CLI commands fail?**
- Verify registry is configured: `axon registry list`
- Check server is running: `curl http://localhost:8080/api/v1/search?q=bert`
- Ensure axon is built: `make build`

## 📚 More Info

See `test/registry/README.md` for full documentation.

