package mcpspec

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
)

// designCatalog mirrors the shape of the Claude Design MCP server's tool
// catalog: a namespaced read/write mix with explicit readOnlyHint annotations
// on the reads and a destructive delete.
const designCatalog = `{
  "server_url": "https://api.anthropic.com/v1/design/mcp",
  "server_info": {"name": "claude-design", "title": "Claude Design", "version": "0.1.0"},
  "tools": [
    {
      "name": "list_projects",
      "description": "List design-system projects the user can write to. Returns name, owner, projectId, updatedAt.",
      "annotations": {"readOnlyHint": true},
      "inputSchema": {"type": "object", "properties": {}}
    },
    {
      "name": "get_file",
      "description": "Read one remote file's content. Capped at 256 KiB.",
      "annotations": {"readOnlyHint": true},
      "inputSchema": {
        "type": "object",
        "properties": {
          "projectId": {"type": "string", "description": "Project to read from"},
          "path": {"type": "string", "description": "File path to read", "minLength": 1}
        },
        "required": ["projectId", "path"]
      }
    },
    {
      "name": "create_project",
      "description": "Create a new design-system project owned by the user.",
      "inputSchema": {
        "type": "object",
        "properties": {"name": {"type": "string", "maxLength": 200}},
        "required": ["name"]
      }
    },
    {
      "name": "write_files",
      "description": "Write files to the project. Every path must be in the finalized plan's writes.",
      "inputSchema": {
        "type": "object",
        "properties": {
          "planId": {"type": "string"},
          "projectId": {"type": "string"},
          "files": {"type": "array", "items": {"type": "object"}}
        },
        "required": ["planId", "projectId", "files"]
      }
    },
    {
      "name": "delete_files",
      "description": "Delete files from the project.",
      "annotations": {"destructiveHint": true},
      "inputSchema": {
        "type": "object",
        "properties": {
          "planId": {"type": "string"},
          "paths": {"type": "array", "items": {"type": "string"}}
        },
        "required": ["planId", "paths"]
      }
    },
    {
      "name": "update_project",
      "description": "Rename a project.",
      "inputSchema": {
        "type": "object",
        "properties": {
          "projectId": {"type": "string"},
          "viewport": {
            "type": "object",
            "properties": {
              "width": {"type": "integer"},
              "height": {"type": "integer"}
            },
            "required": ["width"]
          }
        },
        "required": ["projectId"]
      }
    }
  ]
}`

func TestIsMCPToolsList(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{
			name: "sniffed catalog envelope",
			data: designCatalog,
			want: true,
		},
		{
			name: "raw jsonrpc tools/list response",
			data: `{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"ping","inputSchema":{"type":"object"}}]}}`,
			want: true,
		},
		{
			name: "printed CLI tools-manifest.json is not an MCP catalog",
			data: `{"api_name":"acme","base_url":"https://api.acme.test",
			        "tools":[{"name":"currencies_list","description":"List currencies","method":"GET","path":"/currencies","params":[]}]}`,
			want: false,
		},
		{
			name: "manifest-shaped tools without api_name still rejected on method/path",
			data: `{"tools":[{"name":"currencies_list","description":"d","method":"GET","path":"/currencies"}]}`,
			want: false,
		},
		{
			name: "tool without inputSchema rejected",
			data: `{"tools":[{"name":"ping","description":"d"}]}`,
			want: false,
		},
		{
			name: "empty tool list rejected",
			data: `{"tools":[]}`,
			want: false,
		},
		{
			name: "openapi document rejected",
			data: `{"openapi":"3.0.0","info":{"title":"x"},"paths":{}}`,
			want: false,
		},
		{
			name: "yaml internal spec rejected",
			data: "name: acme\nbase_url: https://api.acme.test\n",
			want: false,
		},
		{
			name: "not json",
			data: `garbage`,
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsMCPToolsList([]byte(tc.data)))
		})
	}
}

