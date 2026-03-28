---
title: "Comprehensive Overhaul: Printing Press Skill v2"
type: fix
status: completed
date: 2026-03-27
origin: synthesis of 2026-03-27-fix-printing-press-post-mortem-notion-run-plan.md + 2026-03-27-docs-lz-cli-honest-quality-assessment.md
---

# Comprehensive Overhaul: Printing Press Skill v2

## Overview

Two independent post-mortems (Notion run, Linear/lz run) converged on the same diagnosis: the printing-press skill does excellent research (Phases 0-1) then throws it away to chase scorecard numbers (Phases 3-4). It produces discrawl clones for every API instead of purpose-built CLIs that beat the actual competition. It invents cute names nobody asked for. It has no test phase. It has no distribution phase. Emboss triggers confusingly.

This plan synthesizes both post-mortems plus the user's direct feedback into 14 concrete changes to SKILL.md.

## The Three Root Problems

### 1. Scorecard-Driven Development

Both runs exhibit the same pattern:
- Phase 0 discovers what competitors do well (schpet/linear-cli has `start`, 4ier/notion-cli has clean names)
- Phase 4 Priority 2 says "Focus on changes that RAISE THE SCORECARD NUMBER"
- The rest of Phase 4 is spent gaming scorecard string patterns instead of building what users need
- Result: type aliases, dead flags, wrapper functions that exist only to match scorecard regexes

**Evidence:**
- lz run: 6 type aliases (`type staleDB = store.Store`) to match `store.` patterns. Hours going from 69->76 instead of building `lz start ENG-123`
- Notion run: `rateLimitErr` wired to `_ = rateLimitErr` to avoid dead code detection. `printPlain` created just because `flags.plain` was "dead"

### 2. The Discrawl Trap

Every CLI gets the same architecture: SQLite + FTS5 + sync + search + stale + health + trends + patterns + similar + forecast + bottleneck. This is fine for communication APIs (Discord) where the data layer IS the product. It's wrong for project management (Linear) where git integration is the product. It's wrong for content platforms (Notion) where markdown export and clean command names are the product.

The skill's vision dimension benchmarks against discrawl. The anti-shortcut rules say "316 commands is worse than 12 commands." This pushes EVERY run toward the same discrawl-shaped output.

**Evidence:**
- lz has 8 insight commands (stale, health, velocity, trends, patterns, bottleneck, orphans, duplicates) and 0 git commands. schpet/linear-cli has `start` (create git branch from issue) with 524 stars.
- Notion noto has 12 insight/workflow commands and ugly command names. 4ier/notion-cli has clean names and human-friendly filters with 91 stars.

### 3. Names Nobody Asked For

