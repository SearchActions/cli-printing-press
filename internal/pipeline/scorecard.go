package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scorecard holds the auto-scored evaluation of a generated CLI against the Steinberger bar.
type Scorecard struct {
	APIName          string       `json:"api_name"`
	Steinberger      SteinerScore `json:"steinberger"`
	CompetitorScores []CompScore  `json:"competitor_scores"`
	OverallGrade     string       `json:"overall_grade"`
	GapReport        []string     `json:"gap_report"`
}

// SteinerScore breaks down the Steinberger bar into 10 dimensions, each 0-10.
type SteinerScore struct {
	OutputModes   int `json:"output_modes"`    // 0-10
	Auth          int `json:"auth"`            // 0-10
	ErrorHandling int `json:"error_handling"`  // 0-10
	TerminalUX    int `json:"terminal_ux"`     // 0-10
	README        int `json:"readme"`          // 0-10
	Doctor        int `json:"doctor"`          // 0-10
	AgentNative   int `json:"agent_native"`    // 0-10
	LocalCache    int `json:"local_cache"`     // 0-10
	Breadth       int `json:"breadth"`         // 0-10: how many commands (penalizes empty CLIs)
	Vision        int `json:"vision"`          // 0-10
	Total         int `json:"total"`           // 0-100
	Percentage    int `json:"percentage"`      // 0-100
}

// CompScore compares our score against a competitor on a single dimension.
type CompScore struct {
	Name       string `json:"name"`
	OurScore   int    `json:"our_score"`
	TheirScore int    `json:"their_score"`
	WeWin      bool   `json:"we_win"`
}

// RunScorecard evaluates generated CLI files and produces a scorecard.
func RunScorecard(outputDir, pipelineDir string) (*Scorecard, error) {
	sc := &Scorecard{}

	// Infer API name from outputDir basename
	sc.APIName = filepath.Base(outputDir)

	// Score each Steinberger dimension by inspecting generated files
	sc.Steinberger.OutputModes = scoreOutputModes(outputDir)
	sc.Steinberger.Auth = scoreAuth(outputDir)
	sc.Steinberger.ErrorHandling = scoreErrorHandling(outputDir)
	sc.Steinberger.TerminalUX = scoreTerminalUX(outputDir)
	sc.Steinberger.README = scoreREADME(outputDir)
	sc.Steinberger.Doctor = scoreDoctor(outputDir)
	sc.Steinberger.AgentNative = scoreAgentNative(outputDir)
	sc.Steinberger.LocalCache = scoreLocalCache(outputDir)
	sc.Steinberger.Breadth = scoreBreadth(outputDir)
	sc.Steinberger.Vision = scoreVision(outputDir)

	sc.Steinberger.Total = sc.Steinberger.OutputModes +
		sc.Steinberger.Auth +
		sc.Steinberger.ErrorHandling +
		sc.Steinberger.TerminalUX +
		sc.Steinberger.README +
		sc.Steinberger.Doctor +
		sc.Steinberger.AgentNative +
		sc.Steinberger.LocalCache +
		sc.Steinberger.Breadth +
		sc.Steinberger.Vision

	if sc.Steinberger.Total > 0 {
		sc.Steinberger.Percentage = (sc.Steinberger.Total * 100) / 100
	}

	// Grade
	sc.OverallGrade = computeGrade(sc.Steinberger.Percentage)

	// Gap report for dimensions below 5
	sc.GapReport = buildGapReport(sc.Steinberger)

	// Competitor comparison from research.json
	sc.CompetitorScores = buildCompetitorScores(sc.Steinberger.Total, pipelineDir)

	// Write scorecard.md
	if err := writeScorecardMD(sc, pipelineDir); err != nil {
		return sc, fmt.Errorf("writing scorecard.md: %w", err)
	}

	return sc, nil
}

