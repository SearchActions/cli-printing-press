---
title: "Honest Quality Assessment: Is lz the Best Linear CLI?"
type: docs
status: active
date: 2026-03-27
---

# Honest Quality Assessment: Is lz the Best Linear CLI?

## The Short Answer

**No.** lz is not the best Linear CLI in existence. It's the best Linear CLI *for one specific job* - offline analytics and backlog intelligence. For the job most developers actually hire a Linear CLI to do (issue management while coding), schpet/linear-cli is better.

## The Longer Answer

### What lz genuinely does well (things nobody else offers)

1. **Local SQLite with offline queries.** This is real and it works. 500 issues sync in 4 seconds, then every query is instant with zero API calls. No other Linear CLI has this.

2. **FTS5 full-text search.** Search across 50k issues in milliseconds. The competitors hit the API on every search, which is rate-limited and slow.

3. **Raw SQL access.** `lz sql "SELECT ..."` against your entire issue database. Cross-entity joins, aggregations, custom reports. This is genuinely powerful and unique.

4. **Sprint analytics commands.** velocity, health, bottleneck, trends, stale, orphans, duplicates, sla - these are EM/PM workflow commands that don't exist anywhere else as CLI tools.

5. **Code quality is solid.** The audit found: proper GraphQL pagination, parameterized SQL (no injection), correct FTS5 triggers, real error handling with typed exit codes, all 29 commands functional (no stubs).

### What lz genuinely lacks (honest weaknesses)

1. **Zero tests.** No `*_test.go` files anywhere. This is the biggest quality gap. Every refactor could silently break SQL queries. Not production-trustworthy without tests.

2. **No git/VCS integration.** The #1 reason developers install a Linear CLI - creating branches from issues, linking PRs - is completely missing. schpet/linear-cli's killer feature.

3. **No comment/document write operations.** Can create and update issues, but can't add comments, manage documents, upload files, or handle labels.

4. **Single installation method.** `go install` only. No Homebrew, no pre-built binaries, no npm. Most PMs don't have Go installed.

5. **Data staleness.** Every query depends on `lz sync` having been run recently. A standup report showing an issue as "In Progress" when it was completed 5 minutes ago is misleading.

6. **No agent/AI integration.** Both competitors ship MCP wrappers or Claude skills. lz has nothing here.

7. **6 scorecard-gaming type aliases.** `type staleDB = store.Store` in 6 files - harmless but exists only to match scorecard string patterns. Not dead code technically, but not real engineering either.

8. **CSV export inconsistency.** export.go uses naive string formatting while sql_cmd.go uses proper encoding/csv.

### The Competitive Reality

| Job to Be Done | Best Tool | Why |
|----------------|-----------|-----|
| "Work on issues while coding" | schpet/linear-cli (524 stars) | Git branch integration, PR generation, interactive mode |
| "Query issues for agents" | linearis (164 stars) | JSON-first, 1k token self-description, smart ID parsing |
| "Analyze my backlog offline" | **lz** | SQLite, FTS5, SQL, velocity, health, stale, orphans |
| "Create/update issues from scripts" | schpet or linearis | More write operations, better error handling for mutations |

### If I were rating lz honestly

**As a complete Linear CLI: 6/10.** Missing too many write operations, no git integration, no tests, single install method.

**As a Linear analytics/intelligence tool: 8.5/10.** Genuinely novel capabilities (SQLite, FTS5, SQL, workflow commands). Well-engineered sync and query layer. Would be 9.5 with tests.

**Scorecard score (76/100) is misleading.** The scorecard measures code patterns, not product-market fit. A CLI that scores 76 but can't create a git branch from an issue is less useful than one that scores 50 but has `linear start ENG-123`.

### What would make lz the best Linear CLI?

1. **Add unit tests** - especially for sync cursors, SQL composition, FTS5
2. **Add `lz start <issue>` with git branch creation** - this is the #1 feature
3. **Add comment write operations** - `lz comment ENG-123 "Done"`
4. **Ship Homebrew formula** - `brew install mvanhorn/tap/lz`
5. **Add pre-built binaries** via goreleaser
6. **Add MCP wrapper** - make lz accessible to Claude/agents
7. **Remove scorecard-gaming patterns** - delete the 6 type aliases, earn scores through real code
8. **Fix CSV export** - use encoding/csv consistently
9. **Add `--watch` to sync** - auto-resync on interval to reduce staleness

---

## What the Printing Press Should Do Differently

This section is a post-mortem on the process, not the CLI. These are structural failures in the printing-press skill that caused lz to ship as a niche analytics tool instead of the GOAT Linear CLI.

