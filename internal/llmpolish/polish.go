package llmpolish

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// PolishRequest defines what the polish pass needs.
type PolishRequest struct {
	APIName    string
	OutputDir  string
	SpecSource string      // OpenAPI spec URL or docs URL
	Research   interface{} // *pipeline.ResearchResult - use interface to avoid circular import
}

// PolishResult summarizes what the polish pass changed.
type PolishResult struct {
	HelpTextsImproved int
	ExamplesAdded     int
	READMERewritten   bool
	TokensUsed        int
	Duration          time.Duration
	Skipped           bool
	SkipReason        string
}

// Polish runs an LLM polish pass over a generated CLI directory.
// It shells out to claude or codex CLI - never calls an LLM API directly.
// If the LLM CLI is unavailable or any step fails, it returns partial results
// rather than crashing the pipeline.
func Polish(req PolishRequest) (*PolishResult, error) {
	start := time.Now()
	result := &PolishResult{}

	llmCmd, llmArgs := findLLMCLI()
	if llmCmd == "" {
		result.Skipped = true
		result.SkipReason = "no LLM CLI found (install claude or codex)"
		result.Duration = time.Since(start)
		return result, nil
	}

	// Step 1: Improve help texts
	helpPrompt := buildHelpPrompt(req.OutputDir)
	if helpPrompt != "" {
		helpOutput, err := runLLM(llmCmd, llmArgs, helpPrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: help text polish failed: %v\n", err)
		} else {
			var improvements []HelpImprovement
			if parseErr := json.Unmarshal(extractJSON(helpOutput), &improvements); parseErr == nil {
				if applyErr := applyHelpTexts(req.OutputDir, improvements); applyErr == nil {
					result.HelpTextsImproved = len(improvements)
				} else {
					fmt.Fprintf(os.Stderr, "warning: applying help texts failed: %v\n", applyErr)
				}
			} else {
				fmt.Fprintf(os.Stderr, "warning: parsing help improvements failed: %v\n", parseErr)
			}
		}
	}

	// Step 2: Add examples
	examplePrompt := buildExamplePrompt(req.OutputDir)
	if examplePrompt != "" {
		exampleOutput, err := runLLM(llmCmd, llmArgs, examplePrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: example polish failed: %v\n", err)
		} else {
			var examples []ExampleSet
			if parseErr := json.Unmarshal(extractJSON(exampleOutput), &examples); parseErr == nil {
				if applyErr := applyExamples(req.OutputDir, examples); applyErr == nil {
					result.ExamplesAdded = countExamples(examples)
				} else {
					fmt.Fprintf(os.Stderr, "warning: applying examples failed: %v\n", applyErr)
				}
			} else {
				fmt.Fprintf(os.Stderr, "warning: parsing examples failed: %v\n", parseErr)
			}
		}
	}

	// Step 3: Rewrite README
	readmePrompt := buildREADMEPrompt(req.OutputDir, req.APIName)
	if readmePrompt != "" {
		readmeOutput, err := runLLM(llmCmd, llmArgs, readmePrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: README polish failed: %v\n", err)
		} else {
			content := extractMarkdown(readmeOutput)
			if content != "" {
				if applyErr := applyREADME(req.OutputDir, content); applyErr == nil {
					result.READMERewritten = true
				} else {
					fmt.Fprintf(os.Stderr, "warning: applying README failed: %v\n", applyErr)
				}
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// findLLMCLI checks for available LLM CLIs in preference order.
// Returns the command name and base args to use.
func findLLMCLI() (string, []string) {
	if path, err := exec.LookPath("claude"); err == nil {
		return path, []string{"--print", "-p"}
	}
	if path, err := exec.LookPath("codex"); err == nil {
		return path, []string{"--quiet", "--prompt"}
	}
	return "", nil
}

// runLLM shells out to the LLM CLI with the given prompt and returns stdout.
func runLLM(llmCmd string, baseArgs []string, prompt string) (string, error) {
	// Write prompt to temp file for debugging/logging
	tmpFile, err := os.CreateTemp("", "polish-prompt-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(prompt); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("writing prompt: %w", err)
	}
	tmpFile.Close()

	args := append(baseArgs, prompt)
	cmd := exec.Command(llmCmd, args...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running %s: %w", filepath.Base(llmCmd), err)
	}
	return string(output), nil
}

// extractJSON finds and returns the first JSON array in the LLM output.
// LLMs often wrap JSON in markdown code blocks.
func extractJSON(output string) []byte {
	// Try to find JSON array directly
	start := -1
	end := -1
	depth := 0
	for i, c := range output {
		if c == '[' {
			if depth == 0 {
				start = i
			}
			depth++
		}
		if c == ']' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if start >= 0 && end > start {
		return []byte(output[start:end])
	}
	return []byte("[]")
}

// extractMarkdown strips markdown code fences from LLM output.
func extractMarkdown(output string) string {
	// Strip leading/trailing whitespace
	s := output
	// Remove ```markdown ... ``` wrapper if present
	if idx := indexOf(s, "```markdown"); idx >= 0 {
		s = s[idx+len("```markdown"):]
		if endIdx := lastIndexOf(s, "```"); endIdx >= 0 {
			s = s[:endIdx]
		}
	} else if idx := indexOf(s, "```md"); idx >= 0 {
		s = s[idx+len("```md"):]
		if endIdx := lastIndexOf(s, "```"); endIdx >= 0 {
			s = s[:endIdx]
		}
	} else if idx := indexOf(s, "```"); idx >= 0 {
		s = s[idx+len("```"):]
		if endIdx := lastIndexOf(s, "```"); endIdx >= 0 {
			s = s[:endIdx]
		}
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func countExamples(examples []ExampleSet) int {
	total := 0
	for _, e := range examples {
		total += len(e.Examples)
	}
	return total
}
