package generator

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mcpBackedSpec builds a spec whose upstream is an MCP server: one read tool
// and one destructive tool, each addressed by a synthetic path under
// MCPEndpointPath and tagged with the tool it maps to.
func mcpBackedSpec(name string) *spec.APISpec {
	apiSpec := minimalSpec(name)
	apiSpec.MCPEndpointPath = "/v1/tools/mcp"
	apiSpec.Resources = map[string]spec.Resource{
		"widgets": {
			Description: "Widgets",
			Endpoints: map[string]spec.Endpoint{
				"list": {
					Method:      "GET",
					Path:        "/v1/tools/mcp/list_widgets",
					Description: "List widgets",
					Meta:        map[string]string{"pp:mcp-tool": "list_widgets", "mcp:read-only": "true"},
					Params: []spec.Param{
						{Name: "limit", Type: "integer"},
					},
				},
				"delete": {
					Method:      "DELETE",
					Path:        "/v1/tools/mcp/delete_widget",
					Description: "Delete a widget",
					Meta:        map[string]string{"pp:mcp-tool": "delete_widget"},
					Body: []spec.Param{
						{Name: "widgetId", Type: "string", Required: true},
					},
				},
			},
		},
	}
	return apiSpec
}

func TestGeneratedMCPTransportCompilesAndWiresEveryTool(t *testing.T) {
	t.Parallel()

	apiSpec := mcpBackedSpec("mcpemit")
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())
	// Compile rather than only string-match: the transport spans two emitted
	// files (client.go calls into mcp.go), so a cross-file contract break is
	// invisible to a Contains assertion.
	requireGeneratedCompiles(t, outputDir)

	mcpGo, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "mcp.go"))
	require.NoError(t, err)
	// gofmt aligns map values, and the padding shifts whenever a sibling key
	// changes length. Collapse space runs so these assertions test the emitted
	// contract rather than the current alignment.
	got := collapseSpaces(string(mcpGo))

	// Every endpoint's synthetic path must resolve back to its tool, or the
	// command silently escapes the JSON-RPC envelope.
	assert.Contains(t, got, `"/v1/tools/mcp/list_widgets": "list_widgets"`)
	assert.Contains(t, got, `"/v1/tools/mcp/delete_widget": "delete_widget"`)

	// Declared argument types drive coercion; without them a schema-declared
	// integer ships as the string "10".
	assert.Contains(t, got, `"limit": "integer"`)

	// The lifecycle notification is required before any other request.
	assert.Contains(t, got, "notifications/initialized")
}

func TestGeneratedClientRoutesMCPPathsThroughTheTransport(t *testing.T) {
	t.Parallel()

	apiSpec := mcpBackedSpec("mcproute")
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	clientGo, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "client.go"))
	require.NoError(t, err)
	got := string(clientGo)

	// Both transports must intercept: do() for mutations, doRead() for reads.
	// A read routed through do() would trip the verify-mode mutation gate.
	assert.Contains(t, got, "c.mcpDispatch(ctx, tool, method, mcpParams, body, headerOverrides, false)")
	assert.Contains(t, got, "c.mcpDispatch(ctx, tool, method, mcpParams, body, headerOverrides, true)")

	// Resolution is query-aware: pathWithQueryValues folds params into the
	// path, and an exact-match lookup would miss those and escape as plain HTTP.
	assert.Contains(t, got, "mcpToolForRequest(path, params)")

	// Reads that ride a mutating wire verb must not wipe the response cache.
	assert.Contains(t, got, "if method != http.MethodGet && !readOnlyIntent && !c.DryRun {")
}

func TestNonMCPSpecEmitsNoMCPTransport(t *testing.T) {
	t.Parallel()

	// A REST spec must regenerate byte-identically to what it produced before
	// MCP intake existed: no mcp.go, and no dispatch calls in client.go that
	// would reference helpers that were never emitted.
	apiSpec := minimalSpec("plainrest")
	apiSpec.Resources = map[string]spec.Resource{
		"widgets": {
			Description: "Widgets",
			Endpoints: map[string]spec.Endpoint{
				"list": {Method: "GET", Path: "/widgets", Description: "List widgets"},
			},
		},
	}
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())
	requireGeneratedCompiles(t, outputDir)

	_, err := os.Stat(filepath.Join(outputDir, "internal", "client", "mcp.go"))
	assert.True(t, os.IsNotExist(err), "a REST spec must not emit the MCP transport")

	clientGo, err := os.ReadFile(filepath.Join(outputDir, "internal", "client", "client.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(clientGo), "mcpToolForRequest",
		"MCP dispatch must not be emitted when no MCP transport file exists")
}

func TestWhitespaceOnlyMCPEndpointPathIsNormalized(t *testing.T) {
	t.Parallel()

	// isMCPSpec trims, but the client template gates on {{if .MCPEndpointPath}},
	// which treats " " as true — emitting dispatch calls while mcp.go is never
	// rendered, producing a CLI that does not compile.
	apiSpec := minimalSpec("wsmcp")
	apiSpec.MCPEndpointPath = "   "
	apiSpec.Resources = map[string]spec.Resource{
		"widgets": {
			Description: "Widgets",
			Endpoints: map[string]spec.Endpoint{
				"list": {Method: "GET", Path: "/widgets", Description: "List widgets"},
			},
		},
	}
	require.NoError(t, apiSpec.Validate())
	assert.Empty(t, apiSpec.MCPEndpointPath, "whitespace-only path normalizes to empty")

	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())
	requireGeneratedCompiles(t, outputDir)
}

// collapseSpaces reduces runs of spaces to one so assertions on emitted Go do
// not depend on gofmt's column alignment.
func collapseSpaces(s string) string {
	return regexp.MustCompile(` +`).ReplaceAllString(s, " ")
}
