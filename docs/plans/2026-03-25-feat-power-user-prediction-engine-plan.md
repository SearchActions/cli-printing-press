---
title: "Power User Prediction Engine for Printing Press"
type: feat
status: active
date: 2026-03-25
---

# Power User Prediction Engine for Printing Press

## Overview

The printing-press generates excellent API wrappers (Discord: 316 commands, Grade A). What it can't do is predict what a *power user* would build on top of those APIs - the domain-specific SQLite schema, the incremental sync, the full-text search, the real-time tail, the compound analytics. These are the features that make discrawl (12 commands, 540 stars) more valuable than a 316-command wrapper.

This plan adds a new **Phase 0.7: Power User Prediction Engine** that runs between Phase 0.5 (Workflow Ideas) and Phase 1 (Deep Research). It uses the LLM's ability to read social signals, analyze API surfaces, and predict human behavior to automatically generate a "data layer specification" - the blueprint for building discrawl-class features without knowing discrawl exists.

## Problem Statement

**The gap today:**

| What printing-press does well | What it misses |
|------|------|
| API wrapper generation (316 commands from OpenAPI spec) | Domain-specific SQLite schema (messages/members/channels tables vs generic JSON blobs) |
| Scorecard-driven quality (Steinberger 10 dimensions) | Incremental sync with domain-aware cursors (updatedAt, message snowflake IDs) |
| Workflow command ideation (Phase 0.5) | Full-text search with domain-specific filters (--author, --channel, --team) |
| Competitor analysis (Phase 1) | Real-time tail via WebSocket/SSE vs REST polling |
| README cookbook generation | Raw SQL access for power user ad-hoc queries |
| Agent-native features (--json, --select, --dry-run) | Trend detection, dependency graphs, semantic search |

**Evidence from two runs:**

- **Discord CLI:** Phase 0.5 predicted channel-health, audit-report, member-report workflows (good). But missed: Gateway WebSocket tail, domain-native SQLite schema, mentions tracking, `sql` command, semantic search. All of these are predictable from the API surface + Reddit/GitHub signals.
- **Linear CLI:** Phase 0.5 predicted stale/velocity/standup/triage/workload/release-notes/my-day (good). But missed: local SQLite with issues/cycles/states tables, incremental sync via updatedAt cursors, FTS5 across issue descriptions, `sql` command, SLA breach trend detection, dependency graph visualization. All predictable.

**The root cause:** Phase 0.5 thinks in *commands*. It asks "what compound CLI commands would a power user want?" But power users don't just want commands - they want a **local data layer** that lets them query, search, trend, and compose in ways the API doesn't natively support. discrawl's killer feature isn't any single command - it's the SQLite database that enables all of them.

## Proposed Solution

### New Phase: 0.7 - Power User Prediction Engine

Insert between Phase 0.5 (Power User Workflows) and Phase 1 (Deep Research). This phase takes the API identity, data profile, and workflow ideas from Phase 0/0.5 and produces a **Data Layer Specification** - a concrete blueprint for what local persistence, sync, search, and analytics the CLI should have.

**The engine runs 5 prediction steps:**

```
Step 1: API Surface Decomposition (analyze what data is available)
Step 2: Social Signal Mining (how do real users combine this data)
Step 3: Data Gravity Analysis (what data accumulates, what's queried most)
Step 4: Compound Feature Synthesis (chain API calls into data-layer features)
Step 5: Write Data Layer Specification (concrete schema + commands)
```

**Total time: 15-25 minutes** (fits within the overall 20-40 minute pipeline)

## Technical Approach

### Step 0.7a: API Surface Decomposition

**Goal:** Map the API's data model into entities, relationships, and temporal patterns.

**Input:** Phase 0 API identity + schema (OpenAPI spec or GraphQL introspection)

**Process:**

1. Extract all entities (resources/types) with their fields
2. Map relationships (foreign keys, nested objects, connection fields)
3. Identify temporal fields (createdAt, updatedAt, timestamp, since/until params)
4. Identify accumulating data (messages, events, audit logs - things that grow over time)
5. Identify reference data (teams, users, roles, labels - things that change rarely)
6. Identify computed data (progress, velocity, health - derived from other fields)

**Output:** Entity classification table:

