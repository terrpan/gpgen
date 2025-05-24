package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/terrpan/gpgen/pkg/manifest"
	"github.com/terrpan/gpgen/pkg/templates"
	"gopkg.in/yaml.v3"
)

// WorkflowGenerator generates GitHub Actions workflows from manifests and templates
type WorkflowGenerator struct {
	templateManager *templates.TemplateManager
}

// NewWorkflowGenerator creates a new workflow generator
func NewWorkflowGenerator(templatesDir string) *WorkflowGenerator {
	return &WorkflowGenerator{
		templateManager: templates.NewTemplateManager(templatesDir),
	}
}

// GitHubActionsWorkflow represents a GitHub Actions workflow
type GitHubActionsWorkflow struct {
	Name string                 `yaml:"name"`
	On   map[string]interface{} `yaml:"on"`
	Jobs map[string]Job         `yaml:"jobs"`
}

// Job represents a GitHub Actions job
type Job struct {
	RunsOn      string            `yaml:"runs-on"`
	Permissions map[string]string `yaml:"permissions,omitempty"`
	Steps       []WorkflowStep    `yaml:"steps"`
}

// WorkflowStep represents a GitHub Actions workflow step
type WorkflowStep struct {
	Name        string            `yaml:"name,omitempty"`
	Uses        string            `yaml:"uses,omitempty"`
	Run         string            `yaml:"run,omitempty"`
	With        map[string]string `yaml:"with,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	If          string            `yaml:"if,omitempty"`
	TimeoutMins int               `yaml:"timeout-minutes,omitempty"`
}

// GenerateWorkflow generates a GitHub Actions workflow from a manifest
func (g *WorkflowGenerator) GenerateWorkflow(m *manifest.Manifest, environment string) (string, error) {
	// Load the template
	tmpl, err := g.templateManager.LoadTemplate(m.Spec.Template)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Get effective inputs for the environment
	inputs := g.getEffectiveInputs(m, environment)

	// Validate inputs against template
	if err := g.templateManager.ValidateInputs(m.Spec.Template, inputs); err != nil {
		return "", fmt.Errorf("input validation failed: %w", err)
	}

	// Generate workflow steps
	steps, err := g.generateSteps(tmpl, m, environment, inputs)
	if err != nil {
		return "", fmt.Errorf("failed to generate steps: %w", err)
	}

	// Create workflow
	workflow := &GitHubActionsWorkflow{
		Name: g.getWorkflowName(m, environment),
		On:   g.getWorkflowTriggers(m, environment),
		Jobs: map[string]Job{
			"build": {
				RunsOn:      "ubuntu-latest",
				Permissions: g.getRequiredPermissions(tmpl, inputs),
				Steps:       steps,
			},
		},
	}

	// Convert to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(workflow); err != nil {
		return "", fmt.Errorf("failed to encode workflow to YAML: %w", err)
	}

	return buf.String(), nil
}

// getEffectiveInputs merges template defaults, base inputs, environment-specific overrides and event context
func (g *WorkflowGenerator) getEffectiveInputs(m *manifest.Manifest, environment string) map[string]interface{} {
	inputs := make(map[string]interface{})

	// Load template to get defaults
	tmpl, err := g.templateManager.LoadTemplate(m.Spec.Template)
	if err != nil {
		// If we can't load the template, proceed without defaults
		// This shouldn't happen as template loading is validated earlier
	} else {
		// Start with template defaults
		for k, inputDef := range tmpl.Inputs {
			if inputDef.Default != nil {
				inputs[k] = inputDef.Default
			}
		}
	}

	// Apply base inputs (overrides template defaults)
	for k, v := range m.Spec.Inputs {
		inputs[k] = v
	}

	// Apply environment-specific overrides
	if environment != "default" {
		if envConfig, exists := m.Spec.Environments[environment]; exists {
			for k, v := range envConfig.Inputs {
				inputs[k] = v
			}
		}
	}

	// Add event-driven context based on environment triggers
	g.addEventDrivenContext(inputs, environment)

	return inputs
}

// addEventDrivenContext adds context-aware settings based on environment and triggers
func (g *WorkflowGenerator) addEventDrivenContext(inputs map[string]interface{}, environment string) {
	// Set default event-driven behavior based on environment
	switch environment {
	case "default", "staging":
		// Default/staging: Build on PRs for validation, but strategic pushing
		if _, exists := inputs["containerBuildOnPR"]; !exists {
			inputs["containerBuildOnPR"] = true
		}
		if _, exists := inputs["containerPushOnProduction"]; !exists {
			inputs["containerPushOnProduction"] = false // Don't push on production events in staging
		}
	case "production":
		// Production: Build and push on production events
		if _, exists := inputs["containerBuildOnPR"]; !exists {
			inputs["containerBuildOnPR"] = false // Don't build on PRs in production env
		}
		if _, exists := inputs["containerBuildOnProduction"]; !exists {
			inputs["containerBuildOnProduction"] = true
		}
		if _, exists := inputs["containerPushOnProduction"]; !exists {
			inputs["containerPushOnProduction"] = true
		}
	}
}

// generateSteps generates workflow steps by merging template steps with custom steps
func (g *WorkflowGenerator) generateSteps(tmpl *templates.Template, m *manifest.Manifest, environment string, inputs map[string]interface{}) ([]WorkflowStep, error) {
	var steps []WorkflowStep

	// Process template steps
	for _, templateStep := range tmpl.Steps {
		step, err := g.processTemplateStep(templateStep, inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to process template step %s: %w", templateStep.ID, err)
		}
		steps = append(steps, step)
	}

	// Apply custom steps
	steps, err := g.applyCustomSteps(steps, m.Spec.CustomSteps, environment, m)
	if err != nil {
		return nil, fmt.Errorf("failed to apply custom steps: %w", err)
	}

	return steps, nil
}

// processTemplateStep processes a template step with input substitution
func (g *WorkflowGenerator) processTemplateStep(templateStep templates.Step, inputs map[string]interface{}) (WorkflowStep, error) {
	step := WorkflowStep{
		Name:        templateStep.Name,
		Uses:        templateStep.Uses,
		TimeoutMins: templateStep.TimeoutMins,
	}

	// Process run command with template substitution
	if templateStep.Run != "" {
		run, err := g.substituteTemplate(templateStep.Run, inputs)
		if err != nil {
			return step, fmt.Errorf("failed to substitute run command: %w", err)
		}
		// Replace GitHub Actions placeholders
		run = g.replaceGitHubActionsPlaceholders(run)
		step.Run = run
	}

	// Process with parameters
	if len(templateStep.With) > 0 {
		step.With = make(map[string]string)
		for k, v := range templateStep.With {
			value, err := g.substituteTemplate(v, inputs)
			if err != nil {
				return step, fmt.Errorf("failed to substitute with parameter %s: %w", k, err)
			}
			// Replace GitHub Actions placeholders
			value = g.replaceGitHubActionsPlaceholders(value)
			step.With[k] = value
		}
	}

	// Process environment variables
	if len(templateStep.Env) > 0 {
		step.Env = make(map[string]string)
		for k, v := range templateStep.Env {
			value, err := g.substituteTemplate(v, inputs)
			if err != nil {
				return step, fmt.Errorf("failed to substitute env variable %s: %w", k, err)
			}
			// Replace GitHub Actions placeholders
			value = g.replaceGitHubActionsPlaceholders(value)
			step.Env[k] = value
		}
	}

	// Process if condition
	if templateStep.If != "" {
		ifCondition, err := g.substituteTemplate(templateStep.If, inputs)
		if err != nil {
			return step, fmt.Errorf("failed to substitute if condition: %w", err)
		}
		// Replace GitHub Actions placeholders
		ifCondition = g.replaceGitHubActionsPlaceholders(ifCondition)
		step.If = ifCondition
	}

	return step, nil
}

// substituteTemplate performs template substitution on a string
func (g *WorkflowGenerator) substituteTemplate(templateStr string, inputs map[string]interface{}) (string, error) {
	tmpl, err := template.New("step").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := map[string]interface{}{
		"Inputs": inputs,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// applyCustomSteps applies custom steps according to their position directives
func (g *WorkflowGenerator) applyCustomSteps(steps []WorkflowStep, customSteps []manifest.CustomStep, environment string, m *manifest.Manifest) ([]WorkflowStep, error) {
	// Get environment-specific custom steps
	allCustomSteps := customSteps
	if environment != "default" {
		if envConfig, exists := m.Spec.Environments[environment]; exists {
			allCustomSteps = append(allCustomSteps, envConfig.CustomSteps...)
		}
	}

	for _, customStep := range allCustomSteps {
		var err error
		steps, err = g.applyCustomStep(steps, customStep)
		if err != nil {
			return nil, fmt.Errorf("failed to apply custom step %s: %w", customStep.Name, err)
		}
	}

	return steps, nil
}

// applyCustomStep applies a single custom step at the specified position
func (g *WorkflowGenerator) applyCustomStep(steps []WorkflowStep, customStep manifest.CustomStep) ([]WorkflowStep, error) {
	newStep := WorkflowStep{
		Name: customStep.Name,
		Uses: customStep.Uses,
		Run:  customStep.Run,
	}

	if customStep.TimeoutMinutes != nil {
		newStep.TimeoutMins = *customStep.TimeoutMinutes
	}

	if len(customStep.With) > 0 {
		newStep.With = customStep.With
	}
	if len(customStep.Env) > 0 {
		newStep.Env = customStep.Env
	}
	if customStep.If != "" {
		newStep.If = customStep.If
	}

	// Parse position directive
	position := customStep.Position
	if position == "" {
		// Default: append to end
		return append(steps, newStep), nil
	}

	parts := strings.SplitN(position, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid position format: %s (expected 'before:step', 'after:step', or 'replace:step')", position)
	}

	directive := parts[0]
	targetStep := parts[1]

	switch directive {
	case "before":
		return g.insertStepBefore(steps, newStep, targetStep)
	case "after":
		return g.insertStepAfter(steps, newStep, targetStep)
	case "replace":
		return g.replaceStep(steps, newStep, targetStep)
	default:
		return nil, fmt.Errorf("unknown position directive: %s", directive)
	}
}

// insertStepBefore inserts a step before the target step
func (g *WorkflowGenerator) insertStepBefore(steps []WorkflowStep, newStep WorkflowStep, targetStep string) ([]WorkflowStep, error) {
	for i, step := range steps {
		if g.matchesStep(step, targetStep) {
			result := make([]WorkflowStep, 0, len(steps)+1)
			result = append(result, steps[:i]...)
			result = append(result, newStep)
			result = append(result, steps[i:]...)
			return result, nil
		}
	}
	return nil, fmt.Errorf("target step not found: %s", targetStep)
}

// insertStepAfter inserts a step after the target step
func (g *WorkflowGenerator) insertStepAfter(steps []WorkflowStep, newStep WorkflowStep, targetStep string) ([]WorkflowStep, error) {
	for i, step := range steps {
		if g.matchesStep(step, targetStep) {
			result := make([]WorkflowStep, 0, len(steps)+1)
			result = append(result, steps[:i+1]...)
			result = append(result, newStep)
			result = append(result, steps[i+1:]...)
			return result, nil
		}
	}
	return nil, fmt.Errorf("target step not found: %s", targetStep)
}

// replaceStep replaces the target step with the new step
func (g *WorkflowGenerator) replaceStep(steps []WorkflowStep, newStep WorkflowStep, targetStep string) ([]WorkflowStep, error) {
	for i, step := range steps {
		if g.matchesStep(step, targetStep) {
			steps[i] = newStep
			return steps, nil
		}
	}
	return nil, fmt.Errorf("target step not found: %s", targetStep)
}

// matchesStep checks if a workflow step matches the target identifier
func (g *WorkflowGenerator) matchesStep(step WorkflowStep, target string) bool {
	// Normalize both strings to lowercase for comparison
	stepName := strings.ToLower(step.Name)
	targetName := strings.ToLower(target)

	// Special case mappings for common patterns
	stepMappings := map[string][]string{
		"run tests":            {"test", "tests"},
		"build application":    {"build"},
		"setup node.js":        {"setup-node", "node"},
		"install dependencies": {"install", "dependencies"},
		"checkout code":        {"checkout"},
	}

	// Check if we have a specific mapping for this step
	if targets, exists := stepMappings[stepName]; exists {
		for _, t := range targets {
			if t == targetName {
				return true
			}
		}
		return false
	}

	// Fallback: check if target is contained in any word of the step name
	stepWords := strings.Fields(stepName)
	for _, word := range stepWords {
		// Remove common prefixes/suffixes and check for match
		word = strings.TrimSuffix(word, "s") // Handle plurals
		if word == targetName || targetName == word {
			return true
		}
		// Also check if target is contained in word (for longer words)
		if len(targetName) >= 4 && strings.Contains(word, targetName) {
			return true
		}
	}

	return false
}

// getWorkflowName generates the workflow name
func (g *WorkflowGenerator) getWorkflowName(m *manifest.Manifest, environment string) string {
	name := m.Metadata.Name
	if environment != "default" {
		name = fmt.Sprintf("%s (%s)", name, environment)
	}
	return name
}

// getWorkflowTriggers generates workflow triggers based on environment
func (g *WorkflowGenerator) getWorkflowTriggers(m *manifest.Manifest, environment string) map[string]interface{} {
	triggers := make(map[string]interface{})

	switch environment {
	case "default", "staging":
		triggers["push"] = map[string]interface{}{
			"branches": []string{"main", "develop"},
		}
		triggers["pull_request"] = map[string]interface{}{
			"branches": []string{"main"},
		}
	case "production":
		triggers["push"] = map[string]interface{}{
			"tags": []string{"v*"},
		}
		triggers["release"] = map[string]interface{}{
			"types": []string{"published"},
		}
	default:
		// Custom environment - use push to main
		triggers["push"] = map[string]interface{}{
			"branches": []string{"main"},
		}
	}

	return triggers
}

// getRequiredPermissions determines the required permissions for the workflow
func (g *WorkflowGenerator) getRequiredPermissions(tmpl *templates.Template, inputs map[string]interface{}) map[string]string {
	permissions := make(map[string]string)

	// Check if Trivy scanning is enabled
	if trivyScanEnabled, exists := inputs["trivyScanEnabled"]; exists {
		if enabled, ok := trivyScanEnabled.(bool); ok && enabled {
			// Add permissions required for uploading SARIF results to GitHub Security tab
			permissions["security-events"] = "write"
			permissions["contents"] = "read"
		}
	}

	// Check if container building/pushing is enabled
	if containerEnabled, exists := inputs["containerEnabled"]; exists {
		if enabled, ok := containerEnabled.(bool); ok && enabled {
			// Add permissions required for container registry operations
			permissions["packages"] = "write"
			if permissions["contents"] == "" {
				permissions["contents"] = "read"
			}
		}
	}

	return permissions
}

// replaceGitHubActionsPlaceholders replaces template placeholders with GitHub Actions syntax
func (g *WorkflowGenerator) replaceGitHubActionsPlaceholders(value string) string {
	// Replace placeholders with GitHub Actions syntax
	value = strings.ReplaceAll(value, "GITHUB_ACTOR_PLACEHOLDER", "${{ github.actor }}")
	value = strings.ReplaceAll(value, "GITHUB_TOKEN_PLACEHOLDER", "${{ secrets.GITHUB_TOKEN }}")
	return value
}
