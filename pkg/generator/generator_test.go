package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terrpan/gpgen/pkg/manifest"
)

func TestWorkflowGenerator_GenerateWorkflow(t *testing.T) {
	generator := NewWorkflowGenerator("")

	t.Run("generate basic node-app workflow", func(t *testing.T) {
		m := &manifest.Manifest{
			APIVersion: "gpgen.dev/v1",
			Kind:       "Pipeline",
			Metadata: &manifest.ManifestMetadata{
				Name: "test-app",
			},
			Spec: manifest.ManifestSpec{
				Template: "node-app",
				Inputs: map[string]interface{}{
					"nodeVersion":    "18",
					"packageManager": "npm",
					"testCommand":    "npm test",
				},
			},
		}

		workflow, err := generator.GenerateWorkflow(m, "default")
		require.NoError(t, err)
		assert.NotEmpty(t, workflow)

		// Check basic YAML structure
		assert.Contains(t, workflow, "name: test-app")
		assert.Contains(t, workflow, "runs-on: ubuntu-latest")
		assert.Contains(t, workflow, "actions/checkout@v4")
		assert.Contains(t, workflow, "actions/setup-node@v4")
	})

	t.Run("generate workflow with environment overrides", func(t *testing.T) {
		m := &manifest.Manifest{
			APIVersion: "gpgen.dev/v1",
			Kind:       "Pipeline",
			Metadata: &manifest.ManifestMetadata{
				Name: "test-app",
			},
			Spec: manifest.ManifestSpec{
				Template: "node-app",
				Inputs: map[string]interface{}{
					"nodeVersion":    "18",
					"packageManager": "npm",
					"testCommand":    "npm test",
				},
				Environments: map[string]manifest.EnvironmentConfig{
					"production": {
						Inputs: map[string]interface{}{
							"nodeVersion": "20",
							"testCommand": "npm run test:all",
						},
					},
				},
			},
		}

		workflow, err := generator.GenerateWorkflow(m, "production")
		require.NoError(t, err)
		assert.NotEmpty(t, workflow)

		// Should use production inputs
		assert.Contains(t, workflow, "node-version: \"20\"")
		assert.Contains(t, workflow, "npm run test:all")
		assert.Contains(t, workflow, "name: test-app (production)")
	})

	t.Run("generate workflow with custom steps", func(t *testing.T) {
		m := &manifest.Manifest{
			APIVersion: "gpgen.dev/v1",
			Kind:       "Pipeline",
			Metadata: &manifest.ManifestMetadata{
				Name: "test-app",
			},
			Spec: manifest.ManifestSpec{
				Template: "node-app",
				Inputs: map[string]interface{}{
					"nodeVersion":    "18",
					"packageManager": "npm",
					"testCommand":    "npm test",
				},
				CustomSteps: []manifest.CustomStep{
					{
						Name:     "security-scan",
						Position: "after:test",
						Uses:     "security/scan-action@v1",
						With: map[string]string{
							"token": "${{ secrets.SECURITY_TOKEN }}",
						},
					},
				},
			},
		}

		workflow, err := generator.GenerateWorkflow(m, "default")
		require.NoError(t, err)
		assert.NotEmpty(t, workflow)

		// Should contain the custom step
		assert.Contains(t, workflow, "name: security-scan")
		assert.Contains(t, workflow, "security/scan-action@v1")
		assert.Contains(t, workflow, "token: ${{ secrets.SECURITY_TOKEN }}")
	})
}

func TestWorkflowGenerator_GetEffectiveInputs(t *testing.T) {
	generator := NewWorkflowGenerator("")

	m := &manifest.Manifest{
		Spec: manifest.ManifestSpec{
			Inputs: map[string]interface{}{
				"nodeVersion":    "18",
				"packageManager": "npm",
				"testCommand":    "npm test",
			},
			Environments: map[string]manifest.EnvironmentConfig{
				"production": {
					Inputs: map[string]interface{}{
						"nodeVersion": "20",
						"testCommand": "npm run test:all",
					},
				},
			},
		},
	}

	t.Run("default environment", func(t *testing.T) {
		inputs := generator.getEffectiveInputs(m, "default")

		assert.Equal(t, "18", inputs["nodeVersion"])
		assert.Equal(t, "npm", inputs["packageManager"])
		assert.Equal(t, "npm test", inputs["testCommand"])
	})

	t.Run("production environment", func(t *testing.T) {
		inputs := generator.getEffectiveInputs(m, "production")

		// Overridden values
		assert.Equal(t, "20", inputs["nodeVersion"])
		assert.Equal(t, "npm run test:all", inputs["testCommand"])

		// Inherited value
		assert.Equal(t, "npm", inputs["packageManager"])
	})
}

