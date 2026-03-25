package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FullRunResult holds everything the press produced for one API.
type FullRunResult struct {
	APIName       string
	Level         string // "EASY", "MEDIUM", "HARD"

	// Step 1: Research
	Research      *ResearchResult
	ResearchError string

	// Step 2: Generate
	OutputDir       string
	GatesPassed     int
	GatesFailed     int
	GatesOutput     string
	CommandCount    int
	ResourceCount   int

	// Step 3: Coverage
	SpecEndpoints   int
	CoveragePercent int

	// Step 4: Dogfood
	Dogfood      *DogfoodResults
	DogfoodError string

	// Step 5: Scorecard
	Scorecard      *Scorecard
	ScorecardError string

	// Step 6: Learnings
	FixPlans []string

	Errors   []string
	Duration time.Duration
}

// MakeBestCLI runs the full printing press pipeline for a single API and
// returns a result summarizing every phase.
func MakeBestCLI(apiName, level, specFlag, specURL, outputDir, pressBinary string) *FullRunResult {
	start := time.Now()
	result := &FullRunResult{
		APIName:   apiName,
		Level:     level,
		OutputDir: outputDir,
	}

	pipelineDir := PipelineDir(apiName)
	os.MkdirAll(pipelineDir, 0o755)

	// Step 1: Research
	research, err := RunResearch(apiName, "catalog", pipelineDir)
	if err != nil {
		result.ResearchError = err.Error()
		result.Errors = append(result.Errors, fmt.Sprintf("research: %v", err))
	}
	result.Research = research

	// Step 2: Generate
	repoRoot := findRepoRootFrom(pressBinary)
	var genArgs []string
	if specFlag == "--docs" {
		genArgs = []string{"generate", "--docs", specURL, "--name", apiName, "--output", outputDir, "--force"}
	} else {
		genArgs = []string{"generate", "--spec", specURL, "--output", outputDir, "--force", "--lenient"}
	}

	cmd := exec.Command(pressBinary, genArgs...)
	cmd.Dir = repoRoot
	genOut, genErr := cmd.CombinedOutput()
	result.GatesOutput = string(genOut)

	for _, line := range strings.Split(result.GatesOutput, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "PASS") {
			result.GatesPassed++
		}
		if strings.Contains(trimmed, "FAIL") {
			result.GatesFailed++
		}
	}

	if genErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("generate: %v", genErr))
		result.Duration = time.Since(start)
		return result
	}

	// Step 3: Count commands and resources
	resources := listResources(outputDir)
	result.ResourceCount = len(resources)
	result.CommandCount = len(resources) + 4 // +4 for root, help, version, doctor

	// Step 4: API Coverage estimate
	skippedCount := strings.Count(result.GatesOutput, "skipping") + strings.Count(result.GatesOutput, "Skipping")
	result.SpecEndpoints = result.ResourceCount + skippedCount
	if result.SpecEndpoints > 0 {
		result.CoveragePercent = (result.ResourceCount * 100) / result.SpecEndpoints
	}

	// Step 5: Dogfood
	cliBinaryPath := filepath.Join(outputDir, apiName+"-cli")
	buildCmd := exec.Command("go", "build", "-o", cliBinaryPath, "./cmd/...")
	buildCmd.Dir = outputDir
	if buildErr := buildCmd.Run(); buildErr != nil {
		result.DogfoodError = fmt.Sprintf("build failed: %v", buildErr)
		result.Errors = append(result.Errors, fmt.Sprintf("dogfood build: %v", buildErr))
	} else {
		cfg := DogfoodConfig{
			BinaryPath:  cliBinaryPath,
			PipelineDir: pipelineDir,
			MaxTier:     1,
			CmdTimeout:  15 * time.Second,
			Resources:   resources,
		}
		dogfood, dogErr := RunDogfood(cfg)
		if dogErr != nil {
			result.DogfoodError = dogErr.Error()
			result.Errors = append(result.Errors, fmt.Sprintf("dogfood: %v", dogErr))
		}
		result.Dogfood = dogfood
	}

	// Step 6: Scorecard
	scorecard, scErr := RunScorecard(outputDir, pipelineDir)
	if scErr != nil {
		result.ScorecardError = scErr.Error()
		result.Errors = append(result.Errors, fmt.Sprintf("scorecard: %v", scErr))
	}
	result.Scorecard = scorecard

	// Step 7: Fix Plans
	if scorecard != nil {
		plans, planErr := GenerateFixPlans(scorecard, pipelineDir)
		if planErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("fix plans: %v", planErr))
		}
		result.FixPlans = plans
	}

	result.Duration = time.Since(start)
	return result
}

// listResources returns the resource names found in the generated CLI's
// internal/cli directory, excluding infrastructure files.
func listResources(outputDir string) []string {
	cliDir := filepath.Join(outputDir, "internal", "cli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return nil
	}

	infraFiles := map[string]bool{
		"helpers.go": true,
		"root.go":    true,
		"doctor.go":  true,
		"auth.go":    true,
	}

	var resources []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if infraFiles[name] {
			continue
		}
		// Resource name is the filename without .go
		resources = append(resources, strings.TrimSuffix(name, ".go"))
	}
	return resources
}

