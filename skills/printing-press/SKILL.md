---
name: printing-press
description: Generate the GOAT CLI for any API. 5-phase loop with dual Steinberger analysis, deep competitor research, complex body field handling, and before/after scoring delta.
version: 1.1.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - WebSearch
  - AskUserQuestion
  - Agent
---

# /printing-press

Generate the best CLI that has ever existed for any API. Five mandatory phases. Dual Steinberger analysis. No shortcuts.

```
/printing-press Notion
/printing-press Plaid payments API
/printing-press --spec ./openapi.yaml
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at `~/cli-printing-press`
- Build binary if missing: `cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press`

## How This Works

Every run produces the GOAT CLI through 8 mandatory phases + 7 comprehensive plan documents:

```
PHASE 0 -> PHASE 0.5 -> PHASE 0.7 -> PHASE 1 -> PHASE 2 -> PHASE 3 -> PHASE 4 -> PHASE 4.5 -> PHASE 5
(3-5m)     (2-3m)       (15-25m)     (5-8m)     (1-2m)     (5-8m)     (5-10m)    (10-20m)      (2-3m)
Visionary  Workflows    Prediction   Research   Generate   Audit      Build      Dogfood       Final
Research   (commands)   Engine       (specs)    (code)     (review)   (fixes)    Emulation     Steinberger
                        (data layer)                                             (spec-test)
```

Total expected time: 45-85 minutes. Phase 4.5 tests every command against spec-derived mocks.

**7 Plan Artifacts Per Run:**

Every phase gate produces a comprehensive plan document in `~/cli-printing-press/docs/plans/`:

```
Phase 0   -> <today>-feat-<api>-cli-visionary-research.md
Phase 0.5 -> <today>-feat-<api>-cli-power-user-workflows.md
Phase 0.7 -> <today>-feat-<api>-cli-data-layer-spec.md
Phase 1   -> <today>-feat-<api>-cli-research.md
Phase 3   -> <today>-fix-<api>-cli-audit.md
Phase 4   -> <today>-fix-<api>-cli-goat-build-log.md
Phase 4.5 -> <today>-fix-<api>-cli-dogfood-report.md
```

Each artifact chains into the next. **Read the previous phase's artifact before starting the next phase.**

**The Steinberger bar:** Peter Steinberger's gogcli is the 10/10 reference. Every generated CLI is scored against it TWICE - once during audit to find gaps, once after fixes to prove improvement. The delta is the proof of work.

**Grade thresholds (10 dimensions, 100 max):**
- **Grade A:** 80+/100 (80%)
- **Grade B:** 65-79/100 (65-79%)
- **Grade C:** 50-64/100 (50-64%)

---

## Artifact Writing: Plan Generation at Each Phase Gate

At the end of each phase, write a comprehensive plan document. This is NOT optional - the artifacts ARE the product.

**Option A: /ce:plan is available (compound-engineering plugin installed)**

Try to invoke the `compound-engineering:ce:plan` skill. If it exists, use it:

```
Skill tool: compound-engineering:ce:plan
Args: "<phase description with all research gathered so far>"
```

The /ce:plan skill produces a full plan document with frontmatter, analysis, acceptance criteria, and sources. Pass it all the research from this phase as the feature description.

**Option B: Built-in plan writer (fallback when compound-engineering is NOT installed)**

If /ce:plan is not available, write the artifact yourself with this structure:

```markdown
---
title: "<Phase Name>: <API> CLI"
type: feat
status: active
date: <today>
phase: "<phase number>"
api: "<api name>"
---

# <Phase Name>: <API> CLI

## Overview
[2-3 paragraph executive summary of what this phase discovered/decided]

## Analysis
[Full analysis with tables, scores, evidence URLs, and reasoning]
[Every claim backed by evidence - WebSearch URLs, star counts, API docs]
[Scoring breakdowns showing how each number was computed]

## Decisions
[What was decided and WHY - rationale for each decision]
[What was rejected and WHY]

## Concrete Outputs
[SQL schemas, command definitions, sync strategies - real code, not pseudocode]
[Every output validated against the actual API]

