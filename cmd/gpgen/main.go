package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gpgen",
	Short: "Golden Path Pipeline Generator",
	Long: `GPGen is a tool for generating GitHub Action workflows based on
pre-defined templates and schemas. It enables teams to standardize their
CI/CD pipelines while allowing customization through user-defined manifest files.`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
}
