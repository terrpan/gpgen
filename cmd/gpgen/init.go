package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// generateManifest creates a manifest using common metadata and environment sections.
// description provides the metadata description annotation.
// baseInputs contains the default input values for the template. The map values
// should include any required quoting.
// envInputs provides environment specific input values keyed by environment name
// (e.g. "staging" or "production").
func generateManifest(name, tmplName, description string, baseInputs map[string]string, envInputs map[string]map[string]string) string {
	var b strings.Builder

	b.WriteString("apiVersion: gpgen.dev/v1\n")
	b.WriteString("kind: Pipeline\n")
	b.WriteString("metadata:\n")
	b.WriteString(fmt.Sprintf("  name: %s\n", name))
	b.WriteString("  annotations:\n")
	b.WriteString("    gpgen.dev/validation-mode: relaxed\n")
	b.WriteString(fmt.Sprintf("    gpgen.dev/description: \"%s\"\n", description))
	b.WriteString("spec:\n")
	b.WriteString(fmt.Sprintf("  template: %s\n", tmplName))
	b.WriteString("  inputs:\n")

	// Render inputs in sorted order for deterministic output
	keys := make([]string, 0, len(baseInputs))
	for k := range baseInputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("    %s: %s\n", k, baseInputs[k]))
	}

	b.WriteString("\n  # Add custom steps here\n  customSteps: []\n\n")

	b.WriteString("  # Environment-specific configurations\n  environments:\n")
	envOrder := []string{"staging", "production"}
	for _, env := range envOrder {
		b.WriteString(fmt.Sprintf("    %s:\n", env))
		b.WriteString("      annotations:\n")
		b.WriteString("        gpgen.dev/validation-mode: strict\n")
		inputs := envInputs[env]
		if len(inputs) == 0 {
			b.WriteString("      inputs: {}\n\n")
			continue
		}
		b.WriteString("      inputs:\n")
		keys := make([]string, 0, len(inputs))
		for k := range inputs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("        %s: %s\n", k, inputs[k]))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func generateNodeAppManifest(name string) string {
	baseInputs := map[string]string{
		"buildCommand":   "\"npm run build\"",
		"nodeVersion":    "\"18\"",
		"packageManager": "npm",
		"testCommand":    "\"npm test\"",
	}
	envInputs := map[string]map[string]string{
		"staging": {
			"testCommand": "\"npm run test:ci\"",
		},
		"production": {
			"nodeVersion": "\"20\"",
			"testCommand": "\"npm run test:all\"",
		},
	}
	return generateManifest(name, "node-app", "Node.js application pipeline", baseInputs, envInputs)
}

func generateGoServiceManifest(name string) string {
	baseInputs := map[string]string{
		"buildCommand":     fmt.Sprintf("\"go build -o bin/%s ./cmd/%s\"", name, name),
		"goVersion":        "\"1.21\"",
		"platforms":        "\"linux/amd64,darwin/amd64\"",
		"testCommand":      "\"go test ./...\"",
		"trivyScanEnabled": "true",
		"trivySeverity":    "\"CRITICAL,HIGH\"",
	}
	envInputs := map[string]map[string]string{
		"staging": {
			"testCommand":   "\"go test -race ./...\"",
			"trivySeverity": "\"CRITICAL,HIGH,MEDIUM\"",
		},
		"production": {
			"goVersion":     "\"1.22\"",
			"testCommand":   "\"go test -race -cover ./...\"",
			"trivySeverity": "\"CRITICAL\"",
		},
	}
	return generateManifest(name, "go-service", "Go service pipeline with security scanning", baseInputs, envInputs)
}

func generatePythonAppManifest(name string) string {
	baseInputs := map[string]string{
		"lintCommand":    "\"flake8\"",
		"packageManager": "pip",
		"pythonVersion":  "\"3.11\"",
		"requirements":   "\"requirements.txt\"",
		"testCommand":    "\"pytest\"",
	}
	envInputs := map[string]map[string]string{
		"staging": {
			"testCommand": "\"pytest --cov=. --cov-report=xml\"",
		},
		"production": {
			"pythonVersion": "\"3.12\"",
			"testCommand":   "\"pytest --cov=. --cov-report=xml --cov-fail-under=80\"",
		},
	}
	return generateManifest(name, "python-app", "Python application pipeline", baseInputs, envInputs)
}
