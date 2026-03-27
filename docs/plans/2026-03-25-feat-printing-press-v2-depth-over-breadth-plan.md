---
title: "feat: Printing Press v2 - Depth Over Breadth, Creative Over Mechanical"
type: feat
status: active
date: 2026-03-25
---

# Printing Press v2: Depth Over Breadth, Creative Over Mechanical

## The Two Failures

### Failure 1: Linear (GraphQL) - Generated Garbage That Can't Work
The press generated 151 commands against REST endpoints that don't exist. Linear is GraphQL-only. Every API call fails. 30 minutes spent polishing something fundamentally broken - fake spec, fake commands, scorecard that doesn't check if anything actually works.

**The missed opportunity:** When asked "could you have built features on top of the APIs that exist?", the answer was: `stale`, `orphans`, `velocity`, `burndown`, `standup`, `triage`, `blocked`, `sla-risk`, `close-stale`, `move-to-cycle`. Real tools for real engineering teams. The data is all there in Linear's GraphQL API - issues have `updatedAt`, `assignee`, `cycle`, `estimate`, `state`, `priority`, `relations`. Just needed to query it and present it usefully.

### Failure 2: Discord (REST) - Generated Breadth Without Depth
The press generated 316 commands covering the full Discord API. Compiles, runs, scores Grade A. But discrawl with 12 commands is more useful because it solves a real problem (archive + offline search) while our CLI is just a 316-endpoint HTTP client with nice flags.

**The missed opportunity:** Phase 0 research scored "local-first message sync + FTS5 search" at 14/16. Architecture decision said "SQLite + FTS5." Then the generator produced a pure API wrapper and Phase 4 polished the README instead of building the features the research demanded.

## Root Cause

**The press treats the generator output as the product.** It's not. The generator produces scaffolding. The product is what a power user would actually want to use. The printing press skill needs to think like osc-newfeature: "What would a power user wish this could do that the raw API can't?"

## Two Plans

---

# PLAN A: Creativity Engine - "What Would a Power User Build?"

## Problem

The press mechanically generates API wrappers. It doesn't think about what problems the API's customers actually have. For Linear, the answer isn't "expose every GraphQL mutation as a CLI flag" - it's "find stale issues, show burndown, automate triage." For Discord, it's not "316 REST commands" - it's "archive my server, search offline, monitor channels."

## Solution: Add Phase 0.5 - "Power User Workflows"

After Phase 0 (Visionary Research) identifies the API domain and competitive landscape, add a new phase that uses Claude's reasoning to predict what compound workflows power users need. This phase outputs a **Workflow Spec** that drives both what gets generated AND what gets hand-built in Phase 4.

### How It Works

#### Step 1: Domain Pattern Recognition

Classify the API into one of these archetypes based on Phase 0 research:

| API Archetype | Signal | Power User Needs | Examples |
|---|---|---|---|
| **Communication** | Messages, channels, threads, reactions | Archive, search, monitor, export conversations | Discord, Slack, Telegram |
| **Project Management** | Issues, tasks, sprints, states, assignees | Hygiene (stale/orphan), velocity, triage, standup, burndown | Linear, Jira, GitHub Issues |
| **Payments** | Charges, subscriptions, invoices, refunds | Reconciliation, webhook replay, fixture flows, revenue reports | Stripe, Plaid |
| **Infrastructure** | Servers, databases, deployments, logs | Sync state, tail logs, deploy orchestration, health checks | Heroku, Fly, AWS |
| **Content** | Documents, pages, blocks, media | Backup, sync to local files, diff, template management | Notion, Confluence |
| **CRM** | Contacts, deals, pipelines, activities | Pipeline reports, stale deal alerts, activity timelines, bulk updates | HubSpot, Salesforce |
| **Developer Platform** | Repos, PRs, issues, CI runs, packages | PR triage, CI monitoring, release management, dependency audit | GitHub, GitLab |

#### Step 2: Generate Workflow Ideas

For each archetype, the press has a library of **workflow templates** - not Go code templates, but *idea templates* that describe what compound commands to build. The LLM reasons about which workflows apply to this specific API.