## Acceptance Criteria
- [ ] [Measurable criteria for this phase's outputs]

## Sources
- [URLs, file paths, competitor repos with star counts]
```

**CRITICAL:** The built-in writer must match /ce:plan depth:
- Full analysis, not bullet summaries
- Evidence with source URLs, not assertions
- Scoring breakdowns, not just final numbers
- Concrete SQL/code examples, not pseudocode
- Validation proof ("I verified the API supports ?after= filtering"), not assumptions
- 200+ lines minimum per artifact

---

## Workflow: `--spec` shortcut

When the user provides `--spec <path-or-url>`, skip Phase 1 spec search (spec is provided). STILL run competitor research (Steps 1.2-1.5). Run all other phases.

## Workflow: Natural Language (Primary)

When the user provides an API name, run ALL five phases.

### Step 0: Parse intent and check known specs

Extract the API name. Check `~/cli-printing-press/skills/printing-press/references/known-specs.md`.

If found in registry: note the URL for Phase 2, but STILL run full Phase 1 research.
If not found: Phase 1 searches for spec too.

---

# PHASE 0: VISIONARY RESEARCH

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Before generating any CLI, understand what a thoughtful developer would build - not just what the OpenAPI spec says.

### Step 0a: API Identity & Domain Understanding

Understand what this API IS:

1. **WebFetch** the API's developer docs landing page
2. **WebSearch**: `"<API name>" developer documentation overview`
3. Extract:
   - **Domain category:** messaging, payments, productivity, infrastructure, analytics
   - **Primary users:** Who uses this API? (e.g., "bot developers", "server admins")
   - **Core entities:** What are the main objects? (e.g., "guilds", "channels", "messages")
   - **Data profile:**
     - Write pattern: append-only, mutable, or event-sourced?
     - Volume: high (millions of records), medium, or low?
     - Real-time: does the API have webhooks/websockets/SSE?
     - Search need: high (users need to find things) or low?

### Step 0b: Usage Pattern Discovery

Discover what people ACTUALLY DO with this API:

**Community Research (run these in parallel):**
1. **WebSearch**: `"<API name>" CLI workflow site:reddit.com`
2. **WebSearch**: `"<API name>" automation script site:github.com`
3. **WebSearch**: `"<API name>" "I built" OR "I made" OR "my tool" site:reddit.com OR site:news.ycombinator.com`
4. **WebSearch**: `"<API name>" tutorial automation workflow 2025 2026`

**Pain Point Research:**
5. **WebSearch**: `"<API name>" API "pain point" OR "limitation" OR "workaround"`
6. **WebSearch**: `site:stackoverflow.com "<API name>" API rate limit OR pagination OR bulk`

From all research, identify the **top 5 usage patterns** ranked by evidence score:

| Source | Weight |
|---|---|
| Existing tool with 100+ stars | 3 points |
| Existing tool with 10-99 stars | 2 points |
| Reddit/HN post with 50+ upvotes | 2 points |
| Reddit/HN post with 10-49 upvotes | 1 point |
| Stack Overflow question with 10+ votes | 1 point |
| Blog post / tutorial | 1 point |
| GitHub issue on competitor | 1 point |
| Cross-platform appearance (same need on 2+ platforms) | +2 bonus |

Score >= 6: Strong evidence. Include in CLI.
Score 3-5: Moderate evidence. Consider as optional.
Score < 3: Weak evidence. Skip.

### Step 0c: Tool Landscape Discovery (The Discrawl Finder)

Find ALL tools for this API, not just API wrappers:

**Tier 1: Direct CLI Search** (existing Phase 1 does this too)
1. **WebSearch**: `"<API name>" CLI tool github`

**Tier 2: Non-Wrapper Tool Search** (CRITICAL - finds discrawl-class tools)
2. **WebSearch**: `"<API name>" sync OR archive OR mirror site:github.com`
3. **WebSearch**: `"<API name>" search engine OR analytics OR dashboard site:github.com`
4. **WebSearch**: `"<API name>" backup OR export OR migration site:github.com`
5. **WebSearch**: `"<API name>" monitor OR watcher OR alerting site:github.com`

**Tier 3: Ecosystem Search**
6. **WebSearch**: `awesome "<API name>" site:github.com`

For each tool found, classify it:

| Type | Description | Example |
|---|---|---|
| **API Wrapper** | Translates HTTP to CLI flags | discli |
| **Data Tool** | Adds local persistence/search | discrawl |
| **Workflow Tool** | Orchestrates multi-step sequences | Stripe fixtures |
| **Environment Tool** | Runs local simulation | Supabase CLI |
| **Integration Tool** | Bridges to other systems | Zapier integration |

**The press should generate CLIs that compete with Data Tools and Workflow Tools, not just API Wrappers.**

### Step 0d: Workflow Analysis

From usage patterns (0b) and tool landscape (0c), identify multi-step workflows:

For each workflow, document:
- **Steps:** The sequence of API calls
- **Frequency:** How often users perform this
- **Pain point:** What makes this hard with the raw API
- **Proposed CLI feature:** What compound command would solve it

### Step 0e: Architecture Planning

Based on data profile and workflows, decide what the CLI needs:

| Data Profile | Architecture |
|---|---|
| High volume + search need | SQLite + FTS5 (Discord, Slack) |
| Transaction data + reconciliation | Local ledger with diff tracking (Stripe, Plaid) |
| Document data + offline editing | Local Markdown/JSON sync (Notion, Confluence) |
| Low volume + simple CRUD | Standard API wrapper is fine (most APIs) |

For each decision area (persistence, real-time, search, bulk, caching), document:
- **Need level:** High / Medium / Low
- **Decision:** What to use
- **Rationale:** Why

### Step 0f: Feature Ideation - "Next 5 Features for the World"

Score each feature idea on 8 dimensions (16-point max):

| Dimension | Weight | Scoring |
|---|---|---|
| **Evidence strength** | 3 | 3=existing tool 100+ stars, 2=Reddit/SO demand, 1=weak, 0=speculation |
| **User impact** | 3 | 3=most users feel this pain, 2=niche, 1=nice-to-have, 0=nobody asked |
| **Implementation feasibility** | 2 | 2=can generate template, 1=needs custom code, 0=major infrastructure |
| **Uniqueness** | 2 | 2=no existing tool, 1=improves on existing, 0=already well-served |
| **Composability** | 2 | 2=great with pipes/agents, 1=somewhat, 0=interactive-only |
| **Data profile fit** | 2 | 2=perfect fit, 1=possible, 0=wrong shape |
| **Maintainability** | 1 | 1=generated code supports it, 0=needs human maintenance |
| **Competitive moat** | 1 | 1=hard to replicate, 0=trivial |

Score >= 12: **Must-have.** Build it.
Score 8-11: **Should-have.** Include as optional.
Score < 8: **Won't-have.** Skip or future work.

### Step 0g: Write the Visionary Research Artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-visionary-research.md`:

```markdown
## Visionary Research: <API> CLI

### API Identity
- Domain: <category>
- Primary users: <who>
- Data profile: <write pattern>, <volume>, <realtime>, <search need>

### Usage Patterns (Top 5 by Evidence)
1. <pattern> (Evidence: X/10) - <what it needs>
2. ...

### Tool Landscape (Beyond API Wrappers)
- <tool> (<stars> stars): <what it does>
- ...

### Workflows
1. <name>: <steps> -> Proposed: `<api>-cli <command>`
2. ...

### Architecture Decisions
- Persistence: <decision> because <rationale>
- Real-time: <decision> because <rationale>
- Search: <decision> because <rationale>
- Bulk: <decision> because <rationale>
- Cache: <decision> because <rationale>

### Top 5 Features for the World
1. <feature> (Score: X/16) - <1-line description>
2. ...
```

### PHASE GATE 0

**STOP.** Verify ALL of these before proceeding:
1. API Identity documented with data profile
2. At least 3 usage patterns with evidence scores
3. Tool landscape includes non-wrapper tools (Tier 2 search done)
4. At least 2 workflows with proposed CLI features
5. Architecture decisions match data profile
6. Top 5 features scored and ranked

**Write Phase 0 Artifact:** Run the Artifact Writing plan generator (see top of skill) with all Phase 0 research as input. Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-visionary-research.md`. Include: API identity, data profile, usage patterns with evidence, tool landscape, architecture decisions, top 5 features with full scoring.

Tell the user: "Phase 0 complete: Domain: [category]. Data profile: [volume]/[realtime]/[search]. Found [N] non-wrapper tools. Top feature: [name] (score [X]/16). Architecture: [key decision]. Proceeding to power user workflows."

---

# PHASE 0.5: POWER USER WORKFLOWS

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

The generator produces API wrappers. Power users want workflow tools. This phase predicts what compound commands would make the CLI genuinely useful - the kind of features that make discrawl (12 commands) more valuable than a 316-command API wrapper.

### Step 0.5a: Classify the API Archetype

Based on Phase 0 research, classify the API:

| Archetype | Signal | Example Workflows |
|---|---|---|
| **Communication** | Messages, channels, threads | Archive, offline search, monitor keywords, export conversations |
| **Project Management** | Issues, tasks, sprints, states | Stale issues, orphan detection, velocity, burndown, standup, triage |
| **Payments** | Charges, subscriptions, invoices | Reconciliation, webhook replay, fixture flows, revenue reports |
| **Infrastructure** | Servers, deployments, logs | State sync, log tailing, deploy orchestration, health dashboards |
| **Content** | Documents, pages, blocks, media | Backup to local files, diff, template management, publish workflows |
| **CRM** | Contacts, deals, pipelines | Pipeline reports, stale deal alerts, activity timelines, bulk updates |
| **Developer Platform** | Repos, PRs, CI runs | PR triage, CI monitoring, release management, dependency audit |

### Step 0.5b: Generate 10-15 Workflow Ideas

For the identified archetype, brainstorm compound workflows. Ask:
- "What does a power user of this API wish they could do in one command?"
- "What multi-step task do people automate with scripts today?"
- "What reporting/hygiene/monitoring task requires manual effort?"
- "What would make an engineering manager's life easier?"

Each workflow should:
- Combine 2+ API calls into one operation
- Solve a real recurring problem
- Be expressible as a single CLI command with flags

### Step 0.5c: Validate Against API Capabilities

For each workflow idea, check:
1. Does the API have the required endpoints/fields?
2. Can the required data be queried/filtered?
3. Are write operations available (for mutation workflows)?
4. For GraphQL APIs: does the schema have the required types?

Drop workflows the API can't support.

### Step 0.5d: Rank by Impact

Score each workflow on:
- **Frequency**: How often would users run this? (daily=3, weekly=2, monthly=1)
- **Pain**: How painful is the manual alternative? (high=3, medium=2, low=1)
- **Feasibility**: How hard to implement? (easy=3, medium=2, hard=1)
- **Uniqueness**: Does any existing tool do this? (no=3, partial=2, yes=0)

### Step 0.5e: Select Top 5-7 for Implementation

These become **mandatory Phase 4 work items**. They are NOT optional polish. They are the PRODUCT.

### PHASE GATE 0.5

**STOP.** Tell the user: "Identified [N] power-user workflows for [API name]. Top 5:
1. [name] - [one-line description] (score [X]/12)
2. ...
These will be built as real commands in Phase 4, alongside the API wrapper."

**Write Phase 0.5 Artifact:** Run the Artifact Writing plan generator (see top of skill) with all Phase 0.5 analysis as input. Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-power-user-workflows.md`. Include: API archetype, all 10-15 workflow ideas, validation results, full scoring table, top 7 with implementation notes.

---

# PHASE 0.7: POWER USER PREDICTION ENGINE

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

The generator produces API wrappers. Power users want a local data layer - domain-specific SQLite tables, incremental sync, full-text search with domain filters, raw SQL access, and trend detection. This phase predicts that data layer from the API surface + social signals, WITHOUT looking at competitors (that's Phase 1's job).

**Read the Phase 0 and Phase 0.5 artifacts before starting this phase.**

### Step 0.7a: Entity Classification

Map every API resource into one of four types by reading the OpenAPI spec or Phase 0's entity list:

| Type | Signal | Example | Persistence Need |
|---|---|---|---|
| **Accumulating** | Grows over time, has timestamps, paginated lists | Messages, Issues, Audit Logs, Commits | SQLite table + incremental sync |
| **Reference** | Changes rarely, small cardinality, referenced by other entities | Users, Teams, Roles, Labels, Channels | SQLite table + periodic refresh |
| **Append-only** | Never edited, only created | Events, Webhooks, Notifications | SQLite table + tail command |
| **Ephemeral** | Short-lived, not worth persisting | OAuth tokens, Rate limit status, Gateway info | API-only, no persistence |

**Heuristics:**
- Has `created_at`/`timestamp` + paginated list endpoint -> Accumulating
- Referenced by 3+ other entities via `_id` fields -> Reference
- Has no UPDATE/PATCH endpoint -> Append-only
- No list endpoint or < 100 expected records -> Ephemeral
- Has `updated_at` or `modified_at` -> needs incremental sync cursor

**Output:** Entity classification table with type, estimated volume, update frequency, and key temporal field for ALL API resources.

### Step 0.7b: Social Signal Mining for Data Patterns

Find evidence of what data power users actually store locally. Run 7 parallel WebSearches:

1. **WebSearch**: `"<API name>" export OR backup OR archive site:github.com`
2. **WebSearch**: `"<API name>" SQLite OR database OR local site:github.com`
3. **WebSearch**: `"<API name>" analytics OR dashboard OR metrics site:github.com`
4. **WebSearch**: `"<API name>" "I wish" OR "would be nice" OR "feature request" data`
5. **WebSearch**: `"<API name>" offline OR search OR "full text" site:reddit.com OR site:news.ycombinator.com`
6. **WebSearch**: `"<API name>" trend OR pattern OR anomaly detection`
7. **WebSearch**: `"<API name>" graph OR visualization OR dependency`

**For each finding, extract:**
- What entities they store locally
- What queries they run (joins, aggregations, time filters)
- What temporal patterns they track (trends, anomalies, velocity)
- What cross-entity relationships they need

**Score using Phase 0 evidence framework.** Anything with score >= 6 informs the data layer.

### Step 0.7c: Data Gravity Scoring

Rank entities by how much value they'd have in a local SQLite database.

**Formula:** `DataGravity = Volume(0-3) + QueryFrequency(0-3) + JoinDemand(0-2) + SearchNeed(0-2) + TemporalValue(0-2)`

| Factor | 0 | 1 | 2 | 3 |
|---|---|---|---|---|
| **Volume** | < 100 records | 100-10k | 10k-1M | > 1M |
| **QueryFrequency** | Rarely queried | Monthly | Weekly | Daily |
| **JoinDemand** | No references | 1-2 entities reference it | 3-4 | 5+ |
| **SearchNeed** | No text fields | 1 text field | 2-3 text fields | Primary text content |
| **TemporalValue** | No time dimension | Created date only | Updated + trends | Core to time-series analysis |

**Thresholds:**
- Score >= 8: **Primary entity** - gets its own SQLite table with proper columns, FTS5 if text-heavy
- Score 5-7: **Support entity** - gets a simpler table
- Score < 5: **API-only** - no local persistence

Score EVERY entity from Step 0.7a. Show the full breakdown.

### Step 0.7d: Schema + Sync + Search Specification

For each Primary entity (score >= 8), produce:

**1. SQLite Schema:**
- Extract columns from the API's response schema (NOT just id + JSON blob)
- Include foreign key columns for joins (e.g., `channel_id`, `author_id`)
- Include the temporal field for sync cursors
- Add indexes on foreign keys and temporal fields
- Create FTS5 virtual table on text fields (title, description, content, body, name)
- Keep a `data JSON NOT NULL` column for the full API response

**2. Sync Strategy:**
- Identify the incremental sync cursor (timestamp field, snowflake ID, cursor pagination)
- **VALIDATE:** Check that the API supports filtering by this cursor - look for `since`, `after`, `updated_after`, `before` query params in the spec
- Determine batch size from API's max `limit` parameter
- Check if API has WebSocket/SSE/Gateway - note for tail command
- If the API doesn't support cursor filtering, fall back to full sync + local dedup

**3. Search Specification:**
- List which text fields to extract into FTS5
- Define domain-specific search filters as SQL WHERE clauses
- Map CLI flags to SQL: `--channel` -> `WHERE channel_id = ?`, `--author` -> `WHERE author_id = ?`
- **VALIDATE:** Confirm these fields actually appear in API list/get responses

**4. Compound Queries:**
- Define 3-5 cross-entity queries (e.g., "messages by author in channel in last N days")
- Validate join columns exist in both tables
- These become Phase 4 workflow commands that use local DB instead of live API

**5. Tail Strategy:**

| Method | When to Use |
|---|---|
| WebSocket/Gateway | API has it (Discord Gateway, Slack RTM) |
| SSE | API has it (GitHub, Linear webhooks) |
| REST Polling | Fallback - GET with ?since= cursor |

Decide which method this API supports. If WebSocket/SSE, the tail command should use it instead of REST polling.

### Step 0.7e: Write the Data Layer Specification Artifact

**Run the Artifact Writing plan generator** (see top of skill) with all Phase 0.7 analysis as input. Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-data-layer-spec.md`.

The artifact MUST include:
- Entity classification table for every API resource
- Data gravity scores with full breakdown per entity
- Complete SQLite schema (CREATE TABLE + CREATE INDEX + FTS5)
- Sync strategy with cursor validation proof
- Domain-specific search filters mapped to SQL WHERE clauses
- 3-5 compound cross-entity queries
- Tail strategy decision with justification
- Commands to build in Phase 4 Priority 0

### PHASE GATE 0.7

**STOP.** Verify ALL of these before proceeding:
1. Entity classification table with type and volume estimates for every API resource
2. At least 3 social signals with evidence scores >= 6
3. Data gravity scores computed, with >= 1 primary entity (score >= 8)
4. SQLite schema with proper columns (NOT generic JSON blobs) for each primary entity
5. FTS5 virtual tables for entities with text fields
6. Sync strategy with cursor fields validated against actual API filter params
7. Domain-specific search filters mapped to SQL WHERE clauses
8. At least 3 compound queries validated (joins work, columns exist)

Tell the user: "Phase 0.7 complete: [N] primary entities for SQLite ([list]), [M] compound queries validated. Sync via [cursor type]. FTS5 on [fields]. Key prediction: [most valuable data-layer feature]. Proceeding to deep research."

---

# PHASE 1: DEEP RESEARCH

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Research the API landscape deeply. You need to understand the competitive terrain, user pain points, and strategic opportunity before generating anything.

### Step 1.1: Search for the OpenAPI spec

If not in known-specs registry:

1. **WebSearch**: `"<API name>" openapi spec site:github.com`
2. **WebSearch**: `"<API name>" openapi.yaml OR openapi.json specification`
3. Try common URL patterns
4. If found, **WebFetch** first 500 bytes to verify

If no spec found: plan to write one from docs in Phase 2.

### Step 1.2: Search for competing CLIs

**WebSearch**: `"<API name>" CLI tool github`
**WebSearch**: `"<API name>" command line client`

Also search for non-wrapper tools discovered in Phase 0:
**WebSearch**: `"<API name>" sync OR archive OR export site:github.com`

For each competitor found, note repo URL, star count, language.

### Step 1.3: Deep competitor analysis (TOP 2 competitors)

For the top 2 competitors by stars, do ALL of the following:

1. **WebFetch** their README - count commands, note features, assess quality
2. **WebFetch** their GitHub repo main page - check:
   - Last commit date (is it maintained?)
   - Open issue count
   - Number of contributors
3. **WebSearch**: `site:github.com/<org>/<repo>/issues` - look for:
   - User complaints about missing features
   - Requests for specific functionality
   - Pain points users report
4. Record at least 2 specific user quotes or pain points.

### Step 1.4: Check demand signals

**WebSearch**: `"<API name>" "need a CLI" OR "command line" site:reddit.com OR site:news.ycombinator.com`

### Step 1.5: Strategic justification

Answer this question explicitly: **"Why should this CLI exist when [best competitor] already has [N] stars?"**

The answer must be SPECIFIC. Not just "agent-native." Examples:
- "[Competitor] hasn't been updated since [date] and doesn't support the latest API version"
- "No existing CLI supports --json + --select + --dry-run for agent workflows"
- "Users on [platform] are asking for [specific feature] which no CLI provides"
- "[Competitor] has [N] open issues about [problem] we can solve"

### Step 1.6: Write the research artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-research.md`:

```markdown
---
title: "Research: <API> CLI"
type: feat
status: active
date: <today>
---

# Research: <API> CLI

## Spec Discovery
- Official OpenAPI spec: <url or "none found - will write from docs">
- Source: <where found>
- Format: <OpenAPI 3.x / Swagger 2.0 / internal YAML>
- Endpoint count: <N>

## Competitors (Deep Analysis)

### <Competitor 1> (<stars> stars)
- Repo: <url>
- Language: <lang>
- Commands: <count>
- Last commit: <date>
- Open issues: <count>
- Maintained: <yes/no>
- Notable features: <list>
- Weaknesses: <what users complain about>

### <Competitor 2> (<stars> stars)
- [same structure]

## User Pain Points
> "<quote from GitHub issue or Reddit>" - <source>
> "<quote>" - <source>

## Auth Method
- Type: <api_key / oauth2 / bearer_token>
- Env var convention: <what competitors use>

## Demand Signals
- <specific posts with URLs, or "none found">

## Strategic Justification
**Why this CLI should exist:** <specific answer, not just "agent-native">

## Target
- Command count: <N - match or beat best competitor>
- Key differentiator: <specific features we'll have that competitors don't>
- Quality bar: Steinberger Grade A (80+/100)
```

### PHASE GATE 1

**STOP.** Verify ALL of these before proceeding:
1. Research artifact exists with Spec Discovery section
2. At least 2 competitors analyzed with maintenance status
3. At least 2 user quotes or pain points documented
4. Strategic justification answers "why should this exist?"
5. Target command count is set

**Write Phase 1 Artifact:** Run the Artifact Writing plan generator with all Phase 1 research as input. Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-research.md`. Include: spec discovery, deep competitor analysis with quotes, demand signals, strategic justification, target command count.

Tell the user: "Phase 1 complete: Found [spec/no spec], [N] competitors. Best: [name] ([stars] stars, [commands] commands, last commit [date]). Strategic angle: [1-sentence justification]. Proceeding to generation."

---

# PHASE 2: GENERATE

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 2.0: API Type Check

Before generating, verify the spec matches the API:

1. **If spec is OpenAPI/Swagger** -> proceed to REST generation (Step 2.1)
2. **If spec is a GraphQL schema** -> STOP. Tell the user:
   "This API is GraphQL-only. The printing press generates REST CLIs. Options:
   a) I'll build the Phase 0.5 workflow commands directly using a GraphQL client (recommended)
   b) Pick a different REST API for generation"
3. **If no spec and API is GraphQL-only** -> same as #2
4. **If the spec describes REST endpoints but the API base URL contains `/graphql`** -> STOP.
   "The spec describes REST endpoints but the API is GraphQL. Generating would produce commands that can't make successful API calls."

**NEVER generate a CLI that can't make a single successful API call.**

### Step 2.1: Get the spec ready

**If OpenAPI spec found:**
```bash
curl -sL -o /tmp/printing-press-spec-<api>.json "<spec-url>" && head -c 200 /tmp/printing-press-spec-<api>.json
```

**If no spec (write from docs):**
1. **WebFetch** the API docs
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Write YAML spec to `/tmp/<api>-spec.yaml` with ALL endpoints
4. Include auth config matching research findings

### Step 2.2: Check for existing output, remove if exists

```bash
cd ~/cli-printing-press && rm -rf <api>-cli 2>/dev/null; echo "CLEAN"
```

### Step 2.3: Run the generator

```bash
cd ~/cli-printing-press && ./printing-press generate \
  --spec /tmp/printing-press-spec-<api>.json \
  --output ./<api>-cli \
  --force --lenient --validate 2>&1
```

### Step 2.4: Note skipped complex body fields

**IMPORTANT:** When the generator outputs "warning: skipping body field X: complex type not supported", note EVERY skipped field. You will handle these in Phase 4.

Run:
```bash
cd ~/cli-printing-press && ./printing-press generate --spec /tmp/printing-press-spec-<api>.json --output ./<api>-cli --force --lenient --validate 2>&1 | grep "skipping body field"
```

Save the list of skipped fields. These are NOT acceptable limitations - they are work items for Phase 4.

### Step 2.5: Handle quality gate failures

Max 3 retries. Read errors carefully and fix spec issues.

### PHASE GATE 2

**STOP.** Verify:
1. CLI directory exists
2. `go build ./...` succeeds
3. List of skipped complex body fields is saved for Phase 3

Tell the user: "Phase 2 complete: Generated <api>-cli with [N] resources, [M] endpoints. [K] complex body fields noted for Phase 4. Proceeding to Steinberger audit."

---

# PHASE 3: STEINBERGER AUDIT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

This phase has TWO parts: (A) code review for tactical fixes, and (B) Steinberger analysis for strategic assessment. Both are required.

## Part A: Code Review

### Step 3.1: Read the generated code

You MUST **Read** these files (not just check they exist):

- `<api>-cli/internal/cli/root.go`
- `<api>-cli/README.md`
- At least 3 resource command files

### Step 3.2: Count commands and compare to target

```bash
cd ~/cli-printing-press/<api>-cli && grep -r "Use:" internal/cli/*.go | grep -v "root.go" | wc -l
```

Compare against target from Phase 1 research.

### Step 3.3: Check help text quality

- Descriptions: developer-friendly or raw spec jargon?
- Examples: realistic values or placeholder garbage ("string", "0", "abc123")?
- Root command: does it explain what the API does?

### Step 3.4: Check for missing endpoints

Compare spec endpoints against generated commands. Note any gaps.

### Step 3.5: Check agent-native features

Verify in root.go: --json, --select, --dry-run, --stdin, --yes, --no-cache, doctor.

### Step 3.6: Check complex body fields

For each field skipped by the generator (from Phase 2 Step 2.4):
1. Is this field critical for the endpoint's purpose?
2. Can the user work around it with `--stdin`?
3. What example JSON would a user pipe in?

## Part B: First Steinberger Analysis

### Step 3.0: Run automated scorecard

Before hand-scoring, run the automated scorecard to get objective baseline numbers:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

Use these numbers as the baseline. The hand-scoring in Step 3.7 should explain WHY each dimension got its score, not re-guess the number.

### Step 3.7: Score against the Steinberger bar

Score each dimension 0-10. For EACH dimension, provide THREE things:
1. **Current score** with justification
2. **What 10/10 looks like** (reference gogcli or best-in-class)
3. **What specific changes would raise the score** (actionable items)

```markdown
## First Steinberger Analysis (Baseline)

| Dimension | Score | What 10 Looks Like | How to Get There |
|-----------|-------|-------------------|-----------------|
| Output modes | X/10 | gogcli: --json, --yaml, --csv, --table, --select, --quiet, --template | Add --yaml output, add --template for custom formats |
| Auth | X/10 | gogcli: OAuth browser flow, token storage, multiple profiles, doctor validates | Add OAuth flow, add profile switching |
| Error handling | X/10 | gogcli: typed exits, retry with backoff, helpful suggestions, link to docs | Add "did you mean?" suggestions |
| Terminal UX | X/10 | gogcli: progress spinners, color themes, pager for long output | Add progress spinners for pagination |
| README | X/10 | gogcli: install, quickstart, every command with example, cookbook, FAQ | Add cookbook section, add FAQ |
| Doctor | X/10 | gogcli: validates auth, API version, rate limits, config file health | Add API version check, config health |
| Agent-native | X/10 | gogcli: --json, --select, --dry-run, --stdin, idempotent, typed exits, no TTY | Already strong if all flags present |
| Local Cache | X/10 | gogcli: file cache + optional embedded DB (bolt/badger), --no-cache bypass, cache clear | [what changes would raise score] |
| Breadth | X/10 | gogcli: 100+ commands covering every API endpoint + convenience wrappers | Add missing commands, add convenience wrappers |
| Vision | X/10 | discrawl: SQLite + FTS5 + sync + search + tail + domain workflows | Add export, search, sync commands based on Phase 0 research |

**Baseline Total: X/100 (Grade X)**
```

### Step 3.8: Write the GOAT improvement plan

Based on the Steinberger analysis, identify:

1. **Top 5 highest-impact improvements** (will raise the score the most)
2. **Commands to ADD** (not just rename - new functionality)
3. **Complex body field examples** to add (top 3 endpoints where --stdin matters most)
4. **What's achievable in Phase 4** vs what's future work

### Step 3.9: Write the audit artifact

**Write** to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-audit.md`:

Include ALL of:
- Command comparison
- Help text quality assessment
- Agent-native checklist
- Specific fixes needed (file paths + what to change)
- First Steinberger Analysis table (full)
- GOAT improvement plan (top 5 + commands to add)
- Complex body field plan

### PHASE GATE 3

**STOP.** Verify ALL of these:
1. Audit artifact exists with Steinberger analysis table
2. Each Steinberger dimension has: score, "what 10 looks like", and "how to get there"
3. GOAT plan has at least 5 specific improvements
4. Complex body fields have a plan (not just "limitation")
5. Baseline total score is recorded

**Write Phase 3 Artifact:** Run the Artifact Writing plan generator with all Phase 3 analysis as input. Write to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-audit.md`. Include: scorecard baseline, full 11-dimension hand-scored table, GOAT improvement plan, complex body field plan, data layer integration notes from Phase 0.7.

Tell the user: "Phase 3 complete: Baseline Steinberger Score: [X]/100 (Grade [X]). Found [N] tactical fixes + [M] GOAT improvements. Top improvement: [description]. Proceeding to GOAT build."

---

# PHASE 4: GOAT BUILD

## THIS PHASE IS MANDATORY.

**The generator output is scaffolding, not the product. The data layer + workflows are the product.**

**Read the Phase 0.7 Data Layer Specification and Phase 3 Audit artifacts before starting.**

Execute in this priority order. Do NOT skip Priority 0 to go straight to workflows.

### Priority 0: Data Layer Foundation (from Phase 0.7)

**This is the most important work in the entire pipeline.** Replace the generic store with domain-specific tables.

1. **Replace `internal/store/store.go`** with domain-specific schema from Phase 0.7 spec:
   - CREATE TABLE with proper columns for each Primary entity (not JSON blobs)
   - CREATE INDEX on foreign keys and temporal fields
   - CREATE VIRTUAL TABLE ... USING fts5() on text fields
   - UpsertMessage/UpsertMember/etc methods that extract fields from JSON

2. **Rewrite `sync` command** with domain-aware sync:
   - Use the cursor field identified in Phase 0.7 (e.g., snowflake ID, updatedAt)
   - Add `--guild`/`--team`/`--project` scoping flags (domain-specific)
   - Add `--since` time filter
   - Paginate with the validated cursor params

3. **Add domain-specific search filters:**
   - Map Phase 0.7's filter table into search command flags
   - `--channel`, `--author`, `--team`, etc -> SQL WHERE clauses
   - FTS5 on extracted text content, not raw JSON

4. **Add `sql` command** for raw read-only queries:
   ```go
   func newSqlCmd(flags *rootFlags) *cobra.Command {
       cmd := &cobra.Command{
           Use:   "sql <query>",
           Short: "Run read-only SQL queries against the local database",
           RunE: func(cmd *cobra.Command, args []string) error {
               // Open DB, execute query, print results as table or JSON
           },
       }
   }
   ```

5. **Add entity-specific list commands** (e.g., `messages`, `members`):
   - Query local SQLite, not the API
   - Support `--channel`, `--author`, `--days`, `--hours`, `--since` filters
   - Support `--sync` flag to trigger on-demand sync before querying

6. **Update tail command** based on Phase 0.7's tail strategy:
   - If WebSocket/Gateway: implement real-time connection
   - If SSE: implement EventSource reader
   - If REST polling: keep current implementation but use domain-aware cursors

### Priority 1: Power User Workflows (from Phase 0.5) - NOW powered by local DB

Implement the top 5-7 workflows identified in Phase 0.5 as real, hand-written Go commands. **Where possible, query the local SQLite database instead of making live API calls.** This makes workflows instant and avoids rate limits.

For each workflow:
1. **Create a dedicated command file** (e.g., `internal/cli/stale.go`, `internal/cli/velocity.go`)
2. **Use the generated client** to make real API calls
3. **Combine 2+ API calls** into one user-facing operation
4. **Add realistic examples** in --help that show the actual workflow
5. **Support --json output** for agent consumption
6. **Register in root.go** alongside the generated commands

Example for a "stale issues" workflow on a project management API:
```go
func newStaleCmd(flags *rootFlags) *cobra.Command {
    var days int
    var team string
    cmd := &cobra.Command{
        Use:   "stale",
        Short: "Find issues with no updates in N days",
        Example: `  linear-cli stale --days 30 --team ENG
  linear-cli stale --days 14 --json --select identifier,title,updatedAt`,
        RunE: func(cmd *cobra.Command, args []string) error {
            c, err := flags.newClient()
            // ... fetch issues, filter by updatedAt, group by team
        },
    }
}
```

### Priority 2: Scorecard-Gap Fixes

Run the scorecard and fix dimensions below 10/10:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

For each dimension below 10/10:
1. **Read** the relevant file
2. **Edit** with surgical changes
3. Focus on changes that RAISE THE SCORECARD NUMBER

Also fix:
- Complex body field --stdin examples for top 3 endpoints
- Lazy descriptions (1-2 word Short fields)
- Placeholder examples ("abc123" -> realistic domain values)

### Priority 3: Polish

Only after Priority 1 and 2 are complete:
1. README cookbook section **showcasing workflow commands** (not just API calls)
2. Command name cleanup
3. FAQ section with domain-specific questions

### Step 4.4: Verify compilation

```bash
cd ~/cli-printing-press/<api>-cli && go build ./... && go vet ./... && echo "ALL FIXES VERIFIED"
```

### PHASE GATE 4

**STOP.** Verify:
1. Data layer implemented: domain-specific SQLite tables (NOT generic JSON blobs)
2. Sync command uses domain-aware cursors (validated in Phase 0.7)
3. Search supports domain-specific filters (--channel, --author, --team, etc.)
4. `sql` command exists for raw read-only queries
5. At least 3 workflow commands implemented (from Phase 0.5)
6. Workflow commands use local DB where possible (not just live API calls)
7. Scorecard gaps addressed
8. `go build ./...` and `go vet ./...` pass
9. README cookbook includes data layer + workflow examples

**Write Phase 4 Artifact:** Run the Artifact Writing plan generator with all Phase 4 work as input. Write to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-goat-build-log.md`. Include: data layer implementation details, workflow commands built, scorecard fixes, what was skipped, before/after scorecard comparison.

Tell the user: "Phase 4 complete: Built [N] data layer tables + [M] workflow commands, applied [K] scorecard fixes. Data layer: [list tables]. Top workflow: [name]. Compilation verified. Proceeding to dogfood emulation."

---

# PHASE 4.5: DOGFOOD EMULATION

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

You don't have real API keys. But the OpenAPI spec already defines every request shape, response schema, error format, and pagination pattern. Test every generated command against spec-derived mock responses. Inspired by [Vercel's emulate](https://github.com/vercel-labs/emulate) - production-fidelity API simulation, zero config.

**Read the OpenAPI spec (from Phase 2) and the Phase 4 GOAT Build Log artifact before starting.**

### Step 4.5a: Generate Synthetic Responses from Spec

For each endpoint in the spec, generate a realistic JSON response by reading the 200/201 response schema:

**Field value heuristics** (domain-aware, not random):

| Field name pattern | Generated value |
|---|---|
| `id`, `*_id` | Realistic format for the API (e.g., Discord snowflake: `"1234567890123456789"`, UUID: `"550e8400-e29b-41d4-a716-446655440000"`) |
| `name`, `username`, `title` | Realistic domain values (e.g., `"general"`, `"test-user"`, `"Bug: Login fails"`) |
| `content`, `description`, `body` | `"Dogfood test content for validation"` |
| `timestamp`, `created_at`, `updated_at` | `"2026-03-26T12:00:00.000Z"` |
| `type` (enum in spec) | First enum value from the spec |
| `url`, `avatar_url`, `icon_url` | `"https://example.com/test.png"` |
| `email` | `"test@example.com"` |
| `count`, `position`, `size` | `1` |
| `boolean` fields | `true` |
| Array fields | 2-3 items with the above heuristics |
| Nested objects | Recursively generate from schema |

Save mocks to `/tmp/<api>-cli-mocks/` for reuse.

### Step 4.5b: Score Every Command on 5 Dimensions

For each generated command, score 0-10 on each dimension (50 max):

**Dimension 1: Request Construction (0-10)**

Run the command with `--dry-run` and inspect the output:

```bash
<api>-cli <resource> <action> <required-args> --dry-run 2>&1
```

| Check | Points |
|---|---|
| Path params replaced (no `{param}` literals in URL) | 2 |
| HTTP method matches spec | 2 |
| Required query params present | 2 |
| Body schema matches spec's requestBody | 2 |
| Auth header present | 2 |

**Dimension 2: Response Parsing (0-10)**

Generate a synthetic response from the spec and verify the command can process it:

| Check | Points |
|---|---|
| Can parse the spec's 200 response schema | 3 |
| --json output is valid JSON | 2 |
| --select works with fields from the response schema | 2 |
| Table output renders without crash | 2 |
| Error responses (401/404/429) produce correct exit codes | 1 |

**Dimension 3: Schema Fidelity (0-10)**

Compare generated flags against the spec's parameters:

| Check | Points |
|---|---|
| All `required: true` params have CLI flags | 3 |
| Flag types match spec types (string/int/bool) | 2 |
| No hallucinated flags (every flag maps to a real spec param) | 3 |
| Help descriptions come from spec, not invented | 2 |

**Dimension 4: Example Quality (0-10)**

Validate every example in --help and README:

| Check | Points |
|---|---|
| Example IDs match realistic format (not "abc123") | 2 |
| --stdin JSON matches spec's requestBody schema | 3 |
| Required flags present in examples | 3 |
| Example commands parse without usage error via --dry-run | 2 |

**Dimension 5: Workflow Integrity (0-10)** (workflow commands only)

| Check | Points |
|---|---|
| All API paths hit by the workflow exist in the spec | 3 |
| Query params sent match spec's parameters | 2 |
| Response fields accessed exist in the response schema | 3 |
| Cross-entity joins reference valid fields | 2 |

### Step 4.5c: Compute Aggregate Scores

| Metric | Formula |
|---|---|
| **Per-command score** | Sum of 5 dimensions (0-50) |
| **Pass rate** | % of commands scoring >= 35/50 (70%) |
| **Critical failure count** | Commands scoring < 25/50 |
| **Overall dogfood score** | Average across all tested commands |

**Thresholds:**
- **PASS:** Pass rate >= 90% AND 0 critical failures
- **WARN:** Pass rate >= 70% AND <= 3 critical failures (auto-fix, then re-score)
- **FAIL:** Pass rate < 70% OR > 3 critical failures (report issues, do NOT proceed)

**Sampling for large CLIs (100+ commands):** Test ALL workflow commands + ALL commands with --stdin examples + random sample of 30 generated commands. Report sample size.

### Step 4.5d: Auto-Fix Issues Found

For each fixable issue:

| Issue Type | Auto-Fix Action |
|---|---|
| Placeholder values ("abc123", "string") | Replace with realistic domain values from spec |
| Missing required flags in examples | Add required flags with domain-realistic values |
| --stdin JSON doesn't match requestBody | Regenerate from spec schema |
| Lazy 1-word Short descriptions | Pull description from spec's endpoint summary |
| Hallucinated flag (not in spec) | Remove the flag and its binding |
| Wrong flag type (string instead of int) | Fix the cobra flag type |

After auto-fixes:
1. Run `go build ./...` and `go vet ./...` to verify fixes compile
2. Re-run the dogfood scoring
3. Report before/after scores

### Step 4.5e: Write the Dogfood Report Artifact

**Run the Artifact Writing plan generator** with all dogfood results as input. Write to `~/cli-printing-press/docs/plans/<today>-fix-<api>-cli-dogfood-report.md`.

The artifact MUST include:
- Per-command score table (sampled commands, all 5 dimensions)
- Top 5 failures with root cause analysis
- Hallucination list (flags/fields not in spec)
- Auto-fixes applied with before/after diff
- Recommendations for remaining manual fixes
- Overall dogfood score, pass rate, critical failure count
- PASS/WARN/FAIL verdict

### PHASE GATE 4.5

**STOP.** Verify ALL of these before proceeding:
1. Every workflow command scored on all 5 dimensions
2. Sample of generated commands scored (30+ or all if < 100)
3. Synthetic responses generated from spec (not invented)
4. Per-command score table computed
5. Auto-fixes applied and compilation verified
6. Re-score after fixes shows improvement
7. Dogfood report artifact written
8. Final verdict: PASS or WARN (FAIL = stop and report)

Tell the user: "Phase 4.5 complete: Dogfood score [X]/50 avg across [N] commands. Pass rate: [Y]%. Critical failures: [Z]. Auto-fixed [K] issues (+[D] point improvement). [PASS/WARN/FAIL]. Proceeding to final Steinberger."

---

# PHASE 5: FINAL STEINBERGER + REPORT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

### Step 5.1: Second Steinberger Analysis (Post-Fix)

Run the automated scorecard again to measure improvement:

```bash
cd ~/cli-printing-press && ./printing-press scorecard --dir ./<api>-cli
```

Re-score ALL 10 dimensions. Show the DELTA from the baseline:

```markdown
## Final Steinberger Analysis (Post-Fix)

| Dimension | Before | After | Delta | What Changed |
|-----------|--------|-------|-------|-------------|
| Output modes | X/10 | Y/10 | +Z | [specific change] |
| Auth | X/10 | Y/10 | +Z | [specific change] |
| Error handling | X/10 | Y/10 | +Z | [specific change] |
| Terminal UX | X/10 | Y/10 | +Z | [specific change] |
| README | X/10 | Y/10 | +Z | [specific change] |
| Doctor | X/10 | Y/10 | +Z | [specific change] |
| Agent-native | X/10 | Y/10 | +Z | [specific change] |
| Local Cache | X/10 | Y/10 | +Z | [specific change] |
| Breadth | X/10 | Y/10 | +Z | [specific change] |
| Vision | X/10 | Y/10 | +Z | [specific change] |

**Before: X/100 -> After: Y/100 (+Z points)**
**Grade: [A/B/C]**
```

### Step 5.2: Remaining gaps

For each dimension still < 8/10:
- What would it take to reach 8+?
- Is this a generator limitation or achievable with more fix time?
- Tag as "future work" with a specific next step

### Step 5.3: Present the final report

Show ALL of these sections:

**1. Summary:**
```
Generated <api>-cli with <N> resources and <M> commands.
Resources: <comma-separated list>
```

**2. Steinberger Score (Before/After):**
```
Steinberger Score: Before X/100 -> After Y/100 (+Z points) - Grade [A/B/C]

[Full before/after table from Step 5.1]
```

**3. Competitor Comparison:**
```
Found <N> competing CLIs.
Best competitor: <name> (<stars> stars, <commands> commands)
Strategic advantage: <why ours is better - from Phase 1 research>
We beat them on: <specific features>
Remaining gap: <what they have that we don't, or "none">
```

**4. Example Commands (with complex body examples):**
```bash
cd ~/cli-printing-press/<api>-cli
go install ./cmd/<api>-cli

export <AUTH_ENV_VAR>="..."

# Basic usage
<api>-cli --help
<api>-cli doctor
<api>-cli <resource> list --json
<api>-cli <resource> get <realistic-id>

# Complex body fields (pipe JSON via stdin)
echo '<realistic-json>' | <api>-cli <resource> create --stdin

# Agent workflow
<api>-cli <resource> list --json --select id,name | jq -r '.[].id'
```

**5. Spec source and limitations**

**6. Future work** (from remaining gaps)

---

## Writing Specs from Docs

When no OpenAPI spec exists:

1. **WebFetch** the API docs
2. **Read** `~/cli-printing-press/skills/printing-press/references/spec-format.md`
3. Read the docs and identify EVERY endpoint
4. Write YAML spec to `/tmp/<api>-spec.yaml`
5. Generate from it

You ARE the brain. Read the docs yourself and write the spec.

## Submit to Catalog

`/printing-press submit <name>` - gather metadata, write `catalog/<name>.yaml`, create PR.

## Safety Gates

- Preview before generating
- Output directory conflict: check before overwriting
- Untrusted specs: note if not from known-specs registry
- Max 3 retries on quality gate failure

## Anti-Shortcut Rules

These phrases indicate a phase was shortcut. If you catch yourself writing them, STOP and re-do the phase:

- "This is a limitation of the generator" (fix it, don't accept it)
- "Complex types not supported" (add --stdin examples)
- "We'll skip this for now" (no skipping - do it or explain why it's impossible)
- "The quality is good enough" (score it against Steinberger, prove it's good enough with numbers)
- "Let's wrap up" (are all 5 phases complete with artifacts?)
- "This API doesn't need local persistence" (Did you run Phase 0? Check the data profile. If search need is high, it needs persistence.)
- "This is just an API wrapper" (Run Phase 0 again. What would a thoughtful developer build?)
- "The API is GraphQL-only but I'll write a REST spec anyway" (STOP. This produces garbage. Use Phase 0.5 workflows or build a GraphQL client.)
- "I'll polish the README instead of building workflows" (Phase 4 Priority 1 is workflows. README is Priority 3. Do not skip ahead.)
- "The Phase 0.5 workflows are future work" (They are the product. Build them now or the CLI is just an API wrapper.)
- "316 commands is better than 12" (discrawl has 12 commands and 539 stars. Depth beats breadth. Build the workflows.)
- "The API doesn't need local persistence" (Check data gravity scores from Phase 0.7. If any entity scores >= 8, it needs SQLite with proper columns.)
- "FTS5 is overkill for this API" (If any entity has 2+ text fields AND data gravity >= 8, it needs FTS5. That's how search works.)
- "REST polling is fine for tail" (Check if the API has WebSocket/SSE/Gateway. If yes, use it. REST polling misses events and wastes rate limit budget.)
- "The generic store is good enough" (Domain-native tables ALWAYS beat JSON blob tables. A `messages` table with `channel_id`, `author_id`, `content` columns enables joins and filters that JSON blobs can't. Write the schema.)
- "I'll build the data layer later" (Phase 0.7 runs BEFORE generation. The data layer spec informs Phase 4 Priority 0. Build it first, then workflows use it.)
- "The artifact is just documentation" (The 7 plan artifacts ARE the product. They capture reasoning, evidence, and decisions. The generated CLI is a side effect.)
- "The CLI compiles so it works" (Compilation proves syntax, not semantics. A command that builds can still 400 on every real call. Run the dogfood.)
- "We can't test without API keys" (The OpenAPI spec defines response schemas. Generate mocks from the spec. Test against them. Zero keys needed.)
- "The dry-run looks right" (Dry-run validates request construction. You also need to feed synthetic responses to validate output parsing, --select, and table rendering.)
