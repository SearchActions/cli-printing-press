---
title: "feat: Visionary Research Phase - Make the Press Think Like Steinberger"
type: feat
status: completed
date: 2026-03-25
---

# Visionary Research Phase - Make the Press Think Like Steinberger

## The Humiliation

Peter Steinberger looked at Discord's API and didn't think "I should generate 307 Cobra commands from the OpenAPI spec." He thought: "Developers need to search their Discord history. The API rate limits make that painful. I'll mirror everything into SQLite with FTS5 and add a WebSocket tail for live updates."

The printing press looked at the same API and generated an API printer. 307 commands. 90/90 Steinberger scorecard. And completely pointless compared to discrawl. A perfect score on a test that measures the wrong thing.

**The scorecard measures whether a CLI looks like gogcli. It doesn't measure whether a CLI is worth using.**

The press has two existing plans for getting smarter (Intelligence Engine + LLM Brain Before Generation). Both still produce API wrappers - just smarter API wrappers. A smarter API wrapper is still an API wrapper.

This plan adds a phase that asks the question Peter asked: **"What would a thoughtful developer actually build for this API?"**

---

## Problem Statement

### What the Press Does Today

```
Phase 1: Find OpenAPI spec, find competitors, find issues
Phase 2: Generate Cobra commands from spec
Phase 3: Score against Steinberger (pattern matching)
Phase 4: Fix scored dimensions
Phase 5: Re-score, report
```

Every API gets the same treatment. Discord gets 307 commands. Stripe gets 200 commands. Notion gets 50 commands. They all look the same: `<api>-cli <resource> <verb> [flags]`. They're all interchangeable. None of them are discrawl.

### What's Missing

The press never asks:

1. **What do people actually DO with this API?** (Discord: archive servers, search history, moderate, monitor)
2. **What tools exist that AREN'T just API wrappers?** (discrawl: SQLite archive with FTS5 search)
3. **What workflows need LOCAL STATE beyond a 5-minute cache?** (transaction sync, message archive, event streaming)
4. **What would make this CLI worth installing over `curl | jq`?** (intelligence, not just convenience)
5. **What are the next 5 features this CLI could have that would be good for the world?**

### The Discrawl Test

For every API, there's a potential discrawl - a tool that a thoughtful human would build that goes beyond the API surface. The press should discover it.

| API | API Wrapper | Discrawl Equivalent |
|---|---|---|
| Discord | 307 CRUD commands | Archive server to SQLite, FTS5 search, WebSocket tail, mention tracking |
| Stripe | Payment CRUD commands | Local transaction ledger, webhook test server, fixture chains, reconciliation engine |
| Notion | Page/database CRUD | Local Markdown sync, offline editing, full-text search, backlinks graph |
| GitHub | Issue/PR CRUD | Codebase analytics, contributor graphs, review queue dashboard, notification triage |
| Slack | Channel/message CRUD | Channel archive, search across workspaces, conversation export, bot management |
| Linear | Issue CRUD | Sprint analytics, team velocity dashboard, issue dependency graph, standup generator |

The press should discover these ideas BEFORE generating anything.

---

## Proposed Solution: Phase 0 - Visionary Research

Add a new phase BEFORE the current Phase 1 that discovers what a thoughtful CLI would look like. This phase uses the osc-newfeature methodology adapted for CLI generation.

### New Pipeline

```
PHASE 0: VISIONARY RESEARCH (new - 8-12 min)
  0a. API Identity & Domain Understanding
  0b. Usage Pattern Discovery (what people DO with this API)
  0c. Tool Landscape Discovery (what exists BEYOND API wrappers)
  0d. Workflow Analysis (what multi-step sequences matter)
  0e. Architecture Planning (what local state, persistence, protocols)
  0f. Feature Ideation (top 5 features scored by evidence)

PHASE 1: DEEP RESEARCH (existing - enhanced)
  Now informed by Phase 0's visionary plan
  Competitor search now includes non-wrapper tools like discrawl

PHASE 2: GENERATE (existing - enhanced)
  Now generates BOTH API wrapper commands AND visionary commands
  Visionary commands may be stubs with TODO markers

PHASE 3-5: AUDIT, FIX, REPORT (existing)
  Scorecard expanded with "Vision" dimension
```

