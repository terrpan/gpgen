package templates

import (
	"fmt"

	"github.com/terrpan/gpgen/pkg/config"
	"github.com/terrpan/gpgen/pkg/models"
)

// Alias shared types from pkg/models for clarity
// Template, Input, and Step are now aliased from pkg/models
// Template represents a golden path template
type Template = models.Template

// Input defines a template input parameter
// Step represents a workflow step in a template
type Input = models.Input
type Step = models.Step

// TemplateManager handles template loading and management
type TemplateManager struct {
	templatesDir string
	templates    map[string]*Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesDir string) *TemplateManager {
	return &TemplateManager{
		templatesDir: templatesDir,
		templates:    make(map[string]*Template),
	}
}

// LoadTemplate loads a template by name
func (tm *TemplateManager) LoadTemplate(name string) (*Template, error) {
	// Check if already loaded
	if template, exists := tm.templates[name]; exists {
		return template, nil
	}

	// For now, return built-in templates
	template, err := getBuiltinTemplate(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", name, err)
	}

	tm.templates[name] = template
	return template, nil
}

// ListTemplates returns available template names
func (tm *TemplateManager) ListTemplates() []string {
	return []string{"node-app", "go-service", "python-app"}
}

// ValidateInputs validates that provided inputs match template requirements
func (tm *TemplateManager) ValidateInputs(templateName string, inputs map[string]interface{}) error {
	template, err := tm.LoadTemplate(templateName)
	if err != nil {
		return err
	}

	// Check required inputs
	for inputName, inputDef := range template.Inputs {
		value, provided := inputs[inputName]

		if inputDef.Required && !provided {
			return fmt.Errorf("required input '%s' not provided", inputName)
		}

		if provided {
			if err := validateInputValue(inputName, value, inputDef); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateInputValue(name string, value interface{}, def Input) error {
	switch def.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("input '%s' must be a string", name)
		}
	case "number":
		switch value.(type) {
		case int, float64:
			// OK
		default:
			return fmt.Errorf("input '%s' must be a number", name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("input '%s' must be a boolean", name)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("input '%s' must be an array", name)
		}
	}

	// Validate options if provided
	if len(def.Options) > 0 {
		strValue := fmt.Sprintf("%v", value)
		for _, option := range def.Options {
			if strValue == option {
				return nil
			}
		}
		return fmt.Errorf("input '%s' must be one of: %v", name, def.Options)
	}

	return nil
}

// getBuiltinTemplate returns built-in template definitions
func getBuiltinTemplate(name string) (*Template, error) {
	switch name {
	case "node-app":
		return getNodeAppTemplate(), nil
	case "go-service":
		return getGoServiceTemplate(), nil
	case "python-app":
		return getPythonAppTemplate(), nil
	default:
		return nil, fmt.Errorf("unknown template: %s", name)
	}
}

func getNodeAppTemplate() *Template {
	// Create base inputs for Node.js language
	baseInputs := map[string]Input{
		"nodeVersion":    createLanguageVersionInput("Node.js", config.DefaultValues["nodeVersion"].(string), config.LanguageVersions["node"]),
		"packageManager": createPackageManagerInput(config.DefaultValues["packageManager"].(map[string]string)["node"], config.PackageManagers["node"]),
		"testCommand":    createCommandInput("Command to run tests", config.DefaultValues["testCommand"].(map[string]string)["node"], true),
		"buildCommand":   createCommandInput("Command to build the application", config.DefaultValues["buildCommand"].(map[string]string)["node"], false),
	}

	// Merge with security and container inputs
	allInputs := mergeInputs(baseInputs, createSecurityInputs(), createContainerInputs())

	// Create base steps
	steps := []Step{
		createCheckoutStep(),
		{
			ID:   "setup-node",
			Name: "Setup Node.js",
			Uses: "actions/setup-node@v4",
			With: map[string]string{
				"node-version": "{{ .Inputs.nodeVersion }}",
				"cache":        "{{ .Inputs.packageManager }}",
			},
		},
		{
			ID:   "install",
			Name: "Install dependencies",
			Run:  "{{ .Inputs.packageManager }} {{ if eq .Inputs.packageManager \"npm\" }}ci{{ else }}install --frozen-lockfile{{ end }}",
		},
		{
			ID:   "test",
			Name: "Run tests",
			Run:  "{{ .Inputs.testCommand }}",
		},
		{
			ID:   "build",
			Name: "Build application",
			Run:  "{{ .Inputs.buildCommand }}",
			If:   "{{ .Inputs.buildCommand }}",
		},
	}

	// Add security and container steps
	steps = append(steps, createSecuritySteps()...)
	steps = append(steps, createContainerSteps()...)

	return &Template{
		Name:        "node-app",
		Description: "Node.js application with testing, building, and deployment",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"nodejs", "javascript", "web"},
		Inputs:      allInputs,
		Steps:       steps,
	}
}

func getGoServiceTemplate() *Template {
	// Create base inputs for Go language
	baseInputs := map[string]Input{
		"goVersion":    createLanguageVersionInput("Go", "1.21", []string{"1.21", "1.22", "1.23", "1.24"}),
		"testCommand":  createCommandInput("Command to run tests", "go test ./...", true),
		"buildCommand": createCommandInput("Command to build the service", "go build -o bin/service ./cmd/service", true),
		"platforms": {
			Type:        "string",
			Description: "Target platforms for cross-compilation",
			Default:     "linux/amd64,darwin/amd64",
			Required:    false,
		},
	}

	// Merge with security and container inputs
	allInputs := mergeInputs(baseInputs, createSecurityInputs(), createContainerInputs())

	// Create base steps
	steps := []Step{
		createCheckoutStep(),
		{
			ID:   "setup-go",
			Name: "Setup Go",
			Uses: "actions/setup-go@v4",
			With: map[string]string{
				"go-version": "{{ .Inputs.goVersion }}",
				"cache":      "true",
			},
		},
		{
			ID:   "test",
			Name: "Run tests",
			Run:  "{{ .Inputs.testCommand }}",
		},
		{
			ID:   "build",
			Name: "Build service",
			Run:  "{{ .Inputs.buildCommand }}",
		},
	}

	// Add security and container steps
	steps = append(steps, createSecuritySteps()...)
	steps = append(steps, createContainerSteps()...)

	return &Template{
		Name:        "go-service",
		Description: "Go service with testing, building, and cross-compilation",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"go", "golang", "service", "api"},
		Inputs:      allInputs,
		Steps:       steps,
	}
}