func scoreOutputModes(dir string) int {
	content := readFileContent(filepath.Join(dir, "internal", "cli", "root.go"))
	score := 0
	if strings.Contains(content, "json") {
		score += 2
	}
	if strings.Contains(content, "plain") {
		score += 2
	}
	if strings.Contains(content, "select") {
		score += 2
	}
	if strings.Contains(content, "table") {
		score += 2
	}
	if strings.Contains(content, "csv") {
		score += 2
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreAuth(dir string) int {
	score := 0
	configContent := readFileContent(filepath.Join(dir, "internal", "config", "config.go"))
	envCount := strings.Count(configContent, "os.Getenv")
	if envCount >= 2 {
		score += 8
	} else if envCount >= 1 {
		score += 5
	}
	if fileExists(filepath.Join(dir, "internal", "cli", "auth.go")) {
		score += 2
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreErrorHandling(dir string) int {
	content := readFileContent(filepath.Join(dir, "internal", "cli", "helpers.go"))
	score := 0
	if strings.Contains(content, "hint:") || strings.Contains(content, "Hint:") {
		score += 5
	}
	// Count typed exit codes (cliError with code:)
	exitCount := strings.Count(content, "code:")
	if exitCount > 5 {
		exitCount = 5
	}
	score += exitCount
	if score > 10 {
		score = 10
	}
	return score
}

func scoreTerminalUX(dir string) int {
	content := readFileContent(filepath.Join(dir, "internal", "cli", "helpers.go"))
	score := 0
	if strings.Contains(content, "colorEnabled") {
		score += 5
	}
	if strings.Contains(content, "NO_COLOR") {
		score += 3
	}
	if strings.Contains(content, "isatty") {
		score += 2
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreREADME(dir string) int {
	content := readFileContent(filepath.Join(dir, "README.md"))
	score := 0
	sections := map[string]int{
		"Quick Start":     2,
		"Output Formats":  2,
		"Agent Usage":     2,
		"Troubleshooting": 2,
		"Doctor":          2,
	}
	for section, pts := range sections {
		if strings.Contains(content, section) {
			score += pts
		}
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreDoctor(dir string) int {
	content := readFileContent(filepath.Join(dir, "internal", "cli", "doctor.go"))
	// Count health check patterns (both stdlib and variable-based calls)
	healthChecks := strings.Count(content, "http.Get") +
		strings.Count(content, "http.Head") +
		strings.Count(content, "http.NewRequest") +
		strings.Count(content, "http.Client") +
		strings.Count(content, "httpClient.Get") +
		strings.Count(content, "httpClient.Head") +
		strings.Count(content, "httpClient.Do")
	score := healthChecks * 2
	if score > 10 {
		score = 10
	}
	return score
}

func scoreAgentNative(dir string) int {
	rootContent := readFileContent(filepath.Join(dir, "internal", "cli", "root.go"))
	helpersContent := readFileContent(filepath.Join(dir, "internal", "cli", "helpers.go"))
	combined := rootContent + helpersContent

	score := 0
	if strings.Contains(combined, "json") {
		score += 2
	}
	if strings.Contains(combined, "select") {
		score += 2
	}
	if strings.Contains(combined, "dry-run") || strings.Contains(combined, "dryRun") || strings.Contains(combined, "dry_run") {
		score += 2
	}
	if strings.Contains(combined, "non-interactive") || strings.Contains(combined, "nonInteractive") {
		score += 1
	}
	// Check for --stdin support
	if strings.Contains(combined, "stdin") {
		score += 1
	}
	// Check for --yes flag
	if strings.Contains(combined, `"yes"`) {
		score += 1
	}
	// Check for idempotency handling (409 or "already exists")
	if strings.Contains(helpersContent, "409") || strings.Contains(helpersContent, "already exists") {
		score += 1
	}
	// Check for --human-friendly flag (agent-safe-by-default coloring)
	if strings.Contains(combined, "human-friendly") || strings.Contains(combined, "humanFriendly") {
		score += 1
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreLocalCache(dir string) int {
	score := 0
	// Check client for cache-related code
	clientContent := readFileContent(filepath.Join(dir, "internal", "client", "client.go"))
	if strings.Contains(clientContent, "cacheDir") || strings.Contains(clientContent, "readCache") || strings.Contains(clientContent, "writeCache") {
		score += 5
	}
	if strings.Contains(clientContent, "no-cache") || strings.Contains(clientContent, "NoCache") {
		score += 2
	}
	// Check for sqlite or advanced caching
	for _, name := range []string{"internal/cache/cache.go", "internal/store/store.go"} {
		content := readFileContent(filepath.Join(dir, name))
		if strings.Contains(content, "sqlite") || strings.Contains(content, "bolt") || strings.Contains(content, "badger") {
			score += 3
		}
	}
	if score > 10 {
		score = 10
	}
	return score
}

func scoreBreadth(dir string) int {
	// Count command files in internal/cli/ (exclude infrastructure files)
	cliDir := filepath.Join(dir, "internal", "cli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		return 0
	}
	infra := map[string]bool{
		"helpers.go": true, "root.go": true, "doctor.go": true, "auth.go": true,
	}
	commandFiles := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if infra[e.Name()] {
			continue
		}
		commandFiles++
	}

	switch {
	case commandFiles >= 60:
		return 10
	case commandFiles >= 41:
		return 9
	case commandFiles >= 21:
		return 7
	case commandFiles >= 11:
		return 5
	case commandFiles >= 5:
		return 3
	default:
		return 0 // < 5 commands = basically empty
	}
}

func scoreVision(dir string) int {
	score := 0
	cliDir := filepath.Join(dir, "internal", "cli")
	if fileExists(filepath.Join(cliDir, "export.go")) {
		score += 2
	}
	if fileExists(filepath.Join(cliDir, "import.go")) {
		score += 1
	}
	if fileExists(filepath.Join(dir, "internal", "store", "store.go")) {
		score += 2
	}
	if fileExists(filepath.Join(cliDir, "search.go")) {
		score += 2
	}
	if fileExists(filepath.Join(cliDir, "sync.go")) {
		score += 1
	}
	if fileExists(filepath.Join(cliDir, "tail.go")) {
		score += 1
	}
	// Check for workflow or compound command files
	entries, err := os.ReadDir(cliDir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if strings.Contains(name, "_workflow") || strings.Contains(name, "_compound") {
				if strings.HasSuffix(name, ".go") {
					score++
					break
				}
			}
		}
	}
	if score > 10 {
		score = 10
	}
	return score
}

func readFileContent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func computeGrade(percentage int) string {
	switch {
	case percentage >= 80:
		return "A"
	case percentage >= 65:
		return "B"
	case percentage >= 50:
		return "C"
	case percentage >= 35:
		return "D"
	default:
		return "F"
	}
}

func buildGapReport(s SteinerScore) []string {
	var gaps []string
	dimensions := []struct {
		name  string
		score int
	}{
		{"output_modes", s.OutputModes},
		{"auth", s.Auth},
		{"error_handling", s.ErrorHandling},
		{"terminal_ux", s.TerminalUX},
		{"readme", s.README},
		{"doctor", s.Doctor},
		{"agent_native", s.AgentNative},
		{"local_cache", s.LocalCache},
		{"breadth", s.Breadth},
		{"vision", s.Vision},
	}
	for _, d := range dimensions {
		if d.score < 5 {
			gaps = append(gaps, fmt.Sprintf("%s scored %d/10 - needs improvement", d.name, d.score))
		}
	}
	return gaps
}

func buildCompetitorScores(ourTotal int, pipelineDir string) []CompScore {
	research, err := LoadResearch(pipelineDir)
	if err != nil {
		return nil
	}
	var scores []CompScore
	for _, alt := range research.Alternatives {
		theirScore := estimateCompetitorTotal(alt)
		scores = append(scores, CompScore{
			Name:       alt.Name,
			OurScore:   ourTotal,
			TheirScore: theirScore,
			WeWin:      ourTotal > theirScore,
		})
	}
	return scores
}

func estimateCompetitorTotal(alt Alternative) int {
	score := 0
	if alt.HasJSON {
		score += 6 // output_modes partial credit
	}
	if alt.HasAuth {
		score += 5 // auth partial credit
	}
	// Assume basic error handling and terminal UX
	score += 3
	score += 3
	// README and doctor are unknowns - give partial credit
	score += 4
	score += 2
	// Agent native: partial if they have JSON
	if alt.HasJSON {
		score += 3
	}
	return score
}

func writeScorecardMD(sc *Scorecard, pipelineDir string) error {
	if err := os.MkdirAll(pipelineDir, 0o755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Scorecard: %s\n\n", sc.APIName))
	b.WriteString(fmt.Sprintf("**Overall Grade: %s** (%d%%)\n\n", sc.OverallGrade, sc.Steinberger.Percentage))

	// Steinberger dimensions table
	b.WriteString("## Steinberger Bar\n\n")
	b.WriteString("| Dimension | Score | Max |\n")
	b.WriteString("|-----------|-------|-----|\n")
	s := sc.Steinberger
	dimensions := []struct {
		name  string
		score int
	}{
		{"Output Modes", s.OutputModes},
		{"Auth", s.Auth},
		{"Error Handling", s.ErrorHandling},
		{"Terminal UX", s.TerminalUX},
		{"README", s.README},
		{"Doctor", s.Doctor},
		{"Agent Native", s.AgentNative},
		{"Local Cache", s.LocalCache},
		{"Breadth", s.Breadth},
		{"Vision", s.Vision},
	}
	for _, d := range dimensions {
		bar := strings.Repeat("#", d.score) + strings.Repeat(".", 10-d.score)
		b.WriteString(fmt.Sprintf("| %s | %d/10 %s |\n", d.name, d.score, bar))
	}
	b.WriteString(fmt.Sprintf("| **Total** | **%d/100** |\n\n", s.Total))

	// Competitor comparison
	if len(sc.CompetitorScores) > 0 {
		b.WriteString("## Competitor Comparison\n\n")
		b.WriteString("| Competitor | Ours | Theirs | Winner |\n")
		b.WriteString("|------------|------|--------|--------|\n")
		for _, cs := range sc.CompetitorScores {
			winner := "Them"
			if cs.WeWin {
				winner = "Us"
			}
			b.WriteString(fmt.Sprintf("| %s | %d | %d | %s |\n", cs.Name, cs.OurScore, cs.TheirScore, winner))
		}
		b.WriteString("\n")
	}

	// Gap report
	if len(sc.GapReport) > 0 {
		b.WriteString("## Gaps\n\n")
		for _, g := range sc.GapReport {
			b.WriteString(fmt.Sprintf("- %s\n", g))
		}
		b.WriteString("\n")
	}

	return os.WriteFile(filepath.Join(pipelineDir, "scorecard.md"), []byte(b.String()), 0o644)
}

// LoadScorecard reads a scorecard from a pipeline directory's scorecard.json.
func LoadScorecard(pipelineDir string) (*Scorecard, error) {
	data, err := os.ReadFile(filepath.Join(pipelineDir, "scorecard.json"))
	if err != nil {
		return nil, err
	}
	var sc Scorecard
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, err
	}
	return &sc, nil
}
