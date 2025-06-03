package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLanguageConstants(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		expected string
	}{
		{
			name:     "Go language constant",
			language: LanguageGo,
			expected: "go",
		},
		{
			name:     "Node language constant",
			language: LanguageNode,
			expected: "node",
		},
		{
			name:     "Python language constant",
			language: LanguagePython,
			expected: "python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.language))
		})
	}
}

func TestPackageManagerConstants(t *testing.T) {
	tests := []struct {
		name           string
		packageManager PackageManager
		expected       string
	}{
		{
			name:           "npm package manager",
			packageManager: PackageManagerNpm,
			expected:       "npm",
		},
		{
			name:           "yarn package manager",
			packageManager: PackageManagerYarn,
			expected:       "yarn",
		},
		{
			name:           "pip package manager",
			packageManager: PackageManagerPip,
			expected:       "pip",
		},
		{
			name:           "poetry package manager",
			packageManager: PackageManagerPoetry,
			expected:       "poetry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.packageManager))
		})
	}
}

func TestSecuritySeverityConstants(t *testing.T) {
	tests := []struct {
		name     string
		severity SecuritySeverity
		expected string
	}{
		{
			name:     "critical severity",
			severity: SeverityCritical,
			expected: "CRITICAL",
		},
		{
			name:     "high severity",
			severity: SeverityHigh,
			expected: "HIGH",
		},
		{
			name:     "critical and high severity",
			severity: SeverityCriticalHigh,
			expected: "CRITICAL,HIGH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.severity))
		})
	}
}

func TestInputFieldConstants(t *testing.T) {
	tests := []struct {
		name       string
		inputField InputField
		expected   string
	}{
		{
			name:       "Go version input field",
			inputField: InputFieldGoVersion,
			expected:   "goVersion",
		},
		{
			name:       "Node version input field",
			inputField: InputFieldNodeVersion,
			expected:   "nodeVersion",
		},
		{
			name:       "package manager input field",
			inputField: InputFieldPackageManager,
			expected:   "packageManager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.inputField))
		})
	}
}

func TestConfiguration_GetLanguageConfig(t *testing.T) {
	tests := []struct {
		name         string
		language     Language
		expectExists bool
	}{
		{
			name:         "get Go config",
			language:     LanguageGo,
			expectExists: true,
		},
		{
			name:         "get Node config",
			language:     LanguageNode,
			expectExists: true,
		},
		{
			name:         "get Python config",
			language:     LanguagePython,
			expectExists: true,
		},
		{
			name:         "get unknown language",
			language:     Language("unknown"),
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, exists := Config.GetLanguageConfig(tt.language)
			assert.Equal(t, tt.expectExists, exists)

			if tt.expectExists {
				assert.NotEmpty(t, config.DefaultVersion)
				assert.NotEmpty(t, config.Versions)
			}
		})
	}
}

