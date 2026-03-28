---
title: "feat: Add Phase 4.9 Agent Readiness Review Loop"
type: feat
status: active
date: 2026-03-27
origin: docs/brainstorms/2026-03-27-agent-readiness-review-loop-requirements.md
---

# feat: Add Phase 4.9 Agent Readiness Review Loop

## Overview

Add a new pipeline phase (4.9) that invokes the compound-engineering plugin's `compound-engineering:cli-agent-readiness-reviewer` agent on generated CLIs, implements its fixes in a severity-gated loop (max 2 passes), and tracks state through the existing pipeline machinery. This deepens agent-readiness from a surface-level flag check (existing "agent-native" 0-10 dimension) to a 7-principle evaluation with framework-idiomatic fixes.

## Problem Frame

The printing-press pipeline scores CLIs on 17 dimensions, but its "agent-native" dimension only checks for flag presence (`--json`, `--select`, `--dry-run`, `--stdin`). The compound-engineering plugin's reviewer agent evaluates 7 deep principles (non-interactive automation, structured output, progressive help, actionable errors, safe retries, composability, bounded responses) with Blocker/Friction/Optimization severity levels and file:line fix recommendations. Integrating this reviewer as a pipeline phase closes the depth gap. (see origin: docs/brainstorms/2026-03-27-agent-readiness-review-loop-requirements.md)

## Requirements Trace

- R1. Phase 4.9 runs after Phase 4.8 (Runtime Verification) and before Phase 5 (Final Quality Score)
- R2. Invokes `compound-engineering:cli-agent-readiness-reviewer` scoped to the generated CLI folder
- R3. Implements all fixes from the reviewer's fix list, with `go build && go vet` after each
- R4. Iterates until zero Blockers and zero Frictions remain (max 2 passes)
- R5. Gracefully skips if the compound-engineering plugin is unavailable
- R6. Declares compound-engineering as a plugin dependency via `.claude/settings.json`
- R7. Existing 17-dimension scorecard is unchanged; Phase 5 captures improvements naturally

## Scope Boundaries

- No new scorecard dimensions or scoring changes
- No forking the reviewer agent definition into this repo
- No changes to the reviewer agent itself
- No opt-in/opt-out flag — Phase 4.9 runs by default, skips gracefully if plugin unavailable
- No custom before/after display for the reviewer's scorecard — Phase 5 handles this

## Context & Research

### Relevant Code and Patterns

- `internal/pipeline/state.go` — Phase constants (`PhasePreflight` through `PhaseShip`), `PhaseOrder` slice, `currentStateVersion = 1`, `LoadState()` migration, `NextPhase()` uses `PlanStatus != PlanStatusCompleted`
- `internal/pipeline/seeds.go` — `seedTemplates` map keyed by phase constant, `RenderSeed()` with `text/template`, `SeedData` struct. Test at `seeds_test.go:19` iterates `PhaseOrder` and validates all seeds
- `internal/pipeline/planner.go` — `GenerateNextPlan()` switch dispatches per-phase; `default` falls back to `RenderSeed()`. Comment: "Preflight, Research use static seeds"
- `internal/pipeline/pipeline.go` — `Init()` at lines 76-89 iterates ALL `PhaseOrder` entries and calls `RenderSeed()` for each
- `skills/printing-press/SKILL.md` — Phase 4.8 (lines 1595-1648) is the structural analog: mandatory header, step labels (4.8a, 4.8b), gate table with PASS/WARN/FAIL, "Tell the user:" message
- SKILL.md frontmatter (lines 1-16): version 1.1.0, allowed-tools includes `Agent`
- SKILL.md phase map (line 233): flow diagram with time budgets
- SKILL.md artifact list (lines 246-254): 7 plan artifacts per run

### Institutional Learnings

- **Skill-within-skill invocation is fragile** — Past fix (`2026-03-25-fix-printing-press-skill-loop-enforcement-plan.md`) documents that Claude treats `Skill()` calls as suggestions and skips phases. Phase 4.9 must use "THIS PHASE IS MANDATORY. DO NOT SKIP IT." header with the single exception path (plugin unavailable) stated explicitly
- **PlanPath uses index-based derivation** — `fmt.Sprintf("%02d-%s-plan.md", i, name)` means inserting a phase shifts all subsequent filenames. Changing to name-based derivation avoids migration complexity
- **Migration bug** — `LoadState()` migration sets `Status: StatusCompleted` but not `PlanStatus: PlanStatusCompleted` for backfilled phases. `NextPhase()` checks `PlanStatus`, so backfilled phases appear pending. Must fix when bumping to version 2

## Key Technical Decisions

