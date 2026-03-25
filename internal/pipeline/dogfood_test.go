package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTier1Commands(t *testing.T) {
	t.Run("zero resources returns 3 commands", func(t *testing.T) {
		cmds := buildTier1Commands("/bin/fake", nil)
		assert.Len(t, cmds, 3)
	})

	t.Run("two resources returns 5 commands", func(t *testing.T) {
		cmds := buildTier1Commands("/bin/fake", []string{"users", "projects"})
		assert.Len(t, cmds, 5)
	})
}

func TestHasCredentials(t *testing.T) {
	t.Run("empty slice returns false", func(t *testing.T) {
		assert.False(t, hasCredentials(nil))
		assert.False(t, hasCredentials([]string{}))
	})

	t.Run("unset vars return false", func(t *testing.T) {
		assert.False(t, hasCredentials([]string{"TOTALLY_UNSET_VAR_12345"}))
	})

	t.Run("set var returns true", func(t *testing.T) {
		t.Setenv("TEST_CRED_ABC", "secret123")
		assert.True(t, hasCredentials([]string{"TEST_CRED_ABC"}))
	})
}

func TestComputeDogfoodScore(t *testing.T) {
	t.Run("all pass returns 40+", func(t *testing.T) {
		r := &DogfoodResults{
			Tier:           1,
			TotalCommands:  5,
			PassedCommands: 5,
		}
		score := computeDogfoodScore(r)
		assert.GreaterOrEqual(t, score, 40)
	})

	t.Run("half pass returns ~20", func(t *testing.T) {
		r := &DogfoodResults{
			Tier:           1,
			TotalCommands:  10,
			PassedCommands: 5,
		}
		score := computeDogfoodScore(r)
		assert.Equal(t, 20, score)
	})

	t.Run("zero pass returns 0", func(t *testing.T) {
		r := &DogfoodResults{
			Tier:           1,
			TotalCommands:  5,
			PassedCommands: 0,
		}
		score := computeDogfoodScore(r)
		assert.Equal(t, 0, score)
	})

	t.Run("zero total returns 0", func(t *testing.T) {
		r := &DogfoodResults{
			Tier:          1,
			TotalCommands: 0,
		}
		score := computeDogfoodScore(r)
		assert.Equal(t, 0, score)
	})

	t.Run("tier bonus adds points", func(t *testing.T) {
		r := &DogfoodResults{
			Tier:           2,
			TotalCommands:  5,
			PassedCommands: 5,
		}
		score := computeDogfoodScore(r)
		assert.Equal(t, 45, score) // 40 base + 5 tier2 bonus
	})
}

func TestWriteAndLoadDogfoodResults(t *testing.T) {
	dir := t.TempDir()
	original := &DogfoodResults{
		Tier:           2,
		TotalCommands:  4,
		PassedCommands: 3,
		FailedCommands: 1,
		Score:          35,
		Commands: []CommandResult{
			{Tier: 1, Command: "--help", ExitCode: 0, Pass: true},
			{Tier: 1, Command: "version", ExitCode: 0, Pass: true},
			{Tier: 1, Command: "doctor", ExitCode: 0, Pass: true},
			{Tier: 2, Command: "users list --json", ExitCode: 1, Pass: false},
		},
	}

	err := writeDogfoodResults(original, dir)
	require.NoError(t, err)

	loaded, err := LoadDogfoodResults(dir)
	require.NoError(t, err)
	assert.Equal(t, 2, loaded.Tier)
	assert.Equal(t, 4, loaded.TotalCommands)
	assert.Equal(t, 3, loaded.PassedCommands)
	assert.Equal(t, 1, loaded.FailedCommands)
	assert.Equal(t, 35, loaded.Score)
	assert.Len(t, loaded.Commands, 4)
}
