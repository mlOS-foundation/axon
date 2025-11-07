.PHONY: build test install clean lint fmt help vet coverage

# Build variables
BINARY_NAME=axon
VERSION?=v0.1.0
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

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

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .
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

validate-pr: ## Run all validation checks before PR (fmt, vet, lint, test, build)
	@./validate-pr.sh

.DEFAULT_GOAL := help
