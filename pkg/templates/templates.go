package templates

import (
	"fmt"
)

// Template represents a golden path template
type Template struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version"`
	Author      string           `yaml:"author"`
	Tags        []string         `yaml:"tags"`
	Inputs      map[string]Input `yaml:"inputs"`
	Steps       []Step           `yaml:"steps"`
}

// Input defines a template input parameter
type Input struct {
	Type        string      `yaml:"type"` // string, number, boolean, array
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Required    bool        `yaml:"required"`
	Options     []string    `yaml:"options,omitempty"` // For enum-like inputs
	Pattern     string      `yaml:"pattern,omitempty"` // For validation
}

// Step represents a workflow step in a template
type Step struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Uses        string            `yaml:"uses,omitempty"`
	Run         string            `yaml:"run,omitempty"`
	With        map[string]string `yaml:"with,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	If          string            `yaml:"if,omitempty"`
	TimeoutMins int               `yaml:"timeout-minutes,omitempty"`
	Position    string            `yaml:"position,omitempty"` // Internal: for step ordering
}

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
	return &Template{
		Name:        "node-app",
		Description: "Node.js application with testing, building, and deployment",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"nodejs", "javascript", "web"},
		Inputs: map[string]Input{
			"nodeVersion": {
				Type:        "string",
				Description: "Node.js version to use",
				Default:     "18",
				Required:    true,
				Options:     []string{"16", "18", "20", "22"},
			},
			"packageManager": {
				Type:        "string",
				Description: "Package manager to use",
				Default:     "npm",
				Required:    true,
				Options:     []string{"npm", "yarn", "pnpm"},
			},
			"testCommand": {
				Type:        "string",
				Description: "Command to run tests",
				Default:     "npm test",
				Required:    true,
			},
			"buildCommand": {
				Type:        "string",
				Description: "Command to build the application",
				Default:     "npm run build",
				Required:    false,
			},
		},
		Steps: []Step{
			{
				ID:   "checkout",
				Name: "Checkout code",
				Uses: "actions/checkout@v4",
			},
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
		},
	}
}

func getGoServiceTemplate() *Template {
	return &Template{
		Name:        "go-service",
		Description: "Go service with testing, building, and cross-compilation",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"go", "golang", "service", "api"},
		Inputs: map[string]Input{
			"goVersion": {
				Type:        "string",
				Description: "Go version to use",
				Default:     "1.21",
				Required:    true,
				Options:     []string{"1.21", "1.22", "1.23", "1.24"},
			},
			"testCommand": {
				Type:        "string",
				Description: "Command to run tests",
				Default:     "go test ./...",
				Required:    true,
			},
			"buildCommand": {
				Type:        "string",
				Description: "Command to build the service",
				Default:     "go build -o bin/service ./cmd/service",
				Required:    true,
			},
			"platforms": {
				Type:        "string",
				Description: "Target platforms for cross-compilation",
				Default:     "linux/amd64,darwin/amd64",
				Required:    false,
			},
			"trivyScanEnabled": {
				Type:        "boolean",
				Description: "Enable Trivy vulnerability scanning",
				Default:     true,
				Required:    false,
			},
			"trivySeverity": {
				Type:        "string",
				Description: "Trivy scan severity levels",
				Default:     "CRITICAL,HIGH",
				Required:    false,
				Options:     []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "CRITICAL,HIGH", "CRITICAL,HIGH,MEDIUM"},
			},
		},
		Steps: []Step{
			{
				ID:   "checkout",
				Name: "Checkout code",
				Uses: "actions/checkout@v4",
			},
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
			{
				ID:   "security-scan",
				Name: "Run Trivy vulnerability scanner",
				Uses: "aquasecurity/trivy-action@master",
				With: map[string]string{
					"scan-type": "fs",
					"scan-ref":  ".",
					"format":    "sarif",
					"output":    "trivy-results.sarif",
					"severity":  "{{ .Inputs.trivySeverity }}",
					"exit-code": "1",
				},
				If: "{{ .Inputs.trivyScanEnabled }}",
			},
			{
				ID:   "upload-sarif",
				Name: "Upload Trivy scan results to GitHub Security tab",
				Uses: "github/codeql-action/upload-sarif@v3",
				With: map[string]string{
					"sarif_file": "trivy-results.sarif",
				},
				If: "{{ .Inputs.trivyScanEnabled }} && always()",
			},
		},
	}
}

func getPythonAppTemplate() *Template {
	return &Template{
		Name:        "python-app",
		Description: "Python application with testing, linting, and packaging",
		Version:     "1.0.0",
		Author:      "GPGen Team",
		Tags:        []string{"python", "web", "application"},
		Inputs: map[string]Input{
			"pythonVersion": {
				Type:        "string",
				Description: "Python version to use",
				Default:     "3.11",
				Required:    true,
				Options:     []string{"3.9", "3.10", "3.11", "3.12"},
			},
			"packageManager": {
				Type:        "string",
				Description: "Package manager to use",
				Default:     "pip",
				Required:    true,
				Options:     []string{"pip", "poetry", "pipenv"},
			},
			"testCommand": {
				Type:        "string",
				Description: "Command to run tests",
				Default:     "pytest",
				Required:    true,
			},
			"lintCommand": {
				Type:        "string",
				Description: "Command to run linting",
				Default:     "flake8",
				Required:    false,
			},
			"requirements": {
				Type:        "string",
				Description: "Requirements file path",
				Default:     "requirements.txt",
				Required:    true,
			},
		},
		Steps: []Step{
			{
				ID:   "checkout",
				Name: "Checkout code",
				Uses: "actions/checkout@v4",
			},
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
		},
	}
}