Example for Project Management archetype (Linear):
```yaml
workflows:
  stale:
    description: "Find issues with no updates in N days"
    requires: [issues.list, issues.updatedAt field]
    implementation: "Query issues where updatedAt < (now - N days), group by team"
    value: "Engineering hygiene - prevent issue rot"

  orphans:
    description: "Find issues missing assignee, project, or cycle"
    requires: [issues.list, issues.assignee/project/cycle fields]
    implementation: "Query issues where assignee=null OR project=null, filterable by team"
    value: "Sprint planning - nothing falls through cracks"

  velocity:
    description: "Completed issues per cycle, trend over time"
    requires: [issues.list with state filter, cycles.list]
    implementation: "Count issues completed per cycle, compute rolling average"
    value: "Engineering metrics - track team throughput"

  standup:
    description: "What changed since yesterday for my team"
    requires: [issues.list with updatedAt filter, comments.list]
    implementation: "Issues updated in last 24h, grouped by assignee"
    value: "Daily standup replacement - async status"

  triage:
    description: "Batch-process untriaged issues"
    requires: [issues.list with state filter, issues.update]
    implementation: "List issues in Triage state, prompt for priority/assignee/cycle per issue"
    value: "Triage workflow - process inbox to zero"

  burndown:
    description: "Current cycle progress vs ideal line"
    requires: [cycles.get with issues, issues.state/completedAt]
    implementation: "Count total vs completed issues, plot progress curve"
    value: "Sprint health - are we on track?"
```

#### Step 3: Validate Against API Capabilities

For each workflow idea, check if the API actually supports it:
- Does the API have the required endpoints?
- Can the required fields be queried/filtered?
- Is pagination available for list operations?
- Are write operations available (for triage/bulk-update workflows)?

For GraphQL APIs: check if the schema has the required types and fields.
For REST APIs: check if the endpoints exist and have the required params.

Drop workflows that can't be implemented. Rank remaining by user impact.

#### Step 4: Output the Workflow Spec

The workflow spec becomes the PRIMARY input to Phase 4 (GOAT Fix). Instead of polishing README descriptions, Phase 4 implements the top 5-7 workflows as real commands.

### Changes to the SKILL.md

**New Phase 0.5 section** (between Phase 0 and Phase 1):

```markdown
# PHASE 0.5: POWER USER WORKFLOWS

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 0.5a: Classify the API Archetype
Based on Phase 0 research, classify the API into one or more archetypes:
Communication, Project Management, Payments, Infrastructure, Content, CRM, Developer Platform.

### Step 0.5b: Generate Workflow Ideas
For the identified archetype(s), brainstorm 10-15 compound workflows that would
solve real problems for the API's core customers. Use the LLM to reason about:
- "What does a power user of this API wish they could do in one command?"
- "What multi-step workflow do people automate with scripts today?"
- "What reporting/hygiene/monitoring task requires manual effort?"

### Step 0.5c: Validate Against API
Check each workflow against the API's actual capabilities (REST endpoints, GraphQL
schema, or developer docs). Drop anything the API can't support.

### Step 0.5d: Rank by Impact
Score each workflow on: frequency of use, pain severity, implementation feasibility,
uniqueness (no existing tool does this).

### Step 0.5e: Select Top 5-7 for Implementation
These become mandatory Phase 4 work items. They are NOT optional polish.
They are the PRODUCT.

### PHASE GATE 0.5
Tell the user: "Identified [N] power-user workflows for [API]. Top 5:
1. [name] - [one-line description]
2. ...
These will be built in Phase 4 alongside the API wrapper."
```

**Modified Phase 4** - currently says "Execute fixes in priority order: scorecard-gap fixes, complex body field examples, command name cleanup, description/README polish." Change to:

```markdown
# PHASE 4: GOAT BUILD

## Priority 1: Power User Workflows (from Phase 0.5)
Implement the top 5-7 workflows as real commands. These are hand-written Go code
that uses the generated API client. Each workflow:
- Has a dedicated command file (e.g., internal/cli/stale.go)
- Uses the generated client to make API calls
- Combines multiple API calls into one user-facing operation
- Has realistic examples in --help
- Outputs structured data (--json) for agent consumption

## Priority 2: Scorecard-Gap Fixes
Fix README, breadth, descriptions - same as today.

## Priority 3: Polish
README cookbook showcasing the workflow commands, not just API calls.
```

### Changes to the Generator

**New `internal/generator/templates/workflow_custom.go.tmpl`** - a template for hand-written workflow commands that:
- Imports the generated client
- Has proper Cobra command structure
- Has --json/--select/--dry-run flags inherited from root
- Has a `RunE` body that's a TODO placeholder with the workflow description as a comment

