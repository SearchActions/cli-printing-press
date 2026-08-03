package mcpspec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
)

// fastMCPCatalog mirrors what a Pydantic-backed MCP server actually advertises:
// every optional parameter is an anyOf union with a null branch and no
// top-level type, and properties carry `title` rather than `description`.
const fastMCPCatalog = `{
  "stdio": {"command": "uvx", "args": ["example-mcp"], "env": ["EXAMPLE_CLI_PATH"], "ready_tool": "check"},
  "server_info": {"name": "example-mcp", "version": "0.4.0"},
  "tools": [
    {
      "name": "read_crawl_data",
      "description": "Read CSV data from an export.",
      "annotations": {"readOnlyHint": true, "openWorldHint": false},
      "inputSchema": {
        "type": "object",
        "properties": {
          "export_id": {"title": "Export Id", "type": "string"},
          "limit": {"title": "Limit", "type": "integer", "default": 100},
          "filter_column": {"anyOf": [{"type": "string"}, {"type": "null"}], "default": null, "title": "Filter Column"},
          "filter_value": {"anyOf": [{"type": "string"}, {"type": "integer"}, {"type": "number"}, {"type": "null"}], "default": null},
          "tags": {"anyOf": [{"type": "array", "items": {"anyOf": [{"type": "string"}, {"type": "null"}]}}, {"type": "null"}]}
        },
        "required": ["export_id"]
      }
    },
    {
      "name": "delete_crawl",
      "description": "Delete a stored crawl.",
      "annotations": {"readOnlyHint": false, "destructiveHint": true},
      "inputSchema": {
        "type": "object",
        "properties": {"db_id": {"type": "string"}},
        "required": ["db_id"]
      }
    }
  ]
}`

func paramByName(t *testing.T, params []spec.Param, name string) spec.Param {
	t.Helper()
	for _, p := range params {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("no param %q in %v", name, params)
	return spec.Param{}
}

func findEndpoint(t *testing.T, s *spec.APISpec, tool string) spec.Endpoint {
	t.Helper()
	for _, res := range s.Resources {
		for _, ep := range res.Endpoints {
			if ep.Meta["pp:mcp-tool"] == tool {
				return ep
			}
		}
	}
	t.Fatalf("no endpoint for tool %q", tool)
	return spec.Endpoint{}
}

// A nullable union is how Pydantic spells "optional", so a parser that only
// reads the top-level `type` sees most Python MCP params as untyped and emits
// every one of them as a string flag.
func TestParseCollapsesNullableUnionsToTheTypedBranch(t *testing.T) {
	parsed, err := Parse("example-mcp", []byte(fastMCPCatalog), ParseOptions{})
	require.NoError(t, err)

	ep := findEndpoint(t, parsed, "read_crawl_data")
	params := ep.Params

	assert.Equal(t, "integer", paramByName(t, params, "limit").Type,
		"a plainly typed param is unaffected")
	assert.Equal(t, "string", paramByName(t, params, "filter_column").Type,
		"anyOf[string,null] is a string flag, not an untyped one")
	assert.Equal(t, "string", paramByName(t, params, "filter_value").Type,
		"a multi-type union collapses to its first non-null branch: a flag carries one typed value")

	tags := paramByName(t, params, "tags")
	assert.Equal(t, "array", tags.Type, "the array branch wins over the null branch")
	assert.Equal(t, "string", tags.ItemType, "a nullable item type resolves too")
}

// The launch command is the only way a printed CLI can reach a stdio server, so
// it has to survive the capture -> spec hop intact.
func TestParseRecordsTheStdioLaunchAndDropsTheHTTPOrigin(t *testing.T) {
	parsed, err := Parse("example-mcp", []byte(fastMCPCatalog), ParseOptions{})
	require.NoError(t, err)

	require.True(t, parsed.IsMCPStdioSource())
	assert.Equal(t, "uvx", parsed.MCPStdio.Command)
	assert.Equal(t, []string{"example-mcp"}, parsed.MCPStdio.Args)
	assert.Equal(t, []string{"EXAMPLE_CLI_PATH"}, parsed.MCPStdio.Env)
	assert.Equal(t, "check", parsed.MCPStdio.Ready)

	assert.Empty(t, parsed.BaseURL, "a subprocess has no HTTP origin")
	assert.False(t, parsed.BaseURLIsPlaceholder,
		"an empty base URL here is correct, not a placeholder generate should reject")
	assert.NotEmpty(t, parsed.MCPEndpointPath,
		"the synthetic route prefix still anchors every tool's path")

	// A stdio server runs as the operator; demanding a token would make doctor
	// require an env var nothing can satisfy.
	assert.Equal(t, "none", parsed.Auth.Type)

	require.NoError(t, parsed.Validate(), "the parsed spec must survive validation")
}

// Semantic verbs drive the generated MCP safety annotations, and the stdio
// transport must not disturb them.
func TestStdioParseKeepsSemanticVerbs(t *testing.T) {
	parsed, err := Parse("example-mcp", []byte(fastMCPCatalog), ParseOptions{})
	require.NoError(t, err)

	assert.Equal(t, "GET", findEndpoint(t, parsed, "read_crawl_data").Method)
	assert.Equal(t, "DELETE", findEndpoint(t, parsed, "delete_crawl").Method)
}

// A capture with both an HTTP origin and a launch command would emit a CLI with
// a base URL it silently ignores.
func TestStdioAndBaseURLAreMutuallyExclusive(t *testing.T) {
	var doc map[string]any
	require.NoError(t, json.Unmarshal([]byte(fastMCPCatalog), &doc))

	parsed, err := Parse("example-mcp", []byte(fastMCPCatalog), ParseOptions{})
	require.NoError(t, err)
	parsed.BaseURL = "https://mcp.example.com"

	err = parsed.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// --name is the operator's slug choice. Dropped from the capture, `generate`
// re-derives from the server's advertised name, which is frequently a sentence.
func TestParsePrefersTheRecordedNameOverTheServerAdvertisedOne(t *testing.T) {
	catalogWithName := `{
  "name": "example",
  "stdio": {"command": "uvx", "args": ["example-mcp"]},
  "server_info": {"name": "Example SEO Tool MCP Server (headless)"},
  "tools": [{"name": "list_records", "description": "List.", "inputSchema": {"type": "object", "properties": {}}}]
}`
	parsed, err := Parse("example-mcp", []byte(catalogWithName), ParseOptions{})
	require.NoError(t, err)
	assert.Equal(t, "example", parsed.Name)

	// An explicit option still wins over the recorded value.
	override, err := Parse("example-mcp", []byte(catalogWithName), ParseOptions{Name: "override"})
	require.NoError(t, err)
	assert.Equal(t, "override", override.Name)
}
