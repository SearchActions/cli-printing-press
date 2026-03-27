---
title: "Printing Press Quality Overhaul: Runtime Verification + Research Improvements"
type: feat
status: completed
date: 2026-03-27
---

# Printing Press Quality Overhaul

## Overview

The printing press generates CLIs that compile but don't work. We proved this by running it on GitHub's API - it produced a 73/100 scorecard where the core feature (sync -> SQLite -> search) 404s on first call, and only 3.9% of commands were tested at runtime. The scorecard measures file contents, not behavior. The research phases miss the competitive landscape. The skill allows skipping mandatory phases.

This plan fixes all of it in two tracks:

**Track A: Runtime Verification** - Make the press prove every CLI works before declaring success. `printing-press verify` builds the binary, tests every command against the real API (read-only) or a mock server, and auto-fixes failures in a loop until 80%+ pass.

**Track B: Research Quality** - Make the press produce better plans by searching the market landscape (not just API wrappers), requiring a product thesis before code generation, consuming prior research, and naming commands after user outcomes.

## Evidence From the GitHub CLI Run

These are the specific failures that motivate each fix. The github-cli is being deleted, but the learnings stay.

| What happened | Root cause | Fix |
|--------------|-----------|-----|
| Sync command 404s | Generic sync hits `/repos` without owner/repo | Track A: verify catches this at runtime |
| Scorecard 73/100 with broken core | Scorecard checks strings not behavior | Track A: verify tests actual execution |
| Missed lazygit (75k stars), Jujutsu (27.4k), Graphite | Phase 0 searches for API wrappers only | Track B: market landscape searches |
| No product pitch, no name, no HN headline | No product thesis phase | Track B: add Phase 0.8 |
| Commands named `actions-health` not `bottleneck` | Names from API resources, not user outcomes | Track B: naming pass in Phase 0.5 |
| Skipped Phase 4.5 and 4.6 | Skill says mandatory but LLM rationalizes | Track A: verify is a Go binary gate, not a skill instruction |
| Tested 5/127 commands, declared PASS | No runtime test suite | Track A: verify tests every command |
| Data pipeline (sync->sql->search) never verified | No end-to-end pipeline test | Track A: verify has dedicated data pipeline test |

## Track A: Runtime Verification

### What's Already Built

`printing-press verify` exists and works. Tested against the github-cli:

```
$ printing-press verify --dir ./github-cli --spec /tmp/spec.json --api-key $GITHUB_TOKEN
Mode: live
Pass Rate: 67% (16/24 passed, 2 critical)
Data Pipeline: FAIL
Verdict: WARN
```

**Files already in the codebase:**
- `internal/pipeline/runtime.go` - TestBackend, mock server, command runner, scoring
- `internal/cli/verify.go` - CLI subcommand with --dir, --spec, --api-key, --env-var, --threshold, --json
- `internal/generator/templates/config.go.tmpl` - Added `<API>_BASE_URL` env var override

### What Remains to Build

#### A1. Fix Loop (`internal/pipeline/fixloop.go`)

When `--fix` flag is set and score < threshold, auto-patch failures:

```
Iteration 1: Classify failures -> Generate patches -> Apply -> Rebuild -> Re-test
Iteration 2: Same for remaining failures
Iteration 3: Same (max 3 iterations, then report what's still broken)
```

**Failure classification and auto-fix map:**

| Failure pattern | Root cause | Patch |
|----------------|-----------|-------|
| `--help` exits non-zero | Command not registered in root.go | Add `rootCmd.AddCommand(newXCmd(&flags))` |
| `--dry-run` fails | Command doesn't check `flags.dryRun` | Add dryRun check in RunE |
| Execute returns non-JSON | `printOutputWithFlags` not called | Wire output through `printOutputWithFlags` |
| Mock received no request | Wrong base URL | Fix config env var reading |
| Sync writes 0 rows | Upsert not wired | Wire sync resource loop to call domain Upsert |
| SQL "no such table" | Migration missing from store.go | Add CREATE TABLE |
| Search returns 0 | FTS5 triggers missing | Add FTS5 insert/update/delete triggers |
| Command naming mismatch | `camelToKebab` produces wrong name | Fix the discovery regex |

Each fix is surgical - read the specific file, find the specific line, generate a targeted patch. Not a rewrite. After each patch: `go build && go vet`. After all patches in an iteration: re-run full verify. If net score decreased, revert that iteration.

