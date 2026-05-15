package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolsManifestEmit(t *testing.T) {
	t.Run("missing dir flag errors", func(t *testing.T) {
		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(nil)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dir")
	})

	t.Run("non-printed directory rejected", func(t *testing.T) {
		dir := t.TempDir()
		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("openapi spec emits valid tools-manifest.json", func(t *testing.T) {
		dir := setupFakeOpenAPICLI(t)

		cmd := newToolsManifestEmitCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--dir", dir, "--json"})

		require.NoError(t, cmd.Execute())

		var result map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &result))
		assert.Equal(t, true, result["passed"])
		assert.Equal(t, "openapi3", result["spec_format"])

		// File should exist at the expected path with a non-trivial body.
		path := filepath.Join(dir, pipeline.ToolsManifestFilename)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		var manifest pipeline.ToolsManifest
		require.NoError(t, json.Unmarshal(data, &manifest))
		assert.NotEmpty(t, manifest.APIName)
		assert.NotEmpty(t, manifest.Tools, "tools list should not be empty")
	})

	t.Run("existing file blocks without --force", func(t *testing.T) {
		dir := setupFakeOpenAPICLI(t)
		require.NoError(t, os.WriteFile(filepath.Join(dir, pipeline.ToolsManifestFilename), []byte("{}"), 0o644))

		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--force")
	})

	t.Run("existing file overwritten with --force", func(t *testing.T) {
		dir := setupFakeOpenAPICLI(t)
		path := filepath.Join(dir, pipeline.ToolsManifestFilename)
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))

		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir, "--force"})

		require.NoError(t, cmd.Execute())
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotEqual(t, "{}", strings.TrimSpace(string(data)))
	})

	t.Run("unsupported spec_format rejected", func(t *testing.T) {
		dir := setupFakeOpenAPICLI(t)
		// Rewrite spec_format to an unsupported value.
		manifest, err := pipeline.ReadCLIManifest(dir)
		require.NoError(t, err)
		manifest.SpecFormat = "asyncapi"
		require.NoError(t, pipeline.WriteCLIManifest(dir, manifest))

		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir, "--force"})

		err = cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported spec_format")
	})
}

// setupFakeOpenAPICLI writes a minimal but valid printed-CLI directory
// shape — .printing-press.json + spec.json — sufficient for emit to
// parse and produce a tools-manifest.json.
func setupFakeOpenAPICLI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := pipeline.CLIManifest{
		SchemaVersion:        1,
		APIName:              "demo",
		CLIName:              "demo-pp-cli",
		PrintingPressVersion: "test",
		Printer:              "test",
		PrinterName:          "test",
		RunID:                "test-run",
		SpecFormat:           "openapi3",
	}
	require.NoError(t, pipeline.WriteCLIManifest(dir, manifest))

	spec := `{
"openapi": "3.0.0",
"info": {"title": "Demo", "version": "1.0.0"},
"servers": [{"url": "https://api.demo.example"}],
"paths": {
  "/widgets": {
    "get": {
      "operationId": "listWidgets",
      "summary": "List widgets",
      "responses": {"200": {"description": "OK"}}
    }
  }
}
}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "spec.json"), []byte(spec), 0o644))
	return dir
}
