# ADR-001: Architecture Overview for GPGen (Golden Path Generator)

## Status
**PARTIALLY ACCEPTED** - Core design decisions made (see ADR-002)

## Date
2025-05-24

## Context
We need to create a "golden path" pipeline tool that generates GitHub Action workflows based on schemas. The tool should:
- Generate GitHub Action workflows from schema definitions
- Validate generated workflows against schemas
- Provide a consistent, repeatable way to create CI/CD pipelines
- **Accept user manifest files** with custom steps and input values for customization

## Decision Areas to Discuss

### 1. Core Architecture
**Options to consider:**
- CLI tool with schema files as input
- Web service with REST API
- Library with programmatic interface
- Hybrid approach (CLI + library)

### 2. Schema Format
**Options to consider:**
- JSON Schema
- YAML Schema
- Custom DSL
- HCL (HashiCorp Configuration Language)
- TOML

### 3. Template Engine
**Options to consider:**
- Go templates (text/template, html/template) - **CONCERNS: Limited logic, poor error messages, hard to debug**
- External templating (Jinja2, Mustache, Handlebars)
- Code generation approach (using Go's AST packages)
- Structured generation (building YAML programmatically)
- Template libraries with better ergonomics (gomplate, sprig)

### 4. Validation Strategy
**Options to consider:**
- Schema validation before generation
- Generated workflow validation
- Both input and output validation
- Integration with GitHub's workflow validation

### 5. Configuration Management
**Options to consider:**
- Single schema file per pipeline
- Multi-file schema composition
- Configuration inheritance
- Environment-specific overrides

## Questions for Discussion
1. What are the primary use cases? (team standardization, compliance, etc.)
2. Who are the target users? (developers, DevOps engineers, platform teams)
3. What level of customization should be supported?
4. Should this integrate with existing tools or be standalone?
5. What's the deployment model? (local CLI, centralized service, etc.)

## Next Steps
- Discuss and decide on architecture approach
- Choose schema format
- Define initial schema structure
- Create prototype implementation

## Notes
- Project is using Go 1.24.0
- Will be hosted at github.com/terrpan/gpgen

## Template Engine Deep Dive

### Why Go Templates Are Problematic for This Use Case

**Pain Points:**
1. **Limited Logic**: No loops with complex conditions, no proper conditionals beyond basic if/else
2. **Poor Error Messages**: Template errors are often cryptic and hard to debug
3. **No Type Safety**: Easy to make mistakes that only surface at runtime
4. **Whitespace Management**: Very difficult to generate clean, properly formatted YAML
5. **Complex Data Access**: Deeply nested data structures become unwieldy (`.Data.Config.Build.Steps`)
6. **No IDE Support**: No syntax highlighting, completion, or validation in most editors

**Example of Go Template Pain:**
```go
{{- range $index, $step := .Config.Steps }}
  {{- if and $step.Enabled (not (eq $step.Type "skip")) }}
    {{- if gt $index 0 }},{{ end }}
    "{{ $step.Name }}": {
      {{- if $step.Uses }}
      "uses": "{{ $step.Uses }}"
      {{- end }}
      {{- if and $step.Uses $step.With }},{{ end }}
      {{- if $step.With }}
      "with": { /* complex nested logic */ }
      {{- end }}
    }
  {{- end }}
{{- end }}
```

### Better Alternatives

**1. Structured Generation (Recommended)**
```go
type WorkflowBuilder struct {
    workflow *Workflow
}

func (w *WorkflowBuilder) AddStep(step Step) *WorkflowBuilder {
    w.workflow.Jobs[0].Steps = append(w.workflow.Jobs[0].Steps, step)
    return w
}

// Then marshal to YAML
```

**2. Template Libraries with Better Ergonomics**
- **gomplate**: More functions, better conditionals
- **sprig**: Rich function library
- **raymond**: Handlebars for Go (more intuitive syntax)

**3. Hybrid Approach**
- Use Go structs to build the workflow structure
- Use simple templates only for string interpolation
- Leverage Go's type system for validation

### Recommendation
Given the complexity of GitHub Actions workflows and the need for robust validation, I recommend **structured generation** with Go structs that marshal to YAML, combined with minimal templating only for dynamic values.

## User Manifest Design

### Core Concept
Users provide a **manifest file** that specifies:
1. **Pipeline Type**: Which golden path template to use (e.g., "node-app", "go-service", "python-lib")
2. **Input Values**: Configuration parameters (versions, environments, etc.)
3. **Custom Steps**: Additional or replacement steps beyond the golden path
4. **Overrides**: Modifications to default golden path behavior

### Manifest Structure Options

**Option 1: YAML Manifest (Recommended)**
```yaml
# gpgen.yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-app-pipeline

spec:
  template: "node-app"  # Golden path template

  inputs:
    nodeVersion: "20"
    packageManager: "npm"
    deployEnvironments: ["staging", "production"]

  customSteps:
    - name: "security-scan"
      position: "after:test"  # Insert after the test step
      uses: "securecodewarrior/github-action-add-sarif@v1"
      with:
        sarif-file: "security-scan.sarif"

    - name: "custom-deploy"
      position: "replace:deploy"  # Replace default deploy step
      run: |
        echo "Custom deployment logic"
        ./deploy.sh ${{ inputs.environment }}

  overrides:
    test:
      timeout: "30m"  # Override default test timeout
    build:
      environment:
        CUSTOM_VAR: "value"
```

**Option 2: Simple Configuration**
```yaml
# Simpler, more opinionated approach
template: node-app
nodeVersion: "20"
environments: [staging, production]

additionalSteps:
  - security-scan
  - custom-tests

customCommands:
  deploy: "./my-deploy.sh"
```

**Option 3: HCL Format (Terraform-style)**
```hcl
pipeline "my-app" {
  template = "node-app"

  inputs = {
    node_version = "20"
    environments = ["staging", "production"]
  }

  step "security-scan" {
    after = "test"
    uses   = "securecodewarrior/github-action-add-sarif@v1"
    with = {
      sarif_file = "security-scan.sarif"
    }
  }
}
```

### Workflow Generation Flow

```
User Manifest → Schema Validation → Template Selection → Golden Path Base → Apply Customizations → Generate Workflow → Validate Output
```

1. **Parse Manifest**: Load and validate user's manifest file
2. **Load Template**: Get the appropriate golden path template
3. **Merge Configuration**: Combine template defaults with user inputs
4. **Apply Customizations**: Insert/replace/modify steps as specified
5. **Generate**: Create the final GitHub Actions workflow
6. **Validate**: Ensure output is valid GitHub Actions YAML

### Benefits of Manifest Approach

1. **Declarative**: Users describe what they want, not how to build it
2. **Versionable**: Manifest can be committed to repo alongside code
3. **Reusable**: Same manifest can generate workflows for different branches/environments
4. **Auditable**: Clear record of what customizations were made
5. **Gradual Adoption**: Teams can start with golden path, add customizations over time
