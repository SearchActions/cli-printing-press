package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DogfoodResults holds the output of the dogfood phase.
type DogfoodResults struct {
	Tier           int             `json:"tier"`
	TotalCommands  int             `json:"total_commands"`
	PassedCommands int             `json:"passed_commands"`
	FailedCommands int             `json:"failed_commands"`
	Commands       []CommandResult `json:"commands"`
	Score          int             `json:"score"` // 0-50
}

// CommandResult records a single dogfood command execution.
type CommandResult struct {
	Tier       int    `json:"tier"`
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	StdoutFile string `json:"stdout_file"`
	Stderr     string `json:"stderr"`
	DurationMs int    `json:"duration_ms"`
	Pass       bool   `json:"pass"`
}

// DogfoodConfig controls what tiers to run.
type DogfoodConfig struct {
	BinaryPath    string
	PipelineDir   string
	MaxTier       int           // 1, 2, or 3
	SandboxSafe   bool          // from KnownSpecs
	CmdTimeout    time.Duration // per-command timeout
	Resources     []string      // resource names from the generated CLI
	AuthEnvVars   []string      // env vars that indicate credentials are available
}

// RunDogfood executes tiered dogfood testing against a generated CLI binary.
func RunDogfood(cfg DogfoodConfig) (*DogfoodResults, error) {
	if cfg.CmdTimeout == 0 {
		cfg.CmdTimeout = 15 * time.Second
	}
	if cfg.MaxTier == 0 {
		cfg.MaxTier = 1
	}

	evidenceDir := filepath.Join(cfg.PipelineDir, "dogfood-evidence")
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating evidence dir: %w", err)
	}

	results := &DogfoodResults{Tier: 1}

	// Tier 1: No auth required - always runs
	tier1Commands := buildTier1Commands(cfg.BinaryPath, cfg.Resources)
	for _, cmd := range tier1Commands {
		cr := runCommand(cfg.BinaryPath, cmd, 1, cfg.CmdTimeout, evidenceDir)
		results.Commands = append(results.Commands, cr)
	}

	// Tier 2: Read-only with auth - if credentials available and tier >= 2
	if cfg.MaxTier >= 2 && hasCredentials(cfg.AuthEnvVars) {
		results.Tier = 2
		tier2Dir := filepath.Join(evidenceDir, "tier2-reads")
		os.MkdirAll(tier2Dir, 0o755)
		tier2Commands := buildTier2Commands(cfg.BinaryPath, cfg.Resources)
		for _, cmd := range tier2Commands {
			cr := runCommand(cfg.BinaryPath, cmd, 2, cfg.CmdTimeout, tier2Dir)
			results.Commands = append(results.Commands, cr)
		}
	}

	// Tier 3: Sandbox write ops - only if sandbox safe and tier >= 3
	if cfg.MaxTier >= 3 && cfg.SandboxSafe {
		results.Tier = 3
		tier3Dir := filepath.Join(evidenceDir, "tier3-writes")
		os.MkdirAll(tier3Dir, 0o755)
		// Tier 3 is API-specific and would need per-API test plans.
		// For now, just record that tier 3 was available.
		fmt.Fprintf(os.Stderr, "info: Tier 3 sandbox testing available but no test plan defined\n")
	}

	// Compute results
	for _, cr := range results.Commands {
		results.TotalCommands++
		if cr.Pass {
			results.PassedCommands++
		} else {
			results.FailedCommands++
		}
	}

	results.Score = computeDogfoodScore(results)

	// Write results
	if err := writeDogfoodResults(results, cfg.PipelineDir); err != nil {
		return results, fmt.Errorf("writing results: %w", err)
	}

	return results, nil
}

// LoadDogfoodResults reads dogfood-results.json from a pipeline directory.
func LoadDogfoodResults(pipelineDir string) (*DogfoodResults, error) {
	data, err := os.ReadFile(filepath.Join(pipelineDir, "dogfood-results.json"))
	if err != nil {
		return nil, err
	}
	var r DogfoodResults
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func buildTier1Commands(binary string, resources []string) []dogfoodCmd {
	cmds := []dogfoodCmd{
		{args: []string{"--help"}, evidenceFile: "tier1-help.txt"},
		{args: []string{"version"}, evidenceFile: "tier1-version.txt"},
		{args: []string{"doctor"}, evidenceFile: "tier1-doctor.txt"},
	}
	// Add per-resource help commands
	resDir := "tier1-resources"
	for _, r := range resources {
		cmds = append(cmds, dogfoodCmd{
			args:         []string{r, "--help"},
			evidenceFile: filepath.Join(resDir, r+"-help.txt"),
		})
	}
	return cmds
}

func buildTier2Commands(binary string, resources []string) []dogfoodCmd {
	var cmds []dogfoodCmd
	for _, r := range resources {
		cmds = append(cmds, dogfoodCmd{
			args:         []string{r, "list", "--json"},
			evidenceFile: r + "-list.txt",
		})
	}
	return cmds
}

type dogfoodCmd struct {
	args         []string
	evidenceFile string
}

func runCommand(binaryPath string, dc dogfoodCmd, tier int, timeout time.Duration, evidenceDir string) CommandResult {
	cr := CommandResult{
		Tier:    tier,
		Command: strings.Join(dc.args, " "),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, binaryPath, dc.args...)
	stdout, err := cmd.CombinedOutput()
	cr.DurationMs = int(time.Since(start).Milliseconds())

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			cr.ExitCode = exitErr.ExitCode()
		} else {
			cr.ExitCode = -1
		}
		cr.Stderr = err.Error()
		cr.Pass = false
	} else {
		cr.ExitCode = 0
		cr.Pass = true
	}

	// Write evidence file
	evidencePath := filepath.Join(evidenceDir, dc.evidenceFile)
	os.MkdirAll(filepath.Dir(evidencePath), 0o755)
	os.WriteFile(evidencePath, stdout, 0o644)
	cr.StdoutFile = evidencePath

	return cr
}

func hasCredentials(envVars []string) bool {
	for _, v := range envVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

func computeDogfoodScore(results *DogfoodResults) int {
	if results.TotalCommands == 0 {
		return 0
	}

	// Base score: percentage of passed commands, scaled to 0-40
	passRate := float64(results.PassedCommands) / float64(results.TotalCommands)
	score := int(passRate * 40)

	// Bonus for higher tiers
	switch results.Tier {
	case 2:
		score += 5
	case 3:
		score += 10
	}

	if score > 50 {
		score = 50
	}
	return score
}

func writeDogfoodResults(results *DogfoodResults, pipelineDir string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pipelineDir, "dogfood-results.json"), data, 0o644)
}