func getPythonAppTemplate() *Template {
	// Create base inputs for Python language
	baseInputs := map[string]Input{
		"pythonVersion":  createLanguageVersionInput("Python", config.DefaultValues["pythonVersion"].(string), config.LanguageVersions["python"]),
		"packageManager": createPackageManagerInput(config.DefaultValues["packageManager"].(map[string]string)["python"], config.PackageManagers["python"]),
		"testCommand":    createCommandInput("Command to run tests", config.DefaultValues["testCommand"].(map[string]string)["python"], true),
		"lintCommand":    createCommandInput("Command to run linting", config.DefaultValues["lintCommand"].(map[string]string)["python"], false),
		"requirements": {
			Type:        "string",
			Description: "Requirements file path",
			Default:     config.DefaultValues["requirements"].(map[string]string)["python"],
			Required:    true,
		},
	}

	// Merge with security and container inputs
	allInputs := mergeInputs(baseInputs, createSecurityInputs(), createContainerInputs())

	// Create base steps
	steps := []Step{
		createCheckoutStep(),
		{
			ID:   "setup-python",
			Name: "Setup Python",
			Uses: "actions/setup-python@v4",
			With: map[string]string{
				"python-version": "{{ .Inputs.pythonVersion }}",
				"cache":          "{{ .Inputs.packageManager }}",
			},
		},
		{
			ID:   "install",
			Name: "Install dependencies",
			Run:  "{{ if eq .Inputs.packageManager \"pip\" }}pip install -r {{ .Inputs.requirements }}{{ else if eq .Inputs.packageManager \"poetry\" }}poetry install{{ else }}pipenv install{{ end }}",
		},
		{
			ID:   "lint",
			Name: "Run linting",
			Run:  "{{ .Inputs.lintCommand }}",
			If:   "{{ .Inputs.lintCommand }}",
		},
		{
			ID:   "test",
			Name: "Run tests",
			Run:  "{{ .Inputs.testCommand }}",
		},
	}

	// Add security and container steps
	steps = append(steps, createSecuritySteps()...)
	steps = append(steps, createContainerSteps()...)

	return &Template{
		Name:        "python-app",
		Description: "Python application with testing, linting, and packaging",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"python", "web", "application"},
		Inputs:      allInputs,
		Steps:       steps,
	}
}

// Helper functions for creating common inputs and steps

// createLanguageVersionInput creates a version input for a programming language
func createLanguageVersionInput(language string, defaultVersion string, versions []string) Input {
	return Input{
		Type:        "string",
		Description: fmt.Sprintf("%s version to use", language),
		Default:     defaultVersion,
		Required:    true,
		Options:     versions,
	}
}

// createPackageManagerInput creates a package manager input
func createPackageManagerInput(defaultManager string, options []string) Input {
	return Input{
		Type:        "string",
		Description: "Package manager to use",
		Default:     defaultManager,
		Required:    true,
		Options:     options,
	}
}

// createCommandInput creates a command input
func createCommandInput(description string, defaultCmd string, required bool) Input {
	return Input{
		Type:        "string",
		Description: description,
		Default:     defaultCmd,
		Required:    required,
	}
}

