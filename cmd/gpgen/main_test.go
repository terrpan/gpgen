package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	// Test root command structure
	assert.Equal(t, "gpgen", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Golden Path Pipeline Generator")
	assert.Contains(t, rootCmd.Long, "GitHub Action workflows")

	// Test that all subcommands are added
	commands := rootCmd.Commands()
	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name()
	}

	assert.Contains(t, commandNames, "init")
	assert.Contains(t, commandNames, "generate")
	assert.Contains(t, commandNames, "validate")
}

func TestRootCommandHelp(t *testing.T) {
	// Test help output
	buf := new(bytes.Buffer)

	// Reset command state before test
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "GPGen is a tool for generating GitHub Action workflows")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "init")
	assert.Contains(t, output, "generate")
	assert.Contains(t, output, "validate")
}

func TestRootCommandVersion(t *testing.T) {
	// Test version output
	buf := new(bytes.Buffer)

	// Create a fresh command instance to avoid global state interference
	testRootCmd := &cobra.Command{
		Use:   "gpgen",
		Short: "Golden Path Pipeline Generator",
		Long: `GPGen is a tool for generating GitHub Action workflows based on
pre-defined templates and schemas. It enables teams to standardize their
CI/CD pipelines while allowing customization through user-defined manifest files.`,
		Version: version,
	}

	testRootCmd.SetOut(buf)
	testRootCmd.SetErr(buf)
	testRootCmd.SetArgs([]string{"-v"})

	err := testRootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "gpgen version")
}

func TestSubcommandHelp(t *testing.T) {
	tests := []struct {
		name    string
		command string
		expects []string
	}{
		{
			name:    "init help",
			command: "init",
			expects: []string{"Initialize", "manifest", "template"},
		},
		{
			name:    "generate help",
			command: "generate",
			expects: []string{"Generate", "workflow", "manifest"},
		},
		{
			name:    "validate help",
			command: "validate",
			expects: []string{"Validate", "manifest", "schema"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			// Reset command state before test
			rootCmd.SetArgs(nil)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{tt.command, "--help"})

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expects {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCommandErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "unknown command",
			args:        []string{"unknown"},
			expectError: true,
		},
		{
			name:        "invalid flag",
			args:        []string{"init", "--invalid-flag"},
			expectError: true,
		},
		{
			name:        "valid command with help",
			args:        []string{"init", "--help"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of rootCmd to avoid state pollution
			cmd := &cobra.Command{
				Use:     rootCmd.Use,
				Short:   rootCmd.Short,
				Long:    rootCmd.Long,
				Version: rootCmd.Version,
			}

			// Add subcommands
			cmd.AddCommand(&cobra.Command{
				Use:   "init [flags]",
				Short: "Initialize a new GPGen manifest",
				RunE:  runInit,
			})
			cmd.AddCommand(&cobra.Command{
				Use:   "generate [manifest-file]",
				Short: "Generate GitHub Actions workflow from manifest",
				RunE:  runGenerate,
			})
			cmd.AddCommand(&cobra.Command{
				Use:   "validate [manifest-file]",
				Short: "Validate a GPGen manifest file",
				RunE:  runValidate,
			})

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main function doesn't panic
	// We can't easily test the actual main() function since it calls os.Exit()
	// But we can test that rootCmd.Execute() works

	// Save original args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Test with version flag
	os.Args = []string{"gpgen", "--version"}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestCommandIntegration(t *testing.T) {
	// Test that commands work together in a typical workflow
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

	// Step 1: Initialize a manifest
	initCmd := &cobra.Command{
		Use:  "init",
		RunE: runInit,
	}
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "node-app", "Template to use")
	initCmd.Flags().StringVarP(&initName, "name", "n", "", "Name for the pipeline")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "manifest.yaml", "Output file path")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing manifest file")

	err = initCmd.Flags().Set("name", "integration-test")
	require.NoError(t, err)

	err = initCmd.RunE(initCmd, []string{})
	assert.NoError(t, err)

	// Step 2: Validate the generated manifest
	validateCmd := &cobra.Command{
		Use:  "validate",
		RunE: runValidate,
	}
	validateCmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "Use strict validation mode")
	validateCmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Only output errors")

	err = validateCmd.RunE(validateCmd, []string{})
	assert.NoError(t, err)

	// Step 3: Generate workflow from manifest
	generateCmd := &cobra.Command{
		Use:  "generate",
		RunE: runGenerate,
	}
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", ".github/workflows", "Output directory")
	generateCmd.Flags().StringVarP(&generateEnv, "environment", "e", "", "Generate for specific environment")
	generateCmd.Flags().BoolVarP(&generateDryRun, "dry-run", "d", false, "Show what would be generated")
	generateCmd.Flags().BoolVarP(&generateOverwrite, "overwrite", "f", false, "Overwrite existing files")

	// Capture output to avoid cluttering test output
	buf := new(bytes.Buffer)
	generateCmd.SetOut(buf)

	err = generateCmd.RunE(generateCmd, []string{})
	assert.NoError(t, err)

	// Verify workflow was created
	workflowPath := ".github/workflows/integration-test.yml"
	assert.FileExists(t, workflowPath)
}