func TestParseDesignCatalog(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	assert.Equal(t, "claude-design", s.Name)
	assert.Equal(t, "Claude Design", s.Description)
	assert.Equal(t, "https://api.anthropic.com", s.BaseURL)
	assert.Equal(t, "/v1/design/mcp", s.MCPEndpointPath)
	assert.False(t, s.BaseURLIsPlaceholder)
	assert.Empty(t, s.GraphQLEndpointPath, "MCP and GraphQL endpoint paths are mutually exclusive")

	// Resources are the pluralized objects of the tool names.
	require.Contains(t, s.Resources, "projects")
	require.Contains(t, s.Resources, "files")

	projects := s.Resources["projects"]
	require.Contains(t, projects.Endpoints, "list")
	require.Contains(t, projects.Endpoints, "create")
	require.Contains(t, projects.Endpoints, "update")

	files := s.Resources["files"]
	require.Contains(t, files.Endpoints, "get")
	require.Contains(t, files.Endpoints, "write")
	require.Contains(t, files.Endpoints, "delete")
}

func TestParseSemanticMethods(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	// Semantic verbs, not the JSON-RPC transport's POST. These drive the
	// generated MCP tool safety annotations.
	assert.Equal(t, "GET", s.Resources["projects"].Endpoints["list"].Method)
	assert.Equal(t, "GET", s.Resources["files"].Endpoints["get"].Method)
	assert.Equal(t, "POST", s.Resources["projects"].Endpoints["create"].Method)
	assert.Equal(t, "PATCH", s.Resources["projects"].Endpoints["update"].Method)
	assert.Equal(t, "POST", s.Resources["files"].Endpoints["write"].Method)
	assert.Equal(t, "DELETE", s.Resources["files"].Endpoints["delete"].Method)
}

func TestParseEndpointPathsResolveToDistinctTools(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	// Each endpoint gets its own synthetic path under the MCP endpoint. The
	// generated transport maps path -> tool, so a duplicate path would send
	// two different commands to the same tool.
	byPath := map[string]string{}
	for resName, res := range s.Resources {
		for epName, ep := range res.Endpoints {
			tool := ep.Meta[mcpToolAnnotation]
			require.NotEmpty(t, tool,
				"resource %s endpoint %s must record its upstream tool name", resName, epName)
			assert.Equal(t, s.MCPEndpointPath+"/"+tool, ep.Path,
				"resource %s endpoint %s path must address its tool", resName, epName)

			prior, clash := byPath[ep.Path]
			assert.False(t, clash,
				"path %s already maps to tool %s; path->tool lookup must be unique", ep.Path, prior)
			byPath[ep.Path] = tool
		}
	}
	assert.Len(t, byPath, 6, "every tool in the catalog gets an addressable endpoint")
}

func TestParseReadOnlyAnnotation(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	// Reads ride JSON-RPC POST, so they must be marked read-only for the
	// transport to route them through doRead() instead of the mutation gate.
	assert.Equal(t, "true", s.Resources["projects"].Endpoints["list"].Meta[mcpReadOnlyAnnotation])
	assert.Equal(t, "true", s.Resources["files"].Endpoints["get"].Meta[mcpReadOnlyAnnotation])

	// Writes must never carry it: a false readOnlyHint on a mutating tool is
	// a real bug, not a missing permission prompt.
	assert.NotContains(t, s.Resources["files"].Endpoints["write"].Meta, mcpReadOnlyAnnotation)
	assert.NotContains(t, s.Resources["files"].Endpoints["delete"].Meta, mcpReadOnlyAnnotation)
	assert.NotContains(t, s.Resources["projects"].Endpoints["create"].Meta, mcpReadOnlyAnnotation)
}

