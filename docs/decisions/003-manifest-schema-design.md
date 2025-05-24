# ADR-003: Manifest Schema Design

## Status
**ACCEPTED**

## Date
2025-05-24

## Context
We need to define a comprehensive schema for GPGen manifest files that supports all the features decided in ADR-002, including positioning strategies, environment handling, overrides, and validation modes.

## Schema Design Decisions

### 1. Kubernetes-Style API Versioning ✅
**Decision**: Use Kubernetes-style `apiVersion` and `kind` fields for future extensibility.

**Rationale**:
- Familiar pattern for DevOps engineers
- Enables schema evolution without breaking changes
- Supports multiple resource types in the future
- Clear versioning strategy

**Structure**:
```yaml
apiVersion: gpgen.dev/v1
kind: Pipeline
```

### 2. Metadata with Annotations ✅
**Decision**: Support metadata with annotations for extensibility and configuration.

**Key Annotations**:
- `gpgen.dev/validation-mode`: Controls strict vs relaxed validation
- `gpgen.dev/description`: Human-readable pipeline description
- Custom annotations allowed for future features

### 3. Template Selection ✅
**Decision**: Simple string enum for template selection with validation.

**Current Templates**:
- `node-app`: Node.js applications
- `go-service`: Go microservices

**Future**: Schema can be extended to support more templates without breaking changes.

### 4. Input Parameter Flexibility ✅
**Decision**: Use `additionalProperties: true` for inputs to support template-specific parameters.

**Benefits**:
- Templates can define their own input schemas
- Users get validation for known parameters
- Unknown parameters are allowed (validated by template)
- Future templates can add new inputs without schema changes

### 5. Position Strategy Implementation ✅
**Decision**: Use regex pattern to enforce simple positioning syntax.

**Pattern**: `^(before|after|replace):[a-z0-9-]+$`

**Examples**:
- `after:test` - Insert after the test step
- `before:deploy` - Insert before deploy step
- `replace:build` - Replace the build step entirely

**Benefits**:
- Simple and intuitive
- Validates at schema level
- Room for complex selectors in future versions

### 6. Step Definition Flexibility ✅
**Decision**: Support both `uses` (actions) and `run` (shell commands) with oneOf constraint.

**Validation**: Ensures steps have either `uses` OR `run`, but not both or neither.

**Properties**:
- Standard GitHub Actions properties supported
- Conditional execution with `if`
- Timeout control
- Error handling options

### 7. Override Granularity ✅
**Decision**: Allow overriding any step property without requiring full step redefinition.

**Structure**:
```yaml
overrides:
  step-name:
    timeout-minutes: 30
    env:
      CUSTOM_VAR: "value"
```

**Benefits**:
- Minimal changes for simple overrides
- No need to redefine entire steps
- Preserves golden path structure

### 8. Environment Configuration ✅
**Decision**: Nested environment objects with inheritance from global configuration.

**Structure**:
```yaml
environments:
  staging:
    inputs: { /* env-specific inputs */ }
    customSteps: [ /* env-specific steps */ ]
    overrides: { /* env-specific overrides */ }
```

**Inheritance Order**:
1. Global `inputs`, `customSteps`, `overrides`
2. Environment-specific configurations override globals
3. Merge strategy preserves both global and environment-specific values

## Schema Features

### Type Safety
- Enum validation for known values
- Pattern validation for position strings
- Required field validation
- Type checking for all properties

### Extensibility
- Additional properties allowed where appropriate
- Annotation system for future features
- API versioning for breaking changes
- Template-specific input validation

### Validation Modes
- **Strict Mode**: Enforces golden path patterns and best practices
- **Relaxed Mode**: Allows any valid GitHub Actions syntax
- Controlled via annotation: `gpgen.dev/validation-mode`

### Error Prevention
- oneOf constraints prevent invalid combinations
- Pattern validation catches syntax errors
- Required fields ensure minimum viable configuration
- Type validation prevents runtime errors

## Implementation Notes

### JSON Schema Choice
- Industry standard for validation
- Good tooling support (IDEs, validators)
- Easy to extend and version
- Works with YAML files (YAML is JSON superset)

### Reference Strategy
- Use JSON Schema `$ref` for reusable components
- Shared definitions for common patterns
- Reduced duplication in schema definition

### Validation Integration
- Schema can be embedded in CLI tool
- IDEs can provide validation and completion
- CI/CD can validate manifests before processing
- Clear error messages for validation failures

## Future Enhancements

### Version 2 Considerations
- Complex position selectors (JSONPath, CSS-style)
- Conditional configurations
- Template composition and inheritance
- Multi-repository pipeline definitions

### Additional Templates
- Schema designed to support unlimited template types
- Template-specific input schemas
- Template validation and discovery

## Examples
See `/examples/manifest-examples.md` for comprehensive usage examples demonstrating all schema features.

## Next Steps
1. Implement schema validation in Go
2. Create template-specific input schemas
3. Build manifest parser with validation
4. Add IDE integration (JSON Schema registration)
5. Create CLI commands for manifest operations
