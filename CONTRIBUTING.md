# Contributing to FlareCert

Thank you for your interest in contributing to FlareCert! This document provides guidelines and information for contributors.

## Table of Contents

1. [Development Setup](#development-setup)
2. [Code Style](#code-style)
3. [Testing](#testing)
4. [Pull Request Process](#pull-request-process)
5. [Issue Reporting](#issue-reporting)

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git
- Make (optional but recommended)

### Getting Started

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/your-username/flarecert.git
cd flarecert
```

3. Set up development environment:
```bash
make setup
```

4. Create a feature branch:
```bash
git checkout -b feature/your-feature-name
```

### Environment Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your test credentials (use staging environment)

### Building and Testing

```bash
# Install dependencies
make deps

# Build the project
make build

# Run tests
make test

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint
```

## Code Style

### Go Style Guidelines

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for public functions and complex logic
- Keep functions small and focused
- Handle errors appropriately

### Example Code Style

```go
// Good: Clear function name and documentation
// GenerateCertificate requests a new SSL certificate for the specified domains
func GenerateCertificate(domains []string, config *Config) (*Certificate, error) {
    if len(domains) == 0 {
        return nil, fmt.Errorf("at least one domain is required")
    }
    
    // Implementation here
    return cert, nil
}

// Good: Error handling
cert, err := GenerateCertificate(domains, config)
if err != nil {
    return fmt.Errorf("failed to generate certificate: %w", err)
}

// Good: Variable naming
clientTimeout := 30 * time.Second
maxRetries := 3
```

### File Organization

```
internal/
├── acme/          # ACME client implementation
├── config/        # Configuration management
├── dns/           # DNS provider implementations
└── utils/         # Utility functions

cmd/              # CLI commands
├── root.go       # Root command
├── cert.go       # Certificate generation
├── renew.go      # Certificate renewal
└── list.go       # Certificate listing
```

## Testing

### Unit Tests

Write unit tests for new functionality:

```go
func TestGenerateCertificate(t *testing.T) {
    tests := []struct {
        name    string
        domains []string
        wantErr bool
    }{
        {
            name:    "valid single domain",
            domains: []string{"example.com"},
            wantErr: false,
        },
        {
            name:    "empty domains",
            domains: []string{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := GenerateCertificate(tt.domains, &Config{})
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateCertificate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

For integration tests that require Cloudflare API access:

1. Use staging Let's Encrypt environment
2. Use test domains that you control
3. Set environment variable `INTEGRATION_TESTS=true`

### Running Tests

```bash
# Unit tests only
go test ./...

# With integration tests (requires setup)
INTEGRATION_TESTS=true go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Pull Request Process

### Before Submitting

1. Ensure tests pass: `make test`
2. Format code: `make fmt`
3. Lint code: `make lint`
4. Update documentation if needed
5. Add tests for new functionality

### PR Requirements

- **Clear description**: Explain what the PR does and why
- **Issue reference**: Link to related issues
- **Tests**: Include tests for new functionality
- **Documentation**: Update README.md or other docs if needed
- **Backwards compatibility**: Avoid breaking changes unless necessary

### PR Template

```markdown
## Description
Brief description of changes

## Related Issues
Fixes #123

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. Automated checks must pass
2. At least one maintainer review required
3. Address feedback and update PR
4. Maintainer will merge when ready

## Issue Reporting

### Bug Reports

Use the bug report template:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. With configuration '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment**
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.24]
- FlareCert version: [e.g. v1.0.0]

**Additional context**
Any other context about the problem.
```

### Feature Requests

Use the feature request template:

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Other solutions you've considered.

**Additional context**
Any other context about the feature request.
```

## Development Guidelines

### Adding New DNS Providers

1. Create new file in `internal/dns/`
2. Implement the DNS provider interface
3. Add configuration options
4. Include comprehensive tests
5. Update documentation

### Adding New Commands

1. Create new file in `cmd/`
2. Follow existing command structure
3. Add to root command in `init()`
4. Include help text and examples
5. Add tests

### Error Handling

- Use wrapped errors: `fmt.Errorf("context: %w", err)`
- Provide clear error messages
- Log appropriately (verbose mode)
- Handle timeouts and retries gracefully

### Configuration

- Use environment variables for configuration
- Provide sensible defaults
- Validate configuration on startup
- Document all configuration options

## Documentation

### Code Documentation

- Document all public functions
- Include examples in doc comments
- Keep documentation up to date

### User Documentation

- Update README.md for user-facing changes
- Update command help text
- Add examples for new features
- Update deployment guide if needed

## Release Process

### Version Numbering

We use semantic versioning (semver):
- Major: Breaking changes
- Minor: New features (backwards compatible)
- Patch: Bug fixes

### Release Checklist

1. Update version in code
2. Update CHANGELOG.md
3. Create release notes
4. Tag release
5. Update documentation

## Community

### Communication

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: General questions and ideas
- Pull Requests: Code contributions

### Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and contribute
- Follow GitHub's community guidelines

## Questions?

If you have questions about contributing:

1. Check existing documentation
2. Search existing issues
3. Create a new issue with the "question" label
4. Reach out via GitHub Discussions

Thank you for contributing to FlareCert!
