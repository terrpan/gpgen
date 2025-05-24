# Getting Started with GPGen

This guide will help you get up and running with GPGen quickly.

## Prerequisites

- **Go 1.21+**: Required for building and running GPGen
- **Git**: For version control and repository management
- **GitHub CLI** (optional): For enhanced GitHub integration

## Installation

### From Source
```bash
# Clone the repository
git clone https://github.com/your-org/gpgen.git
cd gpgen

# Install dependencies
go mod download

# Build the CLI
go build -o bin/gpgen ./cmd/gpgen

# Add to your PATH (optional)
export PATH=$PATH:$(pwd)/bin
```

### Using Go Install
```bash
# Install directly from source
go install github.com/your-org/gpgen/cmd/gpgen@latest
```

## Quick Start

### 1. Initialize a New Project

```bash
# Initialize with Node.js template
gpgen init node-app my-web-app

# Initialize with Go service template
gpgen init go-service my-api

# Initialize with Python template
gpgen init python-app my-python-service

# List available templates
gpgen init --list-templates
```

### 2. Customize Your Manifest

Edit the generated `manifest.yaml`:

```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-web-app
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: npm
    testCommand: "npm run test:ci"
    buildCommand: "npm run build"
```

### 3. Generate Workflows

```bash
# Generate all environments
gpgen generate manifest.yaml

# Generate specific environment
gpgen generate manifest.yaml --environment production

# Dry run (preview without creating files)
gpgen generate manifest.yaml --dry-run

# Custom output directory
gpgen generate manifest.yaml --output .workflows/
```

### 4. Validate Configuration

```bash
# Basic validation
gpgen validate manifest.yaml

# Strict validation for production
gpgen validate manifest.yaml --strict

# Quiet mode (errors only)
gpgen validate manifest.yaml --quiet
```

## Available Commands

### `gpgen init`
Create a new pipeline manifest from a template:

```bash
# Initialize with built-in template
gpgen init node-app my-project

# List available templates
gpgen init --list-templates
```

### `gpgen validate`
Validate a manifest file:

```bash
# Basic validation
gpgen validate manifest.yaml

# Strict validation for production
gpgen validate manifest.yaml --strict

# Quiet mode (errors only)
gpgen validate manifest.yaml --quiet
```

### `gpgen generate`
Generate GitHub Actions workflows:

```bash
# Generate all environments
gpgen generate manifest.yaml

# Generate specific environment
gpgen generate manifest.yaml --environment production

# Dry run (preview without creating files)
gpgen generate manifest.yaml --dry-run

# Custom output directory
gpgen generate manifest.yaml --output .workflows/
```

## Real-World Example

Here's a complete example for a production Node.js API:

```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: ecommerce-api
  annotations:
    gpgen.dev/validation-mode: strict
    gpgen.dev/description: "Production-ready Node.js API pipeline"
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: npm
    testCommand: "npm run test:ci"

  # Custom steps for security and compliance
  customSteps:
    - name: security-scan
      position: after:test
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: security-results.sarif

    - name: dependency-check
      position: before:build
      run: npm audit --audit-level high

    - name: custom-deploy
      position: replace:deploy
      uses: ./.github/actions/custom-deploy
      with:
        environment: ${{ env.ENVIRONMENT }}

  # Environment-specific configurations
  environments:
    staging:
      inputs:
        testCommand: "npm run test:integration"

    production:
      inputs:
        nodeVersion: "20"
        testCommand: "npm run test:all"
      customSteps:
        - name: performance-test
          position: after:test
          run: npm run test:performance
```

This manifest generates environment-specific workflows with security scanning, dependency checks, and custom deployment logic while maintaining the golden path structure.

## Development Workflow

If you're contributing to GPGen:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely with race detection
go test -v -race ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...
```

## Next Steps

- Check out the [Templates Reference](TEMPLATES.md) for detailed template information
- Review [Architecture Documentation](ARCHITECTURE.md) for technical details
- See [examples/](../examples/) for more manifest examples
- Read [Architecture Decision Records](decisions/) for design context
