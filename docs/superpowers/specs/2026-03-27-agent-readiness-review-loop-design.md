# Phase 4.9: Agent Readiness Review Loop

## Problem

The printing-press pipeline generates CLIs that are scored on 17 dimensions including a single "agent-native" dimension (0-10) that checks for surface-level flags (`--json`, `--select`, `--dry-run`, `--stdin`, idempotent). This is a breadth check — it verifies the flags exist but doesn't evaluate whether the CLI is genuinely optimized for autonomous agent use.

The compound-engineering plugin's `cli-agent-readiness-reviewer` agent evaluates CLIs against 7 deep principles with severity-based findings (Blocker/Friction/Optimization) and produces framework-idiomatic fix recommendations with file:line references. This depth is missing from the current pipeline.

## Solution

Add a mandatory Phase 4.9 ("Agent Readiness") to the printing-press pipeline that:
1. Invokes the external reviewer agent on the generated CLI
2. Implements its actionable fixes
3. Iterates until no Blockers or Frictions remain (max 2 passes)

## Where It Fits

```
Phase 4.8 (Runtime Verification)
  → Phase 4.9 (Agent Readiness)    ← NEW
    → Phase 5 (Comparative)
```

Phase 4.9 runs after the CLI is built, audited, and runtime-verified — so the reviewer evaluates a functional CLI, not a broken one. Improvements made here are captured by Phase 5's final scorecard re-run, which already compares against the Phase 3 baseline.

## Plugin Dependency

The `cli-agent-readiness-reviewer` agent lives in the compound-engineering plugin (v2.55.0+), published via the `every-marketplace`.

### Declaration

Create `.claude/settings.json` in this repo:

```json
{
  "enabledPlugins": {
    "compound-engineering@every-marketplace": true
  }
}
```

This follows the established pattern used by other projects (e.g., `context7@claude-plugins-official`, `hookify@claude-plugins-official`). The marketplace must be registered in the user's global `~/.claude/settings.json` under `extraKnownMarketplaces`, which it already is for users who have compound-engineering installed.

### Graceful Degradation

If the agent is unavailable at runtime (plugin not installed, marketplace not registered, agent not found in cached version):

- Log a warning: "Skipping Phase 4.9 (Agent Readiness) — compound-engineering plugin v2.55.0+ required. Install via every-marketplace."
- Skip Phase 4.9 entirely
- Proceed to Phase 5 (Comparative)

The pipeline remains functional without the plugin. The existing "agent-native" scorecard dimension still runs.

## Phase 4.9 Flow

### Step 1: Invoke Reviewer

Dispatch the `compound-engineering:review:cli-agent-readiness-reviewer` agent on the generated CLI folder. The invocation must scope the reviewer to the generated folder only:

```
Run the cli-agent-readiness-reviewer agent on the <name> CLI in <output-dir>.
Do not look at code elsewhere in the repo outside of that folder.
```

The reviewer produces:
- A scorecard table (7 principles x severity)
- A "What's Working Well" section
- A "Top N Recommended Fixes" list with file:line references and concrete fix descriptions

### Step 2: Implement Fixes

Take the Top N Recommended Fixes list and implement each fix sequentially:

1. Read the fix description and target file:line
2. Make the code change
3. Run `go build ./... && go vet ./...` to verify the change compiles
4. If build fails: revert the change, skip this fix, continue to next
5. Move to the next fix

All Blocker and Friction fixes are attempted. Optimization fixes are also attempted from the Top N list — the reviewer already ranks by impact, so all listed fixes are worth implementing. Only the termination check distinguishes severity (Blockers/Frictions trigger another pass; Optimizations alone do not).

### Step 3: Re-run Reviewer

After implementing fixes, re-invoke the reviewer on the same folder. This produces a new scorecard and potentially a new fix list.

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
| **Pass** | Zero Blockers after ≤ 2 passes | Proceed to Phase 5 |
| **Warn** | Frictions remain after 2 passes | Log as known issues, proceed |
| **Fail** | Blockers remain after 2 passes | Log as known issues, proceed (Phase 5 captures impact in final score) |

Note: Phase 4.9 never hard-blocks the pipeline. Even with remaining Blockers, the CLI proceeds to Phase 5. The issues are documented and the final scorecard reflects the state.

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

2. **`internal/pipeline/state.go`** — Add `PhaseAgentReadiness = "agent-readiness"` constant. Insert it in the phase ordering between `PhaseReview` and `PhaseComparative`.

3. **`.claude/settings.json`** (new file) — Create with `enabledPlugins` declaring the compound-engineering dependency.

### Files NOT Changed

- `internal/pipeline/scorecard.go` — No new dimensions or scoring changes
- `internal/pipeline/verify.go` — No new verification layers
- The compound-engineering plugin itself — consumed as-is
- `internal/generator/` — No template changes

## What We're Not Doing

- No custom delta display or before/after table for the reviewer's scorecard. Phase 5's comparative scoring already handles before/after against the Phase 3 baseline.
- No forking the `cli-agent-readiness-reviewer` agent definition into this repo. Single source of truth stays in compound-engineering.
- No changes to the reviewer agent itself.
- No changes to the existing 17-dimension scorecard structure.
- No opt-in/opt-out flag. Phase 4.9 is mandatory (with graceful skip if plugin unavailable).

## Dependencies

- compound-engineering plugin v2.55.0+ (contains `cli-agent-readiness-reviewer` agent)
- Published via `every-marketplace`
- User must have `every-marketplace` registered in global `~/.claude/settings.json`

## Risks

| Risk | Mitigation |
|------|------------|
| Reviewer agent output format changes in future plugin versions | We consume the Top N fix list as natural language, not parsed structure. Format changes are tolerable as long as the reviewer produces actionable fix descriptions. |
| Fix implementation breaks the CLI | Each fix is followed by `go build && go vet`. Failed builds are reverted. |
| Two passes aren't enough for complex CLIs | Remaining issues are logged. The pipeline doesn't block. Users can run emboss mode for further improvement. |
| Plugin not installed | Graceful skip with warning message. Pipeline works standalone. |
| Token cost (~92k tokens per reviewer pass, up to 2 passes) | Accepted tradeoff. Agent-readiness is core to printing-press's value proposition. |
