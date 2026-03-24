---
title: "Autonomous CLI Pipeline - Plan-First, Phase-Chained, From API Name to Finished CLI"
type: feat
status: active
date: 2026-03-24
---

# Autonomous CLI Pipeline - Plan-First, Phase-Chained

## The Philosophy

"The moment I have an idea, it's /ce:plan." - @mvanhorn

Traditional dev is 80% coding, 20% planning. This flips it. The thinking happens in the plan. The execution is mechanical. Every phase of the pipeline follows the same discipline: write a plan.md first, execute it with ce:work, then write the next plan. No phase starts work without a plan. No plan ships without execution.

The pipeline is not a Go binary that runs 7 functions. It's a chain of plan.md files, each one written by ce:plan with full research agents, each one executed by ce:work with test verification. The Go binary is thin - it manages pipeline state and triggers the next plan. The intelligence lives in the plans.

This is the same pattern that produced 70 plan files and 263 commits on /last30days. The same pattern that runs OSC nightnight for 8 hours overnight. The same pattern that turned a lunch transcript into a product proposal.

## Overview

User says: `printing-press print gmail`

What happens:

```
Phase 0: PREFLIGHT
  -> ce:plan writes preflight-plan.md (check environment, find spec)
  -> ce:work executes it
  -> checkpoint updated

Phase 1: SCAFFOLD
  -> ce:plan writes scaffold-plan.md (generate initial CLI from spec)
  -> ce:work executes it (runs printing-press generate, validates 7 gates)
  -> checkpoint updated

Phase 2: ENRICH
  -> ce:plan writes enrich-plan.md (deep-read spec, research API docs, find hints)
  -> ce:work executes it (produces spec overlay YAML)
  -> checkpoint updated

Phase 3: REGENERATE
  -> ce:plan writes regenerate-plan.md (merge overlay, re-generate, re-validate)
  -> ce:work executes it
  -> checkpoint updated

Phase 4: REVIEW
  -> ce:plan writes review-plan.md (dogfood every command, check quality)
  -> ce:work executes it (produces quality score + issue list)
  -> checkpoint updated

Phase 5: SHIP
  -> ce:plan writes ship-plan.md (git init, README, morning report)
  -> ce:work executes it
  -> Done.
```

Each phase is a full ce:plan -> ce:work cycle. Each plan.md gets the benefit of parallel research agents, codebase analysis, and prior learnings. Each ce:work execution gets fresh context, task tracking, and test verification.

The plan.md IS the checkpoint that survives everything. Context gets lost? Start a new session, point it at the plan, pick up where you left off.

## Why Plan-First Per Phase

1. **Fresh context per phase.** ENRICH needs to deeply analyze an API's docs. That's a full context window. If it ran in the same session as SCAFFOLD, context would be exhausted. A new ce:plan session starts clean.

2. **Research agents per phase.** ce:plan launches parallel research agents - repo analysis, learnings search, external docs. DISCOVER's research is different from ENRICH's research is different from REVIEW's research. Each phase needs its own research pass.

3. **Human checkpoint between phases.** Each plan.md is reviewable. In autonomous mode, the mechanical gate reviews it. In interactive mode, the human can read the plan before ce:work starts. "Does this enrichment plan look right before I spend 20 minutes executing it?"

4. **Crash recovery is free.** If a session dies mid-ENRICH, the enrich-plan.md is already on disk. Next session reads it and picks up. No special checkpoint serialization needed - the plan IS the state.

5. **Compounding context.** Each plan.md can reference all prior plans. The REVIEW plan can say "In the scaffold plan, we generated 12 resources. In the enrich plan, we added defaults to 3 of them. Now let's verify all 12 look right." Plans compound like your strategy docs compound.

## Architecture

### The Thin Orchestrator (Go)

The `printing-press print` command is a thin loop:

```go
func RunPipeline(apiName string, opts Options) error {
    state := loadOrCreateState(apiName)

    phases := []string{"preflight", "scaffold", "enrich", "regenerate", "review", "ship"}

    for _, phase := range phases {
        if state.IsCompleted(phase) {
            continue // resume mode - skip done phases
        }

        // Write the plan for this phase
        planPath := state.PlanPath(phase) // e.g., docs/plans/gmail-pipeline/02-scaffold-plan.md
        if !fileExists(planPath) {
            // Generate phase-specific prompt and write it as the plan seed
            writePlanSeed(state, phase, planPath)
        }

        // Mark phase in progress
        state.Start(phase)
        state.Save()

        fmt.Fprintf(os.Stderr, "[%s] Plan at %s\n", phase, planPath)
        fmt.Fprintf(os.Stderr, "[%s] Execute with: /ce:work %s\n", phase, planPath)

        // In autonomous mode: the skill chains ce:plan -> ce:work automatically
        // In interactive mode: the user runs ce:work manually
    }

    return nil
}
```