// findRepoRootFrom walks up from the binary path (or cwd) to find go.mod.
func findRepoRootFrom(binaryPath string) string {
	dir := filepath.Dir(binaryPath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Fallback: try cwd
			cwd, _ := os.Getwd()
			return cwd
		}
		dir = parent
	}
}

// PrintComparisonTable produces a formatted comparison table showing all
// FullRunResults side by side with fixed-width columns.
func PrintComparisonTable(results []*FullRunResult) string {
	if len(results) == 0 {
		return "(no results)\n"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("=== CLI Printing Press - Full Run Comparison ===\n\n")

	// Header
	b.WriteString(fmt.Sprintf("%-25s", "Metric"))
	for _, r := range results {
		b.WriteString(fmt.Sprintf("| %-18s", r.APIName+" ("+r.Level+")"))
	}
	b.WriteString("|\n")

	// Separator
	b.WriteString(strings.Repeat("-", 25))
	for range results {
		b.WriteString("|" + strings.Repeat("-", 19))
	}
	b.WriteString("|\n")

	// Quality Gates
	writeRow(&b, "Quality Gates", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d/7 PASS", r.GatesPassed)
	})

	// Commands
	writeRow(&b, "Commands", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d", r.CommandCount)
	})

	// Resources
	writeRow(&b, "Resources", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d", r.ResourceCount)
	})

	// API Coverage
	writeRow(&b, "API Coverage", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d%%", r.CoveragePercent)
	})

	// Steinberger dimensions
	steinDimensions := []struct {
		label string
		get   func(*Scorecard) int
	}{
		{"Output Modes", func(s *Scorecard) int { return s.Steinberger.OutputModes }},
		{"Auth", func(s *Scorecard) int { return s.Steinberger.Auth }},
		{"Error Handling", func(s *Scorecard) int { return s.Steinberger.ErrorHandling }},
		{"Terminal UX", func(s *Scorecard) int { return s.Steinberger.TerminalUX }},
		{"README", func(s *Scorecard) int { return s.Steinberger.README }},
		{"Doctor", func(s *Scorecard) int { return s.Steinberger.Doctor }},
		{"Agent Native", func(s *Scorecard) int { return s.Steinberger.AgentNative }},
		{"Local Cache", func(s *Scorecard) int { return s.Steinberger.LocalCache }},
	}

	for _, dim := range steinDimensions {
		writeRow(&b, "  "+dim.label, results, func(r *FullRunResult) string {
			if r.Scorecard == nil {
				return "n/a"
			}
			return fmt.Sprintf("%d/10", dim.get(r.Scorecard))
		})
	}

	// Steinberger total + %
	writeRow(&b, "Steinberger Total", results, func(r *FullRunResult) string {
		if r.Scorecard == nil {
			return "n/a"
		}
		return fmt.Sprintf("%d/80 (%d%%)", r.Scorecard.Steinberger.Total, r.Scorecard.Steinberger.Percentage)
	})

	// Grade
	writeRow(&b, "Grade", results, func(r *FullRunResult) string {
		if r.Scorecard == nil {
			return "n/a"
		}
		return r.Scorecard.OverallGrade
	})

	// Competitors found
	writeRow(&b, "Competitors Found", results, func(r *FullRunResult) string {
		if r.Research == nil {
			return "0"
		}
		return fmt.Sprintf("%d", len(r.Research.Alternatives))
	})

	// We Win?
	writeRow(&b, "We Win?", results, func(r *FullRunResult) string {
		if r.Scorecard == nil || len(r.Scorecard.CompetitorScores) == 0 {
			return "n/a"
		}
		wins := 0
		for _, cs := range r.Scorecard.CompetitorScores {
			if cs.WeWin {
				wins++
			}
		}
		return fmt.Sprintf("%d/%d", wins, len(r.Scorecard.CompetitorScores))
	})

	// Dogfood pass rate
	writeRow(&b, "Dogfood Pass Rate", results, func(r *FullRunResult) string {
		if r.Dogfood == nil {
			return "n/a"
		}
		if r.Dogfood.TotalCommands == 0 {
			return "0/0"
		}
		pct := (r.Dogfood.PassedCommands * 100) / r.Dogfood.TotalCommands
		return fmt.Sprintf("%d/%d (%d%%)", r.Dogfood.PassedCommands, r.Dogfood.TotalCommands, pct)
	})

	// Fix plans generated
	writeRow(&b, "Fix Plans", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d", len(r.FixPlans))
	})

	// Duration
	writeRow(&b, "Duration", results, func(r *FullRunResult) string {
		return r.Duration.Round(time.Second).String()
	})

	// Errors
	writeRow(&b, "Errors", results, func(r *FullRunResult) string {
		return fmt.Sprintf("%d", len(r.Errors))
	})

	b.WriteString("\n")
	return b.String()
}

func writeRow(b *strings.Builder, label string, results []*FullRunResult, fn func(*FullRunResult) string) {
	b.WriteString(fmt.Sprintf("%-25s", label))
	for _, r := range results {
		b.WriteString(fmt.Sprintf("| %-18s", fn(r)))
	}
	b.WriteString("|\n")
}