The generator outputs these as scaffolded files. The LLM (Phase 3 synthesis or Phase 4 GOAT build) fills in the actual implementation.

### Changes to the Scorecard

**New dimension: "Workflows" (0-10)** - measures compound commands that combine multiple API calls:
- Count files in `internal/cli/` that contain 2+ client method calls (Get, Post, Put, Delete) in a single RunE
- Score: 0 for API-wrapper-only, 5 for 2-3 compound commands, 10 for 5+ compound commands
- This directly rewards building the kind of tools discrawl and osc-newfeature create

---

# PLAN B: API Type Detection - "Don't Generate Garbage"

## Problem

The press blindly generates REST clients for any spec. When pointed at a GraphQL API (Linear), it generates 151 commands against endpoints that don't exist. It should either (a) refuse gracefully, (b) adapt to GraphQL, or (c) build value-add tools without wrapping the API.

## Solution: Pre-Generation Intelligence

### Step 1: API Type Detection

Before generating anything, detect the API type:

```go
// internal/profiler/apitype.go

type APIType string

const (
    APITypeREST    APIType = "rest"     // Has OpenAPI spec, REST endpoints
    APITypeGraphQL APIType = "graphql"  // GraphQL schema, single POST endpoint
    APITypeGRPC    APIType = "grpc"     // Protocol Buffers, gRPC services
    APITypeUnknown APIType = "unknown"  // Can't determine
)

func DetectAPIType(specPath string) APIType {
    // Check file extension and content
    // .graphql, .gql -> GraphQL
    // Contains "openapi" or "swagger" -> REST
    // Contains "syntax = \"proto" -> gRPC
    // URL ending in /graphql -> GraphQL
}
```

### Step 2: GraphQL Mode

When a GraphQL API is detected, DON'T generate REST commands. Instead:

1. **Parse the GraphQL schema** to extract types, queries, mutations
2. **Generate a GraphQL client** (internal/client/graphql.go) that:
   - Sends POST requests to the single /graphql endpoint
   - Builds query strings from command flags
   - Handles pagination (Linear uses cursor-based Relay pagination)
   - Handles auth (bearer token in Authorization header)
3. **Generate commands from queries and mutations** (not from fake REST paths)
   - Each GraphQL query becomes a `get`/`list` command
   - Each GraphQL mutation becomes a `create`/`update`/`delete` command
   - Arguments map to flags, return types map to --select fields
4. **Still run the profiler and vision pipeline** - GraphQL APIs need sync/search/export too

### Step 3: "No Spec" Mode

When no spec exists at all (or the spec is for the wrong API type), the press should:

1. **Fetch the API docs** and extract endpoints
2. **Build a spec from docs** (existing docspec package handles this)
3. **Or skip generation entirely** and go straight to Phase 0.5 workflows

This is the Linear scenario done right: instead of faking a REST spec, either:
- Build a real GraphQL client from the schema, OR
- Skip generation and hand-build the 10 workflow commands that actually matter

### Changes to SKILL.md

**New section at the top of Phase 2:**

```markdown
### Step 2.0: API Type Check

Before generating, verify the spec matches the API:

1. If spec is OpenAPI/Swagger -> proceed to REST generation
2. If spec is GraphQL schema -> switch to GraphQL generation mode
3. If no spec and API is GraphQL-only -> STOP. Tell the user:
   "This API is GraphQL-only. The printing press currently generates REST CLIs.
   Options:
   a) I can write the workflow commands directly (Phase 0.5 workflows) without
      wrapping the full API
   b) I can write a GraphQL spec and generate a GraphQL-native CLI (experimental)
   c) Pick a different REST API"

4. If the spec has endpoints that don't match the actual API base URL -> STOP.
   "The spec describes REST endpoints but the API is at [url]/graphql.
   This would generate commands that can't make successful API calls."

NEVER generate a CLI that can't make a single successful API call.
```

**New anti-shortcut rule:**

```markdown
- "The API is GraphQL-only but I'll write a REST spec anyway" (STOP - this produces garbage)
```

### Changes to the Generator

**New `internal/generator/graphql_generator.go`** that:
- Reads a GraphQL schema file
- Extracts queries and mutations
- Generates Go commands using a GraphQL HTTP client (POST /graphql with query body)
- Maps GraphQL arguments to Cobra flags
- Maps GraphQL return types to --select fields
- Handles Relay-style cursor pagination