---

## Technical Approach

### Phase 0a: API Identity & Domain Understanding

**Goal:** Understand what this API IS, not just what endpoints it has.

**Method:** LLM reads the API's homepage, developer docs landing page, and "Getting Started" guide. Extracts:

- **Domain category:** messaging, payments, productivity, infrastructure, analytics
- **Primary users:** Who uses this API? (developers building bots, businesses processing payments, teams managing projects)
- **Core entities:** What are the main objects? (Discord: guilds, channels, messages, members. Stripe: customers, charges, subscriptions)
- **Data characteristics:** Is data write-heavy or read-heavy? Append-only or mutable? Real-time or batch? Large or small?

**Implementation:**

```go
// internal/vision/identity.go
type APIIdentity struct {
    Domain        string   // "messaging", "payments", "productivity"
    PrimaryUsers  []string // "bot developers", "server admins"
    CoreEntities  []string // "guilds", "channels", "messages"
    DataProfile   DataProfile
}

type DataProfile struct {
    WritePattern  string // "append-only", "mutable", "event-sourced"
    Volume        string // "high" (millions of records), "medium", "low"
    Realtime      bool   // does the API have webhooks/websockets?
    SearchNeed    string // "high" (users need to find things), "low"
}
```

**Sources:**
- WebFetch the API docs landing page
- WebFetch the "Getting Started" or "Introduction" page
- WebSearch `"<API name>" developer documentation overview`

---

### Phase 0b: Usage Pattern Discovery

**Goal:** Discover what people ACTUALLY DO with this API. Not what the spec says - what humans need.

**Method:** Adapted from osc-newfeature's demand signal methodology.

**Step 1: Community Research (3-4 parallel searches)**

```
WebSearch: "<API name>" CLI workflow site:reddit.com
WebSearch: "<API name>" automation script site:github.com
WebSearch: "<API name>" "I built" OR "I made" OR "my tool" site:reddit.com OR site:news.ycombinator.com
WebSearch: "<API name>" tutorial automation workflow 2025 2026
```

**Step 2: GitHub Non-Wrapper Tool Discovery**

This is the critical search that would have found discrawl:

```
WebSearch: "<API name>" tool NOT "api wrapper" NOT "sdk" site:github.com
WebSearch: "<API name>" sync OR archive OR export OR backup site:github.com
WebSearch: "<API name>" search OR analytics OR dashboard site:github.com
WebSearch: "<API name>" monitor OR watch OR tail site:github.com
```

For each tool found:
- Is it an API wrapper or something MORE?
- What local functionality does it add? (database, cache, search, streaming)
- What workflow does it enable that raw API calls don't?
- How many stars? Is it maintained?

**Step 3: Stack Overflow / Forum Pain Points**

```
WebSearch: "<API name>" API "pain point" OR "limitation" OR "workaround" OR "wish"
WebSearch: site:stackoverflow.com "<API name>" API rate limit OR pagination OR bulk
```

Extract the top 5 pain points people have with the raw API.

**Step 4: Usage Pattern Synthesis**

From all research, identify the **top 5 usage patterns** ranked by evidence:

```markdown
## Usage Patterns (Evidence-Backed)

### 1. Archive & Search (Evidence: 8/10)
- discrawl (539 stars) mirrors Discord to SQLite
- Reddit: 12 posts asking "how to search old Discord messages"
- Pain point: Discord's built-in search is slow and limited
- **What it needs:** Local database, FTS, resumable sync

### 2. Moderation Automation (Evidence: 6/10)
- discli (6 stars) focuses on server management
- GitHub: multiple moderation bots reference the API
- Pain point: managing bans/roles across channels is tedious
- **What it needs:** Bulk operations, templated actions

### 3. Deploy Notifications (Evidence: 5/10)
- Many CI/CD pipelines post to Discord webhooks
- Reddit: common "how to send deploy notification" questions
- **What it needs:** Webhook execute with templates, --stdin
```

