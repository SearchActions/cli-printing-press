package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		wantFile  string                                 // expected output filename
		wantError bool
	}{
		{
			name:     "copies json spec preserving extension",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "input-spec.json")
				require.NoError(t, os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o644))
				return specPath
			},
			wantFile: "spec.json",
		},
		{
			name:     "copies yaml spec preserving extension",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "openapi.yaml")
				require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.0.0\n"), 0o644))
				return specPath
			},
			wantFile: "spec.yaml",
		},
		{
			name:     "copies yml spec preserving extension",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "api.yml")
				require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.0.0\n"), 0o644))
				return specPath
			},
			wantFile: "spec.yml",
		},
		{
			name:     "skips when flag is --docs",
			specFlag: "--docs",
			setup: func(t *testing.T, dir string) string {
				return "https://developers.notion.com/reference"
			},
		},
		{
			name:     "skips when flag is empty",
			specFlag: "",
			setup: func(t *testing.T, dir string) string {
				return ""
			},
		},
		{
			name:     "returns error when spec file missing",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
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

			if tt.wantFile != "" {
				dst := filepath.Join(outputDir, tt.wantFile)
				data, readErr := os.ReadFile(dst)
				require.NoError(t, readErr, "%s should exist in output dir", tt.wantFile)
				expected, _ := os.ReadFile(specURL)
				assert.Equal(t, expected, data, "content should match source")
			} else if !tt.wantError {
				// Verify no spec file was created
				entries, _ := os.ReadDir(outputDir)
				for _, e := range entries {
					assert.False(t, strings.HasPrefix(e.Name(), "spec."), "no spec file should exist, found %s", e.Name())
				}
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
