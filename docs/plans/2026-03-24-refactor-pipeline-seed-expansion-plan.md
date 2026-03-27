---
title: "Refactor Pipeline Seeds to Support ce:plan Expansion"
type: refactor
status: completed
date: 2026-03-24
origin: docs/plans/2026-03-24-feat-autonomous-cli-pipeline-plan.md
---

# Refactor Pipeline Seeds to Support ce:plan Expansion

## Problem Statement

The autonomous pipeline plan (2026-03-24-feat-autonomous-cli-pipeline-plan.md) describes a two-step process per phase:

1. **ce:plan expands the seed** into a full plan with research agents
2. **ce:work executes** the expanded plan

But the current seeds in `internal/pipeline/seeds.go` are already structured like mini-plans - they have implementation units with concrete bash commands, acceptance criteria, and specific file paths. There's nothing for ce:plan to expand. Running ce:plan on them would either:
- Redundantly research what the seed already specifies
- Rewrite the seed and lose the templated context (output dir, spec URL, etc.)

The seeds need to become thinner - just goals + context pointers - so ce:plan can do its job of writing the real plan with research.

## How It Should Work

```
printing-press print petstore
  -> Creates 6 THIN seeds + state.json
  -> Each seed has: phase goal, context from prior phases, pointers to artifacts

Workflow 4 loop (per phase):
  1. Read thin seed
  2. ce:plan expands seed into full plan.md (with research agents, implementation units, verification)
  3. Overwrite the seed file with the expanded plan
  4. ce:work executes the expanded plan
  5. Update state.json -> "completed"
  6. CronCreate -> fresh session for next phase
```

The key insight from the philosophy doc: "The intelligence lives in the plans." Seeds are prompts, not plans. ce:plan adds the intelligence.

## Proposed Solution

### Thin Seed Format

Each seed becomes a structured prompt for ce:plan, not a mini-plan for ce:work:

```markdown
---
title: "petstore CLI Pipeline - Phase 0: Preflight"
type: feat
status: seed
date: 2026-03-24
pipeline_phase: preflight
pipeline_api: petstore
---

# Phase Goal

Validate the environment and discover the OpenAPI spec for petstore.

## Context

- Pipeline directory: docs/plans/petstore-pipeline/
- Output directory: /tmp/petstore-pipeline-test
- Spec URL (from registry): https://petstore3.swagger.io/api/v3/openapi.json
- Spec source: known-specs

## What This Phase Must Produce

- Verified Go environment (1.23+)
- Verified printing-press binary compiles
- Downloaded and validated OpenAPI spec
- conventions.json written to pipeline directory

## Prior Phase Outputs

(none - this is the first phase)

## Codebase Pointers

- Build command: `go build -o ./printing-press ./cmd/printing-press`
- Generate command: `printing-press generate --spec <url> --output <dir>`
- Quality gates: internal/generator/quality.go
- Spec parser: internal/openapi/parser.go
```

Notice what's missing: no implementation units, no bash commands, no step-by-step instructions. That's ce:plan's job. The seed provides WHAT and WHERE, ce:plan figures out HOW.

### Sequential Expansion with Prior Phase Context

Later seeds reference artifacts from earlier phases:

```markdown
# Phase Goal (Scaffold)

Generate the initial CLI from the discovered OpenAPI spec.

## Prior Phase Outputs

- Preflight: conventions.json at docs/plans/petstore-pipeline/conventions.json
- Preflight: spec validated at https://petstore3.swagger.io/api/v3/openapi.json
```

```markdown
# Phase Goal (Enrich)

Deep-read the original spec for hints the parser missed. Research API docs.

## Prior Phase Outputs

- Preflight: conventions.json (auth type, resource count, pagination patterns)
- Scaffold: CLI generated at /tmp/petstore-pipeline-test (resource list, endpoint count)
```

Each phase's seed gets richer because it can point to what prior phases produced.

### Status Field Tracks Expansion

```
status: seed       -> written by printing-press print (thin prompt)
status: active     -> expanded by ce:plan (full plan with implementation units)
status: completed  -> executed by ce:work (all acceptance criteria met)
```

Workflow 4 checks the status field:
- If `seed`: run ce:plan to expand it
- If `active`: run ce:work to execute it
- If `completed`: skip to next phase

This handles crash recovery naturally. If ce:plan crashes mid-expansion, the status is still `seed` and it restarts. If ce:work crashes mid-execution, the status is `active` and it re-runs ce:work.

## Acceptance Criteria

