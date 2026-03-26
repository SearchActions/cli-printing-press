---
title: "Power User Prediction Engine for Printing Press"
type: feat
status: completed
date: 2026-03-26
---

# Power User Prediction Engine for Printing Press

## Overview

The printing-press generates Grade A API wrappers (Discord: 316 commands, 94/110 Steinberger). But it can't predict what a power user would build *on top* of those APIs - the domain-specific SQLite schema, incremental sync, full-text search with domain filters, real-time tail, structured mention tracking, and raw SQL access that make discrawl (12 commands, 540 stars) more valuable than a 316-command wrapper.

This plan adds **Phase 0.7: Power User Prediction Engine** to the SKILL.md. It runs after Phase 0.5 (Power User Workflows) and before Phase 1 (Deep Research). It uses API surface analysis + social signal mining + data gravity scoring to generate a **Data Layer Specification** - a concrete blueprint for building discrawl-class features without knowing discrawl exists.

## Problem Statement

Phase 0.5 thinks in *commands* ("what compound CLI commands would a power user want?"). But power users don't just want commands - they want a **local data layer** that enables all of them. discrawl's killer feature isn't any single command - it's the SQLite database with domain-native tables.

**Evidence from the Discord run (2026-03-25):**

| What we predicted (Phase 0.5) | What we missed (discrawl has) |
|---|---|
| channel-health workflow | Discord-native schema (messages/members/channels tables) |
| audit-report workflow | Gateway WebSocket tail (not REST polling) |
| member-report workflow | `messages` command with --channel/--author/--days filters |
| server-snapshot workflow | `mentions` command with structured mention tracking |
| message-stats workflow | `members search` across bio, pronouns, social handles |
| webhook-test workflow | `sql` command for raw read-only queries |
| prune-preview workflow | FTS5 on extracted content (not JSON blobs) |

The workflows we predicted are good - discrawl doesn't have them. But the data layer underneath is where discrawl wins 13/13 on capabilities.

**Root cause:** The generator creates a generic `resources` table with JSON blobs and a generic FTS5 index on those blobs. It should create domain-specific tables with proper columns and FTS5 on extracted text fields.

## Proposed Solution

### New Phase: 0.7 - Power User Prediction Engine

Insert between Phase 0.5 and Phase 1 in the SKILL.md. Five prediction steps, 15-25 minutes total.

```
Phase 0   -> Phase 0.5 -> [Phase 0.7] -> Phase 1 -> Phase 2 -> Phase 3 -> Phase 4 -> Phase 5
Visionary    Workflows    Prediction     Research    Generate    Audit      Build      Score
Research     (commands)   Engine         (specs)     (code)      (review)   (fixes)    (final)
                          (data layer)
```

---

## Technical Approach

### Step 0.7a: Entity Classification

**Goal:** Map every API resource into one of four types.