// createSecurityInputs creates the standard security configuration inputs
func createSecurityInputs() map[string]Input {
	return map[string]Input{
		"security": {
			Type:        "object",
			Description: "Security scanning configuration",
			Default: map[string]interface{}{ // default security settings
				"trivy": map[string]interface{}{ // trivy object
					"enabled":  true,
					"severity": "CRITICAL,HIGH",
					"exitCode": "1",
				},
			},
			Required: false,
		},
	}
}

// createContainerInputs creates the standard container configuration inputs
func createContainerInputs() map[string]Input {
	return map[string]Input{
		"container": {
			Type:        "object",
			Description: "Container building and registry configuration",
			Default: map[string]interface{}{ // default container settings
				"enabled":      false,
				"registry":     "ghcr.io",
				"imageName":    "${{ github.repository }}",
				"imageTag":     "${{ github.sha }}",
				"dockerfile":   "Dockerfile",
				"buildContext": ".",
				"buildArgs":    "{}",
				"push": map[string]interface{}{ // push settings
					"enabled":      true,
					"onProduction": true,
				},
				"build": map[string]interface{}{ // build settings
					"alwaysBuild":  false,
					"alwaysPush":   false,
					"onPR":         true,
					"onProduction": true,
				},
			},
			Required: false,
		},
	}
}

// mergeInputs merges multiple input maps
func mergeInputs(inputMaps ...map[string]Input) map[string]Input {
	result := make(map[string]Input)
	for _, inputMap := range inputMaps {
		for key, value := range inputMap {
			result[key] = value
		}
	}
	return result
}

// Common step definitions

// createCheckoutStep creates a standard checkout step
func createCheckoutStep() Step {
	return Step{
		ID:   "checkout",
		Name: "Checkout code",
		Uses: "actions/checkout@v4",
	}
}

// createSecuritySteps creates standard security scanning steps
func createSecuritySteps() []Step {
	return []Step{
		{
			ID:   "security-scan",
			Name: "Run Trivy vulnerability scanner",
			Uses: "aquasecurity/trivy-action@master",
			With: map[string]string{
				"scan-type": "fs",
				"scan-ref":  ".",
				"format":    "sarif",
				"output":    "trivy-results.sarif",
				"severity":  "{{ .Inputs.security.trivy.severity }}",
				"exit-code": "1",
			},
			If: "{{ .Inputs.security.trivy.enabled }}",
		},
		{
			ID:   "upload-sarif",
			Name: "Upload Trivy scan results to GitHub Security tab",
			Uses: "github/codeql-action/upload-sarif@v3",
			With: map[string]string{
				"sarif_file": "trivy-results.sarif",
			},
			If: "{{ .Inputs.security.trivy.enabled }} && always()",
		},
	}
}

// createContainerSteps creates standard container building steps
func createContainerSteps() []Step {
	return []Step{
		{
			ID:   "setup-docker-buildx",
			Name: "Set up Docker Buildx",
			Uses: "docker/setup-buildx-action@v3",
			If:   "{{ .Inputs.container.enabled }} && ({{ .Inputs.container.build.alwaysBuild }} || ({{ .Inputs.container.build.onPR }} && github.event_name == 'pull_request') || ({{ .Inputs.container.build.onProduction }} && (github.event_name == 'push' && startsWith(github.ref, 'refs/tags/') || github.event_name == 'release')))",
		},
		{
			ID:   "login-registry",
			Name: "Log in to Container Registry",
			Uses: "docker/login-action@v3",
			With: map[string]string{
				"registry": "{{ .Inputs.container.registry }}",
				"username": "GITHUB_ACTOR_PLACEHOLDER",
				"password": "GITHUB_TOKEN_PLACEHOLDER",
			},
			If: "{{ .Inputs.container.enabled }} && {{ .Inputs.container.push.enabled }} && ({{ .Inputs.container.push.alwaysPush }} || ({{ .Inputs.container.push.onProduction }} && (github.event_name == 'push' && startsWith(github.ref, 'refs/tags/') || github.event_name == 'release')))",
		},
		{
			ID:   "build-and-push",
			Name: "Build and push container image",
			Uses: "docker/build-push-action@v5",
			With: map[string]string{
				"context":    "{{ .Inputs.container.buildContext }}",
				"file":       "{{ .Inputs.container.dockerfile }}",
				"push":       "{{ .Inputs.container.push.enabled }}",
				"tags":       "{{ .Inputs.container.registry }}/{{ .Inputs.container.imageName }}:{{ .Inputs.container.imageTag }}",
				"build-args": "{{ .Inputs.container.buildArgs }}",
				"cache-from": "type=gha",
				"cache-to":   "type=gha,mode=max",
			},
			If: "{{ .Inputs.container.enabled }} && ({{ .Inputs.container.build.alwaysBuild }} || ({{ .Inputs.container.build.onPR }} && github.event_name == 'pull_request') || ({{ .Inputs.container.build.onProduction }} && (github.event_name == 'push' && startsWith(github.ref, 'refs/tags/') || github.event_name == 'release')))",
		},
	}
}