Phase 0.8 says "Do NOT default to `<api>-cli`." This forces creative names: noto (Notion), lz (Linear). These are:
- Confusing ("what the hell is noto?")
- Not discoverable (`brew search notion` won't find `notion-pp-cli`)
- Unnecessary (nobody asked for a creative name)

The generator should produce `<api>-pp-cli` by default (e.g., `notion-pp-cli`, `linear-pp-cli`, `stripe-pp-cli`). The `-pp-` identifies it as a printing press product. If the user wants a different name, they can rename it.

---

## The 14 Changes

### Change 1: Naming - Default to `<api>-pp-cli`

**File:** SKILL.md Phase 0.8 (line ~789)

**Current (broken):**
```markdown
4. **What's the name?** (short, memorable, not confused with the incumbent)
   - Consider: trademark risk, existing tools with that name, domain clarity
   - Test: can you `brew install <name>` without collision?
   - The generator will use this name. Do NOT default to `<api>-cli`.
```

**Proposed:**
```markdown
4. **What's the name?**
   - DEFAULT: `<api>-pp-cli` (e.g., `notion-pp-cli`, `linear-pp-cli`, `stripe-pp-cli`)
   - The `-pp-` identifies it as a printing press product.
   - This is discoverable (`brew search notion` finds `notion-pp-cli`).
   - Creative names are allowed ONLY if the user explicitly requests one.
   - The printing press is a code generator, not a branding agency.
```

Also update the PHASE GATE 0.8 check:
```markdown
3. Name set to `<api>-pp-cli` (default) unless user specified otherwise
```

---

### Change 2: Add Phase 0.6 - Feature Parity Audit (NEW)

**Insert after Phase 0.5 (line ~625).**

This is the missing link between research and build. Currently the skill brainstorms novel features (Phase 0.5) without first checking if we can do what the incumbent already does.

```markdown
# PHASE 0.6: FEATURE PARITY AUDIT

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Before brainstorming novel features, catalog what the competition already ships.

### Step 0.6a: Feature Matrix

For the top 2 competitors by stars (from Phase 0/1 research):

| Feature | Competitor A | Competitor B | Ours | Classification |
|---------|-------------|-------------|------|----------------|

List EVERY command and feature they offer.

### Step 0.6b: Classify Each Feature

- **TABLE STAKES**: >50% of users expect any CLI for this API to have it.
  Examples: `issue create`, `page get`, git branch from issue, clean CRUD names
- **NICE-TO-HAVE**: Useful but not expected. Won't lose users if missing.
  Examples: interactive prompts, TUI mode, plugins
- **ANTI-SCOPE**: Genuinely out of scope with justification.
  Examples: full TUI, mobile app, GUI

**Classification rules:**
- If ANY competitor with >100 stars has it, it's TABLE STAKES unless you
  provide an explicit reason it's anti-scope.
- If users mention it in issues/Reddit with >10 upvotes, it's TABLE STAKES.
- "Complements the incumbent" is NOT a reason to skip a feature. Users
  don't want to install two CLIs.

### Step 0.6c: Table Stakes Become Phase 4 Mandatory Work

Every TABLE STAKES feature becomes a Phase 4 Priority 1 work item.
They are built ALONGSIDE the data layer, not instead of it.

### PHASE GATE 0.6

**STOP.** Verify:
1. Feature matrix complete for top 2 competitors
2. Every feature classified as TABLE STAKES / NICE-TO-HAVE / ANTI-SCOPE
3. TABLE STAKES list has at least 3 items
4. Anti-scope items have explicit cost analysis

Tell the user: "Feature parity audit: [N] table-stakes features identified
from [competitor names]. Top gaps: [list]. These will be built in Phase 4."
```

---

### Change 3: Anti-Scope Requires Cost Analysis

**File:** SKILL.md Phase 0.8 (line ~797)

**Current:**
```markdown
5. **What's the anti-scope?** (what we deliberately do NOT build)
   - Example: "Not a TUI. Not a git replacement. Complements gh, doesn't replace it."
```

**Proposed:**
```markdown
5. **What's the anti-scope and what does it cost?**

   For each anti-scope item, answer:
   - What % of potential users need this feature?
   - Does any competitor with >100 stars offer it?
   - If yes: this is NOT anti-scope, it's a backlog item. Move to Phase 4.

   VALID anti-scope: "Not a TUI" (no competitor offers one either)
   INVALID anti-scope: "Not a git integration" (the top competitor's killer
   feature is git integration - you just excluded the #1 reason people install it)
```

---

### Change 4: Restructure Phase 4 Priorities

**File:** SKILL.md Phase 4 (line ~1226)

**Current order:**
```
Priority 0: Data Layer Foundation
Priority 1: Power User Workflows (from Phase 0.5)
Priority 2: Scorecard-Gap Fixes       <-- THE PROBLEM
Priority 3: Polish
```

**Proposed order:**
```
Priority 0: Data Layer Foundation (unchanged)
Priority 1: Table Stakes Features (NEW - from Phase 0.6 parity audit)
Priority 2: Power User Workflows (from Phase 0.5 - demoted from Priority 1)
Priority 3: Command Name Normalization + Binary Naming (NEW)
Priority 4: Scorecard Gap Fixes (demoted, with anti-gaming rules)
Priority 5: Tests (NEW)
Priority 6: Distribution Scaffold (NEW)
Priority 7: Polish (README cookbook, FAQ)
```

### Priority 1: Table Stakes Features (NEW)

```markdown
### Priority 1: Table Stakes Features (from Phase 0.6)

Build every feature classified as TABLE STAKES in Phase 0.6. These are
features that the top competitor has and that >50% of users expect.

For each table-stakes feature:
1. Read how the competitor implements it (from Phase 1 research)
2. Implement it - don't just match the competitor, make it BETTER
3. Better means: works with --json, supports --dry-run, has --stdin,
   composes with our data layer where possible

**Gate:** Every TABLE STAKES feature from Phase 0.6 must be implemented
before proceeding to Priority 2. No exceptions.
```

### Priority 3: Command Name Normalization + Binary Naming (NEW)

```markdown
### Priority 3: Command Name Normalization + Apply Product Name

**Step 3a: Normalize generated command names**

The generator produces ugly operationId-derived names. Fix them:

| Generated | Normalized | Rule |
|-----------|-----------|------|
| `retrieve-a*` | `get` | Strip "Retrieve a" prefix |
| `delete-a*` | `delete` | Strip "Delete a" prefix |
| `create-a*` | `create` | Strip "Create a" prefix |
| `update-a*` | `update` | Strip "Update a" prefix |
| `post` | `create` | HTTP method -> action |
| `patch` | `update` | HTTP method -> action |
| `get-self` | `me` | Special case |
| `list-*` | `list` | Strip resource suffix |

For each rename:
1. Update the `Use:` field in the command file
2. Rename the file to match
3. Verify `go build` passes

**Step 3b: Apply the product name everywhere**

1. Rename `cmd/<generated-name>/` to `cmd/<product-name>/`
2. Update root.go Use field and version template
3. Update go.mod module path
4. Update client.go User-Agent header
5. `grep -r "<old-name>" . | grep -v "Generated by"` must return 0 hits
6. Update README examples

**Step 3c: Validate API version header**

1. Check what API version the spec uses
2. Check what the generated client sends
3. If they don't match, update client.go
4. If the API uses date-based versions (Notion, Stripe), use the LATEST
```

### Priority 4: Scorecard Gap Fixes (with anti-gaming rule)

```markdown
### Priority 4: Scorecard Gap Fixes (DEMOTED + ANTI-GAMING)

Run the scorecard. Fix real gaps. DO NOT GAME IT.

**ANTI-GAMING RULES:**
- If a function exists only because the scorecard checks for a string
  pattern, DELETE IT.
- If a flag is registered but never checked in any RunE, DELETE IT.
- If an import exists only to put "store." in the file, DELETE IT.
- A CLI that scores 60 but has every table-stakes feature beats one
  that scores 80 with type aliases.
- The scorecard measures proxies for quality. Optimize for actual quality.
```

### Priority 5: Tests (NEW)

```markdown
### Priority 5: Write Tests

A CLI with 0 test files is not shippable.

For each Primary entity in the data layer:
1. Test UpsertX with valid data -> verify row in DB
2. Test UpsertX with missing fields -> verify graceful handling
3. Test SearchX with FTS5 -> verify results match

For each workflow command:
1. Seed DB with test fixtures
2. Run the command's core query
3. Verify result shape and counts

Minimum: 1 test file per package (store, cli).
Use table-driven tests matching Go conventions.
```

### Priority 6: Distribution Scaffold (NEW)

```markdown
### Priority 6: Distribution Scaffold

1. Add `.goreleaser.yaml` for cross-platform binary builds
2. Add Homebrew formula or tap
3. Add install instructions for non-Go users to README
4. Add `.github/workflows/ci.yml` (go test, go vet, goreleaser on tag)

A CLI that can only be installed via `go install` is not a real CLI.
```

---

### Change 5: Fix Emboss Mode UX

**File:** SKILL.md Emboss section (line ~30)

**Current problem:** Emboss can trigger unexpectedly or confusingly. The user said "What the hell is Emboss?" It should be an opt-in follow-up, not something that runs unless requested.

**Proposed changes:**

1. Emboss should NEVER run automatically. It only runs when the user types `emboss` explicitly.
2. After the main run completes (Phase 5 final report), offer emboss as an option:

Add to the end of Phase 5:

```markdown
### Phase 5.9: Offer Emboss

After presenting the final report, ask the user:

"The CLI scored [X]/100 (Grade [Y]). Want me to run an emboss pass
to improve it further? This re-researches the landscape, finds the
top 5 improvements, builds them, and re-scores."

Options:
- "Yes, run emboss" -> proceed to Emboss Mode
- "No, I'm done" -> end the run
- "I'll emboss later" -> tell user they can run `/printing-press emboss ./<api>-cli`

Emboss is a FOLLOW-UP, not an automatic step. The user decides.
```

3. In the Emboss section at the top, add a guard:

```markdown
## Emboss Mode (Second Pass)

**Emboss is opt-in.** It NEVER runs automatically. It runs when:
1. The user explicitly types `/printing-press emboss <dir>`
2. The user selects "Yes, run emboss" from the Phase 5.9 prompt

If the user did NOT request emboss, do NOT mention it, do NOT run it,
do NOT show emboss reports.
```

---

### Change 6: Add Competitor Switch Question (Keep Discrawl as Architecture Reference)

**File:** SKILL.md Phase 4 gate (line ~1331) and anti-shortcut rules

**The printing press has three benchmarks, not one. All three stay:**

1. **gogcli** = "What does quality CLI code look like?" (scoring reference for output modes, auth, error handling, agent-native design). KEEP as the 10/10 reference for code quality dimensions.

2. **discrawl** = "What does a data-layer CLI look like?" (architecture reference for SQLite + FTS5 + sync + domain tables + workflow commands). KEEP as the architecture reference. This is the printing press's moat - most generators produce thin API wrappers. Ours produces discrawl-quality data layers.

3. **Top competitor** = "What do users actually expect?" (feature reference). ADD as the third benchmark. This is what's missing today.

**Current Phase 4 gate:**
```markdown
After Phase 4, ask: "Would a discrawl user switch to this CLI?"
```

**Proposed Phase 4 gate (additive, not replacing):**
```markdown
After Phase 4, ask THREE questions:

1. ARCHITECTURE (discrawl benchmark): "Does this CLI have a real data
   layer - domain-specific SQLite tables, FTS5 search, incremental sync,
   workflow commands that query local data?" If no, Priority 0 isn't done.

2. QUALITY (gogcli benchmark): "Does the code have proper output modes,
   typed errors, agent-native flags, doctor command, README with cookbook?"
   If gaps, Priority 4 scorecard fixes address them.

3. FEATURES (competitor benchmark): "Would a user of [top competitor]
   switch to this CLI?" If no: "What's the ONE feature that would flip
   them?" Build it now before proceeding.

All three must pass. Architecture without features is a toy.
Features without architecture is a thin wrapper. Quality without
either is polished nothing.
```

Update the discrawl-specific anti-shortcut rule to be more nuanced:
- **Current:** `"316 commands is better than 12" (discrawl has 12 commands...)`
- **Proposed:** `"316 commands is better than 12" (Depth beats breadth - discrawl proves this. But depth means building the RIGHT 12 commands, not the same 12 commands for every API. Check the competitor feature matrix.)`

---

### Change 7: Add Competitor Feature Matrix to Phase 3 (Keep gogcli Scoring)

**File:** SKILL.md Phase 3 (line ~1121)

**Keep the existing 10-dimension gogcli scoring table.** It measures real code quality dimensions (output modes, auth, error handling, etc.) that apply to every CLI. Don't touch it.

**Add a SECOND evaluation after the gogcli scoring - the competitor feature matrix:**

```markdown
### Step 3.7b: Head-to-Head Competitor Feature Matrix

The 10-dimension score (above) measures code quality against gogcli.
This step measures feature completeness against the ACTUAL competition.

For the top 2 competitors (from Phase 1):

| Feature | Competitor A | Competitor B | Ours | Status |
|---------|-------------|-------------|------|--------|
| (list every command they have) | Y/N | Y/N | Y/N | HAVE / MISSING / BETTER |

MISSING features with >50% user need become Phase 4 Priority 1 work items.
BETTER features are our differentiators - highlight in README comparison table.

**The gogcli score tells you if the code is good.**
**The competitor matrix tells you if the product is good.**
**You need both.** A CLI that scores 10/10 on output modes but can't
do what the top competitor does is a well-polished toy.
```

---

### Change 8: Phase 5.5 Data Pipeline Smoke Test

**File:** SKILL.md Phase 5.5 (line ~1744)

Add after the live API test sequence:

```markdown
### Step 5.5g: Data Pipeline Smoke Test

After sync, verify data actually flowed through:

1. Query entity counts: pages, blocks, users, etc.
2. **If ANY primary entity has 0 rows:**
   - Verdict: WARN (not PASS)
   - Report the likely cause (permissions, empty workspace, sync bug)
   - Suggest: share a test resource with the integration and re-run
3. **If primary entities have rows, test the read path:**
   - search, stale, health must return non-empty results
4. **If search returns 0 results but rows exist:**
   - FTS5 indexing is broken. Verdict: FAIL.

"0 rows synced" is NEVER a PASS. It's WARN at best, FAIL if the
API token had access.
```

---

### Change 9: Validate Module Path

**File:** SKILL.md Phase 2 (after line ~963)

```markdown
### Step 2.0b: Validate Module Path

Before generation, check the module path:
1. `git config user.name` - this becomes the org
2. If it doesn't match your GitHub username, export the correct one
3. After generation, verify `head -1 <api>-cli/go.mod`
4. If wrong, fix with `go mod edit -module` + find-replace in *.go files
```

---

### Change 10: Validate API Version Header

**File:** SKILL.md Phase 2 (after generation)

```markdown
### Step 2.7: Validate API Version Header

1. Check what version the API docs say to use
2. Check what the generated client sends (`grep "Version" client.go`)
3. If they differ, update client.go to the latest documented version
4. Test against the live API if a token is available
```

---

### Change 11: New Anti-Shortcut Rules

Add these to the existing anti-shortcut section:

```markdown
- "The generated command names are fine" (They're machine names from
  operationIds. Normalize them: retrieve-a -> get, post -> create.)
- "The module path is close enough" (It's a Go import path. It must
  be exact or `go install` fails for everyone.)
- "0 rows synced is still a PASS" (A pipeline that moves no data is
  not tested. It's WARN at best.)
- "Users can go install it" (Most users don't have the Go toolchain.
  Add goreleaser.)
- "I chose the name in Phase 0.8" (Choosing isn't applying. Grep for
  the old name. If it appears, the rename is incomplete.)
- "The scorecard is the objective" (The scorecard measures proxies.
  The objective is: would a user of the top competitor switch?)
- "We complement the incumbent, we don't compete" (Users don't want
  two CLIs. If the incumbent has a feature, you need it too.)
- "That feature is anti-scope" (If a competitor with >100 stars has it,
  it's not anti-scope. It's a backlog item.)
```

---

### Change 12: Remove "Do NOT default to `<api>-cli`" and enforce `<api>-pp-cli`

Grep the skill for any instruction that discourages using the API name:

```
- Phase 0.8 line ~793: "Do NOT default to `<api>-cli`" -> REPLACE with `<api>-pp-cli` default
- Phase 0.5f naming pass: instructions about creative names -> SIMPLIFY
- Any anti-shortcut about naming -> UPDATE to prefer `<api>-pp-cli`
```

The naming pass in Phase 0.5f is fine for WORKFLOW commands (user-friendly
verbs like `stale`, `health`). But the PRODUCT name should be `<api>-pp-cli`.

---

### Change 13: Phase 3 Must Include Competitor Comparison

**File:** SKILL.md Phase 3 (line ~1054)

Add after Step 3.6 (agent-native check):

```markdown
### Step 3.6b: Competitor Command Comparison

Run this comparison for the top 2 competitors:

For each competitor command, check:
- Do we have an equivalent? (Y/N)
- If yes, is ours better? (output, flags, examples)
- If no, why not? (anti-scope with justification, or add to Phase 4 backlog)

This comparison feeds directly into Phase 4 Priority 1.
Any "N" without a strong justification becomes a build task.
```

---

### Change 14: Time Budget Rebalance

**File:** SKILL.md time budget (line ~1900 area)

**Current:**
```
Phase 0-1 (Research): 25%
Phase 2 (Generate): 5%
Phase 3 (Audit): 5%
Phase 4 (GOAT Build): 35%
Phase 4.5 (Dogfood): 10%
Phase 4.6 (Hallucination): 5%
Phase 4.8 (Verify): 10%
Phase 5 (Report): 5%
```

**Proposed:**
```
Phase 0-1 (Research + Parity Audit): 20%
Phase 2 (Generate + Normalize Names): 5%
Phase 3 (Audit + Competitor Comparison): 5%
Phase 4 Priority 0 (Data Layer): 15%
Phase 4 Priority 1 (Table Stakes): 15%        <-- NEW, biggest time allocation
Phase 4 Priority 2 (Workflows): 10%
Phase 4 Priority 3-7 (Names, Score, Tests, Dist): 15%
Phase 4.5-4.8 (Verify): 10%
Phase 5 (Report + Emboss Offer): 5%
```

Table stakes features get 15% - the same as the data layer. Because matching the competition IS the product.

---

## Summary: 14 Changes

| # | Change | Where | Impact |
|---|--------|-------|--------|
| 1 | Default name to `<api>-pp-cli` | Phase 0.8 | Stop inventing confusing names |
| 2 | Add Phase 0.6: Feature Parity Audit | New phase after 0.5 | Prevent shipping without table stakes |
| 3 | Anti-scope requires cost analysis | Phase 0.8 | Stop premature lane narrowing |
| 4 | Restructure Phase 4 (7 priorities) | Phase 4 | Table stakes before scorecard gaming |
| 5 | Fix Emboss UX (opt-in only) | Emboss section + Phase 5.9 | Stop confusing users |
| 6 | Add competitor switch question (keep discrawl+gogcli) | Phase 4 gate | Three benchmarks, not one |
| 7 | Add competitor feature matrix (keep gogcli scoring) | Phase 3 | Ground product in real feature needs |
| 8 | Data pipeline smoke test | Phase 5.5 | 0 rows = WARN, not PASS |
| 9 | Validate module path | Phase 2 | Correct `go install` path |
| 10 | Validate API version header | Phase 2 | Latest API version, not template default |
| 11 | New anti-shortcut rules (8 rules) | Anti-shortcut section | Catch gaming, naming, and testing gaps |
| 12 | Enforce `<api>-pp-cli` naming convention | Phase 0.8, 0.5f | Discoverable, branded, no confusion |
| 13 | Competitor command comparison in Phase 3 | Phase 3 | Feed gaps into Phase 4 Priority 1 |
| 14 | Time budget rebalance | Time budget | Table stakes get 15% (same as data layer) |

## The Meta-Fix

The printing-press skill is optimized for generating impressive-looking CLIs that score well on automated quality checks. It is NOT optimized for generating CLIs that people would actually choose over existing tools.

**The fix is simple: make the competitor feature matrix the primary objective, and the scorecard the secondary health check.** Build what users need first, then polish the code patterns.

The question after every run should not be "what's the scorecard?" It should be: **"Would a user of the top competitor switch to this CLI? If not, what's missing?"**

## Acceptance Criteria

- [ ] Phase 0.8 defaults to `<api>-pp-cli` naming (Change 1, 12)
- [ ] Phase 0.6 Feature Parity Audit exists and is mandatory (Change 2)
- [ ] Anti-scope requires cost analysis with competitor check (Change 3)
- [ ] Phase 4 has 7 priorities with table stakes at Priority 1 (Change 4)
- [ ] Emboss is opt-in only, offered at end of run (Change 5)
- [ ] Phase 4 gate asks "would top competitor's users switch?" (Change 6)
- [ ] Phase 3 includes head-to-head feature matrix (Change 7, 13)
- [ ] Phase 5.5 has data pipeline smoke test (Change 8)
- [ ] Phase 2 validates module path and API version (Change 9, 10)
- [ ] 8 new anti-shortcut rules added (Change 11)
- [ ] Time budget allocates 15% to table stakes (Change 14)
- [ ] Run printing-press on a test API and verify: competitor features built BEFORE scorecard optimization
- [ ] gogcli stays as code quality reference, discrawl stays as architecture reference (Change 6, 7)
- [ ] Run printing-press on Notion again and verify: commands named `get`/`create`/`delete`, binary named `notion-pp-cli`, module path correct

## Sources

- Notion post-mortem: `docs/plans/2026-03-27-fix-printing-press-post-mortem-notion-run-plan.md`
- Linear/lz assessment: `docs/plans/2026-03-27-docs-lz-cli-honest-quality-assessment.md`
- User feedback: "I'm extremely disappointed. The printing-press is so close to being good, but also it's not"
- Current SKILL.md: `~/.claude/skills/printing-press/SKILL.md` (1,915 lines)