- **PlanPath derivation: switch from index-based to name-based** — Current `%02d-%s-plan.md` format means every phase insertion shifts subsequent filenames and requires migration to rename existing files. Changing to `%s-plan.md` (phase name only) eliminates this coupling. The index prefix provides visual ordering in directory listings but is not used programmatically — sorting can rely on `PhaseOrder` instead. This is a minor but load-bearing change that prevents future phase insertions from needing file renames.

- **Planner: use default fallback, no explicit case** — `GenerateNextPlan()` default branch calls `RenderSeed()`. Since Phase 4.9 is entirely LLM-orchestrated (the SKILL.md instructions drive execution, not Go code), the static seed is sufficient. No dynamic plan enrichment from scorecard or dogfood results is needed. If future iterations want to inject prior phase results into the agent-readiness plan, an explicit case can be added later.

- **State version bump to 2 with PlanStatus fix** — The migration must set both `Status: StatusCompleted` and `PlanStatus: PlanStatusCompleted` for backfilled phases. This fixes the latent bug where `NextPhase()` treats backfilled phases as pending. Bumping `currentStateVersion` to 2 triggers the migration for existing state files.

## Open Questions

### Resolved During Planning

- **Should we add a planner.go case?** No — the default fallback to `RenderSeed()` is sufficient. Phase 4.9 is LLM-orchestrated; the Go code only tracks state. (see origin requirement: "The Go code provides state tracking and phase ordering only")
- **Should PlanPath use name-based or index-based naming?** Name-based. Avoids migration file renames and future insertion issues. Sorting in directory listings is a minor cosmetic loss offset by elimination of a maintenance hazard.

### Deferred to Implementation

- **Exact seed template prose** — The required sections (Phase Goal, Context, What This Phase Must Produce, Prior Phase Outputs, Codebase Pointers) are known from the seed template pattern, but the specific content for the agent-readiness phase will be written during implementation
- **SKILL.md time budget for Phase 4.9** — Depends on observed reviewer agent runtime (~3.5 min per pass). Will estimate as 5-10 min in the phase map

## Implementation Units

- [ ] **Unit 1: Create `.claude/settings.json`**

**Goal:** Declare the compound-engineering plugin dependency so anyone cloning the repo gets it automatically

**Requirements:** R6

**Dependencies:** None

**Files:**
- Create: `.claude/settings.json`

**Approach:**
- Create `.claude/` directory and `settings.json` with `enabledPlugins` entry for `compound-engineering@every-marketplace`
- This follows the pattern used by other projects (`context7@claude-plugins-official`, `hookify@claude-plugins-official`)

**Patterns to follow:**
- Claude Code plugin settings format (JSON with `enabledPlugins` key)

**Test scenarios:**
- Happy path: File exists with valid JSON after creation, `enabledPlugins` contains `compound-engineering@every-marketplace: true`
- Edge case: Verify `.claude/settings.json` does not conflict with `.claude-plugin/` directory (different purpose — `.claude-plugin/` is plugin definition, `.claude/settings.json` is plugin dependency declaration)

**Verification:**
- `.claude/settings.json` exists and contains valid JSON with the `enabledPlugins` entry
- `.claude-plugin/` directory is unaffected

---

- [ ] **Unit 2: Add phase constant and fix PlanPath derivation in state.go**

**Goal:** Add `PhaseAgentReadiness` to the pipeline, fix index-based PlanPath to name-based, bump state version with migration fix

**Requirements:** R1, R5

**Dependencies:** None (can be done in parallel with Unit 1)

**Files:**
- Modify: `internal/pipeline/state.go`
- Test: `internal/pipeline/state_test.go`

**Approach:**
- Add `PhaseAgentReadiness = "agent-readiness"` constant
- Insert into `PhaseOrder` between `PhaseReview` and `PhaseComparative`: `[...PhaseReview, PhaseAgentReadiness, PhaseComparative, PhaseShip]`
- Change PlanPath derivation in `NewState()` from `fmt.Sprintf("%02d-%s-plan.md", i, name)` to `fmt.Sprintf("%s-plan.md", name)` — both in `NewState()` (line 87) and `LoadState()` migration (line 131)
- Bump `currentStateVersion` from 1 to 2
- Fix migration block to set both `Status: StatusCompleted` AND `PlanStatus: PlanStatusCompleted` for backfilled phases (matching `Complete()` behavior)
- Migration must also update `PlanPath` for all existing phases from `%02d-%s-plan.md` to `%s-plan.md` format

**Patterns to follow:**
- Existing phase constant declarations at top of state.go
- `Complete()` function (sets both `Status` and `PlanStatus`)
- Migration block structure in `LoadState()`

