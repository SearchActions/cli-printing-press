package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreNovelty(t *testing.T) {
	t.Run("no alternatives returns 10", func(t *testing.T) {
		score := scoreNovelty(nil)
		assert.Equal(t, 10, score)
	})

	t.Run("one alt with 10000 stars returns 2", func(t *testing.T) {
		alts := []Alternative{{Name: "popular-cli", Stars: 10000}}
		score := scoreNovelty(alts)
		assert.Equal(t, 2, score)
	})

	t.Run("one alt with 50 stars returns 7", func(t *testing.T) {
		alts := []Alternative{{Name: "small-cli", Stars: 50}}
		score := scoreNovelty(alts)
		assert.Equal(t, 7, score)
	})
}

func TestDeduplicateAlts(t *testing.T) {
	alts := []Alternative{
		{Name: "cli-a", URL: "https://github.com/org/cli-a"},
		{Name: "cli-b", URL: "https://github.com/org/cli-b"},
		{Name: "cli-a-dup", URL: "https://github.com/org/cli-a"},
	}
	result := deduplicateAlts(alts)
	assert.Len(t, result, 2)
	assert.Equal(t, "cli-a", result[0].Name)
	assert.Equal(t, "cli-b", result[1].Name)
}

func TestRecommend(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{1, "skip"},
		{2, "skip"},
		{3, "skip"},
		{4, "proceed-with-gaps"},
		{5, "proceed-with-gaps"},
		{6, "proceed-with-gaps"},
		{7, "proceed"},
		{8, "proceed"},
		{9, "proceed"},
		{10, "proceed"},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, recommend(tt.score))
		})
	}
}

func TestAnalyzeAlternatives(t *testing.T) {
	alts := []Alternative{
		{Name: "tool-a", Language: "python", HasJSON: false},
		{Name: "tool-b", Language: "typescript", HasJSON: true},
	}
	gaps, patterns := analyzeAlternatives(alts)
	assert.NotEmpty(t, gaps)
	assert.NotEmpty(t, patterns)
}

func TestWriteAndLoadResearch(t *testing.T) {
	dir := t.TempDir()
	result := &ResearchResult{
		APIName:        "test-api",
		NoveltyScore:   8,
		Recommendation: "proceed",
		Alternatives: []Alternative{
			{Name: "alt-1", URL: "https://example.com/alt-1"},
		},
		Gaps:     []string{"no --json"},
		Patterns: []string{"standard CRUD"},
	}

	err := writeResearchJSON(result, dir)
	require.NoError(t, err)

	loaded, err := LoadResearch(dir)
	require.NoError(t, err)
	assert.Equal(t, "test-api", loaded.APIName)
	assert.Equal(t, 8, loaded.NoveltyScore)
	assert.Equal(t, "proceed", loaded.Recommendation)
	assert.Len(t, loaded.Alternatives, 1)
	assert.Equal(t, "alt-1", loaded.Alternatives[0].Name)
}
