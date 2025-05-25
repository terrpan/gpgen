# Templates Reference

This document provides comprehensive information about GPGen's built-in templates and how to use them effectively.

## Available Templates

### Node.js Template (`node-app`)
**Perfect for**: Web applications, APIs, npm packages
**Included Steps**: Checkout, Node.js setup, dependency installation, testing, building

**Configurable Inputs**:
- `nodeVersion`: Node.js version (default: "18")
- `packageManager`: npm, yarn, or pnpm (default: "npm")
- `testCommand`: Test execution command (default: "npm test")
- `buildCommand`: Build command (default: "npm run build")
- `cacheStrategy`: Dependency caching strategy (default: "npm")

**Example Manifest**:
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
```

### Go Template (`go-service`)
**Perfect for**: Microservices, CLI tools, backend services
**Included Steps**: Checkout, Go setup, module download, testing, building, security scanning, container building

**Configurable Inputs**:
- `goVersion`: Go version (default: "1.21", supports: "1.21", "1.22", "1.23", "1.24")
- `testCommand`: Test execution command (default: "go test ./...")
- `buildCommand`: Build command (default: "go build -o bin/app")
- `security.trivy.enabled`: Enable Trivy vulnerability scanning (default: true)
- `security.trivy.severity`: Security scan severity levels (default: "CRITICAL,HIGH")
- `container.enabled`: Enable container image building and pushing (default: false)
- `container.registry`: Container registry to push images to (default: "ghcr.io")
- `container.imageName`: Base name for container images (default: "${{ github.repository }}")
- `container.imageTag`: Tag for container images (default: "${{ github.sha }}")
- `container.dockerfile`: Path to the Dockerfile (default: "Dockerfile")
- `container.buildContext`: Context for container build (default: ".")
- `container.buildArgs`: Additional container build arguments (default: "{}")
- `container.push.enabled`: Enable container image push to registry (default: true)

**Automatic Security Integration**:
- Uses `security.trivy.enabled` and `security.trivy.severity` to configure Trivy scanning
- Automatically adds required GitHub permissions (`contents: read`, `security-events: write`)

**Automatic Container Integration**:
- Uses `container.enabled` and related settings to configure container build and push
- Sets up Docker Buildx and registry login automatically

**Example Manifest**:
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-go-service
spec:
  template: go-service
  inputs:
    security:
      trivy:
        enabled: true
        severity: "CRITICAL,HIGH"
    container:
      enabled: true
```

### Python Template (`python-app`)
**Perfect for**: Web applications, data processing, ML services
**Included Steps**: Checkout, Python setup, dependency installation, testing

**Configurable Inputs**:
- `pythonVersion`: Python version (default: "3.11")
- `dependencyFile`: Requirements file (default: "requirements.txt")
- `testCommand`: Test execution command (default: "pytest")
- `installCommand`: Install command (default: "pip install -r requirements.txt")

**Example Manifest**:
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-python-app
spec:
  template: python-app
  inputs:
    pythonVersion: "3.11"
    testCommand: "pytest --cov=src"
```

## Security Features

GPGen includes built-in security scanning capabilities designed for enterprise compliance and developer productivity.

### 🛡️ **Automated Vulnerability Scanning**

The `go-service` template includes **Trivy security scanning** with automatic GitHub Security tab integration:

```yaml
spec:
  template: go-service
  inputs:
    trivyScanEnabled: true  # Default: enabled
    trivySeverity: "CRITICAL,HIGH"  # Configurable severity levels
```

**Automatic Features**:
- **GitHub Permissions**: Automatically adds `contents: read` and `security-events: write` permissions
- **SARIF Upload**: Security results are uploaded to GitHub's Security tab for tracking
- **Compliance Ready**: SARIF format works with enterprise security workflows
- **Flexible Thresholds**: Configure which severity levels block deployments

### **Environment-Specific Security**

Configure different security policies per environment:

```yaml
environments:
  staging:
    inputs:
      trivySeverity: "CRITICAL,HIGH,MEDIUM"  # Comprehensive scanning for staging
  production:
    inputs:
      trivySeverity: "CRITICAL"  # Only critical issues block production
```

### **Security Best Practices**

- **Zero Configuration**: Security scanning enabled by default in `go-service` template
- **Non-Blocking Development**: Staging environments can scan for all severities
- **Production Safety**: Production deployments focus on critical vulnerabilities
- **Audit Trail**: All security findings tracked in GitHub Security tab
- **Extensible**: Add custom security scanning steps for other templates

## Custom Templates

You can create your own templates by adding them to the templates directory. Each template consists of:
- Template definition with steps and input schema
- Input validation rules and defaults
- Environment-specific trigger configurations

For detailed information on creating custom templates, see the [Architecture Documentation](ARCHITECTURE.md).