**Acceptance criteria for A1:**
- [ ] `printing-press verify --dir ./X --fix` patches at least 3 failure types
- [ ] Fix loop compiles after each patch (`go build && go vet`)
- [ ] Re-runs full verify after each iteration, reports before/after delta
- [ ] Max 3 iterations, stops if no improvement between iterations
- [ ] Net score never decreases (revert iteration if it does)

#### A2. Wire Verify Into Full Pipeline (`internal/pipeline/fullrun.go`)

The current pipeline: `generate -> gates(7) -> dogfood -> verify_static -> scorecard`

New pipeline: `generate -> gates(7) -> dogfood -> verify_static -> scorecard -> VERIFY_RUNTIME -> [fix loop] -> scorecard (re-run)`

The runtime verify runs AFTER the scorecard. After the fix loop, re-run scorecard. Report both numbers: scorecard before and after, verify pass rate before and after.

**Acceptance criteria for A2:**
- [ ] `printing-press print <API>` runs verify as final gate
- [ ] If API key was provided in Phase 0.1, live mode is used
- [ ] Pipeline reports: scorecard (before fix), verify pass rate, fix loop iterations, scorecard (after fix)
- [ ] Pipeline FAIL verdict if verify < 80% after fix loop

#### A3. Improve Command Discovery and Classification

The current verify discovered 24 commands from the github-cli but:
- `p-r-triage` instead of `pr-triage` (camelToKebab splits wrong)
- `s-q-l` instead of `sql` (same issue)
- Workflow commands (stale, actions-health) classified as "read" but need --org/--repo args to work

Fixes:
- [ ] Handle acronyms in camelToKebab (PR, SQL, API should not be split)
- [ ] Parse command files for required args (positional args from `cobra.ExactArgs` or `cobra.MinimumNArgs`)
- [ ] For workflow commands needing --org/--repo, supply test values from the spec (e.g., `--repo mock-owner/mock-repo`)
- [ ] For commands with subcommands, recurse into subcommand tree (currently only tests top-level)

#### A4. Improve Mock Server Fidelity

