package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	assert.Equal(t, "node-app", template.Name)
	assert.NotEmpty(t, template.Description)
	assert.NotEmpty(t, template.Steps)

	// Check required inputs
	nodeVersionInput, exists := template.Inputs["nodeVersion"]
	require.True(t, exists)
	assert.True(t, nodeVersionInput.Required)
	assert.Equal(t, "string", nodeVersionInput.Type)
	assert.Contains(t, nodeVersionInput.Options, "18")

	// Check steps structure
	assert.True(t, len(template.Steps) >= 4) // At least checkout, setup, install, test

	// Verify checkout step
	checkoutStep := template.Steps[0]
	assert.Equal(t, "checkout", checkoutStep.ID)
	assert.Equal(t, "actions/checkout@v4", checkoutStep.Uses)
}

func TestGoServiceTemplate(t *testing.T) {
	template := getGoServiceTemplate()

	assert.Equal(t, "go-service", template.Name)
	assert.NotEmpty(t, template.Description)
	assert.NotEmpty(t, template.Steps)

	// Check Go version input
	goVersionInput, exists := template.Inputs["goVersion"]
	require.True(t, exists)
	assert.True(t, goVersionInput.Required)
	assert.Equal(t, "string", goVersionInput.Type)
	assert.Contains(t, goVersionInput.Options, "1.21")

	// Check that it has Go-specific steps
	hasSetupGo := false
	for _, step := range template.Steps {
		if step.ID == "setup-go" {
			hasSetupGo = true
			assert.Equal(t, "actions/setup-go@v4", step.Uses)
			break
		}
	}
	assert.True(t, hasSetupGo, "Go template should have setup-go step")
}

func TestPythonAppTemplate(t *testing.T) {
	template := getPythonAppTemplate()

	assert.Equal(t, "python-app", template.Name)
	assert.NotEmpty(t, template.Description)
	assert.NotEmpty(t, template.Steps)

	// Check Python version input
	pythonVersionInput, exists := template.Inputs["pythonVersion"]
	require.True(t, exists)
	assert.True(t, pythonVersionInput.Required)
	assert.Equal(t, "string", pythonVersionInput.Type)
	assert.Contains(t, pythonVersionInput.Options, "3.11")

	// Check that it has Python-specific steps
	hasSetupPython := false
	for _, step := range template.Steps {
		if step.ID == "setup-python" {
			hasSetupPython = true
			assert.Equal(t, "actions/setup-python@v4", step.Uses)
			break
		}
	}
	assert.True(t, hasSetupPython, "Python template should have setup-python step")
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
