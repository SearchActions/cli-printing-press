---
title: "Printing Press v2: Comprehensive Anti-Hallucination Overhaul"
type: feat
status: active
date: 2026-03-26
origin: Discord CLI investigation + Linear CLI hallucination audit + discrawl comparison
---

# Printing Press v2: Comprehensive Anti-Hallucination Overhaul

## The Problem in One Sentence

The printing-press produces CLIs that score Grade A on its own scorecard and fail on the first real API call.

## Evidence

| Run | Scorecard | Honest Score | Root Cause |
|-----|-----------|-------------|------------|
| Discord (REST/OpenAPI) | 96/110 Grade A | ~35/110 Grade F | Generator's sync/store/auth templates are disconnected from the spec. 5 critical bugs traced to source lines. |
| Linear (GraphQL) | 91/110 Grade B | ~70/110 Grade C+ | Hand-written code is mostly correct, but 21/36 scorecard points came from dead code written to game string-matching dimensions. |

**Core dynamic:** The scorecard rewards string presence. The SKILL.md tells Claude to raise the scorecard. Claude writes dead code containing trigger strings. The scorecard rewards it. Goodhart's Law in action.

## Solution Architecture

Six phases, each with hard gates. Phase N+1 cannot start until Phase N passes its acceptance test.

```
Phase 1: Honest Scorecard          (fix the measuring stick)
Phase 2: Generator Data Pipeline   (wire the disconnected subsystems)
Phase 3: Generator Auth & Types    (fix protocol-level bugs)
Phase 4: Dogfood Automation        (mechanical verification, not vibes)
Phase 5: GraphQL Support           (the Linear path shouldn't be 100% hand-written)
Phase 6: Skill Overhaul            (prevent the AI from gaming the system)
```

**Why this order:** Phase 1 first because an honest scorecard immediately exposes all other problems. Phase 2-3 fix the generator so the scorecard has something real to measure. Phase 4 automates what Claude currently does by vibes. Phase 5 handles GraphQL (half of modern APIs). Phase 6 updates the skill to use the new tools.

---

# Phase 1: Honest Scorecard

## Goal

Replace the string-matching scorecard with one that tests behavior. After this phase, discord-cli should score D/F (not A) and the delta should be obviously correct to anyone reading the report.

## What's Wrong Now

Every dimension in `internal/pipeline/scorecard.go` works the same way: grep the source files for specific strings. Examples:

| Dimension | What It Checks | What It Should Check |
|---|---|---|
| ErrorHandling (line 183) | `"hint:"` string in helpers.go | Errors from real API calls produce helpful messages |
| LocalCache (line 407) | `"readCache"` string in code | Sync populates the store and queries read from it |
| Vision (line 500) | sync.go file exists | Sync calls domain-specific upsert methods |
| Auth (line 139) | `os.Getenv` string in config.go | Auth header format matches spec's securitySchemes |
| Breadth (line 440) | Count of .go files | Commands hit paths that exist in the spec |

Also: the scorecard header says "10 dimensions" but there are 12. The comparison table in `fullrun.go` only shows 8 dimensions totaling "/80". Both are bugs.

## Tasks

### 1.1 Add Tier 2: Domain Correctness Dimensions (5 new, 50 points)

Add to `scorecard.go`:

**PathValidity (0-10):** For 10 sampled command files, extract the URL path from the `path :=` assignment in RunE. Parse out path segments. Look up each non-param segment in the spec's paths object (load the spec from a `--spec` flag on the scorecard command). Score = (valid paths / total tested) * 10.

```go
// Pseudocode for PathValidity
func scorePathValidity(dir string, specPaths map[string]bool) int {
    commands := sampleCommandFiles(dir, 10)
    valid := 0
    for _, cmd := range commands {
        path := extractPathFromSource(cmd) // regex for `path := "/..."`
        if specContainsPath(specPaths, path) {
            valid++
        }
    }
    return valid * 10 / len(commands)
}
```

**AuthProtocol (0-10):** Read the spec's securitySchemes. Read the generated config.go and client.go. Check:
- Does the auth header name match? (2 pts)
- Does the format match (Bearer vs Bot vs Basic vs raw key)? (3 pts)
- Is `auth.In` respected (header vs query vs cookie)? (3 pts)
- Are env var names derived from the scheme? (2 pts)

**DataPipelineIntegrity (0-10):** Static analysis of the generated code:
- Does sync.go import the store package? (1 pt)
- Does sync.go call any method starting with `Upsert` (not just generic `Upsert`)? (3 pts)
- Does search.go call a domain-specific search method (not generic `Search`)? (3 pts)
- Do domain tables in the store have more columns than just `(id, data, synced_at)`? (3 pts)