// GenerateLearningsPlan writes a markdown plan summarizing consistent gaps
// and recommended fixes across all runs.
func GenerateLearningsPlan(results []*FullRunResult, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	var b strings.Builder
	b.WriteString("# Learnings Plan - CLI Printing Press Full Run\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Summarize runs
	b.WriteString("## Runs\n\n")
	for _, r := range results {
		status := "OK"
		if len(r.Errors) > 0 {
			status = fmt.Sprintf("%d errors", len(r.Errors))
		}
		b.WriteString(fmt.Sprintf("- **%s** (%s) - Gates %d/7, Duration %s, Status: %s\n",
			r.APIName, r.Level, r.GatesPassed, r.Duration.Round(time.Second), status))
	}
	b.WriteString("\n")

	// Find consistent gaps (dimensions scoring <5 across multiple runs)
	type dimTally struct {
		lowCount int
		total    int
		sum      int
	}
	dimNames := []string{"output_modes", "auth", "error_handling", "terminal_ux", "readme", "doctor", "agent_native", "local_cache"}
	dimGetters := []func(*SteinerScore) int{
		func(s *SteinerScore) int { return s.OutputModes },
		func(s *SteinerScore) int { return s.Auth },
		func(s *SteinerScore) int { return s.ErrorHandling },
		func(s *SteinerScore) int { return s.TerminalUX },
		func(s *SteinerScore) int { return s.README },
		func(s *SteinerScore) int { return s.Doctor },
		func(s *SteinerScore) int { return s.AgentNative },
		func(s *SteinerScore) int { return s.LocalCache },
	}

	tallies := make([]dimTally, len(dimNames))
	for _, r := range results {
		if r.Scorecard == nil {
			continue
		}
		for i, getter := range dimGetters {
			score := getter(&r.Scorecard.Steinberger)
			tallies[i].total++
			tallies[i].sum += score
			if score < 5 {
				tallies[i].lowCount++
			}
		}
	}

	b.WriteString("## Consistent Gaps\n\n")
	b.WriteString("Dimensions scoring below 5/10 across multiple runs:\n\n")
	hasGaps := false
	for i, name := range dimNames {
		t := tallies[i]
		if t.total == 0 {
			continue
		}
		avg := t.sum / t.total
		if t.lowCount > 0 {
			hasGaps = true
			b.WriteString(fmt.Sprintf("- **%s** - low in %d/%d runs (avg %d/10)\n", name, t.lowCount, t.total, avg))
		}
	}
	if !hasGaps {
		b.WriteString("No consistent gaps found - all dimensions scored 5+ across runs.\n")
	}
	b.WriteString("\n")

	// Recommended fixes
	b.WriteString("## Recommended Fixes\n\n")
	b.WriteString("Priority order (most impactful first):\n\n")
	priority := 1
	for i, name := range dimNames {
		t := tallies[i]
		if t.total == 0 || t.lowCount == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("%d. **Improve %s templates** - affects %d/%d APIs tested\n",
			priority, name, t.lowCount, t.total))
		if advice, ok := dimensionAdvice[name]; ok {
			// Include first line of advice
			lines := strings.SplitN(advice, "\n", 2)
			b.WriteString(fmt.Sprintf("   - %s\n", strings.TrimSpace(lines[0])))
		}
		priority++
	}
	if priority == 1 {
		b.WriteString("No template fixes needed - all dimensions healthy.\n")
	}
	b.WriteString("\n")

	// Gate failures
	b.WriteString("## Gate Failures\n\n")
	for _, r := range results {
		if r.GatesFailed > 0 {
			b.WriteString(fmt.Sprintf("- **%s** - %d gates failed\n", r.APIName, r.GatesFailed))
		}
	}
	allPassed := true
	for _, r := range results {
		if r.GatesFailed > 0 {
			allPassed = false
			break
		}
	}
	if allPassed {
		b.WriteString("All gates passed across all runs.\n")
	}
	b.WriteString("\n")

	// Dogfood summary
	b.WriteString("## Dogfood Summary\n\n")
	for _, r := range results {
		if r.Dogfood != nil {
			pct := 0
			if r.Dogfood.TotalCommands > 0 {
				pct = (r.Dogfood.PassedCommands * 100) / r.Dogfood.TotalCommands
			}
			b.WriteString(fmt.Sprintf("- **%s** - %d/%d passed (%d%%)\n",
				r.APIName, r.Dogfood.PassedCommands, r.Dogfood.TotalCommands, pct))
		} else {
			b.WriteString(fmt.Sprintf("- **%s** - dogfood not run (%s)\n", r.APIName, r.DogfoodError))
		}
	}
	b.WriteString("\n")

	// Next steps
	b.WriteString("## Next Steps\n\n")
	b.WriteString("1. Fix the highest-priority template gaps listed above\n")
	b.WriteString("2. Re-run this full comparison to verify improvements\n")
	b.WriteString("3. Add more APIs at each difficulty level to broaden coverage\n")

	return os.WriteFile(outputPath, []byte(b.String()), 0o644)
}
