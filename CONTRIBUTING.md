# Contributing to Axon

Thank you for your interest in contributing to Axon! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. We are committed to providing a welcoming and inclusive environment.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional)

### Development Setup

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/mlOS-foundation/axon.git
   cd axon
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Build the project:
   ```bash
   make build
   ```

4. Run tests:
   ```bash
   make test
   ```

## Development Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following our coding standards

3. Write or update tests as needed

4. Ensure all tests pass:
   ```bash
   make test
   ```

5. Run linters:
   ```bash
   make lint
   ```

6. Format your code:
   ```bash
   make fmt
   ```

7. Commit your changes with clear messages:
   ```bash
   git commit -m "Add feature: description of changes"
   ```

8. Push to your fork and create a Pull Request

## Coding Standards

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Keep functions focused and small
- Write clear, descriptive comments for exported functions

### Naming Conventions

- Use neural metaphors where appropriate (see Brand Guidelines)
- Commands should be clear and intuitive
- Internal packages should be in `internal/`
- Public APIs should be in `pkg/`

### Code Organization

```
axon/
├── cmd/axon/         # CLI entry point
├── internal/         # Internal packages (not exported)
│   ├── cache/        # Cache management
│   ├── config/       # Configuration
│   ├── manifest/     # Manifest parsing
│   ├── registry/     # Registry client
│   ├── model/        # Model handling
│   └── ui/           # CLI UI
├── pkg/              # Public packages
│   ├── types/        # Public types
│   └── utils/        # Utilities
└── test/             # Tests
```

## Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Test error cases, not just happy paths

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all CI checks pass
4. Request review from maintainers
5. Address feedback promptly

### PR Title Format

```
Type: Brief description

Examples:
- feat: Add manifest validation
- fix: Fix cache cleanup bug
- docs: Update README with examples
- refactor: Simplify registry client
```

## Project Roadmap

See [MVP_STATUS.md](MVP_STATUS.md) for the current project status and completed features.

Current focus areas:
- ✅ Core types and configuration (Complete)
- ✅ Cache manager and registry client (Complete)
- ✅ CLI commands implementation (Complete)
- 🔄 Registry integration and package extraction (In Progress)

## Questions?

- Open an issue for bug reports or feature requests
- Join our Discord: https://discord.gg/mlos
- Check existing issues and discussions

Thank you for contributing to Axon! 🧠⚡

