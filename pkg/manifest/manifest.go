package manifest

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ValidationMode represents the validation mode for the manifest
type ValidationMode string

const (
	ValidationModeStrict  ValidationMode = "strict"
	ValidationModeRelaxed ValidationMode = "relaxed"
)

// Manifest represents the root structure of a GPGen pipeline manifest
type Manifest struct {
	APIVersion string            `yaml:"apiVersion" json:"apiVersion"`
	Kind       string            `yaml:"kind" json:"kind"`
	Metadata   *ManifestMetadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Spec       ManifestSpec      `yaml:"spec" json:"spec"`
}

// ManifestMetadata contains metadata about the pipeline
type ManifestMetadata struct {
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// ManifestSpec contains the pipeline specification
type ManifestSpec struct {
	Template     string                       `yaml:"template" json:"template"`
	Inputs       map[string]interface{}       `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	CustomSteps  []CustomStep                 `yaml:"customSteps,omitempty" json:"customSteps,omitempty"`
	Overrides    map[string]StepOverride      `yaml:"overrides,omitempty" json:"overrides,omitempty"`
	Environments map[string]EnvironmentConfig `yaml:"environments,omitempty" json:"environments,omitempty"`
}

// CustomStep represents a custom step in the pipeline
type CustomStep struct {
	Name            string            `yaml:"name" json:"name"`
	Position        string            `yaml:"position" json:"position"`
	Uses            string            `yaml:"uses,omitempty" json:"uses,omitempty"`
	Run             string            `yaml:"run,omitempty" json:"run,omitempty"`
	With            map[string]string `yaml:"with,omitempty" json:"with,omitempty"`
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	If              string            `yaml:"if,omitempty" json:"if,omitempty"`
	TimeoutMinutes  *int              `yaml:"timeout-minutes,omitempty" json:"timeout-minutes,omitempty"`
	ContinueOnError *bool             `yaml:"continue-on-error,omitempty" json:"continue-on-error,omitempty"`
}

// StepOverride represents overrides for existing template steps
type StepOverride struct {
	Name            string            `yaml:"name,omitempty" json:"name,omitempty"`
	Uses            string            `yaml:"uses,omitempty" json:"uses,omitempty"`
	Run             string            `yaml:"run,omitempty" json:"run,omitempty"`
	With            map[string]string `yaml:"with,omitempty" json:"with,omitempty"`
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	TimeoutMinutes  *int              `yaml:"timeout-minutes,omitempty" json:"timeout-minutes,omitempty"`
	ContinueOnError *bool             `yaml:"continue-on-error,omitempty" json:"continue-on-error,omitempty"`
	If              string            `yaml:"if,omitempty" json:"if,omitempty"`
}

// EnvironmentConfig represents environment-specific configuration
type EnvironmentConfig struct {
	Inputs      map[string]interface{}  `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	CustomSteps []CustomStep            `yaml:"customSteps,omitempty" json:"customSteps,omitempty"`
	Overrides   map[string]StepOverride `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

var (
	validAPIVersions = []string{"gpgen.dev/v1"}
	validKinds       = []string{"Pipeline"}
	validTemplates   = []string{"node-app", "go-service"}
	positionRegex    = regexp.MustCompile(`^(before|after|replace):[a-z0-9-]+$`)
)

// ParseManifest parses a YAML manifest into a Manifest struct
func ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest

	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if manifest.APIVersion == "" {
		return nil, fmt.Errorf("apiVersion is required")
	}
	if manifest.Kind == "" {
		return nil, fmt.Errorf("kind is required")
	}

	// Check if spec section is present by looking for template field
	// We need to check if spec was provided in the YAML
	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err == nil {
		if _, hasSpec := rawData["spec"]; !hasSpec {
			return nil, fmt.Errorf("spec is required")
		}
	}

	if manifest.Spec.Template == "" {
		return nil, fmt.Errorf("template is required")
	}

	return &manifest, nil
}

// ValidateManifest validates a parsed manifest according to the schema rules
func ValidateManifest(manifest *Manifest) error {
	// Validate API version
	if !contains(validAPIVersions, manifest.APIVersion) {
		return fmt.Errorf("invalid apiVersion: %s, must be one of %v",
			manifest.APIVersion, validAPIVersions)
	}

	// Validate kind
	if !contains(validKinds, manifest.Kind) {
		return fmt.Errorf("invalid kind: %s, must be one of %v",
			manifest.Kind, validKinds)
	}

	// Validate template
	if !contains(validTemplates, manifest.Spec.Template) {
		return fmt.Errorf("invalid template: %s, must be one of %v",
			manifest.Spec.Template, validTemplates)
	}

	// Validate custom steps
	for i, step := range manifest.Spec.CustomSteps {
		if err := validateCustomStep(&step); err != nil {
			return fmt.Errorf("invalid custom step at index %d: %w", i, err)
		}
	}

	// Validate environment custom steps
	for envName, envConfig := range manifest.Spec.Environments {
		for i, step := range envConfig.CustomSteps {
			if err := validateCustomStep(&step); err != nil {
				return fmt.Errorf("invalid custom step at index %d in environment %s: %w", i, envName, err)
			}
		}
	}

	return nil
}

// validateCustomStep validates a custom step
func validateCustomStep(step *CustomStep) error {
	// Validate step name is not empty
	if step.Name == "" {
		return fmt.Errorf("step name cannot be empty")
	}

	// Validate position format
	if err := validatePosition(step.Position); err != nil {
		return err
	}

	// Validate that step has either uses or run, but not both
	hasUses := step.Uses != ""
	hasRun := step.Run != ""

	if !hasUses && !hasRun {
		return fmt.Errorf("step must have either 'uses' or 'run'")
	}

	if hasUses && hasRun {
		return fmt.Errorf("step cannot have both 'uses' and 'run'")
	}

	// Validate timeout if specified
	if step.TimeoutMinutes != nil && (*step.TimeoutMinutes < 1 || *step.TimeoutMinutes > 360) {
		return fmt.Errorf("timeout-minutes must be between 1 and 360")
	}

	return nil
}

// validatePosition validates the position string format
func validatePosition(position string) error {
	if !positionRegex.MatchString(position) {
		return fmt.Errorf("invalid position format: %s, must match pattern '^(before|after|replace):[a-z0-9-]+$'", position)
	}
	return nil
}

// GetValidationMode returns the validation mode from the manifest metadata
func GetValidationMode(manifest *Manifest) ValidationMode {
	if manifest.Metadata == nil || manifest.Metadata.Annotations == nil {
		return ValidationModeStrict
	}

	mode, exists := manifest.Metadata.Annotations["gpgen.dev/validation-mode"]
	if !exists {
		return ValidationModeStrict
	}

	switch mode {
	case string(ValidationModeRelaxed):
		return ValidationModeRelaxed
	default:
		return ValidationModeStrict
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LoadManifestFromFile loads and parses a manifest from a file
func LoadManifestFromFile(filename string) (*Manifest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	manifest, err := ParseManifest(data)
	if err != nil {
		return nil, err
	}

	if err := ValidateManifest(manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return manifest, nil
}