**SyncCorrectness (0-10):**
- Does `defaultSyncResources()` return a non-empty slice? (2 pts)
- Do sync paths contain `{` path params (indicating guild-scoped routes)? (3 pts)
- Does sync read from sync_state before fetching? (2 pts)
- Does sync write to sync_state after fetching? (1 pt)
- Does sync handle pagination (calls paginatedGet or loops with cursor)? (2 pts)

**TypeFidelity (0-10):** Sample 10 flag declarations. For each, check against spec params:
- String type for ID fields (not int)? (3 pts)
- Required params have `MarkFlagRequired` or equivalent? (3 pts)
- Flag descriptions are >5 words (not "After", "Limit")? (2 pts)
- No dummy import guards (`var _ = strings.ReplaceAll`)? (2 pts)

### 1.2 Add Dead Code Detection

New function `scoreDeadCode(dir string) int` that:
1. Reads root.go to find all registered flags (regex for `Flags().BoolVar\|Flags().StringVar` etc.)
2. For each flag field name, grep all RunE functions for `flags.<fieldName>`
3. Flags declared but never checked = dead flags
4. Reads helpers.go, finds all exported function names
5. Greps all other .go files for each function name
6. Functions defined but never called = dead functions
7. Score: 10 - (dead_flags + dead_functions) * 2, floor at 0

### 1.3 Rebalance Tier 1 Weights

Reduce Tier 1 from 120 max to 50 max:

| Dimension | Old Max | New Max |
|---|---|---|
| OutputModes | 10 | 7 |
| Auth | 10 | 5 |
| ErrorHandling | 10 | 5 |
| TerminalUX | 10 | 5 |
| README | 10 | 5 |
| Doctor | 10 | 5 |
| AgentNative | 10 | 7 |
| LocalCache | 10 | 4 |
| Breadth | 10 | 4 |
| Vision | 10 | (moved to Tier 2) |
| Workflows | 10 | (moved to Tier 2) |
| Insight | 10 | 3 |

Tier 2 (50 max): PathValidity (10), AuthProtocol (10), DataPipelineIntegrity (10), SyncCorrectness (10), TypeFidelity (5), DeadCode (5).

New total: 100 max.

### 1.4 Fix Scorecard Reporting Bugs

- Fix header: "12 dimensions" -> actual count, "/120" -> "/100"
- Fix `fullrun.go` comparison table to show ALL dimensions
- Add `--spec` flag to scorecard command for Tier 2 validation

### 1.5 Write Tests

Golden-file tests for each new dimension:
- Test PathValidity with a known-good CLI (paths from spec) vs known-bad (flat paths)
- Test AuthProtocol with Bearer vs Bot vs API key specs
- Test DataPipelineIntegrity with domain-specific vs generic store calls
- Test DeadCode with a file containing unwired flags

## Acceptance Criteria

- [ ] `printing-press scorecard --dir ./discord-cli --spec /tmp/printing-press-spec-discord.json` gives:
  - Tier 1: ~42/50 (infrastructure is genuinely good)
  - Tier 2: <10/50 (domain is broken)
  - Total: ~50/100 Grade D
- [ ] `printing-press scorecard --dir ./linear-cli` gives:
  - Tier 1: ~35/50 (helpers.go dead code penalized)
  - Tier 2: ~30/50 (hand-written code is mostly correct, but some unwired flags)
  - Total: ~65/100 Grade C+
- [ ] Dead code in linear-cli's helpers.go detected and reported
- [ ] All tests pass: `go test ./internal/pipeline/...`
- [ ] Scorecard output is self-documenting: each dimension shows what was checked and what failed

## Estimated Effort

Human team: 3-4 days. CC: 2-3 hours.

---

# Phase 2: Generator Data Pipeline

## Goal

Wire the three disconnected subsystems (schema builder, entity mapper, templates) so that generated sync/store/search commands actually move data through domain-specific tables.

## What's Wrong Now

Three subsystems exist but don't talk to each other:

```
schema_builder.go ──(computes rich schema)──> NOBODY CALLS IT
entity_mapper.go  ──(maps domain entities)──> Only PM workflow templates
sync.go.tmpl      ──(calls generic Upsert)──> resources table (wrong table)
store.go.tmpl     ──(creates bare tables)──> (id, data, synced_at) only
search.go.tmpl    ──(calls generic Search)──> resources_fts (wrong index)
```

After this phase:

```
schema_builder.go ──(computes rich schema)──> store.go.tmpl (domain tables + upsert methods)
entity_mapper.go  ──(maps domain entities)──> sync.go.tmpl + search.go.tmpl + store.go.tmpl
sync.go.tmpl      ──(calls UpsertX())──────> domain tables (correct)
search.go.tmpl    ──(calls SearchX())──────> domain FTS index (correct)
```

## Tasks

### 2.1 Wire BuildSchema() Into Store Template

**Files:** `generator.go`, `store.go.tmpl`