The Go binary doesn't execute phases. It writes plan seeds (structured prompts with all the context each phase needs) and manages state. The Claude Code skill does the actual ce:plan -> ce:work chaining.

### The Claude Code Skill (The Real Engine)

The printing-press skill becomes the orchestrator:

```
User: "print me a gmail CLI"

Skill reads pipeline state
  -> If no state: create it, start at preflight
  -> If existing state: find next incomplete phase

For each phase:
  1. Run ce:plan with phase-specific prompt
     - Prompt includes: API name, prior phase outputs, conventions cache
     - ce:plan research agents analyze the specific needs of THIS phase
     - ce:plan writes the plan.md

  2. Run ce:work on the plan.md
     - ce:work breaks it into tasks, implements, tests, checks off criteria
     - On completion: update pipeline state

  3. Chain to next phase
     - CronCreate schedules the skill to resume in 30 seconds
     - Fresh session, fresh context, next phase
```

This is exactly OSC nightnight's pattern: implement -> checkpoint -> chain -> fresh session -> implement.

### Plan Directory Structure

```
docs/plans/gmail-pipeline/
  00-preflight-plan.md     # Environment checks, spec discovery
  01-scaffold-plan.md      # Initial CLI generation
  02-enrich-plan.md        # Deep spec analysis, overlay creation
  03-regenerate-plan.md    # Merge overlay, re-generate
  04-review-plan.md        # Quality analysis, dogfooding
  05-ship-plan.md          # Finalize, morning report
  state.json               # Pipeline state (which phases are done)
  conventions.json         # Cached spec analysis (reused across phases)
  overlay.yaml             # Spec overlay from ENRICH (consumed by REGENERATE)
  report.md                # Final morning report
```

Each plan.md follows the standard ce:plan format with frontmatter, acceptance criteria, implementation units. They're real plans, not stubs.

### Phase-Specific Plan Seeds

The Go binary writes "plan seeds" - structured prompts that ce:plan uses to write the full plan. Each seed contains:

**Preflight seed:**
- Check Go version, check press binary, check disk space
- Find OpenAPI spec: check known-specs registry, then apis-guru (`APIs-guru/openapi-directory/APIs/<provider>/`), then WebSearch
- Validate spec quality (parses, has 3+ endpoints, has base URL)
- Write conventions cache (auth type, pagination patterns, resource count)
- Acceptance: spec file downloaded, conventions cached, environment validated

**Scaffold seed:**
- Spec path from preflight
- Run `printing-press generate --spec <path> --output <dir>`
- Validate 7 quality gates pass
- Acceptance: CLI compiles, all gates pass, `<cli> --help` works

**Enrich seed:**
- Read original spec from scaffold
- Read conventions cache from preflight
- Deep-read every parameter description for default value hints (e.g., "The special value `me` can be used")
- Scan for `mediaUpload`, `x-google-*` extensions, multipart content types
- Scan response schemas for sync token patterns (`historyId`, `syncToken`)
- Scan for batch endpoint patterns (`:batchGet`, `:batchCreate`)
- Check for endpoints with empty descriptions that could be enriched
- WebSearch for `<api-name> API best practices CLI` to find community patterns
- Write overlay.yaml with all discovered enrichments
- Acceptance: overlay file exists, at least 1 enrichment found

**Regenerate seed:**
- Merge overlay.yaml with original spec
- Re-run `printing-press generate` with merged spec
- Re-validate 7 quality gates
- If gates fail: fall back to original spec, log what broke
- Acceptance: CLI compiles with enrichments, no regressions

**Review seed:**
- Build the CLI binary
- Run `<cli> --help` and `<cli> <resource> --help` for every resource
- Check: no command name > 40 chars, no empty descriptions on resources
- Check: every GET endpoint has at least 1 query param flag
- Check: no duplicate command names
- Check: `<cli> doctor` exits 0
- Check: binary size < 50MB
- Produce quality score (0-100) and issue list
- Acceptance: score > 70, all critical checks pass

**Ship seed:**
- Run `git init` in output directory
- Create initial commit
- Validate goreleaser config if present
- Write morning report with: time per phase, enrichments applied, quality score, next steps for human
- Acceptance: git repo initialized, morning report written

