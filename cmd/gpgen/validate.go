package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/terrpan/gpgen/pkg/manifest"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [manifest-file]",
	Short: "Validate a GPGen manifest file",
	Long: `Validate a GPGen manifest file against the schema and check for errors.
If no file is specified, it will look for manifest.yaml in the current directory.`,
	RunE: runValidate,
}

var (
	validateStrict bool
	validateQuiet  bool
)

func init() {
	validateCmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "Use strict validation mode")
	validateCmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Only output errors, no success messages")
}

func runValidate(cmd *cobra.Command, args []string) error {
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

	if !validateQuiet {
		fmt.Printf("🔍 Validating manifest: %s\n", absPath)
	}

	// Load and validate the manifest
	m, err := manifest.LoadManifestFromFile(absPath)
	if err != nil {
		return fmt.Errorf("❌ Validation failed: %w", err)
	}

	// Apply strict validation if requested
	if validateStrict {
		if m.Metadata.Annotations == nil {
			m.Metadata.Annotations = make(map[string]string)
		}
		m.Metadata.Annotations["gpgen.dev/validation-mode"] = "strict"
	}

	// Validate the manifest
	if err := manifest.ValidateManifest(m); err != nil {
		return fmt.Errorf("❌ Validation failed: %w", err)
	}

	if !validateQuiet {
		fmt.Printf("✅ Manifest is valid\n")
		fmt.Printf("📋 Template: %s\n", m.Spec.Template)
		fmt.Printf("🏷️  Name: %s\n", m.Metadata.Name)

		// Show validation mode
		validationMode := "relaxed"
		if mode, ok := m.Metadata.Annotations["gpgen.dev/validation-mode"]; ok {
			validationMode = mode
		}
		if validateStrict {
			validationMode = "strict (forced)"
		}
		fmt.Printf("🔒 Validation mode: %s\n", validationMode)

		// Show environment info
		if len(m.Spec.Environments) > 0 {
			fmt.Printf("🌍 Environments: ")
			envs := make([]string, 0, len(m.Spec.Environments))
			for env := range m.Spec.Environments {
				envs = append(envs, env)
			}
			fmt.Printf("%v\n", envs)
		}

		// Show custom steps info
		if len(m.Spec.CustomSteps) > 0 {
			fmt.Printf("⚙️  Custom steps: %d\n", len(m.Spec.CustomSteps))
		}
	}

	return nil
}
