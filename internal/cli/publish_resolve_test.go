package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRunDir creates a manuscripts/<key>/<runID>/research/ structure with a dummy file.
func createRunDir(t *testing.T, msRoot, key, runID string) {
	t.Helper()
	dir := filepath.Join(msRoot, key, runID, "research")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "brief.md"), []byte("test"), 0o644))
}

func TestResolveManuscriptDir(t *testing.T) {
	t.Run("exact match on API name", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam-web", "20260401-120000")

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.Equal(t, filepath.Join(msRoot, "steam-web"), dir)
		assert.Equal(t, "20260401-120000", runID)
	})

	t.Run("suffix strip: steam-web from steam-web-api", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam-web", "20260401-120000")

		dir, runID := resolveManuscriptDir(msRoot, "steam-web-api")
		assert.Equal(t, filepath.Join(msRoot, "steam-web"), dir)
		assert.Equal(t, "20260401-120000", runID)
	})

	t.Run("suffix strip: steam from steam-web", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260401-120000")

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.Equal(t, filepath.Join(msRoot, "steam"), dir)
		assert.Equal(t, "20260401-120000", runID)
	})

	t.Run("prefix match: steam dir matches steam-web lookup", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260401-120000")

		// steam-web-service doesn't strip to "steam" directly,
		// but "steam" is a prefix of "steam-web-service"
		dir, runID := resolveManuscriptDir(msRoot, "steam-web-service")
		assert.Equal(t, filepath.Join(msRoot, "steam"), dir)
		assert.Equal(t, "20260401-120000", runID)
	})

	t.Run("no match at all", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "notion", "20260401-120000")

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.Empty(t, dir)
		assert.Empty(t, runID)
	})

	t.Run("empty manuscripts root", func(t *testing.T) {
		msRoot := t.TempDir()

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.Empty(t, dir)
		assert.Empty(t, runID)
	})

	t.Run("directory exists but no runs (empty)", func(t *testing.T) {
		msRoot := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(msRoot, "steam"), 0o755))

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		// Directory exists but findMostRecentRun returns "" — no match
		assert.Empty(t, dir)
		assert.Empty(t, runID)
	})

	t.Run("ADVERSARIAL: prefix collision — steam should NOT match steamgames", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260401-120000")

		// "steamgames" is NOT prefixed by "steam-" — no hyphen boundary
		dir, runID := resolveManuscriptDir(msRoot, "steamgames")
		assert.Empty(t, dir, "steam should NOT match steamgames (no hyphen boundary)")
		assert.Empty(t, runID)
	})

	t.Run("prefix WITH hyphen boundary works: steam matches steam-web", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260401-120000")

		// "steam-web" IS prefixed by "steam-" — hyphen boundary present
		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.NotEmpty(t, dir, "steam should match steam-web (hyphen boundary)")
		assert.Equal(t, "20260401-120000", runID)
	})

	t.Run("ADVERSARIAL: multiple candidates — picks first alphabetically", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260401-120000")
		createRunDir(t, msRoot, "steam-web", "20260401-130000")

		// With API name "steam-web-api", suffix strip finds "steam-web" first
		dir, runID := resolveManuscriptDir(msRoot, "steam-web-api")
		assert.Equal(t, filepath.Join(msRoot, "steam-web"), dir)
		assert.Equal(t, "20260401-130000", runID)
	})

	t.Run("picks most recent run when multiple exist", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "20260331-120000")
		createRunDir(t, msRoot, "steam", "20260401-120000")
		createRunDir(t, msRoot, "steam", "20260330-120000")

		dir, runID := resolveManuscriptDir(msRoot, "steam-web")
		assert.Equal(t, filepath.Join(msRoot, "steam"), dir)
		assert.Equal(t, "20260401-120000", runID) // most recent by lexicographic sort
	})
}

func TestManuscriptLookupPriority(t *testing.T) {
	// Simulates the full lookup chain: CLI name → API name → fuzzy resolve
	// This is what the publish package command does.

	t.Run("prefers CLI name over API name", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam-web-pp-cli", "run-cli")
		createRunDir(t, msRoot, "steam-web", "run-api")
		createRunDir(t, msRoot, "steam", "run-slug")

		cliName := "steam-web-pp-cli"

		// Step 1: CLI name
		cliDir := filepath.Join(msRoot, cliName)
		runID, err := findMostRecentRun(cliDir)
		assert.NoError(t, err)
		assert.Equal(t, "run-cli", runID) // CLI name wins
	})

	t.Run("falls back to API name when CLI name missing", func(t *testing.T) {
		msRoot := t.TempDir()
		// No CLI name directory
		createRunDir(t, msRoot, "steam-web", "run-api")
		createRunDir(t, msRoot, "steam", "run-slug")

		cliName := "steam-web-pp-cli"

		// Step 1: CLI name — fails
		cliDir := filepath.Join(msRoot, cliName)
		cliRunID, _ := findMostRecentRun(cliDir)
		assert.Empty(t, cliRunID)

		// Step 2: API name
		apiDir := filepath.Join(msRoot, "steam-web")
		apiRunID, err := findMostRecentRun(apiDir)
		assert.NoError(t, err)
		assert.Equal(t, "run-api", apiRunID) // API name is second priority
	})

	t.Run("falls back to fuzzy when both CLI and API names missing", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "steam", "run-slug")

		apiName := "steam-web"

		// Steps 1+2 fail, step 3: fuzzy resolve
		dir, runID := resolveManuscriptDir(msRoot, apiName)
		assert.Equal(t, filepath.Join(msRoot, "steam"), dir)
		assert.Equal(t, "run-slug", runID)
	})

	t.Run("returns empty when nothing matches", func(t *testing.T) {
		msRoot := t.TempDir()
		createRunDir(t, msRoot, "notion", "run-notion")

		cliName := "steam-web-pp-cli"
		apiName := "steam-web"

		cliDir := filepath.Join(msRoot, cliName)
		runID, _ := findMostRecentRun(cliDir)
		assert.Empty(t, runID)

		apiDir := filepath.Join(msRoot, apiName)
		runID, _ = findMostRecentRun(apiDir)
		assert.Empty(t, runID)

		_, runID = resolveManuscriptDir(msRoot, apiName)
		assert.Empty(t, runID)
	})
}