## Scope Boundaries

- No real API testing (no credential management in autonomous mode)
- No GitHub repo creation or release publishing
- No Google Discovery format converter (use apis-guru)
- No file upload/download generation
- No batch mode (one API per invocation)
- No Codex delegation for plan writing (ce:plan runs on Claude)

## Implementation Units

### Unit 1: Pipeline State Manager (Go)

**Goal:** The thin Go package that manages pipeline state and writes plan seeds.

**Files:**
- New: `internal/pipeline/state.go` - state types, read/write, phase tracking
- New: `internal/pipeline/seeds.go` - plan seed templates per phase
- New: `internal/pipeline/state_test.go`

**Approach:**

State is minimal JSON:
```go
type PipelineState struct {
    APIName    string                 `json:"api_name"`
    OutputDir  string                 `json:"output_dir"`
    StartedAt  time.Time              `json:"started_at"`
    Phases     map[string]PhaseState  `json:"phases"`
    SpecPath   string                 `json:"spec_path,omitempty"`
}

type PhaseState struct {
    Status   string `json:"status"` // pending, planned, executing, completed, failed
    PlanPath string `json:"plan_path,omitempty"`
}
```

Note the extra status: `planned` means the plan.md exists but hasn't been executed yet. This is the human checkpoint - you can review the plan before ce:work runs.

Plan seeds are Go templates that produce markdown. Each seed is a structured prompt that ce:plan will expand into a full plan with research.

**Verification:** `go test ./internal/pipeline/...`

---

### Unit 2: `print` CLI Command

**Goal:** Add `printing-press print <api-name>` command that creates the pipeline directory and writes plan seeds.

**Files:**
- Update: `internal/cli/root.go` - add print command
- New: `internal/pipeline/pipeline.go` - orchestration logic

**Approach:**

The print command:
```
printing-press print gmail [--output ./gmail-cli] [--force] [--resume]
```

It does NOT execute phases. It:
1. Creates `docs/plans/<api-name>-pipeline/` directory
2. Writes `state.json` with all phases pending
3. Writes plan seeds for each phase as markdown files
4. Prints instructions: "Pipeline ready. Run /ce:work docs/plans/gmail-pipeline/00-preflight-plan.md to start"

In skill mode (autonomous), the skill takes over from here and chains ce:plan -> ce:work for each phase.

**Verification:** `go test ./...`, `printing-press print petstore` creates pipeline directory with plan seeds

---

### Unit 3: Claude Code Skill Update

**Goal:** Update the printing-press skill to support the autonomous pipeline with phase chaining.

**Files:**
- Update: `skills/printing-press/SKILL.md`

**Approach:**

New workflow in the skill:

```
Workflow: Autonomous Pipeline

When user says "print <api-name>":

1. Run `printing-press print <api-name>` to create pipeline directory
2. Read state.json to find next phase
3. For the next incomplete phase:
   a. If plan.md doesn't exist yet:
      - Read the plan seed
      - Run ce:plan with the seed as input (this writes the full plan)
   b. If plan.md exists but phase isn't completed:
      - Run ce:work on the plan.md
   c. After ce:work completes:
      - Update state.json (mark phase completed)
      - CronCreate to chain back to this skill in 30 seconds
      - Fresh session picks up at step 2 with next phase

Budget gate: after each phase, check total elapsed time. If > 3 hours, write
partial morning report and stop. This prevents runaway overnight sessions.
```

The skill is the brain. The Go binary is the skeleton. The plans are the memory.

**Verification:** Manual test in Claude Code: "print me a petstore CLI"

---

### Unit 4: Known-Specs Registry + apis-guru Integration

**Goal:** Make PREFLIGHT's spec discovery actually work for common APIs.

**Files:**
- Update: `skills/printing-press/references/known-specs.md` - add apis-guru patterns
- New: `internal/pipeline/discover.go` - spec discovery logic (used by plan seed generation)

**Approach:**

Expand known-specs.md with apis-guru URL patterns:
```
## apis-guru patterns
For any API not in the registry, try:
https://raw.githubusercontent.com/APIs-guru/openapi-directory/main/APIs/{provider}/{version}/openapi.yaml

Common providers:
- googleapis.com/gmail/v1
- googleapis.com/calendar/v3
- googleapis.com/drive/v3
- stripe.com/2023-10-16
- twilio.com/2010-04-01
```

