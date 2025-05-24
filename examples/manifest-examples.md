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
    - name: "security-scan"
      position: "after:test"
      uses: "securecodewarrior/github-action-add-sarif@v1"
      with:
        sarif-file: "security-scan.sarif"

    - name: "dependency-audit"
      position: "before:build"
      run: |
        yarn audit --level high
        yarn audit --json > audit-results.json
      continue-on-error: false
      timeout-minutes: 10
```

## 3. Go Service with Environment-Specific Configurations

```yaml
# gpgen.yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-go-service
  annotations:
    gpgen.dev/description: "Microservice with staging and production deployments"

spec:
  template: "go-service"

  inputs:
    goVersion: "1.24"
    deployEnvironments: ["staging", "production"]

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

## 4. Advanced Node.js App with Relaxed Validation

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

## 5. Minimal Go Service

```yaml
# gpgen.yaml - Simplest possible configuration
apiVersion: gpgen.dev/v1
kind: Pipeline

spec:
  template: "go-service"

  inputs:
    goVersion: "1.24"
```