Currently `generator.go` renders `store.go.tmpl` with `SyncableResources []string` and `SearchableFields map[string][]string`. These come from the profiler but contain only names, not schema details.

Fix: Call `BuildSchema()` in the generation flow. Pass the full `[]TableDef` output to the store template. The template should iterate over table definitions and generate:

```go
// For each table with rich columns:
CREATE TABLE IF NOT EXISTS {{.Name}} (
    {{range .Columns}}{{.Name}} {{.SQLType}} {{.Constraints}},
    {{end}}
    data JSON NOT NULL
);

// For each table with FTS5:
CREATE VIRTUAL TABLE IF NOT EXISTS {{.Name}}_fts USING fts5(
    {{join .FTSColumns ","}},
    content='{{.Name}}',
    content_rowid='rowid'
);

// Generated upsert method:
func (s *Store) Upsert{{pascal .Name}}(data json.RawMessage) error {
    var obj struct {
        {{range .Columns}}{{pascal .Name}} {{.GoType}} `json:"{{.JSONName}}"`
        {{end}}
    }
    json.Unmarshal(data, &obj)
    _, err := s.db.Exec(`INSERT INTO {{.Name}} (...) VALUES (...)
        ON CONFLICT(id) DO UPDATE SET ...`, ...)
    return err
}
```

Preserve the generic `Upsert`/`Search` methods for backward compatibility (generated API commands still use them for non-primary entities).

### 2.2 Wire Entity Mapper to Sync Template

**Files:** `generator.go`, `sync.go.tmpl`

The sync template currently iterates a flat list of resource names and hits `GET /<resource>`. Fix:

1. Pass the entity mapper's output to the sync template
2. The template should generate API-topology-aware sync:
   - For REST APIs with nested paths: iterate parent resources first, then children scoped to each parent
   - For each entity: call the domain-specific `UpsertX()` method, not generic `Upsert()`
3. `defaultSyncResources()` should return the actual list from the profiler, not `[]string{}`

For the common 2-level nesting pattern (covers Discord, GitHub, Stripe, etc.):

```go
// Template generates:
func syncGuildScoped(c *client.Client, db *store.Store, guildID string) error {
    // Sync channels for this guild
    channels, err := c.Get("/guilds/"+guildID+"/channels", nil)
    // ... parse array, call db.UpsertChannel(guildID, item) for each

    // For each channel, sync messages
    for _, ch := range channelIDs {
        messages, err := paginatedGet(c, "/channels/"+ch+"/messages", ...)
        // ... call db.UpsertMessage(guildID, item) for each
    }
}
```

The key insight: the parser ALREADY extracts the nested path structure (`resourceAndSubFromPath` at line 1366). The entity mapper ALREADY classifies entities by type. We just need to pass both to the template.

### 2.3 Wire Entity Mapper to Search Template

**Files:** `generator.go`, `search.go.tmpl`

Currently calls `db.Search(query, limit)` which queries generic `resources_fts`. Fix: call `db.Search<Primary>(query, guildID, channelID, authorID, limit)` for domain-specific FTS search.

The template should generate domain-specific search flags based on the entity mapper's classifications:
- If entity has a team/guild parent: add `--guild` or `--team` flag
- If entity has an author/user reference: add `--author` flag
- If entity has a temporal field: add `--since`, `--days` flags

### 2.4 Fix Sync Path Construction

**Files:** `sync.go.tmpl`

Replace `path := "/" + resource` (line 60) with actual paths from the spec. The generator should pass each syncable resource's list endpoint path to the template.

For paginated endpoints, use the detected pagination params (cursor name, limit name) from the parser instead of hardcoding `"after"` and `"limit"`.

### 2.5 Write Integration Tests

Create a test spec (small OpenAPI doc with 3-4 nested resources) and verify:
- `go generate` produces a store.go with domain-specific tables
- store.go has `UpsertX()` methods for primary entities
- sync.go calls `UpsertX()` not `Upsert()`
- search.go calls `SearchX()` not `Search()`
- sync paths match the spec's paths

Golden-file approach: store expected output in `testdata/golden/discord-store.go.expected`, compare against generated output.

## Acceptance Criteria

- [ ] Regenerate discord-cli: `store.go` has domain tables with columns (messages, channels, users, members, guilds, threads, roles)
- [ ] Regenerate discord-cli: `sync.go` calls `db.UpsertMessage()`, `db.UpsertChannel()`, etc.
- [ ] Regenerate discord-cli: `search.go` calls `db.SearchMessages()` with `--channel`, `--author` flags
- [ ] Regenerate discord-cli: `defaultSyncResources()` returns non-empty list
- [ ] Regenerate discord-cli: sync paths are guild-scoped (`/guilds/{id}/channels`, `/channels/{id}/messages`)
- [ ] New scorecard Tier 2 DataPipelineIntegrity = 8+/10 for regenerated discord-cli
- [ ] New scorecard Tier 2 SyncCorrectness = 7+/10 for regenerated discord-cli
- [ ] `go test ./internal/generator/...` passes with golden-file comparisons
- [ ] `go build ./...` and `go vet ./...` pass on regenerated discord-cli

