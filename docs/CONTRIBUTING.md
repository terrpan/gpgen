# Contributing to GPGen

We welcome contributions! This guide will help you get started with contributing to GPGen.

## Development Setup

### Prerequisites
- **Go 1.21+**: Required for building and running GPGen
- **Git**: For version control
- **golangci-lint**: For code linting (optional but recommended)

### Getting Started

1. **Fork and Clone**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gpgen.git
   cd gpgen
   ```

2. **Install Dependencies**:
   ```bash
   go mod download
   ```

3. **Verify Setup**:
   ```bash
   go test ./...
   ```

### Project Structure

```
gpgen/
├── cmd/gpgen/          # CLI application
├── pkg/                # Core library packages
│   ├── config/         # Centralized configuration
│   ├── manifest/       # Manifest processing
│   ├── templates/      # Template system
│   └── generator/      # Workflow generation
├── docs/               # Documentation
├── examples/           # Example manifests
└── schemas/v1/         # JSON schemas
```

## Contribution Guidelines

### 🐛 **Bug Reports**
- Use the issue template
- Include manifest examples that reproduce the issue
- Provide expected vs actual behavior
- Include GPGen version and environment details

### ✨ **Feature Requests**
- Start with a discussion in issues
- Include use cases and examples
- Consider backward compatibility
- Provide implementation suggestions if possible

### 🔧 **Code Contributions**

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Follow TDD**: Write tests first, then implementation

3. **Maintain coverage**: Ensure `go test -cover ./...` stays above 90%

4. **Update documentation**: Include relevant ADRs and README updates

5. **Submit PR**: Use the pull request template

## Code Standards

### Formatting and Linting
```bash
# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run

# Run tests with coverage
go test -cover ./...
```

### Testing Requirements
- **Test-driven development**: Write tests before implementation
- **High coverage**: Maintain above 90% test coverage
- **Integration tests**: Include end-to-end scenarios
- **Error handling**: Test error conditions and edge cases

### Documentation
- **Godoc comments**: Include for all public APIs
- **ADRs**: Major changes require Architecture Decision Records
- **Examples**: Include practical examples in documentation
- **Changelog**: Update for user-facing changes

### Error Handling
- Use structured errors with helpful messages
- Provide actionable error guidance
- Include context for debugging
- Handle edge cases gracefully

## Adding New Features

### 📁 **Adding Templates**

1. **Create template function** in `pkg/templates/templates.go`:
   ```go
   func getMyTemplateTemplate() *Template {
       // Use helper functions for consistency
       baseInputs := map[string]Input{
           "myVersion": createLanguageVersionInput("my-lang", "1.0", []string{"1.0", "1.1"}),
       }

       allInputs := mergeInputs(baseInputs, createSecurityInputs(), createContainerInputs())

       return &Template{
           Name: "my-template",
           // ... rest of template definition
       }
   }
   ```

2. **Add to template registry** in the same file

3. **Create comprehensive tests** in `pkg/templates/templates_test.go`

4. **Update documentation** in `docs/TEMPLATES.md`

### 🔧 **Adding Configuration Options**

1. **Update config** in `pkg/config/config.go`:
   ```go
   var LanguageVersions = map[string][]string{
       "my-lang": {"1.0", "1.1", "1.2"},
   }

   var DefaultValues = map[string]interface{}{
       "myLangVersion": "1.0",
   }
   ```

2. **Use in templates** via helper functions

3. **Add tests** for new configuration values

### 🌍 **Adding New Languages**

1. **Update configuration** in `pkg/config/config.go`
2. **Create template** using existing patterns
3. **Add helper functions** if needed for language-specific features
4. **Comprehensive testing** with real-world scenarios
5. **Documentation** and examples

## Testing Strategy

### Test Organization
- Unit tests alongside source code (`*_test.go`)
- Integration tests in `cmd/gpgen/`
- Helper functions for common test patterns
- Table-driven tests for comprehensive coverage

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./pkg/templates

# Verbose output with race detection
go test -v -race ./...
```

### Test Patterns
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    MyInput
        expected MyOutput
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    MyInput{Value: "test"},
            expected: MyOutput{Result: "processed"},
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := MyFunction(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Architecture Decision Records (ADRs)

Major changes require an ADR in `docs/decisions/`:

### ADR Template
```markdown
# ADR-XXX: Title

## Status
Proposed | Accepted | Deprecated | Superseded

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing or have agreed to implement?

## Consequences
What becomes easier or more difficult to do and any risks introduced by this change?
```

### When to Create an ADR
- New packages or major architectural changes
- Breaking changes to APIs
- Significant performance or security decisions
- Template system modifications
- Changes affecting backward compatibility

## Release Process

### Versioning
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Changelog
Update `CHANGELOG.md` with:
- New features and improvements
- Bug fixes
- Breaking changes
- Deprecations

## Community Guidelines

### Code of Conduct
- Be respectful and inclusive
- Focus on what's best for the community
- Use welcoming and inclusive language
- Be collaborative and helpful

### Communication
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community support
- **Pull Requests**: Code contributions and reviews
- **Documentation**: Keep it up-to-date and helpful

### Review Process
1. **Automated checks**: All CI checks must pass
2. **Code review**: At least one approving review required
3. **Testing**: Comprehensive test coverage required
4. **Documentation**: Updates included where needed

## Getting Help

- **Documentation**: Check `docs/` directory first
- **Examples**: Review `examples/` for patterns
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use for questions and community support

## Recognition

Contributors are recognized in:
- GitHub contributor graphs
- Release notes for significant contributions
- Special recognition for major features or fixes

Thank you for contributing to GPGen! 🎉
