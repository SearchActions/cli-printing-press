# Phase 4.9: Agent Readiness Review Loop

## Problem

The printing-press pipeline generates CLIs that are scored on 17 dimensions including a single "agent-native" dimension (0-10) that checks for surface-level flags (`--json`, `--select`, `--dry-run`, `--stdin`, idempotent). This is a breadth check — it verifies the flags exist but doesn't evaluate whether the CLI is genuinely optimized for autonomous agent use.

The compound-engineering plugin's `compound-engineering:cli-agent-readiness-reviewer` reviewer agent evaluates CLIs against 7 deep principles with severity-based findings (Blocker/Friction/Optimization) and produces framework-idiomatic fix recommendations with file:line references. This depth is missing from the current pipeline.

## Solution

Add a mandatory Phase 4.9 ("Agent Readiness") to the printing-press pipeline that:
1. Invokes the `compound-engineering:cli-agent-readiness-reviewer` reviewer agent on the generated CLI
2. Implements its actionable fixes
3. Iterates until no Blockers or Frictions remain (max 2 passes)

## Where It Fits

```
Phase 4.8 (Runtime Verification)
  → Phase 4.9 (Agent Readiness)    ← NEW
    → Phase 5 (Final Quality Score + Report)
```

Phase 4.9 runs after the CLI is built, audited, and runtime-verified — so the reviewer agent evaluates a functional CLI, not a broken one. Improvements made here are captured by Phase 5's final scorecard re-run, which already compares against the Phase 3 baseline.

**Execution model:** Like all printing-press phases, Phase 4.9 is orchestrated by the LLM agent following SKILL.md instructions. The Go code in `internal/pipeline/` provides state tracking and phase ordering only — it does not programmatically invoke the reviewer agent or parse its output. The `PhaseAgentReadiness` constant in `state.go` tracks completion status; Steps 1-4 below are SKILL.md instructions for the LLM agent to follow.

## Plugin Dependency

The `compound-engineering:cli-agent-readiness-reviewer` reviewer agent lives in the compound-engineering plugin (v2.55.0+), published via the `every-marketplace`.

### Declaration

Create and check in `.claude/settings.json` in this repo:

```json
{
  "enabledPlugins": {
    "compound-engineering@every-marketplace": true
  }
}
```

This is a repo-level config file — anyone who clones the repo gets the plugin dependency declared automatically. This follows the established pattern used by other projects (e.g., `context7@claude-plugins-official`, `hookify@claude-plugins-official`).

The `every-marketplace` must be registered in the user's global `~/.claude/settings.json` under `extraKnownMarketplaces`. If the marketplace is not registered, the plugin won't be found and Phase 4.9 will skip with a warning that includes: "Register every-marketplace in ~/.claude/settings.json to enable agent-readiness review."

### Graceful Degradation

If the reviewer agent is unavailable at runtime (plugin not installed, marketplace not registered, agent not found in cached version):

- Log a warning: "Skipping Phase 4.9 (Agent Readiness) — compound-engineering plugin v2.55.0+ required. Register every-marketplace in ~/.claude/settings.json to enable."
- Skip Phase 4.9 entirely
- Proceed to Phase 5

If the reviewer agent becomes unavailable mid-loop (e.g., between pass 1 and pass 2), log a warning and proceed immediately to Phase 5. The pass count does not reset.

The pipeline remains functional without the plugin. The existing "agent-native" scorecard dimension still runs.

## Phase 4.9 Flow

### Step 1: Invoke Reviewer

Dispatch the `compound-engineering:cli-agent-readiness-reviewer` agent on the generated CLI folder. The invocation must scope the reviewer agent to the generated folder only:

```
Run the compound-engineering:cli-agent-readiness-reviewer agent on the <name> CLI in <output-dir>.
Do not look at code elsewhere in the repo outside of that folder.
```

The reviewer agent produces:
- A scorecard table (7 principles x severity)
- A "What's Working Well" section
- A "Top N Recommended Fixes" list with file:line references and concrete fix descriptions

**Error handling:** If the reviewer agent returns output with no parseable fix list (e.g., malformed output, empty response, or an error during evaluation), treat as a pass-through — log a warning and proceed to Phase 5. Do not loop.

### Step 2: Implement Fixes

Implement all fixes from the reviewer agent's Recommended Fixes list sequentially:

1. Read the fix description and target file:line
2. If the referenced file does not exist or the line is out of bounds: skip this fix, log a warning
3. Make the code change
4. Run `go build ./... && go vet ./...` to verify the change compiles and passes vet
5. If build or vet fails: revert the change, skip this fix, continue to next
6. Move to the next fix

All listed fixes are attempted — the reviewer agent already ranks by impact. The termination check (Step 4) distinguishes severity: Blockers and Frictions trigger another pass; Optimizations alone do not.

### Step 3: Re-run Reviewer

After implementing fixes, re-invoke the reviewer agent on the same folder. This produces a new scorecard and potentially a new fix list.

### Step 4: Termination Check

Evaluate the new scorecard:

- **Zero Blockers and zero Frictions** → Phase 4.9 passes. Proceed to Phase 5.
- **Blockers or Frictions remain, pass count < 2** → Return to Step 2 with the new fix list.
- **Blockers or Frictions remain, pass count = 2** → Log remaining issues as known items. Proceed to Phase 5.

### Flow Diagram

