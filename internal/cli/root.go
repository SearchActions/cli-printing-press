package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvanhorn/cli-printing-press/internal/generator"
	"github.com/mvanhorn/cli-printing-press/internal/openapi"
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
	var validate bool
	var refresh bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a Go CLI project from an API spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if specFile == "" {
				return fmt.Errorf("--spec is required")
			}

			var (
				data []byte
				err  error
			)
			if strings.HasPrefix(specFile, "http://") || strings.HasPrefix(specFile, "https://") {
				data, err = fetchOrCacheSpec(specFile, refresh)
				if err != nil {
					return fmt.Errorf("fetching spec from URL: %w", err)
				}
			} else {
				data, err = os.ReadFile(specFile)
				if err != nil {
					return fmt.Errorf("reading spec file: %w", err)
				}
			}

			var apiSpec *spec.APISpec
			if openapi.IsOpenAPI(data) {
				apiSpec, err = openapi.Parse(data)
			} else {
				apiSpec, err = spec.ParseBytes(data)
			}
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
			if validate {
				if err := gen.Validate(); err != nil {
					return fmt.Errorf("validating generated project: %w", err)
				}
			}

			fmt.Fprintf(os.Stderr, "Generated %s-cli at %s\n", apiSpec.Name, absOut)
			return nil
		},
	}

	cmd.Flags().StringVar(&specFile, "spec", "", "Path to API spec (internal YAML or OpenAPI 3.0+)")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: <name>-cli)")
	cmd.Flags().BoolVar(&validate, "validate", true, "Run quality gates on the generated project")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh cached remote spec before generating")

	return cmd
}

func fetchOrCacheSpec(specURL string, refresh bool) ([]byte, error) {
	sum := sha256.Sum256([]byte(specURL))
	cacheKey := hex.EncodeToString(sum[:])

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "printing-press", "specs")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	cachePath := filepath.Join(cacheDir, cacheKey+".json")
	if !refresh {
		info, err := os.Stat(cachePath)
		switch {
		case err == nil && time.Since(info.ModTime()) < 24*time.Hour:
			fmt.Fprintf(os.Stderr, "Using cached spec for %s\n", specURL)
			data, readErr := os.ReadFile(cachePath)
			if readErr != nil {
				return nil, fmt.Errorf("reading cached spec: %w", readErr)
			}
			return data, nil
		case err != nil && !os.IsNotExist(err):
			return nil, fmt.Errorf("checking cached spec: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Fetching spec from %s...\n", specURL)
	resp, err := http.Get(specURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing cached spec: %w", err)
	}

	return data, nil
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
