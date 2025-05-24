package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/terrpan/gpgen/pkg/generator"
	"github.com/terrpan/gpgen/pkg/manifest"
)

var generateCmd = &cobra.Command{
	Use:   "generate [manifest-file]",
	Short: "Generate GitHub Actions workflow from manifest",
	Long: `Generate GitHub Actions workflow files from a GPGen manifest.
If no file is specified, it will look for manifest.yaml in the current directory.`,
	RunE: runGenerate,
}

var (
	generateOutput    string
	generateEnv       string
	generateDryRun    bool
	generateOverwrite bool
)

func init() {
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", ".github/workflows", "Output directory for generated workflows")
	generateCmd.Flags().StringVarP(&generateEnv, "environment", "e", "", "Generate for specific environment (default: all environments)")
	generateCmd.Flags().BoolVarP(&generateDryRun, "dry-run", "d", false, "Show what would be generated without writing files")
	generateCmd.Flags().BoolVarP(&generateOverwrite, "overwrite", "f", false, "Overwrite existing workflow files")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Determine manifest file path
	manifestPath := "manifest.yaml"
	if len(args) > 0 {
		manifestPath = args[0]
	}

	// Check if file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest file not found: %s", manifestPath)
	}

	// Get absolute path for better error messages
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("📄 Loading manifest: %s\n", absPath)

	// Load and validate the manifest
	m, err := manifest.LoadManifestFromFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate the manifest
	if err := manifest.ValidateManifest(m); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}
	fmt.Printf("✅ Manifest loaded and validated\n")
	fmt.Printf("🏗️  Template: %s\n", m.Spec.Template)

	// Create workflow generator
	gen := generator.NewWorkflowGenerator("")

	// Determine which environments to generate
	environments := []string{"default"}
	if generateEnv != "" {
		environments = []string{generateEnv}
	} else if len(m.Spec.Environments) > 0 {
		environments = []string{"default"}
		for env := range m.Spec.Environments {
			environments = append(environments, env)
		}
	}

	// Create output directory if it doesn't exist
	if !generateDryRun {
		if err := os.MkdirAll(generateOutput, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	for _, env := range environments {
		workflowName := fmt.Sprintf("%s.yml", m.Metadata.Name)
		if env != "default" {
			workflowName = fmt.Sprintf("%s-%s.yml", m.Metadata.Name, env)
		}

		outputPath := filepath.Join(generateOutput, workflowName)

		if generateDryRun {
			fmt.Printf("📝 Would generate: %s\n", outputPath)
			fmt.Printf("   Environment: %s\n", env)
			if env != "default" {
				if _, exists := m.Spec.Environments[env]; exists {
					fmt.Printf("   Environment-specific config: yes\n")
				}
			}
			fmt.Printf("   Custom steps: %d\n", len(m.Spec.CustomSteps))
			fmt.Printf("\n")
		} else {
			// Generate the workflow
			fmt.Printf("🔨 Generating workflow for environment: %s\n", env)

			workflowContent, err := gen.GenerateWorkflow(m, env)
			if err != nil {
				return fmt.Errorf("failed to generate workflow for %s: %w", env, err)
			}

			// Check if file exists and handle overwrite
			if _, err := os.Stat(outputPath); err == nil && !generateOverwrite {
				return fmt.Errorf("workflow file %s already exists. Use --overwrite to replace it", outputPath)
			}

			// Write workflow file
			if err := os.WriteFile(outputPath, []byte(workflowContent), 0644); err != nil {
				return fmt.Errorf("failed to write workflow file %s: %w", outputPath, err)
			}

			fmt.Printf("✅ Generated: %s\n", outputPath)
		}
	}

	if generateDryRun {
		fmt.Printf("💡 Run without --dry-run to generate the actual workflow files\n")
	} else {
		fmt.Printf("\n🎉 Successfully generated %d workflow file(s)\n", len(environments))
		fmt.Printf("📁 Output directory: %s\n", generateOutput)
		fmt.Printf("🚀 Commit and push to trigger your workflows!\n")
	}

	return nil
}
