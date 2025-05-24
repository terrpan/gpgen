# ADR-002: Core Design Decisions

## Status
**ACCEPTED**

## Date
2025-05-24

## Context
Based on discussions around the GPGen architecture, we need to make specific decisions about positioning strategy, template scope, override granularity, environment handling, and validation levels.

## Decisions Made

### 1. Positioning Strategy ✅ DECIDED
**Decision**: Start with a simple keyword system for ease of use, but allow for more complex selectors in the future.

**Implementation**:
```yaml
customSteps:
  - name: "security-scan"
    position: "after:test"      # Simple keywords
  - name: "deploy-check"
    position: "before:deploy"   # Simple keywords

# Future enhancement (v2):
  - name: "advanced-step"
    position:
      selector: "jobs.test.steps[?name=='unit-tests']"  # Complex selectors
      placement: "after"
```

**Rationale**:
- Simple keywords cover 80% of use cases
- Easy for developers to understand and use
- Room for growth without breaking existing manifests

### 2. Template Scope ✅ DECIDED
**Decision**: Start with a few common templates (node-app, go-service) and expand based on user feedback.

**Initial Templates**:
1. **node-app**: Node.js applications with npm/yarn, testing, building, and deployment
2. **go-service**: Go services with testing, building, and deployment

**Expansion Strategy**:
- Gather user feedback on most requested templates
- Add templates based on adoption metrics
- Consider: python-lib, docker-app, static-site, microservice

**Template Structure**:
```
templates/
├── node-app/
│   ├── template.yaml      # Base workflow structure
│   ├── schema.json        # Input validation schema
│   └── README.md          # Template documentation
└── go-service/
    ├── template.yaml
    ├── schema.json
    └── README.md
```

### 3. Override Granularity ✅ DECIDED
**Decision**: Allow overrides at both individual step properties and entire steps for maximum flexibility.

**Implementation**:
```yaml
overrides:
  # Property-level override (quick adjustments)
  test:
    timeout: "30m"
    environment:
      NODE_ENV: "test"
      CUSTOM_VAR: "value"

  # Step-level override (complete replacement)
  deploy:
    name: "Custom Deploy"
    run: |
      echo "Completely custom deployment"
      ./my-deploy-script.sh
    environment:
      DEPLOY_KEY: ${{ secrets.DEPLOY_KEY }}
```

**Benefits**:
- Quick property changes don't require redefining entire steps
- Complete step replacement available when needed
- Maintains golden path structure while allowing escape hatches

### 4. Environment Handling ✅ DECIDED
**Decision**: Support environment-specific configurations for deployment scenarios.

**Implementation**:
```yaml
spec:
  template: "node-app"

  # Global inputs
  inputs:
    nodeVersion: "20"

  # Environment-specific configurations
  environments:
    staging:
      inputs:
        deployTarget: "staging-cluster"
        replicas: 2
      customSteps:
        - name: "staging-smoke-test"
          position: "after:deploy"
          run: "npm run test:smoke:staging"

    production:
      inputs:
        deployTarget: "prod-cluster"
        replicas: 5
      overrides:
        deploy:
          timeout: "45m"
      customSteps:
        - name: "production-health-check"
          position: "after:deploy"
          run: "npm run test:health:prod"
```

**Features**:
- Environment-specific input values
- Environment-specific custom steps
- Environment-specific overrides
- Inheritance from global configuration

### 5. Validation Level ✅ DECIDED
**Decision**: Strict validation enforcing golden path patterns with optional "relaxed" mode for advanced users.

**Validation Modes**:

**Strict Mode (Default)**:
- Enforces golden path patterns and best practices
- Validates against predefined step libraries
- Prevents anti-patterns (e.g., hardcoded secrets, missing error handling)
- Ensures compliance and consistency

**Relaxed Mode (Advanced)**:
- Allows any valid GitHub Actions syntax
- Minimal validation (syntax checking only)
- For teams that need escape hatches
- Requires explicit opt-in

**Implementation**:
```yaml
# Strict mode (default)
apiVersion: gpgen.dev/v1
kind: Pipeline
spec:
  template: "node-app"
  # ... rest of config

# Relaxed mode (explicit)
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  annotations:
    gpgen.dev/validation-mode: "relaxed"
spec:
  template: "node-app"
  # ... can use any GitHub Actions syntax
```

**Validation Checks**:
- Schema validation for manifest structure
- Template compatibility
- Security best practices (no hardcoded secrets)
- Performance best practices (reasonable timeouts, caching)
- GitHub Actions syntax validation

## Impact

These decisions create a system that:
1. **Balances simplicity with power**: Easy to start, room to grow
2. **Supports real-world scenarios**: Environment-specific configs, flexible overrides
3. **Maintains consistency**: Strict validation ensures golden path compliance
4. **Allows escape hatches**: Relaxed mode for advanced cases

## Next Steps
1. Implement basic manifest parsing
2. Create node-app and go-service templates
3. Build validation engine with strict/relaxed modes
4. Design environment-specific configuration merging logic
5. Create CLI interface for manifest operations
