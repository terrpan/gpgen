package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifest_ValidMinimalManifest(t *testing.T) {
	yamlContent := `
apiVersion: gpgen.dev/v1
kind: Pipeline
spec:
  template: "go-service"
  inputs:
    goVersion: "1.24"
`

	manifest, err := ParseManifest([]byte(yamlContent))
	require.NoError(t, err)

	assert.Equal(t, "gpgen.dev/v1", manifest.APIVersion)
	assert.Equal(t, "Pipeline", manifest.Kind)
	assert.Equal(t, "go-service", manifest.Spec.Template)
	assert.Equal(t, "1.24", manifest.Spec.Inputs["goVersion"])
}

func TestParseManifest_ValidComplexManifest(t *testing.T) {
	yamlContent := `
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: my-go-service
  annotations:
    gpgen.dev/description: "Test service"
    gpgen.dev/validation-mode: "strict"
spec:
  template: "go-service"
  inputs:
    goVersion: "1.24"
    deployEnvironments: ["staging", "production"]
  customSteps:
    - name: "integration-tests"
      position: "after:test"
      run: "go test -tags=integration ./..."
      timeout-minutes: 15
  overrides:
    test:
      timeout-minutes: 20
      env:
        GO_TEST_TIMEOUT: "15m"
  environments:
    staging:
      inputs:
        deployTarget: "staging-k8s-cluster"
        replicas: 2
`

	manifest, err := ParseManifest([]byte(yamlContent))
	require.NoError(t, err)

	// Test metadata
	assert.Equal(t, "my-go-service", manifest.Metadata.Name)
	assert.Equal(t, "Test service", manifest.Metadata.Annotations["gpgen.dev/description"])
	assert.Equal(t, "strict", manifest.Metadata.Annotations["gpgen.dev/validation-mode"])

	// Test spec
	assert.Equal(t, "go-service", manifest.Spec.Template)
	assert.Equal(t, "1.24", manifest.Spec.Inputs["goVersion"])
	assert.Len(t, manifest.Spec.Inputs["deployEnvironments"], 2)

	// Test custom steps
	require.Len(t, manifest.Spec.CustomSteps, 1)
	step := manifest.Spec.CustomSteps[0]
	assert.Equal(t, "integration-tests", step.Name)
	assert.Equal(t, "after:test", step.Position)
	assert.Equal(t, "go test -tags=integration ./...", step.Run)
	assert.Equal(t, 15, *step.TimeoutMinutes)

	// Test overrides
	testOverride, exists := manifest.Spec.Overrides["test"]
	require.True(t, exists)
	assert.Equal(t, 20, *testOverride.TimeoutMinutes)
	assert.Equal(t, "15m", testOverride.Env["GO_TEST_TIMEOUT"])

	// Test environments
	staging, exists := manifest.Spec.Environments["staging"]
	require.True(t, exists)
	assert.Equal(t, "staging-k8s-cluster", staging.Inputs["deployTarget"])
	assert.Equal(t, 2, staging.Inputs["replicas"])
}

func TestParseManifest_InvalidYAML(t *testing.T) {
	invalidYAML := `
apiVersion: gpgen.dev/v1
kind: Pipeline
spec:
  template: "go-service"
  invalid: [
`

	_, err := ParseManifest([]byte(invalidYAML))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestParseManifest_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		errorMsg string
	}{
		{
			name: "missing apiVersion",
			yaml: `
kind: Pipeline
spec:
  template: "go-service"
`,
			errorMsg: "apiVersion is required",
		},
		{
			name: "missing kind",
			yaml: `
apiVersion: gpgen.dev/v1
spec:
  template: "go-service"
`,
			errorMsg: "kind is required",
		},
		{
			name: "missing spec",
			yaml: `
apiVersion: gpgen.dev/v1
kind: Pipeline
`,
			errorMsg: "spec is required",
		},
		{
			name: "missing template",
			yaml: `
apiVersion: gpgen.dev/v1
kind: Pipeline
spec:
  inputs:
    goVersion: "1.24"
`,
			errorMsg: "template is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(tt.yaml))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestValidateManifest_ValidManifests(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
	}{
		{
			name: "minimal valid manifest",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					Inputs:   map[string]interface{}{"goVersion": "1.24"},
				},
			},
		},
		{
			name: "manifest with custom steps",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "node-app",
					Inputs:   map[string]interface{}{"nodeVersion": "20"},
					CustomSteps: []CustomStep{
						{
							Name:     "test-step",
							Position: "after:test",
							Run:      "echo hello",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			assert.NoError(t, err)
		})
	}
}

