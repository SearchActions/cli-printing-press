package pipeline

import (
	"fmt"
	"os"
	"testing"
)

func TestScorecardOnRealCLI(t *testing.T) {
	outputDir := os.Getenv("SCORECARD_CLI_DIR")
	if outputDir == "" {
		t.Skip("Set SCORECARD_CLI_DIR to run scorecard on a real CLI")
	}
	pipelineDir := os.Getenv("SCORECARD_PIPELINE_DIR")
	if pipelineDir == "" {
		pipelineDir = t.TempDir()
	}

	sc, err := RunScorecard(outputDir, pipelineDir, "")
	if err != nil {
		t.Fatalf("RunScorecard: %v", err)
	}

	fmt.Printf("\n=== STEINBERGER SCORECARD ===\n")
	fmt.Printf("Score: %d/80 (%d%%)\n", sc.Steinberger.Total, sc.Steinberger.Percentage)
	fmt.Printf("Grade: %s\n\n", sc.OverallGrade)
	fmt.Printf("  Output Modes:   %d/10\n", sc.Steinberger.OutputModes)
	fmt.Printf("  Auth:           %d/10\n", sc.Steinberger.Auth)
	fmt.Printf("  Error Handling: %d/10\n", sc.Steinberger.ErrorHandling)
	fmt.Printf("  Terminal UX:    %d/10\n", sc.Steinberger.TerminalUX)
	fmt.Printf("  README:         %d/10\n", sc.Steinberger.README)
	fmt.Printf("  Doctor:         %d/10\n", sc.Steinberger.Doctor)
	fmt.Printf("  Agent Native:   %d/10\n", sc.Steinberger.AgentNative)
	fmt.Printf("  Local Cache:    %d/10\n", sc.Steinberger.LocalCache)
	if len(sc.GapReport) > 0 {
		fmt.Println("\nGaps:")
		for _, g := range sc.GapReport {
			fmt.Printf("  - %s\n", g)
		}
	}
	fmt.Println()
}