**Evidence scoring (adapted from osc-newfeature):**

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
Score 3-5: Moderate evidence. Consider as optional feature.
Score < 3: Weak evidence. Skip or mark as future work.

---

### Phase 0c: Tool Landscape Discovery

**Goal:** Find ALL tools for this API, not just API wrappers. This is where discrawl would be discovered.

**Method:** Three-tier search strategy.

**Tier 1: Direct CLI Search** (existing Phase 1 does this)
```
WebSearch: "<API name>" CLI tool github
```

**Tier 2: Non-Wrapper Tool Search** (NEW - the discrawl finder)
```
WebSearch: "<API name>" sync OR archive OR mirror site:github.com
WebSearch: "<API name>" search engine OR analytics OR dashboard site:github.com
WebSearch: "<API name>" backup OR export OR migration site:github.com
WebSearch: "<API name>" monitor OR watcher OR alerting site:github.com
WebSearch: "<API name>" local OR offline OR self-hosted site:github.com
```

**Tier 3: Ecosystem Tool Search** (NEW - finds complementary tools)
```
WebSearch: "<API name>" tools ecosystem awesome-list
WebSearch: awesome "<API name>" site:github.com
```

For each tool found, classify it:

| Type | Description | Example |
|---|---|---|
| **API Wrapper** | Translates HTTP to CLI flags | discli |
| **Data Tool** | Adds local persistence/search | discrawl |
| **Workflow Tool** | Orchestrates multi-step sequences | Stripe fixtures |
| **Environment Tool** | Runs local simulation | Supabase CLI |
| **Integration Tool** | Bridges to other systems | Zapier Discord integration |

**The press should generate CLIs that compete with Data Tools and Workflow Tools, not just API Wrappers.**

---

### Phase 0d: Workflow Analysis

**Goal:** Identify multi-step sequences that users perform repeatedly.

**Method:** From usage patterns (0b) and tool landscape (0c), extract workflows:

```markdown
## Workflow Analysis

### Workflow 1: Server Backup & Restore
Steps: list channels -> for each channel: list messages -> save to file -> (restore: create channels -> post messages)
Frequency: Monthly for large servers
Pain: Rate limiting makes this take hours manually
**CLI Feature:** `discord-cli backup <guild_id> --output backup.jsonl` (paginate all channels/messages, write JSONL)
                 `discord-cli restore --input backup.jsonl --guild <target_id>` (replay from JSONL)

### Workflow 2: Moderation Sweep
Steps: search messages by user -> review -> bulk delete -> ban user -> log action
Frequency: Daily for active servers
Pain: Requires multiple API calls, no transaction semantics
**CLI Feature:** `discord-cli moderate <guild_id> --user <id> --action ban --delete-messages 7d --reason "spam"`

### Workflow 3: Channel Analytics
Steps: list messages in date range -> count by author -> count by hour -> generate report
Frequency: Weekly for community managers
Pain: No built-in analytics in Discord
**CLI Feature:** `discord-cli analytics <guild_id> --days 7 --by author,hour --format csv`
```

Each workflow maps to a potential **compound command** - a single command that orchestrates multiple API calls.

---

### Phase 0e: Architecture Planning

**Goal:** Based on usage patterns and workflows, decide what local capabilities the CLI needs.

**Decision Matrix:**