```markdown
| Entity | Type | Volume | Update Frequency | Key Temporal Field |
|--------|------|--------|------------------|--------------------|
| Issues | Accumulating | High | Daily | updatedAt |
| Messages | Accumulating | Very High | Continuous | timestamp (snowflake) |
| Teams | Reference | Low | Monthly | updatedAt |
| Users | Reference | Low | Weekly | updatedAt |
| Cycles | Accumulating | Low | Weekly | startsAt/endsAt |
| Audit Logs | Append-only | High | Continuous | createdAt |
```

**Prediction heuristics:**

- Accumulating entities with temporal fields -> needs incremental sync
- High-volume accumulating entities -> needs local persistence (SQLite)
- Reference entities -> needs cache (sync once, refresh periodically)
- Entities with text fields (title, description, body, content) -> needs FTS5
- Entities with relationships to each other -> needs join queries
- Append-only entities -> needs tail/follow command

### Step 0.7b: Social Signal Mining

**Goal:** Discover what data combinations power users actually want, using evidence from the internet.

**Searches (run in parallel via Agent tool):**

```
1. WebSearch: "<API name>" export OR backup OR archive site:github.com
2. WebSearch: "<API name>" SQLite OR database OR local site:github.com
3. WebSearch: "<API name>" analytics OR dashboard OR metrics site:github.com
4. WebSearch: "<API name>" "I wish" OR "would be nice" OR "feature request" data
5. WebSearch: "<API name>" offline OR search OR "full text" site:reddit.com OR site:news.ycombinator.com
6. WebSearch: "<API name>" trend OR pattern OR anomaly detection
7. WebSearch: "<API name>" graph OR visualization OR dependency
```

**For each finding, extract:**
- What data they want locally
- What queries they run against that data
- What temporal patterns they track (trends, anomalies, velocity)
- What joins across entities they need (issues+assignees, messages+channels+authors)

**Score each signal** using the Phase 0 evidence framework (stars, upvotes, cross-platform appearance).

### Step 0.7c: Data Gravity Analysis

**Goal:** Predict which entities will have the most value when stored locally with search + joins.

**Data Gravity = Volume x Query Frequency x Join Demand x Search Need**

For each entity from Step 0.7a, compute:

| Factor | Weight | How to estimate |
|--------|--------|-----------------|
| **Volume** | 3 | From API rate limits, pagination, data profile |
| **Query Frequency** | 3 | From Phase 0.5 workflow analysis (how many workflows touch this entity?) |
| **Join Demand** | 2 | From relationship count (how many other entities reference this?) |
| **Search Need** | 2 | From text field count and social signals (do people search this?) |
| **Temporal Value** | 2 | From temporal field presence (is trend analysis valuable?) |

**Score >= 8: Must be in local SQLite** (primary entity)
**Score 5-7: Should be in SQLite** (reference/support entity)
**Score < 5: API-only** (no local persistence needed)

**Example for Linear:**
- Issues: 3+3+2+2+2 = **12** (primary - high volume, every workflow touches them, many joins, searchable text, temporal trends)
- Cycles: 2+2+1+0+2 = **7** (support - medium volume, velocity workflows, few joins, no text search, temporal)
- Users: 1+2+2+0+0 = **5** (reference - low volume, assignee joins, no search/temporal)
- WorkflowStates: 1+1+1+0+0 = **3** (API-only - tiny, rarely queried independently)

### Step 0.7d: Compound Feature Synthesis

**Goal:** Chain the data layer into concrete features that compose multiple entities.

**For each primary entity (score >= 8), ask:**

1. **Sync pattern:** What's the incremental sync cursor? (updatedAt? cursor pagination? snowflake IDs?)
2. **Search pattern:** What fields should be FTS5-indexed? What domain-specific filters matter?
3. **Trend pattern:** What temporal aggregations would reveal insights? (weekly velocity, daily message volume, SLA breach rate)
4. **Join pattern:** What cross-entity queries would be valuable? (issues-by-assignee-by-state, messages-by-channel-by-author)
5. **Alert pattern:** What threshold crossings should be detectable? (stale issues, SLA breaches, unusual activity spikes)

**For each pattern, validate against the actual API:**
- Can the sync cursor be implemented with available pagination/filter params?
- Do the FTS5 fields actually contain meaningful text?
- Are the temporal fields actually populated and queryable?
- Can the join relationships be resolved from the API?

**Drop any pattern the API can't support.** This is the grounding step - predictions must be API-feasible.

### Step 0.7e: Write Data Layer Specification

**Goal:** Produce a concrete specification that Phase 4 can implement.

**Output artifact:** `~/cli-printing-press/docs/plans/<today>-feat-<api>-cli-data-layer-spec.md`

