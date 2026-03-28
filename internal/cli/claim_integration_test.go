package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimOrForce_DefaultAutoIncrements(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapi-pp-cli")

	// First call claims the base
	got, err := claimOrForce(base, false, false)
	require.NoError(t, err)
	assert.Equal(t, base, got)

	// Second call auto-increments to -2
	got, err = claimOrForce(base, false, false)
	require.NoError(t, err)
	assert.Equal(t, base+"-2", got)

	// Third call auto-increments to -3
	got, err = claimOrForce(base, false, false)
	require.NoError(t, err)
	assert.Equal(t, base+"-3", got)
}

func TestClaimOrForce_ForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapi-pp-cli")

	// Create base with a file in it
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "main.go"), []byte("package main"), 0o644))

	// Force should blow it away and recreate
	got, err := claimOrForce(base, true, false)
	require.NoError(t, err)
	assert.Equal(t, base, got)

	// Old contents should be gone
	entries, err := os.ReadDir(base)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestClaimOrForce_ExplicitOutputErrorsOnCollision(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapi-pp-cli")

	// Create base with content
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "main.go"), []byte("package main"), 0o644))

	// Explicit output without force should error
	_, err := claimOrForce(base, false, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestClaimOrForce_ExplicitOutputWithForce(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "myapi-pp-cli")

	// Create base with content
	require.NoError(t, os.MkdirAll(base, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "main.go"), []byte("package main"), 0o644))

	// Explicit output with force should succeed
	got, err := claimOrForce(base, true, true)
	require.NoError(t, err)
	assert.Equal(t, base, got)
}
