package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		boolFlags      map[string]bool
		expectedError  bool
		expectedOutput string
		setupFunc      func(t *testing.T) string // Returns temp dir
		validateFunc   func(t *testing.T, tempDir string)
	}{
		{
			name: "init with default template",
			args: []string{},
			flags: map[string]string{
				"template": "node-app",
				"name":     "test-project",
				"output":   "manifest.yaml",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string) {
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				assert.FileExists(t, manifestPath)

				content, err := os.ReadFile(manifestPath)
				require.NoError(t, err)

				// Check that the content contains expected template content
				assert.Contains(t, string(content), "name: test-project")
				assert.Contains(t, string(content), "template: node-app")
			},
		},
		{
			name: "init with go-service template",
			args: []string{},
			flags: map[string]string{
				"template": "go-service",
				"name":     "my-service",
				"output":   "manifest.yaml",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string) {
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				assert.FileExists(t, manifestPath)

				content, err := os.ReadFile(manifestPath)
				require.NoError(t, err)

				assert.Contains(t, string(content), "name: my-service")
				assert.Contains(t, string(content), "template: go-service")
			},
		},
		{
			name: "init with python-app template",
			args: []string{},
			flags: map[string]string{
				"template": "python-app",
				"name":     "python-project",
				"output":   "manifest.yaml",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string) {
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				assert.FileExists(t, manifestPath)

				content, err := os.ReadFile(manifestPath)
				require.NoError(t, err)

				assert.Contains(t, string(content), "name: python-project")
				assert.Contains(t, string(content), "template: python-app")
			},
		},
		{
			name: "init with custom output path",
			args: []string{},
			flags: map[string]string{
				"template": "node-app",
				"name":     "test-project",
				"output":   "custom-manifest.yaml",
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string) {
				manifestPath := filepath.Join(tempDir, "custom-manifest.yaml")
				assert.FileExists(t, manifestPath)
			},
		},
		{
			name: "init fails when file exists without force",
			args: []string{},
			flags: map[string]string{
				"template": "node-app",
				"name":     "test-project",
				"output":   "manifest.yaml",
			},
			boolFlags: map[string]bool{
				"force": false,
			},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create existing file
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				err := os.WriteFile(manifestPath, []byte("existing content"), 0644)
				require.NoError(t, err)
				return tempDir
			},
		},
		{
			name: "init succeeds when file exists with force",
			args: []string{},
			flags: map[string]string{
				"template": "node-app",
				"name":     "test-project",
				"output":   "manifest.yaml",
			},
			boolFlags: map[string]bool{
				"force": true,
			},
			expectedError: false,
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create existing file
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				err := os.WriteFile(manifestPath, []byte("existing content"), 0644)
				require.NoError(t, err)
				return tempDir
			},
			validateFunc: func(t *testing.T, tempDir string) {
				manifestPath := filepath.Join(tempDir, "manifest.yaml")
				assert.FileExists(t, manifestPath)

				content, err := os.ReadFile(manifestPath)
				require.NoError(t, err)

				// Should contain new content, not "existing content"
				assert.Contains(t, string(content), "name: test-project")
				assert.NotContains(t, string(content), "existing content")
			},
		},
		{
			name: "init with invalid template",
			args: []string{},
			flags: map[string]string{
				"template": "invalid-template",
				"name":     "test-project",
				"output":   "manifest.yaml",
			},
			expectedError: true,
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
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
				Use:   "init [flags]",
				Short: "Initialize a new GPGen manifest",
				RunE:  runInit,
			}

			// Set flags
			cmd.Flags().StringVarP(&initTemplate, "template", "t", "node-app", "Template to use")
			cmd.Flags().StringVarP(&initName, "name", "n", "", "Name for the pipeline")
			cmd.Flags().StringVarP(&initOutput, "output", "o", "manifest.yaml", "Output file path")
			cmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing manifest file")

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

			// Check error expectation
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Run validation if provided
				if tt.validateFunc != nil {
					tt.validateFunc(t, tempDir)
				}
			}
		})
	}
}

func TestInitCmdFlagsAndHelp(t *testing.T) {
	// Test that all expected flags are present
	assert.NotNil(t, initCmd.Flags().Lookup("template"))
	assert.NotNil(t, initCmd.Flags().Lookup("name"))
	assert.NotNil(t, initCmd.Flags().Lookup("output"))
	assert.NotNil(t, initCmd.Flags().Lookup("force"))

	// Test flag shortcuts
	assert.NotNil(t, initCmd.Flags().ShorthandLookup("t"))
	assert.NotNil(t, initCmd.Flags().ShorthandLookup("n"))
	assert.NotNil(t, initCmd.Flags().ShorthandLookup("o"))
	assert.NotNil(t, initCmd.Flags().ShorthandLookup("f"))

	// Test command help text
	assert.Contains(t, initCmd.Short, "Initialize")
	assert.Contains(t, initCmd.Long, "manifest.yaml")
}