**Required sections:**

```markdown
## Data Layer Specification: <API> CLI

### SQLite Schema

CREATE TABLE issues (
  id TEXT PRIMARY KEY,
  identifier TEXT NOT NULL,      -- e.g., ENG-123
  title TEXT NOT NULL,
  description TEXT,
  priority INTEGER,
  state_type TEXT,               -- completed, started, etc.
  assignee_name TEXT,
  team_key TEXT,
  cycle_number INTEGER,
  updated_at TEXT NOT NULL,
  synced_at TEXT NOT NULL
);

CREATE VIRTUAL TABLE issues_fts USING fts5(
  identifier, title, description,
  content='issues', content_rowid='rowid'
);

-- triggers for FTS sync...

### Sync Strategy

| Entity | Cursor Field | Batch Size | Frequency |
|--------|-------------|------------|-----------|
| Issues | updatedAt | 100 | Every run |
| Cycles | updatedAt | 20 | Every run |
| Users | updatedAt | 100 | Daily |

### Commands to Build

| Command | Description | Entities Used | API Calls |
|---------|-------------|---------------|-----------|
| sync | Incremental sync to local SQLite | All primary | issues(filter: updatedAt > lastSync) |
| search | FTS5 search with domain filters | issues_fts | Local query only |
| sql | Raw SQL access to local database | All | Local query only |
| trends | Temporal trend detection | issues, cycles | Local aggregation |

### Domain-Specific Filters

| Command | Filter | Maps To |
|---------|--------|---------|
| search | --team | WHERE team_key = ? |
| search | --assignee | WHERE assignee_name LIKE ? |
| search | --state | WHERE state_type = ? |
| search | --since | WHERE updated_at > ? |

### Trend Detections

| Trend | Query Pattern | Alert Threshold |
|-------|---------------|-----------------|
| Velocity change | completed issues per cycle, 3-cycle moving average | > 20% drop |
| SLA breach rate | issues where sla_breaches_at < now, grouped by week | > 10% of open |
| Stale issue growth | issues where updated_at < 30d ago, weekly count | > 15% increase |
```

### Phase Gate 0.7

**STOP.** Verify ALL of these before proceeding:
1. Entity classification table with volume/frequency estimates
2. At least 3 social signals with evidence scores
3. Data gravity scores for all entities with >= 1 primary entity (score >= 8)
4. At least 3 compound features validated against the actual API
5. SQLite schema with FTS5 virtual tables for text-heavy entities
6. Sync strategy with cursor fields and batch sizes
7. Domain-specific search filters mapped to SQL WHERE clauses

Tell the user: "Phase 0.7 complete: [N] primary entities for SQLite, [M] compound features validated. Schema: [list tables]. Key prediction: [most valuable feature]. Proceeding to deep research."

## Implementation Phases

### Phase 1: Modify SKILL.md (1-2 hours)

Add Phase 0.7 between Phase 0.5 and Phase 1 in the printing-press SKILL.md.

**Files to modify:**
- `~/.claude/skills/printing-press/SKILL.md` - Insert Phase 0.7 section
- Update the phase flow diagram at the top
- Update total expected time (add 15-25 min)
- Add Phase Gate 0.7 verification checklist
- Update Anti-Shortcut Rules with prediction engine red flags

**New anti-shortcut rules to add:**
- "The API doesn't need local persistence" -> Check data gravity scores. If any entity scores >= 8, it needs SQLite.
- "FTS5 is overkill for this API" -> If any entity has text fields (title, description, body, content) and scores >= 8, it needs FTS5.
- "REST polling is fine for tail" -> Check if the API has WebSocket/SSE/Gateway. If yes, use it.
- "The generic store is good enough" -> Domain-native tables always beat JSON blob tables. Write the schema.

### Phase 2: Update Phase 4 Integration (30 min)

Modify Phase 4's Priority structure to incorporate the data layer:

**Current Phase 4 priorities:**
1. Power User Workflows (from Phase 0.5)
2. Scorecard-Gap Fixes
3. Polish

**New Phase 4 priorities:**
1. **Data Layer** (from Phase 0.7) - SQLite schema, sync, search, sql commands
2. Power User Workflows (from Phase 0.5) - now powered by the data layer
3. Scorecard-Gap Fixes
4. Polish

This means workflow commands should USE the local database when available, not make fresh API calls. Example: `linear-cli stale` should query the local SQLite instead of hitting the API.

