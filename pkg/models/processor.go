package models

import (
	"encoding/json"
	"fmt"
)

// InputProcessor handles the conversion and normalization of workflow inputs
type InputProcessor struct {
	originalInputs map[string]interface{}
}

// hasInput checks if a nested input field was present in the original inputs
func (p *InputProcessor) hasInput(keys ...string) bool {
	if p.originalInputs == nil {
		return false
	}

	var current interface{} = p.originalInputs
	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return false
		}

		v, ok := m[k]
		if !ok {
			return false
		}

		current = v
	}

	return true
}

// NewInputProcessor creates a new input processor
func NewInputProcessor() *InputProcessor {
	return &InputProcessor{}
}

// ProcessInputs converts a map[string]interface{} to strongly typed WorkflowInputs
func (p *InputProcessor) ProcessInputs(rawInputs map[string]interface{}) (*WorkflowInputs, error) {
	// Store original inputs for preserving custom fields
	p.originalInputs = make(map[string]interface{})
	for k, v := range rawInputs {
		p.originalInputs[k] = v
	}

	inputs := &WorkflowInputs{}

	// Convert map to JSON and back to struct for type safety
	jsonData, err := json.Marshal(rawInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs: %w", err)
	}

	if err := json.Unmarshal(jsonData, inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}

	// Apply normalization and defaults
	p.normalizeInputs(inputs)

	return inputs, nil
}

// normalizeInputs applies normalization rules and handles legacy inputs
func (p *InputProcessor) normalizeInputs(inputs *WorkflowInputs) {
	// Normalize security configuration
	p.normalizeSecurityConfig(inputs)

	// Normalize container configuration
	p.normalizeContainerConfig(inputs)

	// Apply default values where needed
	p.applyDefaults(inputs)
}

// normalizeSecurityConfig handles security configuration normalization
func (p *InputProcessor) normalizeSecurityConfig(inputs *WorkflowInputs) {
	// Handle legacy trivy inputs
	if inputs.TrivyScanEnabled != nil {
		inputs.Security.Trivy.Enabled = *inputs.TrivyScanEnabled
	}

	if inputs.TrivySeverity != "" {
		inputs.Security.Trivy.Severity = inputs.TrivySeverity
	}

	// Ensure security config has defaults if not set
	if inputs.Security.Trivy.Severity == "" {
		inputs.Security.Trivy.Severity = "CRITICAL,HIGH"
	}

	if inputs.Security.Trivy.ExitCode == "" {
		inputs.Security.Trivy.ExitCode = "1"
	}
}

// normalizeContainerConfig handles container configuration normalization
func (p *InputProcessor) normalizeContainerConfig(inputs *WorkflowInputs) {
	// Handle legacy container inputs
	if inputs.ContainerEnabled != nil {
		inputs.Container.Enabled = *inputs.ContainerEnabled
	}

	if inputs.ContainerRegistry != "" {
		inputs.Container.Registry = inputs.ContainerRegistry
	}

	if inputs.ContainerImageName != "" {
		inputs.Container.ImageName = inputs.ContainerImageName
	}

	if inputs.ContainerImageTag != "" {
		inputs.Container.ImageTag = inputs.ContainerImageTag
	}

	// Apply defaults if not set
	if inputs.Container.Registry == "" {
		inputs.Container.Registry = "ghcr.io"
	}

	if inputs.Container.ImageName == "" {
		inputs.Container.ImageName = "${{ github.repository }}"
	}

	if inputs.Container.ImageTag == "" {
		inputs.Container.ImageTag = "${{ github.sha }}"
	}

	if inputs.Container.Dockerfile == "" {
		inputs.Container.Dockerfile = "Dockerfile"
	}

	if inputs.Container.BuildContext == "" {
		inputs.Container.BuildContext = "."
	}

	if inputs.Container.BuildArgs == "" {
		inputs.Container.BuildArgs = "{}"
	}
}

