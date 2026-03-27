---
title: "feat: Generate MCP Server Alongside CLI from OpenAPI Spec"
type: feat
status: active
date: 2026-03-26
---

# feat: Generate MCP Server Alongside CLI from OpenAPI Spec

## Overview

Add an `mcp_server.go.tmpl` template to the printing press so every generated CLI automatically gets a companion MCP server binary. The MCP server exposes the same operations as the CLI - same client, same store, same auth - but over the Model Context Protocol (stdin/stdout JSON-RPC) instead of cobra commands.

This means: `printing-press generate --spec discord.json --output ./discord-cli` produces two binaries:
- `discord-cli` (existing) - human + agent CLI via cobra
- `discord-mcp` (new) - MCP server for Claude Desktop, Cursor, Windsurf, etc.

One spec, two interfaces, zero code duplication.

## Problem Statement / Motivation

The printing press currently generates CLIs that are "agent-native" (--json, --select, --dry-run). But the dominant way AI agents actually connect to tools in 2026 is MCP, not shell commands. Every major AI coding tool (Claude Desktop, Cursor, Windsurf, Cline, Continue) uses MCP as the tool discovery and execution protocol.

Right now, if you want an MCP server for an API, you either:
1. Write one by hand (most common - repetitive, error-prone)
2. Use a runtime proxy like Stainless or AWS OpenAPI MCP Server (adds latency, no local data layer)
3. Use a Python generator like FastMCP (no Go, no local SQLite, no offline)

None of these give you what the printing press already has: a local-first data layer (SQLite + FTS5), offline search, incremental sync, typed parameters, and rate-limit-aware clients. The MCP server should inherit all of this.

**Why this is easy**: The CLI already does all the work. Each cobra command's `RunE` function calls the client, parses the response, and formats output. An MCP tool does the exact same thing - call the client, return JSON. The only difference is the transport layer.

## Proposed Solution

### Architecture

```
printing-press generate --spec api.json --output ./api-cli
  |
  |-- cmd/api-cli/main.go      (existing - cobra CLI)
  |-- cmd/api-mcp/main.go      (NEW - MCP server entry point)
  |-- internal/cli/             (existing - cobra commands)
  |-- internal/mcp/             (NEW - MCP tool handlers)
  |-- internal/mcp/tools.go     (NEW - tool registration from spec)
  |-- internal/client/          (shared - same HTTP client)
  |-- internal/store/           (shared - same SQLite store)
  |-- internal/config/          (shared - same config)
```

### How CLI Commands Map to MCP Tools

Every CLI command becomes an MCP tool:

| CLI Command | MCP Tool Name | inputSchema |
|-------------|---------------|-------------|
| `api-cli users list` | `users_list` | `{"limit": int, "offset": int}` |
| `api-cli users get <id>` | `users_get` | `{"id": string (required)}` |
| `api-cli users create --name X` | `users_create` | `{"name": string (required)}` |
| `api-cli search "query"` | `search` | `{"query": string (required), "limit": int}` |
| `api-cli sync` | `sync` | `{"resources": string[], "since": string, "full": bool}` |
| `api-cli sql "SELECT ..."` | `sql` | `{"query": string (required)}` |

Naming convention: `{resource}_{action}` for REST commands, `{command}` for workflow commands. Matches the `operationId` pattern from OpenAPI.

### MCP Server Implementation Pattern

Using `github.com/mark3labs/mcp-go` (the standard Go MCP SDK):

```go
// cmd/api-mcp/main.go
package main

import (
    "fmt"
    "github.com/mark3labs/mcp-go/server"
    "github.com/{owner}/{name}-cli/internal/mcp"
)

func main() {
    s := server.NewMCPServer(
        "{name}-mcp",
        "1.0.0",
        server.WithToolCapabilities(false),
    )

    mcp.RegisterTools(s)

    if err := server.ServeStdio(s); err != nil {
        fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
        os.Exit(1)
    }
}
```