**Test scenarios:**
- Happy path: `NewState()` creates state with `PhaseAgentReadiness` at correct position in PhaseOrder, PlanPath uses name-based format (`agent-readiness-plan.md`)
- Happy path: `NextPhase()` returns `PhaseAgentReadiness` after `PhaseReview` is completed
- Edge case: `LoadState()` with version 1 state file migrates to version 2 — adds `PhaseAgentReadiness` with both `Status: StatusCompleted` and `PlanStatus: PlanStatusCompleted`
- Edge case: Migration updates PlanPath from old `%02d-name-plan.md` to `name-plan.md` for all phases
- Edge case: Version 2 state file is loaded without migration running (idempotent)
- Error path: State file with unknown phase name is preserved (not dropped)

**Verification:**
- `go test ./internal/pipeline/...` passes
- `PhaseAgentReadiness` appears in `PhaseOrder` between `PhaseReview` and `PhaseComparative`
- PlanPath for all phases uses name-based format

---

- [ ] **Unit 3: Add seed template in seeds.go**

**Goal:** Provide the seed template for `PhaseAgentReadiness` so `Init()` and tests don't crash

**Requirements:** R1

**Dependencies:** Unit 2 (PhaseAgentReadiness constant must exist)

**Files:**
- Modify: `internal/pipeline/seeds.go`
- Test: `internal/pipeline/seeds_test.go`

**Approach:**
- Add entry to `seedTemplates` map keyed by `PhaseAgentReadiness`
- Follow the thin-seed format validated by `TestRenderSeedProducesThinSeedTemplates`: YAML frontmatter with `status: seed`, `pipeline_phase: agent-readiness`, sections for Phase Goal, Context, What This Phase Must Produce, Prior Phase Outputs, Codebase Pointers
- Phase Goal: invoke compound-engineering:cli-agent-readiness-reviewer, implement fixes, iterate until no Blockers/Frictions
- What This Phase Must Produce: reviewer scorecard, fix implementation log, pass/warn/degrade verdict
- Prior Phase Outputs: runtime verification results from Phase 4.8

**Patterns to follow:**
- Existing seed templates in `seedTemplates` map (e.g., `PhaseReview`, `PhaseComparative`)
- Test assertions in `seeds_test.go` (must contain required sections, no `## Implementation Units`)

**Test scenarios:**
- Happy path: `RenderSeed(PhaseAgentReadiness, data)` returns valid content with all required sections
- Happy path: `TestRenderSeedProducesThinSeedTemplates` passes with the new phase in `PhaseOrder`
- Edge case: Template renders correctly with SeedData fields populated (APIName, OutputDir, etc.)

**Verification:**
- `go test ./internal/pipeline/...` passes
- Seed template contains `## What This Phase Must Produce` and `## Prior Phase Outputs`

---

- [ ] **Unit 4: Verify planner.go default fallback**

**Goal:** Confirm the default fallback in `GenerateNextPlan()` handles `PhaseAgentReadiness` correctly

**Requirements:** R1

**Dependencies:** Unit 2, Unit 3

**Files:**
- Modify: `internal/pipeline/planner.go` (potentially no changes needed)
- Test: `internal/pipeline/planner_test.go` (if it exists)

**Approach:**
- Read `GenerateNextPlan()` switch statement and verify no explicit case is needed
- The `default` branch calls `RenderSeed(nextPhase, ctx.SeedData)` which will work once the seed template exists (Unit 3)
- If the default branch has any side effects or assumptions that don't apply to the agent-readiness phase, add an explicit case
- Add a comment in the default case noting that `PhaseAgentReadiness` uses the static seed intentionally

**Patterns to follow:**
- Existing default case comment: "Preflight, Research use static seeds"

**Test scenarios:**
- Happy path: `GenerateNextPlan()` with `PhaseAgentReadiness` returns the seed template content
- Integration: Full `Init()` → `NextPhase()` → `GenerateNextPlan()` cycle reaches `PhaseAgentReadiness` without error

**Verification:**
- `go test ./internal/pipeline/...` passes
- No panics or errors when `PhaseAgentReadiness` is the next phase

---

- [ ] **Unit 5: Add Phase 4.9 to SKILL.md**

**Goal:** Add LLM-agent instructions for Phase 4.9 between Phase 4.8 and Phase 5, plus update holistic references

**Requirements:** R1, R2, R3, R4, R5, R7

**Dependencies:** Units 1-4

**Files:**
- Modify: `skills/printing-press/SKILL.md`

**Approach:**

The SKILL.md changes are in three parts:

