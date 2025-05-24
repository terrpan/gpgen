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

## Go Microservice Pipeline

### manifest.yaml
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: user-service
spec:
  template: go-service
  inputs:
    goVersion: "1.21"
    testCommand: "go test -race -coverprofile=coverage.out ./..."
    buildCommand: "CGO_ENABLED=0 go build -o bin/user-service ./cmd/server"

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