func TestWorkflowGenerator_ApplyCustomStep(t *testing.T) {
	generator := NewWorkflowGenerator("")

	originalSteps := []WorkflowStep{
		{Name: "Checkout code"},
		{Name: "Setup Node.js"},
		{Name: "Install dependencies"},
		{Name: "Run tests"},
		{Name: "Build application"},
	}

	t.Run("insert after test", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name:     "Security Scan",
			Position: "after:test",
			Uses:     "security/scan@v1",
		}

		result, err := generator.applyCustomStep(originalSteps, customStep)
		require.NoError(t, err)

		// Should have one more step
		assert.Len(t, result, 6)

		// Find the security scan step
		var found bool
		for i, step := range result {
			if step.Name == "Security Scan" {
				found = true
				// Should be after "Run tests" step
				assert.Greater(t, i, 3)
				assert.Equal(t, "security/scan@v1", step.Uses)
				break
			}
		}
		assert.True(t, found, "Security scan step should be inserted")
	})

	t.Run("insert before build", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name:     "Lint Code",
			Position: "before:build",
			Run:      "npm run lint",
		}

		result, err := generator.applyCustomStep(originalSteps, customStep)
		require.NoError(t, err)

		// Should have one more step
		assert.Len(t, result, 6)

		// Find the lint step
		var found bool
		for i, step := range result {
			if step.Name == "Lint Code" {
				found = true
				// Should be before "Build application" step (which is now at index i+1)
				assert.Equal(t, "Build application", result[i+1].Name)
				assert.Equal(t, "npm run lint", step.Run)
				break
			}
		}
		assert.True(t, found, "Lint step should be inserted")
	})

	t.Run("replace step", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name:     "Custom Build",
			Position: "replace:build",
			Run:      "custom build command",
		}

		result, err := generator.applyCustomStep(originalSteps, customStep)
		require.NoError(t, err)

		// Should have same number of steps
		assert.Len(t, result, 5)

		// Should not have "Build application" anymore
		for _, step := range result {
			assert.NotEqual(t, "Build application", step.Name)
		}

		// Should have "Custom Build"
		var found bool
		for _, step := range result {
			if step.Name == "Custom Build" {
				found = true
				assert.Equal(t, "custom build command", step.Run)
				break
			}
		}
		assert.True(t, found, "Custom build step should replace original")
	})

	t.Run("append when no position", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name: "Deploy",
			Run:  "deploy command",
		}

		result, err := generator.applyCustomStep(originalSteps, customStep)
		require.NoError(t, err)

		// Should have one more step at the end
		assert.Len(t, result, 6)
		assert.Equal(t, "Deploy", result[5].Name)
	})

	t.Run("invalid position format", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name:     "Invalid Step",
			Position: "invalid-position",
			Run:      "some command",
		}

		_, err := generator.applyCustomStep(originalSteps, customStep)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position format")
	})

	t.Run("target step not found", func(t *testing.T) {
		customStep := manifest.CustomStep{
			Name:     "Test Step",
			Position: "after:nonexistent",
			Run:      "some command",
		}

		_, err := generator.applyCustomStep(originalSteps, customStep)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target step not found")
	})
}

