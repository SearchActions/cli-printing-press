package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAugmentREADME_WithMarkers(t *testing.T) {
	dir := t.TempDir()
	evidenceDir := filepath.Join(dir, "evidence")
	require.NoError(t, os.MkdirAll(evidenceDir, 0o755))

	readmePath := filepath.Join(dir, "README.md")
	readmeContent := "# My CLI\n\n<!-- HELP_OUTPUT -->\n\nSome other text.\n"
	require.NoError(t, os.WriteFile(readmePath, []byte(readmeContent), 0o644))

	helpOutput := "Usage: mycli [command]\n\nAvailable commands:\n  list    List items\n  get     Get an item\n"
	require.NoError(t, os.WriteFile(filepath.Join(evidenceDir, "tier1-help.txt"), []byte(helpOutput), 0o644))

	err := AugmentREADME(readmePath, evidenceDir)
	require.NoError(t, err)

	result, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(result), "Usage: mycli [command]")
	assert.NotContains(t, string(result), "<!-- HELP_OUTPUT -->")
}

func TestAugmentREADME_NoEvidence(t *testing.T) {
	dir := t.TempDir()
	evidenceDir := filepath.Join(dir, "evidence-missing")
	// Do not create the evidence dir

	readmePath := filepath.Join(dir, "README.md")
	original := "# My CLI\n\nNothing to see here.\n"
	require.NoError(t, os.WriteFile(readmePath, []byte(original), 0o644))

	err := AugmentREADME(readmePath, evidenceDir)
	require.NoError(t, err)

	result, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Equal(t, original, string(result))
}

func TestAugmentREADME_AppendMode(t *testing.T) {
	dir := t.TempDir()
	evidenceDir := filepath.Join(dir, "evidence")
	require.NoError(t, os.MkdirAll(evidenceDir, 0o755))

	readmePath := filepath.Join(dir, "README.md")
	original := "# My CLI\n\nA simple tool.\n"
	require.NoError(t, os.WriteFile(readmePath, []byte(original), 0o644))

	helpOutput := "Usage: mycli [command]\n"
	require.NoError(t, os.WriteFile(filepath.Join(evidenceDir, "tier1-help.txt"), []byte(helpOutput), 0o644))

	err := AugmentREADME(readmePath, evidenceDir)
	require.NoError(t, err)

	result, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	assert.Contains(t, string(result), "## Real Usage Examples")
	assert.Contains(t, string(result), "Usage: mycli [command]")
	assert.Contains(t, string(result), "# My CLI")
}
