package models

// Template represents a golden path template with inputs and workflow steps
type Template struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version"`
	Author      string           `yaml:"author"`
	Tags        []string         `yaml:"tags"`
	Inputs      map[string]Input `yaml:"inputs"`
	Steps       []Step           `yaml:"steps"`
}

// Input defines a parameter for a template with stronger typing
type Input struct {
	Type        InputType   `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Required    bool        `yaml:"required"`
	Options     []string    `yaml:"options,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty"`
}

// InputType represents the type of an input parameter
type InputType string

const (
	InputTypeString  InputType = "string"
	InputTypeNumber  InputType = "number"
	InputTypeBoolean InputType = "boolean"
	InputTypeArray   InputType = "array"
	InputTypeObject  InputType = "object"
)

// Step represents a GitHub Actions workflow step
type Step struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Uses        string            `yaml:"uses,omitempty"`
	Run         string            `yaml:"run,omitempty"`
	With        map[string]string `yaml:"with,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	If          string            `yaml:"if,omitempty"`
	TimeoutMins int               `yaml:"timeout-minutes,omitempty"`
	Position    string            `yaml:"position,omitempty"`
}

// SecurityConfig represents security scanning configuration
type SecurityConfig struct {
	Trivy TrivyConfig `yaml:"trivy" json:"trivy"`
}

// TrivyConfig represents Trivy vulnerability scanner configuration
type TrivyConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Severity string `yaml:"severity" json:"severity"`
	ExitCode string `yaml:"exitCode" json:"exitCode"`
}

// ContainerConfig represents container building and registry configuration
type ContainerConfig struct {
	Enabled      bool        `yaml:"enabled" json:"enabled"`
	Registry     string      `yaml:"registry" json:"registry"`
	ImageName    string      `yaml:"imageName" json:"imageName"`
	ImageTag     string      `yaml:"imageTag" json:"imageTag"`
	Dockerfile   string      `yaml:"dockerfile" json:"dockerfile"`
	BuildContext string      `yaml:"buildContext" json:"buildContext"`
	BuildArgs    string      `yaml:"buildArgs" json:"buildArgs"`
	Push         PushConfig  `yaml:"push" json:"push"`
	Build        BuildConfig `yaml:"build" json:"build"`
}

// PushConfig represents container push configuration
type PushConfig struct {
	Enabled      bool `yaml:"enabled" json:"enabled"`
	OnProduction bool `yaml:"onProduction" json:"onProduction"`
}

// BuildConfig represents container build configuration
type BuildConfig struct {
	AlwaysBuild  bool `yaml:"alwaysBuild" json:"alwaysBuild"`
	AlwaysPush   bool `yaml:"alwaysPush" json:"alwaysPush"`
	OnPR         bool `yaml:"onPR" json:"onPR"`
	OnProduction bool `yaml:"onProduction" json:"onProduction"`
}

// WorkflowInputs represents all possible workflow inputs with strong typing
type WorkflowInputs struct {
	// Language/Runtime inputs
	NodeVersion   string `json:"nodeVersion,omitempty"`
	GoVersion     string `json:"goVersion,omitempty"`
	PythonVersion string `json:"pythonVersion,omitempty"`

	// Package management
	PackageManager string `json:"packageManager,omitempty"`
	Requirements   string `json:"requirements,omitempty"`

	// Commands
	TestCommand  string `json:"testCommand,omitempty"`
	BuildCommand string `json:"buildCommand,omitempty"`
	LintCommand  string `json:"lintCommand,omitempty"`

	// Configurations
	Security  SecurityConfig  `json:"security,omitempty"`
	Container ContainerConfig `json:"container,omitempty"`

	// Build platforms (Go specific)
	Platforms string `json:"platforms,omitempty"`

	// Legacy compatibility fields (deprecated)
	TrivyScanEnabled   *bool  `json:"trivyScanEnabled,omitempty"`
	TrivySeverity      string `json:"trivySeverity,omitempty"`
	ContainerEnabled   *bool  `json:"containerEnabled,omitempty"`
	ContainerRegistry  string `json:"containerRegistry,omitempty"`
	ContainerImageName string `json:"containerImageName,omitempty"`
	ContainerImageTag  string `json:"containerImageTag,omitempty"`
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Trivy: TrivyConfig{
			Enabled:  true,
			Severity: "CRITICAL,HIGH",
			ExitCode: "1",
		},
	}
}

// DefaultContainerConfig returns the default container configuration
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		Enabled:      false,
		Registry:     "ghcr.io",
		ImageName:    "${{ github.repository }}",
		ImageTag:     "${{ github.sha }}",
		Dockerfile:   "Dockerfile",
		BuildContext: ".",
		BuildArgs:    "{}",
		Push: PushConfig{
			Enabled:      true,
			OnProduction: true,
		},
		Build: BuildConfig{
			AlwaysBuild:  false,
			AlwaysPush:   false,
			OnPR:         true,
			OnProduction: true,
		},
	}
}
