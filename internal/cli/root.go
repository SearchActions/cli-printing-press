package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvanhorn/cli-printing-press/internal/generator"
	"github.com/mvanhorn/cli-printing-press/internal/spec"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func Execute() error {
	rootCmd := &cobra.Command{
		Use:           "printing-press",
		Short:         "Describe your API. Get a production CLI.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	rootCmd.SetVersionTemplate("printing-press {{.Version}}\n")

	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd.Execute()
}

func newGenerateCmd() *cobra.Command {
	var specFile string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a Go CLI project from an API spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if specFile == "" {
				return fmt.Errorf("--spec is required")
			}

			apiSpec, err := spec.Parse(specFile)
			if err != nil {
				return fmt.Errorf("parsing spec: %w", err)
			}

			if outputDir == "" {
				outputDir = apiSpec.Name + "-cli"
			}

			absOut, err := filepath.Abs(outputDir)
			if err != nil {
				return fmt.Errorf("resolving output path: %w", err)
			}

			gen := generator.New(apiSpec, absOut)
			if err := gen.Generate(); err != nil {
				return fmt.Errorf("generating project: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Generated %s-cli at %s\n", apiSpec.Name, absOut)
			return nil
		},
	}

	cmd.Flags().StringVar(&specFile, "spec", "", "Path to API spec YAML file (required)")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: <name>-cli)")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("printing-press %s\n", version)
		},
	}
}
