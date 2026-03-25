package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
)

// Options configures a pipeline run.
type Options struct {
	OutputDir string
	Force     bool
	Resume    bool
	Phase     string
}

// Init creates the pipeline directory, state file, and plan seeds.
// It does NOT execute any phases.
func Init(apiName string, opts Options) (*PipelineState, error) {
	if opts.Resume && StateExists(apiName) {
		return LoadState(apiName)
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "./" + apiName + "-cli"
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("resolving output dir: %w", err)
	}

	pipeDir := PipelineDir(apiName)
	if StateExists(apiName) && !opts.Force {
		return nil, fmt.Errorf("pipeline for %q already exists at %s (use --force to overwrite or --resume to continue)", apiName, pipeDir)
	}

	specURL, specSource, err := DiscoverSpec(apiName)
	if err != nil {
		return nil, fmt.Errorf("discovering spec: %w", err)
	}

	state := NewState(apiName, absOutputDir)
	state.SpecURL = specURL

	if err := os.MkdirAll(pipeDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating pipeline dir: %w", err)
	}

	seedData := SeedData{
		APIName:     apiName,
		OutputDir:   absOutputDir,
		SpecURL:     specURL,
		SpecSource:  specSource,
		PipelineDir: pipeDir,
	}

	// Write only the first two phases as static seeds (preflight + research).
	// Subsequent phases are generated dynamically after prior phases complete.
	staticPhases := []string{PhasePreflight, PhaseResearch}
	for _, phase := range staticPhases {
		content, err := RenderSeed(phase, seedData)
		if err != nil {
			return nil, fmt.Errorf("rendering seed for %s: %w", phase, err)
		}
		planPath := state.PlanPath(phase)
		if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("writing plan seed for %s: %w", phase, err)
		}
		state.MarkSeedWritten(phase)
	}

	// For remaining phases, write placeholder plans that will be replaced
	// dynamically when prior phases complete (via CompleteAndPlanNext).
	for _, phase := range PhaseOrder {
		if phase == PhasePreflight || phase == PhaseResearch {
			continue
		}
		content, err := RenderSeed(phase, seedData)
		if err != nil {
			return nil, fmt.Errorf("rendering seed for %s: %w", phase, err)
		}
		planPath := state.PlanPath(phase)
		if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("writing plan seed for %s: %w", phase, err)
		}
		state.MarkSeedWritten(phase)
	}

	if err := state.Save(); err != nil {
		return nil, fmt.Errorf("saving state: %w", err)
	}

	return state, nil
}