func TestWorkflowGenerator_MatchesStep(t *testing.T) {
	generator := NewWorkflowGenerator("")

	tests := []struct {
		stepName string
		target   string
		expected bool
	}{
		{"Run tests", "test", true},
		{"Run tests", "tests", true},
		{"Build application", "build", true},
		{"Setup Node.js", "setup-node", true},
		{"Install dependencies", "install", true},
		{"Checkout code", "checkout", true},
		{"Run tests", "build", false},
		{"Setup Node.js", "setup-go", false},
	}

	for _, tt := range tests {
		t.Run(tt.stepName+"->"+tt.target, func(t *testing.T) {
			step := WorkflowStep{Name: tt.stepName}
			result := generator.matchesStep(step, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowGenerator_GetWorkflowTriggers(t *testing.T) {
	generator := NewWorkflowGenerator("")
	m := &manifest.Manifest{}

	t.Run("default environment triggers", func(t *testing.T) {
		triggers := generator.getWorkflowTriggers(m, "default")

		assert.Contains(t, triggers, "push")
		assert.Contains(t, triggers, "pull_request")

		pushTrigger := triggers["push"].(map[string]interface{})
		assert.Contains(t, pushTrigger, "branches")
	})

	t.Run("production environment triggers", func(t *testing.T) {
		triggers := generator.getWorkflowTriggers(m, "production")

		assert.Contains(t, triggers, "push")
		assert.Contains(t, triggers, "release")

		pushTrigger := triggers["push"].(map[string]interface{})
		assert.Contains(t, pushTrigger, "tags")
	})

	t.Run("staging environment triggers", func(t *testing.T) {
		triggers := generator.getWorkflowTriggers(m, "staging")

		assert.Contains(t, triggers, "push")
		assert.Contains(t, triggers, "pull_request")
	})
}

func TestWorkflowGenerator_SubstituteTemplate(t *testing.T) {
	generator := NewWorkflowGenerator("")

	inputs := map[string]interface{}{
		"nodeVersion":    "18",
		"packageManager": "npm",
		"testCommand":    "npm test",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple substitution",
			template: "{{ .Inputs.nodeVersion }}",
			expected: "18",
		},
		{
			name:     "conditional substitution",
			template: "{{ if eq .Inputs.packageManager \"npm\" }}npm ci{{ else }}yarn install{{ end }}",
			expected: "npm ci",
		},
		{
			name:     "multiple substitutions",
			template: "node-version: {{ .Inputs.nodeVersion }} manager: {{ .Inputs.packageManager }}",
			expected: "node-version: 18 manager: npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.substituteTemplate(tt.template, inputs)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("invalid template", func(t *testing.T) {
		_, err := generator.substituteTemplate("{{ .Invalid", inputs)
		assert.Error(t, err)
	})
}

func TestWorkflowGenerator_Integration(t *testing.T) {
	generator := NewWorkflowGenerator("")

	// Create a comprehensive manifest
	m := &manifest.Manifest{
		APIVersion: "gpgen.dev/v1",
		Kind:       "Pipeline",
		Metadata: &manifest.ManifestMetadata{
			Name: "ecommerce-api",
			Annotations: map[string]string{
				"gpgen.dev/validation-mode": "strict",
			},
		},
		Spec: manifest.ManifestSpec{
			Template: "node-app",
			Inputs: map[string]interface{}{
				"nodeVersion":    "18",
				"packageManager": "npm",
				"testCommand":    "npm run test:ci",
				"buildCommand":   "npm run build",
			},
			CustomSteps: []manifest.CustomStep{
				{
					Name:     "security-scan",
					Position: "after:test",
					Uses:     "securecodewarrior/github-action-add-sarif@v1",
					With: map[string]string{
						"sarif-file": "security-results.sarif",
					},
				},
				{
					Name:     "dependency-check",
					Position: "before:build",
					Run:      "npm audit --audit-level high",
				},
			},
			Environments: map[string]manifest.EnvironmentConfig{
				"production": {
					Inputs: map[string]interface{}{
						"nodeVersion": "20",
						"testCommand": "npm run test:all",
					},
					CustomSteps: []manifest.CustomStep{
						{
							Name:     "performance-test",
							Position: "after:test",
							Run:      "npm run test:performance",
						},
					},
				},
			},
		},
	}

	t.Run("generate default workflow", func(t *testing.T) {
		workflow, err := generator.GenerateWorkflow(m, "default")
		require.NoError(t, err)

		// Check basic structure
		assert.Contains(t, workflow, "name: ecommerce-api")
		assert.Contains(t, workflow, "node-version: \"18\"")
		assert.Contains(t, workflow, "npm run test:ci")

		// Check custom steps
		assert.Contains(t, workflow, "security-scan")
		assert.Contains(t, workflow, "dependency-check")
		assert.Contains(t, workflow, "securecodewarrior/github-action-add-sarif@v1")
		assert.Contains(t, workflow, "npm audit --audit-level high")

		// Should not contain production-specific steps
		assert.NotContains(t, workflow, "performance-test")
	})

	t.Run("generate production workflow", func(t *testing.T) {
		workflow, err := generator.GenerateWorkflow(m, "production")
		require.NoError(t, err)

		// Check environment-specific changes
		assert.Contains(t, workflow, "name: ecommerce-api (production)")
		assert.Contains(t, workflow, "node-version: \"20\"")
		assert.Contains(t, workflow, "npm run test:all")

		// Check both base and environment-specific custom steps
		assert.Contains(t, workflow, "security-scan")
		assert.Contains(t, workflow, "dependency-check")
		assert.Contains(t, workflow, "performance-test")

		// Check production triggers (tags and releases)
		assert.Contains(t, workflow, "tags:")
		assert.Contains(t, workflow, "release:")
	})
}

func TestWorkflowGenerator_GetRequiredPermissions(t *testing.T) {
	generator := NewWorkflowGenerator("")

	tests := []struct {
		name        string
		inputs      map[string]interface{}
		expected    map[string]string
		description string
	}{
		{
			name: "trivy scanning enabled",
			inputs: map[string]interface{}{
				"trivyScanEnabled": true,
				"goVersion":        "1.22",
			},
			expected: map[string]string{
				"security-events": "write",
				"contents":        "read",
			},
			description: "Should add security permissions when Trivy scanning is enabled",
		},
		{
			name: "trivy scanning disabled",
			inputs: map[string]interface{}{
				"trivyScanEnabled": false,
				"goVersion":        "1.22",
			},
			expected:    map[string]string{},
			description: "Should not add permissions when Trivy scanning is disabled",
		},
		{
			name: "trivy scanning not specified",
			inputs: map[string]interface{}{
				"goVersion": "1.22",
			},
			expected:    map[string]string{},
			description: "Should not add permissions when Trivy scanning is not specified",
		},
		{
			name: "trivy scanning enabled as string (should not trigger)",
			inputs: map[string]interface{}{
				"trivyScanEnabled": "true",
				"goVersion":        "1.22",
			},
			expected:    map[string]string{},
			description: "Should not add permissions when trivyScanEnabled is not a boolean",
		},
		{
			name: "container building enabled",
			inputs: map[string]interface{}{
				"containerEnabled": true,
				"goVersion":        "1.22",
			},
			expected: map[string]string{
				"packages": "write",
				"contents": "read",
			},
			description: "Should add package permissions when container building is enabled",
		},
		{
			name: "container building disabled",
			inputs: map[string]interface{}{
				"containerEnabled": false,
				"goVersion":        "1.22",
			},
			expected:    map[string]string{},
			description: "Should not add permissions when container building is disabled",
		},
		{
			name: "both trivy and container enabled",
			inputs: map[string]interface{}{
				"trivyScanEnabled": true,
				"containerEnabled": true,
				"goVersion":        "1.22",
			},
			expected: map[string]string{
				"security-events": "write",
				"packages":        "write",
				"contents":        "read",
			},
			description: "Should add both security and package permissions when both features are enabled",
		},
		{
			name: "container building not specified",
			inputs: map[string]interface{}{
				"goVersion": "1.22",
			},
			expected:    map[string]string{},
			description: "Should not add permissions when container building is not specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.getRequiredPermissions(nil, tt.inputs)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestWorkflowGenerator_AddEventDrivenContext(t *testing.T) {
	generator := NewWorkflowGenerator("")

	t.Run("default environment context", func(t *testing.T) {
		inputs := make(map[string]interface{})
		generator.addEventDrivenContext(inputs, "default")

		assert.Equal(t, true, inputs["containerBuildOnPR"], "Should set containerBuildOnPR to true for default environment")
		assert.Equal(t, false, inputs["containerPushOnProduction"], "Should set containerPushOnProduction to false for default environment")
	})

	t.Run("staging environment context", func(t *testing.T) {
		inputs := make(map[string]interface{})
		generator.addEventDrivenContext(inputs, "staging")

		assert.Equal(t, true, inputs["containerBuildOnPR"], "Should set containerBuildOnPR to true for staging environment")
		assert.Equal(t, false, inputs["containerPushOnProduction"], "Should set containerPushOnProduction to false for staging environment")
	})

	t.Run("production environment context", func(t *testing.T) {
		inputs := make(map[string]interface{})
		generator.addEventDrivenContext(inputs, "production")

		assert.Equal(t, false, inputs["containerBuildOnPR"], "Should set containerBuildOnPR to false for production environment")
		assert.Equal(t, true, inputs["containerBuildOnProduction"], "Should set containerBuildOnProduction to true for production environment")
		assert.Equal(t, true, inputs["containerPushOnProduction"], "Should set containerPushOnProduction to true for production environment")
	})

	t.Run("preserves existing values", func(t *testing.T) {
		inputs := map[string]interface{}{
			"containerBuildOnPR": false,
		}
		generator.addEventDrivenContext(inputs, "default")

		assert.Equal(t, false, inputs["containerBuildOnPR"], "Should preserve existing containerBuildOnPR value")
		assert.Equal(t, false, inputs["containerPushOnProduction"], "Should still set new values")
	})
}

func TestWorkflowGenerator_NormalizeLegacyInputs(t *testing.T) {
	generator := NewWorkflowGenerator("")

	t.Run("converts legacy security inputs to object structure", func(t *testing.T) {
		inputs := map[string]interface{}{
			"trivyScanEnabled": true,
			"trivySeverity":    "CRITICAL",
		}

		generator.normalizeLegacyInputs(inputs)

		// Check that security object is created
		require.Contains(t, inputs, "security")
		securityObj, ok := inputs["security"].(map[string]interface{})
		require.True(t, ok, "security should be a map")

		require.Contains(t, securityObj, "trivy")
		trivyObj, ok := securityObj["trivy"].(map[string]interface{})
		require.True(t, ok, "security.trivy should be a map")

		assert.Equal(t, true, trivyObj["enabled"], "Should set security.trivy.enabled from trivyScanEnabled")
		assert.Equal(t, "CRITICAL", trivyObj["severity"], "Should set security.trivy.severity from trivySeverity")
		assert.Equal(t, "1", trivyObj["exitCode"], "Should set default exitCode")

		// Check that legacy flat values are preserved for compatibility
		assert.Equal(t, true, inputs["trivyScanEnabled"], "Should preserve legacy trivyScanEnabled")
		assert.Equal(t, "CRITICAL", inputs["trivySeverity"], "Should preserve legacy trivySeverity")
	})

	t.Run("converts legacy container inputs to object structure", func(t *testing.T) {
		inputs := map[string]interface{}{
			"containerEnabled":      true,
			"containerRegistry":     "gcr.io",
			"containerImageName":    "my-app",
			"containerImageTag":     "v1.0.0",
			"containerDockerfile":   "custom.Dockerfile",
			"containerBuildContext": "./app",
			"containerBuildArgs":    `{"VERSION": "1.0"}`,
			"containerPushEnabled":  false,
		}

		generator.normalizeLegacyInputs(inputs)

		// Check that container object is created
		require.Contains(t, inputs, "container")
		containerObj, ok := inputs["container"].(map[string]interface{})
		require.True(t, ok, "container should be a map")

		assert.Equal(t, true, containerObj["enabled"], "Should set container.enabled")
		assert.Equal(t, "gcr.io", containerObj["registry"], "Should set container.registry")
		assert.Equal(t, "my-app", containerObj["imageName"], "Should set container.imageName")
		assert.Equal(t, "v1.0.0", containerObj["imageTag"], "Should set container.imageTag")
		assert.Equal(t, "custom.Dockerfile", containerObj["dockerfile"], "Should set container.dockerfile")
		assert.Equal(t, "./app", containerObj["buildContext"], "Should set container.buildContext")
		assert.Equal(t, `{"VERSION": "1.0"}`, containerObj["buildArgs"], "Should set container.buildArgs")

		// Check nested objects
		require.Contains(t, containerObj, "push")
		pushObj, ok := containerObj["push"].(map[string]interface{})
		require.True(t, ok, "container.push should be a map")
		assert.Equal(t, false, pushObj["enabled"], "Should set container.push.enabled")

		require.Contains(t, containerObj, "build")
		buildObj, ok := containerObj["build"].(map[string]interface{})
		require.True(t, ok, "container.build should be a map")
		assert.Equal(t, true, buildObj["onPR"], "Should set default container.build.onPR")
		assert.Equal(t, true, buildObj["onProduction"], "Should set default container.build.onProduction")

		// Check that legacy flat values are preserved for compatibility
		assert.Equal(t, true, inputs["containerEnabled"], "Should preserve legacy containerEnabled")
		assert.Equal(t, "gcr.io", inputs["containerRegistry"], "Should preserve legacy containerRegistry")
	})

	t.Run("preserves existing object structure", func(t *testing.T) {
		inputs := map[string]interface{}{
			"security": map[string]interface{}{
				"trivy": map[string]interface{}{
					"enabled":  false,
					"severity": "MEDIUM",
					"exitCode": "0",
				},
			},
			"container": map[string]interface{}{
				"enabled": false,
				"push": map[string]interface{}{
					"enabled": false,
				},
			},
		}

		generator.normalizeLegacyInputs(inputs)

		// Check that existing object structure is preserved
		securityObj := inputs["security"].(map[string]interface{})
		trivyObj := securityObj["trivy"].(map[string]interface{})
		assert.Equal(t, false, trivyObj["enabled"], "Should preserve existing security.trivy.enabled")
		assert.Equal(t, "MEDIUM", trivyObj["severity"], "Should preserve existing security.trivy.severity")
		assert.Equal(t, "0", trivyObj["exitCode"], "Should preserve existing security.trivy.exitCode")

		containerObj := inputs["container"].(map[string]interface{})
		assert.Equal(t, false, containerObj["enabled"], "Should preserve existing container.enabled")

		// Check that legacy values are set from object for compatibility
		assert.Equal(t, false, inputs["trivyScanEnabled"], "Should set legacy value from object")
		assert.Equal(t, "MEDIUM", inputs["trivySeverity"], "Should set legacy value from object")
		assert.Equal(t, false, inputs["containerEnabled"], "Should set legacy value from object")
	})

	t.Run("handles missing object properties with defaults", func(t *testing.T) {
		inputs := map[string]interface{}{
			"trivyScanEnabled": true,
			"containerEnabled": true,
		}

		generator.normalizeLegacyInputs(inputs)

		// Check security defaults
		securityObj := inputs["security"].(map[string]interface{})
		trivyObj := securityObj["trivy"].(map[string]interface{})
		assert.Equal(t, true, trivyObj["enabled"], "Should use trivyScanEnabled value")
		assert.Equal(t, "CRITICAL,HIGH", trivyObj["severity"], "Should use default severity")
		assert.Equal(t, "1", trivyObj["exitCode"], "Should use default exitCode")

		// Check container defaults
		containerObj := inputs["container"].(map[string]interface{})
		assert.Equal(t, true, containerObj["enabled"], "Should use containerEnabled value")
		assert.Equal(t, "ghcr.io", containerObj["registry"], "Should use default registry")
		assert.Equal(t, "${{ github.repository }}", containerObj["imageName"], "Should use default imageName")
		assert.Equal(t, "${{ github.sha }}", containerObj["imageTag"], "Should use default imageTag")
		assert.Equal(t, "Dockerfile", containerObj["dockerfile"], "Should use default dockerfile")
		assert.Equal(t, ".", containerObj["buildContext"], "Should use default buildContext")
		assert.Equal(t, "{}", containerObj["buildArgs"], "Should use default buildArgs")

		pushObj := containerObj["push"].(map[string]interface{})
		assert.Equal(t, true, pushObj["enabled"], "Should use default push.enabled")
		assert.Equal(t, true, pushObj["onProduction"], "Should use default push.onProduction")

		buildObj := containerObj["build"].(map[string]interface{})
		assert.Equal(t, false, buildObj["alwaysBuild"], "Should use default build.alwaysBuild")
		assert.Equal(t, false, buildObj["alwaysPush"], "Should use default build.alwaysPush")
		assert.Equal(t, true, buildObj["onPR"], "Should use default build.onPR")
		assert.Equal(t, true, buildObj["onProduction"], "Should use default build.onProduction")
	})
}

func TestWorkflowGenerator_GetEffectiveInputsWithTemplateDefaults(t *testing.T) {
	generator := NewWorkflowGenerator("")

	t.Run("merges template defaults with user inputs and environment overrides", func(t *testing.T) {
		m := &manifest.Manifest{
			Spec: manifest.ManifestSpec{
				Template: "go-service",
				Inputs: map[string]interface{}{
					"goVersion":        "1.23",
					"containerEnabled": true,
				},
				Environments: map[string]manifest.EnvironmentConfig{
					"production": {
						Inputs: map[string]interface{}{
							"trivySeverity": "CRITICAL",
						},
					},
				},
			},
		}

		inputs := generator.getEffectiveInputs(m, "production")

		// Debug: Print some key values to understand what's happening
		t.Logf("Final inputs[containerEnabled] = %v", inputs["containerEnabled"])
		t.Logf("Final inputs[trivySeverity] = %v", inputs["trivySeverity"])
		t.Logf("Final inputs[goVersion] = %v", inputs["goVersion"])

		// Check if container object exists
		if containerObj, ok := inputs["container"]; ok {
			t.Logf("container object: %+v", containerObj)
		}

		// Check if security object exists
		if securityObj, ok := inputs["security"]; ok {
			t.Logf("security object: %+v", securityObj)
		}

		// Check that user inputs override template defaults
		assert.Equal(t, "1.23", inputs["goVersion"], "Should use user input over template default")
		assert.Equal(t, true, inputs["containerEnabled"], "Should use user input over template default")

		// Check that template defaults are applied when not overridden
		assert.Equal(t, "go test ./...", inputs["testCommand"], "Should use template default for testCommand")
		assert.Equal(t, "go build -o bin/service ./cmd/service", inputs["buildCommand"], "Should use template default for buildCommand")

		// Check that environment overrides are applied
		assert.Equal(t, "CRITICAL", inputs["trivySeverity"], "Should use environment override for trivySeverity")

		// Check that event-driven context is added
		assert.Contains(t, inputs, "containerBuildOnPR", "Should add event-driven context")
		assert.Contains(t, inputs, "containerBuildOnProduction", "Should add event-driven context")

		// Check that normalization is applied
		assert.Contains(t, inputs, "security", "Should create security object")
		assert.Contains(t, inputs, "container", "Should create container object")
	})

	t.Run("works with missing template", func(t *testing.T) {
		m := &manifest.Manifest{
			Spec: manifest.ManifestSpec{
				Template: "nonexistent-template",
				Inputs: map[string]interface{}{
					"customInput": "value",
				},
			},
		}

		inputs := generator.getEffectiveInputs(m, "default")

		// Should still work without template defaults
		assert.Equal(t, "value", inputs["customInput"], "Should preserve user inputs")
		assert.Contains(t, inputs, "containerBuildOnPR", "Should still add event-driven context")
	})
}

func TestWorkflowGenerator_GetValue(t *testing.T) {
	tests := []struct {
		name         string
		obj          map[string]interface{}
		key          string
		defaultValue interface{}
		expected     interface{}
	}{
		{
			name:         "returns existing value",
			obj:          map[string]interface{}{"key": "value"},
			key:          "key",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "returns default for missing key",
			obj:          map[string]interface{}{},
			key:          "missing",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "returns nil value if exists",
			obj:          map[string]interface{}{"key": nil},
			key:          "key",
			defaultValue: "default",
			expected:     nil,
		},
		{
			name:         "handles different types",
			obj:          map[string]interface{}{"bool": true, "int": 42},
			key:          "bool",
			defaultValue: false,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValue(tt.obj, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkflowGenerator_ReplaceGitHubActionsPlaceholders(t *testing.T) {
	generator := NewWorkflowGenerator("")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces GITHUB_ACTOR_PLACEHOLDER",
			input:    "username: GITHUB_ACTOR_PLACEHOLDER",
			expected: "username: ${{ github.actor }}",
		},
		{
			name:     "replaces GITHUB_TOKEN_PLACEHOLDER",
			input:    "password: GITHUB_TOKEN_PLACEHOLDER",
			expected: "password: ${{ secrets.GITHUB_TOKEN }}",
		},
		{
			name:     "replaces multiple placeholders",
			input:    "user: GITHUB_ACTOR_PLACEHOLDER, token: GITHUB_TOKEN_PLACEHOLDER",
			expected: "user: ${{ github.actor }}, token: ${{ secrets.GITHUB_TOKEN }}",
		},
		{
			name:     "handles no placeholders",
			input:    "no placeholders here",
			expected: "no placeholders here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.replaceGitHubActionsPlaceholders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