## Estimated Effort

Human team: 1.5-2 weeks. CC: 4-6 hours.

---

# Phase 3: Generator Auth & Types

## Goal

Fix protocol-level bugs so generated CLIs authenticate correctly and use proper types for API-specific value formats.

## Tasks

### 3.1 Fix Auth Header Format

**Files:** `client.go.tmpl` (line 149), `parser.go` (line 250)

The parser already detects bot tokens and captures `auth.Format`, `auth.Header`, and `auth.In`. The client template ignores all three.

Fix `client.go.tmpl`:

```go
// Before (hardcoded):
req.Header.Set("Authorization", "Bearer "+authHeader)

// After (format-aware):
{{if .Auth.Format}}
    formatted := strings.ReplaceAll("{{.Auth.Format}}", "{ {{- .Auth.EnvVarPlaceholder -}} }", token)
    req.Header.Set("{{.Auth.Header}}", formatted)
{{else if eq .Auth.In "query"}}
    q := req.URL.Query()
    q.Set("{{.Auth.Header}}", token)
    req.URL.RawQuery = q.Encode()
{{else}}
    req.Header.Set("{{.Auth.Header}}", "Bearer "+token)
{{end}}
```

### 3.2 Fix Pagination Cursor Casing

**File:** `parser.go` (line 2092)

```go
// Before:
pag.CursorParam = strings.ToLower(name) // loses original casing

// After:
pag.CursorParam = name // preserve original casing
// (keep ToLower only for the matching logic, not the stored value)
```

### 3.3 Fix Positional Arg Indexing

**File:** `command_endpoint.go.tmpl` (line 53)

Add a separate counter for positional args:

```go
{{$posIdx := 0}}
{{range $i, $p := .Params}}
{{if $p.Positional}}
if len(args) < {{add $posIdx 1}} {
    return usageErr(fmt.Errorf("{{$p.Name}} is required"))
}
path = replacePathParam(path, "{{$p.Name}}", args[{{$posIdx}}])
{{$posIdx = add $posIdx 1}}
{{end}}
{{end}}
```

### 3.4 Fix Module Path

**File:** `go.mod.tmpl`

Replace `github.com/USER/` with a configurable module prefix. Add `--module` flag to the `generate` command. Default to `github.com/<current-git-user>/<output-dir-name>` if git is configured.

### 3.5 Fix Example Values

**File:** `generator.go` (lines 490-516, `exampleValue()`)

The function produces `"abc123"` which the scorecard's own `hasPlaceholderValues()` then penalizes. Fix: produce domain-realistic values based on the parameter name:

```go
func exampleValue(param Param, apiName string) string {
    nameLower := strings.ToLower(param.Name)
    switch {
    case strings.Contains(nameLower, "id"):
        return "550e8400-e29b-41d4-a716-446655440000" // UUID
    case strings.Contains(nameLower, "email"):
        return "user@example.com"
    case strings.Contains(nameLower, "url"):
        return "https://example.com/resource"
    case strings.Contains(nameLower, "name"):
        return "my-" + strings.ToLower(apiName) + "-resource"
    case strings.Contains(nameLower, "limit"):
        return "50"
    case param.Type == "integer":
        return "42"
    case param.Type == "boolean":
        return "true"
    default:
        return "example-value"
    }
}
```

### 3.6 Add Missing comm_health Template

**File:** `templates/workflows/comm_health.go.tmpl`

`vision_templates.go` line 111 references this for communication-archetype APIs. Create it or remove the reference. Since communication APIs (Discord, Slack) benefit from a health command, create a template that checks message volume, member activity, and channel staleness.

### 3.7 Write Tests

- Test auth format rendering for: Bearer, Bot, Basic, API key in header, API key in query
- Test pagination cursor casing preservation
- Test positional arg indexing with mixed positional/non-positional params
- Test exampleValue produces non-placeholder values

## Acceptance Criteria

- [ ] Regenerate discord-cli: auth sends `Bot <token>` (not `Bearer`)
- [ ] Regenerate discord-cli: pagination cursors use original casing from spec
- [ ] Regenerate discord-cli: positional args work when non-positional params precede them
- [ ] Regenerate discord-cli: module path is `github.com/mvanhorn/discord-cli` (not `github.com/USER/`)
- [ ] Regenerate discord-cli: examples use UUIDs/snowflakes instead of "abc123"
- [ ] New scorecard AuthProtocol = 8+/10
- [ ] New scorecard TypeFidelity = 7+/10
- [ ] All tests pass

## Estimated Effort

Human team: 3-4 days. CC: 2-3 hours.