The discover.go file provides a `DiscoverSpec(apiName string) (string, error)` function that:
1. Checks known-specs registry (exact match)
2. Tries apis-guru URL pattern
3. Returns the URL or path to the spec

This is used by the preflight plan seed to tell ce:plan where to look.

**Verification:** `go test ./internal/pipeline/...` with test cases for gmail, stripe, petstore

---

### Unit 5: Spec Overlay Types and Merge

**Goal:** Define the overlay format that ENRICH produces and REGENERATE consumes.

**Files:**
- New: `internal/pipeline/overlay.go` - overlay types
- Update: `internal/spec/spec.go` - add MergeOverlay method

**Approach:**

The overlay is a YAML file that mirrors APISpec but every field is optional:
```go
type SpecOverlay struct {
    Resources map[string]ResourceOverlay `yaml:"resources,omitempty"`
}

type ResourceOverlay struct {
    Endpoints map[string]EndpointOverlay `yaml:"endpoints,omitempty"`
}

type EndpointOverlay struct {
    Description *string      `yaml:"description,omitempty"`
    Params      []ParamPatch `yaml:"params,omitempty"`
}

type ParamPatch struct {
    Name    string  `yaml:"name"`
    Default *string `yaml:"default,omitempty"`
}
```

`MergeOverlay` on APISpec applies the overlay: non-nil fields override originals.

**Verification:** `go test ./internal/pipeline/...`, `go test ./internal/spec/...`

---

### Unit 6: End-to-End Test

**Goal:** Run the full pipeline on Petstore to prove it works.

**Files:** No new files - verification only.

**Approach:**
1. `printing-press print petstore`
2. Verify pipeline directory created with plan seeds
3. Manually run ce:work on each plan seed in order
4. Verify final CLI compiles and works
5. Verify morning report generated

**Verification:** Petstore CLI works, all plan.md files exist, morning report is readable

## Dependencies

```
Unit 1 (State) - foundation
Unit 2 (print command) - depends on Unit 1
Unit 3 (Skill) - depends on Unit 2
Unit 4 (Discovery) - independent
Unit 5 (Overlay) - independent
Unit 6 (E2E test) - depends on all
```

Units 1, 4, 5 can be built in parallel.
Unit 2 wires Unit 1 into the CLI.
Unit 3 is the skill layer.
Unit 6 is verification.

## What "Done" Looks Like

```bash
# Two words. That's the input.
printing-press print gmail

# Pipeline creates:
# docs/plans/gmail-pipeline/
#   00-preflight-plan.md    <- ce:plan writes this with full research
#   01-scaffold-plan.md     <- ce:plan writes this after preflight completes
#   02-enrich-plan.md       <- ce:plan deep-analyzes the spec and API docs
#   03-regenerate-plan.md   <- ce:plan plans the overlay merge
#   04-review-plan.md       <- ce:plan plans the quality review
#   05-ship-plan.md         <- ce:plan plans the finalization
#   state.json              <- tracks which phases are done
#   conventions.json        <- cached spec analysis
#   overlay.yaml            <- enrichments from phase 2
#   report.md               <- morning report

# Each plan is a real plan. Research agents. Acceptance criteria. Implementation units.
# Each plan gets executed by ce:work with task tracking and test verification.
# Each phase chains to the next via CronCreate. Fresh context every time.

# Output:
cd gmail-cli && go build -o ./gmail ./cmd/gmail-cli
./gmail auth login --client-id $GOOGLE_CLIENT_ID
./gmail messages list me --limit 5    # 'me' is the default, detected from spec hints
./gmail doctor                        # colored output, all green
```

That's not a scaffold. That's a finished CLI. From two words. Built by 6 plans, each one meticulous, each one researched, each one executed and verified. The same workflow that produces 70 plan files and 263 commits.

Plan, execute, plan, execute. All the way down.

## Sources

- Matt Van Horn, "Every Claude Code Hack I Know (March 2026)" - the plan-first philosophy
- OSC nightnight: `~/.claude/skills/osc-nightnight/SKILL.md` - session chaining, budget gates, CronCreate
- OSC work: `~/.claude/skills/osc-work/SKILL.md` - phased execution, conventions cache
- Compound Engineering ce:plan - parallel research agents, structured plans
- Press generator: `internal/generator/generator.go` - 14 templates, 7 quality gates
- Press parser: `internal/openapi/parser.go` - spec extraction
- apis-guru: `github.com/APIs-guru/openapi-directory` - 2,000+ pre-converted specs
