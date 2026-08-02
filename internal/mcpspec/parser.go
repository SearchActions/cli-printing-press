// Package mcpspec turns an MCP server's advertised tool catalog into an
// APISpec. It is a pure byte->spec parser: the network fetch that produces
// the catalog lives in the `mcp-sniff` command, so this package stays
// unit-testable against fixtures and `generate` stays hermetic.
//
// MCP tools are invoked by POSTing a JSON-RPC `tools/call` envelope to a
// single server endpoint. The spec models that the way the GraphQL parser
// models GraphQL: endpoint Method carries the *semantic* verb (GET for a
// read, DELETE for a destructive call) while the emitted transport rewrites
// every call to a POST at MCPEndpointPath. Semantic methods are what drive
// the generated MCP tool safety annotations, so a read tool must not be
// labeled POST just because JSON-RPC rides POST.
package mcpspec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
)

// ParseOptions carries the server coordinates a bare tools/list response
// does not contain.
type ParseOptions struct {
	// ServerURL is the full MCP endpoint (e.g.
	// https://api.anthropic.com/v1/design/mcp). It supplies BaseURL and
	// MCPEndpointPath. Empty leaves a placeholder base URL, which `generate`
	// rejects with an actionable message.
	ServerURL string
	// Name overrides the derived API slug.
	Name string
}

// catalog is the on-disk envelope `mcp-sniff` writes, and also accepts a
// raw JSON-RPC tools/list response so a catalog captured by any other MCP
// client can be fed in directly.
type catalog struct {
	ServerURL  string      `json:"server_url"`
	ServerInfo *serverInfo `json:"server_info"`
	Tools      []mcpTool   `json:"tools"`

	// Result is the JSON-RPC wrapper: {"result":{"tools":[...]}}.
	Result *struct {
		Tools []mcpTool `json:"tools"`
	} `json:"result"`

	// APIName and BaseURL are tools-manifest.json markers. A printed CLI's
	// tools-manifest.json also has a top-level `tools` array, so detection
	// must be able to rule it out.
	APIName string `json:"api_name"`
	BaseURL string `json:"base_url"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Title   string `json:"title"`
}

type mcpTool struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	InputSchema      *jsonSchema `json:"inputSchema"`
	InputSchemaSnake *jsonSchema `json:"input_schema"`

	Annotations *toolAnnotations `json:"annotations"`

	// Method/Path are tools-manifest.json fields, decoded only so detection
	// can reject a printed CLI's manifest.
	Method string `json:"method"`
	Path   string `json:"path"`
}

func (t mcpTool) schema() *jsonSchema {
	if t.InputSchema != nil {
		return t.InputSchema
	}
	return t.InputSchemaSnake
}

// toolAnnotations mirrors the MCP tool annotation hints. They are advisory
// per spec, but when a server states readOnlyHint we trust it over any
// name-shape heuristic.
type toolAnnotations struct {
	ReadOnlyHint    *bool `json:"readOnlyHint"`
	DestructiveHint *bool `json:"destructiveHint"`
	IdempotentHint  *bool `json:"idempotentHint"`
	OpenWorldHint   *bool `json:"openWorldHint"`
}

type jsonSchema struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Properties  map[string]*jsonSchema `json:"properties"`
	Required    []string               `json:"required"`
	Enum        []any                  `json:"enum"`
	Items       *jsonSchema            `json:"items"`
	Default     any                    `json:"default"`
	Format      string                 `json:"format"`
	Maximum     *float64               `json:"maximum"`
	MinLength   *float64               `json:"minLength"`
	MaxLength   *float64               `json:"maxLength"`
}

// IsMCPToolsList reports whether data looks like an MCP tool catalog.
//
// It must not match a printed CLI's tools-manifest.json, which also carries a
// top-level `tools` array of objects with `name` and `description`. The
// discriminator is inputSchema (MCP) versus method/path (manifest), plus the
// manifest's top-level api_name/base_url markers.
func IsMCPToolsList(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	if !strings.HasPrefix(trimmed, "{") {
		return false
	}
	var doc catalog
	if err := json.Unmarshal(data, &doc); err != nil {
		return false
	}
	// tools-manifest.json, not an MCP catalog.
	if doc.APIName != "" || doc.BaseURL != "" {
		return false
	}
	tools := doc.toolList()
	if len(tools) == 0 {
		return false
	}
	for _, t := range tools {
		if t.Name == "" {
			return false
		}
		// A manifest entry is method+path shaped and never has inputSchema.
		if t.schema() == nil {
			return false
		}
		if t.Method != "" || t.Path != "" {
			return false
		}
	}
	return true
}

func (c catalog) toolList() []mcpTool {
	if len(c.Tools) > 0 {
		return c.Tools
	}
	if c.Result != nil {
		return c.Result.Tools
	}
	return nil
}

// Parse converts an MCP tool catalog into an APISpec.
func Parse(source string, data []byte, opts ParseOptions) (*spec.APISpec, error) {
	var doc catalog
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decoding MCP tool catalog: %w", err)
	}
	tools := doc.toolList()
	if len(tools) == 0 {
		return nil, fmt.Errorf("MCP tool catalog %s advertises no tools", source)
	}

	serverURL := opts.ServerURL
	if serverURL == "" {
		serverURL = doc.ServerURL
	}

	name := opts.Name
	if name == "" {
		name = deriveName(doc, serverURL, source)
	}

	baseURL, endpointPath, err := splitServerURL(serverURL)
	if err != nil {
		return nil, err
	}
	baseURLIsPlaceholder := false
	if baseURL == "" {
		baseURL = spec.PlaceholderBaseURL
		baseURLIsPlaceholder = true
	}
	if endpointPath == "" {
		endpointPath = "/mcp"
	}

	apiSpec := &spec.APISpec{
		Name:                 name,
		Description:          describeServer(doc, name),
		BaseURL:              baseURL,
		BaseURLIsPlaceholder: baseURLIsPlaceholder,
		MCPEndpointPath:      endpointPath,
		Auth:                 defaultAuth(name),
		Config: spec.ConfigSpec{
			Format: "toml",
			Path:   fmt.Sprintf("~/.config/%s-pp-cli/config.toml", name),
		},
		Resources: map[string]spec.Resource{},
		Types:     map[string]spec.TypeDef{},
	}

	// Sort tools so generation is deterministic regardless of catalog order.
	sorted := append([]mcpTool(nil), tools...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	prefix := commonNamePrefix(sorted)

	// Group tools into resources, then name endpoints within each resource.
	grouped := map[string][]classifiedTool{}
	order := []string{}
	for _, t := range sorted {
		ct := classify(t, prefix)
		if _, seen := grouped[ct.Resource]; !seen {
			order = append(order, ct.Resource)
		}
		grouped[ct.Resource] = append(grouped[ct.Resource], ct)
	}
	sort.Strings(order)

	for _, resName := range order {
		res := spec.Resource{
			Description: fmt.Sprintf("MCP tools for %s", strings.ReplaceAll(resName, "-", " ")),
			Endpoints:   map[string]spec.Endpoint{},
		}
		used := map[string]struct{}{}
		for _, ct := range grouped[resName] {
			epName := uniqueEndpointName(ct.Endpoint, used)
			used[epName] = struct{}{}
			res.Endpoints[epName] = buildEndpoint(ct, endpointPath)
		}
		apiSpec.Resources[resName] = res
	}

	return apiSpec, nil
}