---

# Phase 4: Dogfood Automation

## Goal

Replace Claude-driven dogfood (which gave both broken CLIs a PASS) with a mechanical Go command that validates generated CLIs against their source spec. Claude cannot be trusted to honestly fail its own work.

## The Problem

Phase 4.5 is currently performed by Claude during the printing-press skill run. Claude:
1. Runs `--dry-run` on a few commands and eyeballs the output
2. Writes a dogfood report with scores
3. Claims PASS or WARN
4. Proceeds to Phase 5

Both Discord (broken sync, broken auth, broken search) and Linear (dead --csv, dead --stdin, dead --sync, 100% dead helpers.go) got PASS verdicts. The dogfood has no teeth.

## Tasks

### 4.1 Build `printing-press dogfood` Command

New subcommand: `printing-press dogfood --dir ./discord-cli --spec /tmp/spec.json`

This command:
1. Loads the spec
2. Builds the CLI binary
3. Runs `<cli> --help` and parses the command tree
4. For each command (or sampled 30 for large CLIs):
   a. Runs `<cli> <command> --dry-run` with synthetic args
   b. Parses the dry-run output (HTTP method + URL)
   c. Validates the URL path exists in the spec
   d. Validates the HTTP method matches the spec for that path
   e. Checks flag types against spec parameter types
5. For store.go: parses CREATE TABLE statements, checks for domain columns vs generic
6. For sync.go: checks for domain-specific upsert calls vs generic Upsert
7. For search.go: checks for domain-specific search vs generic Search
8. Checks for dead flags (declared in root.go, never read in RunE)
9. Checks for dead functions (defined in helpers.go, never called)
10. Produces a structured report: PASS / WARN / FAIL with evidence

### 4.2 Define Dogfood Thresholds

| Metric | PASS | WARN | FAIL |
|---|---|---|---|
| Path validity | >= 90% | >= 70% | < 70% |
| Auth protocol match | Yes | N/A | No |
| Domain tables (not generic) | >= 1 primary entity | 0 but has store.go | No store at all |
| Dead flags | 0 | 1-2 | 3+ |
| Dead functions | 0 | 1-3 | 4+ |
| Sync calls domain upsert | Yes | Calls generic Upsert | No sync at all |

### 4.3 Integrate Into Pipeline

Update `internal/pipeline/` to run the dogfood command as part of Phase 6 (Review). The dogfood result gates progression: FAIL = stop pipeline, WARN = continue with warnings in final report, PASS = proceed.

### 4.4 Write Tests

- Test dogfood against discord-cli (should FAIL on path validity, auth)
- Test dogfood against linear-cli (should WARN on dead flags)
- Test dogfood against a golden-file CLI (should PASS)

## Acceptance Criteria

- [ ] `printing-press dogfood --dir ./discord-cli --spec /tmp/printing-press-spec-discord.json` outputs FAIL with specific evidence:
  - "Auth: Bearer instead of Bot (spec says BotToken scheme)"
  - "Paths: 12/15 tested paths not found in spec"
  - "Dead flags: csvOutput, stdinInput (declared, never read)"
  - "Sync: calls generic Upsert, not domain-specific UpsertMessage"
- [ ] `printing-press dogfood --dir ./linear-cli` outputs WARN:
  - "Dead flags: csvOutput, stdinInput, syncFirst"
  - "Dead functions: filterFields, outputTabwriter, handleNDJSON, newAuthError, newNotFoundError, newRateLimitError, newConflictError"
- [ ] Dogfood is integrated into pipeline's review phase
- [ ] All tests pass

## Estimated Effort

Human team: 1 week. CC: 3-4 hours.

---

# Phase 5: GraphQL Support

## Goal

When the printing-press encounters a GraphQL API, it should generate more than empty scaffolding. The Linear run was 100% hand-written - the generator contributed nothing beyond go.mod and main.go. For GraphQL APIs, the generator should produce a typed GraphQL client, entity-specific query/mutation functions, and domain-aware sync.

## What the Generator Currently Does for GraphQL

Per `vision_templates.go` and the SKILL.md GraphQL mode instructions:
- Creates project structure (cmd/, internal/cli/, internal/client/, internal/config/, internal/store/)
- Generates go.mod with cobra + SQLite + a GraphQL client
- Generates root.go with global flags
- Generates client.go with a generic `Query(query string, variables map[string]any)` method
- Generates helpers.go, doctor.go, auth.go, store.go from templates
- Does NOT generate any per-entity commands

Everything else (sync, search, workflow commands, entity queries, mutations) must be hand-written in Phase 4 of the skill pipeline.

## Tasks

### 5.1 GraphQL Schema Parser

**New file:** `internal/graphql/parser.go`

