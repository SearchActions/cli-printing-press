package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Phase names in execution order.
const (
	PhasePreflight      = "preflight"
	PhaseResearch       = "research"
	PhaseScaffold       = "scaffold"
	PhaseEnrich         = "enrich"
	PhaseRegenerate     = "regenerate"
	PhaseReview         = "review"
	PhaseAgentReadiness = "agent-readiness"
	PhaseComparative    = "comparative"
	PhaseShip           = "ship"
)

// PhaseOrder defines execution order.
var PhaseOrder = []string{
	PhasePreflight,
	PhaseResearch,
	PhaseScaffold,
	PhaseEnrich,
	PhaseRegenerate,
	PhaseReview,
	PhaseAgentReadiness,
	PhaseComparative,
	PhaseShip,
}

const (
	StatusPending   = "pending"
	StatusPlanned   = "planned"   // plan.md exists but not yet executed
	StatusExecuting = "executing" // ce:work is running on the plan
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

const (
	PlanStatusSeed      = "seed"
	PlanStatusExpanded  = "expanded"
	PlanStatusCompleted = "completed"
)

// PipelineState tracks which phases are done across sessions.
type PipelineState struct {
	Version        int                   `json:"version"`                           // state schema version for migration
	APIName        string                `json:"api_name"`
	OutputDir      string                `json:"output_dir"`
	StartedAt      time.Time             `json:"started_at"`
	Phases         map[string]PhaseState `json:"phases"`
	SpecPath       string                `json:"spec_path,omitempty"`
	SpecURL        string                `json:"spec_url,omitempty"`
	DogfoodTimeout int                   `json:"dogfood_timeout_seconds,omitempty"` // default 600 (10 min)
	DogfoodTier    int                   `json:"dogfood_tier,omitempty"`            // max tier to run (1-3, default 1)
}

const currentStateVersion = 2

// PhaseState tracks a single phase.
type PhaseState struct {
	Status     string `json:"status"`
	PlanPath   string `json:"plan_path,omitempty"`
	PlanStatus string `json:"plan_status,omitempty"`
}

// PipelineDir returns the pipeline state directory path.
func PipelineDir(apiName string) string {
	return filepath.Join("docs", "plans", apiName+"-pipeline")
}

// StatePath returns the state.json path for an API pipeline.
func StatePath(apiName string) string {
	return filepath.Join(PipelineDir(apiName), "state.json")
}

// NewState creates a fresh pipeline state.
func NewState(apiName, outputDir string) *PipelineState {
	phases := make(map[string]PhaseState, len(PhaseOrder))
	for _, name := range PhaseOrder {
		phases[name] = PhaseState{
			Status:   StatusPending,
			PlanPath: filepath.Join(PipelineDir(apiName), fmt.Sprintf("%s-plan.md", name)),
		}
	}
	state := &PipelineState{
		Version:        currentStateVersion,
		APIName:        apiName,
		OutputDir:      outputDir,
		StartedAt:      time.Now(),
		Phases:         phases,
		DogfoodTimeout: 600, // 10 minutes default
		DogfoodTier:    1,   // default to Tier 1 (no auth)
	}
	return state
}

// Save writes state to disk.
func (s *PipelineState) Save() error {
	dir := PipelineDir(s.APIName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating pipeline dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	return os.WriteFile(StatePath(s.APIName), data, 0o644)
}

// LoadState reads existing state from disk, migrating old formats.
func LoadState(apiName string) (*PipelineState, error) {
	data, err := os.ReadFile(StatePath(apiName))
	if err != nil {
		return nil, fmt.Errorf("reading state: %w", err)
	}
	var s PipelineState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}
	// Migrate: add missing phases and update PlanPath from index-based to name-based format.
	if s.Version < currentStateVersion {
		for _, name := range PhaseOrder {
			if _, ok := s.Phases[name]; !ok {
				// Backfill missing phases as completed (both Status and PlanStatus)
				// so NextPhase() doesn't treat them as pending.
				s.Phases[name] = PhaseState{
					Status:     StatusCompleted,
					PlanStatus: PlanStatusCompleted,
					PlanPath:   filepath.Join(PipelineDir(apiName), fmt.Sprintf("%s-plan.md", name)),
				}
			} else {
				// Migrate existing phases from index-based to name-based PlanPath.
				p := s.Phases[name]
				p.PlanPath = filepath.Join(PipelineDir(apiName), fmt.Sprintf("%s-plan.md", name))
				s.Phases[name] = p
			}
		}
		s.Version = currentStateVersion
	}
	return &s, nil
}

// StateExists returns true if a state file exists.
func StateExists(apiName string) bool {
	_, err := os.Stat(StatePath(apiName))
	return err == nil
}

// Start marks a phase as executing.
func (s *PipelineState) Start(phase string) {
	p := s.Phases[phase]
	p.Status = StatusExecuting
	s.Phases[phase] = p
}

// MarkPlanned marks a phase as having its plan.md written.
func (s *PipelineState) MarkPlanned(phase string) {
	p := s.Phases[phase]
	p.Status = StatusPlanned
	s.Phases[phase] = p
}

// MarkSeedWritten marks a phase as having its initial seed plan written.
func (s *PipelineState) MarkSeedWritten(phase string) {
	s.MarkPlanned(phase)
	p := s.Phases[phase]
	p.PlanStatus = PlanStatusSeed
	s.Phases[phase] = p
}

// MarkExpanded marks a phase plan as expanded beyond the initial seed.
func (s *PipelineState) MarkExpanded(phase string) {
	p := s.Phases[phase]
	if p.Status == "" || p.Status == StatusPending {
		p.Status = StatusPlanned
	}
	p.PlanStatus = PlanStatusExpanded
	s.Phases[phase] = p
}

// Complete marks a phase as completed.
func (s *PipelineState) Complete(phase string) {
	p := s.Phases[phase]
	p.Status = StatusCompleted
	p.PlanStatus = PlanStatusCompleted
	s.Phases[phase] = p
}

// CompleteAndPlanNext marks a phase as completed, then generates a dynamic
// plan for the next phase using outputs from all completed phases.
func (s *PipelineState) CompleteAndPlanNext(phase string) error {
	s.Complete(phase)
	nextPhase := s.NextPhase()
	if nextPhase == "" {
		return nil // all phases done
	}

	plan, err := GenerateNextPlan(s, nextPhase)
	if err != nil {
		// Fall back to existing seed plan (already written at init time)
		fmt.Fprintf(os.Stderr, "warning: dynamic plan generation failed for %s, using seed: %v\n", nextPhase, err)
		return nil
	}

	planPath := s.PlanPath(nextPhase)
	if err := os.WriteFile(planPath, []byte(plan), 0o644); err != nil {
		return fmt.Errorf("writing dynamic plan for %s: %w", nextPhase, err)
	}
	s.MarkExpanded(nextPhase)
	return nil
}

// Fail marks a phase as failed.
func (s *PipelineState) Fail(phase string) {
	p := s.Phases[phase]
	p.Status = StatusFailed
	s.Phases[phase] = p
}

// NextPhase returns the name of the next incomplete phase, or "".
func (s *PipelineState) NextPhase() string {
	for _, name := range PhaseOrder {
		if s.Phases[name].PlanStatus != PlanStatusCompleted {
			return name
		}
	}
	return ""
}

// IsComplete returns true if all phases are completed.
func (s *PipelineState) IsComplete() bool {
	return s.NextPhase() == ""
}

// PlanPath returns the plan.md path for a given phase.
func (s *PipelineState) PlanPath(phase string) string {
	return s.Phases[phase].PlanPath
}

// IsSeed reports whether a phase is still at the seed-plan stage.
func (s *PipelineState) IsSeed(phase string) bool {
	return s.Phases[phase].PlanStatus == "" || s.Phases[phase].PlanStatus == PlanStatusSeed
}
