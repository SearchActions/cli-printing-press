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
	PhasePreflight  = "preflight"
	PhaseScaffold   = "scaffold"
	PhaseEnrich     = "enrich"
	PhaseRegenerate = "regenerate"
	PhaseReview     = "review"
	PhaseShip       = "ship"
)

// PhaseOrder defines execution order.
var PhaseOrder = []string{
	PhasePreflight,
	PhaseScaffold,
	PhaseEnrich,
	PhaseRegenerate,
	PhaseReview,
	PhaseShip,
}

const (
	StatusPending   = "pending"
	StatusPlanned   = "planned"   // plan.md exists but not yet executed
	StatusExecuting = "executing" // ce:work is running on the plan
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// PipelineState tracks which phases are done across sessions.
type PipelineState struct {
	APIName        string                `json:"api_name"`
	OutputDir      string                `json:"output_dir"`
	StartedAt      time.Time             `json:"started_at"`
	Phases         map[string]PhaseState `json:"phases"`
	SpecPath       string                `json:"spec_path,omitempty"`
	SpecURL        string                `json:"spec_url,omitempty"`
	DogfoodTimeout int                   `json:"dogfood_timeout_seconds,omitempty"` // default 600 (10 min)
}

// PhaseState tracks a single phase.
type PhaseState struct {
	Status   string `json:"status"`
	PlanPath string `json:"plan_path,omitempty"`
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
	for i, name := range PhaseOrder {
		phases[name] = PhaseState{
			Status:   StatusPending,
			PlanPath: filepath.Join(PipelineDir(apiName), fmt.Sprintf("%02d-%s-plan.md", i, name)),
		}
	}
	state := &PipelineState{
		APIName:        apiName,
		OutputDir:      outputDir,
		StartedAt:      time.Now(),
		Phases:         phases,
		DogfoodTimeout: 600, // 10 minutes default
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

// LoadState reads existing state from disk.
func LoadState(apiName string) (*PipelineState, error) {
	data, err := os.ReadFile(StatePath(apiName))
	if err != nil {
		return nil, fmt.Errorf("reading state: %w", err)
	}
	var s PipelineState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
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

// Complete marks a phase as completed.
func (s *PipelineState) Complete(phase string) {
	p := s.Phases[phase]
	p.Status = StatusCompleted
	s.Phases[phase] = p
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
		if s.Phases[name].Status != StatusCompleted {
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