Parse GraphQL SDL (Schema Definition Language) files to extract:
- Types (objects) with their fields and types
- Queries (root query type fields) with arguments and return types
- Mutations (root mutation type fields) with input types
- Connections (Relay-style pagination: `*Connection` types with `nodes`, `pageInfo`)
- Enums
- Input types

Output: a `GraphQLSchema` struct that maps to the same `APISpec` structure the REST generator uses:
- Each query becomes an endpoint (method: QUERY)
- Each mutation becomes an endpoint (method: MUTATION)
- Types become resources
- Connection types indicate pagination

### 5.2 Typed Query Generator

**New template:** `graphql_queries.go.tmpl`

For each entity with a list query (Connection type), generate:

```go
// Generated from schema type IssueConnection
func (c *Client) ListIssues(filter map[string]any, first int, after string) ([]json.RawMessage, string, bool, error) {
    query := `query($filter: IssueFilter, $first: Int, $after: String) {
        issues(filter: $filter, first: $first, after: $after) {
            nodes { id identifier title description priority ... }
            pageInfo { hasNextPage endCursor }
        }
    }`
    vars := map[string]any{"filter": filter, "first": first}
    if after != "" { vars["after"] = after }
    resp, err := c.Query(query, vars)
    // ... parse nodes, pageInfo, return
}
```

For each entity with a get query:
```go
func (c *Client) GetIssue(id string) (json.RawMessage, error) {
    query := `query($id: String!) { issue(id: $id) { id identifier title ... } }`
    // ...
}
```

For each mutation:
```go
func (c *Client) CreateIssue(input map[string]any) (json.RawMessage, error) {
    mutation := `mutation($input: IssueCreateInput!) { issueCreate(input: $input) { issue { id identifier } } }`
    // ...
}
```

### 5.3 GraphQL-Aware Sync Template

**Updated template:** `sync_graphql.go.tmpl`

For GraphQL APIs, sync should:
1. Use the `filter` argument with `updatedAt: { gt: <cursor> }` for incremental sync
2. Use Relay pagination (`first`/`after`/`endCursor`/`hasNextPage`)
3. Call the generated typed query functions (not raw `client.Query()`)
4. Call domain-specific `UpsertX()` methods from the store

### 5.4 GraphQL-Aware Command Generator

For each entity type, generate CRUD commands:
- `<cli> <entity> list` - calls `ListX()` with filter flags
- `<cli> <entity> get <id>` - calls `GetX()`
- `<cli> <entity> create` - calls `CreateX()` from flags or --stdin
- `<cli> <entity> update <id>` - calls `UpdateX()`
- `<cli> <entity> delete <id>` - calls `DeleteX()` (if mutation exists)

This replaces the current approach where GraphQL = empty scaffolding + hand-write everything.

### 5.5 Write Tests

- Parse the Linear GraphQL schema, verify entity extraction (Issue, Comment, Team, etc.)
- Generate typed queries, verify they compile
- Generate sync, verify it uses Relay pagination
- Golden-file test: known schema -> known output

## Acceptance Criteria

- [ ] `printing-press generate --spec linear-schema.graphql --output ./linear-cli-v2` produces:
  - Typed query functions in `internal/client/` (ListIssues, GetIssue, CreateIssue, etc.)
  - Domain-specific store with tables matching schema types
  - Sync using Relay pagination and incremental updatedAt cursor
  - CRUD commands for all major entity types
- [ ] Generated linear-cli-v2 has 30+ commands without hand-writing
- [ ] Scorecard gives Grade C+ or better on first generation (before Phase 4 polish)
- [ ] `go build` and `go vet` pass

## Estimated Effort

Human team: 2-3 weeks. CC: 6-8 hours.

---

# Phase 6: Skill Overhaul

## Goal

Update SKILL.md so that future printing-press runs can't game the scorecard, can't ship dead code, and spend their time budget where it matters.

## Tasks

### 6.1 Add Phase 4.6: Hallucination & Dead Code Audit

Insert between Phase 4.5 (Dogfood) and Phase 5 (Final Steinberger):

```markdown
# PHASE 4.6: HALLUCINATION & DEAD CODE AUDIT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Run `printing-press dogfood --dir ./<api>-cli --spec <spec-path>` and review the output.

### For every dead flag reported:
1. Either wire the flag into at least one RunE function, OR
2. Remove the flag from root.go

### For every dead function reported:
1. Either call the function from a real command's error/output path, OR
2. Delete the function

### For every ghost table reported (created but never populated):
1. Either wire the sync to populate the table, OR
2. Remove the table from the schema

### For every invalid path reported:
1. Fix the path to match the spec, OR
2. Remove the command if the endpoint doesn't exist

### PHASE GATE 4.6

**STOP.** Re-run `printing-press dogfood`. Verify:
1. Dead flags: 0
2. Dead functions: 0
3. Ghost tables: 0
4. Invalid paths: 0
5. Dogfood verdict: PASS or WARN (FAIL = go back and fix)
```

