package pipeline

import (
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
	s.MarkPlanned(PhaseScaffold)

	require.NoError(t, s.Save())
	defer os.RemoveAll(PipelineDir("roundtrip-test"))

	loaded, err := LoadState("roundtrip-test")
	require.NoError(t, err)

	assert.Equal(t, "roundtrip-test", loaded.APIName)
	assert.Equal(t, "/tmp/spec.yaml", loaded.SpecPath)
	assert.Equal(t, StatusCompleted, loaded.Phases[PhasePreflight].Status)
	assert.Equal(t, StatusPlanned, loaded.Phases[PhaseScaffold].Status)
	assert.Equal(t, StatusPending, loaded.Phases[PhaseEnrich].Status)
}

func TestNextPhase(t *testing.T) {
	s := NewState("next-test", "/tmp/test")
	assert.Equal(t, PhasePreflight, s.NextPhase())

	s.Complete(PhasePreflight)
	assert.Equal(t, PhaseScaffold, s.NextPhase())

	for _, name := range PhaseOrder {
		s.Complete(name)
	}
	assert.Equal(t, "", s.NextPhase())
	assert.True(t, s.IsComplete())
}

func TestPhaseTransitions(t *testing.T) {
	s := NewState("transition-test", "/tmp/test")

	s.MarkPlanned(PhasePreflight)
	assert.Equal(t, StatusPlanned, s.Phases[PhasePreflight].Status)

	s.Start(PhasePreflight)
	assert.Equal(t, StatusExecuting, s.Phases[PhasePreflight].Status)

	s.Complete(PhasePreflight)
	assert.Equal(t, StatusCompleted, s.Phases[PhasePreflight].Status)

	s.Fail(PhaseScaffold)
	assert.Equal(t, StatusFailed, s.Phases[PhaseScaffold].Status)
}