// applyDefaults applies default values for any unset fields
func (p *InputProcessor) applyDefaults(inputs *WorkflowInputs) {
	// Set default security config if empty
	if inputs.Security.Trivy.Severity == "" && inputs.Security.Trivy.ExitCode == "" {
		inputs.Security = DefaultSecurityConfig()
	}

	// Set default container config if completely empty
	if inputs.Container.Registry == "" && inputs.Container.ImageName == "" {
		inputs.Container = DefaultContainerConfig()
	}

	// Ensure push and build configs have defaults applied per field
	def := DefaultContainerConfig()

	if !inputs.Container.Push.Enabled && !p.hasInput("container", "push", "enabled") {
		inputs.Container.Push.Enabled = def.Push.Enabled
	}
	if !inputs.Container.Push.OnProduction && !p.hasInput("container", "push", "onProduction") {
		inputs.Container.Push.OnProduction = def.Push.OnProduction
	}

	if !inputs.Container.Build.AlwaysBuild && !p.hasInput("container", "build", "alwaysBuild") {
		inputs.Container.Build.AlwaysBuild = def.Build.AlwaysBuild
	}
	if !inputs.Container.Build.AlwaysPush && !p.hasInput("container", "build", "alwaysPush") {
		inputs.Container.Build.AlwaysPush = def.Build.AlwaysPush
	}
	if !inputs.Container.Build.OnPR && !p.hasInput("container", "build", "onPR") {
		inputs.Container.Build.OnPR = def.Build.OnPR
	}
	if !inputs.Container.Build.OnProduction && !p.hasInput("container", "build", "onProduction") {
		inputs.Container.Build.OnProduction = def.Build.OnProduction
	}
}

// ToMap converts WorkflowInputs back to a map for template processing
func (p *InputProcessor) ToMap(inputs *WorkflowInputs) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert struct to JSON and back to map
	jsonData, err := json.Marshal(inputs)
	if err != nil {
		return result
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		return result
	}

	// Preserve custom fields from original inputs that aren't part of the struct
	if p.originalInputs != nil {
		knownFields := map[string]bool{
			"nodeVersion": true, "goVersion": true, "pythonVersion": true,
			"packageManager": true, "testCommand": true, "buildCommand": true,
			"lintCommand": true, "requirements": true, "platforms": true,
			"containerEnabled": true, "containerRegistry": true, "containerImageName": true,
			"containerImageTag": true, "trivyScanEnabled": true, "trivySeverity": true,
			"security": true, "container": true,
		}

		for k, v := range p.originalInputs {
			if !knownFields[k] {
				result[k] = v
			}
		}
	}

	return result
}

// GetString safely gets a string value from inputs
func (inputs *WorkflowInputs) GetString(field string) string {
	switch field {
	case "nodeVersion":
		return inputs.NodeVersion
	case "goVersion":
		return inputs.GoVersion
	case "pythonVersion":
		return inputs.PythonVersion
	case "packageManager":
		return inputs.PackageManager
	case "testCommand":
		return inputs.TestCommand
	case "buildCommand":
		return inputs.BuildCommand
	case "lintCommand":
		return inputs.LintCommand
	case "requirements":
		return inputs.Requirements
	case "platforms":
		return inputs.Platforms
	default:
		return ""
	}
}

// GetBool safely gets a boolean value from inputs
func (inputs *WorkflowInputs) GetBool(field string) bool {
	switch field {
	case "security.trivy.enabled":
		return inputs.Security.Trivy.Enabled
	case "container.enabled":
		return inputs.Container.Enabled
	case "container.push.enabled":
		return inputs.Container.Push.Enabled
	case "container.push.onProduction":
		return inputs.Container.Push.OnProduction
	case "container.build.alwaysBuild":
		return inputs.Container.Build.AlwaysBuild
	case "container.build.alwaysPush":
		return inputs.Container.Build.AlwaysPush
	case "container.build.onPR":
		return inputs.Container.Build.OnPR
	case "container.build.onProduction":
		return inputs.Container.Build.OnProduction
	default:
		return false
	}
}

// HasValue checks if a field has a non-empty value
func (inputs *WorkflowInputs) HasValue(field string) bool {
	switch field {
	case "buildCommand":
		return inputs.BuildCommand != ""
	case "lintCommand":
		return inputs.LintCommand != ""
	default:
		return inputs.GetString(field) != ""
	}
}
