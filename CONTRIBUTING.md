# Contributing to Ninjops

Thank you for your interest in contributing to ninjops! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)
- Docker (optional, for containerized testing)
- golangci-lint (for linting)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/ninjops/ninjops.git
cd ninjops

# Install dependencies
go mod download

# Build the binary
make build
# or: go build -o bin/ninjops ./cmd/ninjops

# Run tests
make test
# or: go test -race -cover ./...

# Run linter
make lint
# or: golangci-lint run
```

### Project Structure

```
ninjops/
├── cmd/ninjops/          # Main application entry point
├── internal/
│   ├── app/              # CLI commands and UX
│   ├── config/           # Configuration loading
│   ├── spec/             # QuoteSpec types and validation
│   ├── generate/         # Template rendering
│   ├── agents/           # AI providers
│   ├── invoiceninja/     # API client
│   ├── store/            # Local state management
│   └── httpx/            # HTTP client utilities
├── docs/                 # Documentation
├── examples/             # Example QuoteSpec files
└── testdata/             # Test fixtures
```

## How to Contribute

### Reporting Bugs

Before submitting a bug report:
1. Check existing issues to avoid duplicates
2. Use the bug report template
3. Include:
   - Go version (`go version`)
   - OS and architecture
   - Steps to reproduce
   - Expected vs actual behavior
   - Logs (with secrets redacted!)

### Suggesting Enhancements

1. Check existing issues for similar suggestions
2. Use the feature request template
3. Describe:
   - The problem you're trying to solve
   - Your proposed solution
   - Alternatives you've considered

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our code style
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Run the test suite** to ensure all tests pass
6. **Submit a pull request**

#### PR Checklist

- [ ] Code follows project style (run `make lint`)
- [ ] Tests pass (run `make test`)
- [ ] New code has tests
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow conventional commits

## Code Style

### Go Code

- Run `gofmt -s` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Address all golangci-lint warnings
- Write self-documenting code with clear names

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding/updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(agents): add support for new AI provider
fix(invoiceninja): handle pagination correctly
docs(readme): update installation instructions
test(spec): add edge case tests for validation
```

### Testing

- Write unit tests for new packages
- Use table-driven tests for validation
- Mock external dependencies
- Aim for high coverage on critical paths

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "valid-input", false},
        {"invalid", "bad-input", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Development Workflow

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/spec/...

# With coverage
make test-coverage

# Integration tests
make test-integration
```

### Linting

```bash
# Run all linters
make lint

# Auto-fix issues
golangci-lint run --fix
```

### Building

```bash
# Local build
make build

# All platforms
make build-all

# Docker
make docker
```

## Documentation

- Update README.md for user-facing changes
- Update docs/ for architectural changes
- Add inline comments for complex logic
- Update CLAUDE.md for AI assistant instructions

## Release Process

Maintainers handle releases:
1. Update version in code
2. Update CHANGELOG.md
3. Create git tag
4. Build and publish binaries
5. Create GitHub release

## Getting Help

- **Documentation**: Check docs/ directory
- **Issues**: Search existing issues first
- **Discussions**: Use GitHub Discussions for questions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to ninjops!