```markdown
## Architecture Decisions

### Local Persistence
- **Need:** High (archive use case requires persistent storage)
- **Decision:** SQLite for structured data, JSONL for export
- **Rationale:** discrawl proved SQLite + FTS5 is the right model for message search
- **Implementation:** `internal/store/` package with migration system

### Real-Time Protocol
- **Need:** Medium (monitoring use case benefits from WebSocket)
- **Decision:** Optional `tail` command using Gateway WebSocket
- **Rationale:** discrawl's tail + repair cycle is the gold standard
- **Implementation:** `internal/gateway/` package, opt-in via `tail` subcommand

### Search
- **Need:** High (top usage pattern is finding messages)
- **Decision:** FTS5 on SQLite for local, API search for remote
- **Rationale:** FTS5 is millisecond-fast, API search is rate-limited
- **Implementation:** `internal/search/` package, `search` top-level command

### Bulk Operations
- **Need:** Medium (moderation and analytics require paginating everything)
- **Decision:** `export` and `import` commands with JSONL format
- **Rationale:** JSONL is streaming-friendly, no memory pressure on large exports
- **Implementation:** Built on existing pagination helpers + JSONL writer

### Caching Strategy
- **Need:** Enhanced (5-minute TTL is useless for analytics)
- **Decision:** Persistent cache mode with SQLite backend
- **Rationale:** Users running analytics queries want data to accumulate
- **Implementation:** Extend existing cache with `--cache-persist` flag
```

**Key principle:** The architecture should match the DATA PROFILE from Phase 0a.

| Data Profile | Architecture |
|---|---|
| High volume + search need | SQLite + FTS5 (Discord, Slack) |
| Transaction data + reconciliation | Local ledger with diff tracking (Stripe, Plaid) |
| Document data + offline editing | Local Markdown/JSON sync (Notion, Confluence) |
| Low volume + simple CRUD | Standard API wrapper is fine (most APIs) |

---

### Phase 0f: Feature Ideation - "Next 5 Features for the World"

**Goal:** Propose the top 5 features this CLI should have that go beyond API wrapping. Scored by evidence.

**Method:** Adapted from osc-newfeature's scoring rubric.

**Scoring each feature idea (16-point scale):**

| Dimension | Weight | Scoring |
|---|---|---|
| **Evidence strength** | 3 | 3=strong (existing tool with 100+ stars), 2=moderate (Reddit/SO demand), 1=weak, 0=speculation |
| **User impact** | 3 | 3=solves a pain felt by most users, 2=solves niche pain, 1=nice-to-have, 0=nobody asked |
| **Implementation feasibility** | 2 | 2=can generate template, 1=needs custom code, 0=requires major infrastructure |
| **Uniqueness** | 2 | 2=no existing tool does this, 1=improves on existing, 0=already well-served |
| **Composability** | 2 | 2=works great with pipes/agents, 1=somewhat, 0=interactive-only |
| **Data profile fit** | 2 | 2=perfect fit for this API's data, 1=possible, 0=wrong shape |
| **Maintainability** | 1 | 1=generated code can support it, 0=requires ongoing human maintenance |
| **Competitive moat** | 1 | 1=hard for others to replicate, 0=trivial |

**Score >= 12: Must-have.** Build it.
**Score 8-11: Should-have.** Include as optional feature.
**Score < 8: Won't-have.** Skip or document as future work.

**Output format:**

```markdown
## Top 5 Features for the World

### 1. Local Archive with Full-Text Search (Score: 15/16)
- Evidence: discrawl (539 stars) proves demand
- Impact: Solves the #1 user pain (searching Discord history)
- Implementation: SQLite + FTS5, sync command
- Uniqueness: Only discrawl does this; our version would be integrated
- Template: `sync.go.tmpl`, `search.go.tmpl`, `store.go.tmpl`

### 2. Bulk Export to JSONL (Score: 13/16)
- Evidence: Multiple GitHub repos for "Discord message export"
- Impact: Enables analytics, migration, backup
- Implementation: Pagination loop + JSONL writer
- Uniqueness: No CLI does this well; DiscordChatExporter is GUI-only
- Template: `export.go.tmpl`

### 3. Webhook Templates (Score: 11/16)
...

### 4. Moderation Toolkit (Score: 10/16)
...

### 5. Channel Analytics (Score: 8/16)
...
```

