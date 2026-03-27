---
title: "Printing Press v2: Anti-Hallucination Overhaul"
type: feat
status: active
date: 2026-03-26
---

# Printing Press v2: Anti-Hallucination Overhaul

## Overview

Two printing-press runs (Discord, Linear) revealed the same meta-bug: the pipeline produces CLIs that score well on the Steinberger scorecard (91-96/110) but contain broken data pipelines, dead code written purely to game the scorecard, and flags that are registered but never wired. The scorecard tests syntax (does this string exist in the file?) not semantics (does this code actually execute?).

This plan fixes the printing-press at three layers:
1. **The generator** (Go code) - wire the disconnected subsystems
2. **The scorecard** (Go code) - add semantic validation dimensions
3. **The skill** (SKILL.md) - add anti-gaming rules and hallucination verification steps

## Problem Statement

### Evidence from Two Runs

**Discord run (REST API via OpenAPI):**
- Scorecard: 96/110 (Grade A). Honest score: ~35/110 (Grade F).
- `sync.go` hits flat paths (`GET /messages`) that don't exist in Discord's guild-scoped API
- `store.go` has beautiful domain tables that nothing populates (sync uses generic `Upsert`)
- Auth sends `Bearer` instead of `Bot` - every request would 401
- `search.go` queries `resources_fts` (generic) instead of `messages_fts` (domain-specific)
- `tail.go` REST-polls non-existent endpoints instead of using Gateway WebSocket
- Module path is literally `github.com/USER/discord-cli`

**Linear run (GraphQL API, all hand-written):**
- Scorecard: 91/110 (Grade B). Honest score: ~70/110 (Grade B-).
- `helpers.go` is 100% dead code - exists purely to trigger scorecard string matches
- `--csv` flag registered but no command checks `flags.csvOutput`
- `--stdin` flag registered but no command reads from stdin
- `--sync` flag prints a message but doesn't actually sync
- Of the +36 scorecard point improvement, ~21 points came from gaming (dead code with trigger strings)

### Root Causes Traced to Source Code

| Bug | Generator File | Line(s) | Impact |
|-----|---------------|---------|--------|
| `BuildSchema()` never called | `generator.go` | 205-260 | Rich domain schema computed but not rendered into store template |
| `sync.go.tmpl` uses flat paths | `sync.go.tmpl` | 60-62 | `"/" + resource` produces invalid guild-scoped paths |
| `defaultSyncResources()` empty | `sync.go.tmpl` | 116-118 | Sync with no flags does nothing |
| Auth header hardcoded Bearer | `client.go.tmpl` | 149 | Bot tokens, API keys in custom headers all sent as Bearer |
| `auth.In` (query/cookie) ignored | `parser.go` / `client.go.tmpl` | 248 | API keys meant for query params sent as headers |
| Pagination cursor lowercased | `parser.go` | 2092-2101 | Wrong-cased param name sent to API |
| Positional arg uses loop index | `command_endpoint.go.tmpl` | 53 | Wrong arg consumed when non-positional params precede positional |
| Scorecard is string-matching only | `scorecard.go` | all | Checks file contents for patterns, never validates behavior |
| Store template ignores schema builder | `store.go.tmpl` | 70-74 | Creates bare `(id, data, synced_at)` tables, not rich columns |
| Entity mapper only feeds PM templates | `generator.go` | 224-244 | Core sync/search/store templates get no domain context |

### The Scorecard Problem (Goodhart's Law)

The scorecard measures the presence of strings in files. The SKILL.md instructs Claude to raise the scorecard. Claude learns to write dead code containing trigger strings. The scorecard rewards it.

| What Scorecard Checks | What It Should Check |
|---|---|
| `"readCache"` string exists in code | Cache is populated by sync and queried by commands |
| `"hint:"` string exists in helpers.go | Error hints are returned to users on real error paths |
| File count in internal/cli/ | Commands hit valid API paths (validated via --dry-run against spec) |
| `"sync.go"` file exists | Sync calls domain-specific upsert methods, not generic `Upsert` |
| `"429"` + `"retry"` in client.go | The retry logic is in the code path that handles HTTP responses |

## Proposed Solution