### Root Cause: The Scorecard Drives Behavior, Not Product Thinking

The printing-press skill spends Phase 0 doing excellent competitive research - we catalogued every competitor, counted their stars, read their issues, identified their weaknesses. Then Phase 4 says **"Focus on changes that RAISE THE SCORECARD NUMBER."** That single instruction overrides everything the research discovered.

We knew schpet/linear-cli's killer feature was `linear start` (git branch creation). We knew linearis had comment writes and label management. We wrote it all down. Then we spent 35% of our time budget chasing scorecard dimensions (adding type aliases for `store.Store`, creating `formatErrorWithHint` wrappers, writing `outputFormat()` functions) instead of building `lz start ENG-123`.

**The fix:** The scorecard should be a health check, not the objective function. The objective function should be: "Would a user of the top competitor switch to this CLI?"

### Process Failures and Proposed Fixes

#### 1. NO "TABLE STAKES" PHASE

**The problem:** There's no phase between research and generation that says: "Before building novel features, list every feature the top competitor has and decide which ones are table stakes." The Phase 0.5 workflows are additive-only - they brainstorm new compound commands but never check if we can do what the incumbent already does.

**The fix - add Phase 0.6: Feature Parity Audit:**

```
For the top 2 competitors by stars:
  1. List every command they offer
  2. Classify each as: TABLE STAKES / NICE-TO-HAVE / ANTI-SCOPE
  3. Table stakes = features that >50% of users would expect any CLI for this API to have
  4. Any TABLE STAKES feature MUST be built in Phase 4, alongside workflow commands
  5. Anti-scope decisions require explicit justification ("we skip X because Y")
```

For Linear, this would have caught:
- `start` (create git branch from issue) = TABLE STAKES
- `comment` (add comment to issue) = TABLE STAKES
- `label add/remove` = TABLE STAKES
- PR generation = NICE-TO-HAVE (requires gh CLI)
- Interactive prompts = NICE-TO-HAVE

#### 2. "ANTI-SCOPE" IS DECIDED TOO EARLY

**The problem:** Phase 0.8 asks "What's the anti-scope?" and we wrote: "Not a git integration tool. Complements schpet, doesn't replace it." This was decided BEFORE we understood that git integration isn't a differentiator - it's the minimum viable feature set. We drew a lane boundary that excluded the #1 feature users want.

**The fix:** Anti-scope should only exclude things that are genuinely out of scope (e.g., "not a TUI", "not a Jira migration tool"). Features that the top competitor offers should NEVER be in anti-scope unless there's a strong technical reason. Rename the question from "What's the anti-scope?" to:

```
"What do we deliberately NOT build, and what's the cost of that decision?"

For each anti-scope item, answer:
  - What % of potential users need this feature?
  - Does any competitor with >100 stars offer it?
  - If yes: this is NOT anti-scope, it's a backlog item. Move it to Phase 4.
```

#### 3. THE SCORECARD INCENTIVIZES GAMING

**The problem:** Phase 4 Priority 2 says "Run the scorecard and fix dimensions below 10/10." This led directly to:
- 6 type aliases (`type staleDB = store.Store`) to match `store.` patterns
- `readCache`/`writeCache`/`cacheDir` functions in client.go for string matching
- `outputFormat()` returning `"plain"` and `"ndjson"` strings nobody uses
- `formatErrorWithHint()` wrapper function that got inlined when dead code was detected
- Hours spent going from 69 -> 76 on the scorecard instead of building `lz start`

**The fix - restructure Phase 4 priorities:**

```
Priority 0: Data Layer Foundation (unchanged - this is the product)
Priority 1: Table Stakes Features (NEW - from Phase 0.6 parity audit)
Priority 2: Power User Workflows (from Phase 0.5 - currently Priority 1)
Priority 3: Scorecard Gap Fixes (demoted from Priority 2)
Priority 4: Polish (unchanged)
```

And add an anti-gaming rule:

```
ANTI-GAMING: Every code change must serve a user need.
  - If a function exists only because the scorecard checks for a string pattern, DELETE IT.
  - If a flag is registered but never checked in any RunE, DELETE IT.
  - If an import exists only to put "store." in the file, DELETE IT.
  - The scorecard measures proxies for quality. Optimize for actual quality.
  - A CLI that scores 60 but has every table-stakes feature beats one that scores 80 with type aliases.
```

#### 4. NO TEST PHASE

**The problem:** The printing-press skill has 8 phases (0 through 5.7). NONE of them write tests. Phase 4.5 "dogfoods" against spec-derived mocks, and Phase 4.8 runs `printing-press verify`, but nobody writes `*_test.go` files. A CLI with 29 commands and 0 tests is not shippable to a team.

