package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvanhorn/cli-printing-press/internal/pipeline"
	"github.com/spf13/cobra"
)

// EmbossReport captures the before/after delta from an emboss cycle.
type EmbossReport struct {
	Dir          string          `json:"dir"`
	Spec         string          `json:"spec"`
	Timestamp    string          `json:"timestamp"`
	Before       EmbossSnapshot  `json:"before"`
	After        *EmbossSnapshot `json:"after,omitempty"`
	Delta        *EmbossDelta    `json:"delta,omitempty"`
	Improvements []string        `json:"improvements,omitempty"`
	Mode         string          `json:"mode"` // "audit-only" or "full"
}

type EmbossSnapshot struct {
	ScorecardTotal int     `json:"scorecard_total"`
	ScorecardGrade string  `json:"scorecard_grade"`
	VerifyPassRate float64 `json:"verify_pass_rate"`
	VerifyPassed   int     `json:"verify_passed"`
	VerifyTotal    int     `json:"verify_total"`
	DataPipeline   bool    `json:"data_pipeline"`
	CommandCount   int     `json:"command_count"`
}

type EmbossDelta struct {
	ScorecardDelta int     `json:"scorecard_delta"`
	VerifyDelta    float64 `json:"verify_delta"`
	CommandDelta   int     `json:"command_delta"`
	PipelineFixed  bool    `json:"pipeline_fixed"`
}