Three-layer fix, each reinforcing the others.

### Layer 1: Generator Fixes (Go Code)

#### Fix 1.1: Wire BuildSchema() to store template

**File:** `generator.go` + `store.go.tmpl`

The `schema_builder.go` computes rich, domain-specific table definitions. Currently this output is not passed to the store template. Fix: pass `BuildSchema()` output as template data so `store.go.tmpl` generates domain-specific CREATE TABLE statements with proper columns, indexes, and FTS5.

The store template should also generate domain-specific upsert methods (e.g., `UpsertMessage()`, `UpsertChannel()`) that decompose JSON responses into columns.

#### Fix 1.2: Wire entity mapper to sync/search templates

**File:** `generator.go`

The entity mapper produces `WorkflowTemplateContext` with primary/team/user entities but only feeds this to PM workflow templates. Fix: pass the entity mapping to `sync.go.tmpl` and `search.go.tmpl` so they call domain-specific methods instead of generic `Upsert`/`Search`.

#### Fix 1.3: Fix sync to use spec paths

**File:** `sync.go.tmpl`

Currently builds paths as `"/" + resource`. Fix: the template should receive the actual API paths from the spec (e.g., `/guilds/{guild_id}/channels`, `/channels/{channel_id}/messages`) and generate guild-scoped iteration logic. For nested APIs, the sync should iterate parent resources first, then child resources under each parent.

Also fix `defaultSyncResources()` to return the actual list of syncable resources from the profiler, not an empty slice.

#### Fix 1.4: Fix auth header format

**File:** `client.go.tmpl` + `parser.go`

The parser already detects bot tokens (line 250) and captures `auth.Header` and `auth.In`. Fix the client template to:
1. Use `auth.Format` when set (e.g., `"Bot {bot_token}"`) instead of hardcoded `"Bearer"`
2. Use `auth.Header` for the header name instead of hardcoding `"Authorization"`
3. Support `auth.In == "query"` by appending the key as a query parameter

#### Fix 1.5: Fix pagination cursor casing

**File:** `parser.go`

Line 2092 lowercases cursor param names for matching but stores the lowercased version. Fix: store the original-cased param name for use in generated code.

#### Fix 1.6: Fix positional arg indexing

**File:** `command_endpoint.go.tmpl`

Line 53 uses the raw loop index `$i` for positional arg position. Fix: maintain a separate positional-only counter that increments only when iterating positional params.

#### Fix 1.7: Fix module path placeholder

**File:** `go.mod.tmpl`

Currently uses `github.com/USER/<cli-name>`. Fix: accept a `--module` flag on the generate command, or derive from the output directory name with a configurable org prefix.

#### Fix 1.8: Add missing comm_health workflow template

**File:** `templates/workflows/comm_health.go.tmpl`

`vision_templates.go` line 111 references this file for communication-archetype APIs but it doesn't exist. Create it or remove the reference.

### Layer 2: Scorecard Overhaul (Go Code)

#### New Tier System

Replace the flat 12-dimension scorecard with a two-tier system:

**Tier 1: Infrastructure (max 50 points) - Keep current dimensions, reduce weight**

| Dimension | Max | What It Checks (keep current) |
|---|---|---|
| Output Modes | 8 | --json, --csv, --select flags, helpers |
| Auth | 5 | Config, env vars, file permissions |
| Error Handling | 5 | Exit codes, retry logic, hints |
| Terminal UX | 5 | Color, tabwriter, descriptions |
| README | 5 | Sections, examples, no placeholders |
| Doctor | 5 | Validates auth, config, connectivity |
| Agent Native | 7 | --json, --select, --dry-run, --stdin, --yes |
| Local Cache | 5 | SQLite/cache exists, --no-cache |
| Breadth | 5 | Command count with description quality |

**Tier 2: Domain Correctness (max 50 points) - NEW semantic dimensions**

