package pipeline

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewState(t *testing.T) {
	s := NewState("test-api", "/tmp/test-api-cli")
	assert.Equal(t, "test-api", s.APIName)
	assert.Equal(t, "/tmp/test-api-cli", s.OutputDir)
	assert.Len(t, s.Phases, len(PhaseOrder))

	for _, name := range PhaseOrder {
		assert.Equal(t, StatusPending, s.Phases[name].Status)
		assert.NotEmpty(t, s.Phases[name].PlanPath)
	}
}

func TestStateRoundTrip(t *testing.T) {
	s := NewState("roundtrip-test", "/tmp/rt-cli")
	s.SpecPath = "/tmp/spec.yaml"
	s.Complete(PhasePreflight)
	s.MarkSeedWritten(PhaseScaffold)

	require.NoError(t, s.Save())
	defer os.RemoveAll(PipelineDir("roundtrip-test"))

	loaded, err := LoadState("roundtrip-test")
	require.NoError(t, err)

	assert.Equal(t, "roundtrip-test", loaded.APIName)
	assert.Equal(t, "/tmp/spec.yaml", loaded.SpecPath)
	assert.Equal(t, StatusCompleted, loaded.Phases[PhasePreflight].Status)
	assert.Equal(t, PlanStatusCompleted, loaded.Phases[PhasePreflight].PlanStatus)
	assert.Equal(t, StatusPlanned, loaded.Phases[PhaseScaffold].Status)
	assert.Equal(t, PlanStatusSeed, loaded.Phases[PhaseScaffold].PlanStatus)
	assert.Equal(t, StatusPending, loaded.Phases[PhaseEnrich].Status)
	assert.Empty(t, loaded.Phases[PhaseEnrich].PlanStatus)
}

func TestNextPhase(t *testing.T) {
	s := NewState("next-test", "/tmp/test")
	assert.Equal(t, PhasePreflight, s.NextPhase())

	s.Complete(PhasePreflight)
	assert.Equal(t, PhaseResearch, s.NextPhase())

	s.Complete(PhaseResearch)
	assert.Equal(t, PhaseScaffold, s.NextPhase())

	for _, name := range PhaseOrder {
		s.Complete(name)
	}
	assert.Equal(t, "", s.NextPhase())
	assert.True(t, s.IsComplete())
}

func TestPhaseTransitions(t *testing.T) {
	s := NewState("transition-test", "/tmp/test")

	s.MarkSeedWritten(PhasePreflight)
	assert.Equal(t, StatusPlanned, s.Phases[PhasePreflight].Status)
	assert.Equal(t, PlanStatusSeed, s.Phases[PhasePreflight].PlanStatus)

	s.MarkExpanded(PhasePreflight)
	assert.Equal(t, StatusPlanned, s.Phases[PhasePreflight].Status)
	assert.Equal(t, PlanStatusExpanded, s.Phases[PhasePreflight].PlanStatus)

	s.Start(PhasePreflight)
	assert.Equal(t, StatusExecuting, s.Phases[PhasePreflight].Status)

	s.Complete(PhasePreflight)
	assert.Equal(t, StatusCompleted, s.Phases[PhasePreflight].Status)
	assert.Equal(t, PlanStatusCompleted, s.Phases[PhasePreflight].PlanStatus)

	s.Fail(PhaseScaffold)
	assert.Equal(t, StatusFailed, s.Phases[PhaseScaffold].Status)
}

func TestMarkExpandedFromPendingMarksPlanned(t *testing.T) {
	s := NewState("expanded-test", "/tmp/test")

	s.MarkExpanded(PhaseScaffold)

	assert.Equal(t, StatusPlanned, s.Phases[PhaseScaffold].Status)
	assert.Equal(t, PlanStatusExpanded, s.Phases[PhaseScaffold].PlanStatus)
}

func TestIsSeedBackwardCompatible(t *testing.T) {
	s := NewState("seed-test", "/tmp/test")
	assert.True(t, s.IsSeed(PhasePreflight))

	s.MarkSeedWritten(PhasePreflight)
	assert.True(t, s.IsSeed(PhasePreflight))

	s.MarkExpanded(PhasePreflight)
	assert.False(t, s.IsSeed(PhasePreflight))
}

func TestDefaultOutputDir(t *testing.T) {
	tests := []struct {
		name     string
		apiName  string
		expected string
	}{
		{"simple", "stripe", "shelf/stripe-cli"},
		{"hyphenated", "my-api", "shelf/my-api-cli"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DefaultOutputDir(tt.apiName))
		})
	}
}

func TestPhaseStateJSONIncludesPlanStatus(t *testing.T) {
	state := PhaseState{
		Status:     StatusPlanned,
		PlanPath:   "docs/plans/test.md",
		PlanStatus: PlanStatusSeed,
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	assert.JSONEq(t, `{"status":"planned","plan_path":"docs/plans/test.md","plan_status":"seed"}`, string(data))
}