**Modified `internal/spec/spec.go`** - add `APIType` field:
```go
type APISpec struct {
    Name      string
    APIType   string  // "rest", "graphql", "grpc"
    // ... existing fields
}
```

### Changes to the Scorecard

**New check: "Functional" (pass/fail gate)**
- Before scoring any dimension, verify the CLI can make at least one successful API call
- If the binary exists, try `<cli> doctor --json` and check if `api` field is NOT "degraded"
- If it's degraded with no token set, that's fine (auth issue, not structural)
- If it's degraded because the endpoint doesn't exist (404 on base URL), FAIL the entire scorecard
- Display: "FAIL: API unreachable - CLI may be targeting non-existent endpoints"

This single check would have caught the Linear failure immediately.

---

# Implementation Priorities

## Must-Have (Without These, the Press Still Disappoints)

1. **Phase 0.5 workflow generation** in the SKILL.md - the creative engine
2. **API type detection** - refuse to generate garbage for GraphQL APIs
3. **Scorecard "Functional" gate** - catch broken CLIs before scoring
4. **Modified Phase 4** - workflows first, polish second
5. **Scorecard "Workflows" dimension** - reward compound commands

## Should-Have (Makes It Great)

6. **GraphQL generation mode** - native GraphQL client generation
7. **Workflow template library** - per-archetype workflow ideas
8. **LLM-powered workflow implementation** - use Claude to write the compound command bodies
9. **Domain-specific README cookbook** - generated from workflow commands, not API endpoints

## Nice-to-Have (Makes It Legendary)

10. **Workflow validation** - test that workflow commands actually work against the live API
11. **Interactive triage mode** - for project management APIs, TUI for batch processing
12. **Declarative state mode** - for infrastructure APIs, terraform-like plan/apply

---

## Acceptance Criteria

### Plan A (Creativity Engine)
- [ ] Phase 0.5 section added to SKILL.md with archetype classification and workflow generation
- [ ] Phase 4 rewritten to prioritize workflows over polish
- [ ] Running `/printing-press Linear` produces 5+ workflow commands (stale, orphans, velocity, standup, triage) even though Linear is GraphQL-only
- [ ] Running `/printing-press Discord` produces compound commands (archive, search, monitor) on top of the API wrapper
- [ ] Scorecard has "Workflows" dimension that measures compound commands
- [ ] A CLI with 10 workflow commands scores higher than a CLI with 300 API wrappers and zero workflows

### Plan B (API Type Detection)
- [ ] `DetectAPIType()` correctly identifies REST, GraphQL, gRPC, and unknown APIs
- [ ] Running `/printing-press Linear` does NOT generate 151 fake REST commands
- [ ] Running `/printing-press Linear` either generates a GraphQL client OR builds workflow commands directly
- [ ] Scorecard has "Functional" gate that fails CLIs targeting non-existent endpoints
- [ ] SKILL.md has anti-shortcut rule against faking specs for GraphQL APIs

### Combined
- [ ] A fresh `/printing-press Discord` run produces BOTH API coverage AND discrawl-competitive workflow depth
- [ ] A fresh `/printing-press Linear` run produces something useful, not garbage
- [ ] The press thinks like osc-newfeature: "What would a power user wish this could do?"

## Sources

### Internal
- This conversation's postmortem (discord-cli failure analysis)
- Linear session transcript (GraphQL garbage generation)
- Discord session transcript (breadth without depth, discrawl comparison)
- `docs/plans/2026-03-25-feat-api-shape-intelligence-engine-plan.md` - profiler + vision pipeline (already built)
- `docs/plans/2026-03-24-docs-discrawl-parity-status-report-plan.md` - Tier 1 vs Tier 2 analysis
- `internal/profiler/profiler.go` - API Shape Profiler (Phases 1-2 done)
- `internal/llmpolish/vision.go` - LLM Vision Synthesis (Phase 3 done)

### External
- discrawl (steipete/discrawl) - 12 commands, more useful than 316
- schpet/linear-cli - "second most used CLI after git" (HN)
- Stripe CLI fixtures/trigger/listen - compound workflow design
- AWS s3 vs s3api - two-tier architecture (wrapper + high-level)
- gh CLI pr create - bridges local git state with remote API