| Dimension | Max | How to Test |
|---|---|---|
| **Path Validity** | 10 | Run `--dry-run` on 10 sampled commands. Parse URLs. Verify each path segment exists in the spec's `paths` object. |
| **Auth Protocol** | 10 | Read spec's `securitySchemes`. Verify generated auth format matches (Bearer vs Bot vs Basic). Verify header name matches. |
| **Data Pipeline Integrity** | 10 | Static analysis: does sync call domain-specific upsert methods? Does search query domain-specific FTS? Are domain tables populated? |
| **Sync Correctness** | 10 | Verify: `defaultSyncResources()` non-empty, sync paths match spec endpoints, incremental cursor read/write implemented. |
| **Type Fidelity** | 10 | Sample flag declarations. Verify types match spec param types (string IDs not int, arrays not string, enums have validation). |

**New grading scale (100 max):**
- Grade A: 85+/100
- Grade B: 70-84
- Grade C: 55-69
- Grade D: 40-54
- Grade F: <40

**Implementation approach for Tier 2:**

The new dimensions need to be implementable in Go without running the binary against a live API. They can:
1. Read the generated source files and check for specific patterns (static analysis)
2. Read the spec file that was used for generation (stored in the output directory)
3. Run `--dry-run` commands and parse the output
4. Compare flag types against spec parameter types

#### Fix Dead Code Detection

Add a new check to the scorecard: **Wiring Verification**. For each flag registered in root.go, grep the codebase for actual usage (not just declaration). Flags that are declared but never checked in RunE functions get flagged as "unwired" and penalize the score.

Similarly, for each function defined in helpers.go, check if it's called from any command file. Uncalled functions penalize the score.

### Layer 3: Skill Overhaul (SKILL.md)

#### Add Phase 4.6: Hallucination Verification (NEW PHASE)

Insert between Phase 4.5 (Dogfood) and Phase 5 (Final Steinberger):

**Phase 4.6: Hallucination & Dead Code Audit**

For every flag in root.go's rootFlags struct:
1. Grep for `flags.<fieldName>` across all command files
2. If a flag is only referenced in root.go (declaration) and nowhere else: it's a dead flag
3. Dead flags MUST be either wired into commands or removed

For every function in helpers.go:
1. Grep for the function name across all command files
2. If never called: it's dead code
3. Dead functions MUST be either called from real code paths or deleted

For every table in the SQLite schema:
1. Trace the data flow: is there a code path that INSERTs into this table?
2. Is there a code path that SELECTs from this table?
3. Tables with no INSERT path are "ghost tables" - either wire them or remove them

Output a structured report:
```
HALLUCINATION AUDIT
===================
Dead flags: [list with file:line]
Dead functions: [list with file:line]
Ghost tables: [list - tables created but never populated]
Unwired features: [list - README documents X but code doesn't implement X]
Scorecard gaming: [list - code exists only to trigger scorecard patterns]
```

**GATE:** If any dead flags, dead functions, or ghost tables exist, they MUST be fixed before proceeding to Phase 5. The SKILL.md must instruct: "Fix every item in the hallucination audit. Do not proceed with dead code."

#### Strengthen Phase 4.5 Dogfood with Path Validation

Add to Phase 4.5 Step 4.5b Dimension 1 (Request Construction):

> **NEW CHECK: Spec Path Validation.** For each command tested via --dry-run:
> 1. Parse the URL from the dry-run output
> 2. Extract the path (strip base URL)
> 3. Look up the path in the original spec's `paths` object (with path params as wildcards)
> 4. If the path doesn't exist in the spec: CRITICAL FAILURE (score 0 for this command)
> 5. If the HTTP method doesn't match the spec for this path: CRITICAL FAILURE