---

## How This Changes the Generator

### New Templates for Visionary Features

The generator currently has 18 templates. Add templates for domain-specific features:

| Template | When Generated | What It Creates |
|---|---|---|
| `sync.go.tmpl` | Data Profile = high volume + search need | `sync` command with resumable pagination, SQLite upsert |
| `search.go.tmpl` | Data Profile = search need high | `search` command with FTS5 query, filters |
| `export.go.tmpl` | Always | `export` command with JSONL streaming |
| `import.go.tmpl` | Always | `import` command reading JSONL |
| `tail.go.tmpl` | API has WebSocket/SSE | `tail` command with live event streaming |
| `store.go.tmpl` | Data Profile = high volume | SQLite schema, migration, CRUD helpers |
| `analytics.go.tmpl` | Data Profile = high volume | Basic count/group-by analytics commands |
| `workflow.go.tmpl` | Workflows identified | Compound commands (backup, moderate, etc.) |

**Key design:** Templates are opt-in based on Phase 0's architecture decisions. Not every CLI gets SQLite. An API with 5 endpoints and low data volume gets a clean API wrapper and nothing more.

### New Scorecard Dimension: Vision (0-10)

Add a 10th Steinberger dimension:

| Score | Criteria |
|---|---|
| 0 | Pure API wrapper, no intelligence beyond CRUD |
| 3 | Has export/import for bulk data |
| 5 | Has local persistence (SQLite or equivalent) |
| 7 | Has search + persistence + domain workflows |
| 8 | Has domain-specific compound commands (backup, analytics, moderate) |
| 10 | Matches or exceeds the best non-wrapper tool (discrawl level) |

**New max: 100/100.** Grade A = 80+.

Measurement: Automated via presence of store.go, search.go, export.go, workflow commands + their quality.

### Updated Research Artifact

The Phase 0 research produces `visionary-research.md` alongside the existing research artifact:

```markdown
## Visionary Research: <API> CLI

### API Identity
- Domain: messaging
- Primary users: bot developers, server admins, community managers
- Data profile: high volume, append-only, real-time, high search need

### Usage Patterns (Top 5 by Evidence)
1. Archive & search (8/10)
2. Moderation automation (6/10)
3. Deploy notifications (5/10)
4. Analytics & reporting (4/10)
5. Bot management (3/10)

### Tool Landscape (Beyond API Wrappers)
- discrawl (539 stars): SQLite archive + FTS5 + WebSocket tail
- DiscordChatExporter (GUI): message export to HTML/JSON/CSV
- [none found for moderation CLI]

### Workflows
1. Backup: list -> paginate -> save JSONL
2. Moderation sweep: search -> review -> bulk delete -> ban
3. Analytics: paginate messages -> count/group -> CSV report

### Architecture Decisions
- Persistence: SQLite + FTS5
- Real-time: Optional Gateway WebSocket
- Bulk: JSONL export/import
- Cache: Persistent mode

### Top 5 Features for the World
1. Local archive with FTS search (15/16)
2. Bulk JSONL export (13/16)
3. Webhook templates (11/16)
4. Moderation toolkit (10/16)
5. Channel analytics (8/16)
```

---

## Implementation Phases

### Phase 1: The Research Engine (Core)

Build the visionary research phase as Go code + LLM integration.

**Files to create:**
- `internal/vision/identity.go` - API identity extraction (0a)
- `internal/vision/usage.go` - Usage pattern discovery (0b)
- `internal/vision/landscape.go` - Tool landscape discovery (0c)
- `internal/vision/workflow.go` - Workflow analysis (0d)
- `internal/vision/architecture.go` - Architecture planning (0e)
- `internal/vision/ideation.go` - Feature ideation with scoring (0f)
- `internal/vision/report.go` - Writes visionary-research.md

