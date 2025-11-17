.PHONY: build test install clean lint fmt help vet coverage install-tools

# Build variables
BINARY_NAME=axon
VERSION?=v0.1.0
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Tool versions (matching CI)
GOLANGCI_LINT_VERSION=v1.64.8
GOIMPORTS_VERSION=latest

# Ensure GOPATH/bin is in PATH
export PATH := $(shell go env GOPATH)/bin:$(PATH)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/axon
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

coverage: test-coverage ## Alias for test-coverage

install-tools: ## Install all required CI tools (golangci-lint, goimports)
	@echo "Installing CI tools..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	else \
		echo "✓ golangci-lint already installed"; \
	fi
	@if ! command -v goimports > /dev/null; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION); \
	else \
		echo "✓ goimports already installed"; \
	fi
	@echo "✓ All tools installed. Make sure $(shell go env GOPATH)/bin is in your PATH"

lint: install-tools ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		echo "❌ golangci-lint not found. Run 'make install-tools' first."; \
		exit 1; \
	fi

fmt: install-tools ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "⚠️  goimports not found. Run 'make install-tools' to install it."; \
	fi
	@echo "✓ Code formatted"

fmt-check: ## Check code formatting without modifying
	@echo "Checking code formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted. Run 'make fmt'"; \
		gofmt -s -d .; \
		exit 1; \
	fi
	@echo "✓ Code is properly formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ go vet passed"

install: build ## Install to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/axon
	@echo "✓ Installed $(BINARY_NAME)"

build-local: build ## Build and install to ~/.local/bin (no sudo required)
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p $$HOME/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) $$HOME/.local/bin/$(BINARY_NAME)
	@chmod +x $$HOME/.local/bin/$(BINARY_NAME)
	@echo "✓ Installed $(BINARY_NAME) to $$HOME/.local/bin/$(BINARY_NAME)"
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo ""; \
		echo "⚠️  ~/.local/bin is not in your PATH"; \
		echo "   Add this to your shell config (~/.bashrc, ~/.zshrc, etc.):"; \
		echo "   export PATH=\"$$HOME/.local/bin:\$$PATH\""; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✓ Cleaned"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

tidy: ## Tidy go.mod
	@go mod tidy
	@echo "✓ go.mod tidied"

# Development targets
dev: ## Run in development mode
	@go run ./cmd/axon

# CI targets
ci: fmt-check vet lint test ## Run all CI checks

validate: fmt-check vet lint test ## Alias for ci

validate-pr: install-tools ## Run all validation checks before PR (fmt, vet, lint, test, build)
	@./validate-pr.sh

create-pr: validate-pr ## Create PR with automatic validation (use: make create-pr TITLE="..." BODY="...")
	@echo "✅ Validation passed! Creating PR..."
	@if [ -z "$(TITLE)" ]; then \
		echo "❌ Error: TITLE is required. Usage: make create-pr TITLE=\"...\" BODY=\"...\""; \
		exit 1; \
	fi
	@./scripts/create-pr.sh --title "$(TITLE)" --body "$(BODY)" $(if $(DRAFT),--draft) $(if $(SKIP_VALIDATION),--skip-validation)

# Docker targets for ONNX conversion
docker-build-converter: ## Build Docker image for ONNX conversion
	@echo "🐳 Building Docker image for ONNX conversion..."
	@if [ ! -f docker/Dockerfile.converter ]; then \
		echo "❌ Error: docker/Dockerfile.converter not found"; \
		exit 1; \
	fi
	@docker build -f docker/Dockerfile.converter -t axon-converter:latest .
	@echo "✅ Docker image built: axon-converter:latest"

docker-push-converter: docker-build-converter ## Build and push Docker image to registry
	@echo "📤 Pushing Docker image to registry..."
	@docker tag axon-converter:latest ghcr.io/mlOS-foundation/axon-converter:latest
	@docker push ghcr.io/mlOS-foundation/axon-converter:latest
	@echo "✅ Docker image pushed to registry"

docker-test-converter: docker-build-converter ## Test Docker converter image
	@echo "🧪 Testing Docker converter image..."
	@docker run --rm --entrypoint="" axon-converter:latest python3 --version
	@docker run --rm --entrypoint="" axon-converter:latest python3 -c "import torch; print('PyTorch:', torch.__version__)"
	@docker run --rm --entrypoint="" axon-converter:latest python3 -c "import transformers; print('Transformers:', transformers.__version__)"
	@docker run --rm --entrypoint="" axon-converter:latest python3 -c "import tensorflow as tf; print('TensorFlow:', tf.__version__)"
	@echo "✅ Docker converter image test passed"

.DEFAULT_GOAL := help
