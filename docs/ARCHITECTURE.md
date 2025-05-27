# GPGen Architecture Documentation

This document provides a comprehensive overview of GPGen's architecture, design patterns, and package organization for developers contributing to or extending the project.

## Architecture Overview

GPGen follows a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│                CLI Layer                     │  ← User Interface
│            (cmd/gpgen)                      │
├─────────────────────────────────────────────┤
│              Core Services                   │  ← Business Logic
│  ┌─────────────┬─────────────┬─────────────┐ │
│  │  Templates  │  Generator  │  Manifest   │ │
│  │    (pkg)    │    (pkg)    │    (pkg)    │ │
│  └─────────────┴─────────────┴─────────────┘ │
├─────────────────────────────────────────────┤
│            Configuration                     │  ← Shared State
│              (pkg/config)                   │
└─────────────────────────────────────────────┘
```

## Package Details

### `cmd/gpgen/` - Command Line Interface

**Purpose**: Provides the user-facing CLI application with comprehensive command handling.

**Key Files**:
- `main.go` - Application entry point, root command setup, version handling
- `init.go` - Initialize new manifest files from templates with validation
- `generate.go` - Generate GitHub Actions workflows from manifests
- `validate.go` - Validate manifest syntax, structure, and template compatibility

**Design Patterns**:
- **Command Pattern**: Each subcommand is implemented as a separate file
- **Facade Pattern**: CLI commands provide simple interfaces to complex underlying operations
- **Strategy Pattern**: Different validation modes (strict/relaxed) based on command flags

**Dependencies**:
- `github.com/spf13/cobra` for CLI framework
- All `pkg/*` packages for core functionality

### `pkg/config/` - Centralized Configuration

**Purpose**: Single source of truth for all configuration values, eliminating hardcoded constants throughout the codebase.

**Key Components**:
```go
// Language version mappings
var LanguageVersions = map[string][]string{
    "go":     {"1.21", "1.22", "1.23", "1.24"},
    "node":   {"16", "18", "20", "22"},
    "python": {"3.9", "3.10", "3.11", "3.12"},
}

// Default values for all inputs
var DefaultValues = map[string]interface{}{
    "goVersion":     "1.21",
    "nodeVersion":   "18",
    "pythonVersion": "3.11",
    // ... security, container, command defaults
}

// Package manager mappings
var PackageManagers = map[string][]string{
    "node":   {"npm", "yarn", "pnpm"},
    "python": {"pip", "poetry", "pipenv"},
}
```

**Benefits**:
- **DRY Principle**: Single location for all configuration
- **Maintainability**: Easy to update versions and defaults
- **Extensibility**: Simple to add new languages and options
- **Consistency**: Ensures all templates use the same values

### `pkg/manifest/` - Manifest Processing Engine

**Purpose**: Parse, validate, and structure Kubernetes-style YAML manifests.

**Key Responsibilities**:
1. **YAML Parsing**: Convert YAML files to Go structures
2. **Schema Validation**: Enforce manifest structure and required fields
3. **Business Logic Validation**: Validate template compatibility, step positioning
4. **Error Reporting**: Provide detailed, actionable error messages

**Validation Modes**:
- **Strict Mode**: Enforce all schema rules (production environments)
- **Relaxed Mode**: Allow flexibility for development workflows

**Key Structures**:
```go
type Manifest struct {
    APIVersion string           `yaml:"apiVersion"`
    Kind       string           `yaml:"kind"`
    Metadata   ManifestMetadata `yaml:"metadata"`
    Spec       ManifestSpec     `yaml:"spec"`
}

type ManifestSpec struct {
    Template     string                            `yaml:"template"`
    Inputs       map[string]interface{}           `yaml:"inputs,omitempty"`
    Steps        []CustomStep                     `yaml:"steps,omitempty"`
    Environments map[string]EnvironmentOverride   `yaml:"environments,omitempty"`
}
```

### `pkg/templates/` - Template System

**Purpose**: Define golden path templates with reusable components and helper functions.

**Architecture**: Post-modularization, templates use a **layered modular architecture** combining helper functions with centralized constants and condition builders:

#### Core Modules

**1. `templates.go` - Template Definitions**
```go
// Helper Functions (Reusable)
func createLanguageVersionInput(language string) map[string]Input
func createPackageManagerInput(language string) map[string]Input
func createSecurityInputs() map[string]Input
func createContainerInputs() map[string]Input
func mergeInputs(inputMaps ...map[string]Input) map[string]Input

// Template Definitions (Template-Specific)
func getNodeAppTemplate() *Template
func getPythonAppTemplate() *Template
func getGoServiceTemplate() *Template
```

**2. `conditions.go` - Centralized Constants and Condition Builders**
```go
// Type-Safe Action Version Constants
var GitHubActionVersions = ActionVersions{
    Checkout:            "actions/checkout@v4",
    SetupNode:           "actions/setup-node@v4",
    DockerSetupBuildx:   "docker/setup-buildx-action@v3",
    TrivyAction:         "aquasecurity/trivy-action@master",
    // ... all other action versions
}

// Centralized Placeholder Constants
var GitHubPlaceholders = Placeholders{
    ActorPlaceholder:    "GITHUB_ACTOR_PLACEHOLDER",
    TokenPlaceholder:    "GITHUB_TOKEN_PLACEHOLDER",
}

// Condition Builder for Complex Logic
type ConditionBuilder struct {
    conditions []string
}

// Pre-built Condition Sets
var ContainerCond = ContainerConditions{}
var SecurityCond = SecurityConditions{}
```

#### Modular Benefits

**From Helper Pattern (Phase 1)**:
- **60% Code Reduction**: Eliminated duplication across templates
- **Consistency**: All templates use identical input definitions
- **Maintainability**: Single location to update common functionality
- **Extensibility**: Easy to add new templates with consistent patterns

**From Modularization (Phase 2)**:
- **Type Safety**: Centralized constants prevent typos and version drift
- **Maintainable Conditions**: Complex conditional logic extracted into reusable builders
- **Single Source of Truth**: All action versions and placeholders centralized
- **Simplified Testing**: Modular components are easier to unit test
- **Reduced Cognitive Load**: Complex conditions replaced with descriptive method calls

#### Architecture Layers

```
┌─────────────────────────────────────────────┐
│           Template Definitions               │ ← Business Logic
│         (getNodeAppTemplate, etc.)          │
├─────────────────────────────────────────────┤
│            Helper Functions                  │ ← Reusable Components
│    (createSecurityInputs, mergeInputs)     │
├─────────────────────────────────────────────┤
│         Condition Builders                   │ ← Logic Abstraction
│  (ContainerCond, SecurityCond, Builder)    │
├─────────────────────────────────────────────┤
│        Centralized Constants                │ ← Configuration
│   (GitHubActionVersions, Placeholders)     │
└─────────────────────────────────────────────┘
```

**Template Structure**:
```go
type Template struct {
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Version     string            `yaml:"version"`
    Author      string            `yaml:"author"`
    Tags        []string          `yaml:"tags"`
    Inputs      map[string]Input  `yaml:"inputs"`
    Steps       []Step            `yaml:"steps"`
}
```

#### Example: Before vs After Modularization

**Before** (Hardcoded, Repetitive):
```go
{
    Name: "Checkout code",
    Uses: "actions/checkout@v4",  // Hardcoded version
    If:   "{{ .Inputs.container.enabled }} && ({{ .Inputs.container.build.alwaysBuild }} || ({{ .Inputs.container.build.onPR }} && github.event_name == 'pull_request') || ({{ .Inputs.container.build.onProduction }} && (github.event_name == 'push' && startsWith(github.ref, 'refs/tags/') || github.event_name == 'release')))",  // Complex hardcoded condition
}
```

**After** (Modular, Type-Safe):
```go
{
    Name: "Checkout code",
    Uses: GitHubActionVersions.Checkout,  // Centralized constant
    If:   ContainerCond.BuildCondition(), // Descriptive condition builder
}
```

### `pkg/generator/` - Workflow Generation Engine

**Purpose**: Transform validated manifests into executable GitHub Actions workflows.

**Core Responsibilities**:
1. **Template Resolution**: Load and validate template definitions
2. **Input Normalization**: Handle legacy inputs and apply defaults
3. **Environment Processing**: Apply environment-specific overrides
4. **Custom Step Injection**: Insert, replace, or modify workflow steps
5. **Workflow Rendering**: Generate final GitHub Actions YAML

**Key Features**:

#### Input Normalization
```go
func (wg *WorkflowGenerator) normalizeLegacyInputs(inputs map[string]interface{})
```
- Converts legacy flat inputs to structured objects
- Maintains backward compatibility
- Applies template defaults intelligently

#### Environment Handling
```go
func (wg *WorkflowGenerator) GetEffectiveInputs(manifest *Manifest, envName string, template *Template) map[string]interface{}
```
- Merges template defaults → manifest inputs → environment overrides
- Provides environment-specific configuration

#### Custom Step Processing
```go
func (wg *WorkflowGenerator) ApplyCustomStep(steps []Step, customStep CustomStep) []Step
```
- Supports precise positioning: `before:step`, `after:step`, `replace:step`
- Maintains workflow integrity
- Handles edge cases and validation

### `pkg/models/` - Shared Type Definitions and Input Processing

**Purpose**: Centralized location for shared types, preventing import cycles and providing type-safe input processing.

**Key Responsibilities**:
1. **Shared Types**: Template, Input, Step, and workflow configuration structures
2. **Type Safety**: Strongly typed alternatives to `map[string]interface{}`
3. **Input Processing**: Conversion and normalization of workflow inputs
4. **Default Values**: Centralized default configurations for security, containers, etc.

**Core Types**:
```go
// Shared template structures
type Template struct {
    Name        string           `yaml:"name"`
    Description string           `yaml:"description"`
    Inputs      map[string]Input `yaml:"inputs"`
    Steps       []Step           `yaml:"steps"`
}

// Strongly typed input configurations
type WorkflowInputs struct {
    NodeVersion    string          `json:"nodeVersion,omitempty"`
    Security       SecurityConfig  `json:"security,omitempty"`
    Container      ContainerConfig `json:"container,omitempty"`
    // ... other typed fields
}

type SecurityConfig struct {
    Trivy TrivyConfig `yaml:"trivy" json:"trivy"`
}

type ContainerConfig struct {
    Enabled   bool        `yaml:"enabled" json:"enabled"`
    Registry  string      `yaml:"registry" json:"registry"`
    Push      PushConfig  `yaml:"push" json:"push"`
    Build     BuildConfig `yaml:"build" json:"build"`
    // ... other container fields
}
```

**InputProcessor Features**:
- **Type-Safe Conversion**: Converts `map[string]interface{}` to strongly typed structures
- **Legacy Support**: Handles backward compatibility with old input formats
- **Normalization**: Applies defaults and handles missing values
- **Validation**: Ensures input integrity and consistency

**Benefits**:
- **Import Cycle Prevention**: Shared types break circular dependencies
- **Type Safety**: Compile-time checking prevents runtime errors
- **Code Clarity**: Explicit types make code self-documenting
- **Maintainability**: Centralized type definitions simplify changes
- **Performance**: Reduced runtime type assertions

## Design Patterns & Principles

### 1. **DRY (Don't Repeat Yourself)**
- **Implementation**: Helper functions in templates, centralized configuration
- **Result**: 60% reduction in code duplication
- **Benefit**: Single source of truth for common functionality

### 2. **Single Responsibility Principle**
- Each package has a clear, focused responsibility
- Functions are small and purpose-driven
- Clear separation between parsing, validation, and generation

### 3. **Open/Closed Principle**
- Easy to add new templates without modifying existing code
- Configuration-driven approach allows extension
- Plugin-like architecture for custom step processing

### 4. **Dependency Inversion**
- CLI depends on abstractions, not concrete implementations
- Generator accepts interfaces for flexibility
- Configuration provides contracts for templates

### 5. **Fail Fast Philosophy**
- Comprehensive validation at manifest parsing stage
- Clear error messages with actionable guidance
- Multiple validation modes for different use cases

## Testing Strategy

### Test Organization
```
├── *_test.go files alongside source code
├── Comprehensive test coverage for all packages
├── Integration tests in cmd/gpgen/
└── Unit tests for individual functions
```

### Test Patterns
1. **Table-Driven Tests**: Comprehensive scenario coverage
2. **Helper Functions**: Reusable test utilities (post-DRY refactoring)
3. **Integration Tests**: End-to-end workflow validation
4. **Error Case Testing**: Comprehensive error handling validation

### Modular Test Helpers (New)
```go
func testTemplateStructure(t *testing.T, template *Template)
func testLanguageVersionInput(t *testing.T, template *Template, language string)
func testCommonInputs(t *testing.T, template *Template)
func testCommonSteps(t *testing.T, template *Template)
```

### Template Modularization Testing
The modular template architecture includes comprehensive testing for:

1. **Constant Validation**: Ensures all action versions and placeholders are properly defined
2. **Condition Builder Testing**: Validates complex conditional logic construction
3. **Integration Testing**: Verifies templates use centralized constants instead of hardcoded values
4. **Regression Testing**: Prevents reintroduction of hardcoded values

**Test Structure**:
```go
// conditions_test.go - Unit tests for modular components
func TestGitHubActionVersions(t *testing.T)      // Action version constants
func TestGitHubPlaceholders(t *testing.T)        // Placeholder constants  
func TestConditionBuilder(t *testing.T)          // Condition building logic
func TestContainerConditions(t *testing.T)       // Container-specific conditions
func TestSecurityConditions(t *testing.T)        // Security-specific conditions

// templates_test.go - Integration tests
func TestTemplateUsesConstants(t *testing.T)     // Validates constant usage
func TestConditionBuilderIntegration(t *testing.T) // End-to-end condition testing
```

## Extension Points

### Adding New Templates
1. **Define Template**: Add function in `pkg/templates/templates.go`
2. **Use Helpers**: Leverage existing helper functions for consistency
3. **Add Tests**: Create comprehensive test cases
4. **Update Registry**: Add to template manager

### Adding New Languages
1. **Update Config**: Add to `pkg/config/config.go`
2. **Create Helpers**: Language-specific helper functions if needed
3. **Template Integration**: Use in template definitions
4. **Test Coverage**: Comprehensive test scenarios

### Custom Step Types
1. **Extend Validation**: Update position validation in `pkg/manifest/`
2. **Generator Logic**: Enhance step processing in `pkg/generator/`
3. **Template Support**: Add step templates if needed

## Performance Considerations

### Memory Efficiency
- **Lazy Loading**: Templates loaded on demand
- **Reused Structures**: Common objects shared across templates
- **Minimal Allocations**: Helper functions minimize object creation

### Execution Speed
- **Single Pass Processing**: Minimize file system operations
- **Efficient Validation**: Early termination on errors
- **Optimized Templates**: Pre-compiled template structures

## Future Architecture Considerations

### Potential Enhancements
1. **Plugin System**: Dynamic template loading
2. **Template Marketplace**: Community-contributed templates
3. **Advanced Caching**: Template and validation result caching
4. **Parallel Processing**: Multi-environment generation
5. **Template Composition**: Mixing and matching template components

### Scalability Patterns
- **Microservice Ready**: Clear package boundaries for service extraction
- **Event-Driven**: Potential for async processing
- **Stateless Design**: No shared mutable state
- **Configuration-Driven**: Runtime behavior modification

This architecture provides a solid foundation for GPGen's continued evolution while maintaining simplicity, reliability, and developer productivity.
