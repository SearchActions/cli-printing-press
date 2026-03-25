package llm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Available returns true if any supported LLM CLI is installed.
func Available() bool {
	_, err1 := exec.LookPath("claude")
	_, err2 := exec.LookPath("codex")
	return err1 == nil || err2 == nil
}

// Run sends a prompt to the best available LLM CLI and returns the response.
// It tries claude first, then falls back to codex. The prompt is written to a
// temp file and passed via the -p flag to avoid ARG_MAX issues with large prompts.
func Run(prompt string) (string, error) {
	// Write prompt to temp file to avoid ARG_MAX issues
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("llm-prompt-%d.md", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("writing prompt: %w", err)
	}
	defer os.Remove(tmpFile)

	// Try claude first (--print -p runs non-interactively)
	if path, err := exec.LookPath("claude"); err == nil {
		cmd := exec.Command(path, "--print", "-p", prompt)
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
		// Fall through to codex
	}

	// Try codex (--quiet --prompt runs non-interactively)
	if path, err := exec.LookPath("codex"); err == nil {
		cmd := exec.Command(path, "--quiet", "--prompt", prompt)
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
		return "", fmt.Errorf("codex failed: %w", err)
	}

	return "", fmt.Errorf("no LLM CLI found (install claude or codex)")
}