- [ ] Seeds in `internal/pipeline/seeds.go` are thin prompts (goal + context + pointers, no implementation units)
- [ ] Each seed has `status: seed` in frontmatter
- [ ] Seeds include `pipeline_phase` and `pipeline_api` in frontmatter for programmatic access
- [ ] Later seeds include `Prior Phase Outputs` section with artifact paths
- [ ] SKILL.md Workflow 4 Step 3 checks `status:` field to decide ce:plan vs ce:work
- [ ] state.json `PhaseState` gains a `plan_status` field: `seed | expanded | completed`
- [ ] `printing-press print petstore` creates thin seeds (not mini-plans)
- [ ] ce:plan can expand a thin seed into a full plan (verified manually on preflight)
- [ ] Full pipeline runs: seed -> expand -> execute -> next seed -> expand -> execute -> ... -> report

## Implementation Units

### Unit 1: Rewrite Seed Templates

**Goal:** Replace the current detailed seeds with thin prompt seeds.

**Files:**
- `internal/pipeline/seeds.go` - rewrite all 6 templates
- `internal/pipeline/seeds_test.go` - update tests

**Approach:**
- Each template produces a structured ce:plan prompt, not a ce:work plan
- Include `status: seed` and `pipeline_phase:` in frontmatter
- Preflight seed has no "Prior Phase Outputs"
- Scaffold through Ship seeds have "Prior Phase Outputs" referencing prior artifacts
- "Codebase Pointers" section replaces implementation units - tells ce:plan WHERE to look, not WHAT to do

**Verification:** `go test ./internal/pipeline/...` passes, seed output is thin (no bash commands in implementation units)

### Unit 2: Add plan_status to State

**Goal:** Track whether a phase's plan has been expanded by ce:plan.

**Files:**
- `internal/pipeline/state.go` - add `PlanStatus` field to `PhaseState`

**Approach:**
- Add `PlanStatus string` to PhaseState: `seed`, `expanded`, `completed`
- `MarkPlanned()` becomes `MarkSeedWritten()` (sets `plan_status: seed`)
- New `MarkExpanded()` method (sets `plan_status: expanded`)
- `Complete()` also sets `plan_status: completed`
- `NextPhase()` returns the first phase where `plan_status != completed`

**Verification:** `go test ./internal/pipeline/...`

### Unit 3: Update SKILL.md Workflow 4

**Goal:** Workflow 4 phase loop checks plan status to decide ce:plan vs ce:work.

**Files:**
- `skills/printing-press/SKILL.md` - update Step 3 (phase execution loop)

**Approach:**
Update Step 3 logic:
```
For each phase where plan_status != "completed":
  a. Read the plan file
  b. If status frontmatter is "seed":
     - Run ce:plan with the seed as input
     - ce:plan overwrites the file with the expanded plan (status: active)
     - Update state.json: plan_status = "expanded"
  c. If status frontmatter is "active" (already expanded):
     - Run ce:work on the plan
     - Update state.json: plan_status = "completed", phase status = "completed"
  d. Budget gate, CronCreate chain (unchanged)
```

**Verification:** Manual test with petstore - seed gets expanded, then executed

### Unit 4: Update Pipeline Init

**Goal:** `printing-press print` uses new thin seeds and sets initial plan_status.

**Files:**
- `internal/pipeline/pipeline.go` - update Init() to use new state fields

**Approach:**
- After writing each seed, call `MarkSeedWritten(phase)` instead of `MarkPlanned(phase)`
- state.json now shows `plan_status: seed` for all phases after init

**Verification:** `printing-press print petstore --force` creates state.json with `plan_status: seed` for all phases

## Scope Boundaries

- Don't change the 6-phase structure (preflight through ship)
- Don't change the CronCreate chaining mechanism
- Don't change budget gate logic
- Don't add new phases
- Don't change the overlay system
- Don't change the generator or parser

## Dependencies

- Autonomous pipeline plan (completed) - provides the architecture this implements
- Nightnight chaining plan (completed) - CronCreate mechanism is unchanged
- Dogfood phase plan (completed) - review seed content changes but dogfood tiers don't

## Sources

- **Origin:** [docs/plans/2026-03-24-feat-autonomous-cli-pipeline-plan.md](docs/plans/2026-03-24-feat-autonomous-cli-pipeline-plan.md) - "The plan.md IS the checkpoint", "They're real plans, not stubs", "ce:plan writes the full plan"
- Seed templates: `internal/pipeline/seeds.go`
- State machine: `internal/pipeline/state.go`
- Pipeline init: `internal/pipeline/pipeline.go`
- SKILL.md Workflow 4: `skills/printing-press/SKILL.md:209-307`
- Nightnight chaining: `docs/plans/2026-03-24-feat-pipeline-nightnight-chaining-plan.md`