### Phase 3: Test on Linear and Discord (2-4 hours)

Re-run the printing-press on both APIs with the prediction engine enabled:

1. **Linear re-run:** Verify the engine predicts:
   - issues/cycles/users tables (not generic JSON blobs)
   - Incremental sync via updatedAt
   - FTS5 on issue title+description
   - Domain-specific search filters (--team, --assignee, --state)
   - Velocity trend detection
   - `sql` command

2. **Discord re-run:** Verify the engine predicts:
   - messages/members/channels tables
   - Snowflake-based incremental sync for messages
   - FTS5 on message content + embed text
   - Domain-specific filters (--guild, --channel, --author)
   - Gateway WebSocket for tail (not REST polling)
   - `sql` command
   - Mention tracking

3. **Compare against discrawl:** Did the engine predict discrawl's features without knowing discrawl exists?

### Phase 4: Validate on a Third API (1-2 hours)

Run on an API we haven't tried (e.g., Stripe, Notion, Slack) to verify generalization:
- Does the engine correctly identify primary entities?
- Do the social signals produce actionable feature predictions?
- Is the SQLite schema domain-appropriate?
- Are the sync cursors correct for that API's pagination model?

## Acceptance Criteria

- [ ] Phase 0.7 section added to printing-press SKILL.md with all 5 steps (0.7a-0.7e)
- [ ] Phase Gate 0.7 checklist with 7 verification items
- [ ] Anti-shortcut rules updated with 4 new prediction engine rules
- [ ] Phase 4 priority order updated (data layer first, then workflows)
- [ ] Data Layer Specification template with: schema, sync strategy, commands, filters, trends
- [ ] Phase flow diagram updated to show Phase 0.7
- [ ] Total time estimate updated
- [ ] Linear re-run produces domain-specific SQLite schema (not generic JSON blobs)
- [ ] Discord re-run predicts discrawl-class features without seeing discrawl
- [ ] Entity classification heuristics documented (accumulating vs reference vs computed)
- [ ] Data gravity scoring formula documented with weights
- [ ] Compound feature synthesis validates against actual API capabilities (grounding step)

## What This Enables

**Before (current printing-press):**
```
API Spec -> Generator -> 316 API wrapper commands + generic store
                         + 7 workflow commands (Phase 4 hand-written)
                         = Good API client, weak data tool
```

**After (with prediction engine):**
```
API Spec -> Generator -> 316 API wrapper commands
         -> Prediction Engine -> Domain-specific SQLite schema
                              -> Incremental sync with proper cursors
                              -> FTS5 search with domain filters
                              -> Trend detection + alerting
                              -> sql command for ad-hoc queries
         -> Phase 4 -> Workflow commands powered by local data
                    = Strong API client + strong data tool
                    = discrawl-class without knowing discrawl exists
```

The prediction engine doesn't replace Phase 0.5's workflow ideation - it ADDS a data layer underneath those workflows. `linear-cli stale` goes from "make 3 API calls, filter in memory" to "query local SQLite where updated_at < 30d ago, join with assignees, instant results."

## Risk Analysis

| Risk | Mitigation |
|------|-----------|
| Over-prediction: engine recommends SQLite for simple APIs that don't need it | Data gravity scoring - entities must score >= 8 to warrant local persistence |
| Hallucinated schema: columns that don't exist in the API | Step 0.7d validates every feature against actual API capabilities |
| GraphQL APIs don't have OpenAPI specs | Phase 0.7a works from GraphQL introspection too (schema.graphql) |
| Social signals are noisy | Evidence scoring from Phase 0 (stars, upvotes) filters weak signals |
| Sync cursor identification is wrong | Step 0.7d explicitly verifies the cursor field is queryable/filterable |
| Over-engineering: 25 minutes adds cost to every run | Engine produces a spec document, not code. Phase 4 implements selectively. |

## Sources

- discrawl gap analysis: `docs/plans/2026-03-26-feat-discord-cli-vs-discrawl-gap-analysis-plan.md`
- Linear CLI run: `docs/plans/2026-03-25-feat-linear-cli-visionary-research.md`
- Discord CLI run: `docs/plans/2026-03-25-feat-discord-cli-visionary-research.md`
- Printing-press skill: `~/.claude/skills/printing-press/SKILL.md`
- Profiler source: `~/cli-printing-press/internal/profiler/profiler.go`
- Vision templates: `~/cli-printing-press/internal/generator/vision_templates.go`
me id