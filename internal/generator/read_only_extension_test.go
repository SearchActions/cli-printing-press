package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/openapi"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestXPPReadOnlyExtensionEndToEnd parses a spec through the real OpenAPI
// parser (not a hand-built spec.APISpec) so the assertions cover the full
// wire: extensionPPReadOnly -> Endpoint.Meta -> endpointIsWriteCommand ->
// emitted command annotation, client routing, and MCP tool readOnlyHint.
// POST /v1/load carries x-pp-read-only: true on a body the heuristics alone
// would classify as a write (Cube's actual shape); a control POST with a
// genuinely write-shaped body proves the heuristics still fire without the
// extension.
func TestXPPReadOnlyExtensionEndToEnd(t *testing.T) {
	t.Parallel()

	apiSpec, err := openapi.Parse([]byte(`openapi: "3.0.3"
info:
  title: Read Only Extension Test
  version: "1.0.0"
servers:
  - url: https://api.example.com
paths:
  /v1/load:
    post:
      operationId: loadV1
      x-pp-resource: queries
      x-pp-read-only: true
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                queryDefinition:
                  type: object
      responses:
        "200":
          description: OK
  /reports:
    post:
      operationId: createReport
      x-pp-resource: queries
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                definition:
                  type: object
      responses:
        "200":
          description: OK
`))
	require.NoError(t, err)
	apiSpec.Name = "read-only-ext"
	apiSpec.Config = spec.ConfigSpec{Format: "toml", Path: "~/.config/read-only-ext-pp-cli/config.toml"}

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	extBytes, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "queries_load-v1.go"))
	require.NoError(t, err)
	extSrc := string(extBytes)
	assert.Contains(t, extSrc, `"mcp:read-only": "true"`,
		"x-pp-read-only:true endpoint must emit the mcp:read-only annotation")
	assert.Contains(t, extSrc, "PostQuery",
		"x-pp-read-only:true endpoint must route through the PostQuery* (doRead) client method, not the mutating Post path")

	controlBytes, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "queries_create-report.go"))
	require.NoError(t, err)
	controlSrc := string(controlBytes)
	assert.NotContains(t, controlSrc, `"mcp:read-only": "true"`,
		"control write endpoint must NOT emit the mcp:read-only annotation")

	mcpBytes, err := os.ReadFile(filepath.Join(outputDir, "internal", "mcp", "tools.go"))
	require.NoError(t, err)
	mcpSrc := string(mcpBytes)

	loadToolStart := indexOrFatal(t, mcpSrc, `mcplib.NewTool("queries_load-v1",`)
	loadToolEnd := indexOrFatal(t, mcpSrc[loadToolStart:], "),\n\t\tmakeAPIHandler(") + loadToolStart
	assert.Contains(t, mcpSrc[loadToolStart:loadToolEnd], "mcplib.WithReadOnlyHintAnnotation(true)",
		"queries_load-v1 typed MCP tool must carry WithReadOnlyHintAnnotation(true)")

	reportToolStart := indexOrFatal(t, mcpSrc, `mcplib.NewTool("queries_create-report",`)
	reportToolEnd := indexOrFatal(t, mcpSrc[reportToolStart:], "),\n\t\tmakeAPIHandler(") + reportToolStart
	assert.NotContains(t, mcpSrc[reportToolStart:reportToolEnd], "mcplib.WithReadOnlyHintAnnotation(true)",
		"queries_create-report typed MCP tool must NOT carry WithReadOnlyHintAnnotation(true)")

	requireGeneratedCompiles(t, outputDir)
}

func indexOrFatal(t *testing.T, haystack, needle string) int {
	t.Helper()
	idx := strings.Index(haystack, needle)
	require.GreaterOrEqualf(t, idx, 0, "expected to find %q in generated source", needle)
	return idx
}