func TestServerReadOnlyHintOverridesNameHeuristic(t *testing.T) {
	// A tool whose name reads like a mutation but which the server declares
	// read-only, and one whose name reads like a read but which the server
	// declares NOT read-only. The server wins both ways.
	data := `{
	  "server_url": "https://mcp.test/mcp",
	  "tools": [
	    {"name":"send_preview","description":"d","annotations":{"readOnlyHint":true},
	     "inputSchema":{"type":"object","properties":{}}},
	    {"name":"get_and_stamp_record","description":"d","annotations":{"readOnlyHint":false},
	     "inputSchema":{"type":"object","properties":{}}}
	  ]
	}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	previews := s.Resources["previews"]
	require.Contains(t, previews.Endpoints, "send")
	assert.Equal(t, "GET", previews.Endpoints["send"].Method,
		"explicit readOnlyHint:true must beat the write-verb name heuristic")
	assert.Equal(t, "true", previews.Endpoints["send"].Meta[mcpReadOnlyAnnotation])

	records := s.Resources["and-stamp-records"]
	require.Contains(t, records.Endpoints, "get")
	assert.NotEqual(t, "GET", records.Endpoints["get"].Method,
		"explicit readOnlyHint:false must beat the read-verb name heuristic")
	assert.NotContains(t, records.Endpoints["get"].Meta, mcpReadOnlyAnnotation)
}

func TestParseParamsFromInputSchema(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	// A read puts inputs in Params so the emitted command routes via doRead().
	get := s.Resources["files"].Endpoints["get"]
	assert.Empty(t, get.Body, "a read must not carry a request body")
	names := paramNames(get.Params)
	assert.ElementsMatch(t, []string{"projectId", "path"}, names)
	for _, p := range get.Params {
		assert.True(t, p.Required, "both projectId and path are required")
		assert.Equal(t, "string", p.Type)
	}

	// A write puts inputs in Body.
	write := s.Resources["files"].Endpoints["write"]
	assert.Empty(t, write.Params, "a write's tool arguments belong in the body")
	assert.ElementsMatch(t, []string{"planId", "projectId", "files"}, paramNames(write.Body))

	// Array item types survive.
	del := s.Resources["files"].Endpoints["delete"]
	paths := findParam(t, del.Body, "paths")
	assert.Equal(t, "array", paths.Type)
	assert.Equal(t, "string", paths.ItemType)
}

func TestParseNestedObjectBecomesFields(t *testing.T) {
	s, err := Parse("design.json", []byte(designCatalog), ParseOptions{})
	require.NoError(t, err)

	update := s.Resources["projects"].Endpoints["update"]
	viewport := findParam(t, update.Body, "viewport")
	assert.Equal(t, "object", viewport.Type)
	require.Len(t, viewport.Fields, 2, "nested object properties become sub-fields")
	assert.ElementsMatch(t, []string{"width", "height"}, paramNames(viewport.Fields))
	for _, f := range viewport.Fields {
		assert.Equal(t, "integer", f.Type)
		if f.Name == "width" {
			assert.True(t, f.Required, "nested required list is honored")
		} else {
			assert.False(t, f.Required)
		}
	}
}

func TestParseRequiredParamsSortFirst(t *testing.T) {
	data := `{
	  "server_url":"https://mcp.test/mcp",
	  "tools":[{"name":"create_thing","description":"d","inputSchema":{
	    "type":"object",
	    "properties":{"alpha":{"type":"string"},"zulu":{"type":"string"},"mid":{"type":"string"}},
	    "required":["zulu"]
	  }}]
	}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	body := s.Resources["things"].Endpoints["create"].Body
	require.Len(t, body, 3)
	assert.Equal(t, "zulu", body[0].Name, "required params sort first")
	assert.Equal(t, "alpha", body[1].Name, "remaining params sort by name")
	assert.Equal(t, "mid", body[2].Name)
}

func TestParseIsDeterministic(t *testing.T) {
	// Catalog order must not leak into the spec: two shuffles of the same
	// tool set must produce byte-identical specs.
	a := `{"server_url":"https://mcp.test/mcp","tools":[
	  {"name":"get_alpha","inputSchema":{"type":"object","properties":{"b":{"type":"string"},"a":{"type":"string"}}}},
	  {"name":"get_beta","inputSchema":{"type":"object","properties":{}}}]}`
	b := `{"server_url":"https://mcp.test/mcp","tools":[
	  {"name":"get_beta","inputSchema":{"type":"object","properties":{}}},
	  {"name":"get_alpha","inputSchema":{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"string"}}}}]}`

	sa, err := Parse("a.json", []byte(a), ParseOptions{})
	require.NoError(t, err)
	sb, err := Parse("b.json", []byte(b), ParseOptions{})
	require.NoError(t, err)

	ja, err := json.Marshal(sa)
	require.NoError(t, err)
	jb, err := json.Marshal(sb)
	require.NoError(t, err)
	assert.JSONEq(t, string(ja), string(jb))
}