### 6.2 Add Anti-Gaming Rules

Add to the Anti-Shortcut Rules section:

```markdown
- "I'll add a helpers.go with the patterns the scorecard checks for" (STOP. Every function
  in helpers.go MUST be called by at least one command. Dead code is worse than missing code.
  The dogfood command will catch this.)
- "The error handling score is low, let me add error types" (STOP. Error types must be used
  in actual error paths. Adding newAuthError() that nobody calls is gaming.)
- "I'll add --csv and --quiet flags to root.go" (STOP. Every registered flag must be checked
  in at least one RunE function. Flags nobody reads are dead flags.)
- "I'll add insight command files to match the scorecard prefixes" (STOP. Insight commands must
  query tables that are actually populated. A health command querying an empty database is theater.)
```

### 6.3 Add Spec Path Validation to Phase 4.5

Add to Phase 4.5 Step 4.5b Dimension 1:

```markdown
**NEW CHECK: Spec Path Cross-Reference.** For each command tested via --dry-run:
1. Parse the URL from the dry-run output
2. Extract the path (strip base URL and version prefix)
3. Look up the path in the original spec's paths object (treat {param} as wildcards)
4. If the path doesn't exist in the spec: CRITICAL FAILURE (0 points for this command)
5. If the HTTP method doesn't match: CRITICAL FAILURE

This catches sync hitting GET /messages when the spec only has GET /channels/{id}/messages.
```

### 6.4 Add Data Pipeline Trace to Phase 4 Gate

```markdown
**Data Pipeline Trace (MANDATORY before proceeding from Phase 4):**

For each Primary entity from Phase 0.7:
1. WRITE path: What function calls UpsertX() for this entity? File:line.
2. READ path: What command queries this entity's table? File:line.
3. SEARCH path: If entity has FTS5, what command calls SearchX()? File:line.

If any Primary entity has no WRITE path: the data layer is broken. Fix before proceeding.

Present the trace as a table:
| Entity | Write Path | Read Path | Search Path | Status |
|--------|-----------|-----------|-------------|--------|
```

### 6.5 Restructure Time Budget

```markdown
**Time Budget Guidance:**

The printing-press pipeline should spend its time where value is created:

| Phase | % of Total Time | Why |
|-------|----------------|-----|
| 0, 0.5, 0.7 (Research + Prediction) | 25% | Gets the domain model right |
| 1 (Deep Research) | 10% | Competitor analysis, strategic justification |
| 2 (Generate) | 5% | The generator does the mechanical work |
| 3 (Audit) | 5% | Baseline measurement |
| 4 (GOAT Build) | 35% | THIS IS WHERE THE PRODUCT IS BUILT |
| 4.5 (Dogfood) | 10% | Catch what's broken |
| 4.6 (Hallucination Audit) | 5% | Catch dead code and gaming |
| 5 (Final Report) | 5% | Honest measurement |

Phase 4 gets 35% because that's where workflows are built, data pipelines are wired,
and the CLI becomes useful. Do NOT rush Phase 4 to get to Phase 5.
```

### 6.6 Add Module Path Rule

```markdown
**MANDATORY: Set a real module path.** The go.mod module path MUST be a valid Go import path.
`github.com/USER/<name>` is NEVER acceptable. Use the user's GitHub org or ask.
```

### 6.7 Update Scorecard References

Update all scorecard references in SKILL.md to use the new 100-point scale, Tier 1/Tier 2 language, and new grade thresholds.

### 6.8 Add Discrawl as Benchmark Reference

```markdown
**Benchmark: discrawl (steipete/discrawl)**

For communication APIs (Discord, Slack, Teams), use discrawl as the functional benchmark:
- 11 commands, 551 stars
- Domain-specific SQLite tables with proper columns (NOT JSON blobs)
- FTS5 with normalized content (embeds + attachments + polls + replies in search text)
- Gateway WebSocket tail with 6 event types
- Structured mention tracking (who mentioned whom, target type, timestamp)
- Message event audit log (edit/delete history preserved)
- 80%+ test coverage enforced by CI

After generation + Phase 4 build, ask: "Would a discrawl user switch to this?"
If the answer is "no because [feature X] is missing or broken", that's your Phase 4 work item.
```

## Acceptance Criteria

- [ ] SKILL.md has Phase 4.6 with hallucination audit instructions
- [ ] SKILL.md has 4 new anti-gaming rules
- [ ] SKILL.md has spec path validation in Phase 4.5
- [ ] SKILL.md has data pipeline trace in Phase 4 gate
- [ ] SKILL.md has time budget guidance
- [ ] SKILL.md has module path rule
- [ ] SKILL.md references new 100-point scale
- [ ] Next printing-press run on Discord API produces:
  - 0 dead flags, 0 dead functions, 0 ghost tables
  - Honest scorecard grade (likely C/D on first pass, improvable to B with Phase 4 work)
  - Data pipeline trace showing write/read/search paths for all primary entities

