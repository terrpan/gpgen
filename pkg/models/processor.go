package models

import (
	"encoding/json"
	"fmt"
)

// InputProcessor handles the conversion and normalization of workflow inputs
type InputProcessor struct {
	originalInputs map[string]interface{}
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

	// Determine if push/build configs were explicitly provided in inputs
	var (
		pushProvided  bool
		buildProvided bool
	)
	if rawContainer, ok := p.originalInputs["container"].(map[string]interface{}); ok {
		if _, ok := rawContainer["push"]; ok {
			pushProvided = true
		}
		if _, ok := rawContainer["build"]; ok {
			buildProvided = true
		}
	}

	// Ensure push config has sensible defaults only when not provided
	if !pushProvided && !inputs.Container.Push.Enabled && !inputs.Container.Push.OnProduction {
		inputs.Container.Push = PushConfig{
			Enabled:      true,
			OnProduction: true,
		}
	}

	// Ensure build config has sensible defaults only when not provided
	if !buildProvided && !inputs.Container.Build.OnPR && !inputs.Container.Build.OnProduction &&
		!inputs.Container.Build.AlwaysBuild && !inputs.Container.Build.AlwaysPush {
		inputs.Container.Build = BuildConfig{
			AlwaysBuild:  false,
			AlwaysPush:   false,
			OnPR:         true,
			OnProduction: true,
		}
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