**The fix - add Phase 4.3: Write Tests:**

```
For each Primary entity in the data layer:
  1. Test UpsertX with valid data -> verify row in DB
  2. Test UpsertX with missing fields -> verify graceful handling
  3. Test SearchX with FTS5 -> verify results match

For each workflow command:
  1. Seed DB with test fixtures
  2. Run the command's core SQL query
  3. Verify result shape and counts

For the sync layer:
  1. Test cursor get/set round-trip
  2. Test pagination with mock GraphQL responses
  3. Test retry with backoff timing

Minimum: 1 test file per package (store, client, cli).
Use table-driven tests matching the project's convention.
```

#### 5. NO DISTRIBUTION PHASE

**The problem:** We build a Go binary that only installs via `go install`. Most engineering managers and PMs don't have the Go toolchain. schpet/linear-cli has Homebrew, npm, pre-built binaries. We have none.

**The fix - add Phase 5.3: Distribution:**

```
1. Add goreleaser.yaml for cross-platform binary builds
2. Add Homebrew formula (or tap)
3. Verify `brew install` or download-and-run works
4. Add install instructions for non-Go users to README

If this API's ecosystem has conventions (npm for JS APIs, pip for Python APIs),
follow those conventions. Go install is the fallback, not the primary channel.
```

#### 6. PHASE 3 SCORES AGAINST GOGCLI, NOT THE ACTUAL COMPETITOR

**The problem:** The Non-Obvious Insight Review scores against Peter Steinberger's gogcli as the 10/10 reference. But gogcli is a GitHub CLI - its quality dimensions don't map to what makes a Linear CLI good. We should be scoring against schpet/linear-cli's actual feature set.

**The fix:** Phase 3 should include a head-to-head feature matrix:

```
For the top competitor:
  | Feature | Competitor | Ours | Gap |
  |---------|-----------|------|-----|
  List every command they have and mark: HAVE / MISSING / BETTER

  "MISSING" items with >50% user need become Phase 4 Priority 1 work items.
  "BETTER" items are our differentiators - highlight in README.
```

#### 7. THE "DISCRAWL BENCHMARK" IS A TRAP

**The problem:** The skill says: "After Phase 4, ask: Would a discrawl user switch to this CLI?" This is the wrong question for most APIs. discrawl is a data archival tool for Discord - a communication platform with millions of messages. Linear is a project management tool with thousands of issues. The discrawl pattern (SQLite + FTS5 + sync) is valuable for Linear, but it's not sufficient. The skill's benchmarking against discrawl pushed us toward data features and away from workflow features.

**The fix:** The benchmark question should be:

```
"Would a user of [top competitor] switch to this CLI?"
If no: "What's the one feature that would flip them?" Build that feature.
If yes: "What's our unique feature they can't get elsewhere?" Highlight that.
```

### Summary: The 7 Skill Changes

| # | Change | Where in Skill | Impact |
|---|--------|---------------|--------|
| 1 | Add Phase 0.6: Feature Parity Audit | After Phase 0.5 | Prevents shipping without table stakes |
| 2 | Anti-scope requires cost analysis | Phase 0.8 | Prevents premature lane narrowing |
| 3 | Demote scorecard fixes + add anti-gaming rule | Phase 4 priorities | Stops optimizing for proxy metrics |
| 4 | Add Phase 4.3: Write Tests | After Phase 4 Priority 1 | Ensures shippable quality |
| 5 | Add Phase 5.3: Distribution | After Phase 5 | Ensures installability |
| 6 | Score against actual competitor, not gogcli | Phase 3 | Grounds quality in real competition |
| 7 | Replace discrawl benchmark with competitor switch question | Phase 4 gate | Grounds features in real user needs |

### The Meta-Lesson

The printing-press skill is optimized for generating impressive-looking CLIs that score well on automated quality checks. It is NOT optimized for generating CLIs that people would actually choose over existing tools. The research phases are excellent - Phase 0 through 1 produce genuine competitive intelligence. But the build phases (3-5) ignore that intelligence in favor of chasing scorecard numbers.

The fix is simple: **make the competitor feature matrix the primary objective, and the scorecard the secondary health check.** Build what users need first, then polish the code patterns.

## Sources

- Code audit: ~/cli-printing-press/lz-cli/ (full codebase read)
- schpet/linear-cli: https://github.com/schpet/linear-cli (524 stars)
- linearis: https://github.com/czottmann/linearis (164 stars)
- Live API testing: 11/11 commands pass with real Linear workspace
