package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullRun(t *testing.T) {
	if os.Getenv("FULL_RUN") == "" {
		t.Skip("Set FULL_RUN=1 to run full press test")
	}

	// Build the press binary first
	pressBinary := filepath.Join(t.TempDir(), "printing-press")
	repoRoot := findRepoRoot()
	cmd := exec.Command("go", "build", "-o", pressBinary, "./cmd/printing-press")
	cmd.Dir = repoRoot
	require.NoError(t, cmd.Run(), "failed to build printing-press")

	baseDir := filepath.Join(os.TempDir(), "press-fullrun-"+time.Now().Format("150405"))
	os.MkdirAll(baseDir, 0755)

	apis := []struct {
		name, level, flag, url string
	}{
		{"petstore", "EASY", "--spec", "https://petstore3.swagger.io/api/v3/openapi.json"},
		{"plaid", "MEDIUM", "--spec", "https://raw.githubusercontent.com/plaid/plaid-openapi/master/2020-09-14.yml"},
		{"notion", "HARD", "--docs", "https://developers.notion.com/reference"},
	}

	var results []*FullRunResult
	for _, api := range apis {
		t.Run(api.name, func(t *testing.T) {
			outputDir := filepath.Join(baseDir, api.name+"-cli")
			result := MakeBestCLI(api.name, api.level, api.flag, api.url, outputDir, pressBinary)
			results = append(results, result)

			assert.Equal(t, 7, result.GatesPassed, "%s: all 7 gates should pass", api.name)
			assert.True(t, result.CommandCount > 0, "%s: should have commands", api.name)
			assert.NotNil(t, result.Scorecard, "%s: should have scorecard", api.name)
		})
	}

	// Print comparison table
	table := PrintComparisonTable(results)
	fmt.Println(table)

	// Write learnings plan
	learningsPath := filepath.Join(baseDir, "learnings-plan.md")
	GenerateLearningsPlan(results, learningsPath)
	fmt.Printf("Learnings plan: %s\n", learningsPath)

	// Also write results to file
	os.WriteFile(filepath.Join(baseDir, "comparison-table.txt"), []byte(table), 0644)
	fmt.Printf("Full results at: %s\n", baseDir)
}

func TestCopySpecToOutput(t *testing.T) {
	tests := []struct {
		name      string
		specFlag  string
		setup     func(t *testing.T, dir string) string // returns specURL
		wantCopy  bool
		wantError bool
	}{
		{
			name:     "copies spec when flag is --spec",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "input-spec.json")
				require.NoError(t, os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o644))
				return specPath
			},
			wantCopy: true,
		},
		{
			name:     "skips when flag is --docs",
			specFlag: "--docs",
			setup: func(t *testing.T, dir string) string {
				return "https://developers.notion.com/reference"
			},
			wantCopy: false,
		},
		{
			name:     "returns error when spec file missing",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
			wantCopy:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			outputDir := filepath.Join(dir, "output")
			require.NoError(t, os.MkdirAll(outputDir, 0o755))

			specURL := tt.setup(t, dir)
			err := copySpecToOutput(tt.specFlag, specURL, outputDir)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			dst := filepath.Join(outputDir, "spec.json")
			if tt.wantCopy {
				data, readErr := os.ReadFile(dst)
				require.NoError(t, readErr, "spec.json should exist in output dir")
				expected, _ := os.ReadFile(specURL)
				assert.Equal(t, expected, data, "spec.json content should match source")
			} else if !tt.wantError {
				_, readErr := os.ReadFile(dst)
				assert.True(t, os.IsNotExist(readErr), "spec.json should not exist")
			}
		})
	}
}

func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
