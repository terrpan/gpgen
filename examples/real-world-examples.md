# Real-world Examples

## E-commerce API Pipeline

This example shows a production-ready Node.js API with security scanning, dependency checks, and multi-environment deployment.

### manifest.yaml
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
    buildCommand: "npm run build"

  # Custom steps for security and compliance
  customSteps:
    - name: Security Scan
      position: after:test
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: security-results.sarif

    - name: Dependency Check
      position: before:build
      run: npm audit --audit-level high

    - name: Code Coverage
      position: after:test
      run: npm run coverage

    - name: Deploy to Environment
      position: after:build
      run: npm run deploy:${{ github.event.deployment.environment }}
      if: github.event_name == 'deployment'

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
        - name: Performance Tests
          position: after:test
          run: npm run test:performance
```

### Generated Workflows

#### Default Environment (.github/workflows/ecommerce-api.yml)
- Triggers on push to main/develop and pull requests
- Uses Node.js 18
- Runs unit tests with `npm run test:ci`
- Includes security scanning and code coverage

#### Production Environment (.github/workflows/ecommerce-api-production.yml)
- Triggers on tags and releases
- Uses Node.js 20 for better performance
- Runs comprehensive test suite with performance tests
- Includes all security and compliance checks

## Go Microservice Pipeline with Security Scanning

### manifest.yaml
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: user-service
  annotations:
    gpgen.dev/description: "Secure microservice with vulnerability scanning"
spec:
  template: go-service
  inputs:
    goVersion: "1.22"
    testCommand: "go test -race -coverprofile=coverage.out ./..."
    buildCommand: "CGO_ENABLED=0 go build -o bin/user-service ./cmd/server"
    # Enable Trivy security scanning with GitHub Security tab integration
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH"

  customSteps:
    - name: Static Analysis
      position: after:test
      run: |
        go vet ./...
        staticcheck ./...

    - name: Build Docker Image
      position: after:build
      uses: docker/build-push-action@v5
      with:
        context: .
        push: false
        tags: user-service:${{ github.sha }}

  environments:
    production:
      inputs:
        # Strictest security for production
        trivySeverity: "CRITICAL"
      customSteps:
        - name: Push Docker Image
          position: replace:Build Docker Image
          uses: docker/build-push-action@v5
          with:
            context: .
            push: true
            tags: |
              user-service:latest
              user-service:${{ github.sha }}
```

### Generated Workflow Features

#### Default Environment (.github/workflows/user-service.yml)
- **Automatic permissions**: `contents: read` and `security-events: write` are automatically added
- **Trivy scanning**: Runs vulnerability scanner with CRITICAL,HIGH severity
- **SARIF upload**: Security results automatically uploaded to GitHub Security tab
- **Static analysis**: Custom Go tooling integration

#### Production Environment (.github/workflows/user-service-production.yml)
- **Enhanced security**: Only CRITICAL vulnerabilities block deployment
- **Container registry**: Secure image pushing to production registry
- **All security features**: Inherits Trivy scanning with stricter settings

### Sample Generated Workflow Output

```yaml
# .github/workflows/user-service.yml (generated)
name: user-service
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write  # Automatically added for Trivy SARIF upload
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.22"
          cache: "true"
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...
      - name: Static Analysis
        run: |
          go vet ./...
          staticcheck ./...
      - name: Build service
        run: CGO_ENABLED=0 go build -o bin/user-service ./cmd/server
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: fs
          format: sarif
          output: trivy-results.sarif
          severity: CRITICAL,HIGH
          exit-code: "1"
        if: "true"
      - name: Upload Trivy scan results to GitHub Security tab
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
        if: true && always()
      - name: Build Docker Image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: user-service:${{ github.sha }}
```

## Python Data Pipeline

### manifest.yaml
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: data-processor
spec:
  template: python-app
  inputs:
    pythonVersion: "3.11"
    testCommand: "pytest tests/ --cov=src/"
    installCommand: "pip install -r requirements.txt -r requirements-dev.txt"

  customSteps:
    - name: Data Validation
      position: after:test
      run: python scripts/validate_data_schemas.py

    - name: Type Checking
      position: before:test
      run: mypy src/

    - name: Package Application
      position: after:build
      run: python setup.py sdist bdist_wheel

  environments:
    production:
      inputs:
        installCommand: "pip install -r requirements.txt"
      customSteps:
        - name: Deploy to S3
          position: after:build
          run: aws s3 sync dist/ s3://my-artifacts-bucket/data-processor/
```

## Advanced Features Examples

### Matrix Builds (Coming in Phase 3)
```yaml
spec:
  template: node-app
  matrix:
    nodeVersion: ["16", "18", "20"]
    os: ["ubuntu-latest", "windows-latest"]
    exclude:
      - nodeVersion: "16"
        os: "windows-latest"
```

### Conditional Steps
```yaml
customSteps:
  - name: Deploy Documentation
    position: after:build
    run: npm run docs:deploy
    if: |
      github.ref == 'refs/heads/main' &&
      contains(github.event.head_commit.modified, 'docs/')

  - name: Notify Slack
    position: end
    uses: 8398a7/action-slack@v3
    with:
      status: ${{ job.status }}
    if: always()
```

### Template Inheritance (Coming in Phase 3)
```yaml
spec:
  template: node-app
  extends: security-baseline
  inputs:
    nodeVersion: "18"
```

## Security Features and Permissions

### Automatic GitHub Security Tab Integration

When using the `go-service` template with Trivy scanning enabled:

```yaml
inputs:
  trivyScanEnabled: true  # Default: true
  trivySeverity: "CRITICAL,HIGH"  # Configurable severity levels
```

GPGen automatically:
1. **Adds required permissions** to the workflow:
   - `contents: read` - Required for checking out code
   - `security-events: write` - Required for uploading SARIF results

2. **Configures Trivy scanning** with:
   - Filesystem scanning of the repository
   - SARIF output format for GitHub Security tab
   - Configurable severity levels (CRITICAL, HIGH, MEDIUM, LOW)
   - Automatic upload to GitHub Security tab

3. **Handles conditional behavior**:
   - When `trivyScanEnabled: false`, no security permissions are added
   - Scanning runs on all environments unless explicitly disabled
   - SARIF upload runs even if Trivy finds vulnerabilities (`always()` condition)

### Security Best Practices

```yaml
# Recommended security configuration
spec:
  template: go-service
  inputs:
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH"  # Block on serious vulnerabilities

  environments:
    staging:
      inputs:
        trivySeverity: "CRITICAL,HIGH,MEDIUM"  # More comprehensive scanning
    production:
      inputs:
        trivySeverity: "CRITICAL"  # Only critical issues block production
```

This ensures that:
- Development gets feedback on all security issues
- Staging catches medium-severity vulnerabilities
- Production deployments are only blocked by critical security issues
- All security findings are tracked in GitHub's Security tab for compliance and auditing
