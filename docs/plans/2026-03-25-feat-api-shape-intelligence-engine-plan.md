---
title: "feat: API Shape Intelligence Engine - Predict and Generate Power-User Features from Any OpenAPI Spec"
type: feat
status: active
date: 2026-03-25
---

# API Shape Intelligence Engine

## Overview

Make the printing press think like a power user. Instead of generating pure API wrappers (Tier 1) and hoping someone adds domain intelligence later, the press should analyze the *shape* of any OpenAPI spec and automatically predict + generate the Tier 2 features that power users will inevitably want.

The insight: **every great CLI tool was born because someone looked at an API and asked "what would a power user wish this could do that the raw API can't?"** That question has predictable answers based on the API's shape. Paginated list endpoints predict sync+search. Webhook definitions predict listen+tail. CRUD on long-lived resources predicts declarative state management. The press should encode these heuristics and generate vision features automatically.

## Problem Statement

Today the press has ALL the pieces but they're disconnected:

1. **Vision templates exist** (`export.go.tmpl`, `search.go.tmpl`, `sync.go.tmpl`, `tail.go.tmpl`, `store.go.tmpl`, `analytics.go.tmpl`, `import.go.tmpl`) - 7 templates ready to render
2. **VisionaryPlan types exist** (`internal/vision/vision.go`) - DataProfile, ArchitectureDecision, FeatureIdea with scoring
3. **SelectVisionTemplates() exists** (`internal/generator/vision_templates.go`) - maps plan decisions to template selection
4. **Generate() NEVER calls SelectVisionTemplates()** - the loop is broken. The generator produces Tier 1 only.

The missing link: an **API Shape Profiler** that reads the OpenAPI spec itself (not external research) and produces a `VisionaryPlan` automatically. Phase 0 web research is valuable context, but the core intelligence should come from the spec's structure.

## The Three Meta-Patterns

Every power-user CLI feature falls into one of three categories, each predictable from spec signals:

### Pattern A: "Bridge the Gap" (local-remote sync)
- **Signal:** Resources that have local analogs (files, databases, config, state)
- **Examples:** `aws s3 sync`, `heroku pg:pull`, `gh pr checkout`, `terraform state pull`
- **Spec detection:** Paginated list endpoints, file upload endpoints, config key-value endpoints

### Pattern B: "Collapse the Workflow" (multi-step orchestration)
- **Signal:** Achieving a user goal requires 3+ sequential API calls with response threading
- **Examples:** `stripe trigger`, `gh pr create`, `terraform apply`, `gh repo fork`
- **Spec detection:** Resources with foreign key references to each other, lifecycle state fields

### Pattern C: "Maintain the Connection" (streaming/watching)
- **Signal:** API produces time-series data, events, logs, or async operations
- **Examples:** `stripe listen`, `heroku logs --tail`, `gh run watch`
- **Spec detection:** Webhook definitions, SSE endpoints, operation status endpoints, audit log resources

## Proposed Solution

### The API Shape Profiler (`internal/profiler/`)

A new package that reads a parsed `spec.APISpec` and produces a `vision.VisionaryPlan` purely from structural analysis - no web research needed. This is the brain.

#### Signal Detection Matrix

| Signal in Spec | How to Detect | Feature Predicted | Template |
|---|---|---|---|
| Pagination params (limit/offset/cursor) on list endpoints | Check query params on GET list endpoints | `sync`, `export --all` | sync.go.tmpl, export.go.tmpl |
| No search endpoint but 3+ list endpoints | Count endpoint types per resource | Local FTS5 search | search.go.tmpl, store.go.tmpl |
| Webhook/callback definitions | Check for `webhooks:` key, `x-webhooks`, callback URL schemas | `tail`/`listen` | tail.go.tmpl |
| Full CRUD (GET/POST/PUT/DELETE) on 3+ resources | Check method coverage per resource | Compound workflow commands | workflow.go.tmpl |
| Response schemas with 15+ fields | Count properties in response schemas | `--select` field filtering + analytics | analytics.go.tmpl |
| Resources with `status`/`state` enum fields | Find enum fields with lifecycle values (pending/active/done) | Lifecycle compound commands | workflow.go.tmpl |
| Resources with foreign key IDs to other resources | Find `_id` suffixed fields or `$ref` patterns | Cross-resource compound commands | workflow.go.tmpl |
| Chronological list endpoints (audit logs, events, history) | Detect `created_at`/`timestamp` sort fields, event-shaped schemas | `tail --follow` with polling | tail.go.tmpl |
| File/binary upload endpoints | `multipart/form-data` or `application/octet-stream` | Upload progress, resume | (future) |
| Rate limit headers documented | `X-RateLimit-*` in response headers | Auto-throttle, batch mode | (built into client.go.tmpl already) |
| Read-heavy endpoint ratio (>60% GET) | Count methods across all endpoints | Offline cache, `--cached` flag | store.go.tmpl |