## Estimated Effort

Human team: 2-3 days. CC: 1-2 hours.

---

# Validation Run (After All 6 Phases)

After implementing all 6 phases, run the full printing-press pipeline on Discord:

```bash
/printing-press Discord
```

**Expected outcomes:**

1. Phase 2 generates a discord-cli with:
   - Domain-specific SQLite tables (messages, channels, users, members, etc.)
   - Domain-specific upsert methods (UpsertMessage, UpsertChannel, etc.)
   - Guild-scoped sync paths (/guilds/{id}/channels, /channels/{id}/messages)
   - Auth with Bot prefix
   - Real module path

2. Phase 3 scorecard gives honest baseline:
   - Tier 1: ~40/50 (infrastructure good)
   - Tier 2: ~20/50 (needs Phase 4 work)
   - Total: ~60/100 Grade C

3. Phase 4 builds workflows and raises Tier 2:
   - Sync tested with --dry-run, paths validated against spec
   - Workflow commands query populated domain tables
   - Score rises to ~75/100 Grade B

4. Phase 4.5 dogfood catches remaining issues

5. Phase 4.6 hallucination audit: 0 dead flags, 0 dead functions, 0 ghost tables

6. Phase 5 final score: ~75-80/100 Grade B
   - NOT Grade A. A real Grade B is worth more than a fake Grade A.

7. **Comparison against discrawl:**
   - We have broader API coverage (200+ commands vs 12)
   - discrawl has deeper domain intelligence (normalized content, mention tracking, Gateway WebSocket)
   - Honest assessment of what's better and what's worse

---

# Success Metrics

| Metric | Before (Current) | After (All 6 Phases) |
|--------|------------------|---------------------|
| discord-cli scorecard vs reality gap | 96/110 scorecard, ~35 real (61 point gap) | <10 point gap |
| Dead code in generated output | helpers.go 100% dead, unwired flags, ghost tables | 0 |
| Sync hits valid API paths | 0% | 100% |
| Auth protocol matches spec | 0% | 100% |
| Domain tables populated by sync | 0% | 100% for primary entities |
| GraphQL APIs generate commands | 0 commands generated | 30+ commands |
| Dogfood is automated | Claude self-reports (unreliable) | `printing-press dogfood` command |
| Scorecard tests behavior | 0 semantic dimensions | 6 semantic dimensions |
| Time to first honest Grade B | Never achieved | ~90 minutes |

---

# Risk Analysis

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Phase 2 nested path generation is hard for deeply nested APIs (3+ levels) | High | Start with 2-level (parent/{id}/child). 3+ levels flagged for manual Phase 4 work. Covers 90% of APIs. |
| Phase 5 GraphQL parser is a significant new subsystem | High | Start with SDL parsing only (not introspection). Linear's schema is SDL. Most GraphQL APIs publish SDL. |
| Honest scorecard may discourage users (lower grades) | Medium | The README and final report should explain: "Grade C means the infrastructure works and domain features need polish. This is honest. The old Grade A was lying." |
| Phase 4 automated dogfood may have false positives | Medium | Start with high-confidence checks only (path exists in spec, auth format matches). Add nuanced checks later. |
| Skill changes may make the pipeline take longer | Low | Explicitly acceptable: "I don't care how long it takes. It needs to print epic CLIs." Budget 2 hours. |

---

# Sources

### Internal Analysis (this conversation)
- Discord CLI vs discrawl comparison: `docs/plans/2026-03-26-feat-discord-cli-vs-discrawl-analysis-plan.md`
- Discord investigation: 5 failure classes traced to generator source lines
- Linear hallucination audit: 4 CRITICAL dead code issues, 21/36 points from gaming
- Generator architecture analysis: 30 templates, 12 scorecard dimensions, 8 pipeline phases

### Generator Source (bug locations)
- `internal/generator/generator.go:205-260` - BuildSchema() disconnected
- `internal/generator/templates/sync.go.tmpl:60-62` - Flat paths
- `internal/generator/templates/client.go.tmpl:149` - Hardcoded Bearer
- `internal/openapi/parser.go:2092` - Cursor lowercasing
- `internal/generator/templates/command_endpoint.go.tmpl:53` - Positional arg indexing
- `internal/pipeline/scorecard.go` - All string-matching
- `internal/pipeline/fullrun.go:268-289` - Comparison table shows 8/12 dimensions

### Reference Implementations
- [discrawl](https://github.com/steipete/discrawl) - 551 stars, the gold standard for what "works" means
- [schpet/linear-cli](https://github.com/schpet/linear-cli) - 519 stars, the reference for a real Linear CLI
- Goodhart's Law: "When a measure becomes a target, it ceases to be a good measure"
