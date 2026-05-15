package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

	t.Run("openapi spec.json emits valid tools-manifest", func(t *testing.T) {
		dir := setupFakeCLI(t, "openapi3", "spec.json", openAPIFixture)

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
		// Verify spec path resolved to spec.json (not yaml — even
		// though yaml is tried first, only json exists here).
		assert.True(t, strings.HasSuffix(result["spec"].(string), "spec.json"))

		path := filepath.Join(dir, pipeline.ToolsManifestFilename)
		var manifest pipeline.ToolsManifest
		readJSON(t, path, &manifest)
		assert.NotEmpty(t, manifest.APIName)
		assert.NotEmpty(t, manifest.Tools)
	})

	t.Run("openapi spec.yaml is discovered when spec.json is absent", func(t *testing.T) {
		// Regression test for the v1 bug: hardcoded spec.json default
		// meant CLIs whose spec was archived as spec.yaml were
		// unreachable. spec.yaml comes earlier than spec.json in the
		// candidate list, so the loop must reach it cleanly.
		dir := setupFakeCLI(t, "openapi3", "spec.yaml", openAPIYAMLFixture)

		cmd := newToolsManifestEmitCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--dir", dir, "--json"})

		require.NoError(t, cmd.Execute())

		var result map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &result))
		assert.True(t, strings.HasSuffix(result["spec"].(string), "spec.yaml"),
			"expected spec.yaml in resolved path, got %q", result["spec"])
	})

	t.Run("graphql schema.graphql parses and emits", func(t *testing.T) {
		dir := setupFakeCLI(t, "graphql", "schema.graphql", graphQLFixture)

		cmd := newToolsManifestEmitCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--dir", dir, "--json"})

		require.NoError(t, cmd.Execute())

		var result map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &result))
		assert.Equal(t, "graphql", result["spec_format"])
		assert.True(t, strings.HasSuffix(result["spec"].(string), "schema.graphql"))

		path := filepath.Join(dir, pipeline.ToolsManifestFilename)
		var manifest pipeline.ToolsManifest
		readJSON(t, path, &manifest)
		assert.NotEmpty(t, manifest.Tools)
	})

	t.Run("internal spec.json parses and emits", func(t *testing.T) {
		dir := setupFakeCLI(t, "internal", "spec.json", internalSpecFixture)

		cmd := newToolsManifestEmitCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--dir", dir, "--json"})

		require.NoError(t, cmd.Execute())

		var result map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &result))
		assert.Equal(t, "internal", result["spec_format"])

		path := filepath.Join(dir, pipeline.ToolsManifestFilename)
		var manifest pipeline.ToolsManifest
		readJSON(t, path, &manifest)
		assert.NotEmpty(t, manifest.Tools)
	})

	t.Run("--spec override points at non-default path", func(t *testing.T) {
		// Create a CLI dir with no archived spec at all, then point
		// --spec at a fixture elsewhere on disk.
		dir := setupFakeCLI(t, "openapi3", "", "")
		specFile := filepath.Join(t.TempDir(), "elsewhere.json")
		require.NoError(t, os.WriteFile(specFile, []byte(openAPIFixture), 0o644))

		cmd := newToolsManifestEmitCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"--dir", dir, "--spec", specFile, "--json"})

		require.NoError(t, cmd.Execute())

		var result map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &result))
		assert.Equal(t, specFile, result["spec"])
	})

	t.Run("missing archived spec without --spec errors clearly", func(t *testing.T) {
		dir := setupFakeCLI(t, "openapi3", "", "")

		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no archived spec")
	})

	t.Run("existing file blocks without --force", func(t *testing.T) {
		dir := setupFakeCLI(t, "openapi3", "spec.json", openAPIFixture)
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
		dir := setupFakeCLI(t, "openapi3", "spec.json", openAPIFixture)
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

	t.Run("two --force runs produce byte-identical output", func(t *testing.T) {
		// Substantiates the doc string's idempotency claim.
		dir := setupFakeCLI(t, "openapi3", "spec.json", openAPIFixture)
		path := filepath.Join(dir, pipeline.ToolsManifestFilename)

		runEmit := func() string {
			cmd := newToolsManifestEmitCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{"--dir", dir, "--force"})
			require.NoError(t, cmd.Execute())
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			sum := sha256.Sum256(data)
			return hex.EncodeToString(sum[:])
		}

		first := runEmit()
		second := runEmit()
		assert.Equal(t, first, second, "two emit runs against the same spec must be byte-identical")
	})

	t.Run("unsupported spec_format rejected", func(t *testing.T) {
		dir := setupFakeCLI(t, "openapi3", "spec.json", openAPIFixture)
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

	t.Run("docs spec_format rejected with helpful guidance", func(t *testing.T) {
		// `printing-press generate --docs <url>` writes SpecFormat="docs".
		// Those CLIs have no structured spec to re-parse; the error must
		// say that explicitly rather than the generic "unsupported".
		dir := setupFakeCLI(t, "docs", "", "")

		cmd := newToolsManifestEmitCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--dir", dir})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "documentation URL")
		assert.NotContains(t, err.Error(), "unsupported spec_format",
			"docs format must hit the dedicated error arm, not the generic fallback")
	})
}

// setupFakeCLI writes a minimal printed-CLI directory shape: a
// .printing-press.json with the given spec_format, plus the spec file
// at the given name (skip the spec write by passing name=""). Returns
// the directory path.
func setupFakeCLI(t *testing.T, specFormat, specName, specBody string) string {
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
		SpecFormat:           specFormat,
	}
	require.NoError(t, pipeline.WriteCLIManifest(dir, manifest))

	if specName != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, specName), []byte(specBody), 0o644))
	}
	return dir
}

func readJSON(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, out))
}

const openAPIFixture = `{
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

const openAPIYAMLFixture = `openapi: 3.0.0
info:
  title: Demo
  version: 1.0.0
servers:
  - url: https://api.demo.example
paths:
  /widgets:
    get:
      operationId: listWidgets
      summary: List widgets
      responses:
        "200":
          description: OK
`

const graphQLFixture = `
type Query {
  widgets: [Widget!]!
  widget(id: ID!): Widget
}

type Widget {
  id: ID!
  name: String!
}
`

const internalSpecFixture = `{
"name": "demo",
"display_name": "Demo",
"description": "Demo internal spec for tests.",
"version": "1.0.0",
"base_url": "https://api.demo.example",
"auth": {
  "type": "bearer_token",
  "header": "Authorization",
  "format": "Bearer {DEMO_API_KEY}",
  "env_vars": ["DEMO_API_KEY"]
},
"resources": {
  "widgets": {
    "endpoints": {
      "list": {
        "method": "GET",
        "path": "/widgets",
        "description": "List widgets."
      }
    }
  }
}
}`
