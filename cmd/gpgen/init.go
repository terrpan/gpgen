package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Initialize a new GPGen manifest",
	Long: `Initialize a new GPGen manifest file with a specified template.
This command creates a manifest.yaml file in the current directory with
sensible defaults for the chosen template.`,
	RunE: runInit,
}

var (
	initTemplate string
	initName     string
	initOutput   string
	initForce    bool
)

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "node-app", "Template to use (node-app, go-service, python-app)")
	initCmd.Flags().StringVarP(&initName, "name", "n", "", "Name for the pipeline (defaults to current directory name)")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "manifest.yaml", "Output file path")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing manifest file")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine the pipeline name
	if initName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		initName = filepath.Base(cwd)
		initName = strings.ReplaceAll(initName, " ", "-")
		initName = strings.ToLower(initName)
	}

	// Check if output file exists
	if !initForce {
		if _, err := os.Stat(initOutput); err == nil {
			return fmt.Errorf("manifest file %s already exists. Use --force to overwrite", initOutput)
		}
	}

	// Generate manifest content based on template
	manifestContent, err := generateManifestTemplate(initTemplate, initName)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}

	// Write manifest file
	if err := os.WriteFile(initOutput, []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	fmt.Printf("✅ Initialized %s manifest: %s\n", initTemplate, initOutput)
	fmt.Printf("📝 Edit the manifest to customize your pipeline\n")
	fmt.Printf("🚀 Run 'gpgen generate' to create your GitHub Actions workflow\n")

	return nil
}

func generateManifestTemplate(template, name string) (string, error) {
	switch template {
	case "node-app":
		return generateNodeAppManifest(name), nil
	case "go-service":
		return generateGoServiceManifest(name), nil
	case "python-app":
		return generatePythonAppManifest(name), nil
	default:
		return "", fmt.Errorf("unknown template: %s. Available templates: node-app, go-service, python-app", template)
	}
}

func generateNodeAppManifest(name string) string {
	return fmt.Sprintf(`apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: %s
  annotations:
    gpgen.dev/validation-mode: relaxed
    gpgen.dev/description: "Node.js application pipeline"
spec:
  template: node-app
  inputs:
    nodeVersion: "18"
    packageManager: npm
    testCommand: "npm test"
    buildCommand: "npm run build"

  # Add custom steps here
  customSteps: []

  # Environment-specific configurations
  environments:
    staging:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        testCommand: "npm run test:ci"

    production:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        nodeVersion: "20"
        testCommand: "npm run test:all"
`, name)
}

func generateGoServiceManifest(name string) string {
	return fmt.Sprintf(`apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: %s
  annotations:
    gpgen.dev/validation-mode: relaxed
    gpgen.dev/description: "Go service pipeline with security scanning"
spec:
  template: go-service
  inputs:
    goVersion: "1.21"
    testCommand: "go test ./..."
    buildCommand: "go build -o bin/%s ./cmd/%s"
    platforms: "linux/amd64,darwin/amd64"
    trivyScanEnabled: true
    trivySeverity: "CRITICAL,HIGH"

  # Add custom steps here
  customSteps: []

  # Environment-specific configurations
  environments:
    staging:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        testCommand: "go test -race ./..."
        trivySeverity: "CRITICAL,HIGH,MEDIUM"

    production:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        goVersion: "1.22"
        testCommand: "go test -race -cover ./..."
        trivySeverity: "CRITICAL"
`, name, name, name)
}

func generatePythonAppManifest(name string) string {
	return fmt.Sprintf(`apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: %s
  annotations:
    gpgen.dev/validation-mode: relaxed
    gpgen.dev/description: "Python application pipeline"
spec:
  template: python-app
  inputs:
    pythonVersion: "3.11"
    packageManager: pip
    testCommand: "pytest"
    lintCommand: "flake8"
    requirements: "requirements.txt"

  # Add custom steps here
  customSteps: []

  # Environment-specific configurations
  environments:
    staging:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        testCommand: "pytest --cov=. --cov-report=xml"

    production:
      annotations:
        gpgen.dev/validation-mode: strict
      inputs:
        pythonVersion: "3.12"
        testCommand: "pytest --cov=. --cov-report=xml --cov-fail-under=80"
`, name)
}
