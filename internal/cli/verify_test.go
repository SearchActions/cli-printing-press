package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupVerifyArtifacts(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sample-cli")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cache", "go-build"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cmd", "library", "sample-cli"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample-cli"), []byte("bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample-cli-validation"), []byte("bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample-cli-dogfood"), []byte("bin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".DS_Store"), []byte("finder"), 0o644))

	require.NoError(t, cleanupVerifyArtifacts(dir, true))

	assert.NoFileExists(t, filepath.Join(dir, "sample-cli"))
	assert.NoFileExists(t, filepath.Join(dir, "sample-cli-validation"))
	assert.NoFileExists(t, filepath.Join(dir, "sample-cli-dogfood"))
	assert.NoFileExists(t, filepath.Join(dir, ".DS_Store"))
	assert.NoDirExists(t, filepath.Join(dir, ".cache"))
	assert.NoDirExists(t, filepath.Join(dir, "cmd", "library"))
}

func TestCleanupVerifyArtifacts_NoOpWhenDisabled(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sample-cli")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sample-cli"), []byte("bin"), 0o755))

	require.NoError(t, cleanupVerifyArtifacts(dir, false))

	assert.FileExists(t, filepath.Join(dir, "sample-cli"))
}
