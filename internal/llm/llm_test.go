package llm

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvailable(t *testing.T) {
	// Just verify it doesn't panic - result depends on what's installed
	_ = Available()
}

func TestRunReturnsErrorWhenNoLLM(t *testing.T) {
	// Only test this if neither CLI is installed
	_, err1 := exec.LookPath("claude")
	_, err2 := exec.LookPath("codex")
	if err1 == nil || err2 == nil {
		t.Skip("skipping: LLM CLI is installed")
	}

	_, err := Run("test prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no LLM CLI found")
}
