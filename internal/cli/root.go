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

	"github.com/mvanhorn/cli-printing-press/internal/docspec"
	"github.com/mvanhorn/cli-printing-press/internal/generator"
	"github.com/mvanhorn/cli-printing-press/internal/llm"
	"github.com/mvanhorn/cli-printing-press/internal/llmpolish"
	"github.com/mvanhorn/cli-printing-press/internal/openapi"
	"github.com/mvanhorn/cli-printing-press/internal/pipeline"
	"github.com/mvanhorn/cli-printing-press/internal/spec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	rootCmd.AddCommand(newScorecardCmd())
	rootCmd.AddCommand(newDogfoodCmd())
	rootCmd.AddCommand(newVisionCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newPrintCmd())

	return rootCmd.Execute()
}

func newGenerateCmd() *cobra.Command {
	var specFiles []string
	var cliName string
	var outputDir string
	var validate bool
	var refresh bool
	var force bool
	var lenient bool
	var docsURL string
	var polish bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a Go CLI project from an API spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if docsURL != "" {
				apiName := cliName
				if apiName == "" {
					apiName = "myapi"
				}

				var docSpec *spec.APISpec
				var err error

				if llm.Available() {
					fmt.Fprintln(os.Stderr, "Using LLM to understand API docs...")
					docSpec, err = docspec.GenerateFromDocsLLM(docsURL, apiName)
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: LLM doc-to-spec failed, falling back to regex: %v\n", err)
						docSpec, err = docspec.GenerateFromDocs(docsURL, apiName)
					}
				} else {
					docSpec, err = docspec.GenerateFromDocs(docsURL, apiName)
				}
				if err != nil {
					return fmt.Errorf("generating spec from docs: %w", err)
				}
				docYAML, err := yaml.Marshal(docSpec)
				if err != nil {
					return fmt.Errorf("marshaling doc spec: %w", err)
				}
				// Re-parse through the standard path so validation is consistent
				parsed, err := spec.ParseBytes(docYAML)
				if err != nil {
					return fmt.Errorf("parsing generated spec: %w", err)
				}

				if outputDir == "" {
					outputDir = parsed.Name + "-cli"
				}
				absOut, err := filepath.Abs(outputDir)
				if err != nil {
					return fmt.Errorf("resolving output path: %w", err)
				}
				if force {
					if err := os.RemoveAll(absOut); err != nil {
						return fmt.Errorf("removing existing output dir: %w", err)
					}
				}

				gen := generator.New(parsed, absOut)
				if err := gen.Generate(); err != nil {
					return fmt.Errorf("generating project: %w", err)
				}
				if validate {
					if err := gen.Validate(); err != nil {
						return fmt.Errorf("validating generated project: %w", err)
					}
				}

				if polish {
					fmt.Fprintln(os.Stderr, "Running LLM polish pass...")
					polishResult, polishErr := llmpolish.Polish(llmpolish.PolishRequest{
						APIName:   parsed.Name,
						OutputDir: absOut,
					})
					if polishErr != nil {
						fmt.Fprintf(os.Stderr, "warning: polish failed: %v\n", polishErr)
					} else if polishResult.Skipped {
						fmt.Fprintf(os.Stderr, "polish skipped: %s\n", polishResult.SkipReason)
					} else {
						fmt.Fprintf(os.Stderr, "Polish: %d help texts improved, %d examples added, README %v\n",
							polishResult.HelpTextsImproved, polishResult.ExamplesAdded, polishResult.READMERewritten)
					}
				}

				fmt.Fprintf(os.Stderr, "Generated %s-cli at %s (from docs)\n", parsed.Name, absOut)
				return nil
			}

			if len(specFiles) == 0 {
				return fmt.Errorf("--spec is required")
			}

			var specs []*spec.APISpec
			for _, specFile := range specFiles {
				data, err := readSpec(specFile, refresh)
				if err != nil {
					return fmt.Errorf("reading spec %s: %w", specFile, err)
				}

				var apiSpec *spec.APISpec
				if openapi.IsOpenAPI(data) {
					if lenient {
						apiSpec, err = openapi.ParseLenient(data)
					} else {
						apiSpec, err = openapi.Parse(data)
					}
				} else {
					apiSpec, err = spec.ParseBytes(data)
				}
				if err != nil {
					return fmt.Errorf("parsing spec %s: %w", specFile, err)
				}

				specs = append(specs, apiSpec)
			}

			var apiSpec *spec.APISpec
			if len(specs) == 1 {
				apiSpec = specs[0]
			} else {
				if cliName == "" {
					return fmt.Errorf("--name is required when using multiple specs")
				}
				apiSpec = mergeSpecs(specs, cliName)
			}

			if outputDir == "" {
				outputDir = apiSpec.Name + "-cli"
			}

			absOut, err := filepath.Abs(outputDir)
			if err != nil {
				return fmt.Errorf("resolving output path: %w", err)
			}
			if force {
				if err := os.RemoveAll(absOut); err != nil {
					return fmt.Errorf("removing existing output dir: %w", err)
				}
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

			if polish {
				fmt.Fprintln(os.Stderr, "Running LLM polish pass...")
				polishResult, polishErr := llmpolish.Polish(llmpolish.PolishRequest{
					APIName:   apiSpec.Name,
					OutputDir: absOut,
				})
				if polishErr != nil {
					fmt.Fprintf(os.Stderr, "warning: polish failed: %v\n", polishErr)
				} else if polishResult.Skipped {
					fmt.Fprintf(os.Stderr, "polish skipped: %s\n", polishResult.SkipReason)
				} else {
					fmt.Fprintf(os.Stderr, "Polish: %d help texts improved, %d examples added, README %v\n",
						polishResult.HelpTextsImproved, polishResult.ExamplesAdded, polishResult.READMERewritten)
				}
			}

			fmt.Fprintf(os.Stderr, "Generated %s-cli at %s\n", apiSpec.Name, absOut)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&specFiles, "spec", nil, "Path or URL to API spec (can be repeated)")
	cmd.Flags().StringVar(&cliName, "name", "", "CLI name (required when using multiple specs)")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: <name>-cli)")
	cmd.Flags().BoolVar(&validate, "validate", true, "Run quality gates on the generated project")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh cached remote spec before generating")
	cmd.Flags().BoolVar(&force, "force", false, "Remove existing output directory before generating")
	cmd.Flags().BoolVar(&lenient, "lenient", false, "Skip validation errors from broken $refs in OpenAPI specs")
	cmd.Flags().StringVar(&docsURL, "docs", "", "API documentation URL to generate spec from")
	cmd.Flags().BoolVar(&polish, "polish", false, "Run LLM polish pass on generated CLI (requires claude or codex CLI)")

	return cmd
}