func newEmbossCmd() *cobra.Command {
	var dir string
	var specPath string
	var apiKey string
	var envVar string
	var asJSON bool
	var auditOnly bool
	var saveBaseline bool
	var keepBaseline bool

	cmd := &cobra.Command{
		Use:   "emboss",
		Short: "Second-pass improvement cycle for an existing generated CLI",
		Long: `Run the emboss cycle on an already-generated CLI to make it better.

Step 1: AUDIT - Run verify + scorecard to get a baseline
Step 2: RE-RESEARCH - (skill-driven) Find what's new since v1
Step 3: GAP ANALYSIS - (skill-driven) Identify top 5 improvements
Step 4: IMPROVE - (skill-driven) Build improvements, commit atomically
Step 5: RE-VERIFY - Run verify + scorecard again, compare to baseline
Step 6: REPORT - Output the delta

Use --audit-only to just get the baseline without making changes.
The improvement steps (2-4) are driven by the /printing-press emboss skill.`,
		Example: `  # Full emboss cycle (audit -> improve -> re-verify)
  # Run the skill: /printing-press emboss ./discord-cli
  # Or just get the baseline:
  printing-press emboss --dir ./discord-cli --spec /tmp/spec.json --audit-only

  # Audit with live API testing
  printing-press emboss --dir ./discord-cli --spec /tmp/spec.json --api-key $TOKEN --audit-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			baselinePath := filepath.Join(dir, ".emboss-baseline.json")
			name := filepath.Base(dir)
			report := &EmbossReport{
				Dir:       dir,
				Spec:      specPath,
				Timestamp: time.Now().Format(time.RFC3339),
			}

			if auditOnly {
				report.Mode = "audit-only"
			} else {
				report.Mode = "full"
			}

			if data, err := os.ReadFile(baselinePath); err == nil {
				var baselineReport EmbossReport
				if err := json.Unmarshal(data, &baselineReport); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to load baseline from %s: %v\n", baselinePath, err)
				} else {
					report.Before = baselineReport.Before
					fmt.Fprintln(os.Stderr, "Existing baseline found. Running fresh audit for delta...")
					after := runEmbossAudit(dir, specPath, apiKey, envVar, "after")
					report.After = &after
					report.Delta = computeDelta(report.Before, after)
					report.Mode = "delta"
					reportPath, writeErr := writeEmbossDeltaReport(name, report.Before, after, report.Delta)
					if writeErr != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to write delta report: %v\n", writeErr)
					} else {
						fmt.Fprintf(os.Stderr, "Delta report written: %s\n", reportPath)
					}
					if !keepBaseline {
						if err := os.Remove(baselinePath); err != nil && !os.IsNotExist(err) {
							fmt.Fprintf(os.Stderr, "warning: failed to remove baseline %s: %v\n", baselinePath, err)
						}
					}
					return printEmbossReport(cmd, report, asJSON)
				}
			}

			// Step 1: AUDIT - baseline
			report.Before = runEmbossAudit(dir, specPath, apiKey, envVar, "baseline")

			if auditOnly {
				if saveBaseline {
					if err := saveEmbossBaseline(baselinePath, report); err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to save baseline to %s: %v\n", baselinePath, err)
					}
				}
				return printEmbossReport(cmd, report, asJSON)
			}

			if saveBaseline || report.Mode == "full" {
				if err := saveEmbossBaseline(baselinePath, report); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to save baseline to %s: %v\n", baselinePath, err)
				}
			}

			if err := printEmbossReport(cmd, report, asJSON); err != nil {
				return err
			}

			fmt.Fprintln(os.Stderr, "\nBaseline saved. Now run the skill for improvements:")
			fmt.Fprintf(os.Stderr, "  /printing-press emboss %s\n\n", dir)
			fmt.Fprintln(os.Stderr, "When done, re-run this command to compute the delta:")
			fmt.Fprintf(os.Stderr, "  printing-press emboss --dir %s --spec %s\n", dir, specPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "Path to the generated CLI directory (required)")
	cmd.Flags().StringVar(&specPath, "spec", "", "Path to the OpenAPI spec file")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key for live testing (read-only GETs only)")
	cmd.Flags().StringVar(&envVar, "env-var", "", "Environment variable name for the API key")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&auditOnly, "audit-only", false, "Only run the baseline audit, no improvements")
	cmd.Flags().BoolVar(&saveBaseline, "save-baseline", false, "Save the baseline report to disk for a future delta run")
	cmd.Flags().BoolVar(&keepBaseline, "keep-baseline", false, "Keep the saved baseline after computing a delta")
	_ = cmd.MarkFlagRequired("dir")
	return cmd
}

func runEmbossAudit(dir, specPath, apiKey, envVar, label string) EmbossSnapshot {
	fmt.Fprintf(os.Stderr, "Step 1: AUDIT - Running verify + scorecard for %s...\n", label)

	verifyCfg := pipeline.VerifyConfig{
		Dir:       dir,
		SpecPath:  specPath,
		APIKey:    apiKey,
		EnvVar:    envVar,
		Threshold: 80,
	}
	verifyReport, err := pipeline.RunVerify(verifyCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  verify error: %v (continuing with partial %s)\n", err, label)
	}

	scorecardReport, err := pipeline.RunScorecard(dir, "", specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  scorecard error: %v (continuing with partial %s)\n", err, label)
	}

	snapshot := EmbossSnapshot{}
	if verifyReport != nil {
		snapshot.VerifyPassRate = verifyReport.PassRate
		snapshot.VerifyPassed = verifyReport.Passed
		snapshot.VerifyTotal = verifyReport.Total
		snapshot.DataPipeline = verifyReport.DataPipeline
		snapshot.CommandCount = verifyReport.Total
	}
	if scorecardReport != nil {
		snapshot.ScorecardTotal = scorecardReport.Steinberger.Total
		snapshot.ScorecardGrade = scorecardGrade(scorecardReport.Steinberger.Total)
	}
	return snapshot
}

func saveEmbossBaseline(path string, report *EmbossReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func computeDelta(before, after EmbossSnapshot) *EmbossDelta {
	return &EmbossDelta{
		ScorecardDelta: after.ScorecardTotal - before.ScorecardTotal,
		VerifyDelta:    after.VerifyPassRate - before.VerifyPassRate,
		CommandDelta:   after.CommandCount - before.CommandCount,
		PipelineFixed:  !before.DataPipeline && after.DataPipeline,
	}
}

func writeEmbossDeltaReport(name string, before, after EmbossSnapshot, delta *EmbossDelta) (string, error) {
	reportDir := filepath.Join("docs", "plans")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s-emboss-%s-delta.md", time.Now().Format("2006-01-02"), name)
	path := filepath.Join(reportDir, filename)
	content := strings.Join([]string{
		fmt.Sprintf("# Emboss Delta Report: %s", name),
		"",
		"| Metric | Before | After | Delta |",
		"| --- | --- | --- | --- |",
		fmt.Sprintf("| Scorecard | %d (%s) | %d (%s) | %+d |", before.ScorecardTotal, before.ScorecardGrade, after.ScorecardTotal, after.ScorecardGrade, delta.ScorecardDelta),
		fmt.Sprintf("| Verify | %.0f%% (%d/%d) | %.0f%% (%d/%d) | %+.0f%% |", before.VerifyPassRate, before.VerifyPassed, before.VerifyTotal, after.VerifyPassRate, after.VerifyPassed, after.VerifyTotal, delta.VerifyDelta),
		fmt.Sprintf("| Pipeline | %s | %s | %s |", boolStatus(before.DataPipeline), boolStatus(after.DataPipeline), pipelineDeltaStatus(before.DataPipeline, after.DataPipeline, delta.PipelineFixed)),
		fmt.Sprintf("| Commands | %d | %d | %+d |", before.CommandCount, after.CommandCount, delta.CommandDelta),
		"",
	}, "\n")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func pipelineDeltaStatus(before, after, fixed bool) string {
	if fixed {
		return "FIXED"
	}
	if before == after {
		return "UNCHANGED"
	}
	return "REGRESSED"
}

func printEmbossReport(cmd *cobra.Command, report *EmbossReport, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	name := filepath.Base(report.Dir)
	fmt.Printf("EMBOSS REPORT: %s\n", name)
	fmt.Println("==============================")
	fmt.Printf("Mode: %s\n", report.Mode)
	fmt.Printf("Timestamp: %s\n\n", report.Timestamp)

	fmt.Println("BASELINE:")
	fmt.Printf("  Scorecard:  %d/100 (Grade %s)\n", report.Before.ScorecardTotal, report.Before.ScorecardGrade)
	fmt.Printf("  Verify:     %.0f%% (%d/%d passed)\n", report.Before.VerifyPassRate, report.Before.VerifyPassed, report.Before.VerifyTotal)
	fmt.Printf("  Pipeline:   %s\n", boolStatus(report.Before.DataPipeline))
	fmt.Printf("  Commands:   %d\n", report.Before.CommandCount)

	if report.After != nil {
		fmt.Println("\nAFTER:")
		fmt.Printf("  Scorecard:  %d/100 (Grade %s)\n", report.After.ScorecardTotal, report.After.ScorecardGrade)
		fmt.Printf("  Verify:     %.0f%% (%d/%d passed)\n", report.After.VerifyPassRate, report.After.VerifyPassed, report.After.VerifyTotal)
		fmt.Printf("  Pipeline:   %s\n", boolStatus(report.After.DataPipeline))
		fmt.Printf("  Commands:   %d\n", report.After.CommandCount)
	}

	if report.Delta != nil {
		fmt.Println("\nDELTA:")
		fmt.Printf("  Scorecard:  %+d\n", report.Delta.ScorecardDelta)
		fmt.Printf("  Verify:     %+.0f%%\n", report.Delta.VerifyDelta)
		fmt.Printf("  Commands:   %+d\n", report.Delta.CommandDelta)
		if report.Delta.PipelineFixed {
			fmt.Println("  Pipeline:   FIXED")
		}
	}

	if len(report.Improvements) > 0 {
		fmt.Println("\nIMPROVEMENTS:")
		for i, imp := range report.Improvements {
			fmt.Printf("  %d. %s\n", i+1, imp)
		}
	}

	return nil
}

func scorecardGrade(total int) string {
	switch {
	case total >= 85:
		return "A"
	case total >= 65:
		return "B"
	case total >= 50:
		return "C"
	default:
		return "D"
	}
}

func boolStatus(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