Read the OpenAPI spec (or use Phase 0's entity list) and classify each resource:

| Type | Signal | Example | Persistence Need |
|---|---|---|---|
| **Accumulating** | Grows over time, has timestamps, paginated lists | Messages, Issues, Audit Logs, Commits | SQLite table + incremental sync |
| **Reference** | Changes rarely, small cardinality, referenced by other entities | Users, Teams, Roles, Labels, Channels | SQLite table + periodic refresh |
| **Append-only** | Never edited, only created | Events, Webhooks, Notifications | SQLite table + tail command |
| **Ephemeral** | Short-lived, not worth persisting | OAuth tokens, Rate limit status, Gateway info | API-only, no persistence |

**Output:** Entity classification table with type, estimated volume, update frequency, and key temporal field.

**Heuristics for classification:**
- Has `created_at`/`timestamp` + paginated list endpoint -> Accumulating
- Referenced by 3+ other entities via `_id` fields -> Reference
- Has no UPDATE/PATCH endpoint -> Append-only
- No list endpoint or < 100 expected records -> Ephemeral
- Has `updated_at` or `modified_at` -> needs incremental sync cursor

### Step 0.7b: Social Signal Mining for Data Patterns

**Goal:** Find evidence of what data power users actually store locally.

Run 7 parallel WebSearches:

```
1. "<API name>" export OR backup OR archive site:github.com
2. "<API name>" SQLite OR database OR local site:github.com
3. "<API name>" analytics OR dashboard OR metrics site:github.com
4. "<API name>" "I wish" OR "would be nice" OR "feature request" data
5. "<API name>" offline OR search OR "full text" site:reddit.com OR site:news.ycombinator.com
6. "<API name>" trend OR pattern OR anomaly detection
7. "<API name>" graph OR visualization OR dependency
```

**For each finding, extract:**
- What entities they store locally
- What queries they run (joins, aggregations, time filters)
- What temporal patterns they track
- What cross-entity relationships they need

**Score using Phase 0 evidence framework.** Anything with score >= 6 informs the data layer.

### Step 0.7c: Data Gravity Scoring

**Goal:** Rank entities by how much value they'd have in a local SQLite database.

**Formula:** `DataGravity = Volume(0-3) + QueryFrequency(0-3) + JoinDemand(0-2) + SearchNeed(0-2) + TemporalValue(0-2)`

| Factor | 0 | 1 | 2 | 3 |
|---|---|---|---|---|
| **Volume** | < 100 records | 100-10k | 10k-1M | > 1M |
| **QueryFrequency** | Rarely queried | Monthly | Weekly | Daily |
| **JoinDemand** | No references | 1-2 entities reference it | 3-4 | 5+ |
| **SearchNeed** | No text fields | 1 text field | 2-3 text fields | Primary text content |
| **TemporalValue** | No time dimension | Created date only | Updated + trends | Core to time-series analysis |

**Thresholds:**
- Score >= 8: **Primary entity** - gets its own SQLite table with proper columns
- Score 5-7: **Support entity** - gets a table but simpler schema
- Score < 5: **API-only** - no local persistence

**Example scores that would have predicted discrawl:**

| Entity | Volume | QueryFreq | JoinDemand | SearchNeed | TemporalValue | Total | Decision |
|---|---|---|---|---|---|---|---|
| Messages | 3 | 3 | 2 | 3 | 2 | **13** | Primary |
| Members | 2 | 2 | 2 | 2 | 1 | **9** | Primary |
| Channels | 1 | 2 | 2 | 1 | 0 | **6** | Support |
| Roles | 1 | 1 | 2 | 0 | 0 | **4** | API-only |
| Audit Logs | 2 | 2 | 1 | 1 | 2 | **8** | Primary |

This scoring would have correctly predicted that messages, members, and audit logs need dedicated SQLite tables - exactly what discrawl builds.

### Step 0.7d: Schema + Sync + Search Specification

**Goal:** Write the concrete data layer specification.

For each Primary entity (score >= 8):

**1. SQLite Schema:**
- Extract columns from the API's response schema (not just id + JSON blob)
- Include foreign key columns for joins (e.g., `channel_id`, `author_id`)
- Include the temporal field for sync cursors
- Add indexes on foreign keys and temporal fields
- Create FTS5 virtual table on text fields

**2. Sync Strategy:**
- Identify the incremental sync cursor (timestamp field, snowflake ID, cursor pagination)
- Verify the API supports filtering by this cursor (check query params for `since`, `after`, `updated_after`)
- Determine batch size from API's max `limit` parameter
- Check if API has WebSocket/SSE/Gateway - if yes, note for tail command
- **VALIDATE:** If the API doesn't support filtering by the cursor field, fall back to full sync + local dedup

**3. Search Specification:**
- List which text fields to extract into FTS5 (title, description, content, body, name)
- Define domain-specific search filters as SQL WHERE clauses
- Map CLI flags to SQL: `--channel` -> `WHERE channel_id = ?`, `--author` -> `WHERE author_id = ?`
- **VALIDATE:** Check that the API actually returns these fields in list/get responses

**4. Compound Queries:**
- Define 3-5 useful cross-entity queries (e.g., "messages by author in channel in last N days")
- Validate that the join columns exist in both tables
- These become the basis for workflow commands that use the local database instead of live API calls

### Step 0.7e: Write the Data Layer Specification Artifact

Write to `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-data-layer-spec.md`:

```markdown
## Data Layer Specification: <API> CLI

### Entity Classification
[Table from Step 0.7a]

### Data Gravity Scores
[Table from Step 0.7c]

### SQLite Schema
[CREATE TABLE statements for each Primary/Support entity]
[CREATE INDEX statements]
[CREATE VIRTUAL TABLE ... USING fts5(...) for text-heavy entities]

### Sync Strategy
| Entity | Cursor Field | API Filter Param | Batch Size | Frequency |
[Table]

### Domain-Specific Search Filters
| CLI Flag | SQL Mapping | Entities |
[Table]

### Compound Queries (for Phase 4 Workflows)
| Query Name | SQL | Use Case |
[Table]

### Commands to Build in Phase 4
| Command | Description | Uses Local DB? | API Calls |
| sync | Incremental sync to SQLite | Yes (writes) | GET /entities?since=cursor |
| search | FTS5 search with domain filters | Yes (reads) | None (local only) |
| sql | Raw read-only SQL access | Yes (reads) | None (local only) |
| messages/issues/etc | Entity-specific listing with filters | Yes (reads) | None (local only) |
[Table]

### Tail Strategy
| Method | When to Use | Implementation |
| WebSocket/Gateway | API has it (Discord, Slack) | Connect + event handler |
| SSE | API has it (Linear, GitHub) | EventSource reader |
| REST Polling | Fallback | GET with ?since= cursor |
[Decision for this specific API]
```

### Phase Gate 0.7

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

## Artifact Pipeline: 6 Plans Per Run (via /ce:plan or Built-in)

Every printing-press run should ship **6 comprehensive .md plans** in `docs/plans/`. Each phase pauses at its gate, **runs /ce:plan** (or its built-in equivalent) to write the artifact, then resumes. The artifacts ARE the product - the generated CLI is a side effect.

### How artifact writing works:

At the end of each phase, the skill runs a **plan generation step**. It detects whether the user has compound-engineering installed and routes accordingly:

**Detection logic (add to SKILL.md at the top, before Phase 0):**

```markdown
### Artifact Writing: Plan Generation

At the end of each phase, write a comprehensive plan document. Use this priority:

**Option A: /ce:plan is available (compound-engineering installed)**

Check if the `compound-engineering:ce:plan` skill exists. If yes, invoke it:

    Skill tool: compound-engineering:ce:plan
    Args: "<phase description with all research gathered so far>"

This produces a full plan document with frontmatter, acceptance criteria, technical analysis, and sources. The skill handles the depth - just pass it all the research from this phase as the feature description.

**Option B: Built-in plan writer (fallback when compound-engineering is NOT installed)**

If /ce:plan is not available, write the artifact yourself using this structure:

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

CRITICAL: The built-in writer must match /ce:plan depth. This means:
- Full analysis, not bullet summaries
- Evidence with source URLs, not assertions
- Scoring breakdowns, not just final numbers
- Concrete SQL/code examples, not pseudocode
- Validation proof ("I verified the API supports ?after= filtering"), not assumptions
- 200+ lines minimum per artifact
```

### The 6 artifacts and when to generate them:

```
Phase 0   -> [run plan writer] -> <today>-feat-<api>-cli-visionary-research.md
Phase 0.5 -> [run plan writer] -> <today>-feat-<api>-cli-power-user-workflows.md
Phase 0.7 -> [run plan writer] -> <today>-feat-<api>-cli-data-layer-spec.md
Phase 1   -> [run plan writer] -> <today>-feat-<api>-cli-research.md
Phase 3   -> [run plan writer] -> <today>-fix-<api>-cli-audit.md
Phase 4   -> [run plan writer] -> <today>-fix-<api>-cli-goat-build-log.md
```

### What each plan must contain:

**1. Visionary Research** (Phase 0):

Pass to /ce:plan or built-in writer:
```
"Write a visionary research plan for the <API> CLI. Include:
- API identity with data profile (write pattern, volume, real-time, search need)
- Top 5 usage patterns with evidence scores and source URLs
- Complete tool landscape with star counts and tool type classification
- Architecture decisions with rationale
- Top 5 features scored on 8 dimensions (evidence, impact, feasibility, uniqueness,
  composability, data profile fit, maintainability, competitive moat)
Use the research gathered in Phase 0 steps 0a-0g as input."
```

**2. Power User Workflows** (Phase 0.5):

Pass to /ce:plan or built-in writer:
```
"Write a power user workflow plan for the <API> CLI. Include:
- API archetype classification with justification
- All 10-15 brainstormed workflow ideas with descriptions
- Validation results per workflow (which API endpoints support each one)
- Full scoring table (Frequency x Pain x Feasibility x Uniqueness)
- Top 7 selected with implementation notes
- Which workflows need local data vs live API calls
Use the research from Phase 0 artifact + Phase 0.5 analysis as input."
```

**3. Data Layer Specification** (Phase 0.7):

Pass to /ce:plan or built-in writer:
```
"Write a data layer specification for the <API> CLI. Include:
- Entity classification table (Accumulating/Reference/Append-only/Ephemeral)
- Social signal mining results with evidence scores
- Data gravity scores with full breakdown (Volume + QueryFreq + JoinDemand + SearchNeed + TemporalValue)
- Complete SQLite schema (CREATE TABLE + CREATE INDEX + CREATE VIRTUAL TABLE FTS5)
- Sync strategy with cursor field validation (verified against API filter params)
- Domain-specific search filters mapped to SQL WHERE clauses
- 3-5 compound cross-entity queries
- Tail strategy decision (WebSocket vs SSE vs REST polling with justification)
Use the entity analysis from Phase 0.7 steps 0.7a-0.7d as input."
```

**4. Research** (Phase 1):

Pass to /ce:plan or built-in writer:
```
"Write a deep research plan for the <API> CLI. Include:
- Spec discovery (URL, format, version, endpoint count)
- Deep competitor analysis for top 2 competitors (README, stars, maintenance status,
  user complaints with quotes, open issues)
- At least 2 user quotes with source URLs
- Strategic justification answering 'why should this CLI exist when [competitor] has [N] stars?'
- Target command count and key differentiators
Use the competitor research from Phase 1 steps 1.1-1.5 as input."
```

**5. Steinberger Audit** (Phase 3):

Pass to /ce:plan or built-in writer:
```
"Write a Steinberger audit plan for the <API> CLI. Include:
- Automated scorecard baseline numbers (from running `printing-press scorecard`)
- Hand-scored 11-dimension table with: current score, 'what 10/10 looks like', 'how to get there'
- GOAT improvement plan with top 5 highest-impact fixes
- Complex body field plan (top 3 endpoints needing --stdin examples)
- Data layer integration notes (from Phase 0.7 spec - what schema changes to make)
- Commands to ADD (new functionality, not just fixes)
Use the code review from Phase 3 steps 3.1-3.8 as input."
```

**6. GOAT Build Log** (Phase 4):

Pass to /ce:plan or built-in writer:
```
"Write a build log for the <API> CLI Phase 4. Include:
- Data layer: what schema was implemented, which tables, which FTS5 indexes
- Workflow commands: list of new commands with descriptions and which use local DB
- Scorecard fixes: what was changed and which dimensions improved
- What was skipped: features deferred to future work with reasons
- Compilation verification: go build + go vet results
- Before/after scorecard comparison
Use the actual code changes made in Phase 4 as input."
```

### The chain pattern:

Each artifact feeds the next. The LLM **reads** the previous artifact before writing the next one:
```
Phase 0 artifact  -> read before Phase 0.5 (what workflows match this API?)
Phase 0.5 artifact -> read before Phase 0.7 (what data layer supports these workflows?)
Phase 0.7 artifact -> read before Phase 2   (what schema to generate?)
Phase 0.7 artifact -> read before Phase 4   (what to build in Priority 0: Data Layer)
Phase 3 artifact   -> read before Phase 4   (what scorecard gaps to fix)
```

**IMPORTANT:** Before each phase, the SKILL.md should instruct: "Read the previous phase's artifact file before proceeding. The artifact contains decisions and analysis you need."

---

## Implementation Plan

### Session 1: Update SKILL.md (1-2 hours)

**File:** `~/.claude/skills/printing-press/SKILL.md`

**Key principle:** Each phase artifact is produced by actually running `/ce:plan` (if compound-engineering is installed) or a built-in equivalent. The SKILL.md instructs the LLM to invoke the Skill tool at each phase gate.

**Changes:**

0. **Add Artifact Writing section** at the top of SKILL.md (before Phase 0):
   - Detection logic: check if `compound-engineering:ce:plan` skill is available
   - Option A: invoke `/ce:plan` with phase research as feature description
   - Option B: built-in plan writer with 200+ line minimum, same depth as /ce:plan
   - Both options produce the same artifact structure (frontmatter, analysis, decisions, outputs, acceptance criteria, sources)

1. **Update phase flow diagram** at the top:
```
PHASE 0 -> PHASE 0.5 -> PHASE 0.7 -> PHASE 1 -> PHASE 2 -> PHASE 3 -> PHASE 4 -> PHASE 5
(3-5 min)   (2-3 min)    (15-25 min)  (5-8 min)  (1-2 min)  (5-8 min)  (5-10 min) (2-3 min)
```

2. **Insert Phase 0.7 section** between Phase 0.5 and Phase 1 with:
   - Steps 0.7a through 0.7e (full text from this plan)
   - Phase Gate 0.7 checklist
   - Data Layer Specification template

3. **Update Phase 4 priority order:**

Current:
```
Priority 1: Power User Workflows (from Phase 0.5)
Priority 2: Scorecard-Gap Fixes
Priority 3: Polish
```

New:
```
Priority 0: Data Layer Foundation (from Phase 0.7)
  - Replace generic store.go with domain-specific schema
  - Implement domain-aware sync with proper cursors
  - Add domain-specific search filters
  - Add `sql` command for raw read-only queries
  - Add entity-specific list commands (e.g., `messages --channel --author --days`)
Priority 1: Power User Workflows (from Phase 0.5) - NOW powered by local DB
Priority 2: Scorecard-Gap Fixes
Priority 3: Polish
```

4. **Add to Anti-Shortcut Rules:**
```
- "The API doesn't need local persistence" (Check data gravity scores. If any entity >= 8, it needs SQLite.)
- "FTS5 is overkill for this API" (If entity has 2+ text fields and score >= 8, it needs FTS5.)
- "REST polling is fine for tail" (Check if API has WebSocket/SSE/Gateway. If yes, use it.)
- "The generic store is good enough" (Domain-native tables always beat JSON blobs. Write the schema.)
- "I'll build the data layer later" (Phase 0.7 runs BEFORE generation. The spec informs the schema.)
```

5. **Update Phase 2 to use Phase 0.7 output:**

After generation, the data layer spec should be used to:
- Override the generic store.go.tmpl output with domain-specific tables
- Or: Phase 4 Priority 0 replaces the generated store.go

### Session 2: Test on Discord (1-2 hours)

Run `/printing-press Discord` with the updated SKILL.md and verify:

**Expected Phase 0.7 output for Discord:**

Entity Classification:
- Messages: Accumulating, Very High volume, Continuous updates, `timestamp` (snowflake)
- Members: Accumulating, High volume, Weekly updates, `joined_at`
- Channels: Reference, Low volume, Monthly updates, N/A
- Audit Logs: Append-only, High volume, Continuous, `id`
- Roles: Reference, Low volume, Monthly, N/A
- Emojis: Reference, Low volume, Monthly, N/A

Data Gravity Scores:
- Messages: 13/12 -> Primary
- Members: 9/12 -> Primary
- Audit Logs: 8/12 -> Primary
- Channels: 6/12 -> Support

Schema should include:
```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    guild_id TEXT,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    timestamp TEXT NOT NULL,
    edited_timestamp TEXT,
    type INTEGER DEFAULT 0,
    data JSON NOT NULL,
    synced_at TEXT NOT NULL
);
CREATE INDEX idx_messages_channel ON messages(channel_id);
CREATE INDEX idx_messages_author ON messages(author_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE VIRTUAL TABLE messages_fts USING fts5(content, tokenize='porter unicode61');
```

Sync should use snowflake ID pagination (`?before=OLDEST_ID`).

Search should support `--guild`, `--channel`, `--author` filters.

**Compare against discrawl features (12 commands):**
- [x] init -> guild discovery during sync
- [x] sync -> incremental sync with snowflake cursors
- [x] tail -> Gateway WebSocket (predicted from API surface)
- [x] search -> FTS5 with domain filters
- [x] messages -> channel/author/time filtering
- [x] mentions -> extracted from message content during sync
- [x] members -> list/show/search across profiles
- [x] channels -> list/show from local DB
- [x] sql -> raw read-only SQLite access
- [x] status -> archive statistics
- [x] doctor -> already generated
- [ ] embeddings -> optional, not predicted (acceptable miss)

**Success criteria:** Engine predicts 10/12 discrawl features without seeing discrawl.

### Session 3: Test on a Third API (1-2 hours)

Run on Stripe to verify generalization:

Expected predictions:
- Charges/PaymentIntents: Primary (high volume, temporal, searchable)
- Customers: Primary (reference but high join demand)
- Subscriptions: Primary (lifecycle tracking)
- Events: Append-only (tail command)
- Schema: charges table with amount, currency, customer_id, status, created columns
- Sync: cursor-based pagination via `starting_after`
- Search: FTS5 on charge description, customer name/email

---

## Acceptance Criteria

### SKILL.md Changes
- [ ] Phase 0.7 section inserted between Phase 0.5 and Phase 1
- [ ] Phase flow diagram updated with Phase 0.7 (15-25 min)
- [ ] Steps 0.7a-0.7e documented with clear instructions
- [ ] Phase Gate 0.7 with 8 verification items
- [ ] Entity classification heuristics (Accumulating/Reference/Append-only/Ephemeral)
- [ ] Data gravity scoring formula with weights and thresholds
- [ ] Data Layer Specification template with all required sections
- [ ] Phase 4 priorities reordered (data layer first)
- [ ] 5 new anti-shortcut rules added
- [ ] Total time estimate updated to 35-65 minutes

### Validation
- [ ] Discord re-run predicts messages/members/audit_logs as Primary entities
- [ ] Discord re-run generates domain-specific SQLite schema (not JSON blobs)
- [ ] Discord re-run predicts Gateway WebSocket for tail
- [ ] Discord re-run generates --guild/--channel/--author search filters
- [ ] 10/12 discrawl features predicted without knowing discrawl exists
- [ ] Third API (Stripe) produces sensible predictions (not hallucinated)

### Quality Gates
- [ ] Every predicted feature validates against actual API capabilities (grounding step in 0.7d)
- [ ] Sync cursors verified to be filterable via API query params
- [ ] FTS5 fields verified to contain meaningful text in API responses
- [ ] No hallucinated columns - every column maps to an actual API response field

---

## Risk Analysis

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Over-prediction for simple APIs | Medium | Low | Data gravity threshold (>= 8) prevents unnecessary SQLite for small APIs |
| Hallucinated schema columns | Medium | High | Step 0.7d validates every column against API response schema |
| 15-25 min adds too much to pipeline | Low | Low | Only runs web searches + analysis, no code generation |
| Sync cursor doesn't work | Medium | Medium | Fallback to full sync + local dedup if API lacks filter params |
| GraphQL APIs lack OpenAPI spec | Medium | Medium | Phase 0.7a works from any schema source (OpenAPI, GraphQL introspection, docs) |

---

## What This Enables

**Before:**
```
OpenAPI Spec -> Generator -> 316 commands + generic JSON blob store
             -> Phase 4  -> 7 workflow commands (API calls only)
             = Good API client, weak data tool
```

**After:**
```
OpenAPI Spec -> Generator -> 316 commands
             -> Phase 0.7 -> Data Layer Specification
             -> Phase 4   -> Domain SQLite schema + sync + search + sql
                          -> 7 workflow commands (powered by local DB)
             = Strong API client + strong data tool
             = discrawl-class without knowing discrawl exists
```

The prediction engine doesn't replace Phase 0.5's workflow ideation. It ADDS a data layer underneath. `discord-cli workflow channel-health` goes from "make N API calls per channel" to "query local SQLite with one JOIN, instant results."

---

## Sources

- Discord CLI generation run (2026-03-25): `docs/plans/2026-03-25-feat-discord-cli-visionary-research.md`
- discrawl gap analysis (2026-03-26): `docs/plans/2026-03-26-feat-discord-cli-vs-discrawl-gap-analysis-plan.md`
- [steipete/discrawl](https://github.com/steipete/discrawl) - 540 stars, the reference data-tool CLI
- Printing-press SKILL.md: `~/.claude/skills/printing-press/SKILL.md`
- Generator templates: `~/cli-printing-press/internal/generator/templates/`
- Profiler: `~/cli-printing-press/internal/profiler/profiler.go`
- Vision templates: `~/cli-printing-press/internal/generator/vision_templates.go`
- Scorecard: `~/cli-printing-press/internal/pipeline/scorecard.go`
