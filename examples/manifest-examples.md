# Example Manifest Files for GPGen

## 1. Simple Node.js Application

```yaml
# gpgen.yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-node-app
  annotations:
    gpgen.dev/description: "CI/CD pipeline for my Node.js application"

spec:
  template: "node-app"

  inputs:
    nodeVersion: "20"
    packageManager: "npm"
    deployEnvironments: ["staging", "production"]
```

## 2. Node.js App with Custom Security Scanning

```yaml
# gpgen.yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: secure-node-app

spec:
  template: "node-app"

  inputs:
    nodeVersion: "18"
    packageManager: "yarn"
    deployEnvironments: ["staging", "production"]

  customSteps:
    # Add Trivy scanning to Node.js template (requires manual permissions setup)
    - name: "trivy-security-scan"
      position: "after:test"
      uses: "aquasecurity/trivy-action@master"
      with:
        scan-type: "fs"
        format: "sarif"
        output: "trivy-results.sarif"
        severity: "CRITICAL,HIGH"

    - name: "upload-trivy-results"
      position: "after:trivy-security-scan"
      uses: "github/codeql-action/upload-sarif@v3"
      with:
        sarif_file: "trivy-results.sarif"

    - name: "dependency-audit"
      position: "before:build"
      run: |
        yarn audit --level high
        yarn audit --json > audit-results.json
      continue-on-error: false
      timeout-minutes: 10

# Note: For Node.js templates with custom Trivy scanning, you'll need to manually
# add permissions to your workflow job:
#
# jobs:
#   build:
#     permissions:
#       contents: read
#       security-events: write
#
# The go-service template handles this automatically when trivyScanEnabled: true
```

## 3. Go Service with Security Scanning and Environment-Specific Configurations

```yaml
# gpgen.yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-go-service
  annotations:
    gpgen.dev/description: "Microservice with Trivy security scanning and deployments"

spec:
  template: "go-service"

  inputs:
    goVersion: "1.22"
    deployEnvironments: ["staging", "production"]
    # Security scanning with Trivy (automatically adds GitHub Security tab permissions)
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH"

  # Global custom step
  customSteps:
    - name: "integration-tests"
      position: "after:test"
      run: |
        go test -tags=integration ./...
      timeout-minutes: 15

  # Global overrides
  overrides:
    test:
      timeout-minutes: 20
      env:
        GO_TEST_TIMEOUT: "15m"

  # Environment-specific configurations
  environments:
    staging:
      inputs:
        deployTarget: "staging-k8s-cluster"
        replicas: 2
        # More relaxed security scanning for staging
        trivySeverity: "CRITICAL,HIGH,MEDIUM"
      customSteps:
        - name: "staging-smoke-test"
          position: "after:deploy"
          run: |
            curl -f https://api-staging.example.com/health
            ./scripts/smoke-test.sh staging
          timeout-minutes: 5

    production:
      inputs:
        deployTarget: "prod-k8s-cluster"
        replicas: 5
        # Strict security scanning for production (CRITICAL only)
        trivySeverity: "CRITICAL"
      overrides:
        deploy:
          timeout-minutes: 45
          env:
            DEPLOY_STRATEGY: "rolling"
            MAX_UNAVAILABLE: "25%"
      customSteps:
        - name: "production-health-check"
          position: "after:deploy"
          run: |
            ./scripts/health-check.sh production
            ./scripts/performance-baseline.sh
          timeout-minutes: 10
        - name: "notify-slack"
          position: "after:production-health-check"
          uses: "8398a7/action-slack@v3"
          with:
            status: "success"
            channel: "#deployments"
            webhook_url: "${{ secrets.SLACK_WEBHOOK }}"
          if: "success()"
```

## 4. Go Service with Container Building and Security