**Dependencies:**
- LLM integration (from existing LLM Brain plan)
- WebSearch capability (from skill context or codex CLI)

**Acceptance criteria:**
- [ ] `printing-press vision --api "Discord"` produces visionary-research.md
- [ ] Research finds discrawl-class tools for Discord, Stripe, and Notion
- [ ] Usage patterns are evidence-backed with scores
- [ ] Architecture decisions match data profile
- [ ] Top 5 features are scored and ranked
- [ ] Phase 0 output feeds into Phase 1 research (competitor search now includes non-wrapper tools)

### Phase 2: Domain-Aware Templates

Create opt-in templates for visionary features.

**Files to create:**
- `internal/generator/templates/sync.go.tmpl` - Resumable sync to SQLite
- `internal/generator/templates/search.go.tmpl` - FTS5 search command
- `internal/generator/templates/export.go.tmpl` - JSONL export with pagination
- `internal/generator/templates/import.go.tmpl` - JSONL import
- `internal/generator/templates/store.go.tmpl` - SQLite schema + migrations
- `internal/generator/templates/tail.go.tmpl` - WebSocket/SSE event streaming (stub)
- `internal/generator/templates/analytics.go.tmpl` - Count/group-by commands

**Generator changes:**
- `internal/generator/generator.go` - Accept VisionaryPlan input, conditionally include templates
- `internal/generator/vision_templates.go` - Template selector based on architecture decisions

**Acceptance criteria:**
- [ ] Discord CLI generates with `sync`, `search`, `export` commands when vision plan says data profile = high volume + search
- [ ] Stripe CLI generates with `export`, `import`, `reconcile` when vision plan says data profile = transaction
- [ ] Simple APIs (Petstore) generate WITHOUT extra templates (no unnecessary complexity)
- [ ] Generated SQLite code compiles and passes `go vet`
- [ ] Generated FTS5 search actually works with test data

### Phase 3: Enhanced Scorecard

Add Vision dimension and update grading.

**Files to modify:**
- `internal/pipeline/scorecard.go` - Add Vision dimension (10th)
- SKILL.md Phase 3/5 sections - Score vision dimension

**Vision scoring automation:**
```go
func scoreVision(dir string) int {
    score := 0
    if fileExists(dir, "export.go")  { score += 2 }
    if fileExists(dir, "import.go")  { score += 1 }
    if fileExists(dir, "store.go")   { score += 2 }  // local persistence
    if fileExists(dir, "search.go")  { score += 2 }  // search capability
    if fileExists(dir, "sync.go")    { score += 1 }  // data sync
    if fileExists(dir, "tail.go")    { score += 1 }  // real-time
    if countWorkflowCommands(dir) > 0 { score += 1 } // compound commands
    return min(score, 10)
}
```

**Acceptance criteria:**
- [ ] `printing-press scorecard --dir ./discord-cli` includes Vision dimension
- [ ] New max is 100/100
- [ ] Grade A threshold updated to 80/100
- [ ] discrawl-equivalent CLIs score 8+/10 on Vision
- [ ] Pure API wrappers score 0-3/10 on Vision

### Phase 4: Skill Update

Update the printing-press SKILL.md to include Phase 0.

**Changes to SKILL.md:**
- Add Phase 0 between "Step 0: Parse intent" and current "Phase 1: Deep Research"
- Phase 1 research now uses Phase 0's tool landscape (searches for non-wrapper tools)
- Phase 3 audit includes Vision dimension
- Phase 5 report includes Vision score
- Anti-shortcut rules expanded: "This is just an API wrapper" (run Phase 0 again)

**New anti-shortcut rule:**
> "This API doesn't need local persistence" - Did you run Phase 0? Did you check the data profile? If search need is high, it needs persistence. Don't guess.

---

## Alternative Approaches Considered

