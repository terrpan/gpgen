# GPGen - Golden Path Pipeline Generator

A tool for generating standardized GitHub Actions workflows with customizable templates and environment-specific configurations.

## Quick Start

### Installation

```bash
# From source
git clone https://github.com/your-org/gpgen.git
cd gpgen
go mod download
go build -o bin/gpgen ./cmd/gpgen

# Or via Go install
go install github.com/your-org/gpgen/cmd/gpgen@latest
```

### Usage

```bash
# Initialize a new pipeline manifest
gpgen init <template> <project-name>

# Validate a manifest
gpgen validate manifest.yaml

# Generate workflows
gpgen generate manifest.yaml --output .github/workflows/
```

For detailed guides and references, see the `docs/` directory:

- **Getting Started**: [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)
- **Templates Reference**: [docs/TEMPLATES.md](docs/TEMPLATES.md)
- **Architecture Overview**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Contribution Guide**: [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