The current mock returns hardcoded JSON. Better:
- [ ] Parse the spec's response schemas to generate field-accurate responses
- [ ] Handle pagination (return `Link: <url>; rel="next"` header for list endpoints)
- [ ] Handle nested wrappers (GitHub's `{"workflow_runs": [...], "total_count": 1}`)
- [ ] Return appropriate error codes for missing required params (400 instead of 200)

#### A5. Read-Only Safety Hardening

The runner enforces read-only at the code level, but add belt-and-suspenders:
- [ ] Log every command to stderr before executing (already in plan)
- [ ] Track rate limit consumption via X-RateLimit-Remaining headers
- [ ] Auto-pause if remaining < 50 (don't burn the user's rate limit during testing)
- [ ] Add `--dry-run-only` flag that skips execute tests entirely (for extra caution)

### Track A Implementation Order

```
A3 (fix discovery) -> A1 (fix loop) -> A4 (mock fidelity) -> A2 (wire pipeline) -> A5 (safety)
```

A3 first because the fix loop needs correct command names to work. A2 last because it needs everything else working.

---

## Track B: Research Quality Improvements

### B1. Market Landscape Searches in Phase 0

**The problem:** Phase 0 searches for `"GitHub API" CLI tool` and finds API wrappers. It never searches for `"GitHub CLI" alternative` (which finds lazygit 75k, jj 27.4k, Graphite) or `git TUI` or `stacked PRs tool`.

**The fix:** Add 4 market landscape search queries to Phase 0, Step 0c:

```
# Existing (finds API wrappers):
WebSearch: "<API name>" CLI tool github

# NEW (finds the real competitive landscape):
WebSearch: "<API name>" CLI alternative OR replacement
WebSearch: "<API domain>" TUI OR terminal tool 2026
WebSearch: best "<API domain>" workflow tool
WebSearch: "<API name>" "I switched to" OR "better than"
```

Where `<API domain>` is the domain category from Step 0a (e.g., "git" for GitHub, "payments" for Stripe).

**Where to implement:** Update `skills/printing-press/SKILL.md` Phase 0 Step 0c to add these searches alongside the existing ones.

**Acceptance criteria for B1:**
- [ ] SKILL.md Phase 0c includes 4 market landscape searches
- [ ] Searches use the API's domain category, not just the API name
- [ ] The Three-Lane Framework concept is referenced (Forge CLIs / Workflow Overlays / Domain UX)

### B2. Product Thesis Phase (Phase 0.8)

**The problem:** The press jumps from technical research (Phase 0.7 data layer) to code generation (Phase 2) without ever articulating who the CLI is for or why anyone would use it.

**The fix:** Add Phase 0.8 between Phase 0.7 and Phase 1:

```markdown
# PHASE 0.8: PRODUCT THESIS

## THIS PHASE IS MANDATORY. DO NOT SKIP IT.

Before generating code, answer these five questions:

1. **Who is this for?** (one sentence, specific persona)
   Example: "Engineering managers who need cross-repo PR triage without a dashboard"

2. **What's the comparison table?** (us vs incumbent, 5 rows)
   | Capability | Incumbent | Ours |
   |-----------|-----------|------|

3. **What's the HN headline?** (one sentence that makes a developer click)
   Example: "I built a GitHub CLI that finds stale PRs and lets you SQL query your repos offline"

4. **What's the name?** (short, memorable, not confused with the incumbent)
   Consider: trademark, existing tools, domain clarity

5. **What's the anti-scope?** (what we deliberately do NOT build)
   Example: "Not a TUI. Not a git replacement. Complements gh, doesn't replace it."

### PHASE GATE 0.8

Write a 1-paragraph product thesis. If you can't articulate why someone would install this
in one paragraph, the research phases missed something. Go back.
```

**Where to implement:** Update `skills/printing-press/SKILL.md` to add Phase 0.8.

**Acceptance criteria for B2:**
- [ ] SKILL.md has Phase 0.8 between Phase 0.7 and Phase 1
- [ ] Phase 0.8 requires answers to all 5 questions
- [ ] Phase gate requires a written product thesis paragraph
- [ ] The name chosen in Phase 0.8 is used for the CLI (not `<api>-cli`)

### B3. Consume Prior Research

**The problem:** If the user already researched the API in a separate session, the press starts from scratch. The pre-research competitive landscape plan was better than the press's own Phase 0.

**The fix:** At the start of Phase 0, check for existing research:

```markdown
### Step 0.0: Check for Prior Research

Before starting fresh research, check if the user has already done research:

\`\`\`bash
ls ~/cli-printing-press/docs/plans/*<api-name>* ~/docs/plans/*<api-name>* 2>/dev/null
\`\`\`

If found, read those documents. Extract:
- Competitive landscape and tool rankings
- User pain points with quotes
- Product positioning decisions
- Workflow ideas and command names

Use this as the foundation. Skip redundant research steps. Focus Phase 0 on
filling gaps the prior research didn't cover (data profile, entity classification).
```

**Acceptance criteria for B3:**
- [ ] SKILL.md Phase 0 starts with a check for existing plans matching the API name
- [ ] Found plans are read and their insights are carried forward
- [ ] Redundant searches are skipped when prior research covers them

### B4. Outcome-Based Command Naming Pass

**The problem:** The press names commands after API resources (`actions-health`, `contributors`, `activity`). Good products name commands after user outcomes (`bottleneck`, `review-load`, `standup`).

**The fix:** Add a naming pass at the end of Phase 0.5:

```markdown
### Step 0.5f: Naming Pass

For each workflow command, ask: "If I were an engineering manager, what would I type?"

Map API-oriented names to outcome-oriented names:
- "actions-health" -> "ci-health" or "flaky" (what the user cares about)
- "contributors" -> "leaderboard" or "who-shipped" (the question being answered)
- "activity" -> "standup" (the workflow it serves)

The name should complete this sentence: "I need to check ___"
- "I need to check stale" (good)
- "I need to check actions-health" (bad - nobody talks like that)
```

**Acceptance criteria for B4:**
- [ ] SKILL.md Phase 0.5 includes a naming pass after workflow scoring
- [ ] Each workflow name is tested against the "I need to check ___" sentence
- [ ] The pass produces a name mapping table (API name -> user name)

### B5. Anti-Shortcut Rules for Skipping

**The problem:** The skill says phases are mandatory but the LLM rationalizes skipping them. I skipped Phase 4.5 and 4.6 despite them being labeled "MANDATORY. DO NOT SKIP."

**The fix:** The verify command is the real gate (you can't rationalize past a binary that exits non-zero). But also add explicit anti-shortcut rules:

```markdown
- "I'll skip the dogfood/verify to save time" (Skipping testing is how you produce a CLI
  that scores 73/100 with a broken core feature. The GitHub run proved this. Run verify.)
- "The scorecard is 73 so it's good enough" (The scorecard measures files, not behavior.
  A 73 scorecard with 0% verify pass rate is a CLI that looks good on paper and crashes
  on first use. Run verify.)
- "I tested 5 commands, that's enough" (5/127 is 3.9%. That's not testing. Run verify
  which tests every command automatically.)
```

**Acceptance criteria for B5:**
- [ ] SKILL.md anti-shortcut section includes 3 new rules about testing
- [ ] Phase 4.8 (Runtime Verification) is added as mandatory with PHASE GATE

---

## Implementation Order

Build Track A first (the press can't guarantee quality without verify). Track B improves the research but the press still works without it.

```
Track A (runtime verification):
  A3: Fix command discovery      (runtime.go)
  A1: Build fix loop             (fixloop.go - NEW file)
  A4: Improve mock server        (runtime.go)
  A2: Wire into fullrun pipeline (fullrun.go)
  A5: Read-only safety           (runtime.go)

Track B (research quality):
  B1-B5: All SKILL.md changes    (skills/printing-press/SKILL.md)
```

Estimated: Track A is 4-6 files of Go code. Track B is one file (SKILL.md) with ~100 lines of additions.

## Acceptance Criteria

### The Bar

After this work, running `printing-press print <API>` (or the manual skill phases) should:

1. **Research** finds the full competitive landscape, not just API wrappers
2. **Product thesis** is articulated before code generation
3. **Commands** are named after user outcomes
4. **Every command** is tested at runtime (not just compiled)
5. **Data pipeline** is verified end-to-end (sync -> sql -> search)
6. **Fix loop** auto-patches common failures
7. **Final verdict** is based on runtime pass rate, not scorecard string matching
8. **Read-only guarantee** is enforced in code for live API testing

### Proof It Works

Generate a CLI for the Stripe API (or any API with a public spec + free API key). The verify command should:
- Test 100% of commands at runtime
- Catch any broken sync/search/sql pipeline
- Auto-fix at least 3 issues via fix loop
- Produce a PASS verdict (>= 80% pass rate) or clearly report what's still broken

## Files Changed

| File | Change |
|------|--------|
| `internal/pipeline/runtime.go` | Fix camelToKebab, improve discovery, mock fidelity, rate limit tracking |
| `internal/pipeline/fixloop.go` | **NEW** - Auto-fix loop with failure classification |
| `internal/pipeline/fullrun.go` | Wire verify as final gate after scorecard |
| `internal/cli/verify.go` | Add --fix and --dry-run-only flags |
| `internal/generator/templates/config.go.tmpl` | Already done - BASE_URL env var |
| `skills/printing-press/SKILL.md` | Phase 0.8 (product thesis), market searches, naming pass, anti-shortcut rules, Phase 4.8 (verify) |

## Sources

### From This Session
- Post-mortem: `docs/plans/2026-03-27-fix-printing-press-post-generation-testing-gaps-plan.md`
- Runtime verify plan: `docs/plans/2026-03-27-feat-printing-press-runtime-verification-loop-plan.md`
- Research comparison: `docs/plans/2026-03-27-research-printing-press-plan-quality-comparison-plan.md`
- Verify output: 67% pass rate on github-cli, 2 critical failures, data pipeline FAIL

### From the Pre-Research Session
- Competitive landscape: `~/docs/plans/2026-03-27-research-github-cli-competitive-landscape-plan.md`
- Build plan: `~/docs/plans/2026-03-27-feat-printing-press-github-cli-plan.md`

### Codebase
- `internal/pipeline/scorecard.go` (1,488 lines) - string-matching quality scoring
- `internal/pipeline/dogfood.go` (522 lines) - static analysis validation
- `internal/pipeline/verify.go` (541 lines) - static proof system
- `internal/pipeline/fullrun.go` (584 lines) - pipeline orchestrator
- `internal/generator/validate.go` (123 lines) - 7 compilation quality gates
