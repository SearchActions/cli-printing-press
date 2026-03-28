package pipeline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		name       string
		specFlag   string
		setup      func(t *testing.T, dir string) string // returns specURL
		wantCopy   bool
		wantJSON   bool // if true, verify output is valid JSON
		wantError  bool
	}{
		{
			name:     "copies local json spec",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "input-spec.json")
				require.NoError(t, os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o644))
				return specPath
			},
			wantCopy: true,
			wantJSON: true,
		},
		{
			name:     "converts local yaml spec to json",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "openapi.yaml")
				require.NoError(t, os.WriteFile(specPath, []byte("openapi: \"3.0.0\"\ninfo:\n  title: Test\n"), 0o644))
				return specPath
			},
			wantCopy: true,
			wantJSON: true,
		},
		{
			name:     "converts local yml spec to json",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				specPath := filepath.Join(dir, "api.yml")
				require.NoError(t, os.WriteFile(specPath, []byte("openapi: \"3.0.0\"\n"), 0o644))
				return specPath
			},
			wantCopy: true,
			wantJSON: true,
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
			name:     "skips when specURL is empty",
			specFlag: "--spec",
			setup: func(t *testing.T, dir string) string {
				return ""
			},
		},
		{
			name:     "returns error when local spec file missing",
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

			dst := filepath.Join(outputDir, "spec.json")
			if tt.wantCopy {
				data, readErr := os.ReadFile(dst)
				require.NoError(t, readErr, "spec.json should exist in output dir")
				if tt.wantJSON {
					assert.True(t, json.Valid(data), "spec.json should be valid JSON, got: %s", string(data))
				}
			} else if !tt.wantError {
				_, readErr := os.ReadFile(dst)
				assert.True(t, os.IsNotExist(readErr), "spec.json should not exist")
			}
		})
	}
}

func TestCopySpecToOutput_RemoteURL(t *testing.T) {
	// Test with a local HTTP server to verify remote URL handling
	specJSON := `{"openapi":"3.0.0","info":{"title":"Test"}}`

	ts := httpTestServer(t, specJSON)
	defer ts.Close()

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	err := copySpecToOutput("--spec", ts.URL+"/spec.json", outputDir)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(outputDir, "spec.json"))
	require.NoError(t, err)
	assert.True(t, json.Valid(data))
	assert.Contains(t, string(data), "Test")
}

func TestCopySpecToOutput_RemoteYAML(t *testing.T) {
	specYAML := "openapi: \"3.0.0\"\ninfo:\n  title: RemoteYAML\n"

	ts := httpTestServer(t, specYAML)
	defer ts.Close()

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	err := copySpecToOutput("--spec", ts.URL+"/spec.yaml", outputDir)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(outputDir, "spec.json"))
	require.NoError(t, err)
	assert.True(t, json.Valid(data), "remote YAML should be converted to JSON")
	assert.Contains(t, string(data), "RemoteYAML")
}

func httpTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
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
