package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		boolFlags     map[string]bool
		expectedError bool
		setupFunc     func(t *testing.T) string // Returns temp dir
		validateFunc  func(t *testing.T, tempDir string, err error)
	}{
		{
			name:          "generate with default manifest",
			args:          []string{},
			boolFlags:     map[string]bool{"overwrite": true},
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
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"
  environments:
    default: {}`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)

				// Check that workflow file was created
				workflowPath := filepath.Join(tempDir, ".github/workflows/test-project.yml")
				assert.FileExists(t, workflowPath)

				// Check content
				content, err := os.ReadFile(workflowPath)
				require.NoError(t, err)
				assert.Contains(t, string(content), "name: test-project")
			},
		},
		{
			name:          "generate with custom manifest path",
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
    goVersion: "1.21"
    testCommand: "go test ./..."
    buildCommand: "go build -o bin/service ./cmd/service"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)

				workflowPath := filepath.Join(tempDir, ".github/workflows/custom-project.yml")
				assert.FileExists(t, workflowPath)
			},
		},
		{
			name: "generate with custom output directory",
			args: []string{},
			flags: map[string]string{
				"output": "custom-workflows",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: custom-output
  description: Test custom output directory
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

				workflowPath := filepath.Join(tempDir, "custom-workflows/custom-output.yml")
				assert.FileExists(t, workflowPath)
			},
		},
		{
			name: "generate with specific environment",
			args: []string{},
			flags: map[string]string{
				"environment": "staging",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: env-test
  description: Test environment-specific generation
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"
  environments:
    staging:
      inputs:
        nodeVersion: "20"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)

				// Should generate staging-specific workflow
				workflowPath := filepath.Join(tempDir, ".github/workflows/env-test-staging.yml")
				assert.FileExists(t, workflowPath)
			},
		},
		{
			name: "generate with dry-run flag",
			args: []string{},
			boolFlags: map[string]bool{
				"dry-run": true,
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: dry-run-test
  description: Test dry run functionality
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

				// Should NOT create workflow file in dry-run mode
				workflowPath := filepath.Join(tempDir, ".github/workflows/dry-run-test.yml")
				assert.NoFileExists(t, workflowPath)
			},
		},
		{
			name: "generate with overwrite protection",
			args: []string{},
			boolFlags: map[string]bool{
				"overwrite": false,
			},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: overwrite-test
  description: Test overwrite protection
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)

				// Create existing workflow file
				workflowDir := filepath.Join(tempDir, ".github/workflows")
				err = os.MkdirAll(workflowDir, 0755)
				require.NoError(t, err)

				existingWorkflow := filepath.Join(workflowDir, "overwrite-test.yml")
				err = os.WriteFile(existingWorkflow, []byte("existing content"), 0644)
				require.NoError(t, err)

				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "already exists")
			},
		},
		{
			name: "generate with overwrite flag",
			args: []string{},
			boolFlags: map[string]bool{
				"overwrite": true,
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				validManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: overwrite-allowed
  description: Test overwrite with flag
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"`
				err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
				require.NoError(t, err)

				// Create existing workflow file
				workflowDir := filepath.Join(tempDir, ".github/workflows")
				err = os.MkdirAll(workflowDir, 0755)
				require.NoError(t, err)

				existingWorkflow := filepath.Join(workflowDir, "overwrite-allowed.yml")
				err = os.WriteFile(existingWorkflow, []byte("existing content"), 0644)
				require.NoError(t, err)

				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.NoError(t, err)

				// Should overwrite existing file
				workflowPath := filepath.Join(tempDir, ".github/workflows/overwrite-allowed.yml")
				assert.FileExists(t, workflowPath)

				content, err := os.ReadFile(workflowPath)
				require.NoError(t, err)
				assert.NotContains(t, string(content), "existing content")
			},
		},
		{
			name:          "generate missing manifest file",
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
			name:          "generate invalid manifest",
			args:          []string{},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				invalidManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: invalid-manifest
spec:
  # Missing required template field`
				err := os.WriteFile(manifestPath, []byte(invalidManifest), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string, err error) {
				assert.Error(t, err)
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
				Use:   "generate [manifest-file]",
				Short: "Generate GitHub Actions workflow from manifest",
				RunE:  runGenerate,
			}

			// Set flags
			cmd.Flags().StringVarP(&generateOutput, "output", "o", ".github/workflows", "Output directory")
			cmd.Flags().StringVarP(&generateEnv, "environment", "e", "", "Generate for specific environment")
			cmd.Flags().BoolVarP(&generateDryRun, "dry-run", "d", false, "Show what would be generated")
			cmd.Flags().BoolVarP(&generateOverwrite, "overwrite", "f", false, "Overwrite existing files")

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

			// Capture output to avoid cluttering test output
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			err = cmd.RunE(cmd, tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = originalStdout

			// Drain the pipe
			_, _ = io.ReadAll(r)

			// Run validation
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

func TestGenerateCmdFlagsAndHelp(t *testing.T) {
	// Test that all expected flags are present
	assert.NotNil(t, generateCmd.Flags().Lookup("output"))
	assert.NotNil(t, generateCmd.Flags().Lookup("environment"))
	assert.NotNil(t, generateCmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, generateCmd.Flags().Lookup("overwrite"))

	// Test flag shortcuts
	assert.NotNil(t, generateCmd.Flags().ShorthandLookup("o"))
	assert.NotNil(t, generateCmd.Flags().ShorthandLookup("e"))
	assert.NotNil(t, generateCmd.Flags().ShorthandLookup("d"))
	assert.NotNil(t, generateCmd.Flags().ShorthandLookup("f"))

	// Test command help text
	assert.Contains(t, generateCmd.Short, "Generate")
	assert.Contains(t, generateCmd.Long, "manifest.yaml")
}

func TestGenerateMultiEnvironment(t *testing.T) {
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

	// Create manifest with multiple environments
	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	multiEnvManifest := `apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: multi-env-test
  description: Test multiple environments
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: "npm"
    testCommand: "npm test"
  environments:
    staging:
      inputs:
        nodeVersion: "20"
    production:
      inputs:
        nodeVersion: "18"
        buildCommand: "npm run build:prod"`

	err = os.WriteFile(manifestPath, []byte(multiEnvManifest), 0644)
	require.NoError(t, err)

	// Create command and run without specific environment (should generate all)
	cmd := &cobra.Command{
		Use:  "generate [manifest-file]",
		RunE: runGenerate,
	}
	cmd.Flags().StringVarP(&generateOutput, "output", "o", ".github/workflows", "Output directory")
	cmd.Flags().StringVarP(&generateEnv, "environment", "e", "", "Generate for specific environment")
	cmd.Flags().BoolVarP(&generateDryRun, "dry-run", "d", false, "Show what would be generated")
	cmd.Flags().BoolVarP(&generateOverwrite, "overwrite", "f", false, "Overwrite existing files")

	// Capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.RunE(cmd, []string{})

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout
	_, _ = io.ReadAll(r)

	assert.NoError(t, err)

	// Should generate workflows for default, staging, and production
	defaultWorkflow := filepath.Join(tempDir, ".github/workflows/multi-env-test.yml")
	stagingWorkflow := filepath.Join(tempDir, ".github/workflows/multi-env-test-staging.yml")
	productionWorkflow := filepath.Join(tempDir, ".github/workflows/multi-env-test-production.yml")

	assert.FileExists(t, defaultWorkflow)
	assert.FileExists(t, stagingWorkflow)
	assert.FileExists(t, productionWorkflow)
}
