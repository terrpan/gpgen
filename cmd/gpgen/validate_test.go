package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		boolFlags     map[string]bool
		expectedError bool
		setupFunc     func(t *testing.T) string // Returns temp dir and manifest path
		validateFunc  func(t *testing.T, tempDir string, err error)
	}{
		{
			name:          "validate valid manifest",
			args:          []string{},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: test-project
  description: A test project
spec:
  template: node-app
  inputs:
    node_version: "18"
  steps:
    - name: Run tests
      position: after
      target: install-deps`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:          "validate with custom manifest path",
			args:          []string{"custom-manifest.yaml"},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "custom-manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: custom-project
  description: A custom project
spec:
  template: go-service
  inputs:
    go_version: "1.21"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:          "validate with quiet flag",
			args:          []string{},
			boolFlags:     map[string]bool{"quiet": true},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: quiet-test
  description: A test with quiet flag
spec:
  template: go-service
  inputs:
    goVersion: "1.21"
    testCommand: "go test ./..."
    buildCommand: "go build"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:          "validate with strict flag",
			args:          []string{},
			boolFlags:     map[string]bool{"strict": true},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: strict-test
  description: A test with strict validation
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:          "validate missing manifest file",
			args:          []string{},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				return t.TempDir() // No manifest file created
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "manifest file not found")
			},
		},
		{
			name:          "validate invalid YAML syntax",
			args:          []string{},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				invalidManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: invalid-yaml
  description: [unclosed bracket
spec:
  template: node-app`
				err := os.WriteFile(manifestPath, []byte(invalidManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:          "validate missing required fields",
			args:          []string{},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				invalidManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: missing-template
spec:
  inputs:
    some_value: "test"`
				err := os.WriteFile(manifestPath, []byte(invalidManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:          "validate custom manifest path not found",
			args:          []string{"nonexistent.yaml"},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "manifest file not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directory
			tempDir := tt.setupFunc(t)

			// Change to temp directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(originalDir)
				require.NoError(t, err)
			}()

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Create a fresh command instance
			cmd := &cobra.Command{
				Use:   "validate [manifest-file]",
				Short: "Validate a GPGen manifest file",
				RunE:  runValidate,
			}

			// Set flags
			cmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "Use strict validation mode")
			cmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Only output errors")

			// Apply flag values
			for flag, value := range tt.flags {
				err := cmd.Flags().Set(flag, value)
				require.NoError(t, err)
			}

			for flag, value := range tt.boolFlags {
				if value {
					err := cmd.Flags().Set(flag, "true")
					require.NoError(t, err)
				}
			}

			// Execute command
			err = cmd.RunE(cmd, tt.args)

			// Check error expectation and run validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, tempDir, err)
			} else if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCmdFlagsAndHelp(t *testing.T) {
	// Test that all expected flags are present
	assert.NotNil(t, validateCmd.Flags().Lookup("strict"))
	assert.NotNil(t, validateCmd.Flags().Lookup("quiet"))

	// Test flag shortcuts
	assert.NotNil(t, validateCmd.Flags().ShorthandLookup("s"))
	assert.NotNil(t, validateCmd.Flags().ShorthandLookup("q"))

	// Test command help text
	assert.Contains(t, validateCmd.Short, "Validate")
	assert.Contains(t, validateCmd.Long, "manifest.yaml")
}

func TestValidateManifestFileDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Test with no arguments (should look for manifest.yaml)
	validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: default-detection
  description: Test default file detection
spec:
  template: node-app`

	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	err = os.WriteFile(manifestPath, []byte(validManifest), 0644)
	require.NoError(t, err)

	// Create command and run
	cmd := &cobra.Command{
		Use:  "validate [manifest-file]",
		RunE: runValidate,
	}
	cmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "Use strict validation mode")
	cmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Only output errors")

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}