```yaml
# gpgen.yaml - Complete Go service with security scanning and container building
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: containerized-go-service
  annotations:
    gpgen.dev/validation-mode: strict
    gpgen.dev/description: "Go microservice with security scanning and container building"

spec:
  template: "go-service"

  inputs:
    goVersion: "1.23"
    testCommand: "go test -race ./..."
    buildCommand: "go build -o bin/service ./cmd/service"

    # Security scanning with Trivy (automatically adds GitHub Security tab permissions)
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH"

    # Container building (automatically adds container registry permissions)
    containerEnabled: true
    containerRegistry: "ghcr.io"
    containerImageName: "${{ github.repository }}"
    containerImageTag: "${{ github.sha }}"
    containerDockerfile: "Dockerfile"
    containerBuildContext: "."
    containerBuildArgs: "{\"GO_VERSION\":\"1.23\"}"
    containerPushEnabled: true

  environments:
    development:
      inputs:
        trivySeverity: "CRITICAL,HIGH,MEDIUM,LOW"  # Scan all severities in dev
        containerPushEnabled: false  # Don't push containers in dev

    staging:
      inputs:
        trivySeverity: "CRITICAL,HIGH,MEDIUM"
        containerImageTag: "staging-${{ github.sha }}"

    production:
      inputs:
        trivySeverity: "CRITICAL"  # Only critical issues block production
        containerImageTag: "v${{ github.ref_name }}"  # Use tag for production
        containerBuildArgs: "{\"GO_VERSION\":\"1.23\",\"BUILD_ENV\":\"production\"}"
```

**Generated Features (Automatic)**:
- **Security Permissions**: `contents: read`, `security-events: write` for Trivy scanning
- **Container Permissions**: `packages: write` for GitHub Container Registry
- **Docker Buildx**: Automatic setup for advanced building features
- **Registry Login**: Automatic authentication using GitHub token
- **Build Caching**: GitHub Actions cache optimization for faster builds

## 5. Security-First Go Service

```yaml
# gpgen.yaml - Enterprise security scanning configuration
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: secure-go-service
  annotations:
    gpgen.dev/description: "Go service with comprehensive security scanning"

spec:
  template: "go-service"

  inputs:
    goVersion: "1.23"
    # Enable Trivy vulnerability scanning (automatically adds security-events: write permission)
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH,MEDIUM"

  customSteps:
    # Additional security scanning
    - name: "gosec-security-scan"
      position: "after:test"
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec -fmt sarif -out gosec-results.sarif ./...

    - name: "upload-gosec-results"
      position: "after:gosec-security-scan"
      uses: "github/codeql-action/upload-sarif@v3"
      with:
        sarif_file: "gosec-results.sarif"
        category: "gosec"

  environments:
    staging:
      inputs:
        # Disable Trivy in staging to speed up builds (removes security permissions)
        trivyScanEnabled: false

    production:
      inputs:
        # Maximum security for production
        trivySeverity: "CRITICAL"
      customSteps:
        - name: "compliance-check"
          position: "before:deploy"
          run: |
            echo "Running compliance checks..."
            ./scripts/compliance-scan.sh
```

## 5. Advanced Node.js App with Relaxed Validation

```yaml
# gpgen.yaml - For teams that need escape hatches
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: advanced-node-app
  annotations:
    gpgen.dev/validation-mode: "relaxed"
    gpgen.dev/description: "Advanced pipeline with custom GitHub Actions syntax"

spec:
  template: "node-app"

  inputs:
    nodeVersion: "20"
    packageManager: "pnpm"

  customSteps:
    # This uses advanced GitHub Actions features that might not be in strict mode
    - name: "matrix-e2e-tests"
      position: "after:test"
      uses: "./.github/actions/custom-e2e"
      with:
        matrix: |
          {
            "browser": ["chrome", "firefox", "safari"],
            "viewport": ["desktop", "mobile"]
          }
        parallel: true

    # Replace entire deploy step with complex custom logic
    - name: "advanced-deploy"
      position: "replace:deploy"
      run: |
        # Complex deployment logic that doesn't fit golden path
        if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
          ./deploy.sh production --blue-green
        elif [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
          ./deploy.sh staging --canary
        fi
      env:
        CUSTOM_DEPLOY_KEY: "${{ secrets.CUSTOM_DEPLOY_KEY }}"
        FEATURE_FLAGS: "${{ vars.FEATURE_FLAGS }}"
```

## 6. Minimal Go Service

```yaml
# gpgen.yaml - Simplest possible configuration
apiVersion: gpgen.dev/v1
kind: Pipeline

spec:
  template: "go-service"

  inputs:
    goVersion: "1.23"
    # Trivy scanning enabled by default with CRITICAL,HIGH severity
    # Automatically adds GitHub Security tab permissions
```
