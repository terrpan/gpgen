package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions for modular testing

// templateTestCase defines a test case for template validation
type templateTestCase struct {
	name                string
	template            *Template
	expectedName        string
	expectedSetupStep   string
	expectedSetupAction string
	versionInputKey     string
	expectedVersions    []string
}

// testTemplateStructure validates common template structure requirements
func testTemplateStructure(t *testing.T, tc templateTestCase) {
	t.Helper()

	assert.Equal(t, tc.expectedName, tc.template.Name)
	assert.NotEmpty(t, tc.template.Description)
	assert.NotEmpty(t, tc.template.Steps)
	assert.True(t, len(tc.template.Steps) >= 3, "Template should have at least 3 steps")

	// Verify checkout step (should be first)
	checkoutStep := tc.template.Steps[0]
	assert.Equal(t, "checkout", checkoutStep.ID)
	assert.Equal(t, "actions/checkout@v4", checkoutStep.Uses)
}

// testLanguageVersionInput validates language version input configuration
func testLanguageVersionInput(t *testing.T, template *Template, versionKey string, expectedVersions []string) {
	t.Helper()

	versionInput, exists := template.Inputs[versionKey]
	require.True(t, exists, "Template should have %s input", versionKey)
	assert.True(t, versionInput.Required, "%s should be required", versionKey)
	assert.Equal(t, "string", versionInput.Type, "%s should be string type", versionKey)

	for _, version := range expectedVersions {
		assert.Contains(t, versionInput.Options, version, "Should support %s version %s", versionKey, version)
	}
}

// testLanguageSetupStep validates language-specific setup step
func testLanguageSetupStep(t *testing.T, template *Template, setupStepID, expectedAction string) {
	t.Helper()

	hasSetupStep := false
	for _, step := range template.Steps {
		if step.ID == setupStepID {
			hasSetupStep = true
			assert.Equal(t, expectedAction, step.Uses, "Setup step should use correct action")
			break
		}
	}
	assert.True(t, hasSetupStep, "Template should have %s step", setupStepID)
}

// testCommonInputs validates that all templates have security and container inputs
func testCommonInputs(t *testing.T, template *Template) {
	t.Helper()

	// Check security inputs
	securityInput, exists := template.Inputs["security"]
	assert.True(t, exists, "Template should have security input")
	assert.Equal(t, "object", securityInput.Type)

	// Check container inputs
	containerInput, exists := template.Inputs["container"]
	assert.True(t, exists, "Template should have container input")
	assert.Equal(t, "object", containerInput.Type)
}

// testCommonSteps validates that all templates have security and container steps
func testCommonSteps(t *testing.T, template *Template) {
	t.Helper()

	stepIDs := make(map[string]bool)
	for _, step := range template.Steps {
		stepIDs[step.ID] = true
	}

	// Check for security steps
	assert.True(t, stepIDs["security-scan"], "Template should have security-scan step")
	assert.True(t, stepIDs["upload-sarif"], "Template should have upload-sarif step")

	// Check for container steps
	assert.True(t, stepIDs["setup-docker-buildx"], "Template should have setup-docker-buildx step")
	assert.True(t, stepIDs["login-registry"], "Template should have login-registry step")
	assert.True(t, stepIDs["build-and-push"], "Template should have build-and-push step")
}

func TestTemplateManager_LoadTemplate(t *testing.T) {
	tm := NewTemplateManager("")

	tests := []struct {
		name         string
		templateName string
		expectError  bool
	}{
		{
			name:         "load node-app template",
			templateName: "node-app",
			expectError:  false,
		},
		{
			name:         "load go-service template",
			templateName: "go-service",
			expectError:  false,
		},
		{
			name:         "load python-app template",
			templateName: "python-app",
			expectError:  false,
		},
		{
			name:         "load unknown template",
			templateName: "unknown-template",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := tm.LoadTemplate(tt.templateName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, template)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, template)
				assert.Equal(t, tt.templateName, template.Name)
				assert.NotEmpty(t, template.Description)
				assert.NotEmpty(t, template.Steps)
			}
		})
	}
}