func TestValidateManifest_InvalidManifests(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		errorMsg string
	}{
		{
			name: "invalid API version",
			manifest: &Manifest{
				APIVersion: "v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
				},
			},
			errorMsg: "invalid apiVersion",
		},
		{
			name: "invalid kind",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Workflow",
				Spec: ManifestSpec{
					Template: "go-service",
				},
			},
			errorMsg: "invalid kind",
		},
		{
			name: "invalid template",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "invalid-template",
				},
			},
			errorMsg: "invalid template",
		},
		{
			name: "invalid position format",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:     "test-step",
							Position: "invalid-position",
							Run:      "echo hello",
						},
					},
				},
			},
			errorMsg: "invalid position format",
		},
		{
			name: "step with neither uses nor run",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:     "test-step",
							Position: "after:test",
						},
					},
				},
			},
			errorMsg: "step must have either 'uses' or 'run'",
		},
		{
			name: "step with both uses and run",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:     "test-step",
							Position: "after:test",
							Uses:     "actions/checkout@v4",
							Run:      "echo hello",
						},
					},
				},
			},
			errorMsg: "step cannot have both 'uses' and 'run'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		position string
		valid    bool
	}{
		{"after:test", true},
		{"before:deploy", true},
		{"replace:build", true},
		{"after:integration-tests", true},
		{"before:setup-node", true},
		{"invalid", false},
		{"after:", false},
		{":test", false},
		{"during:test", false},
		{"after:Test", false},      // uppercase not allowed
		{"after:test_case", false}, // underscore not allowed
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			err := validatePosition(tt.position)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetValidationMode(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		expected ValidationMode
	}{
		{
			name: "default strict mode",
			manifest: &Manifest{
				Metadata: &ManifestMetadata{},
			},
			expected: ValidationModeStrict,
		},
		{
			name: "explicit strict mode",
			manifest: &Manifest{
				Metadata: &ManifestMetadata{
					Annotations: map[string]string{
						"gpgen.dev/validation-mode": "strict",
					},
				},
			},
			expected: ValidationModeStrict,
		},
		{
			name: "relaxed mode",
			manifest: &Manifest{
				Metadata: &ManifestMetadata{
					Annotations: map[string]string{
						"gpgen.dev/validation-mode": "relaxed",
					},
				},
			},
			expected: ValidationModeRelaxed,
		},
		{
			name: "nil metadata",
			manifest: &Manifest{
				Metadata: nil,
			},
			expected: ValidationModeStrict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := GetValidationMode(tt.manifest)
			assert.Equal(t, tt.expected, mode)
		})
	}
}

func TestLoadManifestFromFile_Success(t *testing.T) {
	// Create a temporary manifest file
	content := `
apiVersion: gpgen.dev/v1
kind: Pipeline
spec:
  template: "go-service"
  inputs:
    goVersion: "1.24"
`

	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "gpgen.yaml")
	err := os.WriteFile(manifestPath, []byte(content), 0644)
	require.NoError(t, err)

	// Load and validate
	manifest, err := LoadManifestFromFile(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, "gpgen.dev/v1", manifest.APIVersion)
	assert.Equal(t, "go-service", manifest.Spec.Template)
}

func TestLoadManifestFromFile_FileNotFound(t *testing.T) {
	_, err := LoadManifestFromFile("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manifest file")
}

func TestValidateManifest_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		errorMsg string
	}{
		{
			name: "timeout too low",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:           "test-step",
							Position:       "after:test",
							Run:            "echo hello",
							TimeoutMinutes: intPtr(0),
						},
					},
				},
			},
			errorMsg: "timeout-minutes must be between 1 and 360",
		},
		{
			name: "timeout too high",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:           "test-step",
							Position:       "after:test",
							Run:            "echo hello",
							TimeoutMinutes: intPtr(400),
						},
					},
				},
			},
			errorMsg: "timeout-minutes must be between 1 and 360",
		},
		{
			name: "empty step name",
			manifest: &Manifest{
				APIVersion: "gpgen.dev/v1",
				Kind:       "Pipeline",
				Spec: ManifestSpec{
					Template: "go-service",
					CustomSteps: []CustomStep{
						{
							Name:     "",
							Position: "after:test",
							Run:      "echo hello",
						},
					},
				},
			},
			errorMsg: "step name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestParseAndValidateManifest_Integration(t *testing.T) {
	yamlContent := `
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: test-pipeline
  annotations:
    gpgen.dev/validation-mode: "strict"
spec:
  template: "node-app"
  inputs:
    nodeVersion: "20"
    packageManager: "npm"
  customSteps:
    - name: "security-scan"
      position: "after:test"
      uses: "securecodewarrior/github-action-add-sarif@v1"
      with:
        sarif-file: "security-scan.sarif"
    - name: "custom-test"
      position: "before:build"
      run: "npm run custom-test"
      timeout-minutes: 10
  overrides:
    test:
      timeout-minutes: 20
  environments:
    staging:
      inputs:
        deployTarget: "staging"
      customSteps:
        - name: "staging-test"
          position: "after:deploy"
          run: "curl -f https://staging.example.com/health"
`

	// Test full parse and validate flow
	manifest, err := ParseManifest([]byte(yamlContent))
	require.NoError(t, err)

	err = ValidateManifest(manifest)
	require.NoError(t, err)

	// Verify validation mode detection
	mode := GetValidationMode(manifest)
	assert.Equal(t, ValidationModeStrict, mode)

	// Verify complex structure is parsed correctly
	assert.Len(t, manifest.Spec.CustomSteps, 2)
	assert.Len(t, manifest.Spec.Environments["staging"].CustomSteps, 1)
	assert.Equal(t, "staging", manifest.Spec.Environments["staging"].Inputs["deployTarget"])
}

// Helper function for creating int pointers in tests
func intPtr(i int) *int {
	return &i
}