func readSpec(specFile string, refresh bool) ([]byte, error) {
	if strings.HasPrefix(specFile, "http://") || strings.HasPrefix(specFile, "https://") {
		return fetchOrCacheSpec(specFile, refresh)
	}

	return os.ReadFile(specFile)
}

func mergeSpecs(specs []*spec.APISpec, name string) *spec.APISpec {
	if len(specs) == 1 {
		return specs[0]
	}

	merged := &spec.APISpec{
		Name:        name,
		Description: "Combined CLI for multiple API services",
		Version:     specs[0].Version,
		BaseURL:     specs[0].BaseURL,
		BasePath:    specs[0].BasePath,
		Auth:        specs[0].Auth,
		Config: spec.ConfigSpec{
			Format: "toml",
			Path:   fmt.Sprintf("~/.config/%s-cli/config.toml", name),
		},
		Resources: map[string]spec.Resource{},
		Types:     map[string]spec.TypeDef{},
	}

	for _, s := range specs {
		for resourceName, resource := range s.Resources {
			key := resourceName
			if _, exists := merged.Resources[key]; exists {
				key = s.Name + "-" + resourceName
			}
			merged.Resources[key] = resource
		}

		for typeName, typeDef := range s.Types {
			key := typeName
			if _, exists := merged.Types[key]; exists {
				key = s.Name + "-" + typeName
			}
			merged.Types[key] = typeDef
		}

		if s.Auth.AuthorizationURL != "" && merged.Auth.AuthorizationURL == "" {
			merged.Auth = s.Auth
		}
	}

	return merged
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

func newPrintCmd() *cobra.Command {
	var outputDir string
	var force bool
	var resume bool

	cmd := &cobra.Command{
		Use:   "print <api-name>",
		Short: "Create an autonomous CLI generation pipeline",
		Long:  "Creates a pipeline directory with plan seeds for each phase. Use /ce:work on each plan to execute.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiName := args[0]

			state, err := pipeline.Init(apiName, pipeline.Options{
				OutputDir: outputDir,
				Force:     force,
				Resume:    resume,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Pipeline created for %s\n", apiName)
			fmt.Fprintf(os.Stderr, "  Spec: %s\n", state.SpecURL)
			fmt.Fprintf(os.Stderr, "  Output: %s\n", state.OutputDir)
			fmt.Fprintf(os.Stderr, "  Plans:\n")
			for i, phase := range pipeline.PhaseOrder {
				fmt.Fprintf(os.Stderr, "    %d. %s\n", i, state.PlanPath(phase))
			}
			fmt.Fprintf(os.Stderr, "\nStart with: /ce:work %s\n", state.PlanPath(pipeline.PhasePreflight))

			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (default: ./<api-name>-cli)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing pipeline")
	cmd.Flags().BoolVar(&resume, "resume", false, "Resume from existing checkpoint")

	return cmd
}