```go
// internal/mcp/tools.go (generated from template)
package mcp

import (
    "context"
    "encoding/json"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/{owner}/{name}-cli/internal/client"
    "github.com/{owner}/{name}-cli/internal/config"
)

func RegisterTools(s *server.MCPServer) {
    // Generated: one tool per endpoint
    {{range .Resources}}
    {{range .Endpoints}}
    s.AddTool(
        mcp.NewTool("{{snake .Resource}}_{{snake .Name}}",
            mcp.WithDescription("{{.Description}}"),
            {{range .Params}}
            mcp.With{{goMCPType .Type}}("{{.Name}}",
                {{if .Required}}mcp.Required(),{{end}}
                mcp.Description("{{.Description}}"),
            ),
            {{end}}
        ),
        handle_{{snake .Resource}}_{{snake .Name}},
    )
    {{end}}
    {{end}}

    // Vision tools (sync, search, sql) if data layer exists
    {{if .VisionSet.Sync}}
    s.AddTool(
        mcp.NewTool("sync",
            mcp.WithDescription("Sync API data to local SQLite"),
            mcp.WithString("resources", mcp.Description("Comma-separated resources to sync")),
            mcp.WithString("since", mcp.Description("Incremental sync since duration (e.g., 7d, 24h)")),
            mcp.WithBoolean("full", mcp.Description("Full resync ignoring checkpoints")),
        ),
        handleSync,
    )
    {{end}}

    {{if .VisionSet.Search}}
    s.AddTool(
        mcp.NewTool("search",
            mcp.WithDescription("Full-text search across synced data"),
            mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
            mcp.WithNumber("limit", mcp.Description("Max results (default 25)")),
        ),
        handleSearch,
    )
    {{end}}

    {{if .VisionSet.Store}}
    s.AddTool(
        mcp.NewTool("sql",
            mcp.WithDescription("Run read-only SQL against local database"),
            mcp.WithString("query", mcp.Required(), mcp.Description("SQL query (read-only)")),
        ),
        handleSQL,
    )
    {{end}}
}
```

### Tool Handler Pattern

Each handler creates a client, makes the API call, returns JSON:

```go
func handle_users_get(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    id, err := req.RequireString("id")
    if err != nil {
        return mcp.NewToolResultError("id is required"), nil
    }

    c, err := newMCPClient()
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    data, err := c.Get("/users/" + id, nil)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    return mcp.NewToolResultText(string(data)), nil
}

// Shared client factory (reads same config as CLI)
func newMCPClient() (*client.Client, error) {
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }
    return client.New(cfg)
}
```

## Technical Considerations

### Template Data Requirements

The MCP template needs the same data as root.go.tmpl:
- `.Resources` with endpoints, params, descriptions
- `.VisionSet` to know which vision tools to register
- `.Auth` for client configuration
- `.Name`, `.Owner`, `.Version`

Plus a new template function:
- `goMCPType(specType string) string` - maps spec types to mcp-go builder methods: `string` -> `String`, `integer` -> `Number`, `boolean` -> `Boolean`, `array` -> `String` (serialized)

### What the MCP Server Gets for Free

By importing the same internal packages as the CLI:
- **Auth** - same token from config file / env var
- **Rate limiting** - same exponential backoff with Retry-After
- **Caching** - same 5-minute GET cache with --no-cache bypass
- **Local data** - same SQLite store for search/sql/sync
- **Pagination** - same cursor-based pagination (new sync template)

### What the MCP Server Does NOT Do

- No interactive prompts (MCP is non-interactive by design)
- No progress bars (MCP tools return a single result)
- No colored output (MCP returns plain text or JSON)
- No --dry-run (MCP tools execute; the agent decides whether to call)

### Binary Size

The MCP binary will be slightly smaller than the CLI binary since it doesn't import cobra. Rough estimate: 8-12MB (CLI is ~15MB) since the main cost is the SQLite driver.

## Acceptance Criteria

- [ ] `printing-press generate` produces `cmd/{name}-mcp/main.go` alongside `cmd/{name}-cli/main.go`
- [ ] Every REST endpoint in the spec becomes an MCP tool with correct inputSchema
- [ ] Vision tools (sync, search, sql) are registered when the data layer is enabled
- [ ] MCP server reads the same config/env vars as the CLI (no duplicate auth setup)
- [ ] Tool handlers return JSON responses (not formatted tables)
- [ ] Tool handlers return `mcp.NewToolResultError()` on API errors (not panics)
- [ ] `go.mod` includes `github.com/mark3labs/mcp-go` dependency
- [ ] Makefile builds both binaries: `bin/{name}-cli` and `bin/{name}-mcp`
- [ ] GoReleaser config builds both binaries for cross-compilation
- [ ] MCP server works with `claude mcp add` for Claude Code integration
- [ ] Proof-of-Behavior verification passes (tools map to real spec endpoints)