### 1. Always Generate Everything (Rejected)
Generate SQLite + FTS + sync + search for EVERY API. Rejected because most APIs don't need it. Petstore with 8 endpoints doesn't need a local database. The vision phase gates on data profile.

### 2. Post-Generation Only (Rejected)
Run vision research AFTER generating the API wrapper and suggest additions. Rejected because architecture decisions (SQLite vs no-SQLite) affect the entire codebase structure. Must be decided before generation.

### 3. Manual Configuration (Rejected)
Let the user say `--with-sqlite --with-search`. Rejected because the whole point is that the press should figure this out through research, like Peter did. If the user has to tell it what to build, it's not visionary.

### 4. Separate Tool (Rejected)
Build a separate "discrawl generator" tool. Rejected because the value is integration. One tool that generates BOTH the API wrapper AND the visionary features, with the research to decide which visionary features matter.

---

## Success Metrics

### The Discrawl Test
Run the press on Discord. Does it:
- [ ] Discover discrawl during Phase 0c tool landscape search?
- [ ] Identify "archive & search" as the top usage pattern?
- [ ] Decide on SQLite + FTS5 architecture?
- [ ] Generate `sync`, `search`, `export` commands alongside API wrapper commands?
- [ ] Score 8+/10 on the Vision dimension?

### The Stripe Test
Run the press on Stripe. Does it:
- [ ] Discover that Stripe CLI has fixtures and webhook forwarding?
- [ ] Identify "transaction reconciliation" as a key workflow?
- [ ] Generate `export` and `reconcile` commands?
- [ ] Suggest local transaction ledger for analytics?

### The Notion Test
Run the press on Notion. Does it:
- [ ] Identify "offline editing" and "search across pages" as top patterns?
- [ ] Discover notion-to-markdown and similar sync tools?
- [ ] Generate `sync` and `search` commands for Notion databases?

### Quantitative
- Phase 0 adds 8-12 minutes to the pipeline (acceptable given 15-25 minute baseline)
- At least 3 of 5 feature ideas should be evidence-backed (score >= 6)
- Generated visionary features should compile without errors
- Vision scorecard dimension should correlate with actual CLI usefulness

---

## Dependencies & Prerequisites

- **LLM Brain plan** (existing) - Phase 0 needs LLM to read docs and synthesize patterns
- **Intelligence Engine plan** (existing) - Phase 0 feeds into dynamic plan generation
- **WebSearch/WebFetch** - Phase 0b/0c need web research capabilities
- **SQLite Go library** - Templates need `modernc.org/sqlite` (pure Go, no CGO)
- **Generator template system** - Must support conditional template inclusion

---

## Sources & References

### Internal References
- Competitive analysis: `docs/plans/2026-03-25-feat-discord-cli-vs-discrawl-competitive-plan.md`
- Existing intelligence plan: `docs/plans/2026-03-25-feat-press-intelligence-engine-plan.md`
- Existing LLM brain plan: `docs/plans/2026-03-25-feat-llm-brain-before-generation-plan.md`
- OSC newfeature skill: `~/.claude/skills/osc-newfeature/SKILL.md` (research methodology)
- OSC social skill: `~/.claude/skills/osc-social/SKILL.md` (signal extraction)

### External References
- discrawl: https://github.com/steipete/discrawl (the benchmark for "visionary CLI")
- gogcli: https://github.com/steipete/gogcli (Steinberger scorecard reference)
- Stripe CLI fixtures: https://docs.stripe.com/cli/fixtures (workflow orchestration pattern)
- Supabase CLI local dev: https://supabase.com/docs/guides/local-development/overview (environment simulation)
- CLI design principles: https://clig.dev/ (composition, progressive discovery)
- Atlassian CLI principles: https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis (suggest next command)

### The Core Insight
The printing press currently asks: **"What does the OpenAPI spec say?"**
After this plan, it will ask: **"What would a thoughtful developer build?"**

That's the difference between an API printer and a printing press.
