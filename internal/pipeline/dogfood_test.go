package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDogfood(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "cli"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "client"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "store"), 0o755))

	writeTestFile(t, filepath.Join(dir, "internal", "cli", "root.go"), `package cli
type rootFlags struct {
	jsonOutput bool
	csvOutput bool
	stdinInput bool
}
func initFlags(flags *rootFlags) {
	_ = &flags.jsonOutput
	_ = &flags.csvOutput
	_ = &flags.stdinInput
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "cli", "helpers.go"), `package cli
func usedHelper() {}
func deadHelper() {}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "cli", "users_list.go"), `package cli
func usersList() {
	path := "/users/123"
	flags.jsonOutput = true
	usedHelper()
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "cli", "projects_get.go"), `package cli
func projectsGet() {
	path := "/bogus"
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "cli", "sync.go"), `package cli
func runSync(s interface{ UpsertUsers() error }) error {
	return s.UpsertUsers()
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "cli", "search.go"), `package cli
func runSearch(s interface{ SearchUsers() error }) error {
	return s.SearchUsers()
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "client", "client.go"), `package client
func authHeader(token string) string {
	return "Bearer " + token
}
`)
	writeTestFile(t, filepath.Join(dir, "internal", "store", "store.go"), "package store\n"+
		"func schema() string {\n"+
		"\treturn `\n"+
		"\t\tCREATE TABLE IF NOT EXISTS users (\n"+
		"\t\t\tid TEXT PRIMARY KEY,\n"+
		"\t\t\tname TEXT NOT NULL,\n"+
		"\t\t\temail TEXT,\n"+
		"\t\t\tdata JSON NOT NULL\n"+
		"\t\t);\n"+
		"\t\tCREATE TABLE IF NOT EXISTS sync_state (\n"+
		"\t\t\tentity_type TEXT PRIMARY KEY,\n"+
		"\t\t\tlast_sync_at TEXT NOT NULL,\n"+
		"\t\t\tcursor TEXT\n"+
		"\t\t);\n"+
		"\t`\n"+
		"}\n")

	specPath := filepath.Join(dir, "spec.json")
	writeTestFile(t, specPath, `{
  "paths": {
    "/users/{id}": {},
    "/projects/{id}": {}
  },
  "components": {
    "securitySchemes": {
      "BotToken": {
        "type": "http",
        "scheme": "bearer"
      }
    }
  }
}`)

	report, err := RunDogfood(dir, specPath)
	require.NoError(t, err)

	assert.Equal(t, "FAIL", report.Verdict)
	assert.Equal(t, 2, report.PathCheck.Tested)
	assert.Equal(t, 1, report.PathCheck.Valid)
	assert.Equal(t, 50, report.PathCheck.Pct)
	assert.Equal(t, []string{"/bogus"}, report.PathCheck.Invalid)
	assert.False(t, report.AuthCheck.Match)
	assert.Equal(t, 3, report.DeadFlags.Total)
	assert.Equal(t, 2, report.DeadFlags.Dead)
	assert.Equal(t, []string{"csvOutput", "stdinInput"}, report.DeadFlags.Items)
	assert.Equal(t, 2, report.DeadFuncs.Total)
	assert.Equal(t, 1, report.DeadFuncs.Dead)
	assert.Equal(t, []string{"deadHelper"}, report.DeadFuncs.Items)
	assert.True(t, report.PipelineCheck.SyncCallsDomain)
	assert.True(t, report.PipelineCheck.SearchCallsDomain)
	assert.Equal(t, 1, report.PipelineCheck.DomainTables)

	loaded, err := LoadDogfoodResults(dir)
	require.NoError(t, err)
	assert.Equal(t, report.Verdict, loaded.Verdict)
}

func TestCountDomainTables(t *testing.T) {
	storeSource := `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT,
	data JSON NOT NULL
);

CREATE TABLE IF NOT EXISTS sync_state (
	entity_type TEXT PRIMARY KEY,
	last_sync_at TEXT NOT NULL,
	cursor TEXT
);
`
	assert.Equal(t, 1, countDomainTables(storeSource))
}

func TestDeriveDogfoodVerdict(t *testing.T) {
	report := &DogfoodReport{
		PathCheck:     PathCheckResult{Tested: 10, Valid: 10, Pct: 100},
		AuthCheck:     AuthCheckResult{Match: true},
		DeadFlags:     DeadCodeResult{Dead: 1},
		DeadFuncs:     DeadCodeResult{Dead: 0},
		PipelineCheck: PipelineResult{SyncCallsDomain: true},
	}
	assert.Equal(t, "WARN", deriveDogfoodVerdict(report, true))

	report.DeadFlags.Dead = 0
	report.DeadFuncs.Dead = 1
	assert.Equal(t, "WARN", deriveDogfoodVerdict(report, true))

	report.DeadFuncs.Dead = 0
	report.PipelineCheck.SyncCallsDomain = false
	assert.Equal(t, "WARN", deriveDogfoodVerdict(report, true))

	report.PipelineCheck.SyncCallsDomain = true
	assert.Equal(t, "PASS", deriveDogfoodVerdict(report, true))
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