## Implementation Phases

### Phase 1: Core Templates (MVP)

Files to create:
- [ ] `internal/generator/templates/main_mcp.go.tmpl` - MCP server entry point
- [ ] `internal/generator/templates/mcp_tools.go.tmpl` - tool registration (one tool per endpoint)
- [ ] `internal/generator/templates/mcp_handlers.go.tmpl` - per-endpoint handler functions
- [ ] `internal/generator/templates/mcp_helpers.go.tmpl` - shared client factory, param extraction

Files to modify:
- [ ] `internal/generator/generator.go` - add `cmd/{name}-mcp/` dir, register MCP templates
- [ ] `internal/generator/templates/go.mod.tmpl` - add `github.com/mark3labs/mcp-go` dependency
- [ ] `internal/generator/templates/makefile.tmpl` - add `build-mcp` target
- [ ] `internal/generator/templates/goreleaser.yaml.tmpl` - add second builds entry

### Phase 2: Vision Tool Handlers

- [ ] `handleSync` - calls store.Open + sync logic (mirrors sync.go.tmpl RunE)
- [ ] `handleSearch` - calls store.SearchX (mirrors search.go.tmpl RunE)
- [ ] `handleSQL` - calls store.QuerySQL (mirrors sql workflow command)
- [ ] `handleStatus` - returns sync state from store

### Phase 3: Quality & Testing

- [ ] Generate a test MCP server from the petstore spec and verify tools are discoverable
- [ ] Add MCP-specific checks to Proof-of-Behavior: every registered tool name matches a spec endpoint
- [ ] README template updated with MCP usage section
- [ ] Test with `claude mcp add ./bin/{name}-mcp` in Claude Code

## Success Metrics

| Metric | Target |
|--------|--------|
| MCP tools generated per CLI | 1:1 with CLI commands |
| Additional code to maintain | 0 (generated from same spec) |
| Time to generate MCP server | < 5 seconds (template rendering) |
| MCP binary size | < 15MB (same as CLI) |
| Works with Claude Code | `claude mcp add` succeeds |

## Dependencies & Risks

- **Dependency**: `github.com/mark3labs/mcp-go` - 5.6k stars, actively maintained, standard Go MCP SDK
- **Risk**: mcp-go API could change. Mitigation: pin version in go.mod template.
- **Risk**: Large specs (300+ endpoints) could generate a very large tools.go file. Mitigation: split into per-resource files like the CLI does.
- **Risk**: Complex request bodies (embeds, components) can't be expressed as flat MCP params. Mitigation: use a single `body` string param with JSON input for complex endpoints (same --stdin pattern as CLI).

## Sources & References

### External
- [mcp-go SDK](https://github.com/mark3labs/mcp-go) - standard Go MCP implementation
- [mcp-go Getting Started](https://mcp-go.dev/getting-started/) - server creation, tool registration, stdio transport
- [Stainless MCP from OpenAPI](https://www.stainless.com/docs/guides/generate-mcp-server-from-openapi/) - commercial OpenAPI-to-MCP generator
- [Speakeasy MCP Tool Design](https://www.speakeasy.com/mcp/tool-design/generate-mcp-tools-from-openapi) - best practices for mapping OpenAPI to MCP tools
- [openapi-mcp-generator](https://github.com/harsha-iiiv/openapi-mcp-generator) - open source OpenAPI-to-MCP tool (TypeScript)
- [AWS OpenAPI MCP Server](https://awslabs.github.io/mcp/servers/openapi-mcp-server/) - runtime proxy approach

### Internal
- Generator entry point: `internal/generator/generator.go:82-112` - directory creation and template mapping
- Vision template selection: `internal/generator/vision_templates.go` - VisionTemplateSet flags
- Template data: `internal/spec/spec.go` - APISpec struct available to all templates
- CLI main.go pattern: `internal/generator/templates/main.go.tmpl` - pattern for new main_mcp.go.tmpl
- Client template: `internal/generator/templates/client.go.tmpl` - shared HTTP client with retry/cache
- Existing MCP competitors noted in README: "Why CLIs (Not APIs, Not MCP)" - this feature makes the argument moot by providing both