func TestConfiguration_IsValidVersion(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		version  string
		expected bool
	}{
		{
			name:     "valid Go version",
			language: LanguageGo,
			version:  "1.21",
			expected: true,
		},
		{
			name:     "invalid Go version",
			language: LanguageGo,
			version:  "1.15",
			expected: false,
		},
		{
			name:     "valid Node version",
			language: LanguageNode,
			version:  "18",
			expected: true,
		},
		{
			name:     "invalid Node version",
			language: LanguageNode,
			version:  "14",
			expected: false,
		},
		{
			name:     "unknown language",
			language: Language("unknown"),
			version:  "1.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config.IsValidVersion(tt.language, tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfiguration_IsValidPackageManager(t *testing.T) {
	tests := []struct {
		name           string
		language       Language
		packageManager PackageManager
		expected       bool
	}{
		{
			name:           "npm for Node",
			language:       LanguageNode,
			packageManager: PackageManagerNpm,
			expected:       true,
		},
		{
			name:           "pip for Python",
			language:       LanguagePython,
			packageManager: PackageManagerPip,
			expected:       true,
		},
		{
			name:           "npm for Python (invalid)",
			language:       LanguagePython,
			packageManager: PackageManagerNpm,
			expected:       false,
		},
		{
			name:           "any for Go (no package managers)",
			language:       LanguageGo,
			packageManager: PackageManagerNpm,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config.IsValidPackageManager(tt.language, tt.packageManager)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfiguration_IsValidSecuritySeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity SecuritySeverity
		expected bool
	}{
		{
			name:     "valid critical severity",
			severity: SeverityCritical,
			expected: true,
		},
		{
			name:     "valid critical and high severity",
			severity: SeverityCriticalHigh,
			expected: true,
		},
		{
			name:     "invalid severity",
			severity: SecuritySeverity("INVALID"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config.IsValidSecuritySeverity(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfiguration_GetDefaults(t *testing.T) {
	tests := []struct {
		name         string
		language     Language
		expectFields []string
	}{
		{
			name:         "Go defaults",
			language:     LanguageGo,
			expectFields: []string{"version", "testCommand"},
		},
		{
			name:         "Node defaults",
			language:     LanguageNode,
			expectFields: []string{"version", "testCommand", "buildCommand", "packageManager"},
		},
		{
			name:         "Python defaults",
			language:     LanguagePython,
			expectFields: []string{"version", "testCommand", "lintCommand", "requirements", "packageManager"},
		},
		{
			name:         "unknown language",
			language:     Language("unknown"),
			expectFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := Config.GetDefaults(tt.language)

			if len(tt.expectFields) == 0 {
				assert.Empty(t, defaults)
				return
			}

			for _, field := range tt.expectFields {
				assert.Contains(t, defaults, field, "expected field %s to be present", field)
				assert.NotEmpty(t, defaults[field], "expected field %s to have a non-empty value", field)
			}
		})
	}
}

func TestConfiguration_GetPackageManagerOptions(t *testing.T) {
	tests := []struct {
		name         string
		language     Language
		expectLength int
	}{
		{
			name:         "Node package managers",
			language:     LanguageNode,
			expectLength: 3, // npm, yarn, pnpm
		},
		{
			name:         "Python package managers",
			language:     LanguagePython,
			expectLength: 3, // pip, poetry, pipenv
		},
		{
			name:         "Go package managers (none)",
			language:     LanguageGo,
			expectLength: 0,
		},
		{
			name:         "unknown language",
			language:     Language("unknown"),
			expectLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := Config.GetPackageManagerOptions(tt.language)
			assert.Len(t, options, tt.expectLength)

			// Verify all options are strings
			for _, option := range options {
				assert.NotEmpty(t, option)
			}
		})
	}
}

func TestConfiguration_GetSecuritySeverityOptions(t *testing.T) {
	t.Run("get security severity options", func(t *testing.T) {
		options := Config.GetSecuritySeverityOptions()

		assert.Len(t, options, 6) // All defined severity levels
		assert.Contains(t, options, "CRITICAL")
		assert.Contains(t, options, "HIGH")
		assert.Contains(t, options, "MEDIUM")
		assert.Contains(t, options, "LOW")
		assert.Contains(t, options, "CRITICAL,HIGH")
		assert.Contains(t, options, "CRITICAL,HIGH,MEDIUM")
	})
}

func TestConfiguration_ValidateTemplateInput(t *testing.T) {
	tests := []struct {
		name        string
		inputField  InputField
		value       interface{}
		language    Language
		expectError bool
	}{
		{
			name:        "valid Go version",
			inputField:  InputFieldGoVersion,
			value:       "1.21",
			language:    LanguageGo,
			expectError: false,
		},
		{
			name:        "invalid Go version",
			inputField:  InputFieldGoVersion,
			value:       "1.15",
			language:    LanguageGo,
			expectError: true,
		},
		{
			name:        "valid Node package manager",
			inputField:  InputFieldPackageManager,
			value:       "npm",
			language:    LanguageNode,
			expectError: false,
		},
		{
			name:        "invalid Node package manager",
			inputField:  InputFieldPackageManager,
			value:       "invalid",
			language:    LanguageNode,
			expectError: true,
		},
		{
			name:        "non-string version",
			inputField:  InputFieldGoVersion,
			value:       123,
			language:    LanguageGo,
			expectError: true,
		},
		{
			name:        "non-string package manager",
			inputField:  InputFieldPackageManager,
			value:       123,
			language:    LanguageNode,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Config.ValidateTemplateInput(tt.inputField, tt.value, tt.language)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfiguration_ValidateTemplateInputLegacy(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		value       interface{}
		language    Language
		expectError bool
	}{
		{
			name:        "valid legacy goVersion",
			inputName:   "goVersion",
			value:       "1.21",
			language:    LanguageGo,
			expectError: false,
		},
		{
			name:        "invalid legacy input name",
			inputName:   "invalidField",
			value:       "value",
			language:    LanguageGo,
			expectError: true,
		},
		{
			name:        "valid legacy packageManager",
			inputName:   "packageManager",
			value:       "npm",
			language:    LanguageNode,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Config.ValidateTemplateInputLegacy(tt.inputName, tt.value, tt.language)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfiguration_GetValidInputFields(t *testing.T) {
	tests := []struct {
		name         string
		language     Language
		expectFields []InputField
	}{
		{
			name:     "Go input fields",
			language: LanguageGo,
			expectFields: []InputField{
				InputFieldGoVersion,
				InputFieldTestCommand,
				InputFieldBuildCommand,
			},
		},
		{
			name:     "Node input fields",
			language: LanguageNode,
			expectFields: []InputField{
				InputFieldNodeVersion,
				InputFieldPackageManager,
				InputFieldTestCommand,
				InputFieldBuildCommand,
			},
		},
		{
			name:         "unknown language",
			language:     Language("unknown"),
			expectFields: []InputField{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := Config.GetValidInputFields(tt.language)
			assert.ElementsMatch(t, tt.expectFields, fields)
		})
	}
}

func TestConfiguration_GetDefaultValue(t *testing.T) {
	tests := []struct {
		name        string
		inputField  InputField
		language    Language
		expectError bool
		expectValue interface{}
	}{
		{
			name:        "Go version default",
			inputField:  InputFieldGoVersion,
			language:    LanguageGo,
			expectError: false,
			expectValue: "1.21",
		},
		{
			name:        "Node version for Go language (invalid)",
			inputField:  InputFieldNodeVersion,
			language:    LanguageGo,
			expectError: true,
		},
		{
			name:        "package manager for Node",
			inputField:  InputFieldPackageManager,
			language:    LanguageNode,
			expectError: false,
			expectValue: "npm",
		},
		{
			name:        "unknown language",
			inputField:  InputFieldGoVersion,
			language:    Language("unknown"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := Config.GetDefaultValue(tt.inputField, tt.language)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

func TestConfiguration_GetAllDefaults(t *testing.T) {
	tests := []struct {
		name          string
		language      Language
		expectMinKeys int
	}{
		{
			name:          "Go defaults",
			language:      LanguageGo,
			expectMinKeys: 2, // version, testCommand, buildCommand
		},
		{
			name:          "Node defaults",
			language:      LanguageNode,
			expectMinKeys: 3, // version, testCommand, buildCommand, packageManager
		},
		{
			name:          "Python defaults",
			language:      LanguagePython,
			expectMinKeys: 4, // version, testCommand, lintCommand, requirements, packageManager
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaults := Config.GetAllDefaults(tt.language)
			assert.GreaterOrEqual(t, len(defaults), tt.expectMinKeys)

			// Verify all values are non-empty
			for field, value := range defaults {
				assert.NotEmpty(t, value, "field %s should have a non-empty value", field)
			}
		})
	}
}

func TestTypedDefaults(t *testing.T) {
	td := NewTypedDefaults()
	require.NotNil(t, td)

	t.Run("get Go version", func(t *testing.T) {
		version := td.GetGoVersion()
		assert.Equal(t, "1.21", version)
	})

	t.Run("get Node version", func(t *testing.T) {
		version := td.GetNodeVersion()
		assert.Equal(t, "18", version)
	})

	t.Run("get Python version", func(t *testing.T) {
		version := td.GetPythonVersion()
		assert.Equal(t, "3.11", version)
	})

	t.Run("get default package manager for Node", func(t *testing.T) {
		manager, err := td.GetDefaultPackageManager(LanguageNode)
		assert.NoError(t, err)
		assert.Equal(t, PackageManagerNpm, manager)
	})

	t.Run("get default package manager for Go (should error)", func(t *testing.T) {
		_, err := td.GetDefaultPackageManager(LanguageGo)
		assert.Error(t, err)
	})

	t.Run("get default test command", func(t *testing.T) {
		cmd, err := td.GetDefaultTestCommand(LanguageGo)
		assert.NoError(t, err)
		assert.Equal(t, "go test ./...", cmd)
	})

	t.Run("get default build command", func(t *testing.T) {
		cmd, err := td.GetDefaultBuildCommand(LanguageGo)
		assert.NoError(t, err)
		assert.Equal(t, "go build -o bin/service ./cmd/service", cmd)
	})

	t.Run("get default lint command for Python", func(t *testing.T) {
		cmd, err := td.GetDefaultLintCommand(LanguagePython)
		assert.NoError(t, err)
		assert.Equal(t, "flake8", cmd)
	})

	t.Run("get default requirements file for Python", func(t *testing.T) {
		file, err := td.GetDefaultRequirementsFile(LanguagePython)
		assert.NoError(t, err)
		assert.Equal(t, "requirements.txt", file)
	})

	t.Run("access global Defaults instance", func(t *testing.T) {
		assert.NotNil(t, Defaults)
		assert.Equal(t, "1.21", Defaults.GetGoVersion())
	})
}

func TestLegacyCompatibility(t *testing.T) {
	t.Run("LanguageVersions map", func(t *testing.T) {
		assert.NotEmpty(t, LanguageVersions)
		assert.Contains(t, LanguageVersions, "go")
		assert.Contains(t, LanguageVersions, "node")
		assert.Contains(t, LanguageVersions, "python")
	})

	t.Run("PackageManagers map", func(t *testing.T) {
		assert.NotEmpty(t, PackageManagers)
		assert.Contains(t, PackageManagers, "node")
		assert.Contains(t, PackageManagers, "python")
		assert.NotContains(t, PackageManagers, "go") // Go doesn't have package managers
	})

	t.Run("SecuritySeverityLevels slice", func(t *testing.T) {
		assert.NotEmpty(t, SecuritySeverityLevels)
		assert.Contains(t, SecuritySeverityLevels, "CRITICAL")
		assert.Contains(t, SecuritySeverityLevels, "HIGH")
	})
}

func TestLanguageInputFieldsMapping(t *testing.T) {
	t.Run("Go input fields", func(t *testing.T) {
		fields := LanguageInputFields[LanguageGo]
		assert.Contains(t, fields, InputFieldGoVersion)
		assert.Contains(t, fields, InputFieldTestCommand)
		assert.Contains(t, fields, InputFieldBuildCommand)
		assert.NotContains(t, fields, InputFieldPackageManager) // Go doesn't use package managers
	})

	t.Run("Node input fields", func(t *testing.T) {
		fields := LanguageInputFields[LanguageNode]
		assert.Contains(t, fields, InputFieldNodeVersion)
		assert.Contains(t, fields, InputFieldPackageManager)
		assert.Contains(t, fields, InputFieldTestCommand)
		assert.Contains(t, fields, InputFieldBuildCommand)
	})

	t.Run("Python input fields", func(t *testing.T) {
		fields := LanguageInputFields[LanguagePython]
		assert.Contains(t, fields, InputFieldPythonVersion)
		assert.Contains(t, fields, InputFieldPackageManager)
		assert.Contains(t, fields, InputFieldTestCommand)
		assert.Contains(t, fields, InputFieldLintCommand)
		assert.Contains(t, fields, InputFieldRequirements)
	})
}

func TestGetVersionField(t *testing.T) {
	tests := []struct {
		name     string
		language Language
		expected InputField
	}{
		{
			name:     "Go language version field",
			language: LanguageGo,
			expected: InputFieldGoVersion,
		},
		{
			name:     "Node language version field",
			language: LanguageNode,
			expected: InputFieldNodeVersion,
		},
		{
			name:     "Python language version field",
			language: LanguagePython,
			expected: InputFieldPythonVersion,
		},
		{
			name:     "Unknown language fallback",
			language: Language("unknown"),
			expected: InputFieldGoVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVersionField(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigGetVersionsForLanguage(t *testing.T) {
	tests := []struct {
		name        string
		language    Language
		expectError bool
		expected    []string
	}{
		{
			name:        "Go versions",
			language:    LanguageGo,
			expectError: false,
			expected:    []string{"1.21", "1.22", "1.23", "1.24"},
		},
		{
			name:        "Node versions",
			language:    LanguageNode,
			expectError: false,
			expected:    []string{"16", "18", "20", "22"},
		},
		{
			name:        "Python versions",
			language:    LanguagePython,
			expectError: false,
			expected:    []string{"3.9", "3.10", "3.11", "3.12"},
		},
		{
			name:        "Invalid language",
			language:    Language("invalid"),
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := Config.GetVersionsForLanguage(tt.language)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, versions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, versions)
			}
		})
	}
}

func TestConfigGetPackageManagersForLanguage(t *testing.T) {
	tests := []struct {
		name        string
		language    Language
		expectError bool
		expected    []PackageManager
	}{
		{
			name:        "Go package managers (none)",
			language:    LanguageGo,
			expectError: false,
			expected:    []PackageManager{},
		},
		{
			name:        "Node package managers",
			language:    LanguageNode,
			expectError: false,
			expected:    []PackageManager{PackageManagerNpm, PackageManagerYarn, PackageManagerPnpm},
		},
		{
			name:        "Python package managers",
			language:    LanguagePython,
			expectError: false,
			expected:    []PackageManager{PackageManagerPip, PackageManagerPoetry, PackageManagerPipenv},
		},
		{
			name:        "Invalid language",
			language:    Language("invalid"),
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managers, err := Config.GetPackageManagersForLanguage(tt.language)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, managers)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, managers)
			}
		})
	}
}

func TestConfigIsValidInputFieldForLanguage(t *testing.T) {
	tests := []struct {
		name     string
		field    InputField
		language Language
		expected bool
	}{
		{
			name:     "Go version field for Go",
			field:    InputFieldGoVersion,
			language: LanguageGo,
			expected: true,
		},
		{
			name:     "Node version field for Node",
			field:    InputFieldNodeVersion,
			language: LanguageNode,
			expected: true,
		},
		{
			name:     "Package manager for Python",
			field:    InputFieldPackageManager,
			language: LanguagePython,
			expected: true,
		},
		{
			name:     "Go version field for Node (invalid)",
			field:    InputFieldGoVersion,
			language: LanguageNode,
			expected: false,
		},
		{
			name:     "Python version field for Go (invalid)",
			field:    InputFieldPythonVersion,
			language: LanguageGo,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config.IsValidInputFieldForLanguage(tt.field, tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigValidateTemplateInputEnhanced(t *testing.T) {
	tests := []struct {
		name        string
		field       InputField
		value       interface{}
		language    Language
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid Go version",
			field:       InputFieldGoVersion,
			value:       "1.21",
			language:    LanguageGo,
			expectError: false,
		},
		{
			name:        "Invalid Go version",
			field:       InputFieldGoVersion,
			value:       "1.19",
			language:    LanguageGo,
			expectError: true,
			errorMsg:    "invalid go version: 1.19. Supported versions: [1.21 1.22 1.23 1.24]",
		},
		{
			name:        "Empty version string",
			field:       InputFieldGoVersion,
			value:       "",
			language:    LanguageGo,
			expectError: true,
			errorMsg:    "go version cannot be empty",
		},
		{
			name:        "Invalid field for language",
			field:       InputFieldPythonVersion,
			value:       "3.11",
			language:    LanguageGo,
			expectError: true,
			errorMsg:    "input field pythonVersion is not valid for language go",
		},
		{
			name:        "Valid package manager",
			field:       InputFieldPackageManager,
			value:       "npm",
			language:    LanguageNode,
			expectError: false,
		},
		{
			name:        "Invalid package manager",
			field:       InputFieldPackageManager,
			value:       "bower",
			language:    LanguageNode,
			expectError: true,
			errorMsg:    "invalid package manager for node: bower. Supported managers: [npm yarn pnpm]",
		},
		{
			name:        "Empty command",
			field:       InputFieldTestCommand,
			value:       "",
			language:    LanguageGo,
			expectError: true,
			errorMsg:    "testCommand command cannot be empty",
		},
		{
			name:        "Valid command",
			field:       InputFieldTestCommand,
			value:       "go test ./...",
			language:    LanguageGo,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Config.ValidateTemplateInput(tt.field, tt.value, tt.language)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidateAllInputs(t *testing.T) {
	tests := []struct {
		name        string
		inputs      map[InputField]interface{}
		language    Language
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid Go inputs",
			inputs: map[InputField]interface{}{
				InputFieldGoVersion: "1.21",
			},
			language:    LanguageGo,
			expectError: false,
		},
		{
			name: "Valid Node inputs",
			inputs: map[InputField]interface{}{
				InputFieldNodeVersion:    "18",
				InputFieldPackageManager: "npm",
			},
			language:    LanguageNode,
			expectError: false,
		},
		{
			name: "Missing required field",
			inputs: map[InputField]interface{}{
				InputFieldNodeVersion: "18",
				// Missing package manager
			},
			language:    LanguageNode,
			expectError: true,
			errorMsg:    "missing required field: packageManager",
		},
		{
			name: "Invalid language",
			inputs: map[InputField]interface{}{
				InputFieldGoVersion: "1.21",
			},
			language:    Language("invalid"),
			expectError: true,
			errorMsg:    "unsupported language: invalid",
		},
		{
			name: "Invalid input value",
			inputs: map[InputField]interface{}{
				InputFieldGoVersion: "invalid",
			},
			language:    LanguageGo,
			expectError: true,
			errorMsg:    "invalid go version: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Config.ValidateAllInputs(tt.inputs, tt.language)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTypedDefaultsComprehensive(t *testing.T) {
	td := NewTypedDefaults()

	t.Run("GetSupportedLanguages", func(t *testing.T) {
		languages := td.GetSupportedLanguages()
		assert.Len(t, languages, 3)
		assert.Contains(t, languages, LanguageGo)
		assert.Contains(t, languages, LanguageNode)
		assert.Contains(t, languages, LanguagePython)
	})

	t.Run("GetAllVersions", func(t *testing.T) {
		versions := td.GetAllVersions()
		assert.Len(t, versions, 3)
		assert.Equal(t, []string{"1.21", "1.22", "1.23", "1.24"}, versions[LanguageGo])
		assert.Equal(t, []string{"16", "18", "20", "22"}, versions[LanguageNode])
		assert.Equal(t, []string{"3.9", "3.10", "3.11", "3.12"}, versions[LanguagePython])
	})

	t.Run("GetAllPackageManagers", func(t *testing.T) {
		managers := td.GetAllPackageManagers()
		assert.Len(t, managers, 2) // Go has no package managers
		assert.Equal(t, []PackageManager{PackageManagerNpm, PackageManagerYarn, PackageManagerPnpm}, managers[LanguageNode])
		assert.Equal(t, []PackageManager{PackageManagerPip, PackageManagerPoetry, PackageManagerPipenv}, managers[LanguagePython])
	})

	t.Run("GetDefaultSecuritySeverity", func(t *testing.T) {
		severity := td.GetDefaultSecuritySeverity()
		assert.Equal(t, SeverityCriticalHigh, severity)
	})

	t.Run("GetAllSecuritySeverities", func(t *testing.T) {
		severities := td.GetAllSecuritySeverities()
		assert.Len(t, severities, 6)
		assert.Contains(t, severities, SeverityCritical)
		assert.Contains(t, severities, SeverityCriticalHigh)
	})
}
