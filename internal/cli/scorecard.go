package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mvanhorn/cli-printing-press/internal/pipeline"
	"github.com/spf13/cobra"
)

func newScorecardCmd() *cobra.Command {
	var dir string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "scorecard",
		Short: "Score a generated CLI against the Steinberger bar (10 dimensions, max 100)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dir == "" {
				return fmt.Errorf("--dir is required")
			}

			// Use a temp pipeline dir for the scorecard output
			pipelineDir, err := os.MkdirTemp("", "scorecard-*")
			if err != nil {
				return fmt.Errorf("creating temp dir: %w", err)
			}
			defer os.RemoveAll(pipelineDir)

			sc, err := pipeline.RunScorecard(dir, pipelineDir)
			if err != nil {
				return fmt.Errorf("running scorecard: %w", err)
			}

			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sc)
			}

			// Human-readable output
			s := sc.Steinberger
			fmt.Printf("Steinberger Scorecard: %s\n\n", sc.APIName)
			fmt.Printf("  Output Modes   %d/10\n", s.OutputModes)
			fmt.Printf("  Auth           %d/10\n", s.Auth)
			fmt.Printf("  Error Handling %d/10\n", s.ErrorHandling)
			fmt.Printf("  Terminal UX    %d/10\n", s.TerminalUX)
			fmt.Printf("  README         %d/10\n", s.README)
			fmt.Printf("  Doctor         %d/10\n", s.Doctor)
			fmt.Printf("  Agent Native   %d/10\n", s.AgentNative)
			fmt.Printf("  Local Cache    %d/10\n", s.LocalCache)
			fmt.Printf("  Breadth        %d/10\n", s.Breadth)
			fmt.Printf("  Vision         %d/10\n", s.Vision)
			fmt.Printf("  Workflows      %d/10\n", s.Workflows)
			fmt.Printf("\n  Total: %d/110 (%d%%) - Grade %s\n", s.Total, s.Percentage, sc.OverallGrade)

			if len(sc.GapReport) > 0 {
				fmt.Printf("\nGaps:\n")
				for _, g := range sc.GapReport {
					fmt.Printf("  - %s\n", g)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "Path to generated CLI directory")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")

	return cmd
}