func TestTemplateManager_ValidateInputs(t *testing.T) {
	tm := NewTemplateManager("")

	t.Run("valid node-app inputs", func(t *testing.T) {
		inputs := map[string]interface{}{
			"nodeVersion":    "18",
			"packageManager": "npm",
			"testCommand":    "npm test",
			"buildCommand":   "npm run build",
		}

		err := tm.ValidateInputs("node-app", inputs)
		assert.NoError(t, err)
	})

	t.Run("missing required input", func(t *testing.T) {
		inputs := map[string]interface{}{
			"packageManager": "npm",
		}

		err := tm.ValidateInputs("node-app", inputs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required input")
	})

	t.Run("invalid input type", func(t *testing.T) {
		inputs := map[string]interface{}{
			"nodeVersion":    18, // Should be string
			"packageManager": "npm",
			"testCommand":    "npm test",
		}

		err := tm.ValidateInputs("node-app", inputs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
	})

	t.Run("invalid option value", func(t *testing.T) {
		inputs := map[string]interface{}{
			"nodeVersion":    "99", // Invalid version
			"packageManager": "npm",
			"testCommand":    "npm test",
		}

		err := tm.ValidateInputs("node-app", inputs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be one of")
	})
}

func TestNodeAppTemplate(t *testing.T) {
	template := getNodeAppTemplate()

	// Test basic template structure
	testTemplateStructure(t, templateTestCase{
		template:     template,
		expectedName: "node-app",
	})

	// Test Node.js-specific configuration
	testLanguageVersionInput(t, template, "nodeVersion", []string{"16", "18", "20", "22"})
	testLanguageSetupStep(t, template, "setup-node", "actions/setup-node@v4")

	// Test package manager input
	packageManagerInput, exists := template.Inputs["packageManager"]
	require.True(t, exists)
	assert.Equal(t, "string", packageManagerInput.Type)
	assert.Contains(t, packageManagerInput.Options, "npm")
	assert.Contains(t, packageManagerInput.Options, "yarn")
	assert.Contains(t, packageManagerInput.Options, "pnpm")

	// Test common inputs and steps
	testCommonInputs(t, template)
	testCommonSteps(t, template)
}

func TestGoServiceTemplate(t *testing.T) {
	template := getGoServiceTemplate()

	// Test basic template structure
	testTemplateStructure(t, templateTestCase{
		template:     template,
		expectedName: "go-service",
	})

	// Test Go-specific configuration
	testLanguageVersionInput(t, template, "goVersion", []string{"1.21", "1.22", "1.23", "1.24"})
	testLanguageSetupStep(t, template, "setup-go", "actions/setup-go@v4")

	// Test Go-specific inputs
	testCommandInput, exists := template.Inputs["testCommand"]
	require.True(t, exists)
	assert.Equal(t, "string", testCommandInput.Type)
	assert.True(t, testCommandInput.Required)

	buildCommandInput, exists := template.Inputs["buildCommand"]
	require.True(t, exists)
	assert.Equal(t, "string", buildCommandInput.Type)
	assert.True(t, buildCommandInput.Required)

	// Test common inputs and steps
	testCommonInputs(t, template)
	testCommonSteps(t, template)
}

func TestPythonAppTemplate(t *testing.T) {
	template := getPythonAppTemplate()

	// Test basic template structure
	testTemplateStructure(t, templateTestCase{
		template:     template,
		expectedName: "python-app",
	})

	// Test Python-specific configuration
	testLanguageVersionInput(t, template, "pythonVersion", []string{"3.9", "3.10", "3.11", "3.12"})
	testLanguageSetupStep(t, template, "setup-python", "actions/setup-python@v4")

	// Test Python-specific inputs
	requirementsInput, exists := template.Inputs["requirements"]
	require.True(t, exists)
	assert.Equal(t, "string", requirementsInput.Type)

	// Test common inputs and steps
	testCommonInputs(t, template)
	testCommonSteps(t, template)
}

func TestTemplateManager_ListTemplates(t *testing.T) {
	tm := NewTemplateManager("")
	templates := tm.ListTemplates()

	assert.Contains(t, templates, "node-app")
	assert.Contains(t, templates, "go-service")
	assert.Contains(t, templates, "python-app")
	assert.Len(t, templates, 3)
}

func TestValidateInputValue(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		value       interface{}
		inputDef    Input
		expectError bool
	}{
		{
			name:      "valid string",
			inputName: "test",
			value:     "hello",
			inputDef:  Input{Type: "string"},
		},
		{
			name:        "invalid string type",
			inputName:   "test",
			value:       123,
			inputDef:    Input{Type: "string"},
			expectError: true,
		},
		{
			name:      "valid number int",
			inputName: "test",
			value:     123,
			inputDef:  Input{Type: "number"},
		},
		{
			name:      "valid number float",
			inputName: "test",
			value:     123.45,
			inputDef:  Input{Type: "number"},
		},
		{
			name:        "invalid number type",
			inputName:   "test",
			value:       "not-a-number",
			inputDef:    Input{Type: "number"},
			expectError: true,
		},
		{
			name:      "valid boolean true",
			inputName: "test",
			value:     true,
			inputDef:  Input{Type: "boolean"},
		},
		{
			name:      "valid boolean false",
			inputName: "test",
			value:     false,
			inputDef:  Input{Type: "boolean"},
		},
		{
			name:        "invalid boolean type",
			inputName:   "test",
			value:       "true",
			inputDef:    Input{Type: "boolean"},
			expectError: true,
		},
		{
			name:      "valid array",
			inputName: "test",
			value:     []interface{}{"a", "b", "c"},
			inputDef:  Input{Type: "array"},
		},
		{
			name:        "invalid array type",
			inputName:   "test",
			value:       "not-an-array",
			inputDef:    Input{Type: "array"},
			expectError: true,
		},
		{
			name:      "valid option",
			inputName: "test",
			value:     "option1",
			inputDef:  Input{Type: "string", Options: []string{"option1", "option2"}},
		},
		{
			name:        "invalid option",
			inputName:   "test",
			value:       "option3",
			inputDef:    Input{Type: "string", Options: []string{"option1", "option2"}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputValue(tt.inputName, tt.value, tt.inputDef)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