```
┌─────────────────────────┐
│ Check agent availability │
└────────┬────────────────┘
         │ available?
    ┌────┴────┐
    │ no      │ yes
    ▼         ▼
  Skip    ┌──────────────────┐
  with    │ Run reviewer     │◄────────────────┐
  warning │ (pass N)         │                 │
          └────────┬─────────┘                 │
                   ▼                           │
          ┌──────────────────┐                 │
          │ Implement fixes  │                 │
          │ from Top N list  │                 │
          │ (build/vet each) │                 │
          └────────┬─────────┘                 │
                   ▼                           │
          ┌──────────────────┐                 │
          │ Re-run reviewer  │                 │
          └────────┬─────────┘                 │
                   ▼                           │
          ┌──────────────────┐    Blockers/    │
          │ Blockers or      │──Frictions──────┘
          │ Frictions remain?│   & pass < 2
          └────────┬─────────┘
                   │ no, or pass = 2
                   ▼
          ┌──────────────────┐
          │ Proceed to       │
          │ Phase 5          │
          └──────────────────┘
```

## Phase Gate

| Verdict | Condition | Action |
|---------|-----------|--------|
| **Pass** | Zero Blockers and zero Frictions after ≤ 2 passes | Proceed to Phase 5 |
| **Warn** | Frictions remain after 2 passes, zero Blockers | Log Frictions as known issues, proceed to Phase 5 |
| **Degrade** | Blockers remain after 2 passes | Log Blockers as known issues, proceed to Phase 5 (Phase 5 captures impact in final score) |

Note: Phase 4.9 never hard-blocks the pipeline. Even with remaining Blockers (Degrade verdict), the CLI proceeds to Phase 5. The "Degrade" verdict signals that autonomous agent use will have known friction points — it is not a pipeline failure.

## Relationship to Existing Scoring

- The 17-dimension scorecard is unchanged. No new dimensions are added.
- The "agent-native" dimension (0-10) continues to check for flag presence (`--json`, `--select`, `--dry-run`, `--stdin`).
- Phase 4.9's code changes (fixing structured output, non-interactive paths, bounded responses, etc.) naturally improve the "agent-native" dimension score when Phase 5 re-runs the scorecard.
- The reviewer's 7-principle rubric is the evaluation tool; the existing scorecard is the scoring tool. They are complementary, not competing.

## Implementation Scope

### Files to Change

1. **`skills/printing-press/SKILL.md`** — Add Phase 4.9 section between Phase 4.8 and Phase 5. Include:
   - Phase description and gate criteria
   - Agent invocation instructions with folder scoping
   - Fix implementation loop with build/vet verification
   - Termination logic (severity-gated, max 2 passes)
   - Graceful degradation when plugin unavailable

2. **`internal/pipeline/state.go`** — Add `PhaseAgentReadiness = "agent-readiness"` constant. Insert it in `PhaseOrder` between `PhaseReview` and `PhaseComparative`. Bump `currentStateVersion` to 2 so `LoadState` migration adds the new phase to existing state files. Fix migration to set both `Status` and `PlanStatus` to completed for backfilled phases (latent bug in existing migration code that this change would expose).

3. **`internal/pipeline/seeds.go`** — Add a seed template for `PhaseAgentReadiness`. Required because `pipeline.go Init()` iterates all `PhaseOrder` entries and calls `RenderSeed()` for each — missing template causes a crash.

4. **`internal/pipeline/planner.go`** — Add a case for `PhaseAgentReadiness` in `GenerateNextPlan`'s switch statement (or verify the default fallback to `RenderSeed` is sufficient).

5. **`.claude/settings.json`** (new file, checked into repo) — Create with `enabledPlugins` declaring the compound-engineering dependency. This is a repo-level config file so anyone who clones the repo gets the plugin dependency declared.

### Files NOT Changed

- `internal/pipeline/scorecard.go` — No new dimensions or scoring changes
- `internal/pipeline/verify.go` — No new verification layers
- The `compound-engineering:cli-agent-readiness-reviewer` reviewer agent itself — consumed as-is
- `internal/generator/` — No template changes

## What We're Not Doing

- No custom delta display or before/after table for the reviewer's scorecard. Phase 5's comparative scoring already handles before/after against the Phase 3 baseline.
- No forking the `compound-engineering:cli-agent-readiness-reviewer` reviewer agent definition into this repo. Single source of truth stays in compound-engineering.
- No changes to the reviewer agent itself.
- No changes to the existing 17-dimension scorecard structure.
- No opt-in/opt-out flag. Phase 4.9 is mandatory (with graceful skip if plugin unavailable).

## Dependencies

- compound-engineering plugin v2.55.0+ (contains `compound-engineering:cli-agent-readiness-reviewer` reviewer agent)
- Published via `every-marketplace`
- User must have `every-marketplace` registered in global `~/.claude/settings.json`

## Risks

| Risk | Mitigation |
|------|------------|
| Reviewer agent output format changes in future plugin versions | We consume the fix list as natural language, not parsed structure. Format changes are tolerable as long as the reviewer agent produces actionable fix descriptions. If output is unparseable, we skip the loop and proceed to Phase 5. |
| Fix implementation breaks the CLI | Each fix is followed by `go build && go vet`. Failed builds are reverted. |
| Two passes aren't enough for complex CLIs | Remaining issues are logged. The pipeline doesn't block. Users can run emboss mode for further improvement. |
| Plugin not installed | Graceful skip with warning message. Pipeline works standalone. |
| Token cost (~92k tokens per reviewer pass, up to 2 passes) | Accepted tradeoff. Agent-readiness is core to printing-press's value proposition. |
