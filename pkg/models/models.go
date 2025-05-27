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

// Input defines a parameter for a template
type Input struct {
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
	Required    bool        `yaml:"required"`
	Options     []string    `yaml:"options,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty"`
}

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