func TestParseStripsCommonNamespacePrefix(t *testing.T) {
	data := `{"server_url":"https://api.fireflies.ai/mcp","tools":[
	  {"name":"fireflies_get_transcript","inputSchema":{"type":"object","properties":{}}},
	  {"name":"fireflies_list_channels","inputSchema":{"type":"object","properties":{}}}]}`
	s, err := Parse("f.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	// The shared "fireflies_" prefix is the server, not a resource.
	assert.Contains(t, s.Resources, "transcripts")
	assert.Contains(t, s.Resources, "channels")
	assert.NotContains(t, s.Resources, "fireflies")
}

func TestVerbIsFoundMidName(t *testing.T) {
	// A leading service qualifier must not be mistaken for the verb: that
	// yields junk resources like "slacks" instead of "messages".
	data := `{"server_url":"https://mcp.test/mcp","tools":[
	  {"name":"slack_send_message","inputSchema":{"type":"object","properties":{}}},
	  {"name":"qbo_sales_create_invoice","inputSchema":{"type":"object","properties":{}}},
	  {"name":"list_users","inputSchema":{"type":"object","properties":{}}}]}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	require.Contains(t, s.Resources, "messages")
	assert.Contains(t, s.Resources["messages"].Endpoints, "send")

	require.Contains(t, s.Resources, "invoices")
	assert.Contains(t, s.Resources["invoices"].Endpoints, "create")

	require.Contains(t, s.Resources, "users")
	assert.Contains(t, s.Resources["users"].Endpoints, "list")

	assert.NotContains(t, s.Resources, "slacks")
	assert.NotContains(t, s.Resources, "qbos")
}

func TestTrailingVerbIsRecognized(t *testing.T) {
	// Ahrefs-style and manifest-style names put the verb last.
	data := `{"server_url":"https://mcp.test/mcp","tools":[
	  {"name":"currencies_list","inputSchema":{"type":"object","properties":{}}},
	  {"name":"invoice_delete","inputSchema":{"type":"object","properties":{}}}]}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	require.Contains(t, s.Resources, "currencies")
	assert.Contains(t, s.Resources["currencies"].Endpoints, "list")
	assert.Equal(t, "GET", s.Resources["currencies"].Endpoints["list"].Method)

	require.Contains(t, s.Resources, "invoices")
	assert.Equal(t, "DELETE", s.Resources["invoices"].Endpoints["delete"].Method)
}

func TestParseVerblessToolFallsBackToNamespace(t *testing.T) {
	data := `{"server_url":"https://mcp.acme.test/mcp","tools":[
	  {"name":"authenticate","inputSchema":{"type":"object","properties":{}}},
	  {"name":"ping","inputSchema":{"type":"object","properties":{}}}]}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	// Verb-only names have no object to become a resource; they must still
	// land somewhere addressable rather than being dropped.
	total := 0
	for _, res := range s.Resources {
		total += len(res.Endpoints)
	}
	assert.Equal(t, 2, total, "no tool may be silently dropped")
}

func TestEndpointNameCollisionsAreDisambiguated(t *testing.T) {
	// Two tools that classify to the same resource+verb must both survive.
	data := `{"server_url":"https://mcp.test/mcp","tools":[
	  {"name":"get_file","inputSchema":{"type":"object","properties":{}}},
	  {"name":"read_file","inputSchema":{"type":"object","properties":{}}}]}`
	s, err := Parse("t.json", []byte(data), ParseOptions{})
	require.NoError(t, err)

	files := s.Resources["files"]
	assert.Len(t, files.Endpoints, 2, "colliding tools must not overwrite each other")

	tools := map[string]bool{}
	for _, ep := range files.Endpoints {
		tools[ep.Meta[mcpToolAnnotation]] = true
	}
	assert.True(t, tools["get_file"])
	assert.True(t, tools["read_file"])
}

func TestParseServerURLOptionOverridesEnvelope(t *testing.T) {
	s, err := Parse("t.json", []byte(designCatalog), ParseOptions{
		ServerURL: "https://self-hosted.test:8443/custom/mcp",
		Name:      "my-design",
	})
	require.NoError(t, err)
	assert.Equal(t, "my-design", s.Name)
	assert.Equal(t, "https://self-hosted.test:8443", s.BaseURL)
	assert.Equal(t, "/custom/mcp", s.MCPEndpointPath)
}

func TestParseWithoutServerURLLeavesPlaceholder(t *testing.T) {
	// generate rejects a placeholder base URL with an actionable message
	// rather than shipping a CLI whose doctor DNS-fails.
	s, err := Parse("t.json", []byte(`{"tools":[{"name":"get_thing","inputSchema":{"type":"object","properties":{}}}]}`), ParseOptions{})
	require.NoError(t, err)
	assert.True(t, s.BaseURLIsPlaceholder)
	assert.Equal(t, "/mcp", s.MCPEndpointPath)
}

func TestParseRejectsRelativeServerURL(t *testing.T) {
	_, err := Parse("t.json", []byte(designCatalog), ParseOptions{ServerURL: "/just/a/path"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")
}

func TestParseRejectsEmptyCatalog(t *testing.T) {
	_, err := Parse("t.json", []byte(`{"tools":[]}`), ParseOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "advertises no tools")
}

func TestNameDerivation(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		want      string
	}{
		{"path segment beats host", "https://api.anthropic.com/v1/design/mcp", "design"},
		{"version segments skipped", "https://api.example.com/v2/mcp", "example"},
		{"sse suffix skipped", "https://mcp.asana.com/sse", "asana"},
		{"bare mcp path falls back to host", "https://mcp.notion.com/mcp", "notion"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := `{"tools":[{"name":"get_thing","inputSchema":{"type":"object","properties":{}}}]}`
			s, err := Parse("t.json", []byte(data), ParseOptions{ServerURL: tc.serverURL})
			require.NoError(t, err)
			assert.Equal(t, tc.want, s.Name)
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"list_projects", []string{"list", "projects"}},
		{"getFileContent", []string{"get", "file", "content"}},
		{"qbo_sales_create_invoice", []string{"qbo", "sales", "create", "invoice"}},
		{"HTTPServerName", []string{"http", "server", "name"}},
		{"brand-radar-cited-pages", []string{"brand", "radar", "cited", "pages"}},
		{"", nil},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, splitWords(tc.in))
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := map[string]string{
		"project":  "projects",
		"files":    "files",
		"box":      "box",
		"category": "categories",
		"day":      "days",
		"":         "",
	}
	for in, want := range tests {
		assert.Equal(t, want, pluralize(in), "pluralize(%q)", in)
	}
}

func TestEnumStrings(t *testing.T) {
	got := enumStrings([]any{"a", float64(2), true, float64(2.5), []any{"nested"}})
	assert.Equal(t, []string{"a", "2", "true", "2.5"}, got,
		"integral floats render without .0 and non-scalars are dropped")
	assert.Nil(t, enumStrings(nil))
	assert.Nil(t, enumStrings([]any{[]any{"only-nested"}}))
}

func TestFirstSentenceTruncatesOnRuneBoundary(t *testing.T) {
	// A long multi-byte description must not be split mid-character.
	got := firstSentence(strings.Repeat("é", 120))
	assert.True(t, len([]rune(got)) <= 201, "truncated to the rune budget plus ellipsis")
	assert.True(t, json.Valid([]byte(`"`+got+`"`)), "result stays valid UTF-8")
}

func paramNames(params []spec.Param) []string {
	out := make([]string, 0, len(params))
	for _, p := range params {
		out = append(out, p.Name)
	}
	return out
}

func findParam(t *testing.T, params []spec.Param, name string) spec.Param {
	t.Helper()
	for _, p := range params {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("param %q not found in %v", name, paramNames(params))
	return spec.Param{}
}