**Part A — Phase 4.9 section (insert between Phase 4.8 gate and Phase 5 header):**
- Follow Phase 4.8's structure exactly: `# PHASE 4.9: AGENT READINESS REVIEW LOOP`, `## THIS PHASE IS MANDATORY. DO NOT SKIP IT.`, then step labels 4.9a through 4.9d
- Step 4.9a: Check agent availability — attempt to dispatch `compound-engineering:cli-agent-readiness-reviewer` with a test invocation or check if the agent is listed. If unavailable, log warning and skip to Phase 5 (ONLY exception path)
- Step 4.9b: Run reviewer — dispatch the agent scoped to the generated CLI folder with the exact prompt template from the requirements doc
- Step 4.9c: Implement fixes — for each fix in the ranked list: read file:line, make change, `go build ./... && go vet ./...`, revert on failure. All fixes attempted
- Step 4.9d: Termination check — evaluate re-run scorecard. Zero Blockers/Frictions → proceed. Remaining + pass < 2 → repeat from 4.9c. Pass = 2 → log known issues, proceed
- Phase Gate 4.9: pass/warn/degrade table, "Tell the user:" message with verdict, Blocker/Friction/Optimization counts

**Part B — Holistic SKILL.md updates:**
- Frontmatter: bump version from 1.1.0 to 1.2.0
- Phase map diagram (line 233): insert `PHASE 4.9` between `PHASE 4.8` and `PHASE 5` with `(5-10m)` time budget
- Phase count: update "8 mandatory phases" to "9 mandatory phases" (line 230)
- Total time: adjust from "45-85 minutes" to "50-95 minutes" (line 240)

**Part C — Graceful degradation language:**
- Use "THIS PHASE IS MANDATORY. DO NOT SKIP IT." header
- Immediately follow with "**Exception:** Skip ONLY if the `compound-engineering:cli-agent-readiness-reviewer` agent is unavailable (plugin not installed or marketplace not registered). Log a warning and proceed to Phase 5."
- Do not use soft language ("you may skip") — the past fix for skill loop enforcement shows the LLM will aggressively shortcut

**Patterns to follow:**
- Phase 4.8 structure (lines 1595-1648): mandatory header, step labels, gate table, "Tell the user:" message
- Phase 4.5 dogfood emulation: closest analog for an iterative fix loop within a phase

**Test scenarios:**
- Happy path: Phase 4.9 section exists between Phase 4.8 and Phase 5 with all 4 steps
- Happy path: Phase map diagram includes Phase 4.9
- Happy path: Frontmatter version is 1.2.0
- Edge case: Graceful skip path is the ONLY exception and uses "Skip ONLY if" language
- Edge case: Phase gate covers all three verdicts (Pass/Warn/Degrade)

**Verification:**
- SKILL.md contains `# PHASE 4.9: AGENT READINESS REVIEW LOOP` between Phase 4.8 gate and Phase 5 header
- Phase map diagram shows Phase 4.9 with time budget
- Frontmatter version is bumped
- `compound-engineering:cli-agent-readiness-reviewer` is referenced with full namespaced name

## System-Wide Impact

- **Interaction graph:** Phase 4.9 invokes an external Claude Code plugin agent (`compound-engineering:cli-agent-readiness-reviewer`). The agent reads generated CLI files and produces a report. The SKILL.md instructions then drive fix implementation. No callbacks, middleware, or observers are affected.
- **Error propagation:** Agent unavailability → graceful skip. Agent output unparseable → graceful skip. Individual fix build failure → revert + skip. These are all local to Phase 4.9 and don't affect other phases.
- **State lifecycle risks:** PlanPath format change affects all phases, not just the new one. Migration must update paths for all existing phases. New pipelines get name-based paths from the start.
- **API surface parity:** The `printing-press` CLI itself is not affected. The emboss mode (second-pass improvement) could also benefit from agent-readiness review in the future but is explicitly out of scope.

## Risks & Dependencies

- **PlanPath format change is broader than Phase 4.9** — While motivated by the new phase, switching from index-based to name-based PlanPath affects all existing and future pipelines. Mitigation: the change is straightforward (remove the `%02d-` prefix) and migration handles existing state files.
- **External plugin dependency** — compound-engineering v2.55.0+ must be installed and `every-marketplace` registered. Mitigation: graceful skip with diagnostic warning message.
- **LLM may skip the phase** — Past experience shows Claude treats phase gates as suggestions. Mitigation: "THIS PHASE IS MANDATORY" header with only one explicit exception path.

## Sources & References

- **Origin document:** [docs/brainstorms/2026-03-27-agent-readiness-review-loop-requirements.md](docs/brainstorms/2026-03-27-agent-readiness-review-loop-requirements.md)
- Related code: `internal/pipeline/state.go`, `internal/pipeline/seeds.go`, `internal/pipeline/planner.go`, `internal/pipeline/pipeline.go`
- Related skill: `skills/printing-press/SKILL.md` (Phase 4.8 as structural analog)
- External dependency: `compound-engineering:cli-agent-readiness-reviewer` agent in compound-engineering plugin v2.55.0+