#### Data Profile Inference

```go
// internal/profiler/profiler.go

type APIProfile struct {
    HighVolume      bool    // >50% of list endpoints have pagination
    NeedsSearch     bool    // 3+ list endpoints, no search endpoint
    HasRealtime     bool    // webhook/SSE/event endpoints exist
    OfflineValuable bool    // read-heavy (>60% GET) with stable data
    ComplexResources bool   // max response schema >15 fields or 3+ nesting levels
    HasLifecycles   bool    // resources with status/state enum fields
    HasDependencies bool    // resources reference each other via IDs
    HasChronological bool   // audit logs, event lists with timestamps
    HasFileOps      bool    // multipart upload or binary download
    CRUDResources   int     // count of resources with full CRUD
    ListEndpoints   int     // count of paginated list endpoints
    TotalEndpoints  int     // total endpoint count
}

func Profile(s *spec.APISpec) *APIProfile { ... }

func (p *APIProfile) ToVisionaryPlan(apiName string) *vision.VisionaryPlan { ... }
```

The profiler runs in <100ms on any spec (it's just structural analysis, no I/O). It produces the same `VisionaryPlan` that Phase 0 web research produces, but deterministically from the spec alone. Phase 0 research can then *augment* the profiled plan with competitive intelligence and community demand signals.

### LLM Vision Synthesis (`internal/llmpolish/vision.go`)

The profiler decides WHICH templates to include. The LLM decides HOW to customize them. This is where Claude's creative intelligence enters the pipeline - reasoning about what power users actually want, not just what the spec structure suggests mechanically.

The LLM Vision Synthesis step:

1. **Receives** the profiler's `APIProfile` + the `VisionaryPlan` (from profiler and/or Phase 0 research)
2. **Reasons** about power-user workflows using domain knowledge:
   - "This API manages Discord communities. Power users will want to search across all channels for specific conversations, monitor channels for keywords like 'deploy' or 'error', and export threads for external analysis."
   - "This is a payment API. Power users will want to replay webhook events locally, simulate payment flows end-to-end, and reconcile transactions against local records."
3. **Produces** a `VisionCustomization` struct:
   - **Resource priority order** - which resources to sync first (messages before emojis)
   - **FTS5 field selection** - which string fields are actually searchable content vs metadata (message `content` yes, message `nonce` no)
   - **Workflow names** - domain-specific names for compound commands (`archive` not `workflow-1`, `monitor` not `tail`)
   - **Example values** - realistic domain-specific examples in help text (Discord snowflake IDs, not "abc123")
   - **Description overrides** - developer-friendly descriptions that reference actual use cases
   - **Sync strategy hints** - pagination direction (newest-first for messages, alphabetical for members), batch sizes
4. **Falls back gracefully** - if no LLM is available (offline, no API key), the profiler's mechanical output is used as-is. The CLI still compiles and works, just with generic descriptions.

```go
// internal/llmpolish/vision.go

type VisionCustomization struct {
    ResourcePriority  []string            // sync order
    FTSFields         map[string][]string // resource -> searchable fields
    WorkflowNames     map[string]string   // generic -> domain-specific
    ExampleOverrides  map[string]string   // command -> better example
    DescOverrides     map[string]string   // command -> better description
    SyncHints         map[string]SyncHint // resource -> pagination strategy
}

func SynthesizeVision(profile *profiler.APIProfile, plan *vision.VisionaryPlan, spec *spec.APISpec) (*VisionCustomization, error) {
    // Build a prompt with the profile, plan, and spec summary
    // Ask Claude to reason about power-user workflows
    // Parse structured response into VisionCustomization
    // Return nil (not error) if LLM unavailable - graceful degradation
}
```

The pipeline becomes three layers of increasing intelligence:

```
Layer 1: Profiler (deterministic, <100ms)
  -> "This API has paginated lists and no search endpoint"
  -> Decides: include sync.go, search.go, store.go

Layer 2: LLM Synthesis (creative, ~30s, optional)
  -> "Messages are the high-value content. Index content and author_username in FTS5.
      Name the workflow 'archive' not 'workflow'. Use snowflake IDs in examples."
  -> Customizes: field selection, names, examples, descriptions

Layer 3: Phase 0 Research (competitive intel, ~5min, optional)
  -> "discrawl at 539 stars validates SQLite+FTS5. DiscordChatExporter at 10.7k stars
      proves export demand. jackwener/discord-cli shows the architecture works."
  -> Augments: evidence scores, competitive framing, community demand
```

Each layer is optional. Layer 1 alone produces a working CLI. Layer 2 makes it smart. Layer 3 makes it strategic.

### Wire Vision Templates into Generate()

The fix is small but critical. In `generator.go`:

```go
func (g *Generator) Generate() error {
    // ... existing directory creation ...
    // ... existing single files ...
    // ... existing per-resource files ...

    // NEW: Profile the API and generate vision features
    profile := profiler.Profile(g.Spec)
    plan := profile.ToVisionaryPlan(g.Spec.Name)

    // If external VisionaryPlan provided (from Phase 0), merge it
    if g.VisionPlan != nil {
        plan = mergeVisionPlans(plan, g.VisionPlan)
    }

    visionSet := SelectVisionTemplates(plan)

    if visionSet.Store {
        os.MkdirAll(filepath.Join(g.OutputDir, "internal", "store"), 0755)
        g.renderTemplate("store.go.tmpl", filepath.Join("internal", "store", "store.go"), g.Spec)
    }
    for _, tmplName := range visionSet.TemplateNames() {
        if tmplName == "store.go.tmpl" { continue } // already handled
        outPath := filepath.Join("internal", "cli", strings.TrimSuffix(tmplName, ".tmpl"))
        g.renderTemplate(tmplName, outPath, visionData)
    }

    // Register vision commands in root.go (template already supports this)
    // ...
}
```

### Domain-Aware Store Schema Generation

The current `store.go.tmpl` uses a generic `resources` table with `resource_type` + `id` + JSON `data`. This works but is not as powerful as domain-specific tables.

**New approach:** Generate per-resource tables from response schemas.

For each resource with a list endpoint:
1. Extract the response schema's scalar fields -> SQLite columns
2. Keep the full JSON in a `_raw TEXT` column for completeness
3. Create FTS5 indexes on string fields
4. Create foreign key relationships from `_id` fields
5. Create `_sync_state` table with per-resource cursors

Example: For Discord's messages endpoint, instead of a generic `resources` table, generate:

```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL,
    _raw TEXT NOT NULL,
    _synced_at TEXT NOT NULL
);
CREATE INDEX idx_messages_channel ON messages(channel_id);
CREATE VIRTUAL TABLE messages_fts USING fts5(content);
```

This is generated from the OpenAPI response schema - no domain knowledge needed. The spec tells us `id` is the primary key, `channel_id` and `author_id` are references, `content` is a string (searchable), `timestamp` is a date.

### Compound Workflow Generator

When the profiler detects Pattern B (multi-step workflows), generate a `workflow.go.tmpl` that:

1. **Archive workflow** - For APIs with high-volume paginated data:
   ```
   <cli> workflow archive --resource messages --guild <id>
   ```
   Combines: list all parent resources -> paginate all child resources -> store locally -> export to file

2. **Lifecycle workflow** - For APIs with status state machines:
   ```
   <cli> workflow advance <resource_id> --to active
   ```
   Combines: get current state -> validate transition -> update -> verify new state

3. **Audit workflow** - For APIs with audit log endpoints:
   ```
   <cli> workflow audit --since 7d
   ```
   Combines: fetch audit logs -> correlate with resources -> summarize changes

### Scorecard v2: Reward Power-User Thinking

The current Vision scoring (0/10) checks file existence only. This should be replaced with a **two-tier Vision score** that measures both presence and intelligence.

#### New Vision Scoring (0-10)

**Tier 1: Feature Presence (0-5 points)**

Same file checks as today, but worth only half:
- `export.go` exists -> +1
- `store/store.go` exists -> +1
- `search.go` exists -> +1
- `sync.go` exists -> +0.5
- `tail.go` exists -> +0.5
- `import.go` exists -> +0.5
- Any `*_workflow.go` or `*_compound.go` -> +0.5

**Tier 2: Feature Intelligence (0-5 points)**

Checks that vision features are wired to the API's actual shape:

- **Profile match (0-2):** Run the profiler on the spec. Compare recommended features vs actual features present. Score = (matching features / recommended features) * 2. If the profiler says "this API needs sync+search" and both exist, full marks. If the profiler says "this API doesn't need tail" and tail is absent, no penalty.

- **Schema depth (0-1.5):** Check if `store.go` has domain-specific tables (not just a generic `resources` table). Heuristic: count `CREATE TABLE` statements in store.go. >1 table (beyond sync_state) = +0.5 per table, up to 1.5.

- **Wiring check (0-1.5):** Check that vision commands are registered in `root.go`. Grep for `newSyncCmd\|newSearchCmd\|newExportCmd\|newTailCmd\|newWorkflowCmd` in root.go. Each found = +0.3, up to 1.5.

**Total: Tier 1 (0-5) + Tier 2 (0-5) = 0-10**

This means:
- Pure API wrapper: 0/10 (no vision files)
- Stub files that exist but aren't wired: 3-5/10 (Tier 1 only)
- Profiled and wired vision features: 8-10/10 (both tiers)
- **Gaming the scorecard with empty files no longer works**

#### New Breadth Scoring Adjustment

Current Breadth penalizes lazy descriptions. Add a **bonus** for vision commands that match the API profile:

- If profiler recommends sync AND sync.go exists: +0.5 to Breadth
- If profiler recommends search AND search.go exists: +0.5 to Breadth
- Cap: +1 bonus (rewarding that vision features expand the CLI's breadth meaningfully)

## Technical Approach

### Architecture

```
OpenAPI Spec (JSON/YAML)
        |
        v
[OpenAPI Parser]  -->  spec.APISpec (existing)
        |
        v
[API Shape Profiler]  -->  profiler.APIProfile (NEW)
        |
        v
[Profile -> VisionaryPlan]  -->  vision.VisionaryPlan (existing type, new source)
        |                              |
        |                   [Phase 0 Web Research]  (optional, augments)
        |                              |
        v                              v
[SelectVisionTemplates]  -->  VisionTemplateSet (existing)
        |
        v
[Generator.Generate()]  -->  Tier 1 commands + Tier 2 vision features
        |
        v
[Scorecard v2]  -->  Validates profile match + wiring
```

### Implementation Phases

#### Phase 1: The Profiler (Foundation)

Create `internal/profiler/profiler.go`:

- `Profile(s *spec.APISpec) *APIProfile` - structural analysis
- `(p *APIProfile) ToVisionaryPlan(name string) *vision.VisionaryPlan` - convert to plan
- `(p *APIProfile) RecommendedFeatures() []string` - human-readable summary

Test with the three existing specs (Petstore, Stytch, Discord):
- Petstore: small, simple CRUD -> should recommend export only
- Stytch: auth API, moderate complexity -> should recommend export + sync
- Discord: massive, real-time, high-volume -> should recommend all vision features

Files:
- `internal/profiler/profiler.go` - core profiler
- `internal/profiler/profiler_test.go` - test with all three specs
- `internal/profiler/signals.go` - signal detection helpers (isPaginated, hasWebhooks, etc.)

#### Phase 2: Wire the Loop (Core)

Connect profiler output to generator:

1. Add `VisionPlan *vision.VisionaryPlan` field to `Generator` struct
2. In `Generate()`, call profiler and `SelectVisionTemplates()`
3. Render vision templates when selected
4. Register vision commands in root.go (modify `root.go.tmpl` to conditionally include)
5. Add `--store` dependency to `go.mod.tmpl` when Store is selected

Files:
- `internal/generator/generator.go` - add vision rendering
- `internal/generator/templates/root.go.tmpl` - conditional vision command registration
- `internal/generator/templates/go.mod.tmpl` - conditional modernc.org/sqlite dep

#### Phase 3: LLM Vision Synthesis (Creative Layer)

Add the LLM reasoning step that customizes vision features with domain intelligence:

1. Build prompt from profiler output + spec summary (resource names, field names, endpoint descriptions)
2. Ask Claude to identify: which fields are searchable content, which resources to prioritize for sync, what to name workflow commands, what realistic examples look like
3. Parse structured response into `VisionCustomization`
4. Pass customization to template rendering (field selection, descriptions, examples)
5. Graceful fallback: if no LLM key, use profiler defaults (generic but functional)

Files:
- `internal/llmpolish/vision.go` - LLM synthesis prompt + response parsing
- `internal/llmpolish/vision_test.go` - test with mock LLM responses
- `internal/generator/generator.go` - pass VisionCustomization to template rendering

#### Phase 4: Domain-Aware Store (Intelligence Layer)

Replace the generic `resources` table with per-resource tables generated from response schemas:

1. Add `SchemaToSQLite(resourceName string, schema map[string]spec.Param) string` to profiler
2. Modify `store.go.tmpl` to accept a list of resource table definitions
3. Generate FTS5 indexes on string fields automatically
4. Generate foreign key relationships from `_id` fields
5. Update `sync.go.tmpl` to iterate resources and populate domain-specific tables
6. Update `search.go.tmpl` to query domain-specific FTS5 tables

Files:
- `internal/profiler/schema.go` - OpenAPI schema -> SQLite DDL converter
- `internal/profiler/schema_test.go` - tests
- `internal/generator/templates/store.go.tmpl` - domain-aware store
- `internal/generator/templates/sync.go.tmpl` - domain-aware sync
- `internal/generator/templates/search.go.tmpl` - domain-aware search

#### Phase 5: Compound Workflows (Polish)

Generate workflow commands based on resource graph analysis:

1. Detect resource dependency chains (A has B_id -> B has C_id)
2. Generate archive workflow (paginate parent -> paginate children -> store all)
3. Generate audit workflow (if audit log resource exists)
4. Generate fixture workflow (for APIs with sequential dependencies like Stripe)

Files:
- `internal/profiler/graph.go` - resource dependency graph builder
- `internal/generator/templates/workflow.go.tmpl` - compound workflow commands
- `internal/generator/templates/workflow_archive.go.tmpl` - archive-specific workflow

#### Phase 6: Scorecard v2

Implement the new two-tier Vision scoring:

1. Run profiler on the spec in the generated CLI's directory (detect spec from go.mod or config)
2. Compare recommended features vs present features
3. Check schema depth (domain-specific tables vs generic)
4. Check wiring (vision commands registered in root.go)
5. Update scoring thresholds

Files:
- `internal/pipeline/scorecard.go` - new `scoreVisionV2()` function
- `internal/pipeline/scorecard_test.go` - tests against generated CLIs

## How This Solves the Discord Failure

With the profiler, the Discord spec would automatically produce:

```
Profile:
  HighVolume: true       (90%+ of list endpoints have pagination)
  NeedsSearch: true      (20+ list endpoints, no search endpoint for bots)
  HasRealtime: true      (Gateway documented, webhook endpoints exist)
  OfflineValuable: true  (75%+ GET endpoints)
  HasChronological: true (messages have timestamps, audit logs exist)
  HasDependencies: true  (messages reference channels, channels reference guilds)
  CRUDResources: 15+     (guilds, channels, messages, roles, emojis, etc.)

Recommended features:
  - store.go     (HighVolume + NeedsSearch + OfflineValuable)
  - sync.go      (HighVolume + HasChronological)
  - search.go    (NeedsSearch)
  - export.go    (HighVolume, always recommended)
  - import.go    (always recommended)
  - tail.go      (HasRealtime + HasChronological)
  - analytics.go (HighVolume + ComplexResources)
  - workflow.go  (HasDependencies + CRUDResources > 5)
```

All vision features would be generated automatically from `./printing-press generate --spec discord.json`. No web research needed. No manual intervention. The API's shape TELLS you what to build.

For a simple API like Petstore:
```
Profile:
  HighVolume: false      (no pagination on list endpoints)
  NeedsSearch: false     (only 2 list endpoints, one has findByStatus)
  HasRealtime: false     (no webhooks)
  OfflineValuable: false (balanced read/write)

Recommended features:
  - export.go    (always recommended)
  - import.go    (always recommended)
```

Only export and import. No bloat. The profiler is smart enough to NOT recommend features the API doesn't warrant.

## Acceptance Criteria

### Functional Requirements
- [ ] `profiler.Profile()` produces correct `APIProfile` for Petstore, Stytch, and Discord specs
- [ ] `./printing-press generate --spec discord.json` produces vision files (sync, search, store, export, tail) without any manual intervention
- [ ] `./printing-press generate --spec petstore.json` produces only export+import (not sync/search/store/tail)
- [ ] Vision templates compile (`go build ./...`) on first try for all test specs
- [ ] `store.go` has domain-specific tables for Discord (messages, channels, members), not generic `resources` table
- [ ] `search.go` queries FTS5 indexes on domain-specific string fields
- [ ] `sync.go` paginates through list endpoints and upserts to domain-specific tables
- [ ] Vision commands are registered in `root.go` automatically
- [ ] Scorecard v2 gives Discord CLI 8+/10 on Vision (profile match + schema depth + wiring)
- [ ] Scorecard v2 gives Petstore CLI 8+/10 on Vision (fewer features recommended AND fewer present = good match)

### Non-Functional Requirements
- [ ] Profiler runs in <100ms on any spec (pure structural analysis, no I/O)
- [ ] No CGO dependency (use `modernc.org/sqlite`, not `mattn/go-sqlite3`)
- [ ] All generated code passes `go vet` and `go build` without modification
- [ ] Profiler is deterministic (same spec always produces same profile)

### Quality Gates
- [ ] `go test ./internal/profiler/...` passes with 3+ spec fixtures
- [ ] `go test ./internal/pipeline/...` passes with new scorecard v2
- [ ] End-to-end: `generate + scorecard` produces Grade A for Discord spec
- [ ] End-to-end: `generate + scorecard` produces Grade A for Petstore spec (different features, same quality)

## Alternative Approaches Considered

### 1. Keep vision as human-only (rejected)
The discrawl parity report concluded "The press prints Tier 1. Discrawl is Tier 1 + Tier 2." This draws the wrong line. The *API spec itself* contains enough information to predict Tier 2 features. The full-steinberger-parity plan already rejected this framing.

### 2. Require Phase 0 web research for vision (rejected)
External research adds competitive intelligence but is slow, non-deterministic, and requires LLM invocation. The profiler should work from the spec alone. Phase 0 research becomes an optional augmentation, not a prerequisite.

### 3. LLM-based feature prediction (rejected for core, kept as augmentation)
An LLM could analyze the spec and suggest features. But this is non-deterministic, slow, and expensive. The heuristic engine is fast, deterministic, and testable. LLM augmentation (via `internal/llmpolish/`) can improve descriptions and examples after generation.

### 4. Generate all vision features for every API (rejected)
Bloat. A simple API like Petstore doesn't need sync, search, or tail. The profiler's job is to recommend the RIGHT features, not ALL features. The scorecard rewards profile match, not feature count.

## System-Wide Impact

### Interaction Graph
1. `printing-press generate` calls `openapi.Parse()` -> `profiler.Profile()` -> `SelectVisionTemplates()` -> `generator.Generate()` (renders vision templates)
2. `printing-press scorecard` calls `profiler.Profile()` again on the generated CLI's spec to compare recommendations vs actual features
3. SKILL.md Phase 0 research produces a `VisionaryPlan` that gets merged with the profiler's plan, adding community demand signals on top of structural analysis

### State Lifecycle Risks
- Generated vision files might have compile errors if response schemas have edge cases (nested arrays, oneOf types). Mitigation: the quality gates (`go build`, `go vet`) catch this. Generator retries with simplified schema if build fails.
- SQLite schema migration on re-generation. Mitigation: store always runs `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS`.

### Integration Test Scenarios
1. Generate Discord CLI -> `go build` -> `scorecard` -> Vision 8+/10
2. Generate Petstore CLI -> `go build` -> `scorecard` -> Vision 8+/10 (different features, same score)
3. Generate with `--spec` only (no Phase 0) -> vision features still present
4. Generate with Phase 0 plan -> vision features augmented with better descriptions
5. Re-generate over existing output -> vision features update cleanly

## Sources & References

### Internal References
- `internal/generator/vision_templates.go` - existing SelectVisionTemplates() logic
- `internal/vision/vision.go` - existing VisionaryPlan types
- `internal/generator/templates/store.go.tmpl` - existing generic store template
- `internal/generator/templates/sync.go.tmpl` - existing generic sync template
- `internal/pipeline/scorecard.go:494-532` - existing Vision scoring
- `docs/plans/2026-03-24-docs-discrawl-parity-status-report-plan.md` - Tier 1 vs Tier 2 analysis
- `docs/plans/2026-03-24-feat-full-steinberger-parity-plan.md` - rejects the Tier 1 line

### External References
- Stripe CLI: fixtures, trigger, listen - compound features from webhook/event shape
- AWS CLI: s3 sync vs s3api - two-tier architecture (wrapper + high-level)
- gh CLI: pr create compound command - bridges local git state with remote API
- Heroku CLI: pg:push/pg:pull, logs --tail - data sync + streaming
- Terraform: plan/apply/state - declarative management from CRUD APIs
- discrawl: SQLite + FTS5 + sync + search + tail - the gold standard for data-oriented CLIs
- gogcli: OAuth2 + multi-account + least-privilege + keyring - the gold standard for API CLIs
- Cloudflare: auto-generated Terraform provider from OpenAPI spec