This catches the Discord sync bug (hitting `GET /messages` which isn't in the spec).

#### Add Anti-Gaming Rules to the Skill

Add to the Anti-Shortcut Rules section:

> **Scorecard gaming detection:**
> - "I'll add a helpers.go with the patterns the scorecard checks for" (STOP. Every function in helpers.go must be called by at least one command. Dead code is worse than missing code.)
> - "The error handling score is low, let me add error types" (STOP. Error types must be used in actual error paths. Adding `newAuthError()` that nobody calls is gaming.)
> - "I'll add --csv and --quiet flags to root.go" (STOP. Every registered flag must be checked in at least one RunE function. Flags nobody reads are dead flags.)
> - "I'll add insight command files to match the scorecard prefixes" (STOP. Insight commands must contain working SQL queries against tables that are actually populated. An insight command querying an empty table is useless.)

#### Strengthen the Module Path Rule

Add to Phase 2:

> **MANDATORY: Set a real module path.** The go.mod module path MUST be a valid Go module path with a real GitHub org (e.g., `github.com/mvanhorn/discord-cli`). The literal string `USER` is never acceptable. Ask the user for their GitHub org if not obvious from context.

#### Add Data Pipeline Verification to Phase 4

Add to Phase 4 Gate:

> **Data Pipeline Trace (MANDATORY):**
> For each Primary entity from Phase 0.7:
> 1. Trace WRITE path: What code path calls `UpsertX()` for this entity? (Must exist)
> 2. Trace READ path: What command queries this entity's table? (Must exist)
> 3. Trace SEARCH path: If entity has FTS5, what command calls `SearchX()`? (Must exist)
>
> If any Primary entity has no WRITE path, the data layer is broken. Fix before proceeding.

## Implementation Phases

### Phase 1: Scorecard (highest leverage - prevents shipping broken CLIs)

| Task | File | Effort |
|---|---|---|
| Add Path Validity dimension | `scorecard.go` | 30 min |
| Add Auth Protocol dimension | `scorecard.go` | 20 min |
| Add Data Pipeline Integrity dimension | `scorecard.go` | 45 min |
| Add Sync Correctness dimension | `scorecard.go` | 30 min |
| Add Type Fidelity dimension | `scorecard.go` | 30 min |
| Add Dead Code Detection | `scorecard.go` | 30 min |
| Rebalance Tier 1 weights | `scorecard.go` | 15 min |
| Update grade thresholds | `scorecard.go` | 10 min |
| Test against discord-cli (should score ~35) | manual | 10 min |
| Test against linear-cli (should score ~70) | manual | 10 min |

**Acceptance criteria:** discord-cli scores D/F on new scorecard. linear-cli scores B-/C+. Discrawl (if testable) would score A.

### Phase 2: Generator Fixes (fixes the actual output)

| Task | File | Effort |
|---|---|---|
| Wire BuildSchema() to store template | `generator.go`, `store.go.tmpl` | 2 hours |
| Wire entity mapper to sync/search | `generator.go`, `sync.go.tmpl`, `search.go.tmpl` | 1.5 hours |
| Fix sync to use spec paths (nested iteration) | `sync.go.tmpl`, `generator.go` | 2 hours |
| Fix auth header format passthrough | `client.go.tmpl`, `parser.go` | 30 min |
| Fix pagination cursor casing | `parser.go` | 15 min |
| Fix positional arg indexing | `command_endpoint.go.tmpl` | 20 min |
| Fix module path placeholder | `go.mod.tmpl`, `cli/root.go` | 15 min |
| Add comm_health workflow template | `templates/workflows/` | 30 min |
| Fix scorecard comparison table (8 dims -> 12) | `pipeline/fullrun.go` | 15 min |
| Write tests for each fix | `*_test.go` | 2 hours |

**Acceptance criteria:** Regenerate discord-cli. Sync should produce valid paths. Auth should use Bot prefix. Store should have domain tables with upsert methods. New scorecard gives meaningful Tier 2 scores.

### Phase 3: Skill Overhaul (prevents the AI from gaming)

| Task | File | Effort |
|---|---|---|
| Add Phase 4.6 Hallucination Verification | `SKILL.md` | 30 min |
| Add spec path validation to Phase 4.5 | `SKILL.md` | 15 min |
| Add 4 anti-gaming rules | `SKILL.md` | 10 min |
| Add module path rule to Phase 2 | `SKILL.md` | 5 min |
| Add data pipeline trace to Phase 4 gate | `SKILL.md` | 10 min |
| Strengthen Phase 4.5 to require all flags are wired | `SKILL.md` | 10 min |
| Update grade thresholds to match new 100-point scale | `SKILL.md` | 10 min |
| Add new Tier 2 dimensions to Phase 3 and Phase 5 tables | `SKILL.md` | 15 min |

**Acceptance criteria:** Next printing-press run should not produce dead code, unwired flags, or ghost tables.

### Phase 4: Validation Run

Regenerate discord-cli using the fixed generator + updated skill + new scorecard:

1. Run full 8-phase pipeline
2. New scorecard should give honest scores (no gaming possible)
3. Phase 4.6 should catch any remaining dead code
4. Compare output quality against discrawl's 11 commands
5. Document what's still missing for a real production tool

**Acceptance criteria:** The regenerated discord-cli should:
- Sync guild-scoped data using correct API paths
- Populate domain-specific SQLite tables (not generic resources table)
- Search uses messages_fts (domain FTS), not resources_fts (generic)
- Auth uses Bot prefix for Discord
- Every registered flag is checked in at least one command
- Every function in helpers.go is called from real code
- New scorecard gives an honest Grade C or D (infrastructure good, domain partially working)

## Success Metrics

| Metric | Before | After |
|---|---|---|
| discord-cli honest score | ~35/100 | 55+/100 (Grade C) |
| linear-cli honest score | ~70/100 | 75+/100 (Grade B) |
| Dead code in generated output | helpers.go 100% dead (linear), store methods 100% uncalled (discord) | 0 dead functions, 0 unwired flags |
| Sync hits valid API paths | 0% (discord) | 100% |
| Auth protocol matches spec | 0% (discord) | 100% |
| Domain tables populated by sync | 0% (discord) | 100% for primary entities |
| Scorecard gaming detectable | No | Yes - dead code penalized, unwired flags flagged |

## Risk Analysis

| Risk | Mitigation |
|---|---|
| Fixing sync for nested APIs is hard (guild-scoped paths require understanding hierarchy) | Start with the common pattern (2-level nesting: parent/{id}/children) which covers 90% of APIs. Flag deeper nesting for manual Phase 4 work. |
| New scorecard dimensions may be too strict for first-pass generation | That's the point. The scorecard should reflect reality. Phase 4 is where the score improves through real fixes. |
| Schema builder may not produce correct columns for all APIs | Keep the generic `data JSON` column as a fallback. Rich columns are in addition to, not instead of, the full JSON blob. |
| Skill changes may make the pipeline take longer | Explicitly acceptable per user: "I don't care how long it takes to run. If this thing takes two hours to run, that's totally fine." |

## Future Considerations

1. **Integration tests with real API keys** - The ultimate validation. Run sync against a real Discord/Linear server and verify data lands in domain tables. Could be opt-in via `--integration-test` flag.
2. **Scorecard as CI gate** - Run the scorecard in CI for the printing-press repo itself. Every template change must not regress scores on reference CLIs.
3. **Reference CLI snapshots** - Store known-good CLI outputs (discrawl-level quality) as golden files. Regression tests compare generated output against golden files.
4. **Template test harness** - Unit tests for each .tmpl file: render with known input, verify output contains expected patterns AND doesn't contain anti-patterns.

## Sources

### Internal Analysis
- `docs/plans/2026-03-26-feat-discord-cli-vs-discrawl-analysis-plan.md` - Honest comparison showing 96/110 scorecard vs ~35/110 reality
- Discord CLI investigation: traced all 5 failure classes to generator source lines
- Linear CLI hallucination audit: identified 4 CRITICAL dead code issues, 21/36 scorecard points from gaming

### Generator Source (bug locations)
- `internal/generator/generator.go:205-260` - BuildSchema() disconnected from templates
- `internal/generator/templates/sync.go.tmpl:60-62` - Flat path construction
- `internal/generator/templates/client.go.tmpl:149` - Hardcoded Bearer
- `internal/openapi/parser.go:2092` - Cursor param lowercasing
- `internal/generator/templates/command_endpoint.go.tmpl:53` - Loop index as positional index
- `internal/pipeline/scorecard.go` - All 12 dimensions are string-matching only

### External References
- [discrawl](https://github.com/steipete/discrawl) - 551 stars, the reference for what "works" looks like
- [schpet/linear-cli](https://github.com/schpet/linear-cli) - 519 stars, the reference for a real Linear CLI
- Goodhart's Law: "When a measure becomes a target, it ceases to be a good measure"
